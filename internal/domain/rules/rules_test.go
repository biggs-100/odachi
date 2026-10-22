package rules_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// linesOf splits a text into the lines a model reads.
func linesOf(text string) []string {
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// The block is the only text odachi writes that an agent obeys, so it is pinned
// byte for byte: any edit to the voice or to a rule has to pass through this
// test. It is the template that init writes whole and that `level --to`
// rewrites one section of, so the bytes here are the contract both commands
// share.
func TestBlockForNovatoIsPinnedByteForByte(t *testing.T) {
	want := []string{
		"<!-- odachi:begin -->",
		"# Reglas del proyecto (odachi)",
		"",
		"odachi es un binario y cuatro artefactos de texto plano —brief, task doc, prueba y",
		"perfil—. El binario hace cumplir el gate; los artefactos son la memoria. La persona",
		"decide la visión y lo que no se negocia; el agente ejecuta con oficio. Si dudás",
		"dónde está el contexto: `brief.md` para la visión, `odd/tasks/` para el plan,",
		"`.agents/skills/` para el oficio.",
		"",
		"Cuatro reglas duras. No son consejos: `odachi check` las hace cumplir.",
		"",
		"1. Brief antes de código. Si no hay `brief.md` completo, primero `odachi start`.",
		"   Sin visión escrita no se construye.",
		"2. Nada fuera del cerco. Si algo está en \"NO incluye\", no se toca: se pregunta",
		"   con `odachi ask \"<pregunta>\"`. El cerco lo hace cumplir `check`.",
		"3. No decir \"listo\" sin prueba. Listo es `odachi check --run \"<comando>\"` en",
		"   verde, con la salida guardada en `proof/`. Sin eso, no está listo.",
		"4. Pedazos chicos, un commit cada uno. Un paso del task doc, un commit.",
		"",
		"Skills: las que hay en `.agents/skills/`; dentro de este proyecto, cargá `drive` antes de actuar: es el manual.",
		"",
		"## Los seis verbos",
		"",
		"- `odachi start` — la entrevista: escribe brief, perfil, cerco y task doc.",
		"- `odachi ask \"<pregunta>\"` — anota una pregunta abierta; con `--answer` la contesta.",
		"- `odachi level` — muestra tus hechos y propone senior cuando se gana; con `--to <perfil>` aplica el perfil y esta voz.",
		"- `odachi summary` — qué hay hecho en cada feature.",
		"- `odachi check` — ¿está listo? brief, cerco, preguntas y prueba; con `--run` corre la prueba.",
		"- `odachi init` — escribe este bloque, el hook, las skills y la voz del perfil.",
		"",
		"## Cómo hablás",
		"",
		"El objetivo no es solo que funcione: es que la persona aprenda el oficio.",
		"Estas reglas son la diferencia entre un agente que ejecuta y un agente que enseña.",
		"Sos el tutor, no el que hace.",
		"",
		"- Explicá el porqué de cada decisión en UNA línea, en el momento en que la tomás.",
		"- Cada decisión de oficio se anota: odachi ask --craft \"<decisión>\" --because \"<razón>\".",
		"- Cuando frenás algo, decí por qué en una línea: es el momento que más enseña.",
		"- Una de cada cinco decisiones, preguntá de vuelta \"¿por qué elegiste esto?\".",
		"- Nunca un sermón, nunca un glosario, nunca dos explicaciones seguidas.",
		"- Al cerrar una feature, corré odachi summary y contá en tres líneas qué decidió",
		"  la persona y qué concepto apareció.",
		"<!-- odachi:end -->",
	}

	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got := linesOf(block)
	if len(got) != len(want) {
		t.Fatalf("el bloque tiene %d líneas, se esperaban %d:\n%s", len(got), len(want), block)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("línea %d del bloque:\n got: %q\nwant: %q", i+1, got[i], want[i])
		}
	}
}

