package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	roleapp "github.com/biggs-100/odachi/internal/app/role"
	"github.com/biggs-100/odachi/internal/app/start"
)

// runRole parses `odachi role`: one subcommand today, `new <slug>`, because
// writing the source is the first half of the loop and rendering is the second.
func runRole(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "odachi: missing the slug: odachi role new <slug>")
		return start.CodeBlocked
	}
	if args[0] != "new" {
		fmt.Fprintf(stderr, "odachi: unknown subcommand: %s; today the only one is odachi role new <slug>\n", args[0])
		return start.CodeBlocked
	}

	flags := flag.NewFlagSet("odachi role new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	flags.BoolVar(&asJSON, "json", false, "machine output")

	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi role new -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() == 0 {
		fmt.Fprintln(stderr, "odachi: missing the slug: odachi role new <slug>")
		return start.CodeBlocked
	}
	if flags.NArg() > 1 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi role new takes a single slug\n", flags.Arg(1))
		return start.CodeBlocked
	}

	return reportRole(roleapp.New(dir, flags.Arg(0), osFiles{}), asJSON, stdout, stderr)
}

// reportRole prints the result of `odachi role`.
func reportRole(result roleapp.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int      `json:"code"`
			Message string   `json:"message"`
			Lines   []string `json:"lines,omitempty"`
			Files   []string `json:"files"`
		}{Code: result.Code, Message: result.Message, Lines: result.Lines, Files: files}

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
