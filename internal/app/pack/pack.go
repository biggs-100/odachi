// Package pack implements `odachi pack`: it copies the skills bundled in the
// binary into the skills directory of each agent the human uses.
//
// It is pure orchestration over two ports: Files (the disk) and Home (where the
// human's home is). It never imports os, so the whole behaviour — including
// "dry-run writes nothing" — is observable at this seam with an in-memory
// double. The project root arrives as an option, because a project-scoped agent
// reads its skills from inside the project and there is nothing to resolve.
//
// Three rules shape it, and they come from Contract VIII:
//
//   - The destination table is data in internal/domain/agent, never a branch of
//     code, because every path in it was verified and a guess does not ship.
//     The table also carries the scope: a `home` agent resolves against the
//     home, a `project` agent against the project.
//   - The destination is a decision with two escape hatches, so a client this
//     table has never heard of is still served: `--scope` installs a known
//     agent under the other root, and `--dest` installs every named agent into
//     one directory the human typed. The table is a default, not a cage.
//   - A configured file that differs is backed up before it is replaced when the
//     destination is inside the project, where the clone's store and `undo` live;
//     outside it (the home, a --dest) it is replaced and reported as `written`.
//     An identical file is reported as `unchanged` and not rewritten.
//   - Nothing is ever deleted or moved. A skill this tool did not write is
//     reported with its path and left exactly as it was.
package pack

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/agent"
)

// Exit codes: the same two every verb uses.
const (
	// CodeOK means every requested file is in place or already was.
	CodeOK = 0
	// CodeBlocked means a refused agent, an unusable home, or a failed write.
	CodeBlocked = 1
)

// SkillFile is the one file every skill must carry. The agents read exactly this
// name.
const SkillFile = "SKILL.md"

// File is one file of a skill folder, at a path relative to it: "SKILL.md" or
// "references/numbers.md", with forward slashes on every OS.
type File struct {
	Path    string
	Content []byte
}

// Skill is one skill the human gets: its folder name and every file in it,
// sorted by path. A skill is installed as a folder, never as a lone file: the
// references travel with it or the skill links to nothing.
type Skill struct {
	Name  string
	Files []File
}

// Options is one `odachi pack` invocation.
type Options struct {
	// Agents is the comma-separated --agents flag. Empty means "the agents
	// whose config directory exists here", which is what "detected" means —
	// unless --dest is set, where it means "detect nothing" and is refused.
	Agents string
	// Scope is the --scope flag: project or home. Empty keeps each agent's own
	// default scope, and a value installs every named agent under that root.
	Scope string
	// Dest is the --dest flag: one directory for every named agent, absolute or
	// relative to the project. It replaces the table's destination entirely, so
	// the name stops being a lookup and becomes a label.
	Dest string
	// DryRun reports what would happen and writes nothing.
	DryRun bool
	// Dir is the project root a project-scoped agent resolves against, and the
	// root a relative --dest is joined to. Empty means the current directory.
	Dir string
	// Skills is the bundled payload, wired by the CLI from the embedded bytes.
	Skills []Skill
	// Project is the project's own catalog, read from ./skills by the CLI. A
	// project skill with a bundled name replaces it: the bundled one is a
	// starter, the project copy is the project's decision.
	Project []Skill
}

// Files is the disk port: enough to install a file and never lose one.
type Files interface {
	Exists(path string) (bool, error)
	List(dir string) ([]string, error)
	Read(path string) ([]byte, error)
	Backup(path string) (string, error)
	Write(path string, data []byte) error
}

// Home is the port that resolves the human's home directory. The app resolves
// it once; the domain table says where inside it each agent looks.
type Home interface {
	Home() (string, error)
}

// Failure is one file a destination could not write, with the reason the adapter
// gave. It is kept as fields so a caller that composes its own report — `odachi
// init` says which half of the job is pending — does not have to parse a line
// back into a path and a reason.
type Failure struct {
	Path   string
	Reason string
}

// DestResult is what one destination got. The counts are what the summary line
// is built from, so the line and the JSON can never disagree. Names is the label
// a human reads: one agent, or every agent that shares this destination because
// --dest put them together.
type DestResult struct {
	Names     []string
	Dest      string
	Scope     string
	Installed int
	BackedUp  int
	Written   int
	Unchanged int
	Foreign   int
	Failed    int
	// Failure is the first file this destination could not write, or nil.
	Failure *Failure
	Lines   []string
	Files   []string
}

