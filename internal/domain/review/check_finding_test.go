package review_test

import (
	"bytes"
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

// TestEffectiveCheckFindingsUsesOnlyHumanDismissals: reviewer JSON cannot retire
// a deterministic row, while the separate human sidecar can. Path findings keep
// their existing verdict dismissal semantics.
func TestEffectiveCheckFindingsUsesOnlyHumanDismissals(t *testing.T) {
	forged := checkRow()
	forged.Status = review.Dismissed
	forged.Reason = "reviewer-forged waiver"
	forged.Fingerprint = review.Fingerprint(forged)
	path := review.Finding{Path: "a.go", Line: 1, Severity: review.Warning, Why: "a path warning"}
	path = review.Dismiss(path, "the reviewer's cited judgment")

	effective := review.EffectiveCheckFindings([]review.Finding{forged, path}, review.Dismissals{})
	if effective[0].Status != "" || effective[0].Reason != "" || effective[0].Fingerprint != "" {
		t.Errorf("verdict JSON retired the check row: %+v", effective[0])
	}
	if !review.IsDismissed(effective[1]) || effective[1].Reason != path.Reason {
		t.Errorf("path finding changed: %+v", effective[1])
	}

	human := review.Dismissals{Rows: []review.Dismissal{{
		Fingerprint: review.Fingerprint(forged), Check: forged.Check, Reason: "human current waiver",
	}}}
	effective = review.EffectiveCheckFindings([]review.Finding{forged}, human)
	if !review.IsDismissed(effective[0]) || effective[0].Reason != "human current waiver" {
		t.Errorf("human sidecar did not retire the check row: %+v", effective[0])
	}
}

// TestVerdictJSONRejectsCheckBlockedField: the rejected field never entered the
// persisted contract. Strict decoding keeps older and newer binaries on the same
// schema instead of accepting an execution marker they cannot understand.
func TestVerdictJSONRejectsCheckBlockedField(t *testing.T) {
	_, err := review.Parse([]byte(`{"findings":[{"path":"a.go","severity":"WARNING","causal":true,"why":"x","check_blocked":true}]}`))
	if err == nil {
		t.Fatal("Parse accepted the rejected check_blocked field")
	}
}

// TestUnexecutedCheckIsStoredAsSyntheticBlocker: missing approval is a missing
// measurement, regardless of the declared severity. The persisted row uses the
// existing BLOCKER severity and adds no schema field; only a human dismissal can
// retire it.
func TestUnexecutedCheckIsStoredAsSyntheticBlocker(t *testing.T) {
	for _, severity := range []review.Severity{review.Warning, review.Suggestion, review.Critical, review.Blocker} {
		t.Run(string(severity), func(t *testing.T) {
			row := checkRow()
			row.Severity = review.Blocker
			row.Why = "the check lint did not run: this clone has not approved it; declared severity " + string(severity)
			if got := review.Derive([]review.Finding{row}); got != review.ChangesRequested {
				t.Errorf("Derive = %q, want %q", got, review.ChangesRequested)
			}
			sealed, err := review.Seal(review.Verdict{Reviewer: "odachi", Findings: []review.Finding{row}}, "abc1234", "HEAD", now)
			if err != nil {
				t.Fatalf("Seal: %v", err)
			}
			rendered := review.Render(sealed)
			if bytes.Contains(rendered, []byte(`"check_blocked"`)) {
				t.Fatalf("the verdict added the rejected check_blocked field:\n%s", rendered)
			}
			back, err := review.Parse(rendered)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			stored := back.Findings[0]
			if stored.Severity != review.Blocker || !stored.Blocks() {
				t.Errorf("stored row lost its synthetic BLOCKER authority: %+v", stored)
			}
			if !strings.Contains(stored.Why, string(severity)) {
				t.Errorf("stored row lost its declared severity %q: %+v", severity, stored)
			}
			// Supply an otherwise valid counter-citation to prove that check-row
			// authority, not citation shape, prevents a refutation from retiring it.
			refutable := stored
			refutable.Path = "a.go"
			answer := review.Answers{{Tree: "abc1234", Refuter: "refuter", Refutations: []review.Judgment{{
				ID: "F1", Verdict: review.Refuted, Why: "a.go:1 candidate",
			}}}}
			if got := review.DeriveAfter([]review.Finding{refutable}, answer, review.CitationEvidence{Tree: "abc1234", Diff: "+++ b/a.go\n@@ -0,0 +1 @@\n+candidate\n"}); got != review.ChangesRequested {
				t.Errorf("a refutation lifted an unexecuted check: DeriveAfter = %q", got)
			}
			dismissed := review.Dismiss(stored, "human current waiver")
			if dismissed.Blocks() {
				t.Errorf("dismissed blocked check still blocks: %+v", dismissed)
			}
		})
	}
}

// TestEveryBlockingRowIsAnsweredOrTheGateStaysClosed: a verdict can carry more
// than one row that blocks, and a waiver answers one. The gate asks about all of
// them, so waiving the first cannot hide the second.
func TestEveryBlockingRowIsAnsweredOrTheGateStaysClosed(t *testing.T) {
	first := review.Finding{Path: "a.go", Line: 3, Severity: review.Critical, Causal: true, Lens: review.Reliability, Why: "the round trip is lost"}
	second := review.Finding{Path: "b.go", Line: 9, Severity: review.Blocker, Causal: true, Lens: review.Reliability, Why: "the turn is lost"}

	rows := review.BlockingAllAfter([]review.Finding{first, second}, review.Answers{}, review.CitationEvidence{})
	if len(rows) != 2 {
		t.Fatalf("BlockingAllAfter = %d rows, want both blockers", len(rows))
	}
	if got, ok := review.BlockingAfter([]review.Finding{first, second}, review.Answers{}, review.CitationEvidence{}); !ok || got.Why != first.Why {
		t.Errorf("BlockingAfter = %+v, %v; want the first row, unchanged", got, ok)
	}
}
