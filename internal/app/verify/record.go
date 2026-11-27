package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	verifydomain "github.com/biggs-100/odachi/internal/domain/verify"
)

// RecordSchema is the verification record this binary writes when the run was
// managed: the answer the verifier gave, sealed by the tool, with the runtime
// receipt the tool observed stamped beside it. Schema 1 is the unmanaged answer
// and keeps its exact bytes. The domain owns the number so the writer and the
// reader cannot drift.
const RecordSchema = verifydomain.SchemaRuntime

// VerificationRecord is one verification file: the schema-1 answer with the
// tool's seal, and, for a managed run, the validated runtime receipt.
type VerificationRecord struct {
	// Schema is the version of this file, stamped by the tool.
	Schema int `json:"schema"`
	// Tree is the tree this record speaks for. It is also the file's name.
	Tree string `json:"tree"`
	// By names who verified: the command or the model that answered.
	By string `json:"by"`
	// At is the instant the tool sealed this file.
	At string `json:"at"`
	// DeclaredAt is the hour the verifier wrote, kept only when it parses and
	// disagrees with At.
	DeclaredAt string `json:"declared_at,omitempty"`
	// Verdict is what the verifier answered: verified or refuted.
	Verdict string `json:"verdict"`
	// Notes are what the verifier wants read beside the answer.
	Notes []string `json:"notes"`
	// Runtime is the tool's receipt of the managed run. It is absent from a
	// schema-1 record, which claims no runtime.
	Runtime *SealedRuntimeReceipt `json:"runtime,omitempty"`
}

// SealManagedRecord stamps a managed answer into a schema-2 record. The receipt is
// validated first, the answer's own tree is checked against it, and the tool's
// hour is the record's hour. The author is the model the tool observed, so the
// answer contributes its verdict and notes, never its claim about the run or its
// own identity.
func SealManagedRecord(answer verifydomain.Record, receipt SealedRuntimeReceipt, now time.Time) (VerificationRecord, error) {
	if err := receipt.Validate(); err != nil {
		return VerificationRecord{}, fmt.Errorf("the runtime receipt is not a sealed fact: %w", err)
	}
	sealed := answer.Seal(now)
	switch tree := strings.TrimSpace(sealed.Tree); {
	case tree == "":
		// The tool read the tree; an answer that names none is stamped with it.
		sealed.Tree = receipt.Tree
	case !sealed.Covers(receipt.Tree):
		return VerificationRecord{}, fmt.Errorf("the answer names the tree %s and the runtime receipt names %s: a record for another tree covers nothing", tree, receipt.Tree)
	}

	record := VerificationRecord{
		Schema: RecordSchema,
		Tree:   sealed.Tree,
		// The author is the model the tool observed, never the identity the
		// answer claims: a model that names itself is not a fact.
		By:         receipt.SessionModel,
		At:         sealed.At,
		DeclaredAt: sealed.DeclaredAt,
		Verdict:    sealed.Verdict,
		Notes:      sealed.Notes,
		Runtime:    &receipt,
	}
	if err := record.Validate(); err != nil {
		return VerificationRecord{}, err
	}
	return record, nil
}

// Validate refuses a record the tool cannot stand behind. The domain owns both
// schemas: schema 1 is the unmanaged answer and schema 2 must carry one validated
// verifier receipt for the same tree.
func (record VerificationRecord) Validate() error {
	return verifydomain.Record{
		Schema:     record.Schema,
		Tree:       record.Tree,
		By:         record.By,
		At:         record.At,
		DeclaredAt: record.DeclaredAt,
		Verdict:    record.Verdict,
		Notes:      record.Notes,
		Runtime:    record.Runtime,
	}.Validate()
}

// Covers reports whether this record speaks for the given tree.
func (record VerificationRecord) Covers(tree string) bool {
	return strings.EqualFold(strings.TrimSpace(record.Tree), strings.TrimSpace(tree))
}

// ParseVerificationRecord reads either schema. Schema 1 goes through the domain
// reader, so a legacy record reads exactly as it always did; schema 2 is decoded
// strictly and validated, receipt included.
func ParseVerificationRecord(data []byte) (VerificationRecord, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return VerificationRecord{}, errors.New("the verification is empty: there is nothing to read")
	}

	var head struct {
		Schema int `json:"schema"`
	}
	if json.Unmarshal([]byte(text), &head) == nil && head.Schema > RecordSchema {
		return VerificationRecord{}, fmt.Errorf("the verification declares schema %d and this binary reads %d: upgrade odachi", head.Schema, RecordSchema)
	}

	if head.Schema == RecordSchema {
		decoder := json.NewDecoder(strings.NewReader(text))
		decoder.DisallowUnknownFields()

		var parsed VerificationRecord
		if err := decoder.Decode(&parsed); err != nil {
			return VerificationRecord{}, fmt.Errorf("the verification is not valid JSON: %s", oneLine(err.Error()))
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			return VerificationRecord{}, errors.New("the verification is not a single JSON object: answer with the record only")
		}
		if parsed.Notes == nil {
			parsed.Notes = []string{}
		}
		if err := parsed.Validate(); err != nil {
			return VerificationRecord{}, err
		}
		return parsed, nil
	}

	legacy, err := verifydomain.Parse([]byte(text))
	if err != nil {
		return VerificationRecord{}, err
	}
	record := VerificationRecord{
		Schema:     legacy.Schema,
		Tree:       legacy.Tree,
		By:         legacy.By,
		At:         legacy.At,
		DeclaredAt: legacy.DeclaredAt,
		Verdict:    legacy.Verdict,
		Notes:      legacy.Notes,
		Runtime:    legacy.Runtime,
	}
	if err := record.Validate(); err != nil {
		return VerificationRecord{}, err
	}
	return record, nil
}

// RenderVerificationRecord writes the canonical file: indented JSON, an empty
// notes array instead of null, and a trailing newline. A schema-1 record renders
// the same bytes the domain renders.
func RenderVerificationRecord(record VerificationRecord) ([]byte, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}
	canonical := record
	if canonical.Notes == nil {
		canonical.Notes = []string{}
	}
	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		// Unreachable: every field is a string, an int, a slice of strings or the
		// receipt, which encodes as an object of strings.
		return nil, errors.New("could not encode the verification record")
	}
	return append(data, '\n'), nil
}
