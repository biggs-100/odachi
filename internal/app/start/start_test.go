package start_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// The seven questions and their why, verbatim from Contract III. They live in
// the test on purpose: a reworded question would silently stop teaching.
var wantQuestions = []string{
	"What do you want to exist when we finish? (one sentence, in your own words)",
	"Who uses it, and from where? (1 phone · 2 computer · 3 both)",
	"How will you know it worked? (I propose three; you can write your own)",
	"What do you NOT want in this first version? (choose as many as you like)",
	"Is there sensitive data? (1 none · 2 names · 3 health · 4 money)",
	"How much time do you have for the first usable version? (1 today · 2 this week · 3 no rush)",
	"What would you like to learn along the way? (choose one or two)",
}

var wantWhys = []string{
	"without this there is nothing to build; and in your own words it is what you read later to know we are on track",
	"it changes everything that comes after, and it is the most expensive decision to change late",
	`this is called an acceptance criterion: it separates "it works" from "I think it works"`,
	"it is what prevents 40 half-done things; a senior says no, and here the fence enforces it",
	"it defines what cannot leave the project from day one",
	"it defines how much we cut; a project that never finishes teaches nothing",
	"the tutor needs to know what to explain more; this is what takes you past novice",
}

const outcome = "Quiero reservar turnos desde el teléfono"

