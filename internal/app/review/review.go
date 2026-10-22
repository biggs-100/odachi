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
// internal/domain/review; every Spanish line lives here.
package review

import (
	"fmt"
	"path/filepath"
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
const noBrief = "(todavía no hay brief.md: corré odachi start)"

// Options is one `odachi review` invocation.
type Options struct {
	Dir string

	// Base is what the candidate is compared against. Empty means the empty
	// tree: everything counts as new.
	Base string
	// Exec is the reviewer command: it is run with the review prompt as its
	// input, and its answer is the verdict.
	Exec string
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
// check whether this tree has one, and write it.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
}

// Git is the repository port: which tree is here, what changed since the base,
// and the diff itself. The diff is what the prompt is built from, so a
// repository port without it could not describe a candidate.
type Git interface {
	TreeHash(root string) (string, error)
	ChangedPaths(root, base string) ([]string, error)
	Diff(root, base string) (string, error)
}

// Shell is the command port for --exec. A non-zero exit is the command's result,
// not an adapter error: err means it could not be run at all.
//
// RunInput is the form this verb uses: the reviewer is a reviewer because it
// receives the prompt, and how a prompt reaches a command is the shell adapter's
// business (stdin, and a file path for a command that takes one), not a decision
// this layer should know how to make.
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error)
	RunInput(dir, command, input string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error)
}

// Lenses is the port for the four specialist prompts. It returns the file
// verbatim; the domain strips the frontmatter when it builds the prompt, so the
// content stays complete on disk and readable by a human.
type Lenses interface {
	Lens(name string) (string, error)
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
func Run(opts Options, files Files, git Git, shell Shell, lenses Lenses) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	r := &runner{opts: opts, files: files, git: git, shell: shell, lenses: lenses}

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
}

// candidate is what is being reviewed: the brief it is judged against, the base
// it is compared to, the paths that changed and the diff itself.
type candidate struct {
	brief string
	base  string
	paths []string
	diff  string
}

