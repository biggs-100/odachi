package plan_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/plan"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// taskPath is where the first feature lives once odachi start wrote it.
var taskPath = filepath.Join("raiz", "odd", "tasks", "0001-reserva-de-turno.md")

const taskRel = "odd/tasks/0001-reserva-de-turno.md"

// fakeFiles is the disk in memory with a ledger: a test can assert what was
// written, what was backed up and in which order.
type fakeFiles struct {
	files   map[string][]byte
	wrote   []string
	backups []string
	order   []string
}

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

func (f *fakeFiles) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.order = append(f.order, "write "+path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	f.backups = append(f.backups, path)
	f.order = append(f.order, "backup "+path)
	return path + ".bak-1a2b3c4d", nil
}

func (f *fakeFiles) content(t *testing.T) string {
	t.Helper()
	data, ok := f.files[taskPath]
	if !ok {
		t.Fatalf("%s does not exist; files: %v", taskPath, f.paths())
	}
	return string(data)
}

func (f *fakeFiles) paths() []string {
	out := make([]string, 0, len(f.files))
	for path := range f.files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// seed is a task document in the exact canonical shape `odachi start` writes,
// with two requirements, one step already done and one open.
func seed() string {
	return taskdoc.TaskDoc{
		Title:        "reserva-de-turno",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "reservar turnos"}, {ID: "R2", Text: "ver la agenda"}},
		DoneWhen:     []string{"Reservo un turno y aparece en la agenda (R1)"},
		Steps: []taskdoc.Step{
			{ID: "T1", Text: "horario libre u ocupado", Requirement: "R1", Commit: "4a1c9f", Done: true},
			{ID: "T2", Text: "agenda del día", Requirement: "R2"},
		},
		History: []string{`L1 · "reservar turnos"`},
	}.Render()
}

func project() *fakeFiles {
	return newFakeFiles(map[string]string{taskPath: seed()})
}

func assertBlocked(t *testing.T, res plan.Result) string {
	t.Helper()
	if res.Code != plan.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeBlocked, res.Message)
	}
	if strings.TrimSpace(res.Message) == "" {
		t.Error("a blocked result must say why")
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("the message must be one line: %q", res.Message)
	}
	return res.Message
}

func lastStep(t *testing.T, files *fakeFiles) taskdoc.Step {
	t.Helper()
	steps := taskdoc.Parse(files.content(t)).Steps
	return steps[len(steps)-1]
}

// TestPlanAddsTheNextStep: the message Contract IX pins, and the document it
// leaves behind is canonical.
func TestPlanAddsTheNextStep(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Text: "cancelar desde el teléfono"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if want := "step T3 added to " + taskRel; res.Message != want {
		t.Errorf("Message = %q, want %q", res.Message, want)
	}
	if res.Task != taskRel {
		t.Errorf("Task = %q, want %q", res.Task, taskRel)
	}

	step := lastStep(t, files)
	if step.ID != "T3" || step.Text != "cancelar desde el teléfono" || step.Requirement != "" || step.Done {
		t.Errorf("new step = %+v, want T3 open without a requirement", step)
	}
	if got := files.content(t); got != taskdoc.Parse(got).Render() {
		t.Errorf("the document left on disk is not canonical:\n%s", got)
	}
	if want := []string{"backup " + taskPath, "write " + taskPath}; !slices.Equal(files.order, want) {
		t.Errorf("order = %v, want %v (backup before write)", files.order, want)
	}
}

// TestPlanLinksARequirement: --req links the step to an existing R and the link
// survives the canonical render.
func TestPlanLinksARequirement(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Text: "agenda ordenada por hora", Requirement: "R2"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	step := lastStep(t, files)
	if step.Requirement != "R2" {
		t.Errorf("Requirement = %q, want R2", step.Requirement)
	}
}

// TestPlanUnknownRequirementListsTheRequirements: the error tells the agent
// which links exist instead of only refusing.
func TestPlanUnknownRequirementListsTheRequirements(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: "algo", Requirement: "R9"}, files))

	for _, want := range []string{"R9", "R1", "R2"} {
		if !strings.Contains(message, want) {
			t.Errorf("message %q does not name %s", message, want)
		}
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("a blocked plan wrote %v / backed up %v", files.wrote, files.backups)
	}
}

