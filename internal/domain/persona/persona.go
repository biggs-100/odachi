// Package persona renders the two personas odachi writes into a client's own
// mechanism for choosing how the agent talks: novato (the person learns the
// craft) and senior (to the point, no teaching unless asked).
//
// Two axes, two owners. The profile owns the ceremony —how much the binary asks,
// what counts as proof, what it prints— and it lives in .odachi/profile.toml,
// which the model never reads. The persona owns how the agent talks, and it lives
// where the client already lets a human pick it with a keystroke. Conflating them
// is what left `explain = why|none` in a file the model never sees: one word,
// two sources of truth. This package never reads the profile, and every persona
// says so on its second line.
//
// Both personas are written, always. `odachi init` cannot know who the human is,
// and drawing the persona from the profile would push the choice back into the
// TOML this change exists to get it out of. The client's own selector decides,
// and the binary stays the only owner of the gates.
//
// The text is one template. Body is the same bytes for every client; a client
// only supplies the wrapper it needs (frontmatter and directory). The client
// table here is a sibling of the skills table in internal/domain/agent, not a
// copy of it: agent owns which programs odachi knows and where each one reads
// skills, and this table hangs off those same names — a test fails if a row here
// names a client the skills table does not — because a persona destination and a
// skills destination are different facts about the same program, and folding the
// second into the first would make every skills path carry a field that has
// nothing to do with skills.
//
// It is pure: no filesystem, no os. The app layer decides where the project is
// and writes the files.
package persona

import (
	"fmt"
	"strings"
)

// The two personas, by the name the human reads. They are the same two names the
// profiles use, on purpose: the human thinks of one choice, and odachi is the one
// that keeps the two axes apart.
const (
	// Novato is the persona that teaches the craft while it builds.
	Novato = "novato"
	// Senior is the persona that goes to the point.
	Senior = "senior"
)

// generated is the first line of every persona file. Like SKILLS.md, the file
// says out loud that it is written by a command, because a hand edit is lost work
// and the human deserves to know before making it.
const generated = "<!-- generado por odachi init; no lo edites a mano: se reescribe entero cada vez que volvés a correr odachi init -->"

// nonDivergence is the second line of every persona file: the answer to the one
// question two files about the same word create. It stays on a single line
// because it is the line an agent reads when the two disagree.
const nonDivergence = "La ceremonia la decide .odachi/profile.toml; este archivo decide solo cómo hablás. Si los dos no coinciden, decilo y pedí odachi level --to <perfil>."

// voices is the one template: what each persona says, after the two header lines.
// The novato text reuses the sentences of the tutor section in the AGENTS.md
// block on purpose — the tool cannot say one thing in the block and another in
// the persona — and a test in internal/app/init pins that they still match.
var voices = map[string]string{
	Novato: `El objetivo no es solo que funcione: es que la persona aprenda el oficio.
Estas reglas son la diferencia entre un agente que ejecuta y un agente que enseña.
Sos el tutor, no el que hace.

- Explicá el porqué de cada decisión en UNA línea, en el momento en que la tomás.
- Cada decisión de oficio se anota: odachi ask --craft "<decisión>" --because "<razón>".
- Cuando frenás algo, decí por qué en una línea: es el momento que más enseña.
- Una de cada cinco decisiones, preguntá de vuelta "¿por qué elegiste esto?".
- Nunca un sermón, nunca un glosario, nunca dos explicaciones seguidas.
- Al cerrar una feature, corré odachi summary y contá en tres líneas qué decidió
  la persona y qué concepto apareció.
`,
	Senior: `Andá al hueso: respuesta breve, sin enseñanza salvo que la pidan.

- Las mismas reglas duras del proyecto: brief antes de código, nada fuera del
  cerco, no decir "listo" sin prueba, pedazos chicos y un commit cada uno.
- Una línea de porqué solo cuando la decisión no es obvia; en el resto, la
  decisión y nada más.
- Si falta el brief o algo está fuera del cerco, no lo completes: preguntá con
  odachi ask.
`,
}

// descriptions is the one line a client shows next to the persona in its own
// selector. It is part of the wrapper and not of the body because two clients
// store it in two different frontmatter fields; the text is written once here.
// No double quotes: it is rendered inside a YAML string.
var descriptions = map[string]string{
	Novato: "Enseña el oficio mientras construye; una línea de porqué por decisión",
	Senior: "Al hueso; sin enseñanza salvo que la pidan",
}

// Client is one client's native mechanism for a persona: the agent it belongs to
// in the skills table, the project directory that client reads personas from
// (relative to the project root, forward slashes), and how the human picks one
// there.
type Client struct {
	Agent  string
	Dir    string
	Switch string
}

