// Package review implements `odachi review`: one design verdict per tree,
// content-addressed and immutable.
//
// The verb answers the question the gate does not: not "is it proven" but "is it
// well designed". It does that without inheriting a lifecycle: no corrections, no
// budget, no state machine. A verdict is a file named after the tree it judged,
// so an approval cannot float to a tree nobody looked at — the sentence is the
// file name.
//
// The package is pure orchestration over three ports: Files (the store), Git (the
// candidate) and Shell (the reviewer command). Every rule lives in
// internal/domain/review; every English line lives here.
package review

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means the prompt was printed, the verdict was stored, or a verdict
	// was shown.
	CodeOK = 0
	// CodeBlocked means there is nothing to review, the verdict could not be
	// read, the reviewer command failed, or this tree already has a verdict.
	CodeBlocked = 1
)

// DefaultBase is what the candidate is compared against when --base is not given.
const DefaultBase = "HEAD"

// EmptyTree is git's hash of the empty tree: a project with no commit yet is
// compared against nothing, and every file counts as new. `check` uses the same
// fallback, so the two verbs never disagree about what changed.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// ExecTimeoutSeconds bounds --exec. A reviewer that never answers would hang the
// verb, and an LLM review is slower than a test suite; ten minutes is enough for
// a real one and short enough to give up.
const ExecTimeoutSeconds = 600

// briefFile is the vision the review is judged against.
const briefFile = "brief.md"

// noBrief is what the prompt says when there is no brief yet. A review without
// one is still possible — the design defects are still there — and pretending
// otherwise would turn a missing brief into a second gate this verb does not own.
const noBrief = "(there is no brief.md yet: run odachi start)"

// modelToken is what a reviewer command writes where the session's model goes.
// It is a placeholder, not a flag: odachi has no business knowing how each CLI
// spells its model, and the human keeps writing the command they know.
const modelToken = "{model}"

// Options is one `odachi review` invocation.
type Options struct {
	Dir string

	// Base is what the candidate is compared against. Empty means the empty
	// tree: everything counts as new.
	Base string
	// Exec is the reviewer command: it is run with the review prompt as its
	// input, and its answer is the verdict. {model} in it is replaced by the
	// model the session is using, so the reviewer judges with the same one the
	// human is talking to.
	Exec string
	// Model is the raw --model value: the identity that replaces {model}.
	// Empty means the host's answer.
	Model string
	// LensModels is the project's model per lens, from `.odachi/review.toml`.
	// A lens with no entry is left to the reviewer command, which is what the
	// file meant before the table existed.
	LensModels map[string]string
	// Attach is a verdict someone already wrote.
	Attach string
	// Show reads this tree's verdict instead of producing one.
	Show bool
	// Lens is the raw --lens value: one name, a comma list, or `all`. Empty means
	// the lenses the evidence of the diff asks for, and nothing else.
	Lens string
	// Refute is a verdict file, and this run prints the refuter's prompt for it.
	Refute string
	// Now is the instant a stored verdict records. Zero means the wall clock;
	// tests fix it so a stored verdict is byte-stable.
	Now time.Time
}

// Files is the disk port: enough to read the brief, read an attached verdict,
// check whether this tree has one, write it, and list veredicto names the store
// already holds.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	// List answers the file names one directory holds. An absent directory is
	// an empty answer, not an error: a project with no store carries nothing.
	List(dir string) ([]string, error)
}

// Git is the repository port: which tree is here, what changed since the base,
// the diff itself, and whether the repository already ignores a path. The diff is
// what the prompt is built from, so a repository port without it could not
// describe a candidate.
type Git interface {
	TreeHash(root string) (string, error)
	// Candidate is the one read of the worktree the verb shares: the content
	// tree, the paths that differ from base, and the diff, from one throwaway
	// index. The three answers describe the same bytes.
	Candidate(root, base string) (tree string, paths []string, diff string, err error)
	// BlobHash is the content hash of one worktree file - the blob id git
	// already owns. It is how a finding earns its content receipt without
	// anybody declaring it: a hash is measured, not word of mouth.
	// BlobHash(root, path) answers "" for a file that is not in the worktree
	// (the candidate deleted it, or the reviewer invented the path); a real git
	// failure is an err, not an empty receipt.
	BlobHash(root, path string) (string, error)
	// Ignored reports whether the repository's own rules already keep path out
	// of the tree. It is what makes the lazy ignore file unnecessary in a project
	// that ran `odachi init`.
	Ignored(root, path string) (bool, error)
}

// Shell is the command port for --exec. A non-zero exit is the command's result,
// not an adapter error: err means it could not be run at all.
//
// RunInput is the form this verb uses: the reviewer is a reviewer because it
// receives the prompt, and how a prompt reaches a command is the shell adapter's
// business (stdin, and a file path for a command that takes one), not a decision
// this layer should know how to make.
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error)
	RunInput(dir, command, input string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error)
}

// Lenses is the port for the four specialist prompts. It returns the file
// verbatim; the domain strips the frontmatter when it builds the prompt, so the
// content stays complete on disk and readable by a human.
type Lenses interface {
	Lens(name string) (string, error)
}

// Host is the agent odachi runs under. The one thing only it can answer is
// which model the session is using: that identity lives in the host, not in the
// project, and inventing one would put a name in the receipt that never ran.
// "" means the host cannot say, and the verb decides what that costs.
type Host interface {
	Model(dir string) string
}

// Result is what the CLI prints and exits with. It has the same shape as the
// other verbs, plus the two things only this one produces: the prompt, and the
// verdict it read or stored.
type Result struct {
	Code    int
	Message string
	Lines   []string
	// Prompt is the review prompt, printed verbatim. It is the deliverable of
	// the bare form.
	Prompt string
	// Output is what the --exec command printed. A failed review is shown, never
	// summarised away: the agent's own words are the diagnosis.
	Output string
	// Verdict is the verdict that was stored or shown.
	Verdict *domain.Verdict
	// Lenses are the lenses this run asked for, in the fixed order. The prompt
	// carries their rules; the stored verdict records them.
	Lenses []string
	// Refutation is the refutation this run stored or found beside the verdict.
	Refutation *domain.Refutation
	// Files are the files this run wrote, relative to the project.
	Files []string
}

// Run performs one `odachi review`.
func Run(opts Options, files Files, git Git, shell Shell, lenses Lenses, host Host) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	r := &runner{opts: opts, files: files, git: git, shell: shell, lenses: lenses, host: host}

	// --model only exists for {model} in --exec. In any other form it would be
	// accepted and silently ignored, which is the same lie as a missing answer.
	if strings.TrimSpace(opts.Model) != "" && opts.Exec == "" {
		return blocked("--model only means something with --exec: it replaces " + modelToken + " in the reviewer command")
	}

	switch {
	case opts.Show:
		return r.show()
	case opts.Refute != "":
		return r.refute()
	case opts.Attach != "":
		return r.attach()
	case opts.Exec != "":
		return r.exec()
	default:
		return r.prompt()
	}
}

// runner holds one invocation.
type runner struct {
	opts   Options
	files  Files
	git    Git
	shell  Shell
	lenses Lenses
	host   Host
	// work is the candidate read once: tree, changed paths and diff. The
	// reviewer command may move the worktree, so afterReviewer drops it and the
	// store reads again.
	work *worktree
	// storeReady is whether this run already asked the repository whether the
	// store is ignored. The probe is a git process and the answer cannot change
	// between reads, so it is asked once.
	storeReady bool
	// storeIgnoreWritten is whether this run had to write the store's own
	// ignore file, so the result reports only files it really wrote.
	storeIgnoreWritten bool
	// reviewedBy is the model the reviewer command ran with, empty when the
	// command asked for none. The verdict carries it so a later gate can tell
	// this review from one the builder wrote about itself.
	reviewedBy string
}

