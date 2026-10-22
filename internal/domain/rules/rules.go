// Package rules renders the odachi block of AGENTS.md — the one text an agent
// obeys on every session — with the voice of the profile in force as its last
// section.
//
// One owner for one text. `odachi init` writes the whole block and `odachi level
// --to <perfil>` rewrites the voice section in place, and both go through this
// package: a second copy of the rules is a second set of rules, and an agent
// that reads one while the binary enforces the other is being lied to.
//
// The block is the artifact, not a summary of it: it says what odachi is, the
// four hard rules the gate enforces, the six verbs, and how the agent talks. It
// has to make sense to an agent that knows nothing about this project, which is
// why the text is pinned byte for byte by test instead of asserted by pieces.
//
// It is pure: no filesystem, no os. The app layer decides where the project is
// and writes the bytes.
package rules

import (
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

const (
	// File is where the block lives, relative to the project root. Every client
	// odachi knows reads AGENTS.md; that is the whole reason the voice moved
	// here from a per-client file nobody's selector listed.
	File = "AGENTS.md"
	// Begin and End are the markers the block lives between, so replacing it
	// never touches a line the human wrote around it.
	Begin = "<!-- odachi:begin -->"
	End   = "<!-- odachi:end -->"
	// Section is the heading of the block's last section: how the agent talks.
	// It is the one part of the block that changes with the profile, and the
	// only part `odachi level --to` rewrites.
	Section = "## Cómo hablás"
)

// head is the block up to and including the blank line that opens the voice
// section. It is written once and closed by Block, which is what keeps the two
// voices from differing anywhere but in that section: everything here is the
// same bytes whatever voice is in force.
const head = Begin + `
# Reglas del proyecto (odachi)

odachi es un binario y cuatro artefactos de texto plano —brief, task doc, prueba y
perfil—. El binario hace cumplir el gate; los artefactos son la memoria. La persona
decide la visión y lo que no se negocia; el agente ejecuta con oficio. Si dudás
dónde está el contexto: ` + "`brief.md`" + ` para la visión, ` + "`odd/tasks/`" + ` para el plan,
` + "`.agents/skills/`" + ` para el oficio.

Cuatro reglas duras. No son consejos: ` + "`odachi check`" + ` las hace cumplir.

1. Brief antes de código. Si no hay ` + "`brief.md`" + ` completo, primero ` + "`odachi start`" + `.
   Sin visión escrita no se construye.
2. Nada fuera del cerco. Si algo está en "NO incluye", no se toca: se pregunta
   con ` + "`odachi ask \"<pregunta>\"`" + `. El cerco lo hace cumplir ` + "`check`" + `.
3. No decir "listo" sin prueba. Listo es ` + "`odachi check --run \"<comando>\"`" + ` en
   verde, con la salida guardada en ` + "`proof/`" + `. Sin eso, no está listo.
4. Pedazos chicos, un commit cada uno. Un paso del task doc, un commit.

Skills: las que hay en ` + "`.agents/skills/`" + `; dentro de este proyecto, cargá ` + "`drive`" + ` antes de actuar: es el manual.

## Los seis verbos

- ` + "`odachi start`" + ` — la entrevista: escribe brief, perfil, cerco y task doc.
- ` + "`odachi ask \"<pregunta>\"`" + ` — anota una pregunta abierta; con ` + "`--answer`" + ` la contesta.
- ` + "`odachi level`" + ` — muestra tus hechos y propone senior cuando se gana; con ` + "`--to <perfil>`" + ` aplica el perfil y esta voz.
- ` + "`odachi summary`" + ` — qué hay hecho en cada feature.
- ` + "`odachi check`" + ` — ¿está listo? brief, cerco, preguntas y prueba; con ` + "`--run`" + ` corre la prueba.
- ` + "`odachi init`" + ` — escribe este bloque, el hook, las skills y la voz del perfil.

` + Section + `

`

// Block returns the whole block for one voice, trailing newline included. voice
// is a profile name: novato and senior are their own, and `ci` gets senior's.
// An unknown voice is an error naming the two that exist — nothing here guesses
// what an unrecognised name should sound like.
func Block(voice string) (string, error) {
	body, err := persona.Voice(voice)
	if err != nil {
		return "", err
	}
	return head + body + End + "\n", nil
}

// WithBlock returns content with the block replaced in place, or appended when
// the file has no block yet. It is `odachi init`'s write: the block is generated
// whole, so an upgrade refreshes every rule at once.
//
// A file with an opening marker and no closing one is treated as a truncated
// block: it is replaced from the marker to the end instead of gaining a second
// one.
func WithBlock(content, voice string) (string, error) {
	block, err := Block(voice)
	if err != nil {
		return content, err
	}

	begin := strings.Index(content, Begin)
	if begin < 0 {
		return appendBlock(content, block), nil
	}

	end := strings.Index(content[begin:], End)
	if end < 0 {
		return content[:begin] + block, nil
	}
	rest := content[begin+end+len(End):]
	if newline := strings.IndexByte(rest, '\n'); newline >= 0 {
		rest = rest[newline+1:]
	} else {
		rest = ""
	}
	return content[:begin] + block + rest, nil
}

// appendBlock puts the block at the end of a file that does not have one,
// separated by one blank line, and keeps every byte that was already there.
func appendBlock(content, block string) string {
	if strings.TrimSpace(content) == "" {
		return block
	}
	return strings.TrimRight(content, "\n") + "\n\n" + block
}

// WithVoice returns content with the block's voice section rewritten in place.
// It is `odachi level --to`'s write, and it owns exactly one section: every other
// byte of the file —the rules, the verbs, whatever the human wrote above and
// below— is preserved, the way the .gitignore block preserves its file.
//
// A block written before the voice moved into it has no section, and it gains one
// as the last section of the block. A file with no block is returned untouched:
// writing the whole block is init's job, and level never invents a block for a
// file odachi did not write.
//
// changed is false when the bytes are already what they should be, so a second
// run writes nothing at all.
func WithVoice(content, voice string) (string, bool, error) {
	body, err := persona.Voice(voice)
	if err != nil {
		return content, false, err
	}

	start, end, insert, ok := voiceSpan(content)
	if !ok {
		return content, false, nil
	}
	if insert {
		added := content[:start] + "\n" + Section + "\n\n" + body + content[start:]
		return added, true, nil
	}

	updated := content[:start] + "\n" + body + content[end:]
	return updated, updated != content, nil
}

// VoiceIn returns the voice a file already publishes, and whether it publishes
// one at all. A block written before the voice moved into it has no section, and
// that is a state to report, not a voice to guess. The text comes back trimmed,
// because a human's trailing blank line is not a different voice.
func VoiceIn(content string) (string, bool) {
	start, end, insert, ok := voiceSpan(content)
	if !ok || insert {
		return "", false
	}
	return strings.TrimSpace(content[start:end]), true
}

// HasBlock reports whether content carries a complete odachi block. A truncated
// block is not one: `odachi init` repairs it, and a check that called it complete
// would leave the agent without rules.
func HasBlock(content string) bool {
	_, _, _, ok := voiceSpan(content)
	return ok
}

// voiceSpan locates the bytes the voice section's body owns. insert says the
// block has no voice section yet, and then start == end is the line it belongs
// before: the end of the block, so the voice is always the last section.
func voiceSpan(text string) (start, end int, insert, ok bool) {
	begin := lineStart(text, Begin)
	if begin < 0 {
		return 0, 0, false, false
	}

	offset := lineStart(text[begin:], End)
	if offset < 0 {
		return 0, 0, false, false
	}
	endMarker := begin + offset

	heading := lineStart(text[begin:endMarker], Section)
	if heading < 0 {
		return endMarker, endMarker, true, true
	}

	start = begin + heading + len(Section)
	if strings.HasPrefix(text[start:], "\n") {
		start++
	}

	// The body stops at the next heading line, so a section that is not last —
	// a human moved it, or an older block grew one after it — keeps what follows.
	end = endMarker
	for from := start; from < endMarker; {
		newline := strings.IndexByte(text[from:endMarker], '\n')
		if newline < 0 {
			break
		}
		lineStart := from + newline + 1
		if lineStart >= endMarker {
			break
		}
		if text[lineStart] == '#' {
			end = lineStart
			break
		}
		from = lineStart
	}
	return start, end, false, true
}

// lineStart is the index of the marker when it owns its own line, and -1 when it
// does not: a marker inside another line is not a block this tool wrote.
func lineStart(text, marker string) int {
	for from := 0; from < len(text); {
		index := strings.Index(text[from:], marker)
		if index < 0 {
			return -1
		}
		index += from
		if index == 0 || text[index-1] == '\n' {
			return index
		}
		from = index + len(marker)
	}
	return -1
}
