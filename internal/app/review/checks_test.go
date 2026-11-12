package review_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
	storeentry "github.com/biggs-100/odachi/internal/domain/store"
)

// lintCheck is the one tool most of these tests declare.
var lintCheck = domainreview.Check{
	Name:     "lint",
	Command:  "golangci-lint run ./...",
	Severity: domainreview.Warning,
}

// approveChecks is the clone's agreement, written where the store keeps it: the
// declaration travels with the repository and the agreement to run it does not.
func (p *project) approveChecks(checks []domainreview.Check) {
	approval := domainreview.NewCheckApproval(checks, "biggs", "2026-11-12T09:00:00Z")
	p.files.files[filepath.Join(store, storeentry.Checks)] = domainreview.RenderCheckApproval(approval)
}

// TestTheDeclaredCheckRunsOnTheCandidate: the tool runs in the project the prompt
// names, before the reviewer is paid, and its words travel into the prompt.
func TestTheDeclaredCheckRunsOnTheCandidate(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: something is wrong\n", exitCode: 1},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if len(p.shell.commands) != 1 {
		t.Fatalf("commands = %v, want the check to have run once", p.shell.commands)
	}
	if p.shell.commands[0] != lintCheck.Command {
		t.Errorf("the first command was %q, want the declared check", p.shell.commands[0])
	}
	if len(p.shell.dirs) == 0 || p.shell.dirs[0] != dir {
		t.Errorf("the check ran in %v, want the candidate at %q", p.shell.dirs, dir)
	}
	if !strings.Contains(result.Prompt, "## The checks (the tools this project declared, run before any lens)") {
		t.Fatalf("the prompt does not carry the checks:\n%s", result.Prompt)
	}
	for _, want := range []string{"- lint: failed (exit 1)", "a.go:1:1: something is wrong"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q", want)
		}
	}
}

// TestACheckThisCloneNeverApprovedDoesNotRun: the declaration travels with the
// repository and the agreement to execute it does not, so a cloned review.toml
// cannot run anything. The prompt says what was skipped, because a review that
// stayed quiet would report on a tree nobody measured.
func TestACheckThisCloneNeverApprovedDoesNotRun(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("commands = %v, want nothing to run before an approval", p.shell.commands)
	}
	for _, want := range []string{
		"## The checks that did not run",
		"this clone has not approved them",
		"odachi review checks approve",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, result.Prompt)
		}
	}
}

// TestAListTheProjectChangedIsNotApproved: the approval covers the exact list, so
// a command that gained a flag is a tool nobody agreed to run.
func TestAListTheProjectChangedIsNotApproved(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	changed := domainreview.Check{
		Name:     "lint",
		Command:  "golangci-lint run --fix ./...",
		Severity: domainreview.Warning,
	}

	result := p.run(review.Options{Checks: []domainreview.Check{changed}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("commands = %v, want nothing to run for a list that changed", p.shell.commands)
	}
	if !strings.Contains(result.Prompt, "approved a different list") {
		t.Errorf("the prompt does not say the approval covers another list:\n%s", result.Prompt)
	}
}

// TestAPassingCheckSaysSoWithoutOutput: a tool that found nothing is one line, and
// the reviewer reads it as the part of the tree it does not have to inspect.
func TestAPassingCheckSaysSoWithoutOutput(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "- lint: passed") {
		t.Errorf("the prompt does not say the check passed:\n%s", result.Prompt)
	}
}

// TestACheckThatCannotStartIsARowNotARefusal: a tool nobody installed is the
// project's mistake, and it must be readable as a row instead of hiding the
// reviewer's answer behind a crash.
func TestACheckThatCannotStartIsARowNotARefusal(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{err: errors.New("executable file not found in PATH")},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"- lint: could not run", "executable file not found in PATH"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, result.Prompt)
		}
	}
}

// TestACheckThatTimesOutSaysHowLongItWaited: a hung tool is not a pass, and the
// prompt names the ceiling so the human can raise it or drop the tool.
func TestACheckThatTimesOutSaysHowLongItWaited(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{timedOut: true},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{
		Checks:       []domainreview.Check{lintCheck},
		CheckTimeout: 45,
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Prompt, "- lint: timed out · stopped after 45 s") {
		t.Errorf("the prompt does not name the timed-out check:\n%s", result.Prompt)
	}
	if p.shell.timeout != 45 {
		t.Errorf("the check ran with a %ds ceiling, want the project's 45", p.shell.timeout)
	}
}

