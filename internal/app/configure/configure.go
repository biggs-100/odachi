// Package configure is the use case behind the installation and configuration
// TUI: the state a human chose and the ordered operations that apply it.
//
// The package is pure orchestration: it validates the choices, decides the order
// and stops at the first failure. The verbs that touch the disk stay behind the
// Runner interface, which the CLI wires to the operations that already exist.
package configure

import (
	"fmt"
	"sort"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// The exit codes a report carries.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// State is what the wizard collected. Project is the chosen action: true
// initializes the project (init), false configures the machine alone (setup).
type State struct {
	Clients    string
	Skills     string
	Voice      string
	ReviewMode string
	LensModels map[string]string
	Project    bool
}

// Operation is one step of the apply, in the plan's order.
type Operation struct {
	ID    string
	Label string
}

// Status is how an operation ended, or that it never ran.
type Status string

const (
	Done    Status = "done"
	Failed  Status = "failed"
	Pending Status = "pending"
)

// Step is one operation with its outcome.
type Step struct {
	Operation
	Status Status
	Err    error
}

// Result is the whole apply: a code, the line a human reads, and every step.
type Result struct {
	Code    int
	Message string
	Steps   []Step
}

// Runner applies one operation. The CLI wires it to the verbs that already exist.
type Runner interface {
	Install(project bool, clients, skills string) error
	Voice(profile string) error
	Review(mode string) error
	Models(models map[string]string) error
}

// Validate reports every problem in the state, in the order the screens ask.
func Validate(s State) []string {
	var problems []string

	if _, err := persona.NormalizeSelection(s.Clients); err != nil {
		problems = append(problems, "elige al menos un cliente: opencode, claude o both")
	}
	if _, err := persona.NormalizeSkills(s.Skills); err != nil || strings.TrimSpace(s.Skills) == "" {
		problems = append(problems, "elige el scope de skills: home o project")
	}
	if _, ok := profile.Builtin()[strings.ToLower(strings.TrimSpace(s.Voice))]; !ok {
		problems = append(problems, fmt.Sprintf("el perfil %q no existe: usa %s",
			s.Voice, strings.Join(profileNames(), ", ")))
	}
	if _, err := review.NormalizeMode(s.ReviewMode); err != nil {
		problems = append(problems, "el review tiene que ser on u off")
	}
	if _, err := review.NormalizeLensModels(s.LensModels); err != nil {
		problems = append(problems, err.Error())
	}
	return problems
}

// Plan returns the operations a state runs, in order. The install comes first so
// the project and its store exist before anything is configured; the review's
// model table is a line nobody reads when the review is off.
func Plan(s State) []Operation {
	ops := []Operation{
		{ID: "install", Label: "instalar y configurar los clientes"},
		{ID: "voice", Label: "aplicar la voz"},
		{ID: "review", Label: "configurar el review"},
	}
	// The models operation runs whenever the review is on, even with an empty
	// table: an empty map is the human deleting the last model, and the file has
	// to be told. Nil and empty are different answers, and only the caller knows
	// which one it means.
	if strings.EqualFold(strings.TrimSpace(s.ReviewMode), review.ModeOn) {
		ops = append(ops, Operation{ID: "models", Label: "fijar el modelo de cada lente"})
	}
	return ops
}

// Apply validates the state, then runs the plan and stops at the first failure.
// Every operation after the failure is reported pending: the screen has to say
// what remains, not only what broke.
func Apply(s State, run Runner) Result {
	if problems := Validate(s); len(problems) > 0 {
		return Result{Code: CodeBlocked, Message: strings.Join(problems, "; ")}
	}

	plan := Plan(s)
	result := Result{Code: CodeOK, Steps: make([]Step, 0, len(plan))}

	for i, op := range plan {
		err := applyOne(s, s.LensModels, op.ID, run)
		step := Step{Operation: op, Status: Done, Err: err}
		if err != nil {
			step.Status = Failed
			result.Code = CodeBlocked
			result.Message = op.Label + ": " + err.Error()
		}
		result.Steps = append(result.Steps, step)
		if err != nil {
			for _, rest := range plan[i+1:] {
				result.Steps = append(result.Steps, Step{Operation: rest, Status: Pending})
			}
			break
		}
	}

	if result.Code == CodeOK {
		result.Message = "la configuración quedó aplicada"
	}
	return result
}

// applyOne runs one operation through the runner.
func applyOne(s State, models map[string]string, id string, run Runner) error {
	switch id {
	case "install":
		return run.Install(s.Project, s.Clients, s.Skills)
	case "voice":
		return run.Voice(s.Voice)
	case "review":
		return run.Review(s.ReviewMode)
	case "models":
		return run.Models(models)
	}
	return nil
}

// profileNames lists the built-in voices in a stable order for the refusal.
func profileNames() []string {
	names := make([]string, 0, len(profile.Builtin()))
	for name := range profile.Builtin() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