// worktree is one read of the candidate: its content tree, the paths that
// differ from the base it was compared against, and the diff. The three answers
// describe the same bytes and come from the same throwaway index.
type worktree struct {
	tree  string
	base  string
	paths []string
	diff  string
}

// candidate is what is being reviewed: the brief it is judged against, the base
// it is compared to, the paths that changed and the diff itself.
type candidate struct {
	brief string
	base  string
	tree  string
	paths []string
	diff  string
}

// prompt prints the review prompt for the current candidate. It prepares the
// store first: the verdict lives inside the candidate, so the directory has to be
// ignored before the tree is hashed or storing the verdict would change the tree
// that names it.
func (r *runner) prompt() Result {
	cand, problem := r.candidateForReview()
	if problem != nil {
		return *problem
	}

	text, tree, selected, problem := r.promptText(cand)
	if problem != nil {
		return *problem
	}

	files := []string{}
	if r.storeIgnoreWritten {
		// Only what this run actually wrote: a project whose .gitignore already
		// ignores the store gets no second ignore file, and claiming one would be
		// machine output that lies about the worktree.
		files = append(files, ignoreRel())
	}

	return Result{
		Code:    CodeOK,
		Message: promptMessage(tree, selected),
		Prompt:  text,
		Lenses:  selected,
		Files:   files,
	}
}

// promptText is the prompt for one candidate, plus the tree it names and the
// lenses it asks for.
//
// It is one function because --exec and the printed form must hand the reviewer
// the SAME text: two builders would drift, and a reviewer answering a different
// question than the one the gate asked is worse than no review. The store is
// prepared first, exactly as for the printed form: the verdict lives inside the
// candidate, so the directory has to be ignored before the tree is hashed.
func (r *runner) promptText(cand candidate) (text, tree string, selected []string, problem *Result) {
	selected, problem = r.selectLenses(cand.paths, cand.diff)
	if problem != nil {
		return "", "", nil, problem
	}

	text, problem = r.promptFor(cand, selected)
	if problem != nil {
		return "", "", nil, problem
	}
	return text, cand.tree, selected, nil
}

// promptFor builds the prompt for one group of lenses: the same text the printed
// form uses, for the lenses this reviewer owns.
func (r *runner) promptFor(cand candidate, selected []string) (string, *Result) {
	bodies, problem := r.lensBodies(selected)
	if problem != nil {
		return "", problem
	}

	// The store's potent rows are part of the question: a reviewer that knows
	// which findings are already carried files no duplicates, and a review fresh
	// from scratch would re-litigate bytes nobody changed.
	known, cleanRows, _, problem := r.potentRows(cand.tree)
	if problem != nil {
		return "", problem
	}

	return domain.Prompt(domain.PromptInput{
		Brief:       cand.brief,
		Tree:        cand.tree,
		Base:        cand.base,
		Paths:       cand.paths,
		Diff:        cand.diff,
		Lenses:      bodies,
		Known:       known,
		KnownPasses: cleanRows,
		Floor:       domain.FloorFor(cand.paths, cand.diff),
	}), nil
}

// coveredLenses is the lenses this tree's verdict already answered. A missing
// verdict covers nothing; a verdict that cannot be read blocks, because guessing
// would either pay twice or skip a lens nobody judged.
func (r *runner) coveredLenses(tree string) ([]string, *Result) {
	rel, err := domain.Path(tree)
	if err != nil {
		return nil, nil
	}
	path := fullPath(r.opts.Dir, rel)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check the review of this tree (%s): %s; fix the permissions and run odachi review again",
			rel, oneLine(err.Error())))
		return nil, &problem
	}
	if !exists {
		return nil, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		// A verdict that exists but cannot be read is not "no verdict": starting
		// paid reviewers and failing at the write would charge for nothing.
		problem := blocked(fmt.Sprintf("could not read the review of this tree (%s): %s; fix the permissions and run odachi review again",
			rel, oneLine(err.Error())))
		return nil, &problem
	}
	verdict, err := domain.Parse(data)
	if err != nil {
		problem := blocked(fmt.Sprintf("the review of this tree cannot be read (%s): %s; write it again with odachi review --attach <file>",
			rel, oneLine(err.Error())))
		return nil, &problem
	}
	return verdict.Lenses, nil
}

// unanswered is the lenses a run still has to judge: the ones the evidence asks
// for minus the ones this tree's verdict already answered.
func unanswered(required, covered []string) []string {
	done := make(map[string]bool, len(covered))
	for _, lens := range covered {
		done[strings.ToLower(strings.TrimSpace(lens))] = true
	}
	var todo []string
	for _, lens := range required {
		if !done[strings.ToLower(strings.TrimSpace(lens))] {
			todo = append(todo, lens)
		}
	}
	return todo
}

// containsLens reports whether the lens is already answered.
func containsLens(lenses []string, lens string) bool {
	for _, seen := range lenses {
		if strings.EqualFold(strings.TrimSpace(seen), strings.TrimSpace(lens)) {
			return true
		}
	}
	return false
}

// mergeVerdicts adds the lenses a run just judged to the verdict this tree
// already had. The old rows are copied, never recomputed: the file grows by lens,
// and an answer that was already paid for is not replaced by a second opinion.
func mergeVerdicts(existing, incoming domain.Verdict) domain.Verdict {
	merged := existing
	merged.Findings = append(append([]domain.Finding{}, existing.Findings...), incoming.Findings...)
	merged.Passes = append(append([]domain.Pass{}, existing.Passes...), incoming.Passes...)
	merged.Lenses = append(append([]string{}, existing.Lenses...), incoming.Lenses...)
	merged.Reviewer = existing.Reviewer + " + " + incoming.Reviewer
	merged.At = incoming.At
	merged.Verdict = ""

	models := make(map[string]string, len(existing.Models)+len(incoming.Models))
	for lens, model := range existing.Models {
		models[lens] = model
	}
	for lens, model := range incoming.Models {
		models[lens] = model
	}

	// A verdict written before the per-lens map carries one model for every lens
	// it covers. When the halves disagree, that attribution has to survive for the
	// old lenses or the identity gate loses what it is asked to check.
	if existing.Model != "" {
		for _, lens := range existing.Lenses {
			name := strings.ToLower(strings.TrimSpace(lens))
			if _, ok := models[name]; !ok {
				models[name] = existing.Model
			}
		}
	}
	if len(models) == 0 {
		models = nil
	}
	merged.Models = models

	// The single-model field only survives when both halves agree on it.
	if existing.Model == "" || incoming.Model == "" || existing.Model != incoming.Model {
		merged.Model = ""
	}
	return merged
}

// appendUnique adds value when it is not there yet, keeping the first-seen order.
func appendUnique(values []string, value string) []string {
	for _, seen := range values {
		if seen == value {
			return values
		}
	}
	return append(values, value)
}

// refute prints the refuter's prompt for a verdict someone already filed. It is
// the other half of the review: the reviewer argues that something is wrong, and
// the refuter has to be able to show, with a line of the diff, that it is not.
func (r *runner) refute() Result {
	// The evidence is the diff in front of you. Without one there is nothing to
	// cite, and a refutation without a citation is a vote, not a review.
	cand, problem := r.candidate()
	if problem != nil {
		return *problem
	}

	tree := cand.tree

	verdict, problem := r.namedVerdict(r.opts.Refute, tree)
	if problem != nil {
		return *problem
	}

	existing := domain.Refutation{}
	if filed, _ := r.storedRefutation(tree); filed != nil {
		existing = *filed
	}
	if _, blocking := domain.BlockingAfter(verdict.Findings, existing); !blocking {
		return nothingToRefute(tree)
	}

	text := domain.RefutePrompt(domain.RefuteInput{
		Tree:     tree,
		Base:     cand.base,
		Reviewer: verdict.Reviewer,
		Findings: verdict.Findings,
		Diff:     cand.diff,
	})

	return Result{
		Code:    CodeOK,
		Message: "refuter prompt for tree " + tree,
		Prompt:  text,
		Lenses:  verdict.Lenses,
	}
}

