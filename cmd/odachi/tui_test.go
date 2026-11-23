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
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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
	jKey         = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	kKey         = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
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

// TestTUIVoiceScreenOffersOfficePeru: the built-in profile is a real choice in
// the wizard, not only a name in a file (0354).
func TestTUIVoiceScreenOffersOfficePeru(t *testing.T) {
	m := newTUIModel(wizard())
	m, _ = press(t, m, enterKey) // clients
	m, _ = press(t, m, enterKey) // skills
	m, _ = press(t, m, enterKey) // -> voice
	m, _ = press(t, m, downKey)  // novato -> office-peru
	m, _ = press(t, m, enterKey)
	if m.state.Voice != "office-peru" {
		t.Fatalf("Voice = %q, want office-peru", m.state.Voice)
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
	m, _ = press(t, m, downKey)  // novato -> office-peru
	m, _ = press(t, m, downKey)  // office-peru -> senior
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
	limits       limitChoice
	checks       []checkChoice
	fields       []string
	axes         []profile.Axis
	reRead       int
	// catalog is the answer the host gives the picker, and catalogCalls how many
	// times the wizard asked. The probe is the one thing here that would spawn a
	// process, so a test hands back a list instead (0338).
	catalog      []string
	catalogOK    bool
	catalogCalls int
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

func (f *fakeRun) Limits(choice limitChoice) error {
	f.limits = choice
	return f.record("limits")
}

func (f *fakeRun) Checks(edits []checkChoice) error {
	f.checks = edits
	return f.record("checks")
}

func (f *fakeRun) Field(field, value string) error {
	f.fields = append(f.fields, field+"="+value)
	return f.record("field")
}

func (f *fakeRun) Axes() ([]profile.Axis, error) {
	f.reRead++
	return f.axes, nil
}

// Catalog answers what the host serves, without a process.
func (f *fakeRun) Catalog() ([]string, bool) {
	f.catalogCalls++
	return f.catalog, f.catalogOK
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
// key and each client's own file; the limits section reads the four scalars from
// the same file, normalized the way the writer reads them (0337, 0342).
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
		KeepVerdicts: 10,
		CheckTimeout: 300,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("state = %+v, want %+v", got, want)
	}
}

// TestTUIDetectsAModelsRowItCannotRead: one unreadable client file is a row that
// says so, not a verb that refuses to open -the wizard has three sections and the
// other two have nothing to do with that file (0337, the review's warning).
func TestTUIDetectsAModelsRowItCannotRead(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)
	home := withTempHome(t)
	writeFile(t, home, filepath.FromSlash(".config/opencode/opencode.json"), `{"model": "opencode-go/x"}`)
	writeFile(t, home, filepath.FromSlash(".claude/settings.json"), `{"model": `)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v, want the broken file named instead of the verb refused", err)
	}
	if got.ClientModels["opencode"] != "opencode-go/x" {
		t.Errorf("opencode = %q, want the file that could be read", got.ClientModels["opencode"])
	}
	if !strings.Contains(got.ClientModelErrs["claude"], "settings.json") {
		t.Errorf("errs = %v, want the broken file named", got.ClientModelErrs)
	}

	// The row that cannot be read says so instead of claiming there is no model.
	got.Clients = persona.SelectionBoth
	m := toModels(t, newTUIModel(got))
	m, _ = press(t, m, downKey) // verifier -> opencode
	m, _ = press(t, m, downKey) // opencode -> claude
	if view := m.View(); !strings.Contains(view, "(ilegible)") {
		t.Errorf("the row does not say it could not read the file:\n%s", view)
	}
	if view := m.View(); !strings.Contains(view, "settings.json") {
		t.Errorf("the section does not name the file:\n%s", view)
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
		"Proyecto", "sí", "Clientes", "ambos (opencode y claude)", "Skills", "home", "Voz", "novato", "Review", "on",
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

// TestTheModelsSectionRefusesToClearAClientRow: no verb removes a client's key,
// so a row that showed itself empty while the file kept its model would be lying
// about what the run wrote. The verifier's row is the one that can be cleared,
// because its line is written by the review file's own writer (0337).
func TestTheModelsSectionRefusesToClearAClientRow(t *testing.T) {
	in := wizard()
	in.ClientModels = map[string]string{persona.SelectionOpenCode: "opencode-go/x"}
	m := toModels(t, newTUIModel(in))

	m, _ = press(t, m, downKey) // verifier -> opencode
	m, _ = press(t, m, enterKey)
	for i := 0; i < len("opencode-go/x"); i++ {
		m, _ = press(t, m, backspaceKey)
	}
	m, _ = press(t, m, enterKey)

	if !m.editing {
		t.Fatal("editing = false, want the editor still open")
	}
	if m.editErr == "" {
		t.Fatal("editErr is empty, want the refusal on the screen")
	}
	if _, ok := m.modelPending[persona.SelectionOpenCode]; ok {
		t.Fatalf("the refused empty value was kept: %q", m.modelPending[persona.SelectionOpenCode])
	}
	if !strings.Contains(m.View(), "borra la clave model") {
		t.Errorf("the refusal does not say where the line lives:\n%s", m.View())
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

// plainTUI makes the layer render without colour: a terminal that refuses
// escape codes still has to read, and the assertions below are about the words
// and the glyphs, not the bytes (0338).
func plainTUI(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

// colouredTUI forces the profile a modern terminal reports, so a test can prove
// the layer actually styles a row instead of being a no-op (0338).
func colouredTUI(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	dark := lipgloss.HasDarkBackground()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previous)
		lipgloss.SetHasDarkBackground(dark)
	})
}

// TestTUIEveryScreenDrawsTheCursorAsAGlyph: the hub, a choice screen and the
// models section mark the cursor with the layer's glyph instead of composing a
// raw "> " row, and exactly one row carries it (0338).
func TestTUIEveryScreenDrawsTheCursorAsAGlyph(t *testing.T) {
	plainTUI(t)

	hub := newTUIModel(wizard()).View()
	choice := newTUIModel(wizard())
	choice.step = stepClients
	models := toModels(t, newTUIModel(wizard())).View()

	for name, view := range map[string]string{"hub": hub, "choice": choice.View(), "models": models} {
		if !strings.Contains(view, cursorGlyph) {
			t.Errorf("the %s does not draw the cursor glyph %q:\n%s", name, cursorGlyph, view)
		}
		if strings.Contains(view, "> ") {
			t.Errorf("the %s still composes a raw \"> \" row:\n%s", name, view)
		}
		if got := strings.Count(view, cursorGlyph); got != 1 {
			t.Errorf("the %s draws the cursor %d times, want one:\n%s", name, got, view)
		}
	}
}

// TestTUISelectedRowIsStyledNotJustMarked: under colour the selected row is
// actually styled and the unselected one stays plain, so the layer cannot
// silently become a no-op; with colour refused both keep the label column (0338).
func TestTUISelectedRowIsStyledNotJustMarked(t *testing.T) {
	colouredTUI(t)
	selected := tuiLook.row(true, "opciones")
	plain := tuiLook.row(false, "opciones")

	if selected == plain {
		t.Fatalf("the selected row renders like an unselected one: %q", selected)
	}
	if !strings.Contains(selected, "\x1b[") {
		t.Errorf("the selected row carries no colour: %q", selected)
	}
	if strings.Contains(plain, "\x1b[") {
		t.Errorf("the unselected row is styled: %q", plain)
	}
	selectedCol := lipgloss.Width(selected[:strings.Index(selected, "opciones")])
	plainCol := lipgloss.Width(plain[:strings.Index(plain, "opciones")])
	if selectedCol != plainCol {
		t.Errorf("the label column moves: column %d vs %d", selectedCol, plainCol)
	}
}

// TestTUINoticesCarryTheirMarkers: a file that could not be read is a warning
// with a "!" and a refused value is a failure with a "×", each on its own line
// and outside the plain body voice (0338).
func TestTUINoticesCarryTheirMarkers(t *testing.T) {
	plainTUI(t)

	in := wizard()
	in.ClientModelErrs = map[string]string{persona.SelectionClaude: "no se pudo leer ~/.claude/settings.json"}
	m := toModels(t, newTUIModel(in))
	m, _ = press(t, m, downKey) // verifier -> opencode
	m, _ = press(t, m, downKey) // opencode -> claude
	if view := m.View(); !strings.Contains(view, "! no se pudo leer") {
		t.Errorf("the unreadable file is not a warning notice:\n%s", view)
	}

	lens := toReview(t, newTUIModel(wizard()))
	lens, _ = press(t, lens, downKey) // modo -> risk
	lens, _ = press(t, lens, enterKey)
	lens = typeModel(t, lens, "a b")
	lens, _ = press(t, lens, enterKey)
	if view := lens.View(); !strings.Contains(view, "× "+modelRefusal) {
		t.Errorf("the refusal is not a failure notice:\n%s", view)
	}
}

// TestTUINoticeTonesAreDistinct: the warning and the failure are different
// colours, so two states the eye must tell apart cannot collapse into one (0338).
func TestTUINoticeTonesAreDistinct(t *testing.T) {
	colouredTUI(t)

	if got := tuiLook.warning.Render("x"); got == tuiLook.failure.Render("x") {
		t.Errorf("warning and failure render the same: %q", got)
	}
	for name, line := range map[string]string{
		"warning": tuiLook.warningLine("falta un modelo"),
		"failure": tuiLook.failureLine("no se pudo leer"),
	} {
		if !strings.Contains(line, "\x1b[") {
			t.Errorf("the %s notice carries no colour: %q", name, line)
		}
	}
	if !strings.Contains(tuiLook.warningLine("falta un modelo"), "! falta un modelo") {
		t.Error("the warning does not carry its marker")
	}
	if !strings.Contains(tuiLook.failureLine("no se pudo leer"), "× no se pudo leer") {
		t.Error("the failure does not carry its marker")
	}
}

// TestTUINoModelReadsAsAWarning: an empty model is a notice, not the same voice
// as a value that exists (0338).
func TestTUINoModelReadsAsAWarning(t *testing.T) {
	colouredTUI(t)
	view := toModels(t, newTUIModel(wizard())).View()
	if !strings.Contains(view, tuiLook.warning.Render("(ninguno)")) {
		t.Errorf("the missing model is not toned as a notice:\n%q", view)
	}
}

// TestTUIEveryScreenPutsItsKeysInTheSameHelpLine: every screen ends with the
// footer the layer renders, so the keys are one piece of the design and not a
// sentence each screen invents (0338).
func TestTUIEveryScreenPutsItsKeysInTheSameHelpLine(t *testing.T) {
	plainTUI(t)

	type screen struct {
		name string
		view string
		keys string
	}
	screens := []screen{
		{"hub", newTUIModel(wizard()).View(), "↑↓ elegir · enter abrir · q salir"},
		{"summary", toSummary(t, newTUIModel(wizard())).View(), "enter aplicar · q salir"},
	}
	choice := newTUIModel(wizard())
	choice.step = stepClients
	screens = append(screens, screen{"choice", choice.View(), "↑↓ elegir · enter seguir · esc volver · q salir"})
	screens = append(screens, screen{"review", toReview(t, newTUIModel(wizard())).View(), "enter elegir · tab seguir · esc volver · q salir"})
	screens = append(screens, screen{"models", toModels(t, newTUIModel(wizard())).View(), "enter cambiar · esc volver · se aplican en el resumen"})

	profile := newTUIModel(wizard())
	profile.menu = 1
	profile = advance(t, profile)
	screens = append(screens, screen{"profile", profile.View(), "↑↓ elegir · enter cambiar · esc volver · q salir"})

	apply := newTUIModel(wizard())
	apply.step = stepApply
	screens = append(screens, screen{"apply", apply.View(), "q salir"})

	for _, s := range screens {
		if want := tuiLook.helpLine(s.keys); !strings.HasSuffix(s.view, want) {
			t.Errorf("the %s does not end with the layer's help line:\nwant suffix %q\ngot:\n%s", s.name, want, s.view)
		}
	}
}

// TestTUIFooterFollowsTheOpenControl: while a picker or an editor owns the keys,
// the footer names that control's keys; closing it returns the section's own
// footer. The three sections share one state-driven decision, so a given state
// renders the same bytes wherever the control appears (0343).
func TestTUIFooterFollowsTheOpenControl(t *testing.T) {
	plainTUI(t)

	const (
		profileKeys = "↑↓ elegir · enter cambiar · esc volver · q salir"
		sectionKeys = "enter cambiar · esc volver · se aplican en el resumen"
		pickerKeys  = "↑↓ elegir · enter cambiar · esc volver"
		editorKeys  = "escribir · enter cambiar · esc volver"
	)

	assertFooter := func(name, view, want string) {
		t.Helper()
		if got := tuiLook.helpLine(want); !strings.HasSuffix(view, got) {
			t.Errorf("%s footer:\nwant suffix %q\ngot:\n%s", name, got, view)
		}
	}

	// Profile: the value picker opens under a row and esc closes it.
	profile := newTUIModel(wizard())
	profile.menu = 1
	profile = advance(t, profile)
	assertFooter("profile default", profile.View(), profileKeys)
	picking := profile
	picking.picking = true
	assertFooter("profile picker", picking.View(), pickerKeys)
	picking.picking = false
	assertFooter("profile restored", picking.View(), profileKeys)

	// Models: the host list opens the picker; the manual entry opens the editor.
	models := toModels(t, newTUIModel(wizard()))
	assertFooter("models default", models.View(), sectionKeys)
	modelPick := models
	modelPick.pickingModel = true
	assertFooter("models picker", modelPick.View(), pickerKeys)
	modelPick.pickingModel = false
	assertFooter("models picker restored", modelPick.View(), sectionKeys)
	modelEdit := models
	modelEdit.editing = true
	assertFooter("models editor", modelEdit.View(), editorKeys)
	modelEdit.editing = false
	assertFooter("models editor restored", modelEdit.View(), sectionKeys)

	// Limits: the number editor opens under a row and esc closes it.
	limits := toLimits(t, newTUIModel(limitedState()))
	assertFooter("limits default", limits.View(), sectionKeys)
	editing := limits
	editing.editing = true
	assertFooter("limits editor", editing.View(), editorKeys)
	editing.editing = false
	assertFooter("limits restored", editing.View(), sectionKeys)
}

// TestTUIWindowMathAnswersEveryEdge: the window is arithmetic, not state, so it
// answers every corner -an empty list, a cursor outside it, a maximum of zero, a
// list exactly as long as the window- without panicking, and the cursor that
// exists is always inside the slice it returns (0338).
func TestTUIWindowMathAnswersEveryEdge(t *testing.T) {
	cases := []struct {
		name                    string
		cursor, length, maxRows int
		want                    tuiWindow
	}{
		{"empty list", 0, 0, 8, tuiWindow{}},
		{"empty list, maximum zero", 0, 0, 0, tuiWindow{}},
		{"maximum zero still shows one row", 3, 10, 0, tuiWindow{start: 3, end: 4, above: true, below: true}},
		{"cursor before the list", -4, 10, 8, tuiWindow{start: 0, end: 8, below: true}},
		{"cursor after the list", 99, 10, 8, tuiWindow{start: 2, end: 10, above: true}},
		{"exactly as long as the window", 0, 8, 8, tuiWindow{start: 0, end: 8}},
		{"exactly as long, cursor at the end", 7, 8, 8, tuiWindow{start: 0, end: 8}},
		{"cursor at the top", 0, 12, 5, tuiWindow{start: 0, end: 5, below: true}},
		{"cursor just past the window", 5, 12, 5, tuiWindow{start: 1, end: 6, above: true, below: true}},
		{"cursor at the bottom", 11, 12, 5, tuiWindow{start: 7, end: 12, above: true}},
	}

	for _, c := range cases {
		got := windowFor(c.cursor, c.length, c.maxRows)
		if got != c.want {
			t.Errorf("%s: windowFor(%d, %d, %d) = %+v, want %+v",
				c.name, c.cursor, c.length, c.maxRows, got, c.want)
		}
		// The guarantee that matters to the screen: when there is a row at all,
		// the cursor that exists is drawn.
		if c.length > 0 {
			cursor := clamp(c.cursor, c.length)
			if cursor < got.start || cursor >= got.end {
				t.Errorf("%s: cursor %d falls outside the window [%d,%d)", c.name, cursor, got.start, got.end)
			}
		}
	}
}

// TestTUILongListSaysWhatIsHidden: a windowed list draws the rows it keeps and
// says in the muted voice that more waits above or below, while the selected row
// stays on screen (0338).
func TestTUILongListSaysWhatIsHidden(t *testing.T) {
	plainTUI(t)

	const length, maxRows, cursor = 12, 5, 6
	view := tuiLook.windowLines(windowFor(cursor, length, maxRows), func(i int) string {
		return tuiLook.row(i == cursor, fmt.Sprintf("fila-%02d", i))
	})

	for _, want := range []string{"↑ más", "↓ más", "fila-06"} {
		if !strings.Contains(view, want) {
			t.Errorf("the clipped list does not show %q:\n%s", want, view)
		}
	}
	for _, hidden := range []string{"fila-00", "fila-01", "fila-11"} {
		if strings.Contains(view, hidden) {
			t.Errorf("the clipped list draws the hidden row %q:\n%s", hidden, view)
		}
	}
	if got := strings.Count(view, cursorGlyph); got != 1 {
		t.Errorf("the cursor glyph appears %d times, want the selected row once:\n%s", got, view)
	}
}

// TestTUIClippingMarksUseTheMutedTone: the marks are the layer's quiet voice,
// not the plain body and not a notice, so they never compete with a row (0338).
func TestTUIClippingMarksUseTheMutedTone(t *testing.T) {
	colouredTUI(t)

	view := tuiLook.windowLines(windowFor(6, 12, 5), func(i int) string {
		return tuiLook.row(i == 6, fmt.Sprintf("fila-%02d", i))
	})
	for _, mark := range []string{clipAbove, clipBelow} {
		if !strings.Contains(view, tuiLook.subtitle.Render(mark)) {
			t.Errorf("the clipping mark %q does not carry the muted tone:\n%q", mark, view)
		}
	}
}

// TestTUIShortListIsUnchanged: the window is the one thing a person should not
// notice. A list that fits draws exactly the bytes T1 drew, with no mark (0338).
func TestTUIShortListIsUnchanged(t *testing.T) {
	plainTUI(t)

	m := newTUIModel(wizard())
	m.step = stepClients
	want := "odachi 9.9.9 · Clientes\n\n" +
		"  › ambos (opencode y claude)\n" +
		"    opencode\n" +
		"    claude\n" +
		"\n  ↑↓ elegir · enter seguir · esc volver · q salir\n"
	if got := m.View(); got != want {
		t.Errorf("the short list changed:\nwant %q\ngot  %q", want, got)
	}
	if strings.Contains(m.View(), "más") {
		t.Errorf("a list that fits draws a clipping mark:\n%s", m.View())
	}
}

// axesForTest is a profile long enough to scroll: the shipped profiles have six
// axes, but the screen must not depend on that being short (0338).
func axesForTest(count int) []profile.Axis {
	axes := make([]profile.Axis, count)
	for i := range axes {
		axes[i] = profile.Axis{
			Name:    fmt.Sprintf("eje-%02d", i),
			Allowed: []string{"a", "b"},
			Value:   "a",
		}
	}
	return axes
}

// TestTUILongProfileKeepsTheCursorVisible: the profile axes go through the
// window too, and the keys move the cursor exactly as before -only what is drawn
// changes (0338).
func TestTUILongProfileKeepsTheCursorVisible(t *testing.T) {
	plainTUI(t)

	in := wizard()
	in.Axes = axesForTest(12)
	m := newTUIModel(in)
	m.step = stepProfile

	for i := 0; i < 10; i++ {
		m, _ = press(t, m, downKey)
	}
	if m.profileCursor != 10 {
		t.Fatalf("profileCursor after 10 downs = %d, want 10", m.profileCursor)
	}
	view := m.View()
	for _, want := range []string{"↑ más", "↓ más", "eje-10"} {
		if !strings.Contains(view, want) {
			t.Errorf("the scrolled profile does not show %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "eje-00") {
		t.Errorf("the scrolled profile draws a row hidden above:\n%s", view)
	}
	if got := strings.Count(view, cursorGlyph); got != 1 {
		t.Errorf("the cursor glyph appears %d times, want one:\n%s", got, view)
	}

	for i := 0; i < 5; i++ {
		m, _ = press(t, m, upKey)
	}
	if m.profileCursor != 5 {
		t.Fatalf("profileCursor after 5 ups = %d, want 5", m.profileCursor)
	}
	if view := m.View(); !strings.Contains(view, "eje-05") {
		t.Errorf("the selected row after scrolling back is not drawn:\n%s", view)
	}
}

// TestTUIValuePickerScrollsItsOwnWindow: the picker under an axis is a list of
// its own, and its window follows the pick cursor without hiding it (0338).
func TestTUIValuePickerScrollsItsOwnWindow(t *testing.T) {
	plainTUI(t)

	allowed := make([]string, 12)
	for i := range allowed {
		allowed[i] = fmt.Sprintf("valor-%02d", i)
	}
	in := wizard()
	in.Axes = []profile.Axis{{Name: "matiz", Allowed: allowed, Value: allowed[0]}}
	m := newTUIModel(in)
	m.step = stepProfile

	m, _ = press(t, m, enterKey) // open the picker
	if !m.picking {
		t.Fatal("enter on an axis did not open its picker")
	}
	for i := 0; i < 10; i++ {
		m, _ = press(t, m, downKey)
	}
	if m.pickCursor != 10 {
		t.Fatalf("pickCursor = %d, want 10", m.pickCursor)
	}

	view := m.View()
	if !strings.Contains(view, "    › valor-10\n") {
		t.Errorf("the scrolled picker does not draw its selected value:\n%s", view)
	}
	for _, want := range []string{"↑ más", "↓ más"} {
		if !strings.Contains(view, want) {
			t.Errorf("the scrolled picker does not show %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "      valor-00\n") {
		t.Errorf("the scrolled picker draws a value hidden above:\n%s", view)
	}
}

// openModelsSection opens the models section the way a key does, so the test
// receives the command the wizard returns: the probe is fired from the screen that
// shows the rows, never from Init, and a test proves it by running that command
// itself (0338).
func openModelsSection(t *testing.T, m tuiModel) (tuiModel, tea.Cmd) {
	t.Helper()
	m.menu = 2
	next, cmd := press(t, m, enterKey)
	if next.step != stepModels {
		t.Fatalf("step = %d, want the models section", next.step)
	}
	return next, cmd
}

// deliver runs a command and hands its message back to the model, which is the
// step the tea runtime takes between a key and the next frame (0338).
func deliver(t *testing.T, m tuiModel, cmd tea.Cmd) tuiModel {
	t.Helper()
	if cmd == nil {
		t.Fatal("cmd = nil, want the probe's command")
	}
	next, _ := m.Update(cmd())
	got, ok := next.(tuiModel)
	if !ok {
		t.Fatalf("Update returned %T, want tuiModel", next)
	}
	return got
}

// TestTUITheVerifierRowOffersTheHostsList: entering the verifier's row opens the
// catalog the host answered with the value in force marked and selected, and the
// count in the muted voice so the list reads as the host's and not the screen's
// invention (0338).
func TestTUITheVerifierRowOffersTheHostsList(t *testing.T) {
	plainTUI(t)

	in := wizard()
	in.Verify = "b/two"
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: []string{"a/one", "b/two", "c/three"}, catalogOK: true}
	m, cmd := openModelsSection(t, m)

	if view := m.View(); !strings.Contains(view, tuiLook.mutedLine("detectando…")) {
		t.Errorf("the screen does not say the probe is in flight in the muted voice:\n%q", view)
	}
	if m.pickingModel || m.editing {
		t.Fatalf("picking = %v editing = %v, want the rows readable while the probe runs", m.pickingModel, m.editing)
	}

	m = deliver(t, m, cmd)
	m, _ = press(t, m, enterKey)

	if !m.pickingModel {
		t.Fatal("pickingModel = false, want the host's list open")
	}
	view := m.View()
	for _, want := range []string{"    › b/two\n", "a/one", "c/three", "escribir a mano"} {
		if !strings.Contains(view, want) {
			t.Errorf("the picker does not carry %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, tuiLook.mutedLine("3 modelos")) {
		t.Errorf("the count is not in the muted voice:\n%q", view)
	}
}

// TestTUIEnterKeepsThePickedModelWithoutTyping: j moves the picker, enter keeps
// the entry in the section's pending map, and the apply writes it through the verb
// that owns the file -no keystroke of the identifier ever happens (0337, 0338).
func TestTUIEnterKeepsThePickedModelWithoutTyping(t *testing.T) {
	in := wizard()
	in.Verify = "a/one"
	m := newTUIModel(in)
	run := &fakeRun{catalog: []string{"a/one", "b/two"}, catalogOK: true}
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	m, _ = press(t, m, enterKey)
	if !m.pickingModel {
		t.Fatal("the verifier's row did not open the list")
	}
	m, _ = press(t, m, jKey)
	m, _ = press(t, m, enterKey)

	if m.pickingModel || m.editing {
		t.Fatalf("picking = %v editing = %v, want both closed", m.pickingModel, m.editing)
	}
	if got := m.modelPending[modelVerifier]; got != "b/two" {
		t.Fatalf("pending = %q, want the picked entry", got)
	}
	if !strings.Contains(m.View(), "b/two") {
		t.Errorf("the row does not show the picked entry:\n%s", m.View())
	}

	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if run.verify != "b/two" {
		t.Fatalf("verify = %q, want the picked model", run.verify)
	}
}

// TestTUIEscLeavesThePickerAsItWas: esc closes the list and leaves the row with
// the value it had, with nothing pending -the picker writes nothing until enter
// (0338).
func TestTUIEscLeavesThePickerAsItWas(t *testing.T) {
	plainTUI(t)

	in := wizard()
	in.Verify = "a/one"
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: []string{"a/one", "b/two"}, catalogOK: true}
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	m, _ = press(t, m, enterKey)
	m, _ = press(t, m, kKey)
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, escKey)

	if m.pickingModel || m.editing {
		t.Fatalf("picking = %v editing = %v, want both closed", m.pickingModel, m.editing)
	}
	if _, ok := m.modelPending[modelVerifier]; ok {
		t.Fatalf("pending = %v, want nothing kept", m.modelPending)
	}
	view := m.View()
	if !strings.Contains(view, "a/one") {
		t.Errorf("the row lost the value in force:\n%s", view)
	}
	if strings.Contains(view, "escribir a mano") {
		t.Errorf("esc left the list on the screen:\n%s", view)
	}
}

// TestTUITheHostIsAskedOncePerRun: the catalog costs one process, so a second
// visit finds the answer the first asked for instead of spawning again (0338).
func TestTUITheHostIsAskedOncePerRun(t *testing.T) {
	run := &fakeRun{catalog: []string{"a/one"}, catalogOK: true}
	m := newTUIModel(wizard())
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	m, _ = press(t, m, escKey)
	m, again := openModelsSection(t, m)
	if again != nil {
		t.Fatal("the second visit fired the probe again")
	}
	if run.catalogCalls != 1 {
		t.Fatalf("catalogCalls = %d, want one", run.catalogCalls)
	}
	m, _ = press(t, m, enterKey)
	if !m.pickingModel {
		t.Fatal("the second visit lost the list the first one asked for")
	}
}

// TestTUIAFailedProbeLeavesTheEditorAsTheDoor: a host that cannot answer is
// remembered as unable, the picker is not offered, and enter goes straight to the
// line editor the section already had -without retrying on every keystroke (0338).
func TestTUIAFailedProbeLeavesTheEditorAsTheDoor(t *testing.T) {
	run := &fakeRun{catalogOK: false}
	m := newTUIModel(wizard())
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	m, _ = press(t, m, enterKey)
	if m.pickingModel {
		t.Fatal("the picker was offered without a catalog")
	}
	if !m.editing {
		t.Fatal("editing = false, want the line editor")
	}

	m, _ = press(t, m, escKey) // the editor
	m, _ = press(t, m, escKey) // the section
	if _, again := openModelsSection(t, m); again != nil {
		t.Error("the failed probe was retried")
	}
	if run.catalogCalls != 1 {
		t.Errorf("catalogCalls = %d, want one", run.catalogCalls)
	}
}

// TestTUITheFailedProbeIsExplainedAndStillWrites: a host that cannot answer is a
// warning sentence on the screen, not a silent empty picker; enter still opens the
// editor, a value that is not an identifier is still refused, and a value that is
// typed still reaches the apply -the failure explains a list, it does not block the
// verb (0338).
func TestTUITheFailedProbeIsExplainedAndStillWrites(t *testing.T) {
	plainTUI(t)

	run := &fakeRun{catalogOK: false}
	m := newTUIModel(wizard())
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	if m.probe != catalogFailed {
		t.Fatalf("probe = %d, want the failed state", m.probe)
	}
	view := m.View()
	if !strings.Contains(view, tuiLook.warningLine(catalogFailedNote)) {
		t.Errorf("the screen does not explain the failed probe in the warning tone:\n%s", view)
	}
	if strings.Contains(view, catalogEmptyNote) {
		t.Errorf("a failed probe reads as an empty answer:\n%s", view)
	}
	for _, claim := range []string{tuiLook.mutedLine("0 modelos"), tuiLook.mutedLine("1 modelo")} {
		if strings.Contains(view, claim) {
			t.Errorf("the failed screen claims a count %q:\n%s", claim, view)
		}
	}

	m, _ = press(t, m, enterKey)
	if m.pickingModel || !m.editing {
		t.Fatalf("picking = %v editing = %v, want the editor as the door", m.pickingModel, m.editing)
	}

	m.buffer = "open code"
	m, _ = press(t, m, enterKey)
	if m.editErr != modelRefusal || !m.editing {
		t.Fatalf("editErr = %q editing = %v, want the refusal to keep the editor open", m.editErr, m.editing)
	}

	m.buffer = "a/one"
	m, _ = press(t, m, enterKey)
	if m.editing {
		t.Fatal("the editor stayed open after a valid value")
	}
	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if run.verify != "a/one" {
		t.Fatalf("verify = %q, want the typed model written by the apply", run.verify)
	}
	if run.catalogCalls != 1 {
		t.Errorf("catalogCalls = %d, want the probe to stay at one", run.catalogCalls)
	}
}

// TestTUIAnEmptyCatalogIsAnAnswerAndNotAFailure: "the host named nothing" and
// "the host could not be asked" are different facts, so they get different
// sentences; either way the editor stays the door and the value the file carries
// is still on its row (0338).
func TestTUIAnEmptyCatalogIsAnAnswerAndNotAFailure(t *testing.T) {
	plainTUI(t)

	in := wizard()
	in.Verify = "a/one" // a row may keep a model the host does not list
	run := &fakeRun{catalogOK: true}
	m := newTUIModel(in)
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	if m.probe != catalogReady {
		t.Fatalf("probe = %d, want the host's empty answer remembered as ready", m.probe)
	}
	if catalogEmptyNote == catalogFailedNote {
		t.Fatal("the two probe questions share one sentence")
	}
	view := m.View()
	if !strings.Contains(view, tuiLook.warningLine(catalogEmptyNote)) {
		t.Errorf("the screen does not explain the empty answer in the warning tone:\n%s", view)
	}
	if strings.Contains(view, catalogFailedNote) {
		t.Errorf("an empty answer reads as a failed probe:\n%s", view)
	}
	if !strings.Contains(view, "a/one") {
		t.Errorf("the row lost the value the file carries:\n%s", view)
	}

	m, _ = press(t, m, enterKey)
	if m.pickingModel || !m.editing {
		t.Fatalf("picking = %v editing = %v, want the editor as the door", m.pickingModel, m.editing)
	}
}

// TestTUIAClaudeRowShowsTheAliasesAndNeverTheOpenCodeCatalog: the probe answers
// OpenCode's catalogue, and a Claude row would be naming models that client does
// not serve. It opens Claude Code's declared aliases plus the manual door, and the
// muted note says what those names are; the OpenCode row keeps the host's list
// because that client is the host the probe asks (0338).
func TestTUIAClaudeRowShowsTheAliasesAndNeverTheOpenCodeCatalog(t *testing.T) {
	plainTUI(t)

	run := &fakeRun{catalog: []string{"opencode-go/x", "b/two"}, catalogOK: true}
	in := wizard()
	in.ClientModels = map[string]string{persona.SelectionClaude: "claude-sonnet-4-5-20250929"}
	m := newTUIModel(in)
	m.run = run
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	// The OpenCode row belongs to the host the probe asks, so it keeps its list.
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, enterKey)
	if !m.pickingModel || !strings.Contains(m.View(), "opencode-go/x") {
		t.Fatalf("the OpenCode row did not show the host's list:\n%s", m.View())
	}
	m, _ = press(t, m, escKey)

	// The Claude row is another client: its list is the three declared aliases.
	m, _ = press(t, m, downKey)
	m, _ = press(t, m, enterKey)
	if !m.pickingModel {
		t.Fatal("the Claude row did not open its declared list")
	}
	view := m.View()
	for _, alias := range claudeModelAliases {
		if !strings.Contains(view, alias) {
			t.Errorf("the Claude list does not carry the alias %q:\n%s", alias, view)
		}
	}
	if strings.Contains(view, "opencode-go/x") || strings.Contains(view, "b/two") {
		t.Errorf("the Claude list shows the OpenCode catalogue:\n%s", view)
	}
	if !strings.Contains(view, tuiLook.mutedLine(claudeAliasNote)) {
		t.Errorf("the Claude list does not say the names are aliases:\n%s", view)
	}
	if !strings.Contains(view, "escribir a mano") {
		t.Errorf("the Claude list lost the manual door:\n%s", view)
	}
	if !strings.Contains(view, "    › claude-sonnet-4-5-20250929\n") {
		t.Errorf("the value in force is not the selected row:\n%s", view)
	}

	// The last row is the way to type a full identifier, and it reaches the editor.
	for i := 0; i < len(m.modelChoices(m.pickedValue()))-1; i++ {
		m, _ = press(t, m, downKey)
	}
	m, _ = press(t, m, enterKey)
	if m.pickingModel || !m.editing {
		t.Fatalf("picking = %v editing = %v, want the manual door to open the editor", m.pickingModel, m.editing)
	}
	m.buffer = "sonnet"
	m, _ = press(t, m, enterKey)
	if got := m.modelPending[persona.SelectionClaude]; got != "sonnet" {
		t.Fatalf("pending = %q, want the alias typed by hand", got)
	}
}

// TestTUIAModelTheHostStoppedListingIsItsOwnRow: a project may name a model the
// host no longer serves, so the picker shows it first and selects it instead of
// hiding the value in force behind an entry nobody chose (0338).
func TestTUIAModelTheHostStoppedListingIsItsOwnRow(t *testing.T) {
	in := wizard()
	in.Verify = "gone/model"
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: []string{"a/one"}, catalogOK: true}
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)
	m, _ = press(t, m, enterKey)

	view := m.View()
	if !strings.Contains(view, "    › gone/model\n") {
		t.Errorf("the value the host does not list is not the selected row:\n%s", view)
	}
	if !strings.Contains(view, tuiLook.mutedLine("1 modelo")) {
		t.Errorf("the count does not read as one model:\n%q", view)
	}
}

// TestTUIEscribirAManoOpensTheEditor: the picker's last row is the door out of the
// list, and it opens the line editor with the value in force in it, refusal
// included (0338).
func TestTUIEscribirAManoOpensTheEditor(t *testing.T) {
	in := wizard()
	in.Verify = "a/one"
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: []string{"a/one", "b/two"}, catalogOK: true}
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)

	m, _ = press(t, m, enterKey) // the verifier row
	for i := 0; i < len(m.modelChoices("a/one"))-1; i++ {
		m, _ = press(t, m, downKey)
	}
	if !m.pickingModel {
		t.Fatal("the walk through the list closed the picker")
	}
	m, _ = press(t, m, enterKey)

	if m.pickingModel || !m.editing {
		t.Fatalf("picking = %v editing = %v, want the line editor", m.pickingModel, m.editing)
	}
	if m.buffer != "a/one" {
		t.Fatalf("buffer = %q, want the value in force", m.buffer)
	}

	m = typeModel(t, m, "#fast")
	m, _ = press(t, m, enterKey)
	if got := m.modelPending[modelVerifier]; got != "a/one#fast" {
		t.Fatalf("pending = %q, want the typed value", got)
	}
}

