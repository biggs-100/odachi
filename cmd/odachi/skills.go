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

	flags.BoolVar(&check, "check", false, "validate each skill and SKILLS.md; exits 1 when something does not comply")
	flags.BoolVar(&write, "write", false, "regenerate SKILLS.md from the frontmatter")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi skills -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi skills only takes flags (run odachi skills -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}
	if check && write {
		fmt.Fprintln(stderr, "odachi: --check and --write do not go together: use one of the two (odachi skills --check or odachi skills --write)")
		return start.CodeBlocked
	}

	return reportSkills(skillsapp.Run(skillsapp.Options{Dir: dir, Check: check, Write: write}, disk(dir)), asJSON, stdout, stderr)
}

// runSkill parses `odachi skill new <slug>`. `skill` has one subcommand on
// purpose: creating one skill and touching nothing else.
func runSkill(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "odachi: missing the subcommand: the only one is odachi skill new <slug>")
		return start.CodeBlocked
	}
	if args[0] != "new" {
		fmt.Fprintf(stderr, "odachi: unknown subcommand: %s; the only one is odachi skill new <slug>\n", args[0])
		return start.CodeBlocked
	}

	flags := flag.NewFlagSet("odachi skill new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi skill new -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() == 0 {
		fmt.Fprintln(stderr, "odachi: missing the slug: odachi skill new <slug>")
		return start.CodeBlocked
	}
	if flags.NArg() > 1 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi skill new takes a single slug\n", flags.Arg(1))
		return start.CodeBlocked
	}

	return reportSkills(skillsapp.New(skillsapp.Options{Dir: dir, Slug: flags.Arg(0)}, disk(dir)), asJSON, stdout, stderr)
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
			fmt.Fprintln(stderr, "odachi: could not build the JSON output: "+oneLine(err.Error()))
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
