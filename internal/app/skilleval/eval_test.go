package skilleval_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/skilleval"
	domain "github.com/biggs-100/odachi/internal/domain/skilleval"
	corpus "github.com/biggs-100/odachi/internal/skilleval"
)

// withStream is the run the exam is asking for: the test first, the suite red,
// the code after, and the skill loaded on the way in.
const withStream = `{"type":"tool_use","part":{"tool":"skill","state":{"input":{"id":"tests"}}}}
{"type":"tool_use","part":{"tool":"write","state":{"input":{"filePath":"config_test.go"}}}}
{"type":"tool_use","part":{"tool":"bash","state":{"input":{"command":"go test ./..."}}}}
{"type":"tool_use","part":{"tool":"edit","state":{"input":{"path":"config.go"}}}}
`

// withoutStream is a run that wrote the code first: it can score neither
// behaviour, which is the difference the exam exists to see.
const withoutStream = `{"type":"tool_use","part":{"tool":"write","state":{"input":{"filePath":"config.go"}}}}
{"type":"tool_use","part":{"tool":"write","state":{"input":{"filePath":"config_test.go"}}}}
`

type fakeFiles struct{ written map[string][]byte }

func (f *fakeFiles) Write(path string, data []byte) error {
	if f.written == nil {
		f.written = map[string][]byte{}
	}
	f.written[path] = data
	return nil
}
func (f *fakeFiles) MkdirAll(string) error { return nil }
func (f *fakeFiles) Exists(path string) (bool, error) {
	_, ok := f.written[path]
	return ok, nil
}

type fakeShell struct {
	dirs     []string
	commands []string
	inputs   []string
}

func (s *fakeShell) RunInput(dir, command, input string, _ int) (string, string, int, bool, error) {
	s.dirs = append(s.dirs, dir)
	s.commands = append(s.commands, command)
	s.inputs = append(s.inputs, input)
	if strings.Contains(dir, string(filepath.Separator)+"without"+string(filepath.Separator)) {
		return withoutStream, "", 0, false, nil
	}
	return withStream, "", 0, false, nil
}

type fakeHome struct {
	at      string
	took    []string
	restore []string
}

func (h *fakeHome) Installed(string) (string, bool, error) {
	if h.at == "" {
		return filepath.FromSlash("C:/home/.agents/skills/tests"), false, nil
	}
	return h.at, true, nil
}
func (h *fakeHome) Take(name string) (domain.Mark, error) {
	h.took = append(h.took, name)
	return domain.Mark{Skill: name, From: filepath.Join(h.at), To: h.at + ".aside"}, nil
}
func (h *fakeHome) Restore(mark domain.Mark) error {
	h.restore = append(h.restore, mark.Skill)
	return nil
}
func (h *fakeHome) Recover() ([]string, error) { return nil, nil }

func exam() corpus.Exam {
	return corpus.Exam{
		Skill: "tests",
		Task:  "work T1 of the task document\n",
		Files: map[string][]byte{"go.mod": []byte("module exam\n")},
	}
}

func skill() []corpus.File {
	return []corpus.File{{Path: "SKILL.md", Content: []byte("# tests\n")}}
}

// TestTheTwoArmsScoreTheSamePrompt: both arms get the same exam, and the score is
// the difference between their streams -two behaviours for the one that tested
// first, none for the one that wrote the code first.
func TestTheTwoArmsScoreTheSamePrompt(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{}
	got := skilleval.Run(skilleval.Options{
		Scratch: t.TempDir(),
		Exam:    exam(),
		Skill:   skill(),
		Exec:    "agent {prompt}",
		Runs:    1,
		Timeout: 60,
	}, &fakeFiles{}, shell, &fakeHome{})

	if got.Code != skilleval.CodeOK {
		t.Fatalf("Code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if len(got.Draws) != 2 {
		t.Fatalf("draws = %d, want one per arm", len(got.Draws))
	}
	if domain.Met(got.Draws[0].Behaviours) != 2 || !got.Draws[0].Loaded {
		t.Errorf("the arm with the skill = %+v, want both behaviours and the load", got.Draws[0])
	}
	if domain.Met(got.Draws[1].Behaviours) != 0 {
		t.Errorf("the arm without the skill = %+v, want no behaviour", got.Draws[1])
	}

	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "with: 2 of 2") || !strings.Contains(joined, "without: 0 of 2") {
		t.Errorf("the summary does not carry the two arms:\n%s", joined)
	}
	if len(shell.dirs) != 2 {
		t.Fatalf("the command ran %d times, want once per arm", len(shell.dirs))
	}
	if !strings.Contains(shell.commands[0], filepath.Join(got.Draws[0].Arm)) {
		t.Errorf("the command did not receive the prompt path: %s", shell.commands[0])
	}
}

