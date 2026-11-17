package archtest_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	assets "github.com/biggs-100/odachi/internal/assets/review"
)

const normativeBaseline = "normative.baseline"

// TestTheLensesKeepTheirNormativeLines: a byte count cannot tell a rule that
// survived from a rule rewritten to the same length. The four lens prompts are
// the product the reviewer receives, so every line of each one is pinned by its
// words. The sweep runs both ways: a line gone and a line added are the same
// kind of unreviewed change, and both name the line in the failure.
//
// A real edit is still possible: change the line in normative.baseline, put a
// `# why:` comment above it, and the sweep follows. That is a deliberate commit
// a reviewer can read, which is the point of the second file.
func TestTheLensesKeepTheirNormativeLines(t *testing.T) {
	want := readNormativeBaseline(t)
	got := currentNormativeLines()

	for _, issue := range sweepIssues(want, got) {
		t.Error(issue)
	}
}

// sweepIssues compares the two multisets in both directions and names the line
// behind each difference. A count, not just presence, catches a rule pasted
// twice as the same unreviewed edit it would catch a rule removed.
func sweepIssues(want, got map[string]int) []string {
	var issues []string
	for key, count := range want {
		if got[key] != count {
			issues = append(issues, "normative line lost: "+key+
				" ("+strconv.Itoa(count)+" in the baseline, "+strconv.Itoa(got[key])+" in the shipping lens)")
		}
	}
	for key, count := range got {
		if want[key] != count {
			issues = append(issues, "new normative line: "+key+
				" ("+strconv.Itoa(count)+" in the shipping lens, "+strconv.Itoa(want[key])+" in the baseline); write it in "+
				normativeBaseline+" with its reason")
		}
	}
	return issues
}

// TestTheSweepNamesBothDirections: a changed line is one edit seen twice. The
// message has to say what vanished and what arrived, or the reader has to open
// two files to find the same sentence.
func TestTheSweepNamesBothDirections(t *testing.T) {
	want := map[string]int{"lens:risk\t- a rule that was here": 1}
	got := map[string]int{"lens:risk\t- a rule that arrived": 1}

	joined := strings.Join(sweepIssues(want, got), "\n")
	for _, phrase := range []string{
		"normative line lost", "a rule that was here",
		"new normative line", "a rule that arrived", "with its reason",
	} {
		if !strings.Contains(joined, phrase) {
			t.Errorf("the sweep does not say %q:\n%s", phrase, joined)
		}
	}
}

// readNormativeBaseline reads surface<TAB>line rows. Comments carry the reason
// a line changed and are not evidence themselves.
func readNormativeBaseline(t *testing.T) map[string]int {
	t.Helper()
	file, err := os.Open(filepath.Join("..", "..", "internal", "archtest", normativeBaseline))
	if err != nil {
		t.Fatalf("read %s: %v", normativeBaseline, err)
	}
	defer file.Close()

	rows := map[string]int{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(text) == "" || strings.HasPrefix(text, "#") {
			continue
		}
		at := strings.IndexByte(text, '\t')
		if at < 0 {
			t.Fatalf("%s:%d has no tab between the surface and the line: %q", normativeBaseline, line, text)
		}
		rows[text]++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", normativeBaseline, err)
	}
	return rows
}

// currentNormativeLines strips the registry frontmatter and takes every line
// that reaches the model: the inherited rules, the section headings and each
// bullet. A heading that disappears is as much a change as a rule that does.
func currentNormativeLines() map[string]int {
	rows := map[string]int{}
	for _, name := range assets.Names() {
		body, err := assets.Lens(name)
		if err != nil {
			continue
		}
		text := strings.ReplaceAll(body, "\r\n", "\n")
		if strings.HasPrefix(text, "---\n") {
			if end := strings.Index(text[4:], "\n---\n"); end >= 0 {
				text = text[4+end+5:]
			}
		}
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			rows["lens:"+name+"\t"+line]++
		}
	}
	return rows
}
