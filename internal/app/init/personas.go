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
	// Written counts the files this run wrote; Kept counts the ones already
	// right.
	Written int
	Kept    int
	// Files is every agent path this run owns in that directory: the primary
	// mirror and the two subagents.
	Files []string
}

// The two scopes a native agent has.
const (
	ScopeProject = "project"
	ScopeHome    = "home"
)

// writePersonas writes the native agent every client that has a selector gets:
// the primary mirror of the block, plus the bounded worker and the read-only
// verifier the parent invokes as subagents. The primary mirrors the block; the
// subagents carry the handoff contract.
//
// Generated, therefore idempotent: an identical file is not rewritten and a
// differing one is backed up before the rewrite, because the human reads that the
// file is generated on its first line and a hand edit is lost work either way.
func (r *runner) writePersonas() *Result {
	block, err := rules.BodyFor(r.voiceName, r.skillsScope)
	if err != nil {
		return personaBlocked("could not build the native agent", err)
	}

	for _, client := range r.clients {
		primary, hasPrimary := persona.FilesFor(client.Agent, block)
		subs := persona.SubagentFilesFor(client.Agent)

		// A client can read the primary and its subagents from the same folder
		// (OpenCode) or from two (Claude Code). Grouping by folder keeps the
		// report honest: one line per destination, never a file named under the
		// wrong folder.
		if hasPrimary && client.Dir == client.SubagentDir {
			files := append([]persona.File{primary}, subs...)
			if problem := r.writePersonaGroup(client, client.Dir, files, client.Switch, len(subs) > 0); problem != nil {
				return problem
			}
			continue
		}

		if hasPrimary {
			if problem := r.writePersonaGroup(client, client.Dir, []persona.File{primary}, client.Switch, false); problem != nil {
				return problem
			}
		}
		if len(subs) > 0 {
			if problem := r.writePersonaGroup(client, client.SubagentDir, subs, "the subagents are invoked by the agent", false); problem != nil {
				return problem
			}
		}
	}
	return nil
}

// writePersonaGroup writes every generated file of one directory, reports one
// row for that directory, and one line for the human.
func (r *runner) writePersonaGroup(client persona.Client, dir string, files []persona.File, switchText string, subagents bool) *Result {
	dest := PersonaDest{Client: client.Agent, Scope: ScopeProject, Dir: dir}
	var written, kept []string

	for _, file := range files {
		path := filepath.Join(r.dir, filepath.FromSlash(file.Path))
		name := filepath.Base(path)

		exists, err := r.files.Exists(path)
		if err != nil {
			return personaBlocked("could not check "+file.Path, err)
		}
		if exists {
			current, err := r.files.Read(path)
			if err != nil {
				return personaBlocked("could not read "+file.Path, err)
			}
			if string(current) == file.Content {
				dest.Kept++
				dest.Files = append(dest.Files, file.Path)
				kept = append(kept, name)
				continue
			}
			if _, err := r.files.Backup(path); err != nil {
				return personaBlocked("could not back up "+file.Path+" before rewriting it", err)
			}
		}

		if err := r.files.Write(path, []byte(file.Content)); err != nil {
			return personaBlocked("could not write "+file.Path, err)
		}
		dest.Written++
		dest.Files = append(dest.Files, file.Path)
		r.personaFiles = append(r.personaFiles, file.Path)
		written = append(written, name)
	}

	r.personaDests = append(r.personaDests, dest)
	r.personaLabels = append(r.personaLabels, dir)
	r.lines = append(r.lines, personaLine(dir, switchText, written, kept, subagents, r.dryRun))
	return nil
}

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