// TestPlanRejectsBadTextBeforeWriting is the guard on `<texto>`: one line,
// short, not a bullet.
func TestPlanRejectsBadTextBeforeWriting(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"only spaces", "   "},
		{"multiline", "una linea\notra linea"},
		{"leading dash", "- un paso"},
		{"121 characters", strings.Repeat("a", 121)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := project()

			message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: c.text}, files))

			if strings.TrimSpace(message) == "" {
				t.Error("the rejection must say what is wrong with the text")
			}
			if len(files.wrote) != 0 || len(files.backups) != 0 {
				t.Errorf("a rejected text wrote %v / backed up %v", files.wrote, files.backups)
			}
			if got := files.content(t); got != seed() {
				t.Errorf("a rejected text rewrote the document:\n%s", got)
			}
		})
	}
}

// TestPlanNumbersStepsAcrossRuns: two runs append T3 and T4, and the document
// stays canonical, which is the whole reason plan exists.
func TestPlanNumbersStepsAcrossRuns(t *testing.T) {
	files := project()

	for _, text := range []string{"cancelar desde el teléfono", "confirmar por WhatsApp"} {
		if res := plan.Run(plan.Options{Dir: "raiz", Text: text}, files); res.Code != plan.CodeOK {
			t.Fatalf("plan %q: code %d (message: %s)", text, res.Code, res.Message)
		}
	}

	doc := taskdoc.Parse(files.content(t))
	ids := make([]string, 0, len(doc.Steps))
	for _, step := range doc.Steps {
		ids = append(ids, step.ID)
	}
	if want := []string{"T1", "T2", "T3", "T4"}; !slices.Equal(ids, want) {
		t.Fatalf("step ids = %v, want %v", ids, want)
	}
	if got := files.content(t); got != taskdoc.Parse(got).Render() {
		t.Errorf("the document left on disk is not canonical:\n%s", got)
	}
}

// TestPlanDoneMarksAndRecordsTheCommit: Contract IX's second message, and the
// step on disk carries the commit.
func TestPlanDoneMarksAndRecordsTheCommit(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T2", Commit: "deadbee"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if want := "step T2 done · deadbee"; res.Message != want {
		t.Errorf("Message = %q, want %q", res.Message, want)
	}

	step := taskdoc.Parse(files.content(t)).Steps[1]
	if !step.Done || step.Commit != "deadbee" {
		t.Errorf("step = %+v, want done with commit deadbee", step)
	}
	if res.Task != taskRel {
		t.Errorf("Task = %q, want %q", res.Task, taskRel)
	}
	if want := []string{"backup " + taskPath, "write " + taskPath}; !slices.Equal(files.order, want) {
		t.Errorf("order = %v, want %v (backup before write)", files.order, want)
	}
}

// TestPlanDoneWithoutCommitStillClosesTheStep: a step can be done before it is
// committed, or without a commit at all.
func TestPlanDoneWithoutCommitStillClosesTheStep(t *testing.T) {
	files := project()

	res := plan.Run(plan.Options{Dir: "raiz", Done: "t2"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if want := "step T2 done"; res.Message != want {
		t.Errorf("Message = %q, want %q", res.Message, want)
	}
	if step := taskdoc.Parse(files.content(t)).Steps[1]; !step.Done || step.Commit != "" {
		t.Errorf("step = %+v, want done with no commit", step)
	}
}

// TestPlanUnknownStepListsTheOpenOnes: the error names the ids the agent can
// use, and only the ones that are still open.
func TestPlanUnknownStepListsTheOpenOnes(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Done: "T9"}, files))

	if !strings.Contains(message, "T9") {
		t.Errorf("message %q does not name the unknown step", message)
	}
	if !strings.Contains(message, "T2") {
		t.Errorf("message %q does not list the open step T2", message)
	}
	if strings.Contains(message, "T1") {
		t.Errorf("message %q lists T1, which is already done", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked --done wrote %v", files.wrote)
	}
}

// TestPlanAlreadyDoneIsANoOp: closing a closed step is not an error and writes
// nothing, because the log of the work must not change.
func TestPlanAlreadyDoneIsANoOp(t *testing.T) {
	files := project()
	before := files.content(t)

	res := plan.Run(plan.Options{Dir: "raiz", Done: "T1"}, files)

	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, "was already done") {
		t.Errorf("Message = %q, want it to say was already done", res.Message)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("a no-op wrote %v / backed up %v", files.wrote, files.backups)
	}
	if got := files.content(t); got != before {
		t.Errorf("a no-op rewrote the document:\n%s", got)
	}
}

