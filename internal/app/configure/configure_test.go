package configure_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/configure"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// valid is a state the wizard can reach: every choice made, one lens model.
func valid() configure.State {
	return configure.State{
		Clients:    persona.SelectionBoth,
		Skills:     persona.SkillsHome,
		Voice:      "novato",
		ReviewMode: review.ModeOn,
		LensModels: map[string]string{"reliability": "opencode-go/deepseek-v4.1-flash"},
		Project:    true,
	}
}

// TestValidateAcceptsACompleteState: the state the wizard can produce is accepted.
func TestValidateAcceptsACompleteState(t *testing.T) {
	if problems := configure.Validate(valid()); len(problems) != 0 {
		t.Fatalf("Validate = %q, want no problem", problems)
	}
}

// TestValidateRefusesAnEmptyClient: the wizard cannot apply a configuration that
// names no client.
func TestValidateRefusesAnEmptyClient(t *testing.T) {
	in := valid()
	in.Clients = ""
	problems := configure.Validate(in)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, " "), "cliente") {
		t.Fatalf("Validate = %q, want a problem that names the client", problems)
	}
}

// TestValidateRefusesAnEmptySkillsScope: home and project are the two scopes, and
// an empty one is not a third.
func TestValidateRefusesAnEmptySkillsScope(t *testing.T) {
	in := valid()
	in.Skills = ""
	problems := configure.Validate(in)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, " "), "scope") {
		t.Fatalf("Validate = %q, want a problem that names the scope", problems)
	}
}

// TestValidateRefusesAnUnknownProfile: the wizard offers the three voices the
// binary ships, and a name outside them is refused with the ones it knows.
func TestValidateRefusesAnUnknownProfile(t *testing.T) {
	in := valid()
	in.Voice = "guru"
	problems := configure.Validate(in)
	joined := strings.Join(problems, " ")
	for _, want := range []string{"guru", "novato", "senior", "ci"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Validate = %q, want %q", problems, want)
		}
	}
}

// TestValidateRefusesAModelThatIsNotAnIdentifier: a lens model replaces {model}
// inside a command, so it is an identifier and not free text.
func TestValidateRefusesAModelThatIsNotAnIdentifier(t *testing.T) {
	in := valid()
	in.LensModels = map[string]string{"risk": "a b"}
	problems := configure.Validate(in)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, " "), "risk") {
		t.Fatalf("Validate = %q, want a problem that names the lens", problems)
	}
}

// TestPlanOrdersInstallVoiceReviewAndModels: the operations have an order that
// leaves the project writable before it is configured.
func TestPlanOrdersInstallVoiceReviewAndModels(t *testing.T) {
	got := configure.Plan(valid())
	ids := make([]string, 0, len(got))
	for _, op := range got {
		ids = append(ids, op.ID)
	}
	want := []string{"install", "voice", "review", "models"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("Plan ids = %q, want %q", ids, want)
	}
}

// TestPlanSkipsTheModelsWhenTheReviewIsOff: a model of a lens nobody runs is a
// line nobody reads.
func TestPlanSkipsTheModelsWhenTheReviewIsOff(t *testing.T) {
	in := valid()
	in.ReviewMode = review.ModeOff
	got := configure.Plan(in)
	ids := make([]string, 0, len(got))
	for _, op := range got {
		ids = append(ids, op.ID)
	}
	want := []string{"install", "voice", "review"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("Plan ids = %q, want %q", ids, want)
	}
}

