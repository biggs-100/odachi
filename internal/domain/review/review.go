// Package review owns the design verdict: what a review says, how it is
// validated, and the one rule that decides it.
//
// Two ideas hold the whole design together.
//
// The verdict is DERIVED, never administered. There is no correction budget and
// no lifecycle: a causal BLOCKER or CRITICAL requests changes, everything else
// approves. A reviewer supplies findings; the tool does the arithmetic.
//
// A verdict is CONTENT-ADDRESSED. It is stored in the clone's store as
// reviews/<tree>.json, so approval cannot float to a tree nobody reviewed: it is
// literally the name of the file. One tree, one verdict, and a second verdict for
// the same tree is reported instead of applied.
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
	"strconv"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/schema"
)

// The store: one file per tree, named by the tree it judged. It lives in the
// clone's state, outside the worktree, so writing a verdict cannot change the
// tree that names it — the immutability is a consequence of where the file lives,
// not of an ignore rule somebody has to remember to write.
const (
	// Dir is where verdicts live, relative to the clone's store.
	Dir = "reviews"
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
// it, and the evidence. When the tool knows them, it also carries the content
// receipt: the blob id of the bytes judged and the lens that produced it.
//
// Blob and Lens are optional because every verdict written before this decision
// parses exactly as it always did: an old finding is a finding without a
// receipt, and the gate asks for a fresh one instead of inventing coverage.
type Finding struct {
	Path     string   `json:"path"`
	Line     int      `json:"line"`
	Severity Severity `json:"severity"`
	// Causal is true only when the candidate introduced, activated or worsened
	// the defect. A pre-existing problem is a follow-up, not a blocker.
	Causal bool   `json:"causal"`
	Why    string `json:"why"`
	// Blob is the content hash of the file this finding judged, derived from git
	// by the tool and never taken from the reviewer's word. It is what lets a
	// later candidate keep this finding potent while the bytes stay identical.
	Blob string `json:"blob,omitempty"`
	// Lens is the specialist that produced this finding. The same closed
	// vocabulary as the verdict's lens list.
	Lens string `json:"lens,omitempty"`
	// Status is open or dismissed. Empty means open, because every verdict
	// written before dismissals existed carries open findings.
	Status string `json:"status,omitempty"`
	// Reason is the evidence that dismissed the row. It exists only on a
	// dismissed finding: a dismissal without a reason is silence wearing a
	// verdict's clothes.
	Reason string `json:"reason,omitempty"`
	// Fingerprint identifies the defect a dismissal judged, so a later
	// candidate recognizes the same row instead of re-raising it. The tool
	// computes it; it is stored so the verdict says what it discarded.
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Verdict is what a review concluded about one tree.
type Verdict struct {
	// Schema is the version of this file, stamped by the tool and never by the
	// reviewer, the way the tree is. It is what lets a binary older than the next
	// field say "this file is newer than me" instead of "unknown field": the
	// strict decoder cannot tell a typo from a writer that is ahead of it. A
	// verdict stored before the field existed carries none and reads as the first
	// schema.
	Schema   int    `json:"schema"`
	Tree     string `json:"tree"`
	Base     string `json:"base"`
	Reviewer string `json:"reviewer"`
	// Model is the model that judged, when the reviewer command named one. The
	// proof carries the builder's; the same name on both sides is one agent
	// approving the tree it wrote.
	Model string `json:"model,omitempty"`
	// Models is the model per lens when the project assigned one: the review runs
	// one reviewer per model group, and the verdict says who judged what. A lens
	// with no entry was left to the reviewer command.
	Models   map[string]string `json:"models,omitempty"`
	At       string            `json:"at"`
	Findings []Finding         `json:"findings"`
	// Lenses are the specialists the prompt asked this verdict to answer. The tool
	// stamps them, because the reviewer's opinion about which lenses it used is not
	// evidence; the prompt it was given is. A verdict stored before lenses existed
	// has none, and the optional gate treats that as "covers nothing".
	Lenses []string `json:"lenses,omitempty"`
	// Passes are the clean answers: (path, blob, lens) rows a lens audited and
	// found empty. They are the other half of coverage — a finding answers its
	// own row, a pass answers the row of a clean blob. A verdict stored before
	// coverage existed has none, and the gate treats that as "covers nothing".
	Passes []Pass `json:"passes,omitempty"`
	// Inspection is the reviewer's account of what it read. A row says a (file,
	// lens) was answered; it cannot say who answered it, and this is the claim
	// the gate compares with the changed paths. A verdict stored before the
	// declaration existed has none, and still parses.
	Inspection *Inspection `json:"inspection,omitempty"`
	Verdict    string      `json:"verdict"`
}

// Blocks reports whether this finding is the reason not to ship: causal, at
// BLOCKER or CRITICAL, and not retired by the human. It is the one definition -
// the floor asks it before judging earned blocking, and the store asks it to
// decide which findings travel beyond the paths a candidate judged.
func (f Finding) Blocks() bool {
	return !IsDismissed(f) && f.Causal && f.Severity.Blocks()
}

// Missing reports which required lenses this verdict does not cover, in the order
// they were required. It is the question both `--show` and the optional gate ask,
// so it is answered in one place.
func (v Verdict) Missing(required []string) []string {
	covered := make(map[string]bool, len(v.Lenses))
	for _, lens := range v.Lenses {
		covered[strings.ToLower(strings.TrimSpace(lens))] = true
	}

	var missing []string
	for _, lens := range required {
		if !covered[strings.ToLower(strings.TrimSpace(lens))] {
			missing = append(missing, lens)
		}
	}
	return missing
}

// Derive is the whole rule: a causal finding at BLOCKER or CRITICAL requests
// changes, everything else approves. Style findings, non-causal findings and
// suggestions are recorded, and never block.
func Derive(findings []Finding) string {
	return DeriveAfter(findings, Refutation{})
}

// Blocking returns the first finding that blocks: causal, and BLOCKER or
// CRITICAL. "First" is the file order, so the message a human reads is the same
// on every run of the same verdict.
func Blocking(findings []Finding) (Finding, bool) {
	return BlockingAfter(findings, Refutation{})
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
		return Verdict{}, errors.New("the verdict is empty: there is nothing to save")
	}
	if text == NoFindings {
		return Verdict{Findings: []Finding{}}, nil
	}

	// The version is asked before the strict pass, and before anything else is
	// refused: a verdict from a newer odachi carries fields this binary does not
	// know, and naming them as unknown sends the human to fix a file that is not
	// broken. A file that is not JSON fails here in silence and is named by the
	// decoder below, in the words it already used.
	var head struct {
		Schema int `json:"schema"`
	}
	if json.Unmarshal([]byte(text), &head) == nil {
		if _, err := schema.Read(head.Schema); err != nil {
			return Verdict{}, fmt.Errorf("the verdict cannot be read: %w", err)
		}
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Verdict
	if err := decoder.Decode(&parsed); err != nil {
		return Verdict{}, fmt.Errorf("the verdict is not valid JSON: %s", oneLine(err.Error()))
	}

	// Anything after the first object — a second verdict, a commentary
	// paragraph — means the answer was not only the JSON it was asked for.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Verdict{}, errors.New("the verdict is not a single JSON object: answer with the verdict only")
	}

	if parsed.Schema == 0 {
		// A verdict written before the field existed: answered as the first
		// schema, so a caller never compares against "no version".
		parsed.Schema = schema.First
	}
	if parsed.Findings == nil {
		parsed.Findings = []Finding{}
	}
	if parsed.Passes == nil {
		parsed.Passes = []Pass{}
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
	if canonical.Passes == nil {
		canonical.Passes = []Pass{}
	}
	// The declaration's path list is a set, and a set with nothing in it is an
	// empty array, not null: a file a person reads says "and no path" instead of
	// leaving them to guess. The copy keeps Render from writing to the caller's
	// declaration.
	if canonical.Inspection != nil && canonical.Inspection.Paths == nil {
		declared := *canonical.Inspection
		declared.Paths = []string{}
		canonical.Inspection = &declared
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
		return errors.New(`the verdict does not say reviewer: add who reviewed it ("reviewer": "<the command or your name>")`)
	}

	if trimmed := strings.TrimSpace(v.At); trimmed != "" {
		if _, err := time.Parse(time.RFC3339, trimmed); err != nil {
			return fmt.Errorf("the verdict says at %q, which is not RFC3339: use for example 2026-10-21T20:31:00Z", trimmed)
		}
	}

	switch strings.TrimSpace(v.Verdict) {
	case "", Approved, ChangesRequested:
	default:
		return fmt.Errorf(`the verdict says verdict %q; the only values are %q and %q (and the verdict is derived from the findings)`,
			v.Verdict, Approved, ChangesRequested)
	}

	for _, lens := range v.Lenses {
		if !IsLens(lens) {
			return fmt.Errorf("the verdict says it covered the lens %q; the lenses are %s",
				lens, strings.Join(order, ", "))
		}
	}

	// The verdict is derived, never administered: an author who declares one that
	// contradicts their own findings has misread the rule, and storing it would
	// hide the finding with the severity vocabulary.
	if declared := strings.TrimSpace(v.Verdict); declared != "" {
		if derived := Derive(v.Findings); declared != derived {
			return fmt.Errorf(
				"the verdict says %q but the rule derives %q: a causal BLOCKER or CRITICAL finding requests changes, and without them there is no block",
				declared, derived)
		}
	}

	if err := validatePasses(v); err != nil {
		return err
	}

	if err := validateInspection(v.Inspection); err != nil {
		return err
	}

	for i, finding := range v.Findings {
		if strings.TrimSpace(finding.Path) == "" {
			return fmt.Errorf("finding %d does not say path: without the file it cannot be fixed", i+1)
		}
		if !finding.Severity.Valid() {
			return fmt.Errorf("finding %d (%s) says severity %q; the vocabulary is BLOCKER, CRITICAL, WARNING, SUGGESTION",
				i+1, finding.Path, string(finding.Severity))
		}
		if finding.Line < 0 {
			return fmt.Errorf("finding %d (%s) says line %d; use 0 for the whole file or the line number",
				i+1, finding.Path, finding.Line)
		}
		if strings.TrimSpace(finding.Why) == "" {
			return fmt.Errorf("finding %d (%s) does not say why: a finding without evidence is an opinion", i+1, finding.Path)
		}
		if blob := strings.TrimSpace(finding.Blob); blob != "" && !hashLike(blob) {
			return fmt.Errorf("finding %d (%s) says blob %q; the blob is a hash like the tree, and a blob that is not hex names nothing",
				i+1, finding.Path, blob)
		}
		if lens := strings.TrimSpace(finding.Lens); lens != "" && !IsLens(lens) {
			return fmt.Errorf("finding %d (%s) says lens %q; the lenses are %s",
				i+1, finding.Path, lens, strings.Join(order, ", "))
		}

		switch status := strings.ToLower(strings.TrimSpace(finding.Status)); status {
		case "", Open:
			if strings.TrimSpace(finding.Reason) != "" {
				return fmt.Errorf("finding %d (%s) says a reason without being dismissed: a reason is what judged a dismissal, not a note",
					i+1, finding.Path)
			}
		case Dismissed:
			if strings.TrimSpace(finding.Reason) == "" {
				return fmt.Errorf("finding %d (%s) is dismissed without a reason: a dismissal without evidence is silence, not a judgment",
					i+1, finding.Path)
			}
			if fingerprint := strings.TrimSpace(finding.Fingerprint); fingerprint != "" && !hashLike(fingerprint) {
				return fmt.Errorf("finding %d (%s) says fingerprint %q; a fingerprint is hex like a hash",
					i+1, finding.Path, fingerprint)
			}
		default:
			return fmt.Errorf("finding %d (%s) says status %q; the states are %s and %s",
				i+1, finding.Path, finding.Status, Open, Dismissed)
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
		return Verdict{}, errors.New("there is no tree: I cannot save a verdict that does not say what it reviewed")
	}
	if err := v.Validate(); err != nil {
		return Verdict{}, err
	}

	if declared := strings.TrimSpace(v.Tree); declared != "" && declared != tree {
		return Verdict{}, fmt.Errorf(
			"the verdict is for another tree (%s) and this tree is %s: a verdict does not move between trees", declared, tree)
	}

	sealed := v
	sealed.Schema = schema.Current
	sealed.Tree = tree
	sealed.Base = strings.TrimSpace(base)
	sealed.Reviewer = strings.TrimSpace(v.Reviewer)
	sealed.Model = strings.TrimSpace(v.Model)
	models, err := NormalizeLensModels(v.Models)
	if err != nil {
		return Verdict{}, err
	}
	sealed.Models = models
	sealed.At = instant(v.At, now)
	sealed.Lenses = inOrder(v.Lenses)
	sealed.Verdict = Derive(v.Findings)
	if sealed.Findings == nil {
		sealed.Findings = []Finding{}
	}
	return sealed, nil
}

// Name is the file a verdict for one tree lives in. The tree is the name.
func Name(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of a verdict: it must be a hash", tree)
	}
	return tree + ".json", nil
}

// Path is the verdict's path relative to the clone's store, with forward slashes
// so the store has one spelling on every platform.
func Path(tree string) (string, error) {
	name, err := Name(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// PromptPath is where the question that tree's verdict answers lives, beside it
// and under the same name: a prompt nobody files is a prompt somebody improvises
// a home for, and the tree it judged is the only fact that keeps it honest.
func PromptPath(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot name a prompt file: it must be a hash", tree)
	}
	return Dir + "/" + tree + ".prompt.md", nil
}

// GroupPromptPath is where the question one lens group was handed lives, named by
// its lenses. A run with a single group files it under the canonical name the
// printed form writes (the caller decides that); a run split across models asks one
// question per group, and a file holding only the last of them would be a lie about
// the others.
func GroupPromptPath(tree string, lenses []string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot name a prompt file: it must be a hash", tree)
	}
	if len(lenses) == 0 {
		return PromptPath(tree)
	}
	names := make([]string, 0, len(lenses))
	for _, lens := range lenses {
		if !IsLens(lens) {
			return "", fmt.Errorf("%q cannot name a prompt file: the lenses are %s", lens, strings.Join(AllLenses(), ", "))
		}
		names = append(names, strings.ToLower(strings.TrimSpace(lens)))
	}
	return Dir + "/" + tree + ".prompt-" + strings.Join(names, "+") + ".md", nil
}

// PartPromptPath is where the question one part of a candidate was handed lives.
// A split review asks the same lenses several questions, and a single file holding
// only the last of them would be a lie about the others.
func PartPromptPath(tree string, lenses []string, part, of int) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot name a prompt file: it must be a hash", tree)
	}
	if part < 1 || of < 2 || part > of {
		return "", fmt.Errorf("part %d of %d cannot name a prompt file: a part is at least 1 and at most the number of parts", part, of)
	}
	names := make([]string, 0, len(lenses))
	for _, lens := range lenses {
		if !IsLens(lens) {
			return "", fmt.Errorf("%q cannot name a prompt file: the lenses are %s", lens, strings.Join(AllLenses(), ", "))
		}
		names = append(names, strings.ToLower(strings.TrimSpace(lens)))
	}
	name := "-" + strings.Join(names, "+")
	if len(names) == 0 {
		name = ""
	}
	return Dir + "/" + tree + ".prompt" + name + "-part-" + strconv.Itoa(part) + "of" + strconv.Itoa(of) + ".md", nil
}

// AnswerDir is where a reviewer told to write {verdict} leaves its answer. It is
// apart from Dir on purpose: a file under reviews/ that is not a verdict is a
// leftover somebody will read as one.
const AnswerDir = "answers"

// AnswerPath is where a reviewer told to write {verdict} leaves its answer. It is
// scaffolding, not evidence: the parsed answer becomes the verdict file, and this
// one is removed once it has been read.
func AnswerPath(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot name an answer file: it must be a hash", tree)
	}
	return AnswerDir + "/" + tree + ".json", nil
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
