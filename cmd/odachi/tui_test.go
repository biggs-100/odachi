package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// seedTUIProject writes the four facts the first screen reads: the block, the stored
// choice, the voice and the review file.
func seedTUIProject(t *testing.T, dir string) {
	t.Helper()
	initRepo(t, dir)

	block, err := rules.WithBlock("", persona.Novato)
	if err != nil {
		t.Fatalf("rules.WithBlock: %v", err)
	}
	writeFile(t, dir, "AGENTS.md", block)

	store, err := agentstoml.Render(persona.Store{Agents: persona.SelectionOpenCode, Skills: persona.SkillsProject})
	if err != nil {
		t.Fatalf("agentstoml.Render: %v", err)
	}
	writeFile(t, dir, ".git/odachi/agents.toml", string(store))

	writeFile(t, dir, filepath.FromSlash("odd/profile.toml"), "name = \"senior\"\n")

	config, err := reviewtoml.Render(review.ProjectConfig{
		Mode:   review.ModeOn,
		Lenses: map[string]string{"reliability": "opencode-go/deepseek-v4.1-flash"},
	})
	if err != nil {
		t.Fatalf("reviewtoml.Render: %v", err)
	}
	writeFile(t, dir, filepath.FromSlash(review.ModePath), string(config))
}

// keys the wizard answers, in the form the program delivers them.
var (
	enterKey     = tea.KeyMsg{Type: tea.KeyEnter}
	downKey      = tea.KeyMsg{Type: tea.KeyDown}
	upKey        = tea.KeyMsg{Type: tea.KeyUp}
	escKey       = tea.KeyMsg{Type: tea.KeyEsc}
	qKey         = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	tabKey       = tea.KeyMsg{Type: tea.KeyTab}
	backspaceKey = tea.KeyMsg{Type: tea.KeyBackspace}
)

// toReview walks the wizard from the first screen to the review screen.
func toReview(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	for i := 0; i < 4; i++ {
		m, _ = press(t, m, enterKey)
	}
	if m.step != stepReview {
		t.Fatalf("step = %d, want review", m.step)
	}
	return m
}

// wizard is a state with every choice made.
func wizard() tuiState {
	return tuiState{
		Version: "9.9.9",
		Project: true,
		Clients: persona.SelectionBoth,
		Skills:  persona.SkillsHome,
		Voice:   "novato",
		Review:  review.ModeOn,
	}
}

// press sends one key to the model and returns the model it became.
func press(t *testing.T, m tuiModel, key tea.KeyMsg) (tuiModel, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(key)
	got, ok := next.(tuiModel)
	if !ok {
		t.Fatalf("Update returned %T, want tuiModel", next)
	}
	return got, cmd
}

// TestTUIEnterWalksTheSteps: the wizard is linear, and enter commits each choice.
// The review screen edits in place, so the summary is reached with tab.
func TestTUIEnterWalksTheSteps(t *testing.T) {
	m := newTUIModel(wizard())
	for _, want := range []tuiStep{stepClients, stepSkills, stepVoice, stepReview} {
		m, _ = press(t, m, enterKey)
		if m.step != want {
			t.Fatalf("step = %d, want %d", m.step, want)
		}
	}
	m, _ = press(t, m, tabKey)
	if m.step != stepSummary {
		t.Fatalf("step = %d, want summary", m.step)
	}
}

// TestTUINonProjectOffersSetupOrInit: a folder without the block asks what to do
// before it writes anything; setup is the default and init is chosen explicitly.
func TestTUINonProjectOffersSetupOrInit(t *testing.T) {
	in := wizard()
	in.Project = false
	m := newTUIModel(in)

	m, _ = press(t, m, enterKey) // state -> scope
	if m.step != stepScope {
		t.Fatalf("step = %d, want the scope screen", m.step)
	}
	view := m.View()
	for _, want := range []string{"configurar la máquina (setup)", "inicializar el proyecto (init)"} {
		if !strings.Contains(view, want) {
			t.Errorf("the scope screen does not offer %q:\n%s", want, view)
		}
	}

	m, _ = press(t, m, enterKey) // default: the machine
	if m.init {
		t.Fatal("init = true, want the machine default")
	}

	m, _ = press(t, m, escKey) // clients -> scope
	if m.step != stepScope {
		t.Fatalf("step = %d, want the scope screen", m.step)
	}
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, enterKey)
	if !m.init {
		t.Fatal("init = false, want the project chosen")
	}
}

// TestTUIMachinePathSkipsTheSkillsScreen: the machine has one scope, so the
// screen that chooses one would be a question with a single answer.
func TestTUIMachinePathSkipsTheSkillsScreen(t *testing.T) {
	in := wizard()
	in.Project = false
	m := newTUIModel(in)

	m, _ = press(t, m, enterKey) // state -> scope (machine default)
	m, _ = press(t, m, enterKey) // scope -> clients
	m, _ = press(t, m, enterKey) // clients -> voice
	if m.step != stepVoice {
		t.Fatalf("step = %d, want voice: the skills screen is a project question", m.step)
	}
}

// TestTUIClientScreenOffersTheThreeChoices: both, opencode and claude are the
// values the store knows, and the cursor picks one.
func TestTUIClientScreenOffersTheThreeChoices(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // state -> clients

	view := m.View()
	for _, want := range []string{"ambos (opencode y claude)", "opencode", "claude"} {
		if !strings.Contains(view, want) {
			t.Errorf("the client screen does not offer %q:\n%s", want, view)
		}
	}

	m, _ = press(t, m, downKey) // both -> opencode
	m, _ = press(t, m, enterKey)
	if m.state.Clients != persona.SelectionOpenCode {
		t.Fatalf("Clients = %q, want opencode", m.state.Clients)
	}
	if m.step != stepSkills {
		t.Fatalf("step = %d, want skills", m.step)
	}
}

