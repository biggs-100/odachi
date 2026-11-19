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
	"github.com/biggs-100/odachi/internal/domain/profile"
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
	// Axes is the profile in force, one configurable field per row: the editor
	// renders what the file says today, never a second copy (0336).
	Axes []profile.Axis
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
		_, _ = fmt.Fprintln(stderr, "odachi: tui takes no flags: odachi tui")
		return start.CodeBlocked
	}

	if !isTerminal(stdin) || streamTarget(stdout) == nil {
		_, _ = fmt.Fprintln(stderr, "odachi: la tui necesita una terminal; configura con odachi setup (o odachi init) y sus flags")
		return start.CodeBlocked
	}

	state, err := detectTUIState(dir)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	model := newTUIModel(state)
	model.run = newTUIRunner(dir)
	if _, err := tea.NewProgram(model, tea.WithInput(stdin), tea.WithOutput(stdout)).Run(); err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: la tui se cerró con un error: "+oneLine(err.Error()))
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
		state.Axes = profile.Axes(*voice)
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

// tuiStep is where the wizard is: the linear order, with the scope and the
// skills screens skipped when the choice they ask does not exist.
type tuiStep int

const (
	stepState tuiStep = iota
	stepScope
	stepClients
	stepSkills
	stepVoice
	stepReview
	stepSummary
	stepApply
	// stepProfile is the second section of the hub: the six axes of the profile in
	// force, which is what a person comes back to change (0336).
	stepProfile
)

// The options each single-choice screen offers, in the order a human reads them.
var (
	tuiClients = []string{persona.SelectionBoth, persona.SelectionOpenCode, persona.SelectionClaude}
	tuiSkills  = []string{persona.SkillsHome, persona.SkillsProject}
	tuiVoices  = voiceNames()
	tuiLenses  = review.AllLenses()

	// tuiMenu is the hub: the two sections the first screen offers. The menu is
	// what makes the wizard a place to come back to instead of a one-shot run (0336).
	tuiMenu = []string{"instalar y configurar", "perfil"}

	// tuiScopes is the choice a folder that is not a project had to be given: a
	// machine-only setup writes nothing here, init is what creates the project.
	tuiScopes = []string{"machine", "project"}
)

// tuiSteps is every screen the wizard has, in the order the table walks them.
var tuiSteps = []tuiStep{stepState, stepScope, stepClients, stepSkills, stepVoice, stepReview, stepSummary, stepApply, stepProfile}

// tuiRoute is where one screen goes: forward with enter, backward with esc, and
// whether the screen exists for this run at all. The table is the only place
// navigation is written, so a new screen is a row and a test can walk every row
// instead of trusting two switches to stay in step (0336).
type tuiRoute struct {
	forward tuiStep
	back    tuiStep
	when    func(tuiModel) bool
}

var tuiRoutes = map[tuiStep]tuiRoute{
	stepState: {forward: stepScope, back: stepState, when: always},
	// A project already carries the block: there is nothing to decide about the
	// scope, so the question is walked past in both directions.
	stepScope:   {forward: stepClients, back: stepState, when: func(m tuiModel) bool { return !m.state.Project }},
	stepClients: {forward: stepSkills, back: stepScope, when: always},
	// The skills screen belongs to init: a machine-only setup installs the craft
	// once per machine and has no project copy to choose.
	stepSkills:  {forward: stepVoice, back: stepClients, when: func(m tuiModel) bool { return m.init }},
	stepVoice:   {forward: stepReview, back: stepSkills, when: always},
	stepReview:  {forward: stepSummary, back: stepVoice, when: always},
	stepSummary: {forward: stepApply, back: stepReview, when: always},
	stepApply:   {forward: stepApply, back: stepSummary, when: always},
	// The profile section has one screen: esc is the way back to the menu.
	stepProfile: {forward: stepProfile, back: stepState, when: always},
}

// always is the row of a screen that exists for every run.
func always(tuiModel) bool { return true }

// forwardStep walks the table from the screen the wizard is on to the first one
// that exists for this run.
func (m tuiModel) forwardStep() tuiStep {
	next := tuiRoutes[m.step].forward
	for steps := 0; next != m.step && steps < len(tuiSteps); steps++ {
		if tuiRoutes[next].when(m) {
			return next
		}
		next = tuiRoutes[next].forward
	}
	return m.step
}

// backStep is forwardStep in the other direction.
func (m tuiModel) backStep() tuiStep {
	prev := tuiRoutes[m.step].back
	for steps := 0; prev != m.step && steps < len(tuiSteps); steps++ {
		if tuiRoutes[prev].when(m) {
			return prev
		}
		prev = tuiRoutes[prev].back
	}
	return m.step
}

