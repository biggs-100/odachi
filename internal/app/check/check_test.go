package check_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/check"
	"github.com/biggs-100/odachi/internal/domain/gate"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/proof"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// dir is the project root the run sees. It is absolute-looking and never
// touched: every port is a double.
const dir = "/proyecto"

// store is the clone's state, inside the project it describes and outside the
// worktree: the real one is <proyecto>/.git/odachi.
var store = filepath.Join(dir, ".git", "odachi")

// The tree hash the fake git always reports, and the one the seeded proof
// files carry when they are supposed to be current.
const (
	currentTree = "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"
	otherTree   = "9f2b7c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b"
)

// fakeFiles is the Files port in memory. It records every write and backup so a
// test can assert that check writes nothing except the proof --run creates.
type fakeFiles struct {
	files   map[string][]byte
	writes  []string
	removes []string
	backups []string
	err     error

	// denyWrite refuses the writes its predicate names, leaving the rest of the
	// port healthy: a test can pin one write failing without a broken filesystem
	// standing in for a permission problem.
	denyWrite func(path string) error
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		// A path already under the project root (the clone's store lives there)
		// is kept as it is; the rest is project-relative.
		key := filepath.Join(dir, filepath.FromSlash(path))
		if slashed, root := filepath.ToSlash(filepath.FromSlash(path)), filepath.ToSlash(dir)+"/"; strings.HasPrefix(slashed, root) {
			key = filepath.FromSlash(path)
		}
		f.files[key] = []byte(content)
	}
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) List(path string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	var names []string
	for file := range f.files {
		if filepath.Dir(file) == path {
			names = append(names, filepath.Base(file))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Remove(path string) error {
	if f.err != nil {
		return f.err
	}
	f.removes = append(f.removes, path)
	delete(f.files, path)
	return nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	if f.err != nil {
		return f.err
	}
	if f.denyWrite != nil {
		if err := f.denyWrite(path); err != nil {
			return err
		}
	}
	f.writes = append(f.writes, path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	f.backups = append(f.backups, path)
	return path + ".bak-prueba", nil
}

// fakeGit is the Git port: it reports fixed facts and records which base the
// fence asked about.
type fakeGit struct {
	tree       string
	changed    []string
	deps       []string
	changedErr error
	bases      []string

	// ignored is what the repository's rules answer about the evidence, and
	// ignoredErr is the probe failing. A failed probe must never be read as
	// "the tree is already protected".
	ignored    bool
	ignoredErr error
	ignoredAs  []string

	// blobs answers BlobHash per path, and blobErr is git failing for real.
	blobs   map[string]string
	blobErr error

	// diff is what the port answers for the candidate, and diffErr is the diff
	// failing for real.
	diff    string
	diffErr error

	// candidateCalls counts the one read of the worktree this run asked for.
	candidateCalls int

	// filesAt answers FileAt, keyed by "rev:path"; fileAtErr is the ref failing.
	filesAt   map[string]string
	fileAtErr error

	// between answers Between: the paths that moved between the proof's tree and
	// this one, and betweenErr is git failing for real. betweenCalls counts the
	// question, because a proof that is not stale must not ask it.
	between      []string
	betweenErr   error
	betweenCalls int

	// commits says which spellings name the same commit. A spelling that is not
	// here resolves to itself, so a test declares only the pair it cares about.
	commits map[string]string
}

func (g *fakeGit) TreeHash(root string) (string, error) { return g.tree, nil }

// Resolve names the commit a spelling points at. The fake has no repository, so a
// spelling resolves to itself unless the test declared that two spellings are one
// commit: the pair a gate that compares commits must tell apart from two commits.
func (g *fakeGit) Resolve(root, ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", errors.New("no ref to resolve")
	}
	if resolved, ok := g.commits[trimmed]; ok {
		return resolved, nil
	}
	return trimmed, nil
}

// Candidate is the one read: it records every call, because "the fence, the
// lens class and the proof tree describe the same bytes" is the property.
func (g *fakeGit) Candidate(root, base string) (string, []string, string, error) {
	g.candidateCalls++
	g.bases = append(g.bases, base)
	if g.changedErr != nil && base != check.EmptyTree {
		return "", nil, "", g.changedErr
	}
	if g.diffErr != nil {
		return "", nil, "", g.diffErr
	}
	return g.tree, g.changed, g.diff, nil
}

func (g *fakeGit) Between(root, from, to string) ([]string, error) {
	g.betweenCalls++
	if g.betweenErr != nil {
		return nil, g.betweenErr
	}
	return g.between, nil
}

func (g *fakeGit) BlobHash(root, path string) (string, error) {
	if g.blobErr != nil {
		return "", g.blobErr
	}
	return g.blobs[path], nil
}

func (g *fakeGit) AddedDeps(root, base string) ([]string, error) { return g.deps, nil }

// FileAt answers the fence of a ref, or the reason the ref could not be read.
func (g *fakeGit) FileAt(root, rev, path string) ([]byte, bool, error) {
	if g.fileAtErr != nil {
		return nil, false, g.fileAtErr
	}
	data, ok := g.filesAt[rev+":"+path]
	if !ok {
		return nil, false, nil
	}
	return []byte(data), true, nil
}

func (g *fakeGit) Ignored(root, path string) (bool, error) {
	g.ignoredAs = append(g.ignoredAs, path)
	if g.ignoredErr != nil {
		return false, g.ignoredErr
	}
	return g.ignored, nil
}

// fakeShell is the Shell port: it reports the run it was asked for.
type fakeShell struct {
	output   string
	stderr   string
	exitCode int
	timedOut bool
	err      error

	command  string
	timeout  int
	wasAsked bool
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	s.wasAsked = true
	s.command = command
	s.timeout = timeoutSeconds
	return s.output, s.stderr, s.exitCode, s.timedOut, s.err
}

// completeBrief is the smallest brief that passes brief.Parse and Missing.
const completeBrief = `# turnos

## Qué tiene que existir
- reservar turnos

## Listo cuando
- reservo un turno y aparece en la agenda

## NO incluye (primera versión)
- pagos
`

// featureDoc is a task document with the given history lines.
func featureDoc(history ...string) string {
	return taskdoc.TaskDoc{
		Title:        "turnos",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}},
		DoneWhen:     []string{"reservo un turno"},
		History:      history,
	}.Render()
}

// proofFile is a proof produced from tree with output.
func proofFile(tree, command, output string) string {
	return proof.Render(proof.Proof{Tree: tree, Command: command, Output: output})
}

// readyFiles is the base every blocked case breaks in exactly one place.
func readyFiles() map[string]string {
	return map[string]string{
		"brief.md":                 completeBrief,
		"odd/tasks/0007-turnos.md": featureDoc(`L1 · "reservar turnos"`),
		"odd/fence.toml":           "",
		"proof/0007-turnos.txt":    proofFile(currentTree, "go test ./...", "ok\n"),
		"odd/profile.toml":         `name = "novato"`,
	}
}

// run evaluates the current state: no --run, no --proof.
func run(t *testing.T, files *fakeFiles, git *fakeGit, opts check.Options) check.Result {
	t.Helper()
	opts.Dir = dir
	opts.Store = store
	if opts.Base == "" {
		opts.Base = "HEAD"
	}
	if opts.TimeoutSeconds == 0 {
		opts.TimeoutSeconds = 120
	}
	return check.Run(opts, files, git, &fakeShell{})
}

// TestTheBlockNamesTheOpenQuestions: the gate counts them; the task doc knows the
// words. A block that says "1 unanswered question" and hands a bare `ask` sends the
// human to read the file to discover the question and to guess the command that
// answers it.
func TestTheBlockNamesTheOpenQuestions(t *testing.T) {
	files := readyFiles()
	files["odd/tasks/0007-turnos.md"] = featureDoc(
		`L1 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`,
		`L2 · PREGUNTA P2 · "¿Quién puede cancelar?"`,
	)

	got := run(t, newFakeFiles(files), &fakeGit{tree: currentTree}, check.Options{})

	lines := strings.Join(got.Lines, "\n")
	for _, want := range []string{
		"2 unanswered questions",
		`P1 · "¿Los turnos se pueden cancelar?"`,
		`P2 · "¿Quién puede cancelar?"`,
		`next: odachi ask --answer "<answer>" --for P1`,
	} {
		if !strings.Contains(lines, want) {
			t.Errorf("the block does not say %q:\n%s", want, lines)
		}
	}
	if got.Next != `odachi ask --answer "<answer>" --for P1` {
		t.Errorf("Next = %q, want the form that answers the first question", got.Next)
	}
}

// TestTheEightResults is the whole gate table through the app: ready, no brief,
// brief incomplete, out of scope, open question, missing proof, stale proof and
// failed proof.
func TestTheEightResults(t *testing.T) {
	outOfScope := readyFiles()
	outOfScope["odd/fence.toml"] = `[[rule]]
not   = "pagos"
paths = ["payment*", "**/payments/**"]
deps  = ["stripe"]
`

	openQuestion := readyFiles()
	openQuestion["odd/tasks/0007-turnos.md"] = featureDoc(`L1 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`)

	staleProof := readyFiles()
	staleProof["proof/0007-turnos.txt"] = proofFile(otherTree, "go test ./...", "ok\n")

	missingProof := readyFiles()
	delete(missingProof, "proof/0007-turnos.txt")

	noBrief := readyFiles()
	delete(noBrief, "brief.md")

	incompleteBrief := readyFiles()
	incompleteBrief["brief.md"] = "# turnos\n\n## Qué tiene que existir\n- reservar turnos\n"

	tests := []struct {
		name    string
		files   map[string]string
		git     *fakeGit
		opts    check.Options
		shell   *fakeShell
		code    int
		reason  string
		message string
		next    string
	}{
		{
			name:    "ready",
			files:   readyFiles(),
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.Ready),
			reason:  gate.ReasonReady,
			message: "ready: there is proof for this tree",
		},
		{
			name:    "no brief",
			files:   noBrief,
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.NeedsDecision),
			reason:  gate.ReasonNoBrief,
			message: "there is no brief: we do not know yet what must exist",
			next:    "odachi start",
		},
		{
			name:    "brief incomplete",
			files:   incompleteBrief,
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.NeedsDecision),
			reason:  gate.ReasonBriefIncomplete,
			message: "the brief is incomplete: missing done when",
			next:    "odachi start",
		},
		{
			name:    "out of scope",
			files:   outOfScope,
			git:     &fakeGit{tree: currentTree, changed: []string{"internal/payments/stripe.go"}},
			code:    int(gate.NeedsDecision),
			reason:  gate.ReasonOutOfScope,
			message: "out of scope: pagos: internal/payments/stripe.go",
			next:    "odachi ask \"<your decision>\"",
		},
		{
			name:    "open question",
			files:   openQuestion,
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.NeedsDecision),
			reason:  gate.ReasonOpenQuestion,
			message: "1 unanswered question",
			next:    `odachi ask --answer "<answer>" --for P1`,
		},
		{
			name:    "missing proof",
			files:   missingProof,
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.NeedsWork),
			reason:  gate.ReasonProofMissing,
			message: "the proof for this change is missing",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "stale proof",
			files:   staleProof,
			git:     &fakeGit{tree: currentTree},
			code:    int(gate.NeedsWork),
			reason:  gate.ReasonProofStale,
			message: "the proof is from another state of the code",
			next:    "odachi check --proof <file>",
		},
		{
			name:    "failed proof",
			files:   readyFiles(),
			git:     &fakeGit{tree: currentTree},
			opts:    check.Options{Run: "go test ./..."},
			shell:   &fakeShell{output: "ok\nFAIL github.com/biggs-100/odachi 0.02s\n", exitCode: 1},
			code:    int(gate.NeedsWork),
			reason:  gate.ReasonProofFailed,
			message: "the proof failed: FAIL github.com/biggs-100/odachi 0.02s",
			next:    `odachi check --run "go test ./..."`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := newFakeFiles(tc.files)
			shell := tc.shell
			if shell == nil {
				shell = &fakeShell{}
			}
			opts := tc.opts
			opts.Dir = dir
			if opts.Base == "" {
				opts.Base = "HEAD"
			}
			if opts.TimeoutSeconds == 0 {
				opts.TimeoutSeconds = 120
			}

			got := check.Run(opts, files, tc.git, shell)

			if got.Code != tc.code {
				t.Errorf("Code = %d, want %d (reason %q, message %q)", got.Code, tc.code, got.Reason, got.Message)
			}
			if got.Reason != tc.reason {
				t.Errorf("Reason = %q, want %q", got.Reason, tc.reason)
			}
			if got.Message != tc.message {
				t.Errorf("Message = %q, want %q", got.Message, tc.message)
			}
			if got.Next != tc.next {
				t.Errorf("Next = %q, want %q", got.Next, tc.next)
			}
			if strings.Contains(got.Message, "\n") {
				t.Errorf("Message must be one line, got %q", got.Message)
			}
			if got.Code != int(gate.Ready) && got.Next == "" {
				t.Errorf("blocked result %q has no Next", tc.name)
			}
		})
	}
}