// TestTUICursorStaysInsideItsList: holding down at the end does not fall off.
func TestTUICursorStaysInsideItsList(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	if m.client != 0 {
		t.Fatalf("client = %d, want 0", m.client)
	}
	m, _ = press(t, m, upKey)
	if m.client != 0 {
		t.Fatalf("client after up = %d, want 0", m.client)
	}
	for i := 0; i < 5; i++ {
		m, _ = press(t, m, downKey)
	}
	if m.client != len(tuiClients)-1 {
		t.Fatalf("client = %d, want the last option", m.client)
	}
}

// TestTUISkillsScreenChoosesTheScope: home and project are the two scopes.
func TestTUISkillsScreenChoosesTheScope(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	m, _ = press(t, m, enterKey) // -> skills
	m, _ = press(t, m, downKey)  // home -> project
	m, _ = press(t, m, enterKey)

	if m.state.Skills != persona.SkillsProject {
		t.Fatalf("Skills = %q, want project", m.state.Skills)
	}
	if m.step != stepVoice {
		t.Fatalf("step = %d, want voice", m.step)
	}
}

// TestTUIVoiceScreenChoosesTheBuiltIn: the wizard offers the three voices odachi
// ships, and the summary shows the one chosen.
func TestTUIVoiceScreenChoosesTheBuiltIn(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	m, _ = press(t, m, enterKey) // skills
	m, _ = press(t, m, enterKey) // -> voice
	m, _ = press(t, m, upKey)    // novato -> ci, the first of the sorted table
	m, _ = press(t, m, enterKey)

	if m.state.Voice != "ci" {
		t.Fatalf("Voice = %q, want ci", m.state.Voice)
	}
	if m.step != stepReview {
		t.Fatalf("step = %d, want review", m.step)
	}
	m, _ = press(t, m, tabKey) // review -> summary
	if !strings.Contains(m.View(), "ci") {
		t.Errorf("the summary does not show the chosen voice:\n%s", m.View())
	}
}

// TestTUIEscapeGoesBackOneScreen: back is one screen, and the first screen has no
// back.
func TestTUIEscapeGoesBackOneScreen(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	m, _ = press(t, m, enterKey) // skills
	m, _ = press(t, m, enterKey) // voice

	m, _ = press(t, m, escKey)
	if m.step != stepSkills {
		t.Fatalf("step = %d, want skills", m.step)
	}
	m, _ = press(t, m, escKey)
	if m.step != stepClients {
		t.Fatalf("step = %d, want clients", m.step)
	}
	m, _ = press(t, m, escKey)
	if m.step != stepState {
		t.Fatalf("step = %d, want the first screen", m.step)
	}
	m, _ = press(t, m, escKey)
	if m.step != stepState {
		t.Fatalf("step = %d, want the first screen: there is no earlier one", m.step)
	}
}

// TestTUIQuitReturnsTheQuitCommand: q asks the program to close from any screen.
func TestTUIQuitReturnsTheQuitCommand(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	_, cmd := press(t, m, qKey)
	if cmd == nil {
		t.Fatal("cmd = nil, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("cmd() = %T, want tea.QuitMsg", cmd())
	}
}

// TestTUITheChoicesAreShownInTheSummary: the last screen is the decision a human
// reviews before applying it.
func TestTUITheChoicesAreShownInTheSummary(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	m, _ = press(t, m, downKey)  // both -> opencode
	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, downKey) // home -> project
	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, downKey)  // novato -> senior
	m, _ = press(t, m, enterKey) // -> review
	m, _ = press(t, m, tabKey)   // review -> summary

	view := m.View()
	for _, want := range []string{"opencode", "project", "senior"} {
		if !strings.Contains(view, want) {
			t.Errorf("the summary does not show %q:\n%s", want, view)
		}
	}
}

// toSummary walks the wizard from the first screen to the summary.
func toSummary(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	m = toReview(t, m)
	m, _ = press(t, m, tabKey)
	if m.step != stepSummary {
		t.Fatalf("step = %d, want summary", m.step)
	}
	return m
}

// fakeRun records the operations the apply runs.
type fakeRun struct {
	calls        []string
	failAt       string
	clients      string
	skills       string
	project      bool
	voice        string
	review       string
	models       map[string]string
	clientModels map[string]string
	verify       string
	fields       []string
	axes         []profile.Axis
	reRead       int
}

func (f *fakeRun) record(op string) error {
	f.calls = append(f.calls, op)
	if f.failAt == op {
		return errTUI
	}
	return nil
}

func (f *fakeRun) Install(project bool, clients, skills string, models map[string]string) error {
	f.project, f.clients, f.skills, f.clientModels = project, clients, skills, models
	return f.record("install")
}

func (f *fakeRun) Voice(profile string) error {
	f.voice = profile
	return f.record("voice")
}

func (f *fakeRun) Review(mode string) error {
	f.review = mode
	return f.record("review")
}

func (f *fakeRun) Models(models map[string]string) error {
	f.models = models
	return f.record("models")
}

func (f *fakeRun) Verify(model string) error {
	f.verify = model
	return f.record("verify")
}

func (f *fakeRun) Field(field, value string) error {
	f.fields = append(f.fields, field+"="+value)
	return f.record("field")
}

func (f *fakeRun) Axes() ([]profile.Axis, error) {
	f.reRead++
	return f.axes, nil
}

// errTUI is the failure the fake operations return.
var errTUI = tuiError("the operation failed")

type tuiError string

func (e tuiError) Error() string { return string(e) }

