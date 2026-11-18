package skillhome_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/skillhome"
	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// home is a skills folder with the named skills in it and a journal elsewhere, the
// way the clone's store holds it.
func home(t *testing.T, names ...string) skillhome.Mover {
	t.Helper()

	root := t.TempDir()
	for _, name := range names {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("make %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(name+"\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", dir, err)
		}
	}
	return skillhome.Mover{Skills: root, Journal: filepath.Join(t.TempDir(), "evals", "home.json")}
}

// TestTakingASkillWritesTheJournalBeforeItMoves: the mark is what a run that dies
// leaves behind, so it is on disk before the folder moves. A move with no mark is
// a skill nobody knows how to put back.
func TestTakingASkillWritesTheJournalBeforeItMoves(t *testing.T) {
	m := home(t, "tests")

	mark, err := m.Take("tests")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}

	if _, err := os.Stat(mark.From); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s is still there after Take: %v", mark.From, err)
	}
	if _, err := os.Stat(mark.To); err != nil {
		t.Errorf("%s is not there after Take: %v", mark.To, err)
	}
	pending, err := m.Pending()
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	if len(pending) != 1 || pending[0] != "tests" {
		t.Fatalf("pending = %v, want [tests]", pending)
	}
}

// TestRestoringPutsTheSkillBackAndClearsTheJournal: the run ends and the folder is
// where it was, with the journal empty -no file- so the next run measures from a
// clean state.
func TestRestoringPutsTheSkillBackAndClearsTheJournal(t *testing.T) {
	m := home(t, "tests")
	mark, err := m.Take("tests")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}

	if err := m.Restore(mark); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if _, err := os.Stat(filepath.Join(mark.From, "SKILL.md")); err != nil {
		t.Errorf("the skill is not back: %v", err)
	}
	if _, err := os.Stat(mark.To); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s is still there after Restore: %v", mark.To, err)
	}
	if pending, _ := m.Pending(); len(pending) != 0 {
		t.Errorf("pending = %v after a restore, want none", pending)
	}
}

// TestRecoverPutsBackWhatADeadRunLeftOut: the journal is the only thing that
// survives a crash, so the next run reads it first and returns the skill before
// it measures anything.
func TestRecoverPutsBackWhatADeadRunLeftOut(t *testing.T) {
	m := home(t, "tests")
	mark, err := m.Take("tests")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}
	// The run dies here: the journal says the skill is out, and nothing restores it.

	back, err := m.Recover()
	if err != nil {
		t.Fatalf("Recover: %v", err)
	}
	if len(back) != 1 || back[0] != "tests" {
		t.Fatalf("Recover = %v, want [tests]", back)
	}
	if _, err := os.Stat(filepath.Join(mark.From, "SKILL.md")); err != nil {
		t.Errorf("the skill did not come home: %v", err)
	}
	if pending, _ := m.Pending(); len(pending) != 0 {
		t.Errorf("pending = %v after a recover, want none", pending)
	}
}

// TestTakingASkillThatIsNotThereNamesTheFolder: the caller asked for a skill the
// home does not have, and the answer says where it looked instead of moving on.
func TestTakingASkillThatIsNotThereNamesTheFolder(t *testing.T) {
	m := home(t, "tests")

	_, err := m.Take("absent")
	if !errors.Is(err, skilleval.ErrNothingToTake) {
		t.Fatalf("Take(absent) = %v, want ErrNothingToTake", err)
	}
	if err == nil || !strings.Contains(err.Error(), filepath.Join(m.Skills, "absent")) {
		t.Errorf("the refusal does not name the folder it looked at: %v", err)
	}
}

// TestASecondTakeRefusesInsteadOfNesting: a skill already aside is a run that did
// not restore. Taking it again would hide the first folder inside the second and
// the restore would bring back a name nobody can find.
func TestASecondTakeRefusesInsteadOfNesting(t *testing.T) {
	m := home(t, "tests")
	if _, err := m.Take("tests"); err != nil {
		t.Fatalf("the first Take: %v", err)
	}

	_, err := m.Take("tests")
	if err == nil {
		t.Fatal("the second Take moved a skill that was already aside")
	}
	if !strings.Contains(err.Error(), "--recover") {
		t.Errorf("the refusal does not say how to get out: %v", err)
	}
}

// TestRestoringASkillThatIsAlreadyHomeIsNotAnError: a run may restore and then die
// before it clears the journal. Coming home twice is the same state, and the
// second restore has nothing to do.
func TestRestoringASkillThatIsAlreadyHomeIsNotAnError(t *testing.T) {
	m := home(t, "tests")
	mark, err := m.Take("tests")
	if err != nil {
		t.Fatalf("Take: %v", err)
	}
	if err := os.Rename(mark.To, mark.From); err != nil {
		t.Fatalf("put the skill home by hand: %v", err)
	}

	if err := m.Restore(mark); err != nil {
		t.Fatalf("Restore on a skill already home = %v, want no error", err)
	}
}
