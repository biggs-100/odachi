package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// The two things a refuter may say about one finding. There is no third: an
// answer that is not one of these two is not an answer.
const (
	// Stands means the finding survives: it is real, causal and defensible.
	Stands = "stands"
	// Refuted means the finding does not survive, and the refuter could cite the
	// evidence that shows it.
	Refuted = "refuted"
)

// Judgment is one finding's survival, with the evidence that decided it. Why is
// the price of a refutation: no citation, no refutation.
type Judgment struct {
	// ID is the finding's stable name inside the verdict: F1 is the first finding
	// the reviewer filed, F2 the second.
	ID      string `json:"id"`
	Verdict string `json:"verdict"`
	Why     string `json:"why"`
}

// Refutation is the refuter's answer about one verdict. It is stored beside the
// verdict, named after the same tree, and written once.
type Refutation struct {
	Tree        string     `json:"tree"`
	Refuter     string     `json:"refuter,omitempty"`
	At          string     `json:"at"`
	Refutations []Judgment `json:"refutations"`
}

// The two shapes a stored document can have. They share no key and both are
// decoded strictly, so the discriminator is unambiguous.
const (
	// KindVerdict is a review verdict.
	KindVerdict = "verdict"
	// KindRefutation is a refuter's answer.
	KindRefutation = "refutation"
	// KindUnknown is neither.
	KindUnknown = ""
)

// FindingIDs names the findings in the order the reviewer filed them. The ids are
// positions, so they do not move when a human re-reads the same file.
func FindingIDs(findings []Finding) []string {
	ids := make([]string, 0, len(findings))
	for i := range findings {
		ids = append(ids, findingID(i))
	}
	return ids
}

// findingID is the id of the finding at index i.
func findingID(i int) string {
	return "F" + strconv.Itoa(i+1)
}

// ParseRefutation reads a refuter's answer, or a stored refutation: the stored
// document is the answer plus the two facts the tool stamps, so both decode with
// the same function.
//
// Decoding is strict for the same reason a verdict's is: a refutation the tool
// half-understands could lift a block on evidence nobody can point at.
func ParseRefutation(data []byte) (Refutation, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return Refutation{}, errors.New("the refutation is empty: there is nothing to save")
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var parsed Refutation
	if err := decoder.Decode(&parsed); err != nil {
		return Refutation{}, fmt.Errorf("the refutation is not valid JSON: %s", oneLine(err.Error()))
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Refutation{}, errors.New("the refutation is not a single JSON object: answer with the refutation only")
	}

	for i, judgment := range parsed.Refutations {
		if strings.TrimSpace(judgment.ID) == "" {
			return Refutation{}, fmt.Errorf("refutation %d does not say id: each answer names the finding it judges", i+1)
		}
		switch judgment.Verdict {
		case Stands, Refuted:
		default:
			return Refutation{}, fmt.Errorf("the refutation of %s says verdict %q; the only values are %q and %q",
				judgment.ID, judgment.Verdict, Stands, Refuted)
		}
	}

	if parsed.Refutations == nil {
		parsed.Refutations = []Judgment{}
	}
	return parsed, nil
}

// RenderRefutation writes the canonical refutation file: indented JSON, an empty
// array instead of null, and a trailing newline. Two renders of the same
// refutation are the same bytes.
func RenderRefutation(r Refutation) []byte {
	canonical := r
	if canonical.Refutations == nil {
		canonical.Refutations = []Judgment{}
	}

	data, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		// Unreachable: every field is a string or a slice of structs of strings.
		return []byte("{}\n")
	}
	return append(data, '\n')
}

