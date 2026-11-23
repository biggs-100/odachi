package main

import "strings"

// officeCheckLine renders one gate line for the office-peru profile. It is a
// boundary translation: the domain keeps its typed reasons, the JSON keeps its
// keys, and only the human lines change. A line with no known form passes
// through unchanged, because a future note must never disappear (0355).
func officeCheckLine(line string) string {
	switch {
	case strings.HasPrefix(line, "ready: "):
		rest := strings.TrimPrefix(line, "ready: ")
		rest = strings.Replace(rest, "there is proof for this tree", "hay prueba para este árbol", 1)
		return "listo: " + rest
	case strings.HasPrefix(line, "ready for a step: "):
		rest := strings.TrimPrefix(line, "ready for a step: ")
		rest = strings.Replace(rest, "the proof is owed when this feature closes", "la prueba se debe al cerrar este trabajo", 1)
		rest = strings.Replace(rest, ", and the tree that closes it is the one CI judges", ", y el árbol que lo cierra es el que se revisa automáticamente", 1)
		return "listo para un paso: " + rest
	case strings.HasPrefix(line, "adds "):
		rest := strings.TrimPrefix(line, "adds ")
		rest = strings.Replace(rest, " new files: ", " archivos nuevos: ", 1)
		rest = strings.Replace(rest, " and ", " y ", 1)
		rest = strings.Replace(rest, " more; name the reason in the PR", " más; explica el motivo al entregar", 1)
		return "agrega " + rest
	case strings.HasPrefix(line, "no review for this tree"):
		return strings.Replace(line, "no review for this tree", "no hay revisión de este árbol", 1)
	case strings.HasPrefix(line, "next: "):
		return "siguiente: " + strings.TrimPrefix(line, "next: ")
	case strings.HasPrefix(line, "proof: "):
		return "prueba: " + strings.TrimPrefix(line, "proof: ")
	case strings.HasPrefix(line, "undo: "):
		rest := strings.TrimPrefix(line, "undo: ")
		rest = strings.Replace(rest, "could not be judged", "no se pudo juzgar", 1)
		rest = strings.Replace(rest, "the copy did not build", "la copia sin el cambio no compila", 1)
		return "reversión: " + rest
	case strings.HasPrefix(line, "warning: "):
		rest := strings.TrimPrefix(line, "warning: ")
		rest = strings.Replace(rest, "the proof produced no output", "la prueba no mostró salida", 1)
		rest = strings.Replace(rest, "a proof that says nothing is weak evidence", "una prueba que no dice nada es evidencia débil", 1)
		rest = strings.Replace(rest, "run a command that prints what it proved", "ejecuta una comprobación que imprima lo que verificaste", 1)
		return "aviso: " + rest
	case strings.HasPrefix(line, "store: "):
		rest := strings.TrimPrefix(line, "store: ")
		rest = strings.Replace(rest, " would free ", " liberaría ", 1)
		rest = strings.ReplaceAll(rest, " in ", " en ")
		rest = strings.ReplaceAll(rest, " files; ", " archivos; ")
		rest = strings.ReplaceAll(rest, " files", " archivos")
		return "almacén: " + rest
	default:
		return line
	}
}

// officeCheckLines applies the translation to every line the gate prints.
func officeCheckLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, officeCheckLine(line))
	}
	return out
}
