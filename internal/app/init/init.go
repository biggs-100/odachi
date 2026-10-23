// Package init implements `odachi init`: the rules block in AGENTS.md —with the
// voice of the profile at hand as its last section—, the native agent that
// mirrors that block in the project, the door in the human's home that makes
// odachi available in any folder —opt-in, with --global—, the marked block in
// .gitignore that keeps the
// generated evidence out of the tree, the pre-commit hook that refuses a commit
// without a proof for the current tree, and the skills the agent reads — the
// curriculum in the project and the bootstrap skill in the human's home, so an
// agent can learn how to start a project in a folder where nothing exists yet.
//
// One voice, one owner. The profile (in .odachi/profile.toml, owned by `odachi
// start`) decides the ceremony and the voice; the block publishes it where every
// client reads it, the project agent mirrors the same block, and the home agent
// is the door that reads the project's block when there is one. init reads the
// profile only to know which voice to write, and never rewrites it.
//
// Every artifact is idempotent. A file that is already right is not rewritten, a
// file that exists is backed up before the rewrite (except .gitignore, whose
// block is additive and preserves every other byte), and a hook someone else
// wrote is never clobbered: the human gets the three lines to paste instead.
package init

import (
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
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
	// Agents is the client selection this run was asked for. Empty means "the
	// stored one, or both": a plain init after `--agents opencode` must not
	// bring the other client back.
	Agents string
	// AgentSelection is the store of that choice. nil behaves like a project
	// without a file: both clients, and nothing written.
	AgentSelection AgentSelection
	// Home resolves the human's home, where the global agent lives. nil means
	// this run does not install the global half (tests that only care about the
	// project); the binary always passes an adapter.
	Home Home
	// Global says this run installs the door of the home. Es opt-in: un init sin
	// el flag no escribe el home, porque un agente primary en toda carpeta es una
	// decisión del humano y no un efecto de inicializar un proyecto.
	Global bool
}

// AgentSelection is the selection port: the choice a project recorded, when
// there is one, and the canonical bytes that write it. It is a port and not a
// read of the TOML here because the domain owns the names and the adapter owns
// the format.
type AgentSelection interface {
	// Written returns the stored selection, or empty when there is no file.
	Written(dir string) (string, error)
	// Render turns a selection into the bytes of the store.
	Render(selection string) ([]byte, error)
}

