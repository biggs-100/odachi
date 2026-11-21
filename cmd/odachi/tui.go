package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/biggs-100/odachi/internal/app/configure"
	initapp "github.com/biggs-100/odachi/internal/app/init"
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
	// Verify is the model odd/review.toml declares for the verifier, and
	// ClientModels the model each client's own file carries today. The models
	// section shows them and remembers them as the detected value, so a row that
	// is not touched is never a write (0337).
	Verify       string
	ClientModels map[string]string
	// ClientModelErrs is the reason a client's file could not be read, by client.
	// The row says so instead of showing an empty value it cannot vouch for (0337).
	ClientModelErrs map[string]string
	// Axes is the profile in force, one configurable field per row: the editor
	// renders what the file says today, never a second copy (0336).
	Axes []profile.Axis
	// The four scalars odd/review.toml carries, read where the file was read so
	// the limits section shows the values in force and a row kept as it was is a
	// no-op. Zero is a real value for three of them, so the section carries the
	// numbers read, never a second copy of the decision (0342).
	MaxPromptKB  int
	KeepVerdicts int
	KeepPrompts  int
	CheckTimeout int
	// Checks is the declared tools odd/review.toml carries, read where the file
	// was read so the checks section edits the list the project owns. The severity
	// each one carries is shown but never offered for edit: it fixes the verdict
	// and stays the file's decision (0346).
	Checks []review.Check
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
	model.dir = dir
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
	state.Verify = config.Verify
	state.MaxPromptKB = config.MaxPromptKB
	state.KeepVerdicts = config.KeepVerdicts
	state.KeepPrompts = config.KeepPrompts
	state.CheckTimeout = config.CheckTimeout
	state.Checks = config.Checks

	// The clients' models live in each client's own file, so the section reads
	// them where the verbs write them: one reader beside the writers, never a
	// second copy of the choice (0337).
	home, err := osHome{}.Home()
	if err != nil {
		return tuiState{}, fmt.Errorf("could not find the home to read the clients' models: %w", err)
	}
	clientModels, unreadable, err := initapp.ReadClientModels(home, disk(dir))
	if err != nil {
		return tuiState{}, err
	}
	state.ClientModels = clientModels
	// A file this reader cannot parse is a row that says so, not a verb that
	// refuses to open: the wizard configures three sections and only one of them
	// depends on that file (0337).
	state.ClientModelErrs = unreadable
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
	// stepModels is the third section: the model the verifier runs with and the
	// model each installed client reads, which is the choice that decides who
	// answers (0337).
	stepModels
	// stepLimits is the fourth section: the four scalars of odd/review.toml that
	// decide the prompt budget, the two retention lifetimes and the check
	// ceiling (0342).
	stepLimits
	// stepChecks is the fifth section: the declared tools of odd/review.toml. A
	// person can edit each one's name, command and remedy, but not its severity,
	// which fixes the verdict and stays the file's decision (0346).
	stepChecks
)

// The options each single-choice screen offers, in the order a human reads them.
var (
	tuiClients = []string{persona.SelectionBoth, persona.SelectionOpenCode, persona.SelectionClaude}
	tuiSkills  = []string{persona.SkillsHome, persona.SkillsProject}
	tuiVoices  = voiceNames()
	tuiLenses  = review.AllLenses()

	// tuiMenu is the hub: the sections the first screen offers, each with the
	// screen it opens. The install section opens on the hub itself and walks the
	// table from there; the others are screens whose route comes back here
	// (0336, 0337, 0342, 0346).
	tuiMenu = []tuiSection{
		{label: "instalar y configurar", step: stepState},
		{label: "perfil", step: stepProfile},
		{label: "modelos", step: stepModels},
		{label: "límites", step: stepLimits},
		{label: "checks", step: stepChecks},
	}

	// tuiScopes is the choice a folder that is not a project had to be given: a
	// machine-only setup writes nothing here, init is what creates the project.
	tuiScopes = []string{"machine", "project"}
)

// tuiSection is one row of the hub: what a human reads and where enter goes.
type tuiSection struct {
	label string
	step  tuiStep
}

// modelVerifier is the key of the verifier's row in the models section. The
// clients' rows are keyed by the name the client table and --model use, so the
// screen and the verb cannot spell a client two ways (0337).
const modelVerifier = "verifier"

// modelRefusal is the one sentence both editors give: the review screen and the
// models section check the same rule, and a second sentence is how two screens
// start disagreeing (0337).
const modelRefusal = "el modelo no es un identificador: usa provider/name#variant"

// clientModelRefusal answers the one empty value a client's row cannot keep while
// its file names a model: no verb removes that key, so the row has to say where
// the line lives (0337).
const clientModelRefusal = "un cliente no se queda sin modelo desde aquí: escribe uno, o borra la clave model de su archivo"

// tuiSteps is every screen the wizard has, in the order the table walks them.
var tuiSteps = []tuiStep{stepState, stepScope, stepClients, stepSkills, stepVoice, stepReview, stepSummary, stepApply, stepProfile, stepModels, stepLimits, stepChecks}

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
	// The models section is the same shape: one screen, and esc is the menu.
	stepModels: {forward: stepModels, back: stepState, when: always},
	// The limits section is the fourth screen of that shape (0342).
	stepLimits: {forward: stepLimits, back: stepState, when: always},
	// The checks section is the fifth, and esc is the menu (0346).
	stepChecks: {forward: stepChecks, back: stepState, when: always},
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

// tuiRun is what the wizard calls: the operations it applies, plus the host's
// model catalog the picker shows. It embeds configure.Runner instead of growing
// that interface, because the app that applies operations must not gain a read it
// never uses -a verb that only writes has no reason to spawn `opencode models`- and
// the fake in a test implements the same two halves (0338).
type tuiRun interface {
	configure.Runner
	// Catalog answers the models the host serves, and false when it cannot say.
	Catalog() ([]string, bool)
	// Limits writes the four scalars of the limits section. Only the fields the
	// human changed are set (0342).
	Limits(choice limitChoice) error
	// Checks writes the declared checks the human edited. Each edit carries the
	// identity read from the file and the three editable fields; severity is not
	// among them and the app keeps the file's (0346).
	Checks(edits []checkChoice) error
}

// limitChoice is one write of the limits section: a pointer per scalar, nil for
// the three the human left alone. It travels to the review app unchanged so the
// app decides how each value is normalized and what stays preserved (0342).
type limitChoice struct {
	MaxPromptKB  *int
	KeepVerdicts *int
	KeepPrompts  *int
	CheckTimeout *int
}

