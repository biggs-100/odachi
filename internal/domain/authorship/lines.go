// Package authorship measures what a candidate costs the person who wrote it:
// the lines a diff adds and removes, minus what a command in this repository
// owns. It lives in the domain because two verbs ask the same question -the
// close weighs a candidate before paying a verifier, and opening a step weighs
// it before the plan grows- and a rule two verbs enforce cannot live inside one
// of them (0331, 0334).
package authorship

import "strings"

// Lines counts the lines a candidate authors: every added line plus every
// deleted line of the unified diff between a feature's base and the tree that
// stands. Both sides count because the reviewer reads both.
//
// What is not a line: the headers that name a file or mark a hunk (`diff --git`,
// `index`, `---`, `+++`, `@@`, the mode and rename pairs), and the context lines
// a hunk carries so it can be read. What is not authored work: any path under
// proof/, because the proofs are the tool's own evidence named by the tree that
// produced them; any path the repository ignores, which never reaches this text
// because git left it out of the diff; and the closed list of generated files in
// generatedPaths, because a command writes them and a feature that only reruns the
// generator would pay for bytes nobody typed (0331). An empty diff, a pure rename
// and a pure mode change count zero.
//
// A header is read as a header only where headers live: between `diff --git` and
// the first `@@` of that entry. Inside a hunk the first character is the whole
// grammar, so a deleted line whose text begins with `-- ` renders as `--- ...` and
// is still the author's line -counting it as a header would drop it, and reading
// its text as a path would rename the entry for the rest of the hunk (0331).
//
// The text is the diff Candidate reads for the feature's base: the same bytes
// the fence and the review judge, measured once so no two callers can disagree
// about what a candidate costs.
func Lines(diff string) int {
	total := 0
	oldPath, newPath := "", ""
	binary := false
	hunk := false
	for _, raw := range strings.Split(diff, "\n") {
		line := strings.TrimSuffix(raw, "\r")
		switch {
		case strings.HasPrefix(line, "diff --git "):
			oldPath, newPath, binary, hunk = "", "", false, false
		case binary:
			// A binary body is encoded, not authored lines to read.
		case strings.HasPrefix(line, "GIT binary patch"):
			binary = true
		case strings.HasPrefix(line, "@@"):
			// The first hunk of the entry closes the header block: every line
			// after it is content, whatever it starts with.
			hunk = true
		case hunk:
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				if !toolPath(authoredPath(newPath, oldPath)) {
					total++
				}
			}
		case strings.HasPrefix(line, "--- "):
			oldPath = diffPath(line[len("--- "):])
		case strings.HasPrefix(line, "+++ "):
			newPath = diffPath(line[len("+++ "):])
		case strings.HasPrefix(line, "index "),
			strings.HasPrefix(line, "new file mode "),
			strings.HasPrefix(line, "deleted file mode "),
			strings.HasPrefix(line, "old mode "),
			strings.HasPrefix(line, "new mode "),
			strings.HasPrefix(line, "similarity index "),
			strings.HasPrefix(line, "dissimilarity index "),
			strings.HasPrefix(line, "rename from "),
			strings.HasPrefix(line, "rename to "),
			strings.HasPrefix(line, "copy from "),
			strings.HasPrefix(line, "copy to "),
			strings.HasPrefix(line, "Binary files "):
			// Headers: they name the entry or its shape, they are not its lines.
		case strings.HasPrefix(line, "+"), strings.HasPrefix(line, "-"):
			if !toolPath(authoredPath(newPath, oldPath)) {
				total++
			}
		}
	}
	return total
}

// authoredPath names the file the lines land in: the new side, except where the
// entry deletes the file and the new side is /dev/null.
func authoredPath(newPath, oldPath string) string {
	if newPath != "" {
		return newPath
	}
	return oldPath
}

// diffPath reads the repository path a ---/+++ header names, without the a/ or b/
// prefix git adds and without the timestamp some producers append. The empty
// string is the absence of a file: /dev/null on one side of an add or a delete.
func diffPath(header string) string {
	path := header
	if i := strings.IndexByte(path, '\t'); i >= 0 {
		path = path[:i]
	}
	path = strings.Trim(path, "\"")
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	if path == "/dev/null" || path == "dev/null" {
		return ""
	}
	return path
}

// proofPath reports the directory the count refuses by rule: the proofs are the
// tool's own evidence, and it is the same exclusion the ignore file writes, kept
// here with a test so a candidate that forces a proof in does not pay for the
// tool's own evidence.
func proofPath(path string) bool {
	return strings.HasPrefix(path, "proof/")
}

// generatedPath is one file a generator in this repository owns: the command that
// writes it and what a hand edit costs, so a reader can check the claim without
// opening the tool. The two fields are documentation with a job -an entry nobody
// can trace to a writer is a hole in the count-.
type generatedPath struct {
	path  string
	owner string
	edit  string
}

// generatedPaths is the closed list of files the count charges nothing for. They
// are exact paths, one file at a time: a folder or a suffix would forgive the
// human's work that lives beside a generated file -`.agents/skills/<name>/SKILL.md`
// is where a person's own skill lives, and only the registry beside it is written
// by a command- (0331).
var generatedPaths = []generatedPath{
	{
		path:  "SKILLS.md",
		owner: "odachi skills --write",
		edit:  "its first line says it is generated: odachi skills --check reports the drift, odachi doctor names odachi skills --write as the fix, and the next --write replaces the file",
	},
	{
		path:  ".agents/skills/SKILLS.md",
		owner: "odachi skill new --scope project",
		edit:  "the same registry beside the skills a person writes: the next skill new rebuilds it from the files on disk, and a hand edit is lost",
	},
}

// generated reports whether a path is one of the files a command owns. The
// comparison is exact: the count forgives the file the generator writes and nothing
// that merely looks like it (0331).
func generated(path string) bool {
	for _, entry := range generatedPaths {
		if entry.path == path {
			return true
		}
	}
	return false
}

// toolPath reports whether a path's lines belong to the tool instead of to the
// author: the proofs and the generated files. It is the one question both branches
// of the count ask, so a new exclusion cannot be added to one and forgotten in the
// other (0331).
func toolPath(path string) bool {
	return proofPath(path) || generated(path)
}