// attach stores a document someone already wrote — an agent, or a human. The file
// is validated before anything is written: a verdict the tool half-understands
// stored under a tree's name would claim more than it can back.
//
// Two documents can be attached: a verdict, and a refutation. They share no key
// and both are decoded strictly, so `Sniff` says which one this is instead of the
// tool guessing from the file name.
func (r *runner) attach() Result {
	data, err := r.files.Read(explicitPath(r.opts.Dir, r.opts.Attach))
	if err != nil {
		return blocked(fmt.Sprintf("could not read the verdict %s: %s; check the path", r.opts.Attach, oneLine(err.Error())))
	}

	if domain.Sniff(data) == domain.KindRefutation {
		refutation, err := domain.ParseRefutation(data)
		if err != nil {
			return refused(r.opts.Attach, err.Error())
		}
		return r.storeRefutation(refutation, "")
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		return refused(r.opts.Attach, err.Error())
	}
	if err := parsed.Validate(); err != nil {
		return refused(r.opts.Attach, err.Error())
	}

	// The verdict is filed inside the candidate: the store must be ignorable
	// before the tree it names is read. A malformed document was refused above,
	// so a refusal still writes nothing.
	if _, problem := r.ensureStore(); problem != nil {
		return *problem
	}
	return r.store(parsed, r.base(), nil)
}

// reviewerCommand is the command that actually runs for one model group: --exec
// with {model} replaced by the identity this group is judged with. A command that
// asks for a model odachi cannot name is refused instead of started, because a
// verdict from another model is a review nobody asked for.
func (r *runner) reviewerCommand(model string) (string, *Result) {
	asks := strings.Contains(r.opts.Exec, modelToken)

	model = strings.TrimSpace(model)
	if model == "" && asks && r.host != nil {
		model = strings.TrimSpace(r.host.Model(r.opts.Dir))
	}
	// The model replaces {model} inside a command line: an identifier that looks
	// like shell code is refused instead of run.
	checked, err := domain.NormalizeModel(model)
	if err != nil {
		res := blocked(oneLine(err.Error()) + "; the reviewer was not started")
		return "", &res
	}
	model = checked
	r.reviewedBy = model

	switch {
	case asks && model == "":
		res := blocked("--exec asks for " + modelToken + " and no model is known: pass --model <provider/model#variant>, name one in " + domain.ModePath + " or set ODACHI_MODEL; the reviewer was not started")
		return "", &res
	case !asks && model != "":
		res := blocked("a model was given (" + model + ") but --exec has no " + modelToken + " placeholder, so it would change nothing: add it to the command or drop the assignment; the reviewer was not started")
		return "", &res
	}

	return strings.ReplaceAll(r.opts.Exec, modelToken, model), nil
}

// validateCommand refuses a command that cannot run, before the store exists. The
// per-group command does it again with the model it resolved; this one exists so a
// refusal writes nothing, which is the promise the store's preparation depends on.
func (r *runner) validateCommand() *Result {
	asks := strings.Contains(r.opts.Exec, modelToken)
	assigned := strings.TrimSpace(r.opts.Model) != ""
	for _, model := range r.opts.LensModels {
		if strings.TrimSpace(model) != "" {
			assigned = true
			break
		}
	}

	switch {
	case asks && !assigned:
		host := ""
		if r.host != nil {
			host = strings.TrimSpace(r.host.Model(r.opts.Dir))
		}
		if host == "" {
			res := blocked("--exec asks for " + modelToken + " and no model is known: pass --model <provider/model#variant>, name one in " + domain.ModePath + " or set ODACHI_MODEL; the reviewer was not started")
			return &res
		}
	case !asks && assigned:
		res := blocked("a model is assigned (--model or " + domain.ModePath + ") but --exec has no " + modelToken + " placeholder, so it would change nothing: add it to the command or drop the assignment; the reviewer was not started")
		return &res
	}
	return nil
}

// modelFor is the model that judges one lens: --model wins for one run, then the
// project's table, and otherwise the command decides.
func (r *runner) modelFor(lens string) string {
	if model := strings.TrimSpace(r.opts.Model); model != "" {
		return model
	}
	return strings.TrimSpace(r.opts.LensModels[strings.ToLower(strings.TrimSpace(lens))])
}

// lensGroup is the lenses one reviewer run covers: they share a model, so they
// share a command. One group is one call to the model the human pays for.
type lensGroup struct {
	model  string
	lenses []string
}

// lensGroups groups the lenses by the model that judges them, in the order their
// first lens appears, so a run is deterministic.
func (r *runner) lensGroups(lenses []string) []lensGroup {
	var groups []lensGroup
	index := make(map[string]int, len(lenses))
	for _, lens := range lenses {
		name := strings.ToLower(strings.TrimSpace(lens))
		model := r.modelFor(name)
		i, ok := index[model]
		if !ok {
			i = len(groups)
			index[model] = i
			groups = append(groups, lensGroup{model: model})
		}
		groups[i].lenses = append(groups[i].lenses, name)
	}
	return groups
}

// exec runs the reviewer command and stores its answer. A command that fails is
// exec runs one reviewer per model group and stores one verdict. A command that
// fails is not a review: nothing is written, and its output is shown.
func (r *runner) exec() Result {
	// The candidate comes first: without a diff there is nothing to review, and a
	// reviewer must not be started to review nothing. It is read once, so every
	// group judges the same bytes.
	// The command is judged before the store exists: a placeholder nobody can fill
	// is a reviewer that must not be started, and a refused run writes nothing.
	if problem := r.validateCommand(); problem != nil {
		return *problem
	}

	cand, problem := r.candidateForReview()
	if problem != nil {
		return *problem
	}

	selected, problem := r.selectLenses(cand.paths, cand.diff)
	if problem != nil {
		return *problem
	}

	// A verdict for this tree may be incomplete, and a run never answers twice
	// the lenses already answered: the store grows by lens, it is not rewritten.
	covered, problem := r.coveredLenses(cand.tree)
	if problem != nil {
		return *problem
	}
	todo := unanswered(selected, covered)
	if len(todo) == 0 {
		rel, err := domain.Path(cand.tree)
		if err != nil {
			return blocked(oneLine(err.Error()) + "; I wrote nothing")
		}
		return alreadyStored(rel)
	}

	merged := domain.Verdict{Models: map[string]string{}}
	reviewers := make([]string, 0, len(todo))
	judges := make([]string, 0, len(todo))

	for _, group := range r.lensGroups(todo) {
		command, problem := r.reviewerCommand(group.model)
		if problem != nil {
			return *problem
		}

		// The command is a reviewer, so it gets the prompt the verb prints, for the
		// lenses this group owns. A reviewer started without one reviews whatever it
		// decides to look at, and that is not the candidate this verdict is filed
		// under.
		text, problem := r.promptFor(cand, group.lenses)
		if problem != nil {
			return *problem
		}

		stdout, stderr, exitCode, timedOut, err := r.shell.RunInput(r.opts.Dir, command, text, ExecTimeoutSeconds)
		// The answer is stdout. stderr is the command's own diagnosis - banners,
		// warnings, ANSI - and it is shown when something fails, never parsed as
		// part of the verdict.
		diagnosis := stdout
		if stderr != "" {
			diagnosis += stderr
		}
		switch {
		case err != nil:
			return execFailed(command, "could not run the reviewer: "+oneLine(err.Error()), diagnosis)
		case timedOut:
			return execFailed(command, fmt.Sprintf("the review did not finish in %ds", ExecTimeoutSeconds), diagnosis)
		case exitCode != 0:
			return execFailed(command, fmt.Sprintf("the reviewer exited with code %d", exitCode), diagnosis)
		}

		// A refutation is an answer too, and it is stored beside the verdict it judges.
		if domain.Sniff([]byte(stdout)) == domain.KindRefutation {
			refutation, err := domain.ParseRefutation([]byte(stdout))
			if err != nil {
				return execFailed(command, err.Error(), diagnosis)
			}
			// The refuter may have moved the worktree too; what it cites is the diff
			// it was given, and the file is filed under the tree that exists now.
			r.afterReviewer()
			return r.storeRefutation(refutation, command)
		}

		parsed, err := domain.Parse([]byte(stdout))
		if err != nil {
			return execFailed(command, err.Error(), diagnosis)
		}
		// The command is the reviewer: it is the one fact the tool knows about an
		// answer that came out of a shell.
		if strings.TrimSpace(parsed.Reviewer) == "" {
			parsed.Reviewer = command
		}
		// The identity that judged, when the command named one. It is the reviewer's
		// own declaration, and the gate compares it with the proof's.
		parsed.Model = r.reviewedBy
		if err := parsed.Validate(); err != nil {
			return execFailed(r.opts.Exec, err.Error(), diagnosis)
		}

		merged.Findings = append(merged.Findings, parsed.Findings...)
		merged.Passes = append(merged.Passes, parsed.Passes...)
		// The lenses are the ones this reviewer was asked to answer, not the ones
		// it claims: the tool stamps them, and the model that judged them is the
		// one this group ran with.
		for _, lens := range group.lenses {
			merged.Lenses = appendUnique(merged.Lenses, lens)
			if parsed.Model != "" {
				merged.Models[lens] = parsed.Model
			}
		}
		reviewers = appendUnique(reviewers, command)
		judges = appendUnique(judges, parsed.Model)
	}
	if len(merged.Models) == 0 {
		merged.Models = nil
	}
	merged.Reviewer = strings.Join(reviewers, " + ")
	// One model for the whole verdict is the single-model case, and the older
	// field keeps saying it. Several models are only in the per-lens map.
	if len(judges) == 1 {
		merged.Model = judges[0]
	}

	// The reviewer may have moved the worktree; the verdict must be filed under
	// the bytes that exist now, not the ones the prompt named.
	r.afterReviewer()

	// The base is the one the diff was taken against, not the flag: a project
	// with no commit is compared against the empty tree, and the verdict says so.
	return r.store(merged, cand.base, todo)
}

