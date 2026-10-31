package review

import (
	"fmt"
	"strings"
)

// This file owns the one thing a prompt has to fit: the size the project pays a
// reviewer for. A feature whose diff is big enough produces a prompt no model
// closes in the time odachi gives it -114.4 KB stalled the project's reviewer for
// 600s on 0132-, and the tool's answer was a refusal before spending, which is
// honest and leaves the feature reviewed by whatever model can take it. A
// candidate is a set of files, so it can be reviewed in parts instead.

// Split returns the prompts a candidate is reviewed in: one when the prompt fits
// the limit, and the parts that fit when it does not. limitKB is the project's
// max_prompt_kb; zero is the absence of a measured limit and never splits.
//
// A part carries its own paths, the hunks of those paths and its number in the
// header, and keeps the rows the store lends about its paths -plus a finding that
// blocks, which is the reason not to ship and not a row about a path-. The paths
// are handed out in the order git reported them, greedily, so two runs over the
// same candidate produce the same parts.
//
// A path whose own prompt cannot fit is an error naming it: splitting cannot help
// there, and a reviewer handed half a file is not reviewing it.
func Split(in PromptInput, limitKB int) ([]PromptInput, error) {
	whole := Prompt(in)
	if limitKB <= 0 || PromptSizeKB(whole) <= limitKB {
		return []PromptInput{in}, nil
	}

	files, err := diffFiles(in.Diff)
	if err != nil {
		return nil, fmt.Errorf("the prompt is %d KB and odd/review.toml allows %d KB, and the diff cannot be split: %w",
			PromptSizeKB(whole), limitKB, err)
	}
	hunks := make(map[string]string, len(files))
	for _, file := range files {
		hunks[file.path] = file.text
	}

	var parts []PromptInput
	var bin []string
	var carried strings.Builder

	// fits says whether one set of paths renders under the limit. It asks for the
	// longest header a part can carry ("of 99") so the answer is safe whatever the
	// final count is.
	fits := func(paths []string, diff string) bool {
		probe := in
		probe.Paths, probe.Diff = paths, diff
		probe.Part, probe.Of = 1, 99
		return PromptSizeKB(Prompt(probe)) <= limitKB
	}

	for _, path := range in.Paths {
		next := append(append([]string(nil), bin...), path)
		text := carried.String() + hunks[path]
		if fits(next, text) {
			bin, carried = next, strings.Builder{}
			carried.WriteString(text)
			continue
		}
		if len(bin) == 0 {
			return nil, tooBigForAPart(path, in, limitKB)
		}

		parts = append(parts, partOf(in, bin, carried.String()))
		bin, carried = nil, strings.Builder{}
		if !fits([]string{path}, hunks[path]) {
			return nil, tooBigForAPart(path, in, limitKB)
		}
		bin = []string{path}
		carried.WriteString(hunks[path])
	}
	if len(bin) > 0 {
		parts = append(parts, partOf(in, bin, carried.String()))
	}

	for i := range parts {
		parts[i].Part, parts[i].Of = i+1, len(parts)
	}
	// The header grows with the numbers it prints, so the bound is asked again with
	// the numbers in place: a part that no longer fits is a part the reviewer would
	// be handed over the limit.
	for _, part := range parts {
		if size := PromptSizeKB(Prompt(part)); size > limitKB {
			return nil, fmt.Errorf("the prompt of part %d is %d KB and odd/review.toml allows %d KB: raise max_prompt_kb, or review a smaller candidate",
				part.Part, size, limitKB)
		}
	}
	return parts, nil
}

// partOf is one part's input: the shared candidate, the paths it carries, their
// hunks, and the rows that talk about them.
func partOf(in PromptInput, paths []string, diff string) PromptInput {
	part := in
	part.Paths = append([]string(nil), paths...)
	part.Diff = diff
	part.Known = findingsFor(in.Known, paths)
	part.KnownPasses = passesFor(in.KnownPasses, paths)
	return part
}

// findingsFor keeps the findings a part can act on: the ones about its paths, and
// the ones that block. A blocker about a path another part carries is still the
// reason not to ship, and hiding it would let a part approve a tree that cannot
// ship.
func findingsFor(findings []Finding, paths []string) []Finding {
	if len(findings) == 0 {
		return findings
	}
	present := pathSet(paths)
	kept := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if present[RowPath(finding.Path)] || finding.Blocks() {
			kept = append(kept, finding)
		}
	}
	return kept
}

// passesFor keeps the clean rows about the paths this part carries.
func passesFor(passes []Pass, paths []string) []Pass {
	if len(passes) == 0 {
		return passes
	}
	present := pathSet(paths)
	kept := make([]Pass, 0, len(passes))
	for _, pass := range passes {
		if present[RowPath(pass.Path)] {
			kept = append(kept, pass)
		}
	}
	return kept
}

// pathSet is one spelling of a set of paths, the one coverage compares.
func pathSet(paths []string) map[string]bool {
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		set[RowPath(path)] = true
	}
	return set
}

// tooBigForAPart is the refusal of a file no limit can take: it names the file, the
// size it alone costs and the two ways out.
func tooBigForAPart(path string, in PromptInput, limitKB int) error {
	probe := in
	probe.Paths = []string{path}
	probe.Diff = ""
	for _, file := range mustDiffFiles(in.Diff) {
		if RowPath(file.path) == RowPath(path) {
			probe.Diff = file.text
		}
	}
	return fmt.Errorf("the diff of %s alone makes a %d KB prompt and odd/review.toml allows %d KB: raise max_prompt_kb, or review a smaller candidate",
		path, PromptSizeKB(Prompt(probe)), limitKB)
}

// diffFile is one file's slice of a diff.
type diffFile struct {
	path string
	text string
}

// diffFiles splits a unified diff at its `diff --git` headers. A header odachi
// cannot attribute to a path is refused: a part that does not say which file it
// carries is a slice nobody can name.
func diffFiles(diff string) ([]diffFile, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("the diff is empty")
	}

	var files []diffFile
	var current *diffFile
	for _, line := range strings.SplitAfter(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			path, ok := headerPath(line)
			if !ok {
				return nil, fmt.Errorf("the diff header %q does not name a file", strings.TrimSpace(line))
			}
			files = append(files, diffFile{path: path})
			current = &files[len(files)-1]
			current.text += line
			continue
		}
		if current == nil {
			return nil, fmt.Errorf("the diff starts before any file header: %q", strings.TrimSpace(strings.SplitN(line, "\n", 2)[0]))
		}
		current.text += line
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("the diff names no file")
	}
	return files, nil
}

// mustDiffFiles is diffFiles for a message: a refusal is already being built, and a
// diff that cannot be split has no part to measure.
func mustDiffFiles(diff string) []diffFile {
	files, _ := diffFiles(diff)
	return files
}

// headerPath reads the new path out of `diff --git a/x b/x`. Quoted paths (git
// quotes a name with a space or a non-ASCII byte) are refused rather than guessed.
func headerPath(line string) (string, bool) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	at := strings.LastIndex(line, " b/")
	if at < 0 {
		return "", false
	}
	path := line[at+len(" b/"):]
	if path == "" || strings.HasPrefix(path, "\"") || strings.Contains(path, "\"") {
		return "", false
	}
	return path, true
}
