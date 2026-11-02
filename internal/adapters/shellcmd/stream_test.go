package shellcmd_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

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
