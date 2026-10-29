package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// This file owns the human's dismissal: the act that retires one row of a verdict
// without rewriting the verdict.
//
// A reviewer's finding can be dismissed inside the answer - a status, a reason and
// the fingerprint of what it judged - but only the reviewer writes that. The person
// paying for the review had no way to say "this row is wrong" about a warning, and a
// warning is never refutable, so a false one traveled forever: a potent row leaves
// the store only when the bytes it names change. So the human gets the same shape as
// the reviewer's dismissal, stored beside the verdict like a refutation and applied
// when the store is read. The verdict is never edited: it says what the reviewer
// said, and this file says what the human judged.
//
// What it carries is what makes it checkable: the fingerprint binds it to one row,
// and the reason is the human's evidence. A dismissal without a reason is the
// silent edit this exists to replace.

// Dismissal is one row the human retired: what it judged, why, who, and when.
type Dismissal struct {
	// Fingerprint is the identity of the row, computed by the tool from the
	// finding it matched: the same value a reviewer's dismissal carries, which is
	// what makes it travel to a later candidate with unchanged bytes.
	Fingerprint string `json:"fingerprint"`
	// Path and Line are the row's place, for a person reading the file.
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
	// Reason is the human's evidence. It is required: a dismissal without one is
	// silence wearing a decision's clothes.
	Reason string `json:"reason"`
	// By is who judged, when the caller knows it.
	By string `json:"by,omitempty"`
	// At is when, in RFC3339.
	At string `json:"at,omitempty"`
}

// Dismissals is the file beside a verdict: the rows the human retired for it. The
// key is the one that makes Sniffing unambiguous against a verdict or a refutation.
type Dismissals struct {
	Rows []Dismissal `json:"dismissed"`
}

// DismissedName is the file a tree's dismissals live in, beside its verdict.
func DismissedName(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of dismissals: it must be a hash", tree)
	}
	return tree + ".dismissed.json", nil
}

// DismissedPath is the dismissals' path relative to the clone's store.
func DismissedPath(tree string) (string, error) {
	name, err := DismissedName(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// ParseDismissals reads the file. Decoding is strict and a verdict is refused in
// its place: a document the tool half-understands would retire rows nobody judged.
func ParseDismissals(data []byte) (Dismissals, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return Dismissals{}, errors.New("the dismissals are empty: there is nothing to save")
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Dismissals
	if err := decoder.Decode(&parsed); err != nil {
		return Dismissals{}, fmt.Errorf("the dismissals are not valid JSON: %s", oneLine(err.Error()))
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Dismissals{}, errors.New("the dismissals are not a single JSON object: answer with the document only")
	}

	seen := make(map[string]bool, len(parsed.Rows))
	for i, row := range parsed.Rows {
		if !hashLike(strings.TrimSpace(row.Fingerprint)) {
			return Dismissals{}, fmt.Errorf("dismissal %d says fingerprint %q; a fingerprint is hex like a hash, and one that is not names nothing", i+1, row.Fingerprint)
		}
		if strings.TrimSpace(row.Path) == "" {
			return Dismissals{}, fmt.Errorf("dismissal %d does not say path: without the file it judged nothing", i+1)
		}
		if strings.TrimSpace(row.Reason) == "" {
			return Dismissals{}, fmt.Errorf("dismissal %d (%s) does not say why: a dismissal without evidence is silence, not a judgment", i+1, row.Path)
		}
		if row.Line < 0 {
			return Dismissals{}, fmt.Errorf("dismissal %d (%s) says line %d; use 0 for the whole file or the line number", i+1, row.Path, row.Line)
		}
		if trimmed := strings.TrimSpace(row.At); trimmed != "" {
			if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
				return Dismissals{}, fmt.Errorf("dismissal %d says at %q, which is not RFC3339: use for example 2026-10-21T20:31:00Z", i+1, trimmed)
			}
		}
		if seen[row.Fingerprint] {
			return Dismissals{}, fmt.Errorf("dismissal %d repeats the row %s: one judgment per row", i+1, row.Fingerprint)
		}
		seen[row.Fingerprint] = true
	}
	if parsed.Rows == nil {
		parsed.Rows = []Dismissal{}
	}
	return parsed, nil
}

// RenderDismissals writes the canonical file: indented JSON, the rows in the order
// they were added, and a trailing newline, so two renders of the same decision are
// the same bytes.
func RenderDismissals(d Dismissals) []byte {
	canonical := d
	if canonical.Rows == nil {
		canonical.Rows = []Dismissal{}
	}
	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		return []byte("{}\n")
	}
	return append(data, '\n')
}

// Retires reports whether this row was retired by the human, and with which
// judgment. It matches on the fingerprint, which is the row's identity: a
// dismissal written for one defect never retires another.
func (d Dismissals) Retires(finding Finding) (Dismissal, bool) {
	fingerprint := strings.TrimSpace(finding.Fingerprint)
	if fingerprint == "" {
		fingerprint = Fingerprint(finding)
	}
	for _, row := range d.Rows {
		if strings.EqualFold(strings.TrimSpace(row.Fingerprint), fingerprint) {
			return row, true
		}
	}
	return Dismissal{}, false
}

// With returns the document with this judgment added, replacing a judgment about
// the same row so a reason can be corrected instead of duplicated.
func (d Dismissals) With(row Dismissal) Dismissals {
	out := Dismissals{Rows: make([]Dismissal, 0, len(d.Rows)+1)}
	replaced := false
	for _, existing := range d.Rows {
		if strings.EqualFold(strings.TrimSpace(existing.Fingerprint), strings.TrimSpace(row.Fingerprint)) {
			out.Rows = append(out.Rows, row)
			replaced = true
			continue
		}
		out.Rows = append(out.Rows, existing)
	}
	if !replaced {
		out.Rows = append(out.Rows, row)
	}
	return out
}
