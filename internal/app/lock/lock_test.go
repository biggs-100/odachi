package lock_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
)

// now is a fixed instant: the lock line carries a date, and a test that used
// the clock could not assert it.
func now() time.Time {
	return time.Date(2026, 10, 21, 15, 47, 3, 0, time.UTC)
}

// fakeLock is the Lock port in memory. It records what the app asked for, so a
// test can assert the path, the owner and whether the release happened.
type fakeLock struct {
	held bool
	line string
	err  error

	paths    []string
	owners   []string
	releases int
}

func (f *fakeLock) Acquire(path, owner string) (bool, string, error) {
	f.paths = append(f.paths, path)
	f.owners = append(f.owners, owner)
	if f.err != nil {
		return false, "", f.err
	}
	return f.held, f.line, nil
}

func (f *fakeLock) Release() error {
	f.releases++
	return nil
}

// TestNewLineIsTheWholeLockFile pins the file format: one line with the verb,
// the pid and the instant, and nothing else. The stale message and the refusal
// both read those three fields, so a second line would be a lock the tool can
// no longer describe.
func TestNewLineIsTheWholeLockFile(t *testing.T) {
	line := lock.NewLine("start", 4242, now())

	if want := "start 4242 2026-10-21T15:47:03Z\n"; line != want {
		t.Errorf("NewLine = %q, want %q", line, want)
	}
	if strings.Count(strings.TrimSuffix(line, "\n"), "\n") != 0 {
		t.Errorf("NewLine = %q, want exactly one line", line)
	}
}

func TestParseOwnerReadsWhatNewLineWrites(t *testing.T) {
	owner, ok := lock.ParseOwner([]byte(lock.NewLine("undo", 7, now())))
	if !ok {
		t.Fatal("ParseOwner rejected what NewLine wrote")
	}
	if owner.Verb != "undo" || owner.PID != "7" || owner.Instant != "2026-10-21T15:47:03Z" {
		t.Errorf("ParseOwner = %+v, want undo/7/2026-10-21T15:47:03Z", owner)
	}

	// Unexpected content is not an error: the lock is still a lock, and its age
	// still decides. ok=false is what makes the message name the file instead of
	// inventing a verb instead of a pid.
	for _, junk := range []string{
		"",
		"basura",
		"basura sin forma",               // three words, but not a lock line
		"start abc 2026-10-21T15:47:03Z", // pid is not a number
		"start 42 ayer",                  // not an instant
		"start",                          // truncated
		"start 1",                        // truncated
		"start 1 2 3",                    // one line, not the lock's line
		"\x00\x01",                       // binary
	} {
		if _, ok := lock.ParseOwner([]byte(junk)); ok {
			t.Errorf("ParseOwner(%q) = ok; want rejected", junk)
		}
	}
}

// TestAgeIsWhatTheHumanReads: seconds below a minute and minutes above it, so a
// refusal reads "42s ago" and an abandoned lock reads "31m12s ago".
func TestAgeIsWhatTheHumanReads(t *testing.T) {
	cases := []struct {
		age  time.Duration
		want string
	}{
		{42 * time.Second, "42s"},
		{500 * time.Millisecond, "0s"},
		{31*time.Minute + 12*time.Second, "31m12s"},
		{10 * time.Minute, "10m0s"},
		{0, "0s"},
		// A lock whose date could not be measured says so instead of inventing
		// an age: the run is still blocked, so the number is never used to decide.
		{lock.UnknownAge, "a while"},
	}
	for _, c := range cases {
		if got := lock.Age(c.age); got != c.want {
			t.Errorf("Age(%v) = %q, want %q", c.age, got, c.want)
		}
	}
}

// TestStaleLineIsTheAnnouncement: an abandoned lock is taken over, and the run
// says so in one line instead of doing it silently.
func TestStaleLineIsTheAnnouncement(t *testing.T) {
	known := lock.Owner{Verb: "start", PID: "4242", Instant: "2026-10-21T15:47:03Z"}
	if got, want := lock.StaleLine("store/lock", known, 31*time.Minute+12*time.Second), "there was an old lock from start, 31m12s ago; I took it"; got != want {
		t.Errorf("StaleLine = %q, want %q", got, want)
	}

	// A lock whose line cannot be read still gets announced, naming the file:
	// the human has to know a lock was taken over, even when nobody signed it.
	unreadable := lock.StaleLine("store/lock", lock.Owner{}, time.Hour)
	if !strings.Contains(unreadable, "store/lock") {
		t.Errorf("StaleLine for an unreadable lock = %q, want it to name the file", unreadable)
	}
	if !strings.Contains(unreadable, "I took it") {
		t.Errorf("StaleLine for an unreadable lock = %q, want it to say it was taken", unreadable)
	}
}