// check writes nothing: not the brief, not the fence, not the profile. The only
// write in the whole command is the proof --run creates.
func TestCheckWritesNothingWithoutRun(t *testing.T) {
	files := newFakeFiles(readyFiles())

	check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120}, files, &fakeGit{tree: currentTree}, &fakeShell{})

	if len(files.writes) != 0 {
		t.Errorf("check wrote %v, want no writes", files.writes)
	}
	if len(files.backups) != 0 {
		t.Errorf("check backed up %v, want no backups", files.backups)
	}
}

// --run captures the command and writes the proof of the current feature, with
// the tree and the command in the header and the output verbatim.
func TestRunWritesTheProofOfTheCurrentFeature(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	git := &fakeGit{tree: currentTree}
	shell := &fakeShell{output: "ok  github.com/biggs-100/odachi 0.02s\n"}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45}, files, git, shell)

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (reason %q, message %q)", got.Code, got.Reason, got.Message)
	}
	if !shell.wasAsked {
		t.Fatal("the shell was never asked to run the command")
	}
	if shell.command != "go test ./..." {
		t.Errorf("ran %q, want the --run command", shell.command)
	}
	if shell.timeout != 45 {
		t.Errorf("timeout = %d seconds, want 45", shell.timeout)
	}

	wantPath := filepath.Join(dir, "proof", "0007-turnos.txt")
	if !slices.Contains(files.writes, wantPath) {
		t.Fatalf("writes = %v, want %s", files.writes, wantPath)
	}
	for _, write := range files.writes {
		if !strings.HasPrefix(write, filepath.Join(dir, "proof")+string(filepath.Separator)) {
			t.Errorf("check wrote %s, want nothing outside proof/", write)
		}
	}

	want := proofFile(currentTree, "go test ./...", "ok  github.com/biggs-100/odachi 0.02s\n")
	if string(files.files[wantPath]) != want {
		t.Errorf("proof =\n%q\nwant\n%q", files.files[wantPath], want)
	}
	if !strings.Contains(got.Message, "ready") {
		t.Errorf("Message = %q, want the ready sentence", got.Message)
	}
}

