package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
	benchapp "github.com/biggs-100/odachi/internal/app/bench"
	"github.com/biggs-100/odachi/internal/app/start"
)

// runBench parses the flags and wires the two adapters. The verb runs in temp
// directories, so the current project is not read: `dir` is accepted for the
// dispatcher and nothing else.
func runBench(_ string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi bench", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		binary   string
		journeys string
		asJSON   bool
	)

	flags.StringVar(&binary, "binary", "", "el binario que corre cada recorrido (default: el que está corriendo)")
	flags.StringVar(&journeys, "journeys", "", "solo estos recorridos, separados por coma")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi bench -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi bench solo toma flags (corré odachi bench -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	// The default binary is the one that is running: a person who installed a
	// release measures the release, without being told which path to pass.
	if strings.TrimSpace(binary) == "" {
		binary = selfExe()
	}

	result := benchapp.Run(benchapp.Options{
		Binary: binary,
		Only:   list(journeys),
		Env:    os.Environ(),
	}, osSandbox{}, osExec{})

	return reportBench(result, asJSON, stdout, stderr)
}

// list splits a comma-separated flag value. An empty value means "no filter".
func list(value string) []string {
	items := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// reportBench prints the report. The counts and the dead ends are the deliverable,
// so a bench that found a dead end prints on stderr and a clean one on stdout,
// the same way the other verbs separate "listo" from "bloqueado".
func reportBench(result benchapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		journeys := make([]jsonJourney, 0, len(result.Journeys))
		for _, journey := range result.Journeys {
			journeys = append(journeys, jsonJourney{
				Name:      journey.Name,
				Steps:     journey.Steps,
				Blocks:    journey.Blocks,
				Surprises: journey.Surprises,
			})
		}
		deadEnds := make([]jsonDeadEnd, 0, len(result.DeadEnds))
		for _, dead := range result.DeadEnds {
			deadEnds = append(deadEnds, jsonDeadEnd{
				Journey: dead.Journey,
				Command: dead.Command,
				Exit:    dead.Exit,
				Message: dead.Message,
			})
		}

		payload := struct {
			Code        int           `json:"code"`
			Message     string        `json:"message"`
			Journeys    []jsonJourney `json:"journeys"`
			Steps       int           `json:"steps"`
			Blocks      int           `json:"blocks"`
			InBand      int           `json:"in_band"`
			Interactive int           `json:"interactive"`
			OutOfBand   int           `json:"out_of_band"`
			Templated   int           `json:"templated"`
			DeadEnd     int           `json:"dead_end"`
			DeadEnds    []jsonDeadEnd `json:"dead_ends"`
		}{
			Code:      result.Code,
			Message:   result.Message,
			Journeys:  journeys,
			Steps:     result.Steps,
			Blocks:    result.Blocks,
			InBand:    result.InBand,
			OutOfBand: result.OutOfBand,
			Templated: result.Templated,
			DeadEnd:   result.DeadEnd,
			DeadEnds:  deadEnds,
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

// jsonJourney is one journey as a machine reads it.
type jsonJourney struct {
	Name      string   `json:"name"`
	Steps     int      `json:"steps"`
	Blocks    int      `json:"blocks"`
	Surprises []string `json:"surprises,omitempty"`
}

// jsonDeadEnd is one block a human has to look at: the command that did not
// resolve it and what that command said.
type jsonDeadEnd struct {
	Journey string `json:"journey"`
	Command string `json:"command"`
	Exit    int    `json:"exit"`
	Message string `json:"message"`
}

// osSandbox is the sandbox adapter for `odachi bench`: the one place the verb
// creates and removes directories. Writes are atomic, so a bench that is killed
// leaves a directory behind instead of half a file.
type osSandbox struct{}

func (osSandbox) MkdirTemp(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}

func (osSandbox) MkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (osSandbox) Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return fsx.WriteAtomic(path, data, 0o644)
}

// RemoveAll is the promise that a journey leaves the machine as it found it.
func (osSandbox) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// osExec is the process adapter for `odachi bench`: the one place the verb runs a
// program. stdout and stderr are combined because a human reads one stream, and
// the classification is about what the tool said, not about which stream it chose.
//
// Stdin is an empty pipe and never a terminal: a journey must not depend on
// someone typing, and the tool takes its machine path because of it.
type osExec struct{}

func (osExec) Run(dir string, env []string, name string, args ...string) (string, int, error) {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = env
	command.Stdin = strings.NewReader("")

	var out bytes.Buffer
	command.Stdout = &out
	command.Stderr = &out

	err := command.Run()
	var exited *exec.ExitError
	switch {
	case err == nil:
		return out.String(), 0, nil
	case errors.As(err, &exited):
		return out.String(), exited.ExitCode(), nil
	default:
		return out.String(), -1, err
	}
}
