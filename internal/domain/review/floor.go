package review

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// This file is the deterministic floor of ADR 0016: what the candidate's own
// writing is, read from the diff before any model runs. It does three things,
// all reproducible in milliseconds:
//
//   - it names the audit set, the new-file lines a finding may cite;
//   - it names the categories those lines touch, from a fixed vocabulary (the
//     categories the four lenses exist for);
//   - it applies earned blocking: a finding that can block must cite that set,
//     or its severity goes down one tier before the verdict is stored.
//
// It never changes which lenses a tree asks for: that stays ADR 0010's
// path-based tiering, and the fence of this feature keeps it that way. The
// class here is the same class reported to the reviewer, not a second
// selection rule.

// FloorContext is how many lines around a changed line still count as the
// finding's immediate context. Three is enough for a signature and its caller,
// and small enough that a finding about the rest of the file is not "here".
const FloorContext = 3

// FloorRow is one changed file's audit set: the new-file line numbers the
// candidate added, and the signal categories their text carries.
type FloorRow struct {
	Path    string
	Lines   []int
	Signals []string
}

// Floor is the deterministic reading of one candidate. Class is the path class
// the lenses were chosen from, reported so the reviewer reads the same sentence
// the gate does.
type Floor struct {
	Class Tier
	Rows  []FloorRow
}

// floorSignals is the fixed vocabulary the floor reads on added lines. It is
// small on purpose: every token names a category a lens already exists for, and
// a token nobody acts on would be noise dressed as evidence.
var floorSignals = []struct {
	name   string
	tokens []string
}{
	{name: "concurrency", tokens: []string{"go func", "chan ", "sync.", "atomic.", "mutex", "<-"}},
	{name: "secrets", tokens: []string{"password", "secret", "token", "api_key", "apikey", "private_key"}},
	{name: "network", tokens: []string{"http.", "net.", "sql.", "os.open", "os.create", "exec."}},
	{name: "failure", tokens: []string{"panic(", "os.exit", "log.fatal"}},
	{name: "time", tokens: []string{"time.sleep", "sleep("}},
}

// FloorFor reads a unified diff into the candidate's audit set. Only added
// lines are the candidate's own writing; a hunk's context lines belong to
// nobody's citation. A deletion has no new-file line, so it contributes none: a
// finding about a removed file has to cite elsewhere or lose its blocking
// claim. A diff with no readable hunk is a floor with no rows, and a floor with
// no rows judges nothing.
//
// The class is the content-aware one: the path class is the floor and these
// rows can raise it, never lower it (see Assess).
func FloorFor(paths []string, diff string) Floor {
	floor := Floor{Rows: withoutPlumbing(readRows(diff))}
	floor.Class = assessClass(paths, floor.Rows)
	return floor
}

// withoutPlumbing drops the rows of the files the tool installed. Their text
// talks about secrets and commands because that is their job, and a signal read
// there would raise the class of the project that adopted the tool: a skill that
// says "never hardcode a secret" is not a security change in this repository.
func withoutPlumbing(rows []FloorRow) []FloorRow {
	kept := rows[:0]
	for _, row := range rows {
		if isPlumbing(row.Path) {
			continue
		}
		kept = append(kept, row)
	}
	return kept
}

// Assess is the class of the candidate with the diff's content taken into
// account: the path class is the floor, and ordinary code whose added lines
// spawn concurrency or touch secrets is raised to full. Content adds review; it
// never removes it, and prose never turns a document into code.
func Assess(paths []string, diff string) Tier {
	return FloorFor(paths, diff).Class
}

// assessClass is the one rule both Assess and Need answer, so the review
// selection and the gate's demand can never disagree.
func assessClass(paths []string, rows []FloorRow) Tier {
	pathClass := Classify(paths)
	if pathClass != Focused {
		// Passive has nothing for a lens to look at and Full already asks for
		// everything: only ordinary code has room to be raised.
		return pathClass
	}

	for _, row := range rows {
		if slices.Contains(row.Signals, "concurrency") || slices.Contains(row.Signals, "secrets") {
			return Full
		}
	}
	return Focused
}

// rowsFor is the floor rows of one path, compared in the same spelling coverage
// uses.
func (f Floor) rowsFor(path string) []FloorRow {
	normalized := normalizePath(path)
	var rows []FloorRow
	for _, row := range f.Rows {
		if normalizePath(row.Path) == normalized {
			rows = append(rows, row)
		}
	}
	return rows
}

