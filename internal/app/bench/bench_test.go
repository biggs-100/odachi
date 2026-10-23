package bench

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	benchcore "github.com/biggs-100/odachi/internal/bench"
)

// fakeFiles is the sandbox port as data: it hands out the directory names the
// test expects and records what the run did with them.
type fakeFiles struct {
	next    int
	created []string
	written []string
	removed []string
	err     error
}

func (f *fakeFiles) MkdirTemp(string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.next++
	dir := filepath.Join("/sandbox", string(rune('0'+f.next)))
	f.created = append(f.created, dir)
	return dir, nil
}

func (f *fakeFiles) MkdirAll(path string) error {
	if f.err != nil {
		return f.err
	}
	return nil
}

func (f *fakeFiles) Write(path string, _ []byte) error {
	if f.err != nil {
		return f.err
	}
	f.written = append(f.written, path)
	return nil
}

func (f *fakeFiles) RemoveAll(path string) error {
	if f.err != nil {
		return f.err
	}
	f.removed = append(f.removed, path)
	return nil
}

// call is one process the bench ran.
type call struct {
	dir  string
	env  []string
	name string
	args []string
}

// fakeExec answers by program and command, so a test writes the sandbox's
// behaviour instead of a script of side effects.
type fakeExec struct {
	calls []call
	// answer returns the output and exit code for one invocation. The default
	// answers 0 with no output, which is a sandbox where everything works.
	answer func(name string, args []string) (string, int)
	// fail makes one invocation unrunnable, which is how a test breaks a journey
	// halfway without breaking every journey.
	fail func(name string, args []string) error
}

func (e *fakeExec) Run(dir string, env []string, name string, args ...string) (string, int, error) {
	e.calls = append(e.calls, call{dir: dir, env: env, name: name, args: args})
	if e.fail != nil {
		if err := e.fail(name, args); err != nil {
			return "", -1, err
		}
	}
	if e.answer == nil {
		return "", 0, nil
	}
	output, code := e.answer(name, args)
	return output, code, nil
}

// found reports whether the bench ran args, whatever the program was.
func (e *fakeExec) found(args ...string) bool {
	for _, call := range e.calls {
		if strings.Join(call.args, " ") == strings.Join(args, " ") {
			return true
		}
	}
	return false
}

// step is one configured invocation: what the binary prints and what it exits.
type step struct {
	args   []string
	exit   int
	output string
}

// corpus builds an injected corpus of one-step journeys, one per case.
func corpus(steps ...step) []benchcore.Journey {
	journeys := make([]benchcore.Journey, 0, len(steps))
	for i, s := range steps {
		journeys = append(journeys, benchcore.Journey{
			Name:  string(rune('a'+i)) + "-journey",
			Why:   "injected",
			Steps: []benchcore.Step{{Args: s.args, Want: s.exit}},
		})
	}
	return journeys
}

// answers turns the configured steps into the process answer function: the binary
// prints what the step says, and a continuation named in the output answers with
// the exit the test declared for it.
func answers(steps []step, continuations map[string]struct {
	output string
	exit   int
}) func(name string, args []string) (string, int) {
	return func(name string, args []string) (string, int) {
		key := strings.Join(args, " ")
		for _, s := range steps {
			if strings.Join(s.args, " ") == key {
				return s.output, s.exit
			}
		}
		if c, ok := continuations[key]; ok {
			return c.output, c.exit
		}
		return "", 0
	}
}

