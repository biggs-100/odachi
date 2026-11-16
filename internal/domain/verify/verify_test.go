package verify_test

import (
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/domain/verify"
)

const tree = "0123456789abcdef0123456789abcdef01234567"

// TestARecordWithoutAnAuthorIsRefused: a verification nobody signs is a sentence
// about the tree, not evidence about who read it, and the rule is about who.
func TestARecordWithoutAnAuthorIsRefused(t *testing.T) {
	record := verify.Record{Tree: tree, Verdict: verify.Verified, At: "2026-11-16T10:00:00Z"}

	err := record.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want a refusal that names the missing author")
	}
	if !strings.Contains(err.Error(), "who") {
		t.Errorf("the refusal does not say who is missing: %v", err)
	}
}

// TestARecordForAnotherTreeDoesNotCoverThisOne: the tree is the name, so a record
// of a tree that moved cannot speak for the tree that closes.
func TestARecordForAnotherTreeDoesNotCoverThisOne(t *testing.T) {
	other := "fedcba9876543210fedcba9876543210fedcba98"
	record := verify.Record{Tree: other, By: "otro/modelo", Verdict: verify.Verified, At: "2026-11-16T10:00:00Z"}

	if record.Covers(tree) {
		t.Errorf("a record for %s covers %s", other, tree)
	}
	if !record.Covers(other) {
		t.Errorf("a record for %s does not cover itself", other)
	}
}

// TestSealWritesTheClockAndKeepsTheClaimedTime: the hour is the tool's, and what
// the verifier wrote about its own time stays beside it as a claim (0286).
func TestSealWritesTheClockAndKeepsTheClaimedTime(t *testing.T) {
	record := verify.Record{By: "otro/modelo", Verdict: verify.Verified, At: "1999-01-01T00:00:00Z"}

	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	sealed := record.Seal(now)

	if sealed.At != "2026-11-16T12:00:00Z" {
		t.Errorf("At = %q, want the tool's clock", sealed.At)
	}
	if sealed.DeclaredAt != "1999-01-01T00:00:00Z" {
		t.Errorf("DeclaredAt = %q, want the verifier's claim kept", sealed.DeclaredAt)
	}
	if record.At != "1999-01-01T00:00:00Z" {
		t.Errorf("Seal changed the caller's record: At = %q", record.At)
	}
}

// TestARecordWithAnUnknownFieldIsRefused: the decoder is strict, because a file
// half understood would be stored under a name that claims more than it can back.
func TestARecordWithAnUnknownFieldIsRefused(t *testing.T) {
	data := []byte(`{"schema":1,"tree":"` + tree + `","by":"otro/modelo","at":"2026-11-16T10:00:00Z","verdict":"verified","notes":[],"extra":true}`)

	if _, err := verify.Parse(data); err == nil {
		t.Fatal("Parse() = nil, want a refusal that names the unknown field")
	}
}

// TestAVerdictOutsideTheVocabularyIsRefused: verified or refuted are the two
// answers, and a third word is not a verification.
func TestAVerdictOutsideTheVocabularyIsRefused(t *testing.T) {
	record := verify.Record{Tree: tree, By: "otro/modelo", Verdict: "looks fine", At: "2026-11-16T10:00:00Z"}

	if err := record.Validate(); err == nil {
		t.Fatal("Validate() = nil, want a refusal that names the two answers")
	}
}

// TestThePathIsTheTree: the file is named by the tree it verified, in one spelling
// on every platform, and a value that is not a hash is refused.
func TestThePathIsTheTree(t *testing.T) {
	path, err := verify.Path(tree)
	if err != nil {
		t.Fatalf("Path(%q) = %v", tree, err)
	}
	if path != verify.Dir+"/"+tree+".json" {
		t.Errorf("Path = %q, want the tree under %s", path, verify.Dir)
	}
	if _, err := verify.Path("not-a-hash"); err == nil {
		t.Error("Path accepted a value that is not a hash")
	}
}