// Home is the port that answers where the human lives. The global agent is the
// door of every folder, so `init` writes it once per machine; the app never
// touches os.
type Home interface {
	Home() (string, error)
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
// and PersonaDests are what the two generated halves wrote, so `--json` can carry
// the destinations too.
type Result struct {
	Code         int
	Message      string
	Next         string
	Lines        []string
	Files        []string
	SkillDests   []SkillDest
	PersonaDests []PersonaDest
}

// Run performs one `odachi init`.
func Run(opts Options, files Files, skills Skills) Result {
	r := &runner{dir: opts.Dir, files: files, skills: skills, profiles: opts.Profiles, agentSelection: opts.AgentSelection, home: opts.Home}
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

	selection, clients, problem := r.resolveAgents(opts.Agents)
	if problem != nil {
		return *problem
	}
	r.selection = selection
	r.clients = clients

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
	if problem := r.writeAgentStore(opts.Agents != ""); problem != nil {
		// The hook and the rules block are already in place: report how much of
		// the job is done instead of pretending nothing happened.
		problem.Files = append([]string{}, r.changed...)
		return *problem
	}
	if problem := r.writePersonas(); problem != nil {
		// The hook and the rules block are already in place: report how much of
		// the job is done instead of pretending nothing happened.
		problem.Files = append(append([]string{}, r.changed...), r.personaFiles...)
		problem.PersonaDests = r.personaDests
		return *problem
	}
	// The project has its mirror; the home gets the door that appears in every
	// other folder, and only when the human asked for it with --global. A home
	// that cannot be written is pending, never a failed project.
	if opts.Global {
		r.installGlobalPersonas()
	} else {
		r.lines = append(r.lines, "puerta global: no la instalé (es opt-in); odachi init --global la pone en el home y odachi aparece en cualquier carpeta")
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

// resolveAgents knows which clients receive the native agent. The flag wins; a
// run without it keeps the stored choice, and a project with no file keeps both,
// the behavior every existing project already had.
func (r *runner) resolveAgents(requested string) (string, []persona.Client, *Result) {
	if r.agentSelection != nil {
		stored, err := r.agentSelection.Written(r.dir)
		if err != nil {
			problem := blocked("no pude leer " + persona.SelectionPath + ": " + oneLine(err.Error()) +
				"; corregí el archivo o volvé a correr odachi init --agents both")
			return "", nil, &problem
		}
		r.storedSelection = stored
	}

	value := strings.TrimSpace(requested)
	if value == "" {
		value = r.storedSelection
	}
	selection, clients, err := persona.ResolveSelection(value)
	if err != nil {
		problem := blocked("la elección de clientes no es válida: " + oneLine(err.Error()) +
			"; volvé a correr odachi init --agents both")
		return "", nil, &problem
	}
	return selection, clients, nil
}

// writeAgentStore persists the choice only when someone made it: an explicit
// flag, or a store that already exists. A project that never chose keeps no file
// and both clients, so the default install gains no new artifact.
func (r *runner) writeAgentStore(explicit bool) *Result {
	if !explicit && r.storedSelection == "" {
		return nil
	}
	if r.storedSelection == r.selection {
		r.lines = append(r.lines, persona.SelectionPath+": "+r.selection+", sin cambios")
		return nil
	}
	if r.agentSelection == nil {
		problem := blocked("no puedo guardar la elección de clientes: falta el adaptador del store; volvé a correr odachi init --agents both")
		return &problem
	}

	data, err := r.agentSelection.Render(r.selection)
	if err != nil {
		problem := blocked("no pude armar " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; volvé a correr odachi init --agents " + r.selection)
		return &problem
	}

	path := filepath.Join(r.dir, filepath.FromSlash(persona.SelectionPath))
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("no pude revisar " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; revisá los permisos y volvé a correr odachi init --agents " + r.selection)
		return &problem
	}
	if exists {
		if _, err := r.files.Backup(path); err != nil {
			problem := blocked("no pude respaldar " + persona.SelectionPath + " antes de reescribirlo: " + oneLine(err.Error()) +
				"; revisá los permisos y volvé a correr odachi init --agents " + r.selection)
			return &problem
		}
	}
	if err := r.files.Write(path, data); err != nil {
		problem := blocked("no pude escribir " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; revisá los permisos y volvé a correr odachi init --agents " + r.selection)
		return &problem
	}

	r.changed = append(r.changed, persona.SelectionPath)
	r.lines = append(r.lines, persona.SelectionPath+": guardé la elección "+r.selection+"; los archivos de los otros clientes no se tocan")
	return nil
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
	dir            string
	files          Files
	skills         Skills
	profiles       Profiles
	agentSelection AgentSelection
	home           Home
	changed        []string
	lines          []string
	// storedSelection is what the project had before this run: empty means no
	// file and both clients. selection is what this run resolved, and clients is
	// the filtered table every persona write uses.
	storedSelection string
	selection       string
	clients         []persona.Client
	// voiceNote is the sentence for a profile init could not read or name: the
	// block still gets the default voice, and the human learns why.
	voiceNote string
	// voiceName is the voice the block and the native agent were written with,
	// resolved once so the two surfaces cannot disagree.
	voiceName string

	// The skills half. skillFiles are the display paths this run wrote, kept
	// apart from changed because the summary names the project artifacts and
	// not forty-nine files.
	skillDests   []SkillDest
	skillLabels  []string
	skillFiles   []string
	skillBlocked string
	skillNext    string
	skillFatal   bool

	// The persona half: the same split between what a human reads (dirs) and
	// what --json carries (files), for the same reason.
	personaDests  []PersonaDest
	personaLabels []string
	personaFiles  []string
	// homePending is the sentence for a home that could not take the global
	// agent. It is said in the message and never changes the exit code.
	homePending string
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
	files := append(append(append([]string{}, r.changed...), r.skillFiles...), r.personaFiles...)

	if len(r.changed) == 0 && len(r.skillFiles) == 0 && len(r.personaFiles) == 0 {
		message := "sin cambios: " + agentsFile + ", " + ignoreFile + ", el hook " + hookPath + " y el agente nativo ya estaban al día"
		if len(r.skillLabels) > 0 {
			message += "; las skills de " + humanList(r.skillLabels) + " también"
		}
		return Result{
			Code:         r.code(),
			Message:      r.withBlocked(message),
			Next:         r.skillNext,
			Lines:        r.lines,
			SkillDests:   r.skillDests,
			PersonaDests: r.personaDests,
		}
	}

	// What this run finished, in the three kinds the job has: it can be none of
	// the project artifacts and still be real work (a persona that changed, a
	// skill that arrived).
	written := make([]string, 0, 3)
	if len(r.changed) > 0 {
		written = append(written, humanList(r.changed))
	}
	if len(r.skillFiles) > 0 {
		written = append(written, "las skills ("+humanList(r.skillLabels)+")")
	}
	if len(r.personaFiles) > 0 {
		written = append(written, "el agente nativo ("+humanList(r.personaLabels)+")")
	}
	message := "listo: " + humanList(written) + " quedaron al día"

	lines := append([]string{}, r.lines...)
	// The human does not commit: the agent does. Saying "commiteá" here assumed the
	// one thing this tool exists to remove from their plate.
	lines = append(lines, "siguiente: revisá "+agentsFile+"; cuando odachi check --quiet esté en verde, el agente guarda el avance")

	return Result{
		Code:         r.code(),
		Message:      r.withBlocked(message),
		Next:         r.skillNext,
		Lines:        lines,
		Files:        files,
		SkillDests:   r.skillDests,
		PersonaDests: r.personaDests,
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
	if r.skillBlocked != "" {
		message += "; " + r.skillBlocked
	}
	if r.homePending != "" {
		message += "; " + r.homePending
	}
	return message
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
