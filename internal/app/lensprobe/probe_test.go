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
// to measure git. Both arms get their own tree, the way two materializations do.
type fakeGit struct{ paths, diff string }

func (g fakeGit) Candidate(root, base string) (string, []string, string, error) {
	arm := "clean"
	if strings.Contains(root, "seeded") {
		arm = "seeded"
	}
	return "tree-" + arm, []string{g.paths}, g.diff, nil
}

// fakeShell records every command and answers the reviewer from the arm the
// prompt path names, so one fake can answer both arms differently.
type fakeShell struct {
	commands []string
	answer   func(command string) string
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	s.commands = append(s.commands, command)
	if strings.HasPrefix(command, "git ") {
		return "", "", 0, false, nil
	}
	return s.answer(command), "", 0, false, nil
}

func (s *fakeShell) Quote(path string) string { return path }

// fakeLenses answers the four prompt bodies with text the assertions can find.
type fakeLenses struct{}

func (fakeLenses) Lens(name string) (string, error) {
	return "---\nname: " + name + "\n---\n\n## What to look at\n\n- the " + name + " thing\n", nil
}

// verdictJSON is the seeded answer: it names the planted SQL defect the key
// knows, at the line the key names.
const verdictJSON = `{
  "reviewer": "fake-reviewer",
  "model": "fixture/reviewer",
  "inspection": {"status": "completed", "paths": ["server.go"]},
  "findings": [
    {"path": "server.go", "line": 25, "severity": "CRITICAL", "causal": true, "evidence": "cited", "why": "the id is concatenated"}
  ],
  "verdict": "changes_requested"
}`

// emptyJSON is the clean answer: a reviewer that found nothing.
const emptyJSON = `{"reviewer":"fake-reviewer","model":"fixture/reviewer","inspection":{"status":"completed","paths":["server.go"]},"findings":[],"verdict":"approved"}`

func options(exec string) lensprobe.Options {
	// Runs: 1 is the flag's own default, so a test that wants another count says
	// so; a zero here would be the refusal the flag exists for (0267).
	return lensprobe.Options{Dir: "C:\\project", Exec: exec, Model: "fixture/reviewer", Runs: 1}
}

// arm picks the canned answer from the prompt the command was handed.
func arm(clean, seeded string) func(string) string {
	return func(command string) string {
		if strings.Contains(command, "clean") {
			return clean
		}
		return seeded
	}
}

