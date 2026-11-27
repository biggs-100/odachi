// Package verify owns the record one independent verifier leaves about one tree.
//
// Three ideas hold it together.
//
// A verification is a RECORD, not a step somebody marked done. The tool seals the
// tree, the author and the hour into a file, the way it seals a verdict, so the
// gate compares bytes instead of trusting a sentence in a task document.
//
// A record is CONTENT-ADDRESSED. It lives in the clone's store under the name of
// the tree it verified, so "a record of another tree does not cover this one" is
// the path of the file and not a rule somebody has to apply.
//
// The author is a DECLARATION the record cannot drop. A verification that names
// nobody is the tree telling itself it looks fine, and the gate refuses it for the
// same reason it refuses an unattributed verdict: an identity nobody declared is
// an identity nobody can check.
//
// The package is pure: bytes and time in, bytes and time out. It reads no files,
// runs no commands and knows nothing about git; the tree is a fact the caller
// brings.
package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/oneline"
)

// The store: one file per tree, named by the tree it verified. It lives in the
// clone's state, outside the worktree, so writing a record cannot change the tree
// that names it.
const (
	// Dir is where records live, relative to the clone's store. It is a folder of
	// its own beside reviews/, because a verification and a verdict answer
	// different questions about the same bytes.
	Dir = "verifications"
	// Schema is the version of this file, and this package's own counter. The
	// verdict's counter is shared by five files; a sixth file that starts here
	// moves none of them, and a binary that never heard of this file never reads
	// it.
	Schema = 1
	// SchemaRuntime is the managed record: the same answer, sealed by the tool,
	// with the runtime receipt the tool observed stamped beside it. Schema 1 is
	// the unmanaged answer and keeps its exact bytes.
	SchemaRuntime = 2
)

// The two answers a verifier gives. They are a closed vocabulary: a third word is
// not a verification, and a record that carries one is refused before it is
// written.
const (
	Verified = "verified"
	Refuted  = "refuted"
)

// Record is what one verifier concluded about one tree.
type Record struct {
	// Schema is the version of this file, stamped by the tool and never by the
	// verifier, the way the tree is.
	Schema int `json:"schema"`
	// Tree is the tree this record speaks for. It is also the file's name.
	Tree string `json:"tree"`
	// By names who verified: the command or the model that answered. The gate
	// refuses a record without it.
	By string `json:"by"`
	// At is the instant the tool sealed this file: the clock of the clone that
	// wrote it, never the verifier's word.
	At string `json:"at"`
	// DeclaredAt is the time the verifier wrote, kept only when it parses and
	// disagrees with At, so a wrong clock stays visible instead of erased.
	DeclaredAt string `json:"declared_at,omitempty"`
	// Verdict is what the verifier answered: verified or refuted.
	Verdict string `json:"verdict"`
	// Notes are what the verifier wants read beside the answer. A refutation that
	// says nothing is not a refusal of anything, and this is where it says why.
	Notes []string `json:"notes"`
	// Runtime is the tool's receipt of the managed run. It is absent from a
	// schema-1 record, which claims no runtime.
	Runtime *Receipt `json:"runtime,omitempty"`
}

// Parse reads a record. Decoding is strict -an unknown field is refused, a second
// JSON object is refused- because a record the tool half-understands is worse than
// no record: it would be stored under a name that claims more than it can back.
func Parse(data []byte) (Record, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return Record{}, errors.New("the verification is empty: there is nothing to save")
	}

	// The version is asked before the strict pass: a record from a newer odachi
	// carries fields this binary does not know, and calling them unknown sends the
	// human to fix a file that is not broken.
	var head struct {
		Schema int `json:"schema"`
	}
	if json.Unmarshal([]byte(text), &head) == nil && head.Schema > SchemaRuntime {
		return Record{}, fmt.Errorf("the verification declares schema %d and this binary writes 1 and %d: upgrade odachi", head.Schema, SchemaRuntime)
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Record
	if err := decoder.Decode(&parsed); err != nil {
		return Record{}, fmt.Errorf("the verification is not valid JSON: %s", oneLine(err.Error()))
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Record{}, errors.New("the verification is not a single JSON object: answer with the record only")
	}

	if parsed.Schema == 0 {
		parsed.Schema = Schema
	}
	if parsed.Notes == nil {
		parsed.Notes = []string{}
	}
	return parsed, nil
}

