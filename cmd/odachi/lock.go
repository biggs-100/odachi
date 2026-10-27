package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
)

// osLock is the Lock adapter: the `lock` file inside the clone's store, created
// with O_EXCL so two runs
// cannot both believe they created it, dated by the filesystem so a lock left by
// a run that died can be told apart from a live one, and removed by the run that
// took it.
//
// Nothing here decides policy. Age is measured here because only the disk knows
// it; whether ten minutes is "abandoned", and what the human reads, belongs to
// internal/app/lock.
type osLock struct {
	// path and mine are what this run wrote. Release compares before removing:
	// a run longer than the staleness window can have its lock taken over by a
	// newer one, and deleting that would break the rule that a live lock is
	// never removed by someone else.
	path string
	mine []byte
	held bool
}

// Acquire takes the lock at path, or reports the one that is already there.
func (l *osLock) Acquire(path, owner string) (bool, string, error) {
	line := []byte(lock.NewLine(owner, os.Getpid(), time.Now()))

	// Two passes at most: the second one exists for the race where the other run
	// released (or was taken over) between the exclusive create and the stat.
	for attempt := 0; attempt < 2; attempt++ {
		err := createExclusive(path, line)
		if err == nil {
			l.path, l.mine, l.held = path, line, true
			return true, "", nil
		}
		if !errors.Is(err, os.ErrExist) {
			return false, "", err
		}

		info, statErr := os.Stat(path)
		switch {
		case errors.Is(statErr, os.ErrNotExist):
			continue
		case statErr != nil:
			// The lock is there and its age cannot be measured. It is never
			// ignored, so the run stops and the message names the file.
			return false, lock.Refusal(path, lock.Owner{}, lock.UnknownAge), nil
		}

		// The modification time decides staleness, never the content: a lock
		// written by another version, or cut in half by a crash, still says how
		// long it has been there.
		age := time.Since(info.ModTime())
		who, known := readOwner(path)
		if !known {
			who = lock.Owner{}
		}

		// Exactly ten minutes is still a live lock: when the age is the boundary,
		// waiting one more second costs nothing and refusing a live run costs an
		// incoherent project.
		if age <= lock.StaleAfter {
			return false, lock.Refusal(path, who, age), nil
		}

		// Abandoned. Taking it over is remove-then-exclusive-create, so two runs
		// that find the same old lock cannot both win: the one that loses the
		// create looks again and sees the winner's lock.
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, "", err
		}
		if err := createExclusive(path, line); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return false, "", err
		}
		l.path, l.mine, l.held = path, line, true
		return true, lock.StaleLine(path, who, age), nil
	}

	return false, lock.Refusal(path, lock.Owner{}, lock.UnknownAge), nil
}

// Release removes the lock, and only if it is still the one this run wrote.
func (l *osLock) Release() error {
	if !l.held {
		return nil
	}
	l.held = false

	current, err := os.ReadFile(l.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return err
	}
	if !bytes.Equal(current, l.mine) {
		return nil
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// createExclusive writes line to path in one exclusive step. The lock's parent
// directory is the first thing odachi writes in a fresh project, so it is
// created here; a failure to create it is a failure to take the lock.
func createExclusive(path string, line []byte) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(line); err != nil {
		file.Close()
		// A half-written lock would say the wrong owner for ten minutes; the
		// create was ours, so undoing it is safe.
		os.Remove(path)
		return err
	}
	return file.Close()
}

// readOwner reads the lock's one line. A file that cannot be read is as unknown
// as one that says something unexpected: both are still locks.
func readOwner(path string) (lock.Owner, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return lock.Owner{}, false
	}
	return lock.ParseOwner(data)
}