// clients is the verified table, in the order the skills table lists the same
// programs. Each destination was read in the client's own documentation or in
// its own code before it shipped; the shape is per client because the mechanism
// is per client.
var clients = []Client{
	// OpenCode: `.opencode/agents/<name>.md`, frontmatter `description` and
	// `mode: primary`, and the markdown body becomes the agent's system prompt.
	// Primary is what puts the persona in the selector the human cycles with Tab;
	// a subagent is only reachable by the model. The scanner also accepts the
	// singular `agent/`; the documented directory the current docs describe is
	// the plural one, so that is the one written here.
	{
		Agent:  "opencode",
		Dir:    ".opencode/agents",
		Switch: "se eligen con Tab",
	},
	// Claude Code: `.claude/output-styles/<name>.md`, frontmatter `name`,
	// `description` and `keep-coding-instructions: true`, and the body becomes
	// the output style's instructions. keep-coding-instructions is not optional:
	// without it the style replaces Claude Code's own software engineering
	// instructions, and this file is about how the agent talks, never about
	// dropping its tools. The human picks it in /config (the /output-style
	// command that used to do it only prints a deprecation notice now).
	{
		Agent:  "claude",
		Dir:    ".claude/output-styles",
		Switch: "se eligen con /config (estilo de salida)",
	},
}

// wrappers renders the frontmatter one client needs around the shared body. It is
// keyed by agent name, so a row without a wrapper cannot exist: the two tables are
// built together and a test fails on a client that has no wrapper.
var wrappers = map[string]func(name, description string) string{
	"opencode": func(name, description string) string {
		return "---\ndescription: \"" + description + "\"\nmode: primary\n---\n"
	},
	"claude": func(name, description string) string {
		return "---\nname: " + styleName(name) + "\ndescription: \"" + description + "\"\nkeep-coding-instructions: true\n---\n"
	},
}

// File is one persona file: where it goes relative to the project root, and the
// bytes that go there.
type File struct {
	Path    string
	Content string
}

// Names returns the two personas, in the order init writes them.
func Names() []string { return []string{Novato, Senior} }

// Body returns the full body every client shares for this persona: the generated
// note, the non-divergence line, and the voice. An unknown name is an error that
// names the two that exist: nothing here guesses.
func Body(name string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	voice, ok := voices[key]
	if !ok {
		return "", fmt.Errorf("persona desconocida: %q; odachi escribe %s", name, strings.Join(Names(), " y "))
	}
	return header() + "\n" + voice, nil
}

// header is the two lines every persona file starts with, identical for both.
func header() string { return generated + "\n" + nonDivergence }

// Clients returns the clients that have a native mechanism, safe to keep: a
// caller that reorders or trims the result cannot change what this package
// resolves.
func Clients() []Client { return append([]Client(nil), clients...) }

// Files returns every file one client gets: one per persona. ok is false when the
// client has no native mechanism — `ci` is not here, because nobody picks an
// agent in CI — so nothing is guessed for a client odachi does not know.
func Files(client string) (files []File, ok bool) {
	row, found := find(client)
	if !found {
		return nil, false
	}
	out := make([]File, 0, len(voices))
	for _, name := range Names() {
		out = append(out, fileFor(row, name))
	}
	return out, true
}

// All returns every file odachi writes: both personas for every client that has a
// native mechanism, clients in table order.
func All() []File {
	out := make([]File, 0, len(clients)*len(voices))
	for _, row := range clients {
		for _, name := range Names() {
			out = append(out, fileFor(row, name))
		}
	}
	return out
}

// fileFor composes the one file for one client and one persona: that client's
// wrapper around the body every client shares.
func fileFor(row Client, name string) File {
	return File{
		Path:    row.Dir + "/" + fileName(name),
		Content: wrappers[row.Agent](name, descriptions[name]) + header() + "\n" + voices[name],
	}
}

// styleName is what a client calls the persona in its own selector, and the file
// every client gets. The odachi prefix keeps a generated persona from colliding
// with a style or an agent the human wrote and called novato; nothing odachi owns
// takes a name the human already used. The name has no extension: Claude Code
// shows `name` as written, next to the styles the human already has.
func styleName(name string) string { return "odachi-" + name }

// fileName is the file the style is stored in.
func fileName(name string) string { return styleName(name) + ".md" }

// find returns the row for an already-normalized name. It trims spaces and
// ignores capitals, because a human types both.
func find(client string) (Client, bool) {
	key := strings.ToLower(strings.TrimSpace(client))
	for _, row := range clients {
		if row.Agent == key {
			return row, true
		}
	}
	return Client{}, false
}