// tuiModel is the wizard: the detected facts, the choices and where it stands.
type tuiModel struct {
	state        tuiState
	step         tuiStep
	init         bool // the chosen action: initialize the project, or the machine only
	scope        int
	client       int
	skills       int
	voice        int
	menu         int
	reviewCursor int
	// The profile section: which row the cursor is on (the axes, then the apply
	// row), the value picker it opened, the values the human chose but did not
	// write yet, and the line the last apply left (0336).
	profileCursor int
	picking       bool
	pickCursor    int
	pending       map[string]string
	fieldMsg      string
	editing       bool
	buffer        string
	editErr       string
	run           configure.Runner
	applied       bool
	result        configure.Result
}

// newTUIModel is the model a run starts with, positioned on what the project
// already chose.
func newTUIModel(state tuiState) tuiModel {
	if state.Models == nil {
		state.Models = map[string]string{}
	}
	// A caller that did not read the file -a test, and the first day- still sees the
	// axes of the profile it named: the screen never renders an empty list (0336).
	if len(state.Axes) == 0 && state.Voice != "" {
		if resolved, err := profile.Resolve(profile.Builtin(), state.Voice); err == nil {
			state.Axes = profile.Axes(resolved)
		}
	}
	return tuiModel{
		state:   state,
		init:    state.Project,
		client:  indexOf(tuiClients, state.Clients),
		skills:  indexOf(tuiSkills, state.Skills),
		voice:   indexOf(tuiVoices, state.Voice),
		pending: map[string]string{},
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
		if m.step == stepProfile {
			return m.profileKey(msg)
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc":
			m = m.back()
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
	}
	return m, nil
}

// next advances one screen, committing the choice the screen was showing. The
// step it lands on is the table's answer, not a chain of ifs (0336).
func (m tuiModel) next() (tea.Model, tea.Cmd) {
	m = m.commit()
	// The menu is the one screen whose enter picks its destination: that is what a
	// menu is, and the table still says where each section starts (0336).
	if m.step == stepState && m.menu == 1 {
		m.step = stepProfile
		return m, nil
	}
	m.step = m.forwardStep()
	return m, nil
}

// commit records the choice the screen is showing. Nothing here touches the disk.
func (m tuiModel) commit() tuiModel {
	switch m.step {
	case stepScope:
		m.init = tuiScopes[m.scope] == "project"
	case stepClients:
		m.state.Clients = tuiClients[m.client]
	case stepSkills:
		m.state.Skills = tuiSkills[m.skills]
	case stepVoice:
		m.state.Voice = tuiVoices[m.voice]
	}
	return m
}

// back is one screen towards the first one, skipping the screens this run does
// not have.
func (m tuiModel) back() tuiModel {
	if m.step == stepState {
		return m
	}
	m.step = m.backStep()
	return m
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
			m.editErr = "el modelo no es un identificador: usa provider/name#variant"
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
	case stepState:
		m.menu = clamp(m.menu+delta, len(tuiMenu))
	case stepScope:
		m.scope = clamp(m.scope+delta, len(tuiScopes))
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
	case stepScope:
		return m.viewChoices("Qué configurar", scopeLabels(), m.scope)
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
	case stepProfile:
		return m.viewProfile()
	default:
		return m.viewState()
	}
}

// viewState is the hub: where the project is, and the two sections a person can
// open. It is the screen the wizard comes back to, which is what makes a change of
// mind possible at all (0336).
func (m tuiModel) viewState() string {
	var out strings.Builder
	out.WriteString(m.header("instalación y configuración"))
	out.WriteString(m.facts())
	out.WriteString("\n")
	for i, item := range tuiMenu {
		marker := "  "
		if i == m.menu {
			marker = "> "
		}
		out.WriteString("  " + marker + item + "\n")
	}
	out.WriteString("\n  ↑↓ elegir · enter abrir · q salir\n")
	return out.String()
}

// viewProfile renders the six axes of the profile in force: what each one says
// today, the values it may take, and the row that writes them. The section is a
// map of the file, never a second copy: an empty value means the profile inherits
// it, and a value chosen here is a choice until it is applied (0336).
func (m tuiModel) viewProfile() string {
	var out strings.Builder
	out.WriteString(m.header("perfil " + m.state.Voice))
	for i, axis := range m.state.Axes {
		marker := "  "
		if i == m.profileCursor && !m.picking {
			marker = "> "
		}
		out.WriteString("  " + marker + fmt.Sprintf("%-12s %-12s %s\n",
			axis.Name, shownAxis(m.value(axis)), strings.Join(axis.Allowed, " ")))
		if m.picking && i == m.profileCursor {
			for j, allowed := range axis.Allowed {
				pick := "   "
				if j == m.pickCursor {
					pick = " > "
				}
				out.WriteString("    " + pick + allowed + "\n")
			}
		}
	}
	marker := "  "
	if m.profileCursor == len(m.state.Axes) {
		marker = "> "
	}
	out.WriteString("  " + marker + "aplicar los cambios\n")
	if m.fieldMsg != "" {
		out.WriteString("\n  " + m.fieldMsg + "\n")
	}
	out.WriteString("\n  ↑↓ elegir · enter cambiar · esc volver · q salir\n")
	return out.String()
}