// TestTUISummaryAppliesTheChoices: enter on the summary runs the plan with the
// state the screens collected, and every operation is reported.
func TestTUISummaryAppliesTheChoices(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"reliability": "opencode-go/x"}
	m := toSummary(t, newTUIModel(in))
	run := &fakeRun{}
	m.run = run

	m, _ = press(t, m, enterKey)

	if m.step != stepApply || !m.applied {
		t.Fatalf("step = %d applied = %v, want the apply screen", m.step, m.applied)
	}
	if got := strings.Join(run.calls, ","); got != "install,voice,review,models" {
		t.Fatalf("calls = %q, want install,voice,review,models", got)
	}
	if run.clients != persona.SelectionBoth || run.skills != persona.SkillsHome || !run.project {
		t.Errorf("install = %q/%q project=%v, want both/home/true", run.clients, run.skills, run.project)
	}
	if run.voice != "novato" || run.review != review.ModeOn {
		t.Errorf("voice = %q review = %q", run.voice, run.review)
	}
	if run.models["reliability"] != "opencode-go/x" {
		t.Errorf("models = %v, want the lens table", run.models)
	}

	view := m.View()
	for _, want := range []string{
		"instalar y configurar los clientes", "aplicar la voz", "configurar el review",
		"fijar el modelo de cada lente", "la configuración quedó aplicada",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("the apply screen does not carry %q:\n%s", want, view)
		}
	}
}

// TestTUISummaryStopsOnFailure: a failed operation is named and the rest stays
// pending, on the screen and in the runner.
func TestTUISummaryStopsOnFailure(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"reliability": "opencode-go/x"}
	m := toSummary(t, newTUIModel(in))
	run := &fakeRun{failAt: "voice"}
	m.run = run

	m, _ = press(t, m, enterKey)

	if got := strings.Join(run.calls, ","); got != "install,voice" {
		t.Fatalf("calls = %q, want install,voice", got)
	}
	view := m.View()
	for _, want := range []string{"✗ aplicar la voz", "configurar el review", "fijar el modelo de cada lente"} {
		if !strings.Contains(view, want) {
			t.Errorf("the apply screen does not carry %q:\n%s", want, view)
		}
	}
}

// TestTUICancelBeforeApplyLeavesTheRunnerAlone: q on the summary is the way out
// that writes nothing.
func TestTUICancelBeforeApplyLeavesTheRunnerAlone(t *testing.T) {
	m := toSummary(t, newTUIModel(wizard()))
	run := &fakeRun{}
	m.run = run

	_, cmd := press(t, m, qKey)

	if len(run.calls) != 0 {
		t.Fatalf("calls = %q, want none", strings.Join(run.calls, ","))
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want tea.Quit")
	}
}

// TestTUIRunnerInstallChoosesInitOrSetup: a project is initialized, a folder that
// is not one configures the machine only.
func TestTUIRunnerInstallChoosesInitOrSetup(t *testing.T) {
	var initArgs, setupArgs []string
	r := newTUIRunner("proyecto")
	r.initFn = func(_ string, args []string, _, _ io.Writer) int {
		initArgs = args
		return 0
	}
	r.setupFn = func(args []string, _, _ io.Writer) int {
		setupArgs = args
		return 0
	}

	if err := r.Install(true, persona.SelectionOpenCode, persona.SkillsProject, nil); err != nil {
		t.Fatalf("Install(project): %v", err)
	}
	if got := strings.Join(initArgs, " "); got != "--agents opencode --skills project" {
		t.Fatalf("init args = %q", got)
	}

	if err := r.Install(false, persona.SelectionBoth, persona.SkillsHome, nil); err != nil {
		t.Fatalf("Install(machine): %v", err)
	}
	if got := strings.Join(setupArgs, " "); got != "--agents both" {
		t.Fatalf("setup args = %q", got)
	}
}

// TestTUIRunnerReportsTheVerbError: the verb's own first line is the failure the
// apply screen shows.
func TestTUIRunnerReportsTheVerbError(t *testing.T) {
	r := newTUIRunner("proyecto")
	r.initFn = func(_ string, _ []string, _, stderr io.Writer) int {
		_, _ = fmt.Fprintln(stderr, "odachi: the store is broken")
		return 1
	}

	err := r.Install(true, persona.SelectionBoth, persona.SkillsHome, nil)
	if err == nil || !strings.Contains(err.Error(), "the store is broken") {
		t.Fatalf("err = %v, want the verb's line", err)
	}
}

// TestTUIRunnerModelsWritesTheFile: the lens table reaches the review file
// through the same writer `review mode` uses.
func TestTUIRunnerModelsWritesTheFile(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	r := newTUIRunner(dir)

	if err := r.Models(map[string]string{"reliability": "opencode-go/x"}); err != nil {
		t.Fatalf("Models: %v", err)
	}
	stored, err := os.ReadFile(filepath.Join(dir, "odd", "review.toml"))
	if err != nil {
		t.Fatalf("read odd/review.toml: %v", err)
	}
	if !strings.Contains(string(stored), `reliability = "opencode-go/x"`) {
		t.Errorf("the file does not carry the model:\n%s", stored)
	}
}

// TestTUIRunnerFieldWritesTheProfile: the editor's apply row reaches level.Set over
// a real project, and Axes reads the same file back: the two halves of the editor
// are pinned together instead of only against a fake (0336).
func TestTUIRunnerFieldWritesTheProfile(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	r := newTUIRunner(dir)

	if err := r.Field("output", "detail"); err != nil {
		t.Fatalf("Field: %v", err)
	}
	stored, err := os.ReadFile(filepath.Join(dir, "odd", "profile.toml"))
	if err != nil {
		t.Fatalf("read odd/profile.toml: %v", err)
	}
	if !strings.Contains(strings.Join(strings.Fields(string(stored)), " "), `output = "detail"`) {
		t.Errorf("the file does not carry the value:\n%s", stored)
	}

	axes, err := r.Axes()
	if err != nil {
		t.Fatalf("Axes: %v", err)
	}
	for _, axis := range axes {
		if axis.Name == "output" && axis.Value != "detail" {
			t.Errorf("Axes reads output = %q, want the value just written", axis.Value)
		}
	}
}

