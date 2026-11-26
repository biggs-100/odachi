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

const cleanReviewerAnswer = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

func requireCausalCheck(t *testing.T, row domainreview.Finding, name string, severity domainreview.Severity, wants ...string) {
	t.Helper()
	if row.Check != name || row.Severity != severity || !row.Causal || row.Evidence != domainreview.EvidenceCited || row.Status != "" || row.Reason != "" {
		t.Fatalf("check row = %+v, want open causal %s at %s", row, name, severity)
	}
	for _, want := range wants {
		if !strings.Contains(row.Why, want) {
			t.Errorf("%s does not carry %q: %q", name, want, row.Why)
		}
	}
}

func requireChangesRequested(t *testing.T, verdict domainreview.Verdict) {
	t.Helper()
	if verdict.Verdict != domainreview.ChangesRequested {
		t.Errorf("verdict = %q, want %q", verdict.Verdict, domainreview.ChangesRequested)
	}
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

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{lintCheck}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"- lint: could not run", "executable file not found in PATH"} {
		if !strings.Contains(p.shell.input, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, p.shell.input)
		}
	}
	stored := p.storedVerdictOf(t)
	if len(stored.Findings) != 1 {
		t.Fatalf("findings = %+v, want the missing executable's row", stored.Findings)
	}
	requireCausalCheck(t, stored.Findings[0], "lint", domainreview.Blocker,
		"could not run", "executable file not found in PATH", "status could not run",
		"declared severity WARNING", "measurement did not complete")
	requireChangesRequested(t, stored)
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
		Exec:         "revisor",
		Checks:       []domainreview.Check{lintCheck},
		CheckTimeout: 45,
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(p.shell.input, "- lint: timed out · stopped after 45 s") {
		t.Errorf("the prompt does not name the timed-out check:\n%s", p.shell.input)
	}
	stored := p.storedVerdictOf(t)
	if len(stored.Findings) != 1 {
		t.Fatalf("findings = %+v, want the timed-out check's row", stored.Findings)
	}
	requireCausalCheck(t, stored.Findings[0], "lint", domainreview.Blocker,
		"stopped by the ceiling", "stopped after 45 s", "status timed out",
		"declared severity WARNING", "measurement did not complete")
	requireChangesRequested(t, stored)
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

// TestUnapprovedDeclaredChecksProduceBlockingRows: missing or drifted approval
// executes nothing, but every tool remains visible as an open causal row.
func TestUnapprovedDeclaredChecksProduceBlockingRows(t *testing.T) {
	blocking := lintCheck
	blocking.Severity = domainreview.Critical
	vuln := domainreview.Check{
		Name: "vuln", Command: "govulncheck ./...",
		Severity: domainreview.Warning, Remedy: "upgrade the vulnerable module",
	}
	suggestion := domainreview.Check{
		Name: "format", Command: "gofmt -l .",
		Severity: domainreview.Suggestion, Remedy: "format the changed files",
	}
	changed := lintCheck
	changed.Command = "golangci-lint run --fix ./..."
	changed.Severity = domainreview.Critical
	for _, tc := range []struct {
		name     string
		approved []domainreview.Check
		checks   []domainreview.Check
		reason   string
	}{
		{name: "missing approval", checks: []domainreview.Check{blocking, vuln}, reason: "has not approved them"},
		{name: "warning without approval", checks: []domainreview.Check{vuln}, reason: "has not approved them"},
		{name: "suggestion without approval", checks: []domainreview.Check{suggestion}, reason: "has not approved them"},
		{name: "approval drift", approved: []domainreview.Check{lintCheck}, checks: []domainreview.Check{changed}, reason: "approved a different list"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject()
			if len(tc.approved) > 0 {
				p.approveChecks(tc.approved)
			}
			p.shell.output = cleanReviewerAnswer
			result := p.run(review.Options{Exec: "revisor", Checks: tc.checks})
			if result.Code != review.CodeOK {
				t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
			}
			if len(p.shell.commands) != 1 || p.shell.commands[0] != "revisor" {
				t.Fatalf("commands = %v, want only the reviewer: unapproved tools must never execute", p.shell.commands)
			}
			stored := p.storedVerdictOf(t)
			if len(stored.Findings) != len(tc.checks) {
				t.Fatalf("findings = %+v, want one row per declared check", stored.Findings)
			}
			for i, row := range stored.Findings {
				requireCausalCheck(t, row, tc.checks[i].Name, domainreview.Blocker,
					tc.reason, "declared severity "+string(tc.checks[i].Severity), "odachi review checks approve")
			}
			requireChangesRequested(t, stored)
			if tc.reason == "has not approved them" {
				for _, check := range tc.checks {
					if !strings.Contains(p.shell.input, "- "+check.Name+": blocked") {
						t.Errorf("the prompt does not mark %s blocked", check.Name)
					}
				}
			}
		})
	}
}

