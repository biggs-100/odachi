// Package doctor implements `odachi doctor`: eleven read-only health checks,
// each with one English line and, when something is wrong, the phrase that
// repairs it.
//
// It answers the question a person has after installing the tool — "is
// everything installed right?" — without them reading the repository. The most valuable of
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
	"path"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/brief"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/layout"
	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/skill"
	"github.com/biggs-100/odachi/internal/domain/store"
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
	profileFile = layout.Profile
	briefFile   = "brief.md"
	skillsDir   = "skills"
	skillFile   = "SKILL.md"
	registry    = skill.RegistryFile
	// installedSkillsDir is where `odachi init` puts the skills every client
	// reads. The row that installs there belongs to the agent table — it is the
	// `standard` row — and a test pins that this string still matches it, so the
	// count doctor reports is the count init wrote.
	installedSkillsDir = ".agents/skills"
	proofDir           = "proof"
	ignoreFile         = ".gitignore"
)

// Options is one `odachi doctor` invocation.
type Options struct {
	Dir string
	// Version is what this binary reports, so the first line answers "what did
	// I install?" before anything else.
	Version string
	// Exe is the running binary. It is only used to name the folder a human has
	// to add to the PATH, and the message stays honest when it is empty.
	Exe string
	// Home is the human's home, where the global door lives. Empty means the
	// caller could not resolve it; the check skips the door instead of guessing.
	Home string
	// Agents is the client selection the project stored. nil means no store:
	// both clients, the default of a project that never chose.
	Agents Agents
}

// Agents is the selection port: which clients the project chose for the native
// agent, and whether it carries them at all. It only reads: doctor has no way to
// change the choice.
type Agents interface {
	Written(dir string) (persona.Store, error)
}

// Files is the disk port. It can only read: doctor has no way to write, which is
// what "strictly read-only" means structurally.
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
		r.taskDocs(),
		r.legacyFolder(),
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
// A read error is not a verdict. Saying "this folder is not a project" about a
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

	found, err := r.isDir(layout.Odd)
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
			"it is not on the PATH: the agent cannot run the rules",
			domaindoctor.PathFix(r.opts.Exe))
	}
	return domaindoctor.OK(domaindoctor.CheckOnPath, "yes: "+filepath.ToSlash(path))
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
			return domaindoctor.OK(domaindoctor.CheckGit, "yes: worktree (the real .git is in the main repository)")
		}
		return domaindoctor.Broken(domaindoctor.CheckGit, "there is no "+gitDir+": no hook is possible", "run git init")
	}
	return domaindoctor.OK(domaindoctor.CheckGit, "yes")
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
			"there is no "+agentsFile+": the agent does not read the rules",
			"run odachi init")
	}
	if !domaindoctor.HasBlock(content) {
		return domaindoctor.Broken(domaindoctor.CheckAgents,
			agentsFile+" does not have the odachi block (missing "+domaindoctor.AgentsBegin+" and "+domaindoctor.AgentsEnd+"): the agent does not read the rules",
			"run odachi init")
	}
	if detail, stale := r.staleBlock(content); stale {
		return domaindoctor.Warning(domaindoctor.CheckAgents, detail)
	}
	return domaindoctor.OK(domaindoctor.CheckAgents, "yes: "+agentsFile+" with the block")
}

// staleBlock reports whether the block AGENTS.md carries differs from the one this
// binary writes for the project's voice. Presence is not currency: a project that
// upgraded odachi keeps the rules it adopted with, and a rule that arrived later never
// reaches its agent - the one about other folders, or the four hard rules when they
// changed shape. It is a warning and not a blocker: the old block still works, it is
// simply older. Line endings and trailing spaces are not a stale block, for the reason
// 0031 already wrote for the brief: a Windows checkout is not an older rule.
func (r *runner) staleBlock(content string) (string, bool) {
	// The parser folds line endings once (rules.fold), so a CRLF checkout reads here too;
	// what is left to ignore is what a text editor does: trailing spaces and the blank
	// lines at the end.
	body, ok := rules.BodyIn(content)
	if !ok {
		return "", false
	}
	name, _, err := r.profiles.Resolve(r.opts.Dir)
	if err != nil {
		// The profile port is what names the voice; a port that cannot answer leaves
		// nothing to generate the block with. Check 7 owns that refusal.
		return "", false
	}
	want, err := rules.Body(name)
	if err != nil {
		return "", false
	}
	if sameBlock(body, want) {
		return "", false
	}
	return agentsFile + " carries an older block than this binary writes: the rules that arrived after this project adopted odachi are not in it (run odachi init)", true
}

