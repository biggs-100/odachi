package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/domain/schema"
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

// CitationStatus is the one decision every refutation consumer shares. Valid is
// the only status that lends a vote; invalid and ambiguous both keep the finding
// standing.
type CitationStatus string

const (
	CitationValid     CitationStatus = "valid"
	CitationInvalid   CitationStatus = "invalid"
	CitationAmbiguous CitationStatus = "ambiguous"
)

// CitationEvidence is the candidate evidence available to check a locator. Diff
// is authoritative when present; Source is the cited file's contents when no
// diff is available. Tree binds both to the candidate the refutation names.
type CitationEvidence struct {
	Tree   string
	Diff   string
	Source string
}

// CitationInput is one refutation judgment bound to the finding it answers.
type CitationInput struct {
	FindingID    string
	Finding      Finding
	Locator      string
	DeclaredTree string
	Evidence     CitationEvidence
}

// CitationResult explains the shared decision without turning ambiguity into a
// vote. Reason is empty only for CitationValid.
type CitationResult struct {
	Status CitationStatus
	Reason string
}

var (
	lineCitation  = regexp.MustCompile(`(?:^|[\s"'` + "`" + `(\[{])([^\r\n"'` + "`" + `()\[\]{}]+?):([1-9][0-9]*)\b`)
	quoteCitation = regexp.MustCompile(`(?:^|[\s"'` + "`" + `(\[{])([^\r\n"'` + "`" + `()\[\]{}]+?)(?:\s*:\s*|\s+)(?:"([^"\n]+)"|'([^'\n]+)'|` + "`([^`\n]+)`" + `)`)
)

// SealedFrom is the schema from which a refutation's `at` is the instant this
// clone sealed the file and the refuter's own hour, when it disagrees, is kept in
// DeclaredAt. A file written under it carries a time nothing can attribute: it is
// either the refuter's word or the clock of the clone that stored it, and no field
// says which (0291).
const SealedFrom = 8