// checkChoice is one check the wizard edited. Identity is the name read from the
// file, so a rename still locates its row; the other three are the new values.
// There is no severity field on purpose: a person cannot change the verdict from
// here (0346).
type checkChoice struct {
	Identity string
	Name     string
	Command  string
	Remedy   string
}

// The three editable fields of a check, in the order the editor walks them. The
// severity is deliberately absent from this list: it is shown on the row but no
// key opens it (0346).
const (
	checkFieldName = iota
	checkFieldCommand
	checkFieldRemedy
	checkFieldCount
)

// checkFieldLabel is the name a person reads for an editable field. The labels
// fit the row's gutter the same way the other nested lists do (0346).
func checkFieldLabel(field int) string {
	switch field {
	case checkFieldName:
		return "nombre"
	case checkFieldCommand:
		return "comando"
	case checkFieldRemedy:
		return "remedio"
	}
	return ""
}

// The four rows of the limits section, keyed by the TOML key they write. The
// keys are the ones a person reads in odd/review.toml, so a row and the line it
// edits cannot drift apart (0342).
const (
	limitMaxPromptKB  = "max_prompt_kb"
	limitKeepVerdicts = "keep_verdicts"
	limitKeepPrompts  = "keep_prompts"
	limitCheckTimeout = "check_timeout_s"
)

// limitRefusal is the sentence a value the domain refuses earns. A number that
// cannot be parsed is the other refusal, and both keep the editor open (0342).
const limitNumberRefusal = "el valor tiene que ser un número entero"

// catalogState is what the wizard knows about the host's list: nothing yet, one
// probe in flight, the answer, or the fact that the host could not answer. It is
// a state and not a pair of flags because "still asking" and "asked and unable"
// are the two a screen has to tell apart, and only one of them is worth a
// sentence (0338).
type catalogState int

const (
	catalogIdle catalogState = iota
	catalogAsking
	catalogReady
	catalogFailed
)

// catalogMsg is the host's answer as the tea runtime delivers it, off the key
// loop: the screen never blocks on a process (0338).
type catalogMsg struct {
	models []string
	ok     bool
}

// The two sentences the picker adds to the host's entries: the row that opens the
// line editor, and the state the screen shows while the probe runs (0338).
const (
	manualModelRow   = "escribir a mano"
	detectingCatalog = "detectando…"
)

// The sentences a probe that produced no list earns. They are two because they are
// two facts: one host could not be asked, the other answered that there is nothing
// here, and a screen that used one sentence for both would be guessing (0338).
const (
	catalogFailedNote = "no se pudo leer la lista de modelos del host: escribe el id a mano, o comprueba que el host responde"
	catalogEmptyNote  = "el host no nombró ningún modelo aquí: escribe el id a mano"
)

// claudeModelAliases are the model aliases Claude Code documents for --model. The
// probe answers OpenCode's catalogue, so offering that list on a Claude row would
// name models this client does not serve. The three names Claude Code accepts were
// checked against its CLI reference in 2026-11; a slice cannot be a constant in Go,
// so this is the one declaration to update when the client renames one (0338).
var claudeModelAliases = []string{"sonnet", "opus", "haiku"}

// claudeAliasNote explains a list that is not the host's: these are short names,
// and the last row of the picker is where the full identifier is written (0338).
const claudeAliasNote = `son alias de Claude Code: el id completo está en "escribir a mano"`

// tuiModel is the wizard: the detected facts, the choices and where it stands.
type tuiModel struct {
	state tuiState
	// dir is the folder the wizard configures. It is the context the header
	// names under the title; empty means nothing to name and the line is left
	// out, which keeps the screens readable in a test (0338).
	dir          string
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
	// The models section: which row the cursor is on, and the value the human
	// chose but did not write yet, keyed by row. A row that is not here was never
	// changed, which is what keeps the plan (and the disk) empty (0337).
	modelsCursor int
	modelPending map[string]string
	// The limits section: which row the cursor is on and the scalar the human
	// chose but did not write yet, keyed by row. A row that is not here was never
	// changed, which is what keeps the write (and the checks array) alone (0342).
	limitsCursor  int
	limitsPending map[string]int
	// The checks section: which check row the cursor is on, whether its three
	// fields are open, which field owns the buffer, and the edited values keyed
	// by the identity read from the file. A check that was never opened has no
	// pending entry, which is what keeps the write empty (0346).
	checksCursor  int
	checkOpen     bool
	checkField    int
	checksPending map[string]checkChoice
	checkNotice   string
	// The host's catalog and the picker that shows it: the state of the one probe
	// the run makes, the entries it answered, whether the list is open, and which
	// entry its cursor is on. Both screens that edit a model -the review lens table
	// and the models section- share this one picker (0338).
	probe           catalogState
	catalog         []string
	pickingModel    bool
	modelPickCursor int
	// edited is every axis the editor wrote in this run, by name: the voice op
	// recomputes the whole profile from the builtin, so these are re-applied after
	// it and an explicit edit is the last word (0336).
	edited  map[string]string
	editing bool
	buffer  string
	editErr string
	run     tuiRun
	applied bool
	result  configure.Result
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
		state:         state,
		init:          state.Project,
		client:        indexOf(tuiClients, state.Clients),
		skills:        indexOf(tuiSkills, state.Skills),
		voice:         indexOf(tuiVoices, state.Voice),
		pending:       map[string]string{},
		limitsPending: map[string]int{},
		checksPending: map[string]checkChoice{},
		edited:        map[string]string{},
	}
}

// Init does nothing: the first screen is already the state.
func (m tuiModel) Init() tea.Cmd { return nil }

