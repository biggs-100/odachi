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
	// Setup is the machine half: the craft goes to the home, because before a
	// project exists the home is the only place a skill can live.
	r.skillsScope = persona.SkillsHome
	r.skillsSource = skillsFromSetup
	r.dryRun = opts.DryRun

	selection, clients, problem := r.setupClients(opts.Agents)
	if problem != nil {
		return *problem
	}
	r.selection = selection
	r.clients = clients

	r.installGlobalPersonas()
	r.installOpenCodeDefault()
	r.installSkills()

	// The clients detection did not find are named after the work: they are the
	// one thing a machine refresh leaves alone, and silence about them reads as an
	// omission.
	if len(r.absent) > 0 {
		r.lines = append(r.lines, "not installed here: "+humanList(r.absent)+"; setup does not create another tool's config folder")
	}
	// A read that failed is reported as what it is. The absent list above is a
	// fact about the machine; this one is a fact about this run.
	if len(r.unchecked) > 0 {
		r.lines = append(r.lines, "could not check: "+humanList(r.unchecked)+"; check the permissions and run odachi setup again")
	}

	if len(clients) == 0 {
		r.lines = append(r.lines, "no agent client found in the home: the bootstrap skill is installed; install OpenCode and run odachi setup again")
	}

	result := r.result()
	// A rehearsal is not ready for anything: it wrote nothing, and the message
	// from result() says what it would write. Keeping "odachi is ready" here made
	// the plan read like a fact in the one output a machine consumes.
	if !opts.DryRun {
		result.Message = "odachi is ready: open your agent and ask for the project"
	}
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
		if err != nil {
			// A read that failed is not a client that is absent: saying "not
			// installed here" about a folder odachi could not look at is a claim
			// about the machine it does not have.
			r.unchecked = append(r.unchecked, client.Agent+" (~/"+configDir+"): "+oneLine(err.Error()))
			continue
		}
		if !exists {
			// A client that is not here is named, not dropped: an installer that
			// lists three destinations and stays quiet about the fourth looks like
			// it forgot one.
			r.absent = append(r.absent, client.Agent+" (~/"+configDir+")")
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
