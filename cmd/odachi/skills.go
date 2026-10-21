package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
	"github.com/biggs-100/odachi/internal/app/start"
)

// runSkills parses the flags and wires the Files adapter: listing, --check and
// --write are three intents of one command.
func runSkills(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi skills", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		check  bool
		write  bool
		asJSON bool
	)

	flags.BoolVar(&check, "check", false, "valida cada skill y SKILLS.md; sale con 1 si algo no cumple")
	flags.BoolVar(&write, "write", false, "regenera SKILLS.md desde los frontmatter")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi skills -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi skills solo toma flags (corré odachi skills -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}
	if check && write {
		fmt.Fprintln(stderr, "odachi: --check y --write no van juntos: usá uno de los dos (odachi skills --check o odachi skills --write)")
		return start.CodeBlocked
	}

	return reportSkills(skillsapp.Run(skillsapp.Options{Dir: dir, Check: check, Write: write}, osFiles{}), asJSON, stdout, stderr)
}

// runSkill parses `odachi skill new <slug>`. `skill` has one subcommand on
// purpose: creating one skill and touching nothing else.
func runSkill(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "odachi: falta el subcomando: el único es odachi skill new <slug>")
		return start.CodeBlocked
	}
	if args[0] != "new" {
		fmt.Fprintf(stderr, "odachi: subcomando desconocido: %s; el único es odachi skill new <slug>\n", args[0])
		return start.CodeBlocked
	}

	flags := flag.NewFlagSet("odachi skill new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi skill new -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() == 0 {
		fmt.Fprintln(stderr, "odachi: falta el slug: odachi skill new <slug>")
		return start.CodeBlocked
	}
	if flags.NArg() > 1 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi skill new toma un solo slug\n", flags.Arg(1))
		return start.CodeBlocked
	}

	return reportSkills(skillsapp.New(skillsapp.Options{Dir: dir, Slug: flags.Arg(0)}, osFiles{}), asJSON, stdout, stderr)
}

// skillEntry is one row of the machine listing.
type skillEntry struct {
	Name    string `json:"name"`
	Trigger string `json:"trigger"`
	Path    string `json:"path"`
}

// reportSkills prints the skills result. Unlike the other verbs it prints every
// line on a failed check, because each line is a different file a human has to
// fix; a machine gets the same list under "lines".
func reportSkills(result skillsapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		entries := make([]skillEntry, 0, len(result.Skills))
		for _, s := range result.Skills {
			name := s.Name
			if name == "" {
				name = s.Dir
			}
			entries = append(entries, skillEntry{Name: name, Trigger: s.Trigger, Path: s.Path})
		}

		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int          `json:"code"`
			Message string       `json:"message"`
			Lines   []string     `json:"lines,omitempty"`
			Files   []string     `json:"files"`
			Skills  []skillEntry `json:"skills"`
		}{
			Code:    result.Code,
			Message: result.Message,
			Lines:   result.Lines,
			Files:   files,
			Skills:  entries,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	lines := result.Lines
	if len(lines) == 0 {
		lines = []string{result.Message}
	}
	out := stdout
	if result.Code != start.CodeOK {
		out = stderr
	}
	for _, line := range lines {
		if result.Code != start.CodeOK {
			line = "odachi: " + line
		}
		fmt.Fprintln(out, line)
	}
	return result.Code
}