// Result is what the CLI prints and exits with. Files carries the display paths
// this run wrote, or would write under --dry-run. Dests carries one entry per
// destination, never one per agent: a summary that repeats the same folder is a
// summary a human has to decode.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
	Dests   []DestResult
	DryRun  bool
}

// Run performs one `odachi pack`.
func Run(opts Options, files Files, home Home) Result {
	skills, replaced := merge(opts.Skills, opts.Project)
	if err := validate(skills); err != nil {
		return blocked(err.Error())
	}
	if len(skills) == 0 {
		return blocked("there are no skills bundled in the binary: this is an incomplete install; download odachi again")
	}

	// The flags are checked before anything is read or written: --dest and --scope
	// both answer "where", and two answers are no answer.
	if err := checkFlags(opts); err != nil {
		return blocked(err.Error())
	}

	scope, err := agent.ParseScope(opts.Scope)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}
	dest := strings.TrimSpace(opts.Dest)

	// With --dest the human typed the path their client reads, so the names are
	// labels and the table is not consulted at all — a client odachi has never
	// heard of is exactly the case this flag exists for. Without it the table
	// resolves every name and --scope, if any, moves it to the other root.
	var labels []string
	var selected []agent.Agent
	if dest != "" {
		labels = agent.Labels(opts.Agents)
	} else {
		selected, err = agent.ResolveAll(opts.Agents)
		if err != nil {
			return blocked(oneLine(err.Error()))
		}
		for i, a := range selected {
			selected[i] = a.In(scope)
		}
	}

	roots := roots{project: strings.TrimSpace(opts.Dir)}
	if roots.project == "" {
		roots.project = "."
	}

	// The home is asked for only when a destination lives in it: an install that
	// never touches the home must not fail because it is unavailable, and it must
	// not read the home at all. An empty selection means detection, which looks at
	// both roots.
	if dest == "" && (len(selected) == 0 || anyHomeScoped(selected)) {
		root, err := home.Home()
		if err != nil {
			return blocked("could not find your home: " + oneLine(err.Error()) + "; set HOME (or USERPROFILE on Windows) and run odachi pack again")
		}
		root = strings.TrimSpace(root)
		if root == "" {
			return blocked("your home came back empty; set HOME (or USERPROFILE on Windows) and run odachi pack again")
		}
		roots.home = root
	}

	if dest == "" && len(selected) == 0 {
		selected, err = detect(roots, files)
		if err != nil {
			return blocked(err.Error())
		}
		if len(selected) == 0 {
			message := "no agent detected: name the ones you use with odachi pack --agents claude,codex, or install into the folder every client reads with odachi pack --agents standard (supported: " + agent.List() + ")"
			return Result{Code: CodeOK, Message: message, Lines: []string{message}, DryRun: opts.DryRun}
		}
	}

	targets := targetsFor(selected, labels, dest, roots)

	out := Result{Code: CodeOK, DryRun: opts.DryRun}
	// The replacements come first: they say which bytes the agent will read
	// before the destinations say where they went.
	out.Lines = append(out.Lines, replaced...)
	summaries := make([]string, 0, len(targets))
	dests := make([]string, 0, len(targets))
	for _, t := range targets {
		result := install(t, skills, files, opts.DryRun)
		out.Dests = append(out.Dests, result)
		out.Lines = append(out.Lines, result.Lines...)
		out.Files = append(out.Files, result.Files...)
		summaries = append(summaries, result.summary(opts.DryRun))
		dests = append(dests, t.label)
		if result.Failed > 0 {
			out.Code = CodeBlocked
		}
	}
	out.Lines = append(out.Lines, summaries...)

	switch {
	case out.Code != CodeOK:
		out.Message = "could not install everything: " + firstFailure(out.Lines)
	case opts.DryRun:
		out.Message = "--dry-run: wrote nothing; this is what it would do in " + strings.Join(dests, ", ")
	default:
		out.Message = fmt.Sprintf("ready: %d %s in %s", len(skills), plural(len(skills), "skill", "skills"), strings.Join(dests, ", "))
	}

	// A dry run says so on the first line, before anyone can mistake the plan
	// for a result.
	if out.Code == CodeOK && opts.DryRun {
		out.Lines = append([]string{out.Message}, out.Lines...)
	}
	return out
}

// roots are the two directories an agent's skills directory can live in: the
// home a human owns, and the project being worked on.
type roots struct {
	home    string
	project string
}

// base is the root an agent's destination belongs to. The table says the scope;
// this is the only place that turns it into a directory.
func (r roots) base(a agent.Agent) string {
	if a.Scope == agent.ScopeProject {
		return r.project
	}
	return r.home
}