// show reads this tree's verdict, or says there is none. It writes nothing: a
// question is not a reason to touch the project.
//
// It also answers the two questions the gate will ask: which lenses this tree
// requires, which ones the verdict already covers, and whether a refutation was
// filed beside it.
func (r *runner) show() Result {
	work, problem := r.worktree()
	assessed := problem == nil
	if problem != nil {
		// --show answers even when the base cannot be compared: the verdict is
		// named by the tree, and a tree the tool can still read is enough to
		// find it. Hiding a verdict because a flag is wrong would be worse.
		tree, fallback := r.treeHash()
		if fallback != nil {
			return *fallback
		}
		work = &worktree{tree: tree}
	}
	tree := work.tree

	rel, err := domain.Path(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; check the git tree and run odachi review --show again")
	}

	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		if !assessed {
			// The candidate could not be read: the lenses it asks for are
			// unknown, and guessing them would answer a question the tool did
			// not measure.
			return missing(tree, nil, "")
		}
		required := r.requiredLenses(work)
		result := missing(tree, required, domain.Reason(work.paths, work.diff))
		result.Lines = append(result.Lines, r.modelPlan(required)...)
		return result
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		return blocked(fmt.Sprintf("the review of this tree cannot be read (%s): %s; write it again with odachi review --attach <file>",
			rel, oneLine(err.Error())))
	}

	lines := shown(tree, rel, parsed)
	lines = append(lines, r.lensCoverage(work, parsed)...)
	if assessed {
		lines = append(lines, r.modelPlan(r.requiredLenses(work))...)
	}

	result := Result{
		Code:    CodeOK,
		Message: "review of " + tree + ": " + parsed.Verdict,
		Lines:   lines,
		Verdict: &parsed,
		Lenses:  parsed.Lenses,
		Files:   []string{rel},
	}

	if refutation, refRel := r.storedRefutation(tree); refutation != nil {
		result.Refutation = refutation
		result.Files = append(result.Files, refRel)
		result.Lines = append(result.Lines, refutationLine(refutation, refRel, parsed.Findings))
	}
	return result
}

// modelPlan is one line per required lens: which model will judge it. It is the
// plan the human pays for, so --show names it before any reviewer runs. Without
// an assignment the command decides, and saying "the session's" would be a promise
// the tool does not keep when the command has no placeholder.
func (r *runner) modelPlan(required []string) []string {
	if len(required) == 0 {
		return nil
	}
	lines := make([]string, 0, len(required))
	for _, lens := range required {
		name := strings.ToLower(strings.TrimSpace(lens))
		model := strings.TrimSpace(r.opts.LensModels[name])
		if model == "" {
			model = "the reviewer command decides"
		}
		lines = append(lines, "model "+name+": "+model)
	}
	return lines
}

// lensCoverage is the two lines that make the optional gate predictable: what
// this tree requires and what the verdict covers.
func (r *runner) lensCoverage(work *worktree, verdict domain.Verdict) []string {
	required := r.requiredLenses(work)

	lines := []string{lensLine(required, verdict)}
	missing := verdict.Missing(required)
	if len(missing) == 0 {
		return lines
	}
	return append(lines, "missing lenses: "+strings.Join(missing, ", ")+" (odachi review --lens "+missing[0]+")")
}

// store seals a verdict for the current tree and writes it, unless this tree
// already has one: the store is immutable, and a second verdict is reported
// instead of applied.
//
// The lenses are stamped here, like the tree and the instant: the reviewer never
// declares which lenses it used, and the candidate is what says which ones were
// required. A stored verdict that does not cover them would leave the optional
// gate with no way out, because the store cannot be rewritten.
func (r *runner) store(parsed domain.Verdict, base string, judged []string) Result {
	work, problem := r.worktree()
	if problem != nil {
		return *problem
	}
	tree := work.tree

	// The floor is one reading of the candidate: it fixes the class this verdict
	// is asked under, the audit set earned blocking judges, and the rows the
	// reviewer's own findings may cite. Measuring it once keeps all three
	// answers from drifting apart.
	floor, problem := r.floor()
	if problem != nil {
		return *problem
	}

	// A run that judged a subset says exactly which lenses it answered: the class
	// lenses the candidate requires may include ones another run already paid for,
	// and claiming them here would both lie and collide with the stored verdict.
	var stamped domain.Verdict
	if judged != nil {
		stamped = parsed
		stamped.Lenses = domain.Add(nil, judged)
		if len(stamped.Lenses) == 1 {
			for i, finding := range stamped.Findings {
				if strings.TrimSpace(finding.Lens) == "" {
					stamped.Findings[i].Lens = stamped.Lenses[0]
				}
			}
		}
	} else {
		stamped, problem = r.stampLenses(parsed, floor.Class)
		if problem != nil {
			return *problem
		}
	}

	stamped, problem = r.stampBlobs(stamped)
	if problem != nil {
		return *problem
	}

	var downgraded int
	stamped.Findings, downgraded = domain.Earned(stamped.Findings, floor)

	stamped, carried, notes, problem := r.inherit(stamped, tree)
	if problem != nil {
		return *problem
	}

	sealed, err := domain.Seal(stamped, tree, base, r.opts.Now)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; I wrote nothing")
	}

	rel, err := domain.Path(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; I wrote nothing")
	}
	target := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(target)
	if err != nil {
		return blocked(fmt.Sprintf("could not check whether %s already exists: %s; check the permissions", rel, oneLine(err.Error())))
	}
	if exists {
		// A verdict that already answered these lenses is not rewritten: the run
		// answers only what the tree still owes. A shared lens means this is not
		// that run, and the store says so instead of replacing an answer.
		existing, problem := r.storedVerdictFor(tree)
		if problem != nil {
			return *problem
		}
		for _, lens := range sealed.Lenses {
			if containsLens(existing.Lenses, lens) {
				return alreadyStored(rel)
			}
		}
		sealed = mergeVerdicts(existing, sealed)
	}

	if err := r.files.Write(target, domain.Render(sealed)); err != nil {
		return blocked(fmt.Sprintf("could not write the verdict %s: %s; check the permissions and run odachi review --attach <file> again",
			rel, oneLine(err.Error())))
	}

	message := "verdict saved in " + rel + ": " + sealed.Verdict
	lines := []string{message, "reviewer: " + sealed.Reviewer + " · " + sealed.At, lenses(sealed.Lenses)}
	if downgraded > 0 {
		lines = append(lines, fmt.Sprintf("earned blocking: %d finding(s) downgraded for citing outside the candidate", downgraded))
	}
	if carried.findings+carried.passes > 0 {
		lines = append(lines, fmt.Sprintf("carried from the store: %d potent finding(s) and %d potent pass(es) about unchanged bytes", carried.findings, carried.passes))
	}
	if carried.suppressed > 0 {
		lines = append(lines, fmt.Sprintf("kept dismissed: %d re-raised finding(s) discarded by fingerprint", carried.suppressed))
	}
	lines = append(lines, notes...)
	return Result{
		Code:    CodeOK,
		Message: message,
		Lines:   lines,
		Verdict: &sealed,
		Lenses:  sealed.Lenses,
		Files:   []string{rel},
	}
}

