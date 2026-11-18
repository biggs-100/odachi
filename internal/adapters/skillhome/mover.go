// Package skillhome takes a skill out of the home for the length of one exam and
// puts it back. It exists because the client reads `~/.agents/skills` and offers no
// switch to leave a folder out: its configuration only adds paths, so the arm
// without the skill is the folder absent. The move is loud and reversible -the
// journal is written before the run and read by the next one- because a
// measurement that eats a skill is worse than no measurement.
package skillhome

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// Mover is the home's skills folder and the journal that remembers what left it.
type Mover struct {
	// Skills is the folder the client reads, `~/.agents/skills`.
	Skills string
	// Journal is the file in the clone's store that survives a run that died.
	Journal string
}

// hidden is the name a taken skill carries while the run measures. It is a
// sibling of the original so the move stays on one volume: a rename across
// volumes is a copy, and a copy that fails halfway leaves two folders and no
// answer.
const hidden = ".odachi-eval-"

// Take moves the named skill aside and answers the marks the journal now holds.
// Nothing is measured before the journal names the move: the order is what makes
// a crash recoverable.
func (m Mover) Take(name string) (taken skilleval.Mark, problem error) {
	from := filepath.Join(m.Skills, name)
	to := filepath.Join(m.Skills, hidden+name)
	if _, err := os.Stat(to); err == nil {
		return skilleval.Mark{}, fmt.Errorf("%s is already there: an earlier run took the skill out and did not put it back; run the verb with --recover first", to)
	} else if !errors.Is(err, os.ErrNotExist) {
		return skilleval.Mark{}, fmt.Errorf("could not look at %s: %w", to, err)
	}

	if _, err := os.Stat(from); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return skilleval.Mark{}, fmt.Errorf("%s: %w", from, skilleval.ErrNothingToTake)
		}
		return skilleval.Mark{}, fmt.Errorf("could not look at %s: %w", from, err)
	}

	mark := skilleval.Mark{Skill: name, From: from, To: to}
	marks, err := m.read()
	if err != nil {
		return skilleval.Mark{}, err
	}
	// The journal is written first: a move with no mark is a skill nobody will
	// put back, and a mark with no move is one the next run finds already home.
	if err := m.write(append(marks, mark)); err != nil {
		return skilleval.Mark{}, err
	}
	if err := os.Rename(from, to); err != nil {
		// The move failed and the mark is a lie: drop it, or the next run reports
		// a skill taken out that never left.
		_ = m.write(marks)
		return skilleval.Mark{}, fmt.Errorf("could not move %s out of the way: %w", from, err)
	}
	return mark, nil
}

// Restore puts one taken skill back and drops its mark. A skill already home is a
// success, not an error: the run may have restored it before dying, and the guest
// that comes back twice is the same guest.
func (m Mover) Restore(mark skilleval.Mark) error {
	marks, err := m.read()
	if err != nil {
		return err
	}

	atHomeErr := errOf(os.Stat(mark.From))
	asideErr := errOf(os.Stat(mark.To))
	switch {
	case atHomeErr == nil && asideErr == nil:
		// Both: the move was copied, not renamed, or a second folder appeared.
		return fmt.Errorf("%s and %s both exist: the skill was copied, not moved, and this tool will not guess which one is yours", mark.From, mark.To)
	case asideErr == nil:
		if err := os.Rename(mark.To, mark.From); err != nil {
			return fmt.Errorf("could not put %s back: %w", mark.From, err)
		}
	case atHomeErr == nil:
		// Already home: a run restored it and died before it cleared the journal.
	case errors.Is(atHomeErr, os.ErrNotExist) && errors.Is(asideErr, os.ErrNotExist):
		// Neither is there: the skill is gone and the journal cannot invent it.
		return fmt.Errorf("%s and %s are both missing: the skill %q is not on this machine any more", mark.From, mark.To, mark.Skill)
	default:
		return fmt.Errorf("could not look at %s or %s: %v; %v", mark.From, mark.To, atHomeErr, asideErr)
	}

	return m.write(without(marks, mark.Skill))
}

// Recover puts back everything a dead run left out and answers what came home.
// It is the first thing a run does: a journal with marks is a run that ended
// before its restore, and measuring on top of it would take the skill out twice.
func (m Mover) Recover() ([]string, error) {
	marks, err := m.read()
	if err != nil {
		return nil, err
	}
	var back []string
	for _, mark := range marks {
		if err := m.Restore(mark); err != nil {
			return back, err
		}
		back = append(back, mark.Skill)
	}
	return back, nil
}

// Installed answers where the named skill lives in the home, and whether it is
// there at all. It reads; it does not move.
func (m Mover) Installed(name string) (string, bool, error) {
	path := filepath.Join(m.Skills, name)
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return path, true, nil
	case errors.Is(err, os.ErrNotExist):
		return path, false, nil
	default:
		return path, false, fmt.Errorf("could not look at %s: %w", path, err)
	}
}

// Pending answers the skills a previous run left out, without touching them.
func (m Mover) Pending() ([]string, error) {
	marks, err := m.read()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(marks))
	for _, mark := range marks {
		out = append(out, mark.Skill)
	}
	return out, nil
}

func (m Mover) read() ([]skilleval.Mark, error) {
	data, err := os.ReadFile(m.Journal)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read the journal %s: %w", m.Journal, err)
	}
	return skilleval.ParseJournal(data)
}

func (m Mover) write(marks []skilleval.Mark) error {
	data, err := skilleval.RenderJournal(marks)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		if err := os.Remove(m.Journal); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("could not clear the journal %s: %w", m.Journal, err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(m.Journal), 0o755); err != nil {
		return fmt.Errorf("could not make room for the journal %s: %w", m.Journal, err)
	}
	if err := os.WriteFile(m.Journal, data, 0o644); err != nil {
		return fmt.Errorf("could not write the journal %s: %w", m.Journal, err)
	}
	return nil
}

// errOf answers the error a Stat left, and nil when the path is there: the three
// answers a restore needs are "here", "gone" and "the disk said something else".
func errOf(_ os.FileInfo, err error) error { return err }

func without(marks []skilleval.Mark, skill string) []skilleval.Mark {
	out := make([]skilleval.Mark, 0, len(marks))
	for _, mark := range marks {
		if mark.Skill != skill {
			out = append(out, mark)
		}
	}
	return out
}
