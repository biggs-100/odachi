// Package mutated owns the verdict and the file of the row that proves a guard by
// breaking it: the run names a file and a string, the string is taken out in a
// copy, the same command runs there, and the failure has to name what was expected.
//
// It is pure string in, string out: the package reads no files and runs no
// commands. The caller exports the tree, removes the string in the copy, runs the
// command and hands the exit code and the output over (0297).
package mutated

import (
	"fmt"
	"strings"
)

// Headers are the exact prefixes of the lines a mutated file starts with, in the
// canonical order. OutputHeader ends them: everything after it is the command's
// output, so a first line that looks like a header stays in the quotation.
const (
	MutatedHeader  = "# mutated "
	CommandHeader  = "# command "
	FileHeader     = "# file "
	RemovedHeader  = "# removed "
	ExpectedHeader = "# expected "
	VerdictHeader  = "# verdict "
	NoteHeader     = "# note "
	OutputHeader   = "# output"
)

// buildFailed is Go's own marker for a package that did not compile: a copy that
// does not build says nothing about the guard.
const buildFailed = "[build failed]"

// Verdict is what the row saw. Only two of them block a close.
type Verdict string

const (
	// Noticed is the answer the row exists for: the suite failed and its output
	// names the string the run expected. The guard fires when the source of truth
	// is gone.
	Noticed Verdict = "noticed"
	// Missed is a suite that stayed green with the string removed: the guard does
	// not exist, whatever its name says.
	Missed Verdict = "missed"
	// WrongReason is a suite that failed without naming the expected string. A red
	// for another cause is a red, not evidence about this guard.
	WrongReason Verdict = "wrong_reason"
	// Unknown is a copy that could not be judged: it did not build, or the command
	// never ran. It never blocks, and it never passes.
	Unknown Verdict = "unknown"
	// NoText is a string that is not in the file: a mutation that does not mutate
	// proves nothing, and the row is refused before anything runs.
	NoText Verdict = "no_text"
)

// Evidence is one mutated run: the tree it judged, the command, the file and the
// string that were taken out, the text the failure had to name, the verdict, an
// optional note, and the output exactly as it came out.
type Evidence struct {
	Tree     string
	Command  string
	File     string
	Removed  string
	Expected string
	Verdict  Verdict
	// Note carries the cause a verdict cannot: the timeout of a run that never
	// finished, or the sentence that says the string was not there.
	Note   string
	Output string
}

// Classify reads a finished run: green is the block the row exists for, a copy that
// does not compile is unknown, and a failure counts only when it names the expected
// text (0297).
func Classify(exitCode int, output, expected string) Verdict {
	switch {
	case exitCode < 0:
		return Unknown
	case exitCode == 0:
		return Missed
	case strings.Contains(output, buildFailed):
		return Unknown
	case strings.Contains(output, expected):
		return Noticed
	default:
		return WrongReason
	}
}

// Render writes the evidence: the headers in the canonical order, then the output
// verbatim.
func Render(e Evidence) string {
	var out strings.Builder
	out.WriteString(MutatedHeader + e.Tree + "\n")
	out.WriteString(CommandHeader + e.Command + "\n")
	out.WriteString(FileHeader + e.File + "\n")
	out.WriteString(RemovedHeader + e.Removed + "\n")
	out.WriteString(ExpectedHeader + e.Expected + "\n")
	out.WriteString(VerdictHeader + string(e.Verdict) + "\n")
	if e.Note != "" {
		out.WriteString(NoteHeader + e.Note + "\n")
	}
	out.WriteString(OutputHeader + "\n")
	out.WriteString(e.Output)
	return out.String()
}

// Parse reads a mutated file. A file that starts with `# tree` is a proof, and one
// that starts with `# undone` is the other row: both are refused by name, because a
// row read as another row is evidence nobody can trust.
func Parse(text string) (Evidence, error) {
	first, rest, ok := cutLine(text)
	if !ok || !strings.HasPrefix(first, MutatedHeader) {
		switch {
		case strings.HasPrefix(first, "# tree "):
			return Evidence{}, fmt.Errorf("mutated: the file is a proof, not a mutation: a proof is a green run, and this row removes a string")
		case strings.HasPrefix(first, "# undone "):
			return Evidence{}, fmt.Errorf("mutated: the file is the undo row, not a mutation: the undo reverts production files, and this row removes a string from one")
		}
		return Evidence{}, fmt.Errorf(`mutated: the file has no "# mutated <tree>" header: I cannot know which tree it judged`)
	}

	values := map[string]string{
		CommandHeader:  "",
		FileHeader:     "",
		RemovedHeader:  "",
		ExpectedHeader: "",
	}
	order := []string{CommandHeader, FileHeader, RemovedHeader, ExpectedHeader}
	for i, want := range order {
		line, after, ok := cutLine(rest)
		if !ok || !strings.HasPrefix(line, want) {
			return Evidence{}, fmt.Errorf("mutated: the headers are out of order: %q has to be the one after %q", want, at(i))
		}
		values[want] = strings.TrimSpace(strings.TrimPrefix(line, want))
		rest = after
	}

	verdictLine, rest, ok := cutLine(rest)
	if !ok || !strings.HasPrefix(verdictLine, VerdictHeader) {
		return Evidence{}, fmt.Errorf(`mutated: the file has no "# verdict" line: the output alone does not say what the row decided`)
	}
	verdict := Verdict(strings.TrimSpace(strings.TrimPrefix(verdictLine, VerdictHeader)))
	switch verdict {
	case Noticed, Missed, WrongReason, Unknown, NoText:
	default:
		return Evidence{}, fmt.Errorf("mutated: %q is not a verdict; the verdicts are noticed, missed, wrong_reason, unknown, no_text", verdict)
	}

	note := ""
	if line, after, ok := cutLine(rest); ok && strings.HasPrefix(line, NoteHeader) {
		note = strings.TrimSpace(strings.TrimPrefix(line, NoteHeader))
		next, _, ok := cutLine(after)
		if !ok || next != OutputHeader {
			return Evidence{}, fmt.Errorf(`mutated: the "# note" line is not followed by "# output": the headers are out of order`)
		}
		rest = after
	}

	terminator, output, ok := cutLine(rest)
	if !ok || terminator != OutputHeader {
		return Evidence{}, fmt.Errorf(`mutated: the file does not end its headers with "# output": the quotation has no beginning`)
	}

	return Evidence{
		Tree:     strings.TrimSpace(strings.TrimPrefix(first, MutatedHeader)),
		Command:  values[CommandHeader],
		File:     values[FileHeader],
		Removed:  values[RemovedHeader],
		Expected: values[ExpectedHeader],
		Verdict:  verdict,
		Note:     note,
		Output:   output,
	}, nil
}

// at is the header before this one, for a message that says where the file broke.
func at(i int) string {
	if i == 0 {
		return "the tree"
	}
	return []string{CommandHeader, FileHeader, RemovedHeader, ExpectedHeader}[i-1]
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
