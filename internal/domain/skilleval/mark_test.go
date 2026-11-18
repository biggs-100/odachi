package skilleval_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// TestAnEmptyJournalIsNotAFailure: a clone that never ran an exam carries no
// file, and the first run has to read that as "nothing to put back", not as a
// broken journal.
func TestAnEmptyJournalIsNotAFailure(t *testing.T) {
	t.Parallel()

	marks, err := skilleval.ParseJournal(nil)
	if err != nil {
		t.Fatalf("ParseJournal(nil) = %v, want no error", err)
	}
	if len(marks) != 0 {
		t.Fatalf("ParseJournal(nil) = %d marks, want none", len(marks))
	}
}

// TestAMarkWithoutItsTwoPathsIsRefused: a mark that says which skill left but not
// where it went is a skill nobody can put back, and a journal that carries one is
// refused before a run trusts it.
func TestAMarkWithoutItsTwoPathsIsRefused(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`[{"skill":"tests"}]`,
		`[{"skill":"tests","from":"C:/home/.agents/skills/tests"}]`,
		`[{"from":"C:/home/.agents/skills/tests","to":"C:/home/.agents/skills/.odachi-eval-tests"}]`,
	} {
		if _, err := skilleval.ParseJournal([]byte(body)); err == nil {
			t.Errorf("ParseJournal(%s) = nil error, want a refusal", body)
		}
	}
}

// TestTheJournalRoundTrips: what the tool writes is what it reads. A journal that
// drifts between the two is a skill left out for good.
func TestTheJournalRoundTrips(t *testing.T) {
	t.Parallel()

	want := []skilleval.Mark{{
		Skill: "tests",
		From:  "C:/home/.agents/skills/tests",
		To:    "C:/home/.agents/skills/.odachi-eval-tests",
	}}
	data, err := skilleval.RenderJournal(want)
	if err != nil {
		t.Fatalf("RenderJournal: %v", err)
	}
	got, err := skilleval.ParseJournal(data)
	if err != nil {
		t.Fatalf("ParseJournal: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("journal = %+v, want %+v", got, want)
	}
}

// TestAnEmptyJournalRendersNothing: a cleared journal is no file at all, because
// a file with "[]" in it is a journal somebody has to explain.
func TestAnEmptyJournalRendersNothing(t *testing.T) {
	t.Parallel()

	data, err := skilleval.RenderJournal(nil)
	if err != nil {
		t.Fatalf("RenderJournal(nil): %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("RenderJournal(nil) = %q, want nothing", data)
	}
}

// TestAMarkWithAnEmptySkillIsRefused: the name is how a restore finds the folder,
// so a mark without one is a mark the next run cannot use.
func TestAMarkWithAnEmptySkillIsRefused(t *testing.T) {
	t.Parallel()

	if _, err := skilleval.ParseJournal([]byte(`[{"skill":"","from":"a","to":"b"}]`)); err == nil {
		t.Fatal("ParseJournal accepted a mark with no skill")
	}
}
