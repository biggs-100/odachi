package init

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// linesOf splits the block into the lines a model reads.
func linesOf(block string) []string {
	return strings.Split(strings.TrimSuffix(block, "\n"), "\n")
}

// headingsOf is every line that starts a section, in order: the skeleton a
// reader navigates by.
func headingsOf(block string) []string {
	headings := []string{}
	for _, line := range linesOf(block) {
		if strings.HasPrefix(line, "#") {
			headings = append(headings, line)
		}
	}
	return headings
}

// The block is the only text odachi writes that an agent obeys, so it is pinned
// byte for byte: any edit to the voice or to a rule has to pass through this
// test. The previous limit (L28) was `contains` assertions, which let the tutor
// section be dropped without a single failure.
func TestAgentsBlockIsPinnedByteForByte(t *testing.T) {
	want := []string{
		"<!-- odachi:begin -->",
		"# Reglas del proyecto (odachi)",
		"",
		"odachi es un binario y cuatro artefactos de texto plano —brief, task doc, prueba y",
		"perfil—. El binario hace cumplir el gate; los artefactos son la memoria. La persona",
		"decide la visión y lo que no se negocia; el agente ejecuta con oficio de senior y",
		"explica lo que la persona necesita aprender. Si dudás dónde está el contexto:",
		"`brief.md` para la visión, `odd/tasks/` para el plan, `SKILLS.md` para el oficio.",
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
		"Skills: ver SKILLS.md; cargá la que matchee el trigger antes de escribir código.",
		"",
		"## Si la persona es novata",
		"",
		"No lo adivines: si no sabés el nivel, preguntá una vez",
		"\"¿preferís que te explique el porqué o que vaya al hueso?\" y aplicá lo de abajo",
		"solo si la respuesta es novato.",
		"",
		"El objetivo no es solo que funcione: es que la persona aprenda el oficio. Estas",
		"reglas son la diferencia entre un agente que ejecuta y un agente que enseña.",
		"",
		"- Explicá el porqué de cada decisión en UNA línea, en el momento en que la tomás.",
		"- Cada decisión de oficio se anota: `odachi ask --craft \"<decisión>\" --because \"<razón>\"`.",
		"- Cuando frenás algo, decí por qué en una línea: es el momento que más enseña.",
		"- Una de cada cinco decisiones, preguntá de vuelta \"¿por qué elegiste esto?\".",
		"- Nunca un sermón, nunca un glosario, nunca dos explicaciones seguidas.",
		"- Al cerrar una feature, corré `odachi summary` y contá en tres líneas qué",
		"  decidió la persona y qué concepto apareció.",
		"",
		"## Los seis verbos",
		"",
		"- `odachi start` — la entrevista: escribe brief, perfil, cerco y task doc.",
		"- `odachi ask \"<pregunta>\"` — anota una pregunta abierta; con `--answer` la contesta.",
		"- `odachi level` — muestra tus hechos y propone senior cuando se gana.",
		"- `odachi summary` — qué hay hecho en cada feature.",
		"- `odachi check` — ¿está listo? brief, cerco, preguntas y prueba; con `--run` corre la prueba.",
		"- `odachi init` — escribe este bloque y el hook pre-commit.",
		"<!-- odachi:end -->",
	}

	got := linesOf(agentsBlock)
	if len(got) != len(want) {
		t.Fatalf("el bloque tiene %d líneas, se esperaban %d:\n%s", len(got), len(want), agentsBlock)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("línea %d del bloque:\n got: %q\nwant: %q", i+1, got[i], want[i])
		}
	}
}

// The first paragraph answers what the agent is working with and who decides
// what, before it is handed the rules. Without it the block talks to an agent
// that already knows.
func TestAgentsBlockSaysWhatOdachiIsAndWhoDecides(t *testing.T) {
	// Wrapping is formatting, not content: the byte pin guards the line breaks and
	// this test guards the meaning, so it reads the block as one line.
	flat := strings.Join(strings.Fields(agentsBlock), " ")
	for _, want := range []string{
		"odachi es un binario y cuatro artefactos de texto plano",
		"El binario hace cumplir el gate",
		"los artefactos son la memoria",
		"La persona decide la visión y lo que no se negocia",
		"el agente ejecuta con oficio de senior",
		"explica lo que la persona necesita aprender",
		"`brief.md` para la visión",
		"`odd/tasks/` para el plan",
		"`SKILLS.md` para el oficio",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("el bloque no dice %q:\n%s", want, agentsBlock)
		}
	}

	// The profile file is read by the binary only; that contract is not the
	// block's to publish.
	if strings.Contains(agentsBlock, ".odachi/profile.toml") {
		t.Errorf("el bloque manda al agente al archivo de perfil:\n%s", agentsBlock)
	}
}

