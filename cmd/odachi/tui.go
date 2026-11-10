package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// tuiState is what the first screen reads: the facts the wizard starts from. It
// is detected, never written: opening the TUI changes nothing.
type tuiState struct {
	Version string
	Project bool
	Clients string
	Skills  string
	Voice   string
	Review  string
}

// runTUI opens the installation and configuration wizard. It refuses before the
// first screen when there is no terminal: a wizard that writes escape codes into a
// pipe is a script's problem disguised as a screen.
func runTUI(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {

	if len(args) > 0 {
		if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: tui takes no flags: odachi tui")
		return start.CodeBlocked
	}

	if !isTerminal(stdin) || streamTarget(stdout) == nil {
		fmt.Fprintln(stderr, "odachi: la tui necesita una terminal; configurá con odachi setup (o odachi init) y sus flags")
		return start.CodeBlocked
	}

	state, err := detectTUIState(dir)
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	model := newTUIModel(state)
	if _, err := tea.NewProgram(model, tea.WithInput(stdin), tea.WithOutput(stdout)).Run(); err != nil {
		fmt.Fprintln(stderr, "odachi: la tui se cerró con un error: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	return start.CodeOK
}

// detectTUIState reads the facts the wizard shows first. Every read is a read: no
// file is written and no state is changed.
func detectTUIState(dir string) (tuiState, error) {
	state := tuiState{Version: version, Review: "off"}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	switch {
	case err == nil:
		state.Project = rules.HasBlock(string(content))
	case !errors.Is(err, os.ErrNotExist):
		return tuiState{}, fmt.Errorf("could not read AGENTS.md: %w", err)
	}

	stored, err := agentSelection{}.Written(dir)
	if err != nil {
		return tuiState{}, fmt.Errorf("could not read %s: %w", persona.SelectionPath, err)
	}
	state.Clients = strings.TrimSpace(stored.Agents)
	if state.Clients == "" {
		state.Clients = persona.SelectionBoth
	}
	state.Skills = strings.TrimSpace(stored.Skills)
	if state.Skills == "" {
		state.Skills = persona.SkillsHome
	}

	voice, code, message := currentProfile(dir, "odachi tui", false)
	if code != start.CodeOK {
		return tuiState{}, errors.New(message)
	}
	if voice != nil {
		state.Voice = voice.Name
	}

	config, _, problem := cleanLimits(dir)
	if problem != "" {
		return tuiState{}, errors.New(problem)
	}
	if strings.EqualFold(config.Mode, review.ModeOn) {
		state.Review = review.ModeOn
	}
	return state, nil
}

// tuiStep is where the wizard is. The review and the apply screens arrive in the
// next steps; the summary is where they will hang.
type tuiStep int

const (
	stepState tuiStep = iota
	stepClients
	stepSkills
	stepVoice
	stepSummary
)

// The options each single-choice screen offers, in the order a human reads them.
var (
	tuiClients = []string{persona.SelectionBoth, persona.SelectionOpenCode, persona.SelectionClaude}
	tuiSkills  = []string{persona.SkillsHome, persona.SkillsProject}
	tuiVoices  = []string{"novato", "senior", "ci"}
)

// tuiModel is the wizard: the detected facts, the choices and where it stands.
type tuiModel struct {
	state  tuiState
	step   tuiStep
	width  int
	client int
	skills int
	voice  int
}

// newTUIModel is the model a run starts with, positioned on what the project
// already chose.
func newTUIModel(state tuiState) tuiModel {
	return tuiModel{
		state:  state,
		client: indexOf(tuiClients, state.Clients),
		skills: indexOf(tuiSkills, state.Skills),
		voice:  indexOf(tuiVoices, state.Voice),
	}
}

// Init does nothing: the first screen is already the state.
func (m tuiModel) Init() tea.Cmd { return nil }

// Update handles the keys the wizard answers. Nothing here touches the disk: the
// choices are the model's until the apply screen runs them.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.step > stepState {
				m.step--
			}
		case "enter":
			return m.next()
		case "up", "k":
			m = m.move(-1)
		case "down", "j":
			m = m.move(1)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

// next advances one screen, committing the choice the screen was showing.
func (m tuiModel) next() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepState:
		m.step = stepClients
	case stepClients:
		m.state.Clients = tuiClients[m.client]
		m.step = stepSkills
	case stepSkills:
		m.state.Skills = tuiSkills[m.skills]
		m.step = stepVoice
	case stepVoice:
		m.state.Voice = tuiVoices[m.voice]
		m.step = stepSummary
	}
	return m, nil
}