// TestTUIReviewScreenTogglesTheMode: the switch is one row, and enter flips it.
func TestTUIReviewScreenTogglesTheMode(t *testing.T) {
	m := toReview(t, newTUIModel(wizard()))

	m, _ = press(t, m, enterKey)
	if m.state.Review != review.ModeOff {
		t.Fatalf("Review = %q, want off", m.state.Review)
	}
	m, _ = press(t, m, enterKey)
	if m.state.Review != review.ModeOn {
		t.Fatalf("Review = %q, want on", m.state.Review)
	}
}

// TestTUIReviewScreenEditsALensModel: a lens row opens a line editor, and the
// identifier is saved on enter.
func TestTUIReviewScreenEditsALensModel(t *testing.T) {
	m := toReview(t, newTUIModel(wizard()))
	m, _ = press(t, m, downKey) // modo -> risk
	m, _ = press(t, m, enterKey)

	if !m.editing {
		t.Fatal("editing = false, want the line editor open")
	}
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("opencode-go/x")})
	m, _ = press(t, m, enterKey)

	if m.editing {
		t.Fatal("editing = true, want the editor closed")
	}
	if got := m.state.Models["risk"]; got != "opencode-go/x" {
		t.Fatalf("risk = %q, want opencode-go/x", got)
	}
}

// TestTUIReviewScreenRefusesAModelThatIsNotAnIdentifier: the editor stays open
// and the table is untouched.
func TestTUIReviewScreenRefusesAModelThatIsNotAnIdentifier(t *testing.T) {
	m := toReview(t, newTUIModel(wizard()))
	m, _ = press(t, m, downKey) // modo -> risk
	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a b")})
	m, _ = press(t, m, enterKey)

	if !m.editing {
		t.Fatal("editing = false, want the editor still open")
	}
	if m.editErr == "" {
		t.Fatal("editErr is empty, want the refusal on the screen")
	}
	if _, ok := m.state.Models["risk"]; ok {
		t.Fatalf("risk was saved: %q", m.state.Models["risk"])
	}
}

// TestTUIReviewScreenRemovesAModelWhenCleared: an empty value means the lens has
// no model of its own, and the line goes away.
func TestTUIReviewScreenRemovesAModelWhenCleared(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"reliability": "opencode-go/x"}
	m := toReview(t, newTUIModel(in))

	for i := 0; i < 3; i++ { // modo -> risk -> readability -> reliability
		m, _ = press(t, m, downKey)
	}
	m, _ = press(t, m, enterKey)
	for i := 0; i < len("opencode-go/x"); i++ {
		m, _ = press(t, m, backspaceKey)
	}
	m, _ = press(t, m, enterKey)

	if _, ok := m.state.Models["reliability"]; ok {
		t.Fatalf("reliability = %q, want the line removed", m.state.Models["reliability"])
	}
}

// TestTUIReviewScreenTabReachesTheSummary: tab leaves the review screen, and the
// summary shows the model that will be written.
func TestTUIReviewScreenTabReachesTheSummary(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"reliability": "opencode-go/x"}
	m := toReview(t, newTUIModel(in))
	m, _ = press(t, m, tabKey)

	if m.step != stepSummary {
		t.Fatalf("step = %d, want summary", m.step)
	}
	if !strings.Contains(m.View(), "reliability=opencode-go/x") {
		t.Errorf("the summary does not show the lens model:\n%s", m.View())
	}
}

// TestTUIReviewScreenEscapesToVoice: back is one screen, also from the review.
func TestTUIReviewScreenEscapesToVoice(t *testing.T) {
	m := toReview(t, newTUIModel(wizard()))
	m, _ = press(t, m, escKey)
	if m.step != stepVoice {
		t.Fatalf("step = %d, want voice", m.step)
	}
}

// TestTUIDetectsTheState: the first screen starts from what the project already
// chose, not from the wizard's defaults. The models section reads the verifier's
// key and each client's own file (0337).
func TestTUIDetectsTheState(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)
	home := withTempHome(t)
	writeFile(t, home, filepath.FromSlash(".config/opencode/opencode.json"), `{"model": "opencode-go/x"}`)
	writeFile(t, home, filepath.FromSlash(".claude/settings.json"), `{"model": "sonnet"}`)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}

	want := tuiState{
		Version:      version,
		Project:      true,
		Clients:      persona.SelectionOpenCode,
		Skills:       persona.SkillsProject,
		Voice:        "senior",
		Review:       review.ModeOn,
		Models:       map[string]string{"reliability": "opencode-go/deepseek-v4.1-flash"},
		ClientModels: map[string]string{"opencode": "opencode-go/x", "claude": "sonnet"},
		Axes:         profile.Axes(profile.Builtin()["senior"]),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("state = %+v, want %+v", got, want)
	}
}

// TestTUIDetectsAFreshFolder: a folder odachi never touched reads as not a
// project, with the machine's defaults and the review off.
func TestTUIDetectsAFreshFolder(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	withTempHome(t)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}

	if got.Project {
		t.Error("Project = true, want false: there is no block")
	}
	if got.Clients != persona.SelectionBoth {
		t.Errorf("Clients = %q, want both", got.Clients)
	}
	if got.Skills != persona.SkillsHome {
		t.Errorf("Skills = %q, want home", got.Skills)
	}
	if got.Voice != "novato" {
		t.Errorf("Voice = %q, want the inferred novato", got.Voice)
	}
	if got.Review != review.ModeOff {
		t.Errorf("Review = %q, want off", got.Review)
	}
}

