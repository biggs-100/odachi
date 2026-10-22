package skills_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// bundled is every skill directory in this package's directory, read from disk
// instead of listed here: adding a skill must not require editing this file.
var bundled = func() []string {
	entries, err := os.ReadDir(".")
	if err != nil {
		panic("read skills directory: " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names
}()

// TestAllCarriesEveryStarterSkill: the binary knows them by name, sorted, so
// `pack` writes the same set on every machine and in the same order.
func TestAllCarriesEveryStarterSkill(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.Name)
		if len(s.Content) == 0 {
			t.Errorf("%s came back empty", s.Name)
		}
	}
	if !slices.Equal(names, bundled) {
		t.Fatalf("All() = %v, want %v", names, bundled)
	}
	if got := skills.Names(); !slices.Equal(got, bundled) {
		t.Errorf("Names() = %v, want %v", got, bundled)
	}
}

// TestEmbeddedBytesAreTheFilesInTheRepo: the embed is the source, not a copy of
// it. If a skill is edited on disk and the bytes here do not match, the binary
// is shipping a version nobody reviewed.
func TestEmbeddedBytesAreTheFilesInTheRepo(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	for _, s := range all {
		onDisk, err := os.ReadFile(s.Name + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s/SKILL.md: %v", s.Name, err)
		}
		if string(s.Content) != string(onDisk) {
			t.Errorf("%s: the embedded bytes differ from %s/SKILL.md", s.Name, s.Name)
		}
	}
}

// TestDelegationPolicyLivesInSlicesAndDrive: the policy that decides when a
// subagent is used is content, and content drifts. `slices` owns the triggers and
// the handoff; `drive` is what the block loads before acting, so it is the one
// that has to send the agent to `slices` when the request does not fit.
func TestDelegationPolicyLivesInSlicesAndDrive(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	for _, want := range []string{
		"gana esta",
		"Nunca por tamaño",
		"path del task doc",
		"≤2k tokens",
		"Nunca delegues la decisión de visión",
	} {
		if !strings.Contains(byName["slices"], want) {
			t.Errorf("slices no dice %q: la política de delegación se perdió", want)
		}
	}
	if !strings.Contains(byName["drive"], "cargá `slices` antes de partir") {
		t.Errorf("drive no manda a cargar slices cuando el pedido no entra en una lectura")
	}
}

// TestBootstrapIsTheEntryPoint: the payload has exactly one skill that teaches
// how to start a project, and it is the one `odachi init` installs in the
// human's home. Splitting it out of the curriculum is the whole reason the
// helper exists: a project carries craft, the home carries how to arrive.
func TestBootstrapIsTheEntryPoint(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatalf("el payload no tiene la skill %q: un agente no puede aprender a arrancar", skills.BootstrapName)
	}
	if bootstrap.Name != skills.BootstrapName {
		t.Errorf("Bootstrap() = %q, want %q", bootstrap.Name, skills.BootstrapName)
	}
	if len(bootstrap.Content) == 0 {
		t.Error("la skill de arranque vino vacía")
	}
}

// TestCurriculumIsEverySkillButTheEntryPoint: the project copy is the craft and
// nothing else, so init never installs the same two skills in both roots.
func TestCurriculumIsEverySkillButTheEntryPoint(t *testing.T) {
	curriculum, err := skills.Curriculum()
	if err != nil {
		t.Fatalf("Curriculum(): %v", err)
	}

	want := make([]string, 0, len(bundled))
	for _, name := range bundled {
		if name != skills.BootstrapName {
			want = append(want, name)
		}
	}
	got := make([]string, 0, len(curriculum))
	for _, s := range curriculum {
		got = append(got, s.Name)
		if s.Name == skills.BootstrapName {
			t.Errorf("el currículum del proyecto incluye la skill de arranque")
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Curriculum() = %v, want %v", got, want)
	}
}