// move changes the cursor of the screen the wizard is on. The detection and the
// summary are not choices.
func (m tuiModel) move(delta int) tuiModel {
	switch m.step {
	case stepClients:
		m.client = clamp(m.client+delta, len(tuiClients))
	case stepSkills:
		m.skills = clamp(m.skills+delta, len(tuiSkills))
	case stepVoice:
		m.voice = clamp(m.voice+delta, len(tuiVoices))
	}
	return m
}

// View renders the screen the wizard is on. The copy is Spanish because the human
// reads it; the identifiers and the code stay English.
func (m tuiModel) View() string {
	switch m.step {
	case stepClients:
		return m.viewChoices("Clientes", clientLabels(), m.client)
	case stepSkills:
		return m.viewChoices("Skills", tuiSkills, m.skills)
	case stepVoice:
		return m.viewChoices("Voz", tuiVoices, m.voice)
	case stepSummary:
		return m.viewSummary()
	default:
		return m.viewState()
	}
}

// viewState is the first screen: where the project is before any choice.
func (m tuiModel) viewState() string {
	return m.header("instalación y configuración") + m.facts() + "\n  enter seguir · q salir\n"
}

// viewSummary is the last screen before the apply: the choices as they will be.
func (m tuiModel) viewSummary() string {
	return m.header("resumen") + m.facts() + "\n  q salir\n"
}

// viewChoices renders one single-choice screen with its cursor.
func (m tuiModel) viewChoices(title string, options []string, cursor int) string {
	var out strings.Builder
	out.WriteString(m.header(title))
	for i, option := range options {
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		out.WriteString("  " + marker + option + "\n")
	}
	out.WriteString("\n  ↑↓ elegir · enter seguir · esc volver · q salir\n")
	return out.String()
}

// header is the line every screen starts with.
func (m tuiModel) header(title string) string {
	return lipgloss.NewStyle().Bold(true).Render("odachi "+m.state.Version) + " · " + title + "\n\n"
}

// facts renders the state as the rows a human reads.
func (m tuiModel) facts() string {
	project := "no: esta carpeta no tiene el bloque de odachi"
	if m.state.Project {
		project = "sí: el bloque de odachi está en AGENTS.md"
	}

	rows := [][2]string{
		{"Proyecto", project},
		{"Clientes", m.state.Clients},
		{"Skills", m.state.Skills},
		{"Voz", m.state.Voice},
		{"Review", m.state.Review},
	}

	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "  %-10s %s\n", row[0], row[1])
	}
	return out.String()
}

// clientLabels names the three client choices the way the screen reads them.
func clientLabels() []string {
	labels := make([]string, len(tuiClients))
	for i, value := range tuiClients {
		labels[i] = value
		if value == persona.SelectionBoth {
			labels[i] = "ambos (opencode y claude)"
		}
	}
	return labels
}

// indexOf is the cursor a value sits on, or the first option for a value the
// screen does not offer.
func indexOf(options []string, value string) int {
	for i, option := range options {
		if option == value {
			return i
		}
	}
	return 0
}

// clamp keeps a cursor inside its list.
func clamp(value, length int) int {
	if value < 0 {
		return 0
	}
	if value >= length {
		return length - 1
	}
	return value
}
