// Package doctor implements `odachi doctor`: eleven read-only health checks,
// each with one Spanish line and, when something is wrong, the phrase that
// repairs it.
//
// It answers the question a person has after installing the tool — "¿está todo
// bien instalado?" — without them reading the repository. The most valuable of
// the eleven is whether `odachi` resolves from a shell: in a real OpenCode
// session where it did not, the agent could not execute a single rule and spent
// its budget searching for the binary instead.
//
// A folder that is not a project yet gets one situation and one step instead of
// the eleven checks: `odachi init` leaves the rules the agent reads, and `odachi
// start` is the interview. Eleven checks name three different commands, and none
// of them is the sentence a beginner needs.
//
// The package is pure orchestration over three ports: Files (the disk), Lookup
// (whether a shell can resolve a name) and Profiles (the human's profile). The
// Files port has no write method, so "doctor never writes" is a property of the
// type and not a promise in a comment.
package doctor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/brief"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/skill"
)

// Exit codes, the same shape as the other verbs.
const (
	// CodeOK means no check found a problem. Warnings do not fail the run.
	CodeOK = 0
	// CodeProblems means at least one check found something to fix.
	CodeProblems = 1
)

// The paths and layouts the checks read, relative to the project root. They are
// the same ones `init`, `start`, `check`, `review` and `skills --write` write;
// doctor owns none of them and writes none of them.
const (
	gitDir      = ".git"
	hookRel     = ".git/hooks/pre-commit"
	agentsFile  = "AGENTS.md"
	odachiDir   = ".odachi"
	profileFile = odachiDir + "/profile.toml"
	briefFile   = "brief.md"
	skillsDir   = "skills"
	skillFile   = "SKILL.md"
	registry    = skill.RegistryFile
	// installedSkillsDir is where `odachi init` puts the skills every client
	// reads. The row that installs there belongs to the agent table — it is the
	// `standard` row — and a test pins that this string still matches it, so the
	// count doctor reports is the count init wrote.
	installedSkillsDir = ".agents/skills"
	reviewStore        = odachiDir + "/reviews"
	proofDir           = "proof"
	ignoreFile         = ".gitignore"
)

// Options is one `odachi doctor` invocation.
type Options struct {
	Dir string
	// Version is what this binary reports, so the first line answers "¿qué
	// instalé?" before anything else.
	Version string
	// Exe is the running binary. It is only used to name the folder a human has
	// to add to the PATH, and the message stays honest when it is empty.
	Exe string
}

// Files is the disk port. It can only read: doctor has no way to write, which is
// what "estrictamente de solo lectura" means structurally.
type Files interface {
	Exists(path string) (bool, error)
	IsDir(path string) (bool, error)
	Read(path string) ([]byte, error)
	List(dir string) ([]string, error)
}

// Lookup is the shell port: can a shell resolve this name, and where? It is a
// port and not a call to os/exec so a test can put `odachi` off the PATH without
// touching the machine it runs on.
type Lookup interface {
	LookPath(name string) (string, error)
}

// Profiles is the profile port. It answers the same question the CLI answers
// before every verb that reads a profile: which profile is in force, and is it
// valid? inferred says there is no file and the profile comes from what the tool
// can see, which is a valid answer and not a problem.
type Profiles interface {
	Resolve(dir string) (name string, inferred bool, err error)
}

// Result is what the CLI prints and exits with. Checks carries the eleven lines
// in order so --json prints the same content as the human block, without printing
// twice. A folder that is not a project yet has no checks: Lines and Message are
// the situation and the one step, and they are what --json carries.
type Result struct {
	Code     int
	Message  string
	Lines    []string
	Checks   []domaindoctor.Check
	Problems int
	Warnings int
}

