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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/fence"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/proof"
	reviewdomain "github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
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
	fenceFile = layout.Fence
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
	// ReasonReviewCoverage means the verdict exists but its rows do not answer
	// about the bytes git measures now: stale receipts, missing rows, or the
	// old format that carries no receipts at all.
	ReasonReviewCoverage = "review_coverage"
	// ReasonReviewSelf means the verdict names the same model the proof names:
	// one agent approving the tree it wrote. Both identities are what each side
	// declared, so this is disclosure, not proof of who signed.
	ReasonReviewSelf = "review_self"
	// ReasonReviewUnattributed means the proof names a model and the verdict
	// names none, so the identity gate has nothing to compare: a silence would
	// otherwise be the cheapest way past it.
	ReasonReviewUnattributed = "review_unattributed"
)

// nextReview is the continuation for a missing verdict, and the command that
// shows what a changes_requested verdict says.
const nextReview = "odachi review"

// Options is one `odachi check` invocation.
type Options struct {
	Dir string

	// Store is the absolute root of the clone's state, resolved by the caller
	// from the repository's git common directory. The review gate reads the
	// verdict for this tree from there; an empty value means no store was found
	// and no verdict can exist, which is the same answer as "not reviewed yet".
	Store string

	// Feature is the number of the task document this run judges. Zero means the
	// highest, which is the feature the human is working on now.
	Feature int

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
	// Model is the model of the session that runs --run, as the host reports it.
	// It travels into the proof so a verdict by the same model can be refused as
	// a self-review. Empty means there is no agent to name.
	Model string
	// RequireReview turns the optional design gate on for THIS tree: a stored
	// verdict for this exact tree, and not a changes_requested one. It is the
	// flag, and it is off by default because a beginner must not be blocked by a
	// verb they have never heard of.
	RequireReview bool
	// ReviewMode is the project's switch. It asks once per feature: the tree that
	// closes the feature owes the one verdict that judged its diff, and the trees
	// inside an open feature owe only their proof.
	ReviewMode bool
	// NoReview turns the gate off for one run, whatever the mode or the flag say.
	NoReview bool
	// FenceFrom is a git ref to read the fence from instead of the worktree. CI
	// uses it to obey the policy of the base branch, which the candidate cannot
	// loosen. Empty means the worktree's own fence.
	FenceFrom string
}

// Files is the disk port: enough to read the project and write one proof.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
	Write(path string, data []byte) error
	// Remove deletes one file. A missing file may be an error; the caller that
	// only needs the state ignores it.
	Remove(path string) error
}

// Git is the repository port: which tree is here, what changed since a base,
// which dependencies appeared, and whether the repository already ignores a path.
type Git interface {
	TreeHash(root string) (string, error)
	// Candidate is the one read of the worktree: its content tree, the paths
	// that differ from base, and the diff, from one throwaway index. The fence
	// reads the paths, the lens class reads the diff, and the proof is named
	// after the tree; all three describe the same bytes.
	Candidate(root, base string) (tree string, paths []string, diff string, err error)
	// Between names the paths that differ between two trees. The proof gate knows
	// the two hashes differ; this is what tells a human which file moved, because a
	// stray backup, an editor and a real edit look identical in one line.
	Between(root, from, to string) ([]string, error)
	// FileAt reads a file from a git ref, which is how the fence is read from a
	// base branch the candidate cannot edit.
	FileAt(root, rev, path string) ([]byte, bool, error)
	AddedDeps(root, base string) ([]string, error)
	// BlobHash is the receipt of one worktree file: the blob id git already
	// owns, measured now. The coverage answer compares a verdict's rows against
	// exactly this, so a measurement is load-bearing; it answers "" for a file
	// that is not in the worktree, and a real git failure is an err.
	BlobHash(root, path string) (string, error)
	// Resolve names the commit a ref points at, in full. The review gate compares
	// bases by commit, not by spelling: `6aeb877` and its forty characters are the
	// same diff, and a receipt that names either one answers it.
	Resolve(root, ref string) (string, error)
	// Ignored reports whether the repository's own rules already keep path out
	// of the tree. It is what makes the lazy ignore file unnecessary in a project
	// that ran `odachi init`.
	Ignored(root, path string) (bool, error)
}