// storeRefutation seals a refuter's answer and writes it beside the verdict,
// once. Everything that can be wrong with it is refused before the write, so a
// refusal leaves no file and nothing to clean up.
func (r *runner) storeRefutation(answer domain.Refutation, refuter string) Result {
	// The order matters: the checks that decide whether this document may exist
	// run before the store is prepared, so a refusal writes nothing at all.
	work, problem := r.worktree()
	if problem != nil {
		return *problem
	}
	tree := work.tree

	verdict, problem := r.storedVerdictFor(tree)
	if problem != nil {
		return *problem
	}

	if strings.TrimSpace(answer.Refuter) == "" {
		answer.Refuter = refuter
	}

	sealed, err := domain.SealRefutation(answer, verdict.Findings, tree, r.opts.Now)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; I wrote nothing")
	}

	wrote, problem := r.ensureStore()
	if problem != nil {
		return *problem
	}
	if wrote {
		r.work = nil
		after, problem := r.worktree()
		if problem != nil {
			return *problem
		}
		if after.tree != tree {
			return blocked("the tree changed while preparing the review store: run odachi review again")
		}
	}

	rel, err := domain.RefutationPath(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; I wrote nothing")
	}
	target := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(target)
	if err != nil {
		return blocked(fmt.Sprintf("could not check whether %s already exists: %s; check the permissions", rel, oneLine(err.Error())))
	}
	if exists {
		return alreadyRefuted(rel)
	}

	if err := r.files.Write(target, domain.RenderRefutation(sealed)); err != nil {
		return blocked(fmt.Sprintf("could not write the refutation %s: %s; check the permissions", rel, oneLine(err.Error())))
	}

	stands, refuted := counts(sealed)
	message := "refutation saved in " + rel + " · " + domain.Stands + ": " + fmt.Sprint(stands) + " · " + domain.Refuted + ": " + fmt.Sprint(refuted)
	return Result{
		Code:       CodeOK,
		Message:    "refutation saved in " + rel,
		Lines:      []string{message},
		Refutation: &sealed,
		Lenses:     verdict.Lenses,
		Files:      []string{rel},
	}
}

// stampLenses records the lenses this candidate's class asks for, plus the ones
// the human asked for, and whatever an attached verdict declared on its own. The
// class comes from the floor, the same reading the prompt used. When exactly one
// lens answers this tree, every finding without a lens is its answer: with one
// specialist running, the attribution is not in doubt.
func (r *runner) stampLenses(verdict domain.Verdict, class domain.Tier) (domain.Verdict, *Result) {
	extra, err := domain.ParseLenses(r.opts.Lens)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; I wrote nothing")
		return domain.Verdict{}, &problem
	}

	stamped := verdict
	stamped.Lenses = domain.Add(domain.Add(domain.Lenses(class), extra), verdict.Lenses)

	if len(stamped.Lenses) == 1 {
		for i, finding := range stamped.Findings {
			if strings.TrimSpace(finding.Lens) == "" {
				stamped.Findings[i].Lens = stamped.Lenses[0]
			}
		}
	}
	return stamped, nil
}

// stampBlobs gives every finding its content receipt: the blob id of the file as
// it stands in this worktree, measured by the git port and never read from the
// verdict. Whatever the reviewer declared is overwritten — a hash is a fact
// about bytes, and facts of this kind are computed, not declared.
//
// A path the port cannot hash answers "" with no error: the candidate deleted
// that file, or the reviewer invented the path. That finding is stored without a
// receipt, and the gate later has no implicit coverage to pretend: it is the
// blob-less finding the coverage arithmetic refuses to extend.
func (r *runner) stampBlobs(verdict domain.Verdict) (domain.Verdict, *Result) {
	stamped := verdict
	for i, finding := range stamped.Findings {
		blob, problem := r.stampBlob("the finding's content receipt", finding.Path)
		if problem != nil {
			return domain.Verdict{}, problem
		}
		stamped.Findings[i].Blob = blob
	}

	// A pass is stamped with the same rule as a finding: a row that is clean
	// must name the bytes that were measured clean. Only the reviewer's fresh
	// rows reach this function, so a stamping here can only strengthen the
	// answer, never make it older than it is.
	for i, clean := range stamped.Passes {
		blob, problem := r.stampBlob("the pass's content receipt", clean.Path)
		if problem != nil {
			return domain.Verdict{}, problem
		}
		stamped.Passes[i].Blob = blob
	}
	return stamped, nil
}

// stampBlob is one content receipt measured. A path the port cannot hash gets
// no receipt and no error: a file that is not in the worktree (the candidate
// deleted it, or the reviewer invented the path) has no bytes to stamp. A real
// git failure blocks the storing — receipts of that kind are guesses.
func (r *runner) stampBlob(what, path string) (string, *Result) {
	blob, err := r.git.BlobHash(r.opts.Dir, path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not hash %s to stamp %s: %s; I wrote nothing", path, what, oneLine(err.Error())))
		return "", &problem
	}
	return blob, nil
}

// carriedRows is how much the store lent to one verdict, and how many fresh
// claims it refused because a dismissal already judged them.
type carriedRows struct {
	findings   int
	passes     int
	suppressed int
}

// inherit adds the rows the store already holds potent about these exact
// bytes: findings that stood in their verdict and still describe what git
// measures now, and passes about unchanged bytes. The reviewer's own answer
// goes first, so a re-report of the same defect is filed once.
//
// Potency is per blob: a receipt for other bytes is a file that changed since
// that verdict, and the answer is dead. A finding refuted with evidence dies
// with its verdict — a refutation is index-based and answers about one
// verdict's findings, so it never travels.
func (r *runner) inherit(verdict domain.Verdict, tree string) (domain.Verdict, carriedRows, []string, *Result) {
	knowFindings, knowPasses, notes, problem := r.potentRows(tree)
	if problem != nil {
		return domain.Verdict{}, carriedRows{}, nil, problem
	}

	carried := carriedRows{}

	// A reused dismissal wins over a fresh claim, deterministically: the same
	// fingerprint was judged with evidence, and a re-report does not undo a
	// judgment. The row still travels, dismissed, so the next candidate keeps
	// the reason.
	dismissed := make(map[string]bool)
	for _, finding := range knowFindings {
		if domain.IsDismissed(finding) {
			dismissed[domain.Fingerprint(finding)] = true
		}
	}

	merged := verdict
	kept := make([]domain.Finding, 0, len(verdict.Findings))
	for _, finding := range verdict.Findings {
		if !domain.IsDismissed(finding) && dismissed[domain.Fingerprint(finding)] {
			carried.suppressed++
			continue
		}
		kept = append(kept, finding)
	}
	merged.Findings = kept

	seen := make(map[string]bool, len(merged.Findings)+len(knowFindings))
	for _, finding := range merged.Findings {
		seen[findingKey(finding)] = true
	}
	for _, finding := range knowFindings {
		if seen[findingKey(finding)] {
			continue
		}
		seen[findingKey(finding)] = true
		merged.Findings = append(merged.Findings, finding)
		carried.findings++
	}

	clean := make(map[string]bool, len(verdict.Passes)+len(knowPasses))
	for _, pass := range merged.Passes {
		clean[passKey(pass)] = true
	}
	for _, pass := range knowPasses {
		if clean[passKey(pass)] {
			continue
		}
		clean[passKey(pass)] = true
		merged.Passes = append(merged.Passes, pass)
		carried.passes++
	}
	return merged, carried, notes, nil
}

