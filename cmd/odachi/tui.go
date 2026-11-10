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

// tuiModel is the first screen: the detected facts and the way out. The screens
// that choose and apply arrive in the next steps.
type tuiModel struct {
	state tuiState
	width int
}

// newTUIModel is the model a run starts with.
func newTUIModel(state tuiState) tuiModel {
	return tuiModel{state: state}
}

// Init does nothing: the first screen is already the state.
func (m tuiModel) Init() tea.Cmd { return nil }

// Update handles the window size and the one key the first screen answers.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

// View renders the first screen. The copy is Spanish because the human reads it;
// the identifiers and the code stay English.
func (m tuiModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("odachi " + m.state.Version)
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
	out.WriteString(title + " · instalación y configuración\n\n")
	for _, row := range rows {
		fmt.Fprintf(&out, "  %-10s %s\n", row[0], row[1])
	}
	out.WriteString("\n  q salir\n")
	return out.String()
}