// SealRefutation completes a refuter's answer into the document that is stored:
// it validates every id against the findings that actually block, stamps the
// tree and the instant, and applies the one rule the whole verb exists for.
//
// A blocking finding the refuter did not answer stands. A refutation with no
// evidence is not a refutation. Doubt always votes for keeping the finding.
func SealRefutation(r Refutation, findings []Finding, tree string, now time.Time) (Refutation, error) {
	if strings.TrimSpace(tree) == "" {
		return Refutation{}, errors.New("there is no tree: I cannot save a refutation that does not say what it refuted")
	}

	blocking := blockingIDs(findings)

	answered := make(map[string]Judgment, len(r.Refutations))
	for _, judgment := range r.Refutations {
		id := strings.TrimSpace(judgment.ID)
		if _, ok := blocking[id]; !ok {
			return Refutation{}, fmt.Errorf(
				"the refutation talks about %s, which is not a causal BLOCKER or CRITICAL of this verdict: the refuter can only judge the ones that block", id)
		}
		if _, twice := answered[id]; twice {
			return Refutation{}, fmt.Errorf("the refutation repeats %s: one answer per finding", id)
		}
		switch judgment.Verdict {
		case Stands, Refuted:
		default:
			return Refutation{}, fmt.Errorf("the refutation of %s says verdict %q; the only values are %q and %q",
				id, judgment.Verdict, Stands, Refuted)
		}
		answered[id] = normalized(judgment, id)
	}

	sealed := Refutation{
		Tree:        tree,
		Refuter:     strings.TrimSpace(r.Refuter),
		At:          instant(r.At, now),
		Refutations: make([]Judgment, 0, len(blocking)),
	}
	// The stored record is in the verdict's file order and covers every blocking
	// finding, so a reader a month later sees the whole question and not just the
	// half the refuter answered.
	for i := range findings {
		id := findingID(i)
		if _, ok := blocking[id]; !ok {
			continue
		}
		judgment, ok := answered[id]
		if !ok {
			judgment = Judgment{ID: id, Verdict: Stands, Why: "no answer from the refuter: the finding stands"}
		}
		sealed.Refutations = append(sealed.Refutations, judgment)
	}
	return sealed, nil
}

// normalized applies the evidence rule to one judgment: a refutation without a
// citation is not a refutation.
func normalized(judgment Judgment, id string) Judgment {
	why := strings.TrimSpace(judgment.Why)
	if judgment.Verdict == Refuted && why == "" {
		return Judgment{ID: id, Verdict: Stands, Why: ""}
	}
	return Judgment{ID: id, Verdict: judgment.Verdict, Why: why}
}

// blockingIDs is the set of ids the refuter may judge: causal BLOCKER and
// CRITICAL findings, and nothing else.
func blockingIDs(findings []Finding) map[string]struct{} {
	ids := make(map[string]struct{}, len(findings))
	for i, finding := range findings {
		if finding.Causal && finding.Severity.Blocks() {
			ids[findingID(i)] = struct{}{}
		}
	}
	return ids
}

// Refutes reports whether this refutation lifts the finding with that id. It is
// false for anything else — an unanswered finding, a `stands`, or a "refuted"
// with no evidence — because the default is always that the finding stands.
func (r Refutation) Refutes(id string) bool {
	for _, judgment := range r.Refutations {
		if judgment.ID != id {
			continue
		}
		return judgment.Verdict == Refuted && strings.TrimSpace(judgment.Why) != ""
	}
	return false
}

// BlockingAfter returns the first finding that still blocks, in the verdict's
// file order: causal, BLOCKER or CRITICAL, and not refuted with evidence.
func BlockingAfter(findings []Finding, r Refutation) (Finding, bool) {
	for i, finding := range findings {
		if !finding.Causal || !finding.Severity.Blocks() {
			continue
		}
		if r.Refutes(findingID(i)) {
			continue
		}
		return finding, true
	}
	return Finding{}, false
}

// DeriveAfter is the whole rule with refutations taken into account: a causal
// finding at BLOCKER or CRITICAL that nobody refuted requests changes,
// everything else approves.
func DeriveAfter(findings []Finding, r Refutation) string {
	if _, blocking := BlockingAfter(findings, r); blocking {
		return ChangesRequested
	}
	return Approved
}

// RefutationName is the file a refutation for one tree lives in.
func RefutationName(tree string) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of a refutation: it must be a hash", tree)
	}
	return tree + ".refuted.json", nil
}

// RefutationPath is the refutation's path relative to the project root.
func RefutationPath(tree string) (string, error) {
	name, err := RefutationName(tree)
	if err != nil {
		return "", err
	}
	return Dir + "/" + name, nil
}

