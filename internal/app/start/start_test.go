package start_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggz/odachi/internal/app/start"
	"github.com/biggz/odachi/internal/domain/brief"
	"github.com/biggz/odachi/internal/domain/profile"
	"github.com/biggz/odachi/internal/domain/taskdoc"
)

// The seven questions and their why, verbatim from Contract III. They live in
// the test on purpose: a reworded question would silently stop teaching.
var wantQuestions = []string{
	"¿Qué querés que exista cuando terminemos? (una frase, en tus palabras)",
	"¿Quién lo usa y desde dónde? (1 celular · 2 compu · 3 los dos)",
	"¿Cómo vas a saber que funcionó? (te propongo tres; podés escribir la tuya)",
	"¿Qué NO querés en esta primera versión? (elegí las que quieras)",
	"¿Hay datos sensibles? (1 ninguno · 2 nombres · 3 salud · 4 plata)",
	"¿Cuánto tiempo tenés para la primera versión usable? (1 hoy · 2 esta semana · 3 sin apuro)",
	"¿Qué te gustaría aprender en el camino? (elegí una o dos)",
}

var wantWhys = []string{
	"sin esto no hay nada que construir; y en tus palabras es lo que leés después para saber si vamos bien",
	"cambia todo lo que viene después, y es la decisión más cara de cambiar tarde",
	`esto se llama criterio de aceptación: es lo que separa "anda" de "creo que anda"`,
	"es lo que evita 40 cosas a medio hacer; un senior dice no, y acá el cerco lo hace cumplir",
	"define qué no puede salir del proyecto desde el día uno",
	"define cuánto recortamos; un proyecto que no se termina no enseña nada",
	"el tutor necesita saber qué explicarte más; esto es lo que te saca de novato",
}

const outcome = "Quiero reservar turnos desde el teléfono"

// expectedBrief is the exact artifact the interview produces for the answers
// below: canonical headers (Contract A), the outcome verbatim, and one durable
// decision per answered question.
const expectedBrief = `# quiero-reservar-turnos

## Qué tiene que existir
- ` + outcome + `

## Listo cuando
- Reservo un turno y aparece en la agenda

## NO incluye (primera versión)
- pagos
- login

## Decisiones tomadas
- Se usa desde el celular.
- Datos sensibles: plata; no salen del proyecto desde el día uno.
- Primera versión usable: esta semana.
- Quiere aprender: tests.
`

const expectedProfile = `name         = "novato"
extends      = ""
brief_source = "interview"
questions    = "one-by-one"
assumptions  = "never"
output       = "line"
proof        = "command"
explain      = "why"
`

const expectedFence = `[[rule]]
not   = "pagos"
paths = ["payment*", "**/payments/**"]
deps  = ["stripe", "mercadopago"]

[[rule]]
not   = "login"
paths = ["**/login*", "**/auth/login*"]
deps  = []
`

const expectedTask = `# quiero-reservar-turnos

## Requisitos (textual, del usuario)
R1 · "` + outcome + `"

## Listo cuando
- Reservo un turno y aparece en la agenda

## Pasos

## Historial
L1 · "` + outcome + `"
`

// interviewAnswers answers all seven questions in the order the interview asks
// them, mixing numbers (the human types "1") with free text.
func interviewAnswers() []string {
	return []string{
		outcome,
		"1",
		"Reservo un turno y aparece en la agenda",
		"pagos, login",
		"4",
		"2",
		"tests",
	}
}

// flagsCarryTheSameAnswers describes the same run without a human: every
// answer arrives as a flag.
func flagsCarryTheSameAnswers() start.Options {
	return start.Options{
		Outcome:   outcome,
		Audience:  "phone",
		Done:      []string{"Reservo un turno y aparece en la agenda"},
		Exclude:   []string{"pagos", "login"},
		Sensitive: "money",
		Deadline:  "week",
		Learn:     []string{"tests"},
	}
}