// Shell is the command port. A non-zero exit is the command's result, not an
// error: err means the command could not be run at all. The two streams are
// reported separately: the proof composes both, and the answer of a reviewer is
// stdout only.
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error)
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
	// Note is what the run has to say and still be ready: the review a close did
	// not make, or the debt it leaves. It is printed even under --quiet, because
	// the hook is the only moment where a reminder is read.
	Note string
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

	work, problem := r.worktree()
	if problem != nil {
		return *problem
	}
	r.tree = work.tree

	state := gate.State{TreeHash: work.tree}

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

	loosened, problem := r.fenceLoosenings()
	if problem != nil {
		return *problem
	}
	state.FenceLoosened = loosened

	feature, problem := r.currentFeature()
	if problem != nil {
		return *problem
	}
	r.feature = feature
	closedNow := false
	if feature != nil {
		state.OpenQuestions = len(feature.Doc.OpenQuestions())
		// An open feature owes the proof at its close, not at every step: the tree
		// that closes it is the one a pull request and CI judge. A document with no
		// steps yet is open too - the plan lands before the work.
		if done, total := feature.Doc.Progress(); total == 0 || done < total {
			state.ProofDeferred = true
		} else {
			closedNow = true
		}
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
	// The model the evidence came from: the recorded proof's, or the one this
	// run stamps when it writes a new one.
	r.proofModel = recorded.Model

	if opts.Run != "" {
		ran, problem := r.runProof(proofPath)
		if problem != nil {
			return *problem
		}
		state.ProofTreeHash = work.tree
		state.ProofOutput = ran.output
		state.ProofLastLine = proof.Proof{Output: ran.output}.LastLine()
		state.ProofFailed = ran.failed
		state.ProofTimedOut = ran.timedOut
		state.ProofTimeoutSeconds = opts.TimeoutSeconds
		state.ProofCommand = opts.Run
		r.proofModel = opts.Model
		haveProof = haveProof || (!ran.failed && !ran.timedOut)
	}

	// One read of what moved between the proof and this tree. The note a block
	// prints and the paperwork question below are the same evidence, and ask once:
	// an editor, a stray backup and a real edit look identical in one line, and two
	// questions about the same bytes is one question too many.
	var moved []string
	if state.ProofTreeHash != "" && state.ProofTreeHash != state.TreeHash {
		if paths, err := r.git.Between(r.opts.Dir, state.ProofTreeHash, state.TreeHash); err == nil {
			moved = paths
			r.staleNote = pathNote(paths)
		}
	}

	// The proof's twin of theCloseOwes: after the close, a tree that only moved
	// documents is paperwork -a release note cannot fail a suite- and the proof that
	// judged the code still speaks for it, because the pull request runs the suite
	// from the base ref. The question is the review's own, asked of the same
	// classifier, so the two gates cannot disagree about what a candidate is; the
	// close itself is not paperwork, and neither is a tree whose paths git cannot
	// name - nor one that moved shipped text, which the suite reads. The lane asks
	// ShippedText and not its own copy of the list: the installed skill under
	// .agents/ is plumbing for the lenses (0043) and still owes this suite.
	if closedNow && len(moved) > 0 && reviewdomain.Classify(moved) == reviewdomain.Passive && !reviewdomain.ShippedText(moved) {
		if already, problem := r.closedBefore(feature); problem == nil && already {
			state.ProofPaperwork = true
			r.paperworkNote = fmt.Sprintf("the proof is for tree %s, and since then only documents moved: %s",
				state.ProofTreeHash, strings.Join(moved, ", "))
		}
	}

	verdict := gate.Evaluate(state)
	if verdict.Code == gate.Ready {
		// The review gate is last on purpose: everything before it is something
		// the agent can act on, and sending an agent to `odachi review` while the
		// proof gate is still failing would be a dead end that resolves nothing.
		if problem := r.reviewGate(feature); problem != nil {
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
	opts  Options
	files Files
	git   Git
	shell Shell
	tree  string
	// feature is the open task document: the gate counts its open questions, and
	// the report names them. The domain gate reads no files, so the words live here.
	feature  *tasks.Document
	proofRel string
	// failedRel is the failed output written by this run, empty when the run
	// passed or the output could not be kept.
	failedRel string
	// proofModel is the model the evidence came from: the recorded proof's, or
	// the one this run stamps. Empty when no session could be named.
	proofModel string
	// staleNote names the paths that differ between the proof's tree and this
	// one, when the proof is stale. Empty means the diff could not be read.
	staleNote string
	// paperworkNote says which tree the proof belongs to when it is not owed
	// again: the documents that moved since are not the code the suite ran on.
	paperworkNote string
	// reviewNote is what the review has to say when it did not stop the close:
	// the reminder the hook prints, and the line a ready report carries.
	reviewNote string
	// work is the one read of the candidate: tree, changed paths and diff. The
	// fence, the lens class and the proof all describe these same bytes.
	work *worktree
}

// worktree is one read of the candidate: the content tree that names the proof,
// the paths that differ from base, and the diff the lens class reads.
type worktree struct {
	tree  string
	base  string
	paths []string
	diff  string
}

// worktree reads the candidate once. A project with no commit yet has no HEAD,
// so the comparison falls back to the empty tree: everything uncommitted is the
// candidate. A base the human named is never silently replaced.
func (r *runner) worktree() (*worktree, *Result) {
	if r.work != nil {
		return r.work, nil
	}

	base := baseRef(r.opts.Base)
	tree, paths, diff, err := r.git.Candidate(r.opts.Dir, base)
	if err != nil && base == head {
		if retryTree, retryPaths, retryDiff, retryErr := r.git.Candidate(r.opts.Dir, EmptyTree); retryErr == nil {
			tree, paths, diff, err = retryTree, retryPaths, retryDiff, nil
			base = EmptyTree
		}
	}
	if err != nil {
		message := "could not read the git tree: " + oneLine(err.Error()) + "; run odachi check inside a git repository and check that git is installed"
		if base != head {
			message = baseProblem(base, err)
		}
		problem := blocked(message)
		return nil, &problem
	}

	r.work = &worktree{tree: tree, base: base, paths: paths, diff: diff}
	return r.work, nil
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

	work, problem := r.worktree()
	if problem != nil {
		return nil, problem
	}
	changed := work.paths

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

// currentFeature is the feature this run judges: the one the caller selected, or
// the highest-numbered task document -the feature the human is working on now-.
// No document is not an error when no feature was named: a project with no
// features yet has no open questions and can still be verified. A named number on
// that project is an error, exactly as it is for every other verb: a selector
// never answers nothing.
func (r *runner) currentFeature() (*tasks.Document, *Result) {
	if r.opts.Feature > 0 {
		doc, err := tasks.Feature(r.files, r.opts.Dir, r.opts.Feature)
		if err != nil {
			problem := blocked(oneLine(err.Error()))
			return nil, &problem
		}
		return &doc, nil
	}

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
		// The failed file is not a malformed proof: it is the evidence of a run
		// that did not pass, and saying so is more useful than "no tree header".
		if strings.HasPrefix(string(data), proof.FailedHeader) {
			problem := refused(shown, "it is the failed output of a run, not a proof: fix what failed and run the gate green")
			return proof.Proof{}, &problem
		}
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
	stdout, stderr, exitCode, timedOut, err := r.shell.Run(r.opts.Dir, r.opts.Run, r.opts.TimeoutSeconds)
	// The proof is what the command said, both streams: a failing test writes
	// its verdict to stdout and its build errors to stderr, and hiding either
	// would make the proof a partial quotation.
	output := stdout + stderr
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
		r.keepFailed(proofPath, output, exitCode, true)
		return ranProof{output: output, timedOut: true}, nil
	}
	if exitCode != 0 {
		r.keepFailed(proofPath, output, exitCode, false)
		return ranProof{output: output, failed: true}, nil
	}

	// A green run removes the failed evidence of an earlier run of the same
	// feature: proof/ must never show a failure beside the proof of the tree that
	// just passed. Best effort: the green proof is the truth either way.
	_ = r.files.Remove(failedFile(proofPath))

	recorded := proof.Proof{Tree: r.tree, Command: r.opts.Run, Fence: r.opts.FenceFrom, Model: r.opts.Model, Output: output}
	if err := r.files.Write(proofPath, []byte(proof.Render(recorded))); err != nil {
		problem := blocked(fmt.Sprintf("could not write the proof %s: %s; check the permissions and run odachi check --run %q again",
			r.proofRel, oneLine(err.Error()), r.opts.Run))
		return ranProof{}, &problem
	}
	return ranProof{output: output}, nil
}

// keepFailed writes the whole output of a run that did not pass, next to where
// the green proof would live, so a red gate can be diagnosed after its one-line
// message. It is best effort: the run already failed, and a write problem must
// not replace the failure the human has to fix. Nothing is claimed when the file
// is not written.
func (r *runner) keepFailed(proofPath, output string, exitCode int, timedOut bool) {
	recorded := proof.Proof{Tree: r.tree, Command: r.opts.Run, Fence: r.opts.FenceFrom, Output: output}
	path := failedFile(proofPath)
	if err := r.files.Write(path, []byte(proof.RenderFailed(recorded, exitCode, timedOut))); err != nil {
		return
	}
	rel, err := filepath.Rel(r.opts.Dir, path)
	if err != nil {
		return
	}
	r.failedRel = filepath.ToSlash(rel)
}

// failedFile names the evidence of a failed run next to the proof of a green
// one: proof/NNNN-<slug>.failed.txt. The suffix is the whole difference: a
// failed run is evidence, not a proof, and nothing looks for it as one.
func failedFile(path string) string {
	if strings.HasSuffix(path, ".txt") {
		return strings.TrimSuffix(path, ".txt") + ".failed.txt"
	}
	return path + ".failed.txt"
}

// reviewGate decides whether this tree owes a verdict and which one. The flag
// asks for this tree; the project mode asks once per feature, and names the base
// its verdict has to judge. A feature with no steps, or a document that records no
// base, has no unit to review once: the mode keeps asking per tree there.
func (r *runner) reviewGate(feature *tasks.Document) *Result {
	if r.opts.NoReview {
		return nil
	}
	base, closed, waived, known := featureReview(feature)

	// The human can waive this feature's review. The decision is recorded in the
	// document, it outranks the per-run flag -the lane exists so a small feature
	// does not pay a verdict in CI either- and it removes the verdict, never the
	// proof or the fence. The report names it.
	if known && closed && waived {
		return nil
	}

	if !r.opts.RequireReview {
		if !r.opts.ReviewMode {
			return nil
		}
		// Inside an open feature the proof is the gate: the review is asked once,
		// for the tree that closes it, not for every step that builds it.
		if known && !closed {
			return nil
		}
		// A feature stays closed after its close. The verdict is owed by the change
		// that marked the last step, not by every document that follows it.
		if known && closed {
			owed, problem := r.theCloseOwes(feature)
			if problem != nil {
				return problem
			}
			if !owed {
				return nil
			}
		}
	}
	if known && closed {
		return r.softened(r.requireReview(base))
	}
	return r.softened(r.requireReview(""))
}

// softened turns a review refusal into a note unless the human asked to be refused.
// The review does not authorise delivery (ADR 0043): with the project's switch on,
// a close names what the review says and the commit proceeds. The one refusal that
// survives is a finding the review made -a causal BLOCKER is the reason not to ship,
// not a requirement that a review exists- and `--require-review` is the human asking
// to be stopped.
func (r *runner) softened(result *Result) *Result {
	if result == nil || r.opts.RequireReview {
		return result
	}
	if result.Reason == ReasonReviewChanges {
		return result
	}
	note := result.Message
	if next := strings.TrimSpace(result.Next); next != "" {
		note += "; " + next
	}
	r.reviewNote = note
	return nil
}

// theCloseOwes reports whether this tree is the one that closes the feature. Two
// things ask for the verdict: code, whenever it appears; and the documents that
// marked the last step done. A release note written after the close is paperwork:
// the verdict it would ask for already judged the same code.
func (r *runner) theCloseOwes(feature *tasks.Document) (bool, *Result) {
	required, _, _, problem := r.requiredLenses()
	if problem != nil {
		return false, problem
	}
	if len(required) > 0 {
		// Code after the close is a change nobody reviewed: the gate stays on.
		return true, nil
	}

	// The feature's own diff decides, and it is the diff a verdict would judge: from
	// the document's base to the tree that stands. A feature that only wrote documents
	// asks for no lens, so there is nothing to judge and the close is ready - the twin
	// 0096 gave the proof, asked of the same classifier the review uses. Measuring the
	// closing tree alone would answer "no code" about a feature that wrote plenty: that
	// tree only marks the last step.
	passive, problem := r.featurePassive(feature)
	if problem != nil {
		return false, problem
	}
	if passive {
		return false, nil
	}

	// The anchor is the verdict that judged the close: the store names the bytes a
	// review judged, so what arrived after them is the question. The candidate cannot
	// answer it - it is measured against HEAD, and a committed change has already
	// disappeared from it, which is how a local check answered "ready" for a tree CI
	// would block.
	if judged, ok := r.lastVerdictTree(feature); ok {
		work, problem := r.worktree()
		if problem != nil {
			return false, problem
		}
		paths, err := r.git.Between(r.opts.Dir, judged, work.tree)
		if err != nil {
			problem := blocked(fmt.Sprintf("could not read what changed since the verdict of %s: %s; check git and run odachi check again",
				judged, oneLine(err.Error())))
			return false, &problem
		}
		return reviewdomain.Classify(paths) != reviewdomain.Passive, nil
	}

	closed, problem := r.closedBefore(feature)
	if problem != nil {
		return false, problem
	}
	return !closed, nil
}

// lastVerdictTree names the bytes this feature's review judged: the tree of the newest
// verdict whose base is the document's base. A feature's verdicts share that base, so
// the newest one is the furthest the review has looked, and the store is the tool's own
// evidence - asking it beats guessing the close commit out of the document's history.
func (r *runner) lastVerdictTree(feature *tasks.Document) (string, bool) {
	base, _, _, known := featureReview(feature)
	if !known || base == "" {
		return "", false
	}

	names, err := r.files.List(r.storeFile(reviewdomain.Dir))
	if err != nil {
		return "", false
	}

	best, bestAt := "", ""
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := r.files.Read(r.storeFile(reviewdomain.Dir + "/" + name))
		if err != nil {
			continue
		}
		// A scope record, a dismissal or a refused answer is not a verdict: the strict
		// parser is what tells them apart, and one it cannot read is not evidence.
		verdict, err := reviewdomain.Parse(data)
		if err != nil || strings.TrimSpace(verdict.Tree) == "" {
			continue
		}
		if !sameCommit(r.git, r.opts.Dir, verdict.Base, base) {
			continue
		}
		if verdict.At > bestAt {
			best, bestAt = verdict.Tree, verdict.At
		}
	}
	return best, best != ""
}

// featurePassive asks the review's classifier about the diff this feature merges:
// from the document's base to the tree that stands. It is the same function the
// review asks, so the gate and the tool cannot answer differently about whether a
// candidate deserves a verdict.
func (r *runner) featurePassive(feature *tasks.Document) (bool, *Result) {
	base, _, _, known := featureReview(feature)
	if !known || base == "" {
		return false, nil
	}
	work, problem := r.worktree()
	if problem != nil {
		return false, problem
	}
	paths, err := r.git.Between(r.opts.Dir, base, work.tree)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read what changed since %s: %s; check git and run odachi check again",
			base, oneLine(err.Error())))
		return false, &problem
	}
	return reviewdomain.Classify(paths) == reviewdomain.Passive, nil
}

