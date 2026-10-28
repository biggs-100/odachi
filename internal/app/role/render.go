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

// SkillPath is where the rendered skill lands, relative to the project root. It
// is the folder every client reads, so one write serves them all.
func SkillPath(slug string) string { return ".agents/skills/" + slug + "/SKILL.md" }

// Render writes the artifacts of one role from its source: the agent file for
// each chosen client and the role's skill. A second run writes nothing when the
// bytes already match, so the command is safe to repeat.
func Render(dir, slug, block string, clients []string, files Files) Result {
	source, problem := load(dir, slug, files)
	if problem != nil {
		return *problem
	}

	skill, err := skillText(source)
	if err != nil {
		return blocked(err.Error() + " (in " + Dir + "/" + slug + FileExt + ")")
	}

	wanted, unknown := targets(slug, skill, block, source, clients)
	if unknown != "" {
		return blocked("the client " + unknown + " has no native agent file: the ones that do are " + strings.Join(clientNames(), ", "))
	}

	lines := make([]string, 0, len(wanted))
	written := make([]string, 0, len(wanted))
	for _, target := range wanted {
		path := filepath.Join(rootOf(dir), filepath.FromSlash(target.Path))
		current, err := files.Read(path)
		switch {
		case err == nil && string(current) == target.Content:
			lines = append(lines, "unchanged "+target.Path)
			continue
		case err != nil:
			if exists, existsErr := files.Exists(path); existsErr != nil || exists {
				return blocked("could not read " + target.Path + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role " + slug + " again")
			}
		}
		if err := files.Write(path, []byte(target.Content)); err != nil {
			return blocked("could not write " + target.Path + ": " + oneLine(err.Error()) + "; check the permissions and run odachi role " + slug + " again")
		}
		lines = append(lines, "wrote "+target.Path)
		written = append(written, target.Path)
	}

	message := "rendered " + slug + " from " + Dir + "/" + slug + FileExt
	if len(written) == 0 {
		message = slug + " was already up to date"
	}
	return Result{Code: CodeOK, Message: message, Lines: lines, Files: written}
}

// List reports every role and whether its artifacts match the source.
func List(dir, block string, clients []string, files Files) Result {
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
		if !matches(dir, slug, skill, block, source, clients, files) {
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
func targets(slug, skill, block string, source role.Role, clients []string) (files []persona.File, unknown string) {
	files = make([]persona.File, 0, len(clients)+1)
	files = append(files, persona.File{Path: SkillPath(slug), Content: skill})
	for _, client := range clients {
		file, ok := persona.RoleFile(client, slug, source.Description, block, sectionText(source))
		if !ok {
			return nil, client
		}
		files = append(files, file)
	}
	return files, ""
}

// matches reports whether every artifact of the role is already the bytes the
// source would render.
func matches(dir, slug, skill, block string, source role.Role, clients []string, files Files) bool {
	wanted, unknown := targets(slug, skill, block, source, clients)
	if unknown != "" {
		return false
	}
	for _, target := range wanted {
		current, err := files.Read(filepath.Join(rootOf(dir), filepath.FromSlash(target.Path)))
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