// prompt prints the review prompt for the current candidate. It prepares the
// store first: the verdict lives inside the candidate, so the directory has to be
// ignored before the tree is hashed or storing the verdict would change the tree
// that names it.
func (r *runner) prompt() Result {
	cand, problem := r.candidate()
	if problem != nil {
		return *problem
	}

	text, tree, selected, problem := r.promptText(cand)
	if problem != nil {
		return *problem
	}

	return Result{
		Code:    CodeOK,
		Message: promptMessage(tree, selected),
		Prompt:  text,
		Lenses:  selected,
		Files:   []string{ignoreRel()},
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
	selected, problem = r.selectLenses(cand.paths)
	if problem != nil {
		return "", "", nil, problem
	}

	bodies, problem := r.lensBodies(selected)
	if problem != nil {
		return "", "", nil, problem
	}

	tree, problem = r.prepareStore()
	if problem != nil {
		return "", "", nil, problem
	}

	text = domain.Prompt(domain.PromptInput{
		Brief:  cand.brief,
		Tree:   tree,
		Base:   cand.base,
		Paths:  cand.paths,
		Diff:   cand.diff,
		Lenses: bodies,
	})
	return text, tree, selected, nil
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

	tree, problem := r.treeHash()
	if problem != nil {
		return *problem
	}

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
		Message: "prompt del refutador del árbol " + tree,
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
		return blocked(fmt.Sprintf("no pude leer el veredicto %s: %s; revisá la ruta", r.opts.Attach, oneLine(err.Error())))
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

	return r.store(parsed, r.base())
}

// exec runs the reviewer command and stores its answer. A command that fails is
// not a review: nothing is written, and its output is shown.
func (r *runner) exec() Result {
	// The candidate comes first: without a diff there is nothing to review, and
	// a reviewer must not be started to review nothing.
	cand, problem := r.candidate()
	if problem != nil {
		return *problem
	}

	// The command is a reviewer, so it gets the prompt the verb prints. A
	// reviewer started without one reviews whatever it decides to look at, and
	// that is not the candidate this verdict is filed under.
	text, _, _, problem := r.promptText(cand)
	if problem != nil {
		return *problem
	}

	output, exitCode, timedOut, err := r.shell.RunInput(r.opts.Dir, r.opts.Exec, text, ExecTimeoutSeconds)
	switch {
	case err != nil:
		return execFailed(r.opts.Exec, "no pude correr el revisor: "+oneLine(err.Error()), output)
	case timedOut:
		return execFailed(r.opts.Exec, fmt.Sprintf("la revisión no terminó en %ds", ExecTimeoutSeconds), output)
	case exitCode != 0:
		return execFailed(r.opts.Exec, fmt.Sprintf("el revisor salió con código %d", exitCode), output)
	}

	// A refutation is an answer too, and it is stored beside the verdict it judges.
	if domain.Sniff([]byte(output)) == domain.KindRefutation {
		refutation, err := domain.ParseRefutation([]byte(output))
		if err != nil {
			return execFailed(r.opts.Exec, err.Error(), output)
		}
		return r.storeRefutation(refutation, r.opts.Exec)
	}

	parsed, err := domain.Parse([]byte(output))
	if err != nil {
		return execFailed(r.opts.Exec, err.Error(), output)
	}
	// The command is the reviewer: it is the one fact the tool knows about an
	// answer that came out of a shell.
	if strings.TrimSpace(parsed.Reviewer) == "" {
		parsed.Reviewer = r.opts.Exec
	}
	if err := parsed.Validate(); err != nil {
		return execFailed(r.opts.Exec, err.Error(), output)
	}

	// The base is the one the diff was taken against, not the flag: a project
	// with no commit is compared against the empty tree, and the verdict says so.
	return r.store(parsed, cand.base)
}

// show reads this tree's verdict, or says there is none. It writes nothing: a
// question is not a reason to touch the project.
//
// It also answers the two questions the gate will ask: which lenses this tree
// requires, which ones the verdict already covers, and whether a refutation was
// filed beside it.
func (r *runner) show() Result {
	tree, problem := r.treeHash()
	if problem != nil {
		return *problem
	}

	rel, err := domain.Path(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; revisá el árbol de git y volvé a correr odachi review --show")
	}

	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		return missing(tree)
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		return blocked(fmt.Sprintf("el review de este árbol no se puede leer (%s): %s; volvé a escribirlo con odachi review --attach <archivo>",
			rel, oneLine(err.Error())))
	}

	lines := shown(tree, rel, parsed)
	lines = append(lines, r.lensCoverage(tree, parsed)...)

	result := Result{
		Code:    CodeOK,
		Message: "review de " + tree + ": " + parsed.Verdict,
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

// lensCoverage is the two lines that make the optional gate predictable: what
// this tree requires and what the verdict covers.
func (r *runner) lensCoverage(tree string, verdict domain.Verdict) []string {
	required := r.requiredLenses()

	lines := []string{lensLine(required, verdict)}
	missing := verdict.Missing(required)
	if len(missing) == 0 {
		return lines
	}
	return append(lines, "faltan lentes: "+strings.Join(missing, ", ")+" (odachi review --lens "+missing[0]+")")
}

// store seals a verdict for the current tree and writes it, unless this tree
// already has one: the store is immutable, and a second verdict is reported
// instead of applied.
//
// The lenses are stamped here, like the tree and the instant: the reviewer never
// declares which lenses it used, and the candidate is what says which ones were
// required. A stored verdict that does not cover them would leave the optional
// gate with no way out, because the store cannot be rewritten.
func (r *runner) store(parsed domain.Verdict, base string) Result {
	tree, problem := r.prepareStore()
	if problem != nil {
		return *problem
	}

	stamped, problem := r.stampLenses(parsed)
	if problem != nil {
		return *problem
	}

	sealed, err := domain.Seal(stamped, tree, base, r.opts.Now)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; no escribí nada")
	}

	rel, err := domain.Path(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; no escribí nada")
	}
	target := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(target)
	if err != nil {
		return blocked(fmt.Sprintf("no pude revisar si %s ya existe: %s; revisá los permisos", rel, oneLine(err.Error())))
	}
	if exists {
		return alreadyStored(rel)
	}

	if err := r.files.Write(target, domain.Render(sealed)); err != nil {
		return blocked(fmt.Sprintf("no pude escribir el veredicto %s: %s; revisá los permisos y volvé a correr odachi review --attach <archivo>",
			rel, oneLine(err.Error())))
	}

	message := "veredicto guardado en " + rel + ": " + sealed.Verdict
	return Result{
		Code:    CodeOK,
		Message: message,
		Lines:   []string{message, "reviewer: " + sealed.Reviewer + " · " + sealed.At, lenses(sealed.Lenses)},
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
	tree, problem := r.treeHash()
	if problem != nil {
		return *problem
	}

	verdict, problem := r.storedVerdictFor(tree)
	if problem != nil {
		return *problem
	}

	if strings.TrimSpace(answer.Refuter) == "" {
		answer.Refuter = refuter
	}

	sealed, err := domain.SealRefutation(answer, verdict.Findings, tree, r.opts.Now)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; no escribí nada")
	}

	prepared, problem := r.prepareStore()
	if problem != nil {
		return *problem
	}
	if prepared != tree {
		return blocked("el árbol cambió mientras preparaba el store de reviews: volvé a correr odachi review")
	}

	rel, err := domain.RefutationPath(tree)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; no escribí nada")
	}
	target := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(target)
	if err != nil {
		return blocked(fmt.Sprintf("no pude revisar si %s ya existe: %s; revisá los permisos", rel, oneLine(err.Error())))
	}
	if exists {
		return alreadyRefuted(rel)
	}

	if err := r.files.Write(target, domain.RenderRefutation(sealed)); err != nil {
		return blocked(fmt.Sprintf("no pude escribir la refutación %s: %s; revisá los permisos", rel, oneLine(err.Error())))
	}

	stands, refuted := counts(sealed)
	message := "refutación guardada en " + rel + " · " + domain.Stands + ": " + fmt.Sprint(stands) + " · " + domain.Refuted + ": " + fmt.Sprint(refuted)
	return Result{
		Code:       CodeOK,
		Message:    "refutación guardada en " + rel,
		Lines:      []string{message},
		Refutation: &sealed,
		Lenses:     verdict.Lenses,
		Files:      []string{rel},
	}
}