// Update handles the keys the wizard answers. Nothing here touches the disk: the
// choices are the model's until the apply screen runs them.
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case catalogMsg:
		m.catalog = msg.models
		if msg.ok {
			// An answer with no entries is still an answer: the host looked and
			// named nothing here, which is a different fact from not being asked
			// and earns its own sentence (0338).
			m.probe = catalogReady
		} else {
			// Remembered as unable, not forgotten: a second visit finds the fact
			// instead of spawning the same process on every key (0338).
			m.probe = catalogFailed
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.pickingModel {
			return m.modelPickKey(msg)
		}
		if m.editing {
			return m.editKey(msg), nil
		}
		if m.step == stepProfile {
			return m.profileKey(msg)
		}
		if m.step == stepModels {
			return m.modelsKey(msg)
		}
		if m.step == stepLimits {
			return m.limitsKey(msg)
		}
		if m.step == stepChecks {
			return m.checksKey(msg)
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
	// menu is, and the table still says where each section starts (0336). The
	// row itself names the screen, so a fourth section is data and not a branch
	// (0337, 0342).
	if m.step == stepState && m.menu > 0 {
		section := tuiMenu[m.menu]
		m.step = section.step
		// The hub applies the voice in one section and edits the profile in another:
		// the rows are re-read on entry so they describe the file that stands (0336).
		if section.step == stepProfile && m.run != nil {
			if axes, err := m.run.Axes(); err == nil && len(axes) > 0 {
				m.state.Axes = axes
			}
		}
		return m.withCatalog()
	}
	m.step = m.forwardStep()
	return m.withCatalog()
}

// withCatalog fires the host's catalog probe when the wizard lands on a screen
// that picks a model, and only once per run. The answer is a command, so the
// screen is readable while the process runs; a second visit finds the state the
// first one left instead of asking the host again (0338).
func (m tuiModel) withCatalog() (tea.Model, tea.Cmd) {
	if m.step != stepReview && m.step != stepModels {
		return m, nil
	}
	cmd := m.detectModels()
	if cmd != nil {
		m.probe = catalogAsking
	}
	return m, cmd
}

// detectModels is the one place the wizard asks the host what it serves. It
// returns nil when the question was already asked, which is what keeps a run to
// one process: idle is the only state that asks (0338).
func (m tuiModel) detectModels() tea.Cmd {
	if m.probe != catalogIdle || m.run == nil {
		return nil
	}
	run := m.run
	return func() tea.Msg {
		models, ok := run.Catalog()
		return catalogMsg{models: models, ok: ok}
	}
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
// cursor is on: the host's list when the probe answered one, the line editor
// otherwise (0338).
func (m tuiModel) enterReview() (tea.Model, tea.Cmd) {
	if m.reviewCursor == 0 {
		if m.state.Review == review.ModeOn {
			m.state.Review = review.ModeOff
		} else {
			m.state.Review = review.ModeOn
		}
		return m, nil
	}
	return m.openModelRow()
}

// openModelRow opens what a model row offers: the list the row shows when there is
// one to pick from, and the line editor otherwise. Typing is not removed by the
// picker, it is moved behind the list's last row (0338).
func (m tuiModel) openModelRow() (tea.Model, tea.Cmd) {
	if list, _ := m.modelList(m.pickerRow()); len(list) > 0 {
		value := m.pickedValue()
		m.pickingModel = true
		m.modelPickCursor = indexOf(m.modelChoices(value), value)
		m.editErr = ""
		return m, nil
	}
	return m.openModelEditor(), nil
}

// openModelEditor is the door that stays: the buffer the screen has always had,
// opened on the value in force, checked on enter by the rule of the row (0337).
func (m tuiModel) openModelEditor() tuiModel {
	m.editing = true
	m.buffer = m.pickedValue()
	m.editErr = ""
	return m
}

// modelCatalog is the host's list when there is one to pick from. An empty or
// failed probe is no list: the picker is not offered and the row keeps the line
// editor as its only door (0338).
func (m tuiModel) modelCatalog() []string {
	if m.probe != catalogReady || len(m.catalog) == 0 {
		return nil
	}
	return m.catalog
}

// pickerRow is the row the picker or the editor was opened on. The models section
// names it by the cursor; the review screen's rows are lenses, which always show
// the host's list, so the empty row says exactly that (0338).
func (m tuiModel) pickerRow() string {
	if m.step == stepModels {
		return m.modelRows()[m.modelsIndex()]
	}
	return ""
}

// modelList answers which list a row shows: the host's catalogue for the verifier,
// the review lenses and the OpenCode client -the host the probe asks is the one
// that serves them-, and Claude Code's declared aliases for a Claude client, which
// the host does not serve. The rule lives here and nowhere else, so a screen never
// branches on which client a row belongs to (0338).
//
// The second answer is the sentence that explains a list that is not the host's; a
// host list has none, and the picker's count line says how long it is.
func (m tuiModel) modelList(row string) ([]string, string) {
	if row == persona.SelectionClaude {
		return claudeModelAliases, claudeAliasNote
	}
	return m.modelCatalog(), ""
}

// catalogNotice is what a model screen says about the probe under its rows: the
// one process still running, the host that could not be asked, or the answer that
// named nothing. A list with entries explains itself in the picker's count, so
// there is nothing to say (0338).
func (m tuiModel) catalogNotice() string {
	switch m.probe {
	case catalogAsking:
		return tuiLook.mutedLine(detectingCatalog)
	case catalogFailed:
		return tuiLook.warningLine(catalogFailedNote)
	case catalogReady:
		if len(m.catalog) == 0 {
			return tuiLook.warningLine(catalogEmptyNote)
		}
	}
	return ""
}

// modelChoices is what the picker shows for the value in force: the row's entries,
// the value first when the list does not carry it -a project may name a model that
// was renamed or retired, and hiding it would invite a change nobody asked for-,
// and the manual door last (0338).
func (m tuiModel) modelChoices(value string) []string {
	list, _ := m.modelList(m.pickerRow())
	choices := make([]string, 0, len(list)+2)
	if !slices.Contains(list, value) {
		choices = append(choices, value)
	}
	choices = append(choices, list...)
	return append(choices, manualModelRow)
}

// pickedValue is the value in force of the row the picker was opened on. The two
// screens share one picker, so the screen says which row it is and the picker only
// reads it (0337, 0338).
func (m tuiModel) pickedValue() string {
	if m.step == stepModels {
		return m.modelValue(m.modelRows()[m.modelsIndex()])
	}
	return m.state.Models[tuiLenses[m.reviewCursor-1]]
}

// modelPickKey drives the host's list: the cursor, enter keeping the entry, and
// esc dropping the whole thing -nothing is written before the apply (0338).
func (m tuiModel) modelPickKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	choices := m.modelChoices(m.pickedValue())
	switch msg.String() {
	case "esc":
		m.pickingModel = false
	case "up", "k":
		m.modelPickCursor = clamp(m.modelPickCursor-1, len(choices))
	case "down", "j":
		m.modelPickCursor = clamp(m.modelPickCursor+1, len(choices))
	case "enter":
		choice := choices[clamp(m.modelPickCursor, len(choices))]
		if choice == manualModelRow {
			m.pickingModel = false
			return m.openModelEditor(), nil
		}
		m = m.keepPickedModel(choice)
		// A refused choice -a client's row kept empty- leaves the list open with
		// the sentence on the screen, so the refusal is not swallowed (0338).
		if m.editErr == "" {
			m.pickingModel = false
		}
	}
	return m, nil
}

// keepPickedModel records a value chosen from the host's list. The models
// section ends in the rule its editor already applied; the review screen writes
// the lens table directly, which is what the editor did before there was a list
// (0337, 0338).
func (m tuiModel) keepPickedModel(value string) tuiModel {
	if m.step == stepModels {
		return m.keepModelValue(m.modelRows()[m.modelsIndex()], value)
	}
	return m.keepLens(value)
}

// modelCount says how many entries the host answered, so a person knows the list
// is the agent's and not the screen's invention (0338).
func modelCount(n int) string {
	if n == 1 {
		return "1 modelo"
	}
	return fmt.Sprintf("%d modelos", n)
}

// modelPicker renders the list under the row that opened it: the window every
// other list uses, the entries in the nested row shape, and the line that explains
// the list. A host list carries its count in the muted voice; a list that is not
// the host's carries the sentence that says what it is. Either line is furniture,
// not a row: it is not selectable (0338).
func (m tuiModel) modelPicker(value string) string {
	list, note := m.modelList(m.pickerRow())
	choices := m.modelChoices(value)
	var out strings.Builder
	out.WriteString(tuiLook.windowLines(windowFor(m.modelPickCursor, len(choices), modelPickRows), func(i int) string {
		return tuiLook.subRow(i == m.modelPickCursor, shownModel(choices[i]))
	}))
	if note != "" {
		out.WriteString(tuiLook.mutedLine(note))
	} else {
		out.WriteString(tuiLook.mutedLine(modelCount(len(list))))
	}
	return out.String()
}

// editKey edits the row the editor was opened on. The buffer keys are one for
// both editors; what the value means belongs to the screen, so enter hands it to
// the keeper of that screen (0336, 0337).
func (m tuiModel) editKey(msg tea.KeyMsg) tuiModel {
	switch msg.Type {
	case tea.KeyEsc:
		m.editing = false
		m.buffer = ""
		m.editErr = ""
	case tea.KeyEnter:
		if m.step == stepModels {
			return m.keepModel()
		}
		if m.step == stepLimits {
			return m.keepLimit()
		}
		if m.step == stepChecks {
			return m.keepCheckField()
		}
		return m.keepLens(m.buffer)
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

// keepModel records the value of the models row the editor was opened on. It is
// the lens screen's rule, checked by the same function and refused with the same
// sentence; an empty value is the verifier's row meaning "no model", which the
// plan reads as a line to remove.
//
// A client's row is not the verifier's: no verb removes that key, and a screen
// that showed the row empty while the file kept its model would be lying about
// what the run wrote. It refuses the empty value and names where the line lives.
func (m tuiModel) keepModel() tuiModel {
	return m.keepModelValue(m.modelRows()[m.modelsIndex()], strings.TrimSpace(m.buffer))
}

// keepModelValue is the one place a models row's choice is accepted: the editor
// and the picker both end here, so a value typed by hand and a value chosen from
// the host's list cannot disagree about what a row may keep (0338).
func (m tuiModel) keepModelValue(row, value string) tuiModel {
	value = strings.TrimSpace(value)
	// Only a row whose file carries a model is refusing to lose it: a row that
	// already has none has nothing to lie about, and keeping it empty is a no-op
	// (0337).
	if value == "" && row != modelVerifier && m.state.ClientModels[row] != "" {
		m.editErr = clientModelRefusal
		return m
	}
	if value != "" {
		checked, err := review.NormalizeModel(value)
		if err != nil {
			m.editErr = modelRefusal
			return m
		}
		value = checked
	}
	if m.modelPending == nil {
		m.modelPending = map[string]string{}
	}
	m.modelPending[row] = value
	m.editing = false
	m.buffer = ""
	m.editErr = ""
	return m
}

// keepLens records the review screen's model for the lens the editor or the
// picker was opened on. An empty value removes the line, which is the only way a
// lens goes back to the session's own model (0337).
func (m tuiModel) keepLens(value string) tuiModel {
	lens := tuiLenses[m.reviewCursor-1]
	value = strings.TrimSpace(value)
	if value == "" {
		delete(m.state.Models, lens)
		m.editing = false
		m.buffer = ""
		m.editErr = ""
		return m
	}
	checked, err := review.NormalizeModel(value)
	if err != nil {
		m.editErr = modelRefusal
		return m
	}
	m.state.Models[lens] = checked
	m.editing = false
	m.buffer = ""
	m.editErr = ""
	return m
}

// modelRows is the section's rows: the verifier and each client the hub already
// detected, in the client table's order. The rows are the detection, never a
// second copy of which clients this machine runs (0337).
func (m tuiModel) modelRows() []string {
	rows := []string{modelVerifier}
	_, clients, err := persona.ResolveSelection(m.state.Clients)
	if err != nil {
		return rows
	}
	for _, client := range clients {
		rows = append(rows, client.Agent)
	}
	return rows
}

// modelsIndex is the row the cursor is on, clamped to the rows the section shows
// now: the client choice lives in another section, and a row that is gone must
// not be the one the next enter edits (0337).
func (m tuiModel) modelsIndex() int {
	return clamp(m.modelsCursor, len(m.modelRows()))
}

// modelValue is what a row shows: the value chosen in this run, or the one the
// file carries today.
func (m tuiModel) modelValue(row string) string {
	if chosen, ok := m.modelPending[row]; ok {
		return chosen
	}
	if row == modelVerifier {
		return m.state.Verify
	}
	return m.state.ClientModels[row]
}

// shownModel names an empty row the way the file means it: no model at all.
func shownModel(value string) string {
	if value == "" {
		return "(ninguno)"
	}
	return value
}

// modelUnreadable is the reason a client's file could not be read, when the row
// was not edited in this run. The row says it cannot vouch for the value instead
// of showing an empty one, and the reason sits under the list (0337).
func (m tuiModel) modelUnreadable(row string) string {
	reason, bad := m.state.ClientModelErrs[row]
	if !bad {
		return ""
	}
	if _, edited := m.modelPending[row]; edited {
		return ""
	}
	return reason
}

// modelLabel is the name a human reads for a row.
func modelLabel(row string) string {
	switch row {
	case modelVerifier:
		return "verificador"
	case persona.SelectionOpenCode:
		return "OpenCode"
	case persona.SelectionClaude:
		return "Claude"
	}
	return row
}

// viewModels renders the models section: who answers, and who answers each
// client. The values are the file's, and a row that was edited shows the choice
// until it is applied (0337).
func (m tuiModel) viewModels() string {
	var out strings.Builder
	out.WriteString(m.header("modelos"))
	rows := m.modelRows()
	cursor := clamp(m.modelsCursor, len(rows))
	out.WriteString(tuiLook.windowLines(windowFor(cursor, len(rows), modelsRows), func(i int) string {
		row := rows[i]
		value := shownModel(m.modelValue(row))
		notice := value == "(ninguno)"
		if m.modelUnreadable(row) != "" {
			value = "(ilegible)"
			notice = true
		}
		if m.editing && i == cursor {
			value = m.buffer + "_"
			notice = false
		}
		var line strings.Builder
		// One cursor at a time: while the picker is open under this row, the
		// nested list is what the keys move, so the row itself drops its glyph
		// (the profile editor already drew it this way) (0338).
		nested := m.pickingModel && i == cursor
		selected := i == cursor && !nested
		if notice {
			line.WriteString(tuiLook.rowNotice(selected, modelLabel(row), value))
		} else {
			line.WriteString(tuiLook.rowValues(selected, modelLabel(row), value))
		}
		if nested {
			line.WriteString(m.modelPicker(m.pickedValue()))
		}
		return line.String()
	}))
	out.WriteString(m.catalogNotice())
	if m.editErr != "" {
		out.WriteString("\n")
		out.WriteString(tuiLook.failureLine(m.editErr))
	}
	if reason := m.modelUnreadable(rows[cursor]); reason != "" {
		out.WriteString("\n")
		out.WriteString(tuiLook.warningLine(reason))
	}
	out.WriteString(tuiLook.helpLine(m.helpKeys("enter cambiar · esc volver · se aplican en el resumen")))
	return out.String()
}

// modelsKey drives the section: the rows, the picker or the editor they open, and
// nothing that touches the disk. What is chosen here is written only by the apply,
// and only when it differs from the file (0337, 0338).
func (m tuiModel) modelsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	switch msg.String() {
	case "esc":
		m.step = m.backStep()
	case "up", "k":
		m.modelsCursor = clamp(m.modelsCursor-1, len(m.modelRows()))
	case "down", "j":
		m.modelsCursor = clamp(m.modelsCursor+1, len(m.modelRows()))
	case "enter":
		return m.openModelRow()
	}
	return m, nil
}

// limitRows is the section's rows, in the order the file writes them: the prompt
// budget, the two retention lifetimes and the check ceiling. The [[check]] array
// is not among them -it fixes the gate's severity and belongs to the file (0342).
func limitRows() []string {
	return []string{limitMaxPromptKB, limitKeepVerdicts, limitKeepPrompts, limitCheckTimeout}
}

// limitLabel is the name a human reads for a row. It carries the unit so a value
// is never an unlabelled number, and it fits the twelve-cell label column the
// other sections use (0342).
func limitLabel(row string) string {
	switch row {
	case limitMaxPromptKB:
		return "prompt KB"
	case limitKeepVerdicts:
		return "veredictos"
	case limitKeepPrompts:
		return "prompts"
	case limitCheckTimeout:
		return "timeout s"
	}
	return row
}

// limitDetected is the value the file carried when the wizard read it.
func (m tuiModel) limitDetected(row string) int {
	switch row {
	case limitMaxPromptKB:
		return m.state.MaxPromptKB
	case limitKeepVerdicts:
		return m.state.KeepVerdicts
	case limitKeepPrompts:
		return m.state.KeepPrompts
	case limitCheckTimeout:
		return m.state.CheckTimeout
	}
	return 0
}

// limitValue is what a row shows: the value chosen in this run, or the one the
// file carries today.
func (m tuiModel) limitValue(row string) int {
	if chosen, ok := m.limitsPending[row]; ok {
		return chosen
	}
	return m.limitDetected(row)
}

// limitsIndex is the row the cursor is on, clamped to the four the section shows.
func (m tuiModel) limitsIndex() int {
	return clamp(m.limitsCursor, len(limitRows()))
}

// viewLimits renders the four scalars in the same dense, label-aligned shape as
// the models section. The values are the file's until a row is edited, and a
// refused value keeps its sentence on the screen (0342).
func (m tuiModel) viewLimits() string {
	var out strings.Builder
	out.WriteString(m.header("límites"))
	rows := limitRows()
	cursor := clamp(m.limitsCursor, len(rows))
	out.WriteString(tuiLook.windowLines(windowFor(cursor, len(rows), limitsRows), func(i int) string {
		row := rows[i]
		value := strconv.Itoa(m.limitValue(row))
		if m.editing && i == cursor {
			value = m.buffer + "_"
		}
		return tuiLook.rowValues(i == cursor, limitLabel(row), value)
	}))
	if m.editErr != "" {
		out.WriteString("\n")
		out.WriteString(tuiLook.failureLine(m.editErr))
	}
	out.WriteString(tuiLook.helpLine(m.helpKeys("enter cambiar · esc volver · se aplican en el resumen")))
	return out.String()
}

// limitsKey drives the section: the four rows and the number editor they open.
// Esc leaves with the file untouched; what is typed here is written only by the
// apply (0342).
func (m tuiModel) limitsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	switch msg.String() {
	case "esc":
		m.step = m.backStep()
	case "up", "k":
		m.limitsCursor = clamp(m.limitsCursor-1, len(limitRows()))
	case "down", "j":
		m.limitsCursor = clamp(m.limitsCursor+1, len(limitRows()))
	case "enter":
		// A number is replaced, not appended to: the row just showed the value in
		// force, so a fresh buffer makes one short edit enough to change it.
		m.editing = true
		m.buffer = ""
		m.editErr = ""
	}
	return m, nil
}

// keepLimit records the value typed into a limits row. The domain owns each
// scalar's rule, so the same normalizer the writer uses decides here what a row
// may keep; a refused value stays in the editor with its sentence (0342).
func (m tuiModel) keepLimit() tuiModel {
	row := limitRows()[m.limitsIndex()]
	value, err := strconv.Atoi(strings.TrimSpace(m.buffer))
	if err != nil {
		m.editErr = limitNumberRefusal
		return m
	}
	switch row {
	case limitMaxPromptKB:
		value, err = review.NormalizeMaxPromptKB(value)
	case limitKeepVerdicts:
		value, err = review.NormalizeKeepVerdicts(value)
	case limitKeepPrompts:
		value, err = review.NormalizeKeepPrompts(value)
	case limitCheckTimeout:
		value, err = review.NormalizeCheckTimeout(value)
	}
	if err != nil {
		m.editErr = "no se puede fijar " + limitLabel(row) + ": " + err.Error()
		return m
	}
	if m.limitsPending == nil {
		m.limitsPending = map[string]int{}
	}
	m.limitsPending[row] = value
	m.editing = false
	m.buffer = ""
	m.editErr = ""
	return m
}

// limitsChoice is the write the apply sends: one pointer per scalar whose chosen
// value differs from what the file carried. A section opened and left alone, or
// a row retyped with its own value, sends nothing at all (0342).
func (m tuiModel) limitsChoice() *limitChoice {
	var choice limitChoice
	changed := false
	for _, row := range limitRows() {
		chosen, ok := m.limitsPending[row]
		if !ok || chosen == m.limitDetected(row) {
			continue
		}
		value := chosen
		switch row {
		case limitMaxPromptKB:
			choice.MaxPromptKB = &value
		case limitKeepVerdicts:
			choice.KeepVerdicts = &value
		case limitKeepPrompts:
			choice.KeepPrompts = &value
		case limitCheckTimeout:
			choice.CheckTimeout = &value
		}
		changed = true
	}
	if !changed {
		return nil
	}
	return &choice
}

// detectedCheck is the check the cursor is on as the file declared it. The
// identity is its name: it stays the key even after the human renames the row
// (0346).
func (m tuiModel) detectedCheck() review.Check {
	return m.state.Checks[m.checksIndex()]
}

// checksIndex is the row the cursor is on, clamped to the declared checks.
func (m tuiModel) checksIndex() int {
	return clamp(m.checksCursor, len(m.state.Checks))
}

// checkEdit is the pending edit of a check. Until a field is kept, the values
// shown are the file's; keeping one records all three in this entry so the next
// field opens on the same edit. The identity never moves (0346).
func (m tuiModel) checkEdit(detected review.Check) checkChoice {
	edit, ok := m.checksPending[detected.Name]
	if !ok {
		edit = checkChoice{
			Identity: detected.Name,
			Name:     detected.Name,
			Command:  detected.Command,
			Remedy:   detected.Remedy,
		}
	}
	return edit
}

// checkFieldValue is the value a field shows in the nested list: the pending
// edit once one was kept, otherwise the value the file declared (0346).
func (m tuiModel) checkFieldValue(field int) string {
	edit := m.checkEdit(m.detectedCheck())
	switch field {
	case checkFieldName:
		return edit.Name
	case checkFieldCommand:
		return edit.Command
	case checkFieldRemedy:
		return edit.Remedy
	}
	return ""
}

// keepCheckField records the value typed into an open check's field. The domain
// owns the rule -the same NormalizeChecks a hand-written file obeys- so a blank
// name, command or remedy stays in the editor with the reason and no pending
// edit is written. Severity is copied from the file, never offered to the editor
// (0346).
func (m tuiModel) keepCheckField() tuiModel {
	detected := m.detectedCheck()
	edit := m.checkEdit(detected)
	value := strings.TrimSpace(m.buffer)
	switch m.checkField {
	case checkFieldName:
		edit.Name = value
	case checkFieldCommand:
		edit.Command = value
	case checkFieldRemedy:
		edit.Remedy = value
	}
	candidate := review.Check{
		Name:     edit.Name,
		Command:  edit.Command,
		Severity: detected.Severity,
		Remedy:   edit.Remedy,
	}
	if _, err := review.NormalizeChecks([]review.Check{candidate}); err != nil {
		m.editErr = oneLine(err.Error())
		return m
	}
	if m.checksPending == nil {
		m.checksPending = map[string]checkChoice{}
	}
	m.checksPending[detected.Name] = edit
	m.editing = false
	m.buffer = ""
	m.editErr = ""
	return m
}

// viewChecks renders one row per declared check, with the severity it keeps as a
// read-only value beside the name. Opening a row reveals its three editable
// fields as a nested list; the help line follows whichever control owns the keys
// (0346).
func (m tuiModel) viewChecks() string {
	var out strings.Builder
	out.WriteString(m.header("checks"))
	rows := m.state.Checks
	if len(rows) == 0 {
		out.WriteString(tuiLook.mutedLine("no hay checks declarados en odd/review.toml"))
	} else {
		cursor := m.checksIndex()
		out.WriteString(tuiLook.windowLines(windowFor(cursor, len(rows), checksRows), func(i int) string {
			check := rows[i]
			name := check.Name
			if edit, ok := m.checksPending[check.Name]; ok {
				name = edit.Name
			}
			return tuiLook.rowValues(i == cursor && !m.checkOpen, name, "severity "+string(check.Severity))
		}))
	}
	if m.checkOpen {
		out.WriteString(tuiLook.windowLines(windowFor(m.checkField, checkFieldCount, checkFieldCount), func(i int) string {
			value := m.checkFieldValue(i)
			if m.editing && i == m.checkField {
				value = m.buffer + "_"
			}
			return tuiLook.subRow(i == m.checkField && !m.editing, checkFieldLabel(i)+": "+value)
		}))
	}
	if m.editErr != "" {
		out.WriteString("\n")
		out.WriteString(tuiLook.failureLine(m.editErr))
	}
	out.WriteString(tuiLook.helpLine(m.helpKeys("enter editar · esc volver · se aplican en el resumen")))
	return out.String()
}

// checksKey drives the section: the check rows, the three fields of an open one
// and the buffer of the field being edited. Esc closes the innermost control
// first; only esc on the rows leaves the section, with the file untouched until
// the apply (0346).
func (m tuiModel) checksKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	if len(m.state.Checks) == 0 {
		if msg.String() == "esc" {
			m.step = m.backStep()
		}
		return m, nil
	}
	if m.checkOpen {
		switch msg.String() {
		case "esc":
			m.checkOpen = false
			m.checkField = 0
			m.editErr = ""
		case "up", "k":
			m.checkField = clamp(m.checkField-1, checkFieldCount)
		case "down", "j":
			m.checkField = clamp(m.checkField+1, checkFieldCount)
		case "enter":
			// A field is replaced, not appended to: the row already showed its
			// value, so a fresh buffer makes one short edit enough to change it.
			m.editing = true
			m.buffer = ""
			m.editErr = ""
		}
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.step = m.backStep()
	case "up", "k":
		m.checksCursor = clamp(m.checksCursor-1, len(m.state.Checks))
	case "down", "j":
		m.checksCursor = clamp(m.checksCursor+1, len(m.state.Checks))
	case "enter":
		// The first field opens with the file's value; a pending entry appears
		// only once a field is actually kept (0346).
		m.checkOpen = true
		m.checkField = 0
		m.editErr = ""
	}
	return m, nil
}

