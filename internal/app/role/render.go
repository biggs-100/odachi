package role

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/role"
	"github.com/biggs-100/odachi/internal/domain/skill"
)

// MaxSkillBody is the most lines the rendered skill may have. A role whose rules
// do not fit is refused, not truncated: the cap is a contract of the renderer,
// and a role that cannot be loaded is not a role.
const MaxSkillBody = skill.MaxBodyLines

// skillFileName is the one file a skill is: every client reads exactly this name.
const skillFileName = "SKILL.md"

// SkillDest is where a role's skill goes: the folder every client reads, in the
// scope this project chose. Root is the absolute folder Base hangs from -the
// project, or the human's home-, Base is that folder in slash form, and Label is
// how a human reads it: `./.agents/skills` in the project, `~/.agents/skills` in
// the machine.
//
// The role does not decide the scope. A role that wrote into `.agents/skills`
// whatever the project chose would hand the adoption inference a copy odachi wrote
// itself, and the next plain init would install the whole curriculum there; so the
// caller resolves the scope and passes the destination down.
type SkillDest struct {
	Root  string
	Base  string
	Label string
}

// Path is the absolute path of one role's skill.
func (d SkillDest) Path(slug string) string {
	return filepath.Join(d.Root, filepath.FromSlash(d.Base), slug, skillFileName)
}

// Display is the same path as a human reads it.
func (d SkillDest) Display(slug string) string {
	return d.Label + "/" + slug + "/" + skillFileName
}

// target is one artifact a role owns: the bytes it must have and where they go.
// Root is absolute, because a role's skill can live outside the project.
type target struct {
	Root    string
	Path    string
	Label   string
	Content string
}

// abs is the path the disk port is asked about.
func (t target) abs() string { return filepath.Join(t.Root, filepath.FromSlash(t.Path)) }