// TestTUIFirstScreenNamesTheState: the first thing a human reads is where they
// are and what odachi will configure.
func TestTUIFirstScreenNamesTheState(t *testing.T) {
	view := newTUIModel(tuiState{
		Version: "9.9.9",
		Project: true,
		Clients: persona.SelectionBoth,
		Skills:  persona.SkillsHome,
		Voice:   "novato",
		Review:  review.ModeOn,
	}).View()

	for _, want := range []string{
		"odachi 9.9.9", "instalación y configuración",
		"Proyecto", "sí", "Clientes", "both", "Skills", "home", "Voz", "novato", "Review", "on",
		"q salir",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("the first screen does not carry %q:\n%s", want, view)
		}
	}
}

// TestTUIDetectionWritesNothing: opening the wizard is a read. This tree is the
// one the human will configure, and looking at it must not move it.
func TestTUIDetectionWritesNothing(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)
	withTempHome(t)
	before := manifest(t, dir)

	if _, err := detectTUIState(dir); err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}
	newTUIModel(tuiState{Version: version}).View()

	if after := manifest(t, dir); after != before {
		t.Errorf("the tree moved while detecting:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestTUIRefusesWithoutATerminal: a pipe is not a screen, and the refusal names
// the command that configures without one.
func TestTUIRefusesWithoutATerminal(t *testing.T) {
	code, out, errOut := runDir(t, t.TempDir(), []string{"tui"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want blocked (stdout %q)", out)
	}
	message := out + errOut
	for _, want := range []string{"terminal", "odachi setup"} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, message)
		}
	}
}

// TestTUIWithAFlagIsRefused: the wizard has no flags; the flags belong to the
// verbs it applies.
func TestTUIWithAFlagIsRefused(t *testing.T) {
	code, _, errOut := runDir(t, t.TempDir(), []string{"tui", "--json"}, "")

	if code == 0 {
		t.Fatal("code = 0, want blocked")
	}
	if !strings.Contains(errOut, "takes no flags") {
		t.Errorf("the refusal does not say the verb takes no flags:\n%s", errOut)
	}
}

// TestTUITheMenuOffersEachSectionAndEscReturnsToIt: the wizard opens on a menu,
// enter walks the section the cursor names, and esc comes back to the menu from
// any of them, so a person can change one thing and leave (0336).
func TestTUITheMenuOffersEachSectionAndEscReturnsToIt(t *testing.T) {
	m := newTUIModel(tuiState{Project: true, Clients: "both", Skills: "home", Voice: "novato"})

	if m.step != stepState {
		t.Fatalf("the wizard opened on step %d, want the menu", m.step)
	}
	view := m.View()
	for _, want := range []string{"instalar y configurar", "perfil"} {
		if !strings.Contains(view, want) {
			t.Errorf("the menu does not offer %q:\n%s", want, view)
		}
	}

	install, _ := m.next()
	if got := install.(tuiModel).step; got != stepClients {
		t.Errorf("install landed on step %d, want the first choice of the section", got)
	}

	m.menu = 1
	opened, _ := m.next()
	editor := opened.(tuiModel)
	if editor.step != stepProfile {
		t.Fatalf("perfil landed on step %d, want the profile editor", editor.step)
	}
	profileView := editor.View()
	for _, want := range []string{"brief_source", "output", "explain"} {
		if !strings.Contains(profileView, want) {
			t.Errorf("the profile screen does not show %q:\n%s", want, profileView)
		}
	}
	if back := editor.back(); back.step != stepState {
		t.Errorf("esc from the profile landed on step %d, want the menu", back.step)
	}
}

// advance presses enter through next and gives the model back, which next returns
// as the tea.Model interface.
func advance(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	next, _ := m.next()
	model, ok := next.(tuiModel)
	if !ok {
		t.Fatalf("next returned %T, want tuiModel", next)
	}
	return model
}

// TestTheProfileEditorWritesOnlyWhenApplied: a value chosen in the editor stays in
// the model, esc leaves the section with the file untouched, and only the apply row
// writes -one field at a time (0336).
func TestTheProfileEditorWritesOnlyWhenApplied(t *testing.T) {
	run := &fakeRun{axes: profile.Axes(profile.Builtin()["novato"])}
	m := newTUIModel(tuiState{Project: true, Clients: "both", Skills: "home", Voice: "novato"})
	m.run = run
	m.menu = 1
	m = advance(t, m)
	if m.step != stepProfile {
		t.Fatalf("step = %d, want the profile editor", m.step)
	}

	// output is the fourth axis, and detail is its second value.
	m.profileCursor = 3
	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, enterKey)
	if len(run.calls) != 0 {
		t.Fatalf("choosing a value wrote %v", run.calls)
	}
	if got := m.value(m.state.Axes[3]); got != "detail" {
		t.Fatalf("the row shows %q, want detail", got)
	}

	m, _ = press(t, m, escKey)
	if m.step != stepState {
		t.Fatalf("esc landed on step %d, want the menu", m.step)
	}
	if len(run.calls) != 0 {
		t.Fatalf("leaving the section wrote %v", run.calls)
	}

	m.menu = 1
	m = advance(t, m)
	m.profileCursor = len(m.state.Axes)
	m, _ = press(t, m, enterKey)
	if want := []string{"field"}; !slices.Equal(run.calls, want) {
		t.Fatalf("calls = %v, want %v", run.calls, want)
	}
	if want := []string{"output=detail"}; !slices.Equal(run.fields, want) {
		t.Errorf("fields = %v, want %v", run.fields, want)
	}
	if !strings.Contains(m.fieldMsg, "output=detail") {
		t.Errorf("the apply line does not say what it wrote: %q", m.fieldMsg)
	}
}