func novatoProfile(t *testing.T) *profile.Profile {
	t.Helper()
	p, err := profile.Resolve(profile.Builtin(), "novato")
	if err != nil {
		t.Fatalf("resolve novato: %v", err)
	}
	return &p
}

func TestScriptedInterviewWritesExactlyTheFourArtifacts(t *testing.T) {
	var events []string
	files := newFakeFiles(nil, &events)
	interview := newScriptedInterview(&events, interviewAnswers()...)

	got := start.Run(start.Options{}, interview, files)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	wantFiles := []string{
		"brief.md",
		".odachi/profile.toml",
		".odachi/fence.toml",
		"odd/tasks/0001-quiero-reservar-turnos.md",
	}
	if strings.Join(got.Files, "|") != strings.Join(wantFiles, "|") {
		t.Errorf("Files = %v, want %v", got.Files, wantFiles)
	}

	if briefPath := filepath.Join(".odachi", "profile.toml"); files.content(t, briefPath) != expectedProfile {
		t.Errorf("profile.toml =\n%s\nwant\n%s", files.content(t, briefPath), expectedProfile)
	}
	if fencePath := filepath.Join(".odachi", "fence.toml"); files.content(t, fencePath) != expectedFence {
		t.Errorf("fence.toml =\n%s\nwant\n%s", files.content(t, fencePath), expectedFence)
	}
	taskPath := filepath.Join("odd", "tasks", "0001-quiero-reservar-turnos.md")
	if got := files.content(t, taskPath); got != expectedTask {
		t.Errorf("task doc =\n%s\nwant\n%s", got, expectedTask)
	}

	parsed := brief.Parse(files.content(t, "brief.md"))
	if len(parsed.Outcome) != 1 || parsed.Outcome[0] != outcome {
		t.Errorf("brief outcome = %q, want the answer verbatim %q", parsed.Outcome, outcome)
	}
	if len(parsed.DoneWhen) != 1 || parsed.DoneWhen[0] != "Reservo un turno y aparece en la agenda" {
		t.Errorf("brief done-when = %q", parsed.DoneWhen)
	}
	if len(parsed.NonGoals) != 2 || parsed.NonGoals[0] != "pagos" || parsed.NonGoals[1] != "login" {
		t.Errorf("brief non-goals = %q, want [pagos login]", parsed.NonGoals)
	}
	if missing := parsed.Missing(); len(missing) != 0 {
		t.Errorf("the generated brief is incomplete: %q", missing)
	}

	doc := taskdoc.Parse(files.content(t, taskPath))
	if len(doc.Requirements) != 1 || doc.Requirements[0].ID != "R1" || doc.Requirements[0].Text != outcome {
		t.Errorf("R1 = %+v, want the outcome verbatim", doc.Requirements)
	}
	if len(doc.History) != 1 || doc.History[0] != `L1 · "`+outcome+`"` {
		t.Errorf("L1 = %q, want the outcome verbatim", doc.History)
	}
	if len(doc.Steps) != 0 {
		t.Errorf("Steps = %v, want none: the agent writes them, not start", doc.Steps)
	}

	if got := files.content(t, "brief.md"); got != expectedBrief {
		t.Errorf("brief.md =\n%s\nwant\n%s", got, expectedBrief)
	}
}

// TestWhyTravelsWithTheQuestionAndNeverAlone is the tutor rule: the explanation
// is part of the question block, never a follow-up message the human can miss.
func TestWhyTravelsWithTheQuestionAndNeverAlone(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)

	start.Run(start.Options{}, interview, newFakeFiles(nil, &events))

	if len(interview.questions) != len(wantQuestions) {
		t.Fatalf("asked %d questions, want %d", len(interview.questions), len(wantQuestions))
	}
	for i, question := range interview.questions {
		if question.Text != wantQuestions[i] {
			t.Errorf("question %d = %q, want %q", i+1, question.Text, wantQuestions[i])
		}
		if question.Why != wantWhys[i] {
			t.Errorf("why %d = %q, want %q", i+1, question.Why, wantWhys[i])
		}
	}
	for _, line := range interview.said {
		for _, why := range wantWhys {
			if strings.Contains(line, why) {
				t.Errorf("the why was printed as its own message: %q", line)
			}
		}
	}
}