// checksChoice is the write the apply sends: one edit per declared check whose
// name, command or remedy differs from what the file carried. A section opened
// and left alone, or a field retyped with its own value, sends nothing (0346).
func (m tuiModel) checksChoice() []checkChoice {
	var edits []checkChoice
	for _, detected := range m.state.Checks {
		edit, ok := m.checksPending[detected.Name]
		if !ok || edit == (checkChoice{Identity: detected.Name, Name: detected.Name, Command: detected.Command, Remedy: detected.Remedy}) {
			continue
		}
		edits = append(edits, edit)
	}
	return edits
}

// checksNeedApproval reports whether any edit changed a field inside the
// fingerprint -the name or the command-. The remedy is advice and is left out of
// the hash, so rewording it alone asks nothing of the human (0346).
func (m tuiModel) checksNeedApproval(edits []checkChoice) bool {
	for _, edit := range edits {
		for _, detected := range m.state.Checks {
			if detected.Name == edit.Identity &&
				(edit.Name != detected.Name || edit.Command != detected.Command) {
				return true
			}
		}
	}
	return false
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

// helpKeys is the one decision that names the keys in the footer. While a picker
// or an editor owns the keyboard, its keys replace the section's: arrows move the
// nested list, typing fills the buffer, enter keeps the value and esc closes the
// control. With no control open, the section keeps its own keys. The open control
// is state of the machine, not of the screen, so profile, models and limits share
// this one helper instead of each repeating the conditional (0343).
func (m tuiModel) helpKeys(section string) string {
	switch {
	case m.picking || m.pickingModel:
		return "↑↓ elegir · enter cambiar · esc volver"
	case m.editing:
		return "escribir · enter cambiar · esc volver"
	case m.checkOpen:
		return "↑↓ campo · enter editar · esc cerrar"
	default:
		return section
	}
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
	case stepModels:
		return m.viewModels()
	case stepLimits:
		return m.viewLimits()
	case stepChecks:
		return m.viewChecks()
	default:
		return m.viewState()
	}
}

