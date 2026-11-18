package skilleval_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// TestTheIdentityNamesTheBytesAndNothingElse: the refusal to pay twice rests on
// this hash, so the same text must answer the same name and any change must
// answer a different one.
func TestTheIdentityNamesTheBytesAndNothingElse(t *testing.T) {
	t.Parallel()

	files := map[string][]byte{"SKILL.md": []byte("rules")}
	same := skilleval.Identity("tests", files, "the task")
	if again := skilleval.Identity("tests", files, "the task"); again != same {
		t.Fatalf("the same text answered %q and %q", same, again)
	}
	if changed := skilleval.Identity("tests", map[string][]byte{"SKILL.md": []byte("other rules")}, "the task"); changed == same {
		t.Errorf("a changed skill answered the same identity %q", same)
	}
	if other := skilleval.Identity("tests", files, "another task"); other == same {
		t.Errorf("a changed exam answered the same identity %q", same)
	}
}

// TestARecordRoundTrips: what the run files is what the next run reads, or the
// refusal names bytes nobody measured.
func TestARecordRoundTrips(t *testing.T) {
	t.Parallel()

	want := skilleval.Record{
		Schema:   skilleval.RecordSchema,
		Skill:    "tests",
		Identity: "abc123",
		At:       "2026-11-18T00:00:00Z",
		Draws: []skilleval.Draw{{
			Arm: "with",
			Run: 1,
			Behaviours: []skilleval.Behaviour{{
				ID: skilleval.TestBeforeCode, Label: "the test was written before the code", Met: true, Why: "yes",
			}},
		}},
	}
	data, err := skilleval.RenderRecord(want)
	if err != nil {
		t.Fatalf("RenderRecord: %v", err)
	}
	got, err := skilleval.ParseRecord(data)
	if err != nil {
		t.Fatalf("ParseRecord: %v", err)
	}
	if got.Skill != want.Skill || got.Identity != want.Identity || len(got.Draws) != 1 {
		t.Fatalf("record = %+v, want %+v", got, want)
	}
	if !strings.Contains(got.Summarize(), "1 of 1") {
		t.Errorf("the summary does not count the behaviours: %s", got.Summarize())
	}
}

// TestARecordWithoutItsIdentityIsRefused: a record that does not say what it
// measured cannot refuse anything, and guessing on its behalf is what the file
// exists to end.
func TestARecordWithoutItsIdentityIsRefused(t *testing.T) {
	t.Parallel()

	if _, err := skilleval.ParseRecord([]byte(`{"skill":"tests"}`)); err == nil {
		t.Fatal("ParseRecord accepted a record with no identity")
	}
	if _, err := skilleval.ParseRecord([]byte("not json")); err == nil {
		t.Fatal("ParseRecord accepted a file that is not JSON")
	}
}
