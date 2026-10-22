package init

import (
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// PersonaDest is one client's native agent directory and what this run did
// there. It mirrors SkillDest: one row a human reads in a line and a machine
// reads in --json.
type PersonaDest struct {
	// Client is the name the skills table uses for the same program.
	Client string
	// Scope is "project" or "home": the same client has a mirror inside an
	// initialized project and a door in the home for every other folder.
	Scope string
	// Dir is the directory the client reads agents from, as the human reads it
	// (`./.opencode/agents` or `~/.config/opencode/agents`).
	Dir string
	// Written counts the file this run wrote; Kept is 1 when it was already
	// right.
	Written int
	Kept    int
	// Files is every agent path this run owns in that directory.
	Files []string
}

// The two scopes a native agent has.
const (
	ScopeProject = "project"
	ScopeHome    = "home"
)

// writePersonas writes the native agent every client that has a selector gets.
// The file mirrors the block: same rules, same voice, so a human who picks it
// with Tab reads exactly what the agent reads from AGENTS.md.
//
// Generated, therefore idempotent: an identical file is not rewritten and a
// differing one is backed up before the rewrite, because the human reads that the
// file is generated on its first line and a hand edit is lost work either way.
func (r *runner) writePersonas() *Result {
	block, err := rules.Body(r.voiceName)
	if err != nil {
		return personaBlocked("no pude armar el agente nativo", err)
	}

	for _, client := range persona.Clients() {
		file, ok := persona.FilesFor(client.Agent, block)
		if !ok {
			// Unreachable: the client came from the same table. Skipping instead
			// of writing a file to a directory nobody verified.
			continue
		}

		dest := PersonaDest{Client: client.Agent, Scope: ScopeProject, Dir: client.Dir}
		path := filepath.Join(r.dir, filepath.FromSlash(file.Path))
		name := filepath.Base(path)

		exists, err := r.files.Exists(path)
		if err != nil {
			return personaBlocked("no pude revisar "+file.Path, err)
		}
		if exists {
			current, err := r.files.Read(path)
			if err != nil {
				return personaBlocked("no pude leer "+file.Path, err)
			}
			if string(current) == file.Content {
				dest.Kept = 1
				dest.Files = append(dest.Files, file.Path)
				r.personaDests = append(r.personaDests, dest)
				r.personaLabels = append(r.personaLabels, client.Dir)
				r.lines = append(r.lines, personaLine(client.Dir, client.Switch, nil, []string{name}))
				continue
			}
			if _, err := r.files.Backup(path); err != nil {
				return personaBlocked("no pude respaldar "+file.Path+" antes de reescribirlo", err)
			}
		}

		if err := r.files.Write(path, []byte(file.Content)); err != nil {
			return personaBlocked("no pude escribir "+file.Path, err)
		}
		dest.Written = 1
		dest.Files = append(dest.Files, file.Path)
		r.personaDests = append(r.personaDests, dest)
		r.personaLabels = append(r.personaLabels, client.Dir)
		r.personaFiles = append(r.personaFiles, file.Path)
		r.lines = append(r.lines, personaLine(client.Dir, client.Switch, []string{name}, nil))
	}
	return nil
}

// installGlobalPersonas writes the door in the human's home: the agent that
// appears in every folder, not only in the project that ran init. A home that
// cannot be written is `pendiente:` and exit 0, because the project half is done;
// a failure inside the project is the caller's blocked result.
func (r *runner) installGlobalPersonas() {
	if r.home == nil {
		return
	}
	home, err := r.home.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		reason := "el sistema no reportó ningún home"
		if err != nil {
			reason = oneLine(err.Error())
		}
		r.pendGlobal("no pude resolver el home para instalar el agente global: " + reason + "; volvé a correr odachi init")
		return
	}

	for _, client := range persona.Clients() {
		file, ok := persona.GlobalFilesFor(client.Agent)
		if !ok {
			continue
		}

		display := "~/" + file.Path
		dest := PersonaDest{Client: client.Agent, Scope: ScopeHome, Dir: "~/" + client.HomeDir}
		path := filepath.Join(home, filepath.FromSlash(file.Path))

		exists, err := r.files.Exists(path)
		if err != nil {
			r.pendGlobal("no pude revisar " + display + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			continue
		}
		if exists {
			current, err := r.files.Read(path)
			if err != nil {
				r.pendGlobal("no pude leer " + display + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
				continue
			}
			if string(current) == file.Content {
				dest.Kept = 1
				dest.Files = append(dest.Files, display)
				r.personaDests = append(r.personaDests, dest)
				r.personaLabels = append(r.personaLabels, dest.Dir)
				r.lines = append(r.lines, personaLine(dest.Dir, client.Switch, nil, []string{persona.FileName}))
				continue
			}
			if _, err := r.files.Backup(path); err != nil {
				r.pendGlobal("no pude respaldar " + display + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
				continue
			}
		}

		if err := r.files.Write(path, []byte(file.Content)); err != nil {
			r.pendGlobal("no pude escribir " + display + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			continue
		}
		dest.Written = 1
		dest.Files = append(dest.Files, display)
		r.personaDests = append(r.personaDests, dest)
		r.personaLabels = append(r.personaLabels, dest.Dir)
		r.personaFiles = append(r.personaFiles, display)
		r.lines = append(r.lines, personaLine(dest.Dir, client.Switch, []string{persona.FileName}, nil))
	}
}

// pendGlobal keeps one pending line for the home half. It never changes the exit
// code: the project is ready, and the human can retry the home half alone.
func (r *runner) pendGlobal(message string) {
	if r.homePending == "" {
		r.homePending = "pendiente: " + message
	}
	r.lines = append(r.lines, "pendiente: "+message)
}

// personaBlocked is the persona half's refusal: the file, the reason the adapter
// gave, and the command that retries the run.
func personaBlocked(what string, err error) *Result {
	result := blocked(what + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
	return &result
}

// personaLine is one destination in one line: what arrived, what was already
// there, and how the human selects the agent in that client. The last part is the
// point of the whole job — a file nobody knows how to pick is a file nobody uses.
func personaLine(dir, switchText string, written, kept []string) string {
	parts := make([]string, 0, 2)
	if len(written) > 0 {
		parts = append(parts, "escribí "+humanList(written))
	}
	if len(kept) > 0 {
		parts = append(parts, humanList(kept)+" sin cambios")
	}
	line := dir + ": " + strings.Join(parts, ", ")
	if len(written) > 0 {
		line += "; " + switchText
	}
	return line
}
