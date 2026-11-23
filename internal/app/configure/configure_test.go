package configure_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/configure"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
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

// TestPlanWritesTheVoiceWhereTheHumanChose: the machine scope writes the
// default outside the project, so the apply must call a different runner op
// and never touch odd/profile.toml (0357).
func TestPlanWritesTheVoiceWhereTheHumanChose(t *testing.T) {
	in := valid()
	in.VoiceScope = configure.VoiceMachine
	ids := make([]string, 0)
	for _, op := range configure.Plan(in) {
		ids = append(ids, op.ID)
	}
	want := []string{"install", "globalvoice", "review", "models"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("Plan ids = %q, want %q", ids, want)
	}

	run := &fakeRunner{}
	if result := configure.Apply(in, run); result.Code != configure.CodeOK {
		t.Fatalf("Apply = %d, want OK: %s", result.Code, result.Message)
	}
	if !slices.Contains(run.ops(), "globalvoice") || slices.Contains(run.ops(), "voice") {
		t.Fatalf("calls = %q, want globalvoice without voice", run.ops())
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

// TestPlanWritesTheVerifierOnlyWhenItDiffers: a row that still shows the value
// the file carries is not a write, and a section nobody opened is not in the
// plan at all (0337).
func TestPlanWritesTheVerifierOnlyWhenItDiffers(t *testing.T) {
	same := "cheap/model"
	in := valid()
	in.Verify = &configure.ModelChoice{Chosen: same, Detected: same}
	if ids := planIDs(in); slices.Contains(ids, "verify") {
		t.Errorf("plan = %v, want no verifier write for the value already there", ids)
	}

	changed := "other/model"
	in.Verify = &configure.ModelChoice{Chosen: changed, Detected: same}
	if ids := planIDs(in); !slices.Contains(ids, "verify") {
		t.Errorf("plan = %v, want the verifier write", ids)
	}
}

// TestApplyWritesTheVerifierThroughTheRunner: the chosen model reaches the verb
// that owns odd/review.toml, and it is the last write of the plan (0337).
func TestApplyWritesTheVerifierThroughTheRunner(t *testing.T) {
	in := valid()
	in.Verify = &configure.ModelChoice{Chosen: "cheap/model#fast", Detected: ""}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if run.verify != "cheap/model#fast" {
		t.Fatalf("verify = %q, want the chosen model", run.verify)
	}
	if ids := run.ops(); ids[len(ids)-1] != "verify" {
		t.Errorf("calls = %v, want the verifier written after the rest", ids)
	}
}

// TestApplyLeavesTheVerifierAloneWhenTheSectionWasNeverOpened: nil is a caller
// that never showed the section, and the file keeps its line.
func TestApplyLeavesTheVerifierAloneWhenTheSectionWasNeverOpened(t *testing.T) {
	run := &fakeRunner{}

	got := configure.Apply(valid(), run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if slices.Contains(run.ops(), "verify") {
		t.Errorf("calls = %v, want no verifier operation", run.ops())
	}
	if run.verify != "" {
		t.Errorf("verify = %q, want nothing written", run.verify)
	}
}

// TestApplyClearsTheVerifierWhenTheRowIsEmpty: an empty row is the project
// naming no verifier, and the operation runs with the empty value so Render
// removes the line (0337).
func TestApplyClearsTheVerifierWhenTheRowIsEmpty(t *testing.T) {
	in := valid()
	in.Verify = &configure.ModelChoice{Chosen: "", Detected: "cheap/model"}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if !slices.Contains(run.ops(), "verify") {
		t.Fatalf("calls = %v, want the verifier operation", run.ops())
	}
	if run.verify != "" {
		t.Errorf("verify = %q, want an empty value", run.verify)
	}
}

// TestApplySendsOnlyTheChangedClientModelsToInstall: a client whose file already
// carries the chosen model is not named in the install, which is what makes an
// unchanged section write nothing (0337).
func TestApplySendsOnlyTheChangedClientModelsToInstall(t *testing.T) {
	in := valid()
	in.ClientModels = map[string]configure.ModelChoice{
		"opencode": {Chosen: "new/model", Detected: "old/model"},
		"claude":   {Chosen: "sonnet", Detected: "sonnet"},
	}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	want := map[string]string{"opencode": "new/model"}
	if !reflect.DeepEqual(run.clientModels, want) {
		t.Fatalf("install models = %v, want %v", run.clientModels, want)
	}
}

// TestApplyDropsAClearedClientModel: an empty row means "no model", and there is
// no verb that removes a client's line, so the install is told nothing about it
// rather than writing an empty identifier (0337).
func TestApplyDropsAClearedClientModel(t *testing.T) {
	in := valid()
	in.ClientModels = map[string]configure.ModelChoice{
		"claude": {Chosen: "", Detected: "sonnet"},
	}
	run := &fakeRunner{}

	got := configure.Apply(in, run)

	if got.Code != configure.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, configure.CodeOK, got.Message)
	}
	if run.clientModels != nil {
		t.Fatalf("install models = %v, want none", run.clientModels)
	}
}

// TestValidateRefusesAModelThatIsNotAnIdentifierInTheModelsSection: the section
// obeys the same rule the verb obeys, and the state is refused before a write.
func TestValidateRefusesAModelThatIsNotAnIdentifierInTheModelsSection(t *testing.T) {
	in := valid()
	in.Verify = &configure.ModelChoice{Chosen: "a b"}
	in.ClientModels = map[string]configure.ModelChoice{"claude": {Chosen: "x;rm"}}

	problems := strings.Join(configure.Validate(in), " ")
	for _, want := range []string{"verificador", "claude"} {
		if !strings.Contains(problems, want) {
			t.Errorf("Validate = %q, want it to name %q", problems, want)
		}
	}
}

// planIDs names the operations a state runs.
func planIDs(s configure.State) []string {
	ops := configure.Plan(s)
	ids := make([]string, 0, len(ops))
	for _, op := range ops {
		ids = append(ids, op.ID)
	}
	return ids
}

// fakeRunner records the operations it ran and can fail one of them.
type fakeRunner struct {
	calls        []call
	failAt       string
	models       map[string]string
	clientModels map[string]string
	verify       string
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

func (f *fakeRunner) Install(project bool, clients, skills string, models map[string]string) error {
	f.clientModels = models
	f.calls = append(f.calls, call{op: "install", project: project, value: clients + "/" + skills})
	if f.failAt == "install" {
		return errFailed
	}
	return nil
}

func (f *fakeRunner) Voice(profile string) error { return f.record("voice", profile) }

func (f *fakeRunner) GlobalVoice(profile string) error { return f.record("globalvoice", profile) }

func (f *fakeRunner) Review(mode string) error { return f.record("review", mode) }

func (f *fakeRunner) Models(models map[string]string) error {
	f.models = models
	return f.record("models", "")
}

func (f *fakeRunner) Verify(model string) error {
	f.verify = model
	return f.record("verify", model)
}

func (f *fakeRunner) Field(field, value string) error {
	return f.record("field", field+"="+value)
}

func (f *fakeRunner) Axes() ([]profile.Axis, error) { return nil, nil }

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
