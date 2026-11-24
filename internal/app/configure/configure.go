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

	initapp "github.com/biggs-100/odachi/internal/app/init"
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
	Clients string
	Skills  string
	Voice   string
	// VoiceScope is where the chosen profile is written: machine for the
	// default in the home, project for odd/profile.toml (0357).
	VoiceScope string
	ReviewMode string
	// LensModels is the table the review runs with. Nil means the caller did not
	// touch it; an empty map means the human cleared it, and the operation runs.
	LensModels map[string]string
	// Verify is the verifier's row of the models section. Nil means the section
	// was never opened: odd/review.toml keeps its line. The plan writes only a
	// difference, so choosing the value the file already carries writes nothing.
	Verify *ModelChoice
	// ClientModels is the row of each client the section showed, by client. Only
	// a client whose choice differs reaches the install operation, which is why
	// an unchanged section never names one (0337).
	ClientModels map[string]ModelChoice
	Project      bool
}

// ModelChoice is one row of the models section: the value the human chose and
// the value the file carried when the section read it. The two together are what
// makes "nothing changed" a fact instead of a guess (0337).
type ModelChoice struct {
	Chosen   string
	Detected string
}

// The two places a voice can be written. Project is the default so a caller
// that never opened the new scope keeps the behavior it had before 0357.
const (
	VoiceMachine = "machine"
	VoiceProject = "project"
)

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
	// Install runs init or setup. models is the model of each client whose row
	// the human changed, and it travels as `--model client=model` so the verb
	// that owns the file is still the only writer (0337).
	Install(project bool, clients, skills string, models map[string]string) error
	Voice(profile string) error
	// GlobalVoice writes the machine's default profile without touching a project.
	GlobalVoice(profile string) error
	Review(mode string) error
	Models(models map[string]string) error
	// Verify writes the verifier's model into odd/review.toml, keeping the lens
	// table and the numbers the section does not touch (0337).
	Verify(model string) error
	// Field writes one axis of the profile: the editor changes one row and this is
	// the only write it makes (0336).
	Field(field, value string) error
	// Axes re-reads the profile in force, so the editor shows the file that stands
	// instead of the one the wizard read at the start (0336).
	Axes() ([]profile.Axis, error)
}

// normalize returns the state with the values the writer expects and the problems
// that prevent an apply. One function answers both questions, so the validation
// and the write cannot disagree about what a valid table is, and no error is
// discarded on the way to the runner.
func normalize(s State) (State, []string) {
	var problems []string

	if clients, err := persona.NormalizeSelection(s.Clients); err != nil {
		problems = append(problems, "elige al menos un cliente: opencode, claude o both")
	} else {
		s.Clients = clients
	}
	if skills, err := persona.NormalizeSkills(s.Skills); err != nil || strings.TrimSpace(s.Skills) == "" {
		problems = append(problems, "elige el scope de skills: home o project")
	} else {
		s.Skills = skills
	}
	voice := strings.ToLower(strings.TrimSpace(s.Voice))
	if _, ok := profile.Builtin()[voice]; !ok {
		problems = append(problems, fmt.Sprintf("el perfil %q no existe: usa %s",
			s.Voice, strings.Join(profileNames(), ", ")))
	} else {
		s.Voice = voice
	}
	scope := strings.ToLower(strings.TrimSpace(s.VoiceScope))
	if scope == "" {
		scope = VoiceProject
	}
	if scope != VoiceMachine && scope != VoiceProject {
		problems = append(problems, fmt.Sprintf("el destino de la voz %q no existe: usa machine o project", s.VoiceScope))
	} else {
		s.VoiceScope = scope
	}
	if mode, err := review.NormalizeMode(s.ReviewMode); err != nil {
		problems = append(problems, "el review tiene que ser on u off")
	} else {
		s.ReviewMode = mode
	}
	if s.LensModels != nil {
		models, err := review.NormalizeLensModels(s.LensModels)
		switch {
		case err != nil:
			problems = append(problems, err.Error())
		case models == nil:
			// An empty table is the human clearing it: keep the map non-nil so
			// the plan knows the table was touched.
			s.LensModels = map[string]string{}
		default:
			s.LensModels = models
		}
	}
	// The models section obeys the rule each writer already obeys: the verifier's
	// value replaces {model} inside a command, and a client's value ends up in a
	// file that client parses. An empty value means different things on the two
	// rows and each writer decides it: the verifier's line is removed, because
	// the review file's writer can, and a client's row writes nothing, because no
	// verb removes that key (0337).
	if s.Verify != nil {
		choice := *s.Verify
		choice.Chosen = strings.TrimSpace(choice.Chosen)
		choice.Detected = strings.TrimSpace(choice.Detected)
		if choice.Chosen != "" {
			checked, err := review.NormalizeModel(choice.Chosen)
			if err != nil {
				problems = append(problems, "el verificador: "+err.Error())
			} else {
				choice.Chosen = checked
			}
		}
		s.Verify = &choice
	}
	if len(s.ClientModels) > 0 {
		chosen := make(map[string]string, len(s.ClientModels))
		for client, choice := range s.ClientModels {
			value := strings.TrimSpace(choice.Chosen)
			if value != "" {
				chosen[client] = value
			}
		}
		models, err := initapp.NormalizeModels(chosen)
		if err != nil {
			problems = append(problems, "los modelos de los clientes: "+err.Error())
		} else {
			checked := make(map[string]ModelChoice, len(s.ClientModels))
			for client, choice := range s.ClientModels {
				// The client table and the flag lowercase the name, so the row
				// key is folded the same way: two spellings must not survive as
				// two clients. A client's empty value is "write nothing",
				// because no verb removes the line (0337).
				name := strings.ToLower(strings.TrimSpace(client))
				checked[name] = ModelChoice{
					Chosen:   strings.TrimSpace(models[name]),
					Detected: strings.TrimSpace(choice.Detected),
				}
			}
			s.ClientModels = checked
		}
	}
	return s, problems
}