// closedBefore says whether the document at HEAD already carried the close: the
// tree that closed a feature is the tree that owes its evidence, and the trees
// after it are paperwork. A document that was not at HEAD yet was closed here, and
// a document git cannot read is not paperwork either: the lane never opens on a
// question the tool could not answer.
func (r *runner) closedBefore(feature *tasks.Document) (bool, *Result) {
	before, ok, err := r.git.FileAt(r.opts.Dir, "HEAD", feature.Rel)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s at HEAD: %s; check git and run odachi check again", feature.Rel, oneLine(err.Error())))
		return false, &problem
	}
	if !ok {
		return false, nil
	}
	done, total := taskdoc.Parse(string(before)).Progress()
	return total > 0 && done == total, nil
}

// featureReview reads the feature's review state from its document: the base its
// diff starts from, whether every step is done, and whether the document is the
// one that can be reviewed once at all.
func featureReview(feature *tasks.Document) (base string, closed, waived, known bool) {
	if feature == nil {
		return "", false, false, false
	}
	base = strings.TrimSpace(feature.Doc.Base)
	done, total := feature.Doc.Progress()
	if base == "" {
		return "", false, false, false
	}
	// A document with no steps yet is open too: the plan lands before the work,
	// the same rule the proof follows above. Without it the plan commit -which
	// writes the fence, ordinary code on purpose because it is a human decision-
	// is a stranger, and check asks it for the verdict of a tree nobody built.
	if total == 0 {
		return base, false, feature.Doc.ReviewWaived, true
	}
	return base, done == total, feature.Doc.ReviewWaived, true
}