// TestEmptyOutcomeIsRejectedAndAskedAgain: question 1 carries the whole
// project, so an empty answer is never silently accepted.
func TestEmptyOutcomeIsRejectedAndAskedAgain(t *testing.T) {
	var events []string
	answers := append([]string{"", "   ", outcome}, interviewAnswers()[1:]...)
	interview := newScriptedInterview(&events, answers...)

	got := start.Run(start.Options{}, interview, newFakeFiles(nil, &events))

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(interview.questions) != len(wantQuestions)+2 {
		t.Fatalf("asked %d questions, want %d (question 1 asked three times)", len(interview.questions), len(wantQuestions)+2)
	}
	for i := 0; i < 3; i++ {
		if interview.questions[i].Text != wantQuestions[0] {
			t.Errorf("question %d = %q, want question 1", i+1, interview.questions[i].Text)
		}
	}
	if len(interview.said) != 2 {
		t.Fatalf("said %q, want two rejection messages", interview.said)
	}
	for _, line := range interview.said {
		assertOneLine(t, line)
		if !strings.Contains(line, "escribila de nuevo") {
			t.Errorf("rejection message %q does not ask again", line)
		}
	}
}

// TestInvalidAudienceAnswerIsExplainedAndAskedAgain keeps the write path
// lenient: a human who types "tablet" gets told what the options are.
func TestInvalidAudienceAnswerIsExplainedAndAskedAgain(t *testing.T) {
	var events []string
	answers := []string{outcome, "tablet", "1"}
	answers = append(answers, interviewAnswers()[2:]...)
	interview := newScriptedInterview(&events, answers...)

	got := start.Run(start.Options{}, interview, newFakeFiles(nil, &events))

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(interview.said) != 1 {
		t.Fatalf("said %q, want one explanation for the invalid answer", interview.said)
	}
	if !strings.Contains(interview.said[0], "no entendí") {
		t.Errorf("explanation = %q, want it to name the invalid answer", interview.said[0])
	}
}

// TestBatchedShowsEveryAnswerTogetherBeforeWriting: a senior answers all seven
// and sees them as a block before a single byte hits the disk.
func TestBatchedShowsEveryAnswerTogetherBeforeWriting(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	senior, err := profile.Resolve(profile.Builtin(), "senior")
	if err != nil {
		t.Fatalf("resolve senior: %v", err)
	}

	got := start.Run(start.Options{Profile: &senior}, interview, newFakeFiles(nil, &events))

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(interview.questions) != len(wantQuestions) {
		t.Errorf("asked %d questions, want all %d", len(interview.questions), len(wantQuestions))
	}

	shown := strings.Join(interview.said, "\n")
	if !strings.Contains(shown, outcome) {
		t.Errorf("the answers were not shown together before writing: %q", shown)
	}
	for _, line := range interview.said {
		assertOneLine(t, line)
	}

	lastSay := -1
	for i, event := range events {
		if strings.HasPrefix(event, "say ") {
			lastSay = i
		}
	}
	firstWrite := firstIndex(events, "write ")
	if firstWrite < 0 {
		t.Fatal("nothing was written")
	}
	if lastSay < 0 || lastSay > firstWrite {
		t.Errorf("answers must be shown before the first write; last say at %d, first write at %d (%v)", lastSay, firstWrite, events)
	}
}

