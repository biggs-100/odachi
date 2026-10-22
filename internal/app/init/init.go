// Package init implements `odachi init`: the rules block in AGENTS.md —with the
// voice of the profile at hand as its last section—, the marked block in
// .gitignore that keeps the generated evidence out of the tree, the pre-commit
// hook that refuses a commit without a proof for the current tree, and the skills
// the agent reads — the curriculum in the project and the bootstrap skill in the
// human's home, so an agent can learn how to start a project in a folder where
// nothing exists yet.
//
// The voice is a section of the block, and no longer a file per client. One
// channel, every client that reads AGENTS.md, and no dependency on a selector
// that does not list what odachi writes.
//
// Two axes, and this command writes both without joining them: the profile (in
// .odachi/profile.toml, owned by `odachi start`) decides the ceremony, and the
// block's last section decides the voice. init reads the profile only to know
// which voice to write, and never rewrites it.
//
// Every artifact is idempotent. A file that is already right is not rewritten, a
// file that exists is backed up before the rewrite (except .gitignore, whose
// block is additive and preserves every other byte), and a hook someone else
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

// The artifacts this command owns.
const (
	gitDir   = ".git"
	hookPath = ".git/hooks/pre-commit"
)

// hookMarker names the hook as odachi's. A hook without it is somebody else's
// work and is never replaced.
const hookMarker = "# odachi: no se commitea sin prueba de este árbol"

// hookFor is the hook itself. It calls this binary by absolute path when the
// binary is not on PATH, because a downloaded binary is the normal case for a
// beginner and a hook that says "command not found" blocks every commit without
// explaining anything. If that path later stops existing it falls back to
// `odachi` from PATH, and the last line stays the rule: no commit without a
// green check.
func hookFor(exe string) string {
	if exe == "" {
		exe = "odachi"
	}
	return "#!/bin/sh\n" + hookMarker +
		"\nODACHI='" + exe + "'" +
		"\n[ -x \"$ODACHI\" ] || ODACHI=odachi" +
		"\n\"$ODACHI\" check --quiet || exit 1\n"
}

// Options is one `odachi init` invocation.
type Options struct {
	Dir string
	// Exe is how the hook reaches this binary. Empty means `odachi` from PATH.
	Exe string
	// PathMissing says a shell cannot run `odachi` by name. The rules block this
	// command writes tells the agent to run exactly that command, so an
	// unreachable name makes every rule unexecutable while the hook still works
	// from its absolute path. init says so instead of reporting a clean install.
	PathMissing bool
	// Profiles is the profile at hand, read and not inferred. nil means the
	// tool's default audience: novato.
	Profiles Profiles
}