// TestTheEditorReadsTheProfileAgainOnEntry: the hub applies the voice in one
// section and edits the profile in another, so the rows describe the file that
// stands when the section opens, not the one the wizard read at the start (0336).
func TestTheEditorReadsTheProfileAgainOnEntry(t *testing.T) {
	axes := profile.Axes(profile.Builtin()["novato"])
	for i := range axes {
		if axes[i].Name == "output" {
			axes[i].Value = "json"
		}
	}
	run := &fakeRun{axes: axes}
	m := newTUIModel(tuiState{Project: true, Clients: "both", Skills: "home", Voice: "novato"})
	m.run = run
	m.menu = 1

	m = advance(t, m)

	if run.reRead != 1 {
		t.Errorf("the section read the profile %d times, want one", run.reRead)
	}
	if got := m.value(m.state.Axes[3]); got != "json" {
		t.Errorf("output reads %q, want the value the file carries now", got)
	}
}

// TestTheExplicitEditsSurviveTheVoice: the voice op recomputes every field from the
// builtin, whatever order the sections ran in, so an axis written in this run is
// re-applied after it: the last word is the human's edit (0336).
func TestTheExplicitEditsSurviveTheVoice(t *testing.T) {
	run := &fakeRun{axes: profile.Axes(profile.Builtin()["novato"])}
	m := newTUIModel(tuiState{Project: true, Clients: "both", Skills: "home", Voice: "novato", Review: review.ModeOff})
	m.run = run
	m.menu = 1
	m = advance(t, m)

	// The human edits one axis and applies it in the editor.
	m.profileCursor = 3
	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, enterKey)
	m.profileCursor = len(m.state.Axes)
	m, _ = press(t, m, enterKey)

	// Then a voice is chosen and the install section applies: the voice rewrites
	// the profile from the builtin, and the edit has to survive it.
	run.calls = nil
	run.fields = nil
	m.state.Voice = "senior"
	m.step = stepSummary
	m, _ = press(t, m, enterKey)

	voice := slices.Index(run.calls, "voice")
	if voice < 0 {
		t.Fatalf("the voice was not applied: %v", run.calls)
	}
	last := len(run.calls) - 1
	if run.calls[last] != "field" || last < voice {
		t.Fatalf("the edit was not re-applied after the voice: %v", run.calls)
	}
	if want := []string{"output=detail"}; !slices.Equal(run.fields, want) {
		t.Errorf("fields = %v, want %v", run.fields, want)
	}
}

// withTempHome points the home at a folder the test owns: the models section
// reads the client configs there, and no test touches the real home (0337).
func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	previous := homeDir
	homeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { homeDir = previous })
	return home
}

// toModels opens the models section from the hub.
func toModels(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	m.menu = 2
	m = advance(t, m)
	if m.step != stepModels {
		t.Fatalf("step = %d, want the models section", m.step)
	}
	return m
}

// typeModel types a whole value into the open editor.
func typeModel(t *testing.T, m tuiModel, value string) tuiModel {
	t.Helper()
	m, _ = press(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)})
	return m
}

// TestTUITheMenuOffersTheModelsSection: the hub lists the third section, enter
// opens it, esc comes back to the hub and q still quits (0337).
func TestTUITheMenuOffersTheModelsSection(t *testing.T) {
	m := newTUIModel(tuiState{Project: true, Clients: persona.SelectionBoth, Skills: persona.SkillsHome, Voice: "novato"})

	if view := m.View(); !strings.Contains(view, "modelos") {
		t.Fatalf("the hub does not offer the models section:\n%s", view)
	}

	opened := toModels(t, m)
	if view := opened.View(); !strings.Contains(view, "verificador") {
		t.Errorf("the section does not show the verifier:\n%s", view)
	}

	if back := opened.back(); back.step != stepState {
		t.Errorf("esc landed on step %d, want the hub", back.step)
	}

	if _, cmd := press(t, opened, qKey); cmd == nil {
		t.Error("q does not quit from the models section")
	}
}

// TestTheModelsSectionShowsTheVerifierAndTheInstalledClients: the rows are the
// verifier and the clients the hub detected, with the model each file carries
// today and a word for the ones that carry none (0337).
func TestTheModelsSectionShowsTheVerifierAndTheInstalledClients(t *testing.T) {
	both := tuiState{
		Project: true, Clients: persona.SelectionBoth, Skills: persona.SkillsHome, Voice: "novato",
		Verify:       "cheap/model",
		ClientModels: map[string]string{"opencode": "opencode-go/x"},
	}
	view := toModels(t, newTUIModel(both)).View()
	for _, want := range []string{"verificador", "cheap/model", "OpenCode", "opencode-go/x", "Claude", "(ninguno)"} {
		if !strings.Contains(view, want) {
			t.Errorf("the section does not carry %q:\n%s", want, view)
		}
	}

	one := both
	one.Clients = persona.SelectionClaude
	view = toModels(t, newTUIModel(one)).View()
	if strings.Contains(view, "OpenCode") {
		t.Errorf("a machine that runs Claude alone was asked about OpenCode:\n%s", view)
	}
	if !strings.Contains(view, "Claude") || !strings.Contains(view, "(ninguno)") {
		t.Errorf("the Claude row is missing:\n%s", view)
	}
}

