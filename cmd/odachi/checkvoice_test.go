package main

import (
	"bytes"
	"reflect"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

// TestOfficeCheckLinesTranslateTheObservedGate: these are the lines the walkthrough
// actually printed in an office-peru project. They must read as plain Spanish,
// without changing commands, paths or exit codes (0355).
func TestOfficeCheckLinesTranslateTheObservedGate(t *testing.T) {
	cases := []struct{ in, want string }{
		{
			"ready: there is proof for this tree",
			"listo: hay prueba para este árbol",
		},
		{
			"ready for a step: the proof is owed when this feature closes, and the tree that closes it is the one CI judges",
			"listo para un paso: la prueba se debe al cerrar este trabajo, y el árbol que lo cierra es el que se revisa automáticamente",
		},
		{
			"adds 95 new files: .agents/skills/acceptance/SKILL.md and 87 more; name the reason in the PR",
			"agrega 95 archivos nuevos: .agents/skills/acceptance/SKILL.md y 87 más; explica el motivo al entregar",
		},
		{
			"no review for this tree; odachi review",
			"no hay revisión de este árbol; odachi review",
		},
		{
			"next: odachi review",
			"siguiente: odachi review",
		},
		{
			"proof: proof/0001-un-informe-corto.txt",
			"prueba: proof/0001-un-informe-corto.txt",
		},
		{
			"undo: could not be judged: the copy did not build",
			"reversión: no se pudo juzgar: la copia sin el cambio no compila",
		},
		{
			"warning: the proof produced no output: a proof that says nothing is weak evidence; run a command that prints what it proved",
			"aviso: la prueba no mostró salida: una prueba que no dice nada es evidencia débil; ejecuta una comprobación que imprima lo que verificaste",
		},
		{
			"store: 4.1 MB in 1499 files; odachi clean --apply would free 1.9 MB in 150 files",
			"almacén: 4.1 MB en 1499 archivos; odachi clean --apply liberaría 1.9 MB en 150 archivos",
		},
	}
	for _, c := range cases {
		if got := officeCheckLine(c.in); got != c.want {
			t.Errorf("officeCheckLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestOfficeCheckLinesLeaveUnknownLinesAlone: a future gate note that has no
// translation must still be printed, not swallowed by an over-eager mapper.
func TestOfficeCheckLinesLeaveUnknownLinesAlone(t *testing.T) {
	in := []string{"a note this feature never saw", "next: git status --short"}
	want := []string{"a note this feature never saw", "siguiente: git status --short"}
	if got := officeCheckLines(in); !reflect.DeepEqual(got, want) {
		t.Fatalf("officeCheckLines = %q, want %q", got, want)
	}
}

// TestOfficeLineTranslatesInit: these are the lines init really prints (the
// result message, the unchanged rerun, the two dry-run notices, the git line
// main.go prepends, and the next: note). Commands and paths stay as they came.
func TestOfficeLineTranslatesInit(t *testing.T) {
	cases := []struct{ in, want string }{
		{
			"done: AGENTS.md and .gitignore are up to date",
			"listo: AGENTS.md y .gitignore están al día",
		},
		{
			"unchanged: AGENTS.md, .gitignore, the hook .git/hooks/pre-commit and the native agent were already up to date",
			"sin cambios: AGENTS.md, .gitignore, el gancho .git/hooks/pre-commit y el agente nativo ya estaban al día",
		},
		{
			"dry run: this is what odachi setup would write, and it writes nothing",
			"simulación: esto es lo que odachi setup escribiría, y no escribe nada",
		},
		{
			"dry run: nothing to write, this machine is already current",
			"simulación: nada que escribir, esta máquina ya está al día",
		},
		{
			"project history: I created it in this folder (git init); it is what lets me prove that a proof belongs to this state of the code",
			"historia del proyecto: lo creé en esta carpeta (git init); es lo que me permite probar que una prueba pertenece a este estado del código",
		},
		{
			"next: check AGENTS.md; when odachi check --quiet is green, the agent saves the progress",
			"siguiente: check AGENTS.md; cuando odachi check --quiet esté en verde, el agente guarda el progreso",
		},
	}
	for _, c := range cases {
		if got := officeLine(c.in); got != c.want {
			t.Errorf("officeLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestOfficeLineTranslatesStart: real lines from start's result, the fence note,
// the refusal that names missing flags and the answers preview.
func TestOfficeLineTranslatesStart(t *testing.T) {
	cases := []struct{ in, want string }{
		{
			"done: wrote odd/tasks/0359-los-verbos-init.md; a light feature is one line, and it owes the same proof at its close",
			"listo: escribió odd/tasks/0359-los-verbos-init.md; una funcionalidad ligera es una línea y debe la misma prueba al cerrarse",
		},
		{
			"done: wrote odd/tasks/foo.md; commit what start wrote (the plan) before you build",
			"listo: escribió odd/tasks/foo.md; confirma lo que start escribió (el plan) antes de construir",
		},
		{
			"the fence kept 6 rules it already had",
			"la cerca conservó 6 reglas que ya tenía",
		},
		{
			"the fence protects files, so what names no paths stays in brief.md: 3 non-goals",
			"la cerca protege archivos, así que lo que no nombra rutas queda en brief.md: 3 no objetivos",
		},
		{
			"to prove a rule works: touch a path it protects and run odachi check - it must refuse",
			"para probar que una regla funciona: toca una ruta que protege y ejecuta odachi check - debe negarse",
		},
		{
			"answers are missing and this run cannot ask: --outcome, --audience; run odachi start again with those flags (or pass --brief <file>)",
			"faltan respuestas y esta ejecución no puede preguntar: --outcome, --audience; ejecuta odachi start de nuevo con esas opciones (o pasa --brief <file>)",
		},
		{
			"this is how your answers look; if they are right, I write the files",
			"así se ven tus respuestas; si son correctas, escribo los archivos",
		},
	}
	for _, c := range cases {
		if got := officeLine(c.in); got != c.want {
			t.Errorf("officeLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestOfficeLineTranslatesLevel: real lines from level's facts, proposal, apply
// and the AGENTS.md note. The profile names, counts, paths and commands stay.
func TestOfficeLineTranslatesLevel(t *testing.T) {
	cases := []struct{ in, want string }{
		{
			"profile: novato · features with steps: 12 · answered questions: 30 · requirement changes: 4 · craft decisions: 8 · open questions: 2",
			"perfil: novato · funcionalidades con pasos: 12 · preguntas respondidas: 30 · cambios de requisito: 4 · decisiones de oficio: 8 · preguntas abiertas: 2",
		},
		{
			"you earned the senior profile: 12 features with done steps and 34 vision decisions (answered questions and requirement changes); to accept it run odachi level --to senior",
			"ganaste el perfil senior: 12 funcionalidades con pasos terminados y 34 decisiones de visión (preguntas respondidas y cambios de requisito); para aceptarlo ejecuta odachi level --to senior",
		},
		{
			"you are already at senior: there is nothing to propose",
			"ya estás en senior: no hay nada que proponer",
		},
		{
			"the senior profile is not earned yet: keep adding features with done steps and vision decisions (answered questions or requirement changes)",
			"el perfil senior aún no se ganó: sigue agregando funcionalidades con pasos terminados y decisiones de visión (preguntas respondidas o cambios de requisito)",
		},
		{
			"done: applied the senior profile in odd/profile.toml and the block voice is senior",
			"listo: aplicaste el perfil senior en odd/profile.toml y la voz del bloque es senior",
		},
		{
			"the profile was already senior",
			"el perfil ya era senior",
		},
		{
			"the previous profile is at odd/profile.toml.bak-1",
			"el perfil anterior está en odd/profile.toml.bak-1",
		},
		{
			"AGENTS.md: there is no odachi block: the voice is not in the project; run odachi init",
			"AGENTS.md: no hay bloque odachi: la voz no está en el proyecto; ejecuta odachi init",
		},
	}
	for _, c := range cases {
		if got := officeLine(c.in); got != c.want {
			t.Errorf("officeLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestOfficeLineAppliesCheckRulesFirst: the verb entry point reuses the gate's
// mapper, so a shared note such as next: keeps one Spanish form for every verb.
func TestOfficeLineAppliesCheckRulesFirst(t *testing.T) {
	if got, want := officeLine("next: odachi review"), "siguiente: odachi review"; got != want {
		t.Fatalf("officeLine = %q, want %q", got, want)
	}
}

// TestOfficeLinePassesUnknownThrough: a future init/start/level note with no
// known form must still be printed, exactly as the gate's mapper does (0359).
func TestOfficeLinePassesUnknownThrough(t *testing.T) {
	in := []string{"a verb note this feature never saw", "a future sentence about a thing not yet named"}
	want := []string{"a verb note this feature never saw", "a future sentence about a thing not yet named"}
	if got := officeLines(in); !reflect.DeepEqual(got, want) {
		t.Fatalf("officeLines = %q, want %q", got, want)
	}
}

// TestOfficeVoiceLinesLeaveOtherProfilesByteIdentical: only office-peru gets
// Spanish. novato, senior, ci (and no profile at all) keep the exact bytes the
// domain wrote (0359).
func TestOfficeVoiceLinesLeaveOtherProfilesByteIdentical(t *testing.T) {
	in := []string{
		"done: wrote odd/tasks/foo.md; commit what start wrote (the plan) before you build",
		"profile: novato · features with steps: 12 · answered questions: 30",
	}
	for _, profileName := range []string{"novato", "senior", "ci", ""} {
		got := voiceLines(profileName, in)
		if !reflect.DeepEqual(got, in) {
			t.Errorf("voiceLines(%q) = %q, want the untouched %q", profileName, got, in)
		}
	}
}

// TestOfficeVoiceLinesTranslateOfficePeru: the profile gate opens the same
// mapper the three report functions use.
func TestOfficeVoiceLinesTranslateOfficePeru(t *testing.T) {
	in := []string{"done: wrote odd/tasks/foo.md; commit what start wrote (the plan) before you build"}
	want := officeLines(in)
	if got := voiceLines("office-peru", in); !reflect.DeepEqual(got, want) {
		t.Fatalf("voiceLines(office-peru) = %q, want %q", got, want)
	}
}

// TestReportInitJSONIsIdenticalForEveryProfile: the voice is a printing choice.
// A machine reading --json gets the untouched English message and keys.
func TestReportInitJSONIsIdenticalForEveryProfile(t *testing.T) {
	result := initapp.Result{
		Code:    0,
		Message: "done: AGENTS.md and .gitignore are up to date",
		Next:    "odachi init",
		Lines:   []string{"done: AGENTS.md and .gitignore are up to date"},
		Files:   []string{"AGENTS.md", ".gitignore"},
	}
	var officeOut, officeErr, plainOut, plainErr bytes.Buffer
	reportInit(result, "office-peru", true, &officeOut, &officeErr)
	reportInit(result, "novato", true, &plainOut, &plainErr)
	if officeOut.String() != plainOut.String() {
		t.Fatalf("--json changed with the profile:\noffice-peru: %s\nnovato: %s", officeOut.String(), plainOut.String())
	}
	if !bytes.Contains(officeOut.Bytes(), []byte("done: AGENTS.md and .gitignore are up to date")) {
		t.Fatalf("--json no longer carries the English message: %s", officeOut.String())
	}
	if officeErr.Len() != 0 || plainErr.Len() != 0 {
		t.Fatalf("--json wrote to stderr: %q / %q", officeErr.String(), plainErr.String())
	}
}

// TestReportOutcomeJSONIsIdenticalForEveryProfile: start and level add a voice
// to the outcome, but the JSON payload has the same keys and bytes as before.
func TestReportOutcomeJSONIsIdenticalForEveryProfile(t *testing.T) {
	o := outcome{
		Code:    0,
		Message: "done: wrote odd/tasks/foo.md; commit what start wrote (the plan) before you build",
		Lines:   []string{"done: wrote odd/tasks/foo.md; commit what start wrote (the plan) before you build"},
		Profile: "office-peru",
		Files:   []string{"odd/tasks/foo.md"},
		Voice:   "office-peru",
	}
	var officeOut, officeErr, plainOut, plainErr bytes.Buffer
	reportOutcome(o, true, &officeOut, &officeErr)
	o.Voice = "novato"
	reportOutcome(o, true, &plainOut, &plainErr)
	if officeOut.String() != plainOut.String() {
		t.Fatalf("--json changed with the voice:\noffice-peru: %s\nnovato: %s", officeOut.String(), plainOut.String())
	}
	if !bytes.Contains(officeOut.Bytes(), []byte("commit what start wrote")) {
		t.Fatalf("--json no longer carries the English message: %s", officeOut.String())
	}
}
