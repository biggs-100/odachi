package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	verifyapp "github.com/biggs-100/odachi/internal/app/verify"
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

// TestTheAdapterCauseReachesTheVerbByName: the adapter names the clock and the
// verb keeps its own word for it, so the two vocabularies meet in one mapping.
// A cast by order would report silence as the deadline the day either constant
// block grows a case, which is the defect this change removes.
func TestTheAdapterCauseReachesTheVerbByName(t *testing.T) {
	for _, tc := range []struct {
		from shellcmd.Stop
		want verifyapp.Stop
	}{
		{shellcmd.StopNone, verifyapp.StopNone},
		{shellcmd.StopDeadline, verifyapp.StopDeadline},
		{shellcmd.StopSilence, verifyapp.StopSilence},
	} {
		if got := runInputStop(tc.from); got != tc.want {
			t.Errorf("runInputStop(%v) = %d, want %d", tc.from, got, tc.want)
		}
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