// sameBlock compares two blocks the way a person reads them: line endings, trailing
// spaces and the blank lines at the end are not a difference; the words are.
func sameBlock(carried, want string) bool {
	return normalizeBlock(carried) == normalizeBlock(want)
}

// normalizeBlock folds what a checkout does to a file: CRLF becomes LF, a trailing
// space goes, and the text ends where its last word does.
func normalizeBlock(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
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
			"you are in a worktree: hooks live in the main repository, so there is no local gate here",
			"run odachi init from the main repository (where "+gitDir+" is a directory)")
	case !exists:
		return domaindoctor.Broken(domaindoctor.CheckHook,
			"there is no "+hookRel+": there is no local gate",
			"run odachi init")
	case !domaindoctor.IsOurHook(content):
		return domaindoctor.Broken(domaindoctor.CheckHook,
			hookRel+" is not odachi's: there is no local gate",
			"run odachi init and paste its block into the hook")
	default:
		return domaindoctor.OK(domaindoctor.CheckHook, "yes: "+hookRel)
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
		return domaindoctor.OK(domaindoctor.CheckHookExe, "there is no odachi hook to check")
	}

	exe := domaindoctor.HookExe(content)
	if exe == "" {
		// A hook with no ODACHI= line runs `odachi` from the PATH.
		exe = "odachi"
	}

	if isBareName(exe) {
		if _, err := r.lookup.LookPath(exe); err != nil {
			return domaindoctor.Broken(domaindoctor.CheckHookExe,
				"the hook runs «"+exe+"» and a shell cannot find it: the hook fails on every commit",
				domaindoctor.PathFix(r.opts.Exe))
		}
		return domaindoctor.OK(domaindoctor.CheckHookExe, "yes: "+exe+" (from the PATH)")
	}

	found, err := r.regular(exe)
	if err != nil {
		return r.unreadable(domaindoctor.CheckHookExe, exe, err)
	}
	if !found {
		return domaindoctor.Broken(domaindoctor.CheckHookExe,
			"the hook runs «"+filepath.ToSlash(exe)+"», which no longer exists: the hook fails on every commit",
			"run odachi init so it points at the current binary")
	}
	return domaindoctor.OK(domaindoctor.CheckHookExe, "yes: "+filepath.ToSlash(exe))
}

// profile: a profile that does not resolve breaks the ceremony for every verb
// that reads it. A missing file is not a problem — the profile is inferred from
// what the tool can see, and that always resolves.
func (r *runner) profile() domaindoctor.Check {
	name, inferred, err := r.profiles.Resolve(r.opts.Dir)
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckProfile,
			profileFile+" does not resolve: "+oneLine(err.Error()),
			"fix the file (or delete it) and run odachi doctor again")
	}
	if inferred {
		return domaindoctor.OK(domaindoctor.CheckProfile, "there is no file: the "+name+" profile is inferred")
	}
	return domaindoctor.OK(domaindoctor.CheckProfile, "profile "+name)
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
			"there is no "+briefFile+": without a vision there is nothing to build",
			"run odachi start")
	}

	if missing := brief.Parse(content).Missing(); len(missing) > 0 {
		return domaindoctor.Broken(domaindoctor.CheckBrief,
			"missing from "+briefFile+": "+strings.Join(missing, ", "),
			"run odachi start --force (or complete the file)")
	}
	return domaindoctor.OK(domaindoctor.CheckBrief, "complete")
}