// Refutation is the refuter's answer about one verdict. It is stored beside the
// verdict, named after the same tree, and written once.
type Refutation struct {
	// Schema is the version of this file's shape, stamped by the tool. The
	// refutation was the fourth strict decoder and the only one with no number to
	// announce itself: without it, the next field it gains reads as an unknown key
	// in every binary older than the change (0291).
	Schema int    `json:"schema"`
	Tree   string `json:"tree"`
	// Refuter names the account that answered. The quorum counts accounts, so two
	// answers from one name are one opinion written twice.
	Refuter string `json:"refuter,omitempty"`
	// At is the instant the tool sealed this file: the clock of the clone that
	// wrote it, never the refuter's word. A refutation stored before SealedFrom
	// carries the refuter's own time here, and nothing can tell the two apart; the
	// field is honest from the schema that introduced DeclaredAt onwards.
	At string `json:"at"`
	// DeclaredAt is the time the refuter wrote, kept only when it parses and
	// disagrees with At. A refuter that claims a round hour is evidence about the
	// refuter, and dropping it would erase the only trace of a wrong clock.
	DeclaredAt  string     `json:"declared_at,omitempty"`
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
	// Compatibility is the first decision: a record from a newer odachi must
	// receive the upgrade instruction before its citations or unknown fields can
	// be mistaken for damage. Non-JSON input still falls through to the decoder's
	// existing refusal below.
	var head struct {
		Schema int `json:"schema"`
	}
	if json.Unmarshal(data, &head) == nil {
		if _, err := schema.Read(head.Schema); err != nil {
			return Refutation{}, fmt.Errorf("the refutation cannot be read: %w", err)
		}
	}

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
		if judgment.Verdict == Refuted {
			result := ValidateCitation(CitationInput{FindingID: judgment.ID, Locator: judgment.Why})
			if result.Status == CitationInvalid {
				return Refutation{}, fmt.Errorf("the refutation of %s: %s", judgment.ID, result.Reason)
			}
		}
	}

	if parsed.Schema == 0 {
		// A refutation written before the field existed: answered as the first
		// schema, so a caller never compares against "no version".
		parsed.Schema = schema.First
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
func SealRefutation(r Refutation, findings []Finding, evidence CitationEvidence, now time.Time) (Refutation, error) {
	tree := strings.TrimSpace(evidence.Tree)
	if tree == "" {
		return Refutation{}, errors.New("there is no tree: I cannot save a refutation that does not say what it refuted")
	}
	if declared := strings.TrimSpace(r.Tree); declared != "" && declared != tree {
		return Refutation{}, fmt.Errorf("the refutation is from tree %s and this tree is %s: matching finding ids do not bind evidence to another candidate", declared, tree)
	}
	if strings.TrimSpace(r.Tree) == "" {
		r.Tree = tree
	}
	for i := range r.Refutations {
		r.Refutations[i] = normalized(r.Refutations[i], strings.TrimSpace(r.Refutations[i].ID))
	}
	if err := r.validateCitationResults(findings, evidence); err != nil {
		return Refutation{}, err
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
		Schema:      schema.Current,
		Tree:        tree,
		Refuter:     strings.TrimSpace(r.Refuter),
		At:          stamp(now),
		DeclaredAt:  declared(r.At, now),
		Refutations: make([]Judgment, 0, len(blocking)),
	}
	// A draft that already carries a claim is one this tool sealed before, and then
	// its At is the stamp this tool wrote -not the refuter's word-. Reading that as
	// a claim would print "the refuter wrote" a time the clone wrote and would drop
	// the claim the first seal read, so the preserved field decides, not the empty
	// one: the answer is not allowed to depend on how far apart the two seals fell
	// (0291).
	if claim := strings.TrimSpace(r.DeclaredAt); claim != "" {
		sealed.DeclaredAt = claim
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

// ValidateCitation is the single decision used by parsing, sealing, inheritance
// and the check gate. Every locator in the prose is considered: the first
// evidence-valid candidate wins, ambiguity survives when no candidate can be
// decided, and invalid is returned only when every candidate contradicts the
// available evidence.
func ValidateCitation(in CitationInput) CitationResult {
	candidates := parseCitationsForFinding(in.Locator, in.Finding.Path)
	if len(candidates) == 0 {
		if located, ok := parseCitation(in.Locator); ok {
			candidates = append(candidates, located)
		}
	}
	if len(candidates) == 0 {
		if path := strings.TrimSpace(in.Finding.Path); path != "" {
			return CitationResult{CitationInvalid, fmt.Sprintf("needs a citation: use %s:%d or %s plus a quoted fragment from the candidate", path, max(in.Finding.Line, 1), path)}
		}
		return CitationResult{CitationInvalid, "needs a citation: use path:line or a path plus a quoted candidate fragment"}
	}

	invalid := CitationResult{Status: CitationInvalid}
	ambiguous := CitationResult{}
	for _, located := range candidates {
		result := validateLocatedCitation(in, located)
		switch result.Status {
		case CitationValid:
			return result
		case CitationAmbiguous:
			if ambiguous.Status == "" {
				ambiguous = result
			}
		default:
			if invalid.Reason == "" {
				invalid = result
			}
		}
	}
	if ambiguous.Status != "" {
		return ambiguous
	}
	return invalid
}

func validateLocatedCitation(in CitationInput, located citation) CitationResult {
	if path := strings.TrimSpace(in.Finding.Path); path != "" && RowPath(located.path) != RowPath(path) {
		return CitationResult{CitationInvalid, fmt.Sprintf("cites %s, but the finding is about %s", located.path, path)}
	}
	if strings.TrimSpace(in.FindingID) == "" || strings.TrimSpace(in.Finding.Path) == "" {
		return CitationResult{CitationAmbiguous, "ambiguous: the finding identity or path is missing"}
	}
	declared := strings.TrimSpace(in.DeclaredTree)
	available := strings.TrimSpace(in.Evidence.Tree)
	if declared == "" || available == "" {
		return CitationResult{CitationAmbiguous, "ambiguous: the declared or available evidence tree is missing"}
	}
	if declared != available {
		return CitationResult{CitationInvalid, fmt.Sprintf("the refutation names tree %s, but the available evidence is tree %s", declared, available)}
	}

	diff := in.Evidence.Diff
	source := in.Evidence.Source
	if strings.TrimSpace(diff) == "" && strings.TrimSpace(source) == "" {
		return CitationResult{CitationAmbiguous, "ambiguous: the candidate diff and source evidence are missing or empty"}
	}

	if located.line > 0 {
		if strings.TrimSpace(diff) != "" {
			file, known, checkable := citationDiffFile(diff, located.path)
			if !checkable {
				return CitationResult{CitationAmbiguous, "ambiguous: the candidate diff cannot be read"}
			}
			if !known {
				return CitationResult{CitationInvalid, fmt.Sprintf("cites %s, which is not in the available candidate diff", located.path)}
			}
			if !diffHasLine(file.text, located.line) {
				return CitationResult{CitationInvalid, fmt.Sprintf("cites %s:%d, which is not a checkable line in the candidate", located.path, located.line)}
			}
		} else if !sourceHasLine(source, located.line) {
			return CitationResult{CitationInvalid, fmt.Sprintf("cites %s:%d, which is not a line in the available source", located.path, located.line)}
		}
		return CitationResult{Status: CitationValid}
	}

	quote := strings.TrimSpace(located.quote)
	if strings.TrimSpace(diff) != "" {
		file, known, checkable := citationDiffFile(diff, located.path)
		if !checkable {
			return CitationResult{CitationAmbiguous, "ambiguous: the candidate diff cannot be read"}
		}
		if !known {
			return CitationResult{CitationInvalid, fmt.Sprintf("cites %s, which is not in the available candidate diff", located.path)}
		}
		if !strings.Contains(candidateDiffText(file.text), quote) {
			return CitationResult{CitationInvalid, fmt.Sprintf("quotes %q, which is not a fragment in the candidate's %s", quote, located.path)}
		}
	} else if !strings.Contains(source, quote) {
		return CitationResult{CitationInvalid, fmt.Sprintf("quotes %q, which is not a fragment in the available source for %s", quote, located.path)}
	}
	return CitationResult{Status: CitationValid}
}

// CitationResults applies ValidateCitation to every refuted judgment. Stands and
// unanswered rows lend no votes and need no locator.
func (r Refutation) CitationResults(findings []Finding, evidence CitationEvidence) []CitationResult {
	blocking := make(map[string]Finding, len(findings))
	for _, entry := range blockingFindings(findings) {
		blocking[entry.id] = entry.finding
	}
	var results []CitationResult
	for _, judgment := range r.Refutations {
		if judgment.Verdict != Refuted {
			continue
		}
		id := strings.TrimSpace(judgment.ID)
		finding, ok := blocking[id]
		if !ok {
			results = append(results, CitationResult{CitationInvalid, fmt.Sprintf("the refutation talks about %s, which is not a causal BLOCKER or CRITICAL of this verdict", id)})
			continue
		}
		result := ValidateCitation(CitationInput{
			FindingID:    id,
			Finding:      finding,
			Locator:      judgment.Why,
			DeclaredTree: r.Tree,
			Evidence:     evidence,
		})
		if result.Reason != "" {
			result.Reason = "the refutation of " + id + ": " + result.Reason
		}
		results = append(results, result)
	}
	return results
}

func (r Refutation) validateCitationResults(findings []Finding, evidence CitationEvidence) error {
	for _, result := range r.CitationResults(findings, evidence) {
		if result.Status != CitationValid {
			return errors.New(result.Reason)
		}
	}
	return nil
}

// citation is the concrete locator carried inside why. The persisted shape stays
// unchanged: a path and line, or a path and a quoted candidate fragment.
type citation struct {
	path  string
	line  int
	quote string
}

// parseCitation reads path:line or path "quoted fragment" from why without
// changing the persisted shape.
func parseCitation(why string) (citation, bool) {
	if match := lineCitation.FindStringSubmatch(why); len(match) == 3 {
		if line, err := strconv.Atoi(match[2]); err == nil && line > 0 {
			return citation{path: strings.TrimSpace(match[1]), line: line}, true
		}
	}
	match := quoteCitation.FindStringSubmatch(why)
	if len(match) != 5 {
		return citation{}, false
	}
	for _, quote := range match[2:] {
		if strings.TrimSpace(quote) != "" {
			return citation{path: strings.TrimSpace(match[1]), quote: quote}, true
		}
	}
	return citation{}, false
}

// parseCitationsForFinding captures every locator anchored to the exact finding
// path, including paths with spaces. Evaluation decides which mention is useful;
// parsing never lets an earlier quoted mention hide a later valid line.
func parseCitationsForFinding(why, path string) []citation {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	var candidates []citation
	for at := 0; at < len(why); {
		rel := strings.Index(why[at:], path)
		if rel < 0 {
			break
		}
		start := at + rel
		at = start + len(path)
		if start > 0 && !citationBoundary(why[start-1]) {
			continue
		}
		if located, ok := parseCitationTail(path, why[at:]); ok {
			candidates = append(candidates, located)
		}
	}
	return candidates
}

func parseCitationTail(path, tail string) (citation, bool) {
	tail = strings.TrimLeftFunc(tail, func(r rune) bool { return r == ' ' || r == '\t' || r == '\r' || r == '\n' })
	if rest, ok := strings.CutPrefix(tail, ":"); ok {
		tail = strings.TrimSpace(rest)
		end := 0
		for end < len(tail) && tail[end] >= '0' && tail[end] <= '9' {
			end++
		}
		if end > 0 && (end == len(tail) || !citationWord(tail[end])) {
			if line, err := strconv.Atoi(tail[:end]); err == nil && line > 0 {
				return citation{path: path, line: line}, true
			}
		}
	}
	if tail == "" {
		return citation{}, false
	}
	close := tail[0]
	if close != '"' && close != '\'' && close != '`' {
		return citation{}, false
	}
	if end := strings.IndexByte(tail[1:], close); end >= 0 && strings.TrimSpace(tail[1:end+1]) != "" {
		return citation{path: path, quote: tail[1 : end+1]}, true
	}
	return citation{}, false
}

func citationBoundary(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n', '"', '\'', '`', '(', '[', '{':
		return true
	default:
		return false
	}
}

func citationWord(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '_'
}

// citationDiffFile answers whether the diff is readable and whether it carries
// the cited path. An unreadable diff is insufficient evidence, not a contradiction.
func citationDiffFile(diff, path string) (diffFile, bool, bool) {
	files, err := diffFiles(diff)
	if err != nil {
		return diffFile{}, false, false
	}
	for _, file := range files {
		if RowPath(file.path) == RowPath(path) {
			if !hasReadableHunk(file.text) {
				return file, true, false
			}
			return file, true, true
		}
	}
	return diffFile{}, false, true
}

func hasReadableHunk(text string) bool {
	for line := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if _, ok := strings.CutPrefix(line, "@@ "); ok {
			if _, ok := hunkStart(line); ok {
				return true
			}
		}
	}
	return false
}

// candidateDiffText keeps candidate-side hunk lines, including their diff prefix
// so both `+code` and `code` quotations remain checkable. Removed lines and file
// headers are not candidate evidence.
func candidateDiffText(text string) string {
	var candidate strings.Builder
	inHunk := false
	for line := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "@@ ") {
			inHunk = true
			continue
		}
		if !inHunk {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, " ") {
			candidate.WriteString(line)
			candidate.WriteByte('\n')
		}
	}
	return candidate.String()
}

