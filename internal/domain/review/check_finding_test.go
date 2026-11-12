package review_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// checkRow is one failing tool as the verdict carries it: a name, no path, and the
// tool's own words as the evidence.
func checkRow() review.Finding {
	return review.Finding{
		Check:    "lint",
		Severity: review.Critical,
		Causal:   true,
		Evidence: review.EvidenceCited,
		Why:      "the check lint exited 1\n",
	}
}

// TestACheckRowIsNotAPath: the row a tool files names the tool, never a file it did
// not read. A verdict that invented a path would send a person to fix the wrong
// file.
func TestACheckRowIsNotAPath(t *testing.T) {
	verdict := review.Verdict{Reviewer: "odachi", Findings: []review.Finding{checkRow()}}
	if err := verdict.Validate(); err != nil {
		t.Fatalf("Validate = %v, want a check row to be a finding", err)
	}

	nameless := review.Verdict{Reviewer: "odachi", Findings: []review.Finding{{
		Severity: review.Critical,
		Causal:   true,
		Why:      "something is wrong",
	}}}
	if err := nameless.Validate(); err == nil {
		t.Fatal("Validate accepted a finding with neither a path nor a check")
	}
}

// TestACheckRowNamesNoLine: a tool's output is not a place in the diff, so a line
// beside a check is a claim nobody measured.
func TestACheckRowNamesNoLine(t *testing.T) {
	row := checkRow()
	row.Line = 3

	verdict := review.Verdict{Reviewer: "odachi", Findings: []review.Finding{row}}
	if err := verdict.Validate(); err == nil {
		t.Fatal("Validate accepted a check row that claims a line")
	}
}

// TestACheckRowSurvivesTheStore: the name travels with the verdict, so a reader a
// week later knows what failed instead of reading a finding about nowhere.
func TestACheckRowSurvivesTheStore(t *testing.T) {
	sealed, err := review.Seal(review.Verdict{
		Reviewer: "odachi",
		Findings: []review.Finding{checkRow()},
	}, "abc1234", "def5678", now)
	if err != nil {
		t.Fatalf("Seal = %v, want a check row to be stored", err)
	}

	back, err := review.Parse(review.Render(sealed))
	if err != nil {
		t.Fatalf("Parse(Render(sealed)) = %v, want no error", err)
	}
	if len(back.Findings) != 1 || back.Findings[0].Check != "lint" {
		t.Fatalf("the stored verdict lost the check's name: %+v", back.Findings)
	}
}

// TestACheckFailureRequestsChangesLikeAnyOther: a tool that fails at CRITICAL
// blocks, and that is the whole point: the model cannot approve past it.
func TestACheckFailureRequestsChangesLikeAnyOther(t *testing.T) {
	if got := review.Derive([]review.Finding{checkRow()}); got != review.ChangesRequested {
		t.Errorf("Derive = %q, want %q", got, review.ChangesRequested)
	}

	warned := checkRow()
	warned.Severity = review.Warning
	if got := review.Derive([]review.Finding{warned}); got != review.Approved {
		t.Errorf("Derive = %q for a warning, want %q", got, review.Approved)
	}
}

// TestTheFloorLeavesACheckRowAlone: the floor judges what a finding may cite, and
// a check row cites a tool's output. Downgrading it would let a tree pass because
// a linter is not a line of the diff.
func TestTheFloorLeavesACheckRowAlone(t *testing.T) {
	floor := review.FloorFor([]string{"a.go"}, "+++ b/a.go\n@@ -0,0 +1 @@\n+x := 1\n")

	earned, downgraded := review.Earned([]review.Finding{checkRow()}, floor)
	if downgraded != 0 {
		t.Fatalf("Earned downgraded %d rows, want none", downgraded)
	}
	if earned[0].Severity != review.Critical {
		t.Errorf("severity = %q, want CRITICAL kept", earned[0].Severity)
	}
}

// TestTwoChecksWithOneWhyAreTwoDefects: the name is part of the identity, so a
// linter and a vulnerability scanner that print the same sentence are two rows and
// dismissing one does not dismiss the other.
func TestTwoChecksWithOneWhyAreTwoDefects(t *testing.T) {
	lint := checkRow()
	vuln := checkRow()
	vuln.Check = "vuln"

	if review.Fingerprint(lint) == review.Fingerprint(vuln) {
		t.Error("two checks that failed the same way share one fingerprint")
	}
}

// TestAnOrdinaryFindingKeepsItsIdentity: the check's name is part of the identity
// only when there is one, so every dismissal a store already holds still names the
// same defect.
func TestAnOrdinaryFindingKeepsItsIdentity(t *testing.T) {
	row := review.Finding{Path: "a.go", Line: 3, Lens: "risk", Why: "x"}

	sum := sha256.Sum256([]byte("a.go|3|risk|x"))
	want := hex.EncodeToString(sum[:])[:16]
	if got := review.Fingerprint(row); got != want {
		t.Errorf("Fingerprint = %q, want %q: an old row must hash as it did", got, want)
	}
}

// TestACheckRowCanBeDismissed: the way out of a false positive is a recorded
// reason, exactly like every other row.
func TestACheckRowCanBeDismissed(t *testing.T) {
	dismissed := review.Dismiss(checkRow(), "the ruleset flags a pattern this file must keep")
	if !review.IsDismissed(dismissed) {
		t.Fatal("Dismiss did not mark the row")
	}
	if strings.TrimSpace(dismissed.Reason) == "" || dismissed.Fingerprint == "" {
		t.Errorf("the dismissal carries no reason or no fingerprint: %+v", dismissed)
	}
	if got := review.Derive([]review.Finding{dismissed}); got != review.Approved {
		t.Errorf("Derive = %q, want %q: a dismissed row never blocks", got, review.Approved)
	}
}