// Run performs one `odachi doctor`. Every check is independent: it reads what it
// needs and cannot change what the next one sees.
//
// The pre-flight comes first. In a folder where odachi was never started, the
// checks name three different commands in five problem lines, and none of them
// says the only thing that person needs: this folder is not a project yet, start
// here. One situation and one step is what a beginner can act on.
func Run(opts Options, files Files, lookup Lookup, profiles Profiles) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	r := &runner{opts: opts, files: files, lookup: lookup, profiles: profiles}

	if !r.isProject() {
		message, lines := domaindoctor.NotAProject()
		return Result{
			Code:     CodeProblems,
			Message:  message,
			Lines:    lines,
			Problems: 1,
		}
	}

	checks := []domaindoctor.Check{
		r.version(),
		r.onPath(),
		r.repository(),
		r.agents(),
		r.hook(),
		r.hookExe(),
		r.profile(),
		r.brief(),
		r.skills(),
		r.ignored(),
		r.voice(),
	}

	lines := make([]string, 0, len(checks)+1)
	for _, check := range checks {
		lines = append(lines, domaindoctor.Line(check))
	}
	summary := domaindoctor.Summary(checks)
	lines = append(lines, summary)

	problems, warnings := domaindoctor.Tally(checks)
	return Result{
		Code:     domaindoctor.Code(checks),
		Message:  summary,
		Lines:    lines,
		Checks:   checks,
		Problems: problems,
		Warnings: warnings,
	}
}

// runner holds the facts one run needs.
type runner struct {
	opts     Options
	files    Files
	lookup   Lookup
	profiles Profiles
}

// isProject answers the question the checks take for granted: is there anything
// here at all? Git, the rules, a brief or the odachi folder each mean somebody
// already started, and a project with only one of them still gets every check
// with its per-check fix.
//
// A read error is not a verdict. Saying "esta carpeta no es un proyecto" about a
// folder odachi could not read would be a lie, so it falls through to the checks
// and the unreadable path becomes its own line.
func (r *runner) isProject() bool {
	// .git is a directory in a normal clone and a file in a linked worktree; both
	// mean git is already in charge here.
	for _, rel := range []string{gitDir, briefFile, agentsFile} {
		found, err := r.exists(rel)
		if err != nil {
			return true
		}
		if found {
			return true
		}
	}

	found, err := r.isDir(odachiDir)
	if err != nil {
		return true
	}
	return found
}

// version says what was installed. It is the one check with no problem: a
// version is a fact, not a state to fix.
func (r *runner) version() domaindoctor.Check {
	version := strings.TrimSpace(r.opts.Version)
	if version == "" {
		version = "dev"
	}
	return domaindoctor.OK(domaindoctor.CheckVersion, "odachi "+version)
}

// onPath is the check that matters most. Every rule in AGENTS.md tells the agent
// to run `odachi`; when a shell cannot resolve that name, the agent cannot obey a
// single one of them.
func (r *runner) onPath() domaindoctor.Check {
	path, err := r.lookup.LookPath("odachi")
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckOnPath,
			"no está en el PATH: el agente no puede ejecutar las reglas",
			domaindoctor.PathFix(r.opts.Exe))
	}
	return domaindoctor.OK(domaindoctor.CheckOnPath, "sí: "+filepath.ToSlash(path))
}

// repository: without .git there is nowhere to install the hook.
func (r *runner) repository() domaindoctor.Check {
	found, err := r.isDir(gitDir)
	if err != nil {
		return r.unreadable(domaindoctor.CheckGit, gitDir, err)
	}
	if !found {
		// In a linked worktree .git is a file pointing at the main repository.
		// Saying "there is no .git, run git init" there is a lie and a dead end:
		// git init is already done, in the main repository.
		if exists, err := r.files.Exists(filepath.Join(r.opts.Dir, filepath.FromSlash(gitDir))); err == nil && exists {
			return domaindoctor.OK(domaindoctor.CheckGit, "sí: worktree (el .git real está en el repo principal)")
		}
		return domaindoctor.Broken(domaindoctor.CheckGit, "no hay "+gitDir+": no hay hook posible", "corré git init")
	}
	return domaindoctor.OK(domaindoctor.CheckGit, "sí")
}

// inWorktree reports whether this project is a linked worktree, where the hook
// lives in the main repository and init cannot install it here.
func (r *runner) inWorktree() bool {
	isDir, err := r.isDir(gitDir)
	if err != nil || isDir {
		return false
	}
	exists, err := r.files.Exists(filepath.Join(r.opts.Dir, filepath.FromSlash(gitDir)))
	return err == nil && exists
}

