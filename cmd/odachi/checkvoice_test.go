package main

import (
	"reflect"
	"testing"
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