// TestProbeWithoutExecPrintsTheShippingPrompts: the bare form is the inspection
// lane. It composes the same document `review` would send, one per lens and arm,
// and starts no reviewer, so a person can read the question before paying for the
// answer.
func TestProbeWithoutExecPrintsTheShippingPrompts(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{answer: arm(emptyJSON, verdictJSON)}
	result := lensprobe.Run(options(""), newFiles(), fakeGit{paths: "server.go", diff: "+query := \"x\" + id"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	for _, want := range []string{
		"# ---- clean / risk ----",
		"# ---- seeded / risk ----",
		"# ---- seeded / resilience ----",
		"### risk",
		"+query := \"x\" + id",
		"tree-seeded",
	} {
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

// TestProbeScoresEveryLensAndTheCleanArm: with a reviewer, the seeded arm is
// scored against the key and the clean arm is measured too - a finding on code
// with no defect is the noise the lenses raise, and it is the other half of the
// promise.
func TestProbeScoresEveryLensAndTheCleanArm(t *testing.T) {
	t.Parallel()

	files := newFiles()
	shell := &fakeShell{answer: arm(emptyJSON, verdictJSON)}
	result := lensprobe.Run(options("reviewer {prompt}"), files, fakeGit{paths: "server.go", diff: "+x"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	want := "runs 1 · risk 1/1, readability 0/1, reliability 0/1, resilience 0/2 · noise 0 · clean 0"
	if result.Summary != want {
		t.Errorf("Summary = %q, want %q", result.Summary, want)
	}
	if strings.Count(strings.Join(result.Lines, "\n"), result.Summary) > 0 {
		t.Errorf("the summary is a row of the table, and the CLI prints it again:\n%s", strings.Join(result.Lines, "\n"))
	}
	var clean, seeded bool
	for path := range files.files {
		if strings.Contains(path, "reviews") {
			t.Errorf("the probe wrote into the review store: %s", path)
		}
		if strings.Contains(path, "probe") && strings.Contains(path, "-clean-") {
			clean = true
		}
		if strings.Contains(path, "probe") && strings.Contains(path, "-seeded-") {
			seeded = true
		}
	}
	if !clean || !seeded {
		t.Errorf("the probe stored %v; want both arms under the probe folder", paths(files))
	}

	// Every "stored:" line names a file the run wrote: a report that points at a
	// path nobody created sends a person to an empty folder.
	stored := 0
	for _, line := range result.Lines {
		if !strings.HasPrefix(line, "stored: ") {
			continue
		}
		stored++
		path := strings.TrimPrefix(line, "stored: ")
		if exists, _ := files.Exists(path); !exists {
			t.Errorf("the report names a path that was never written: %s", path)
		}
	}
	if stored != 8 {
		t.Errorf("the report names %d stored answers, want one per lens and arm", stored)
	}
}

// TestProbeCountsTheCleanArmsFindingsAsNoise: a reviewer that reports a defect
// where the key planted none is the number the clean arm exists for. It is
// counted and named, and it never disappears into the seeded score.
func TestProbeCountsTheCleanArmsFindingsAsNoise(t *testing.T) {
	t.Parallel()

	result := lensprobe.Run(options("reviewer {prompt}"), newFiles(), fakeGit{paths: "server.go", diff: "+x"}, &fakeShell{answer: arm(verdictJSON, emptyJSON)}, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	if !strings.HasSuffix(result.Summary, "· clean 4") {
		t.Errorf("Summary = %q, want the four clean answers counted as noise", result.Summary)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "clean") {
		t.Errorf("the report does not carry the clean arm:\n%s", strings.Join(result.Lines, "\n"))
	}
}

// TestProbeKeepsEveryRun: the fixture's tree is the same on every run, so the
// store names each answer with the first free number. A second measurement is
// what a person compares; overwriting the first one loses it.
func TestProbeKeepsEveryRun(t *testing.T) {
	t.Parallel()

	files := newFiles()
	for i := 0; i < 2; i++ {
		if result := lensprobe.Run(options("reviewer {prompt}"), files, fakeGit{paths: "server.go", diff: "+x"}, &fakeShell{answer: arm(emptyJSON, verdictJSON)}, fakeLenses{}); result.Code != lensprobe.CodeOK {
			t.Fatalf("run %d = %d: %s", i+1, result.Code, result.Summary)
		}
	}

	count := 0
	for path := range files.files {
		if strings.Contains(path, "-risk") && strings.Contains(path, "probe") {
			count++
		}
	}
	if count < 4 {
		t.Errorf("the second run overwrote the first: %d risk answers stored, want the seeded and clean answers of both runs", count)
	}
}

// TestProbeRefusesAnAnswerThatIsNotAVerdict: a reviewer that answers prose is a
// broken run, not a zero: the number would be a guess about a file the probe
// could not read.
func TestProbeRefusesAnAnswerThatIsNotAVerdict(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{answer: func(string) string { return "LGTM, looks fine to me" }}
	result := lensprobe.Run(options("reviewer {prompt}"), newFiles(), fakeGit{paths: "a.go", diff: "+x"}, shell, fakeLenses{})

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

// TestProbeRepeatsTheMeasurement: --runs asks every lens and arm once per draw,
// and the reading says how many draws it holds. The aggregation itself is pinned
// by TestSummarize, so this only proves the loop is wired.
func TestProbeRepeatsTheMeasurement(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{answer: arm(emptyJSON, verdictJSON)}
	opts := options("reviewer {prompt}")
	opts.Runs = 3
	result := lensprobe.Run(opts, newFiles(), fakeGit{paths: "server.go", diff: "+x"}, shell, fakeLenses{})

	if result.Code != lensprobe.CodeOK {
		t.Fatalf("Run = %d: %s", result.Code, result.Summary)
	}
	reviewers := 0
	for _, command := range shell.commands {
		if !strings.HasPrefix(command, "git ") {
			reviewers++
		}
	}
	if want := 3 * 8; reviewers != want {
		t.Errorf("reviewers run = %d, want %d (three draws of eight)", reviewers, want)
	}
	if !strings.HasPrefix(result.Summary, "runs 3 · ") {
		t.Errorf("Summary = %q, want it to name three draws", result.Summary)
	}
}

// TestProbeRefusesARunCountOutsideTheRange: a count the measurement cannot hold
// is refused before any reviewer runs, and the message names the bound and what
// one draw costs.
func TestProbeRefusesARunCountOutsideTheRange(t *testing.T) {
	t.Parallel()

	for _, runs := range []int{-1, 0, lensprobe.MaxRuns + 1} {
		shell := &fakeShell{answer: arm(emptyJSON, verdictJSON)}
		opts := options("reviewer {prompt}")
		opts.Runs = runs
		result := lensprobe.Run(opts, newFiles(), fakeGit{paths: "server.go", diff: "+x"}, shell, fakeLenses{})
		if result.Code != lensprobe.CodeBlocked {
			t.Fatalf("Run(--runs %d) = %d, want blocked: %s", runs, result.Code, result.Summary)
		}
		for _, want := range []string{"1", "10", "reviewer"} {
			if !strings.Contains(result.Summary, want) {
				t.Errorf("message %q does not name %q", result.Summary, want)
			}
		}
		for _, command := range shell.commands {
			if !strings.HasPrefix(command, "git ") {
				t.Errorf("--runs %d started a reviewer: %q", runs, command)
			}
		}
	}
}
