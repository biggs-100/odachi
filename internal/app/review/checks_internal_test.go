package review

import (
	"strings"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/review"
)

// TestEveryUnmeasuredCheckStateProducesBlockingRow: a declared tool that did not
// complete a measurement cannot approve a candidate, regardless of its declared
// severity or whether this binary currently knows the status name.
func TestEveryUnmeasuredCheckStateProducesBlockingRow(t *testing.T) {
	for _, status := range []domain.CheckStatus{
		domain.CheckBlocked,
		domain.CheckBroken,
		domain.CheckTimedOut,
		"pending",
		"",
	} {
		for _, severity := range []domain.Severity{
			domain.Warning,
			domain.Suggestion,
			domain.Critical,
			domain.Blocker,
		} {
			name := string(status)
			if strings.TrimSpace(name) == "" {
				name = "empty"
			}
			t.Run(name+"/"+string(severity), func(t *testing.T) {
				results := []domain.CheckResult{{
					Name: "lint", Command: "lint ./...", Severity: severity,
					Status: status, Note: "the measurement stopped early",
				}}
				r := &runner{ran: &results}
				rows, problem := r.checkFindings()
				if problem != nil {
					t.Fatalf("checkFindings: %s", problem.Message)
				}
				if len(rows) != 1 {
					t.Fatalf("findings = %+v, want one synthetic row", rows)
				}
				shownStatus := string(status)
				if strings.TrimSpace(shownStatus) == "" {
					shownStatus = "unknown"
				}
				for _, want := range []string{
					"lint", "status " + shownStatus, "declared severity " + string(severity),
					"measurement did not complete", "the measurement stopped early",
				} {
					if !strings.Contains(rows[0].Why, want) {
						t.Errorf("the row does not carry %q: %q", want, rows[0].Why)
					}
				}
				if rows[0].Severity != domain.Blocker || !rows[0].Causal || rows[0].Evidence != domain.EvidenceCited {
					t.Errorf("the incomplete measurement is not an open causal BLOCKER: %+v", rows[0])
				}
				if got := domain.Derive(rows); got != domain.ChangesRequested {
					t.Errorf("Derive = %q, want %q", got, domain.ChangesRequested)
				}
			})
		}
	}
}

// TestMeasuredCheckResultsKeepDeclaredSeverity: a completed measurement is the
// project's result at its chosen severity; only incomplete evidence is promoted.
func TestMeasuredCheckResultsKeepDeclaredSeverity(t *testing.T) {
	for _, severity := range []domain.Severity{domain.Warning, domain.Suggestion, domain.Critical, domain.Blocker} {
		t.Run(string(severity), func(t *testing.T) {
			results := []domain.CheckResult{{Name: "lint", Severity: severity, Status: domain.CheckFailed, Exit: 1}}
			r := &runner{ran: &results}
			rows, problem := r.checkFindings()
			if problem != nil {
				t.Fatalf("checkFindings: %s", problem.Message)
			}
			if len(rows) != 1 || rows[0].Severity != severity {
				t.Fatalf("findings = %+v, want one measured %s row", rows, severity)
			}
			if strings.Contains(rows[0].Why, "measurement did not complete") {
				t.Errorf("a measured failure was described as incomplete: %q", rows[0].Why)
			}

			results[0].Status = domain.CheckPassed
			rows, problem = r.checkFindings()
			if problem != nil {
				t.Fatalf("checkFindings after pass: %s", problem.Message)
			}
			if len(rows) != 0 {
				t.Errorf("a measured pass filed rows: %+v", rows)
			}
		})
	}
}

type absentCheckDismissals struct{ Files }

func (absentCheckDismissals) Exists(string) (bool, error) { return false, nil }

// TestFutureUnmeasuredCheckCannotClearPredecessorFailure: a status this binary
// does not recognize is still incomplete evidence, so it preserves the measured
// failure and adds its own synthetic BLOCKER.
func TestFutureUnmeasuredCheckCannotClearPredecessorFailure(t *testing.T) {
	prior := domain.Finding{
		Check: "lint", Severity: domain.Critical, Causal: true,
		Evidence: domain.EvidenceCited, Why: "the check lint exited 1\nprior output",
	}
	results := []domain.CheckResult{{Name: "lint", Severity: domain.Warning, Status: "pending"}}
	r := &runner{
		opts:  Options{Dir: "project", Store: "store"},
		files: absentCheckDismissals{},
		ran:   &results,
	}
	rows, problem := r.checkFindings()
	if problem != nil {
		t.Fatalf("checkFindings: %s", problem.Message)
	}
	kept, _, note := r.reconcileCheckFindings(
		"abc1234", []domain.Finding{prior}, rows,
		[]domain.Check{{Name: "lint", Severity: domain.Warning}}, results,
	)
	if note != "" {
		t.Fatalf("unexpected dismissal authority note: %s", note)
	}
	if len(kept) != 2 {
		t.Fatalf("findings = %+v, want the predecessor and the incomplete current row", kept)
	}
	if kept[0] != prior {
		t.Errorf("the measured predecessor changed: %+v", kept[0])
	}
	if kept[1].Severity != domain.Blocker || !strings.Contains(kept[1].Why, "status pending") {
		t.Errorf("the future state did not produce a synthetic BLOCKER: %+v", kept[1])
	}
}