// TestTUILongModelListWindows: the picker is a list like every other one, so a
// catalog of forty draws a window with its marks, not forty rows, and the selected
// entry stays on screen (0338).
func TestTUILongModelListWindows(t *testing.T) {
	plainTUI(t)

	catalog := make([]string, 40)
	for i := range catalog {
		catalog[i] = fmt.Sprintf("host/model-%02d", i)
	}
	in := wizard()
	in.Verify = catalog[39]
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: catalog, catalogOK: true}
	m, cmd := openModelsSection(t, m)
	m = deliver(t, m, cmd)
	m, _ = press(t, m, enterKey)

	view := m.View()
	for _, want := range []string{"↑ más", "host/model-39", "40 modelos"} {
		if !strings.Contains(view, want) {
			t.Errorf("the windowed picker does not show %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "host/model-00") {
		t.Errorf("the picker drew an entry hidden above its window:\n%s", view)
	}
	// One cursor, and it belongs to the list: the row that opened it drops its
	// glyph so the keys have a single owner (0338).
	if got := strings.Count(view, cursorGlyph); got != 1 {
		t.Errorf("the cursor glyph appears %d times, want only the picker's entry:\n%s", got, view)
	}
}

// TestTUITheLensRowOffersTheHostsList: the review screen's lens rows open the same
// picker, and a chosen entry lands in the lens table the apply writes -the editor
// stays as the door for a value nobody lists (0338).
func TestTUITheLensRowOffersTheHostsList(t *testing.T) {
	in := wizard()
	in.Models = map[string]string{"risk": "a/one"}
	m := newTUIModel(in)
	m.run = &fakeRun{catalog: []string{"a/one", "b/two"}, catalogOK: true}

	var cmd tea.Cmd
	for i := 0; i < 4; i++ {
		m, cmd = press(t, m, enterKey)
	}
	if m.step != stepReview {
		t.Fatalf("step = %d, want the review screen", m.step)
	}
	m = deliver(t, m, cmd)

	m, _ = press(t, m, downKey) // modo -> risk
	m, _ = press(t, m, enterKey)
	if !m.pickingModel {
		t.Fatal("the lens row did not open the host's list")
	}
	m, _ = press(t, m, jKey) // a/one -> b/two
	m, _ = press(t, m, enterKey)

	if m.pickingModel || m.editing {
		t.Fatalf("picking = %v editing = %v, want both closed", m.pickingModel, m.editing)
	}
	if got := m.state.Models["risk"]; got != "b/two" {
		t.Fatalf("risk = %q, want the picked entry", got)
	}
}

// TestTUIRunnerCatalogAsksTheProbeFromTheProject: the wizard's catalog is the same
// bounded probe the review verb validates its lens table against, run from the
// folder the wizard configures (0338).
func TestTUIRunnerCatalogAsksTheProbeFromTheProject(t *testing.T) {
	var dir string
	r := newTUIRunner("proyecto")
	r.catalogFn = func(want string) ([]string, bool) {
		dir = want
		return []string{"a/one"}, true
	}

	models, ok := r.Catalog()
	if !ok || len(models) != 1 || models[0] != "a/one" {
		t.Fatalf("Catalog = %v %v, want the probe's answer", models, ok)
	}
	if dir != "proyecto" {
		t.Fatalf("probe dir = %q, want the folder the wizard configures", dir)
	}
}

// limitedState is a wizard state with the four scalars the new section shows, so
// a row kept as it was compares equal to what the file carries (0342).
func limitedState() tuiState {
	in := wizard()
	in.MaxPromptKB = 120
	in.KeepVerdicts = 4
	in.KeepPrompts = 2
	in.CheckTimeout = 45
	return in
}

// menuStep finds the hub row that opens a step. A new section is a new row, so a
// helper that assumed "last" would silently open the wrong screen (0346).
func menuStep(step tuiStep) int {
	for i, section := range tuiMenu {
		if section.step == step {
			return i
		}
	}
	return -1
}

// toLimits opens the limits section from the hub.
func toLimits(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	m.menu = menuStep(stepLimits)
	m = advance(t, m)
	if m.step != stepLimits {
		t.Fatalf("step = %d, want the limits section", m.step)
	}
	return m
}

// checkedState is a wizard state that carries one declared check, the kind the
// checks section edits. The severity is the project's gate and stays in the file
// (0346).
func checkedState() tuiState {
	in := limitedState()
	in.Checks = []review.Check{
		{Name: "lint", Command: "go vet ./...", Severity: review.Critical, Remedy: "fix it"},
	}
	return in
}

// toChecks opens the checks section from the hub.
func toChecks(t *testing.T, m tuiModel) tuiModel {
	t.Helper()
	m.menu = menuStep(stepChecks)
	m = advance(t, m)
	if m.step != stepChecks {
		t.Fatalf("step = %d, want the checks section", m.step)
	}
	return m
}

// editCheckField edits one of an open check's fields, by its index, and keeps the
// value. The check must already be open with its cursor on the first field.
func editCheckField(t *testing.T, m tuiModel, field int, value string) tuiModel {
	t.Helper()
	m.checkField = field
	m, _ = press(t, m, enterKey)
	m = typeModel(t, m, value)
	m, _ = press(t, m, enterKey)
	return m
}

// TestTUILimitsSectionIsInTheMenuAndRoute: the fourth section is a row of the hub
// and a row of the route table, reachable forward and back like the other two
// (0336, 0342).
func TestTUILimitsSectionIsInTheMenuAndRoute(t *testing.T) {
	m := newTUIModel(limitedState())
	if view := m.View(); !strings.Contains(view, "límites") {
		t.Fatalf("the hub does not offer the limits section:\n%s", view)
	}
	if got := menuStep(stepLimits); got < 0 {
		t.Errorf("the menu has no row opening the limits step")
	}
	route, ok := tuiRoutes[stepLimits]
	if !ok || route.back != stepState {
		t.Errorf("the limits route = %+v, want back to the menu", route)
	}
	opened := toLimits(t, m)
	for _, want := range []string{"prompt KB", "veredictos", "prompts", "timeout s"} {
		if view := opened.View(); !strings.Contains(view, want) {
			t.Errorf("the section does not show %q:\n%s", want, view)
		}
	}
	if back := opened.back(); back.step != stepState {
		t.Errorf("esc from the limits section landed on step %d, want the menu", back.step)
	}
}

// TestTheLimitsSectionEditsAndAppliesOnlyOnApply: a scalar chosen in the section
// stays in the model, nothing is written before the summary, and only the scalar
// that changed reaches the review writer on apply (0342).
func TestTheLimitsSectionEditsAndAppliesOnlyOnApply(t *testing.T) {
	run := &fakeRun{}
	m := toLimits(t, newTUIModel(limitedState()))
	m.run = run

	m, _ = press(t, m, enterKey) // prompt KB opens the editor
	if !m.editing {
		t.Fatal("enter did not open the number editor")
	}
	m = typeModel(t, m, "200")
	m, _ = press(t, m, enterKey)
	if view := m.View(); !strings.Contains(view, "200") {
		t.Errorf("the row does not show the chosen value:\n%s", view)
	}
	if len(run.calls) != 0 {
		t.Fatalf("choosing a value wrote %v", run.calls)
	}

	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if m.step != stepApply {
		t.Fatalf("step = %d, want the apply screen", m.step)
	}
	if want := []string{"install", "voice", "review", "models", "limits"}; !slices.Equal(run.calls, want) {
		t.Fatalf("calls = %v, want %v", run.calls, want)
	}
	if run.limits.MaxPromptKB == nil || *run.limits.MaxPromptKB != 200 {
		t.Errorf("MaxPromptKB = %v, want 200", run.limits.MaxPromptKB)
	}
	if run.limits.KeepVerdicts != nil || run.limits.KeepPrompts != nil || run.limits.CheckTimeout != nil {
		t.Errorf("untouched scalars were written: %+v", run.limits)
	}
}

// TestAnUnchangedLimitsSectionWritesNothing: a section nobody opened and a row
// retyped with the value it already carried add no write at all (0342).
func TestAnUnchangedLimitsSectionWritesNothing(t *testing.T) {
	never := toSummary(t, newTUIModel(limitedState()))
	never.run = &fakeRun{}
	never, _ = press(t, never, enterKey)
	if slices.Contains(never.run.(*fakeRun).calls, "limits") {
		t.Errorf("calls = %v, want no limits write from a section nobody opened", never.run.(*fakeRun).calls)
	}

	m := toLimits(t, newTUIModel(limitedState()))
	run := &fakeRun{}
	m.run = run
	m, _ = press(t, m, enterKey) // prompt KB
	m = typeModel(t, m, "120")   // the value the file already carries
	m, _ = press(t, m, enterKey)

	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if slices.Contains(run.calls, "limits") {
		t.Errorf("calls = %v, want no limits write for the value already there", run.calls)
	}
}

// TestSetLimitsRefusesAnInvalidValue: a value the domain cannot accept keeps the
// editor open with the refusal on the screen, and nothing is pending (0342).
func TestSetLimitsRefusesAnInvalidValue(t *testing.T) {
	m := toLimits(t, newTUIModel(limitedState()))
	m, _ = press(t, m, downKey) // prompt KB -> veredictos
	m, _ = press(t, m, enterKey)
	m = typeModel(t, m, "not a number")
	m, _ = press(t, m, enterKey)
	if !m.editing || m.editErr == "" {
		t.Fatalf("editing = %v editErr = %q, want the refusal on the screen", m.editing, m.editErr)
	}
	if _, ok := m.limitsPending[limitKeepVerdicts]; ok {
		t.Fatalf("a refused value was kept: %v", m.limitsPending)
	}
}

// TestTUITheMenuOffersTheChecksSection: the fifth section is a row of the hub and
// a row of the route table, reachable and left like the other four (0346).
func TestTUITheMenuOffersTheChecksSection(t *testing.T) {
	m := newTUIModel(checkedState())
	if view := m.View(); !strings.Contains(view, "checks") {
		t.Fatalf("the hub does not offer the checks section:\n%s", view)
	}
	route, ok := tuiRoutes[stepChecks]
	if !ok || route.back != stepState {
		t.Errorf("the checks route = %+v, want back to the menu", route)
	}
	opened := toChecks(t, m)
	if back := opened.back(); back.step != stepState {
		t.Errorf("esc from the checks section landed on step %d, want the menu", back.step)
	}
}

// TestTheChecksSectionListsEditsAndAppliesOnlyOnApply: the section lists the
// declared check with the severity it keeps, opens the three editable fields,
// writes nothing before the summary, and sends the identity with the three new
// values on apply (0346).
func TestTheChecksSectionListsEditsAndAppliesOnlyOnApply(t *testing.T) {
	run := &fakeRun{}
	m := toChecks(t, newTUIModel(checkedState()))
	m.run = run

	if view := m.View(); !strings.Contains(view, "lint") || !strings.Contains(view, "CRITICAL") {
		t.Fatalf("the section does not list the check and its severity:\n%s", view)
	}
	m, _ = press(t, m, enterKey) // open the check's fields
	if !m.checkOpen {
		t.Fatal("enter did not open the check fields")
	}
	view := m.View()
	for _, field := range []string{"nombre", "comando", "remedio"} {
		if !strings.Contains(view, field) {
			t.Errorf("the open check does not show %q:\n%s", field, view)
		}
	}

	m = editCheckField(t, m, checkFieldName, "vet")
	m = editCheckField(t, m, checkFieldCommand, "go vet ./...")
	m = editCheckField(t, m, checkFieldRemedy, "fix the vet findings")
	if len(run.calls) != 0 {
		t.Fatalf("editing a check wrote %v", run.calls)
	}

	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if m.step != stepApply {
		t.Fatalf("step = %d, want the apply screen", m.step)
	}
	if !slices.Contains(run.calls, "checks") {
		t.Fatalf("calls = %v, want a checks write", run.calls)
	}
	if len(run.checks) != 1 {
		t.Fatalf("checks = %+v, want one edit", run.checks)
	}
	edit := run.checks[0]
	if edit.Identity != "lint" || edit.Name != "vet" || edit.Command != "go vet ./..." || edit.Remedy != "fix the vet findings" {
		t.Errorf("check edit = %+v, want the identity and the three new fields", edit)
	}
}

// TestAnUnchangedChecksSectionWritesNothing: a section nobody opened, and a field
// retyped with the value it already carried, add no write at all (0346).
func TestAnUnchangedChecksSectionWritesNothing(t *testing.T) {
	never := toSummary(t, newTUIModel(checkedState()))
	never.run = &fakeRun{}
	never, _ = press(t, never, enterKey)
	if slices.Contains(never.run.(*fakeRun).calls, "checks") {
		t.Errorf("calls = %v, want no checks write from a section nobody opened", never.run.(*fakeRun).calls)
	}

	m := toChecks(t, newTUIModel(checkedState()))
	run := &fakeRun{}
	m.run = run
	m, _ = press(t, m, enterKey)                     // open the check
	m = editCheckField(t, m, checkFieldName, "lint") // the value the file already carries
	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if slices.Contains(run.calls, "checks") {
		t.Errorf("calls = %v, want no checks write for the value already there", run.calls)
	}
}

// TestSetCheckRefusesAnEmptyNameOrCommand: a field the domain cannot accept stays
// in the editor with its refusal, and nothing is kept as a pending edit (0346).
func TestSetCheckRefusesAnEmptyNameOrCommand(t *testing.T) {
	m := toChecks(t, newTUIModel(checkedState()))
	m, _ = press(t, m, enterKey) // open the check

	m = editCheckField(t, m, checkFieldName, "  ")
	if !m.editing || m.editErr == "" {
		t.Fatalf("editing = %v editErr = %q, want the blank name refused", m.editing, m.editErr)
	}
	if edit, ok := m.checksPending["lint"]; ok && edit.Name == "" {
		t.Fatalf("a blank name was kept: %+v", m.checksPending)
	}

	m = typeModel(t, m, "vet")
	m, _ = press(t, m, enterKey)
	m = editCheckField(t, m, checkFieldCommand, "  ")
	if !m.editing || m.editErr == "" {
		t.Fatalf("editing = %v editErr = %q, want the blank command refused", m.editing, m.editErr)
	}
}

// TestChecksApplyNoticeNamesApprovalWhenTheFingerprintChanges: a new name or a
// new command is a new list, so the apply screen says the clone must approve it
// again with the exact command that does (0346).
func TestChecksApplyNoticeNamesApprovalWhenTheFingerprintChanges(t *testing.T) {
	for _, tc := range []struct {
		name  string
		field int
		value string
	}{
		{"name", checkFieldName, "vet"},
		{"command", checkFieldCommand, "go test ./..."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := toChecks(t, newTUIModel(checkedState()))
			m.run = &fakeRun{}
			m, _ = press(t, m, enterKey) // open the check
			m = editCheckField(t, m, tc.field, tc.value)
			m.step = stepSummary
			m, _ = press(t, m, enterKey)
			if view := m.View(); !strings.Contains(view, "odachi review checks approve") {
				t.Errorf("the apply notice does not name the approval command:\n%s", view)
			}
		})
	}
}

// TestChecksApplyNoticeIsQuietForARemedyOnlyEdit: the remedy is advice, not part
// of the fingerprint, so rewording it alone asks nothing of the human (0346).
func TestChecksApplyNoticeIsQuietForARemedyOnlyEdit(t *testing.T) {
	m := toChecks(t, newTUIModel(checkedState()))
	m.run = &fakeRun{}
	m, _ = press(t, m, enterKey) // open the check
	m = editCheckField(t, m, checkFieldRemedy, "run go vet and fix what it says")
	m.step = stepSummary
	m, _ = press(t, m, enterKey)
	if view := m.View(); strings.Contains(view, "odachi review checks approve") {
		t.Errorf("a remedy-only edit asked for a new approval:\n%s", view)
	}
}
