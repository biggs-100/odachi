package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resolveTree seeds a project whose only feature declares three steps: one with
// a project skill and a home one, one with no skills, and one with a name that
// exists nowhere. The home is a temp directory, never the real one.
func resolveTree(t *testing.T) (dir, home string) {
	t.Helper()
	dir = t.TempDir()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	write := func(root, rel, body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	write(dir, "odd/tasks/0001-demo.md", `# demo
Base: abc1234

## Requirements (verbatim, from the user)
R1 · "demo"

## Done when

## Steps
- [ ] T1 · one   R1 · route:delegated/odachi-worker · skills:tests,drive
- [ ] T2 · two   R1 · route:inline
- [ ] T3 · three   R1 · route:delegated/odachi-worker · skills:ghost

## History
L1 · "demo"
`)
	write(dir, ".agents/skills/tests/SKILL.md", "---\nname: tests\n---\n")
	write(home, ".agents/skills/drive/SKILL.md", "---\nname: drive\n---\n")
	return dir, home
}

// TestSkillsResolveNamesTheProjectCopyAndTheHomeOne: the surface answers with the
// exact path and the scope for each name the step declares, in the step's order.
func TestSkillsResolveNamesTheProjectCopyAndTheHomeOne(t *testing.T) {
	dir, home := resolveTree(t)

	code, out, errOut := runDir(t, dir, []string{"skills", "--resolve", "T1"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stdout %q, stderr %q)", code, out, errOut)
	}
	projectPath := filepath.ToSlash(filepath.Join(dir, ".agents", "skills", "tests", "SKILL.md"))
	homePath := filepath.ToSlash(filepath.Join(home, ".agents", "skills", "drive", "SKILL.md"))
	for _, want := range []string{"tests", projectPath, "(project)", "drive", homePath, "(home)"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not carry %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "tests") > strings.Index(out, "drive") {
		t.Errorf("the names are not in the step's order:\n%s", out)
	}
}

// TestSkillsResolveSaysWhenASkillIsMissing: a name no scope carries is reported
// with both places it looked, and the run is blocked because the hand-off it
// describes would fail.
func TestSkillsResolveSaysWhenASkillIsMissing(t *testing.T) {
	dir, home := resolveTree(t)

	code, out, errOut := runDir(t, dir, []string{"skills", "--resolve", "T3"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want blocked: a skill that exists nowhere is a broken hand-off")
	}
	projectPath := filepath.ToSlash(filepath.Join(dir, ".agents", "skills", "ghost", "SKILL.md"))
	homePath := filepath.ToSlash(filepath.Join(home, ".agents", "skills", "ghost", "SKILL.md"))
	message := out + errOut
	for _, want := range []string{"ghost", projectPath, homePath} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, message)
		}
	}
}

// TestSkillsResolveSaysWhenTheStepNamesNone: an empty list is an answer, not a
// silence, and it exits 0.
func TestSkillsResolveSaysWhenTheStepNamesNone(t *testing.T) {
	dir, _ := resolveTree(t)

	code, out, errOut := runDir(t, dir, []string{"skills", "--resolve", "T2"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	if !strings.Contains(out, "T2") || !strings.Contains(out, "no skills") {
		t.Errorf("the answer does not say the step names none:\n%s", out)
	}
}

// TestSkillsResolveNamesTheStepsWhenTheIDIsUnknown: an id the document does not
// have is refused with the ones it does, so the next run is one line.
func TestSkillsResolveNamesTheStepsWhenTheIDIsUnknown(t *testing.T) {
	dir, _ := resolveTree(t)

	code, out, errOut := runDir(t, dir, []string{"skills", "--resolve", "T9"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want blocked: T9 is not a step")
	}
	message := out + errOut
	for _, want := range []string{"T9", "T1", "T3"} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, message)
		}
	}
}

// TestSkillsResolveWritesNothing: the surface is read-only, so every byte of the
// tree is the same after the run.
func TestSkillsResolveWritesNothing(t *testing.T) {
	dir, _ := resolveTree(t)
	before := manifest(t, dir)

	code, _, errOut := runDir(t, dir, []string{"skills", "--resolve", "T1"}, "")

	if code != 0 {
		t.Fatalf("code = %d, want 0 (stderr %q)", code, errOut)
	}
	if after := manifest(t, dir); after != before {
		t.Errorf("the tree moved:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestSkillsResolveRefusesTheOtherIntents: --resolve is one intent, and mixing it
// with --check or --write is refused before anything is read.
func TestSkillsResolveRefusesTheOtherIntents(t *testing.T) {
	dir, _ := resolveTree(t)

	for _, other := range []string{"--check", "--write"} {
		t.Run(other, func(t *testing.T) {
			code, _, errOut := runDir(t, dir, []string{"skills", "--resolve", "T1", other}, "")
			if code == 0 {
				t.Fatalf("code = 0, want blocked")
			}
			if !strings.Contains(errOut, "--resolve") || !strings.Contains(errOut, other) {
				t.Errorf("the refusal does not name both flags (%q):\n%s", other, errOut)
			}
		})
	}
}

// manifest is every path and content digest under dir, in walk order.
func manifest(t *testing.T, dir string) string {
	t.Helper()
	var out strings.Builder
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			fmt.Fprintf(&out, "%s/\n", filepath.ToSlash(rel))
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "%s %x\n", filepath.ToSlash(rel), sha256.Sum256(data))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out.String()
}
