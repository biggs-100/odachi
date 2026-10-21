package start_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/lock"
	"github.com/biggs-100/odachi/internal/app/start"
)

// fakeLock is the Lock port in memory: it answers with what the test decides and
// remembers what the verb asked it, so the lock can be asserted without a disk.
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

// busyLock is a lock another run holds right now.
func busyLock() (*fakeLock, string) {
	line := lock.Refusal(lock.Owner{Verb: "start", PID: "4242", Instant: "2026-10-21T15:47:03Z"}, 42*time.Second)
	return &fakeLock{held: false, line: line}, line
}

// TestStartIsRefusedWhileAnotherRunHoldsTheLock is the whole point of T17: two
// runs writing the same four artifacts would leave every file valid and the
// project incoherent, so the second one stops before reading a single artifact.
func TestStartIsRefusedWhileAnotherRunHoldsTheLock(t *testing.T) {
	var events []string
	files := newFakeFiles(nil, &events)
	port, refusal := busyLock()

	opts := flagsCarryTheSameAnswers()
	opts.Machine = true

	got := start.Run(opts, nil, files, port)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	for _, want := range []string{
		"otra corrida de odachi está escribiendo en este proyecto",
		"start, pid 4242",
		"hace 42s",
		".odachi/lock",
	} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", got.Message, want)
		}
	}
	if got.Message != refusal {
		t.Errorf("Message = %q, want the refusal the port reported (%q)", got.Message, refusal)
	}

	// Nothing touched: no artifact is read, backed up or written before the
	// refusal. The fake records every call the verb makes through the port.
	if len(events) != 0 {
		t.Errorf("la corrida bloqueada tocó el proyecto: %v", events)
	}
	if written := files.paths(); len(written) != 0 {
		t.Errorf("archivos escritos durante el bloqueo: %v", written)
	}
	if port.releases != 0 {
		t.Errorf("a refusal released the lock %d times; it was never ours", port.releases)
	}
	if len(port.owners) != 1 || port.owners[0] != "start" {
		t.Errorf("Acquire owners = %v, want one call with %q", port.owners, "start")
	}
	// The port speaks the platform's separator; the file it names is .odachi/lock.
	wantPath := filepath.Join(".odachi", "lock")
	if len(port.paths) != 1 || !strings.HasSuffix(port.paths[0], wantPath) {
		t.Errorf("Acquire paths = %v, want one call ending in %q", port.paths, wantPath)
	}
}

// TestStartTakesOverAStaleLockAndSaysSo: a lock older than ten minutes belongs
// to a run that died. It is taken over, the human is told in one line, and the
// run completes.
func TestStartTakesOverAStaleLockAndSaysSo(t *testing.T) {
	const notice = "había un lock viejo de start, hace 31m12s; lo tomé"
	files := newFakeFiles(nil, nil)
	port := &fakeLock{held: true, line: notice}

	opts := flagsCarryTheSameAnswers()
	opts.Machine = true

	got := start.Run(opts, nil, files, port)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(got.Files) != 4 {
		t.Errorf("Files = %v, want the four artifacts", got.Files)
	}
	if !strings.Contains(got.Message, notice) {
		t.Errorf("Message = %q, want the announcement %q", got.Message, notice)
	}
	// The announcement is its own line: it is something that happened, not part
	// of the verb's answer.
	if len(got.Lines) != 2 || got.Lines[0] != notice {
		t.Errorf("Lines = %v, want the announcement first and the result second", got.Lines)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want 1", port.releases)
	}
}

// TestStartReleasesTheLockWhenTheVerbFails: the lock is released with defer, so
// a run that stops on "brief.md already exists" does not block the project for
// ten minutes.
func TestStartReleasesTheLockWhenTheVerbFails(t *testing.T) {
	files := newFakeFiles(map[string]string{"brief.md": "# ya existe\n"}, nil)
	port := &fakeLock{held: true}

	opts := flagsCarryTheSameAnswers()
	opts.Machine = true

	got := start.Run(opts, nil, files, port)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	if !strings.Contains(got.Message, "--force") {
		t.Errorf("Message = %q, want it to name --force", got.Message)
	}
	if port.releases != 1 {
		t.Errorf("port released %d times, want exactly 1 after a failed run", port.releases)
	}
}

// TestStartStopsWhenTheLockCannotBeTaken: a lock that cannot be created is not a
// reason to write anyway. The message names the file and the failure.
func TestStartStopsWhenTheLockCannotBeTaken(t *testing.T) {
	var events []string
	files := newFakeFiles(nil, &events)
	port := &fakeLock{err: errors.New("permiso denegado")}

	opts := flagsCarryTheSameAnswers()
	opts.Machine = true

	got := start.Run(opts, nil, files, port)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	for _, want := range []string{".odachi/lock", "permiso denegado"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", got.Message, want)
		}
	}
	if len(events) != 0 {
		t.Errorf("la corrida sin lock tocó el proyecto: %v", events)
	}
	if port.releases != 0 {
		t.Errorf("port released %d times, want 0", port.releases)
	}
}