// readRows is the diff parser: the candidate's added lines, with the signals
// their text carries.
func readRows(diff string) []FloorRow {
	var rows []FloorRow

	var row *FloorRow
	newLine := 0
	inHunk := false

	for _, line := range strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "+++ "):
			row = nil
			inHunk = false
			path := strings.TrimPrefix(line, "+++ ")
			path = strings.TrimPrefix(path, "b/")
			if path == "" || path == "/dev/null" {
				continue
			}
			rows = append(rows, FloorRow{Path: path})
			row = &rows[len(rows)-1]

		case strings.HasPrefix(line, "@@ "):
			inHunk = true
			target, ok := hunkStart(line)
			if !ok {
				inHunk = false
				continue
			}
			newLine = target

		case !inHunk || row == nil:
			continue

		case strings.HasPrefix(line, "+"):
			row.Lines = append(row.Lines, newLine)
			row.Signals = addSignals(row.Signals, strings.TrimPrefix(line, "+"))
			newLine++

		case strings.HasPrefix(line, "-"):
			// A removed line has no new-file number: it is not an audit target.

		case strings.HasPrefix(line, "\\"):
			// "\ No newline at end of file" is metadata, not a line.

		case strings.HasPrefix(line, " "):
			newLine++
		}
	}

	return rows
}

// hunkStart reads the new-file first line of a hunk header: `@@ -a,b +c,d @@`
// answers c and whether the header was readable.
func hunkStart(header string) (int, bool) {
	parts := strings.Split(header, " ")
	for _, part := range parts {
		if !strings.HasPrefix(part, "+") {
			continue
		}
		span := strings.TrimPrefix(part, "+")
		if comma := strings.Index(span, ","); comma >= 0 {
			span = span[:comma]
		}
		start, err := strconv.Atoi(span)
		if err != nil {
			return 0, false
		}
		return start, true
	}
	return 0, false
}

// addSignals adds the categories one added line carries, keeping the fixed
// order and no duplicates, so the same diff always produces the same row bytes.
func addSignals(signals []string, line string) []string {
	lower := strings.ToLower(line)
	for _, signal := range floorSignals {
		if slices.Contains(signals, signal.name) {
			continue
		}
		for _, token := range signal.tokens {
			if strings.Contains(lower, token) {
				signals = append(signals, signal.name)
				break
			}
		}
	}
	return signals
}

// Earned applies the floor's citation rule to the findings that can block: a
// causal BLOCKER or CRITICAL must cite one of the changed lines or a line
// within FloorContext of one. Without that citation it goes down one tier, and
// the count says how many did. Everything else is untouched: a WARNING was
// never going to block, and a non-causal finding is a follow-up by definition.
//
// A finding about a path the floor does not know is not judged: a deleted file
// or an invented path has no audit set to compare with, and silence is not
// evidence either way.
func Earned(findings []Finding, floor Floor) ([]Finding, int) {
	earned := make([]Finding, len(findings))
	copy(earned, findings)

	downgraded := 0
	for i, finding := range earned {
		if !finding.Blocks() {
			continue
		}
		row, known := floor.Row(finding.Path)
		if !known || row.Earns(finding.Line) {
			continue
		}
		earned[i].Severity = Downgrade(finding.Severity)
		downgraded++
	}
	return earned, downgraded
}

// Row answers one path's audit set, if the floor knows it.
func (f Floor) Row(path string) (FloorRow, bool) {
	normalized := normalizePath(path)
	for _, row := range f.Rows {
		if normalizePath(row.Path) == normalized {
			return row, true
		}
	}
	return FloorRow{}, false
}

// Earns reports whether a finding's line cites this row: one of the added
// lines, or a line within FloorContext of one. Line 0 cites nothing.
func (r FloorRow) Earns(line int) bool {
	if line <= 0 {
		return false
	}
	for _, changed := range r.Lines {
		if math.Abs(float64(line-changed)) <= FloorContext {
			return true
		}
	}
	return false
}

// Downgrade is one tier down: BLOCKER to CRITICAL, CRITICAL to WARNING, WARNING
// to SUGGESTION, and SUGGESTION stays the lowest honest word.
func Downgrade(severity Severity) Severity {
	switch severity {
	case Blocker:
		return Critical
	case Critical:
		return Warning
	case Warning:
		return Suggestion
	default:
		return Suggestion
	}
}

// FloorLines renders the audit set the way the prompt states it: ranges a human
// and a model read the same way.
func FloorLines(lines []int) string {
	if len(lines) == 0 {
		return "none"
	}

	sorted := append([]int(nil), lines...)
	slices.Sort(sorted)

	var parts []string
	start, end := sorted[0], sorted[0]
	flush := func() {
		if start == end {
			parts = append(parts, strconv.Itoa(start))
			return
		}
		parts = append(parts, fmt.Sprintf("%d-%d", start, end))
	}

	for _, line := range sorted[1:] {
		switch {
		case line == end+1:
			end = line
		case line <= end:
			// a duplicate is the same line
		default:
			flush()
			start, end = line, line
		}
	}
	flush()

	return strings.Join(parts, ", ")
}
