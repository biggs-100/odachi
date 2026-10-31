package review

import (
	"fmt"
	"slices"
	"strings"
)

// This file owns the coverage arithmetic of ADR 0016: what a verdict must
// carry so the gate can ask "was every (file, lens) pair answered about the
// exact bytes that stand now".
//
// Two kinds of row answer it. A finding covers its own (path, blob, lens)
// row by being the reported defect; a Pass covers the row of a blob a lens
// audited and found clean. Both carry their content receipt, because coverage
// without one would be an answer about bytes nobody named.

// Pass is one lens's clean answer for one blob: the lens looked and reported
// nothing. The blob is the receipt; the lens is who answered; the path is the
// file the receipt was measured on.
type Pass struct {
	Path string `json:"path"`
	Blob string `json:"blob"`
	Lens string `json:"lens"`
}

// Gap is one unanswered row: which file must answer, and which lens of its path
// class does not hold a current answer.
type Gap struct {
	Path string
	Lens string
}

// Uncovered asks the coverage question for a candidate whose files need the
// answers in `need` (path → the lenses its path class requires) and whose
// current content is `blobs` (path → the blob id git measured now).
//
// A row is covered when a finding or a pass names the same path and lens and a
// blob equal to the current one. Any other case is a gap:
//
//   - a verdict written before content receipts has no usable rows, so it
//     covers nothing: the gate asks for a fresh review instead of an implicit
//     yes;
//   - a stale blob means the receipt describes other bytes, which is the
//     potency boundary — this file changed, so its answer is dead;
//   - a missing lens means the row was never answered.
//
// The rows are returned in a deterministic order (path, then the caller's lens
// order), so two calls with the same candidate produce the same message.
func (v Verdict) Uncovered(need map[string][]string, blobs map[string]string) []Gap {
	var gaps []Gap
	for _, path := range sortedPaths(need) {
		for _, lens := range need[path] {
			if !rowCovered(v, path, lens, blobs[path]) {
				gaps = append(gaps, Gap{Path: path, Lens: lens})
			}
		}
	}
	return gaps
}

// rowCovered reports whether one (path, lens) row holds a current answer:
// a finding or a pass that names this path and lens and whose blob is exactly
// the blob of these bytes. Everything else is a gap.
func rowCovered(v Verdict, path, lens, blob string) bool {
	for _, finding := range v.Findings {
		if sameRow(finding.Path, finding.Lens, finding.Blob, path, lens, blob) {
			return true
		}
	}
	for _, pass := range v.Passes {
		if sameRow(pass.Path, pass.Lens, pass.Blob, path, lens, blob) {
			return true
		}
	}
	return false
}

// sameRow compares one stored answer with one asked row: the same file, the
// same specialist, and the same bytes. A receipt for other bytes is the potency
// boundary — the answer is dead — and an absent receipt answers nothing.
func sameRow(answerPath, answerLens, answerBlob, path, lens, blob string) bool {
	if RowPath(answerPath) != RowPath(path) {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(answerLens), lens) {
		return false
	}
	return strings.TrimSpace(answerBlob) != "" && strings.TrimSpace(answerBlob) == blob
}

// RowPath is one path in the spelling coverage and the store compare: forward
// slashes and lower case, so Windows and Linux name the same file the same way,
// and two places that ask "is this the same path" cannot answer differently.
func RowPath(path string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(path), "\\", "/"))
}

// sortedPaths is the paths of a need map in alphabetical order, so two calls
// with the same candidate produce the same gap order and the gate's message is
// stable.
func sortedPaths(need map[string][]string) []string {
	paths := make([]string, 0, len(need))
	for path := range need {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	return paths
}

// validatePasses is the part of Validate that belongs to coverage rows. It is a
// separate function so review.go keeps its finding checks and this file keeps
// its arithmetic; Validate calls both in order.
//
// A pass without a blob is the reviewer's answer waiting for the tool's stamp —
// exactly what a fresh answer carries. It is inert at use time (no receipt, no
// coverage), so it is refused nowhere; a blob that is not a hash names bytes
// nobody measured and is refused, like any invented fact.
func validatePasses(v Verdict) error {
	for i, pass := range v.Passes {
		row := i + 1
		if strings.TrimSpace(pass.Path) == "" {
			return fmt.Errorf("pass %d does not say path: without the file it cannot be a coverage row", row)
		}
		if blob := strings.TrimSpace(pass.Blob); blob != "" && !hashLike(blob) {
			return fmt.Errorf("pass %d (%s) says blob %q; the blob is a hash like the tree, and a blob that is not hex names nothing",
				row, pass.Path, pass.Blob)
		}
		if strings.TrimSpace(pass.Lens) == "" {
			return fmt.Errorf("pass %d (%s) does not say lens: without the specialist it answers no question", row, pass.Path)
		}
		if !IsLens(pass.Lens) {
			return fmt.Errorf("pass %d (%s) says lens %q; the lenses are %s",
				row, pass.Path, pass.Lens, strings.Join(order, ", "))
		}
	}
	return nil
}