// ReviewOwed is the project-mode review question, exported so a report tells the
// truth the gate tells instead of its own version of it: whether this tree owes a
// verdict, and the base that review must name when it does. It writes nothing and
// starts nothing.
func ReviewOwed(opts Options, files Files, git Git) (bool, string, error) {
	// The question is about the tree that stands. A caller that names no base gets
	// HEAD, the default `odachi check` uses: an empty base means "compare against
	// nothing" for the fence, and here it would turn every file into code.
	if strings.TrimSpace(opts.Base) == "" {
		opts.Base = "HEAD"
	}

	r := &runner{opts: opts, files: files, git: git}

	feature, problem := r.currentFeature()
	if problem != nil {
		return false, "", errors.New(problem.Message)
	}

	// The same questions reviewGate asks, in the same order. Sharing the
	// answers is the point: two callers deriving them would drift.
	base, closed, waived, known := featureReview(feature)

	// A waived feature owes no verdict: the human decided it lands without one,
	// and the report says so instead of inventing a review.
	if known && closed && waived {
		return false, "", nil
	}

	if !opts.RequireReview {
		if !opts.ReviewMode {
			return false, "", nil
		}
		if known && !closed {
			return false, "", nil
		}
		if known && closed {
			owed, problem := r.theCloseOwes(feature)
			if problem != nil {
				return false, "", errors.New(problem.Message)
			}
			if !owed {
				return false, "", nil
			}
			return true, base, nil
		}
	}
	if known && closed {
		return true, base, nil
	}

	required, _, _, problem := r.requiredLenses()
	if problem != nil {
		return false, "", errors.New(problem.Message)
	}
	return len(required) > 0, "", nil
}