// agents: the rules block is what makes the agent read them at all.
func (r *runner) agents() domaindoctor.Check {
	content, exists, err := r.read(agentsFile)
	if err != nil {
		return r.unreadable(domaindoctor.CheckAgents, agentsFile, err)
	}
	if !exists {
		return domaindoctor.Broken(domaindoctor.CheckAgents,
			"no hay "+agentsFile+": el agente no lee las reglas",
			"corré odachi init")
	}
	if !domaindoctor.HasBlock(content) {
		return domaindoctor.Broken(domaindoctor.CheckAgents,
			agentsFile+" no tiene el bloque de odachi (faltan "+domaindoctor.AgentsBegin+" y "+domaindoctor.AgentsEnd+"): el agente no lee las reglas",
			"corré odachi init")
	}
	return domaindoctor.OK(domaindoctor.CheckAgents, "sí: "+agentsFile+" con el bloque")
}

// hook: the local gate. A hook odachi did not write is somebody else's work and
// `init` refuses to replace it, so the repair is to run init and paste its block.
func (r *runner) hook() domaindoctor.Check {
	content, exists, err := r.read(hookRel)
	if err != nil {
		return r.unreadable(domaindoctor.CheckHook, hookRel, err)
	}
	switch {
	case !exists && r.inWorktree():
		return domaindoctor.Broken(domaindoctor.CheckHook,
			"estás en un worktree: los hooks viven en el repo principal, así que acá no hay gate local",
			"corré odachi init desde el repo principal (donde "+gitDir+" es un directorio)")
	case !exists:
		return domaindoctor.Broken(domaindoctor.CheckHook,
			"no hay "+hookRel+": no hay gate local",
			"corré odachi init")
	case !domaindoctor.IsOurHook(content):
		return domaindoctor.Broken(domaindoctor.CheckHook,
			hookRel+" no es de odachi: no hay gate local",
			"corré odachi init y pegá su bloque en el hook")
	default:
		return domaindoctor.OK(domaindoctor.CheckHook, "sí: "+hookRel)
	}
}

// hookExe: a hook is only a gate if the binary it calls still exists. When there
// is no odachi hook, this check reports nothing extra: the previous line already
// named the one cause.
func (r *runner) hookExe() domaindoctor.Check {
	content, exists, err := r.read(hookRel)
	if err != nil {
		return r.unreadable(domaindoctor.CheckHookExe, hookRel, err)
	}
	if !exists || !domaindoctor.IsOurHook(content) {
		return domaindoctor.OK(domaindoctor.CheckHookExe, "no hay hook de odachi que revisar")
	}

	exe := domaindoctor.HookExe(content)
	if exe == "" {
		// A hook with no ODACHI= line runs `odachi` from the PATH.
		exe = "odachi"
	}

	if isBareName(exe) {
		if _, err := r.lookup.LookPath(exe); err != nil {
			return domaindoctor.Broken(domaindoctor.CheckHookExe,
				"el hook corre «"+exe+"» y un shell no lo encuentra: el hook falla en cada commit",
				domaindoctor.PathFix(r.opts.Exe))
		}
		return domaindoctor.OK(domaindoctor.CheckHookExe, "sí: "+exe+" (desde el PATH)")
	}

	found, err := r.regular(exe)
	if err != nil {
		return r.unreadable(domaindoctor.CheckHookExe, exe, err)
	}
	if !found {
		return domaindoctor.Broken(domaindoctor.CheckHookExe,
			"el hook corre «"+filepath.ToSlash(exe)+"», que ya no existe: el hook falla en cada commit",
			"corré odachi init para que apunte al binario actual")
	}
	return domaindoctor.OK(domaindoctor.CheckHookExe, "sí: "+filepath.ToSlash(exe))
}

