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
	// Exec is the reviewer command: it is run, and its answer is the verdict.
	Exec string
	// Attach is a verdict someone already wrote.
	Attach string
	// Show reads this tree's verdict instead of producing one.
	Show bool
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
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (output string, exitCode int, timedOut bool, err error)
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
	// Files are the files this run wrote, relative to the project.
	Files []string
}

// Run performs one `odachi review`.
func Run(opts Options, files Files, git Git, shell Shell) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	r := &runner{opts: opts, files: files, git: git, shell: shell}

	switch {
	case opts.Show:
		return r.show()
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
	opts  Options
	files Files
	git   Git
	shell Shell
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

	tree, problem := r.prepareStore()
	if problem != nil {
		return *problem
	}

	text := domain.Prompt(domain.PromptInput{
		Brief: cand.brief,
		Tree:  tree,
		Base:  cand.base,
		Paths: cand.paths,
		Diff:  cand.diff,
	})

	return Result{
		Code:    CodeOK,
		Message: "prompt de review del árbol " + tree,
		Prompt:  text,
		Files:   []string{ignoreRel()},
	}
}

// attach stores a verdict someone already wrote — an agent, or a human. The file
// is validated before anything is written: a verdict the tool half-understands
// stored under a tree's name would claim more than it can back.
func (r *runner) attach() Result {
	data, err := r.files.Read(explicitPath(r.opts.Dir, r.opts.Attach))
	if err != nil {
		return blocked(fmt.Sprintf("no pude leer el veredicto %s: %s; revisá la ruta", r.opts.Attach, oneLine(err.Error())))
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

	output, exitCode, timedOut, err := r.shell.Run(r.opts.Dir, r.opts.Exec, ExecTimeoutSeconds)
	switch {
	case err != nil:
		return execFailed(r.opts.Exec, "no pude correr el revisor: "+oneLine(err.Error()), output)
	case timedOut:
		return execFailed(r.opts.Exec, fmt.Sprintf("la revisión no terminó en %ds", ExecTimeoutSeconds), output)
	case exitCode != 0:
		return execFailed(r.opts.Exec, fmt.Sprintf("el revisor salió con código %d", exitCode), output)
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

	result := Result{
		Code:    CodeOK,
		Message: "review de " + tree + ": " + parsed.Verdict,
		Lines:   shown(tree, rel, parsed),
		Verdict: &parsed,
		Files:   []string{rel},
	}
	return result
}

// store seals a verdict for the current tree and writes it, unless this tree
// already has one: the store is immutable, and a second verdict is reported
// instead of applied.
func (r *runner) store(parsed domain.Verdict, base string) Result {
	tree, problem := r.prepareStore()
	if problem != nil {
		return *problem
	}

	sealed, err := domain.Seal(parsed, tree, base, r.opts.Now)
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
		Lines:   []string{message, "reviewer: " + sealed.Reviewer + " · " + sealed.At},
		Verdict: &sealed,
		Files:   []string{rel},
	}
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
