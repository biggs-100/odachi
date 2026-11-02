package shellcmd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
)

// syncBuffer is the mirror a test hands the adapter: two streams are copied by
// two goroutines, so the writer has to be safe for concurrent use, exactly as the
// terminal is.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// watchBuffer is the mirror that says when it was written: the signal is what
// lets a test look at the bytes while the command is still running.
type watchBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
	ch  chan struct{}
}

func newWatchBuffer() *watchBuffer {
	return &watchBuffer{ch: make(chan struct{}, 1)}
}

func (w *watchBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.buf.Write(p)
	w.mu.Unlock()
	select {
	case w.ch <- struct{}{}:
	default:
	}
	return n, err
}

func (w *watchBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// The child of the timing test reads these: it prints, then waits for the release
// file the parent writes only after seeing both lines.
const (
	helperEnv  = "ODACHI_STREAM_HELPER"
	releaseEnv = "ODACHI_STREAM_RELEASE"
)

// TestRunStreamWritesWhileTheCommandRuns: the point of the stream is when the
// bytes arrive, not only that they do. Every other test here reads the mirror
// after the command ended, so an implementation that captured everything and
// wrote it at the end would pass all of them and bring back the silence.
//
// The child is this same test binary re-run: it prints to both streams and blocks
// until the parent says so, which means it cannot exit before the parent has seen
// what it printed. If nothing arrives, the child never exits and the run reports
// its own timeout instead of hanging the suite.
func TestRunStreamWritesWhileTheCommandRuns(t *testing.T) {
	if os.Getenv(helperEnv) == "1" {
		fmt.Fprintln(os.Stdout, "ready-out")
		fmt.Fprintln(os.Stderr, "ready-err")
		release := os.Getenv(releaseEnv)
		// The wait is bounded: a parent that dies leaves no spinner behind. The
		// bound is never the assertion -the parent releases the child as soon as
		// it sees both lines- it is what keeps a failed run from leaking a
		// process that holds the build cache.
		for waited := time.Duration(0); waited < 60*time.Second; waited += 5 * time.Millisecond {
			if _, err := os.Stat(release); err == nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		return
	}

	release := filepath.Join(t.TempDir(), "release")
	t.Setenv(helperEnv, "1")
	t.Setenv(releaseEnv, release)
	// Whatever happens below, the child hears the release: a test that fails must
	// not leave the process it started waiting for a file nobody will write.
	defer os.WriteFile(release, []byte("fail\n"), 0o600)

	seen := newWatchBuffer()
	type result struct {
		stdout, stderr string
		exit           int
		timedOut       bool
		err            error
	}
	done := make(chan result, 1)

	command := shellcmd.Quote(os.Args[0]) + " -test.run=^TestRunStreamWritesWhileTheCommandRuns$"
	go func() {
		stdout, stderr, exit, timedOut, err := shellcmd.RunStream(".", command, 30, seen)
		done <- result{stdout: stdout, stderr: stderr, exit: exit, timedOut: timedOut, err: err}
	}()

	// The guard is a bound, not the assertion: a broken stream must fail the test
	// instead of leaving it waiting forever.
	deadline := time.After(60 * time.Second)
	for !strings.Contains(seen.String(), "ready-out") || !strings.Contains(seen.String(), "ready-err") {
		select {
		case <-seen.ch:
		case got := <-done:
			t.Fatalf("the run ended before both lines arrived (mirror %q, timedOut %v, err %v)",
				seen.String(), got.timedOut, got.err)
		case <-deadline:
			t.Fatalf("nothing reached the mirror while the command was alive (mirror %q)", seen.String())
		}
	}

	// Both lines arrived while the child was blocked: release it and let the run
	// finish, so the test also proves the stream never replaced the captured answer.
	//
	// The run must still be open: an implementation that writes the mirror when
	// the process ends leaves it ready, and that is the silence this test exists
	// to refuse.
	select {
	case got := <-done:
		t.Fatalf("the lines arrived only once the run had ended (timedOut %v, err %v)", got.timedOut, got.err)
	default:
	}
	if err := os.WriteFile(release, []byte("go\n"), 0o600); err != nil {
		t.Fatalf("release the child: %v", err)
	}
	got := <-done

	if got.err != nil {
		t.Fatalf("RunStream: %v", got.err)
	}
	if got.timedOut {
		t.Fatalf("the run timed out: the child was never released")
	}
	if !strings.Contains(got.stdout, "ready-out") {
		t.Errorf("stdout = %q, want the line the child printed", got.stdout)
	}
	if !strings.Contains(got.stderr, "ready-err") {
		t.Errorf("stderr = %q, want the line the child printed to the second stream", got.stderr)
	}
}

// TestRunStreamMirrorsWhatTheCommandPrints: the proof is written when the command
// ends, so without a mirror a human watching a lock sees nothing at all. The
// mirror receives what the command says, and the two returned streams keep the
// separation the proof depends on.
func TestRunStreamMirrorsWhatTheCommandPrints(t *testing.T) {
	seen := &syncBuffer{}

	stdout, stderr, exitCode, timedOut, err := shellcmd.RunStream(".", "go version", 60, seen)
	if err != nil {
		t.Fatalf("RunStream: %v", err)
	}
	if exitCode != 0 || timedOut {
		t.Fatalf("exit = %d, timedOut = %v, want a clean run", exitCode, timedOut)
	}
	if !strings.Contains(seen.String(), strings.TrimSpace(stdout)) {
		t.Errorf("mirror = %q, want the command's own output %q", seen.String(), stdout)
	}
	if strings.Contains(stdout, strings.TrimSpace(stderr)) && strings.TrimSpace(stderr) != "" {
		t.Errorf("stdout = %q carries the second stream: the proof composes them apart", stdout)
	}
}

// TestRunStreamWithoutAMirrorIsRun: the nil writer is the quiet call the reviewer
// and every existing caller already make.
func TestRunStreamWithoutAMirrorIsRun(t *testing.T) {
	streamed, _, code, _, err := shellcmd.RunStream(".", "go version", 60, nil)
	if err != nil {
		t.Fatalf("RunStream: %v", err)
	}

	plain, _, plainCode, _, err := shellcmd.Run(".", "go version", 60)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if streamed != plain || code != plainCode {
		t.Errorf("RunStream(nil) = %q/%d, want Run's %q/%d", streamed, code, plain, plainCode)
	}
}