// TestTheProofKeepsBothStreams: a proof is what the command said. The test
// verdict goes to stdout and build errors go to stderr; hiding either would
// make the proof a partial quotation.
func TestTheProofKeepsBothStreams(t *testing.T) {
	files := newFakeFiles(readyFiles())
	git := &fakeGit{tree: currentTree}
	shell := &fakeShell{output: "ok  github.com/biggs-100/odachi 0.02s\n", stderr: "warning: one test was skipped\n"}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "go test ./...", Base: "HEAD", TimeoutSeconds: 45}, files, git, shell)
	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (%s)", got.Code, got.Message)
	}

	wantPath := filepath.Join(dir, "proof", "0007-turnos.txt")
	want := proofFile(currentTree, "go test ./...", "ok  github.com/biggs-100/odachi 0.02s\nwarning: one test was skipped\n")
	if string(files.files[wantPath]) != want {
		t.Errorf("proof =\n%q\nwant\n%q", files.files[wantPath], want)
	}
}

// The proof says which session produced it. It is the only way a later gate can
// tell a verdict by another model from one the same model wrote about itself.
func TestTheProofNamesTheSessionThatRanIt(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120,
		Model: "provider/model#variant"}, files, &fakeGit{tree: currentTree}, &fakeShell{output: "ok\n"})

	body := string(files.files[filepath.Join(dir, "proof", "0007-turnos.txt")])
	parsed, err := proof.Parse(body)
	if err != nil {
		t.Fatalf("the written proof does not parse: %v", err)
	}
	if parsed.Model != "provider/model#variant" {
		t.Errorf("Model = %q, want the session that ran it", parsed.Model)
	}
	if parsed.Output != "ok\n" {
		t.Errorf("Output = %q, want the verbatim output", parsed.Output)
	}
}