// requireReview is the optional design gate: a stored verdict for THIS tree, one
// that covers every lens the candidate requires, and no causal BLOCKER or
// CRITICAL finding left standing after the refutation is taken into account.
//
// The lenses come first because they answer whether this candidate asks for a
// review at all. A documentation-only candidate has nothing for a specialist lens
// to look at, so requiring a verdict would be ceremony, not evidence. A closed
// feature is the exception: its closing tree only carries the document, its own
// lenses are none, and it still owes the verdict that judged the feature's diff -
// base is that diff's base, and the verdict has to name it.
//
// It writes nothing. The store's ignore file is created by the verbs that produce
// a verdict, so a tree with no verdict has no store to prepare either.
func (r *runner) requireReview(base string) *Result {
	required, changed, diff, problem := r.requiredLenses()
	if problem != nil {
		return problem
	}
	if len(required) == 0 && base == "" {
		return nil
	}

	next := nextReview
	if base != "" {
		next = "odachi review --base " + base
	}

	rel, err := reviewdomain.Path(r.tree)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; check the git tree and run odachi check again")
		return &problem
	}

	data, err := r.files.Read(filepath.Join(r.opts.Store, filepath.FromSlash(rel)))
	if err != nil {
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewMissing,
			Message: "no review for this tree",
			Next:    next,
			Lines:   []string{"no review for this tree", "next: " + next},
		}
		if base != "" {
			problem.Message = "the feature is closed and its tree has no verdict"
			problem.Lines = []string{
				problem.Message,
				"next: " + next,
				"why: the closing commit only carries the document; the verdict judges the diff the feature merges",
			}
		}
		// A tree without a verdict may be a correction of one that has it: the
		// store keeps the predecessor, this run knows the feature's base, and the
		// command that pays for the delta is one line away instead of a hash the
		// human has to go looking for.
		if predecessor, ok := r.lastVerdictTree(r.feature); ok && predecessor != r.tree {
			problem.Lines = append(problem.Lines, "or correct it: odachi review --since "+predecessor)
		}
		return &problem
	}

	verdict, err := reviewdomain.Parse(data)
	if err != nil {
		problem := Result{
			Code:   int(gate.NeedsDecision),
			Reason: ReasonReviewMissing,
			Message: fmt.Sprintf("this tree's review cannot be read (%s): %s; write it again with odachi review --attach <file>",
				r.storeFile(rel), oneLine(err.Error())),
			Next: nextReview,
		}
		problem.Lines = []string{problem.Message, "next: odachi review --attach <file>"}
		return &problem
	}

	// A closed feature owes the verdict of its own diff. A verdict that judged
	// something else - the closing commit, another base - does not answer the
	// question this tree asks, and accepting it would make the base decoration.
	if base != "" && !sameCommit(r.git, r.opts.Dir, verdict.Base, base) {
		message := fmt.Sprintf("the review of this tree judged another diff (base %q): the feature starts at %s", strings.TrimSpace(verdict.Base), base)
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewMissing,
			Message: message,
			Next:    next,
			Lines: []string{
				message,
				"next: " + next,
				"why: a verdict describes the bytes it judged; the closing commit is not the feature",
			},
		}
		return &problem
	}

	// Identity before coverage: a lens judged by the model that produced the proof
	// is the tree approving itself. Both names are declarations, so this refuses
	// what can be seen, not what cannot be proven.
	if lenses := selfJudgedLenses(r.proofModel, verdict, required); len(lenses) > 0 {
		message := fmt.Sprintf("the review of %s was written by the model that produced the proof (%s): a model does not judge its own tree",
			strings.Join(lenses, ", "), strings.TrimSpace(r.proofModel))
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewSelf,
			Message: message,
			Next:    "odachi review --model <provider/model>",
			Lines: []string{
				message,
				"next: odachi review --model <provider/model>",
				"why: a verdict from another model is evidence; the builder's own is not",
			},
		}
		return &problem
	}

	// Silence is not an identity. The gate compares two declarations, so a verdict
	// that names no model would pass it by saying less - the cheapest way around a
	// rule - and the tree would be approved by nobody in particular.
	if lenses := unattributedLenses(r.proofModel, verdict, required); len(lenses) > 0 {
		message := fmt.Sprintf("the review of %s (%s) does not say who judged it, and the proof names %s: a verdict that names no model cannot be compared with the tree's builder",
			strings.Join(lenses, ", "), r.storeFile(rel), strings.TrimSpace(r.proofModel))
		problem := Result{
			Code:    int(gate.NeedsDecision),
			Reason:  ReasonReviewUnattributed,
			Message: message,
			Next:    "odachi review --model <provider/model>",
			Lines: []string{
				message,
				"next: odachi review --model <provider/model>",
				"or record the decision that there is no second model: odachi start --waive-review",
				"why: an identity nobody declared is an identity nobody can check",
			},
		}
		return &problem
	}

	if missing := verdict.Missing(required); len(missing) > 0 {
		problem := lensMissing(missing, r.storeFile(rel))
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

	// Coverage is per row: every changed file whose bytes git can measure owes
	// one current answer per lens its path class requires. A receipt that
	// describes other bytes is a file that changed since the verdict — its
	// answer is dead, and the gate names the gap instead of reusing it.
	blobs, problem := r.currentBlobs(changed)
	if problem != nil {
		return problem
	}
	need := reviewdomain.Need(changed, diff)
	for path, blob := range blobs {
		if blob == "" {
			delete(need, path)
		}
	}
	if gaps := verdict.Uncovered(need, blobs); len(gaps) > 0 {
		problem := coverageMissing(verdict, gaps, r.storeFile(rel))
		return &problem
	}

	return nil
}