// profile: a profile that does not resolve breaks the ceremony for every verb
// that reads it. A missing file is not a problem — the profile is inferred from
// what the tool can see, and that always resolves.
func (r *runner) profile() domaindoctor.Check {
	name, inferred, err := r.profiles.Resolve(r.opts.Dir)
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckProfile,
			profileFile+" no resuelve: "+oneLine(err.Error()),
			"corregí el archivo (o borralo) y volvé a correr odachi doctor")
	}
	if inferred {
		return domaindoctor.OK(domaindoctor.CheckProfile, "no hay archivo: se infiere el perfil "+name)
	}
	return domaindoctor.OK(domaindoctor.CheckProfile, "perfil "+name)
}

// brief: without a vision there is nothing to build, and an incomplete brief
// blocks the gate with the same sentence every time.
func (r *runner) brief() domaindoctor.Check {
	content, exists, err := r.read(briefFile)
	if err != nil {
		return r.unreadable(domaindoctor.CheckBrief, briefFile, err)
	}
	if !exists {
		return domaindoctor.Broken(domaindoctor.CheckBrief,
			"no hay "+briefFile+": sin visión no hay nada que construir",
			"corré odachi start")
	}

	if missing := brief.Parse(content).Missing(); len(missing) > 0 {
		return domaindoctor.Broken(domaindoctor.CheckBrief,
			"a "+briefFile+" le falta "+strings.Join(missing, ", "),
			"corré odachi start --force (o completá el archivo)")
	}
	return domaindoctor.OK(domaindoctor.CheckBrief, "completo")
}

// skills answers two questions `init` made easy to confuse, with one answer each
// and neither of them hidden: how many skills the project registers in skills/
// with an up-to-date SKILLS.md, and how many are installed where the clients read
// them. A project that just ran `odachi init` read "ninguna: no hay skills/"
// while the skills init had installed sat in .agents/skills, and the two facts
// are not the same fact: the registry is the project's own craft and the
// installed copy is what a client loads.
func (r *runner) skills() domaindoctor.Check {
	registered, fix, err := r.skillRegistry()
	if err != nil {
		return r.unreadable(domaindoctor.CheckSkills, skillsDir, err)
	}

	installed, err := r.installedSkills()
	if err != nil {
		return r.unreadable(domaindoctor.CheckSkills, installedSkillsDir, err)
	}

	detail := registered + "; " + installedSkillsDir + ": " + installed
	if fix != "" {
		return domaindoctor.Broken(domaindoctor.CheckSkills, detail, fix)
	}
	return domaindoctor.OK(domaindoctor.CheckSkills, detail)
}

// skillRegistry is the half that existed before: the project's own skills/ and
// the catalogue generated from them. fix is empty unless the catalogue drifted,
// and a project without skills/ has nothing to drift.
func (r *runner) skillRegistry() (detail, fix string, err error) {
	found, err := r.isDir(skillsDir)
	if err != nil {
		return "", "", err
	}
	if !found {
		return skillsDir + "/: no existe", "", nil
	}

	loaded, err := r.loadSkills()
	if err != nil {
		return "", "", err
	}
	if len(loaded) == 0 {
		return skillsDir + "/: ninguna", "", nil
	}

	want := skill.Registry(loaded)
	content, exists, err := r.read(registry)
	if err != nil {
		return "", "", err
	}
	if !exists || content != want {
		return skillsDir + "/: " + registry + " no está al día: el catálogo se desvió", "corré odachi skills --write", nil
	}
	return skillsDir + "/: " + count(len(loaded), "skill", "skills") + ", " + registry + " al día", "", nil
}

// installedSkills counts the directories that hold a SKILL.md under the folder
// init installs into. A directory without a SKILL.md is not a skill, which is the
// same rule `odachi skills` applies, and a stray file is not counted either.
func (r *runner) installedSkills() (string, error) {
	found, err := r.isDir(installedSkillsDir)
	if err != nil {
		return "", err
	}
	if !found {
		return "ninguna", nil
	}

	dirPath := filepath.Join(r.opts.Dir, filepath.FromSlash(installedSkillsDir))
	names, err := r.files.List(dirPath)
	if err != nil {
		return "", err
	}

	installed := 0
	for _, name := range names {
		entry := filepath.Join(dirPath, name)
		isDir, err := r.files.IsDir(entry)
		if err != nil {
			return "", err
		}
		if !isDir {
			continue
		}
		has, err := r.files.Exists(filepath.Join(entry, skillFile))
		if err != nil {
			return "", err
		}
		if has {
			installed++
		}
	}
	if installed == 0 {
		return "ninguna", nil
	}
	return count(installed, "instalada", "instaladas"), nil
}