// stampLenses records the lenses this candidate required plus the ones the human
// asked for, and whatever an attached verdict declared on its own.
func (r *runner) stampLenses(verdict domain.Verdict) (domain.Verdict, *Result) {
	extra, err := domain.ParseLenses(r.opts.Lens)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; no escribí nada")
		return domain.Verdict{}, &problem
	}

	// A tree that cannot be compared yet still has its declared lenses: the
	// evidence list is empty, not fatal.
	paths, _, pathErr := r.changedPaths()
	if pathErr != nil {
		paths = nil
	}

	stamped := verdict
	stamped.Lenses = domain.Add(domain.Add(domain.Lenses(domain.Classify(paths)), extra), verdict.Lenses)
	return stamped, nil
}

// selectLenses is the selection for one candidate: what the evidence asks for,
// plus what the flags asked for. Extra lenses add, never remove.
func (r *runner) selectLenses(paths []string) ([]string, *Result) {
	extra, err := domain.ParseLenses(r.opts.Lens)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; corré odachi review -h para ver los flags")
		return nil, &problem
	}
	return domain.Add(domain.Lenses(domain.Classify(paths)), extra), nil
}

// requiredLenses is the selection for a tree read without a candidate: `--show`
// must still answer when git cannot compare, so a failure to read the changed
// paths is not a block here.
func (r *runner) requiredLenses() []string {
	extra, err := domain.ParseLenses(r.opts.Lens)
	if err != nil {
		extra = nil
	}

	paths, _, pathErr := r.changedPaths()
	if pathErr != nil {
		paths = nil
	}
	return domain.Add(domain.Lenses(domain.Classify(paths)), extra)
}

// lensBodies loads the prompts of the selected lenses through the port.
func (r *runner) lensBodies(names []string) ([]domain.Lens, *Result) {
	bodies := make([]domain.Lens, 0, len(names))
	for _, name := range names {
		if r.lenses == nil {
			problem := blocked("este binario no tiene lentes: volvé a instalarlo desde una release y corré odachi review")
			return nil, &problem
		}

		body, err := r.lenses.Lens(name)
		if err != nil {
			problem := blocked("no pude cargar la lente " + name + ": " + oneLine(err.Error()) + "; corré odachi review sin --lens para usar las que pide la evidencia")
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
		problem := blocked(oneLine(err.Error()) + "; revisá el árbol de git y volvé a correr odachi review")
		return domain.Verdict{}, &problem
	}

	data, err := r.files.Read(fullPath(r.opts.Dir, rel))
	if err != nil {
		problem := noVerdictToRefute(tree)
		return domain.Verdict{}, &problem
	}

	parsed, err := domain.Parse(data)
	if err != nil {
		problem := blocked(fmt.Sprintf("el review de este árbol no se puede leer (%s): %s; volvé a escribirlo con odachi review --attach <archivo>",
			rel, oneLine(err.Error())))
		return domain.Verdict{}, &problem
	}
	return parsed, nil
}

// noVerdictToRefute says a refutation needs a verdict to judge, and names the
// command that produces one. Without it, the refuter would be arguing with
// findings nobody filed.
func noVerdictToRefute(tree string) Result {
	message := "no hay review para este árbol: una refutación juzga los findings de un veredicto; corré odachi review"
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{"no hay review para este árbol (" + tree + ")", "siguiente: odachi review"},
	}
}

