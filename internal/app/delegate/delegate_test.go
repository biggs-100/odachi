package delegate_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/delegate"
)

// doc is a task document with one delegated step that records its craft.
const doc = `# demo

## Requirements (verbatim, from the user)
R1 · "implementalo"

## Steps
- [ ] T1 · implementar el parser   R1 · route:delegated/odachi-worker · skills:tests,drive
- [ ] T2 · escribir la nota   R1 · route:inline
`

type fakeFiles map[string][]byte

func (f fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f[path]
	if !ok {
		return nil, errors.New("no such file")
	}
	return data, nil
}

type fakeResolver map[string]string

func (r fakeResolver) Path(name string) (string, bool) {
	path, ok := r[name]
	return path, ok
}

func (r fakeResolver) Names() []string {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	return names
}

// setup is the common wiring: one project root, one task doc, two skills.
func setup(t *testing.T) (delegate.Options, fakeFiles) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	return delegate.Options{
		Dir:      dir,
		TaskDoc:  path,
		Step:     "T1",
		Surfaces: []string{"config/config.go", "config/config_test.go"},
		Commands: []string{"go test ./..."},
		Resolver: fakeResolver{
			"tests": `C:\skills\tests\SKILL.md`,
			"drive": `C:\skills\drive\SKILL.md`,
		},
	}, fakeFiles{path: []byte(doc)}
}

// TestComposeCarriesTheStepAndTheCraft: the prompt names the document, the step,
// the verbatim requirement, the surfaces, the commands and the exact skill paths,
// so the parent writes no section by hand.
func TestComposeCarriesTheStepAndTheCraft(t *testing.T) {
	opts, files := setup(t)

	result := delegate.Run(opts, files)
	if result.Code != delegate.CodeOK {
		t.Fatalf("Run: %s", result.Message)
	}

	for _, want := range []string{
		"# T1 · implementar el parser",
		"Task doc: " + opts.TaskDoc,
		"Step: T1 · implementar el parser",
		"Requirement R1 (verbatim):",
		"> implementalo",
		"- config/config.go",
		"- go test ./...",
		"## Skills to load before work",
		"Read these exact files before reading, writing or testing:",
		`- C:\skills\tests\SKILL.md`,
		`- C:\skills\drive\SKILL.md`,
		"Return your report in the format your prompt names.",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, result.Prompt)
		}
	}
	// The section is always present: an absent one is what the worker refuses.
	if strings.Index(result.Prompt, "## Skills to load before work") > strings.Index(result.Prompt, "Return your report") {
		t.Error("the skills section is not before the closing line")
	}
}

// TestAStepWithoutSkillsSaysNone: a section that says there are none lets the
// worker start; an absent section does not.
func TestAStepWithoutSkillsSaysNone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	plain := strings.Replace(doc, " · skills:tests,drive", "", 1)
	opts := delegate.Options{Dir: dir, TaskDoc: path, Step: "T1"}
	result := delegate.Run(opts, fakeFiles{path: []byte(plain)})

	if result.Code != delegate.CodeOK {
		t.Fatalf("Run: %s", result.Message)
	}
	if !strings.Contains(result.Prompt, "None: this step owes no skill.") {
		t.Errorf("the prompt does not say none:\n%s", result.Prompt)
	}
}

// TestAnInlineStepIsRefused: the parent runs it, so there is no hand-off to write.
func TestAnInlineStepIsRefused(t *testing.T) {
	opts, files := setup(t)
	opts.Step = "T2"

	result := delegate.Run(opts, files)
	if result.Code != delegate.CodeBlocked {
		t.Fatal("an inline step got a prompt")
	}
	if !strings.Contains(result.Message, "inline") {
		t.Errorf("message %q does not name the route", result.Message)
	}
}

// TestAMissingSkillIsNamed: the parent fixes the plan, and the error says which
// name failed and which ones exist.
func TestAMissingSkillIsNamed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "task.md")
	opts := delegate.Options{Dir: dir, TaskDoc: path, Step: "T1", Resolver: fakeResolver{"drive": `C:\skills\drive\SKILL.md`}}

	result := delegate.Run(opts, fakeFiles{path: []byte(doc)})
	if result.Code != delegate.CodeBlocked {
		t.Fatal("a missing skill produced a prompt")
	}
	if !strings.Contains(result.Message, `"tests"`) || !strings.Contains(result.Message, "drive") {
		t.Errorf("message %q does not name the missing skill and the known ones", result.Message)
	}
}

// TestAMissingStepListsTheSteps: the parent knows what the document has without
// opening it again.
func TestAMissingStepListsTheSteps(t *testing.T) {
	opts, files := setup(t)
	opts.Step = "T9"

	result := delegate.Run(opts, files)
	if result.Code != delegate.CodeBlocked {
		t.Fatal("an unknown step produced a prompt")
	}
	if !strings.Contains(result.Message, "T1") || !strings.Contains(result.Message, "T2") {
		t.Errorf("message %q does not list the steps", result.Message)
	}
}