// --run keeps the evidence out of the tree it describes, so two runs on the
// same source tree hash the same worktree.
func TestRunWritesTheProofIgnoreFile(t *testing.T) {
	files := newFakeFiles(readyFiles())

	check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "ok\n"})

	ignore := filepath.Join(dir, "proof", ".gitignore")
	content, ok := files.files[ignore]
	if !ok {
		t.Fatalf("proof/.gitignore was not written; writes: %v", files.writes)
	}
	if !strings.Contains(string(content), "*") {
		t.Errorf("proof/.gitignore = %q, want the proof files ignored", content)
	}

	// A second run finds it and does not write it again.
	files.writes = nil
	check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "ok\n"})
	if slices.Contains(files.writes, ignore) {
		t.Errorf("a rerun rewrote %s, want it written once", ignore)
	}
}

// Two runs on the same source tree produce the same proof bytes: no timestamp,
// no absolute path, nothing that changes between runs.
func TestTwoRunsOnTheSameTreeProduceTheSameProof(t *testing.T) {
	first := newFakeFiles(readyFiles())
	delete(first.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	second := newFakeFiles(readyFiles())
	delete(second.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	shell := &fakeShell{output: "ok\n"}
	opts := check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120}

	check.Run(opts, first, &fakeGit{tree: currentTree}, shell)
	check.Run(opts, second, &fakeGit{tree: currentTree}, shell)

	path := filepath.Join(dir, "proof", "0007-turnos.txt")
	if string(first.files[path]) != string(second.files[path]) {
		t.Errorf("two runs produced different proofs:\n%q\nvs\n%q", first.files[path], second.files[path])
	}
}

// A failing command is a failed proof: exit 1, the last line the command said,
// and the same command as the continuation. Nothing is written as a proof.
func TestFailingRunDoesNotWriteAProof(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "exit 1", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "boom\nla que importa\n\n", exitCode: 1})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Code/Reason = %d/%q, want 1/%q", got.Code, got.Reason, gate.ReasonProofFailed)
	}
	if got.Message != "the proof failed: la que importa" {
		t.Errorf("Message = %q, want the last non-empty output line", got.Message)
	}
	if got.Next != `odachi check --run "exit 1"` {
		t.Errorf("Next = %q, want the same command", got.Next)
	}
	for _, write := range files.writes {
		if strings.HasSuffix(write, "0007-turnos.txt") {
			t.Errorf("a failed run wrote the proof %s: a failed command is not evidence", write)
		}
	}
}