// TestTheThreeCounts: one block that resolves, one with no command, one that
// does not resolve, and one templated. The four numbers are the whole instrument.
func TestTheThreeCounts(t *testing.T) {
	steps := []step{
		{args: []string{"check"}, exit: 1, output: "odachi: falta la prueba\nsiguiente: odachi review\n"},
		{args: []string{"level"}, exit: 1, output: "odachi: algo salió mal y no digo qué hacer\n"},
		{args: []string{"undo"}, exit: 1, output: "odachi: no hay backups\nsiguiente: odachi undo --file brief\n"},
		{args: []string{"plan"}, exit: 1, output: "odachi: falta la prueba\nsiguiente: odachi check --proof <archivo>\n"},
	}

	exec := &fakeExec{answer: answers(steps, map[string]struct {
		output string
		exit   int
	}{
		"review":            {output: "el prompt\n", exit: 0},
		"undo --file brief": {output: "no hay backups\n", exit: 1},
	})}

	result := Run(Options{Binary: "/bin/odachi", Corpus: corpus(steps...)}, &fakeFiles{}, exec)

	if result.Blocks != 4 {
		t.Fatalf("four blocks, got %d:\n%s", result.Blocks, strings.Join(result.Lines, "\n"))
	}
	if result.InBand != 1 || result.OutOfBand != 1 || result.DeadEnd != 1 || result.Templated != 1 {
		t.Fatalf("in_band=%d out_of_band=%d dead_end=%d templated=%d; want 1,1,1,1",
			result.InBand, result.OutOfBand, result.DeadEnd, result.Templated)
	}
	if result.Code != CodeFailed {
		t.Fatalf("a dead end must fail the bench, got %d", result.Code)
	}
	if len(result.DeadEnds) != 1 {
		t.Fatalf("one dead end, got %d", len(result.DeadEnds))
	}
	dead := result.DeadEnds[0]
	if dead.Command != "odachi undo --file brief" {
		t.Fatalf("the dead end must name the command that failed to resolve, got %q", dead.Command)
	}
	if dead.Exit != 1 {
		t.Fatalf("the dead end must record what the command exited, got %d", dead.Exit)
	}
	if !strings.Contains(dead.Message, "no hay backups") {
		t.Fatalf("the dead end must carry the message that named it, got %q", dead.Message)
	}
}

// TestTemplatedIsNeverRun: a command with a placeholder cannot be run alone, and
// the bench must not guess what to put in the slot.
func TestTemplatedIsNeverRun(t *testing.T) {
	steps := []step{
		{args: []string{"check"}, exit: 1, output: "siguiente: odachi check --proof <archivo>\n"},
	}
	exec := &fakeExec{answer: answers(steps, nil)}
	result := Run(Options{Binary: "/bin/odachi", Corpus: corpus(steps...)}, &fakeFiles{}, exec)

	if result.Templated != 1 || result.DeadEnd != 0 {
		t.Fatalf("templated=%d dead_end=%d; want 1,0", result.Templated, result.DeadEnd)
	}
	if exec.found("check", "--proof", "<archivo>") {
		t.Fatalf("the bench ran a templated command: %v", exec.calls)
	}
	if exec.found() {
		t.Fatalf("the bench ran nothing after the only measured step: %v", exec.calls)
	}
}

// TestNoDeadEndIsGreen: the instrument only fails for a real dead end.
func TestNoDeadEndIsGreen(t *testing.T) {
	steps := []step{
		{args: []string{"check"}, exit: 1, output: "siguiente: odachi review\n"},
	}
	exec := &fakeExec{answer: answers(steps, nil)}
	result := Run(Options{Binary: "/bin/odachi", Corpus: corpus(steps...)}, &fakeFiles{}, exec)

	if result.Code != CodeOK {
		t.Fatalf("no dead end must exit 0, got %d:\n%s", result.Code, strings.Join(result.Lines, "\n"))
	}
	if result.InBand != 1 {
		t.Fatalf("in_band = %d, want 1", result.InBand)
	}
	if !strings.Contains(result.Message, "no dead ends") {
		t.Fatalf("the verdict must say there is no dead end, got %q", result.Message)
	}
}

// TestTheBinaryUnderTestRunsTheContinuation: the message says `odachi`, and the
// bench runs the binary it was given. A machine with no `odachi` on the PATH must
// still be measurable.
func TestTheBinaryUnderTestRunsTheContinuation(t *testing.T) {
	steps := []step{
		{args: []string{"check"}, exit: 1, output: "siguiente: odachi review --lens risk\n"},
	}
	exec := &fakeExec{answer: answers(steps, nil)}
	Run(Options{Binary: "/tmp/bin/odachi", Corpus: corpus(steps...)}, &fakeFiles{}, exec)

	for _, call := range exec.calls {
		if call.name != "/tmp/bin/odachi" {
			t.Fatalf("every invocation runs the binary under test, got %q", call.name)
		}
	}
	if !exec.found("review", "--lens", "risk") {
		t.Fatalf("the continuation must keep its flags: %v", exec.calls)
	}
}

