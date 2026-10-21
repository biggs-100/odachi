// Package review owns the design verdict: what a review says, how it is
// validated, and the one rule that decides it.
//
// Two ideas hold the whole design together.
//
// The verdict is DERIVED, never administered. There is no correction budget and
// no lifecycle: a causal BLOCKER or CRITICAL requests changes, everything else
// approves. A reviewer supplies findings; the tool does the arithmetic.
//
// A verdict is CONTENT-ADDRESSED. It is stored as .odachi/reviews/<tree>.json,
// so approval cannot float to a tree nobody reviewed: it is literally the name
// of the file. One tree, one verdict, and a second verdict for the same tree is
// reported instead of applied.
//
// The package is pure: strings, bytes and time in, strings, bytes and time out.
// It reads no files, runs no commands and knows nothing about git; the tree hash
// is a fact the caller brings.
package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// The store: one file per tree, and the ignore file that keeps the store out of
// the tree it describes. Without it, storing a verdict would change the tree
// that names it, and every verdict would invalidate itself.
const (
	// Dir is where verdicts live, relative to the project root.
	Dir = ".odachi/reviews"
	// IgnoreName is the file inside Dir that keeps the verdicts out of git.
	IgnoreName = ".gitignore"
	// IgnoreContent ignores everything in Dir except the ignore file itself,
	// which is versioned: the rule that keeps verdicts out of the tree travels
	// with the project instead of living inside the tool.
	IgnoreContent = "# odachi: un veredicto por árbol y se generan; no se versionan\n*\n!.gitignore\n"
)

// NoFindings is the canonical empty answer. A reviewer with nothing to report
// says exactly this and nothing more, which is what stops it inventing findings
// to look useful.
const NoFindings = "No findings."

// Severity is the closed vocabulary: four values, no more.
type Severity string

const (
	Blocker    Severity = "BLOCKER"
	Critical   Severity = "CRITICAL"
	Warning    Severity = "WARNING"
	Suggestion Severity = "SUGGESTION"
)

// Valid reports whether the severity is one of the four.
func (s Severity) Valid() bool {
	switch s {
	case Blocker, Critical, Warning, Suggestion:
		return true
	default:
		return false
	}
}

// Blocks reports whether this severity can block a candidate when it is causal.
// A warning is information; a blocker is a reason not to ship.
func (s Severity) Blocks() bool {
	return s == Blocker || s == Critical
}

// The two verdicts. There are no others: the rule derives one of these and a
// stored file never carries a third value.
const (
	Approved         = "approved"
	ChangesRequested = "changes_requested"
)

// Finding is one reported defect: where, how bad, whether this candidate caused
// it, and the evidence.
type Finding struct {
	Path     string   `json:"path"`
	Line     int      `json:"line"`
	Severity Severity `json:"severity"`
	// Causal is true only when the candidate introduced, activated or worsened
	// the defect. A pre-existing problem is a follow-up, not a blocker.
	Causal bool   `json:"causal"`
	Why    string `json:"why"`
}

// Verdict is what a review concluded about one tree.
type Verdict struct {
	Tree     string    `json:"tree"`
	Base     string    `json:"base"`
	Reviewer string    `json:"reviewer"`
	At       string    `json:"at"`
	Findings []Finding `json:"findings"`
	Verdict  string    `json:"verdict"`
}

// Derive is the whole rule: a causal finding at BLOCKER or CRITICAL requests
// changes, everything else approves. Style findings, non-causal findings and
// suggestions are recorded, and never block.
func Derive(findings []Finding) string {
	if _, blocking := Blocking(findings); blocking {
		return ChangesRequested
	}
	return Approved
}

// Blocking returns the first finding that blocks: causal, and BLOCKER or
// CRITICAL. "First" is the file order, so the message a human reads is the same
// on every run of the same verdict.
func Blocking(findings []Finding) (Finding, bool) {
	for _, finding := range findings {
		if finding.Causal && finding.Severity.Blocks() {
			return finding, true
		}
	}
	return Finding{}, false
}

// Parse reads a verdict. It accepts exactly two shapes: the canonical empty
// answer ("No findings.") and the JSON the prompt asks for.
//
// Decoding is strict — an unknown field, a severity outside the vocabulary or a
// second JSON object is refused — because a verdict the tool half-understands is
// worse than no verdict: it would be stored under a name that claims more than
// it can back.
func Parse(data []byte) (Verdict, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return Verdict{}, errors.New("el veredicto está vacío: no hay nada que guardar")
	}
	if text == NoFindings {
		return Verdict{Findings: []Finding{}}, nil
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Verdict
	if err := decoder.Decode(&parsed); err != nil {
		return Verdict{}, fmt.Errorf("el veredicto no es JSON válido: %s", oneLine(err.Error()))
	}

	// Anything after the first object — a second verdict, a commentary
	// paragraph — means the answer was not only the JSON it was asked for.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Verdict{}, errors.New("el veredicto no es un solo objeto JSON: contestá solo con el veredicto")
	}

	if parsed.Findings == nil {
		parsed.Findings = []Finding{}
	}
	return parsed, nil
}

