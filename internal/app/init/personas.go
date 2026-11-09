package init

import (
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
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
	// Written counts the files this run wrote; Kept counts the ones already
	// right.
	Written int
	Kept    int
	// Files is every agent path this run owns in that directory: the primary
	// mirror and the two subagents.
	Files []string
}

// The scope a native agent has.
const ScopeHome = "home"

// installGlobalPersonas writes the door in the human's home: the agent that
// appears in every folder, not only in the project that ran init. A home that
// cannot be written is `pending:` and exit 0, because the project half is done;
// a failure inside the project is the caller's blocked result.
func (r *runner) installGlobalPersonas() {
	if r.home == nil {
		return
	}
	home, err := r.home.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		reason := "the system reported no home"
		if err != nil {
			reason = oneLine(err.Error())
		}
		r.pendGlobal("could not resolve the home to install the global agent: " + reason + "; run odachi init again")
		return
	}

	for _, client := range r.clients {
		file, ok := persona.GlobalFilesFor(client.Agent)
		if !ok {
			continue
		}

		display := "~/" + file.Path
		dest := PersonaDest{Client: client.Agent, Scope: ScopeHome, Dir: "~/" + client.HomeDir}
		path := filepath.Join(home, filepath.FromSlash(file.Path))

		exists, err := r.files.Exists(path)
		if err != nil {
			r.pendGlobal("could not check " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
			continue
		}
		if exists {
			current, err := r.files.Read(path)
			if err != nil {
				r.pendGlobal("could not read " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
				continue
			}
			if string(current) == file.Content {
				dest.Kept = 1
				dest.Files = append(dest.Files, display)
				r.personaDests = append(r.personaDests, dest)
				r.personaLabels = append(r.personaLabels, dest.Dir)
				r.lines = append(r.lines, personaLine(dest.Dir, client.Switch, nil, []string{persona.FileName}, false, r.dryRun))
				continue
			}
			if _, err := r.files.Backup(path); err != nil {
				r.pendGlobal("could not back up " + display + " before rewriting it: " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
				continue
			}
		}

		if err := r.files.Write(path, []byte(file.Content)); err != nil {
			r.pendGlobal("could not write " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
			continue
		}
		dest.Written = 1
		dest.Files = append(dest.Files, display)
		r.personaDests = append(r.personaDests, dest)
		r.personaLabels = append(r.personaLabels, dest.Dir)
		r.personaFiles = append(r.personaFiles, display)
		r.lines = append(r.lines, personaLine(dest.Dir, client.Switch, []string{persona.FileName}, nil, false, r.dryRun))
	}

	r.installGlobalSubagents(home)
}

// installGlobalSubagents writes the worker, the verifier and the reviewer in the
// human's home: the door is global, so the agents it invokes are global too, and
// a project that never writes `.opencode/` still has them.
func (r *runner) installGlobalSubagents(home string) {
	for _, client := range r.clients {
		files := persona.GlobalSubagentFilesFor(client.Agent)
		if len(files) == 0 {
			continue
		}

		dir := "~/" + client.HomeSubagentDir
		written := make([]string, 0, len(files))
		kept := make([]string, 0, len(files))
		for _, file := range files {
			display := "~/" + file.Path
			path := filepath.Join(home, filepath.FromSlash(file.Path))

			exists, err := r.files.Exists(path)
			if err != nil {
				r.pendGlobal("could not check " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
				continue
			}
			if exists {
				current, err := r.files.Read(path)
				if err != nil {
					r.pendGlobal("could not read " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
					continue
				}
				if string(current) == file.Content {
					kept = append(kept, filepath.Base(file.Path))
					continue
				}
				if _, err := r.files.Backup(path); err != nil {
					r.pendGlobal("could not back up " + display + " before rewriting it: " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
					continue
				}
			}

			if err := r.files.Write(path, []byte(file.Content)); err != nil {
				r.pendGlobal("could not write " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
				continue
			}
			written = append(written, filepath.Base(file.Path))
			r.personaFiles = append(r.personaFiles, display)
		}
		r.lines = append(r.lines, personaLine(dir, "", written, kept, true, r.dryRun))
	}
}

// pendGlobal keeps one pending line for the home half. It never changes the exit
// code: the project is ready, and the human can retry the home half alone.
func (r *runner) pendGlobal(message string) {
	if r.homePending == "" {
		r.homePending = "pending: " + message
	}
	r.lines = append(r.lines, "pending: "+message)
}

// personaBlocked is the persona half's refusal: the file, the reason the adapter
// gave, and the command that retries the run.
func personaBlocked(what string, err error) *Result {
	result := blocked(what + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
	return &result
}

// personaLine is one destination in one line: what arrived, what was already
// there, and how the human selects the agent in that client. The last part is the
// point of the whole job — a file nobody knows how to pick is a file nobody uses.
//
// dry is the rehearsal: the line names the file as one it would write, because a
// report that says "wrote" over zero bytes is the lie the flag exists to avoid.
func personaLine(dir, switchText string, written, kept []string, subagents, dry bool) string {
	parts := make([]string, 0, 2)
	if len(written) > 0 {
		verb := "wrote "
		if dry {
			verb = "would write "
		}
		parts = append(parts, verb+humanList(written))
	}
	if len(kept) > 0 {
		parts = append(parts, humanList(kept)+" unchanged")
	}
	line := dir + ": " + strings.Join(parts, ", ")
	if len(written) > 0 && switchText != "" {
		line += "; " + switchText
	}
	if subagents {
		line += "; " + humanList(persona.SubagentNames()) + " are invoked by the agent"
	}
	return line
}
