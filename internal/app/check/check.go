// Package check implements `odachi check`: the four gates — brief, fence, open
// questions and proof — in the order `gate.Evaluate` documents, so the first
// unanswered thing is always the one the human sees.
//
// It is pure orchestration over three ports: Files (the disk), Git (the tree and
// what changed) and Shell (the proof command). It reads nothing else and writes
// nothing except the proof that `--run` creates. The domain packages own the
// rules: this package only gathers the facts the gate needs.
package check

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/fence"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/proof"
	reviewdomain "github.com/biggs-100/odachi/internal/domain/review"
)

// DefaultTimeout is how long the proof command may take when --timeout is not
// given, in seconds. A gate that can hang is a gate nobody trusts.
const DefaultTimeout = 120

// EmptyTree is git's well-known hash of the empty tree. Comparing against it is
// what "no commits: everything uncommitted" means: every file is new because
// nothing was ever committed, and the fence still runs instead of failing.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// head is the default base: what the worktree is compared against.
const head = "HEAD"

// The paths this command reads, and the directory it may write into.
const (
	briefFile = "brief.md"
	fenceFile = ".odachi/fence.toml"
	proofDir  = "proof"
)

// The proof of a project with no feature yet: a project can be verified before
// it has one, so the evidence needs a name even then.
const (
	noFeatureNumber = 0
	noFeatureSlug   = "check"
)

// The three reasons the optional review gate can add. They are not gate reasons:
// the gate is untouched, and a review only ever participates when the human asks
// for it with --require-review.
const (
	// ReasonReviewMissing means this tree has no verdict yet.
	ReasonReviewMissing = "review_missing"
	// ReasonReviewChanges means the verdict for this tree requests changes, and
	// the run names the first finding that blocks and nobody refuted.
	ReasonReviewChanges = "review_changes_requested"
	// ReasonReviewLensMissing means the verdict exists but does not cover a lens
	// the evidence requires.
	ReasonReviewLensMissing = "review_lens_missing"
	// ReasonReviewRefutedUnreadable means the refutation beside the verdict cannot
	// be read, so nothing it might say can be taken into account.
	ReasonReviewRefutedUnreadable = "review_refutation_unreadable"
)

// nextReview is the continuation for a missing verdict, and the command that
// shows what a changes_requested verdict says.
const nextReview = "odachi review"

// Options is one `odachi check` invocation.
type Options struct {
	Dir string

	// Run is the command to run. Its exit code is never the exit code of
	// `check`: the gate decides, and a failing command is a failed proof.
	Run string
	// Proof is a proof file an agent or CI already wrote.
	Proof string
	// Base is what the fence is compared against. "HEAD" by default; an empty
	// value means "compare against nothing", which is the empty tree.
	Base string
	// TimeoutSeconds bounds Run. Zero or less means DefaultTimeout.
	TimeoutSeconds int
	// Quiet changes nothing about the result: the printer decides whether a
	// ready run says anything (the pre-commit hook depends on the silence).
	Quiet bool
	// RequireReview turns the optional design gate on: a stored verdict for this
	// exact tree, and not a changes_requested one. It is off by default because a
	// beginner must not be blocked by a verb they have never heard of.
	RequireReview bool
}

// Files is the disk port: enough to read the project and write one proof.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
	Write(path string, data []byte) error
}

// Git is the repository port: which tree is here, what changed since a base,
// which dependencies appeared, and whether the repository already ignores a path.
type Git interface {
	TreeHash(root string) (string, error)
	ChangedPaths(root, base string) ([]string, error)
	AddedDeps(root, base string) ([]string, error)
	// Ignored reports whether the repository's own rules already keep path out
	// of the tree. It is what makes the lazy ignore file unnecessary in a project
	// that ran `odachi init`.
	Ignored(root, path string) (bool, error)
}

// Shell is the command port. A non-zero exit is the command's result, not an
// error: err means the command could not be run at all.
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error)
}

// Result is what the CLI prints and exits with. It has the same shape as the
// other verbs; Reason and Next make the gate's decision visible to machines and
// to the pre-commit hook.
type Result struct {
	Code    int
	Reason  string
	Message string
	Next    string
	Lines   []string
	Proof   string
	Files   []string
}