// noProfileFile is the second half of check 11 when there is nothing to compare:
// the voice comes from the block, and the profile file is the only thing that can
// contradict it. Saying which voice is in force is still information.
const noProfileFile = "sin " + profileFile + " no hay perfil que comparar"

// voice answers the question a directory listing cannot: does the block talk the
// way the profile says it should, and does the native agent still mirror it? The
// binary writes both, so the halves disagreeing is a project that prints one
// register and talks in another, and nobody notices until the agent sounds wrong.
//
// No profile file is not a problem: there is no claim to compare against, and the
// block carries the voice `odachi init` wrote. The persona files v0.10.0 left
// behind are reported in the same line as leftovers — ok, not a warning: nothing
// reads them and nothing is broken.
func (r *runner) voice() domaindoctor.Check {
	content, exists, err := r.read(agentsFile)
	if err != nil {
		return r.unreadable(domaindoctor.CheckVoice, agentsFile, err)
	}
	if !exists {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"no hay "+agentsFile+": la voz no está en el proyecto",
			"corré odachi init"), "")
	}
	if !rules.HasBlock(content) {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			agentsFile+" no tiene el bloque de odachi: no hay voz que comparar",
			"corré odachi init"), content)
	}

	published, hasVoice := rules.VoiceIn(content)

	name, inferred, err := r.profiles.Resolve(r.opts.Dir)
	if err != nil {
		// The profile is check 7's to report. Here it only means there is
		// nothing to compare the voice against.
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
			"no comparé la voz: el perfil no resuelve (ver el chequeo 7)"), content)
	}
	if inferred {
		detail := noProfileFile
		if hasVoice {
			detail += "; la voz del bloque es la de " + named(published)
		} else {
			detail += " y el bloque no tiene sección de voz"
		}
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice, detail), content)
	}

	want, err := persona.Voice(name)
	if err != nil {
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
			"no comparé la voz: "+profileFile+" nombra un perfil sin voz ("+oneLine(err.Error())+")"), content)
	}

	if !hasVoice {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"el bloque no tiene la sección «"+rules.Section+"» y "+profileFile+" dice "+name+": la voz no está en el bloque",
			"corré odachi init"), content)
	}

	if published != strings.TrimSpace(want) {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"la voz del bloque es la de "+named(published)+" y "+profileFile+" dice "+name+
				": el agente habla distinto de lo que el perfil promete",
			"corré odachi level --to "+name), content)
	}

	return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
		"la voz del bloque es la de "+name+", la misma que dice "+profileFile), content)
}

// voiceCheck appends the native agent mirror and the leftovers of v0.10.0 to a
// line about the voice. Neither changes the status: the block is the invariant,
// the mirror is a convenience, and odachi does not delete files it did not write
// in this version.
func (r *runner) voiceCheck(check domaindoctor.Check, content string) domaindoctor.Check {
	notes := make([]string, 0, 2)
	if note := r.personaNote(content); note != "" {
		notes = append(notes, note)
	}
	note, err := r.leftovers()
	if err != nil {
		return r.unreadable(domaindoctor.CheckVoice, agentsFile, err)
	}
	if note != "" {
		notes = append(notes, note)
	}
	if len(notes) == 0 {
		return check
	}
	detail := strings.Join(notes, "; ")
	switch {
	case check.Detail == "":
		check.Detail = detail
	default:
		check.Detail += "; " + detail
	}
	return check
}

