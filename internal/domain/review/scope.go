package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/schema"
)

// This file owns the record of a raised scope: the decision somebody made before a
// run judged anything.
//
// `--lens` adds lenses the evidence did not ask for, and the verdict only stamps the
// lenses it answered: a later reader sees four lenses and cannot tell whether the
// change asked for four or a person did, and a replay that drops the flag leaves no
// trace at all. The record is not the reviewer's answer, so it does not live in the
// verdict; it is the question that was asked, so it lives beside the verdict -the
// same shape 0091 gave the human's dismissal, and for the same reason: the verdict
// says what the reviewer said, and this file says what the run was asked to do.
//
// What it carries is what makes it checkable: which lenses the evidence required,
// which ones were added, the flag as it was given, and -when the caller supplied
// them- the words and the person. A record that raised nothing is refused, because
// the file exists to show a decision and a decision nobody made is noise.

// Scope is one run's widened scope: what the evidence asked for and what was added
// on top of it.
type Scope struct {
	// Schema is the version of this file, stamped by the tool. See
	// internal/domain/schema: it is what lets an older binary say "newer than me"
	// instead of naming a field it has never heard of as unknown.
	Schema int `json:"schema"`
	// Required is what the evidence asked for, in the fixed order: the lenses the
	// classifier would have run without anybody saying a word.
	Required []string `json:"required"`
	// Added is what --lens added beyond the evidence. It is never empty: a record
	// without it would say nothing.
	Added []string `json:"added"`
	// Asked is the flag exactly as the caller wrote it, because a person reading
	// the file later needs the input, not a reconstruction of it.
	Asked string `json:"asked"`
	// Because is the caller's reason, when they gave one. The tool never invents
	// it: a record that says why in the tool's words would be the tool talking.
	Because string `json:"because,omitempty"`
	// By is who ran it, when the caller knows.
	By string `json:"by,omitempty"`
	// At is when, in RFC3339.
	At string `json:"at,omitempty"`
}

// Raised is what a run added beyond the evidence: the lenses in extra that required
// does not already hold, in the fixed order, so the same scope always writes the
// same bytes.
func Raised(required, extra []string) []string {
	asked := make(map[string]bool, len(required))
	for _, lens := range required {
		asked[strings.ToLower(strings.TrimSpace(lens))] = true
	}

	added := make([]string, 0, len(extra))
	for _, lens := range extra {
		if !asked[strings.ToLower(strings.TrimSpace(lens))] {
			added = append(added, lens)
		}
	}
	return inOrder(added)
}

// ScopeName is the file a tree's scope lives in, beside its verdict.
func ScopeName(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of a scope: it must be a hash", tree)
	}
	return tree + ".scope.json", nil
}

// ScopePath is the scope's path relative to the clone's store.
func ScopePath(tree string) (string, error) {
	name, err := ScopeName(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// RenderScope writes the canonical file: indented JSON, fields in the documented
// order, the schema stamped, and a trailing newline. An empty required list is an
// empty array, not null: a person reading the file sees "the evidence asked for
// nothing" instead of having to guess.
func RenderScope(s Scope) []byte {
	canonical := s
	canonical.Schema = schema.Current
	if canonical.Required == nil {
		canonical.Required = []string{}
	}
	if canonical.Added == nil {
		canonical.Added = []string{}
	}

	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		return []byte("{}\n")
	}
	return append(data, '\n')
}

// ParseScope reads the file. Decoding is strict and the schema is read first, for
// the reason every file odachi writes carries one: a newer writer's key is not a
// typo. A record that raised nothing, or that cannot say which flag asked for it, is
// refused: the file is evidence, and evidence that says nothing is noise.
func ParseScope(data []byte) (Scope, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return Scope{}, errors.New("the scope record is empty: there is nothing to read")
	}

	var head struct {
		Schema int `json:"schema"`
	}
	if json.Unmarshal([]byte(text), &head) == nil {
		if _, err := schema.Read(head.Schema); err != nil {
			return Scope{}, fmt.Errorf("the scope record cannot be read: %w", err)
		}
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Scope
	if err := decoder.Decode(&parsed); err != nil {
		return Scope{}, fmt.Errorf("the scope record is not valid JSON: %s", oneLine(err.Error()))
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Scope{}, errors.New("the scope record is not a single JSON object: answer with the document only")
	}

	if parsed.Schema == 0 {
		parsed.Schema = schema.First
	}
	if parsed.Required == nil {
		parsed.Required = []string{}
	}
	if len(parsed.Added) == 0 {
		return Scope{}, errors.New("the scope record names no added lens: a record that raised nothing is not evidence")
	}
	if strings.TrimSpace(parsed.Asked) == "" {
		return Scope{}, errors.New("the scope record does not say which --lens asked for it: the flag is what a replay has to repeat")
	}
	for _, lens := range append(append([]string{}, parsed.Required...), parsed.Added...) {
		if !IsLens(lens) {
			return Scope{}, fmt.Errorf("the scope record names the lens %q, which does not exist: the lenses are %s", lens, strings.Join(order, ", "))
		}
	}
	if trimmed := strings.TrimSpace(parsed.At); trimmed != "" {
		if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
			return Scope{}, fmt.Errorf("the scope record says at %q, which is not RFC3339: use for example 2026-10-21T20:31:00Z", trimmed)
		}
	}
	return parsed, nil
}
