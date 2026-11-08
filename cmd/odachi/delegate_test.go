package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// delegatedProject writes a task doc with one delegated step that owes `tests`,
// and installs that skill in the project, so the delegation resolves it.
func delegatedProject(t *testing.T) (dir, taskPath, skillPath string) {
	t.Helper()
	dir = t.TempDir()

	craft := filepath.Join(dir, ".agents", "skills", "tests")
	if err := os.MkdirAll(craft, 0o755); err != nil {
		t.Fatalf("mkdir craft: %v", err)
	}
	skillPath = filepath.Join(craft, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("---\nname: tests\ndescription: x\n---\n"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	tasks := filepath.Join(dir, "odd", "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatalf("mkdir tasks: %v", err)
	}
	taskPath = filepath.Join(tasks, "0001-x.md")
	doc := `# demo

## Requirements (verbatim, from the user)
R1 · "implementalo"

## Steps
- [ ] T1 · implementar el parser   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T2 · escribir la nota   R1 · route:inline
`
	if err := os.WriteFile(taskPath, []byte(doc), 0o644); err != nil {
		t.Fatalf("write task doc: %v", err)
	}
	return dir, filepath.Join("odd", "tasks", "0001-x.md"), skillPath
}

// TestDelegateComposesThePromptForADelegatedStep: one command prints the whole
// hand-off, and the skills section is part of it.
func TestDelegateComposesThePromptForADelegatedStep(t *testing.T) {
	t.Parallel()
	dir, taskPath, skillPath := delegatedProject(t)

	code, stdout, stderr := runDir(t, dir, []string{"delegate", taskPath, "--step", "T1", "--surface", "config/config.go", "--command", "go test ./..."}, "")
	if code != 0 {
		t.Fatalf("delegate: exit %d, stderr=%s", code, stderr)
	}

	for _, want := range []string{
		"# T1 · implementar el parser",
		"Requirement R1 (verbatim):",
		"> implementalo",
		"- config/config.go",
		"- go test ./...",
		"## Skills to load before work",
		"- " + skillPath,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the prompt does not carry %q:\n%s", want, stdout)
		}
	}
}

// TestDelegateRefusesAnInlineStep: a step the parent runs has no hand-off.
func TestDelegateRefusesAnInlineStep(t *testing.T) {
	t.Parallel()
	dir, taskPath, _ := delegatedProject(t)

	code, _, stderr := runDir(t, dir, []string{"delegate", taskPath, "--step", "T2"}, "")
	if code == 0 {
		t.Fatal("an inline step got a prompt")
	}
	if !strings.Contains(stderr, "inline") {
		t.Errorf("stderr %q does not name the route", stderr)
	}
}