// anyHomeScoped reports whether one of these agents needs the home at all.
func anyHomeScoped(agents []agent.Agent) bool {
	for _, a := range agents {
		if a.Scope != agent.ScopeProject {
			return true
		}
	}
	return false
}

// checkFlags refuses the combinations that have no single answer, each with the
// flag that wins where. Nothing is read or written before this passes.
func checkFlags(opts Options) error {
	dest := strings.TrimSpace(opts.Dest)
	scope := strings.TrimSpace(opts.Scope)
	named := strings.TrimSpace(opts.Agents) != ""

	if dest != "" && scope != "" {
		return errors.New("--dest and --scope do not go together: choose one; --dest already says where to install (odachi pack --agents claude --dest ./skills-x) and --scope only works with the table (odachi pack --agents opencode --scope home)")
	}
	if !named && dest != "" {
		return errors.New("--dest needs you to name the agents: odachi pack --agents claude --dest ./skills-x")
	}
	if !named && scope != "" {
		return errors.New("--scope needs you to name the agents: odachi pack --agents opencode --scope home")
	}
	return nil
}

// target is one directory `pack` installs into, with the agents that read it.
// Several agents can share one target — --dest puts every named agent in the same
// folder — and the report names the destination once, not once per agent.
type target struct {
	names []string
	dir   string // the directory on disk: absolute, or relative to the project
	label string // what a human reads: ~/..., ./..., or the --dest path they typed
	scope string
	// insideProject says whether the clone's store can hold a backup of a file
	// here and `undo` can read it. The home and a --dest outside the project have
	// no such store, so a replaced file there is written without a sidecar.
	insideProject bool
}

// targetsFor turns the selection into the directories to install into. With
// --dest there is exactly one, and its names are labels rather than lookups.
// Without it every agent keeps its own destination under the root its scope
// names, and two agents that resolve to the same directory share one target.
func targetsFor(selected []agent.Agent, labels []string, dest string, roots roots) []target {
	if dest != "" {
		dir := filepath.FromSlash(dest)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(roots.project, dir)
		}
		return []target{{names: labels, dir: dir, label: destLabel(dest), insideProject: withinProject(roots.project, dir)}}
	}

	targets := make([]target, 0, len(selected))
	at := make(map[string]int, len(selected))
	for _, a := range selected {
		dir := filepath.Join(roots.base(a), filepath.FromSlash(a.Dest))
		if i, ok := at[dir]; ok {
			targets[i].names = append(targets[i].names, a.Name)
			continue
		}
		at[dir] = len(targets)
		targets = append(targets, target{
			names:         []string{a.Name},
			dir:           dir,
			label:         a.Display(),
			scope:         string(a.Scope),
			insideProject: withinProject(roots.project, dir),
		})
	}
	return targets
}

