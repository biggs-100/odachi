package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// A linked worktree has .git as a file pointing at the main repository. Telling
// that project "no hay .git, corré git init" is a lie and a dead end: git init is
// already done, in the main repository.
func TestDoctorUnderstandsAWorktree(t *testing.T) {
	dir := t.TempDir()
	pointer := "gitdir: /otro/repo/.git/worktrees/rama\n"
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(pointer), 0o644); err != nil {
		t.Fatalf("write .git: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "worktree") {
		t.Fatalf("el reporte no reconoce el worktree: %q", out)
	}
	if strings.Contains(out, "corré git init") {
		t.Fatalf("el reporte manda a git init en un worktree: %q", out)
	}
}

// A profile whose name nobody shipped and that inherits from nothing resolves to
// a profile with every field empty: it validates and configures nothing. Doctor
// has to call that a problem, because every verb would run with a blank ceremony.
func TestDoctorRefusesAProfileThatInheritsFromNobody(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".odachi", "profile.toml"), []byte("name = \"inventado\"\n"), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if code != 1 {
		t.Fatalf("code = %d, want 1: un perfil que no hereda de nadie es un problema", code)
	}
	if !strings.Contains(out, "no hereda de nadie") {
		t.Fatalf("el reporte no dice por qué: %q", out)
	}
	if !strings.Contains(out, "extends") {
		t.Fatalf("el reporte no dice cómo arreglarlo: %q", out)
	}
}

// A profile that inherits from a shipped one is a legitimate configuration, and
// must stay green so the rule only catches the empty ceremony.
func TestDoctorAcceptsAProfileThatInherits(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".odachi"), 0o755); err != nil {
		t.Fatalf("mkdir .odachi: %v", err)
	}
	content := "name = \"mio\"\nextends = \"senior\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".odachi", "profile.toml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write profile.toml: %v", err)
	}

	_, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	out := stdout + stderr

	if !strings.Contains(out, "perfil mio") {
		t.Fatalf("el reporte no resuelve el perfil heredado: %q", out)
	}
	if strings.Contains(out, "no hay archivo") {
		t.Fatalf("el reporte dice que no hay archivo cuando sí lo hay: %q", out)
	}
}

// checkLine returns the one line of the report that carries a check number. The
// number is how a report and a test name the same check, so the helper matches on
// it and never on the position: a check inserted in the middle cannot make a test
// read its neighbour's line.
func checkLine(t *testing.T, report string, number int) string {
	t.Helper()

	prefix := " " + strconv.Itoa(number) + " · "
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, prefix) {
			return line
		}
	}
	t.Fatalf("el reporte no trae el chequeo %d:\n%s", number, report)
	return ""
}

// Un proyecto que acaba de correr `odachi init` tiene las skills en
// .agents/skills y ningún registro propio, y el chequeo 9 decía solo "ninguna: no
// hay skills/": el humano leía que no había nada cuando init sí había instalado
// las suyas. Son dos preguntas distintas —el registro del proyecto y lo instalado
// donde el cliente lo lee— y la línea las responde sin mezclarlas.
func TestDoctorCheckNineSaysBothSkillFacts(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"doctor"})
	line := checkLine(t, out, 9)

	curriculum, err := bundledskills.Curriculum()
	if err != nil {
		t.Fatalf("las skills incluidas: %v", err)
	}
	for _, want := range []string{
		"skills/: no existe",
		fmt.Sprintf(".agents/skills: %d instaladas", len(curriculum)),
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 9 no dice %q: %s", want, line)
		}
	}
}