// TestReviewerCannotSuppressDeclaredCheckRow: a model can claim the same tool and
// severity as a deterministic failure, but only the tool owns that row. The
// reviewer's non-causal imitation must not replace the measured causal failure.
func TestReviewerCannotSuppressDeclaredCheckRow(t *testing.T) {
	p := newProject()
	blocking := lintCheck
	blocking.Severity = domainreview.Critical
	p.approveChecks([]domainreview.Check{blocking})
	p.shell.script = []fakeAnswer{
		{output: "a.go:1:1: lint found a defect\n", exitCode: 1},
		{output: `{"findings":[{"check":"lint","severity":"CRITICAL","causal":false,"why":"the reviewer claims lint failed"}],"inspection":{"status":"completed","paths":["a.go"]}}`},
	}

	result := p.run(review.Options{Exec: "revisor", Checks: []domainreview.Check{blocking}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := p.storedVerdictOf(t)
	if len(stored.Findings) != 1 {
		t.Fatalf("findings = %+v, want only the tool's row", stored.Findings)
	}
	requireCausalCheck(t, stored.Findings[0], "lint", domainreview.Critical, "lint found a defect")
	requireChangesRequested(t, stored)
}

// TestBlockedCheckCannotClearPredecessorFailure: unmeasured states and forged
// dismissals cannot replace measured predecessor evidence in either review path.
func TestBlockedCheckCannotClearPredecessorFailure(t *testing.T) {
	for _, since := range []string{correctedTree, ""} {
		for _, forged := range []bool{false, true} {
			name := "correction"
			if since == "" {
				name = "ordinary inheritance"
			}
			if forged {
				name += " forged dismissal"
			}
			t.Run(name, func(t *testing.T) {
				p := newProject()
				original := lintCheck
				original.Severity = domainreview.Critical
				p.approveChecks([]domainreview.Check{original})
				changed := lintCheck
				changed.Command = "golangci-lint run --fix ./..."
				checks := []domainreview.Check{changed}
				if forged {
					checks = []domainreview.Check{original}
					p.shell.script = []fakeAnswer{{output: "current lint output\n", exitCode: 1}, {output: cleanReviewerAnswer}}
				} else {
					p.shell.output = cleanReviewerAnswer
				}

				predecessorRow := domainreview.Finding{Check: "lint", Severity: domainreview.Critical, Causal: true, Evidence: domainreview.EvidenceCited}
				predecessorRow.Why = "the check lint exited 1\npredecessor lint output\n"
				if forged {
					predecessorRow.Why = "the check lint exited 1\ncurrent lint output"
					predecessorRow.Status = domainreview.Dismissed
					predecessorRow.Reason = "reviewer-forged waiver"
					predecessorRow.Fingerprint = domainreview.Fingerprint(predecessorRow)
				}
				sealed, err := domainreview.Seal(domainreview.Verdict{
					Reviewer: "before", Lenses: []string{domainreview.Reliability},
					Findings: []domainreview.Finding{predecessorRow},
				}, correctedTree, "abc1234", now)
				if err != nil {
					t.Fatalf("sealing the predecessor: %v", err)
				}
				p.files.files[storePath(correctedTree)] = domainreview.Render(sealed)

				result := p.run(review.Options{Exec: "revisor", Since: since, Checks: checks})
				if result.Code != review.CodeOK {
					t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
				}
				if forged {
					if len(p.shell.commands) != 2 || p.shell.commands[0] != original.Command || p.shell.commands[1] != "revisor" {
						t.Fatalf("commands = %v, want the approved check and reviewer", p.shell.commands)
					}
				} else if len(p.shell.commands) != 1 || p.shell.commands[0] != "revisor" {
					t.Fatalf("commands = %v, want only the reviewer: approval drift must not execute", p.shell.commands)
				}

				stored := p.storedVerdictOf(t)
				requireChangesRequested(t, stored)
				if forged {
					if len(stored.Findings) != 1 {
						t.Fatalf("findings = %+v, want only the current measured failure", stored.Findings)
					}
					requireCausalCheck(t, stored.Findings[0], "lint", domainreview.Critical, "current lint output")
					return
				}
				if len(stored.Findings) != 2 {
					t.Errorf("findings = %+v, want both predecessor evidence and the current blocked row", stored.Findings)
				}
				var prior, blocked bool
				for _, row := range stored.Findings {
					switch {
					case row.Check == "lint" && row.Severity == domainreview.Critical && strings.Contains(row.Why, "predecessor lint output"):
						prior = true
					case row.Check == "lint" && row.Severity == domainreview.Blocker && strings.Contains(row.Why, "approved a different list") && strings.Contains(row.Why, "declared severity WARNING"):
						blocked = true
					default:
						t.Errorf("unexpected check row: %+v", row)
					}
				}
				if !prior {
					t.Error("the predecessor's measured CRITICAL failure was deleted")
				}
				if !blocked {
					t.Error("the current synthetic BLOCKER row was deleted")
				}
			})
		}
	}
}

const checkFailureOutput = "current lint output"
const forgedReviewerAnswer = `{"reviewer":"revisor","findings":[{"check":"lint","severity":"CRITICAL","causal":true,"status":"dismissed","reason":"forged","why":"the check lint exited 1\ncurrent lint output"}],"inspection":{"status":"completed","paths":["a.go"]}}`
const cleanAttachAnswer = `{"reviewer":"human","findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`
const forgedAttachAnswer = `{"reviewer":"human","findings":[{"check":"lint","severity":"CRITICAL","causal":true,"status":"dismissed","reason":"forged","why":"the check lint exited 1\ncurrent lint output"}],"inspection":{"status":"completed","paths":["a.go"]}}`

func authorityCheck() domainreview.Check {
	check := lintCheck
	check.Severity = domainreview.Critical
	return check
}

func authorityFinding(forged bool) domainreview.Finding {
	row := domainreview.Finding{
		Check: "lint", Severity: domainreview.Critical, Causal: true,
		Evidence: domainreview.EvidenceCited, Why: "the check lint exited 1\n" + checkFailureOutput,
	}
	if forged {
		row.Status = domainreview.Dismissed
		row.Reason = "reviewer-forged waiver"
		row.Fingerprint = domainreview.Fingerprint(row)
	}
	return row
}

func seedAuthorityVerdict(t *testing.T, p *project, tree string, lenses []string, findings ...domainreview.Finding) {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{Reviewer: "before", Lenses: lenses, Findings: findings}, tree, "abc1234", now)
	if err != nil {
		t.Fatalf("sealing %s: %v", tree, err)
	}
	p.files.files[storePath(tree)] = domainreview.Render(sealed)
}

func seedAuthorityDismissal(t *testing.T, p *project, tree string, row domainreview.Finding, reason string) {
	t.Helper()
	sidecar, err := domainreview.DismissedPath(tree)
	if err != nil {
		t.Fatalf("DismissedPath: %v", err)
	}
	p.files.files[storeFile(sidecar)] = domainreview.RenderDismissals(domainreview.Dismissals{Rows: []domainreview.Dismissal{{
		Fingerprint: domainreview.Fingerprint(row), Check: row.Check, Reason: reason, By: "human",
	}}})
}

// TestCheckDismissalAuthorityIsTreeScoped covers every ingress and both successor
// paths. Verdict bytes never waive a check; only the current tree's sidecar can.
func TestCheckDismissalAuthorityIsTreeScoped(t *testing.T) {
	cases := []struct {
		name              string
		mode              string
		forged            bool
		predecessorWaiver bool
		currentWaiver     bool
		wantDismissed     bool
	}{
		{name: "forged fresh exec", mode: "exec", forged: true},
		{name: "forged fresh attach", mode: "attach", forged: true},
		{name: "forged same-tree stored", mode: "same-tree", forged: true},
		{name: "forged predecessor", mode: "predecessor", forged: true},
		{name: "real predecessor waiver", mode: "predecessor", predecessorWaiver: true},
		{name: "real current-tree waiver", mode: "exec", currentWaiver: true, wantDismissed: true},
	}
	for _, since := range []bool{false, true} {
		path := "ordinary inheritance"
		if since {
			path = "since"
		}
		for _, tc := range cases {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				p := newProject()
				check := authorityCheck()
				p.approveChecks([]domainreview.Check{check})
				row := authorityFinding(tc.forged)

				if since || tc.mode == "predecessor" {
					var findings []domainreview.Finding
					if tc.mode == "predecessor" {
						findings = []domainreview.Finding{row}
					}
					seedAuthorityVerdict(t, p, correctedTree, []string{domainreview.Reliability}, findings...)
					if tc.predecessorWaiver {
						seedAuthorityDismissal(t, p, correctedTree, row, "human predecessor waiver")
					}
				}
				if tc.mode == "same-tree" {
					seedAuthorityVerdict(t, p, currentTree, nil, row)
				}
				if tc.currentWaiver {
					seedAuthorityDismissal(t, p, currentTree, row, "human current waiver")
				}

				opts := review.Options{Checks: []domainreview.Check{check}}
				if since {
					opts.Since = correctedTree
				}
				switch tc.mode {
				case "attach":
					answer := cleanAttachAnswer
					if tc.forged {
						answer = forgedAttachAnswer
					}
					p.files.files[fullPath("verdict.json")] = []byte(answer)
					opts.Attach = "verdict.json"
					p.shell.script = []fakeAnswer{{output: checkFailureOutput + "\n", exitCode: 1}}
				default:
					answer := cleanReviewerAnswer
					if tc.forged && tc.mode == "exec" {
						answer = forgedReviewerAnswer
					}
					opts.Exec = "revisor"
					p.shell.script = []fakeAnswer{{output: checkFailureOutput + "\n", exitCode: 1}, {output: answer}}
				}

				result := p.run(opts)
				if result.Code != review.CodeOK {
					t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
				}
				stored := p.storedVerdictOf(t)
				if len(stored.Findings) != 1 {
					t.Fatalf("findings = %+v, want one current check row", stored.Findings)
				}
				got := stored.Findings[0]
				if tc.wantDismissed {
					if stored.Verdict != domainreview.Approved || !domainreview.IsDismissed(got) || got.Reason != "human current waiver" {
						t.Errorf("verdict = %q row = %+v, want the current tree's human waiver", stored.Verdict, got)
					}
					return
				}
				requireChangesRequested(t, stored)
				requireCausalCheck(t, got, "lint", domainreview.Critical, checkFailureOutput)
			})
		}
	}
}