// skills answers two questions `init` made easy to confuse, with one answer each
// and neither of them hidden: how many skills the project registers in skills/
// with an up-to-date SKILLS.md, and how many are installed where the clients read
// them. A project that just ran `odachi init` read "none: there is no skills/"
// while the skills init had installed sat in .agents/skills, and the two facts
// are not the same fact: the registry is the project's own craft and the
// installed copy is what a client loads.
func (r *runner) skills() domaindoctor.Check {
	registered, fix, err := r.skillRegistry()
	if err != nil {
		return r.unreadable(domaindoctor.CheckSkills, skillsDir, err)
	}

	installed, compared, drift, err := r.installedSkills()
	if err != nil {
		return r.unreadable(domaindoctor.CheckSkills, installedSkillsDir, err)
	}
	home, homeCompared, homeDrift, err := r.homeSkills()
	if err != nil {
		return r.unreadable(domaindoctor.CheckSkills, "~/.agents/skills", err)
	}

	detail := registered + "; " + installedSkillsDir + ": " + installed
	if compared > 0 && len(drift) == 0 {
		detail += "; " + "compared: " + count(compared, "match", "matches") + " in " + skillsDir + "/"
	}
	if home != "" {
		detail += "; " + home
		if homeCompared > 0 && len(homeDrift) == 0 {
			detail += "; " + "compared: " + count(homeCompared, "match", "matches") + " in " + skillsDir + "/"
		}
	}
	drift = append(drift, homeDrift...)
	switch {
	case fix != "":
		return domaindoctor.Broken(domaindoctor.CheckSkills, detail, fix)
	case len(drift) > 0:
		return domaindoctor.Warning(domaindoctor.CheckSkills, detail+
			"; do not match "+skillsDir+"/: "+strings.Join(drift, ", ")+"; run odachi init (or odachi init --skills home)")
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
		return skillsDir + "/: does not exist", "", nil
	}

	loaded, err := r.loadSkills()
	if err != nil {
		return "", "", err
	}
	if len(loaded) == 0 {
		return skillsDir + "/: none", "", nil
	}

	want := skill.Registry(loaded)
	content, exists, err := r.read(registry)
	if err != nil {
		return "", "", err
	}
	if !exists || content != want {
		return skillsDir + "/: " + registry + " is not up to date: the catalogue drifted", "run odachi skills --write", nil
	}
	return skillsDir + "/: " + count(len(loaded), "skill", "skills") + ", " + registry + " up to date", "", nil
}

// installedSkills counts the directories that hold a SKILL.md under the folder
// init installs into, and compares every one that also exists in the project's
// source `skills/`. A directory without a SKILL.md is not a skill, which is the
// same rule `odachi skills` applies, and a stray file is not counted either.
//
// compared is how many installed skills had a source to compare against; drift
// names the ones whose bytes differ. A project without `skills/` has no source:
// it gets the count and no comparison, because the installed copy is the only
// copy it has.
func (r *runner) installedSkills() (detail string, compared int, drift []string, err error) {
	found, err := r.isDir(installedSkillsDir)
	if err != nil {
		return "", 0, nil, err
	}
	if !found {
		return "none", 0, nil, nil
	}

	dirPath := filepath.Join(r.opts.Dir, filepath.FromSlash(installedSkillsDir))
	names, err := r.files.List(dirPath)
	if err != nil {
		return "", 0, nil, err
	}

	installed := 0
	for _, name := range names {
		entry := filepath.Join(dirPath, name)
		isDir, err := r.files.IsDir(entry)
		if err != nil {
			return "", 0, nil, err
		}
		if !isDir {
			continue
		}

		installedPath := filepath.Join(entry, skillFile)
		has, err := r.files.Exists(installedPath)
		if err != nil {
			return "", 0, nil, err
		}
		if !has {
			continue
		}
		installed++

		sourcePath := filepath.Join(r.opts.Dir, skillsDir, name, skillFile)
		sourceExists, err := r.files.Exists(sourcePath)
		if err != nil {
			return "", 0, nil, err
		}
		if !sourceExists {
			continue
		}
		compared++

		installedData, err := r.files.Read(installedPath)
		if err != nil {
			return "", 0, nil, err
		}
		sourceData, err := r.files.Read(sourcePath)
		if err != nil {
			return "", 0, nil, err
		}
		if string(installedData) != string(sourceData) {
			drift = append(drift, installedSkillsDir+"/"+name)
		}
	}

	if installed == 0 {
		return "none", 0, nil, nil
	}
	return count(installed, "installed", "installed"), compared, drift, nil
}