func TestTimedOutRunSaysHowLongItWaited(t *testing.T) {
	files := newFakeFiles(readyFiles())

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "sleep 60", Base: "HEAD", TimeoutSeconds: 7},
		files, &fakeGit{tree: currentTree}, &fakeShell{timedOut: true, exitCode: -1})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Code/Reason = %d/%q, want 1/%q", got.Code, got.Reason, gate.ReasonProofFailed)
	}
	if got.Message != "the proof did not finish in 7s" {
		t.Errorf("Message = %q, want %q", got.Message, "the proof did not finish in 7s")
	}
	if got.Next != `odachi check --run "sleep 60" --timeout 14` {
		t.Errorf("Next = %q, want the same command with twice the wait", got.Next)
	}
	for _, write := range files.writes {
		if strings.HasSuffix(write, "0007-turnos.txt") {
			t.Errorf("a timed-out run wrote the proof %s", write)
		}
	}
}

// TestFailingRunKeepsTheWholeOutput: the message is one line because a terminal
// is one line; the output goes to proof/<feature>.failed.txt so the failing
// test, the assertion and the build error are still there when the line is not
// enough.
func TestFailingRunKeepsTheWholeOutput(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "exit 1", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree},
		&fakeShell{output: "--- FAIL: TestX\nboom\nla que importa\n", stderr: "detalle del build\n", exitCode: 1})

	rel := "proof/0007-turnos.failed.txt"
	body, ok := files.files[filepath.Join(dir, "proof", "0007-turnos.failed.txt")]
	if !ok {
		t.Fatalf("the failed run kept no output; writes: %v", files.writes)
	}
	text := string(body)
	for _, want := range []string{"# failed " + currentTree, "# command exit 1", "# exit 1", "--- FAIL: TestX", "boom", "la que importa", "detalle del build"} {
		if !strings.Contains(text, want) {
			t.Errorf("the failed output does not carry %q:\n%s", want, text)
		}
	}
	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, "output: "+rel) {
		t.Errorf("the gate does not name the failed output:\n%s", lines)
	}
	if strings.Contains(lines, "proof/0007-turnos.txt") {
		t.Errorf("the gate names a green proof that does not exist:\n%s", lines)
	}
}