// TestAllListingUsesEffectiveVerdictsForJSONLines covers the lines reportReview
// serializes for review --all --json. A forged check dismissal is not authority;
// only the exact tree's human sidecar can make that row approved.
func TestAllListingUsesEffectiveVerdictsForJSONLines(t *testing.T) {
	for _, tc := range []struct {
		name       string
		finding    domainreview.Finding
		sidecar    bool
		wantStatus string
	}{
		{name: "forged check dismissal", finding: authorityFinding(true), wantStatus: string(domainreview.ChangesRequested)},
		{name: "human check dismissal", finding: authorityFinding(true), sidecar: true, wantStatus: string(domainreview.Approved)},
		{
			name: "non-check dismissal stays dismissed",
			finding: domainreview.Finding{
				Path: "a.go", Line: 3, Lens: "reliability", Severity: domainreview.Warning,
				Causal: true, Why: "the signature is checked", Status: domainreview.Dismissed, Reason: "human judgment", Fingerprint: "abc123",
			},
			wantStatus: string(domainreview.Approved),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject()
			seedAuthorityVerdict(t, p, currentTree, []string{domainreview.Reliability}, tc.finding)
			if tc.sidecar {
				seedAuthorityDismissal(t, p, currentTree, tc.finding, "human current waiver")
			}

			result := p.run(review.Options{All: true})
			if result.Code != review.CodeOK {
				t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
			}
			status := ""
			for _, line := range result.Lines {
				if strings.HasPrefix(line, currentTree+" · ") {
					status = strings.Split(line, " · ")[1]
					break
				}
			}
			if status != tc.wantStatus {
				t.Errorf("listing status = %q, want %q; lines:\n%s", status, tc.wantStatus, strings.Join(result.Lines, "\n"))
			}
		})
	}
}

