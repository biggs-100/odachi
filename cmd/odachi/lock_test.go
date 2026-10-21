package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
)

// lockRel is the lock, relative to the project root. Contract XI pins the name.
const lockRel = ".odachi/lock"

// lockPath is that file inside a throwaway project.
func lockPath(dir string) string {
	return filepath.Join(dir, filepath.FromSlash(lockRel))
}

// handLock writes a lock by hand, the way a run that died leaves one, and dates
// it age ago. A negative age leaves the date alone.
func handLock(t *testing.T, dir, content string, age time.Duration) {
	t.Helper()
	path := lockPath(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", lockRel, err)
	}
	if age >= 0 {
		when := time.Now().Add(-age)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatalf("age %s: %v", lockRel, err)
		}
	}
}

// lockExists reports whether the lock file is on disk.
func lockExists(t *testing.T, dir string) bool {
	t.Helper()
	_, err := os.Stat(lockPath(dir))
	switch {
	case err == nil:
		return true
	case os.IsNotExist(err):
		return false
	default:
		t.Fatalf("stat %s: %v", lockRel, err)
		return false
	}
}

// TestTheLockIsOneLineAndGoesAway pins the adapter's half of Contract XI: an
// exclusive create, one line with verb/pid/instant, and a release that removes
// it.
func TestTheLockIsOneLineAndGoesAway(t *testing.T) {
	dir := t.TempDir()
	guard := &osLock{}

	held, line, err := guard.Acquire(lockPath(dir), "start")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if !held {
		t.Fatalf("Acquire held = false on a free lock (line: %q)", line)
	}
	if line != "" {
		t.Errorf("Acquire line = %q, want empty when the lock was free", line)
	}
	if !lockExists(t, dir) {
		t.Fatal("Acquire did not leave the lock file on disk")
	}

	data, err := os.ReadFile(lockPath(dir))
	if err != nil {
		t.Fatalf("read %s: %v", lockRel, err)
	}
	content := string(data)
	if !strings.HasSuffix(content, "\n") || strings.Count(content, "\n") != 1 {
		t.Fatalf("lock content = %q, want exactly one line", content)
	}
	owner, ok := lock.ParseOwner(data)
	if !ok {
		t.Fatalf("lock content = %q, want verb, pid and instant", content)
	}
	if owner.Verb != "start" {
		t.Errorf("lock verb = %q, want start", owner.Verb)
	}
	if owner.PID != strconv.Itoa(os.Getpid()) {
		t.Errorf("lock pid = %q, want %d", owner.PID, os.Getpid())
	}
	if _, err := time.Parse(time.RFC3339, owner.Instant); err != nil {
		t.Errorf("lock instant = %q, want an RFC3339 date: %v", owner.Instant, err)
	}

	if err := guard.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if lockExists(t, dir) {
		t.Error("Release left the lock behind")
	}
	if err := guard.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
}

// TestALiveLockRefusesAndIsNeverDeleted: a fresh lock is another run writing
// right now. odachi names it, refuses, and leaves it exactly where it is.
func TestALiveLockRefusesAndIsNeverDeleted(t *testing.T) {
	dir := t.TempDir()
	first := &osLock{}
	if held, _, err := first.Acquire(lockPath(dir), "start"); err != nil || !held {
		t.Fatalf("first Acquire: held=%v err=%v", held, err)
	}
	before, err := os.ReadFile(lockPath(dir))
	if err != nil {
		t.Fatalf("read %s: %v", lockRel, err)
	}

	second := &osLock{}
	held, line, err := second.Acquire(lockPath(dir), "undo")
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}
	if held {
		t.Fatal("second Acquire held a lock another run is holding")
	}
	for _, want := range []string{
		"otra corrida de odachi está escribiendo en este proyecto",
		"start, pid " + strconv.Itoa(os.Getpid()),
		"hace ",
		lockRel,
	} {
		if !strings.Contains(line, want) {
			t.Errorf("refusal = %q, want it to contain %q", line, want)
		}
	}

	// Never deleted, and never rewritten: the owner is still there for whoever
	// comes next.
	after, err := os.ReadFile(lockPath(dir))
	if err != nil {
		t.Fatalf("read %s after the refusal: %v", lockRel, err)
	}
	if string(after) != string(before) {
		t.Errorf("the refusal rewrote %s: %q -> %q", lockRel, before, after)
	}
	if err := second.Release(); err != nil {
		t.Fatalf("Release on a refusal: %v", err)
	}
	if !lockExists(t, dir) {
		t.Error("the refusal deleted the owner's lock")
	}

	if err := first.Release(); err != nil {
		t.Fatalf("first Release: %v", err)
	}
	if lockExists(t, dir) {
		t.Error("the owner's Release left the lock behind")
	}
}