// namedVerdict reads a verdict file the human named and refuses one from another
// tree: the evidence a refuter cites is the diff in front of it.
func (r *runner) namedVerdict(path, tree string) (domain.Verdict, *Result) {
	data, err := r.files.Read(explicitPath(r.opts.Dir, path))
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer el veredicto %s: %s; revisá la ruta", path, oneLine(err.Error())))
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
		problem := blocked(fmt.Sprintf("el veredicto %s es del árbol %s y este árbol es %s: un veredicto se refuta con el diff que lo vio nacer",
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

// candidate gathers the brief, the changed paths and the diff. Without a diff
// there is nothing to review, and the run stops there — before any write.
func (r *runner) candidate() (candidate, *Result) {
	briefText, problem := r.brief()
	if problem != nil {
		return candidate{}, problem
	}

	paths, base, err := r.changedPaths()
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer qué cambió contra %q: %s; corregí --base %q o sacá el flag para comparar contra HEAD",
			r.base(), oneLine(err.Error()), r.base()))
		return candidate{}, &problem
	}
	if len(paths) == 0 {
		problem := nothingToReview()
		return candidate{}, &problem
	}

	diff, err := r.git.Diff(r.opts.Dir, base)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer el diff contra %q: %s; corregí --base %q o sacá el flag para comparar contra HEAD",
			base, oneLine(err.Error()), r.base()))
		return candidate{}, &problem
	}
	if strings.TrimSpace(diff) == "" {
		problem := nothingToReview()
		return candidate{}, &problem
	}

	return candidate{brief: briefText, base: base, paths: paths, diff: diff}, nil
}

// brief reads brief.md. A missing brief is a state, not an error: the review is
// still possible, and the prompt says the brief is missing instead of the verb
// holding a second gate it does not own.
func (r *runner) brief() (string, *Result) {
	path := fullPath(r.opts.Dir, briefFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi review", briefFile, oneLine(err.Error())))
		return "", &problem
	}
	if !exists {
		return noBrief, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer %s: %s; revisá los permisos y volvé a correr odachi review", briefFile, oneLine(err.Error())))
		return "", &problem
	}
	if trimmed := strings.TrimSpace(string(data)); trimmed != "" {
		return trimmed, nil
	}
	return noBrief, nil
}

// changedPaths is what changed and which base actually answered. A project with
// no commit yet has no HEAD, so the comparison falls back to the empty tree:
// everything uncommitted is the candidate. A base the human named is never
// silently replaced.
func (r *runner) changedPaths() ([]string, string, error) {
	base := r.base()

	paths, err := r.git.ChangedPaths(r.opts.Dir, base)
	if err == nil {
		return paths, base, nil
	}
	if base != DefaultBase {
		return nil, base, err
	}

	paths, retryErr := r.git.ChangedPaths(r.opts.Dir, EmptyTree)
	if retryErr != nil {
		return nil, base, err
	}
	return paths, EmptyTree, nil
}

// base is the ref the candidate is compared against. An empty --base means the
// empty tree, exactly as in `check`: compare against nothing.
func (r *runner) base() string {
	if trimmed := strings.TrimSpace(r.opts.Base); trimmed != "" {
		return trimmed
	}
	return EmptyTree
}

// prepareStore makes the store's directory ignorable and then hashes the
// worktree. The order is the whole trick: the verdict is written inside the
// candidate, so without the ignore file, writing it would change the tree that
// names it and every verdict would invalidate itself.
func (r *runner) prepareStore() (string, *Result) {
	rel := ignoreRel()
	path := fullPath(r.opts.Dir, rel)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude revisar %s: %s; revisá los permisos", rel, oneLine(err.Error())))
		return "", &problem
	}
	if !exists {
		if err := r.files.Write(path, []byte(domain.IgnoreContent)); err != nil {
			problem := blocked(fmt.Sprintf("no pude preparar %s: %s; revisá los permisos", rel, oneLine(err.Error())))
			return "", &problem
		}
	}

	return r.treeHash()
}

// treeHash is the current worktree, which is what the verdict is named after.
func (r *runner) treeHash() (string, *Result) {
	tree, err := r.git.TreeHash(r.opts.Dir)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer el árbol de git: %s; corré odachi review dentro de un repositorio git y revisá que git esté instalado",
			oneLine(err.Error())))
		return "", &problem
	}
	return tree, nil
}