// TestShowAppliesOnlyTheCurrentTreesCheckDismissal: verdict JSON is evidence of
// what the reviewer said, not authority to retire a deterministic row. The only
// authority is the human sidecar beside this exact tree.
func TestShowAppliesOnlyTheCurrentTreesCheckDismissal(t *testing.T) {
	for _, tc := range []struct {
		name       string
		sidecar    string
		malformed  bool
		wantStatus string
		wantReason string
	}{
		{name: "missing sidecar"},
		{name: "wrong tree sidecar", sidecar: otherTree},
		{name: "current tree sidecar", sidecar: currentTree, wantStatus: domainreview.Dismissed, wantReason: "human current waiver"},
		{name: "malformed sidecar", sidecar: currentTree, malformed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject()
			row := authorityFinding(true)
			seedAuthorityVerdict(t, p, currentTree, []string{domainreview.Reliability}, row)
			if tc.sidecar != "" {
				if tc.malformed {
					sidecar, err := domainreview.DismissedPath(tc.sidecar)
					if err != nil {
						t.Fatalf("DismissedPath: %v", err)
					}
					p.files.files[storeFile(sidecar)] = []byte("not json")
				} else {
					seedAuthorityDismissal(t, p, tc.sidecar, row, "human current waiver")
				}
			}

			shown := p.run(review.Options{Show: true})
			if shown.Code != review.CodeOK {
				t.Fatalf("code = %d, want 0 (%s)", shown.Code, shown.Message)
			}
			if shown.Verdict == nil || len(shown.Verdict.Findings) != 1 {
				t.Fatalf("shown verdict = %+v, want one effective check row", shown.Verdict)
			}
			got := shown.Verdict.Findings[0]
			if got.Status != tc.wantStatus || got.Reason != tc.wantReason {
				t.Errorf("row = %+v, want status %q and reason %q", got, tc.wantStatus, tc.wantReason)
			}
			if tc.wantStatus == "" && shown.Verdict.Verdict != domainreview.ChangesRequested {
				t.Errorf("verdict = %q, want changes_requested without current-tree authority", shown.Verdict.Verdict)
			}
			if tc.malformed && !strings.Contains(strings.Join(shown.Lines, "\n"), "cannot be read") {
				t.Errorf("--show hides the malformed sidecar: %v", shown.Lines)
			}
		})
	}
}