// Run performs one `odachi check`.
func Run(opts Options, files Files, git Git, shell Shell) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.TimeoutSeconds <= 0 {
		opts.TimeoutSeconds = DefaultTimeout
	}

	r := &runner{opts: opts, files: files, git: git, shell: shell}

	if opts.Run != "" {
		if problem := r.ensureProofDir(); problem != nil {
			return *problem
		}
	}

	treeHash, err := git.TreeHash(opts.Dir)
	if err != nil {
		return blocked(fmt.Sprintf(
			"could not read the git tree: %s; run odachi check inside a git repository and check that git is installed",
			oneLine(err.Error())))
	}
	r.tree = treeHash

	state := gate.State{TreeHash: treeHash}

	hasBrief, missing, problem := r.brief()
	if problem != nil {
		return *problem
	}
	state.HasBrief = hasBrief
	state.BriefMissing = missing

	violations, problem := r.fenceViolations()
	if problem != nil {
		return *problem
	}
	state.Violations = violations

	feature, problem := r.currentFeature()
	if problem != nil {
		return *problem
	}
	if feature != nil {
		state.OpenQuestions = len(feature.Doc.OpenQuestions())
	}

	proofPath, proofRel := proofPathFor(opts.Dir, feature)
	r.proofRel = proofRel
	if opts.Proof != "" {
		r.proofRel = opts.Proof
	}

	recorded, haveProof, problem := r.recordedProof(proofPath)
	if problem != nil {
		return *problem
	}
	state.ProofTreeHash = recorded.Tree
	state.ProofOutput = recorded.Output
	state.ProofCommand = recorded.Command

	if opts.Run != "" {
		ran, problem := r.runProof(proofPath)
		if problem != nil {
			return *problem
		}
		state.ProofTreeHash = treeHash
		state.ProofOutput = ran.output
		state.ProofLastLine = proof.Proof{Output: ran.output}.LastLine()
		state.ProofFailed = ran.failed
		state.ProofTimedOut = ran.timedOut
		state.ProofTimeoutSeconds = opts.TimeoutSeconds
		state.ProofCommand = opts.Run
		haveProof = haveProof || (!ran.failed && !ran.timedOut)
	}

	verdict := gate.Evaluate(state)
	if opts.RequireReview && verdict.Code == gate.Ready {
		// The review gate is last on purpose: everything before it is something
		// the agent can act on, and sending an agent to `odachi review` while the
		// proof gate is still failing would be a dead end that resolves nothing.
		if problem := r.requireReview(); problem != nil {
			return *problem
		}
	}

	return r.report(verdict, haveProof)
}

// proofIgnore keeps the evidence out of the tree the evidence describes:
// without it, writing the proof changes the worktree, the next run computes a
// different tree hash and every proof invalidates the one before it. The ignore
// file itself is versioned, so the rule that keeps proofs out of git lives with
// the project instead of inside the tool.
const proofIgnore = "# odachi: proofs are generated, not versioned\n*\n!.gitignore\n"

// ensureProofDir keeps the evidence out of the tree the evidence describes:
// without an ignore rule, writing the proof changes the worktree, the next run
// computes a different tree hash and every proof invalidates the one before it.
//
// The rule is asked of the repository first. A project that ran `odachi init`
// carries the block in its own .gitignore, and there the per-directory file would
// only be a new path in the worktree — the very shift this function exists to
// avoid. A probe that cannot answer is NOT a reason to write nothing: evidence
// must never be allowed to move the tree because git failed, so the fallback is
// the per-directory file, exactly as before, written before the tree is hashed.
func (r *runner) ensureProofDir() *Result {
	if ignored, err := r.git.Ignored(r.opts.Dir, proofDir); err == nil && ignored {
		return nil
	}

	path := filepath.Join(r.opts.Dir, proofDir, ".gitignore")

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s/.gitignore: %s; check the permissions and run odachi check --run %q again",
			proofDir, oneLine(err.Error()), r.opts.Run))
		return &problem
	}
	if exists {
		return nil
	}

	if err := r.files.Write(path, []byte(proofIgnore)); err != nil {
		problem := blocked(fmt.Sprintf("could not prepare %s/: %s; check the permissions and run odachi check --run %q again",
			proofDir, oneLine(err.Error()), r.opts.Run))
		return &problem
	}
	return nil
}

// runner holds the facts of one run.
type runner struct {
	opts     Options
	files    Files
	git      Git
	shell    Shell
	tree     string
	proofRel string
}