// TestForceBacksUpTheBriefAndWritesTheSameBytes: the same answers produce the
// same brief, and the file the human may have edited is never lost.
func TestForceBacksUpTheBriefAndWritesTheSameBytes(t *testing.T) {
	var firstEvents []string
	first := newFakeFiles(nil, &firstEvents)
	opts := flagsCarryTheSameAnswers()
	opts.Profile = novatoProfile(t)

	if got := start.Run(opts, nil, first); got.Code != start.CodeOK {
		t.Fatalf("first run Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	firstBrief := first.content(t, "brief.md")
	firstProfile := first.content(t, filepath.Join(".odachi", "profile.toml"))
	firstFence := first.content(t, filepath.Join(".odachi", "fence.toml"))

	var secondEvents []string
	second := &fakeFiles{files: first.files, events: &secondEvents}
	opts.Force = true

	got := start.Run(opts, nil, second)
	if got.Code != start.CodeOK {
		t.Fatalf("second run Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	if backup, write := firstIndex(secondEvents, "backup brief.md"), firstIndex(secondEvents, "write brief.md"); backup < 0 || backup > write {
		t.Errorf("brief.md must be backed up before it is rewritten; events: %v", secondEvents)
	}
	if second.content(t, "brief.md") != firstBrief {
		t.Errorf("brief.md changed with the same answers:\n%s\nvs\n%s", second.content(t, "brief.md"), firstBrief)
	}
	if second.content(t, filepath.Join(".odachi", "profile.toml")) != firstProfile {
		t.Errorf("profile.toml changed with the same answers")
	}
	if second.content(t, filepath.Join(".odachi", "fence.toml")) != firstFence {
		t.Errorf("fence.toml changed with the same answers")
	}
	taskDoc := "odd/tasks/0001-quiero-reservar-turnos.md"
	if got.Task != taskDoc {
		t.Errorf("second task doc = %q, want the existing document reused", got.Task)
	}
	for _, event := range secondEvents {
		if event == "write "+taskDoc || event == "backup "+taskDoc {
			t.Errorf("second run touched the task doc (%s); its steps and history are the memory of the work", event)
		}
	}
}

// TestExistingBriefWithoutForceIsBlocked names the flag that unblocks it.
func TestExistingBriefWithoutForceIsBlocked(t *testing.T) {
	var events []string
	files := newFakeFiles(map[string]string{"brief.md": "viejo\n"}, &events)
	opts := flagsCarryTheSameAnswers()

	got := start.Run(opts, nil, files)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "--force") {
		t.Errorf("message %q does not name --force", got.Message)
	}
	if !strings.Contains(got.Message, "odachi") {
		t.Errorf("message %q does not name the next action", got.Message)
	}
	if files.content(t, "brief.md") != "viejo\n" {
		t.Errorf("brief.md changed even though the run was blocked")
	}
	if firstIndex(events, "write ") >= 0 || firstIndex(events, "backup ") >= 0 {
		t.Errorf("a blocked run wrote files: %v", events)
	}
}

// TestNonInteractiveWithoutAnswersAsksTheHuman: exit 2 names every missing flag
// and never invents an answer.
func TestNonInteractiveWithoutAnswersAsksTheHuman(t *testing.T) {
	var events []string
	files := newFakeFiles(nil, &events)

	got := start.Run(start.Options{Machine: true}, nil, files)

	if got.Code != start.CodeNeedsAnswer {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeNeedsAnswer, got.Message)
	}
	assertOneLine(t, got.Message)
	for _, flag := range []string{"--outcome", "--audience", "--done", "--sensitive", "--deadline"} {
		if !strings.Contains(got.Message, flag) {
			t.Errorf("message %q does not name the missing %s", got.Message, flag)
		}
	}
	if !strings.Contains(got.Message, "odachi") {
		t.Errorf("message %q does not name the next action", got.Message)
	}
	if len(files.files) != 0 {
		t.Errorf("files written without answers: %v", files.paths())
	}
}

// TestAllPacingNeverAsks: the machine profile uses the flags or fails, it never
// blocks waiting for a human who is not there.
func TestAllPacingNeverAsks(t *testing.T) {
	var events []string
	ci, err := profile.Resolve(profile.Builtin(), "ci")
	if err != nil {
		t.Fatalf("resolve ci: %v", err)
	}
	interview := newScriptedInterview(&events) // no answers: any ask would fail

	got := start.Run(start.Options{Profile: &ci, Machine: true}, interview, newFakeFiles(nil, &events))

	if got.Code != start.CodeNeedsAnswer {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeNeedsAnswer, got.Message)
	}
	if len(interview.questions) != 0 {
		t.Errorf("all pacing asked %d questions, want 0", len(interview.questions))
	}

	// The same profile with the flags writes the artifacts and asks nothing.
	opts := flagsCarryTheSameAnswers()
	opts.Profile = &ci
	opts.Machine = true
	interview = newScriptedInterview(&events)

	got = start.Run(opts, interview, newFakeFiles(nil, &events))
	if got.Code != start.CodeOK {
		t.Fatalf("Code with flags = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(interview.questions) != 0 || len(interview.said) != 0 {
		t.Errorf("all pacing talked to a human: asked %d, said %q", len(interview.questions), interview.said)
	}
}

// TestInvalidFlagIsBlockedNamingTheFlagAndItsValues: the write path is lenient
// about what the human typed but never about what it silently accepts.
func TestInvalidFlagIsBlockedNamingTheFlagAndItsValues(t *testing.T) {
	var events []string
	opts := flagsCarryTheSameAnswers()
	opts.Audience = "tablet"

	got := start.Run(opts, nil, newFakeFiles(nil, &events))

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "--audience") || !strings.Contains(got.Message, "phone") {
		t.Errorf("message %q does not name --audience and its values", got.Message)
	}
	if len(events) != 0 {
		t.Errorf("invalid flags reached the filesystem: %v", events)
	}
}

