package level_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/level"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// fakeFiles is the disk in memory with an ordered event log, so a test can
// assert that the backup happens before the rewrite and not only after.
type fakeFiles struct {
	files  map[string][]byte
	events []string
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[path] = []byte(content)
	}
	return f
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) List(dir string) ([]string, error) {
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == dir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	data, ok := f.files[path]
	if !ok {
		return "", fmt.Errorf("no se puede respaldar %s: no existe", path)
	}
	f.events = append(f.events, "backup "+path)
	backup := path + ".bak-1a2b3c4d"
	f.files[backup] = data
	return backup, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.events = append(f.events, "write "+path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) content(t *testing.T, path string) string {
	t.Helper()
	data, ok := f.files[path]
	if !ok {
		t.Fatalf("%q no existe; archivos: %v", path, keys(f.files))
	}
	return string(data)
}

func keys(files map[string][]byte) []string {
	out := make([]string, 0, len(files))
	for path := range files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// profilePath is where the resolved profile lives.
var profilePath = filepath.Join("raiz", ".odachi", "profile.toml")

// taskPath is odd/tasks/<NNNN>-<slug>.md inside the fake project.
func taskPath(number int, slug string) string {
	return filepath.Join("raiz", "odd", "tasks", fmt.Sprintf("%04d-%s.md", number, slug))
}

// feature renders a canonical task document with the given steps and history.
func feature(outcome string, doneSteps, totalSteps int, history ...string) string {
	doc := taskdoc.TaskDoc{
		Title:        "feature",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: outcome}},
		History:      history,
	}
	for i := 1; i <= totalSteps; i++ {
		doc.Steps = append(doc.Steps, taskdoc.Step{
			ID:     "T" + strconv.Itoa(i),
			Text:   "paso " + strconv.Itoa(i),
			Done:   i <= doneSteps,
			Commit: "4a1c9f",
		})
	}
	return doc.Render()
}

// builtin resolves one shipped profile.
func builtin(t *testing.T, name string) profile.Profile {
	t.Helper()
	p, err := profile.Resolve(profile.Builtin(), name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}
	return p
}

// TestLevelWithoutEvidenceShowsTheFactsAndProposesNothing: a beginner with one
// feature started is not told to graduate.
func TestLevelWithoutEvidenceShowsTheFactsAndProposesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 2,
			`L1 · "reservar turnos"`,
			`L2 · PREGUNTA P1 · "¿Se pueden cancelar?"`,
			`L3 · VOS P1 · "No todavía"`,
			`L4 · OFICIO · "usé SQLite" (porque no hay dos escritores)`,
		),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("level wrote without --to: %v", files.events)
	}

	block := strings.Join(res.Lines, "\n")
	for _, want := range []string{"perfil: novato", "features con pasos: 1", "preguntas contestadas: 1", "decisiones de oficio: 1", "preguntas abiertas: 0"} {
		if !strings.Contains(block, want) {
			t.Errorf("block does not report %q:\n%s", want, block)
		}
	}
	if strings.Contains(block, "--to") {
		t.Errorf("level proposed a graduation with no evidence behind it:\n%s", block)
	}
}

