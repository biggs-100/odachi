// Package schema names the version of the files odachi writes and reads, so an
// older binary that meets a newer file says "upgrade" instead of calling a file
// that is perfectly well formed broken.
//
// It exists because three decoders -the verdict, agents.toml and
// odd/review.toml- are strict on purpose, and strictness cannot tell a typo from
// a newer writer. A field that arrived after the installed binary did reads as an
// unknown key, and the message sends the human to fix a file that is not broken.
// With a version written in the file, the binary that knows the field can answer
// the true thing: this file is newer than me.
package schema

import "fmt"

const (
	// First is the schema of every file written before the field existed. An
	// absent value means this one, today and always: the day Current moves, the
	// oldest files do not silently gain fields they never had.
	First = 1
	// Current is the schema this binary writes, and the newest it understands.
	//
	// One counter serves the three files on purpose: they are one answer to the
	// same question - can this binary read what is on disk - and a per-file
	// counter would let a file the binary cannot fully read pass a check that
	// only looks at its own number. The cost is accepted and paid at each move:
	// a key added to one file makes all three declare Current, so a binary older
	// than this one refuses a verdict whose shape did not change. The words are
	// still true where it matters - the file on disk is newer than that binary -
	// and the upgrade is the same command either way.
	Current = 2
)

// Read answers which schema a file with this value was written with. A value this
// binary does not know is refused in the words that name the way out: the file is
// not broken, the binary is old.
//
// A written zero is read as absence, because no odachi writes it and a decoder
// that counts cannot tell the two apart; the field is a number, not a pointer,
// and a schema-0 file would be shaped like the first schema anyway.
func Read(written int) (int, error) {
	switch {
	case written == 0:
		return First, nil
	case written < First:
		return 0, fmt.Errorf("schema %d is not a version odachi writes; remove the line or write schema %d", written, First)
	case written > Current:
		return 0, fmt.Errorf("this file was written by a newer odachi (schema %d); this binary reads schema %d - upgrade odachi, or use the binary that wrote it", written, Current)
	}
	return written, nil
}