// TestRetireReadsEffectiveStoredCheckRows exercises storedVerdictFor: a legacy
// dismissal in verdict JSON is reopened, a current sidecar is already retired, and
// a sidecar for another tree or an unreadable one supplies no authority here.
func TestRetireReadsEffectiveStoredCheckRows(t *testing.T) {
	for _, tc := range []struct {
		name       string
		sidecar    string
		malformed  bool
		wantCode   int
		wantDetail string
	}{
		{name: "legacy forged dismissal", wantCode: review.CodeOK},
		{name: "wrong tree sidecar", sidecar: otherTree, wantCode: review.CodeOK},
		{name: "current tree sidecar", sidecar: currentTree, wantCode: review.CodeBlocked, wantDetail: "no open row"},
		{name: "malformed current sidecar", sidecar: currentTree, malformed: true, wantCode: review.CodeBlocked, wantDetail: "dismissals"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject()
			row := authorityFinding(true)
			seedAuthorityVerdict(t, p, currentTree, []string{domainreview.Reliability}, row)
			if tc.sidecar != "" {
				if tc.malformed {
					sidecar, err := domainreview.DismissedPath(tc.sidecar)
					if err != nil {
						t.Fatalf("DismissedPath: %v", err)
					}
					p.files.files[storeFile(sidecar)] = []byte("not json")
				} else {
					seedAuthorityDismissal(t, p, tc.sidecar, row, "human current waiver")
				}
			}

			got := p.run(review.Options{Dismiss: "check:lint", Because: "a fresh human judgment"})
			if got.Code != tc.wantCode {
				t.Fatalf("code = %d, want %d (%s)", got.Code, tc.wantCode, got.Message)
			}
			if tc.wantDetail != "" && !strings.Contains(got.Message, tc.wantDetail) {
				t.Errorf("message = %q, want %q", got.Message, tc.wantDetail)
			}
		})
	}
}

