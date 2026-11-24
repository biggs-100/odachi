package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

func isolatedHome(t *testing.T) string {
	t.Helper()
	return withTempHome(t)
}

// TestGlobalProfileRoundTrip: the machine default is one file outside every
// repository, and only a profile this binary ships can be written there (0357).
func TestGlobalProfileRoundTrip(t *testing.T) {
	isolatedHome(t)
	if name, found, err := loadGlobalProfile(); err != nil || found || name != "" {
		t.Fatalf("empty home = %q, %v, %v; want not found", name, found, err)
	}
	if err := saveGlobalProfile("office-peru"); err != nil {
		t.Fatalf("saveGlobalProfile: %v", err)
	}
	name, found, err := loadGlobalProfile()
	if err != nil || !found || name != "office-peru" {
		t.Fatalf("loadGlobalProfile = %q, %v, %v; want office-peru, found", name, found, err)
	}
	if err := saveGlobalProfile("not-a-profile"); err == nil {
		t.Fatal("saveGlobalProfile accepted an unknown profile")
	}
}

// TestResolveProfileOrder: the project wins over the published block, the
// block wins over the machine default, and the default wins over inference.
func TestResolveProfileOrder(t *testing.T) {
	home := isolatedHome(t)
	dir := t.TempDir()

	// With nothing written, the machine default decides over inference.
	if err := saveGlobalProfile("office-peru"); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, explicit, err := resolveProfile(dir, "odachi test", false)
	if err != nil || explicit || got.Name != "office-peru" {
		t.Fatalf("default = %+v, explicit %v, err %v; want office-peru", got, explicit, err)
	}

	// A voice already published in AGENTS.md is a project choice.
	block, err := rules.BlockFor(persona.Novato, persona.SkillsHome)
	if err != nil {
		t.Fatalf("BlockFor: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(block), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	got, explicit, err = resolveProfile(dir, "odachi test", false)
	if err != nil || !explicit || got.Name != "novato" {
		t.Fatalf("block = %+v, explicit %v, err %v; want novato", got, explicit, err)
	}

	// The project's file is the last word over both.
	if err := os.MkdirAll(filepath.Join(dir, "odd"), 0o755); err != nil {
		t.Fatalf("mkdir odd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "odd", "profile.toml"), []byte("name = \"ci\"\n"), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	got, explicit, err = resolveProfile(dir, "odachi test", false)
	if err != nil || !explicit || got.Name != "ci" {
		t.Fatalf("project = %+v, explicit %v, err %v; want ci", got, explicit, err)
	}

	// Nothing at any level falls back to inference; a brief means senior.
	clean := t.TempDir()
	if err := os.WriteFile(filepath.Join(clean, ".odachi-does-not-exist"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Remove(filepath.Join(home, ".odachi", "profile.toml")); err != nil {
		t.Fatalf("remove global: %v", err)
	}
	got, explicit, err = resolveProfile(clean, "odachi test", false)
	if err != nil || explicit || got.Name != "novato" {
		t.Fatalf("inference = %+v, explicit %v, err %v; want novato", got, explicit, err)
	}
}
