// Package delegate composes the prompt a parent hands a delegated worker. It is
// the mechanism the seventh probe asked for: the parent does not write the skills
// section by hand, so it cannot drop it. The step records the craft at plan time
// and this verb resolves it to the exact paths the worker reads.
//
// It is pure: it reads the task doc through the Files port and asks the Resolver
// where a skill lives. The CLI wires both.
package delegate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes, the same shape as the other verbs.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// Options is one `odachi delegate` invocation.
type Options struct {
	// Dir is the project root, used to resolve a relative task-doc path and to
	// print the absolute path the worker must read.
	Dir string
	// TaskDoc is the document the step lives in, as the parent wrote it.
	TaskDoc string
	// Step is the T id to hand over.
	Step string
	// Surfaces are the paths the worker may edit.
	Surfaces []string
	// Commands are the verification commands the worker must run.
	Commands []string
	// Resolver answers where a skill name lives. It is the same order every
	// client reads: the project's copy first, then the machine's.
	Resolver Resolver
}

// Resolver is the skill lookup the CLI wires: the exact SKILL.md for a name, and
// the names that exist, for an error that tells the reader what it can use.
type Resolver interface {
	Path(name string) (string, bool)
	Names() []string
}

// Files is the disk port: enough to read the task doc.
type Files interface {
	Read(path string) ([]byte, error)
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Prompt  string
	Task    string
}

// Run composes the prompt for one delegated step.
func Run(opts Options, files Files) Result {
	if strings.TrimSpace(opts.TaskDoc) == "" {
		return blocked(`missing the task doc: write odachi delegate odd/tasks/<file>.md --step T1`)
	}
	stepID := strings.ToUpper(strings.TrimSpace(opts.Step))
	if stepID == "" {
		return blocked(`missing the step: write odachi delegate odd/tasks/<file>.md --step T1`)
	}

	path := opts.TaskDoc
	if !filepath.IsAbs(path) {
		path = filepath.Join(opts.Dir, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	data, err := files.Read(abs)
	if err != nil {
		return blocked("could not read " + opts.TaskDoc + ": " + oneLine(err.Error()))
	}
	doc := taskdoc.Parse(string(data))

	step, ok := stepByID(doc, stepID)
	if !ok {
		return blocked(fmt.Sprintf("the task doc has no step %s; its steps are %s", stepID, stepIDs(doc)))
	}
	if step.Route != taskdoc.RouteDelegated {
		return blocked(fmt.Sprintf("step %s runs %s, not delegated: a step the parent runs needs no prompt", stepID, routeOf(step)))
	}

	skills, problem := resolve(opts.Resolver, step.Skills)
	if problem != "" {
		return blocked(problem)
	}

	requirement, hasRequirement := requirementOf(doc, step.Requirement)

	prompt := compose(abs, step, requirement, hasRequirement, opts.Surfaces, opts.Commands, skills)
	return Result{Code: CodeOK, Prompt: prompt, Task: opts.TaskDoc}
}

// skill is one resolved path, in the order the step declared it.
type skill struct{ name, path string }

// resolve turns the recorded names into exact paths. A missing skill is named
// with the ones that exist: the parent must fix the plan, not guess a path.
func resolve(resolver Resolver, names []string) ([]skill, string) {
	if len(names) == 0 {
		return nil, ""
	}
	if resolver == nil {
		return nil, "no skill resolver is wired: the delegation cannot resolve " + strings.Join(names, ", ")
	}
	resolved := make([]skill, 0, len(names))
	for _, name := range names {
		path, ok := resolver.Path(name)
		if !ok {
			return nil, fmt.Sprintf("the skill %q is not installed; the ones I know are %s", name, strings.Join(resolver.Names(), ", "))
		}
		resolved = append(resolved, skill{name: name, path: path})
	}
	return resolved, ""
}

// compose writes the prompt the worker reads. The skills section is always
// present: when the step owes none it says so, because an absent section is what
// the worker refuses.
func compose(taskDoc string, step taskdoc.Step, requirement taskdoc.Requirement, hasRequirement bool, surfaces, commands []string, skills []skill) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s · %s\n\n", step.ID, step.Text)
	fmt.Fprintf(&b, "Task doc: %s\n", taskDoc)
	fmt.Fprintf(&b, "Step: %s · %s\n", step.ID, step.Text)

	if hasRequirement {
		fmt.Fprintf(&b, "\nRequirement %s (verbatim):\n> %s\n", requirement.ID, requirement.Text)
	}

	if len(surfaces) > 0 {
		b.WriteString("\nSurfaces (edit nothing outside them):\n")
		for _, surface := range surfaces {
			fmt.Fprintf(&b, "- %s\n", surface)
		}
	}

	if len(commands) > 0 {
		b.WriteString("\nVerification (run these):\n")
		for _, command := range commands {
			fmt.Fprintf(&b, "- %s\n", command)
		}
	}

	b.WriteString("\n## Skills to load before work\n\n")
	if len(skills) == 0 {
		b.WriteString("None: this step owes no skill.\n")
	} else {
		b.WriteString("Read these exact files before reading, writing or testing:\n\n")
		for _, s := range skills {
			fmt.Fprintf(&b, "- %s\n", s.path)
		}
	}

	b.WriteString("\nReturn your report in the format your prompt names.\n")
	return b.String()
}

// stepByID finds the step the parent names.
func stepByID(doc taskdoc.TaskDoc, id string) (taskdoc.Step, bool) {
	for _, step := range doc.Steps {
		if strings.EqualFold(step.ID, id) {
			return step, true
		}
	}
	return taskdoc.Step{}, false
}

// requirementOf reads the requirement the step cites, when it cites one.
func requirementOf(doc taskdoc.TaskDoc, id string) (taskdoc.Requirement, bool) {
	if id == "" {
		return taskdoc.Requirement{}, false
	}
	for _, requirement := range doc.Requirements {
		if strings.EqualFold(requirement.ID, id) {
			return requirement, true
		}
	}
	return taskdoc.Requirement{}, false
}

// stepIDs lists the ids a document has, for an error that names what exists.
func stepIDs(doc taskdoc.TaskDoc) string {
	if len(doc.Steps) == 0 {
		return "none"
	}
	ids := make([]string, 0, len(doc.Steps))
	for _, step := range doc.Steps {
		ids = append(ids, step.ID)
	}
	return strings.Join(ids, ", ")
}

// routeOf spells a step's route for the error that refuses it.
func routeOf(step taskdoc.Step) string {
	if step.Route == "" {
		return "with no route declared"
	}
	if step.Agent != "" {
		return step.Route + " by " + step.Agent
	}
	return step.Route
}

// blocked is the result a refusal returns.
func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: message}
}

// oneLine keeps an error message on one line.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
