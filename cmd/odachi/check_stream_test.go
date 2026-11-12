package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// blockedWriter is a stream nobody drains: the first write never returns. It is
// the shape of the bug -a check --run whose stderr is captured by a caller that
// reads it only at the end- without needing a real pipe.
type blockedWriter struct {
	release <-chan struct{}
}

func (w blockedWriter) Write(p []byte) (int, error) {
	<-w.release
	return len(p), nil
}

// TestCheckRunFinishesWhenStderrIsNotDrained: the command's output is a courtesy
// for a human at a terminal, and the proof keeps both streams. A writer that is
// not a terminal cannot take it: a pipe nobody reads blocks the command, and the
// run then waits for its own timeout and leaves the child behind.
func TestCheckRunFinishesWhenStderrIsNotDrained(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	initRepo(t, dir)
	seedProject(t, dir)

	release := make(chan struct{})
	defer close(release)

	var out bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- runIn(dir, []string{"check", "--run", "echo hola"}, strings.NewReader(""), &out, blockedWriter{release: release})
	}()

	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("code = %d, want 0 (stdout: %q)", code, out.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("check --run did not finish: a stderr nobody drains blocked the command")
	}
}

// TestTheGateStreamsOnlyToATerminal is the decision itself: a buffer, a discard
// writer and a regular file lose nothing, because the proof is what keeps the
// bytes; only a character device receives the live copy.
func TestTheGateStreamsOnlyToATerminal(t *testing.T) {
	if got := streamTarget(&bytes.Buffer{}); got != nil {
		t.Fatalf("streamTarget(buffer) = %T, want nil: a buffer is not a terminal", got)
	}
	if got := streamTarget(io.Discard); got != nil {
		t.Fatalf("streamTarget(io.Discard) = %T, want nil", got)
	}
	file, err := os.CreateTemp(t.TempDir(), "stream")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer func() { _ = file.Close() }()
	if got := streamTarget(file); got != nil {
		t.Fatalf("streamTarget(regular file) = %T, want nil: only a character device is a terminal", got)
	}

	if devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0); err == nil {
		defer func() { _ = devNull.Close() }()
		if got := streamTarget(devNull); got == nil {
			t.Fatalf("streamTarget(character device) = nil, want the writer: a terminal receives the live copy")
		}
	}
}