// TestFailedRunWithoutItsEvidenceNamesNoPath: keeping the output is best effort.
// When the file cannot be written the gate still fails with its one line and its
// next command, and it names no path: an "output:" line would send the human to
// a file that is not there.
func TestFailedRunWithoutItsEvidenceNamesNoPath(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	files.denyWrite = func(path string) error {
		if strings.HasSuffix(path, ".failed.txt") {
			return errors.New("disco lleno")
		}
		return nil
	}

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "exit 1", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree},
		&fakeShell{output: "--- FAIL: TestX\nla que importa\n", exitCode: 1})

	if got.Code != int(gate.NeedsWork) || got.Reason != gate.ReasonProofFailed {
		t.Fatalf("Code/Reason = %d/%q, want %d/%q (the run failed either way)",
			got.Code, got.Reason, int(gate.NeedsWork), gate.ReasonProofFailed)
	}
	if !strings.Contains(got.Message, "la que importa") {
		t.Errorf("Message = %q, want the last line of the output: it is the only evidence left", got.Message)
	}
	lines := strings.Join(got.Lines, "\n")
	if strings.Contains(lines, "output:") {
		t.Errorf("the gate names evidence it could not write:\n%s", lines)
	}
	if !strings.Contains(lines, "next: ") {
		t.Errorf("the gate lost the command that resolves it:\n%s", lines)
	}
	for _, write := range files.writes {
		if strings.HasSuffix(write, ".failed.txt") {
			t.Errorf("a refused write was recorded as written: %s", write)
		}
	}
}

