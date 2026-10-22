package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	doctorapp "github.com/biggs-100/odachi/internal/app/doctor"
	"github.com/biggs-100/odachi/internal/app/start"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/profile"
)

// runDoctor parses the flags and wires the three adapters. The command takes no
// positional arguments and writes nothing, so there is nothing else to check.
func runDoctor(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi doctor", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi doctor -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi doctor no toma argumentos (corré odachi doctor -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	result := doctorapp.Run(doctorapp.Options{
		Dir:     dir,
		Version: version,
		Exe:     selfExe(),
	}, osFiles{}, osLookup{}, osProfiles{})

	return reportDoctor(result, asJSON, stdout, stderr)
}

// reportDoctor prints the ten lines and the summary. The report is the whole
// deliverable, so a broken install prints it on stderr and a clean one on
// stdout, the same way the other verbs separate "listo" from "bloqueado".
func reportDoctor(result doctorapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		// Lines is the block verbatim, so --json never says less than the human
		// report: a folder that is not a project yet has one situation and one
		// step and no checks, and a machine reads exactly that.
		payload := struct {
			Code     int         `json:"code"`
			Message  string      `json:"message"`
			Problems int         `json:"problems"`
			Warnings int         `json:"warnings"`
			Lines    []string    `json:"lines"`
			Checks   []jsonCheck `json:"checks"`
		}{
			Code:     result.Code,
			Message:  result.Message,
			Problems: result.Problems,
			Warnings: result.Warnings,
			Lines:    result.Lines,
			Checks:   jsonChecks(result.Checks),
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	out := stdout
	if result.Code != start.CodeOK {
		out = stderr
	}
	for _, line := range result.Lines {
		fmt.Fprintln(out, line)
	}
	return result.Code
}

// jsonCheck is one check as a machine reads it. The fields are the same ones the
// human line carries, so --json never says less than the block.
type jsonCheck struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
	Fix    string `json:"fix,omitempty"`
}

func jsonChecks(checks []domaindoctor.Check) []jsonCheck {
	out := make([]jsonCheck, 0, len(checks))
	for _, check := range checks {
		out = append(out, jsonCheck{
			Number: check.Number,
			Title:  check.Title,
			Status: check.Status.String(),
			Detail: check.Detail,
			Fix:    check.Fix,
		})
	}
	return out
}

// osLookup is the PATH adapter for `odachi doctor`: the one place the command
// asks a shell whether a name resolves. It is a port so a test can take `odachi`
// off the PATH without touching the machine it runs on.
type osLookup struct{}

func (osLookup) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// osProfiles answers, for `odachi doctor`, the same question the CLI answers
// before every verb that reads a profile: which profile is in force, and does it
// resolve? A missing file is not an error — the profile is inferred from what the
// tool can see, which is exactly what `start` and `level` do.
type osProfiles struct{}

func (osProfiles) Resolve(dir string) (string, bool, error) {
	// One owner for the decision. Answering this twice is how the two answers
	// drift apart, and doctor exists to report what the verbs would do.
	loaded, code, message := loadProfile(dir, "odachi doctor")
	if code != start.CodeOK {
		return "", false, errors.New(message)
	}
	if loaded == nil {
		name, err := inferProfile(dir)
		if err != nil {
			return "", false, err
		}
		return name, true, nil
	}
	return loaded.Name, false, nil
}

// inferProfile is what the profile would be with no file at all: the inference
// `start` and `level` already use.
func inferProfile(dir string) (string, error) {
	files := osFiles{}

	hasBrief, err := files.Exists(filepath.Join(dir, "brief.md"))
	if err != nil {
		return "", fmt.Errorf("no pude revisar brief.md: %w", err)
	}
	hasRepo, err := files.Exists(filepath.Join(dir, ".git"))
	if err != nil {
		return "", fmt.Errorf("no pude revisar .git: %w", err)
	}
	return profile.Infer(profile.Facts{HasBrief: hasBrief, HasRepo: hasRepo}), nil
}

// selfExe is the path of the binary that is running. `odachi doctor` uses it to
// name the exact folder to add to the PATH: the advice that leaves a beginner
// where they were is not advice.
func selfExe() string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}
	return self
}