// TestAStaleLockIsTakenOver: older than ten minutes means the run that wrote it
// died. It is taken over — one line says so — and the new run owns the file.
func TestAStaleLockIsTakenOver(t *testing.T) {
	dir := t.TempDir()
	handLock(t, dir, "start 111 2026-10-21T10:00:00Z\n", lock.StaleAfter+time.Minute)

	guard := &osLock{}
	held, line, err := guard.Acquire(lockPath(dir), "undo")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if !held {
		t.Fatalf("Acquire held = false on an abandoned lock (line: %q)", line)
	}
	for _, want := range []string{"había un lock viejo de start", "lo tomé"} {
		if !strings.Contains(line, want) {
			t.Errorf("announcement = %q, want it to contain %q", line, want)
		}
	}

	data, err := os.ReadFile(lockPath(dir))
	if err != nil {
		t.Fatalf("read %s: %v", lockRel, err)
	}
	if owner, ok := lock.ParseOwner(data); !ok || owner.Verb != "undo" || owner.PID != strconv.Itoa(os.Getpid()) {
		t.Errorf("lock content = %q, want the new owner (undo, %d)", data, os.Getpid())
	}

	if err := guard.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if lockExists(t, dir) {
		t.Error("Release left the taken-over lock behind")
	}
}

// TestAnUnreadableLockIsJudgedByItsAge: a lock with unexpected content is never
// ignored. Old, it is taken over with a line; fresh, it blocks naming the file.
func TestAnUnreadableLockIsJudgedByItsAge(t *testing.T) {
	fresh := t.TempDir()
	handLock(t, fresh, "basura sin forma\n", 30*time.Second)

	guard := &osLock{}
	held, line, err := guard.Acquire(lockPath(fresh), "start")
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if held {
		t.Fatal("Acquire took a fresh lock it could not read")
	}
	for _, want := range []string{"otra corrida de odachi está escribiendo en este proyecto", "no pude leer " + lockRel, "hace "} {
		if !strings.Contains(line, want) {
			t.Errorf("refusal = %q, want it to contain %q", line, want)
		}
	}
	if !lockExists(t, fresh) {
		t.Error("the refusal deleted a lock it could not read")
	}

	old := t.TempDir()
	handLock(t, old, "basura sin forma\n", 30*time.Minute)
	taken := &osLock{}
	held, line, err = taken.Acquire(lockPath(old), "start")
	if err != nil {
		t.Fatalf("Acquire on the old lock: %v", err)
	}
	if !held {
		t.Fatalf("Acquire held = false on an old lock with unexpected content (line: %q)", line)
	}
	for _, want := range []string{"había un lock viejo", lockRel, "lo tomé"} {
		if !strings.Contains(line, want) {
			t.Errorf("announcement = %q, want it to contain %q", line, want)
		}
	}
	if err := taken.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

// TestReleaseNeverDeletesSomebodyElsesLock: a run that outlives the staleness
// window can have its lock taken over. Its Release must not remove the new
// owner's lock.
func TestReleaseNeverDeletesSomebodyElsesLock(t *testing.T) {
	dir := t.TempDir()
	guard := &osLock{}
	if held, _, err := guard.Acquire(lockPath(dir), "start"); err != nil || !held {
		t.Fatalf("Acquire: held=%v err=%v", held, err)
	}

	// Another run took it over after ten minutes of interview.
	taken := "undo 999 2026-10-21T16:00:00Z\n"
	if err := os.WriteFile(lockPath(dir), []byte(taken), 0o644); err != nil {
		t.Fatalf("write %s: %v", lockRel, err)
	}

	if err := guard.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	data, err := os.ReadFile(lockPath(dir))
	if err != nil {
		t.Fatalf("read %s after Release: %v", lockRel, err)
	}
	if string(data) != taken {
		t.Errorf("lock content after Release = %q, want the other run's %q", data, taken)
	}
}

// startArgs is a full `start` invocation with every answer as a flag. JSON is
// left out so the test reads the human line.
func startArgs(extra ...string) []string {
	args := []string{"start"}
	args = append(args, extra...)
	for _, flag := range allFlags() {
		if flag != "--json" {
			args = append(args, flag)
		}
	}
	return args
}

// TestStartLeavesNoLockBehind is the ordinary case end to end: the lock exists
// while start writes and is gone when it returns.
func TestStartLeavesNoLockBehind(t *testing.T) {
	dir := t.TempDir()

	if code, _, stderr := runDir(t, dir, startArgs(), ""); code != 0 {
		t.Fatalf("start: exit %d, stderr=%s", code, stderr)
	}
	if lockExists(t, dir) {
		t.Error("start left .odachi/lock behind")
	}
}

// TestStartIsRefusedByALiveLockAndTouchesNothing: the run stops before reading
// an artifact, and every byte in the project is exactly as it was.
func TestStartIsRefusedByALiveLockAndTouchesNothing(t *testing.T) {
	dir := t.TempDir()
	if code, _, stderr := runDir(t, dir, startArgs(), ""); code != 0 {
		t.Fatalf("first start: exit %d, stderr=%s", code, stderr)
	}
	handLock(t, dir, "start 4242 2026-10-21T15:47:03Z\n", 30*time.Second)
	before := tree(t, dir)

	code, stdout, stderr := runDir(t, dir, startArgs("--force"), "")
	if code != 1 {
		t.Fatalf("second start: exit %d, want 1 (stdout=%s stderr=%s)", code, stdout, stderr)
	}
	for _, want := range []string{
		"otra corrida de odachi está escribiendo en este proyecto",
		"start, pid 4242",
		"hace ",
		lockRel,
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to contain %q", stderr, want)
		}
	}
	if strings.Count(strings.TrimSpace(stderr), "\n") != 0 {
		t.Errorf("the refusal must be one line: %q", stderr)
	}

	after := tree(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("the refused run changed the project:\nbefore: %v\nafter:  %v", keys(before), keys(after))
		for path, content := range after {
			if before[path] != content {
				t.Errorf("  %s: %q -> %q", path, before[path], content)
			}
		}
	}

	// --json carries the same message.
	code, stdout, _ = runDir(t, dir, append(startArgs("--force"), "--json"), "")
	if code != 1 {
		t.Fatalf("start --json: exit %d, want 1 (stdout=%s)", code, stdout)
	}
	if !strings.Contains(stdout, "otra corrida de odachi está escribiendo en este proyecto") {
		t.Errorf("--json = %q, want the same refusal", stdout)
	}
}

