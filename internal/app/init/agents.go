package init

import "strings"

// The AGENTS.md block lives between two HTML comments, so replacing it never
// touches a line the human wrote around it.
const (
	agentsFile  = "AGENTS.md"
	agentsBegin = "<!-- odachi:begin -->"
	agentsEnd   = "<!-- odachi:end -->"
)

// agentsBlock is the fixed block `odachi init` writes: what odachi is, the four
// hard rules and the six verbs, each with its use in one line. It is pinned byte
// for byte by test because it is the contract the agent reads on every session,
// and it has to make sense to an agent that knows nothing about this project.
const agentsBlock = agentsBegin + `
# Reglas del proyecto (odachi)

odachi es un binario y cuatro artefactos de texto plano —brief, task doc, prueba y
perfil—. El binario hace cumplir el gate; los artefactos son la memoria. La persona
decide la visión y lo que no se negocia; el agente ejecuta con oficio de senior y
explica lo que la persona necesita aprender. Si dudás dónde está el contexto:
` + "`brief.md`" + ` para la visión, ` + "`odd/tasks/`" + ` para el plan, ` + "`.agents/skills/`" + ` para el oficio.

Cuatro reglas duras. No son consejos: ` + "`odachi check`" + ` las hace cumplir.

1. Brief antes de código. Si no hay ` + "`brief.md`" + ` completo, primero ` + "`odachi start`" + `.
   Sin visión escrita no se construye.
2. Nada fuera del cerco. Si algo está en "NO incluye", no se toca: se pregunta
   con ` + "`odachi ask \"<pregunta>\"`" + `. El cerco lo hace cumplir ` + "`check`" + `.
3. No decir "listo" sin prueba. Listo es ` + "`odachi check --run \"<comando>\"`" + ` en
   verde, con la salida guardada en ` + "`proof/`" + `. Sin eso, no está listo.
4. Pedazos chicos, un commit cada uno. Un paso del task doc, un commit.

Skills: las que hay en ` + "`.agents/skills/`" + `; dentro de este proyecto, cargá ` + "`drive`" + ` antes de actuar: es el manual.

## Si la persona es novata

No lo adivines: si no sabés el nivel, preguntá una vez
"¿preferís que te explique el porqué o que vaya al hueso?" y aplicá lo de abajo
solo si la respuesta es novato.

El objetivo no es solo que funcione: es que la persona aprenda el oficio. Estas
reglas son la diferencia entre un agente que ejecuta y un agente que enseña.

- Explicá el porqué de cada decisión en UNA línea, en el momento en que la tomás.
- Cada decisión de oficio se anota: ` + "`odachi ask --craft \"<decisión>\" --because \"<razón>\"`" + `.
- Cuando frenás algo, decí por qué en una línea: es el momento que más enseña.
- Una de cada cinco decisiones, preguntá de vuelta "¿por qué elegiste esto?".
- Nunca un sermón, nunca un glosario, nunca dos explicaciones seguidas.
- Al cerrar una feature, corré ` + "`odachi summary`" + ` y contá en tres líneas qué
  decidió la persona y qué concepto apareció.

## Los seis verbos

- ` + "`odachi start`" + ` — la entrevista: escribe brief, perfil, cerco y task doc.
- ` + "`odachi ask \"<pregunta>\"`" + ` — anota una pregunta abierta; con ` + "`--answer`" + ` la contesta.
- ` + "`odachi level`" + ` — muestra tus hechos y propone senior cuando se gana.
- ` + "`odachi summary`" + ` — qué hay hecho en cada feature.
- ` + "`odachi check`" + ` — ¿está listo? brief, cerco, preguntas y prueba; con ` + "`--run`" + ` corre la prueba.
- ` + "`odachi init`" + ` — escribe este bloque, el hook, las skills y las dos personas.
` + agentsEnd + "\n"

// withBlock returns content with the odachi block replaced in place, or appended
// when the file has no block yet. A file with an opening marker and no closing
// one is treated as a truncated block: it is replaced from the marker to the end
// instead of gaining a second one.
func withBlock(content string) string {
	begin := strings.Index(content, agentsBegin)
	if begin < 0 {
		return appendBlock(content)
	}

	end := strings.Index(content[begin:], agentsEnd)
	if end < 0 {
		return content[:begin] + agentsBlock
	}
	rest := content[begin+end+len(agentsEnd):]
	if newline := strings.IndexByte(rest, '\n'); newline >= 0 {
		rest = rest[newline+1:]
	} else {
		rest = ""
	}
	return content[:begin] + agentsBlock + rest
}

// appendBlock puts the block at the end of a file that does not have one,
// separated by one blank line, and keeps every byte that was already there.
func appendBlock(content string) string {
	if strings.TrimSpace(content) == "" {
		return agentsBlock
	}
	return strings.TrimRight(content, "\n") + "\n\n" + agentsBlock
}
