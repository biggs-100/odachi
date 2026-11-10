package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// seedTUIProject writes the four facts the first screen reads: the block, the stored
// choice, the voice and the review file.
func seedTUIProject(t *testing.T, dir string) {
	t.Helper()
	initRepo(t, dir)

	block, err := rules.WithBlock("", persona.Novato)
	if err != nil {
		t.Fatalf("rules.WithBlock: %v", err)
	}
	writeFile(t, dir, "AGENTS.md", block)

	store, err := agentstoml.Render(persona.Store{Agents: persona.SelectionOpenCode, Skills: persona.SkillsProject})
	if err != nil {
		t.Fatalf("agentstoml.Render: %v", err)
	}
	writeFile(t, dir, ".git/odachi/agents.toml", string(store))

	writeFile(t, dir, filepath.FromSlash("odd/profile.toml"), "name = \"senior\"\n")

	config, err := reviewtoml.Render(review.ProjectConfig{Mode: review.ModeOn})
	if err != nil {
		t.Fatalf("reviewtoml.Render: %v", err)
	}
	writeFile(t, dir, filepath.FromSlash(review.ModePath), string(config))
}

// TestTUIDetectsTheState: the first screen starts from what the project already
// chose, not from the wizard's defaults.
func TestTUIDetectsTheState(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}

	want := tuiState{
		Version: version,
		Project: true,
		Clients: persona.SelectionOpenCode,
		Skills:  persona.SkillsProject,
		Voice:   "senior",
		Review:  review.ModeOn,
	}
	if got != want {
		t.Fatalf("state = %+v, want %+v", got, want)
	}
}

// TestTUIDetectsAFreshFolder: a folder odachi never touched reads as not a
// project, with the machine's defaults and the review off.
func TestTUIDetectsAFreshFolder(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	got, err := detectTUIState(dir)
	if err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}

	if got.Project {
		t.Error("Project = true, want false: there is no block")
	}
	if got.Clients != persona.SelectionBoth {
		t.Errorf("Clients = %q, want both", got.Clients)
	}
	if got.Skills != persona.SkillsHome {
		t.Errorf("Skills = %q, want home", got.Skills)
	}
	if got.Voice != "novato" {
		t.Errorf("Voice = %q, want the inferred novato", got.Voice)
	}
	if got.Review != review.ModeOff {
		t.Errorf("Review = %q, want off", got.Review)
	}
}

// TestTUIFirstScreenNamesTheState: the first thing a human reads is where they
// are and what odachi will configure.
func TestTUIFirstScreenNamesTheState(t *testing.T) {
	view := newTUIModel(tuiState{
		Version: "9.9.9",
		Project: true,
		Clients: persona.SelectionBoth,
		Skills:  persona.SkillsHome,
		Voice:   "novato",
		Review:  review.ModeOn,
	}).View()

	for _, want := range []string{
		"odachi 9.9.9", "instalación y configuración",
		"Proyecto", "sí", "Clientes", "both", "Skills", "home", "Voz", "novato", "Review", "on",
		"q salir",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("the first screen does not carry %q:\n%s", want, view)
		}
	}
}

// TestTUIDetectionWritesNothing: opening the wizard is a read. This tree is the
// one the human will configure, and looking at it must not move it.
func TestTUIDetectionWritesNothing(t *testing.T) {
	dir := t.TempDir()
	seedTUIProject(t, dir)
	before := manifest(t, dir)

	if _, err := detectTUIState(dir); err != nil {
		t.Fatalf("detectTUIState: %v", err)
	}
	newTUIModel(tuiState{Version: version}).View()

	if after := manifest(t, dir); after != before {
		t.Errorf("the tree moved while detecting:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestTUIRefusesWithoutATerminal: a pipe is not a screen, and the refusal names
// the command that configures without one.
func TestTUIRefusesWithoutATerminal(t *testing.T) {
	code, out, errOut := runDir(t, t.TempDir(), []string{"tui"}, "")

	if code == 0 {
		t.Fatalf("code = 0, want blocked (stdout %q)", out)
	}
	message := out + errOut
	for _, want := range []string{"terminal", "odachi setup"} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal does not name %q:\n%s", want, message)
		}
	}
}

// TestTUIWithAFlagIsRefused: the wizard has no flags; the flags belong to the
// verbs it applies.
func TestTUIWithAFlagIsRefused(t *testing.T) {
	code, _, errOut := runDir(t, t.TempDir(), []string{"tui", "--json"}, "")

	if code == 0 {
		t.Fatal("code = 0, want blocked")
	}
	if !strings.Contains(errOut, "takes no flags") {
		t.Errorf("the refusal does not say the verb takes no flags:\n%s", errOut)
	}
}
