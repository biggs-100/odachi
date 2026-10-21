package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/biggs-100/odachi/internal/app/pack"
	"github.com/biggs-100/odachi/internal/app/start"
	bundledskills "github.com/biggs-100/odachi/skills"
)

// runPack parses the flags and wires the three pieces: the skills embedded in
// the binary, the Files adapter and the Home adapter. `pack` takes the project
// directory because a project-scoped agent — opencode — reads its skills from
// inside the project; a home-scoped one still resolves against the home, and the
// project's `skills/` and `SKILLS.md` are never touched (those belong to
// `odachi skills`).
func runPack(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi pack", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		agents string
		dryRun bool
		asJSON bool
	)

	flags.StringVar(&agents, "agents", "", "solo estos agentes, separados por coma (default: los detectados)")
	flags.BoolVar(&dryRun, "dry-run", false, "qué haría, sin escribir nada")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi pack -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi pack solo toma flags (corré odachi pack -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	embedded, err := bundledskills.All()
	if err != nil {
		fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error())+"; volvé a bajar odachi")
		return start.CodeBlocked
	}

	payload := make([]pack.Skill, 0, len(embedded))
	for _, s := range embedded {
		payload = append(payload, pack.Skill{Name: s.Name, Content: s.Content})
	}

	return reportPack(pack.Run(pack.Options{Agents: agents, DryRun: dryRun, Dir: dir, Skills: payload}, osFiles{}, osHome{}), asJSON, stdout, stderr)
}

// packAgent is one row of the machine report: the same numbers the summary line
// a human reads is built from.
type packAgent struct {
	Name      string   `json:"name"`
	Dest      string   `json:"dest"`
	Scope     string   `json:"scope"`
	Installed int      `json:"installed"`
	BackedUp  int      `json:"backed_up"`
	Unchanged int      `json:"unchanged"`
	Foreign   int      `json:"foreign"`
	Failed    int      `json:"failed"`
	Files     []string `json:"files"`
}

// reportPack prints one line per action and one summary line per agent. On a
// blocked run every line goes to stderr, because each of them is a file a human
// has to look at.
func reportPack(result pack.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		agents := make([]packAgent, 0, len(result.Agents))
		for _, a := range result.Agents {
			files := a.Files
			if files == nil {
				files = []string{}
			}
			agents = append(agents, packAgent{
				Name:      a.Name,
				Dest:      a.Dest,
				Scope:     a.Scope,
				Installed: a.Installed,
				BackedUp:  a.BackedUp,
				Unchanged: a.Unchanged,
				Foreign:   a.Foreign,
				Failed:    a.Failed,
				Files:     files,
			})
		}

		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			DryRun  bool        `json:"dry_run"`
			Lines   []string    `json:"lines,omitempty"`
			Files   []string    `json:"files"`
			Agents  []packAgent `json:"agents"`
		}{
			Code:    result.Code,
			Message: result.Message,
			DryRun:  result.DryRun,
			Lines:   result.Lines,
			Files:   files,
			Agents:  agents,
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