// TestSandboxesAreRemoved: a journey writes in a temp directory and the bench
// leaves the machine as it found it, even when the journey failed.
func TestSandboxesAreRemoved(t *testing.T) {
	files := &fakeFiles{}
	exec := &fakeExec{answer: answers([]step{
		{args: []string{"check"}, exit: 1, output: "siguiente: odachi review\n"},
	}, nil)}
	Run(Options{Binary: "/bin/odachi", Corpus: corpus(step{args: []string{"check"}, exit: 1})}, files, exec)

	if len(files.created) != 1 {
		t.Fatalf("one sandbox per journey, got %d", len(files.created))
	}
	if len(files.removed) != 1 || files.removed[0] != files.created[0] {
		t.Fatalf("the sandbox must be removed: created %v, removed %v", files.created, files.removed)
	}
}

// TestAPartialJourneyStillReportsItsBlocks: a journey that breaks halfway still
// gives a reading for the blocks it already found. Dropping them would hide the
// dead end behind the failure that stopped the journey.
func TestAPartialJourneyStillReportsItsBlocks(t *testing.T) {
	steps := []step{{args: []string{"check"}, exit: 1, output: "siguiente: odachi undo\n"}}
	exec := &fakeExec{
		answer: answers(steps, map[string]struct {
			output string
			exit   int
		}{"undo": {output: "no hay backups\n", exit: 1}}),
		fail: func(_ string, args []string) error {
			if len(args) > 0 && args[0] == "level" {
				return errors.New("no se pudo ejecutar")
			}
			return nil
		},
	}
	journeys := []benchcore.Journey{{
		Name:  "half-way",
		Why:   "injected",
		Steps: []benchcore.Step{{Args: []string{"check"}, Want: 1}, {Args: []string{"level"}, Want: 0}},
	}}

	result := Run(Options{Binary: "/bin/odachi", Corpus: journeys}, &fakeFiles{}, exec)

	if !result.Journeys[0].Failed {
		t.Fatalf("the journey must say it could not finish: %+v", result.Journeys[0])
	}
	if result.DeadEnd != 1 {
		t.Fatalf("the block found before the failure must still be classified, got dead_end=%d", result.DeadEnd)
	}
	if result.Code != CodeFailed {
		t.Fatalf("a dead end fails the bench even in a partial journey, got %d", result.Code)
	}
}

// TestSandboxGetsAControlledHome: the journey must not read or write the home of
// whoever ran the bench.
func TestSandboxGetsAControlledHome(t *testing.T) {
	files := &fakeFiles{}
	exec := &fakeExec{}
	Run(Options{
		Binary: "/bin/odachi",
		Env:    []string{"HOME=/home/ana", "PATH=/usr/bin"},
		Corpus: corpus(step{args: []string{"check"}}),
	}, files, exec)

	if len(exec.calls) == 0 {
		t.Fatalf("the journey must run")
	}
	home := filepath.Join(files.created[0], "home")
	env := strings.Join(exec.calls[0].env, "\n")
	if strings.Contains(env, "/home/ana") {
		t.Fatalf("the sandbox inherited the human's home:\n%s", env)
	}
	if !strings.Contains(env, "HOME="+home) {
		t.Fatalf("the sandbox must set HOME inside the journey, want HOME=%s:\n%s", home, env)
	}
}

// TestSetupRunsBeforeTheSteps: a journey arrives at the state it is about before
// anything is measured.
func TestSetupRunsBeforeTheSteps(t *testing.T) {
	files := &fakeFiles{}
	exec := &fakeExec{}
	journeys := []benchcore.Journey{{
		Name:  "with-repo",
		Why:   "injected",
		Setup: []benchcore.Action{{Command: []string{"git", "init", "-q"}}, {Path: "notas.md", Content: "# notas\n"}},
		Steps: []benchcore.Step{{Args: []string{"check"}}},
	}}

	Run(Options{Binary: "/bin/odachi", Corpus: journeys}, files, exec)

	if len(exec.calls) != 2 {
		t.Fatalf("setup then one step, got %d calls: %v", len(exec.calls), exec.calls)
	}
	if !exec.found("init", "-q") {
		t.Fatalf("the setup command must run first: %v", exec.calls)
	}
	if len(files.written) != 1 || !strings.HasSuffix(files.written[0], "notas.md") {
		t.Fatalf("the setup file must be written in the sandbox: %v", files.written)
	}
}