// storeFile names a path inside the store as the file a human can delete: the
// absolute path, with forward slashes so a message reads the same on every
// platform. `odachi check` may run from anywhere, so a project-relative spelling
// would not be runnable.
func (r *runner) storeFile(rel string) string {
	return filepath.ToSlash(filepath.Join(r.opts.Store, filepath.FromSlash(rel)))
}

// sameCommit reports whether two spellings name the same commit. A base is a
// commit, not a text: `6aeb877` and its forty characters describe the same diff,
// and comparing the strings made a paid review unusable the day the flag was
// typed short. What cannot be resolved is not accepted: a receipt that names an
// unreadable ref answers nothing.
func sameCommit(git Git, root, judged, start string) bool {
	judged = strings.TrimSpace(judged)
	start = strings.TrimSpace(start)
	if judged == start {
		return true
	}
	if judged == "" || start == "" {
		return false
	}
	left, err := git.Resolve(root, judged)
	if err != nil {
		return false
	}
	right, err := git.Resolve(root, start)
	if err != nil {
		return false
	}
	return left == right
}

// requiredLenses is the lens selection the candidate asks for. It is the same
// question the review verb answers before printing a prompt, asked from the same
// changed paths and the same diff, so the two can never disagree about what this
// tree needs. The changed paths and the diff are returned with it: coverage is
// per file, and the same evidence that picks the lenses is the one whose rows
// must answer.
func (r *runner) requiredLenses() ([]string, []string, string, *Result) {
	work, problem := r.worktree()
	if problem != nil {
		return nil, nil, "", problem
	}

	return reviewdomain.Lenses(reviewdomain.Assess(work.paths, work.diff)), work.paths, work.diff, nil
}