func TestTimedOutRunKeepsTheOutputToo(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "sleep 60", Base: "HEAD", TimeoutSeconds: 7},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "starting\n", timedOut: true, exitCode: -1})

	body := string(files.files[filepath.Join(dir, "proof", "0007-turnos.failed.txt")])
	if !strings.Contains(body, "# timed out") || !strings.Contains(body, "starting") {
		t.Errorf("the timed-out output does not say it timed out or lost the output:\n%s", body)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "proof/0007-turnos.failed.txt") {
		t.Errorf("the gate does not name the timed-out output:\n%s", strings.Join(got.Lines, "\n"))
	}
}

func TestGreenRunRemovesTheStaleFailedFile(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	stale := filepath.Join(dir, "proof", "0007-turnos.failed.txt")
	files.files[stale] = []byte("# failed una-otra\n")

	check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "ok\n"})

	if _, still := files.files[stale]; still {
		t.Errorf("a green run left the stale failed file %s: proof/ must not show a failure beside the proof of the tree that passed", stale)
	}
	for _, write := range files.writes {
		if strings.HasSuffix(write, ".failed.txt") {
			t.Errorf("a green run wrote a failed file: %s", write)
		}
	}
}

// TestProofFlagRefusesTheFailedEvidence: --proof takes evidence already written,
// and the failed file is evidence of the opposite. The refusal names it instead
// of saying "no tree header", because the difference is the whole point.
func TestProofFlagRefusesTheFailedEvidence(t *testing.T) {
	files := newFakeFiles(readyFiles())
	failed := filepath.Join(dir, "manual.failed.txt")
	files.files[failed] = []byte(proof.RenderFailed(proof.Proof{
		Tree:    currentTree,
		Command: "go test ./...",
		Output:  "--- FAIL: TestX\nFAIL\n",
	}, 1, false))

	got := check.Run(check.Options{Dir: dir, Store: store, Proof: "manual.failed.txt", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.NeedsWork) {
		t.Fatalf("Code = %d, want %d (a refusal the agent can act on)", got.Code, gate.NeedsWork)
	}
	if !strings.Contains(got.Message, "failed output") {
		t.Errorf("Message = %q, want it to say the file is the failed output of a run", got.Message)
	}
}

// --proof takes a proof an agent or CI already wrote, and refuses one that does
// not say which tree produced it.
func TestProofFlagRefusesAFileWithoutTheHeader(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, "manual.txt")] = []byte("ok\n")

	got := check.Run(check.Options{Dir: dir, Store: store, Proof: "manual.txt", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.NeedsWork) {
		t.Fatalf("Code = %d, want %d (a refusal the agent can act on)", got.Code, gate.NeedsWork)
	}
	if !strings.Contains(got.Message, "--run") {
		t.Errorf("Message = %q, want it to name --run", got.Message)
	}
	if got.Next == "" || !strings.HasPrefix(got.Next, "odachi ") {
		t.Errorf("Next = %q, want a command that resolves it", got.Next)
	}
}

func TestProofFlagUsesTheGivenProof(t *testing.T) {
	files := newFakeFiles(readyFiles())
	delete(files.files, filepath.Join(dir, "proof", "0007-turnos.txt"))
	files.files[filepath.Join(dir, "otra.txt")] = []byte(proofFile(currentTree, "go test ./...", "ok\n"))

	got := check.Run(check.Options{Dir: dir, Store: store, Proof: "otra.txt", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (reason %q, message %q)", got.Code, got.Reason, got.Message)
	}
	if len(files.writes) != 0 {
		t.Errorf("check wrote %v with --proof, want no writes", files.writes)
	}
}

// No task document is not an error: a project can be verified before it has a
// feature. Open questions are zero and the proof lives at a stable path.
func TestNoTaskDocumentMeansZeroOpenQuestions(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"brief.md": completeBrief,
	})

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Reason != gate.ReasonProofMissing {
		t.Fatalf("Reason = %q, want %q: with no feature there is nothing to have asked", got.Reason, gate.ReasonProofMissing)
	}
}

