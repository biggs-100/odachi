package lensprobe_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/lensprobe"
)

// fakeFiles is the probe's Files port in memory: the fixture is written here and
// read back by the assertions, so a test can see exactly what the probe wrote.
type fakeFiles struct {
	files map[string][]byte
}

func newFiles() *fakeFiles { return &fakeFiles{files: map[string][]byte{}} }

func (f *fakeFiles) Read(path string) ([]byte, error)     { return f.files[path], nil }
func (f *fakeFiles) Write(path string, data []byte) error { f.files[path] = data; return nil }
func (f *fakeFiles) Exists(path string) (bool, error)     { _, ok := f.files[path]; return ok, nil }

// fakeGit answers the candidate with fixed bytes: the probe's job is to ask, not
// to measure git.
type fakeGit struct{ tree, paths, diff string }

func (g fakeGit) Candidate(root, base string) (string, []string, string, error) {
	return g.tree, []string{g.paths}, g.diff, nil
}

// fakeShell records every command and answers the reviewer's with a canned
// verdict. The fixture's git commands get an empty answer and exit 0.
type fakeShell struct {
	commands []string
	answer   string
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	s.commands = append(s.commands, command)
	if strings.HasPrefix(command, "git ") {
		return "", "", 0, false, nil
	}
	return s.answer, "", 0, false, nil
}

func (s *fakeShell) Quote(path string) string { return path }

// fakeLenses answers the four prompt bodies with text the assertions can find.
type fakeLenses struct{}

func (fakeLenses) Lens(name string) (string, error) {
	return "---\nname: " + name + "\n---\n\n## What to look at\n\n- the " + name + " thing\n", nil
}

const verdictJSON = `{
  "reviewer": "fake-reviewer",
  "model": "fixture/reviewer",
  "inspection": {"status": "completed", "paths": ["server.go"]},
  "findings": [
    {"path": "server.go", "line": 28, "severity": "CRITICAL", "causal": true, "evidence": "cited", "why": "the id is concatenated"}
  ],
  "verdict": "changes_requested"
}`

func options(exec string) lensprobe.Options {
	return lensprobe.Options{Dir: "C:\\project", Exec: exec, Model: "fixture/reviewer"}
}

// TestProbeWithoutExecPrintsTheShippingPrompts: the bare form is the inspection
// lane. It composes the same document `review` would send -one per lens- and
// starts no reviewer, so a person can read the question before paying for the
// answer.
func TestProbeWithoutExecPrintsTheShippingPrompts(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{}
	result := lensprobe.Run(options(""), newFiles(), fakeGit{tree: "abc123", paths: "server.go", diff: "+query := \"x\" + id"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	for _, want := range []string{"# ---- risk ----", "# ---- readability ----", "# ---- reliability ----", "# ---- resilience ----", "### risk", "+query := \"x\" + id", "abc123"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the printed prompts do not carry %q", want)
		}
	}
	for _, command := range shell.commands {
		if !strings.HasPrefix(command, "git ") {
			t.Errorf("the bare form ran a reviewer: %q", command)
		}
	}
}

// TestProbeScoresEveryLensAgainstTheKey: with a reviewer, each lens answers and
// the report counts what the key planted. A defect found by one lens is that
// lens's catch, and the same finding in another lens's answer is not noise.
func TestProbeScoresEveryLensAgainstTheKey(t *testing.T) {
	t.Parallel()

	files := newFiles()
	shell := &fakeShell{answer: verdictJSON}
	result := lensprobe.Run(options("reviewer {prompt}"), files, fakeGit{tree: "abc123", paths: "server.go", diff: "+x"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	want := "risk 1/1, readability 0/1, reliability 0/1, resilience 0/2 · noise 0"
	if result.Summary != want {
		t.Errorf("Summary = %q, want %q", result.Summary, want)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "noise 0") {
		t.Errorf("the table does not carry the summary:\n%s", strings.Join(result.Lines, "\n"))
	}
	// The answer is stored where `check` cannot read it as a verdict for a tree.
	found := false
	for path := range files.files {
		if strings.Contains(path, "probe") && strings.Contains(path, "risk") {
			found = true
		}
		if strings.Contains(path, "reviews") {
			t.Errorf("the probe wrote into the review store: %s", path)
		}
	}
	if !found {
		t.Errorf("the probe stored no answer under the probe folder: %v", paths(files))
	}
}

// TestProbeRefusesAnAnswerThatIsNotAVerdict: a reviewer that answers prose is a
// broken run, not a zero: the number would be a guess about a file the probe
// could not read.
func TestProbeRefusesAnAnswerThatIsNotAVerdict(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{answer: "LGTM, looks fine to me"}
	result := lensprobe.Run(options("reviewer {prompt}"), newFiles(), fakeGit{tree: "abc", paths: "a.go", diff: "+x"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeBlocked {
		t.Fatalf("Run = %d, want blocked: %s", result.Code, result.Summary)
	}
	if !strings.Contains(result.Summary, "not a verdict") {
		t.Errorf("Summary = %q, want it to say the answer is not a verdict", result.Summary)
	}
}

func paths(f *fakeFiles) []string {
	out := make([]string, 0, len(f.files))
	for path := range f.files {
		out = append(out, path)
	}
	return out
}
