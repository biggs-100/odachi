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

	"github.com/biggz/odachi/internal/app/tasks"
	"github.com/biggz/odachi/internal/domain/brief"
	"github.com/biggz/odachi/internal/domain/fence"
	"github.com/biggz/odachi/internal/domain/gate"
	"github.com/biggz/odachi/internal/domain/proof"
)

// DefaultTimeout is how long the proof command may take when --timeout is not
// given, in seconds. A gate that can hang is a gate nobody trusts.
const DefaultTimeout = 120

// EmptyTree is git's well-known hash of the empty tree. Comparing against it is
// what "sin commits: todo lo no commiteado" means: every file is new because
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
}

// Files is the disk port: enough to read the project and write one proof.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
	Write(path string, data []byte) error
}

// Git is the repository port: which tree is here, what changed since a base,
// and which dependencies appeared.
type Git interface {
	TreeHash(root string) (string, error)
	ChangedPaths(root, base string) ([]string, error)
	AddedDeps(root, base string) ([]string, error)
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
			"no pude leer el árbol de git: %s; corré odachi check dentro de un repositorio git y revisá que git esté instalado",
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

	return r.report(gate.Evaluate(state), haveProof)
}

// proofIgnore keeps the evidence out of the tree the evidence describes:
// without it, writing the proof changes the worktree, the next run computes a
// different tree hash and every proof invalidates the one before it. The ignore
// file itself is versioned, so the rule that keeps proofs out of git lives with
// the project instead of inside the tool.
const proofIgnore = "# odachi: las pruebas se generan, no se versionan\n*\n!.gitignore\n"

// ensureProofDir creates proof/.gitignore before the tree is hashed, so the
// first run and every later run hash the same worktree. Only --run writes it:
// a plain check never touches the disk.
func (r *runner) ensureProofDir() *Result {
	path := filepath.Join(r.opts.Dir, proofDir, ".gitignore")

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude revisar %s/.gitignore: %s; revisá los permisos y volvé a correr odachi check --run %q",
			proofDir, oneLine(err.Error()), r.opts.Run))
		return &problem
	}
	if exists {
		return nil
	}

	if err := r.files.Write(path, []byte(proofIgnore)); err != nil {
		problem := blocked(fmt.Sprintf("no pude preparar %s/: %s; revisá los permisos y volvé a correr odachi check --run %q",
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
		problem := blocked(fmt.Sprintf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi check", briefFile, oneLine(err.Error())))
		return false, nil, &problem
	}
	if !exists {
		return false, nil, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("no pude leer %s: %s; revisá los permisos y volvé a correr odachi check", briefFile, oneLine(err.Error())))
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
		problem := blocked(oneLine(err.Error()) + "; revisá los permisos de " + tasks.Dir + " y volvé a correr odachi check")
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
		problem := blocked(fmt.Sprintf("no pude revisar la prueba %s: %s; revisá los permisos y volvé a correr odachi check", r.proofRel, oneLine(err.Error())))
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
		problem := refused(shown, "no pude leerla: "+oneLine(err.Error()))
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
			Message: "no pude correr la prueba: " + oneLine(err.Error()),
			Next:    runCommand(r.opts.Run),
			Lines:   []string{"no pude correr la prueba: " + oneLine(err.Error()), "siguiente: " + runCommand(r.opts.Run)},
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
		problem := blocked(fmt.Sprintf("no pude escribir la prueba %s: %s; revisá los permisos y volvé a correr odachi check --run %q",
			r.proofRel, oneLine(err.Error()), r.opts.Run))
		return ranProof{}, &problem
	}
	return ranProof{output: output}, nil
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
		result.Lines = []string{verdict.Message, "siguiente: " + verdict.Next}
		return result
	}

	result.Lines = []string{verdict.Message}
	if haveProof && r.proofRel != "" {
		result.Proof = r.proofRel
		result.Lines = append(result.Lines, "prueba: "+r.proofRel)
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
	return fmt.Sprintf("no pude comparar el cerco contra %q: %s; corregí --base %q o sacá el flag para comparar contra HEAD",
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
	message := fmt.Sprintf(`la prueba %s no sirve: %s; corré odachi check --run "<comando>" para que la escriba la herramienta`, shown, reason)
	return Result{
		Code:    int(gate.NeedsWork),
		Reason:  gate.ReasonProofMissing,
		Message: message,
		Next:    `odachi check --run "<comando>"`,
		Lines:   []string{message, `siguiente: odachi check --run "<comando>"`},
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
		Lines:   []string{message, "siguiente: odachi check"},
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
