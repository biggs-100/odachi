// Package init implements `odachi init`: the rules block in AGENTS.md and the
// pre-commit hook that refuses a commit without a proof for the current tree.
//
// Both halves are idempotent. A file that is already right is not rewritten, a
// file that exists is backed up before the rewrite, and a hook someone else
// wrote is never clobbered: the human gets the three lines to paste instead.
package init

import (
	"path/filepath"
	"strings"
)

// Exit codes: the same two every verb uses when nothing needs a human answer.
const (
	// CodeOK means both halves are in place.
	CodeOK = 0
	// CodeBlocked means the job could not be finished and says why.
	CodeBlocked = 1
)

// The two artifacts this command owns.
const (
	gitDir   = ".git"
	hookPath = ".git/hooks/pre-commit"
)

// hookMarker names the hook as odachi's. A hook without it is somebody else's
// work and is never replaced.
const hookMarker = "# odachi: no se commitea sin prueba de este árbol"

// hookContent is the hook itself: exactly three lines, the same three a human
// is told to paste when the hook slot is already taken.
const hookContent = "#!/bin/sh\n" + hookMarker + "\nodachi check --quiet || exit 1\n"

// Options is one `odachi init` invocation.
type Options struct {
	Dir string
}

// Files is the disk port: enough to read, back up and replace the two files,
// and to make the hook executable.
type Files interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Backup(path string) (string, error)
	MakeExecutable(path string) error
}

// Result is what the CLI prints and exits with. It has the same shape as the
// other verbs; Next names the command that finishes a blocked run.
type Result struct {
	Code    int
	Message string
	Next    string
	Lines   []string
	Files   []string
}

// Run performs one `odachi init`.
func Run(opts Options, files Files) Result {
	r := &runner{dir: opts.Dir, files: files}
	if r.dir == "" {
		r.dir = "."
	}

	hasGit, err := r.files.Exists(filepath.Join(r.dir, gitDir))
	if err != nil {
		return blocked("no pude revisar " + gitDir + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
	}
	if !hasGit {
		// No repository means nowhere to install the hook. Writing only the
		// AGENTS.md half would report success for a job that is half done.
		return blocked("no hay " + gitDir + " en este directorio: no puedo instalar el hook " + hookPath + "; corré git init y volvé a correr odachi init")
	}

	if problem := r.writeAgents(); problem != nil {
		return *problem
	}
	if problem := r.installHook(); problem != nil {
		// The rules block may already be written: report it, so the human knows
		// exactly how much of the job is done.
		problem.Files = append([]string{}, r.changed...)
		return *problem
	}

	return r.result()
}

// runner holds what one run touched and what it reports.
type runner struct {
	dir     string
	files   Files
	changed []string
	lines   []string
}

// writeAgents writes the rules block, replacing it in place when it is already
// there and backing up the human's file before touching it.
func (r *runner) writeAgents() *Result {
	path := filepath.Join(r.dir, agentsFile)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("no pude revisar " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	current := ""
	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("no pude leer " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
		current = string(data)
	}

	updated := withBlock(current)
	if exists && updated == current {
		r.lines = append(r.lines, agentsFile+": sin cambios")
		return nil
	}

	if exists {
		backup, err := r.files.Backup(path)
		if err != nil {
			problem := blocked("no pude respaldar " + agentsFile + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
		r.lines = append(r.lines, agentsFile+": bloque actualizado (el anterior quedó en "+relative(r.dir, backup)+")")
	} else {
		r.lines = append(r.lines, agentsFile+": creado con el bloque de odachi")
	}

	if err := r.files.Write(path, []byte(updated)); err != nil {
		problem := blocked("no pude escribir " + agentsFile + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}
	r.changed = append(r.changed, agentsFile)
	return nil
}

// installHook installs the pre-commit hook, keeps odachi's own hook up to date,
// and refuses to replace a hook someone else wrote.
func (r *runner) installHook() *Result {
	path := filepath.Join(r.dir, hookPath)

	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("no pude revisar " + hookPath + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("no pude leer " + hookPath + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}

		if !strings.Contains(string(data), hookMarker) {
			problem := foreignHook()
			return &problem
		}
		if string(data) == hookContent {
			r.lines = append(r.lines, hookPath+": sin cambios")
			return nil
		}

		if _, err := r.files.Backup(path); err != nil {
			problem := blocked("no pude respaldar " + hookPath + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
	}

	if err := r.files.Write(path, []byte(hookContent)); err != nil {
		problem := blocked("no pude escribir " + hookPath + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}
	if err := r.files.MakeExecutable(path); err != nil {
		problem := blocked("no pude hacer ejecutable " + hookPath + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
		return &problem
	}

	r.changed = append(r.changed, hookPath)
	r.lines = append(r.lines, hookPath+": instalado")
	return nil
}

// foreignHook is the refusal: the three lines to paste, and nothing touched.
func foreignHook() Result {
	message := "ya existe " + hookPath + " y no es de odachi: no lo pisé; pegá estas tres líneas dentro de tu hook y volvé a correr odachi init"
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Next:    "odachi init",
		Lines: []string{
			message,
			"",
			"#!/bin/sh",
			hookMarker,
			"odachi check --quiet || exit 1",
			"",
			"siguiente: odachi init",
		},
	}
}

// result reports what changed. A rerun that changed nothing says so instead of
// claiming work it did not do.
func (r *runner) result() Result {
	if len(r.changed) == 0 {
		message := "sin cambios: " + agentsFile + " y el hook " + hookPath + " ya estaban al día"
		return Result{Code: CodeOK, Message: message, Lines: r.lines}
	}

	message := "listo: " + humanList(r.changed) + " quedaron al día"
	lines := append([]string{}, r.lines...)
	lines = append(lines, "siguiente: revisá "+agentsFile+" y commiteá cuando odachi check --quiet esté en verde")

	return Result{Code: CodeOK, Message: message, Lines: lines, Files: r.changed}
}

// blocked is a result the human can act on, with the command that retries it.
func blocked(message string) Result {
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Next:    "odachi init",
		Lines:   []string{message, "siguiente: odachi init"},
	}
}

// humanList joins two or three paths the way a Spanish sentence does.
func humanList(items []string) string {
	switch len(items) {
	case 0:
		return "nada"
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " y " + items[len(items)-1]
	}
}

// relative reports a path the way the human typed it: inside the project, not
// an absolute path from a temporary directory.
func relative(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