// TestApplyRunsEveryOperationInOrder: the runner receives the operations in the
// plan's order, and every step ends done.
func TestApplyRunsEveryOperationInOrder(t *testing.T) {
	run := &fakeRunner{}
	got := configure.Apply(valid(), run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if want := []string{"install", "voice", "review", "models"}; !reflect.DeepEqual(run.ops(), want) {
		t.Fatalf("runner calls = %q, want %q", run.ops(), want)
	}
	for _, step := range got.Steps {
		if step.Status != configure.Done {
			t.Errorf("step %s = %s, want done", step.ID, step.Status)
		}
	}
}

// TestApplyStopsAtTheFirstFailureAndLeavesTheRestPending: a failure is named, the
// operations after it are not run, and the report says they remain.
func TestApplyStopsAtTheFirstFailureAndLeavesTheRestPending(t *testing.T) {
	run := &fakeRunner{failAt: "voice"}
	got := configure.Apply(valid(), run)

	if got.Code != configure.CodeBlocked {
		t.Fatalf("Code = %d, want %d", got.Code, configure.CodeBlocked)
	}
	if want := []string{"install", "voice"}; !reflect.DeepEqual(run.ops(), want) {
		t.Fatalf("runner calls = %q, want %q: the run continued after the failure", run.ops(), want)
	}
	if got.Message == "" {
		t.Fatal("Message is empty, want the failed operation named")
	}
	statuses := map[string]configure.Status{}
	for _, step := range got.Steps {
		statuses[step.ID] = step.Status
	}
	if statuses["voice"] != configure.Failed {
		t.Errorf("voice = %s, want failed", statuses["voice"])
	}
	for _, id := range []string{"review", "models"} {
		if statuses[id] != configure.Pending {
			t.Errorf("%s = %s, want pending", id, statuses[id])
		}
	}
}

// TestApplyClearsTheLensTableWhenTheReviewIsOn: an empty table is a decision the
// human made by deleting the last model, and the runner receives it as one; nil
// is the caller that never touched the table.
func TestApplyClearsTheLensTableWhenTheReviewIsOn(t *testing.T) {
	in := valid()
	in.LensModels = map[string]string{}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if run.models == nil || len(run.models) != 0 {
		t.Fatalf("models = %v, want an explicit empty table", run.models)
	}
	if got := strings.Join(run.ops(), ","); got != "install,voice,review,models" {
		t.Fatalf("calls = %q, want the models operation with an empty table", got)
	}
}

// TestApplyNormalisesTheLensTable: the runner receives the canonical table the
// file format expects, not the spelling the caller typed.
func TestApplyNormalisesTheLensTable(t *testing.T) {
	in := valid()
	in.LensModels = map[string]string{" Risk ": " opencode-go/x "}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if got := run.models["risk"]; got != "opencode-go/x" {
		t.Fatalf("models = %v, want the folded key and the trimmed value", run.models)
	}
	if _, ok := run.models[" Risk "]; ok {
		t.Fatalf("models = %v, want no raw key", run.models)
	}
}

// TestApplyLeavesTheTableAloneWhenItIsNil: a caller that never touched the table
// does not clear it, and the models operation is not planned at all.
func TestApplyLeavesTheTableAloneWhenItIsNil(t *testing.T) {
	in := valid()
	in.LensModels = nil
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if calls := strings.Join(run.ops(), ","); calls != "install,voice,review" {
		t.Fatalf("calls = %q, want no models operation for a nil table", calls)
	}
	if run.models != nil {
		t.Fatalf("models = %v, want nil: the table was never touched", run.models)
	}
}

// TestApplyRefusesAnInvalidStateWithoutTouchingTheRunner: validation runs before
// the first write, so a state the wizard should have refused changes nothing.
func TestApplyRefusesAnInvalidStateWithoutTouchingTheRunner(t *testing.T) {
	in := valid()
	in.Voice = "guru"
	run := &fakeRunner{}
	got := configure.Apply(in, run)

	if got.Code != configure.CodeBlocked {
		t.Fatalf("Code = %d, want %d", got.Code, configure.CodeBlocked)
	}
	if len(run.calls) != 0 {
		t.Fatalf("runner calls = %q, want none", run.ops())
	}
}

// TestApplyPassesTheDetectedProjectToInstall: the same state configures the
// machine or initializes the folder, and the runner is told which.
func TestApplyPassesTheDetectedProjectToInstall(t *testing.T) {
	run := &fakeRunner{}
	configure.Apply(valid(), run)
	if len(run.calls) == 0 || !run.calls[0].project {
		t.Fatalf("install project = false, want true")
	}

	run = &fakeRunner{}
	in := valid()
	in.Project = false
	configure.Apply(in, run)
	if len(run.calls) == 0 || run.calls[0].project {
		t.Fatalf("install project = true, want false")
	}
}

// fakeRunner records the operations it ran and can fail one of them.
type fakeRunner struct {
	calls  []call
	failAt string
	models map[string]string
}

type call struct {
	op      string
	project bool
	value   string
}

func (f *fakeRunner) record(op, value string) error {
	f.calls = append(f.calls, call{op: op, value: value})
	if f.failAt == op {
		return errFailed
	}
	return nil
}

func (f *fakeRunner) Install(project bool, clients, skills string) error {
	f.calls = append(f.calls, call{op: "install", project: project, value: clients + "/" + skills})
	if f.failAt == "install" {
		return errFailed
	}
	return nil
}

func (f *fakeRunner) Voice(profile string) error { return f.record("voice", profile) }

func (f *fakeRunner) Review(mode string) error { return f.record("review", mode) }

func (f *fakeRunner) Models(models map[string]string) error {
	f.models = models
	return f.record("models", "")
}

func (f *fakeRunner) ops() []string {
	out := make([]string, 0, len(f.calls))
	for _, c := range f.calls {
		out = append(out, c.op)
	}
	return out
}

// errFailed is the failure a fake operation returns.
var errFailed = failedError("the operation failed")

type failedError string

func (e failedError) Error() string { return string(e) }
