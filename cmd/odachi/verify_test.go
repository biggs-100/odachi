package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyWithoutExecPrintsTheQuestionItWouldAsk: the verb reads the project's
// own words -the requirement the human wrote and the risk the commits declared- and
// prints them, running nobody. It is the cheap way to see what a verifier will be
// asked before paying for an answer.
func TestVerifyWithoutExecPrintsTheQuestionItWouldAsk(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	writeFile(t, dir, "README.md", "# demo\n")
	commitAll(t, dir, "base")
	base := headOf(t, dir)

	writeFile(t, dir, "odd/tasks/0001-la-cerca.md", `# la-cerca

Base: `+base+`

## Requirements (verbatim, from the user)
R1 · "quién escribe no se verifica a sí mismo"

## Done when
- la puerta bloquea cuando no hay registro
`)
	writeFile(t, dir, "internal/app/turnos.go", "package app\n")
	commitAll(t, dir, "a risky change\n\nRisk: contract")

	code, out, errOut := runDir(t, dir, []string{"verify"}, "")

	if code != 0 {
		t.Fatalf("exit = %d, want 0 (stderr: %s)", code, errOut)
	}
	for _, want := range []string{
		"quien escribe no se verifica a si mismo",
		"la puerta bloquea cuando no hay registro",
		"contract",
		"verified", "refuted",
		`odachi verify --exec`,
	} {
		if !strings.Contains(normalize(out), normalize(want)) {
			t.Errorf("the prompt does not name %q:\n%s", want, out)
		}
	}
	if _, err := filepath.Glob(filepath.Join(dir, ".git", "odachi", "verifications", "*.json")); err != nil {
		t.Fatalf("glob: %v", err)
	}
	records, _ := filepath.Glob(filepath.Join(dir, ".git", "odachi", "verifications", "*.json"))
	if len(records) != 0 {
		t.Errorf("a run with no command stored %v", records)
	}
}

// normalize compares what a terminal shows: the prompt quotes the human's Spanish,
// and the console that prints it may not carry the accents the file holds.
func normalize(text string) string {
	replacer := strings.NewReplacer(
		"í", "i", "ó", "o", "á", "a", "é", "e", "ú", "u", "ñ", "n",
		"Í", "I", "Ó", "O", "Á", "A", "É", "E", "Ú", "U", "Ñ", "N",
	)
	return replacer.Replace(text)
}