// TestLevelProposesSeniorWhenEarned: three features with a done step and five
// answered vision questions earn the proposal, and the tool says how to accept
// it without accepting it.
func TestLevelProposesSeniorWhenEarned(t *testing.T) {
	history := []string{`L1 · "feature"`}
	for i := 1; i <= 5; i++ {
		id := "P" + strconv.Itoa(i)
		history = append(history,
			fmt.Sprintf(`L%d · PREGUNTA %s · "¿pregunta %d?"`, 2*i, id, i),
			fmt.Sprintf(`L%d · VOS %s · "respuesta %d"`, 2*i+1, id, i),
		)
	}

	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 2, history...),
		taskPath(2, "facturar-obras"):  feature("facturar obras", 1, 1, `L1 · "facturar obras"`),
		taskPath(3, "cobrar-senas"):    feature("cobrar senas", 1, 3, `L1 · "cobrar senas"`),
		// A feature with no done step is not counted: work in progress is not
		// evidence of a finished feature.
		taskPath(4, "agenda-del-dia"): feature("agenda del dia", 0, 2, `L1 · "agenda"`),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}

	block := strings.Join(res.Lines, "\n")
	if !strings.Contains(block, "features con pasos: 3") {
		t.Errorf("block does not count three features with a done step:\n%s", block)
	}
	if !strings.Contains(block, "preguntas contestadas: 5") {
		t.Errorf("block does not count five answered questions:\n%s", block)
	}
	if !strings.Contains(block, "odachi level --to senior") {
		t.Errorf("block does not say how to accept the proposal:\n%s", block)
	}
	if len(files.events) != 0 {
		t.Errorf("level wrote the profile on its own: %v", files.events)
	}
}

// TestLevelForASeniorProposesNothing: the graduation is offered once, to a
// novato. A senior is not told that something is still missing.
func TestLevelForASeniorProposesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		taskPath(1, "reservar-turnos"): feature("reservar turnos", 1, 1, `L1 · "reservar turnos"`),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior")}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	block := strings.Join(res.Lines, "\n")
	if strings.Contains(block, "--to") {
		t.Errorf("level proposed a profile to someone who is not a novato:\n%s", block)
	}
	if strings.Contains(block, "todavía no está ganado") {
		t.Errorf("level told a senior that senior is not earned yet:\n%s", block)
	}
}

// TestLevelToBacksUpBeforeWriting is the whole contract of `level --to`: the
// human's file is copied aside, then replaced, and the new file keeps all eight
// fields.
func TestLevelToBacksUpBeforeWriting(t *testing.T) {
	before := "name         = \"novato\"\nextends      = \"\"\nbrief_source = \"interview\"\nquestions    = \"one-by-one\"\nassumptions  = \"never\"\noutput       = \"line\"\nproof        = \"command\"\nexplain      = \"why\"\n"
	files := newFakeFiles(map[string]string{profilePath: before})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}

	if len(files.events) != 2 {
		t.Fatalf("events = %v, want exactly a backup and a write", files.events)
	}
	if !strings.HasPrefix(files.events[0], "backup ") {
		t.Errorf("first event = %q, want the backup: the human's file is never the price of the change", files.events[0])
	}
	if files.content(t, profilePath+".bak-1a2b3c4d") != before {
		t.Errorf("the backup does not hold the previous profile")
	}

	after := files.content(t, profilePath)
	lines := strings.Split(strings.TrimSuffix(after, "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("the new profile has %d lines, want all eight fields:\n%s", len(lines), after)
	}

	got := map[string]string{}
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("line %q is not a key = value line", line)
		}
		got[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	want := map[string]string{
		"name":         "senior",
		"extends":      "novato",
		"brief_source": "repo",
		"questions":    "batched",
		"assumptions":  "declare",
		"output":       "detail",
		"proof":        "suite-or-ci",
		"explain":      "none",
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("profile.%s = %q, want %q", key, got[key], value)
		}
	}

	block := strings.Join(res.Lines, "\n")
	for _, want := range []string{"questions", "one-by-one", "batched", "proof", "suite-or-ci"} {
		if !strings.Contains(block, want) {
			t.Errorf("the change block does not report %q:\n%s", want, block)
		}
	}
	if res.Message == "" || strings.Contains(res.Message, "\n") {
		t.Errorf("message %q must be one line", res.Message)
	}
}

