package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/hostmodel"
	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	verifyapp "github.com/biggs-100/odachi/internal/app/verify"
)

// verifyProject builds a project with one feature to verify: a task document
// naming the requirement, and a commit after it that declared a risk. A verifier
// is asked for a change, so a tree that changed nothing proves nothing here.
func verifyProject(t *testing.T) string {
	t.Helper()

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

	return dir
}

// An echo is not a verification, so these runs end blocked; the identity reaches
// the command all the same, and the exit line carries what the command printed.
const echoModelAndFail = "echo {model} && exit 1"

// TestVerifyWithoutExecPrintsTheQuestionItWouldAsk: the verb reads the project's
// own words -the requirement the human wrote and the risk the commits declared- and
// prints them, running nobody. It is the cheap way to see what a verifier will be
// asked before paying for an answer.
func TestVerifyWithoutExecPrintsTheQuestionItWouldAsk(t *testing.T) {
	dir := verifyProject(t)

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

// TestVerifyRefusesAPlaceholderWithoutAModel: {model} with no model behind it is a
// verifier nobody can name, so the verb stops before it starts one and names the
// three sources that could have answered.
func TestVerifyRefusesAPlaceholderWithoutAModel(t *testing.T) {
	// The host is this process: cleared so the test does not depend on the agent
	// that runs the suite. It also means this test cannot be parallel.
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	code, _, errOut := runDir(t, dir, []string{"verify", "--exec", "echo {model}"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stderr: %q)", code, errOut)
	}
	for _, want := range []string{"{model}", "--model", "odd/review.toml", "ODACHI_MODEL"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr %q does not name %q", errOut, want)
		}
	}
}