// TestTaskNumberIsOneMoreThanTheHighestExisting: numbers are never reused, and
// a backup file is not a task document.
func TestTaskNumberIsOneMoreThanTheHighestExisting(t *testing.T) {
	seed := map[string]string{
		filepath.Join("odd", "tasks", "0007-viejo.md"):          "# viejo\n",
		filepath.Join("odd", "tasks", "0011-otro.md"):           "# otro\n",
		filepath.Join("odd", "tasks", "0002-respaldo.md.bak-x"): "# respaldo\n",
		filepath.Join("odd", "tasks", "notas.txt"):              "notas\n",
	}
	opts := flagsCarryTheSameAnswers()
	opts.Profile = novatoProfile(t)
	opts.Machine = true

	got := start.Run(opts, nil, newFakeFiles(seed, nil))

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	want := "odd/tasks/0012-quiero-reservar-turnos.md"
	if got.Task != want {
		t.Errorf("Task = %q, want %q", got.Task, want)
	}
}

// TestFreeTextNonGoalBecomesItsOwnRule: a preset uses its catalog paths and
// deps; free text earns a slug pattern of its own.
func TestFreeTextNonGoalBecomesItsOwnRule(t *testing.T) {
	opts := flagsCarryTheSameAnswers()
	opts.Profile = novatoProfile(t)
	opts.Machine = true
	opts.Exclude = []string{"Pagos", "datos de pacientes"}

	files := newFakeFiles(nil, nil)
	start.Run(opts, nil, files)

	want := `[[rule]]
not   = "pagos"
paths = ["payment*", "**/payments/**"]
deps  = ["stripe", "mercadopago"]

[[rule]]
not   = "datos de pacientes"
paths = ["datos-de-pacientes*"]
deps  = []
`
	if got := files.content(t, filepath.Join(".odachi", "fence.toml")); got != want {
		t.Errorf("fence.toml =\n%s\nwant\n%s", got, want)
	}
}