// TestTheModelsSectionEditsTheVerifier: enter opens the buffer, enter keeps the
// value in the model, and esc drops what was typed without dropping the choice
// that was already there (0337).
func TestTheModelsSectionEditsTheVerifier(t *testing.T) {
	m := toModels(t, newTUIModel(wizard()))
	m, _ = press(t, m, enterKey) // the verifier row
	if !m.editing {
		t.Fatal("editing = false, want the line editor open")
	}

	m = typeModel(t, m, "cheap/model")
	m, _ = press(t, m, enterKey)
	if m.editing {
		t.Fatal("editing = true, want the editor closed")
	}
	if got := m.modelPending[modelVerifier]; got != "cheap/model" {
		t.Fatalf("pending = %q, want the value kept", got)
	}
	if !strings.Contains(m.View(), "cheap/model") {
		t.Errorf("the row does not show what was kept:\n%s", m.View())
	}

	m, _ = press(t, m, enterKey)
	m = typeModel(t, m, "dropped")
	m, _ = press(t, m, escKey)
	if m.editing {
		t.Fatal("editing = true, want esc to close the editor")
	}
	if got := m.modelPending[modelVerifier]; got != "cheap/model" {
		t.Fatalf("pending = %q, want the value esc did not touch", got)
	}
	if strings.Contains(m.View(), "dropped") {
		t.Errorf("the row kept what esc dropped:\n%s", m.View())
	}
}

// TestTheModelsSectionRefusesAValueThatIsNotAnIdentifier: the section obeys the
// lens screen's rule and gives its sentence, because a second rule is how two
// screens start disagreeing (0337).
func TestTheModelsSectionRefusesAValueThatIsNotAnIdentifier(t *testing.T) {
	m := toModels(t, newTUIModel(wizard()))
	m, _ = press(t, m, enterKey)
	m = typeModel(t, m, "a b")
	m, _ = press(t, m, enterKey)

	if !m.editing {
		t.Fatal("editing = false, want the editor still open")
	}
	if m.editErr == "" {
		t.Fatal("editErr is empty, want the refusal on the screen")
	}
	if _, ok := m.modelPending[modelVerifier]; ok {
		t.Fatalf("the refused value was kept: %q", m.modelPending[modelVerifier])
	}

	lens := toReview(t, newTUIModel(wizard()))
	lens, _ = press(t, lens, downKey) // modo -> risk
	lens, _ = press(t, lens, enterKey)
	lens = typeModel(t, lens, "a b")
	lens, _ = press(t, lens, enterKey)
	if lens.editErr != m.editErr {
		t.Errorf("the lens screen says %q and the models section %q: one rule, one sentence", lens.editErr, m.editErr)
	}
}

// TestTheModelsSectionNamesNoModelWhenCleared: an empty value is the row going
// back to none, and the plan will remove the verifier's line (0337).
func TestTheModelsSectionNamesNoModelWhenCleared(t *testing.T) {
	in := wizard()
	in.Verify = "cheap/model"
	m := toModels(t, newTUIModel(in))
	m, _ = press(t, m, enterKey)
	for i := 0; i < len("cheap/model"); i++ {
		m, _ = press(t, m, backspaceKey)
	}
	m, _ = press(t, m, enterKey)

	value, ok := m.modelPending[modelVerifier]
	if !ok || value != "" {
		t.Fatalf("pending = %q (present %v), want an explicit empty choice", value, ok)
	}
	if !strings.Contains(m.View(), "(ninguno)") {
		t.Errorf("the row does not say there is no model:\n%s", m.View())
	}
}

// TestTheModelsSectionReachesTheInstallOperation: a client's model is applied by
// the verb that owns the client file, so the install receives --model instead of
// the wizard writing opencode.json itself (0337).
func TestTheModelsSectionReachesTheInstallOperation(t *testing.T) {
	m := toModels(t, newTUIModel(wizard()))
	run := &fakeRun{}
	m.run = run

	m, _ = press(t, m, downKey) // verifier -> opencode
	m, _ = press(t, m, enterKey)
	m = typeModel(t, m, "opencode-go/x")
	m, _ = press(t, m, enterKey)

	m.step = stepSummary
	m, _ = press(t, m, enterKey)

	if m.step != stepApply {
		t.Fatalf("step = %d, want the apply screen", m.step)
	}
	if got := run.clientModels["opencode"]; got != "opencode-go/x" {
		t.Fatalf("install models = %v, want opencode=opencode-go/x", run.clientModels)
	}
	if slices.Contains(run.calls, "verify") {
		t.Errorf("calls = %v, want no verifier write", run.calls)
	}
}

// TestAnUnchangedModelsSectionWritesNothing: a section the human never opened,
// and a row that was opened and kept as it was, add no operation and no --model
// (0337).
func TestAnUnchangedModelsSectionWritesNothing(t *testing.T) {
	never := toSummary(t, newTUIModel(wizard()))
	run := &fakeRun{}
	never.run = run
	never, _ = press(t, never, enterKey)
	if slices.Contains(run.calls, "verify") {
		t.Errorf("calls = %v, want no verifier write from a section nobody opened", run.calls)
	}
	if run.clientModels != nil {
		t.Errorf("install models = %v, want none", run.clientModels)
	}

	in := wizard()
	in.Verify = "cheap/model"
	in.ClientModels = map[string]string{"opencode": "opencode-go/x"}
	m := toModels(t, newTUIModel(in))
	run = &fakeRun{}
	m.run = run
	m, _ = press(t, m, enterKey) // open the verifier
	m, _ = press(t, m, enterKey) // keep the value it shows

	m.step = stepSummary
	m, _ = press(t, m, enterKey)

	if slices.Contains(run.calls, "verify") {
		t.Errorf("calls = %v, want no verifier write for the value already there", run.calls)
	}
	if run.clientModels != nil {
		t.Errorf("install models = %v, want none: the client row did not change", run.clientModels)
	}
}