// viewState is the hub: where the project is, and the sections a person can
// open. It is the screen the wizard comes back to, which is what makes a change of
// mind possible at all (0336).
func (m tuiModel) viewState() string {
	var out strings.Builder
	out.WriteString(m.header("instalación y configuración"))
	out.WriteString(m.facts())
	out.WriteString("\n")
	out.WriteString(tuiLook.windowLines(windowFor(m.menu, len(tuiMenu), hubRows), func(i int) string {
		return tuiLook.row(i == m.menu, tuiMenu[i].label)
	}))
	out.WriteString(tuiLook.helpLine("↑↓ elegir · enter abrir · q salir"))
	return out.String()
}

// viewProfile renders the six axes of the profile in force: what each one says
// today, the values it may take, and the row that writes them. The section is a
// map of the file, never a second copy: an empty value means the profile inherits
// it, and a value chosen here is a choice until it is applied (0336).
func (m tuiModel) viewProfile() string {
	var out strings.Builder
	out.WriteString(m.header("perfil " + m.state.Voice))
	axes := len(m.state.Axes)
	out.WriteString(tuiLook.windowLines(windowFor(m.profileCursor, axes+1, profileRows), func(i int) string {
		if i == axes {
			return tuiLook.row(m.profileCursor == axes, "aplicar los cambios")
		}
		var row strings.Builder
		axis := m.state.Axes[i]
		row.WriteString(tuiLook.rowValues(i == m.profileCursor && !m.picking, axis.Name,
			fmt.Sprintf("%-12s", shownAxis(m.value(axis))), strings.Join(axis.Allowed, " ")))
		if m.picking && i == m.profileCursor {
			row.WriteString(tuiLook.windowLines(windowFor(m.pickCursor, len(axis.Allowed), valueRows), func(j int) string {
				return tuiLook.subRow(j == m.pickCursor, axis.Allowed[j])
			}))
		}
		return row.String()
	}))
	if m.fieldMsg != "" {
		out.WriteString("\n")
		switch {
		case strings.HasPrefix(m.fieldMsg, "falló"):
			out.WriteString(tuiLook.failureLine(m.fieldMsg))
		case strings.HasPrefix(m.fieldMsg, "aplicado:"):
			out.WriteString(tuiLook.successLine(m.fieldMsg))
		default:
			out.WriteString(tuiLook.warningLine(m.fieldMsg))
		}
	}
	out.WriteString(tuiLook.helpLine(m.helpKeys("↑↓ elegir · enter cambiar · esc volver · q salir")))
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
		// Recorded per write, not after the loop: an axis that landed before a
		// later failure is still on disk and has to be defended (0336).
		m.edited[axis.Name] = value
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
	return m.header("resumen") + m.facts() + tuiLook.helpLine("enter aplicar · q salir")
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
		switch step.Status {
		case configure.Done:
			out.WriteString(tuiLook.successLine(line))
		case configure.Failed:
			out.WriteString(tuiLook.failureLine(line))
		default:
			out.WriteString("  " + tuiLook.subtitle.Render(line) + "\n")
		}
	}
	if m.applied {
		out.WriteString("\n")
		if m.result.Code == configure.CodeBlocked {
			out.WriteString(tuiLook.failureLine(m.result.Message))
		} else {
			out.WriteString(tuiLook.successLine(m.result.Message))
		}
	}
	if m.checkNotice != "" {
		out.WriteString(tuiLook.warningLine(m.checkNotice))
	}
	out.WriteString(tuiLook.helpLine("q salir"))
	return out.String()
}