// TestPresetNonGoalIsMatchedCaseAndAccentInsensitively keeps a human from
// creating a second rule for the same thing by typing it differently.
func TestPresetNonGoalIsMatchedCaseAndAccentInsensitively(t *testing.T) {
	opts := flagsCarryTheSameAnswers()
	opts.Profile = novatoProfile(t)
	opts.Machine = true
	opts.Exclude = []string{"Historia Clínica", "Recordatorios"}

	files := newFakeFiles(nil, nil)
	if got := start.Run(opts, nil, files); got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	fenceText := files.content(t, filepath.Join(".odachi", "fence.toml"))
	if strings.Count(fenceText, "[[rule]]") != 2 {
		t.Errorf("fence.toml has %d rules, want 2:\n%s", strings.Count(fenceText, "[[rule]]"), fenceText)
	}
	for _, want := range []string{`not   = "historia-clinica"`, `not   = "recordatorios"`} {
		if !strings.Contains(fenceText, want) {
			t.Errorf("fence.toml is missing %q:\n%s", want, fenceText)
		}
	}
}

// TestBriefFlagReadsAnExistingBriefInsteadOfAsking covers the senior/ci path.
func TestBriefFlagReadsAnExistingBriefInsteadOfAsking(t *testing.T) {
	examplePath := filepath.Join("..", "..", "..", "testdata", "examples", "brief.md")
	data, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("read %s: %v", examplePath, err)
	}

	var events []string
	files := newFakeFiles(map[string]string{"mi-brief.md": string(data)}, &events)
	opts := start.Options{BriefPath: "mi-brief.md", Machine: true}

	got := start.Run(opts, nil, files)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if _, ok := files.files["brief.md"]; ok {
		t.Errorf("--brief must not rewrite the human's brief")
	}
	wantFiles := []string{
		".odachi/profile.toml",
		".odachi/fence.toml",
		"odd/tasks/0001-turnos-consultorio.md",
	}
	if strings.Join(got.Files, "|") != strings.Join(wantFiles, "|") {
		t.Errorf("Files = %v, want %v", got.Files, wantFiles)
	}

	firstOutcome := "Un paciente reserva un turno con un médico en menos de 1 minuto desde el teléfono."
	taskPath := filepath.Join("odd", "tasks", "0001-turnos-consultorio.md")
	doc := taskdoc.Parse(files.content(t, taskPath))
	if len(doc.Requirements) != 1 || doc.Requirements[0].Text != firstOutcome {
		t.Errorf("R1 = %+v, want the brief outcome verbatim", doc.Requirements)
	}
	if !strings.Contains(files.content(t, taskPath), `L1 · "`+firstOutcome+`"`) {
		t.Errorf("task doc does not quote the outcome in L1:\n%s", files.content(t, taskPath))
	}
	if len(doc.DoneWhen) != 5 {
		t.Errorf("task doc done-when = %q, want the five lines from the brief", doc.DoneWhen)
	}

	fenceText := files.content(t, filepath.Join(".odachi", "fence.toml"))
	if strings.Count(fenceText, "[[rule]]") != 5 {
		t.Errorf("fence.toml has %d rules, want one per non-goal:\n%s", strings.Count(fenceText, "[[rule]]"), fenceText)
	}
	for _, want := range []string{
		`not   = "Historia clínica y recetas."`,
		`paths = ["historia-clinica-y-recetas*"]`,
	} {
		if !strings.Contains(fenceText, want) {
			t.Errorf("fence.toml is missing %q:\n%s", want, fenceText)
		}
	}
}

// TestMissingBriefFileIsBlocked: a path that does not exist costs one line, not
// a stack trace.
func TestMissingBriefFileIsBlocked(t *testing.T) {
	opts := start.Options{BriefPath: "no-existe.md", Machine: true}

	got := start.Run(opts, nil, newFakeFiles(nil, nil))

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "--brief") || !strings.Contains(got.Message, "no-existe.md") {
		t.Errorf("message %q does not name --brief and the path", got.Message)
	}
}
