// Package persona owns the two voices odachi writes and the native agent file
// that lets a client select the project's agent with its own keystroke.
//
// One voice, one owner. `.odachi/profile.toml` decides the ceremony and the
// voice; `AGENTS.md` publishes that voice as the last section of the block; the
// native agent file mirrors the same block for the clients that have a selector.
// The mirror is generated, never hand-written, and `odachi init`, `odachi level
// --to` and `odachi start` keep it in step. There are no two axes: a human who
// wants a different voice runs `odachi level --to`, and every surface follows.
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
	"slices"
	"strings"
)

// The two voices, by the name the human reads. They are the same two names the
// profiles use, on purpose: the human thinks of one choice.
const (
	// Novato is the voice that teaches the craft while it builds.
	Novato = "novato"
	// Senior is the voice that goes to the point.
	Senior = "senior"
	// CI is the profile of a machine: it gets the senior voice, because nobody
	// is reading the answers.
	CI = "ci"
)

// AgentName is the agent a client shows in its selector, and FileName is the one
// file odachi writes for it. The odachi prefix keeps a generated agent from
// colliding with one the human wrote and called odachi; one file per client
// means the profile changing is a rewrite, never a second stale persona.
const (
	AgentName = "odachi"
	FileName  = AgentName + ".md"
)

// generated is the first line of every native agent file. Like SKILLS.md, the
// file says out loud that it is written by a command, because a hand edit is
// lost work and the human deserves to know before making it.
const generated = "<!-- generado por odachi init; no lo edites a mano: se reescribe entero cada vez que volvés a correr odachi init -->"

// voiceSeal is the one line every native agent file carries after the generated
// note: it says where the voice comes from, so an agent that finds the two files
// disagreeing knows which command repairs them.
const voiceSeal = "La voz de este agente es la del perfil en .odachi/profile.toml y la del bloque de AGENTS.md; si no coinciden, pedí odachi level --to <perfil>."

// voices is the one template: what each voice says. The novato text reuses the
// sentences of the tutor section in the AGENTS.md block on purpose — the tool
// cannot say one thing in the block and another in the native agent — and a test
// in internal/app/init pins that they still match.
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

// Client is one client's native mechanism for an agent: the name the skills
// table uses, the project directory that client reads agents from (relative to
// the project root, forward slashes), and how the human picks one there.
type Client struct {
	Agent  string
	Dir    string
	Switch string
}

// clients is the verified table, in the order the skills table lists the same
// programs. Each destination was read in the client's own documentation or in
// its own code before it shipped.
var clients = []Client{
	// OpenCode: `.opencode/agents/<name>.md`, frontmatter `description` and
	// `mode: primary`, and the markdown body becomes the agent's system prompt.
	// Primary is what puts the agent in the selector the human cycles with Tab;
	// a subagent is only reachable by the model. The scanner also accepts the
	// singular `agent/`; the documented directory is the plural one. Measured on
	// OpenCode 2.1.3: both `.opencode/agents/` and `.opencode/mode/` are listed
	// as primary, and a directory without `.git` lists nothing at all.
	{
		Agent:  "opencode",
		Dir:    ".opencode/agents",
		Switch: "se elige con Tab",
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
		Switch: "se elige con /config (estilo de salida)",
	},
}

// wrappers renders the frontmatter one client needs around the shared mirror.
// It is keyed by agent name, so a row without a wrapper cannot exist: the two
// tables are built together and a test fails on a client that has no wrapper.
var wrappers = map[string]string{
	"opencode": "---\ndescription: \"odachi: las reglas del proyecto y la voz del perfil\"\nmode: primary\n---\n",
	"claude":   "---\nname: " + AgentName + "\ndescription: \"odachi: las reglas del proyecto y la voz del perfil\"\nkeep-coding-instructions: true\n---\n",
}

// File is one native agent file: where it goes relative to the project root, and
// the bytes that go there.
type File struct {
	Path    string
	Content string
}

// Names returns the two voices, in the order they are compared and reported.
func Names() []string { return []string{Novato, Senior} }

// Clients returns the clients that have a native mechanism, safe to keep: a
// caller that reorders or trims the result cannot change what this package
// resolves.
func Clients() []Client { return append([]Client(nil), clients...) }

// Files returns the native agent file every client gets, given the block without
// its markers (rules.Body). One file per client: it mirrors the block, so the
// voice the human selected and the rules the agent obeys cannot drift.
func Files(block string) []File {
	out := make([]File, 0, len(clients))
	for _, row := range clients {
		out = append(out, fileFor(row, block))
	}
	return out
}

// FilesFor returns the one file a named client gets, and false for a client
// without a native mechanism. `init` and `level` report one destination at a
// time, and asking by name is clearer than relying on the table's order.
func FilesFor(client, block string) (File, bool) {
	key := strings.ToLower(strings.TrimSpace(client))
	for _, row := range clients {
		if row.Agent == key {
			return fileFor(row, block), true
		}
	}
	return File{}, false
}

// fileFor composes the one mirror for one client: its wrapper, the generated
// note, and the block the agent reads.
func fileFor(row Client, block string) File {
	return File{
		Path:    row.Dir + "/" + FileName,
		Content: wrappers[row.Agent] + generated + "\n" + voiceSeal + "\n\n" + block,
	}
}

// VoiceName resolves the voice a profile gets: novato and senior are their own,
// ci gets senior's, and anything else is refused naming the names that exist.
func VoiceName(profileName string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(profileName))
	if key == CI {
		return Senior, nil
	}
	if _, ok := voices[key]; ok {
		return key, nil
	}
	return "", fmt.Errorf("voz desconocida: %q; odachi escribe la de %s (%s usa la de %s)",
		profileName, strings.Join(Names(), " o "), CI, Senior)
}

// Voice returns the text of the voice one profile gets, ending in exactly one
// newline. It is the body of the block's last section, without the heading: the
// renderer owns the heading, so there is one copy of it.
func Voice(profileName string) (string, error) {
	name, err := VoiceName(profileName)
	if err != nil {
		return "", err
	}
	return voices[name], nil
}

// Deprecated is what v0.10.0 wrote: two persona files per client, one per
// profile. They still work as a voice, but they no longer mirror the block, so
// odachi reports them as leftovers and never deletes them: a binary that removes
// files it no longer writes is a binary nobody trusts with a project.
func Deprecated() []string {
	return []string{
		".opencode/agents/odachi-novato.md",
		".opencode/agents/odachi-senior.md",
		".claude/output-styles/odachi-novato.md",
		".claude/output-styles/odachi-senior.md",
	}
}

// DeprecatedDirs is the same leftovers as the directories a human reads in one
// line. It is derived from Deprecated, not written twice: a second list is a
// second thing to drift.
func DeprecatedDirs() []string {
	out := []string{}
	for _, file := range Deprecated() {
		dir := file
		if i := strings.LastIndexByte(file, '/'); i > 0 {
			dir = file[:i]
		}
		if !slices.Contains(out, dir) {
			out = append(out, dir)
		}
	}
	return out
}