// TestTheChecksRunOnceHoweverManyGroupsPay: two models read two prompts about one
// candidate, and the tools that measured it ran once. A tool paid twice for one
// tree is money thrown at the same answer.
func TestTheChecksRunOnceHoweverManyGroupsPay(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"risk": "sharp/model"},
		Checks:     []domainreview.Check{lintCheck},
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	runs := 0
	for _, command := range p.shell.commands {
		if command == lintCheck.Command {
			runs++
		}
	}
	if runs != 1 {
		t.Errorf("the check ran %d times over one candidate, want once: %v", runs, p.shell.commands)
	}
}

// TestAProjectThatDeclaresNoChecksRunsNothing: the absence of the table is the
// absence of the feature, and the reviewer's command is the only thing executed.
func TestAProjectThatDeclaresNoChecksRunsNothing(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 1 || p.shell.commands[0] == lintCheck.Command {
		t.Errorf("commands = %v, want only the reviewer", p.shell.commands)
	}
	if strings.Contains(result.Prompt, "## The checks (the tools this project declared, run before any lens)") {
		t.Errorf("the prompt claims checks nobody declared:\n%s", result.Prompt)
	}
}

// storedVerdictOf is what the run filed for the current tree.
func (p *project) storedVerdictOf(t *testing.T) domainreview.Verdict {
	t.Helper()
	verdict, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	return verdict
}

// TestAFailingCheckIsAFindingTheVerdictCarries: the tool filed the row, so the
// reviewer cannot approve past it. A warning is information and a critical blocks,
// exactly as it does for a row a person wrote.
func TestAFailingCheckIsAFindingTheVerdictCarries(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: something is wrong\n", exitCode: 1},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := p.storedVerdictOf(t)
	if len(stored.Findings) != 1 {
		t.Fatalf("findings = %+v, want the check's row", stored.Findings)
	}
	row := stored.Findings[0]
	if row.Check != "lint" {
		t.Errorf("the row does not name the check: %+v", row)
	}
	if row.Path != "" || row.Line != 0 {
		t.Errorf("the row claims a place: path %q line %d, want neither", row.Path, row.Line)
	}
	if row.Severity != domainreview.Warning {
		t.Errorf("severity = %q, want the declared WARNING", row.Severity)
	}
	if !strings.Contains(row.Why, "a.go:1:1: something is wrong") {
		t.Errorf("the row does not carry the tool's words: %q", row.Why)
	}
	if stored.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want %q: a warning does not block", stored.Verdict, domainreview.Approved)
	}
}

// TestACheckDeclaredAtCriticalBlocksTheTree: the severity is the project's, and the
// project that declares a blocker gets a blocker.
func TestACheckDeclaredAtCriticalBlocksTheTree(t *testing.T) {
	p := newProject()
	blocking := lintCheck
	blocking.Severity = domainreview.Critical
	p.approveChecks([]domainreview.Check{blocking})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: something is wrong\n", exitCode: 1},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{blocking}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := p.storedVerdictOf(t)
	if stored.Verdict != domainreview.ChangesRequested {
		t.Errorf("verdict = %q, want %q", stored.Verdict, domainreview.ChangesRequested)
	}
}