// TestTheModelsSectionKeepsTheCursorInsideItsRows: the client choice lives in
// another section, and a row that is no longer shown cannot be the one enter
// edits (0337).
func TestTheModelsSectionKeepsTheCursorInsideItsRows(t *testing.T) {
	m := toModels(t, newTUIModel(wizard()))
	m, _ = press(t, m, downKey) // verifier -> opencode
	m, _ = press(t, m, downKey) // opencode -> claude
	m, _ = press(t, m, escKey)

	m.state.Clients = persona.SelectionOpenCode
	m = toModels(t, m)
	if view := m.View(); strings.Contains(view, "Claude") {
		t.Fatalf("the section still shows a client the hub no longer names:\n%s", view)
	}

	m, _ = press(t, m, enterKey)
	if !m.editing {
		t.Fatal("enter on the row that exists did not open the editor")
	}
	m = typeModel(t, m, "opencode-go/x")
	m, _ = press(t, m, enterKey)
	if got := m.modelPending[persona.SelectionOpenCode]; got != "opencode-go/x" {
		t.Fatalf("pending = %v, want the OpenCode row", m.modelPending)
	}
}

// TestTheVerifierRowReachesTheReviewWriter: a changed verifier is an operation of
// the plan, and it reaches the app that owns odd/review.toml (0337).
func TestTheVerifierRowReachesTheReviewWriter(t *testing.T) {
	m := toModels(t, newTUIModel(wizard()))
	run := &fakeRun{}
	m.run = run

	m, _ = press(t, m, enterKey) // the verifier row
	m = typeModel(t, m, "cheap/model#fast")
	m, _ = press(t, m, enterKey)

	m.step = stepSummary
	m, _ = press(t, m, enterKey)

	if !slices.Contains(run.calls, "verify") {
		t.Fatalf("calls = %v, want the verifier written", run.calls)
	}
	if run.verify != "cheap/model#fast" {
		t.Fatalf("verify = %q, want the chosen model", run.verify)
	}
}

// TestTheVerifierRowWritesTheReviewFile: the row reaches SetConfig over a real
// project, and the lens table the section does not touch survives (0337).
func TestTheVerifierRowWritesTheReviewFile(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)
	r := newTUIRunner(dir)

	if err := r.Verify("cheap/model#fast"); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	stored, err := os.ReadFile(filepath.Join(dir, "odd", "review.toml"))
	if err != nil {
		t.Fatalf("read odd/review.toml: %v", err)
	}
	for _, want := range []string{`verify = "cheap/model#fast"`, `reliability = "opencode-go/deepseek-v4.1-flash"`} {
		if !strings.Contains(string(stored), want) {
			t.Errorf("the write lost %q:\n%s", want, stored)
		}
	}
}

// TestTUIRunnerInstallCarriesTheClientModels: the client's model travels as the
// flag the verb already understands, in a stable order (0337).
func TestTUIRunnerInstallCarriesTheClientModels(t *testing.T) {
	var initArgs, setupArgs []string
	r := newTUIRunner("proyecto")
	r.initFn = func(_ string, args []string, _, _ io.Writer) int {
		initArgs = args
		return 0
	}
	r.setupFn = func(args []string, _, _ io.Writer) int {
		setupArgs = args
		return 0
	}
	models := map[string]string{"opencode": "opencode-go/x", "claude": "sonnet"}

	if err := r.Install(true, persona.SelectionBoth, persona.SkillsHome, models); err != nil {
		t.Fatalf("Install(project): %v", err)
	}
	if got := strings.Join(initArgs, " "); got != "--agents both --skills home --model claude=sonnet --model opencode=opencode-go/x" {
		t.Errorf("init args = %q", got)
	}

	if err := r.Install(false, persona.SelectionBoth, persona.SkillsHome, models); err != nil {
		t.Fatalf("Install(machine): %v", err)
	}
	if got := strings.Join(setupArgs, " "); got != "--agents both --model claude=sonnet --model opencode=opencode-go/x" {
		t.Errorf("setup args = %q", got)
	}
}

// TestTUIWalksTheRouterTable: every screen has a row, and walking forward then
// back returns to where it started. A screen without a row is a screen the
// keyboard can enter and not leave, and the conditional screens are walked past
// in both directions instead of being hardcoded in two switches (0336).
func TestTUIWalksTheRouterTable(t *testing.T) {
	for _, step := range tuiSteps {
		if _, ok := tuiRoutes[step]; !ok {
			t.Errorf("step %d has no row in the router table", step)
		}
	}

	m := newTUIModel(tuiState{Project: true, Clients: "both", Skills: "home", Voice: "novato"})
	var forward []tuiStep
	m.step = stepState
	for i := 0; i < len(tuiSteps); i++ {
		next := m.forwardStep()
		if next == m.step {
			break
		}
		m.step = next
		forward = append(forward, next)
	}
	if len(forward) == 0 || forward[len(forward)-1] != stepApply {
		t.Fatalf("forward = %v, want it to end on the apply screen", forward)
	}
	for _, skipped := range forward {
		if skipped == stepScope {
			t.Errorf("a project walked through the scope screen: %v", forward)
		}
	}

	forward = nil
	m = newTUIModel(tuiState{Project: false, Clients: "both", Skills: "home", Voice: "novato"})
	m.step = stepState
	for i := 0; i < len(tuiSteps); i++ {
		next := m.forwardStep()
		if next == m.step {
			break
		}
		m.step = next
		forward = append(forward, next)
	}
	if !slices.Contains(forward, stepScope) {
		t.Errorf("a folder that is not a project skipped the scope question: %v", forward)
	}
	if slices.Contains(forward, stepSkills) {
		t.Errorf("a machine-only setup asked which skills to install in the project: %v", forward)
	}

	m.step = stepApply
	for i := 0; i < len(tuiSteps); i++ {
		next := m.backStep()
		if next == m.step {
			break
		}
		m.step = next
	}
	if m.step != stepState {
		t.Errorf("walking back ended on step %d, want the first screen", m.step)
	}
}