// expectedBrief is the exact artifact the interview produces for the answers
// below: canonical headers (Contract A), the outcome verbatim, and one durable
// decision per answered question.
const expectedBrief = `# quiero-reservar-turnos

## What must exist
- ` + outcome + `

## Done when
- Reservo un turno y aparece en la agenda

## Not included (first version)
- pagos
- login

## Decisions taken
- It is used from the phone.
- Sensitive data: money; it does not leave the project from day one.
- First usable version: this week.
- Wants to learn: tests.
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

## Requirements (verbatim, from the user)
R1 · "` + outcome + `"

## Done when
- Reservo un turno y aparece en la agenda

## Steps

## History
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
// answer arrives as a flag. Store is the root the caller resolved from the
// repository; the lock lives there, so the run that passes a port takes it.
func flagsCarryTheSameAnswers() start.Options {
	return start.Options{
		Store:     "store",
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

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	wantFiles := []string{
		"brief.md",
		"odd/profile.toml",
		"odd/fence.toml",
		"odd/tasks/0001-quiero-reservar-turnos.md",
	}
	if strings.Join(got.Files, "|") != strings.Join(wantFiles, "|") {
		t.Errorf("Files = %v, want %v", got.Files, wantFiles)
	}

	if briefPath := filepath.Join("odd", "profile.toml"); files.content(t, briefPath) != expectedProfile {
		t.Errorf("profile.toml =\n%s\nwant\n%s", files.content(t, briefPath), expectedProfile)
	}
	if fencePath := filepath.Join("odd", "fence.toml"); files.content(t, fencePath) != expectedFence {
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
// TestStartTellsTheAgentToCommitThePlan: start writes four artifacts (brief,
// profile, fence and task doc) and the plan is all of them. If they do not land
// together, the pair can split (0031) or a reset can lose them.
func TestStartTellsTheAgentToCommitThePlan(t *testing.T) {
	var events []string
	files := newFakeFiles(nil, &events)
	interview := newScriptedInterview(&events, interviewAnswers()...)

	got := start.Run(start.Options{}, interview, files, nil)

	for _, want := range []string{"commit what start wrote", "the plan", "before you build"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to contain %q", got.Message, want)
		}
	}
}

func TestWhyTravelsWithTheQuestionAndNeverAlone(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)

	start.Run(start.Options{}, interview, newFakeFiles(nil, &events), nil)

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

	got := start.Run(start.Options{}, interview, newFakeFiles(nil, &events), nil)

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
		if !strings.Contains(line, "write it again") {
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

	got := start.Run(start.Options{}, interview, newFakeFiles(nil, &events), nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if len(interview.said) != 1 {
		t.Fatalf("said %q, want one explanation for the invalid answer", interview.said)
	}
	if !strings.Contains(interview.said[0], "did not understand") {
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

	got := start.Run(start.Options{Profile: &senior}, interview, newFakeFiles(nil, &events), nil)

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

	if got := start.Run(opts, nil, first, nil); got.Code != start.CodeOK {
		t.Fatalf("first run Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	firstBrief := first.content(t, "brief.md")
	firstProfile := first.content(t, filepath.Join("odd", "profile.toml"))
	firstFence := first.content(t, filepath.Join("odd", "fence.toml"))

	var secondEvents []string
	second := &fakeFiles{files: first.files, events: &secondEvents}
	opts.Force = true

	got := start.Run(opts, nil, second, nil)
	if got.Code != start.CodeOK {
		t.Fatalf("second run Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	if backup, write := firstIndex(secondEvents, "backup brief.md"), firstIndex(secondEvents, "write brief.md"); backup < 0 || backup > write {
		t.Errorf("brief.md must be backed up before it is rewritten; events: %v", secondEvents)
	}
	if second.content(t, "brief.md") != firstBrief {
		t.Errorf("brief.md changed with the same answers:\n%s\nvs\n%s", second.content(t, "brief.md"), firstBrief)
	}
	if second.content(t, filepath.Join("odd", "profile.toml")) != firstProfile {
		t.Errorf("profile.toml changed with the same answers")
	}
	if second.content(t, filepath.Join("odd", "fence.toml")) != firstFence {
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

// closedDocument renders one task document whose only step is done: the memory of
// a finished feature, which a new one must not plan into.
func closedDocument(done bool) string {
	return taskdoc.TaskDoc{
		Title:        "the-review-prompt",
		Base:         "abc1234",
		Requirements: []taskdoc.Requirement{{ID: "R1", Text: "the review prompt carries each row once"}},
		DoneWhen:     []string{"each row appears once"},
		Steps:        []taskdoc.Step{{ID: "T1", Text: "the rows", Done: done, Requirement: "R1"}},
		History:      []string{`L1 · "the review prompt carries each row once"`},
	}.Render()
}

// openDocument is the same document with its only step still unchecked.
// openDocument is the same shape with its only step unchecked.
func openDocument() string {
	return closedDocument(false)
}

// TestStartDoesNotReuseAClosedDocument: a document with every step done is the
// memory of a finished feature. A new feature whose first three words repeat it
// gets the next number, instead of planning into that file and reporting done
// without writing anything - what happened to 0046.
func TestStartDoesNotReuseAClosedDocument(t *testing.T) {
	var events []string
	files := newFakeFiles(map[string]string{
		"odd/tasks/0046-the-review-prompt.md": closedDocument(true),
	}, &events)
	opts := flagsCarryTheSameAnswers()
	opts.Outcome = "the review prompt carries each row once, again"
	opts.Profile = novatoProfile(t)

	got := start.Run(opts, nil, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	want := "odd/tasks/0047-the-review-prompt.md"
	if got.Task != want {
		t.Fatalf("Task = %q, want the next number %q", got.Task, want)
	}
	if _, ok := files.files[want]; !ok {
		t.Errorf("%s was not written", want)
	}
	for _, event := range events {
		if event == "write odd/tasks/0046-the-review-prompt.md" || event == "backup odd/tasks/0046-the-review-prompt.md" {
			t.Errorf("the closed document was touched: %s", event)
		}
	}
}

// TestStartReusesAnOpenDocument: the other half. Running start again on the same
// unfinished feature must not leave the agent two documents to choose from.
func TestStartReusesAnOpenDocument(t *testing.T) {
	var events []string
	// The steady state this feature creates: the finished feature's closed
	// document and the open one that followed it share the same three words.
	files := newFakeFiles(map[string]string{
		"odd/tasks/0046-the-review-prompt.md": closedDocument(true),
		"odd/tasks/0047-the-review-prompt.md": openDocument(),
	}, &events)
	opts := flagsCarryTheSameAnswers()
	opts.Outcome = "the review prompt carries each row once, again"
	opts.Profile = novatoProfile(t)

	got := start.Run(opts, nil, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if want := "odd/tasks/0047-the-review-prompt.md"; got.Task != want {
		t.Errorf("Task = %q, want the open document %q", got.Task, want)
	}
	if _, ok := files.files["odd/tasks/0048-the-review-prompt.md"]; ok {
		t.Error("a third document was written for the same open feature")
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "odd/tasks/0047-the-review-prompt.md") {
		t.Errorf("the run does not say which document it kept:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestExistingBriefWithoutForceIsBlocked names the flag that unblocks it.
func TestExistingBriefWithoutForceIsBlocked(t *testing.T) {
	var events []string
	files := newFakeFiles(map[string]string{"brief.md": "viejo\n"}, &events)
	opts := flagsCarryTheSameAnswers()

	got := start.Run(opts, nil, files, nil)

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

	got := start.Run(start.Options{Machine: true}, nil, files, nil)

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

// TestAnInterviewThatCannotReadNamesTheFlags: a terminal that dies mid-question
// and a run with no terminal are the same state. The message names the flags
// that unblock the run instead of the adapter's read error.
func TestAnInterviewThatCannotReadNamesTheFlags(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events) // no answers: Ask returns an error
	files := newFakeFiles(nil, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeNeedsAnswer {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeNeedsAnswer, got.Message)
	}
	assertOneLine(t, got.Message)
	for _, flag := range []string{"--outcome", "--audience", "--done", "--sensitive", "--deadline"} {
		if !strings.Contains(got.Message, flag) {
			t.Errorf("message %q does not name the missing %s", got.Message, flag)
		}
	}
	if strings.Contains(got.Message, "could not read the answer") {
		t.Errorf("message %q still reports the adapter error instead of the flags", got.Message)
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

	got := start.Run(start.Options{Profile: &ci, Machine: true}, interview, newFakeFiles(nil, &events), nil)

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

	got = start.Run(opts, interview, newFakeFiles(nil, &events), nil)
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

	got := start.Run(opts, nil, newFakeFiles(nil, &events), nil)

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

	got := start.Run(opts, nil, newFakeFiles(seed, nil), nil)

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
	start.Run(opts, nil, files, nil)

	want := `[[rule]]
