// Package bench is the corpus and the rules of `odachi bench`.
//
// The instrument answers one question mechanically: does every blocked message
// name a command that actually resolves the block? A blocked message is worth
// nothing if it sends a person to a place with no exit, and "worth nothing" is
// not something a review reliably catches. So each block is classified by what
// its message names, and a block whose message names an unreachable command
// fails the bench.
//
// This package is pure: the journeys as data, the extraction of the command a
// message names, and the classification. The app layer supplies the sandbox, the
// binary and the process runner. It cannot run anything itself, which is why the
// instrument can promise never to execute a command that is not `odachi`.
package bench

import (
	"strings"
)

// Kind is how one block was classified.
type Kind string

const (
	// InBand means the message named a concrete command and, run in the same
	// sandbox, it exited 0: a real way out.
	InBand Kind = "in_band"
	// OutOfBand means the message named no odachi command at all: there is
	// nothing to run and the person is on their own.
	OutOfBand Kind = "out_of_band"
	// DeadEnd means the message named a command that, run, does not resolve the
	// block. This is the one that fails the bench: it is the instrument's
	// ratchet.
	DeadEnd Kind = "dead_end"
	// Interactive means the command ran and exited 2: by this tool's own
	// contract that is "only the human can unblock", so the machine could not
	// finish, and calling it a dead end would be the instrument measuring its own
	// blindness. It is counted apart and listed, never hidden.
	Interactive Kind = "interactive"
	// Templated means the command carries a <placeholder> and cannot be run
	// alone. It is counted apart and never fails the bench, because the bench's
	// inability to fill the placeholder is not the message's defect.
	Templated Kind = "templated"
)

// Action is one thing a sandbox does outside a measured invocation: create a
// file, or run a program. Setup is how a journey arrives at the state it is
// about, and it is never measured.
type Action struct {
	// Path is a sandbox-relative file to create, with Content as its bytes.
	Path    string
	Content string
	// Command is a program and its arguments, run inside the sandbox. It is how
	// a journey gets a repository, and nothing else: a message's Next is never
	// run through this field.
	Command []string
}

// Step is one invocation of the binary under test, with the exit code it is
// expected to have.
type Step struct {
	Args []string
	// Want is the exit this step should have. A different one is reported as a
	// surprise: the journey's premise changed, and a bench that says nothing
	// about it would be measuring a product that no longer exists.
	Want int
	// Then is what the sandbox does after this step, before the next one.
	Then []Action
}

// Journey is one journey: a name, why it exists, the state it starts from, and
// the invocations that are measured. A journey is data, never code, so the corpus
// can grow without the instrument growing with it.
type Journey struct {
	Name  string
	Why   string
	Setup []Action
	Steps []Step
}

// The binary is always called `odachi` in a message; the bench runs the binary
// under test instead. That replacement is the only thing it is allowed to
// change, and it is what makes `--binary` meaningful.
const binaryName = "odachi"

// startOutcome is the answer the journeys give to question 1. It is a real
// sentence and not "test": the slug, the brief and the task document are derived
// from it, so a journey that measures the artifacts measures real ones.
const startOutcome = "una lista de gastos compartida"

// startArgs is the interview answered by flags, which is how a machine runs it.
func startArgs(extra ...string) []string {
	args := []string{
		"start",
		"--outcome", startOutcome,
		"--audience", "cpu",
		"--done", "ver la lista en el teléfono",
		"--sensitive", "none",
		"--deadline", "week",
	}
	return append(args, extra...)
}

