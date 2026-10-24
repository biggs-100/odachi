package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// dismissal is a finding judged and discarded with evidence: the same row a
// refuter kills with a citation.
func dismissal(why, reason string) review.Finding {
	return review.Finding{
		Path:        "internal/app/turnos.go",
		Line:        12,
		Severity:    review.Critical,
		Causal:      true,
		Why:         why,
		Lens:        "reliability",
		Status:      review.Dismissed,
		Reason:      reason,
		Fingerprint: review.Fingerprint(review.Finding{Path: "internal/app/turnos.go", Line: 12, Why: why, Lens: "reliability"}),
	}
}

// TestADismissalNeedsAReason: a dismissal without evidence is silence wearing a
// verdict's clothes. The status vocabulary is closed, and a reason only means
// something on a dismissed row.
func TestADismissalNeedsAReason(t *testing.T) {
	cases := []struct {
		name    string
		finding review.Finding
		want    string
	}{
		{
			name:    "dismissed without reason",
			finding: review.Finding{Path: "a.go", Severity: review.Warning, Why: "algo", Status: review.Dismissed},
			want:    "reason",
		},
		{
			name:    "unknown status",
			finding: review.Finding{Path: "a.go", Severity: review.Warning, Why: "algo", Status: "ignored"},
			want:    "status",
		},
		{
			name:    "reason on an open finding",
			finding: review.Finding{Path: "a.go", Severity: review.Warning, Why: "algo", Reason: "porque si"},
			want:    "reason",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			verdict := review.Verdict{
				Tree:     tree,
				Reviewer: "agent",
				Verdict:  review.Approved,
				Findings: []review.Finding{testCase.finding},
			}

			err := verdict.Validate()
			if err == nil {
				t.Fatal("the finding must be refused")
			}
			if !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("the error must name %q, got: %s", testCase.want, err)
			}
		})
	}
}

// TestADismissalRoundTrips: the reason and the fingerprint travel with the
// finding, because the next candidate reads them to avoid re-raising it.
func TestADismissalRoundTrips(t *testing.T) {
	finding := dismissal("pierde el turno", "el turno se guarda en la transaccion, la cita lo prueba")
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{finding},
	}
	if err := verdict.Validate(); err != nil {
		t.Fatalf("a complete dismissal must validate, got: %s", err)
	}

	parsed, err := review.Parse(review.Render(verdict))
	if err != nil {
		t.Fatalf("render then parse must round trip, got: %s", err)
	}
	got := parsed.Findings[0]
	if got.Status != review.Dismissed || got.Reason != finding.Reason || got.Fingerprint != finding.Fingerprint {
		t.Errorf("finding = %+v, want the dismissal with its reason and fingerprint", got)
	}
}

// TestAFingerprintIsContentIdentity: two descriptions of the same defect at the
// same place hash the same, whatever the spacing or the case; a different
// place, lens or evidence is a different defect.
func TestAFingerprintIsContentIdentity(t *testing.T) {
	base := review.Finding{Path: "internal/app/turnos.go", Line: 12, Lens: "reliability", Why: "pierde el turno"}

	same := review.Finding{Path: "internal/app/turnos.go", Line: 12, Lens: "RELIABILITY", Why: "  pierde   el   turno "}
	if review.Fingerprint(base) != review.Fingerprint(same) {
		t.Error("the same defect in another spelling must hash the same")
	}

	other := []review.Finding{
		{Path: "internal/app/turnos.go", Line: 13, Lens: "reliability", Why: "pierde el turno"},
		{Path: "internal/app/otro.go", Line: 12, Lens: "reliability", Why: "pierde el turno"},
		{Path: "internal/app/turnos.go", Line: 12, Lens: "risk", Why: "pierde el turno"},
		{Path: "internal/app/turnos.go", Line: 12, Lens: "reliability", Why: "pierde el turno al cancelar"},
	}
	for _, finding := range other {
		if review.Fingerprint(base) == review.Fingerprint(finding) {
			t.Errorf("%+v must not hash like %+v", finding, base)
		}
	}

	if got := review.Fingerprint(base); len(got) != 16 {
		t.Errorf("fingerprint = %q, want 16 hex characters", got)
	}
}

// TestADismissedFindingNeverBlocks: the dismissal was judged with evidence, so
// the row is recorded and never requests changes — not even at CRITICAL.
func TestADismissedFindingNeverBlocks(t *testing.T) {
	verdict := review.Verdict{
		Tree:     tree,
		Reviewer: "agent",
		Verdict:  review.Approved,
		Findings: []review.Finding{dismissal("pierde el turno", "la transaccion lo cubre, la cita lo prueba")},
	}
	if err := verdict.Validate(); err != nil {
		t.Fatalf("the dismissed verdict must validate: %s", err)
	}

	if got := review.Derive(verdict.Findings); got != review.Approved {
		t.Errorf("derived = %q, want approved: a dismissal never blocks", got)
	}
	if finding, blocking := review.Blocking(verdict.Findings); blocking {
		t.Errorf("blocking = %+v, want none", finding)
	}
}