// brief reads brief.md when it exists. A missing brief is a state, not an
// error: it is the first gate and it belongs to the human.
func (r *runner) brief() (bool, []string, *Result) {
	path := filepath.Join(r.opts.Dir, briefFile)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi check again", briefFile, oneLine(err.Error())))
		return false, nil, &problem
	}
	if !exists {
		return false, nil, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi check again", briefFile, oneLine(err.Error())))
		return false, nil, &problem
	}

	parsed := brief.Parse(string(data))
	return true, parsed.Missing(), nil
}

// fenceViolations runs the human's fence against what changed. A project with
// no fence has no rules and nothing to compare, so git is not asked at all.
func (r *runner) fenceViolations() ([]string, *Result) {
	rules, problem := r.fenceRules()
	if problem != nil {
		return nil, problem
	}
	if len(rules) == 0 {
		return nil, nil
	}

	base := baseRef(r.opts.Base)

	changed, err := r.git.ChangedPaths(r.opts.Dir, base)
	if err != nil && base == head {
		// No commit yet: everything uncommitted is the comparison, and the
		// empty tree says exactly that. A base the human named is never
		// silently replaced.
		changed, err = r.git.ChangedPaths(r.opts.Dir, EmptyTree)
	}
	if err != nil {
		problem := blocked(baseProblem(base, err))
		return nil, &problem
	}

	deps, err := r.git.AddedDeps(r.opts.Dir, base)
	if err != nil && base == head {
		deps, err = r.git.AddedDeps(r.opts.Dir, EmptyTree)
	}
	if err != nil {
		problem := blocked(baseProblem(base, err))
		return nil, &problem
	}

	violations := fence.Check(rules, changed, deps)
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Rule+": "+violation.What)
	}
	return messages, nil
}

// currentFeature is the highest-numbered task document: the feature the human
// is working on now. No document is not an error — a project with no features
// yet has no open questions and can still be verified.
func (r *runner) currentFeature() (*tasks.Document, *Result) {
	docs, err := tasks.All(r.files, r.opts.Dir)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; check the permissions of " + tasks.Dir + " and run odachi check again")
		return nil, &problem
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return &docs[len(docs)-1], nil
}

// recordedProof is the proof already on disk for this feature. Without --proof
// it is the feature's own file; an explicit --proof must carry the header,
// because a file that does not say which tree produced it proves nothing.
func (r *runner) recordedProof(proofPath string) (proof.Proof, bool, *Result) {
	if r.opts.Proof != "" {
		recorded, problem := r.readProof(explicitPath(r.opts.Dir, r.opts.Proof), r.opts.Proof)
		if problem != nil {
			return proof.Proof{}, false, problem
		}
		return recorded, true, nil
	}

	exists, err := r.files.Exists(proofPath)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check the proof %s: %s; check the permissions and run odachi check again", r.proofRel, oneLine(err.Error())))
		return proof.Proof{}, false, &problem
	}
	if !exists {
		return proof.Proof{}, false, nil
	}

	recorded, problem := r.readProof(proofPath, r.proofRel)
	if problem != nil {
		return proof.Proof{}, false, problem
	}
	return recorded, true, nil
}

// readProof reads and parses one proof file. A file that cannot be read or does
// not carry the header is refused: it cannot say which tree produced it, and
// guessing is the defect this tool exists to prevent.
func (r *runner) readProof(path, shown string) (proof.Proof, *Result) {
	data, err := r.files.Read(path)
	if err != nil {
		problem := refused(shown, "could not read it: "+oneLine(err.Error()))
		return proof.Proof{}, &problem
	}

	parsed, err := proof.Parse(string(data))
	if err != nil {
		problem := refused(shown, oneLine(err.Error()))
		return proof.Proof{}, &problem
	}
	return parsed, nil
}

// ranProof is the outcome of one `--run`.
type ranProof struct {
	output   string
	failed   bool
	timedOut bool
}

// runProof runs the proof command and records the proof when, and only when,
// the command succeeded: a failed command is not evidence. The exit code of the
// command never becomes the exit code of `check`.
func (r *runner) runProof(proofPath string) (ranProof, *Result) {
	output, exitCode, timedOut, err := r.shell.Run(r.opts.Dir, r.opts.Run, r.opts.TimeoutSeconds)
	if err != nil {
		problem := Result{
			Code:    int(gate.NeedsWork),
			Reason:  gate.ReasonProofFailed,
			Message: "could not run the proof: " + oneLine(err.Error()),
			Next:    runCommand(r.opts.Run),
			Lines:   []string{"could not run the proof: " + oneLine(err.Error()), "next: " + runCommand(r.opts.Run)},
		}
		return ranProof{}, &problem
	}

	if timedOut {
		return ranProof{output: output, timedOut: true}, nil
	}
	if exitCode != 0 {
		return ranProof{output: output, failed: true}, nil
	}

	recorded := proof.Proof{Tree: r.tree, Command: r.opts.Run, Output: output}
	if err := r.files.Write(proofPath, []byte(proof.Render(recorded))); err != nil {
		problem := blocked(fmt.Sprintf("could not write the proof %s: %s; check the permissions and run odachi check --run %q again",
			r.proofRel, oneLine(err.Error()), r.opts.Run))
		return ranProof{}, &problem
	}
	return ranProof{output: output}, nil
}

