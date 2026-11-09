package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// routerProjectMain is a project skill with one reference, in the shape the
// check accepts: the output contract and the inventory.
const routerProjectMain = "---\nname: router\ndescription: Trigger: cuando hay que enrutar\n---\n\n## When\n\ncuando\n\n## Rules\n\n- regla\n\n## Never\n\nnunca\n\n## Output\n\nthe file.\n\n## References\n\n- `references/numbers.md` — the scale.\n"

// TestPackInstallsAProjectSkillWithItsReferences: the project's own catalog is a
// folder too, so the join from ./skills to the agent's directory carries the
// references, not only the SKILL.md.
func TestPackInstallsAProjectSkillWithItsReferences(t *testing.T) {
	home := t.TempDir()
	withHome(t, home)
	dir := t.TempDir()

	writeFile(t, dir, "skills/router/SKILL.md", routerProjectMain)
	writeFile(t, dir, "skills/router/references/numbers.md", "the scale\n")

	if code, _, errOut := runDir(t, dir, []string{"skills", "--write"}, ""); code != 0 {
		t.Fatalf("skills --write code = %d (stderr %q)", code, errOut)
	}

	code, out, errOut := runDir(t, dir, []string{"pack", "--agents", "claude"}, "")
	if code != 0 {
		t.Fatalf("pack code = %d (stderr %q)", code, errOut)
	}

	for file, want := range map[string]string{
		"SKILL.md":              routerProjectMain,
		"references/numbers.md": "the scale\n",
	} {
		path := filepath.Join(home, ".claude", "skills", "router", filepath.FromSlash(file))
		got, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s was not installed: %v", file, err)
			continue
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", file, got, want)
		}
	}
	if !strings.Contains(out, "router/") {
		t.Errorf("the run does not report the project skill as a folder:\n%s", out)
	}
}