// currentBlobs measures the receipt each changed path earns now. A path the
// port answers "" for is not in the worktree — the candidate deleted it, and
// absence owes no coverage. A real git failure is reported, never swallowed:
// the measurement is what makes a receipt a receipt.
func (r *runner) currentBlobs(paths []string) (map[string]string, *Result) {
	blobs := make(map[string]string, len(paths))
	for _, path := range paths {
		blob, err := r.git.BlobHash(r.opts.Dir, path)
		if err != nil {
			problem := blocked(fmt.Sprintf("could not measure the blob of %s: %s; fix git and run odachi check again", path, oneLine(err.Error())))
			return nil, &problem
		}
		blobs[path] = blob
	}
	return blobs, nil
}

// coverageMissing is the block for a verdict whose rows do not answer coverage:
// stale receipts, missing rows, or — for a verdict written before content
// receipts — nothing computable at all. The store cannot be rewritten, so the
// message names the file the human must delete and the command that replaces it;
// the gate never deletes anything itself.
func coverageMissing(verdict reviewdomain.Verdict, gaps []reviewdomain.Gap, rel string) Result {
	next := nextReview

	// A verdict written before content receipts owes the diagnosis of its
	// format, not of its rows: an approval the tool cannot prove is an approval
	// it must not use.
	receipted := len(verdict.Passes) > 0
	for _, finding := range verdict.Findings {
		if strings.TrimSpace(finding.Blob) != "" {
			receipted = true
			break
		}
	}

	var message string
	if !receipted {
		message = fmt.Sprintf("this tree's review was written before content receipts (%s): it cannot serve coverage", rel)
	} else {
		gap := gaps[0]
		message = fmt.Sprintf("the review of this tree has no current answer for %s under the %s lens (%s)", gap.Path, gap.Lens, rel)
	}
	return Result{
		Code:    int(gate.NeedsDecision),
		Reason:  ReasonReviewCoverage,
		Message: message,
		Next:    next,
		Lines: []string{
			message,
			"a verdict is not rewritten: delete " + rel + " and run " + next,
			"next: " + next,
		},
	}
}