// homeSkills is the other scope of the same question: `init --skills home`
// installs the whole payload in ~/.agents/skills, so a doctor that only counted
// the project copy would call that project empty. It compares the same way, when
// the project has a source `skills/` to compare against.
func (r *runner) homeSkills() (detail string, compared int, drift []string, err error) {
	if r.opts.Home == "" {
		return "", 0, nil, nil
	}

	dirPath := filepath.Join(r.opts.Home, filepath.FromSlash(installedSkillsDir))
	found, err := r.files.IsDir(dirPath)
	if err != nil {
		return "", 0, nil, err
	}
	if !found {
		return "~/.agents/skills: none", 0, nil, nil
	}

	names, err := r.files.List(dirPath)
	if err != nil {
		return "", 0, nil, err
	}

	installed := 0
	for _, name := range names {
		entry := filepath.Join(dirPath, name)
		isDir, err := r.files.IsDir(entry)
		if err != nil {
			return "", 0, nil, err
		}
		if !isDir {
			continue
		}

		installedPath := filepath.Join(entry, skillFile)
		has, err := r.files.Exists(installedPath)
		if err != nil {
			return "", 0, nil, err
		}
		if !has {
			continue
		}
		installed++

		sourcePath := filepath.Join(r.opts.Dir, skillsDir, name, skillFile)
		sourceExists, err := r.files.Exists(sourcePath)
		if err != nil {
			return "", 0, nil, err
		}
		if !sourceExists {
			continue
		}
		compared++

		installedData, err := r.files.Read(installedPath)
		if err != nil {
			return "", 0, nil, err
		}
		sourceData, err := r.files.Read(sourcePath)
		if err != nil {
			return "", 0, nil, err
		}
		if string(installedData) != string(sourceData) {
			drift = append(drift, "~/.agents/skills/"+name)
		}
	}

	if installed == 0 {
		return "~/.agents/skills: none", 0, nil, nil
	}
	return "~/.agents/skills: " + count(installed, "installed", "installed"), compared, drift, nil
}

// noProfileFile is the second half of check 11 when there is nothing to compare:
// the voice comes from the block, and the profile file is the only thing that can
// contradict it. Saying which voice is in force is still information.
const noProfileFile = "without " + profileFile + " there is no profile to compare"

// voice answers the question a directory listing cannot: does the block talk the
// way the profile says it should, and does the native agent still mirror it? The
// binary writes both, so the halves disagreeing is a project that prints one
// register and talks in another, and nobody notices until the agent sounds wrong.
//
// No profile file is not a problem: there is no claim to compare against, and the
// block carries the voice `odachi init` wrote. The persona files v0.10.0 left
// behind are reported in the same line as leftovers — ok, not a warning: nothing
// reads them and nothing is broken.
// taskDocs names the documents in odd/tasks that odachi does not read. It never
// fails the run: another shape of task document is a project decision, not a
// broken install, and the line exists so the human hears what the tool skips
// instead of wondering why summary counts one.
func (r *runner) taskDocs() domaindoctor.Check {
	skipped, err := tasks.Skipped(r.files, r.opts.Dir)
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckTaskDocs, oneLine(err.Error()), "check the permissions of "+tasks.Dir)
	}
	if len(skipped) == 0 {
		return domaindoctor.OK(domaindoctor.CheckTaskDocs, "")
	}
	return domaindoctor.Warning(domaindoctor.CheckTaskDocs, tasks.SkippedLine(skipped)+"; the project keeps two memories")
}

func (r *runner) voice() domaindoctor.Check {
	content, exists, err := r.read(agentsFile)
	if err != nil {
		return r.unreadable(domaindoctor.CheckVoice, agentsFile, err)
	}
	if !exists {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"there is no "+agentsFile+": the voice is not in the project",
			"run odachi init"), "")
	}
	if !rules.HasBlock(content) {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			agentsFile+" does not have the odachi block: there is no voice to compare",
			"run odachi init"), content)
	}

	published, hasVoice := rules.VoiceIn(content)

	name, inferred, err := r.profiles.Resolve(r.opts.Dir)
	if err != nil {
		// The profile is check 7's to report. Here it only means there is
		// nothing to compare the voice against.
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
			"I did not compare the voice: the profile does not resolve (see check 7)"), content)
	}
	if inferred {
		detail := noProfileFile
		if hasVoice {
			detail += "; the block voice is " + named(published)
		} else {
			detail += " and the block has no voice section"
		}
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice, detail), content)
	}

	want, err := persona.Voice(name)
	if err != nil {
		return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
			"I did not compare the voice: "+profileFile+" names a profile without a voice ("+oneLine(err.Error())+")"), content)
	}

	if !hasVoice {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"the block has no «"+rules.Section+"» section and "+profileFile+" says "+name+": the voice is not in the block",
			"run odachi init"), content)
	}

	if published != strings.TrimSpace(want) {
		return r.voiceCheck(domaindoctor.Broken(domaindoctor.CheckVoice,
			"the block voice is "+named(published)+" and "+profileFile+" says "+name+
				": the agent talks differently from what the profile promises",
			"run odachi level --to "+name), content)
	}

	return r.voiceCheck(domaindoctor.OK(domaindoctor.CheckVoice,
		"the block voice is "+name+", the same one "+profileFile+" says"), content)
}

