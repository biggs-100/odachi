package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/biggs-100/odachi/internal/app/start"
)

// terminalInterview is the Interview adapter: it prints one question block (the
// question, its why in two or three lines, and the options) and reads one answer
// line.
type terminalInterview struct {
	in  *bufio.Reader
	out io.Writer
}

// Ask shows the block and returns the raw line. The question and its why are
// never split into two messages: the why is the teaching, not a follow-up. The
// lines after the first are indented under "why: " so the block still reads as
// one piece.
func (t *terminalInterview) Ask(q start.Question) (string, error) {
	fmt.Fprintf(t.out, "\n? %s\n", q.Text)
	if q.Why != "" {
		why := strings.Split(q.Why, "\n")
		fmt.Fprintf(t.out, "  why: %s\n", why[0])
		for _, line := range why[1:] {
			fmt.Fprintf(t.out, "       %s\n", line)
		}
	}
	if len(q.Options) > 0 {
		fmt.Fprintf(t.out, "  options: %s\n", strings.Join(q.Options, ", "))
	}
	fmt.Fprint(t.out, "> ")

	line, err := t.in.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", fmt.Errorf("the input ended before the answer")
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// Say shows one line: a retry, a warning, or the batched answers block.
func (t *terminalInterview) Say(line string) error {
	_, err := fmt.Fprintln(t.out, line)
	return err
}
