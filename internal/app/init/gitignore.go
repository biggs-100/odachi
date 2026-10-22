package init

import (
	"path/filepath"
	"strings"
)

// The block `odachi init` keeps in the project's .gitignore. It is the reason the
// generated evidence never shifts the tree it certifies: proof/ and the review
// store live inside the worktree, and their own ignore files used to arrive one
// run late, invalidating the proof taken before them.
//
// The rules live in the project instead of inside the tool: they survive a clone,
// a human can read them, and `odachi init` is the single owner of the block.
const (
	ignoreFile  = ".gitignore"
	ignoreBegin = "# odachi:begin"
	ignoreEnd   = "# odachi:end"

	// The two directories the block keeps out of the tree, without the slashes
	// the rule spells them with.
	proofRule  = "proof"
	reviewRule = ".odachi/reviews"
)

// ignoreRules are the two generated directories. Both are written as directory
// rules with a leading slash, so each one means "this folder at the root of THIS
// project" and nothing else.
var ignoreRules = []string{"/" + proofRule + "/", "/" + reviewRule + "/"}

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
		problem := blocked("no pude revisar " + ignoreFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	var current []byte
	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("no pude leer " + ignoreFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
		current = data
	}

	updated, changed := withIgnoreBlock(current)
	if !changed {
		r.lines = append(r.lines, ignoreFile+": sin cambios, ya ignora /"+proofRule+"/ y /"+reviewRule+"/")
		return nil
	}

	if err := r.files.Write(path, updated); err != nil {
		problem := blocked("no pude escribir " + ignoreFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	action := "bloque actualizado"
	if !exists {
		action = "creado con el bloque de odachi"
	}
	r.changed = append(r.changed, ignoreFile)
	r.lines = append(r.lines, ignoreFile+": "+action+": /"+proofRule+"/ y /"+reviewRule+"/ no ensucian el árbol que certifican")
	return nil
}