// Validate reports every problem in the state, in the order the screens ask.
func Validate(s State) []string {
	_, problems := normalize(s)
	return problems
}

// Plan returns the operations a state runs, in order. The install comes first so
// the project and its store exist before anything is configured; the review's
// model table is a line nobody reads when the review is off.
func Plan(s State) []Operation {
	voiceOp := Operation{ID: "voice", Label: "aplicar la voz"}
	if strings.EqualFold(strings.TrimSpace(s.VoiceScope), VoiceMachine) {
		voiceOp = Operation{ID: "globalvoice", Label: "fijar el perfil por defecto de la máquina"}
	}
	ops := []Operation{
		{ID: "install", Label: "instalar y configurar los clientes"},
		voiceOp,
		{ID: "review", Label: "configurar el review"},
	}
	// The models operation runs whenever the review is on and the table is not
	// nil: an empty map is the human deleting the last model, and the file has to
	// be told; nil is the caller that never touched it, and the file keeps what it
	// has.
	if strings.EqualFold(strings.TrimSpace(s.ReviewMode), review.ModeOn) && s.LensModels != nil {
		ops = append(ops, Operation{ID: "models", Label: "fijar el modelo de cada lente"})
	}
	// The verifier's row is written only when it differs from the value the file
	// showed: a section that was opened and left alone is not a write (0337).
	if s.Verify != nil && s.Verify.Chosen != s.Verify.Detected {
		ops = append(ops, Operation{ID: "verify", Label: "fijar el modelo del verificador"})
	}
	return ops
}

// clientModels is what the install operation is told: one entry per client whose
// choice differs from the file, and none for a cleared row -there is no verb that
// removes a client's line, so nothing is written for it- (0337).
func clientModels(s State) map[string]string {
	changed := map[string]string{}
	for client, choice := range s.ClientModels {
		if choice.Chosen == "" || choice.Chosen == choice.Detected {
			continue
		}
		changed[client] = choice.Chosen
	}
	if len(changed) == 0 {
		return nil
	}
	return changed
}

// Apply validates the state, then runs the plan and stops at the first failure.
// Every operation after the failure is reported pending: the screen has to say
// what remains, not only what broke.
func Apply(s State, run Runner) Result {
	normalized, problems := normalize(s)
	if len(problems) > 0 {
		return Result{Code: CodeBlocked, Message: strings.Join(problems, "; ")}
	}
	s = normalized

	plan := Plan(s)
	result := Result{Code: CodeOK, Steps: make([]Step, 0, len(plan))}

	for i, op := range plan {
		err := applyOne(s, op.ID, run)
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
func applyOne(s State, id string, run Runner) error {
	switch id {
	case "install":
		return run.Install(s.Project, s.Clients, s.Skills, clientModels(s))
	case "voice":
		return run.Voice(s.Voice)
	case "globalvoice":
		return run.GlobalVoice(s.Voice)
	case "review":
		return run.Review(s.ReviewMode)
	case "models":
		return run.Models(s.LensModels)
	case "verify":
		return run.Verify(s.Verify.Chosen)
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