// keys lists a snapshot's paths, sorted: a failure reads better than a map dump.
func keys(files map[string]string) []string {
	out := make([]string, 0, len(files))
	for path := range files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// TestAnOldLockIsTakenOverEndToEnd: the line reaches the human and the run
// completes with the lock released.
func TestAnOldLockIsTakenOverEndToEnd(t *testing.T) {
	dir := t.TempDir()
	handLock(t, dir, "start 111 2026-10-21T10:00:00Z\n", 30*time.Minute)

	code, stdout, stderr := runDir(t, dir, startArgs(), "")
	if code != 0 {
		t.Fatalf("start with an old lock: exit %d, stderr=%s", code, stderr)
	}
	for _, want := range []string{"había un lock viejo de start", "lo tomé"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
	if lockExists(t, dir) {
		t.Error("start left .odachi/lock behind after taking over an old one")
	}
}

// TestTheSingleFileVerbsTakeNoLock: plan, ask and level write one file each,
// atomically, so a lock would protect nothing and only add friction.
func TestTheSingleFileVerbsTakeNoLock(t *testing.T) {
	dir, _ := startedProject(t)
	if lockExists(t, dir) {
		t.Fatal("the project started with a lock on disk")
	}

	commands := [][]string{
		{"plan", "un paso más"},
		{"ask", "¿Los turnos se pueden cancelar?"},
		{"level"},
	}
	for _, args := range commands {
		if code, _, stderr := runDir(t, dir, args, ""); code != 0 {
			t.Fatalf("%v: exit %d, stderr=%s", args, code, stderr)
		}
		if lockExists(t, dir) {
			t.Errorf("%v created %s", args, lockRel)
		}
	}
}

// TestAFailingStartLeavesNoLockBehind: a flag error never reaches the verb, and
// a verb-level failure releases with defer. Neither leaves a lock that would
// block the project for ten minutes.
func TestAFailingStartLeavesNoLockBehind(t *testing.T) {
	dir := t.TempDir()

	// A flag the command does not have: start never runs.
	if code, _, stderr := runDir(t, dir, []string{"start", "--no-existe"}, ""); code != 1 {
		t.Fatalf("start --no-existe: exit %d, want 1 (stderr=%s)", code, stderr)
	}
	if lockExists(t, dir) {
		t.Error("a failed flag parse created .odachi/lock")
	}

	// A run that starts, takes the lock and then stops on its own rule.
	if code, _, stderr := runDir(t, dir, startArgs(), ""); code != 0 {
		t.Fatalf("first start: exit %d, stderr=%s", code, stderr)
	}
	code, _, stderr := runDir(t, dir, startArgs(), "") // brief.md exists, no --force
	if code != 1 {
		t.Fatalf("second start without --force: exit %d, want 1 (stderr=%s)", code, stderr)
	}
	if !strings.Contains(stderr, "--force") {
		t.Errorf("stderr = %q, want it to name --force", stderr)
	}
	if lockExists(t, dir) {
		t.Error("a start that failed on its own rule left .odachi/lock behind")
	}
}
