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
// One voice, one owner. The profile (in odd/profile.toml, owned by `odachi
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

	"github.com/biggs-100/odachi/internal/domain/layout"
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
const hookMarker = "# odachi: no commit with the gate red"

// hookMarkerOld is the marker written before the proof moved to the feature's
// close. It stays a read-only alias, like the Spanish one: the hook is still
// odachi's and is rewritten in place, so an existing project does not have to paste
// the lines by hand.
const hookMarkerOld = "# odachi: no commit without a proof for this tree"

// hookMarkerES is the marker written before the English migration. It stays a
// read-only alias: the hook is still odachi's and gets rewritten in place, so an
// existing project does not have to paste the lines by hand.
const hookMarkerES = "# odachi: no se commitea sin prueba de este árbol"

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
	// Global says this run installs the door of the home. It is opt-in: an init
	// without the flag does not write the home, because a primary agent in every
	// folder is the human's decision and not an effect of initializing a project.
	Global bool
	// NoProjectAgents skips the project's native agent files, so a project that
	// keeps its agents in the home carries only the rules and the plan. It needs
	// Global: skipping both halves would leave no agent at all.
	NoProjectAgents bool
	// Skills is where this run was asked to put the craft. Empty means "the
	// stored one, or the project": a plain init after `--skills home` must not
	// rebuild the project copy (the same rule 0081 gave the agent choice).
	Skills string
	// DryRun is the rehearsal: the run reads the machine and reports every file
	// it would write, and writes none. Only `setup` sets it -the verb a machine
	// refresh runs- and an init that ignored it would be a flag that lies.
	DryRun bool
	// Existing says this repository already has a commit: it is a project that
	// already lives, and adopting odachi there proposes the minimal footprint -
	// the rules and the hook here, the craft in the machine. It is a fact the CLI
	// reads from git, never a guess: a folder with no history still gets the full
	// install, because there the project is starting now.
	Existing bool
}

// AgentSelection is the selection port: the choice a project recorded, when
// there is one, and the canonical bytes that write it. It is a port and not a
// read of the TOML here because the domain owns the names and the adapter owns
// the format.
type AgentSelection interface {
	// Written returns the stored choices, or an empty Store when there is no
	// file.
	Written(dir string) (persona.Store, error)
	// Target names the file the selection lives in: the absolute path to write,
	// and the label a human reads. The file is in the clone's store, so the app
	// cannot build it from the project root.
	Target(dir string) (path, label string, err error)
	// Render turns the choices into the bytes of the store.
	Render(store persona.Store) ([]byte, error)
}

// Home is the port that answers where the human lives. The global agent is the
// door of every folder, so `init` writes it once per machine; the app never
// touches os.
type Home interface {
	Home() (string, error)
}

// Profiles is the profile port, one question wide: which profile did the human
// write down, if any? Written returns the profile named in odd/profile.toml
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
	// DryRun says this run was a rehearsal: the files above are the ones it would
	// write. A machine reading the payload cannot tell a plan from a fact
	// otherwise, and the two look identical by design.
	DryRun bool
}