// Render writes the canonical file: indented JSON, an empty notes array instead of
// null, and a trailing newline. Two renders of the same record are the same bytes.
func Render(r Record) []byte {
	canonical := r
	if canonical.Notes == nil {
		canonical.Notes = []string{}
	}
	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		// Unreachable: every field is a string, an int, or a slice of strings.
		return []byte("{}\n")
	}
	return append(data, '\n')
}

// Validate reports what only the author of a record can fix: who verified, which
// tree it speaks for, what it answered, and the hour the tool wrote.
func (r Record) Validate() error {
	if trimmed := strings.TrimSpace(r.Tree); trimmed == "" {
		return errors.New("the verification names no tree: a record that does not say what it verified covers nothing")
	} else if !hashLike(trimmed) {
		return fmt.Errorf("the verification names %q as its tree, which is not a hash", trimmed)
	}
	if strings.TrimSpace(r.By) == "" {
		return errors.New(`the verification does not say who verified: add "by" with the command or the model that answered`)
	}
	if trimmed := strings.TrimSpace(r.At); trimmed == "" {
		return errors.New("the verification carries no hour: the tool seals it, and a record without one cannot be ordered with the tree")
	} else if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
		return fmt.Errorf("the verification says at %q, which is not RFC3339: use for example 2026-10-21T20:31:00Z", trimmed)
	}
	switch strings.TrimSpace(r.Verdict) {
	case Verified, Refuted:
	default:
		return fmt.Errorf(`the verification says %q; the only answers are %q and %q`, r.Verdict, Verified, Refuted)
	}

	switch r.Schema {
	case 0, Schema:
		if r.Runtime != nil {
			return errors.New("the verification declares schema 1 and carries a runtime receipt: schema 1 is the unmanaged answer, and it claims no runtime")
		}
	case SchemaRuntime:
		if r.Runtime == nil {
			return errors.New("the verification declares schema 2 and carries no runtime receipt: a managed verification stamps the tool's receipt beside the answer")
		}
		if err := r.Runtime.Validate(); err != nil {
			return fmt.Errorf("the verification's runtime receipt is not sealed: %w", err)
		}
		if r.Runtime.Role != RoleVerifier {
			return fmt.Errorf("the verification carries a %q receipt; a verification carries the %q receipt", r.Runtime.Role, RoleVerifier)
		}
		if !strings.EqualFold(r.Runtime.Tree, r.Tree) {
			return fmt.Errorf("the verification names the tree %s and its runtime receipt names %s: they must be the same bytes", r.Tree, r.Runtime.Tree)
		}
		// The author of a managed verification is the model the tool observed.
		// A record whose author differs from the receipt is a claim the receipt
		// does not back, so it is refused whole.
		if r.By != r.Runtime.SessionModel {
			return fmt.Errorf("the verification names %q as its author and its runtime receipt observed %q: a managed verification is authored by the model the tool observed", r.By, r.Runtime.SessionModel)
		}
	default:
		return fmt.Errorf("the verification declares schema %d; this binary writes schema 1 or %d", r.Schema, SchemaRuntime)
	}
	return nil
}

// Seal copies the record with the tool's hour in At, and keeps the time the
// verifier wrote in DeclaredAt when it parses and disagrees. The caller's record is
// not touched.
func (r Record) Seal(now time.Time) Record {
	sealed := r
	sealed.Schema = Schema
	claimed := strings.TrimSpace(sealed.At)
	sealed.At = now.UTC().Format(time.RFC3339)
	sealed.DeclaredAt = ""
	if claimed != "" && claimed != sealed.At {
		if _, err := time.Parse(time.RFC3339, claimed); err == nil {
			sealed.DeclaredAt = claimed
		}
	}
	return sealed
}

// Covers reports whether this record speaks for the given tree. A verification of
// another tree says nothing about this one: the tree is what was read.
func (r Record) Covers(tree string) bool {
	return strings.EqualFold(strings.TrimSpace(r.Tree), strings.TrimSpace(tree))
}

// Name is the file a record for one tree lives in. The tree is the name.
func Name(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of a verification: it must be a hash", tree)
	}
	return tree + ".json", nil
}

// Path is the record's path relative to the clone's store, with forward slashes so
// the store has one spelling on every platform.
func Path(tree string) (string, error) {
	name, err := Name(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// hashLike reports whether a value can name a tree: the hex git writes, whole.
func hashLike(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// oneLine keeps a decoder's message inside the sentence that carries it.
func oneLine(text string) string {
	return oneline.Line(text)
}