// diffHasLine reports whether a new-file line is present as added or context
// evidence in one file's diff. Removed lines have no new-file number.
func diffHasLine(text string, line int) bool {
	current := 0
	inHunk := false
	for raw := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(raw, "@@ ") {
			start, ok := hunkStart(raw)
			if !ok {
				inHunk = false
				continue
			}
			current, inHunk = start, true
			continue
		}
		if !inHunk {
			continue
		}
		if raw == "" {
			continue
		}
		switch raw[0] {
		case '+', ' ':
			if current == line {
				return true
			}
			current++
		case '-', '\\':
			// A deletion or the no-newline marker contributes no new-file line.
		}
	}
	return false
}

func sourceHasLine(source string, line int) bool {
	if line <= 0 || strings.TrimSpace(source) == "" {
		return false
	}
	lines := strings.Split(strings.TrimSuffix(strings.ReplaceAll(source, "\r\n", "\n"), "\n"), "\n")
	return line <= len(lines)
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
// without a citation — because the default is always that the finding stands.
func (r Refutation) Refutes(id string, finding Finding, evidence CitationEvidence) bool {
	for _, judgment := range r.Refutations {
		if strings.TrimSpace(judgment.ID) != strings.TrimSpace(id) {
			continue
		}
		if judgment.Verdict != Refuted {
			return false
		}
		return ValidateCitation(CitationInput{
			FindingID:    id,
			Finding:      finding,
			Locator:      judgment.Why,
			DeclaredTree: r.Tree,
			Evidence:     evidence,
		}).Status == CitationValid
	}
	return false
}

// MaxWitnesses is the most refuters one tree may answer with. A cited claim needs
// one and an argued one two; three is the ceiling, because a block decided by a
// crowd is a block nobody can read.
const MaxWitnesses = 3

// RequiredWitnesses is how many cited refutations lift this finding: one when the
// reviewer quoted the line, two when it only argued. The class is the reviewer's
// own account, and absence reads as argued.
func RequiredWitnesses(f Finding) int {
	if f.EvidenceClass() == EvidenceCited {
		return 1
	}
	return 2
}

// Answers is every refuter's answer filed for one tree, in the order they landed.
// One answer is the common case; two lift an argued finding, and three is the
// ceiling.
type Answers []Refutation

// RefutedBy is how many different accounts refute the finding with a citation. A
// `stands`, or a `refuted` without one, is not one: the default never lifts.
// Counting accounts and not files is the whole promise -two answers from the same
// refuter are one opinion written twice, and a retry would buy a quorum nobody
// paid for.
func (a Answers) RefutedBy(id string, finding Finding, evidence CitationEvidence) int {
	seen := make(map[string]bool, len(a))
	for _, answer := range a {
		if !answer.Refutes(id, finding, evidence) {
			continue
		}
		seen[strings.ToLower(strings.TrimSpace(answer.Refuter))] = true
	}
	return len(seen)
}

// Lifts reports whether the answers are enough to lift this finding, which is the
// whole rule: the class decides how many, and a witness that stands is no witness.
func (a Answers) Lifts(id string, f Finding, evidence CitationEvidence) bool {
	return a.RefutedBy(id, f, evidence) >= RequiredWitnesses(f)
}

// IndexOf names the position of one finding inside its verdict, which is the id a
// refuter answers about. It compares every field that can be reported, not only
// the place: two findings on the same line are two findings, and naming the wrong
// one would count the votes of a row nobody is looking at.
func IndexOf(findings []Finding, finding Finding) int {
	for i := range findings {
		if findings[i] == finding {
			return i
		}
	}
	return -1
}

// BlockingAfter returns the first finding that still blocks, in the verdict's
// file order: causal, BLOCKER or CRITICAL, not dismissed with evidence, and not
// lifted by the refutations filed for this tree.
func BlockingAfter(findings []Finding, answers Answers, evidence CitationEvidence) (Finding, bool) {
	for _, finding := range BlockingAllAfter(findings, answers, evidence) {
		return finding, true
	}
	return Finding{}, false
}

// BlockingAllAfter is every row that still blocks, in the same order and by the
// same rule. A verdict can stand behind more than one row, so a gate that answers
// them has to see all of them: answering the first and stopping is how the second
// one disappears.
func BlockingAllAfter(findings []Finding, answers Answers, evidence CitationEvidence) []Finding {
	var blocking []Finding
	for i, finding := range findings {
		if IsDismissed(finding) {
			continue
		}
		if !finding.Causal || !finding.Severity.Blocks() {
			continue
		}
		if answers.Lifts(findingID(i), finding, evidence) {
			continue
		}
		blocking = append(blocking, finding)
	}
	return blocking
}

// DeriveAfter is the whole rule with refutations taken into account: a causal
// finding at BLOCKER or CRITICAL that nobody lifted requests changes, everything
// else approves.
func DeriveAfter(findings []Finding, answers Answers, evidence CitationEvidence) string {
	if _, blocking := BlockingAfter(findings, answers, evidence); blocking {
		return ChangesRequested
	}
	return Approved
}

// Rounds is every refutation a tree may collect, in the order they land. The
// readers loop over it so the store's shape is decided in one place.
func Rounds() []int {
	rounds := make([]int, 0, MaxWitnesses)
	for round := 1; round <= MaxWitnesses; round++ {
		rounds = append(rounds, round)
	}
	return rounds
}

// RefutationName is the file one refuter's answer for one tree lives in. The first
// answer keeps the name a refutation always had, so a tree refuted before the
// votes existed is still read; the next two are numbered, and a fourth has no name.
func RefutationName(tree string, round int) (string, error) {
	if !hashLike(tree) {
		return "", fmt.Errorf("the tree %q cannot be the name of a refutation: it must be a hash", tree)
	}
	if round < 1 || round > MaxWitnesses {
		return "", fmt.Errorf("a tree takes between 1 and %d refutations and this is round %d: the ceiling is what keeps a block from being outvoted", MaxWitnesses, round)
	}
	if round == 1 {
		return tree + ".refuted.json", nil
	}
	return tree + ".refuted." + strconv.Itoa(round) + ".json", nil
}

// RefutationPath is the refutation's path relative to the project root.
func RefutationPath(tree string, round int) (string, error) {
	name, err := RefutationName(tree, round)
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
	// Witness is which of the at most MaxWitnesses answers this prompt asks for.
	// Zero means the run did not say, and the prompt leaves the count out.
	Witness int
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
		out.WriteString("### " + entry.id + " · " + where(entry.finding) + " · " + string(entry.finding.Severity) + " · " + entry.finding.EvidenceClass() + "\n")
		out.WriteString(entry.finding.Why + "\n\n")
	}

	if in.Witness > 0 {
		out.WriteString("## Which witness you are\n")
		out.WriteString("This is answer " + strconv.Itoa(in.Witness) + " of at most " + strconv.Itoa(MaxWitnesses) + " for this tree. A finding whose evidence is\n")
		out.WriteString(EvidenceCited + " falls with one citation; one that is " + EvidenceArgued + " needs two. Your answer counts the same as\n")
		out.WriteString("any other: it is a vote with a reason, not a verdict.\n\n")
	}

	out.WriteString("## The diff (the only evidence that counts)\n")
	out.WriteString(section(in.Diff))
	out.WriteString("\n")

	out.WriteString("## The three questions, one answer per finding\n")
	out.WriteString("1. Is it causal of the diff, or was it already there? Cite the diff line that shows it.\n")
	out.WriteString("2. Is it a real defect that affects the user, or a preference?\n")
	out.WriteString("3. Is the evidence enough to defend it before the author?\n\n")

	out.WriteString("## How it is decided\n")
	out.WriteString("- A citation names the finding's path and either `path:line` or a quoted fragment from the diff.\n")
	out.WriteString("- Without a citation from the diff, the finding is not refuted: it stays in `" + Stands + "`. Doubt does not refute.\n")
	out.WriteString("- `" + Refuted + "` only when the three answers hold with the citation.\n")
	out.WriteString("- When in doubt: `" + Stands + "`. Over-refuting is worse than under-refuting.\n")
	out.WriteString("- WARNING and SUGGESTION are not judged: they do not block, and do not add new findings.\n")
	out.WriteString("- The five review rules still hold: a real defect, defensible with evidence,\n")
	out.WriteString("  and causal of this candidate. What was already there is a follow-up.\n\n")

	out.WriteString("## What to return\n")
	out.WriteString(`{"refutations":[{"id":"` + firstID(blocking) + `","verdict":"` + Stands + `|` + Refuted + `","why":"<the three answers and the citation: path:line or path \"quoted diff fragment\">"}]}` + "\n\n")
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