// One template, two voices: the block for senior and the block for novato differ
// only in the body of the last section. Anything else drifting means the rules
// depend on who is reading, which is how the two copies odachi used to write
// contradicted each other.
func TestTheTwoVoicesDifferOnlyInTheVoiceSection(t *testing.T) {
	novato, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}

	if novato == senior {
		t.Fatalf("las dos voces producen el mismo bloque")
	}

	// Everything up to and including the heading is the template, and it is the
	// same bytes for both.
	head := strings.Index(novato, rules.Section) + len(rules.Section)
	if head < len(rules.Section) {
		t.Fatalf("el bloque no tiene la sección %q", rules.Section)
	}
	if novato[:head] != senior[:head] {
		t.Fatalf("el bloque cambia antes de la sección de voz:\nnovato: %q\nsenior: %q", novato[:head], senior[:head])
	}

	novatoBody, ok := rules.VoiceIn(novato)
	if !ok {
		t.Fatalf("la sección de voz no se deja leer:\n%s", novato)
	}
	seniorBody, ok := rules.VoiceIn(senior)
	if !ok {
		t.Fatalf("la sección de voz no se deja leer:\n%s", senior)
	}
	wantNovato, _ := persona.Voice(persona.Novato)
	wantSenior, _ := persona.Voice(persona.Senior)
	if novatoBody != strings.TrimSpace(wantNovato) {
		t.Errorf("la voz del bloque novato no es la de la persona:\n%q\nvs\n%q", novatoBody, wantNovato)
	}
	if seniorBody != strings.TrimSpace(wantSenior) {
		t.Errorf("la voz del bloque senior no es la de la persona:\n%q\nvs\n%q", seniorBody, wantSenior)
	}
	if strings.Contains(seniorBody, "Sos el tutor") {
		t.Errorf("la voz senior repite la del tutor:\n%s", seniorBody)
	}
}

// ci is a profile, not a person: an agent running in CI has nobody to teach, so
// the block it reads carries the senior voice.
func TestCIUsesTheSeniorVoice(t *testing.T) {
	ci, err := rules.Block("ci")
	if err != nil {
		t.Fatalf("Block(ci): %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if ci != senior {
		t.Errorf("el bloque de ci no es el de senior:\n%s", ci)
	}
}

// TestUnknownVoiceIsRefused: nothing is guessed for a name that is not a voice.
func TestUnknownVoiceIsRefused(t *testing.T) {
	if _, err := rules.Block("principiante"); err == nil {
		t.Fatalf("Block(principiante) devolvió un bloque")
	} else {
		for _, want := range []string{"principiante", persona.Novato, persona.Senior} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("el error no nombra %q: %v", want, err)
			}
		}
	}
}

// The first paragraph answers what the agent is working with and who decides
// what, before it is handed the rules. Without it the block talks to an agent
// that already knows.
func TestBlockSaysWhatOdachiIsAndWhereTheVoiceComesFrom(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	flat := strings.Join(strings.Fields(block), " ")
	for _, want := range []string{
		"odachi es un binario y cuatro artefactos de texto plano",
		"El binario hace cumplir el gate",
		"los artefactos son la memoria",
		"La persona decide la visión y lo que no se negocia",
		"el agente ejecuta con oficio",
		"`brief.md` para la visión",
		"`odd/tasks/` para el plan",
		"`.agents/skills/` para el oficio",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("el bloque no dice %q:\n%s", want, block)
		}
	}

	// The profile file is read by the binary only; that contract is not the
	// block's to publish.
	if strings.Contains(block, ".odachi/profile.toml") {
		t.Errorf("el bloque manda al agente al archivo de perfil:\n%s", block)
	}
}

// The voice no longer asks the agent to detect a level: the binary writes the
// voice it read from the profile, and an agent that asks again would be guessing
// what a file already decided.
func TestBlockDoesNotAskTheAgentToGuessTheVoice(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	for _, unwanted := range []string{
		"No lo adivines",
		"## Si la persona es novata",
		"¿preferís que te explique el porqué o que vaya al hueso?",
	} {
		if strings.Contains(block, unwanted) {
			t.Errorf("el bloque todavía le pide al agente que adivine el nivel (%q):\n%s", unwanted, block)
		}
	}

	// The voice is the last section, and it says so where the agent looks for it.
	headings := []string{}
	for _, line := range linesOf(block) {
		if strings.HasPrefix(line, "#") {
			headings = append(headings, line)
		}
	}
	if len(headings) == 0 || headings[len(headings)-1] != rules.Section {
		t.Errorf("la voz tiene que ser la última sección del bloque, encabezados: %v", headings)
	}
}

// TestWithBlockReplacesInPlaceAndKeepsTheRest: init owns the block and nothing
// outside it. A human's lines above and below survive every run.
func TestWithBlockReplacesInPlaceAndKeepsTheRest(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got, err := rules.WithBlock("# mío\n\n"+block+"\nmás mío\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if got != "# mío\n\n"+senior+"\nmás mío\n" {
		t.Errorf("el bloque no se reemplazó en su lugar:\n%q", got)
	}

	// A file with no block gains one at the end, after one blank line.
	appended, err := rules.WithBlock("# mío\n", persona.Novato)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	if appended != "# mío\n\n"+block {
		t.Errorf("el bloque no se agregó al final:\n%q", appended)
	}

	// A truncated block is repaired instead of duplicated.
	truncated, err := rules.WithBlock("antes\n<!-- odachi:begin -->\n# Reglas viejas\n", persona.Novato)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	if truncated != "antes\n"+block {
		t.Errorf("el bloque truncado no se reparó:\n%q", truncated)
	}
}