func TestRunWithoutATaskDocumentStillWritesAProof(t *testing.T) {
	files := newFakeFiles(map[string]string{"brief.md": completeBrief})

	got := check.Run(check.Options{Dir: dir, Store: store, Run: "echo ok", Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{output: "ok\n"})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (reason %q, message %q)", got.Code, got.Reason, got.Message)
	}
	want := filepath.Join(dir, "proof", "0000-check.txt")
	if !slices.Contains(files.writes, want) {
		t.Errorf("writes = %v, want %s", files.writes, want)
	}
}

// --base "" must not fail: it means "compare against nothing", so every path in
// the worktree counts as new and the fence is still checked.
func TestEmptyBaseComparesAgainstTheEmptyTree(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash(layout.Fence))] = []byte(`[[rule]]
not   = "pagos"
paths = ["**/payments/**"]
`)
	git := &fakeGit{tree: currentTree, changed: []string{"internal/payments/stripe.go"}}

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "", TimeoutSeconds: 120}, files, git, &fakeShell{})

	if got.Reason != gate.ReasonOutOfScope {
		t.Fatalf("Reason = %q, want %q (base \"\" must still run the fence)", got.Reason, gate.ReasonOutOfScope)
	}
	if len(git.bases) == 0 || git.bases[0] != check.EmptyTree {
		t.Errorf("bases asked = %v, want %q", git.bases, check.EmptyTree)
	}
}

// A repository with no commit has no HEAD. That is not a failure: everything
// uncommitted is what the fence compares against.
func TestNoCommitFallsBackToTheEmptyTree(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash(layout.Fence))] = []byte(`[[rule]]
not   = "pagos"
paths = ["**/payments/**"]
`)
	git := &fakeGit{tree: currentTree, changedErr: fmt.Errorf("git diff: exit status 128: unknown revision HEAD")}

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120}, files, git, &fakeShell{})

	if got.Code != int(gate.Ready) {
		t.Fatalf("Code = %d, want 0 (reason %q, message %q)", got.Code, got.Reason, got.Message)
	}
	if !slices.Contains(git.bases, check.EmptyTree) {
		t.Errorf("bases asked = %v, want a fallback to %q", git.bases, check.EmptyTree)
	}
}

// A base the human named and git cannot resolve is a block that names --base,
// not a silent fallback.
func TestUnknownBaseBlocksAndNamesTheFlag(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash(layout.Fence))] = []byte(`[[rule]]
not   = "pagos"
paths = ["**/payments/**"]
`)
	git := &fakeGit{tree: currentTree, changedErr: fmt.Errorf("git diff: exit status 128: unknown revision origin/main")}

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "origin/main", TimeoutSeconds: 120}, files, git, &fakeShell{})

	if got.Code != int(gate.NeedsWork) {
		t.Fatalf("Code = %d, want %d", got.Code, gate.NeedsWork)
	}
	if !strings.Contains(got.Message, "--base") {
		t.Errorf("Message = %q, want it to name --base", got.Message)
	}
	if !strings.Contains(got.Message, "origin/main") {
		t.Errorf("Message = %q, want it to name the base that failed", got.Message)
	}
}

// A brief that exists but cannot be parsed still blocks as incomplete, with the
// parts it is missing, instead of crashing.
func TestUnreadableFenceBlocksWithAHumanSentence(t *testing.T) {
	files := newFakeFiles(readyFiles())
	files.files[filepath.Join(dir, filepath.FromSlash(layout.Fence))] = []byte("[[rule]\nnot = \n")

	got := check.Run(check.Options{Dir: dir, Store: store, Base: "HEAD", TimeoutSeconds: 120},
		files, &fakeGit{tree: currentTree}, &fakeShell{})

	if got.Code != int(gate.NeedsWork) {
		t.Fatalf("Code = %d, want %d", got.Code, gate.NeedsWork)
	}
	if !strings.Contains(got.Message, "fence.toml") {
		t.Errorf("Message = %q, want it to name the file", got.Message)
	}
}