// Journeys is the corpus Contract XV asks for: the eight journeys a person
// actually makes in the first hour. Each one is here because it has blocked
// someone, or because it is the state where a message has to be trusted.
func Journeys() []Journey {
	return []Journey{
		{
			Name: "repo-nuevo-start",
			Why:  "the first minute: an empty directory and the interview answered by flags",
			Steps: []Step{
				{Args: startArgs(), Want: 0},
			},
		},
		{
			Name: "start-dos-veces",
			Why:  "the person runs the interview again because they do not remember whether they ran it",
			Steps: []Step{
				{Args: startArgs(), Want: 0},
				{Args: startArgs(), Want: 1},
			},
		},
		{
			Name:  "check-sin-prueba",
			Why:   "the agent asks whether it is ready before having any proof",
			Setup: []Action{gitInit()},
			Steps: []Step{
				{Args: startArgs(), Want: 0},
				{Args: []string{"check"}, Want: 1},
			},
		},
		{
			Name:  "check-run-verde",
			Why:   "the proof runs, is saved with its tree, and the next run is ready",
			Setup: []Action{gitInit()},
			Steps: []Step{
				{Args: startArgs(), Want: 0},
				{Args: []string{"check", "--run", "true"}, Want: 0},
				{Args: []string{"check"}, Want: 0},
			},
		},
		{
			Name:  "archivo-prohibido",
			Why:   "the fence is crossed: a file appears from something the person said was out of scope",
			Setup: []Action{gitInit()},
			Steps: []Step{
				{
					Args: startArgs("--exclude", "pagos"),
					Want: 0,
					Then: []Action{{
						Path:    "internal/payments/pay.go",
						Content: "package payments\n",
					}},
				},
				{Args: []string{"check"}, Want: 2},
			},
		},
		{
			Name: "plan-y-undo",
			Why:  "the agent adds a step, closes it with its commit, and the person changes their mind",
			Steps: []Step{
				{Args: startArgs(), Want: 0},
				{Args: []string{"plan", "save the expenses in sqlite", "--req", "R1"}, Want: 0},
				{Args: []string{"plan", "--done", "T1", "--commit", "abc1234"}, Want: 0},
				{Args: []string{"undo"}, Want: 0},
			},
		},
		{
			Name:  "review-solo-docs",
			Why:   "a change that is only documentation: review must not ask for lenses that do not apply",
			Setup: []Action{gitInit(), {Path: "notas.md", Content: "# notas\n\nlo que decidimos hoy\n"}},
			Steps: []Step{
				{Args: []string{"review"}, Want: 0},
				{Args: []string{"review", "--show"}, Want: 1},
			},
		},
		{
			Name:  "repo-sin-commits",
			Why:   "git init and nothing else: the project does not have a history to compare yet",
			Setup: []Action{gitInit(), {Path: "notas.md", Content: "# notas\n"}},
			Steps: []Step{
				{Args: []string{"check"}, Want: 2},
			},
		},
	}
}

// gitInit is how a journey gets a repository. It is setup, not a measurement:
// the bench measures what odachi says, not what git does.
func gitInit() Action {
	return Action{Command: []string{"git", "init", "-q"}}
}

// Named selects the journeys `--journeys` asked for, in corpus order. A name that
// does not exist is returned instead of ignored: a filter that silently matches
// nothing would report "all good" about a corpus that never ran.
func Named(all []Journey, names []string) (selected []Journey, unknown []string) {
	if len(names) == 0 {
		return all, nil
	}

	wanted := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		wanted[name] = true
	}

	known := map[string]bool{}
	for _, journey := range all {
		known[journey.Name] = true
		if wanted[journey.Name] {
			selected = append(selected, journey)
		}
	}
	for name := range wanted {
		if !known[name] {
			unknown = append(unknown, name)
		}
	}
	return selected, unknown
}

// KindOf classifies one block. command is what the message named (nil when it
// named nothing) and exit is what running it returned.
func KindOf(command []string, exit int) Kind {
	switch {
	case len(command) == 0:
		return OutOfBand
	case IsTemplated(command):
		return Templated
	case exit == 0:
		return InBand
	case exit == 2:
		return Interactive
	default:
		return DeadEnd
	}
}

// IsTemplated reports whether a command carries a <placeholder>. Such a command
// cannot be run alone, and a bench that ran it would be testing its own guess.
func IsTemplated(command []string) bool {
	for _, word := range command {
		if strings.Contains(word, "<") && strings.Contains(word, ">") {
			return true
		}
	}
	return false
}

// Command finds the `odachi ...` invocation a blocked message names.
//
// The messages are sentences with the command inside, sometimes after a label
// and sometimes in the middle of the sentence. The search is for the name of
// this tool followed by a verb, plus the flags and values that belong to it, so
// `odachi start --force (or delete brief.md)` yields the command and not the
// aside.
//
// Only `odachi` commands are ever returned. That is the instrument's safety
// property: a message could name anything, and the bench is not a shell.
func Command(message string) []string {
	for _, line := range strings.Split(normalize(message), "\n") {
		words := splitWords(line)
		for i, word := range words {
			if unquote(word) != binaryName {
				continue
			}
			if command := commandFrom(words[i:]); len(command) > 1 {
				return command
			}
		}
	}
	return nil
}