// refutation reads the refutation filed beside this tree's verdict. A tree with
// no refutation is the normal case and is not an error; a file that cannot be
// read is, because silence would let a broken artifact look like an approval.
func (r *runner) refutation(verdict reviewdomain.Verdict) (reviewdomain.Refutation, *Result) {
	rel, err := reviewdomain.RefutationPath(r.tree)
	if err != nil {
		return reviewdomain.Refutation{}, nil
	}

	data, err := r.files.Read(filepath.Join(r.opts.Store, filepath.FromSlash(rel)))
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

// selfJudgedLenses is the required lenses this verdict judged with the model that
// produced the proof, in the gate's order. A lens answers with its own model when
// the project assigned one, and with the verdict's single model otherwise.
//
// Two unknowns are not a match: an old proof, a verdict written by a human or a
// lens the command did not name has no model, and calling that a self-review
// would block every project that ran the gate before the field existed. The
// comparison ignores case because the same identity is spelled both ways.
func selfJudgedLenses(proofModel string, verdict reviewdomain.Verdict, required []string) []string {
	proofModel = strings.TrimSpace(proofModel)
	if proofModel == "" {
		return nil
	}

	var lenses []string
	for _, lens := range required {
		name := strings.ToLower(strings.TrimSpace(lens))
		// Only a lens this verdict answered can be self-judged: an uncovered one
		// is the coverage gate's business, and naming it here would hide the gap
		// behind an identity refusal.
		if !answered(verdict.Lenses, name) {
			continue
		}
		model := strings.TrimSpace(verdict.Models[name])
		if model == "" {
			model = strings.TrimSpace(verdict.Model)
		}
		if model != "" && strings.EqualFold(model, proofModel) {
			lenses = append(lenses, name)
		}
	}
	return lenses
}

// unattributedLenses is the required lenses this verdict answered while naming no
// model, on a tree whose proof names one. It is the mirror of selfJudgedLenses: one
// compares what both sides declare, and this one refuses the case where a side
// declares nothing at all.
func unattributedLenses(proofModel string, verdict reviewdomain.Verdict, required []string) []string {
	proofModel = strings.TrimSpace(proofModel)
	if proofModel == "" {
		return nil
	}

	var lenses []string
	for _, lens := range required {
		name := strings.ToLower(strings.TrimSpace(lens))
		if !answered(verdict.Lenses, name) {
			continue
		}
		model := strings.TrimSpace(verdict.Models[name])
		if model == "" {
			model = strings.TrimSpace(verdict.Model)
		}
		if model == "" {
			lenses = append(lenses, name)
		}
	}
	return lenses
}

// answered reports whether the verdict covers this lens.
func answered(lenses []string, name string) bool {
	for _, lens := range lenses {
		if strings.EqualFold(strings.TrimSpace(lens), name) {
			return true
		}
	}
	return false
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

// pathNote names paths in one line, capped so a block stays readable: the human
// needs to recognize what moved, not to read a manifest.
func pathNote(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	const limit = 8
	if len(paths) > limit {
		return strings.Join(paths[:limit], ", ") + fmt.Sprintf(" and %d more", len(paths)-limit)
	}
	return strings.Join(paths, ", ")
}

// report turns the gate's verdict into the verb's result.
// openQuestions is what the gate only counts: the words the human is being asked.
// It answers nothing and writes nothing; it only makes the block readable.
func (r *runner) openQuestions(verdict gate.Result) []taskdoc.Question {
	if verdict.Reason != gate.ReasonOpenQuestion || r.feature == nil {
		return nil
	}
	return r.feature.Doc.OpenQuestions()
}

// report is the block the human reads: the message, what moved, and the command
// that unblocks it.
func (r *runner) report(verdict gate.Result, haveProof bool) Result {
	result := Result{
		Code:    int(verdict.Code),
		Reason:  verdict.Reason,
		Message: verdict.Message,
		Next:    verdict.Next,
	}

	if verdict.Code != gate.Ready {
		result.Lines = []string{verdict.Message}
		// The gate counts what blocks; the document knows the words. A block that
		// only counts sends the human to read a file to discover what they are
		// being asked, and to guess the command that answers it.
		questions := r.openQuestions(verdict)
		for _, question := range questions {
			result.Lines = append(result.Lines, fmt.Sprintf("%s · %q", question.ID, question.Text))
		}
		if r.staleNote != "" {
			result.Lines = append(result.Lines, "changed since the proof: "+r.staleNote)
		}
		if r.failedRel != "" {
			result.Lines = append(result.Lines, "output: "+r.failedRel)
		}
		next := verdict.Next
		if len(questions) > 0 {
			next = fmt.Sprintf("odachi ask --answer \"<answer>\" --for %s", questions[0].ID)
			result.Next = next
		}
		result.Lines = append(result.Lines, "next: "+next)
		return result
	}

	result.Lines = []string{verdict.Message}
	if haveProof && r.proofRel != "" {
		result.Proof = r.proofRel
		result.Lines = append(result.Lines, "proof: "+r.proofRel)
	}
	// A ready tree the suite never ran on says so: the proof it keeps proved the
	// code, and a silent ready would be the kind of quiet this gate exists to avoid.
	if r.paperworkNote != "" {
		result.Lines = append(result.Lines, r.paperworkNote)
	}
	// A waived review is a decision, not a silence: the ready tree says so, or
	// the next reader believes a verdict was made and lost.
	if _, closed, waived, known := featureReview(r.feature); known && closed && waived {
		result.Lines = append(result.Lines, "review: waived by the human")
	}
	// The close names the debt it leaves. The rows are the list a person reads
	// before choosing the next feature; a close that does not name them leaves the
	// reading to memory, and memory is what the file exists to replace.
	if note := r.debtNote(); note != "" {
		result.Lines = append(result.Lines, note)
	}
	// What the review has to say and did not stop: the missing verdict of a close,
	// or why the one on file does not count. A ready tree says it out loud -the hook
	// is where a reminder is read- and keeps its exit code.
	if r.reviewNote != "" {
		result.Lines = append(result.Lines, r.reviewNote)
		result.Note = r.reviewNote
	}
	return result
}

// debtNote is the review debt this tree leaves, read from the file a person reads:
// the open rows of docs/review-debt.md, never a second copy of the list. It is a
// note and not a gate -a row never blocks, and a file that cannot be read says
// nothing instead of failing a ready tree-.
func (r *runner) debtNote() string {
	data, err := r.files.Read(filepath.Join(r.opts.Dir, "docs", "review-debt.md"))
	if err != nil {
		return ""
	}
	rows := openDebtRows(string(data))
	if len(rows) == 0 {
		return ""
	}
	return fmt.Sprintf("review debt: %d open row(s) in docs/review-debt.md: %s", len(rows), strings.Join(rows, "; "))
}

// openDebtRows reads the labels under "## Open" until the next section. The file
// writes one row per "- **<label>**", and everything else in the section is the
// prose that explains them.
func openDebtRows(text string) []string {
	var rows []string
	open := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			open = strings.EqualFold(trimmed, "## Open")
			continue
		}
		if !open || !strings.HasPrefix(trimmed, "- **") {
			continue
		}
		label := strings.TrimPrefix(trimmed, "- **")
		if end := strings.Index(label, "**"); end >= 0 {
			label = label[:end]
		}
		if label = strings.TrimSpace(label); label != "" {
			rows = append(rows, label)
		}
	}
	return rows
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
