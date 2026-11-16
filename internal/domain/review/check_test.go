package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestNormalizeChecksIsTheOrderTheFileWrote: the checks run in the order the
// project declared them, because the first failing tool a human reads is the
// first one they wrote. The name is the identity the row and the finding carry,
// so it arrives trimmed and lowercased and the command arrives as typed.
func TestNormalizeChecksIsTheOrderTheFileWrote(t *testing.T) {
	got, err := review.NormalizeChecks([]review.Check{
		{Name: " Lint ", Command: " golangci-lint run ./... ", Severity: "warning", Remedy: "fix it"},
		{Name: "vuln", Command: "govulncheck ./...", Severity: review.Critical, Remedy: "fix it"},
	})
	if err != nil {
		t.Fatalf("NormalizeChecks = %v, want no error", err)
	}
	if len(got) != 2 {
		t.Fatalf("NormalizeChecks kept %d checks, want 2", len(got))
	}
	if got[0].Name != "lint" || got[0].Command != "golangci-lint run ./..." || got[0].Severity != review.Warning {
		t.Errorf("first check = %+v, want the name lowercased, the command trimmed and WARNING", got[0])
	}
	if got[1].Name != "vuln" || got[1].Severity != review.Critical {
		t.Errorf("second check = %+v, want vuln at CRITICAL, in the order the file wrote", got[1])
	}
}

// TestNormalizeChecksRefusesACheckTheReviewCannotRun: a check without a name, a
// command or a severity is a line that costs a run and answers nothing.
func TestNormalizeChecksRefusesACheckTheReviewCannotRun(t *testing.T) {
	cases := map[string]review.Check{
		"no name":     {Command: "go vet ./...", Severity: review.Warning, Remedy: "fix it"},
		"no command":  {Name: "vet", Severity: review.Warning, Remedy: "fix it"},
		"no severity": {Name: "vet", Command: "go vet ./...", Remedy: "fix it"},
		"bad severity": {
			Name: "vet", Command: "go vet ./...", Severity: "SERIOUS", Remedy: "fix it",
		},
		"name with a space": {
			Name: "go vet", Command: "go vet ./...", Severity: review.Warning, Remedy: "fix it",
		},
		"no remedy": {
			Name: "vet", Command: "go vet ./...", Severity: review.Warning,
		},
	}

	for name, check := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := review.NormalizeChecks([]review.Check{check}); err == nil {
				t.Fatalf("NormalizeChecks(%+v) = %+v, want a refusal", check, got)
			}
		})
	}
}

// TestTheRefusalNamesWhatTheRemedyIsFor: a check whose red names no repair is the
// one row that teaches nothing, so the refusal says what the key is for.
func TestTheRefusalNamesWhatTheRemedyIsFor(t *testing.T) {
	_, err := review.NormalizeChecks([]review.Check{
		{Name: "vet", Command: "go vet ./...", Severity: review.Warning},
	})
	if err == nil {
		t.Fatal("NormalizeChecks accepted a check with no remedy")
	}
	for _, want := range []string{"remedy", "guessing"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not say %q", err, want)
		}
	}
}

// TestAnOldFileOwesNoRemedy: the requirement belongs to the schema that made it,
// so the same row is refused at 10 and accepted at 9. Without this the reader
// would refuse files its own writer had produced one release earlier (0312).
func TestAnOldFileOwesNoRemedy(t *testing.T) {
	raw := []review.Check{{Name: "vet", Command: "go vet ./...", Severity: review.Warning}}

	if got, err := review.NormalizeChecksAt(raw, review.RemedySchema-1); err != nil {
		t.Fatalf("NormalizeChecksAt(9) = %v, want the old file to keep reading: %+v", err, got)
	}
	if _, err := review.NormalizeChecksAt(raw, review.RemedySchema); err == nil {
		t.Fatal("NormalizeChecksAt(10) accepted a check with no remedy")
	}
}

// TestTheRefusalNamesTheSeverityVocabulary: the human who wrote a word odachi does
// not know must read the four it does.
func TestTheRefusalNamesTheSeverityVocabulary(t *testing.T) {
	_, err := review.NormalizeChecks([]review.Check{
		{Name: "vet", Command: "go vet ./...", Severity: "SERIOUS", Remedy: "fix it"},
	})
	if err == nil {
		t.Fatal("NormalizeChecks accepted a severity nobody knows")
	}
	for _, want := range []string{"BLOCKER", "CRITICAL", "WARNING", "SUGGESTION"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not name %q", err, want)
		}
	}
}

// TestTwoChecksWithOneNameAreRefused: the name is the identity a row is filed
// under, so two tools that share it would pay one row and lose the other.
func TestTwoChecksWithOneNameAreRefused(t *testing.T) {
	_, err := review.NormalizeChecks([]review.Check{
		{Name: "lint", Command: "golangci-lint run ./...", Severity: review.Warning, Remedy: "fix it"},
		{Name: "LINT", Command: "gofmt -l .", Severity: review.Suggestion, Remedy: "fix it"},
	})
	if err == nil {
		t.Fatal("NormalizeChecks accepted two checks with one name")
	}
	if !strings.Contains(err.Error(), "lint") {
		t.Errorf("the refusal %q does not name the check written twice", err)
	}
}

// TestAnEmptyCheckListIsNoChecks: a project that declares none runs none, which
// is the absence of the feature and not an error.
func TestAnEmptyCheckListIsNoChecks(t *testing.T) {
	got, err := review.NormalizeChecks(nil)
	if err != nil {
		t.Fatalf("NormalizeChecks(nil) = %v, want no error", err)
	}
	if len(got) != 0 {
		t.Errorf("NormalizeChecks(nil) = %+v, want none", got)
	}
}

// TestCheckTimeoutIsACeilingOrTheDefault: a tool that hangs must not hang a
// review, so the project may lower it; zero is the absence of a choice and means
// the ceiling odachi ships.
func TestCheckTimeoutIsACeilingOrTheDefault(t *testing.T) {
	if got, err := review.NormalizeCheckTimeout(0); err != nil || got != review.DefaultCheckTimeout {
		t.Errorf("NormalizeCheckTimeout(0) = %d, %v; want the default", got, err)
	}
	if got, err := review.NormalizeCheckTimeout(45); err != nil || got != 45 {
		t.Errorf("NormalizeCheckTimeout(45) = %d, %v; want 45", got, err)
	}
	for _, bad := range []int{-1, review.MaxCheckTimeout + 1} {
		if _, err := review.NormalizeCheckTimeout(bad); err == nil {
			t.Errorf("NormalizeCheckTimeout(%d) accepted a ceiling that cannot be one", bad)
		}
	}
}
