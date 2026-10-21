// Package pack implements `odachi pack`: it copies the skills bundled in the
// binary into the skills directory of each agent the human uses.
//
// It is pure orchestration over two ports: Files (the disk) and Home (where the
// human's home is). It never imports os, so the whole behaviour — including
// "dry-run writes nothing" — is observable at this seam with an in-memory
// double.
//
// Three rules shape it, and they come from Contract VIII:
//
//   - The destination table is data in internal/domain/agent, never a branch of
//     code, because every path in it was verified and a guess does not ship.
//   - A configured file that differs is backed up before it is replaced; an
//     identical file is reported as `sin cambios` and not rewritten.
//   - Nothing is ever deleted or moved. A skill this tool did not write is
//     reported with its path and left exactly as it was.
package pack

import (
	"bytes"
	"fmt"
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

// SkillFile is the one file a skill is. The agents read exactly this name.
const SkillFile = "SKILL.md"

// Skill is one bundled skill: its name (its directory) and its exact bytes.
type Skill struct {
	Name    string
	Content []byte
}

// Options is one `odachi pack` invocation.
type Options struct {
	// Agents is the comma-separated --agents flag. Empty means "the agents
	// whose config directory exists here", which is what "detectado" means.
	Agents string
	// DryRun reports what would happen and writes nothing.
	DryRun bool
	// Skills is the bundled payload, wired by the CLI from the embedded bytes.
	Skills []Skill
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

// AgentResult is what one agent got. The counts are what the summary line is
// built from, so the line and the JSON can never disagree.
type AgentResult struct {
	Name      string
	Dest      string
	Installed int
	BackedUp  int
	Unchanged int
	Foreign   int
	Failed    int
	Lines     []string
	Files     []string
}

// Result is what the CLI prints and exits with. Files carries the display paths
// this run wrote, or would write under --dry-run.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
	Agents  []AgentResult
	DryRun  bool
}

// Run performs one `odachi pack`.
func Run(opts Options, files Files, home Home) Result {
	skills := normalize(opts.Skills)
	if err := validate(skills); err != nil {
		return blocked(err.Error())
	}
	if len(skills) == 0 {
		return blocked("no hay skills incluidas en el binario: es una instalación incompleta; volvé a bajar odachi")
	}

	// The agent list is resolved before anything is read or written: a typo in
	// --agents refuses the whole run, so it can never leave half a payload.
	selected, err := agent.ResolveAll(opts.Agents)
	if err != nil {
		return blocked(oneLine(err.Error()))
	}

	root, err := home.Home()
	if err != nil {
		return blocked("no pude encontrar tu home: " + oneLine(err.Error()) + "; definí HOME (o USERPROFILE en Windows) y volvé a correr odachi pack")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return blocked("tu home llegó vacío; definí HOME (o USERPROFILE en Windows) y volvé a correr odachi pack")
	}

	if len(selected) == 0 {
		selected, err = detect(root, files)
		if err != nil {
			return blocked(err.Error())
		}
	}
	if len(selected) == 0 {
		message := "no detecté ningún agente: nombrá los que usás con odachi pack --agents claude,codex (los soportados son: " + agent.List() + ")"
		return Result{Code: CodeOK, Message: message, Lines: []string{message}, DryRun: opts.DryRun}
	}

	out := Result{Code: CodeOK, DryRun: opts.DryRun}
	summaries := make([]string, 0, len(selected))
	names := make([]string, 0, len(selected))
	for _, a := range selected {
		result := install(root, a, skills, files, opts.DryRun)
		out.Agents = append(out.Agents, result)
		out.Lines = append(out.Lines, result.Lines...)
		out.Files = append(out.Files, result.Files...)
		summaries = append(summaries, result.summary(opts.DryRun))
		names = append(names, a.Name)
		if result.Failed > 0 {
			out.Code = CodeBlocked
		}
	}
	out.Lines = append(out.Lines, summaries...)

	switch {
	case out.Code != CodeOK:
		out.Message = "no pude instalar todo: " + firstFailure(out.Lines)
	case opts.DryRun:
		out.Message = "--dry-run: no escribí nada; esto es lo que haría en " + strings.Join(names, ", ")
	default:
		out.Message = fmt.Sprintf("listo: %d %s en %s", len(skills), plural(len(skills), "skill", "skills"), strings.Join(names, ", "))
	}

	// A dry run says so on the first line, before anyone can mistake the plan
	// for a result.
	if out.Code == CodeOK && opts.DryRun {
		out.Lines = append([]string{out.Message}, out.Lines...)
	}
	return out
}

// install puts every bundled skill where one agent looks, and reports what it
// found there besides its own files.
func install(root string, a agent.Agent, skills []Skill, files Files, dryRun bool) AgentResult {
	result := AgentResult{Name: a.Name, Dest: a.Display()}
	destDir := filepath.Join(root, filepath.FromSlash(a.Dest))

	seen := make(map[string]bool, len(skills))
	for _, skill := range skills {
		seen[skill.Name] = true
		path := filepath.Join(destDir, skill.Name, SkillFile)
		label := a.Display() + "/" + skill.Name + "/" + SkillFile

		exists, err := files.Exists(path)
		if err != nil {
			result.fail(a.Name, label, err)
			continue
		}

		if !exists {
			if dryRun {
				result.Installed++
				result.Lines = append(result.Lines, a.Name+" · instalaría · "+label)
				continue
			}
			if err := files.Write(path, skill.Content); err != nil {
				result.fail(a.Name, label, err)
				continue
			}
			result.Installed++
			result.Files = append(result.Files, label)
			result.Lines = append(result.Lines, a.Name+" · instalada · "+label)
			continue
		}

		current, err := files.Read(path)
		if err != nil {
			result.fail(a.Name, label, err)
			continue
		}
		if bytes.Equal(current, skill.Content) {
			result.Unchanged++
			result.Lines = append(result.Lines, a.Name+" · sin cambios · "+label)
			continue
		}

		if dryRun {
			result.BackedUp++
			result.Lines = append(result.Lines, a.Name+" · respaldaría y escribiría · "+label)
			continue
		}
		// The human's bytes go aside before they go away: a backup that can be
		// skipped is not a backup.
		if _, err := files.Backup(path); err != nil {
			result.fail(a.Name, label, err)
			continue
		}
		if err := files.Write(path, skill.Content); err != nil {
			result.fail(a.Name, label, err)
			continue
		}
		result.BackedUp++
		result.Files = append(result.Files, label)
		result.Lines = append(result.Lines, a.Name+" · respaldada y escrita · "+label)
	}

	// A skill odachi did not write belongs to the human. It is reported, never
	// touched: deletion is not this command's job.
	names, err := files.List(destDir)
	if err != nil {
		result.fail(a.Name, a.Display(), err)
		return result
	}
	for _, name := range names {
		if seen[name] {
			continue
		}
		label := a.Display() + "/" + name + "/" + SkillFile
		exists, err := files.Exists(filepath.Join(destDir, name, SkillFile))
		if err != nil {
			result.fail(a.Name, label, err)
			continue
		}
		if !exists {
			// An entry that is not a skill (a stray file, a directory without
			// SKILL.md) is not this command's business either.
			continue
		}
		result.Foreign++
		result.Lines = append(result.Lines, a.Name+" · no lo escribió odachi, no lo toco · "+label)
	}
	return result
}

// fail records one failed file and keeps going: one unreadable directory must
// not cost the human the other forty-nine files.
func (r *AgentResult) fail(agentName, label string, err error) {
	r.Failed++
	r.Lines = append(r.Lines, agentName+" · falló · "+label+": "+oneLine(err.Error()))
}

// summary is the one line Contract VIII fixes: `claude · 5 instaladas ·
// ~/.claude/skills`. Only the counts that happened show up, so a quiet second
// run says `5 sin cambios` and nothing else.
func (r AgentResult) summary(dryRun bool) string {
	parts := []string{r.Name}
	add := func(n int, one, many string) {
		if n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, plural(n, one, many)))
		}
	}

	switch {
	case dryRun && r.Installed > 0:
		parts = append(parts, fmt.Sprintf("%d por instalar", r.Installed))
	default:
		add(r.Installed, "instalada", "instaladas")
	}

	switch {
	case dryRun && r.BackedUp > 0:
		parts = append(parts, fmt.Sprintf("%d por respaldar", r.BackedUp))
	default:
		add(r.BackedUp, "respaldada", "respaldadas")
	}

	add(r.Unchanged, "sin cambios", "sin cambios")
	add(r.Failed, "falló", "fallaron")

	parts = append(parts, r.Dest)
	return strings.Join(parts, " · ")
}

