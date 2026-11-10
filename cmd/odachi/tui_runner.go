package main

import (
	"bytes"
	"errors"
	"io"
	"strings"

	appreview "github.com/biggs-100/odachi/internal/app/review"
)

// tuiRunner applies one operation of the wizard by calling the verb that already
// owns it. The functions are fields so a test can drive the branch without
// touching the home or the project.
type tuiRunner struct {
	dir     string
	initFn  func(dir string, args []string, stdout, stderr io.Writer) int
	setupFn func(args []string, stdout, stderr io.Writer) int
	levelFn func(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int
	modeFn  func(dir string, args []string, stdout, stderr io.Writer) int
}

// newTUIRunner wires the runner to the same functions the CLI dispatches.
func newTUIRunner(dir string) tuiRunner {
	return tuiRunner{
		dir:     dir,
		initFn:  runInit,
		setupFn: runSetup,
		levelFn: runLevel,
		modeFn:  runReviewMode,
	}
}

// Install runs init when the human chose the project, setup when they chose the
// machine alone. The clients and the skills scope are the two choices init
// records; setup writes no store, which is why a folder without a repository is
// asked which one it wants.
func (r tuiRunner) Install(project bool, clients, skills string) error {
	var out, errOut bytes.Buffer
	code := 0
	if project {
		code = r.initFn(r.dir, []string{"--agents", clients, "--skills", skills}, &out, &errOut)
	} else {
		code = r.setupFn([]string{"--agents", clients}, &out, &errOut)
	}
	return runnerError(code, out.String(), errOut.String())
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