// withinProject reports whether dir lives under the project root. That is the
// only place a backup can go: the store hangs from the clone and `undo` reads it
// from there, so a file outside it has nowhere to be undone.
func withinProject(project, dir string) bool {
	if strings.TrimSpace(project) == "" {
		return false
	}
	root, err := filepath.Abs(project)
	if err != nil {
		return false
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// destLabel is how a human reads the --dest path. An absolute path stays as
// typed — the human wrote it, and it is not a path this tool has to hide — and a
// relative one is shown from the project, the way every other project path is.
func destLabel(dest string) string {
	clean := filepath.ToSlash(filepath.Clean(dest))
	if filepath.IsAbs(dest) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return clean
	}
	return "./" + strings.TrimPrefix(clean, "./")
}

// install puts every bundled skill where one destination expects them, and
// reports what it found there besides its own files. A skill is one line, no
// matter how many references it carries: the summary counts skills, not files.
func install(t target, skills []Skill, files Files, dryRun bool) DestResult {
	result := DestResult{Names: t.names, Dest: t.label, Scope: t.scope}
	// who is the prefix every line carries: one agent, or every agent that shares
	// this destination.
	who := strings.Join(t.names, ", ")

	seen := make(map[string]bool, len(skills))
	for _, skill := range skills {
		seen[skill.Name] = true
		result.installSkill(t, who, skill, files, dryRun)
	}

	// A skill odachi did not write belongs to the human. It is reported, never
	// touched: deletion is not this command's job.
	names, err := files.List(t.dir)
	if err != nil {
		result.fail(who, t.label, err)
		return result
	}
	for _, name := range names {
		if seen[name] {
			continue
		}
		label := t.label + "/" + name + "/" + SkillFile
		exists, err := files.Exists(filepath.Join(t.dir, name, SkillFile))
		if err != nil {
			result.fail(who, label, err)
			continue
		}
		if !exists {
			// An entry that is not a skill (a stray file, a directory without
			// SKILL.md) is not this command's business either.
			continue
		}
		result.Foreign++
		result.Lines = append(result.Lines, who+" · odachi did not write it, not touching it · "+label)
	}
	return result
}

// installSkill writes one skill folder, file by file, each one backed up before
// it is replaced. The skill is reported once: installed when its SKILL.md was not
// there, unchanged when every file matches, backed up and written when a file the
// human had is replaced, written when the skill only gained a file.
func (r *DestResult) installSkill(t target, who string, skill Skill, files Files, dryRun bool) {
	label := t.label + "/" + skill.Name + "/"
	mainExisted := false
	identical := true
	backedUp := false

	for _, file := range skill.Files {
		path := filepath.Join(t.dir, skill.Name, filepath.FromSlash(file.Path))
		display := label + file.Path

		exists, err := files.Exists(path)
		if err != nil {
			r.fail(who, display, err)
			return
		}
		if file.Path == SkillFile {
			mainExisted = exists
		}
		if exists {
			current, err := files.Read(path)
			if err != nil {
				r.fail(who, display, err)
				return
			}
			if bytes.Equal(current, file.Content) {
				continue
			}
		}

		identical = false
		r.Files = append(r.Files, display)
		// A backup is undo material, and undo lives inside a project: the clone's
		// store can hold it and `undo` can read it. Outside the project (the home,
		// a --dest) there is no store to read it from, so the file is replaced and
		// reported as written. A file that was not there needs none, and a skill
		// that only gained one is never reported as backed up (0207).
		if exists && t.insideProject {
			backedUp = true
			if !dryRun {
				if _, err := files.Backup(path); err != nil {
					r.fail(who, display, err)
					return
				}
			}
		}
		if dryRun {
			continue
		}
		if err := files.Write(path, file.Content); err != nil {
			r.fail(who, display, err)
			return
		}
	}

	switch {
	case identical:
		r.Unchanged++
		r.Lines = append(r.Lines, who+" · unchanged · "+label)
	case !mainExisted:
		r.Installed++
		verb := "installed"
		if dryRun {
			verb = "would install"
		}
		r.Lines = append(r.Lines, who+" · "+verb+" · "+label)
	case backedUp:
		r.BackedUp++
		verb := "backed up and written"
		if dryRun {
			verb = "would back up and write"
		}
		r.Lines = append(r.Lines, who+" · "+verb+" · "+label)
	default:
		r.Written++
		verb := "written"
		if dryRun {
			verb = "would write"
		}
		r.Lines = append(r.Lines, who+" · "+verb+" · "+label)
	}
}

// fail records one failed file and keeps going: one unreadable directory must
// not cost the human the other forty-nine files. The first failure is also kept
// as fields, because it is the one a caller quotes.
func (r *DestResult) fail(who, label string, err error) {
	reason := oneLine(err.Error())
	r.Failed++
	if r.Failure == nil {
		r.Failure = &Failure{Path: label, Reason: reason}
	}
	r.Lines = append(r.Lines, who+" · failed · "+label+": "+reason)
}

// summary is the one line Contract VIII fixes: `claude · 5 installed ·
// ~/.claude/skills`. Only the counts that happened show up, so a quiet second
// run says `5 unchanged` and nothing else. The destination is printed once: two
// agents that share a folder share a line.
func (r DestResult) summary(dryRun bool) string {
	parts := []string{strings.Join(r.Names, ", ")}
	add := func(n int, one, many string) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, plural(n, one, many)))
		}
	}

	switch {
	case dryRun && r.Installed > 0:
		parts = append(parts, fmt.Sprintf("%d to install", r.Installed))
	default:
		add(r.Installed, "installed", "installed")
	}

	switch {
	case dryRun && r.BackedUp > 0:
		parts = append(parts, fmt.Sprintf("%d to back up", r.BackedUp))
	default:
		add(r.BackedUp, "backed up", "backed up")
	}

	switch {
	case dryRun && r.Written > 0:
		parts = append(parts, fmt.Sprintf("%d to write", r.Written))
	default:
		add(r.Written, "written", "written")
	}

	add(r.Unchanged, "unchanged", "unchanged")
	add(r.Failed, "failed", "failed")

	parts = append(parts, r.Dest)
	return strings.Join(parts, " · ")
}