// commandFrom reads one command out of the words that start with `odachi`. A
// command with no verb is not concrete — "I installed odachi" names a product,
// not a command — so it is not one.
func commandFrom(words []string) []string {
	command := []string{binaryName}
	if len(words) < 2 {
		return command
	}

	verb := unquote(words[1])
	if !verbs[verb] {
		// A word after the tool's name is only a command when it is one of the
		// verbs: "I installed odachi and came back" names a product, not a
		// command.
		return command
	}
	command = append(command, verb)

	previous := verb
	for _, word := range words[2:] {
		switch {
		case strings.HasPrefix(word, "-"):
			command = append(command, word)
			previous = word
		case isPlaceholder(word):
			command = append(command, unquote(word))
			previous = unquote(word)
		case takesValue(previous):
			// `--to senior`, `--lens risk`: the value belongs to the command.
			command = append(command, unquote(word))
			previous = unquote(word)
		default:
			return command
		}
	}
	return command
}

// verbs is the tool's own surface. A message that does not name one of these is
// not naming a command, and running a guess would be the bench inventing an
// instrument reading instead of taking one.
var verbs = map[string]bool{
	"start":   true,
	"ask":     true,
	"plan":    true,
	"level":   true,
	"summary": true,
	"undo":    true,
	"check":   true,
	"review":  true,
	"doctor":  true,
	"bench":   true,
	"init":    true,
	"skills":  true,
	"skill":   true,
	"pack":    true,
	"version": true,
}

// booleanFlags are the flags that take no value. Every other flag does, which is
// what lets `odachi level --to senior` stay one command instead of becoming
// `odachi level --to`.
var booleanFlags = map[string]bool{
	"-h":               true,
	"--help":           true,
	"--version":        true,
	"--force":          true,
	"--json":           true,
	"--quiet":          true,
	"--dry-run":        true,
	"--show":           true,
	"--list":           true,
	"--check":          true,
	"--write":          true,
	"--require-review": true,
}

// takesValue reports whether the previous word expects a value.
func takesValue(previous string) bool {
	return strings.HasPrefix(previous, "-") && !booleanFlags[previous]
}

// isPlaceholder reports whether a word is a `<...>` slot.
func isPlaceholder(word string) bool {
	return strings.Contains(word, "<") && strings.Contains(word, ">")
}

// splitWords splits one line into words, keeping a double-quoted argument
// together: `--run "go test ./..."` is one argument after the flag, not three.
// The quotes are dropped, because the command is run as arguments and not through
// a shell.
func splitWords(line string) []string {
	words := []string{}
	var current strings.Builder
	quoted := false

	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	for _, r := range line {
		switch {
		case r == '"':
			quoted = !quoted
		case (r == ' ' || r == '\t') && !quoted:
			flush()
		default:
			current.WriteRune(r)
		}
	}
	flush()
	return words
}

// unquote removes the punctuation a sentence wraps an argument in: `(odachi`,
// `«odachi»` and `odachi,` all name the same command. A colon is left alone on
// purpose: every blocked line the CLI prints starts with the label `odachi:`, and
// that label is not an invocation.
func unquote(word string) string {
	return strings.Trim(word, `()[]{}«»"',;¡!¿?`)
}

// normalize makes the line endings the same on every platform.
func normalize(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

// SandboxEnv is the environment one journey runs in: the caller's environment
// with HOME, USERPROFILE and the git config pointed inside the sandbox, so a
// journey never reads or writes the home of whoever ran the bench and never
// inherits a global git setting that changes its behaviour.
func SandboxEnv(base []string, home string) []string {
	overrides := map[string]string{
		"HOME":              home,
		"USERPROFILE":       home,
		"GIT_CONFIG_GLOBAL": home + "/.gitconfig",
		"GIT_CONFIG_SYSTEM": home + "/.gitconfig-system",
	}

	env := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		name, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, replaced := overrides[name]; replaced {
			continue
		}
		env = append(env, entry)
	}

	// Fixed order, so two runs of the same bench produce the same environment and
	// a failure is reproducible.
	for _, name := range []string{"HOME", "USERPROFILE", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM"} {
		env = append(env, name+"="+overrides[name])
	}
	return env
}
