package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// approved is the list most of these tests approve.
var approved = []review.Check{
	{Name: "lint", Command: "golangci-lint run ./...", Severity: review.Warning},
	{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical},
}

// TestTheFingerprintIsTheListNotTheWords: two spellings of one list -different
// spacing, another key order in the file- describe the same tools and must be
// approved once.
func TestTheFingerprintIsTheListNotTheWords(t *testing.T) {
	first := review.ChecksFingerprint(approved)
	same := review.ChecksFingerprint([]review.Check{
		{Name: " Lint ", Command: " golangci-lint run ./... ", Severity: "warning"},
		{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical},
	})
	if first != same {
		t.Errorf("the same list has two fingerprints: %q and %q", first, same)
	}
	other := review.ChecksFingerprint(approved[:1])
	if first == other {
		t.Error("a shorter list has the same fingerprint as the whole one")
	}
}

// TestOneWordChangedIsAnotherList: approving a linter is not approving whatever
// command somebody puts in its place later.
func TestOneWordChangedIsAnotherList(t *testing.T) {
	changed := []review.Check{
		{Name: "lint", Command: "golangci-lint run --fix ./...", Severity: review.Warning},
		{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical},
	}
	if review.ChecksFingerprint(approved) == review.ChecksFingerprint(changed) {
		t.Error("a command that gained --fix keeps the approval of the one that did not")
	}

	raised := []review.Check{
		{Name: "lint", Command: "golangci-lint run ./...", Severity: review.Critical},
		{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical},
	}
	if review.ChecksFingerprint(approved) == review.ChecksFingerprint(raised) {
		t.Error("a check that became CRITICAL keeps the approval of the one that was a warning")
	}
}

// TestAnApprovalCoversTheListItWasWrittenFor: the record survives a reread and
// stops covering the day the list changes.
func TestAnApprovalCoversTheListItWasWrittenFor(t *testing.T) {
	approval := review.NewCheckApproval(approved, "biggs", "2026-11-12T09:00:00Z")
	if !approval.Covers(approved) {
		t.Error("the approval does not cover the list it was written for")
	}
	if approval.Covers(approved[:1]) {
		t.Error("the approval covers a list it never saw")
	}
	if approval.Count != 2 || approval.By != "biggs" {
		t.Errorf("the approval says %d tools by %q, want 2 by biggs", approval.Count, approval.By)
	}
}

// TestTheApprovalRoundTrips: the clone's store keeps the record, and the record
// says what it approved, when and by whom.
func TestTheApprovalRoundTrips(t *testing.T) {
	approval := review.NewCheckApproval(approved, "biggs", "2026-11-12T09:00:00Z")

	back, err := review.ParseCheckApproval(review.RenderCheckApproval(approval))
	if err != nil {
		t.Fatalf("ParseCheckApproval(Render(...)) = %v, want no error", err)
	}
	if back != approval {
		t.Errorf("the approval came back as %+v, want %+v", back, approval)
	}
}

// TestAnApprovalThatCannotBeReadBlocks: a record this binary cannot parse is not
// "never approved" -it is a file to look at, and reading it as absence would run
// nothing while the human believes the tools run.
func TestAnApprovalThatCannotBeReadBlocks(t *testing.T) {
	for name, data := range map[string]string{
		"no fingerprint": "# at 2026-11-12T09:00:00Z\n",
		"bad count":      "# fingerprint " + review.ChecksFingerprint(approved) + "\n# count many\n",
		"no count":       "# fingerprint " + review.ChecksFingerprint(approved) + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := review.ParseCheckApproval([]byte(data)); err == nil {
				t.Fatalf("ParseCheckApproval(%q) = no error, want a refusal", data)
			}
		})
	}
}

// TestAnEmptyApprovalIsNoTools: a project that declared none is approved by
// definition, and the record says zero.
func TestAnEmptyApprovalIsNoTools(t *testing.T) {
	if got := review.ChecksFingerprint(nil); got != "" {
		t.Errorf("ChecksFingerprint(nil) = %q, want empty", got)
	}
	if !review.NewCheckApproval(nil, "biggs", "2026-11-12T09:00:00Z").Covers(nil) {
		t.Error("an empty list is not covered by its own approval")
	}
	if strings.TrimSpace(string(review.RenderCheckApproval(review.NewCheckApproval(nil, "biggs", "2026-11-12T09:00:00Z")))) == "" {
		t.Error("an empty approval renders nothing, so it cannot be told from no record")
	}
}
