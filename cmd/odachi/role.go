package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	roleapp "github.com/biggs-100/odachi/internal/app/role"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// runRole parses `odachi role`: `new <slug>` writes the source, `<slug>`
// renders the agent and the skill from it, and no slug lists every role.
//
// The two flags are read by hand instead of by flag.FlagSet because the slug is
// positional and `role marketing --json` has to mean the same as
// `role --json marketing`: the standard parser stops at the first positional and
// would call --json an extra argument.
func runRole(dir string, args []string, stdout, stderr io.Writer) int {
	var asJSON bool
	kept := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		case "-h", "--help":
			usage(stdout)
			return start.CodeOK
		default:
			kept = append(kept, arg)
		}
	}
	args = kept

	sub := ""
	if len(args) > 0 {
		sub = args[0]
		args = args[1:]
	}

	if sub == "new" {
		if len(args) == 0 {
			fmt.Fprintln(stderr, "odachi: missing the slug: odachi role new <slug>")
			return start.CodeBlocked
		}
		if len(args) > 1 {
			fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi role new takes a single slug\n", args[1])
			return start.CodeBlocked
		}
		return reportRole(roleapp.New(dir, args[0], disk(dir)), asJSON, stdout, stderr)
	}

	block, clients, code, message := roleContext(dir)
	if code != start.CodeOK {
		fmt.Fprintln(stderr, "odachi: "+message)
		return code
	}
	if sub == "" {
		return reportRole(roleapp.List(dir, block, clients, disk(dir)), asJSON, stdout, stderr)
	}
	if len(args) > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi role takes a single slug\n", args[0])
		return start.CodeBlocked
	}
	return reportRole(roleapp.Render(dir, sub, block, clients, disk(dir)), asJSON, stdout, stderr)
}

// roleContext resolves what a render needs: the block the agent will read, with
// the voice the profile names, and the clients the project chose.
func roleContext(dir string) (block string, clients []string, code int, message string) {
	name := "novato"
	if loaded, code, message := loadProfile(dir, "odachi role"); code != start.CodeOK {
		return "", nil, code, message
	} else if loaded != nil {
		name = loaded.Name
	}

	voice, err := persona.VoiceName(name)
	if err != nil {
		return "", nil, start.CodeBlocked, oneLine(err.Error()) + "; run odachi role again"
	}
	// The role mirrors the block the project already carries: the same words an
	// agent reads in AGENTS.md, because two texts about where the craft lives are
	// one text too many. A project with no block gets the block for its voice.
	block, err = projectBlock(dir, voice)
	if err != nil {
		return "", nil, start.CodeBlocked, "could not build the rules block: " + oneLine(err.Error()) + "; run odachi role again"
	}

	selection, code, message := storedAgents(dir, "odachi role")
	if code != start.CodeOK {
		return "", nil, code, message
	}
	_, resolved, err := persona.ResolveSelection(selection)
	if err != nil {
		return "", nil, start.CodeBlocked, oneLine(err.Error()) + "; fix the client selection in the clone's store and run odachi role again"
	}
	clients = make([]string, 0, len(resolved))
	for _, client := range resolved {
		clients = append(clients, client.Agent)
	}
	return block, clients, start.CodeOK, ""
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

// projectBlock is the block this project already carries, so a role written here
// says what AGENTS.md says. A project with no AGENTS.md, or one whose block the
// parser cannot find, gets the block for its voice, which is what init would
// write there: those are states. A file that is there and cannot be read is not
// the same fact, so it is refused instead of written over.
func projectBlock(dir, voice string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return rules.Body(voice)
		}
		return "", fmt.Errorf("could not read AGENTS.md: %s", oneLine(err.Error()))
	}
	if body, ok := rules.BodyIn(string(data)); ok {
		return body, nil
	}
	return rules.Body(voice)
}