// personaNote compares the native agent files with the block the human has. A
// file that is missing or different is the one thing a selector cannot explain:
// the agent a human picks would read a voice the project no longer promises. It
// is a note and not a Problem because the block still works without a selector;
// the fix is the same command that installs the mirror.
func (r *runner) personaNote(content string) string {
	body, ok := rules.BodyIn(content)
	if !ok {
		return ""
	}

	missing, stale := []string{}, []string{}
	for _, file := range persona.Files(body) {
		current, exists, err := r.read(file.Path)
		switch {
		case err != nil:
			return "no pude leer " + file.Path + ": " + oneLine(err.Error()) + "; corré odachi init"
		case !exists:
			missing = append(missing, file.Path)
		case current != file.Content:
			stale = append(stale, file.Path)
		}
	}

	dirs := make([]string, 0, len(persona.Clients()))
	for _, client := range persona.Clients() {
		dirs = append(dirs, client.Dir)
	}

	switch {
	case len(missing) > 0:
		return "el agente nativo no está en " + strings.Join(missing, ", ") + "; corré odachi init (el bloque alcanza, pero no se elige con Tab)"
	case len(stale) > 0:
		return "el agente nativo no coincide con el bloque en " + strings.Join(stale, ", ") + "; corré odachi init"
	default:
		return "el agente nativo coincide con el bloque en " + strings.Join(dirs, " y ")
	}
}

// leftovers names the per-client persona files v0.10.0 wrote, when they are still
// there: one line for the human to read, or empty when there are none. They still
// work as a voice, but they no longer mirror the block, so the new agent is the
// one to pick.
func (r *runner) leftovers() (string, error) {
	found := make([]string, 0, len(persona.Deprecated()))
	for _, file := range persona.Deprecated() {
		exists, err := r.exists(file)
		if err != nil {
			return "", err
		}
		if exists {
			found = append(found, file)
		}
	}
	if len(found) == 0 {
		return "", nil
	}
	return "quedaron de v0.10.0 " + strings.Join(persona.DeprecatedDirs(), " y ") +
		" (las personas viejas): el agente de esta versión es " + persona.FileName + "; podés borrarlas", nil
}

// named is the voice a body belongs to, by name. Text that is neither is text
// somebody edited by hand, and saying so is more useful than naming a voice that
// is not there.
func named(body string) string {
	for _, name := range persona.Names() {
		voice, err := persona.Voice(name)
		if err != nil {
			continue
		}
		if strings.TrimSpace(voice) == body {
			return name
		}
	}
	return "ninguna de las dos: la editaron a mano"
}

// loadSkills reads every skills/<name>/SKILL.md. A directory without a SKILL.md
// is not a skill, which is the same rule `odachi skills` applies.
func (r *runner) loadSkills() ([]skill.Skill, error) {
	dirPath := filepath.Join(r.opts.Dir, skillsDir)
	names, err := r.files.List(dirPath)
	if err != nil {
		return nil, err
	}

	loaded := make([]skill.Skill, 0, len(names))
	for _, name := range names {
		entry := filepath.Join(dirPath, name)
		isDir, err := r.files.IsDir(entry)
		if err != nil {
			return nil, err
		}
		if !isDir {
			continue
		}

		rel := skillsDir + "/" + name + "/" + skillFile
		content, exists, err := r.read(rel)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		loaded = append(loaded, skill.Parse(name, rel, content))
	}
	return loaded, nil
}

