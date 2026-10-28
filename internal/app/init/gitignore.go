package init

import (
	"path/filepath"
	"strings"
)

// The block `odachi init` keeps in the project's .gitignore. It is the reason the
// generated evidence never shifts the tree it certifies: proof/ lives inside the
// worktree, and its own ignore file used to arrive one run late, invalidating the
// proof taken before it. The backups `plan` and `ask --craft` write next to the
// task doc are generated too: they are undo material, not project content, and a
// clean `git status` is part of the deal.
//
// The verdicts are not here: they live in the clone's store, outside the worktree,
// so no ignore rule has to keep them out of the tree they describe.
//
// The rules live in the project instead of inside the tool: they survive a clone,
// a human can read them, and `odachi init` is the single owner of the block.
const (
	ignoreFile  = ".gitignore"
	ignoreBegin = "# odachi:begin"
	ignoreEnd   = "# odachi:end"

	// The one directory the block keeps out of the tree, without the slashes the
	// rule spells it with, and the backup suffix every rewrite leaves.
	proofRule  = "proof"
	backupRule = "*.bak-*"
)

// ignoreRules is the generated directory and the backup suffix. The directory is
// written with a leading slash, so it means "this folder at the root of THIS
// project" and nothing else; the backup pattern matches the `file.bak-<8 hex>`
// files `undo` reads.
var ignoreRules = []string{"/" + proofRule + "/", backupRule}

// ignoreBlock is the canonical content of the block, trailing newline included.
func ignoreBlock() string {
	return ignoreBegin + "\n" + strings.Join(ignoreRules, "\n") + "\n" + ignoreEnd + "\n"
}

// withIgnoreBlock returns the .gitignore that carries the canonical block, and
// whether that differs from current. current is nil when the file does not exist.
//
// The rules, in the order they were decided:
//   - no file: the block alone, with its trailing newline.
//   - a block already there: replaced in place, so an upgrade refreshes it and
//     whatever the human wrote above and below survives.
//   - no block: appended after the existing content, after exactly one blank
//     line. A file whose last line has no newline keeps every byte and is not
//     glued to the marker.
//
// Running it twice changes nothing the second time, which is what makes a rerun
// of `odachi init` write nothing at all.
func withIgnoreBlock(current []byte) ([]byte, bool) {
	block := ignoreBlock()
	text := string(current)

	if text == "" {
		return []byte(block), true
	}

	if start, end, ok := ignoreSpan(text); ok {
		updated := text[:start] + block + text[end:]
		return []byte(updated), updated != text
	}

	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return []byte(text + "\n" + block), true
}

// ignoreSpan finds the bytes the block owns: from the begin marker's line to the
// end of the end marker's line. A begin marker with no end marker is a truncated
// block, and its tail belongs to the block too, so the replacement repairs it
// instead of leaving a second marker behind.
func ignoreSpan(text string) (start, end int, ok bool) {
	start = lineStart(text, ignoreBegin)
	if start < 0 {
		return 0, 0, false
	}

	rest := text[start:]
	offset := strings.Index(rest, ignoreEnd)
	if offset < 0 {
		return start, len(text), true
	}

	end = start + offset + len(ignoreEnd)
	if strings.HasPrefix(text[end:], "\n") {
		end++
	}
	return start, end, true
}

// lineStart is the index of the marker when it owns its own line, and -1 when it
// does not: a marker inside another line is not a block this tool wrote.
func lineStart(text, marker string) int {
	for from := 0; from < len(text); {
		index := strings.Index(text[from:], marker)
		if index < 0 {
			return -1
		}
		index += from
		if index == 0 || text[index-1] == '\n' {
			return index
		}
		from = index + len(marker)
	}
	return -1
}

// writeIgnore keeps the block in the project's .gitignore, writing only when it
// reports a change. The human's file is never backed up for this: the block is
// additive and the human's own lines are preserved byte for byte, so a backup
// would only add noise to the project.
func (r *runner) writeIgnore() *Result {
	path := filepath.Join(r.dir, ignoreFile)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("could not check " + ignoreFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
		return &problem
	}

	var current []byte
	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("could not read " + ignoreFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
			return &problem
		}
		current = data
	}

	updated, changed := withIgnoreBlock(current)
	if !changed {
		r.lines = append(r.lines, ignoreFile+": unchanged, it already ignores /"+proofRule+"/ and the "+backupRule+" backups")
		return nil
	}

	if err := r.files.Write(path, updated); err != nil {
		problem := blocked("could not write " + ignoreFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
		return &problem
	}

	action := "block updated"
	if !exists {
		action = "created with the odachi block"
	}
	r.changed = append(r.changed, ignoreFile)
	r.lines = append(r.lines, ignoreFile+": "+action+": /"+proofRule+"/ does not dirty the tree it certifies, and the "+backupRule+" backups stay out of git")
	return nil
}
