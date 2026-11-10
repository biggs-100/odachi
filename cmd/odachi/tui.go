package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/biggs-100/odachi/internal/app/configure"
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
	Models  map[string]string
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
	model.run = newTUIRunner(dir)
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
	state.Models = make(map[string]string, len(config.Lenses))
	for lens, model := range config.Lenses {
		state.Models[lens] = model
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
	stepReview
	stepSummary
	stepApply
)

// The options each single-choice screen offers, in the order a human reads them.
var (
	tuiClients = []string{persona.SelectionBoth, persona.SelectionOpenCode, persona.SelectionClaude}
	tuiSkills  = []string{persona.SkillsHome, persona.SkillsProject}
	tuiVoices  = []string{"novato", "senior", "ci"}
	tuiLenses  = review.AllLenses()
)

// tuiModel is the wizard: the detected facts, the choices and where it stands.
type tuiModel struct {
	state        tuiState
	step         tuiStep
	width        int
	client       int
	skills       int
	voice        int
	reviewCursor int
	editing      bool
	buffer       string
	editErr      string
	run          configure.Runner
	applied      bool
	result       configure.Result
}

// newTUIModel is the model a run starts with, positioned on what the project
// already chose.
func newTUIModel(state tuiState) tuiModel {
	if state.Models == nil {
		state.Models = map[string]string{}
	}
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
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.editing {
			return m.editKey(msg), nil
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc":
			if m.step > stepState {
				m.step--
			}
		case "enter":
			if m.step == stepReview {
				return m.enterReview()
			}
			if m.step == stepSummary {
				return m.apply()
			}
			return m.next()
		case "tab":
			if m.step == stepReview {
				m.step = stepSummary
			}
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
		m.step = stepReview
	}
	return m, nil
}

// enterReview toggles the switch on its row, or opens the model of the lens the
// cursor is on.
func (m tuiModel) enterReview() (tea.Model, tea.Cmd) {
	if m.reviewCursor == 0 {
		if m.state.Review == review.ModeOn {
			m.state.Review = review.ModeOff
		} else {
			m.state.Review = review.ModeOn
		}
		return m, nil
	}
	lens := tuiLenses[m.reviewCursor-1]
	m.editing = true
	m.buffer = m.state.Models[lens]
	m.editErr = ""
	return m, nil
}

// editKey edits the model of a lens. The value is an identifier, so it is
// checked before it is saved; an empty value removes the lens from the table.
func (m tuiModel) editKey(msg tea.KeyMsg) tuiModel {
	switch msg.Type {
	case tea.KeyEsc:
		m.editing = false
		m.buffer = ""
		m.editErr = ""
	case tea.KeyEnter:
		lens := tuiLenses[m.reviewCursor-1]
		value := strings.TrimSpace(m.buffer)
		if value == "" {
			delete(m.state.Models, lens)
			m.editing = false
			m.buffer = ""
			m.editErr = ""
			return m
		}
		checked, err := review.NormalizeModel(value)
		if err != nil {
			m.editErr = "el modelo no es un identificador: usá provider/name#variant"
			return m
		}
		m.state.Models[lens] = checked
		m.editing = false
		m.buffer = ""
		m.editErr = ""
	case tea.KeyBackspace:
		runes := []rune(m.buffer)
		if len(runes) > 0 {
			m.buffer = string(runes[:len(runes)-1])
		}
	case tea.KeySpace:
		m.buffer += " "
	case tea.KeyRunes:
		m.buffer += string(msg.Runes)
	}
	return m
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
	case stepReview:
		m.reviewCursor = clamp(m.reviewCursor+delta, 1+len(tuiLenses))
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
	case stepReview:
		return m.viewReview()
	case stepSummary:
		return m.viewSummary()
	case stepApply:
		return m.viewApply()
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
	return m.header("resumen") + m.facts() + "\n  enter aplicar · q salir\n"
}

// viewApply renders the outcome of every planned operation.
func (m tuiModel) viewApply() string {
	var out strings.Builder
	out.WriteString(m.header("aplicar"))
	for _, step := range m.result.Steps {
		mark := "·"
		switch step.Status {
		case configure.Done:
			mark = "✓"
		case configure.Failed:
			mark = "✗"
		}
		line := mark + " " + step.Label
		if step.Err != nil {
			line += ": " + step.Err.Error()
		}
		out.WriteString("  " + line + "\n")
	}
	if m.applied {
		out.WriteString("\n  " + m.result.Message + "\n")
	}
	out.WriteString("\n  q salir\n")
	return out.String()
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

// apply runs the chosen configuration through the runner and shows every
// operation with its outcome. Nothing runs until the human presses enter on the
// summary: q before it leaves the tree alone.
func (m tuiModel) apply() (tea.Model, tea.Cmd) {
	if m.run == nil {
		return m, nil
	}
	m.result = configure.Apply(configure.State{
		Clients:    m.state.Clients,
		Skills:     m.state.Skills,
		Voice:      m.state.Voice,
		ReviewMode: m.state.Review,
		LensModels: m.state.Models,
		Project:    m.state.Project,
	}, m.run)
	m.applied = true
	m.step = stepApply
	return m, nil
}

// viewReview renders the switch and the four lens models. The model being edited
// shows its buffer and a cursor; the error stays on the screen until it is fixed.
func (m tuiModel) viewReview() string {
	var out strings.Builder
	out.WriteString(m.header("Review"))
	labels := append([]string{"modo"}, tuiLenses...)
	for i, label := range labels {
		marker := "  "
		if i == m.reviewCursor {
			marker = "> "
		}
		value := m.state.Review
		if i > 0 {
			lens := tuiLenses[i-1]
			value = m.state.Models[lens]
			if m.editing && i == m.reviewCursor {
				value = m.buffer + "_"
			}
		}
		out.WriteString("  " + marker + fmt.Sprintf("%-12s %s\n", label, value))
	}
	if m.editErr != "" {
		out.WriteString("\n  " + m.editErr + "\n")
	}
	out.WriteString("\n  enter elegir · tab seguir · esc volver · q salir\n")
	return out.String()
}

// modelsLine renders the lens table for the summary.
func modelsLine(models map[string]string) string {
	if len(models) == 0 {
		return "ninguno: usa el modelo de la sesión"
	}
	names := make([]string, 0, len(models))
	for lens := range models {
		names = append(names, lens)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, lens := range names {
		parts = append(parts, lens+"="+models[lens])
	}
	return strings.Join(parts, ", ")
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
		{"Modelos", modelsLine(m.state.Models)},
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