// profileKey drives the editor: the rows, the value picker they open, and the row
// that writes what the human chose. Esc leaves the section and the file alone, so
// nothing reaches the disk before that last row (0336).
func (m tuiModel) profileKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	if m.picking {
		switch msg.String() {
		case "esc":
			m.picking = false
		case "up", "k":
			m.pickCursor = clamp(m.pickCursor-1, len(m.state.Axes[m.profileCursor].Allowed))
		case "down", "j":
			m.pickCursor = clamp(m.pickCursor+1, len(m.state.Axes[m.profileCursor].Allowed))
		case "enter":
			axis := m.state.Axes[m.profileCursor]
			m.pending[axis.Name] = axis.Allowed[m.pickCursor]
			m.picking = false
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.step = m.backStep()
	case "up", "k":
		m.profileCursor = clamp(m.profileCursor-1, len(m.state.Axes)+1)
	case "down", "j":
		m.profileCursor = clamp(m.profileCursor+1, len(m.state.Axes)+1)
	case "enter":
		if m.profileCursor == len(m.state.Axes) {
			return m.applyProfile()
		}
		m.picking = true
		m.pickCursor = indexOf(m.state.Axes[m.profileCursor].Allowed, m.value(m.state.Axes[m.profileCursor]))
	}
	return m, nil
}

// value is what a row shows: the value chosen in this run, or what the file says.
func (m tuiModel) value(axis profile.Axis) string {
	if chosen, ok := m.pending[axis.Name]; ok {
		return chosen
	}
	return axis.Value
}

// applyProfile writes the axes the human changed, one at a time, and stops at the
// first failure: a half-written profile is reported, never hidden (0336).
func (m tuiModel) applyProfile() (tea.Model, tea.Cmd) {
	if len(m.pending) == 0 {
		m.fieldMsg = "no hay cambios que aplicar"
		return m, nil
	}
	if m.run == nil {
		return m, nil
	}

	written := make([]string, 0, len(m.pending))
	for _, axis := range m.state.Axes {
		value, ok := m.pending[axis.Name]
		if !ok {
			continue
		}
		if err := m.run.Field(axis.Name, value); err != nil {
			m.fieldMsg = "falló " + axis.Name + ": " + err.Error()
			return m, nil
		}
		written = append(written, axis.Name+"="+value)
		m.state.Axes = withAxisValue(m.state.Axes, axis.Name, value)
	}
	m.pending = map[string]string{}
	m.fieldMsg = "aplicado: " + strings.Join(written, ", ")
	return m, nil
}

// withAxisValue records what was written, so the screen shows the file it just
// produced instead of the one it read at the start.
func withAxisValue(axes []profile.Axis, name, value string) []profile.Axis {
	for i := range axes {
		if axes[i].Name == name {
			axes[i].Value = value
		}
	}
	return axes
}

// shownAxis names an empty value the way the file means it: inherited.
func shownAxis(value string) string {
	if value == "" {
		return "(hereda)"
	}
	return value
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
		Project:    m.init,
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

	action := "solo la máquina (setup); la elección no se guarda sin un proyecto"
	skills := "home (la máquina)"
	if m.init {
		action = "inicializar el proyecto (init)"
		skills = m.state.Skills
	}

	rows := [][2]string{
		{"Proyecto", project},
		{"Configurar", action},
		{"Clientes", m.state.Clients},
		{"Skills", skills},
		{"Voz", m.state.Voice},
		{"Review", m.state.Review},
		{"Modelos", modelsLine(m.state.Models)},
	}

	var out strings.Builder
	for _, row := range rows {
		_, _ = fmt.Fprintf(&out, "  %-10s %s\n", row[0], row[1])
	}
	return out.String()
}

// voiceNames is the profiles the binary ships, read from the one table the
// validation reads. The screen does not carry a second copy of the list.
func voiceNames() []string {
	names := make([]string, 0, len(profile.Builtin()))
	for name := range profile.Builtin() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// scopeLabels names the two actions a folder that is not a project can take.
func scopeLabels() []string {
	return []string{"configurar la máquina (setup)", "inicializar el proyecto (init)"}
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