// TestUnmeasuredCheckPreservesExistingFailure: broken, timed-out and blocked
// tools did not measure the candidate. Their own row is added, but neither a
// predecessor nor a same-tree stored failure may disappear.
func TestUnmeasuredCheckPreservesExistingFailure(t *testing.T) {
	for _, source := range []string{"predecessor", "same tree"} {
		for _, state := range []string{"broken", "timed out", "blocked"} {
			t.Run(source+"/"+state, func(t *testing.T) {
				p := newProject()
				check := authorityCheck()
				prior := authorityFinding(false)
				prior.Why = "the check lint exited 1\npredecessor lint output"
				opts := review.Options{Exec: "revisor", Checks: []domainreview.Check{check}}
				if source == "predecessor" {
					seedAuthorityVerdict(t, p, correctedTree, []string{domainreview.Reliability}, prior)
					opts.Since = correctedTree
				} else {
					seedAuthorityVerdict(t, p, currentTree, nil, prior)
				}

				switch state {
				case "broken":
					p.approveChecks([]domainreview.Check{check})
					p.shell.script = []fakeAnswer{{err: errors.New("executable file not found in PATH")}, {output: cleanReviewerAnswer}}
				case "timed out":
					p.approveChecks([]domainreview.Check{check})
					p.shell.script = []fakeAnswer{{timedOut: true}, {output: cleanReviewerAnswer}}
				case "blocked":
					p.shell.output = cleanReviewerAnswer
				}

				result := p.run(opts)
				if result.Code != review.CodeOK {
					t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
				}
				stored := p.storedVerdictOf(t)
				if len(stored.Findings) != 2 {
					t.Fatalf("findings = %+v, want the stored failure and the unmeasured row", stored.Findings)
				}
				var priorKept, currentKept bool
				for _, row := range stored.Findings {
					if strings.Contains(row.Why, "predecessor lint output") {
						priorKept = true
					}
					current := false
					switch state {
					case "broken":
						current = strings.Contains(row.Why, "could not run")
					case "timed out":
						current = strings.Contains(row.Why, "stopped by the ceiling")
					case "blocked":
						current = strings.Contains(row.Why, "did not run")
					}
					if !current {
						continue
					}
					if row.Severity != domainreview.Blocker || !strings.Contains(row.Why, "measurement did not complete") ||
						!strings.Contains(row.Why, "declared severity CRITICAL") {
						t.Errorf("the incomplete %s check is not a described synthetic BLOCKER: %+v", state, row)
					}
					currentKept = true
				}
				if !priorKept || !currentKept {
					t.Errorf("prior=%v current=%v, want both rows preserved: %+v", priorKept, currentKept, stored.Findings)
				}
			})
		}
	}
}