// TestAPassingCheckFilesNoRow: a tool that found nothing wrote nothing, and the
// verdict is the reviewer's alone.
func TestAPassingCheckFilesNoRow(t *testing.T) {
	p := newProject()
	p.approveChecks([]domainreview.Check{lintCheck})
	p.shell.script = []fakeAnswer{
		{},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if stored := p.storedVerdictOf(t); len(stored.Findings) != 0 {
		t.Errorf("findings = %+v, want none for a passing check", stored.Findings)
	}
}

// TestACheckThisCloneNeverApprovedFilesNoRow: nothing ran, so nothing is filed, and
// the verdict never claims a measurement nobody made.
func TestACheckThisCloneNeverApprovedFilesNoRow(t *testing.T) {
	p := newProject()
	p.shell.script = []fakeAnswer{
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if stored := p.storedVerdictOf(t); len(stored.Findings) != 0 {
		t.Errorf("findings = %+v, want none: the tools never ran", stored.Findings)
	}
}

// fileCheck is one failing blocking check, run and stored, ready to be dismissed.
func fileCheck(t *testing.T) *project {
	t.Helper()
	t.Parallel()

	p := newProject()
	blocking := lintCheck
	blocking.Severity = domainreview.Critical
	p.approveChecks([]domainreview.Check{blocking})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: something is wrong\n", exitCode: 1},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}
	if result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{blocking}}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	return p
}

// TestACheckRowIsDismissedByNamingTheTool: an analyzer has no reasoning to refute,
// so the human's recorded reason is the whole judgment -and the gate opens only
// with that reason written beside the verdict.
func TestACheckRowIsDismissedByNamingTheTool(t *testing.T) {
	p := fileCheck(t)

	result := p.run(review.Options{
		Dismiss: "check:lint",
		Because: "the ruleset flags a pattern this file must keep",
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "dismissed check:lint") {
		t.Errorf("message = %q, want it to name the tool", result.Message)
	}

	sidecar, err := domainreview.DismissedPath(currentTree)
	if err != nil {
		t.Fatalf("DismissedPath = %v", err)
	}
	dismissals, err := domainreview.ParseDismissals(p.files.files[filepath.Join(store, filepath.FromSlash(sidecar))])
	if err != nil {
		t.Fatalf("the dismissals do not parse: %v", err)
	}
	if len(dismissals.Rows) != 1 || dismissals.Rows[0].Check != "lint" {
		t.Fatalf("the dismissal does not name the tool: %+v", dismissals.Rows)
	}
}

// TestShowNamesTheToolThatFiledTheRow: the human reads `check:lint` and copies it
// to dismiss the row, so the face of the verdict says exactly what the verb takes.
func TestShowNamesTheToolThatFiledTheRow(t *testing.T) {
	p := fileCheck(t)

	result := p.run(review.Options{Show: true})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	printed := strings.Join(append([]string{result.Message}, result.Lines...), "\n")
	if !strings.Contains(printed, "check:lint") {
		t.Errorf("--show does not name the tool:\n%s", printed)
	}
	if strings.Contains(printed, " · 0") {
		t.Errorf("--show prints a place for a row that has none:\n%s", printed)
	}
}

// TestADismissalOfACheckThatIsNotThereNamesTheToolsItHas: the human copies a name,
// and a name that is not in the verdict reads the ones that are.
func TestADismissalOfACheckThatIsNotThereNamesTheToolsItHas(t *testing.T) {
	p := fileCheck(t)

	result := p.run(review.Options{
		Dismiss: "check:vuln",
		Because: "not this one",
	})
	if result.Code == review.CodeOK {
		t.Fatalf("dismissing a row that is not there = 0, want a refusal (%s)", result.Message)
	}
	if !strings.Contains(result.Message, "check:lint") {
		t.Errorf("the refusal %q does not name the rows the verdict holds", result.Message)
	}
}

// TestShowPrintsTheRowAWaiverAnswers: the human copies the identity the gate would
// refuse on, instead of learning it by being stopped.
func TestShowPrintsTheRowAWaiverAnswers(t *testing.T) {
	p := newProject()
	blocking := domainreview.Finding{
		Path: "a.go", Line: 3, Severity: domainreview.Critical, Causal: true,
		Lens: domainreview.Reliability, Why: "the round trip is lost", Blob: gitBlob,
	}
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "revisor",
		Lenses:   []string{domainreview.Reliability},
		Findings: []domainreview.Finding{blocking},
	}, currentTree, review.DefaultBase, now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(sealed)

	result := p.run(review.Options{Show: true})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	printed := strings.Join(append([]string{result.Message}, result.Lines...), "\n")
	if !strings.Contains(printed, "row "+domainreview.Fingerprint(blocking)) {
		t.Errorf("--show does not print the row a waiver answers:\n%s", printed)
	}
}