// potentRows is what the store holds potent right now: rows of previous
// verdicts whose receipts still describe what git measures, minus the findings
// the refuter killed with evidence. `notes` reports neighbor files that could
// not be read so neither — a broken neighbor never blocks a verdict, but
// nobody is told it said nothing.
func (r *runner) potentRows(verdictTree string) (findings []domain.Finding, passes []domain.Pass, notes []string, problem *Result) {
	names, err := r.files.List(fullPath(r.opts.Dir, domain.Dir))
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read the review store to weigh the potent rows: %s; I wrote nothing", oneLine(err.Error())))
		return nil, nil, nil, &problem
	}

	own, err := domain.Path(verdictTree)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; I wrote nothing")
		return nil, nil, nil, &problem
	}

	blobs := make(map[string]string)
	measure := func(path string) (string, error) {
		if blob, ok := blobs[path]; ok {
			return blob, nil
		}
		blob, err := r.git.BlobHash(r.opts.Dir, path)
		if err != nil {
			return "", err
		}
		blobs[path] = blob
		return blob, nil
	}

	for _, name := range names {
		tree, isVerdict := verdictName(name)
		if !isVerdict || name == filepath.Base(own) {
			continue
		}

		origin, note := r.storedNeighbor(tree)
		if note != "" {
			notes = append(notes, note)
			continue
		}

		killed, note := r.refutedFindings(tree)
		if note != "" {
			notes = append(notes, note)
			continue
		}

		for i, finding := range origin.Findings {
			if reason, killed := killed[domain.FindingIDs(origin.Findings)[i]]; killed {
				finding = domain.Dismiss(finding, reason)
			}
			if !domain.IsLens(finding.Lens) || strings.TrimSpace(finding.Blob) == "" {
				continue
			}
			if domain.IsDismissed(finding) {
				if strings.TrimSpace(finding.Reason) == "" {
					continue
				}
				if strings.TrimSpace(finding.Fingerprint) == "" {
					finding.Fingerprint = domain.Fingerprint(finding)
				}
			}
			blob, err := measure(finding.Path)
			if err != nil {
				problem := blocked(fmt.Sprintf("could not hash %s to weigh the store's potent rows: %s; I wrote nothing", finding.Path, oneLine(err.Error())))
				return nil, nil, nil, &problem
			}
			if blob != "" && blob == strings.TrimSpace(finding.Blob) {
				findings = append(findings, finding)
			}
		}

		for _, pass := range origin.Passes {
			if !domain.IsLens(pass.Lens) || strings.TrimSpace(pass.Blob) == "" {
				continue
			}
			blob, err := measure(pass.Path)
			if err != nil {
				problem := blocked(fmt.Sprintf("could not hash %s to weigh the store's potent rows: %s; I wrote nothing", pass.Path, oneLine(err.Error())))
				return nil, nil, nil, &problem
			}
			if blob != "" && blob == strings.TrimSpace(pass.Blob) {
				passes = append(passes, pass)
			}
		}
	}
	return findings, passes, notes, nil
}

// verdictName is the tree a store file names, when the name has the verdict's
// shape: one hash and the .json suffix. Refutations end differently and are
// read through their own path; anything else is not ours.
func verdictName(name string) (string, bool) {
	if !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".refuted.json") {
		return "", false
	}
	tree := strings.TrimSuffix(name, ".json")
	if !hexName(tree) {
		return "", false
	}
	return tree, true
}

// hexName reports whether every letter of the name is a hex digit, which is
// what makes a store name a tree and keeps a path from hiding inside one.
func hexName(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

// storedNeighbor reads one store verdict. A file this tool cannot read is
// reported, not skipped in silence: silence would be a lie about what the
// store said.
func (r *runner) storedNeighbor(tree string) (domain.Verdict, string) {
	data, err := r.files.Read(fullPath(r.opts.Dir, domain.Dir+"/"+tree+".json"))
	if err != nil {
		return domain.Verdict{}, fmt.Sprintf("the stored verdict %s could not be read; its potent rows are not carried", tree+".json")
	}
	parsed, err := domain.Parse(data)
	if err != nil {
		return domain.Verdict{}, fmt.Sprintf("the stored verdict %s cannot be parsed; its potent rows are not carried", tree+".json")
	}
	return parsed, ""
}

// refutedFindings is the ids the refuter killed with evidence, and the reason it
// cited. A verdict with no refutation has nothing killed, which is the normal
// case, not an error. A refutation whose entry has no evidence refutes nothing:
// the default is always that the finding stands.
func (r *runner) refutedFindings(tree string) (map[string]string, string) {
	rel, err := domain.RefutationPath(tree)
	if err != nil {
		return nil, fmt.Sprintf("the refutation path of %s cannot be built: %s", tree, oneLine(err.Error()))
	}
	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		return nil, "" // no refutation stored: nothing was killed
	}
	parsed, err := domain.ParseRefutation(data)
	if err != nil {
		return nil, fmt.Sprintf("the refutation of %s cannot be read; its verdict's rows are not carried", tree)
	}

	killed := make(map[string]string, len(parsed.Refutations))
	for _, judgment := range parsed.Refutations {
		if judgment.Verdict == domain.Refuted && strings.TrimSpace(judgment.Why) != "" {
			killed[judgment.ID] = strings.TrimSpace(judgment.Why)
		}
	}
	return killed, ""
}

// findingKey is the identity a finding is filed under: the same file, bytes,
// lens, place and severity is one defect, however many verdicts saw it.
func findingKey(finding domain.Finding) string {
	return finding.Path + "|" + finding.Blob + "|" + finding.Lens + "|" +
		strconv.Itoa(finding.Line) + "|" + string(finding.Severity)
}

// passKey is the identity of one clean row: the same file, bytes and lens.
func passKey(pass domain.Pass) string {
	return pass.Path + "|" + pass.Blob + "|" + pass.Lens
}

// selectLenses is the selection for one candidate: what the evidence asks for,
// the path class raised by the diff's content, plus what the flags asked for.
// Extra lenses add, never remove.
func (r *runner) selectLenses(paths []string, diff string) ([]string, *Result) {
	extra, problem := r.lensFlag()
	if problem != nil {
		return nil, problem
	}
	return domain.Add(domain.Lenses(domain.Assess(paths, diff)), extra), nil
}

// lensFlag is the --lens vocabulary, parsed once. A flag that names no lens is
// refused before anything is read or written.
func (r *runner) lensFlag() ([]string, *Result) {
	extra, err := domain.ParseLenses(r.opts.Lens)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; run odachi review -h to see the flags")
		return nil, &problem
	}
	return extra, nil
}

// requiredLenses is the selection for a worktree read without a candidate:
// `--show` must still answer when git cannot compare, so it uses whatever the
// one read produced and the path class answers alone when the diff is missing.
func (r *runner) requiredLenses(work *worktree) []string {
	extra, problem := r.lensFlag()
	if problem != nil {
		extra = nil
	}
	return domain.Add(domain.Lenses(domain.Assess(work.paths, work.diff)), extra)
}