// requireReview is the optional design gate: a stored verdict for THIS tree, one
// that covers every lens the candidate requires, and no causal BLOCKER or
// CRITICAL finding left standing after the refutation is taken into account.
//
// The lenses come first because they answer whether this candidate asks for a
// review at all. A documentation-only candidate has nothing for a specialist lens
// to look at, so requiring a verdict would be ceremony, not evidence.
//
// It writes nothing. The store's ignore file is created by the verbs that produce
// a verdict, so a tree with no verdict has no store to prepare either.
func (r *runner) requireReview() *Result {
	required, problem := r.requiredLenses()
	if problem != nil {
		return problem
	}
	if len(required) == 0 {
		return nil
	}

	rel, err := reviewdomain.Path(r.tree)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; check the git tree and run odachi check --require-review again")
		return &problem
	}

	data, err := r.files.Read(filepath.Join(r.opts.Dir, filepath.FromSlash(rel)))
	if err != nil {
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewMissing,
			Message: "no review for this tree",
			Next:    nextReview,
			Lines:   []string{"no review for this tree", "next: " + nextReview},
		}
		return &problem
	}

	verdict, err := reviewdomain.Parse(data)
	if err != nil {
		problem := Result{
			Code:   int(gate.NeedsDecision),
			Reason: ReasonReviewMissing,
			Message: fmt.Sprintf("this tree's review cannot be read (%s): %s; write it again with odachi review --attach <file>",
				rel, oneLine(err.Error())),
			Next: nextReview,
		}
		problem.Lines = []string{problem.Message, "next: odachi review --attach <file>"}
		return &problem
	}

	if missing := verdict.Missing(required); len(missing) > 0 {
		problem := lensMissing(missing, rel)
		return &problem
	}

	refutation, problem := r.refutation(verdict)
	if problem != nil {
		return problem
	}

	if finding, blocking := reviewdomain.BlockingAfter(verdict.Findings, refutation); blocking {
		message := "the review asks for changes: " + oneLine(reviewFinding(finding))
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewChanges,
			Message: message,
			Next:    "odachi review --show",
			Lines: []string{
				message,
				"next: odachi review --show",
			},
		}
		return &problem
	}

	return nil
}

// requiredLenses is the lens selection the candidate asks for. It is the same
// question the review verb answers before printing a prompt, asked from the same
// changed paths, so the two can never disagree about what this tree needs.
func (r *runner) requiredLenses() ([]string, *Result) {
	base := baseRef(r.opts.Base)

	changed, err := r.git.ChangedPaths(r.opts.Dir, base)
	if err != nil && base == head {
		changed, err = r.git.ChangedPaths(r.opts.Dir, EmptyTree)
	}
	if err != nil {
		problem := blocked(baseProblem(base, err))
		return nil, &problem
	}

	return reviewdomain.Lenses(reviewdomain.Classify(changed)), nil
}

// refutation reads the refutation filed beside this tree's verdict. A tree with
// no refutation is the normal case and is not an error; a file that cannot be
// read is, because silence would let a broken artifact look like an approval.
func (r *runner) refutation(verdict reviewdomain.Verdict) (reviewdomain.Refutation, *Result) {
	rel, err := reviewdomain.RefutationPath(r.tree)
	if err != nil {
		return reviewdomain.Refutation{}, nil
	}

	data, err := r.files.Read(filepath.Join(r.opts.Dir, filepath.FromSlash(rel)))
	if err != nil {
		return reviewdomain.Refutation{}, nil
	}

	parsed, err := reviewdomain.ParseRefutation(data)
	if err != nil {
		detail := ""
		if finding, blocking := reviewdomain.Blocking(verdict.Findings); blocking {
			detail = "; the finding " + reviewFinding(finding) + " is still blocking"
		}

		message := fmt.Sprintf("this tree's refutation cannot be read (%s): %s%s", rel, oneLine(err.Error()), detail)
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewRefutedUnreadable,
			Message: message,
			Next:    "odachi review --show",
			Lines:   []string{message, "next: odachi review --show"},
		}
		return reviewdomain.Refutation{}, &problem
	}

	return parsed, nil
}

