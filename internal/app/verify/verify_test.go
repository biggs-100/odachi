package verify_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/verify"
)

const (
	tree  = "0123456789abcdef0123456789abcdef01234567"
	store = "/store"
)

type fakeFiles struct {
	files  map[string]string
	events *[]string
}

func newFakeFiles() *fakeFiles { return &fakeFiles{files: map[string]string{}} }

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, errors.New("no such file: " + path)
	}
	return []byte(data), nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.files[path] = string(data)
	if f.events != nil {
		*f.events = append(*f.events, "write "+path)
	}
	return nil
}

type fakeGit struct{ tree string }

func (g fakeGit) TreeHash(string) (string, error) { return g.tree, nil }

type fakeShell struct {
	calls   int
	command string
	input   string
	stdout  string
	stderr  string
	exit    int
	stop    verify.Stop
	err     error
}

func (s *fakeShell) RunInputStopped(_, command, input string, _ int, _, _ time.Duration) (string, string, int, verify.Stop, error) {
	s.calls++
	s.command = command
	s.input = input
	return s.stdout, s.stderr, s.exit, s.stop, s.err
}

func options() verify.Options {
	return verify.Options{
		Dir:            ".",
		Store:          store,
		Exec:           "agent {prompt}",
		Requirement:    "quién escribe no se verifica a sí mismo",
		DoneWhen:       []string{"la puerta bloquea cuando no hay registro"},
		Risks:          []string{"contract"},
		TimeoutSeconds: 60,
		Now:            time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC),
	}
}

func answer(extra string) string {
	return `{"schema":1,"tree":"` + tree + `","by":"otro/modelo","verdict":"verified","notes":[]` + extra + `}`
}

// TestWithoutExecThePromptIsPrintedAndNobodyRuns: the verb composes the question
// and stops; the human reads what the verifier will be asked before paying for it.
func TestWithoutExecThePromptIsPrintedAndNobodyRuns(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{}
	opts := options()
	opts.Exec = ""

	result := verify.Run(opts, files, fakeGit{tree: tree}, shell)

	if shell.calls != 0 {
		t.Errorf("the shell ran %d times without --exec", shell.calls)
	}
	if len(files.files) != 0 {
		t.Errorf("a run with no command wrote %v", files.files)
	}
	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{opts.Requirement, opts.DoneWhen[0], "contract", tree, "verified", "refuted", "focused probes"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the prompt does not name %q:\n%s", want, joined)
		}
	}
}

// TestTheAnswerIsSealedIntoARecordNamedByTheTree: what the verifier answered is
// stored under the tree the tool read, with the tool's hour and the answer as it
// was given.
func TestTheAnswerIsSealedIntoARecordNamedByTheTree(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: answer("")}
	opts := options()

	result := verify.Run(opts, files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", result.Code, verify.CodeOK, result.Message)
	}
	want := filepath.Join(store, "verifications", tree+".json")
	stored, ok := files.files[want]
	if !ok {
		t.Fatalf("no record at %s; wrote %v", want, files.files)
	}
	for _, wantLine := range []string{`"tree": "` + tree + `"`, `"by": "otro/modelo"`, `"at": "2026-11-16T12:00:00Z"`, `"verdict": "verified"`} {
		if !strings.Contains(stored, wantLine) {
			t.Errorf("the record does not carry %s:\n%s", wantLine, stored)
		}
	}
	if shell.command != opts.Exec {
		t.Errorf("the shell ran %q, want %q", shell.command, opts.Exec)
	}
	if !strings.Contains(shell.input, opts.Requirement) {
		t.Error("the command did not receive the prompt with the requirement")
	}
}

// TestAnAnswerWithoutAnAuthorWritesNothing: the record is evidence about who read
// the tree, so an answer that names nobody is refused before it is stored.
func TestAnAnswerWithoutAnAuthorWritesNothing(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: `{"schema":1,"tree":"` + tree + `","by":"","verdict":"verified","notes":[]}`}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	if len(files.files) != 0 {
		t.Errorf("an answer without an author wrote %v", files.files)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "who verified") {
		t.Errorf("the refusal does not say what is missing: %v", result.Lines)
	}
}

