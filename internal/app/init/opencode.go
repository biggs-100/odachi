package init

import (
	"path"
	"path/filepath"
	"strings"

	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// installOpenCodeDefault makes the door the agent a session starts with. The
// setting lives in OpenCode's global config, beside the agents folder; the door
// is opt-in (--global), so this write is too. A default the human picked is
// named, never replaced: installing a door is not hijacking a selector.
func (r *runner) installOpenCodeDefault() {
	if r.home == nil {
		return
	}

	selected := false
	for _, client := range r.clients {
		if client.Agent == "opencode" {
			selected = true
			break
		}
	}
	if !selected {
		return
	}

	home, err := r.home.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		// installGlobalPersonas already reported why the home is missing.
		return
	}

	door, ok := persona.GlobalFilesFor("opencode")
	if !ok {
		return
	}

	// The default points at the door: writing it for a door that was not
	// installed would name an agent the client cannot find. The persona half
	// already said why the door is missing.
	doorPath := filepath.Join(home, filepath.FromSlash(door.Path))
	doorExists, err := r.files.Exists(doorPath)
	if err != nil || !doorExists {
		return
	}

	// The config lives one level above the agents folder: ~/.config/opencode.
	rel := path.Join(path.Dir(path.Dir(door.Path)), "opencode.json")
	jsoncRel := strings.TrimSuffix(rel, ".json") + ".jsonc"

	target := filepath.Join(home, filepath.FromSlash(rel))
	display := "~/" + rel

	exists, err := r.files.Exists(target)
	if err != nil {
		r.pendGlobal("could not check " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init --global again")
		return
	}
	if !exists {
		jsoncPath := filepath.Join(home, filepath.FromSlash(jsoncRel))
		jsoncExists, err := r.files.Exists(jsoncPath)
		if err != nil {
			r.pendGlobal("could not check ~/" + jsoncRel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init --global again")
			return
		}
		if jsoncExists {
			target = jsoncPath
			display = "~/" + jsoncRel
			exists = true
		}
	}

	var data []byte
	if exists {
		data, err = r.files.Read(target)
		if err != nil {
			r.pendGlobal("could not read " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init --global again")
			return
		}
	}

	updated, outcome := domainopencode.EnsureDefaultAgent(data, persona.AgentName)
	switch outcome.Status {
	case domainopencode.Added:
		if exists {
			if _, err := r.files.Backup(target); err != nil {
				r.pendGlobal("could not back up " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init --global again")
				return
			}
		}
		if err := r.files.Write(target, updated); err != nil {
			r.pendGlobal("could not write " + display + ": " + oneLine(err.Error()) + "; check the permissions and run odachi init --global again")
			return
		}
		r.lines = append(r.lines, display+": "+persona.AgentName+" is the default agent")
	case domainopencode.Already:
		r.lines = append(r.lines, display+": "+persona.AgentName+" is already the default agent")
	case domainopencode.Chosen:
		r.lines = append(r.lines, display+": kept your default agent ("+outcome.Agent+"); edit the file to choose "+persona.AgentName)
	case domainopencode.Unreadable:
		r.pendGlobal("could not set the default agent in " + display + ": " + outcome.Reason + "; add \"default_agent\": \"" + persona.AgentName + "\" by hand")
	}
}