// Profiles is the profile port, one question wide: which profile did the human
// write down, if any? Written returns the profile named in .odachi/profile.toml
// and an empty name when there is no file. It is a port and not a read of the
// TOML here because the domain owns the names and the adapter owns the format,
// and because init must not infer a profile: inference is a guess about a person
// and the voice is a decision.
type Profiles interface {
	Written(dir string) (name string, err error)
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
// other verbs; Next names the command that finishes a blocked run, and SkillDests
// is what the generated half wrote, so `--json` can carry the destinations too.
type Result struct {
	Code       int
	Message    string
	Next       string
	Lines      []string
	Files      []string
	SkillDests []SkillDest
}

// Run performs one `odachi init`.
func Run(opts Options, files Files, skills Skills) Result {
	r := &runner{dir: opts.Dir, files: files, skills: skills, profiles: opts.Profiles}
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
	if r.voiceNote != "" {
		r.lines = append(r.lines, r.voiceNote)
	}
	if problem := r.writeIgnore(); problem != nil {
		// The rules block may already be written: report it, so the human knows
		// exactly how much of the job is done.
		problem.Files = append([]string{}, r.changed...)
		return *problem
	}
	if problem := r.installHook(opts.Exe); problem != nil {
		// The rules block may already be written: report it, so the human knows
		// exactly how much of the job is done.
		problem.Files = append([]string{}, r.changed...)
		return *problem
	}

	r.installSkills()

	// Nothing here is broken: an earlier version wrote one persona file per
	// client, they are still on disk, and the human is the one who decides
	// whether they go. One line says what they are, not what to do with them.
	r.reportLeftovers()

	if opts.PathMissing {
		r.lines = append(r.lines, pathWarning(opts.Exe)...)
	}

	return r.result()
}

// pathWarning is the half-broken install: the hook works from the absolute path,
// the rules do not work at all. It names the folder to add, because "ponelo en el
// PATH" is the kind of advice that leaves a beginner exactly where they were.
func pathWarning(exe string) []string {
	lines := []string{
		"odachi no está en tu PATH: el bloque de AGENTS.md le pide al agente que corra `odachi`, y así no va a poder",
	}
	if dir := filepath.Dir(exe); exe != "" && dir != "." && dir != string(filepath.Separator) {
		return append(lines, "el hook sí funciona porque usa la ruta absoluta; agregá esa carpeta al PATH (en sh: export PATH=\""+dir+":$PATH\")")
	}
	return append(lines, "el hook sí funciona porque usa la ruta absoluta; agregá la carpeta donde instalaste odachi al PATH y volvé a correr odachi init")
}

// runner holds what one run touched and what it reports.
type runner struct {
	dir      string
	files    Files
	skills   Skills
	profiles Profiles
	changed  []string
	lines    []string
	// voiceNote is the sentence for a profile init could not read or name: the
	// block still gets the default voice, and the human learns why.
	voiceNote string

	// The skills half. skillFiles are the display paths this run wrote, kept
	// apart from changed because the summary names the project artifacts and
	// not forty-nine files.
	skillDests   []SkillDest
	skillLabels  []string
	skillFiles   []string
	skillBlocked string
	skillNext    string
	skillFatal   bool
}

// installHook installs the pre-commit hook, keeps odachi's own hook up to date,
// and refuses to replace a hook someone else wrote.
func (r *runner) installHook(exe string) *Result {
	path := filepath.Join(r.dir, hookPath)
	hook := hookFor(exe)

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
			problem := foreignHook(exe)
			return &problem
		}
		if string(data) == hook {
			r.lines = append(r.lines, hookPath+": sin cambios")
			return nil
		}

		if _, err := r.files.Backup(path); err != nil {
			problem := blocked("no pude respaldar " + hookPath + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi init")
			return &problem
		}
	}

	if err := r.files.Write(path, []byte(hook)); err != nil {
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
func foreignHook(exe string) Result {
	message := "ya existe " + hookPath + " y no es de odachi: no lo pisé; pegá estas líneas dentro de tu hook y volvé a correr odachi init"
	lines := []string{message, ""}
	lines = append(lines, strings.Split(strings.TrimRight(hookFor(exe), "\n"), "\n")...)
	lines = append(lines, "", "siguiente: odachi init")
	return Result{
		Code:    CodeBlocked,
		Message: message,
		Next:    "odachi init",
		Lines:   lines,
	}
}

// result reports what changed. A rerun that changed nothing says so instead of
// claiming work it did not do, skills included.
func (r *runner) result() Result {
	files := append(append([]string{}, r.changed...), r.skillFiles...)

	if len(r.changed) == 0 && len(r.skillFiles) == 0 {
		message := "sin cambios: " + agentsFile + ", " + ignoreFile + " y el hook " + hookPath + " ya estaban al día"
		if len(r.skillLabels) > 0 {
			message += "; las skills de " + humanList(r.skillLabels) + " también"
		}
		return Result{
			Code:       r.code(),
			Message:    r.withBlocked(message),
			Next:       r.skillNext,
			Lines:      r.lines,
			SkillDests: r.skillDests,
		}
	}

	// What this run finished, in the two kinds the job has: it can be none of
	// the project artifacts and still be real work (a skill that arrived).
	written := make([]string, 0, 2)
	if len(r.changed) > 0 {
		written = append(written, humanList(r.changed))
	}
	if len(r.skillFiles) > 0 {
		written = append(written, "las skills ("+humanList(r.skillLabels)+")")
	}
	message := "listo: " + humanList(written) + " quedaron al día"

	lines := append([]string{}, r.lines...)
	// The human does not commit: the agent does. Saying "commiteá" here assumed the
	// one thing this tool exists to remove from their plate.
	lines = append(lines, "siguiente: revisá "+agentsFile+"; cuando odachi check --quiet esté en verde, el agente guarda el avance")

	return Result{
		Code:       r.code(),
		Message:    r.withBlocked(message),
		Next:       r.skillNext,
		Lines:      lines,
		Files:      files,
		SkillDests: r.skillDests,
	}
}

// code is the exit code of a finished run. A block inside the project is a
// blocked run: the artifacts are written, but the agent would work without the
// skills it reads. A block in the home is not, because the project is done.
func (r *runner) code() int {
	if r.skillFatal {
		return CodeBlocked
	}
	return CodeOK
}

// withBlocked adds the pending half to the one-line message, so a caller that
// reads only `message` still learns the run left something behind.
func (r *runner) withBlocked(message string) string {
	if r.skillBlocked == "" {
		return message
	}
	return message + "; " + r.skillBlocked
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