// TestUnknownJourneyIsBlocked: a filter that matches nothing would report "all
// good" about a corpus that never ran.
func TestUnknownJourneyIsBlocked(t *testing.T) {
	files := &fakeFiles{}
	result := Run(Options{Binary: "/bin/odachi", Only: []string{"inventado"}}, files, &fakeExec{})

	if result.Code != CodeFailed {
		t.Fatalf("an unknown journey must fail, got %d", result.Code)
	}
	if !strings.Contains(result.Message, "inventado") || !strings.Contains(result.Message, "repo-nuevo-start") {
		t.Fatalf("the message must name the typo and the corpus: %q", result.Message)
	}
	if len(files.created) != 0 {
		t.Fatalf("nothing may run when the filter is wrong")
	}
}

// TestOnlyFilters: --journeys runs the ones it names and only those.
func TestOnlyFilters(t *testing.T) {
	files := &fakeFiles{}
	result := Run(Options{Binary: "/bin/odachi", Only: []string{"plan-y-undo"}}, files, &fakeExec{})

	if len(result.Journeys) != 1 || result.Journeys[0].Name != "plan-y-undo" {
		t.Fatalf("only the named journey runs, got %v", result.Journeys)
	}
	if len(files.created) != 1 {
		t.Fatalf("one sandbox, got %d", len(files.created))
	}
}

// TestMissingBinaryIsBlocked: the bench drives a binary, so it needs one.
func TestMissingBinaryIsBlocked(t *testing.T) {
	result := Run(Options{}, &fakeFiles{}, &fakeExec{})
	if result.Code != CodeFailed || !strings.Contains(result.Message, "--binary") {
		t.Fatalf("exit %d, message %q", result.Code, result.Message)
	}
}

// TestSurpriseIsReportedNotSwallowed: when a step exits something other than
// what the journey says, the product changed and the bench says so. It does not
// fail the run: only a dead end does that.
func TestSurpriseIsReportedNotSwallowed(t *testing.T) {
	journeys := []benchcore.Journey{{
		Name:  "change",
		Why:   "injected",
		Steps: []benchcore.Step{{Args: []string{"check"}, Want: 0}},
	}}
	exec := &fakeExec{answer: answers([]step{{args: []string{"check"}, exit: 2}}, nil)}

	result := Run(Options{Binary: "/bin/odachi", Corpus: journeys}, &fakeFiles{}, exec)
	if len(result.Journeys) != 1 || len(result.Journeys[0].Surprises) != 1 {
		t.Fatalf("the surprise must be reported: %+v", result.Journeys)
	}
	if !strings.Contains(result.Journeys[0].Surprises[0], "expected 0") {
		t.Fatalf("the surprise must say what was expected, got %q", result.Journeys[0].Surprises[0])
	}
}

// TestSandboxFailureIsAReport: a machine with no temp directory is a line, not a
// panic, and every journey still reports.
func TestSandboxFailureIsAReport(t *testing.T) {
	files := &fakeFiles{err: errors.New("no hay lugar")}
	result := Run(Options{Binary: "/bin/odachi", Corpus: corpus(step{args: []string{"check"}})}, files, &fakeExec{})

	if result.Code != CodeFailed {
		t.Fatalf("a journey that could not run must fail the bench, got %d", result.Code)
	}
	if len(result.Journeys) != 1 {
		t.Fatalf("the journey must still report: %v", result.Journeys)
	}
	if !result.Journeys[0].Failed {
		t.Fatalf("the journey must say it could not be measured: %+v", result.Journeys[0])
	}
	if len(result.Journeys[0].Surprises) == 0 {
		t.Fatalf("the failure must be said: %+v", result.Journeys[0])
	}
}

// TestHumanLinesCarryTheCounts is the report the contract asks for: journeys,
// blocks, the counts, and the dead ends with their command and message.
func TestHumanLinesCarryTheCounts(t *testing.T) {
	steps := []step{
		{args: []string{"check"}, exit: 1, output: "siguiente: odachi undo\n"},
	}
	exec := &fakeExec{answer: answers(steps, map[string]struct {
		output string
		exit   int
	}{"undo": {output: "no hay backups\n", exit: 1}})}
	result := Run(Options{Binary: "/bin/odachi", Corpus: corpus(steps...)}, &fakeFiles{}, exec)

	report := strings.Join(result.Lines, "\n")
	for _, want := range []string{
		"1 journey",
		"1 step",
		"1 block",
		"in_band 0",
		"out_of_band 0",
		"templated 0",
		"dead_end 1",
		"dead ends:",
		"odachi undo",
		"no hay backups",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("the report must carry %q:\n%s", want, report)
		}
	}
}
