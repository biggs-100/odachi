package persona_test

import (
	"os/exec"
	"path/filepath"
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

// TestFilesMirrorTheBlockForEveryClient: one generated file per client, in the
// client's own selector directory, and the block is the last thing in it. The
// mirror is the whole point: the agent a human picks reads the same words as the
// block, so the voice and the rules cannot drift.
func TestFilesMirrorTheBlockForEveryClient(t *testing.T) {
	block := "# Reglas del proyecto (odachi)\n\nla voz del perfil\n"
	files := persona.Files(block)

	if len(files) != len(persona.Clients()) {
		t.Fatalf("Files() = %d archivos, want %d (uno por cliente)", len(files), len(persona.Clients()))
	}
	wantPaths := []string{
		".opencode/agents/odachi.md",
		".claude/output-styles/odachi.md",
	}
	for i, want := range wantPaths {
		if files[i].Path != want {
			t.Errorf("Files()[%d].Path = %q, want %q", i, files[i].Path, want)
		}
		if !strings.HasSuffix(files[i].Content, block) {
			t.Errorf("%s no termina con el bloque:\n%s", files[i].Path, files[i].Content)
		}
		if !strings.Contains(files[i].Content, "generado por odachi init") {
			t.Errorf("%s no dice que es generado:\n%s", files[i].Path, files[i].Content)
		}
	}
}

// TestOpenCodeFileIsAPrimaryAgent: Tab only cycles primary agents. A file the
// model alone can reach is not the selector the human asked for.
func TestOpenCodeFileIsAPrimaryAgent(t *testing.T) {
	for _, file := range persona.Files("bloque\n") {
		if file.Path != ".opencode/agents/odachi.md" {
			continue
		}
		for _, want := range []string{"mode: primary", "description:", "odachi:"} {
			if !strings.Contains(file.Content, want) {
				t.Errorf("%s no lleva %q:\n%s", file.Path, want, file.Content)
			}
		}
		return
	}
	t.Fatalf("Files() no escribió el agente de OpenCode")
}

// TestClaudeFileKeepsTheCodingInstructions: a style without that flag replaces
// Claude Code's own software engineering instructions, which is not a voice.
func TestClaudeFileKeepsTheCodingInstructions(t *testing.T) {
	for _, file := range persona.Files("bloque\n") {
		if file.Path != ".claude/output-styles/odachi.md" {
			continue
		}
		for _, want := range []string{"name: odachi", "keep-coding-instructions: true"} {
			if !strings.Contains(file.Content, want) {
				t.Errorf("%s no lleva %q:\n%s", file.Path, want, file.Content)
			}
		}
		return
	}
	t.Fatalf("Files() no escribió el agente de Claude")
}

// TestGlobalFilesAreTheDoorInTheHome: the file the human finds in any folder is
// a door, not a project mirror. It lives in the client's global directory, reads
// the project's block when there is one, and sends a new folder to bootstrap.
func TestGlobalFilesAreTheDoorInTheHome(t *testing.T) {
	files := persona.GlobalFiles()

	wantPaths := []string{
		".config/opencode/agents/odachi.md",
		".claude/output-styles/odachi.md",
	}
	if len(files) != len(wantPaths) {
		t.Fatalf("GlobalFiles() = %d archivos, want %d", len(files), len(wantPaths))
	}
	for i, want := range wantPaths {
		if files[i].Path != want {
			t.Errorf("GlobalFiles()[%d].Path = %q, want %q", i, files[i].Path, want)
		}
		for _, marker := range []string{
			"puerta de entrada",
			"AGENTS.md",
			"<!-- odachi:begin -->",
			"drive/SKILL.md",
			"bootstrap/SKILL.md",
			"odachi init",
			"odachi start",
		} {
			if !strings.Contains(files[i].Content, marker) {
				t.Errorf("%s no dice %q:\n%s", want, marker, files[i].Content)
			}
		}
		if strings.Contains(files[i].Content, "# Reglas del proyecto (odachi)") {
			t.Errorf("%s copia las reglas del proyecto: la puerta no es un segundo dueño", want)
		}
	}
}

// TestProjectAndGlobalShareTheName: the project copy wins inside an initialized
// project, and the global copy is the same name everywhere else. Two names would
// make the human learn which one to pick.
func TestProjectAndGlobalShareTheName(t *testing.T) {
	project := persona.Files("bloque\n")
	global := persona.GlobalFiles()
	if len(project) != len(global) {
		t.Fatalf("project %d files, global %d", len(project), len(global))
	}
	for i := range project {
		if filepath.Base(project[i].Path) != filepath.Base(global[i].Path) {
			t.Errorf("%s y %s no comparten nombre", project[i].Path, global[i].Path)
		}
	}
}

// TestSubagentFilesAreSubagentsForEveryClient: the worker and the verifier are
// native subagents, not primary agents. OpenCode gets `mode: subagent` and a
// permission block; Claude Code gets the name and, for the verifier, a tools
// allowlist without Edit or Write.
func TestSubagentFilesAreSubagentsForEveryClient(t *testing.T) {
	files := persona.SubagentFiles()

	wantPaths := []string{
		".opencode/agents/odachi-worker.md",
		".opencode/agents/odachi-verify.md",
		".claude/agents/odachi-worker.md",
		".claude/agents/odachi-verify.md",
	}
	if len(files) != len(wantPaths) {
		t.Fatalf("SubagentFiles() = %d archivos, want %d", len(files), len(wantPaths))
	}
	for i, want := range wantPaths {
		if files[i].Path != want {
			t.Errorf("SubagentFiles()[%d].Path = %q, want %q", i, files[i].Path, want)
		}
		if !strings.Contains(files[i].Content, "generado por odachi init") {
			t.Errorf("%s no dice que es generado", want)
		}
	}
	for _, file := range files {
		switch file.Path {
		case ".opencode/agents/odachi-worker.md", ".opencode/agents/odachi-verify.md":
			if !strings.Contains(file.Content, "mode: subagent") {
				t.Errorf("%s no es un subagente de OpenCode:\n%s", file.Path, file.Content)
			}
		case ".claude/agents/odachi-worker.md", ".claude/agents/odachi-verify.md":
			if !strings.Contains(file.Content, "name: odachi-") {
				t.Errorf("%s no declara el nombre de Claude:\n%s", file.Path, file.Content)
			}
		}
		if strings.Contains(file.Content, "mode: primary") {
			t.Errorf("%s declara mode: primary: los subagentes no se eligen con Tab", file.Path)
		}
	}
}

// TestTheVerifierIsReadOnly: the verifier must not be able to edit. OpenCode
// enforces it with `edit: deny`; Claude Code by leaving Edit and Write out of
// the tools allowlist. The worker keeps edit access and cannot delegate.
func TestTheVerifierIsReadOnly(t *testing.T) {
	for _, file := range persona.SubagentFiles() {
		switch file.Path {
		case ".opencode/agents/odachi-verify.md":
			for _, want := range []string{"edit: deny", "task: deny"} {
				if !strings.Contains(file.Content, want) {
					t.Errorf("%s no lleva %q:\n%s", file.Path, want, file.Content)
				}
			}
		case ".opencode/agents/odachi-worker.md":
			if !strings.Contains(file.Content, "task: deny") {
				t.Errorf("%s no prohíbe delegar:\n%s", file.Path, file.Content)
			}
		case ".claude/agents/odachi-verify.md":
			if !strings.Contains(file.Content, "tools: Read, Grep, Glob, Bash") {
				t.Errorf("%s no declara tools de solo lectura:\n%s", file.Path, file.Content)
			}
			if strings.Contains(file.Content, "Edit") || strings.Contains(file.Content, "Write") {
				t.Errorf("%s permite editar:\n%s", file.Path, file.Content)
			}
		}
	}
}

// TestSubagentFilesForUnknownClient: nothing is guessed for a client this table
// does not know.
func TestSubagentFilesForUnknownClient(t *testing.T) {
	if got := persona.SubagentFilesFor("inventado"); got != nil {
		t.Errorf("SubagentFilesFor(inventado) = %v, want nil", got)
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
