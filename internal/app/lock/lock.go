// Package lock is the write lock that `odachi start` and `odachi undo` hold
// while they write more than one file.
//
// Atomic writes protect each file; they do not protect the set. Two runs that
// interleave can leave the brief of one beside the fence of the other — every
// file valid, the project incoherent. The lock is the only thing that makes a
// multi-file verb exclusive, and only the multi-file verbs take it: `plan`, `ask`
// and `level --to` write a single file atomically, so a lock there would protect
// nothing and be pure friction.
//
// The domain takes no part: there is no business rule here, only exclusion.
// Everything that touches the filesystem — O_EXCL, the modification time, the
// removal — belongs to the adapter behind the Lock port.
package lock

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// File is the lock, relative to the project root. One name, so every message
// that names the file names the same one.
const File = ".odachi/lock"

// StaleAfter is when a lock stops meaning "someone is writing right now" and
// starts meaning "someone was writing and never finished". Ten minutes: longer
// than any interview, shorter than a coffee break.
const StaleAfter = 10 * time.Minute

// UnknownAge is the age of a lock whose date could not be measured. It is
// negative on purpose: no real age can compare with it by accident, and Age
// says so in words instead of printing "-1s".
const UnknownAge = time.Duration(-1)

// Lock is the exclusivity port. The adapter creates the file with O_EXCL, reads
// its modification time and removes it; this package only ever sees the outcome.
type Lock interface {
	// Acquire tries to take the lock at path for owner (the verb's name).
	// held reports whether the lock is ours now.
	//
	// stale carries the one line the human has to see about the lock that was
	// already there: the announcement when an abandoned lock was taken over
	// ("there was an old lock from …, … ago; I took it"), or the refusal when a
	// live one blocks the run. It is empty only when the file was free.
	Acquire(path, owner string) (held bool, stale string, err error)

	// Release frees the lock, and only the lock this port took: a lock another
	// run took over in the meantime is never removed.
	Release() error
}

// Owner is what the lock file says: who is writing, and since when.
type Owner struct {
	Verb    string
	PID     string
	Instant string
}

// NewLine is the whole content of a fresh lock: the verb, the pid and the
// instant, on one line. Three fields, because the refusal has to name a verb, a
// pid and an age, and a file with more than one line is a file no message can
// summarise.
func NewLine(verb string, pid int, now time.Time) string {
	return fmt.Sprintf("%s %d %s\n", verb, pid, now.UTC().Format(time.RFC3339))
}

// ParseOwner reads the line NewLine writes. ok is false when the content is not
// that shape — a lock written by an older version, a truncated file, a note a
// human left. The pid and the instant are checked, not only counted: three
// words that are not a pid and an instant are exactly the "content we cannot
// read" the refusal is supposed to name. The lock is still a lock: its age still
// decides, and the messages name the file instead of inventing a verb and a pid.
func ParseOwner(data []byte) (Owner, bool) {
	fields := strings.Fields(string(data))
	if len(fields) != 3 {
		return Owner{}, false
	}
	if _, err := strconv.Atoi(fields[1]); err != nil {
		return Owner{}, false
	}
	if _, err := time.Parse(time.RFC3339, fields[2]); err != nil {
		return Owner{}, false
	}
	return Owner{Verb: fields[0], PID: fields[1], Instant: fields[2]}, true
}

// Age is how an age reads in a message: seconds below a minute, minutes above
// it, and "a while" when the date could not be measured. A human reading "42s
// ago" knows to wait; one reading "900s ago" has to do arithmetic.
func Age(age time.Duration) string {
	if age < 0 {
		return "a while"
	}
	if age < time.Second {
		return "0s"
	}
	return age.Round(time.Second).String()
}

// StaleLine is the announcement a run prints when it took over an abandoned
// lock: one line, because taking a lock that is not yours is something the human
// has to know.
func StaleLine(owner Owner, age time.Duration) string {
	if owner.Verb == "" {
		return fmt.Sprintf("there was an old lock at %s, %s ago; I took it", File, Age(age))
	}
	return fmt.Sprintf("there was an old lock from %s, %s ago; I took it", owner.Verb, Age(age))
}

// Refusal is the message when a live lock blocks a run. It names who is writing,
// how long ago, and the file a human can delete by hand once they are sure
// nothing is running: odachi itself never deletes a live lock.
func Refusal(owner Owner, age time.Duration) string {
	if owner.Verb == "" || owner.PID == "" {
		return fmt.Sprintf(
			"another odachi run is writing to this project (could not read %s, %s ago); wait and try again; if you know nothing is running, delete %s by hand",
			File, Age(age), File)
	}
	return fmt.Sprintf(
		"another odachi run is writing to this project (%s, pid %s, %s ago); wait and try again; if you know nothing is running, delete %s by hand",
		owner.Verb, owner.PID, Age(age), File)
}

// Held is the lock as one verb sees it: taken, announced, or refused.
type Held struct {
	// Notice is the announcement to show before the verb's own answer, when an
	// abandoned lock was taken over. Empty when the lock was free or the run was
	// refused.
	Notice string
	// Refusal is the one line that stops the run before it reads anything.
	// Empty when the lock is ours, or when there is no port to take it with.
	Refusal string

	port Lock
	held bool
}

// Take acquires the lock for verb in dir. The caller writes
// `defer held.Release()` right away: a run that fails while holding the lock
// would block the project until the lock grew old.
//
// A nil port takes nothing. That is for a test that is not about the lock; the
// binary always wires one, so every real run of a multi-file verb takes it.
func Take(port Lock, dir, verb string) Held {
	if port == nil {
		return Held{}
	}

	path := filepath.Join(dir, filepath.FromSlash(File))
	held, line, err := port.Acquire(path, verb)

	switch {
	case err != nil:
		// No lock means no exclusive write, and a verb that writes anyway is the
		// incoherent project this whole file exists to prevent.
		return Held{Refusal: oneLine(fmt.Sprintf(
			"could not take %s: %s; check the project permissions and try again",
			File, err.Error()))}
	case !held:
		if strings.TrimSpace(line) == "" {
			// An adapter that refuses without saying why still has to name the
			// file: a refusal that cannot be acted on is a dead end.
			line = Refusal(Owner{}, UnknownAge)
		}
		return Held{Refusal: oneLine(line)}
	}

	return Held{Notice: oneLine(line), port: port, held: true}
}

// Release frees the lock. It is safe to call twice and safe to call on a
// refusal or a missing port, so every verb can defer it without a condition.
func (h *Held) Release() error {
	if h == nil || !h.held || h.port == nil {
		return nil
	}
	h.held = false
	return h.port.Release()
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