// The tutor rules are only usable if the agent knows it can ask first: a novice
// who gets senior treatment learns nothing, and a senior who gets taught the
// basics is being talked down to.
func TestAgentsBlockAsksBeforeTeaching(t *testing.T) {
	if got := headingsOf(agentsBlock); !slices.Contains(got, "## Si la persona es novata") {
		t.Fatalf("falta el encabezado del tutor: %v", got)
	}
	if strings.Contains(agentsBlock, "Si el perfil es novato") {
		t.Errorf("el encabezado viejo sobrevive: nombra un archivo que el agente no lee")
	}

	const ask = "¿preferís que te explique el porqué o que vaya al hueso?"
	if !strings.Contains(agentsBlock, ask) {
		t.Fatalf("el bloque no dice qué preguntar: falta %q", ask)
	}
	if !strings.Contains(agentsBlock, "No lo adivines") {
		t.Errorf("el bloque no prohíbe adivinar el nivel")
	}

	// The question comes before the tutor rules: it gates them.
	question := strings.Index(agentsBlock, ask)
	firstRule := strings.Index(agentsBlock, "- Explicá el porqué")
	if question < 0 || firstRule < 0 || question > firstRule {
		t.Errorf("la pregunta tiene que ir antes de las reglas del tutor")
	}
}

// The four hard rules and the six verbs are what the block exists for, so a
// count and a set, not a sentence that happens to mention them.
func TestAgentsBlockKeepsTheFourRulesAndTheSixVerbs(t *testing.T) {
	heading := strings.Index(agentsBlock, "# Reglas del proyecto (odachi)")
	verbsHeading := strings.Index(agentsBlock, "## Los seis verbos")
	if heading < 0 || verbsHeading < 0 || verbsHeading < heading {
		t.Fatalf("el bloque perdió su encabezado o la sección de verbos")
	}

	rules := []string{
		"1. Brief antes de código",
		"2. Nada fuera del cerco",
		"3. No decir \"listo\" sin prueba",
		"4. Pedazos chicos, un commit cada uno",
	}
	numbered := regexp.MustCompile(`(?m)^[0-9]+\. `)
	if got := numbered.FindAllString(agentsBlock, -1); len(got) != len(rules) {
		t.Fatalf("el bloque tiene %d reglas numeradas, se esperaban %d", len(got), len(rules))
	}
	for _, want := range rules {
		if !strings.Contains(agentsBlock, want) {
			t.Errorf("falta la regla %q", want)
		}
	}

	wantVerbs := []string{"start", "ask", "level", "summary", "check", "init"}
	verbs := []string{}
	for _, line := range linesOf(agentsBlock[verbsHeading:])[1:] {
		if line == agentsEnd {
			break
		}
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "- `odachi ") {
			t.Fatalf("la sección de verbos tiene una línea que no es un verbo: %q", line)
		}
		rest := strings.TrimPrefix(line, "- `odachi ")
		name := rest
		if i := strings.IndexAny(rest, " `"); i >= 0 {
			name = rest[:i]
		}
		verbs = append(verbs, name)
	}
	if !slices.Equal(verbs, wantVerbs) {
		t.Fatalf("los verbos son %v, se esperaban %v", verbs, wantVerbs)
	}
}

// flatText is one string compared by meaning instead of by line breaks: wrapping
// is formatting, and the two texts wrap differently on purpose. Backticks are the
// block's markdown, not part of what it says.
func flatText(text string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(text, "`", "")), " ")
}

// bulletsOf joins the wrapped continuation lines of a list, so a rule that spans
// two lines is compared as one sentence.
func bulletsOf(section string) []string {
	var bullets []string
	for _, line := range linesOf(section) {
		switch {
		case strings.HasPrefix(line, "- "):
			bullets = append(bullets, line[2:])
		case strings.HasPrefix(line, "  ") && len(bullets) > 0:
			bullets[len(bullets)-1] += " " + strings.TrimSpace(line)
		}
	}
	return bullets
}

// TestTheNovatoPersonaRepeatsTheTutorSection: the block teaches the agent how to
// behave with a novice, and the persona says the same thing in the file the
// client loads as a system prompt. Two texts, one voice: if they drift, odachi
// contradicts itself depending on which file the agent read, and the human gets
// two rules for the same situation.
func TestTheNovatoPersonaRepeatsTheTutorSection(t *testing.T) {
	body, err := persona.Body(persona.Novato)
	if err != nil {
		t.Fatalf("persona.Body(novato): %v", err)
	}
	flatBody := flatText(body)

	begin := strings.Index(agentsBlock, "## Si la persona es novata")
	end := strings.Index(agentsBlock, "## Los seis verbos")
	if begin < 0 || end < begin {
		t.Fatalf("el bloque perdió la sección del tutor")
	}
	section := agentsBlock[begin:end]

	for _, want := range []string{
		"El objetivo no es solo que funcione: es que la persona aprenda el oficio.",
		"Estas reglas son la diferencia entre un agente que ejecuta y un agente que enseña.",
	} {
		if !strings.Contains(flatBody, want) {
			t.Errorf("la persona novato no repite la frase del bloque %q:\n%s", want, body)
		}
	}

	rules := bulletsOf(section)
	if len(rules) == 0 {
		t.Fatalf("la sección del tutor no tiene reglas")
	}
	for _, rule := range rules {
		if !strings.Contains(flatBody, flatText(rule)) {
			t.Errorf("la persona novato no repite la regla del bloque %q:\n%s", flatText(rule), body)
		}
	}
}