// TestLevelToWithoutAProfileFileWritesOne: a project that never wrote
// .odachi/profile.toml gets it now, with nothing to back up.
func TestLevelToWithoutAProfileFileWritesOne(t *testing.T) {
	files := newFakeFiles(nil)

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 1 || !strings.HasPrefix(files.events[0], "write ") {
		t.Fatalf("events = %v, want one write and no backup", files.events)
	}
	if !strings.Contains(files.content(t, profilePath), `name         = "senior"`) {
		t.Errorf("profile was not written:\n%s", files.content(t, profilePath))
	}
}

// TestLevelToUnknownProfileNamesTheAvailableOnes: a typo is blocked, and the
// message offers what exists instead of a dead end.
func TestLevelToUnknownProfileNamesTheAvailableOnes(t *testing.T) {
	files := newFakeFiles(nil)

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "sinior"}, files)

	if res.Code != level.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, level.CodeBlocked)
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("message must be one line: %q", res.Message)
	}
	for _, want := range []string{"sinior", "novato", "senior", "ci"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("message %q does not mention %q", res.Message, want)
		}
	}
	if len(files.events) != 0 {
		t.Errorf("a blocked level wrote %v", files.events)
	}
}

// TestLevelToTheSameProfileWritesNothing: applying what is already in force is
// not an event.
func TestLevelToTheSameProfileWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{profilePath: "name = \"senior\"\n"})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("level rewrote an unchanged profile: %v", files.events)
	}
	if !strings.Contains(res.Message, "senior") {
		t.Errorf("message %q does not say the profile was already senior", res.Message)
	}
}

// agentsPath is the file whose last section holds the voice.
var agentsPath = filepath.Join("raiz", "AGENTS.md")

// blockOf returns the block the renderer writes for one voice, or a failed test.
func blockOf(t *testing.T, voice string) string {
	t.Helper()
	block, err := rules.Block(voice)
	if err != nil {
		t.Fatalf("rules.Block(%q): %v", voice, err)
	}
	return block
}

// TestLevelToRewritesTheVoiceSectionAndNothingElse: accepting a profile changes
// how the agent talks, so the block's last section follows — and every other byte
// of the file, including what the human wrote around the block, survives. It is
// the same promise the .gitignore block makes.
func TestLevelToRewritesTheVoiceSectionAndNothingElse(t *testing.T) {
	before := "mío arriba\n\n" + blockOf(t, persona.Novato) + "\nmío abajo\n"
	files := newFakeFiles(map[string]string{agentsPath: before})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	want := "mío arriba\n\n" + blockOf(t, persona.Senior) + "\nmío abajo\n"
	if got := files.content(t, agentsPath); got != want {
		t.Fatalf("la voz no se reemplazó en su lugar:\n%q\nvs\n%q", got, want)
	}
	if !slices.Contains(res.Files, "AGENTS.md") {
		t.Errorf("Files = %v, want it to name AGENTS.md", res.Files)
	}
	if !strings.Contains(strings.Join(res.Lines, "\n"), "voz senior") {
		t.Errorf("las líneas no reportan la voz:\n%s", strings.Join(res.Lines, "\n"))
	}
	backup := slices.Index(files.events, "backup "+agentsPath)
	rewrite := slices.Index(files.events, "write "+agentsPath)
	if backup < 0 || rewrite < 0 || backup > rewrite {
		t.Errorf("events = %v, want the block backed up before it is rewritten", files.events)
	}
}