// TestASkillInTheHomeNeedsPermission: the arm without the skill is the folder
// absent, so a skill installed in the home is a refusal that names it until the
// human says --move-home.
func TestASkillInTheHomeNeedsPermission(t *testing.T) {
	t.Parallel()

	home := &fakeHome{at: filepath.FromSlash("C:/home/.agents/skills/tests")}
	shell := &fakeShell{}
	got := skilleval.Run(skilleval.Options{
		Scratch: t.TempDir(),
		Exam:    exam(),
		Skill:   skill(),
		Exec:    "agent {prompt}",
		Runs:    1,
	}, &fakeFiles{}, shell, home)

	if got.Code != skilleval.CodeBlocked {
		t.Fatalf("Code = %d, want blocked", got.Code)
	}
	if !strings.Contains(got.Message, home.at) || !strings.Contains(got.Message, "--move-home") {
		t.Errorf("the refusal does not name the folder and the way out: %s", got.Message)
	}
	if len(shell.dirs) != 0 {
		t.Error("a reviewer ran before the home was settled")
	}
	if len(home.took) != 0 {
		t.Error("the skill was taken without permission")
	}
}

// TestTheSkillGoesBackAfterTheRun: the permission covers one run, not the
// machine: whatever the draws did, the folder is home again when the verb
// answers.
func TestTheSkillGoesBackAfterTheRun(t *testing.T) {
	t.Parallel()

	home := &fakeHome{at: filepath.FromSlash("C:/home/.agents/skills/tests")}
	got := skilleval.Run(skilleval.Options{
		Scratch:  t.TempDir(),
		Exam:     exam(),
		Skill:    skill(),
		Exec:     "agent {prompt}",
		Runs:     1,
		Timeout:  60,
		MoveHome: true,
	}, &fakeFiles{}, &fakeShell{}, home)

	if got.Code != skilleval.CodeOK {
		t.Fatalf("Code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if len(home.took) != 1 || len(home.restore) != 1 || home.restore[0] != "tests" {
		t.Fatalf("took = %v, restore = %v; want one of each for tests", home.took, home.restore)
	}
}

// TestAnAnswerAlreadyPaidForIsNotPaidAgain: the record is the receipt. The same
// skill text over the same exam is refused, and --again is the human saying the
// second run is worth it.
func TestAnAnswerAlreadyPaidForIsNotPaidAgain(t *testing.T) {
	t.Parallel()

	files := &fakeFiles{}
	record := filepath.FromSlash("C:/store/evals/tests-abc123.json")
	opts := skilleval.Options{
		Scratch:  t.TempDir(),
		Exam:     exam(),
		Skill:    skill(),
		Exec:     "agent {prompt}",
		Runs:     1,
		Timeout:  60,
		Record:   record,
		Identity: "abc123",
		Now:      time.Unix(0, 0),
	}

	first := skilleval.Run(opts, files, &fakeShell{}, &fakeHome{})
	if first.Code != skilleval.CodeOK {
		t.Fatalf("the first run = %d, want 0 (%s)", first.Code, first.Message)
	}
	if _, ok := files.written[record]; !ok {
		t.Fatalf("the run filed no record; it wrote %v", keys(files.written))
	}

	second := skilleval.Run(opts, files, &fakeShell{}, &fakeHome{})
	if second.Code != skilleval.CodeBlocked {
		t.Fatalf("the second run = %d, want blocked", second.Code)
	}
	if !strings.Contains(second.Message, "--again") || !strings.Contains(second.Message, "abc123") {
		t.Errorf("the refusal does not say what was paid and how to override it: %s", second.Message)
	}

	opts.Again = true
	if third := skilleval.Run(opts, files, &fakeShell{}, &fakeHome{}); third.Code != skilleval.CodeOK {
		t.Fatalf("the run with --again = %d, want 0 (%s)", third.Code, third.Message)
	}
}

func keys(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}

// TestNoCommandPrintsThePrompt: without --exec the verb asks the question and
// runs nobody, the way the probe does.
func TestNoCommandPrintsThePrompt(t *testing.T) {
	t.Parallel()

	shell := &fakeShell{}
	got := skilleval.Run(skilleval.Options{Scratch: t.TempDir(), Exam: exam(), Skill: skill()}, &fakeFiles{}, shell, &fakeHome{})

	if got.Code != skilleval.CodeOK {
		t.Fatalf("Code = %d, want 0 (%s)", got.Code, got.Message)
	}
	if len(shell.dirs) != 0 {
		t.Error("a command ran without --exec")
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "work T1 of the task document") {
		t.Errorf("the prompt is not printed:\n%s", strings.Join(got.Lines, "\n"))
	}
}