// lensMissing is the block for a verdict that does not cover a required lens.
// The store cannot be rewritten, so the message says which file to remove and
// which command produces a new verdict; the gate never deletes anything itself.
func lensMissing(missing []string, rel string) Result {
	message := fmt.Sprintf("missing the %s lens in this tree's review (%s)", missing[0], rel)
	return Result{
		Code:    int(gate.NeedsDecision),
		Reason:  ReasonReviewLensMissing,
		Message: message,
		Next:    nextReview,
		Lines: []string{
			message,
			"a verdict is not rewritten: delete " + rel + " and run " + nextReview + " --lens " + missing[0],
			"next: " + nextReview,
		},
	}
}

// reviewFinding is the one finding a human has to read first, in one line: where
// it is, how bad it is, and the evidence.
func reviewFinding(finding reviewdomain.Finding) string {
	where := finding.Path
	if finding.Line > 0 {
		where = fmt.Sprintf("%s:%d", finding.Path, finding.Line)
	}
	return where + " " + string(finding.Severity) + " " + finding.Why
}

// report turns the gate's verdict into the verb's result.
func (r *runner) report(verdict gate.Result, haveProof bool) Result {
	result := Result{
		Code:    int(verdict.Code),
		Reason:  verdict.Reason,
		Message: verdict.Message,
		Next:    verdict.Next,
	}

	if verdict.Code != gate.Ready {
		result.Lines = []string{verdict.Message, "next: " + verdict.Next}
		return result
	}

	result.Lines = []string{verdict.Message}
	if haveProof && r.proofRel != "" {
		result.Proof = r.proofRel
		result.Lines = append(result.Lines, "proof: "+r.proofRel)
	}
	return result
}

// proofPathFor names the proof of one feature: proof/NNNN-<slug>.txt, the same
// shape as the task document, so one feature has exactly one proof. Before the
// first feature the number is 0000 and the slug is stable, so --run and a later
// check still agree on the same file.
func proofPathFor(dir string, feature *tasks.Document) (path, rel string) {
	number, slug := noFeatureNumber, noFeatureSlug
	if feature != nil {
		number, slug = feature.Number, feature.Slug
	}

	name := fmt.Sprintf("%04d-%s.txt", number, slug)
	return filepath.Join(dir, proofDir, name), proofDir + "/" + name
}

// baseRef is the base the fence compares against: an empty --base means the
// empty tree, which is "compare against nothing".
func baseRef(base string) string {
	if strings.TrimSpace(base) == "" {
		return EmptyTree
	}
	return strings.TrimSpace(base)
}

// baseProblem says which base failed and how to fix it, in one line.
func baseProblem(base string, err error) string {
	return fmt.Sprintf("could not compare the fence against %q: %s; fix --base %q or drop the flag to compare against HEAD",
		base, oneLine(err.Error()), base)
}

// runCommand is the continuation that reruns the exact command.
func runCommand(command string) string {
	return `odachi check --run "` + strings.ReplaceAll(command, `"`, `\"`) + `"`
}

// refused is the result for a proof file the tool will not guess about. The
// message names --run because that is the command that writes a proof the tool
// can trust.
func refused(shown, reason string) Result {
	message := fmt.Sprintf(`the proof %s is not valid: %s; run odachi check --run "<command>" so the tool writes it`, shown, reason)
	return Result{
		Code:    int(gate.NeedsWork),
		Reason:  gate.ReasonProofMissing,
		Message: message,
		Next:    `odachi check --run "<command>"`,
		Lines:   []string{message, `next: odachi check --run "<command>"`},
	}
}

// blocked is a result the agent can act on, with the one command that unblocks
// it already in the message. It carries no gate reason: nothing was evaluated,
// the project could not be read.
func blocked(message string) Result {
	return Result{
		Code:    int(gate.NeedsWork),
		Message: message,
		Next:    "odachi check",
		Lines:   []string{message, "next: odachi check"},
	}
}

// explicitPath resolves a path the human typed. A relative one lives inside the
// project, so the command works the same from any directory.
func explicitPath(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