// Sniff says which stored document these bytes are. A verdict and a refutation
// share no key, and both are decoded strictly, so the top-level "refutations" key
// decides and nothing is guessed.
func Sniff(data []byte) string {
	var probe struct {
		Refutations json.RawMessage `json:"refutations"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return KindUnknown
	}
	if probe.Refutations != nil {
		return KindRefutation
	}
	// A verdict is only a verdict if it would be accepted as one: otherwise the
	// caller would have to guess what a half-written document was meant to be.
	if parsed, err := Parse(data); err == nil && parsed.Validate() == nil {
		return KindVerdict
	}
	return KindUnknown
}

// RefuteInput is what the refuter is shown: which verdict, which findings block,
// and the diff it has to cite.
type RefuteInput struct {
	Tree     string
	Base     string
	Reviewer string
	Findings []Finding
	Diff     string
}

// RefutePrompt renders the refuter's prompt. Content is fixed and pinned by
// tests: this is the only place the "default stands" rule is stated to a model,
// and a prompt that drifts would refute by default.
func RefutePrompt(in RefuteInput) string {
	var out strings.Builder

	out.WriteString("# odachi refuter · tree " + in.Tree + " · base " + in.Base + "\n\n")
	out.WriteString("You are the refuter of this verdict. Your job is not to approve or reject the change:\n")
	out.WriteString("it is to tear down the findings that do not hold, and nothing else.\n\n")

	out.WriteString("## The verdict you are judging\n")
	out.WriteString("- tree: " + in.Tree + "\n")
	out.WriteString("- base: " + in.Base + "\n")
	if strings.TrimSpace(in.Reviewer) != "" {
		out.WriteString("- reviewer: " + in.Reviewer + "\n")
	}
	out.WriteString("\n")

	out.WriteString("## The only thing you can refute\n")
	blocking := blockingFindings(in.Findings)
	if len(blocking) == 0 {
		out.WriteString("(none: this verdict has no causal BLOCKER or CRITICAL, there is nothing to refute)\n")
	}
	for _, entry := range blocking {
		out.WriteString("### " + entry.id + " · " + where(entry.finding) + " · " + string(entry.finding.Severity) + "\n")
		out.WriteString(entry.finding.Why + "\n\n")
	}

	out.WriteString("## The diff (the only evidence that counts)\n")
	out.WriteString(section(in.Diff))
	out.WriteString("\n")

	out.WriteString("## The three questions, one answer per finding\n")
	out.WriteString("1. Is it causal of the diff, or was it already there? Cite the diff line that shows it.\n")
	out.WriteString("2. Is it a real defect that affects the user, or a preference?\n")
	out.WriteString("3. Is the evidence enough to defend it before the author?\n\n")

	out.WriteString("## How it is decided\n")
	out.WriteString("- Without a citation from the diff, the finding is not refuted: it stays in `" + Stands + "`. Doubt does not refute.\n")
	out.WriteString("- `" + Refuted + "` only when the three answers hold with the citation.\n")
	out.WriteString("- When in doubt: `" + Stands + "`. Over-refuting is worse than under-refuting.\n")
	out.WriteString("- WARNING and SUGGESTION are not judged: they do not block, and do not add new findings.\n")
	out.WriteString("- The five review rules still hold: a real defect, defensible with evidence,\n")
	out.WriteString("  and causal of this candidate. What was already there is a follow-up.\n\n")

	out.WriteString("## What to return\n")
	out.WriteString(`{"refutations":[{"id":"` + firstID(blocking) + `","verdict":"` + Stands + `|` + Refuted + `","why":"<the three answers and the citation from the diff>"}]}` + "\n\n")
	out.WriteString("One object per finding in the list above, with its id, in the same order.\n")
	out.WriteString("Answer with the JSON only, without text around it.\n")

	return out.String()
}

// blockingEntry is one refutable finding with the id the answer must use.
type blockingEntry struct {
	id      string
	finding Finding
}

// blockingFindings is the refutable subset, in file order, with ids.
func blockingFindings(findings []Finding) []blockingEntry {
	entries := make([]blockingEntry, 0, len(findings))
	for i, finding := range findings {
		if !finding.Causal || !finding.Severity.Blocks() {
			continue
		}
		entries = append(entries, blockingEntry{id: findingID(i), finding: finding})
	}
	return entries
}

// firstID is the id the answer template shows, so the shape has a real example.
func firstID(entries []blockingEntry) string {
	if len(entries) == 0 {
		return "F1"
	}
	return entries[0].id
}

// where is one finding's location, in the shape the review prompt used.
func where(finding Finding) string {
	if finding.Line > 0 {
		return finding.Path + ":" + strconv.Itoa(finding.Line)
	}
	return finding.Path
}
