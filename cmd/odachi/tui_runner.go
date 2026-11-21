package main

import (
	"bytes"
	"errors"
	"io"
	"sort"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/hostmodel"
	"github.com/biggs-100/odachi/internal/app/level"
	appreview "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/profile"
)

// tuiRunner applies one operation of the wizard by calling the verb that already
// owns it. The functions are fields so a test can drive the branch without
// touching the home or the project.
type tuiRunner struct {
	dir       string
	initFn    func(dir string, args []string, stdout, stderr io.Writer) int
	setupFn   func(args []string, stdout, stderr io.Writer) int
	levelFn   func(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int
	modeFn    func(dir string, args []string, stdout, stderr io.Writer) int
	catalogFn func(dir string) ([]string, bool)
}

// newTUIRunner wires the runner to the same functions the CLI dispatches.
func newTUIRunner(dir string) tuiRunner {
	return tuiRunner{
		dir:       dir,
		initFn:    runInit,
		setupFn:   runSetup,
		levelFn:   runLevel,
		modeFn:    runReviewMode,
		catalogFn: hostmodel.Models,
	}
}

// Install runs init when the human chose the project, setup when they chose the
// machine alone. The clients and the skills scope are the two choices init
// records; setup writes no store, which is why a folder without a repository is
// asked which one it wants. The client models travel as `--model client=model`,
// so the verb that owns opencode.json and ~/.claude/settings.json stays the only
// writer of them (0337).
func (r tuiRunner) Install(project bool, clients, skills string, models map[string]string) error {
	args := []string{"--agents", clients}
	if project {
		args = append(args, "--skills", skills)
	}
	args = appendModelArgs(args, models)

	var out, errOut bytes.Buffer
	code := 0
	if project {
		code = r.initFn(r.dir, args, &out, &errOut)
	} else {
		code = r.setupFn(args, &out, &errOut)
	}
	return runnerError(code, out.String(), errOut.String())
}

// appendModelArgs adds one --model client=model per client, in a stable order:
// the map iterates in no fixed order, and a command a human reads should not
// change between two runs (0337).
func appendModelArgs(args []string, models map[string]string) []string {
	clients := make([]string, 0, len(models))
	for client := range models {
		clients = append(clients, client)
	}
	sort.Strings(clients)
	for _, client := range clients {
		args = append(args, "--model", client+"="+models[client])
	}
	return args
}

// Voice applies the profile through the same path `odachi level --to` uses.
func (r tuiRunner) Voice(profile string) error {
	var out, errOut bytes.Buffer
	code := r.levelFn(r.dir, []string{"--to", profile}, strings.NewReader(""), &out, &errOut)
	return runnerError(code, out.String(), errOut.String())
}

// Review flips the switch through `odachi review mode`.
func (r tuiRunner) Review(mode string) error {
	var out, errOut bytes.Buffer
	code := r.modeFn(r.dir, []string{mode}, &out, &errOut)
	return runnerError(code, out.String(), errOut.String())
}

// Models writes the lens table through the review app, which preserves the
// prompt limit and the retention the wizard does not touch.
func (r tuiRunner) Models(models map[string]string) error {
	result := appreview.SetConfig(appreview.ConfigOptions{Dir: r.dir, Models: models}, disk(r.dir), reviewModeCodec{})
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

// Verify writes the verifier's model through the same review app: SetConfig
// reads the file, changes the one key and renders the rest, so the lens table and
// the numbers the section does not touch survive (0337).
func (r tuiRunner) Verify(model string) error {
	result := appreview.SetConfig(appreview.ConfigOptions{Dir: r.dir, Verify: &model}, disk(r.dir), reviewModeCodec{})
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

// Limits writes the four scalars through the same review app the models section
// uses: SetConfig reads the file, changes only the named fields and renders the
// rest, so the [[check]] array keeps the gate severity the project declared. A
// pointer left nil keeps the value the file carried (0342).
func (r tuiRunner) Limits(choice limitChoice) error {
	result := appreview.SetConfig(appreview.ConfigOptions{
		Dir:          r.dir,
		MaxPromptKB:  choice.MaxPromptKB,
		KeepVerdicts: choice.KeepVerdicts,
		KeepPrompts:  choice.KeepPrompts,
		CheckTimeout: choice.CheckTimeout,
	}, disk(r.dir), reviewModeCodec{})
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

// Checks writes the checks the wizard edited through the same review app the
// other sections use. SetConfig locates each entry by the identity read from the
// file, keeps the severity the project declared and refuses a value the domain
// cannot accept, so the wizard is a caller of the use case, never a second
// writer. A name or command change is a new fingerprint and needs a fresh
// `odachi review checks approve`; a remedy-only edit does not (0346).
func (r tuiRunner) Checks(edits []checkChoice) error {
	checks := make([]appreview.CheckEdit, 0, len(edits))
	for _, edit := range edits {
		checks = append(checks, appreview.CheckEdit{
			Identity: edit.Identity,
			Name:     edit.Name,
			Command:  edit.Command,
			Remedy:   edit.Remedy,
		})
	}
	result := appreview.SetConfig(appreview.ConfigOptions{
		Dir:    r.dir,
		Checks: checks,
	}, disk(r.dir), reviewModeCodec{})
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

// Field writes one axis of the profile through the app that owns the file, the
// same way Models writes the lens table: the wizard is a caller of the use case,
// never a second writer (0336).
func (r tuiRunner) Field(field, value string) error {
	current, code, message := currentProfile(r.dir, "odachi tui", false)
	if code != start.CodeOK {
		return errors.New(message)
	}
	if current == nil {
		return errors.New("no hay perfil que cambiar: run odachi start")
	}
	result := level.Set(level.SetOptions{Dir: r.dir, Current: *current, Field: field, Value: value}, disk(r.dir))
	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

// Axes re-reads the profile in force: the hub applies the voice in one section and
// the editor lives in another, so the rows have to describe the file that stands
// now, not the one the wizard read when it opened (0336).
func (r tuiRunner) Axes() ([]profile.Axis, error) {
	current, code, message := currentProfile(r.dir, "odachi tui", false)
	if code != start.CodeOK {
		return nil, errors.New(message)
	}
	if current == nil {
		return nil, errors.New("no hay perfil: run odachi start")
	}
	return profile.Axes(*current), nil
}

// Catalog answers the models the host serves through the same bounded probe the
// review verb validates its lens table against, so the picker offers what the
// verbs already accept. It runs from the home, outside the project, and it only
// reads: opening the list writes nothing (0338).
func (r tuiRunner) Catalog() ([]string, bool) {
	return r.catalogFn(r.dir)
}

// runnerError turns a verb's exit code and output into the line the apply screen
// shows. The first non-empty line is the message the verb already wrote.
func runnerError(code int, out, errOut string) error {
	if code == 0 {
		return nil
	}
	for _, text := range []string{errOut, out} {
		for _, line := range strings.Split(text, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				return errors.New(line)
			}
		}
	}
	return errors.New("la operación falló")
}