// TestPlanCommitWithoutDoneIsBlocked: a commit with no step names nothing to
// mark, and the message names --done.
func TestPlanCommitWithoutDoneIsBlocked(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Commit: "deadbee"}, files))

	if !strings.Contains(message, "--done") {
		t.Errorf("message %q does not name --done", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked --commit wrote %v", files.wrote)
	}
}

// TestPlanOneFormAtATime: adding and closing are two different commands.
func TestPlanOneFormAtATime(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: "un paso", Done: "T2"}, files))

	if !strings.Contains(message, "one form") {
		t.Errorf("message %q does not explain that one call does one thing", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked mixed run wrote %v", files.wrote)
	}
}

// TestPlanRequirementWithoutTextIsBlocked: --req only makes sense with a new
// step, and the message says so instead of ignoring the flag.
func TestPlanRequirementWithoutTextIsBlocked(t *testing.T) {
	files := project()

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Requirement: "R1"}, files))

	if !strings.Contains(message, `odachi plan "<text>"`) {
		t.Errorf("message %q does not show the shape of the command", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked --req wrote %v", files.wrote)
	}
}

// TestPlanWithoutATaskDocumentNamesStart: no document to plan into, and the
// message names the command that creates one.
func TestPlanWithoutATaskDocumentNamesStart(t *testing.T) {
	files := newFakeFiles(nil)

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: "primer paso"}, files))

	if !strings.Contains(message, "odachi start") {
		t.Errorf("message %q does not name odachi start", message)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("a blocked plan wrote %v / backed up %v", files.wrote, files.backups)
	}
}

// TestPlanRefusesToDropAHandwrittenSection: the shared guard, exercised through
// the second verb that uses it. Nothing is written, asserted on the disk.
func TestPlanRefusesToDropAHandwrittenSection(t *testing.T) {
	handwritten := seed() + "\n## Notas\nuna idea que el humano agregó a mano\n"
	files := newFakeFiles(map[string]string{taskPath: handwritten})

	message := assertBlocked(t, plan.Run(plan.Options{Dir: "raiz", Text: "un paso más"}, files))

	if !strings.Contains(message, "## History") {
		t.Errorf("message %q does not say where the notes belong", message)
	}
	if len(files.wrote) != 0 {
		t.Errorf("a blocked plan wrote %v", files.wrote)
	}
	if len(files.backups) != 0 {
		t.Errorf("a blocked plan backed up %v", files.backups)
	}
	if got := files.content(t); got != handwritten {
		t.Errorf("the document was rewritten:\n%s", got)
	}
}

// TestPlanTouchesOnlyTheTaskDocument: brief, profile and fence belong to other
// verbs and are byte-identical after a plan.
func TestPlanTouchesOnlyTheTaskDocument(t *testing.T) {
	briefPath := filepath.Join("raiz", "brief.md")
	profilePath := filepath.Join("raiz", ".odachi", "profile.toml")
	fencePath := filepath.Join("raiz", ".odachi", "fence.toml")
	files := newFakeFiles(map[string]string{
		taskPath:    seed(),
		briefPath:   "# reserva-de-turno\n",
		profilePath: "name = \"novato\"\n",
		fencePath:   "[[rule]]\nnot = \"pagos\"\n",
	})

	res := plan.Run(plan.Options{Dir: "raiz", Text: "un paso más"}, files)
	if res.Code != plan.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, plan.CodeOK, res.Message)
	}

	if got := string(files.files[briefPath]); got != "# reserva-de-turno\n" {
		t.Errorf("brief.md changed: %q", got)
	}
	if got := string(files.files[profilePath]); got != "name = \"novato\"\n" {
		t.Errorf("profile.toml changed: %q", got)
	}
	if got := string(files.files[fencePath]); got != "[[rule]]\nnot = \"pagos\"\n" {
		t.Errorf("fence.toml changed: %q", got)
	}
	if len(files.backups) != 1 || files.backups[0] != taskPath {
		t.Errorf("backups = %v, want only the task document", files.backups)
	}
}