// viewChoices renders one single-choice screen with its cursor.
func (m tuiModel) viewChoices(title string, options []string, cursor int) string {
	var out strings.Builder
	out.WriteString(m.header(title))
	out.WriteString(tuiLook.windowLines(windowFor(cursor, len(options), choiceRows), func(i int) string {
		return tuiLook.row(i == cursor, options[i])
	}))
	out.WriteString(tuiLook.helpLine("↑↓ elegir · enter seguir · esc volver · q salir"))
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
		Clients:      m.state.Clients,
		Skills:       m.state.Skills,
		Voice:        m.state.Voice,
		ReviewMode:   m.state.Review,
		LensModels:   m.state.Models,
		Verify:       m.verifyChoice(),
		ClientModels: m.clientModelChoices(),
		Project:      m.init,
	}, m.run)
	// The voice op recomputes every field from the builtin, whatever order the run
	// took, so the axes the human wrote here are re-applied after it: an explicit
	// edit is the last word, and a write that landed before a failure is defended
	// too (0336).
	for _, axis := range m.state.Axes {
		value, ok := m.edited[axis.Name]
		if !ok {
			continue
		}
		err := m.run.Field(axis.Name, value)
		step := configure.Step{
			Operation: configure.Operation{ID: "field:" + axis.Name, Label: "reafirmar " + axis.Name + "=" + value},
			Status:    configure.Done,
			Err:       err,
		}
		if err != nil {
			step.Status = configure.Failed
			m.result.Code = configure.CodeBlocked
			m.result.Message = step.Label + ": " + err.Error()
		}
		m.result.Steps = append(m.result.Steps, step)
	}
	// The limits section writes through the same review app the models section
	// uses: SetConfig changes only the named scalars and renders the rest, so the
	// [[check]] array keeps the gate severity the project declared. A section
	// opened and left alone sends no choice, so nothing is written; a plan that
	// already failed writes nothing either (0342).
	if m.result.Code == configure.CodeOK {
		if choice := m.limitsChoice(); choice != nil {
			err := m.run.Limits(*choice)
			step := configure.Step{
				Operation: configure.Operation{ID: "limits", Label: "fijar los límites de la revisión"},
				Status:    configure.Done,
				Err:       err,
			}
			if err != nil {
				step.Status = configure.Failed
				m.result.Code = configure.CodeBlocked
				m.result.Message = step.Label + ": " + err.Error()
			}
			m.result.Steps = append(m.result.Steps, step)
		}
	}
	// The checks section writes through the same review app the other sections
	// use: SetConfig changes only the named entries and keeps each severity the
	// file declared. A name or command is part of the approval fingerprint, so the
	// apply says the clone must approve the list again; a remedy-only edit stays
	// quiet (0346).
	if m.result.Code == configure.CodeOK {
		edits := m.checksChoice()
		if len(edits) > 0 {
			err := m.run.Checks(edits)
			step := configure.Step{
				Operation: configure.Operation{ID: "checks", Label: "editar los checks de la revisión"},
				Status:    configure.Done,
				Err:       err,
			}
			if err != nil {
				step.Status = configure.Failed
				m.result.Code = configure.CodeBlocked
				m.result.Message = step.Label + ": " + err.Error()
			} else if m.checksNeedApproval(edits) {
				m.checkNotice = "nombre o comando cambiados: vuelve a ejecutar odachi review checks approve"
			}
			m.result.Steps = append(m.result.Steps, step)
		}
	}
	m.applied = true
	m.step = stepApply
	return m, nil
}

