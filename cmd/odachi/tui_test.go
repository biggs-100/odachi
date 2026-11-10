package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
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
	m, _ = press(t, m, downKey)  // novato -> senior
	m, _ = press(t, m, downKey)  // senior -> ci
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
	calls   []string
	failAt  string
	clients string
	skills  string
	project bool
	voice   string
	review  string
	models  map[string]string
}

func (f *fakeRun) record(op string) error {
	f.calls = append(f.calls, op)
	if f.failAt == op {
		return errTUI
	}
	return nil
}

func (f *fakeRun) Install(project bool, clients, skills string) error {
	f.project, f.clients, f.skills = project, clients, skills
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

	if err := r.Install(true, persona.SelectionOpenCode, persona.SkillsProject); err != nil {
		t.Fatalf("Install(project): %v", err)
	}
	if got := strings.Join(initArgs, " "); got != "--agents opencode --skills project" {
		t.Fatalf("init args = %q", got)
	}

	if err := r.Install(false, persona.SelectionBoth, persona.SkillsHome); err != nil {
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
		fmt.Fprintln(stderr, "odachi: the store is broken")
		return 1
	}

	err := r.Install(true, persona.SelectionBoth, persona.SkillsHome)
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
// chose, not from the wizard's defaults.
func TestTUIDetectsTheState(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}

	want := tuiState{
		Version: version,
		Project: true,
		Clients: persona.SelectionOpenCode,
		Skills:  persona.SkillsProject,
		Voice:   "senior",
		Review:  review.ModeOn,
		Models:  map[string]string{"reliability": "opencode-go/deepseek-v4.1-flash"},
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
