package check

import "strings"

// authoredLines counts the lines a candidate authors: every added line plus
// every deleted line of the unified diff between a feature's base and the tree
// that closes it. Both sides count because the reviewer reads both.
//
// What is not a line: the headers that name a file or mark a hunk (`diff --git`,
// `index`, `---`, `+++`, `@@`, the mode and rename pairs), and the context lines
// a hunk carries so it can be read. What is not authored work: any path under
// proof/, because the proofs are the tool's own evidence named by the tree that
// produced them; and any path the repository ignores, which never reaches this
// text because git left it out of the diff. An empty diff, a pure rename and a
// pure mode change count zero.
//
// A header is read as a header only where headers live: between `diff --git` and
// the first `@@` of that entry. Inside a hunk the first character is the whole
// grammar, so a deleted line whose text begins with `-- ` renders as `--- ...` and
// is still the author's line -counting it as a header would drop it, and reading
// its text as a path would rename the entry for the rest of the hunk (0331).
//
// The text is the diff Candidate reads for the feature's base: the same bytes
// the fence and the review judge, measured once so the gate and any later caller
// cannot disagree about what a candidate costs.
func authoredLines(diff string) int {
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
				if !proofPath(authoredPath(newPath, oldPath)) {
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
			if !proofPath(authoredPath(newPath, oldPath)) {
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

// proofPath reports the one path the count refuses by rule. It is the same
// exclusion the ignore file writes, kept here with a test so a candidate that
// forces a proof in does not pay for the tool's own evidence.
func proofPath(path string) bool {
	return strings.HasPrefix(path, "proof/")
}
