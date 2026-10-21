// Package proof owns the proof file format: the two header lines the tool
// writes and never a human, followed by the verbatim output of the command that
// produced them.
//
// # tree <hash del worktree>
// # command <el comando que se corrió>
// <salida real, stdout+stderr, verbatim>
//
// A proof is only evidence if it says what it was produced from, so a file
// without the header is refused instead of guessed at. The format is pure
// string in, string out: this package reads no files and runs no commands.
package proof

import (
	"errors"
	"fmt"
	"strings"
)

// Headers are the exact prefixes of the two lines every proof starts with.
const (
	TreeHeader    = "# tree "
	CommandHeader = "# command "
)

// ErrNoHeader means the file does not say which tree produced it.
var ErrNoHeader = errors.New(`no tiene el header "# tree <hash>": no puedo saber de qué árbol salió`)

// Proof is one recorded run: the tree it was produced from, the command that
// produced it, and the combined stdout+stderr exactly as it came out.
type Proof struct {
	Tree    string
	Command string
	Output  string
}

// Render writes the canonical proof: tree, command, then the output verbatim.
// There is no timestamp and no absolute path, so two runs on the same tree with
// the same command produce the same bytes.
func Render(p Proof) string {
	var out strings.Builder
	out.WriteString(TreeHeader + p.Tree + "\n")
	out.WriteString(CommandHeader + p.Command + "\n")
	out.WriteString(p.Output)
	return out.String()
}

// Parse reads a proof file. Both header lines are required and must come first,
// in the canonical order: the tool writes them, and a file without them was
// written by hand or by something else. Everything after the second line is the
// output, byte for byte.
func Parse(text string) (Proof, error) {
	first, rest, ok := cutLine(text)
	if !ok || !strings.HasPrefix(first, TreeHeader) {
		return Proof{}, fmt.Errorf("proof: %w", ErrNoHeader)
	}

	second, output, ok := cutLine(rest)
	if !ok || !strings.HasPrefix(second, CommandHeader) {
		return Proof{}, fmt.Errorf(`proof: el archivo tiene "# tree" pero no el header "# command <el comando>": no puedo decir qué comando produjo la prueba`)
	}

	return Proof{
		Tree:    strings.TrimSpace(strings.TrimPrefix(first, TreeHeader)),
		Command: strings.TrimSpace(strings.TrimPrefix(second, CommandHeader)),
		Output:  output,
	}, nil
}

// Matches reports whether the proof was produced from this exact tree. An empty
// hash never matches: an unknown tree is not the same tree.
func (p Proof) Matches(treeHash string) bool {
	return treeHash != "" && p.Tree == treeHash
}

// LastLine is the last non-empty line of the output: what the human needs to
// read when the command failed.
func (p Proof) LastLine() string {
	lines := strings.Split(p.Output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(strings.TrimSuffix(lines[i], "\r")); line != "" {
			return line
		}
	}
	return ""
}

// cutLine splits the first line from the rest. The terminator is not part of the
// line; a CRLF file reports the same header as an LF file. ok is false when the
// text has no line at all.
func cutLine(text string) (line, rest string, ok bool) {
	if text == "" {
		return "", "", false
	}
	index := strings.IndexByte(text, '\n')
	if index < 0 {
		return strings.TrimSuffix(text, "\r"), "", true
	}
	return strings.TrimSuffix(text[:index], "\r"), text[index+1:], true
}
