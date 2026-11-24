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

// officeVerbLine applies the init/start/level rules to one line. It runs after
// officeCheckLine, so a note shared with the gate keeps the gate's form. Only
// prose changes: commands, flags, paths and numbers are left as they came in
// (0359). A line with no known form passes through unchanged.
func officeVerbLine(line string) string {
	// Order matters: the whole phrases go before the fragments they contain,
	// and the generic one-word rules go last, so no replacement eats the text a
	// later rule needs to recognise.
	replacements := [][2]string{
		// level's "you earned the X profile:" shares the word "profile" with the
		// facts line; handle its colon form before the generic prefix below.
		{" profile:", ":"},

		// init and start both open with "done: wrote"; keep that whole opening
		// before the shorter "done: " rule.
		{"done: wrote ", "listo: escribió "},

		// init
		{"I created it in this folder (git init); it is what lets me prove that a proof belongs to this state of the code", "lo creé en esta carpeta (git init); es lo que me permite probar que una prueba pertenece a este estado del código"},
		{"this is what ", "esto es lo que "},
		{" would write, and it writes nothing", " escribiría, y no escribe nada"},
		{"nothing to write, this machine is already current", "nada que escribir, esta máquina ya está al día"},
		{" and the native agent were already up to date", " y el agente nativo ya estaban al día"},
		{"the hook ", "el gancho "},
		{"when odachi check --quiet is green, the agent saves the progress", "cuando odachi check --quiet esté en verde, el agente guarda el progreso"},

		// start
		{"; a light feature is one line, and it owes the same proof at its close", "; una funcionalidad ligera es una línea y debe la misma prueba al cerrarse"},
		{"; commit what start wrote (the plan) before you build", "; confirma lo que start escribió (el plan) antes de construir"},
		{"the fence protects files, so what names no paths stays in ", "la cerca protege archivos, así que lo que no nombra rutas queda en "},
		{"to prove a rule works: touch a path it protects and run ", "para probar que una regla funciona: toca una ruta que protege y ejecuta "},
		{" - it must refuse", " - debe negarse"},
		{"answers are missing and this run cannot ask: ", "faltan respuestas y esta ejecución no puede preguntar: "},
		{"; run odachi start again with those flags (or pass ", "; ejecuta odachi start de nuevo con esas opciones (o pasa "},
		{"this is how your answers look; if they are right, I write the files", "así se ven tus respuestas; si son correctas, escribo los archivos"},
		{"non-goals", "no objetivos"},
		{"non-goal", "no objetivo"},
		{" rules it already had", " reglas que ya tenía"},
		{"the fence kept ", "la cerca conservó "},

		// level
		{" · features with steps: ", " · funcionalidades con pasos: "},
		{" · answered questions: ", " · preguntas respondidas: "},
		{" · requirement changes: ", " · cambios de requisito: "},
		{" · craft decisions: ", " · decisiones de oficio: "},
		{" · open questions: ", " · preguntas abiertas: "},
		// This sentence carries the same "features with done steps and"
		// fragment as the earned line, so its whole form must be handled first.
		{"the senior profile is not earned yet: keep adding features with done steps and vision decisions (answered questions or requirement changes)", "el perfil senior aún no se ganó: sigue agregando funcionalidades con pasos terminados y decisiones de visión (preguntas respondidas o cambios de requisito)"},
		{"you earned the ", "ganaste el perfil "},
		{" features with done steps and ", " funcionalidades con pasos terminados y "},
		{" vision decisions (answered questions and requirement changes); to accept it run ", " decisiones de visión (preguntas respondidas y cambios de requisito); para aceptarlo ejecuta "},
		{"you are already at ", "ya estás en "},
		{": there is nothing to propose", ": no hay nada que proponer"},
		{"applied the ", "aplicaste el perfil "},
		{" profile in ", " en "},
		{" and the block voice is ", " y la voz del bloque es "},
		{"the profile was already ", "el perfil ya era "},
		{"the previous profile is at ", "el perfil anterior está en "},
		{"there is no odachi block: the voice is not in the project; run ", "no hay bloque odachi: la voz no está en el proyecto; ejecuta "},

		// shared opening tags
		{"unchanged: ", "sin cambios: "},
		{"project history: ", "historia del proyecto: "},
		{"dry run: ", "simulación: "},
		{"pending: ", "pendiente: "},
		{"failed: ", "falló: "},
		{"done: ", "listo: "},
		{"profile: ", "perfil: "},

		// generic fragments, last
		{" are up to date", " están al día"},
		{" and ", " y "},
	}
	for _, r := range replacements {
		line = strings.Replace(line, r[0], r[1], 1)
	}
	return line
}

// officeLine is the one binding for an init/start/level line under office-peru:
// the gate rules are applied first, then the verb rules.
func officeLine(line string) string {
	return officeVerbLine(officeCheckLine(line))
}

// officeLines applies the verb translation to a whole human block.
func officeLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, officeLine(line))
	}
	return out
}

// voiceLines is the printing gate. Only office-peru is translated; every other
// profile keeps the exact bytes the domain wrote, including an empty voice.
func voiceLines(profileName string, lines []string) []string {
	if profileName != "office-peru" {
		return lines
	}
	return officeLines(lines)
}

// voiceLine is voiceLines for the single blocked message a verb prints.
func voiceLine(profileName, line string) string {
	if profileName != "office-peru" {
		return line
	}
	return officeLine(line)
}
