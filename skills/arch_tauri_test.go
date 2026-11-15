package skills_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// TestArchTauriKeepsTheBoundaryRules: the skill this catalog carries for a
// Tauri v2 project is the one thing that keeps the framework out of the domain. The
// rules are pinned because they are the whole value: a body that lost the ACL gate
// or the generated contract would still look like a foundation (0302).
func TestArchTauriKeepsTheBoundaryRules(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	content, ok := byName["arch-tauri"]
	if !ok {
		t.Fatal("the payload has no arch-tauri skill: a desktop project has no door")
	}
	for _, want := range []string{
		"never knows Tauri exists",
		"thin adapter",
		"src/services/",
		"tauri-specta",
		"serializable",
		"source-scan",
		"has_app_acl_manifest",
		"three layers",
		"bootstrap",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("arch-tauri does not say %q", want)
		}
	}
}

// TestArchTauriCarriesItsDepth: the body is a door, and the four pieces are
// what a person actually follows. A reference the body names and the folder does not
// hold is a dead link in a skill that installs on every machine (0302).
func TestArchTauriCarriesItsDepth(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var skill skills.Skill
	for _, s := range all {
		if s.Name == "arch-tauri" {
			skill = s
		}
	}
	if skill.Name == "" {
		t.Fatal("the payload has no arch-tauri skill")
	}

	files := map[string]bool{}
	for _, file := range skill.Files {
		files[file.Path] = true
	}
	for _, want := range []string{
		"references/tauri-scaffold.md",
		"references/tauri-bridge.md",
		"references/rust-guard-recipes.md",
		"references/ui-verification.md",
		"assets/tauri-layout.txt",
	} {
		if !files[want] {
			t.Errorf("arch-tauri does not carry %q; it carries %v", want, files)
		}
	}
}

// TestArchTauriScaffoldKeepsTheOrder: the nine steps existed in the retired skill and
// the port lost them, so the references had every piece and no sequence. The steps are
// pinned because the order is the rule: the guard before the business code, the error
// type before the first command, the capability with the permission file (0308).
func TestArchTauriScaffoldKeepsTheOrder(t *testing.T) {
	skill := mustSkill(t, "arch-tauri")

	var scaffold string
	for _, file := range skill.Files {
		if file.Path == "references/tauri-scaffold.md" {
			scaffold = string(file.Content)
		}
	}
	if scaffold == "" {
		t.Fatal("arch-tauri carries no references/tauri-scaffold.md")
	}

	for _, want := range []string{
		"1. **Scaffold the workspace**",
		"2. **Write the domain-purity guard before the business code**",
		"3. **Define the serializable error type**",
		"4. **Write one thin command as the template**",
		"5. **Generate the TypeScript bindings from Rust**",
		"6. **Create the single service layer**",
		"7. **Wire the three verification levels**",
		"8. **Write the first ADR**",
		"9. **If a file landed under `src-tauri/permissions/`",
	} {
		if !strings.Contains(scaffold, want) {
			t.Errorf("the scaffold lost a step: %q", want)
		}
	}

	// The two details a person forgets and the compiler does not catch.
	for _, want := range []string{"build.rs", "generate_handler!", "no test specified"} {
		if !strings.Contains(scaffold, want) {
			t.Errorf("the scaffold does not name %q", want)
		}
	}
}
