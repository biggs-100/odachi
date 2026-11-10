package main

import (
	"path/filepath"
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

	config, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn})
	if err != nil {
		t.Fatalf("reviewtoml.Render: %v", err)
	}
	writeFile(t, dir, filepath.FromSlash(review.ModePath), string(config))
}

// keys the wizard answers, in the form the program delivers them.
var (
	enterKey = tea.KeyMsg{Type: tea.KeyEnter}
	downKey  = tea.KeyMsg{Type: tea.KeyDown}
	upKey    = tea.KeyMsg{Type: tea.KeyUp}
	escKey   = tea.KeyMsg{Type: tea.KeyEsc}
	qKey     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
)

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
func TestTUIEnterWalksTheSteps(t *testing.T) {
	m := newTUIModel(wizard())
	for want := stepClients; want <= stepSummary; want++ {
		m, _ = press(t, m, enterKey)
		if m.step != want {
			t.Fatalf("step = %d, want %d", m.step, want)
		}
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
	if m.step != stepSummary {
		t.Fatalf("step = %d, want summary", m.step)
	}
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
	m, _ = press(t, m, downKey) // novato -> senior
	m, _ = press(t, m, enterKey)

	view := m.View()
	for _, want := range []string{"opencode", "project", "senior"} {
		if !strings.Contains(view, want) {
			t.Errorf("the summary does not show %q:\n%s", want, view)
		}
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
	}
	if got != want {
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
