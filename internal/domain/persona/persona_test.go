package persona_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestNovatoTeachesAndSeniorGoesToThePoint: the two personas are two voices, not
// two spellings of the same text.
func TestNovatoTeachesAndSeniorGoesToThePoint(t *testing.T) {
	novato, err := persona.Voice(persona.Novato)
	if err != nil {
		t.Fatalf("Voice(novato): %v", err)
	}
	senior, err := persona.Voice(persona.Senior)
	if err != nil {
		t.Fatalf("Voice(senior): %v", err)
	}

	for _, want := range []string{"aprenda el oficio", "odachi ask --craft", "odachi summary", "preguntá de vuelta", "sermón"} {
		if !strings.Contains(novato, want) {
			t.Errorf("la voz novato no contiene %q:\n%s", want, novato)
		}
	}
	for _, want := range []string{"salvo que la pidan", "breve", "prueba"} {
		if !strings.Contains(senior, want) {
			t.Errorf("la voz senior no contiene %q:\n%s", want, senior)
		}
	}
	if strings.Contains(senior, "odachi ask --craft") {
		t.Errorf("la voz senior repite la ceremonia del novato:\n%s", senior)
	}
	if novato == senior {
		t.Errorf("las dos voces son el mismo texto")
	}
}

// TestVoicesEndInExactlyOneNewline: the renderer puts the text straight after the
// section's blank line and straight before the closing marker, so the newlines
// are the renderer's to compose and the text's to get right.
func TestVoicesEndInExactlyOneNewline(t *testing.T) {
	for _, name := range persona.Names() {
		voice, err := persona.Voice(name)
		if err != nil {
			t.Fatalf("Voice(%q): %v", name, err)
		}
		if !strings.HasSuffix(voice, "\n") || strings.HasSuffix(voice, "\n\n") {
			t.Errorf("%s: la voz no termina en exactamente un salto de línea: %q", name, voice)
		}
		if strings.HasPrefix(voice, "\n") {
			t.Errorf("%s: la voz arranca con un salto de línea: %q", name, voice)
		}
	}
}

// TestCIUsesTheSeniorVoice: ci is a profile, not a person. There is nobody to
// teach in a pipeline, and teaching an agent that is not a person is noise.
func TestCIUsesTheSeniorVoice(t *testing.T) {
	name, err := persona.VoiceName(persona.CI)
	if err != nil {
		t.Fatalf("VoiceName(ci): %v", err)
	}
	if name != persona.Senior {
		t.Fatalf("VoiceName(ci) = %q, want %q", name, persona.Senior)
	}

	ci, err := persona.Voice(persona.CI)
	if err != nil {
		t.Fatalf("Voice(ci): %v", err)
	}
	senior, err := persona.Voice(persona.Senior)
	if err != nil {
		t.Fatalf("Voice(senior): %v", err)
	}
	if ci != senior {
		t.Errorf("la voz de ci no es la de senior:\n%s", ci)
	}
}

// TestVoiceNameTrimsAndIgnoresCase: a name arrives from a TOML file and from a
// flag, and a human types both.
func TestVoiceNameTrimsAndIgnoresCase(t *testing.T) {
	for input, want := range map[string]string{
		"  Novato ": persona.Novato,
		"SENIOR":    persona.Senior,
		"ci":        persona.Senior,
	} {
		got, err := persona.VoiceName(input)
		if err != nil {
			t.Fatalf("VoiceName(%q): %v", input, err)
		}
		if got != want {
			t.Errorf("VoiceName(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestUnknownVoiceIsRefusedNamingWhatExists: nothing is guessed for a name that
// is not a voice, and the error says which two there are.
func TestUnknownVoiceIsRefusedNamingWhatExists(t *testing.T) {
	_, err := persona.Voice("principiante")
	if err == nil {
		t.Fatalf("Voice(principiante) devolvió una voz que no existe")
	}
	for _, want := range []string{"principiante", persona.Novato, persona.Senior} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("el error no nombra %q: %v", want, err)
		}
	}
}

// TestDeprecatedIsTheV0100Install: the files the previous version wrote, named
// exactly, because the report and the check both point at them and the fix is the
// human deleting them.
func TestDeprecatedIsTheV0100Install(t *testing.T) {
	want := []string{
		".opencode/agents/odachi-novato.md",
		".opencode/agents/odachi-senior.md",
		".claude/output-styles/odachi-novato.md",
		".claude/output-styles/odachi-senior.md",
	}
	got := persona.Deprecated()
	if len(got) != len(want) {
		t.Fatalf("Deprecated() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Deprecated()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	dirs := persona.DeprecatedDirs()
	if len(dirs) != 2 || dirs[0] != ".opencode/agents" || dirs[1] != ".claude/output-styles" {
		t.Fatalf("DeprecatedDirs() = %v, want the two client directories", dirs)
	}
}

// TestTheRendererDoesNotReadTheProfile: the voices do not depend on
// .odachi/profile.toml. If the two axes ever get wired together, this fails.
func TestTheRendererDoesNotReadTheProfile(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".").CombinedOutput()
	if err != nil {
		t.Skipf("go list no corrió en este entorno: %v: %s", err, out)
	}
	for _, path := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.Contains(path, "profile") {
			t.Errorf("el renderer de voces importa %q: el ceremonial y la voz son dos ejes y no se tocan", path)
		}
	}
}
