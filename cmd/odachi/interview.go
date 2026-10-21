package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/biggz/odachi/internal/app/start"
)

// terminalInterview is the Interview adapter: it prints one question block (the
// question, its why on one line, and the options) and reads one answer line.
type terminalInterview struct {
	in  *bufio.Reader
	out io.Writer
}

// Ask shows the block and returns the raw line. The question and its why are
// never split into two messages: the why is the teaching, not a follow-up.
func (t *terminalInterview) Ask(q start.Question) (string, error) {
	fmt.Fprintf(t.out, "\n? %s\n", q.Text)
	if q.Why != "" {
		fmt.Fprintf(t.out, "  por qué: %s\n", q.Why)
	}
	if len(q.Options) > 0 {
		fmt.Fprintf(t.out, "  opciones: %s\n", strings.Join(q.Options, ", "))
	}
	fmt.Fprint(t.out, "> ")

	line, err := t.in.ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", fmt.Errorf("se terminó la entrada antes de la respuesta")
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// Say shows one line: a retry, a warning, or the batched answers block.
func (t *terminalInterview) Say(line string) error {
	_, err := fmt.Fprintln(t.out, line)
	return err
}