// detect returns the agents installed here: the ones a human actually uses in
// this home or this project. A home agent is here when its config directory
// exists in the home; a project agent when its config directory exists in the
// project. An agent that is not installed is not invented into existence.
func detect(roots roots, files Files) ([]agent.Agent, error) {
	found := make([]agent.Agent, 0, len(agent.Supported()))
	for _, a := range agent.Supported() {
		marker := filepath.Join(roots.base(a), filepath.FromSlash(a.Marker()))
		exists, err := files.Exists(marker)
		if err != nil {
			return nil, fmt.Errorf("could not check %s: %s; check the permissions and run odachi pack --agents %s again", a.Display(), oneLine(err.Error()), a.Name)
		}
		if exists {
			found = append(found, a)
		}
	}
	return found, nil
}

// normalize trims, sorts and de-duplicates the payload, so the summary and the
// write order are the same on every run.
func normalize(skills []Skill) []Skill {
	out := make([]Skill, 0, len(skills))
	seen := make(map[string]bool, len(skills))
	for _, skill := range skills {
		name := strings.TrimSpace(skill.Name)
		if seen[name] {
			continue
		}
		seen[name] = true
		files := append([]File(nil), skill.Files...)
		slices.SortStableFunc(files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
		out = append(out, Skill{Name: name, Files: files})
	}
	slices.SortStableFunc(out, func(a, b Skill) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// merge puts the project's skills on top of the bundled ones. A project skill
// with a bundled name replaces it, and the replacement is named: the bundled
// skill is a starter, the project copy is the project's decision, and a human
// has to know which bytes the agent will read.
func merge(bundled, project []Skill) ([]Skill, []string) {
	merged := normalize(bundled)
	if len(project) == 0 {
		return merged, nil
	}

	byName := make(map[string]int, len(merged))
	for i, skill := range merged {
		byName[skill.Name] = i
	}

	lines := make([]string, 0, len(project))
	for _, skill := range normalize(project) {
		if i, ok := byName[skill.Name]; ok {
			merged[i] = skill
			lines = append(lines, "project skill "+skill.Name+" replaces the bundled one")
			continue
		}
		byName[skill.Name] = len(merged)
		merged = append(merged, skill)
	}

	slices.SortStableFunc(merged, func(a, b Skill) int { return strings.Compare(a.Name, b.Name) })
	return merged, lines
}

// validate refuses a payload that could write outside the agent's directory.
// The embed is fixed at build time, so this can only fire on a broken build —
// and a broken build must not be allowed to write.
func validate(skills []Skill) error {
	for _, skill := range skills {
		name := skill.Name
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
			return fmt.Errorf("invalid bundled skill: %q; the binary is built wrong, download odachi again", name)
		}
		if len(skill.Files) == 0 {
			return fmt.Errorf("invalid bundled skill %q: it carries no files; the binary is built wrong, download odachi again", name)
		}
		main := false
		for _, file := range skill.Files {
			if file.Path == SkillFile {
				main = true
			}
			if !validFilePath(file.Path) {
				return fmt.Errorf("invalid bundled file %q in skill %q: a path must stay inside the folder; the binary is built wrong, download odachi again", file.Path, name)
			}
		}
		if !main {
			return fmt.Errorf("invalid bundled skill %q: it carries no %s; the binary is built wrong, download odachi again", name, SkillFile)
		}
	}
	return nil
}

// validFilePath keeps every write inside the skill folder: a relative path with
// forward slashes, SKILL.md, a markdown reference under references/, or a flat
// file under assets/.
func validFilePath(rel string) bool {
	if rel == "" || strings.Contains(rel, `\`) || strings.HasPrefix(rel, "/") {
		return false
	}
	if path.Clean(rel) != rel || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return false
	}
	if rel == SkillFile || strings.HasPrefix(rel, "references/") {
		return true
	}
	return strings.HasPrefix(rel, "assets/") && !strings.Contains(strings.TrimPrefix(rel, "assets/"), "/")
}

// firstFailure is the failure the one-line message quotes, so a blocked run
// names an agent and a path even when it only prints one line.
func firstFailure(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, " · failed · ") {
			return line
		}
	}
	if len(lines) > 0 {
		return lines[0]
	}
	return "could not write the skills"
}

// blocked is a refused run: one English line that says what to do next.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

// plural picks the form a human reads: "1 installed", "5 installed".
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// oneLine keeps every message to a single line, whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