// detect returns the agents whose config directory exists in this home: the
// ones a human actually uses here. An agent that is not installed is not
// invented into existence.
func detect(root string, files Files) ([]agent.Agent, error) {
	found := make([]agent.Agent, 0, len(agent.Supported()))
	for _, a := range agent.Supported() {
		marker := filepath.Join(root, filepath.FromSlash(a.Marker()))
		exists, err := files.Exists(marker)
		if err != nil {
			return nil, fmt.Errorf("no pude revisar %s: %s; revisá los permisos y volvé a correr odachi pack --agents %s", a.Display(), oneLine(err.Error()), a.Name)
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
		out = append(out, Skill{Name: name, Content: skill.Content})
	}
	slices.SortStableFunc(out, func(a, b Skill) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// validate refuses a payload that could write outside the agent's directory.
// The embed is fixed at build time, so this can only fire on a broken build —
// and a broken build must not be allowed to write.
func validate(skills []Skill) error {
	for _, skill := range skills {
		name := skill.Name
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
			return fmt.Errorf("skill incluida inválida: %q; el binario está mal armado, volvé a bajar odachi", name)
		}
	}
	return nil
}

// firstFailure is the failure the one-line message quotes, so a blocked run
// names an agent and a path even when it only prints one line.
func firstFailure(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, " · falló · ") {
			return line
		}
	}
	if len(lines) > 0 {
		return lines[0]
	}
	return "no pude escribir las skills"
}

// blocked is a refused run: one Spanish line that says what to do next.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

// plural picks the form a human reads: "1 instalada", "5 instaladas".
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