// shown is the block `--show` prints: what the verdict is, who wrote it, and the
// findings in the order they were filed.
func shown(tree, rel string, verdict domain.Verdict) []string {
	lines := []string{
		"review de " + tree + ": " + verdict.Verdict,
		"reviewer: " + verdict.Reviewer + " · " + verdict.At + " · base " + verdict.Base,
		"archivo: " + rel,
	}

	if len(verdict.Findings) == 0 {
		return append(lines, "sin findings: "+domain.NoFindings)
	}

	for _, finding := range verdict.Findings {
		lines = append(lines, "- "+findingLine(finding))
	}
	return lines
}

// findingLine is one finding, in the same shape the prompt asked for it: the
// file, the severity, whether the candidate caused it, and the evidence.
func findingLine(finding domain.Finding) string {
	where := finding.Path
	if finding.Line > 0 {
		where = fmt.Sprintf("%s:%d", finding.Path, finding.Line)
	}

	causal := "previo"
	if finding.Causal {
		causal = "causal"
	}
	return where + " · " + string(finding.Severity) + " · " + causal + " · " + finding.Why
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
		Message: "no hay nada que revisar en este árbol",
		Lines:   []string{"no hay nada que revisar en este árbol"},
	}
}

// nothingToRefute is the answer for a verdict with nothing blocking: an approval
// has no blocker to argue with, and inventing work is worse than saying so.
func nothingToRefute(tree string) Result {
	message := "el veredicto de " + tree + " no tiene BLOCKER ni CRITICAL causal: no hay nada que refutar"
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message, "siguiente: odachi review --show"},
	}
}

// alreadyRefuted says this tree already has a refutation. It is written once, so
// the second answer is reported and dropped instead of overwriting the first.
func alreadyRefuted(rel string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: "ya hay una refutación para este árbol",
		Lines: []string{
			"ya hay una refutación para este árbol: " + rel,
			"siguiente: odachi review --show",
		},
	}
}

// promptMessage names the tree, and the lenses when the evidence asked for any.
func promptMessage(tree string, lenses []string) string {
	message := "prompt de review del árbol " + tree
	if len(lenses) > 0 {
		message += " · lentes " + strings.Join(lenses, ", ")
	}
	return message
}

// lenses is the one line a stored verdict adds about its coverage.
func lenses(covered []string) string {
	if len(covered) == 0 {
		return "lentes: ninguna (la evidencia no pidió ninguna)"
	}
	return "lentes: " + strings.Join(covered, ", ")
}

// lensLine is the coverage `--show` prints: what the verdict covers and what this
// tree requires.
func lensLine(required []string, verdict domain.Verdict) string {
	covered := "ninguna"
	if len(verdict.Lenses) > 0 {
		covered = strings.Join(verdict.Lenses, ", ")
	}

	needed := "ninguna"
	if len(required) > 0 {
		needed = strings.Join(required, ", ")
	}
	return "lentes: " + covered + " · requeridas: " + needed
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

	line := fmt.Sprintf("refutación: %s · se quedan: %d · refutados: %d", rel, stands, refuted)
	if _, blocking := domain.BlockingAfter(findings, *refutation); !blocking {
		return line + " · ya no bloquea nada"
	}
	return line + " · todavía bloquea"
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

// missing is the answer for a tree without a verdict.
func missing(tree string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: "no hay review para este árbol",
		Lines: []string{
			"no hay review para este árbol (" + tree + ")",
			"siguiente: odachi review",
		},
	}
}

// alreadyStored says the tree is already judged. The store is immutable, so the
// second verdict is reported and dropped.
func alreadyStored(rel string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: "ya hay un veredicto para este árbol",
		Lines: []string{
			"ya hay un veredicto para este árbol: " + rel,
			"siguiente: odachi review --show",
		},
	}
}

// refused is an attached verdict the tool will not store: one line, and nothing
// written.
func refused(path, reason string) Result {
	message := fmt.Sprintf("el veredicto %s no sirve: %s; no escribí nada", path, oneLine(reason))
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Lines:   []string{message},
	}
}

// execFailed is a reviewer that did not answer a verdict. Its own output is part
// of the result: it is the only thing that says why.
func execFailed(command, reason, output string) Result {
	message := fmt.Sprintf("%s no dio un veredicto: %s; no escribí nada", command, oneLine(reason))
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