// TestRefusalIsTheExactSentence: the message Contract XI pins down, with the
// verb, the pid, the age, and the file a human can delete by hand. odachi never
// deletes a live lock.
func TestRefusalIsTheExactSentence(t *testing.T) {
	owner := lock.Owner{Verb: "start", PID: "4242", Instant: "2026-10-21T15:47:03Z"}
	got := lock.Refusal("store/lock", owner, 42*time.Second)

	want := "another odachi run is writing to this project (start, pid 4242, 42s ago); wait and try again; if you know nothing is running, delete store/lock by hand"
	if got != want {
		t.Errorf("Refusal = %q, want %q", got, want)
	}

	// A lock with unexpected content blocks just the same; it names the file
	// instead of inventing a verb and a pid.
	junk := lock.Refusal("store/lock", lock.Owner{}, 42*time.Second)
	for _, want := range []string{"another odachi run is writing to this project", "store/lock", "42s ago"} {
		if !strings.Contains(junk, want) {
			t.Errorf("Refusal for an unreadable lock = %q, want it to contain %q", junk, want)
		}
	}
	if strings.Contains(junk, "pid") {
		t.Errorf("Refusal for an unreadable lock = %q, want no invented pid", junk)
	}
}

// TestTakeTakesTheLockInTheStore pins the path and the owner: the file lives in
// the store root the caller resolved, and the owner is the verb, so the next run
// can say who wrote.
func TestTakeTakesTheLockInTheStore(t *testing.T) {
	port := &fakeLock{held: true}
	held := lock.Take(port, "proyecto", "start")

	wantPath := filepath.Join("proyecto", lock.File)
	if len(port.paths) != 1 || port.paths[0] != wantPath {
		t.Errorf("Acquire paths = %v, want one call with %q", port.paths, wantPath)
	}
	if len(port.owners) != 1 || port.owners[0] != "start" {
		t.Errorf("Acquire owners = %v, want one call with %q", port.owners, "start")
	}
	if held.Refusal != "" || held.Notice != "" {
		t.Errorf("Take on a free lock = %+v, want neither refusal nor notice", held)
	}
}

// TestReleaseIsIdempotentAndOnlyAfterHolding: a verb writes `defer
// held.Release()`, so releasing twice, or releasing a refusal, must do nothing.
func TestReleaseIsIdempotentAndOnlyAfterHolding(t *testing.T) {
	port := &fakeLock{held: true}
	held := lock.Take(port, "proyecto", "start")
	if err := held.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if err := held.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want exactly 1", port.releases)
	}

	busy := &fakeLock{held: false, line: "another odachi run is writing to this project"}
	refused := lock.Take(busy, "proyecto", "start")
	_ = refused.Release()
	if busy.releases != 0 {
		t.Errorf("a refusal released the lock %d times; the lock is not ours to release", busy.releases)
	}
}

// TestTakeReportsARefusal: when another run holds the lock, the verb gets the
// line to print and no lock to release.
func TestTakeReportsARefusal(t *testing.T) {
	const line = "another odachi run is writing to this project (start, pid 4242, 42s ago)"
	held := lock.Take(&fakeLock{held: false, line: line}, "proyecto", "undo")

	if held.Refusal != line {
		t.Errorf("Refusal = %q, want %q", held.Refusal, line)
	}
	if held.Notice != "" {
		t.Errorf("Notice = %q, want empty on a refusal", held.Notice)
	}
}

// TestTakeAnnouncesAStaleLock: an abandoned lock is taken, and the line that
// says so is carried to the verb's output.
func TestTakeAnnouncesAStaleLock(t *testing.T) {
	const line = "there was an old lock from start, 31m12s ago; I took it"
	port := &fakeLock{held: true, line: line}
	held := lock.Take(port, "proyecto", "start")

	if held.Notice != line {
		t.Errorf("Notice = %q, want %q", held.Notice, line)
	}
	if held.Refusal != "" {
		t.Errorf("Refusal = %q, want empty when the lock was taken", held.Refusal)
	}
	if err := held.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want 1", port.releases)
	}
}

// TestTakeBlocksWhenTheLockCannotBeTaken: a lock file that cannot be created is
// a refusal that names the file, never a run that writes anyway.
func TestTakeBlocksWhenTheLockCannotBeTaken(t *testing.T) {
	port := &fakeLock{err: errors.New("permission denied")}
	held := lock.Take(port, "proyecto", "start")

	if held.Refusal == "" {
		t.Fatal("Refusal empty: a lock that cannot be taken must stop the run")
	}
	for _, want := range []string{lock.File, "permission denied"} {
		if !strings.Contains(held.Refusal, want) {
			t.Errorf("Refusal = %q, want it to contain %q", held.Refusal, want)
		}
	}
	if err := held.Release(); err != nil {
		t.Fatalf("Release after a failed Acquire: %v", err)
	}
	if port.releases != 0 {
		t.Errorf("port released %d times, want 0", port.releases)
	}
}

// TestTakeWithoutAPortDoesNothing: a test that is not about the lock passes no
// port, and the verb behaves exactly as before.
func TestTakeWithoutAPortDoesNothing(t *testing.T) {
	held := lock.Take(nil, "proyecto", "start")
	if held.Refusal != "" || held.Notice != "" {
		t.Errorf("Take(nil) = %+v, want neither refusal nor notice", held)
	}
	if err := held.Release(); err != nil {
		t.Fatalf("Release(nil): %v", err)
	}
}