// TestLevelToRefreshesTheNativeAgentWhenItExists: the selector and the block are
// two surfaces of one voice. After a profile change they read the same words, and
// the old file is backed up before the rewrite.
func TestLevelToRefreshesTheNativeAgentWhenItExists(t *testing.T) {
	oldBody, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	seed := map[string]string{
		profilePath: "name = \"novato\"\n",
		agentsPath:  blockOf(t, persona.Novato),
	}
	for _, file := range persona.Files(oldBody) {
		seed[filepath.Join("raiz", filepath.FromSlash(file.Path))] = file.Content
	}
	files := newFakeFiles(seed)

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	newBody, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Body(senior): %v", err)
	}
	for _, file := range persona.Files(newBody) {
		path := filepath.Join("raiz", filepath.FromSlash(file.Path))
		if got := files.content(t, path); got != file.Content {
			t.Errorf("%s no quedó espejado:\n%q\nvs\n%q", file.Path, got, file.Content)
		}
		if !slices.Contains(res.Files, file.Path) {
			t.Errorf("Files = %v, want it to name %s", res.Files, file.Path)
		}
		backup := slices.Index(files.events, "backup "+path)
		rewrite := slices.Index(files.events, "write "+path)
		if backup < 0 || rewrite < 0 || backup > rewrite {
			t.Errorf("events = %v, want %s backed up before the rewrite", files.events, file.Path)
		}
	}
	if !strings.Contains(strings.Join(res.Lines, "\n"), "agente nativo") {
		t.Errorf("las líneas no reportan el agente nativo:\n%s", strings.Join(res.Lines, "\n"))
	}
}

// TestLevelToRefreshesOnlyTheChosenClient: the stored choice filters the mirror
// a profile change refreshes, and the other client keeps its bytes.
func TestLevelToRefreshesOnlyTheChosenClient(t *testing.T) {
	oldBody, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	seed := map[string]string{
		profilePath: "name = \"novato\"\n",
		agentsPath:  blockOf(t, persona.Novato),
	}
	for _, file := range persona.Files(oldBody) {
		seed[filepath.Join("raiz", filepath.FromSlash(file.Path))] = file.Content
	}
	files := newFakeFiles(seed)

	res := level.Run(level.Options{
		Dir:     "raiz",
		Current: builtin(t, "novato"),
		To:      "senior",
		Agents:  persona.SelectionOpenCode,
	}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	newBody, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Body(senior): %v", err)
	}
	chosen, _ := persona.FilesFor(persona.SelectionOpenCode, newBody)
	if got := files.content(t, filepath.Join("raiz", filepath.FromSlash(chosen.Path))); got != chosen.Content {
		t.Errorf("%s no quedó espejado:\n%q\nvs\n%q", chosen.Path, got, chosen.Content)
	}
	other, _ := persona.FilesFor(persona.SelectionClaude, oldBody)
	otherPath := filepath.Join("raiz", filepath.FromSlash(other.Path))
	if got := files.content(t, otherPath); got != other.Content {
		t.Errorf("%s se refrescó aunque no estaba elegido:\n%q\nvs\n%q", other.Path, got, other.Content)
	}
	if slices.Contains(res.Files, other.Path) {
		t.Errorf("Files = %v, want it not to name %s", res.Files, other.Path)
	}
}

// TestLevelToDoesNotInstallAMissingNativeAgent: level refreshes what init
// installed; it does not turn a missing file into a second installer.
func TestLevelToDoesNotInstallAMissingNativeAgent(t *testing.T) {
	files := newFakeFiles(map[string]string{
		profilePath: "name = \"novato\"\n",
		agentsPath:  blockOf(t, persona.Novato),
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	for _, file := range persona.Files("") {
		path := filepath.Join("raiz", filepath.FromSlash(file.Path))
		if slices.Contains(files.events, "write "+path) {
			t.Errorf("level escribió %s: instalar es init", file.Path)
		}
	}
}

// TestLevelToTheSameVoiceTwiceWritesNothing: the second run of the same command
// is not an event, and the message says so instead of claiming work it did not do.
func TestLevelToTheSameVoiceTwiceWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{
		profilePath: "name = \"senior\"\n",
		agentsPath:  blockOf(t, persona.Novato),
	})

	if res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior"), To: "senior"}, files); res.Code != level.CodeOK {
		t.Fatalf("first run Code = %d (message: %s)", res.Code, res.Message)
	}
	files.events = nil

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "senior"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("el segundo level escribió: %v", files.events)
	}
	if !strings.Contains(res.Message, "no escribí nada") {
		t.Errorf("Message = %q, want it to say nothing was written", res.Message)
	}
}