// Run performs one `odachi init`.
func Run(opts Options, files Files, skills Skills) Result {
	r := &runner{dir: opts.Dir, files: files, skills: skills, profiles: opts.Profiles, agentSelection: opts.AgentSelection, home: opts.Home}
	if r.dir == "" {
		r.dir = "."
	}
	if opts.NoProjectAgents && !opts.Global {
		return blocked("skipping the project's agent files needs the home door: without it there would be no agent at all; run odachi init --global")
	}

	hasGit, err := r.files.Exists(filepath.Join(r.dir, gitDir))
	if err != nil {
		return blocked("could not check " + gitDir + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
	}
	if !hasGit {
		// No repository means nowhere to install the hook. Writing only the
		// AGENTS.md half would report success for a job that is half done.
		return blocked("there is no " + gitDir + " in this directory: cannot install the hook " + hookPath + "; run git init and run odachi init again")
	}

	selection, clients, problem := r.resolveAgents(opts.Agents)
	if problem != nil {
		return *problem
	}
	r.selection = selection
	r.clients = clients
	r.existing = opts.Existing
	r.minimal = r.derivesMinimal()
	r.projectAgents = r.resolveProjectAgents(opts)
	r.skillsScope = r.resolveSkillsScope(opts)
	r.skillsSource = r.resolveSkillsSource(opts)

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
	if problem := r.writeAgentStore(opts.Agents != "" || opts.NoProjectAgents || opts.Skills != ""); problem != nil {
		// The hook and the rules block are already in place: report how much of
		// the job is done instead of pretending nothing happened.
		problem.Files = append([]string{}, r.changed...)
		return *problem
	}
	if r.projectAgents == persona.ProjectAgentsNone {
		r.skipProjectPersonas(opts.NoProjectAgents)
	} else if problem := r.writePersonas(); problem != nil {
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
		r.installOpenCodeDefault()
	} else {
		r.lines = append(r.lines, "global door: I did not install it (it is opt-in); odachi init --global puts it in the home and odachi shows up in any folder")
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
			problem := blocked("could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) +
				"; fix the file or run odachi init --agents both")
			return "", nil, &problem
		}
		r.storedSelection = stored.Agents
		r.storedProject = stored.ProjectAgents
		r.storedSkills = stored.Skills
	}

	value := strings.TrimSpace(requested)
	if value == "" {
		value = r.storedSelection
	}
	selection, clients, err := persona.ResolveSelection(value)
	if err != nil {
		problem := blocked("the client selection is not valid: " + oneLine(err.Error()) +
			"; run odachi init --agents both")
		return "", nil, &problem
	}
	return selection, clients, nil
}

// effectiveProject is what a store asks for: the absent key is the mirror in the
// tree, the answer every project had before the key existed. Comparing the
// effective answers keeps a plain run from rewriting an old store only to add a
// line that says what its silence already said.
func effectiveProject(store persona.Store) persona.Store {
	if store.ProjectAgents == "" {
		store.ProjectAgents = persona.ProjectAgentsWrite
	}
	return store
}

// effectiveSkills is the same rule for the other half of the store: an absent
// scope is the project, which is the answer the tool always gave.
func effectiveSkills(store persona.Store) persona.Store {
	if store.Skills == "" {
		store.Skills = persona.SkillsProject
	}
	return store
}

// derivesMinimal is the footprint a repository that already lives gets: the
// project never chose, and none of odachi's folders are here yet, so the rules
// stay in the tree and the craft stays in the machine. A project that already
// carries any of the three has adopted odachi the old way, and an inference does
// not move a project that already has the craft.
func (r *runner) derivesMinimal() bool {
	if !r.existing || r.storedSkills != "" || r.storedProject != "" {
		return false
	}
	for _, rel := range []string{".agents", ".opencode", ".claude"} {
		if found, err := r.files.Exists(filepath.Join(r.dir, rel)); err == nil && found {
			return false
		}
	}
	return true
}

// resolveProjectAgents decides where the native agent files live. The flag wins;
// an explicit --agents asks for the project's own copy, which is the only way out
// of the home-only choice and needs no second flag; otherwise the stored answer
// stands, a repository that already lives keeps them in the machine, and a
// project that never chose keeps the mirror in the tree.
func (r *runner) resolveProjectAgents(opts Options) string {
	switch {
	case opts.NoProjectAgents:
		return persona.ProjectAgentsNone
	case strings.TrimSpace(opts.Agents) != "":
		return persona.ProjectAgentsWrite
	case r.storedProject != "":
		return r.storedProject
	case r.minimal:
		return persona.ProjectAgentsNone
	default:
		return persona.ProjectAgentsWrite
	}
}