// TestVerifyModelFlagReplacesThePlaceholder: --model is the human's override for
// one run, and the command that runs is the one carrying the identity.
func TestVerifyModelFlagReplacesThePlaceholder(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", echoModelAndFail, "--model", "human/model"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a record (stdout: %q, stderr: %q)", code, out, errOut)
	}
	joined := out + errOut
	if !strings.Contains(joined, "human/model") {
		t.Errorf("the resolved command is not shown (stdout: %q, stderr: %q)", out, errOut)
	}
	if strings.Contains(joined, "{model}") {
		t.Errorf("the placeholder reached the verifier (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestVerifyReadsTheModelTheProjectDeclares: odd/review.toml names the verifier's
// model, so a project chooses it once instead of on every command line.
func TestVerifyReadsTheModelTheProjectDeclares(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	writeFile(t, dir, "odd/review.toml", "mode = \"off\"\nverify = \"project/model\"\n")

	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", echoModelAndFail}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a record (stdout: %q, stderr: %q)", code, out, errOut)
	}
	joined := out + errOut
	if !strings.Contains(joined, "project/model") {
		t.Errorf("the declared model did not reach the command (stdout: %q, stderr: %q)", out, errOut)
	}
	if strings.Contains(joined, "{model}") {
		t.Errorf("the placeholder reached the verifier (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestTheFlagWinsOverTheProjectModel: the file is the project's default and the
// flag is the exception for one run.
func TestTheFlagWinsOverTheProjectModel(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	writeFile(t, dir, "odd/review.toml", "mode = \"off\"\nverify = \"project/model\"\n")

	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", echoModelAndFail, "--model", "flag/model"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a record (stdout: %q, stderr: %q)", code, out, errOut)
	}
	joined := out + errOut
	if !strings.Contains(joined, "flag/model") {
		t.Errorf("--model did not win (stdout: %q, stderr: %q)", out, errOut)
	}
	if strings.Contains(joined, "project/model") {
		t.Errorf("the project's model overrode the flag (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestTheSessionModelIsTheLastResort: with no flag and no key, the model odachi
// runs under is the one the verifier receives.
func TestTheSessionModelIsTheLastResort(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "session/model")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", echoModelAndFail}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a record (stdout: %q, stderr: %q)", code, out, errOut)
	}
	joined := out + errOut
	if !strings.Contains(joined, "session/model") {
		t.Errorf("the session's model did not reach the command (stdout: %q, stderr: %q)", out, errOut)
	}
	if strings.Contains(joined, "{model}") {
		t.Errorf("the placeholder reached the verifier (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestTheDeclaredModelIsIgnoredByACommandThatCannotUseIt: the key is a default for
// {model}, not a model every command must accept. A command without the
// placeholder runs exactly as it did before the key existed, and the verb does not
// blame the human for a value they never typed (the review's critical finding).
func TestTheDeclaredModelIsIgnoredByACommandThatCannotUseIt(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	writeFile(t, dir, "odd/review.toml", "mode = \"off\"\nverify = \"project/model\"\n")

	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", "echo sin-placeholder && exit 1"}, "")

	if strings.Contains(errOut, "no {model} placeholder") || strings.Contains(errOut, "--model only means") {
		t.Fatalf("the declared model refused a command that never asked for one (stderr: %q)", errOut)
	}
	if strings.Contains(out, "project/model") {
		t.Errorf("the declared model reached a command that cannot use it (stdout: %q)", out)
	}
	if code != 1 {
		t.Fatalf("code = %d, want 1: an echo is not a record (stdout: %q, stderr: %q)", code, out, errOut)
	}
}

// TestVerifyRefusesAModelWithoutThePlaceholder: a model that replaces nothing
// would be accepted and ignored, so the verb says which placeholder gives it a
// meaning and runs nobody.
func TestVerifyRefusesAModelWithoutThePlaceholder(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	code, out, errOut := runDir(t, dir, []string{"verify", "--exec", "echo sin-modelo && exit 1", "--model", "human/model"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(errOut, "{model}") {
		t.Errorf("stderr %q does not name the placeholder", errOut)
	}
	if strings.Contains(out+errOut, "sin-modelo") {
		t.Errorf("the command ran with a model nothing would use (stdout: %q, stderr: %q)", out, errOut)
	}
}

// TestVerifyRefusesAModelWithoutExec: --model only means something with --exec, so
// the flag alone is refused instead of printing a prompt it would not reach.
func TestVerifyRefusesAModelWithoutExec(t *testing.T) {
	t.Setenv(hostmodel.EnvModel, "")
	t.Setenv(hostmodel.SessionEnv, "")

	dir := verifyProject(t)
	code, out, errOut := runDir(t, dir, []string{"verify", "--model", "human/model"}, "")

	if code != 1 {
		t.Fatalf("code = %d, want 1 (stdout: %q, stderr: %q)", code, out, errOut)
	}
	if !strings.Contains(errOut, "--exec") {
		t.Errorf("stderr %q does not name --exec", errOut)
	}
	if strings.Contains(out, "the prompt the verifier will read") {
		t.Errorf("the verb printed a prompt with --model set (stdout: %q)", out)
	}
}

// TestVerifyHelpNamesTheModelFlag: the flag is discoverable from the verb's own
// section, next to the command it modifies, not from another verb's.
func TestVerifyHelpNamesTheModelFlag(t *testing.T) {
	code, out, errOut := runDir(t, t.TempDir(), []string{"verify", "-h"}, "")

	if code != 0 {
		t.Fatalf("exit = %d, want 0 (stderr: %s)", code, errOut)
	}
	section, found := between(out, "verify flags:", "plan flags:")
	if !found {
		t.Fatalf("the help has no verify section:\n%s", out)
	}
	for _, want := range []string{"--model id", "replaces {model}", "odd/review.toml"} {
		if !strings.Contains(section, want) {
			t.Errorf("the verify section does not name %q:\n%s", want, section)
		}
	}
}

// between is the part of a text that starts at one marker and ends at the next,
// so a check reads one help section instead of the whole page.
func between(text, start, end string) (string, bool) {
	from := strings.Index(text, start)
	if from < 0 {
		return "", false
	}
	rest := text[from:]
	if to := strings.Index(rest, end); to >= 0 {
		return rest[:to], true
	}
	return rest, true
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