// ignored: the review store and proof/ live inside the tree they certify, so
// without an ignore rule every write invalidates the evidence that was already
// there. Two mechanisms count, and either one is enough: the project's own
// .gitignore (the block `odachi init` writes) and the legacy per-directory
// .gitignore this tool used to write lazily. Only when neither is there is the
// evidence moving the tree, and then the fix is the command that writes the rule:
// init owns it now, and `check`/`review` are what produce evidence, not what
// protects it.
//
// An existing but empty directory is still a warning: nothing is broken yet, and
// the next write is the one that would be.
func (r *runner) ignored() domaindoctor.Check {
	type part struct {
		rel    string
		ignore string
		detail string
	}

	parts := []part{
		{
			rel:    reviewStore,
			ignore: reviewStore + "/" + ignoreFile,
			detail: reviewStore + " no está ignorado: cada review invalidaría la prueba",
		},
		{
			rel:    proofDir,
			ignore: proofDir + "/" + ignoreFile,
			detail: proofDir + "/ no está ignorado: cada prueba invalidaría la siguiente",
		},
	}

	gitignore, hasGitignore, err := r.read(ignoreFile)
	if err != nil {
		return r.unreadable(domaindoctor.CheckIgnored, ignoreFile, err)
	}

	protected := []string{}
	problems := []string{}
	warnings := []string{}

	for _, p := range parts {
		found, err := r.isDir(p.rel)
		if err != nil {
			return r.unreadable(domaindoctor.CheckIgnored, p.rel, err)
		}
		if !found {
			continue
		}

		// The repository's rules first: that is what init writes, and it survives
		// a clone. The lazy file is the second mechanism, kept for the projects
		// that never ran init.
		if hasGitignore && domaindoctor.IgnoredBy(gitignore, p.rel) {
			protected = append(protected, p.rel+" por "+ignoreFile)
			continue
		}

		has, err := r.exists(p.ignore)
		if err != nil {
			return r.unreadable(domaindoctor.CheckIgnored, p.ignore, err)
		}
		if has {
			protected = append(protected, p.rel+" con su propio "+ignoreFile)
			continue
		}

		files, err := r.files.List(filepath.Join(r.opts.Dir, filepath.FromSlash(p.rel)))
		if err != nil {
			return r.unreadable(domaindoctor.CheckIgnored, p.rel, err)
		}
		if len(files) == 0 {
			warnings = append(warnings, p.rel+"/ existe y está vacío: todavía no hay nada que ignorar")
			continue
		}

		problems = append(problems, p.detail)
	}

	switch {
	case len(problems) > 0:
		return domaindoctor.Broken(domaindoctor.CheckIgnored, strings.Join(problems, "; "), "corré odachi init")
	case len(warnings) > 0:
		return domaindoctor.Warning(domaindoctor.CheckIgnored, strings.Join(warnings, "; "))
	case len(protected) == 0:
		return domaindoctor.OK(domaindoctor.CheckIgnored, "no hay review ni pruebas todavía: nada que ignorar")
	default:
		return domaindoctor.OK(domaindoctor.CheckIgnored, "sí: "+strings.Join(protected, " y ")+": no mueven el árbol que certifican")
	}
}

// unreadable turns an adapter error into a line a human can act on. doctor reads
// a project it did not write, so a permission error is a line and never a panic.
func (r *runner) unreadable(number int, rel string, err error) domaindoctor.Check {
	return domaindoctor.Broken(number,
		"no pude leer "+rel+": "+oneLine(err.Error()),
		"revisá los permisos y volvé a correr odachi doctor")
}

// read is one project-relative read. exists is false when the file is not there,
// which every check treats as a state and not as an error.
func (r *runner) read(rel string) (content string, exists bool, err error) {
	path := filepath.Join(r.opts.Dir, filepath.FromSlash(rel))

	exists, err = r.files.Exists(path)
	if err != nil || !exists {
		return "", false, err
	}

	data, err := r.files.Read(path)
	if err != nil {
		return "", true, err
	}
	return string(data), true, nil
}

// exists is read's question without the bytes.
func (r *runner) exists(rel string) (bool, error) {
	path := filepath.Join(r.opts.Dir, filepath.FromSlash(rel))
	return r.files.Exists(path)
}

// isDir is exists for a directory: a file with that name is not what the check
// is asking about.
func (r *runner) isDir(rel string) (bool, error) {
	path := filepath.Join(r.opts.Dir, filepath.FromSlash(rel))
	return r.files.IsDir(path)
}

// regular is "a file, not a directory, not nothing": the hook names a binary, and
// a directory with that name is not one.
func (r *runner) regular(path string) (bool, error) {
	found, err := r.files.Exists(path)
	if err != nil || !found {
		return false, err
	}
	isDir, err := r.files.IsDir(path)
	if err != nil {
		return false, err
	}
	return !isDir, nil
}

// isBareName reports whether the hook named the binary by name instead of by
// path: a bare name is resolved through the PATH, a path is resolved by existing.
func isBareName(name string) bool {
	return !strings.ContainsAny(name, `/\`)
}

// count is how a Spanish sentence counts things.
func count(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

// oneLine keeps every message on one line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
