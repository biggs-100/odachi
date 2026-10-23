package summary_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/summary"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// fakeFiles is the read-only disk this verb needs.
type fakeFiles struct{ files map[string][]byte }

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[path] = []byte(content)
	}
	return f
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("missing file: %s", path)
	}
	return data, nil
}

func taskPath(number int, slug string) string {
	return filepath.Join("raiz", "odd", "tasks", fmt.Sprintf("%04d-%s.md", number, slug))
}

// feature renders a canonical task document.
func feature(outcome string, doneSteps, totalSteps int, history ...string) string {
	doc := taskdoc.TaskDoc{
		Title:        "feature",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: outcome}},
		DoneWhen:     []string{"algo observable (R1)"},
		History:      history,
	}
	for i := 1; i <= totalSteps; i++ {
		doc.Steps = append(doc.Steps, taskdoc.Step{
			ID:     "T" + strconv.Itoa(i),
			Text:   "paso " + strconv.Itoa(i),
			Done:   i <= doneSteps,
			Commit: "4a1c9f",
		})
	}
	return doc.Render()
}

// TestSummaryOnAProjectWithoutFeaturesSaysSo: nothing to report is a sentence,
// never an error.
func TestSummaryOnAProjectWithoutFeaturesSaysSo(t *testing.T) {
	res := summary.Run(summary.Options{Dir: "raiz"}, newFakeFiles(nil))

	if res.Code != summary.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, summary.CodeOK, res.Message)
	}
	if len(res.Lines) > 1 {
		t.Errorf("an empty project printed %d lines, want one:\n%s", len(res.Lines), strings.Join(res.Lines, "\n"))
	}
	if !strings.Contains(res.Message, "odachi start") {
		t.Errorf("message %q does not name the command that creates the first feature", res.Message)
	}
}

// TestSummaryReportsEveryFeature is the derived report: per feature the slug,
// R1 verbatim, the steps, the open questions and the last three history lines;
// then one line of totals.
func TestSummaryReportsEveryFeature(t *testing.T) {
	first := "quiero reservar turnos desde el teléfono"
	second := "quiero facturar a mis obras sociales"

	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature(first, 1, 3,
			`L1 · "`+first+`"`,
			`L2 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"`,
			`L3 · VOS P1 · "No en la primera versión"`,
			`L4 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
			`L5 · OFICIO · "usé SQLite" (porque no hay dos escritores)`,
		),
		taskPath(2, "facturar-obras"): feature(second, 2, 2, `L1 · "`+second+`"`),
	})

	res := summary.Run(summary.Options{Dir: "raiz"}, files)

	if res.Code != summary.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, summary.CodeOK, res.Message)
	}

	block := strings.Join(res.Lines, "\n")

	for _, want := range []string{
		"0001-reservar-turnos",
		"0002-facturar-obras",
		`R1 · "` + first + `"`,
		`R1 · "` + second + `"`,
		"steps: 1 of 3",
		"steps: 2 of 2",
		"1 open question",
		"0 open questions",
		`L5 · OFICIO · "usé SQLite" (porque no hay dos escritores)`,
		`L4 · PREGUNTA P2 · "¿Hace falta WhatsApp?"`,
		`L3 · VOS P1 · "No en la primera versión"`,
	} {
		if !strings.Contains(block, want) {
			t.Errorf("the report does not contain %q:\n%s", want, block)
		}
	}

	// The first three history lines are beyond the window: the report shows the
	// last three, so the log fits on a screen.
	if strings.Contains(block, `L1 · "`+first+`"`) {
		t.Errorf("the report shows more than the last three history lines:\n%s", block)
	}

	last := res.Lines[len(res.Lines)-1]
	if !strings.Contains(last, "2 features") || !strings.Contains(last, "3 steps done of 5") || !strings.Contains(last, "1 open question") {
		t.Errorf("the totals line is %q", last)
	}
}

// TestSummaryNeverInventsNarration: the binary reports what is on disk. The
// teaching about the week is the agent's job, not the tool's.
func TestSummaryNeverInventsNarration(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 1, `L1 · "reservar turnos"`),
	})

	res := summary.Run(summary.Options{Dir: "raiz"}, files)

	for _, line := range res.Lines {
		for _, forbidden := range []string{"you learned", "learning", "lesson", "week summary"} {
			if strings.Contains(strings.ToLower(line), forbidden) {
				t.Errorf("the binary narrated: %q", line)
			}
		}
	}
}

// TestSummaryShowsADocumentWithoutHistory: an agent may still be working on the
// only step; the report says so instead of hiding the feature.
func TestSummaryShowsADocumentWithoutHistory(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 0, 1),
	})

	res := summary.Run(summary.Options{Dir: "raiz"}, files)

	if res.Code != summary.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, summary.CodeOK, res.Message)
	}
	block := strings.Join(res.Lines, "\n")
	if !strings.Contains(block, "0001-reservar-turnos") || !strings.Contains(block, "steps: 0 of 1") {
		t.Errorf("report:\n%s", block)
	}
}