// TestARefutationIsKeptAsGiven: the verifier's answer is the record; the tool
// seals it and does not overrule it.
func TestARefutationIsKeptAsGiven(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: `{"schema":1,"tree":"` + tree + `","by":"otro/modelo","verdict":"refuted","notes":["go test ./internal/app/check falls"]}`}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	want := filepath.Join(store, "verifications", tree+".json")
	if _, ok := files.files[want]; !ok {
		t.Fatalf("a refutation was not stored; wrote %v (message: %s)", files.files, result.Message)
	}
	if !strings.Contains(files.files[want], `"verdict": "refuted"`) {
		t.Errorf("the record does not keep the answer:\n%s", files.files[want])
	}
	if !strings.Contains(result.Message, "refuted") {
		t.Errorf("the line does not say what the verifier answered: %q", result.Message)
	}
}

// TestAnAnswerForAnotherTreeIsRefused: the tool read the tree, and a record that
// names another one would cover bytes nobody verified.
func TestAnAnswerForAnotherTreeIsRefused(t *testing.T) {
	files := newFakeFiles()
	other := "fedcba9876543210fedcba9876543210fedcba98"
	shell := &fakeShell{stdout: `{"schema":1,"tree":"` + other + `","by":"otro/modelo","verdict":"verified","notes":[]}`}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	if len(files.files) != 0 {
		t.Errorf("an answer for another tree wrote %v", files.files)
	}
}

// TestADeadlineStopNamesTheDeadlineAndTheFlagThatRaisesIt: the run was given a
// total budget and did not answer inside it. The message names that number, and
// the line below names the command that raises it: the human has to know which
// clock to move.
func TestADeadlineStopNamesTheDeadlineAndTheFlagThatRaisesIt(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stop: verify.StopDeadline}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{"60", "total deadline", `--timeout`, `odachi verify --exec`} {
		if !strings.Contains(joined, want) {
			t.Errorf("the deadline message does not name %q:\n%s", want, joined)
		}
	}
	if len(files.files) != 0 {
		t.Errorf("a run stopped by the deadline wrote %v", files.files)
	}
}

// TestASilenceStopNamesItsOwnBudgetAndNotTheDeadline: a silence clock fires long
// before the total deadline, so sending the human to --timeout would point at a
// flag that was not the cause. The message names the silence budget it was given
// and says the deadline was not the reason; the line below says to run the
// verifier by hand.
func TestASilenceStopNamesItsOwnBudgetAndNotTheDeadline(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stop: verify.StopSilence}
	opts := options()
	opts.FirstOutput = 90 * time.Second
	opts.Idle = 90 * time.Second

	result := verify.Run(opts, files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{"90", "silence", "not the reason", "by hand"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the silence message does not name %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "--timeout") {
		t.Errorf("the silence message sends the human to --timeout, which was not the cause:\n%s", joined)
	}
	if len(files.files) != 0 {
		t.Errorf("a run stopped by silence wrote %v", files.files)
	}
}

// TestAFailingCommandWritesNothingAndNamesTheExit: a verifier that could not run
// verified nothing, and the line says what to look at.
func TestAFailingCommandWritesNothingAndNamesTheExit(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: "", stderr: "the agent is not installed", exit: 3}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	joined := strings.Join(result.Lines, "\n")
	for _, want := range []string{"3", "the agent is not installed"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the refusal does not name %q: %s", want, joined)
		}
	}
	if len(files.files) != 0 {
		t.Errorf("a failing command wrote %v", files.files)
	}
}

// TestAnAnswerThatIsNotOneObjectIsRefused: the decoder is strict, and the refusal
// names the four keys the answer has.
func TestAnAnswerThatIsNotOneObjectIsRefused(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{stdout: "I looked at everything and it seems fine.\n"}

	result := verify.Run(options(), files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", result.Code)
	}
	if len(files.files) != 0 {
		t.Errorf("a prose answer wrote %v", files.files)
	}
}

// TestTheDeclaredRisksAreWhatAsksForTheVerifier: a feature that declared no risk
// asks for no verification, and the prompt says so instead of running anyway.
func TestWithoutDeclaredRisksThePromptIsStillPrinted(t *testing.T) {
	files := newFakeFiles()
	shell := &fakeShell{}
	opts := options()
	opts.Exec = ""
	opts.Risks = nil

	result := verify.Run(opts, files, fakeGit{tree: tree}, shell)

	if result.Code != verify.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", result.Code, verify.CodeOK, result.Message)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "none") {
		t.Errorf("the prompt does not say the commits declared no risk: %v", result.Lines)
	}
}
