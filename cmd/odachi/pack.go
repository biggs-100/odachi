package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/biggs-100/odachi/internal/app/pack"
	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/skill"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// runPack parses the flags and wires the three pieces: the skills embedded in
// the binary, the Files adapter and the Home adapter. `pack` takes the project
// directory because a project-scoped agent — opencode — reads its skills from
// inside the project, because a relative --dest is resolved against it, and
// because a home-scoped agent still resolves against the home; the project's
// `skills/` and `SKILLS.md` are never touched (those belong to `odachi skills`).
func runPack(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi pack", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		agents string
		scope  string
		dest   string
		dryRun bool
		asJSON bool
	)

	flags.StringVar(&agents, "agents", "", "only these agents, comma-separated (default: the detected ones; standard = the folder every client reads)")
	flags.StringVar(&scope, "scope", "", "project | home: the same agent, in the other root (default: the one in the table)")
	flags.StringVar(&dest, "dest", "", "install everything in this folder: absolute, or relative to the project")
	flags.BoolVar(&dryRun, "dry-run", false, "what it would do, without writing anything")
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi pack -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi pack only takes flags (run odachi pack -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	embedded, err := bundledskills.All()
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error())+"; download odachi again")
		return start.CodeBlocked
	}

	payload := make([]pack.Skill, 0, len(embedded))
	for _, s := range embedded {
		payload = append(payload, pack.Skill{Name: s.Name, Content: s.Content})
	}

	// The project's own catalog: a skill created with `odachi skill new` lives
	// in ./skills and the agent reads .agents/skills, so this is the join. A
	// project without skills/ contributes nothing, and a broken catalog is
	// refused by name: installing half a skill is worse than not installing it.
	own := skillsapp.Run(skillsapp.Options{Dir: dir}, osFiles{})
	if own.Code != start.CodeOK {
		// A catalog that cannot be read is not an empty catalog: installing only
		// the bundled set would hide the project's own skills without saying so.
		fmt.Fprintln(stderr, "odachi: "+own.Message)
		return start.CodeBlocked
	}
	if problems := projectProblems(own.Skills); len(problems) > 0 {
		for _, line := range problems {
			fmt.Fprintln(stderr, "odachi: "+line)
		}
		fmt.Fprintln(stderr, "odachi: the project's catalog does not pass its check: fix it and run odachi pack again (odachi skills --check lists every problem)")
		return start.CodeBlocked
	}
	project := make([]pack.Skill, 0, len(own.Skills))
	for _, s := range own.Skills {
		project = append(project, pack.Skill{Name: s.Name, Content: s.Raw})
	}

	return reportPack(pack.Run(pack.Options{Agents: agents, Scope: scope, Dest: dest, DryRun: dryRun, Dir: dir, Skills: payload, Project: project}, osFiles{}, osHome{}), asJSON, stdout, stderr)
}

// packDest is one row of the machine report: the same numbers the summary line a
// human reads is built from. Agents is a list because --dest can put several
// named agents into one destination, and the destination is reported once.
type packDest struct {
	Agents    []string `json:"agents"`
	Dest      string   `json:"dest"`
	Scope     string   `json:"scope"`
	Installed int      `json:"installed"`
	BackedUp  int      `json:"backed_up"`
	Unchanged int      `json:"unchanged"`
	Foreign   int      `json:"foreign"`
	Failed    int      `json:"failed"`
	Files     []string `json:"files"`
}

// reportPack prints one line per action and one summary line per destination. On
// a blocked run every line goes to stderr, because each of them is a file a human
// has to look at.
func reportPack(result pack.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		dests := make([]packDest, 0, len(result.Dests))
		for _, d := range result.Dests {
			files := d.Files
			if files == nil {
				files = []string{}
			}
			agents := d.Names
			if agents == nil {
				agents = []string{}
			}
			dests = append(dests, packDest{
				Agents:    agents,
				Dest:      d.Dest,
				Scope:     d.Scope,
				Installed: d.Installed,
				BackedUp:  d.BackedUp,
				Unchanged: d.Unchanged,
				Foreign:   d.Foreign,
				Failed:    d.Failed,
				Files:     files,
			})
		}

		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int        `json:"code"`
			Message string     `json:"message"`
			DryRun  bool       `json:"dry_run"`
			Lines   []string   `json:"lines,omitempty"`
			Files   []string   `json:"files"`
			Dests   []packDest `json:"dests"`
		}{
			Code:    result.Code,
			Message: result.Message,
			DryRun:  result.DryRun,
			Lines:   result.Lines,
			Files:   files,
			Dests:   dests,
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

// projectProblems lists every violation in the project's own skills, one line
// each, so pack refuses a broken catalog by the file a human has to open
// instead of installing half a skill.
func projectProblems(loaded []skill.Skill) []string {
	lines := []string{}
	for _, s := range loaded {
		for _, problem := range s.Problems() {
			lines = append(lines, problem.File+": "+problem.Reason)
		}
	}
	return lines
}