// Render writes the artifacts of one role from its source: the agent file for
// each chosen client and the role's skill, in the destination the scope named. A
// second run writes nothing when the bytes already match, so the command is safe
// to repeat.
func Render(dir, slug, block string, clients []string, dest SkillDest, files Files) Result {
	source, problem := load(dir, slug, files)
	if problem != nil {
		return *problem
	}

	skill, err := skillText(source)
	if err != nil {
		return blocked(err.Error() + " (in " + Dir + "/" + slug + FileExt + ")")
	}

	wanted, unknown := targets(dir, slug, skill, block, source, clients, dest)
	if unknown != "" {
		return blocked("the client " + unknown + " has no native agent file: the ones that do are " + strings.Join(clientNames(), ", "))
	}

	lines := make([]string, 0, len(wanted))
	written := make([]string, 0, len(wanted))
	for _, target := range wanted {
		current, err := files.Read(target.abs())
		switch {
		case err == nil && string(current) == target.Content:
			lines = append(lines, "unchanged "+target.Label)
			continue
		case err != nil:
			if exists, existsErr := files.Exists(target.abs()); existsErr != nil || exists {
				return blocked("could not read " + target.Label + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role " + slug + " again")
			}
		}
		if err := files.Write(target.abs(), []byte(target.Content)); err != nil {
			return blocked("could not write " + target.Label + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role " + slug + " again")
		}
		lines = append(lines, "wrote "+target.Label)
		written = append(written, target.Label)
	}

	message := "rendered " + slug + " from " + Dir + "/" + slug + FileExt
	if len(written) == 0 {
		message = slug + " was already up to date"
	}
	return Result{Code: CodeOK, Message: message, Lines: lines, Files: written}
}

// List reports every role and whether its artifacts match the source.
func List(dir, block string, clients []string, dest SkillDest, files Files) Result {
	names, err := files.List(filepath.Join(rootOf(dir), filepath.FromSlash(Dir)))
	if err != nil {
		return blocked("could not list " + Dir + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role again")
	}

	lines := make([]string, 0, len(names))
	for _, name := range names {
		if !strings.HasSuffix(name, FileExt) {
			continue
		}
		slug := strings.TrimSuffix(name, FileExt)
		source, problem := load(dir, slug, files)
		if problem != nil {
			lines = append(lines, slug+": "+problem.Message)
			continue
		}
		skill, err := skillText(source)
		if err != nil {
			lines = append(lines, slug+": "+err.Error())
			continue
		}
		if !matches(dir, slug, skill, block, source, clients, dest, files) {
			lines = append(lines, slug+": stale, run odachi role "+slug)
			continue
		}
		lines = append(lines, slug+": up to date")
	}
	if len(lines) == 0 {
		return blocked("there are no roles yet: odachi role new <slug> writes the first source in " + Dir)
	}
	return Result{Code: CodeOK, Message: strings.Join(lines, "\n"), Lines: lines}
}

// targets is the one definition of what a role owns: its skill and one persona file
// per client. Render writes this list and matches compares it, so a target added to
// one path is added to both. Two copies would let the listing answer about a file
// nobody writes, with every test green.
func targets(dir, slug, skill, block string, source role.Role, clients []string, dest SkillDest) ([]target, string) {
	files := make([]target, 0, len(clients)+1)
	files = append(files, target{
		Root:    dest.Root,
		Path:    dest.Base + "/" + slug + "/" + skillFileName,
		Label:   dest.Display(slug),
		Content: skill,
	})
	for _, client := range clients {
		file, ok := persona.RoleFile(client, slug, source.Description, block, sectionText(source))
		if !ok {
			return nil, client
		}
		files = append(files, target{Root: rootOf(dir), Path: file.Path, Label: file.Path, Content: file.Content})
	}
	return files, ""
}

// matches reports whether every artifact of the role is already the bytes the
// source would render.
func matches(dir, slug, skill, block string, source role.Role, clients []string, dest SkillDest, files Files) bool {
	wanted, unknown := targets(dir, slug, skill, block, source, clients, dest)
	if unknown != "" {
		return false
	}
	for _, target := range wanted {
		current, err := files.Read(target.abs())
		if err != nil || string(current) != target.Content {
			return false
		}
	}
	return true
}

// load reads and validates one source.
func load(dir, slug string, files Files) (role.Role, *Result) {
	rel := Dir + "/" + slug + FileExt
	data, err := files.Read(filepath.Join(rootOf(dir), filepath.FromSlash(rel)))
	if err != nil {
		return role.Role{}, failure(blocked("could not read " + rel + ": " + oneLine(err.Error()) + "; write it with odachi role new " + slug))
	}
	parsed := role.Parse(rel, string(data))
	if problems := parsed.Problems(slug); len(problems) > 0 {
		lines := make([]string, 0, len(problems))
		for _, problem := range problems {
			lines = append(lines, problem.File+": "+problem.Reason)
		}
		return role.Role{}, failure(Result{Code: CodeBlocked, Message: "the source does not pass its check", Lines: lines})
	}
	return parsed, nil
}

// skillText renders the craft of the role as a skill: When from the trigger,
// Rules from what it asks and how it works, Never from what it refuses.
func skillText(source role.Role) (string, error) {
	var b strings.Builder
	b.WriteString("---\nname: " + source.Name + "\n")
	b.WriteString("description: " + source.Description + "\n---\n\n")
	b.WriteString("## When\n\n" + source.Trigger + "\n\n")
	b.WriteString("## Rules\n\n")
	b.WriteString("Before starting:\n" + source.Bodies["What it asks before starting"] + "\n\n")
	b.WriteString("How it works:\n" + source.Bodies["How it works"] + "\n\n")
	b.WriteString("It knows it worked when:\n" + source.Bodies["How it knows it worked"] + "\n\n")
	b.WriteString("## Never\n\n" + source.Bodies["What it refuses"] + "\n")
	text := b.String()

	body := strings.TrimSpace(strings.SplitN(text, "---\n\n", 2)[1])
	if lines := strings.Count(body, "\n") + 1; lines > MaxSkillBody {
		return "", fmt.Errorf("the role renders a skill of %d lines and the maximum is %d: cut the role, do not split the skill", lines, MaxSkillBody)
	}
	return text, nil
}

// sectionText is the role as the agent reads it: the professional, not the probe
// (the probe is how the human verifies, not what the agent follows).
func sectionText(source role.Role) string {
	var b strings.Builder
	b.WriteString("## The role you play\n\n" + source.Bodies["Who it is"] + "\n\n")
	b.WriteString("### What you ask before starting\n\n" + source.Bodies["What it asks before starting"] + "\n\n")
	b.WriteString("### How you work\n\n" + source.Bodies["How it works"] + "\n\n")
	b.WriteString("### What you refuse\n\n" + source.Bodies["What it refuses"] + "\n\n")
	b.WriteString("### How you know it worked\n\n" + source.Bodies["How it knows it worked"] + "\n\n")
	b.WriteString("### What you do not know\n\n" + source.Bodies["What it does not know"] + "\n\n")
	b.WriteString("### Where this comes from\n\n" + source.Bodies["Where this comes from"] + "\n")
	return b.String()
}

// clientNames is the list of clients that have a native agent file, for a
// message that has to name them.
func clientNames() []string {
	clients := persona.Clients()
	names := make([]string, 0, len(clients))
	for _, client := range clients {
		names = append(names, client.Agent)
	}
	return names
}

// failure boxes a result for the internal call chain.
func failure(result Result) *Result {
	return &result
}
