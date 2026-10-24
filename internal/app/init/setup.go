package init

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// RunSetup installs the half a machine needs before a project exists: the
// bootstrap skill in the home, the door for the agent clients this machine
// already has, and OpenCode's default agent. It reads and writes nothing in the
// current folder, so the installer can run it anywhere and the human's first
// conversation finds an agent that knows the rest.
func RunSetup(opts Options, files Files, skills Skills) Result {
	r := &runner{dir: ".", files: files, skills: skills, profiles: opts.Profiles, home: opts.Home}
	if r.home == nil {
		return blocked("there is no home to install the global half: run odachi setup again")
	}

	selection, clients, problem := r.setupClients(opts.Agents)
	if problem != nil {
		return *problem
	}
	r.selection = selection
	r.clients = clients

	r.installGlobalPersonas()
	r.installOpenCodeDefault()
	r.installSkills()

	if len(clients) == 0 {
		r.lines = append(r.lines, "no agent client found in the home: the bootstrap skill is installed; install OpenCode and run odachi setup again")
	}

	result := r.result()
	result.Message = "odachi is ready: open your agent and ask for the project"
	return result
}

// setupClients is the selection setup uses: the flag when the human named one,
// otherwise the clients this machine already has. A client that is not here
// gets no door: an installer does not create another tool's config folder.
func (r *runner) setupClients(requested string) (string, []persona.Client, *Result) {
	if strings.TrimSpace(requested) != "" {
		selection, clients, err := persona.ResolveSelection(requested)
		if err != nil {
			problem := blocked("the client selection is not valid: " + oneLine(err.Error()) + "; run odachi setup --agents opencode")
			return "", nil, &problem
		}
		return selection, clients, nil
	}

	home, err := r.home.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		reason := "the system reported no home"
		if err != nil {
			reason = oneLine(err.Error())
		}
		problem := blocked("could not resolve the home to install the global half: " + reason + "; run odachi setup again")
		return "", nil, &problem
	}

	found := make([]persona.Client, 0, len(persona.Clients()))
	names := make([]string, 0, len(found))
	for _, client := range persona.Clients() {
		if strings.TrimSpace(client.HomeDir) == "" {
			continue
		}
		// HomeDir is where the agent file lives (.config/opencode/agents); the
		// client is installed when its config directory exists, one level up.
		configDir := path.Dir(client.HomeDir)
		exists, err := r.files.Exists(filepath.Join(home, filepath.FromSlash(configDir)))
		if err != nil || !exists {
			continue
		}
		found = append(found, client)
		names = append(names, client.Agent)
	}

	selection := strings.Join(names, ",")
	if len(found) == len(persona.Clients()) {
		selection = persona.SelectionBoth
	}
	return selection, found, nil
}