// verifyChoice is the verifier's row as the plan reads it: the value the human
// chose and the one the file carried. nil means the row was never edited, and the
// file keeps its line (0337).
func (m tuiModel) verifyChoice() *configure.ModelChoice {
	chosen, ok := m.modelPending[modelVerifier]
	if !ok {
		return nil
	}
	return &configure.ModelChoice{Chosen: chosen, Detected: m.state.Verify}
}

// clientModelChoices is the client rows the human edited, with the value each
// file carried. A row that was not edited is not here at all (0337).
func (m tuiModel) clientModelChoices() map[string]configure.ModelChoice {
	choices := map[string]configure.ModelChoice{}
	for _, row := range m.modelRows() {
		if row == modelVerifier {
			continue
		}
		chosen, ok := m.modelPending[row]
		if !ok {
			continue
		}
		choices[row] = configure.ModelChoice{Chosen: chosen, Detected: m.state.ClientModels[row]}
	}
	if len(choices) == 0 {
		return nil
	}
	return choices
}

// viewReview renders the switch and the four lens models. A model row opens the
// host's picker or the editor, and either one shows its state in place: the
// buffer with a cursor, the list with the entry the cursor is on. The error stays
// on the screen until it is fixed (0337, 0338).
func (m tuiModel) viewReview() string {
	var out strings.Builder
	out.WriteString(m.header("Review"))
	labels := append([]string{"modo"}, tuiLenses...)
	out.WriteString(tuiLook.windowLines(windowFor(m.reviewCursor, len(labels), reviewRows), func(i int) string {
		value := m.state.Review
		if i > 0 {
			lens := tuiLenses[i-1]
			value = m.state.Models[lens]
			// An unset lens is not a blank: it runs with the session's model, and
			// the row says so instead of leaving a column empty (0338).
			if value == "" {
				value = "(la sesión)"
			}
			if m.editing && i == m.reviewCursor {
				value = m.buffer + "_"
			}
		}
		var row strings.Builder
		// One cursor at a time, as in the models section: the picker under the row
		// owns the keys while it is open (0338).
		nested := m.pickingModel && i > 0 && i == m.reviewCursor
		row.WriteString(tuiLook.rowValues(i == m.reviewCursor && !nested, labels[i], value))
		if nested {
			row.WriteString(m.modelPicker(m.pickedValue()))
		}
		return row.String()
	}))
	out.WriteString(m.catalogNotice())
	if m.editErr != "" {
		out.WriteString("\n")
		out.WriteString(tuiLook.failureLine(m.editErr))
	}
	out.WriteString(tuiLook.helpLine("enter elegir · tab seguir · esc volver · q salir"))
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

// header is the line every screen starts with: the name, the version and the
// section, then the folder it configures when the wizard knows it. The layout
// convention is title, blank line, body, blank line, keys (0338).
func (m tuiModel) header(title string) string {
	line := tuiLook.titleLine("odachi " + m.state.Version + " · " + title)
	if m.dir != "" {
		line += "\n" + tuiLook.subtitleLine(m.dir)
	}
	return line + "\n\n"
}

// factWidth is the label column of the hub's facts. It fits the longest label, so
// the values never start a column late (0338).
const factWidth = 12

// clientFacts names the client selection the way the screen that chooses it does,
// not as the word the store keeps: the hub is read by a person (0338).
func clientFacts(clients string) string {
	labels := clientLabels()
	if at := indexOf(tuiClients, clients); at >= 0 && at < len(labels) {
		return labels[at]
	}
	return clients
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
		{"Clientes", clientFacts(m.state.Clients)},
		{"Skills", skills},
		{"Voz", m.state.Voice},
		{"Review", m.state.Review},
		{"Modelos", modelsLine(m.state.Models)},
	}

	var out strings.Builder
	for _, row := range rows {
		// The label column is wide enough for the longest one -"Configurar"- so
		// the values line up; a column that one word breaks is what made the hub
		// read as a dump (0338).
		_, _ = fmt.Fprintf(&out, "  %s %s\n", tuiLook.factLabel.Render(fmt.Sprintf("%-*s", factWidth, row[0])), row[1])
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