// TestLevelToGivesAnOldBlockAVoiceSection: a project initialised before the voice
// moved into the block has no section. It gains one at the end of the block, and
// the rules it already had stay exactly as they were.
func TestLevelToGivesAnOldBlockAVoiceSection(t *testing.T) {
	old := "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas de la versión vieja\n<!-- odachi:end -->\n"
	files := newFakeFiles(map[string]string{
		profilePath: "name = \"novato\"\n",
		agentsPath:  old,
	})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}

	want := "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas de la versión vieja\n\n" +
		rules.Section + "\n\n" + voiceOf(t, persona.Senior) + "<!-- odachi:end -->\n"
	if got := files.content(t, agentsPath); got != want {
		t.Fatalf("la sección no se agregó al final del bloque:\n%q\nvs\n%q", got, want)
	}
}

// TestLevelToWithoutABlockNamesInit: a file odachi did not write is not odachi's
// to invent. The profile is applied, the block is left alone, and the line names
// the command that writes it.
func TestLevelToWithoutABlockNamesInit(t *testing.T) {
	files := newFakeFiles(map[string]string{agentsPath: "solo mío\n"})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "senior"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if got := files.content(t, agentsPath); got != "solo mío\n" {
		t.Errorf("level escribió en un archivo sin bloque: %q", got)
	}
	if !strings.Contains(strings.Join(res.Lines, "\n"), "odachi init") {
		t.Errorf("las líneas no nombran el comando que escribe el bloque:\n%s", strings.Join(res.Lines, "\n"))
	}
}

// TestLevelToCIWritesTheSeniorVoice: ci is a profile, not a person, and the block
// it gets is the one nobody has to teach.
func TestLevelToCIWritesTheSeniorVoice(t *testing.T) {
	files := newFakeFiles(map[string]string{agentsPath: blockOf(t, persona.Novato)})

	res := level.Run(level.Options{Dir: "raiz", Current: builtin(t, "novato"), To: "ci"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if got := files.content(t, agentsPath); got != blockOf(t, persona.Senior) {
		t.Fatalf("el bloque de ci no es el de senior:\n%q", got)
	}
	if !strings.Contains(res.Message, "senior") {
		t.Errorf("Message = %q, want it to name the voice that was written", res.Message)
	}
}

// voiceOf is the body of one voice, or a failed test.
func voiceOf(t *testing.T, name string) string {
	t.Helper()
	voice, err := persona.Voice(name)
	if err != nil {
		t.Fatalf("persona.Voice(%q): %v", name, err)
	}
	return voice
}

// TestLevelToAProfileWithoutAVoiceSaysWhichNamesHaveOne: a human may write a
// profile of their own that extends a shipped one. The ceremony is still theirs to
// accept, the block is left alone, and the line says why and which names have a
// voice — a blank or a guess would both be worse.
func TestLevelToAProfileWithoutAVoiceSaysWhichNamesHaveOne(t *testing.T) {
	mine, err := profile.Resolve(profile.Builtin(), "senior")
	if err != nil {
		t.Fatalf("Resolve(senior): %v", err)
	}
	mine.Name = "mio"

	files := newFakeFiles(map[string]string{agentsPath: blockOf(t, persona.Novato)})

	res := level.Run(level.Options{Dir: "raiz", Current: mine, To: "mio"}, files)

	if res.Code != level.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, level.CodeOK, res.Message)
	}
	if got := files.content(t, agentsPath); got != blockOf(t, persona.Novato) {
		t.Errorf("la voz se tocó para un perfil sin voz:\n%q", got)
	}
	block := strings.Join(res.Lines, "\n") + res.Message
	for _, want := range []string{"mio", "novato", "senior", "ci"} {
		if !strings.Contains(block, want) {
			t.Errorf("las líneas no nombran %q:\n%s", want, block)
		}
	}
}