// TestWithVoiceRewritesOnlyTheSection: what `level --to` owns. Every other byte
// of the file is the human's and survives, including text the agent wrote
// outside the block.
func TestWithVoiceRewritesOnlyTheSection(t *testing.T) {
	novato, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got, changed, err := rules.WithVoice("mío arriba\n\n"+novato+"\nmío abajo\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if !changed {
		t.Fatalf("WithVoice dijo que no cambió nada")
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if got != "mío arriba\n\n"+senior+"\nmío abajo\n" {
		t.Errorf("una sola sección cambiaba y cambió más:\n%q", got)
	}

	// Identical bytes are never a change, so a rerun writes nothing.
	again, changed, err := rules.WithVoice(got, persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if changed || again != got {
		t.Errorf("WithVoice reescribió una voz idéntica")
	}
}

// A block written before the voice moved into it has no voice section. It gains
// one as the last section, and every byte that was already there stays.
func TestWithVoiceAddsTheSectionToAnOldBlock(t *testing.T) {
	old := "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas viejas\n<!-- odachi:end -->\n"

	got, changed, err := rules.WithVoice("mío\n\n"+old, persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if !changed {
		t.Fatalf("WithVoice no agregó la sección")
	}

	want := "mío\n\n<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas viejas\n\n" +
		rules.Section + "\n\n" + mustVoice(t, persona.Senior) + "<!-- odachi:end -->\n"
	if got != want {
		t.Errorf("la sección no se agregó al final del bloque:\n got %q\nwant %q", got, want)
	}

	// A file with no block is not level's to invent.
	untouched, changed, err := rules.WithVoice("solo mío\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if changed || untouched != "solo mío\n" {
		t.Errorf("WithVoice escribió en un archivo sin bloque: %q", untouched)
	}
	if !rules.HasBlock(old) {
		t.Errorf("HasBlock no reconoció un bloque completo")
	}
	if rules.HasBlock("solo mío\n") {
		t.Errorf("HasBlock dijo que sí a un archivo sin bloque")
	}
	if _, ok := rules.VoiceIn(old); ok {
		t.Errorf("VoiceIn inventó una voz para un bloque sin sección")
	}
}

// TestBodyIsTheBlockWithoutItsMarkers: the native agent mirrors exactly what the
// agent reads, and the markers are plumbing for in-place edits, not rules.
func TestBodyIsTheBlockWithoutItsMarkers(t *testing.T) {
	block, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	body, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("Body(senior): %v", err)
	}

	if strings.Contains(body, rules.Begin) || strings.Contains(body, rules.End) {
		t.Errorf("el cuerpo del bloque lleva los marcadores:\n%s", body)
	}
	if strings.HasPrefix(body, "\n") {
		t.Errorf("el cuerpo arranca con un salto de línea: %q", body[:10])
	}
	if !strings.HasSuffix(body, "\n") || strings.HasSuffix(body, "\n\n") {
		t.Errorf("el cuerpo no termina en exactamente un salto de línea: %q", body[len(body)-10:])
	}
	want := rules.Begin + "\n" + body + rules.End + "\n"
	if block != want {
		t.Errorf("Block no es Begin + Body + End:\n%q\nvs\n%q", block, want)
	}

	got, ok := rules.BodyIn(block)
	if !ok {
		t.Fatalf("BodyIn no reconoció el bloque")
	}
	if got != body {
		t.Errorf("BodyIn no devuelve el cuerpo:\n%q\nvs\n%q", got, body)
	}
	for _, content := range []string{"", "sin bloque\n", rules.Begin + "\ntruncado\n"} {
		if _, ok := rules.BodyIn(content); ok {
			t.Errorf("BodyIn inventó un cuerpo para %q", content)
		}
	}
}

// mustVoice is the voice text, or a failed test: the template tests compare
// against the persona text, so a missing one is not a test to skip.
func mustVoice(t *testing.T, name string) string {
	t.Helper()
	voice, err := persona.Voice(name)
	if err != nil {
		t.Fatalf("persona.Voice(%q): %v", name, err)
	}
	return voice
}