not   = "pagos"
paths = ["payment*", "**/payments/**"]
deps  = ["stripe", "mercadopago"]

[[rule]]
not   = "datos de pacientes"
paths = ["datos-de-pacientes*"]
deps  = []
`
	if got := files.content(t, filepath.Join("odd", "fence.toml")); got != want {
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
	if got := start.Run(opts, nil, files, nil); got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	fenceText := files.content(t, filepath.Join("odd", "fence.toml"))
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
	files := newFakeFiles(map[string]string{"my-brief.md": string(data)}, &events)
	opts := start.Options{BriefPath: "my-brief.md", Machine: true}

	got := start.Run(opts, nil, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	if _, ok := files.files["brief.md"]; ok {
		t.Errorf("--brief must not rewrite the human's brief")
	}
	wantFiles := []string{
		"odd/profile.toml",
		"odd/fence.toml",
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

	fenceText := files.content(t, filepath.Join("odd", "fence.toml"))
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
	opts := start.Options{BriefPath: "does-not-exist.md", Machine: true}

	got := start.Run(opts, nil, newFakeFiles(nil, nil), nil)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	assertOneLine(t, got.Message)
	if !strings.Contains(got.Message, "--brief") || !strings.Contains(got.Message, "does-not-exist.md") {
		t.Errorf("message %q does not name --brief and the path", got.Message)
	}
}

// TestStartDoesNotInferSeniorFromRepository: `odachi init` creates `.git` before
// `start` decides the profile, so a repository is not evidence of experience. A
// beginner's first project must start on novato; senior is earned with
// `odachi level --to` or inferred from a brief the human brought.
func TestStartDoesNotInferSeniorFromRepository(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	files := newFakeFiles(map[string]string{".git": "gitdir: ../.git/worktrees/x\n"}, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}
	profile := files.content(t, "odd/profile.toml")
	if !strings.Contains(profile, `name         = "novato"`) {
		t.Errorf("a repo without a brief must start on novato:\n%s", profile)
	}
}

// TestStartPutsTheVoiceOfTheProfileItWrote: the profile decides how the agent
// talks, and this run is what decided the profile. A block left at the old voice
// is an agent that speaks differently from what the project promises, which is
// the one thing doctor's eleventh check can catch and nobody wants to read.
func TestStartPutsTheVoiceOfTheProfileItWrote(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	files := newFakeFiles(map[string]string{"AGENTS.md": blockText(t, persona.Senior)}, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}

	blockBytes, err := files.Read("AGENTS.md")
	if err != nil {
		t.Fatalf("AGENTS.md: %v", err)
	}
	if want := blockText(t, persona.Novato); string(blockBytes) != want {
		t.Errorf("the block kept the wrong voice:\n%s", blockBytes)
	}
}

// TestStartRefreshesTheNativeAgentItFinds: the selector mirrors the block, and
// this run just changed the block's voice. Leaving the old mirror behind would
// make the agent a human picks read a voice the project no longer promises.
func TestStartRefreshesTheNativeAgentItFinds(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)

	seed := map[string]string{"AGENTS.md": blockText(t, persona.Senior)}
	oldBody, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Body(senior): %v", err)
	}
	for _, file := range persona.Files(oldBody) {
		seed[file.Path] = file.Content
	}
	files := newFakeFiles(seed, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}
	newBody, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(newBody) {
		if got := files.content(t, file.Path); got != file.Content {
			t.Errorf("%s did not get the new voice:\n%q\nvs\n%q", file.Path, got, file.Content)
		}
	}
}

// TestStartRefreshesOnlyTheChosenClient: the stored choice also governs what a
// later start refreshes. A client that was not chosen keeps its old file.
func TestStartRefreshesOnlyTheChosenClient(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)

	oldBody, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("rules.Body(senior): %v", err)
	}
	seed := map[string]string{"AGENTS.md": blockText(t, persona.Senior)}
	for _, file := range persona.Files(oldBody) {
		seed[file.Path] = file.Content
	}
	files := newFakeFiles(seed, &events)

	got := start.Run(start.Options{Agents: persona.SelectionOpenCode}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}
	newBody, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	chosen, _ := persona.FilesFor(persona.SelectionOpenCode, newBody)
	if got := files.content(t, chosen.Path); got != chosen.Content {
		t.Errorf("%s did not get the new voice:\n%q\nvs\n%q", chosen.Path, got, chosen.Content)
	}
	other, _ := persona.FilesFor(persona.SelectionClaude, oldBody)
	if got := files.content(t, other.Path); got != other.Content {
		t.Errorf("%s was refreshed although it was not the chosen client:\n%q\nvs\n%q", other.Path, got, other.Content)
	}
}

// TestStartLeavesAProjectWithoutABlockAlone: the block belongs to init. Creating
// it here would give the same artifact two owners.
func TestStartLeavesAProjectWithoutABlockAlone(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	files := newFakeFiles(nil, &events)

	start.Run(start.Options{}, interview, files, nil)

	if exists, _ := files.Exists("AGENTS.md"); exists {
		t.Error("start created AGENTS.md; the block belongs to odachi init")
	}
}

// TestStartDoesNotRewriteABlockAlreadyInTheRightVoice: a second run that changes
// nothing must write nothing, byte for byte.
func TestStartDoesNotRewriteABlockAlreadyInTheRightVoice(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	files := newFakeFiles(map[string]string{"AGENTS.md": blockText(t, persona.Novato)}, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}
	for _, written := range got.Files {
		if written == "AGENTS.md" {
			t.Errorf("start rewrote a block that was already right: %v", got.Files)
		}
	}
}

// TestStartRecordsTheFeatureBase: the task doc carries the commit the feature
// starts from, because the closing gate asks for the review of that diff, not of
// the commit that closes the feature. Without a commit, nothing is recorded.
func TestStartRecordsTheFeatureBase(t *testing.T) {
	const sha = "8ee3e55f31ee50a088bb1f8aa8bca4afdbc40319"
	var events []string
	files := newFakeFiles(nil, &events)
	interview := newScriptedInterview(&events, interviewAnswers()...)

	got := start.Run(start.Options{Base: sha}, interview, files, nil)
	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}

	taskPath := filepath.Join("odd", "tasks", "0001-quiero-reservar-turnos.md")
	doc := taskdoc.Parse(files.content(t, taskPath))
	if doc.Base != sha {
		t.Errorf("Base = %q, want %q", doc.Base, sha)
	}
}

// TestStartNamesTheDocumentsItDoesNotRead: a project that keeps ODD names gets
// the truth once, at the moment it adopts odachi, instead of a summary that
// counts one feature where there are ten.
func TestStartNamesTheDocumentsItDoesNotRead(t *testing.T) {
	var events []string
	files := newFakeFiles(map[string]string{"odd/tasks/qonta-einvoicing.md": "# qonta-einvoicing\n"}, &events)
	interview := newScriptedInterview(&events, interviewAnswers()...)

	got := start.Run(start.Options{}, interview, files, nil)
	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeOK, got.Message)
	}
	block := strings.Join(got.Lines, "\n")
	if !strings.Contains(block, "not features odachi reads") || !strings.Contains(block, "qonta-einvoicing.md") {
		t.Errorf("the run does not name the document it skips:\n%s", block)
	}
}

// block renders the canonical block for one voice, so the tests compare against
// the renderer instead of a copy of its text.
func blockText(t *testing.T, voice string) string {
	t.Helper()
	content, err := rules.Block(voice)
	if err != nil {
		t.Fatalf("rules.Block(%s): %v", voice, err)
	}
	return content
}

// TestStartDoesNotInstallTheAgentsItDoesNotFind: the mirror is refreshed, never
// installed. A project that keeps its agents in the home stays without them.
func TestStartDoesNotInstallTheAgentsItDoesNotFind(t *testing.T) {
	var events []string
	interview := newScriptedInterview(&events, interviewAnswers()...)
	files := newFakeFiles(map[string]string{"AGENTS.md": blockText(t, persona.Novato)}, &events)

	got := start.Run(start.Options{}, interview, files, nil)

	if got.Code != start.CodeOK {
		t.Fatalf("Code = %d, want %d (%s)", got.Code, start.CodeOK, got.Message)
	}
	body, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, file := range persona.Files(body) {
		if _, ok := files.files[files.key(file.Path)]; ok {
			t.Errorf("%s was installed by start; the mirror is refreshed, never installed", file.Path)
		}
	}
}

// TestStartStopsWhenTheMatchingDocumentCannotBeRead: a document whose slug matches
// is read to tell a re-run from a new feature that repeats its first three words.
// A read that fails is a block the agent can fix - and it happens after the brief and
// the fence are written - so the message names the document instead of planning into
// one nobody could read.
func TestStartStopsWhenTheMatchingDocumentCannotBeRead(t *testing.T) {
	var events []string
	files := newFakeFiles(map[string]string{
		"odd/tasks/0001-el-turno-queda.md": closedDocument(true),
	}, &events)
	files.failRead = func(path string) error {
		if strings.HasSuffix(path, "0001-el-turno-queda.md") {
			return errors.New("permiso denegado")
		}
		return nil
	}
	opts := flagsCarryTheSameAnswers()
	opts.Outcome = "el turno queda"

	got := start.Run(opts, nil, files, nil)

	if got.Code != start.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", got.Code, start.CodeBlocked, got.Message)
	}
	for _, want := range []string{"could not read", "0001-el-turno-queda.md"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want %q", got.Message, want)
		}
	}
}
