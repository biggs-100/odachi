// Package undone owns the verdict and the file of the row that reverts a
// change: did the suite notice, stay green, or could it not be judged?
//
// It is pure string in, string out: the package reads no files and runs no
// commands. The caller reverts the code in a copy, runs the command there and
// hands the exit code and the output over (0052).
package undone

import (
	"fmt"
	"strings"
)

// Headers are the exact prefixes of the lines an undone file starts with, in the
// canonical order. OutputHeader ends them: everything after it is the command's
// output, so a first line that looks like a header stays in the quotation.
const (
	UndoneHeader  = "# undone "
	CommandHeader = "# command "
	VerdictHeader = "# verdict "
	OutputHeader  = "# output"
)

// buildFailed is Go's own marker for a package that did not compile. It is the
// one marker this package knows: a command that is not a suite has no such line,
// and its non-zero reads as noticed (0052).
const buildFailed = "[build failed]"

// Verdict is what the row saw. Only one of them blocks a close.
type Verdict string

const (
	// Missed is a suite that stayed green with the change undone: the tests do
	// not notice the code being reverted, and this is the row's only block.
	Missed Verdict = "missed"
	// Noticed is a suite that failed with the change undone: the tests depend on
	// the code, whether they quoted a line or the package refused to build for
	// another reason.
	Noticed Verdict = "noticed"
	// Unknown is a copy that could not be judged: it did not build, or the
	// command never ran. It is neither a pass nor a miss, and it never blocks.
	Unknown Verdict = "unknown"
)

// Evidence is one undone run: the tree it judged, the command, the verdict and
// the output exactly as it came out.
type Evidence struct {
	Tree    string
	Command string
	Verdict Verdict
	Output  string
}

// Classify reads a finished run. An exit of zero is the block the row exists
// for; a build failure is unknown, because a copy that does not compile says
// nothing about the tests; any other non-zero is the suite noticing.
func Classify(exitCode int, output string) Verdict {
	switch {
	case exitCode < 0:
		return Unknown
	case exitCode == 0:
		return Missed
	case strings.Contains(output, buildFailed):
		return Unknown
	default:
		return Noticed
	}
}

// Render writes the evidence: the three header fields, then the output verbatim.
func Render(e Evidence) string {
	var out strings.Builder
	out.WriteString(UndoneHeader + e.Tree + "\n")
	out.WriteString(CommandHeader + e.Command + "\n")
	out.WriteString(VerdictHeader + string(e.Verdict) + "\n")
	out.WriteString(OutputHeader + "\n")
	out.WriteString(e.Output)
	return out.String()
}

// Parse reads an undone file. A file that starts with `# tree` is a proof, not
// an undone, and is refused by name.
func Parse(text string) (Evidence, error) {
	first, rest, ok := cutLine(text)
	if !ok || !strings.HasPrefix(first, UndoneHeader) {
		if strings.HasPrefix(first, "# tree ") {
			return Evidence{}, fmt.Errorf("undone: the file is a proof, not an undone run: a proof is a green run, and this row's file is a reverted one")
		}
		return Evidence{}, fmt.Errorf(`undone: the file has no "# undone <tree>" header: I cannot know which tree it judged`)
	}

	second, rest, ok := cutLine(rest)
	if !ok || !strings.HasPrefix(second, CommandHeader) {
		return Evidence{}, fmt.Errorf(`undone: the file has "# undone" but not the "# command <the command>" line: I cannot tell what ran`)
	}

	third, rest, ok := cutLine(rest)
	if !ok || !strings.HasPrefix(third, VerdictHeader) {
		return Evidence{}, fmt.Errorf(`undone: the file has no "# verdict missed|noticed|unknown" line: the output alone does not say what the row decided`)
	}
	verdict := Verdict(strings.TrimSpace(strings.TrimPrefix(third, VerdictHeader)))
	switch verdict {
	case Missed, Noticed, Unknown:
	default:
		return Evidence{}, fmt.Errorf("undone: %q is not a verdict; the verdicts are missed, noticed, unknown", verdict)
	}

	fourth, output, ok := cutLine(rest)
	if !ok || fourth != OutputHeader {
		return Evidence{}, fmt.Errorf(`undone: the file does not end its headers with "# output": the quotation has no beginning`)
	}

	return Evidence{
		Tree:    strings.TrimSpace(strings.TrimPrefix(first, UndoneHeader)),
		Command: strings.TrimSpace(strings.TrimPrefix(second, CommandHeader)),
		Verdict: verdict,
		Output:  output,
	}, nil
}

// cutLine splits the first line from the rest. The terminator is not part of the
// line; a CRLF file reports the same header as an LF file.
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
