package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

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

// Las personas generadas se leen desde el selector del cliente. Si una no está
// donde ese cliente la busca, el selector aparece vacío y el humano se entera ahí:
// el chequeo tiene que decirlo antes, con el comando que las vuelve a escribir.
func TestDoctorReportsAMissingPersonaFile(t *testing.T) {
	dir := configuredProject(t)

	missing := filepath.Join(dir, ".opencode", "agents", "odachi-senior.md")
	if err := os.Remove(missing); err != nil {
		t.Fatalf("borrar %s: %v", missing, err)
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("una persona que falta deja el selector vacío: exit %d, want 1:\n%s%s", code, stdout, stderr)
	}
	line := checkLine(t, stdout+stderr, 11)
	for _, want := range []string{
		"mal 11",
		".opencode/agents/odachi-senior.md",
		"el selector de opencode queda vacío",
		"corré odachi init",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 11 no dice %q: %s", want, line)
		}
	}
}

// Un proyecto que no usa ningún cliente con selector no está roto: no hay
// personas que leer y no hay nada que arreglar. Se dice con la misma forma con la
// que doctor dice "no hay nada que revisar todavía" en el resto de los chequeos,
// y no se falla por eso.
func TestDoctorSaysThereAreNoPersonasYet(t *testing.T) {
	dir := configuredProject(t)

	for _, rel := range []string{".opencode", ".claude"} {
		if err := os.RemoveAll(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("borrar %s: %v", rel, err)
		}
	}

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 0 {
		t.Fatalf("sin clientes no hay nada que arreglar: exit %d:\n%s%s", code, stdout, stderr)
	}
	line := checkLine(t, stdout, 11)
	if !strings.HasPrefix(line, "ok 11") {
		t.Fatalf("sin personas y sin clientes el chequeo no es un problema: %s", line)
	}
	if !strings.Contains(line, "todavía no hay personas") {
		t.Fatalf("el chequeo 11 tiene que decir que todavía no hay personas: %s", line)
	}
}

// Un proyecto configurado dice dónde están las personas: la mitad útil de la
// línea verde es el directorio que el cliente lee.
func TestDoctorNamesTheDirectoriesWhereThePersonasAre(t *testing.T) {
	dir := configuredProject(t)

	out, _ := output(t, dir, []string{"doctor"})
	line := checkLine(t, out, 11)

	for _, want := range []string{".opencode/agents", ".claude/output-styles"} {
		if !strings.Contains(line, want) {
			t.Fatalf("el chequeo 11 no nombra %q: %s", want, line)
		}
	}
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
