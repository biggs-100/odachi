package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// runSkills parses the flags and wires the Files adapter: listing, --check and
// --write are three intents of one command.
func runSkills(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi skills", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		check   bool
		write   bool
		resolve string
		asJSON  bool
	)

	flags.BoolVar(&check, "check", false, "validate each skill and SKILLS.md; exits 1 when something does not comply")
	flags.BoolVar(&write, "write", false, "regenerate SKILLS.md from the frontmatter")
	flags.StringVar(&resolve, "resolve", "", "the skills a step of the current feature would receive, with their path and scope")
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
	// Visit tells "the flag was absent" apart from "the flag was given empty":
	// `--resolve "$STEP"` with an unset variable must not fall through to the
	// plain listing and report a hand-off nobody asked for.
	resolveSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "resolve" {
			resolveSet = true
		}
	})
	if resolveSet {
		if strings.TrimSpace(resolve) == "" {
			fmt.Fprintln(stderr, "odachi: --resolve needs the step: odachi skills --resolve T1")
			return start.CodeBlocked
		}
		other := ""
		switch {
		case check:
			other = "--check"
		case write:
			other = "--write"
		case asJSON:
			other = "--json"
		}
		if other != "" {
			fmt.Fprintf(stderr, "odachi: --resolve and %s do not go together: use one of them (odachi skills --resolve <step>, or odachi skills %s)\n", other, other)
			return start.CodeBlocked
		}
		return resolveSkills(dir, resolve, stdout, stderr)
	}

	return reportSkills(skillsapp.Run(skillsapp.Options{Dir: dir, Check: check, Write: write}, disk(dir)), asJSON, stdout, stderr)
}

// resolveSkills answers what a step of the current feature would receive: the
// exact SKILL.md each name resolves to and the scope it came from. It is the
// read-only twin of `delegate`, and it writes nothing.
func resolveSkills(dir, stepID string, stdout, stderr io.Writer) int {
	current, err := tasks.Current(disk(dir), dir)
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	stepID = strings.ToUpper(strings.TrimSpace(stepID))
	var step taskdoc.Step
	found := false
	for _, candidate := range current.Doc.Steps {
		if candidate.ID == stepID {
			step = candidate
			found = true
			break
		}
	}
	if !found {
		if len(current.Doc.Steps) == 0 {
			fmt.Fprintf(stderr, "odachi: %s has no steps yet: run odachi plan \"<the step>\" first\n", current.Rel)
			return start.CodeBlocked
		}
		ids := make([]string, 0, len(current.Doc.Steps))
		for _, candidate := range current.Doc.Steps {
			ids = append(ids, candidate.ID)
		}
		fmt.Fprintf(stderr, "odachi: %s is not a step of %s: the steps are %s\n", stepID, current.Rel, strings.Join(ids, ", "))
		return start.CodeBlocked
	}

	if len(step.Skills) == 0 {
		fmt.Fprintf(stdout, "%s names no skills\n", step.ID)
		return start.CodeOK
	}

	blocked := false
	for _, name := range step.Skills {
		found, ok := craftPath(dir, name)
		if !ok {
			looked := craftPaths(dir, name)
			places := make([]string, 0, len(looked))
			for _, path := range looked {
				places = append(places, filepath.ToSlash(path))
			}
			fmt.Fprintf(stderr, "odachi: %s has no SKILL.md in %s\n", name, strings.Join(places, " or "))
			blocked = true
			continue
		}
		scope := "home"
		if found == filepath.Join(dir, ".agents", "skills", name, "SKILL.md") {
			scope = "project"
		}
		fmt.Fprintf(stdout, "%s  %s  (%s)\n", name, filepath.ToSlash(found), scope)
	}
	if blocked {
		return start.CodeBlocked
	}
	return start.CodeOK
}

// runSkill parses `odachi skill new <slug> --scope project|home`. `skill` has one
// subcommand on purpose: creating one skill and touching nothing else.
func runSkill(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "odachi: missing the subcommand: the only one is odachi skill new <slug> --scope project|home")
		return start.CodeBlocked
	}
	if args[0] != "new" {
		fmt.Fprintf(stderr, "odachi: unknown subcommand: %s; the only one is odachi skill new <slug> --scope project|home\n", args[0])
		return start.CodeBlocked
	}

	flags := flag.NewFlagSet("odachi skill new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		asJSON bool
		scope  string
	)
	flags.BoolVar(&asJSON, "json", false, "machine output")
	flags.StringVar(&scope, "scope", "", "where the skill lives: project (./.agents/skills) | home (~/.agents/skills)")

	// Go's flag package stops at the first positional, and the slug comes first:
	// split the words myself so `skill new ejemplo --scope home` and
	// `skill new --scope home ejemplo` both work.
	rest := args[1:]
	var flagArgs, positionals []string
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		switch {
		case arg == "--scope" || arg == "-scope":
			flagArgs = append(flagArgs, arg)
			if i+1 < len(rest) {
				i++
				flagArgs = append(flagArgs, rest[i])
			}
		case strings.HasPrefix(arg, "--scope=") || strings.HasPrefix(arg, "-scope="):
			flagArgs = append(flagArgs, arg)
		case strings.HasPrefix(arg, "-") && arg != "-":
			flagArgs = append(flagArgs, arg)
		default:
			positionals = append(positionals, arg)
		}
	}

	if err := flags.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi skill new -h to see the flags")
		return start.CodeBlocked
	}
	if len(positionals) == 0 {
		fmt.Fprintln(stderr, "odachi: missing the slug: odachi skill new <slug> --scope project|home")
		return start.CodeBlocked
	}
	if len(positionals) > 1 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi skill new takes a single slug\n", positionals[1])
		return start.CodeBlocked
	}
	slug := positionals[0]
	if scope != "" {
		if _, err := persona.NormalizeSkills(scope); err != nil {
			fmt.Fprintf(stderr, "odachi: --scope has to be \"project\" or \"home\", not %q; project writes .agents/skills/<slug>, home writes ~/.agents/skills/<slug>\n", scope)
			return start.CodeBlocked
		}
		scope = strings.ToLower(strings.TrimSpace(scope))
	}

	// The home is the CLI's to resolve: the app never touches os. A home that
	// cannot be found is refused before the skill is written.
	home := ""
	if scope == persona.SkillsHome {
		resolved, err := homeDir()
		if err != nil || strings.TrimSpace(resolved) == "" {
			reason := "the system reported no home"
			if err != nil {
				reason = oneLine(err.Error())
			}
			fmt.Fprintln(stderr, "odachi: could not find your home to write ~/.agents/skills/<slug>: "+reason+"; run odachi skill new <slug> --scope project, or set a home and run it again")
			return start.CodeBlocked
		}
		home = resolved
	}

	return reportSkills(skillsapp.New(skillsapp.Options{Dir: dir, Slug: slug, Scope: scope, Home: home}, disk(dir)), asJSON, stdout, stderr)
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