// resolveSkillsScope decides where the craft lives, by the same rule: --skills
// wins and is recorded, the stored answer stands when the flag is silent, a
// repository that already lives leaves the craft in the machine, and a project
// that never chose keeps the project copy.
func (r *runner) resolveSkillsScope(opts Options) string {
	switch {
	case strings.TrimSpace(opts.Skills) != "":
		return strings.ToLower(strings.TrimSpace(opts.Skills))
	case r.storedSkills != "":
		return r.storedSkills
	case r.minimal:
		return persona.SkillsHome
	default:
		return persona.SkillsProject
	}
}

// The five answers to "who decided where the craft goes", spelled once so the
// output and the tests name the same thing.
const (
	skillsFromFlag     = "flag"
	skillsFromStore    = "store"
	skillsFromDefault  = "default"
	skillsFromSetup    = "setup"
	skillsFromAdoption = "adoption"
)

// resolveSkillsSource says who decided where the craft goes.
func (r *runner) resolveSkillsSource(opts Options) string {
	switch {
	case strings.TrimSpace(opts.Skills) != "":
		return skillsFromFlag
	case r.storedSkills != "":
		return skillsFromStore
	case r.minimal:
		return skillsFromAdoption
	default:
		return skillsFromDefault
	}
}

// writeAgentStore persists the choice only when someone made it: an explicit
// flag, or a store that already exists. A project that never chose keeps no file
// and both clients, so the default install gains no new artifact.
func (r *runner) writeAgentStore(explicit bool) *Result {
	want := effectiveSkills(effectiveProject(persona.Store{Agents: r.selection, ProjectAgents: r.projectAgents, Skills: r.skillsScope}))
	stored := effectiveSkills(effectiveProject(persona.Store{Agents: r.storedSelection, ProjectAgents: r.storedProject, Skills: r.storedSkills}))
	if !explicit && r.storedSelection == "" {
		return nil
	}
	if stored == want {
		r.lines = append(r.lines, persona.SelectionPath+": "+r.selection+", "+want.ProjectAgents+", "+want.Skills+", unchanged")
		return nil
	}
	if r.agentSelection == nil {
		// A caller that wired no store (the tests that only care about the
		// project) has nowhere to record the choice; the run still honors it.
		return nil
	}

	data, err := r.agentSelection.Render(want)
	if err != nil {
		problem := blocked("could not build " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; run odachi init --agents " + r.selection + " again")
		return &problem
	}

	path, label, err := r.agentSelection.Target(r.dir)
	if err != nil {
		problem := blocked("could not find the clone's store: " + oneLine(err.Error()) +
			"; run odachi init --agents " + r.selection + " again")
		return &problem
	}
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("could not check " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; check the permissions and run odachi init --agents " + r.selection + " again")
		return &problem
	}
	if exists {
		if _, err := r.files.Backup(path); err != nil {
			problem := blocked("could not back up " + persona.SelectionPath + " before rewriting it: " + oneLine(err.Error()) +
				"; check the permissions and run odachi init --agents " + r.selection + " again")
			return &problem
		}
	}
	if err := r.files.Write(path, data); err != nil {
		problem := blocked("could not write " + persona.SelectionPath + ": " + oneLine(err.Error()) +
			"; check the permissions and run odachi init --agents " + r.selection + " again")
		return &problem
	}

	r.changed = append(r.changed, label)
	r.lines = append(r.lines, label+": saved the choice "+r.selection+"; the other clients files are not touched")
	return nil
}

// pathWarning is the half-broken install: the hook works from the absolute path,
// the rules do not work at all. It names the folder to add, because "put it on
// the PATH" is the kind of advice that leaves a beginner exactly where they were.
func pathWarning(exe string) []string {
	lines := []string{
		"odachi is not on your PATH: the AGENTS.md block asks the agent to run `odachi`, and it will not be able to",
	}
	if dir := filepath.Dir(exe); exe != "" && dir != "." && dir != string(filepath.Separator) {
		return append(lines, "the hook still works because it uses the absolute path; add that folder to the PATH (in sh: export PATH=\""+dir+":$PATH\")")
	}
	return append(lines, "the hook still works because it uses the absolute path; add the folder where you installed odachi to the PATH and run odachi init again")
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
	// storedProject is where the project said its agent lives, and projectAgents
	// is what this run decided: the flag wins, an explicit client choice asks for
	// the mirror, and otherwise the stored answer stands. Empty means the project
	// never chose, which keeps the mirror in the tree.
	storedProject string
	projectAgents string
	// storedSkills and skillsScope are the same pair for the craft: where the
	// store said the skills live and what this run decided. skillsSource says who
	// decided, so the output names where the answer came from.
	storedSkills string
	skillsScope  string
	skillsSource string
	// dryRun is the rehearsal: every writer still compares, and the lines say
	// "would write" because no byte moved. It is the runner's job and not the
	// Files port's: a port that silently swallowed a write would hide the report.
	dryRun bool
	// existing says the repository already has a commit, and minimal is the
	// answer this run derived from it: the project never chose, nothing of
	// odachi's footprint is here yet, and the machine is where the craft lives.
	existing bool
	minimal  bool
	// absent is what detection did not find: a client whose config folder is not
	// in this home. It is reported, because an installer that stays quiet about a
	// client looks like it forgot one.
	absent []string
	// unchecked is what detection could not look at: a read that failed, which is
	// not the same thing as a client that is not here.
	unchecked []string
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
		problem := blocked("could not check " + hookPath + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
		return &problem
	}

	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("could not read " + hookPath + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
			return &problem
		}

		if !strings.Contains(string(data), hookMarker) && !strings.Contains(string(data), hookMarkerOld) && !strings.Contains(string(data), hookMarkerES) {
			problem := foreignHook(exe)
			return &problem
		}
		if string(data) == hook {
			r.lines = append(r.lines, hookPath+": unchanged")
			return nil
		}

		if _, err := r.files.Backup(path); err != nil {
			problem := blocked("could not back up " + hookPath + " before rewriting it: " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
			return &problem
		}
	}

	if err := r.files.Write(path, []byte(hook)); err != nil {
		problem := blocked("could not write " + hookPath + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
		return &problem
	}
	if err := r.files.MakeExecutable(path); err != nil {
		problem := blocked("could not make executable " + hookPath + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init again")
		return &problem
	}

	r.changed = append(r.changed, hookPath)
	r.lines = append(r.lines, hookPath+": installed")
	return nil
}

// foreignHook is the refusal: the three lines to paste, and nothing touched.
func foreignHook(exe string) Result {
	message := "there is already " + hookPath + " and it is not odachi's: I did not overwrite it; paste these lines inside your hook and run odachi init again"
	lines := []string{message, ""}
	lines = append(lines, strings.Split(strings.TrimRight(hookFor(exe), "\n"), "\n")...)
	lines = append(lines, "", "next: odachi init")
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
	// The rehearsal says so first, before any destination: a line list that opens
	// with "would write" and never says "dry run" reads like a report of work done.
	if r.dryRun {
		notice := "dry run: this is what odachi setup would write, and it writes nothing"
		if len(r.changed) == 0 && len(r.skillFiles) == 0 && len(r.personaFiles) == 0 {
			notice = "dry run: nothing to write, this machine is already current"
		}
		r.lines = append([]string{notice}, r.lines...)
	}

	files := append(append(append([]string{}, r.changed...), r.skillFiles...), r.personaFiles...)

	if len(r.changed) == 0 && len(r.skillFiles) == 0 && len(r.personaFiles) == 0 {
		message := "unchanged: " + agentsFile + ", " + ignoreFile + ", the hook " + hookPath + " and the native agent were already up to date"
		if len(r.skillLabels) > 0 {
			message += "; the skills of " + humanList(r.skillLabels) + " as well"
		}
		if r.dryRun {
			message = "dry run: nothing to write, everything is already current"
			if len(r.skillLabels) > 0 {
				message = "dry run: nothing to write, everything is already current (" + humanList(r.skillLabels) + " included)"
			}
		}
		return Result{
			Code:         r.code(),
			Message:      r.withBlocked(message),
			Next:         r.skillNext,
			Lines:        r.lines,
			SkillDests:   r.skillDests,
			PersonaDests: r.personaDests,
			DryRun:       r.dryRun,
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
		written = append(written, "the skills ("+humanList(r.skillLabels)+")")
	}
	if len(r.personaFiles) > 0 {
		written = append(written, "the native agent ("+humanList(r.personaLabels)+")")
	}
	message := "done: " + humanList(written) + " are up to date"
	if r.dryRun {
		message = "dry run: would write " + humanList(written) + "; nothing was touched"
	}

	lines := append([]string{}, r.lines...)
	lines = append(lines, r.legacyMoves()...)
	// The human does not commit: the agent does. Saying "commit" here assumed the
	// one thing this tool exists to remove from their plate.
	lines = append(lines, "next: "+r.next())

	return Result{
		Code:         r.code(),
		Message:      r.withBlocked(message),
		Next:         r.skillNext,
		Lines:        lines,
		Files:        files,
		SkillDests:   r.skillDests,
		PersonaDests: r.personaDests,
		DryRun:       r.dryRun,
	}
}

// next is the command that finishes what this run started. A rehearsal's next
// step is the same command without the flag, because the point of a dry run is
// the run that follows it.
func (r *runner) next() string {
	if r.skillNext != "" {
		return r.skillNext
	}
	if r.dryRun {
		return "odachi setup"
	}
	return "check " + agentsFile + "; when odachi check --quiet is green, the agent saves the progress"
}

// legacyMoves names the rules an older release kept in .odachi/ and the exact git
// command that moves each one. init prints them and never runs them: a tracked
// file belongs to git, and the human commits the move.
func (r *runner) legacyMoves() []string {
	present := []string{}
	for _, to := range layout.LegacyRules[0:3] {
		rel := filepath.ToSlash(filepath.Join(layout.LegacyDir, filepath.Base(to)))
		if found, err := r.files.Exists(filepath.Join(r.dir, filepath.FromSlash(rel))); err == nil && found {
			present = append(present, rel)
		}
	}
	if found, err := r.files.Exists(filepath.Join(r.dir, filepath.FromSlash(layout.LegacyDir), filepath.Base(layout.Roles))); err == nil && found {
		present = append(present, filepath.ToSlash(filepath.Join(layout.LegacyDir, filepath.Base(layout.Roles))))
	}

	moves := layout.LegacyMoves(present)
	if len(moves) == 0 {
		return nil
	}
	return append([]string{"the rules an older release kept in " + layout.LegacyDir + "/ move with git (I do not touch a tracked file):"}, moves...)
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
		Lines:   []string{message, "next: odachi init"},
	}
}

// humanList joins two or three paths the way an English sentence does.
func humanList(items []string) string {
	switch len(items) {
	case 0:
		return "nothing"
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
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

// skipProjectPersonas is the clean mode: the project keeps the rules and the
// plan, and the agents it would have written stay in the home, where the door
// and its subagents are the ones that run.
func (r *runner) skipProjectPersonas(byFlag bool) {
	reason := "the store says the agent lives in the home"
	if byFlag {
		reason = "--no-project-agents"
	}
	if r.minimal && r.storedProject == "" {
		reason = "this repository already lives, so the agent stays in the machine"
	}
	r.lines = append(r.lines, "project agents: not written ("+reason+"); the door in the home is the one that runs, and --agents "+r.selection+" puts the mirrors back")
	// The minimal footprint skips both halves here, so the machine is the only
	// place the agent can come from. Saying it is not advice: without setup the
	// project has no agent at all.
	if r.minimal {
		r.lines = append(r.lines, "if this machine never ran odachi setup, do it once: it installs the door and the craft the whole machine reads")
	}
}