// lensBodies loads the prompts of the selected lenses through the port.
func (r *runner) lensBodies(names []string) ([]domain.Lens, *Result) {
	bodies := make([]domain.Lens, 0, len(names))
	for _, name := range names {
		if r.lenses == nil {
			problem := blocked("this binary has no lenses: install it again from a release and run odachi review")
			return nil, &problem
		}

		body, err := r.lenses.Lens(name)
		if err != nil {
			problem := blocked("could not load the lens " + name + ": " + oneLine(err.Error()) + "; run odachi review without --lens to use the ones the evidence asks for")
			return nil, &problem
		}
		bodies = append(bodies, domain.Lens{Name: name, Body: body})
	}
	return bodies, nil
}

// storedVerdictFor reads the verdict stored for one tree. A refutation needs the
// findings it judges, so it cannot exist without a verdict.
func (r *runner) storedVerdictFor(tree string) (domain.Verdict, *Result) {
	rel, err := domain.Path(tree)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; check the git tree and run odachi review again")
		return domain.Verdict{}, &problem
	}

	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		problem := noVerdictToRefute(tree)
		return domain.Verdict{}, &problem
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		problem := blocked(fmt.Sprintf("the review of this tree cannot be read (%s): %s; write it again with odachi review --attach <file>",
			rel, oneLine(err.Error())))
		return domain.Verdict{}, &problem
	}
	return parsed, nil
}

// noVerdictToRefute says a refutation needs a verdict to judge, and names the
// command that produces one. Without it, the refuter would be arguing with
// findings nobody filed.
func noVerdictToRefute(tree string) Result {
	message := "there is no review for this tree: a refutation judges the findings of a verdict; run odachi review"
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{"there is no review for this tree (" + tree + ")", "next: odachi review"},
	}
}

// namedVerdict reads a verdict file the human named and refuses one from another
// tree: the evidence a refuter cites is the diff in front of it.
func (r *runner) namedVerdict(path, tree string) (domain.Verdict, *Result) {
	data, err := r.files.Read(explicitPath(r.opts.Dir, path))
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read the verdict %s: %s; check the path", path, oneLine(err.Error())))
		return domain.Verdict{}, &problem
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		problem := refused(path, err.Error())
		return domain.Verdict{}, &problem
	}
	if err := parsed.Validate(); err != nil {
		problem := refused(path, err.Error())
		return domain.Verdict{}, &problem
	}

	if parsed.Tree != "" && parsed.Tree != tree {
		problem := blocked(fmt.Sprintf("the verdict %s is from tree %s and this tree is %s: a verdict is refuted with the diff that saw it born",
			path, parsed.Tree, tree))
		return domain.Verdict{}, &problem
	}
	return parsed, nil
}

// storedRefutation reads the refutation filed beside this tree's verdict, if
// any. An absent one is not an error: it is the normal state of a reviewed tree.
func (r *runner) storedRefutation(tree string) (*domain.Refutation, string) {
	rel, err := domain.RefutationPath(tree)
	if err != nil {
		return nil, ""
	}

	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		return nil, ""
	}

	parsed, err := domain.ParseRefutation(data)
	if err != nil {
		return nil, rel
	}
	return &parsed, rel
}

// base is the ref the candidate is compared against. An empty --base means the
// empty tree, exactly as in `check`: compare against nothing.
func (r *runner) base() string {
	if trimmed := strings.TrimSpace(r.opts.Base); trimmed != "" {
		return trimmed
	}
	return EmptyTree
}

// worktree reads the current candidate once. Every later question about the
// candidate - the prompt, the floor, the blobs, the store - reads this value;
// a second read of the same bytes was the cost this verb used to pay.
func (r *runner) worktree() (*worktree, *Result) {
	if r.work != nil {
		return r.work, nil
	}

	base := r.base()
	tree, paths, diff, err := r.git.Candidate(r.opts.Dir, base)
	if err != nil && base == DefaultBase {
		// A project with no commit yet has no HEAD: everything uncommitted is
		// the candidate, and the empty tree says exactly that. A base the human
		// named is never silently replaced.
		if retryTree, retryPaths, retryDiff, retryErr := r.git.Candidate(r.opts.Dir, EmptyTree); retryErr == nil {
			tree, paths, diff, err = retryTree, retryPaths, retryDiff, nil
			base = EmptyTree
		}
	}
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read what changed against %q: %s; fix --base %q or drop the flag to compare against HEAD",
			r.base(), oneLine(err.Error()), r.base()))
		return nil, &problem
	}

	r.work = &worktree{tree: tree, base: base, paths: paths, diff: diff}
	return r.work, nil
}

// afterReviewer drops the candidate after the reviewer command ran: the command
// is arbitrary, and a reviewer that edits the tree must not have its answer
// filed under bytes it did not judge.
func (r *runner) afterReviewer() {
	r.work = nil
	r.storeReady = false
}

// candidate gathers the brief and the worktree. Without a diff there is nothing
// to review, and the run stops there - before any write.
func (r *runner) candidate() (candidate, *Result) {
	briefText, problem := r.brief()
	if problem != nil {
		return candidate{}, problem
	}

	work, problem := r.worktree()
	if problem != nil {
		return candidate{}, problem
	}
	if len(work.paths) == 0 || strings.TrimSpace(work.diff) == "" {
		problem := nothingToReview()
		return candidate{}, &problem
	}

	return candidate{brief: briefText, base: work.base, tree: work.tree, paths: work.paths, diff: work.diff}, nil
}

// candidateForReview is candidate() plus the store preparation: a verdict is
// filed inside the candidate, so its directory must be ignorable before the
// tree it names is final. The write usually changes nothing; when it does, the
// candidate is read again so the prompt names the tree the verdict will carry.
func (r *runner) candidateForReview() (candidate, *Result) {
	cand, problem := r.candidate()
	if problem != nil {
		return candidate{}, problem
	}

	// The flag is validated before the store write: an invented lens must not
	// leave a file behind while refusing.
	if _, problem := r.lensFlag(); problem != nil {
		return candidate{}, problem
	}

	wrote, problem := r.ensureStore()
	if problem != nil {
		return candidate{}, problem
	}
	if !wrote {
		return cand, nil
	}

	r.work = nil
	return r.candidate()
}

// floor is the deterministic audit set of the candidate being stored: which
// lines the reviewer's own findings may cite, and which categories they touch.
// It fails closed — without the diff there is no citation rule, and a verdict
// whose blocking claims nobody can check is worse than no verdict.
func (r *runner) floor() (domain.Floor, *Result) {
	work, problem := r.worktree()
	if problem != nil {
		return domain.Floor{}, problem
	}
	return domain.FloorFor(work.paths, work.diff), nil
}

// brief reads brief.md. A missing brief is a state, not an error: the review is
// still possible, and the prompt says the brief is missing instead of the verb
// holding a second gate it does not own.
func (r *runner) brief() (string, *Result) {
	path := fullPath(r.opts.Dir, briefFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s: %s; check the permissions and run odachi review again", briefFile, oneLine(err.Error())))
		return "", &problem
	}
	if !exists {
		return noBrief, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s; check the permissions and run odachi review again", briefFile, oneLine(err.Error())))
		return "", &problem
	}
	if trimmed := strings.TrimSpace(string(data)); trimmed != "" {
		return trimmed, nil
	}
	return noBrief, nil
}

