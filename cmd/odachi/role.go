package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	roleapp "github.com/biggs-100/odachi/internal/app/role"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/agent"
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

	inputs, code, message := roleContext(dir)
	if code != start.CodeOK {
		fmt.Fprintln(stderr, "odachi: "+message)
		return code
	}
	if sub == "" {
		return reportRole(roleapp.List(dir, inputs.block, inputs.clients, inputs.dest, disk(dir)), asJSON, stdout, stderr)
	}
	if len(args) > 0 {
		fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi role takes a single slug\n", args[0])
		return start.CodeBlocked
	}
	return reportRole(roleapp.Render(dir, sub, inputs.block, inputs.clients, inputs.dest, disk(dir)), asJSON, stdout, stderr)
}

// roleInputs is what a render and a listing both need, resolved once: the block
// the agent will read, the clients the project chose, and the folder a role's
// skill goes to.
type roleInputs struct {
	block   string
	clients []string
	dest    roleapp.SkillDest
}

// roleContext resolves what a render needs: the block the agent will read, with
// the voice the profile names, the clients the project chose, and the destination
// of a role's skill, from the scope the project has.
func roleContext(dir string) (roleInputs, int, string) {
	name := "novato"
	if loaded, code, message := loadProfile(dir, "odachi role"); code != start.CodeOK {
		return roleInputs{}, code, message
	} else if loaded != nil {
		name = loaded.Name
	}

	voice, err := persona.VoiceName(name)
	if err != nil {
		return roleInputs{}, start.CodeBlocked, oneLine(err.Error()) + "; run odachi role again"
	}
	// The role mirrors the block the project already carries: the same words an
	// agent reads in AGENTS.md, because two texts about where the craft lives are
	// one text too many. A project with no block gets the block for its voice.
	block, err := projectBlock(dir, voice)
	if err != nil {
		return roleInputs{}, start.CodeBlocked, "could not build the rules block: " + oneLine(err.Error()) + "; run odachi role again"
	}

	stored, err := agentSelection{}.Written(dir)
	if err != nil {
		return roleInputs{}, start.CodeBlocked, "could not read " + persona.SelectionPath + ": " + oneLine(err.Error()) + "; fix the file or run odachi init --agents both, and run odachi role again"
	}
	_, resolved, err := persona.ResolveSelection(stored.Agents)
	if err != nil {
		return roleInputs{}, start.CodeBlocked, oneLine(err.Error()) + "; fix the client selection in the clone's store and run odachi role again"
	}
	clients := make([]string, 0, len(resolved))
	for _, client := range resolved {
		clients = append(clients, client.Agent)
	}

	dest, code, message := roleSkillDest(dir, stored.Skills)
	if code != start.CodeOK {
		return roleInputs{}, code, message
	}
	return roleInputs{block: block, clients: clients, dest: dest}, start.CodeOK, ""
}

// roleSkillDest is the folder a role's skill goes to: the one every client reads,
// in the scope this project has. A role that wrote into `.agents/skills` whatever
// the scope would hand the adoption inference a copy odachi wrote itself, and the
// next plain init would install the whole curriculum there.
func roleSkillDest(dir, stored string) (roleapp.SkillDest, int, string) {
	adopted, code, message := adoptionHere(dir, stored)
	if code != start.CodeOK {
		return roleapp.SkillDest{}, code, message
	}
	// The same function init asks, so the two verbs cannot answer differently
	// about the same project.
	scope, _ := persona.SkillsScope{Stored: stored, Adopted: adopted}.Resolve()

	row, err := agent.Resolve(agent.Standard)
	if err != nil {
		return roleapp.SkillDest{}, start.CodeBlocked, "could not resolve the folder every client reads: " + oneLine(err.Error()) + "; download odachi again"
	}
	if scope != persona.SkillsHome {
		in := row.In(agent.ScopeProject)
		return roleapp.SkillDest{Root: dir, Base: in.Dest, Label: in.Display()}, start.CodeOK, ""
	}

	home, err := osHome{}.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		reason := "the system reported no home"
		if err != nil {
			reason = oneLine(err.Error())
		}
		return roleapp.SkillDest{}, start.CodeBlocked, "could not find your home, and this project keeps its craft there: " + reason + "; run odachi role again"
	}
	in := row.In(agent.ScopeHome)
	return roleapp.SkillDest{Root: home, Base: in.Dest, Label: in.Display()}, start.CodeOK, ""
}

// adoptionHere is the inference `init` draws, asked by a verb that only reads: a
// repository that already lives, no scope recorded, and no copy of the craft here.
// A folder that is not a repository, or has nothing committed yet, is not an
// adoption -that is a state- and a check that fails is refused, never read as an
// absence (0125's rule).
func adoptionHere(dir, stored string) (bool, int, string) {
	if stored != "" {
		return false, start.CodeOK, ""
	}
	gitPath := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		if os.IsNotExist(err) {
			return false, start.CodeOK, ""
		}
		return false, start.CodeBlocked, "could not check .git: " + oneLine(err.Error()) + "; check the permissions and run odachi role again"
	}
	lives, err := repositoryLives(dir, false)
	if err != nil || !lives {
		if err != nil {
			return false, start.CodeBlocked, oneLine(err.Error())
		}
		return false, start.CodeOK, ""
	}
	copyPath := filepath.Join(dir, filepath.FromSlash(".agents/skills"))
	if _, err := os.Stat(copyPath); err != nil {
		if os.IsNotExist(err) {
			return true, start.CodeOK, ""
		}
		return false, start.CodeBlocked, "could not check .agents/skills: " + oneLine(err.Error()) + "; check the permissions and run odachi role again"
	}
	// The copy is here: the craft lives in this project, and the role writes here.
	return false, start.CodeOK, ""
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