// voiceCheck appends the native agent mirror, the global door and the leftovers
// of v0.10.0 to a line about the voice. None changes the status: the block is the
// invariant, the agents are a convenience, and odachi does not delete files it
// did not write in this version.
func (r *runner) voiceCheck(check domaindoctor.Check, content string) domaindoctor.Check {
	notes := make([]string, 0, 4)
	selection, clients, selectionNote, readable := r.clientSelection()
	if readable {
		if note := r.personaNote(content, selection, clients); note != "" {
			notes = append(notes, note)
		}
		if note := r.globalPersonaNote(clients); note != "" {
			notes = append(notes, note)
		}
	}
	if selectionNote != "" {
		notes = append(notes, selectionNote)
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

// clientSelection reads the choice the project stored. A project that never
// chose, or a doctor without the port, keeps both clients: the same default init
// has. A file doctor cannot read is a note, not a failed check: the block still
// works, and the line tells the human which file to fix instead of guessing a
// client and reporting files that were never meant to exist.
func (r *runner) clientSelection() (string, []persona.Client, string, bool) {
	if r.opts.Agents == nil {
		return persona.SelectionBoth, persona.Clients(), "", true
	}
	stored, err := r.opts.Agents.Written(r.opts.Dir)
	if err != nil {
		return "", nil, "could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; fix the file or run odachi init --agents both", false
	}
	name, clients, err := persona.ResolveSelection(stored.Agents)
	if err != nil {
		return "", nil, "could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; fix the file or run odachi init --agents both", false
	}
	note := ""
	if name != persona.SelectionBoth {
		note = "I only check " + name + ", as " + persona.SelectionPath + " says"
	}
	return name, clients, note, true
}

// personaNote compares the native agent files with the block the human has. A
// file that is missing or different is the one thing a selector cannot explain:
// the agent a human picks would read a voice the project no longer promises. It
// is a note and not a Problem because the block still works without a selector;
// the fix is the same command that installs the mirror.
func (r *runner) personaNote(content, selection string, clients []persona.Client) string {
	body, ok := rules.BodyIn(content)
	if !ok {
		return ""
	}

	missing, stale := []string{}, []string{}
	for _, client := range clients {
		file, ok := persona.FilesFor(client.Agent, body)
		if !ok {
			continue
		}
		current, exists, err := r.read(file.Path)
		switch {
		case err != nil:
			return "could not read " + file.Path + ": " + oneLine(err.Error()) + "; run odachi init"
		case !exists:
			missing = append(missing, file.Path)
		case current != file.Content:
			stale = append(stale, file.Path)
		}
		for _, sub := range persona.SubagentFilesFor(client.Agent) {
			current, exists, err := r.read(sub.Path)
			switch {
			case err != nil:
				return "could not read " + sub.Path + ": " + oneLine(err.Error()) + "; run odachi init"
			case !exists:
				missing = append(missing, sub.Path)
			case current != sub.Content:
				stale = append(stale, sub.Path)
			}
		}
	}

	dirs := make([]string, 0, len(clients))
	subDirs := make([]string, 0, len(clients))
	for _, client := range clients {
		dirs = append(dirs, client.Dir)
		subDirs = append(subDirs, client.SubagentDir)
	}

	choice := r.storedProjectChoice()
	home := r.homeCarriesTheAgents(clients)
	store := persona.SelectionPath + ": project_agents = "

	switch {
	case len(missing) > 0 && choice == persona.ProjectAgentsNone && home:
		// The store is the authority: the absence of the mirrors is the answer it
		// already recorded, not a gap the home happens to explain.
		return "the project has no agent files, and the store says so (" + store +
			"none): the door in the home is the one that runs"
	case len(missing) > 0 && choice == persona.ProjectAgentsNone:
		return "the project has no agent files because the store keeps the agent in the home (" + store +
			"none), and the home does not carry the door: run odachi init --global"
	case len(missing) > 0 && choice == persona.ProjectAgentsWrite:
		// A recorded write is a decision too: the mirror is what is wrong, and
		// the home carrying the door does not turn that decision into silence.
		return "the store says this project writes its agents (" + store +
			"write), and they are missing in " + strings.Join(missing, ", ") + ": run odachi init"
	case len(missing) > 0 && home:
		// The home explains the absence, and nothing recorded the choice: a
		// teammate who clones this project would not know where its agent lives.
		return "the project has no agent files: the door in the home is the one that runs, and " + persona.SelectionPath +
			" does not say so; run odachi init --global --no-project-agents to record it"
	case len(missing) > 0:
		return "the native agent or its subagents are missing in " + strings.Join(missing, ", ") + "; run odachi init"
	case len(stale) > 0 && choice == persona.ProjectAgentsNone:
		// The store keeps the agent in the home, so init skips these files: the
		// advice has to be something that actually resolves them.
		return "the store keeps the agent in the home (" + store + "none), and " +
			strings.Join(stale, ", ") + " in the project are leftovers nobody reads; odachi does not delete them: remove them, or run odachi init --agents " + selection + " to put them back in use"
	case len(stale) > 0:
		return "the native agent or its subagents do not match what was generated in " + strings.Join(stale, ", ") + "; run odachi init"
	case choice == persona.ProjectAgentsNone:
		return "the native agent matches the block in " + strings.Join(dirs, " and ") +
			"; the subagents match in " + strings.Join(subDirs, " and ") +
			"; the store keeps the agent in the home (" + store + "none), so these files are leftovers nobody reads"
	default:
		return "the native agent matches the block in " + strings.Join(dirs, " and ") +
			"; the subagents match in " + strings.Join(subDirs, " and ")
	}
}

// storedProjectChoice is where the store says the project's agent lives, or
// empty when nothing recorded it. A store doctor cannot read is the selection
// note's business, not this one's.
func (r *runner) storedProjectChoice() string {
	if r.opts.Agents == nil {
		return ""
	}
	stored, err := r.opts.Agents.Written(r.opts.Dir)
	if err != nil {
		return ""
	}
	return stored.ProjectAgents
}

// homeCarriesTheAgents reports whether the home already carries the door and the
// subagents that make the project's agent files optional. It is a question about
// the home, not a guess: no home, or half of what it needs, means no clean mode.
func (r *runner) homeCarriesTheAgents(clients []persona.Client) bool {
	if r.opts.Home == "" {
		return false
	}
	for _, client := range clients {
		file, ok := persona.GlobalFilesFor(client.Agent)
		if !ok {
			continue
		}
		exists, err := r.files.Exists(filepath.Join(r.opts.Home, filepath.FromSlash(file.Path)))
		if err != nil || !exists {
			return false
		}
		for _, sub := range persona.GlobalSubagentFilesFor(client.Agent) {
			subExists, subErr := r.files.Exists(filepath.Join(r.opts.Home, filepath.FromSlash(sub.Path)))
			if subErr != nil || !subExists {
				return false
			}
		}
	}
	return len(clients) > 0
}

// globalPersonaNote compares the door in the human's home with the files this
// binary would install. It is a note, not a Problem: a project with the block and
// the project agent still works without the global door; what is lost is odachi
// appearing in every other folder. The fix is the same `odachi init`.
func (r *runner) globalPersonaNote(clients []persona.Client) string {
	if r.opts.Home == "" {
		return ""
	}

	missing, stale := []string{}, []string{}
	for _, client := range clients {
		file, ok := persona.GlobalFilesFor(client.Agent)
		if !ok {
			continue
		}
		path := filepath.Join(r.opts.Home, filepath.FromSlash(file.Path))
		display := "~/" + file.Path

		exists, err := r.files.Exists(path)
		if err != nil {
			return "could not check the global door " + display + ": " + oneLine(err.Error()) + "; run odachi init"
		}
		if !exists {
			missing = append(missing, display)
			continue
		}

		data, err := r.files.Read(path)
		if err != nil {
			return "could not read the global door " + display + ": " + oneLine(err.Error()) + "; run odachi init"
		}
		if string(data) != file.Content {
			stale = append(stale, display)
		}
	}

	var result string
	switch {
	case len(missing) > 0:
		result = "the global door is not in " + strings.Join(missing, ", ") + "; it is opt-in: run odachi init --global if you want odachi in any folder"
	case len(stale) > 0:
		result = "the global door does not match this binary in " + strings.Join(stale, ", ") + "; run odachi init --global"
	default:
		result = "the global door is up to date in the home"
	}

	// The door is global and the agents it invokes are too, so they are compared
	// by the same command and named by the same line.
	if extra := r.globalSubagentsNote(clients); extra != "" {
		result += "; " + extra
	}

	if extra := r.opencodeDefaultNote(clients); extra != "" {
		if result == "" {
			return extra
		}
		return result + "; " + extra
	}
	return result
}

// globalSubagentsNote compares the subagents in the home with the files this
// binary would install there. Like the door it is a note: the project works
// without them, and what is lost is delegating and verifying in a project that
// never writes `.opencode/`.
func (r *runner) globalSubagentsNote(clients []persona.Client) string {
	if r.opts.Home == "" {
		return ""
	}

	missing, stale := []string{}, []string{}
	dirs := []string{}
	for _, client := range clients {
		files := persona.GlobalSubagentFilesFor(client.Agent)
		if len(files) == 0 {
			continue
		}
		dirs = append(dirs, "~/"+client.HomeSubagentDir)

		for _, file := range files {
			path := filepath.Join(r.opts.Home, filepath.FromSlash(file.Path))
			display := "~/" + file.Path

			exists, err := r.files.Exists(path)
			if err != nil {
				return "could not check the global subagents " + display + ": " + oneLine(err.Error()) + "; run odachi init --global"
			}
			if !exists {
				missing = append(missing, display)
				continue
			}

			data, err := r.files.Read(path)
			if err != nil {
				return "could not read the global subagents " + display + ": " + oneLine(err.Error()) + "; run odachi init --global"
			}
			if string(data) != file.Content {
				stale = append(stale, display)
			}
		}
	}

	switch {
	case len(missing) > 0:
		return "the global subagents are not in " + strings.Join(dirs, ", ") + "; run odachi init --global to install them"
	case len(stale) > 0:
		return "the global subagents do not match this binary in " + strings.Join(stale, ", ") + "; run odachi init --global"
	case len(dirs) == 0:
		return ""
	default:
		return "the global subagents are up to date in " + strings.Join(dirs, ", ")
	}
}

// opencodeDefaultNote says whether the global OpenCode config starts a session
// with the door. It is a note: the door works without it, the human's own
// default is respected, and the fix is the same `odachi init --global`.
func (r *runner) opencodeDefaultNote(clients []persona.Client) string {
	if r.opts.Home == "" {
		return ""
	}

	hasOpenCode := false
	for _, client := range clients {
		if client.Agent == "opencode" {
			hasOpenCode = true
			break
		}
	}
	if !hasOpenCode {
		return ""
	}

	door, ok := persona.GlobalFilesFor("opencode")
	if !ok {
		return ""
	}

	// A missing door is already the previous note's sentence; this one only
	// speaks when the door is installed and the config does not choose it.
	doorPath := filepath.Join(r.opts.Home, filepath.FromSlash(door.Path))
	exists, err := r.files.Exists(doorPath)
	if err != nil || !exists {
		return ""
	}

	rel := path.Join(path.Dir(path.Dir(door.Path)), "opencode.json")
	jsoncRel := strings.TrimSuffix(rel, ".json") + ".jsonc"
	target := filepath.Join(r.opts.Home, filepath.FromSlash(rel))
	display := "~/" + rel

	exists, err = r.files.Exists(target)
	if err != nil {
		return "could not check " + display + ": " + oneLine(err.Error()) + "; run odachi init --global"
	}
	if !exists {
		if jsoncExists, err := r.files.Exists(filepath.Join(r.opts.Home, filepath.FromSlash(jsoncRel))); err == nil && jsoncExists {
			target = filepath.Join(r.opts.Home, filepath.FromSlash(jsoncRel))
			display = "~/" + jsoncRel
			exists = true
		}
	}
	if !exists {
		return "the door is installed but " + display + " has no default_agent; run odachi init --global"
	}

	data, err := r.files.Read(target)
	if err != nil {
		return "could not read " + display + ": " + oneLine(err.Error()) + "; run odachi init --global"
	}

	agent, found, err := domainopencode.CurrentAgent([]byte(data))
	if err != nil {
		return "could not read the default agent in " + display + ": " + oneLine(err.Error()) + "; run odachi init --global"
	}
	if !found {
		return "the door is installed but " + display + " does not choose it; run odachi init --global"
	}
	if agent != persona.AgentName {
		return "your default agent is " + agent + " in " + display + ", not " + persona.AgentName + "; set default_agent if you want odachi in any folder"
	}
	return ""
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
	return "left over from v0.10.0: " + strings.Join(persona.DeprecatedDirs(), " and ") +
		" (the old personas): the agent of this version is " + persona.FileName + "; you can delete them", nil
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
	return "neither one: they edited it by hand"
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

// ignored: proof/ lives inside the tree it certifies, so without an ignore rule
// every write invalidates the evidence that was already there. Two mechanisms
// count, and either one is enough: the project's own .gitignore (the block
// `odachi init` writes) and the per-directory .gitignore the proof itself lands
// with. Only when neither is there is the evidence moving the tree, and then the
// fix is the command that writes the rule: init owns it now, and `check` is what
// produces evidence, not what protects it.
//
// The verdicts are not checked here: they live in the clone's store, outside the
// worktree, so there is no rule they could need.
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
			rel:    proofDir,
			ignore: proofDir + "/" + ignoreFile,
			detail: proofDir + "/ is not ignored: every proof would invalidate the next one",
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
			protected = append(protected, p.rel+" via "+ignoreFile)
			continue
		}

		has, err := r.exists(p.ignore)
		if err != nil {
			return r.unreadable(domaindoctor.CheckIgnored, p.ignore, err)
		}
		if has {
			protected = append(protected, p.rel+" with its own "+ignoreFile)
			continue
		}

		files, err := r.files.List(filepath.Join(r.opts.Dir, filepath.FromSlash(p.rel)))
		if err != nil {
			return r.unreadable(domaindoctor.CheckIgnored, p.rel, err)
		}
		if len(files) == 0 {
			warnings = append(warnings, p.rel+"/ exists and is empty: there is nothing to ignore yet")
			continue
		}

		problems = append(problems, p.detail)
	}

	switch {
	case len(problems) > 0:
		return domaindoctor.Broken(domaindoctor.CheckIgnored, strings.Join(problems, "; "), "run odachi init")
	case len(warnings) > 0:
		return domaindoctor.Warning(domaindoctor.CheckIgnored, strings.Join(warnings, "; "))
	case len(protected) == 0:
		return domaindoctor.OK(domaindoctor.CheckIgnored, "there are no proofs yet: nothing to ignore")
	default:
		return domaindoctor.OK(domaindoctor.CheckIgnored, "yes: "+strings.Join(protected, " and ")+": they do not move the tree they certify")
	}
}

