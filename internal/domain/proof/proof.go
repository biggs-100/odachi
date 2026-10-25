// Package proof owns the proof file format: the two header lines the tool
// writes and never a human, followed by the verbatim output of the command that
// produced them.
//
// # tree <worktree hash>
// # command <the command that was run>
// # fence <ref>            (optional: CI, the policy the gate obeyed)
// # model <provider/name>  (optional: the model of the session that ran it)
// # output
// <real output, stdout+stderr, verbatim>
//
// The output terminator is what makes the output verbatim: without it, a
// command whose first line is "# model x" would lose that line to the header
// parser. Older proofs without the line still parse; the ambiguity stays for
// those files only, because what is already written is not rewritten.
//
// A proof is only evidence if it says what it was produced from, so a file
// without the header is refused instead of guessed at. The format is pure
// string in, string out: this package reads no files and runs no commands.
package proof

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Headers are the exact prefixes of the lines every proof starts with. Fence is
// optional: it names the ref the fence came from when the gate ran in CI against
// a policy the candidate cannot edit. Exit and Timeout belong to the evidence of
// a failed run, which is not a proof.
const (
	TreeHeader    = "# tree "
	CommandHeader = "# command "
	FenceHeader   = "# fence "
	// ModelHeader names the session that ran the command. It is what makes a
	// self-review visible: the verdict carries the model that judged, and the
	// same name on both sides is one agent approving itself.
	ModelHeader = "# model "
	// OutputHeader ends the headers. Everything after it is the command's
	// output, so a first line that looks like a header stays in the quotation.
	OutputHeader = "# output"
	ExitHeader   = "# exit "
	TimeoutHeader = "# timed out"
	// FailedHeader starts the evidence of a failed run. It is not a variant of
	// TreeHeader: a failed file never matches the first line of a proof, so Parse
	// refuses it before it can read a tree out of it.
	FailedHeader = "# failed "
)

// ErrNoHeader means the file does not say which tree produced it.
var ErrNoHeader = errors.New(`it has no "# tree <hash>" header: I cannot know which tree it came from`)

// Proof is one recorded run: the tree it was produced from, the command that
// produced it, and the combined stdout+stderr exactly as it came out.
type Proof struct {
	Tree    string
	Command string
	// Fence is the ref the fence was read from, empty when the proof used the
	// worktree's own fence.
	Fence string
	// Model is the model of the session that ran the command, empty when there
	// was no agent to name (a human at a terminal, or CI).
	Model  string
	Output string
}

// Render writes the canonical proof: tree, command, an optional fence ref, then
// the output verbatim. There is no timestamp and no absolute path, so two runs
// on the same tree with the same command produce the same bytes.
func Render(p Proof) string {
	var out strings.Builder
	out.WriteString(TreeHeader + p.Tree + "\n")
	out.WriteString(CommandHeader + p.Command + "\n")
	if p.Fence != "" {
		out.WriteString(FenceHeader + p.Fence + "\n")
	}
	if p.Model != "" {
		out.WriteString(ModelHeader + p.Model + "\n")
	}
	out.WriteString(OutputHeader + "\n")
	out.WriteString(p.Output)
	return out.String()
}

// RenderFailed writes the evidence of a run that did not pass: the tree and the
// command, an optional fence ref, then the exit code (or the timeout) and the
// output verbatim. It is not a proof —a proof is a green run— and its first line
// is `# failed <tree>`, which no proof starts with: Parse refuses it by
// construction, so the gate cannot read a failure as evidence of a pass.
func RenderFailed(p Proof, exitCode int, timedOut bool) string {
	var out strings.Builder
	out.WriteString(FailedHeader + p.Tree + "\n")
	out.WriteString(CommandHeader + p.Command + "\n")
	if p.Fence != "" {
		out.WriteString(FenceHeader + p.Fence + "\n")
	}
	if timedOut {
		out.WriteString(TimeoutHeader + "\n")
	} else {
		out.WriteString(ExitHeader + strconv.Itoa(exitCode) + "\n")
	}
	out.WriteString(OutputHeader + "\n")
	out.WriteString(p.Output)
	return out.String()
}

// Parse reads a proof file. Both header lines are required and must come first,
// in the canonical order; an optional "# fence <ref>" line may follow them. An
// old proof without the fence line keeps parsing: the candidate's own fence was
// the one it used. Everything after the headers is the output, byte for byte.
func Parse(text string) (Proof, error) {
	first, rest, ok := cutLine(text)
	if !ok || !strings.HasPrefix(first, TreeHeader) {
		return Proof{}, fmt.Errorf("proof: %w", ErrNoHeader)
	}

	second, output, ok := cutLine(rest)
	if !ok || !strings.HasPrefix(second, CommandHeader) {
		return Proof{}, fmt.Errorf(`proof: the file has "# tree" but not the "# command <the command>" header: I cannot tell which command produced the proof`)
	}

	fence := ""
	if line, rest, ok := cutLine(output); ok && strings.HasPrefix(line, FenceHeader) {
		fence = strings.TrimSpace(strings.TrimPrefix(line, FenceHeader))
		output = rest
	}

	// The terminator is optional: a proof written before it existed ends its
	// headers wherever they stop, and that file keeps its old reading.
	//
	// A model line only exists in the format with the terminator. In an older
	// proof a line that looks like one is part of the output — the old writer
	// never wrote that header — and dropping it would corrupt the quotation.
	model := ""
	if line, rest, ok := cutLine(output); ok && line == OutputHeader {
		output = rest
	} else if line, rest, ok := cutLine(output); ok && strings.HasPrefix(line, ModelHeader) {
		if next, after, ok := cutLine(rest); ok && next == OutputHeader {
			model = strings.TrimSpace(strings.TrimPrefix(line, ModelHeader))
			output = after
		}
	}

	return Proof{
		Tree:    strings.TrimSpace(strings.TrimPrefix(first, TreeHeader)),
		Command: strings.TrimSpace(strings.TrimPrefix(second, CommandHeader)),
		Fence:   fence,
		Model:   model,
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