// ensureStore makes the store's directory ignorable before the tree that names
// the verdict is read. The order is the whole trick: the verdict is written
// inside the candidate, so without an ignore rule, writing it would change the
// tree that names it and every verdict would invalidate itself.
//
// The rule is asked of the repository first. A project that ran `odachi init`
// carries the block in its own .gitignore, and there the store's own file would
// only be a new path in the worktree - the shift this function exists to avoid.
// A probe that cannot answer is not permission to skip the file: a verdict must
// never move the tree because git failed, so the fallback is the store's own
// ignore file, written before the tree is read.
//
// It reports whether it wrote the file: a write can remove earlier verdicts from
// the candidate, and the caller decides whether the read has to happen again.
func (r *runner) ensureStore() (bool, *Result) {
	if r.storeReady {
		return false, nil
	}
	r.storeReady = true

	if ignored, err := r.git.Ignored(r.opts.Dir, domain.Dir); err == nil && ignored {
		return false, nil
	}

	rel := ignoreRel()
	path := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s: %s; check the permissions", rel, oneLine(err.Error())))
		return false, &problem
	}
	if exists {
		return false, nil
	}
	if err := r.files.Write(path, []byte(domain.IgnoreContent)); err != nil {
		problem := blocked(fmt.Sprintf("could not prepare %s: %s; check the permissions", rel, oneLine(err.Error())))
		return false, &problem
	}
	r.storeIgnoreWritten = true
	return true, nil
}

// treeHash is the current worktree, which is what the verdict is named after.
func (r *runner) treeHash() (string, *Result) {
	tree, err := r.git.TreeHash(r.opts.Dir)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read the git tree: %s; run odachi review inside a git repository and check that git is installed",
			oneLine(err.Error())))
		return "", &problem
	}
	return tree, nil
}

// shown is the block `--show` prints: what the verdict is, who wrote it, and the
// findings in the order they were filed.
func shown(tree, rel string, verdict domain.Verdict) []string {
	lines := []string{
		"review of " + tree + ": " + verdict.Verdict,
		"reviewer: " + verdict.Reviewer + " · " + verdict.At + " · base " + verdict.Base,
		"file: " + rel,
	}

	if len(verdict.Findings) == 0 {
		return append(lines, "no findings: "+domain.NoFindings)
	}

	for _, finding := range verdict.Findings {
		lines = append(lines, "- "+findingLine(finding))
	}
	return lines
}

// findingLine is one finding, in the same shape the prompt asked for it: the
// file, the severity, whether the candidate caused it, and the evidence. A
// dismissal also prints the reason that judged it: a verdict says what it
// discarded, not only what it kept.
func findingLine(finding domain.Finding) string {
	where := finding.Path
	if finding.Line > 0 {
		where = fmt.Sprintf("%s:%d", finding.Path, finding.Line)
	}

	causal := "prior"
	if finding.Causal {
		causal = "causal"
	}
	line := where + " · " + string(finding.Severity) + " · " + causal + " · " + finding.Why
	if domain.IsDismissed(finding) {
		line += " · dismissed: " + finding.Reason
	}
	return line
}

// ignoreRel is the store's ignore file, relative to the project.
func ignoreRel() string {
	return domain.Dir + "/" + domain.IgnoreName
}

// fullPath resolves a project-relative path, accepted with forward slashes on
// every platform.
func fullPath(dir, rel string) string {
	return filepath.Join(dir, filepath.FromSlash(rel))
}

// explicitPath resolves a path the human typed. A relative one lives inside the
// project, so the command works the same from any directory; an absolute one is
// used exactly as given, because resolving it inside the project would turn a
// real file into an unreadable name.
func explicitPath(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return fullPath(dir, path)
}

// nothingToReview is the answer when the worktree is identical to the base: no
// candidate, no review. It is the one message the spec fixes verbatim.
func nothingToReview() Result {
	return Result{
		Code:    CodeBlocked,
		Message: "there is nothing to review in this tree",
		Lines:   []string{"there is nothing to review in this tree"},
	}
}

// nothingToRefute is the answer for a verdict with nothing blocking: an approval
// has no blocker to argue with, and inventing work is worse than saying so.
func nothingToRefute(tree string) Result {
	message := "the verdict of " + tree + " has no causal BLOCKER or CRITICAL: there is nothing to refute"
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message, "next: odachi review --show"},
	}
}

// alreadyRefuted says this tree already has a refutation. It is written once, so
// the second answer is reported and dropped instead of overwriting the first.
func alreadyRefuted(rel string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: "this tree already has a refutation",
		Lines: []string{
			"this tree already has a refutation: " + rel,
			"next: odachi review --show",
		},
	}
}

// promptMessage names the tree, and the lenses when the evidence asked for any.
func promptMessage(tree string, lenses []string) string {
	message := "review prompt for tree " + tree
	if len(lenses) > 0 {
		message += " · lenses " + strings.Join(lenses, ", ")
	}
	return message
}

// lenses is the one line a stored verdict adds about its coverage.
func lenses(covered []string) string {
	if len(covered) == 0 {
		return "lenses: none (the evidence asked for none)"
	}
	return "lenses: " + strings.Join(covered, ", ")
}

// lensLine is the coverage `--show` prints: what the verdict covers and what this
// tree requires.
func lensLine(required []string, verdict domain.Verdict) string {
	covered := "none"
	if len(verdict.Lenses) > 0 {
		covered = strings.Join(verdict.Lenses, ", ")
	}

	needed := "none"
	if len(required) > 0 {
		needed = strings.Join(required, ", ")
	}
	return "lenses: " + covered + " · required: " + needed
}

// refutationLine is the refutation as `--show` reports it: which file, and how
// many findings survived it.
func refutationLine(refutation *domain.Refutation, rel string, findings []domain.Finding) string {
	stands := 0
	refuted := 0
	for _, judgment := range refutation.Refutations {
		if judgment.Verdict == domain.Refuted {
			refuted++
			continue
		}
		stands++
	}

	line := fmt.Sprintf("refutation: %s · stands: %d · refuted: %d", rel, stands, refuted)
	if _, blocking := domain.BlockingAfter(findings, *refutation); !blocking {
		return line + " · no longer blocks anything"
	}
	return line + " · still blocking"
}

// counts is the same arithmetic for the message a refutation prints when stored.
func counts(refutation domain.Refutation) (stands, refuted int) {
	for _, judgment := range refutation.Refutations {
		if judgment.Verdict == domain.Refuted {
			refuted++
			continue
		}
		stands++
	}
	return stands, refuted
}

// missing is the answer for a tree without a verdict. When the candidate could
// be read it also says which lenses this tree requires and why: the gate is
// invisible until it runs, and --show is where a human decides whether the
// review pays before paying for it. The line says "required", not "asks for":
// --lens adds lenses the tree did not ask for, and the sentence must stay true.
func missing(tree string, required []string, reason string) Result {
	lines := []string{"there is no review for this tree (" + tree + ")"}
	switch {
	case reason == "":
		lines = append(lines, "next: odachi review")
	case len(required) == 0:
		lines = append(lines, "none is due: "+reason, "next: odachi review --lens all (ask anyway)")
	default:
		lines = append(lines, "required: "+strings.Join(required, ", ")+": "+reason, "next: odachi review")
	}
	return Result{
		Code:    CodeBlocked,
		Message: "there is no review for this tree",
		Lines:   lines,
	}
}

// alreadyStored says the tree is already judged. The store is immutable, so the
// second verdict is reported and dropped.
func alreadyStored(rel string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: "this tree already has a verdict",
		Lines: []string{
			"this tree already has a verdict: " + rel,
			"next: odachi review --show",
		},
	}
}

// refused is an attached verdict the tool will not store: one line, and nothing
// written.
func refused(path, reason string) Result {
	message := fmt.Sprintf("the verdict %s is no good: %s; I wrote nothing", path, oneLine(reason))
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message},
	}
}

// execFailed is a reviewer that did not answer a verdict. Its own output is part
// of the result: it is the only thing that says why.
func execFailed(command, reason, output string) Result {
	message := fmt.Sprintf("%s did not give a verdict: %s; I wrote nothing", command, oneLine(reason))
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message},
		Output:  output,
	}
}

// blocked is a result the agent can act on, in one line.
func blocked(message string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message},
	}
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