// legacyFolder reports the folder older releases used. The distinction that matters
// is whether the rules inside it are still the ones being read: a project that
// never moved its fence has a fence nobody enforces, and that is a problem, not a
// notice. Once the new path exists too, the old file is dead weight and the line
// is a warning.
func (r *runner) legacyFolder() domaindoctor.Check {
	old := layout.LegacyDir

	oldFiles := []string{}
	for _, name := range []string{filepath.Base(layout.Profile), filepath.Base(layout.Fence), filepath.Base(layout.Review), store.Agents} {
		if r.has(filepath.FromSlash(filepath.Join(old, name))) {
			oldFiles = append(oldFiles, filepath.ToSlash(filepath.Join(old, name)))
		}
	}
	for _, name := range []string{store.Reviews, store.Backups} {
		if found, err := r.isDir(filepath.Join(old, name)); err == nil && found {
			oldFiles = append(oldFiles, filepath.ToSlash(filepath.Join(old, name))+"/")
		}
	}

	if len(oldFiles) == 0 && !r.has(filepath.Join(old, filepath.Base(layout.Roles))) {
		return domaindoctor.OK(domaindoctor.CheckLeftovers, "there is no "+old+"/: the rules live in odd/ and the store in the clone")
	}

	// A rule that moved in name but not in place: the file the team wrote is not
	// the one this binary reads.
	notRead := []string{}
	for _, to := range []string{layout.Profile, layout.Fence, layout.Review, layout.Roles} {
		from := filepath.ToSlash(filepath.Join(old, filepath.Base(to)))
		if r.has(filepath.FromSlash(from)) && !r.has(to) {
			notRead = append(notRead, from+" (today "+to+")")
		}
	}
	if len(notRead) > 0 {
		return domaindoctor.Broken(domaindoctor.CheckLeftovers,
			"the rules this project wrote are not read any more: "+strings.Join(notRead, ", ")+"; a fence nobody enforces is worse than no fence",
			"run odachi store move (it adopts the evidence and prints the git mv for the rules)")
	}

	return domaindoctor.Warning(domaindoctor.CheckLeftovers,
		old+"/ is still there ("+strings.Join(oldFiles, ", ")+") and nothing reads it; run odachi store move to adopt the evidence, then the git mv it prints")
}

// has reports whether a project-relative path exists.
func (r *runner) has(rel string) bool {
	found, err := r.exists(rel)
	return err == nil && found
}

// unreadable turns an adapter error into a line a human can act on. doctor reads
// a project it did not write, so a permission error is a line and never a panic.
func (r *runner) unreadable(number int, rel string, err error) domaindoctor.Check {
	return domaindoctor.Broken(number,
		"could not read "+rel+": "+oneLine(err.Error()),
		"check the permissions and run odachi doctor again")
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

// count is how an English sentence counts things.
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