// Render writes the canonical file: indented JSON, fields in the documented
// order, an empty findings array instead of null, and a trailing newline. Two
// renders of the same verdict are the same bytes, which is what makes the file
// comparable and its immutability checkable.
func Render(v Verdict) []byte {
	canonical := v
	if canonical.Findings == nil {
		canonical.Findings = []Finding{}
	}

	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		// Unreachable: every field of Verdict is a string, an int, a bool, or a
		// slice of those. Keeping the fallback explicit beats a panic.
		return []byte("{}\n")
	}
	return append(data, '\n')
}

// Validate reports what only the author of a verdict can fix: who reviewed it,
// and whether every finding is complete and uses the closed vocabulary.
//
// It is the last check before a write, and the reason an incomplete attached
// verdict leaves no file behind.
func (v Verdict) Validate() error {
	if strings.TrimSpace(v.Reviewer) == "" {
		return errors.New(`el veredicto no dice reviewer: agregá quién revisó ("reviewer": "<el comando o tu nombre>")`)
	}

	if trimmed := strings.TrimSpace(v.At); trimmed != "" {
		if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
			return fmt.Errorf("el veredicto dice at %q, que no es RFC3339: usá por ejemplo 2026-10-21T20:31:00Z", trimmed)
		}
	}

	switch strings.TrimSpace(v.Verdict) {
	case "", Approved, ChangesRequested:
	default:
		return fmt.Errorf(`el veredicto dice verdict %q; los únicos valores son %q y %q (y el veredicto se deriva de los findings)`,
			v.Verdict, Approved, ChangesRequested)
	}

	// The verdict is derived, never administered: an author who declares one that
	// contradicts their own findings has misread the rule, and storing it would
	// hide the finding with the severity vocabulary.
	if declared := strings.TrimSpace(v.Verdict); declared != "" {
		if derived := Derive(v.Findings); declared != derived {
			return fmt.Errorf(
				"el veredicto dice %q pero la regla deriva %q: un finding causal BLOCKER o CRITICAL pide cambios, y sin ellos no hay bloqueo",
				declared, derived)
		}
	}

	for i, finding := range v.Findings {
		if strings.TrimSpace(finding.Path) == "" {
			return fmt.Errorf("el finding %d no dice path: sin el archivo no se puede arreglar", i+1)
		}
		if !finding.Severity.Valid() {
			return fmt.Errorf("el finding %d (%s) dice severity %q; el vocabulario es BLOCKER, CRITICAL, WARNING, SUGGESTION",
				i+1, finding.Path, string(finding.Severity))
		}
		if finding.Line < 0 {
			return fmt.Errorf("el finding %d (%s) dice line %d; usá 0 para todo el archivo o el número de línea",
				i+1, finding.Path, finding.Line)
		}
		if strings.TrimSpace(finding.Why) == "" {
			return fmt.Errorf("el finding %d (%s) no dice why: un finding sin evidencia es una opinión", i+1, finding.Path)
		}
	}

	return nil
}

// Seal validates a verdict, stamps the parts only the tool knows — the tree, the
// base, the instant — and derives the verdict from the findings.
//
// A verdict that names another tree is refused: a review of one tree is not a
// review of another, and the store's file name says so. Everything else that can
// be wrong is already refused by Validate, before anything is written.
func Seal(v Verdict, tree, base string, now time.Time) (Verdict, error) {
	if strings.TrimSpace(tree) == "" {
		return Verdict{}, errors.New("no hay árbol: no puedo guardar un veredicto que no dice qué revisó")
	}
	if err := v.Validate(); err != nil {
		return Verdict{}, err
	}

	if declared := strings.TrimSpace(v.Tree); declared != "" && declared != tree {
		return Verdict{}, fmt.Errorf(
			"el veredicto es de otro árbol (%s) y este árbol es %s: un veredicto no se mueve de árbol", declared, tree)
	}

	sealed := v
	sealed.Tree = tree
	sealed.Base = strings.TrimSpace(base)
	sealed.Reviewer = strings.TrimSpace(v.Reviewer)
	sealed.At = instant(v.At, now)
	sealed.Verdict = Derive(v.Findings)
	if sealed.Findings == nil {
		sealed.Findings = []Finding{}
	}
	return sealed, nil
}

// Name is the file a verdict for one tree lives in. The tree is the name.
func Name(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("el árbol %q no puede ser el nombre de un veredicto: tiene que ser un hash", tree)
	}
	return tree + ".json", nil
}

// Path is the verdict's path relative to the project root, with forward slashes
// so the store has one spelling on every platform.
func Path(tree string) (string, error) {
	name, err := Name(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// hashLike reports whether s can be a tree hash. Only hex is accepted, which is
// what makes it impossible for a verdict name to describe a path: no separator,
// no dots, nothing a filesystem would read as more than a file name.
func hashLike(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			continue
		default:
			return false
		}
	}
	return true
}

// instant is the instant a verdict records. An author may write one; it is
// normalized to UTC so two stored verdicts can be compared. Without one, the
// tool's clock is the only honest answer.
func instant(at string, now time.Time) string {
	if trimmed := strings.TrimSpace(at); trimmed != "" {
		if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
	}
	if now.IsZero() {
		now = time.Now()
	}
	return now.UTC().Format(time.RFC3339)
}

// oneLine keeps a message on one line whatever a decoder said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