// La voz es la última sección del bloque y la escribe el binario. Si el bloque
// dice una cosa y el perfil que el binario lee dice otra, el agente habla en un
// registro que la herramienta no prometió, y el chequeo lo tiene que decir antes
// de que el humano lo note leyendo al agente: nombra los dos y el comando que los
// pone de acuerdo.
func TestDoctorReportsAVoiceThatDoesNotMatchTheProfile(t *testing.T) {
	dir := configuredProject(t)

	// El proyecto quedó inicializado con la voz de novato y el perfil que `start`
	// escribió. Volver el bloque a novato es el estado de un proyecto al que le
	// cambiaron el perfil sin refrescar la voz.
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	writeFile(t, dir, "AGENTS.md", block)
	profile := readProfileName(t, dir)

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("una voz que no coincide con el perfil: exit %d, want 1: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout+stderr, 11)
	for _, want := range []string{"mal 11", "novato", profile, "odachi level --to " + profile} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 11 no dice %q: %s", want, line)
		}
	}

	// El comando que la línea nombra arregla el problema, y el chequeo vuelve a ok.
	if code, _, stderr := runDir(t, dir, []string{"level", "--to", profile}, ""); code != 0 {
		t.Fatalf("level --to %s: exit %d: %s", profile, code, stderr)
	}
	out, _ := output(t, dir, []string{"doctor"})
	if line := checkLine(t, out, 11); !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("después de level --to el chequeo 11 sigue mal: %s", line)
	}
}

// Un bloque sin sección de voz y un perfil que existe: la voz no está en el
// proyecto, y el comando que la escribe es init.
func TestDoctorReportsABlockWithoutTheVoiceSection(t *testing.T) {
	dir := configuredProject(t)
	writeFile(t, dir, "AGENTS.md", "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas\n<!-- odachi:end -->\n")

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("un bloque sin voz y con perfil: exit %d, want 1: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout+stderr, 11)
	for _, want := range []string{"mal 11", rules.Section, "odachi init"} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 11 no dice %q: %s", want, line)
		}
	}
}

// Las personas por perfil de v0.10.0 ya no son el agente de esta versión. Son archivos del humano, no
// un defecto: el chequeo lo dice en la misma línea verde, nombra las carpetas, y
// no falla por eso.
func TestDoctorSaysTheV0100PersonaFilesAreLeftovers(t *testing.T) {
	dir := configuredProject(t)
	writeFile(t, dir, ".opencode/agents/odachi-senior.md", "viejo")

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 0 {
		t.Fatalf("los leftovers no son un problema: exit %d: %s %s", code, stdout, stderr)
	}
	line := checkLine(t, stdout, 11)
	if !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("el chequeo 11 tiene que ser ok con leftovers: %s", line)
	}
	for _, want := range []string{"v0.10.0", ".opencode/agents", ".claude/output-styles", "podés borrarlas"} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 11 no dice %q: %s", want, line)
		}
	}
}

// Un proyecto sin .odachi/profile.toml no tiene contra qué comparar la voz: el
// bloque lleva la que init escribió y no hay nada roto. Fallar acá le diría a
// alguien que acaba de correr init que su proyecto está mal.
func TestDoctorSaysNoProfileIsNothingToCompare(t *testing.T) {
	dir := configuredProject(t)
	if err := os.Remove(filepath.Join(dir, ".odachi", "profile.toml")); err != nil {
		t.Fatalf("borrar .odachi/profile.toml: %v", err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 0 {
		t.Fatalf("sin perfil no hay voz que comparar: exit %d: %s%s", code, stdout, stderr)
	}
	line := checkLine(t, stdout, 11)
	if !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("el chequeo 11 tiene que ser ok sin perfil: %s", line)
	}
	if !strings.Contains(line, "no hay perfil que comparar") {
		t.Fatalf("el chequeo 11 no dice que no hay con qué comparar: %s", line)
	}
}

// readProfileName is the profile the binary reads, for the lines that name it.
func readProfileName(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".odachi", "profile.toml"))
	if err != nil {
		t.Fatalf("leer .odachi/profile.toml: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "name" {
			return strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	t.Fatalf("el perfil no dice su nombre: %s", data)
	return ""
}

// La línea de cierre dice cuántos chequeos corrieron: con uno nuevo son once, y
// el uso del verbo no puede seguir diciendo diez.
func TestDoctorUsageAndSummaryCountElevenChecks(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"doctor"})
	if !strings.Contains(out, "todo en orden: 11 chequeos") {
		t.Fatalf("el resumen cuenta once chequeos:\n%s", out)
	}

	usageOut, _ := output(t, dir, []string{"doctor", "-h"})
	if !strings.Contains(usageOut, "once chequeos de salud") {
		t.Fatalf("el uso no puede seguir diciendo diez:\n%s", usageOut)
	}
}