// TestMeasuredCheckReconcilesExistingFailure: only a complete measurement clears
// or replaces stored tool evidence, and different failure output is a new row.
func TestMeasuredCheckReconcilesExistingFailure(t *testing.T) {
	for _, source := range []string{"predecessor", "same tree"} {
		for _, passed := range []bool{true, false} {
			name := source + "/failure"
			if passed {
				name = source + "/pass"
			}
			t.Run(name, func(t *testing.T) {
				p := newProject()
				check := authorityCheck()
				p.approveChecks([]domainreview.Check{check})
				prior := authorityFinding(false)
				prior.Why = "the check lint exited 1\npredecessor lint output"
				opts := review.Options{Exec: "revisor", Checks: []domainreview.Check{check}}
				if source == "predecessor" {
					seedAuthorityVerdict(t, p, correctedTree, []string{domainreview.Reliability}, prior)
					opts.Since = correctedTree
				} else {
					seedAuthorityVerdict(t, p, currentTree, nil, prior)
				}
				answer := fakeAnswer{output: "replacement lint output\n", exitCode: 1}
				if passed {
					answer = fakeAnswer{}
				}
				p.shell.script = []fakeAnswer{answer, {output: cleanReviewerAnswer}}

				result := p.run(opts)
				if result.Code != review.CodeOK {
					t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
				}
				stored := p.storedVerdictOf(t)
				if passed {
					if len(stored.Findings) != 0 {
						t.Fatalf("findings = %+v, want a measured pass to clear the old failure", stored.Findings)
					}
					return
				}
				if len(stored.Findings) != 1 {
					t.Fatalf("findings = %+v, want only the replacement failure", stored.Findings)
				}
				requireCausalCheck(t, stored.Findings[0], "lint", domainreview.Critical, "replacement lint output")
				if strings.Contains(stored.Findings[0].Why, "predecessor lint output") {
					t.Errorf("the old output survived the measured failure: %+v", stored.Findings[0])
				}
			})
		}
	}
}

// TestSyntheticBlockedCheckIsNotDeduplicatedByLegacyRow: a legacy row can carry
// the same words at the declared WARNING severity. Reconciliation must keep the
// current synthetic BLOCKER instead of deduplicating away the missing measurement.
func TestSyntheticBlockedCheckIsNotDeduplicatedByLegacyRow(t *testing.T) {
	p := newProject()
	check := lintCheck
	why := "the check lint did not run: missing approval: the project declares 1 check and this clone has not approved them: run `odachi review checks approve` to run them; declared severity WARNING"
	prior := domainreview.Finding{
		Check: "lint", Severity: domainreview.Warning, Causal: true,
		Evidence: domainreview.EvidenceCited,
		Why:      why,
	}
	seedAuthorityVerdict(t, p, correctedTree, []string{domainreview.Reliability}, prior)
	p.shell.output = cleanReviewerAnswer

	result := p.run(review.Options{Exec: "revisor", Since: correctedTree, Checks: []domainreview.Check{check}})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := p.storedVerdictOf(t)
	requireChangesRequested(t, stored)
	if len(stored.Findings) != 2 {
		t.Fatalf("findings = %+v, want the legacy WARNING and current synthetic BLOCKER", stored.Findings)
	}
	var legacy, blocked bool
	for _, row := range stored.Findings {
		switch {
		case row.Check == "lint" && row.Severity == domainreview.Warning && row.Why == why:
			legacy = true
		case row.Check == "lint" && row.Severity == domainreview.Blocker && strings.Contains(row.Why, "missing approval") &&
			strings.Contains(row.Why, "measurement did not complete"):
			blocked = true
		default:
			t.Errorf("unexpected check row: %+v", row)
		}
	}
	if !legacy || !blocked {
		t.Fatalf("legacy=%v blocked=%v, want both rows preserved: %+v", legacy, blocked, stored.Findings)
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
