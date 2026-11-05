package init

import (
	"bytes"
	"path/filepath"

	assetsdesign "github.com/biggs-100/odachi/internal/assets/design"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// DesignPluginRel is where OpenCode loads a local plugin once per machine. The
// plugin is a door too: it loads in every folder and watches the design queue of
// the project the canvas was opened in. The doctor reads the same path through
// the Plugin option, so the two verbs cannot drift.
const DesignPluginRel = ".config/opencode/plugins/odachi-design/index.ts"

// installDesignPlugin writes the push half of the canvas: the plugin that reads the
// queue `odachi design open` writes and delivers it to the session. It is the global
// half because a plugin lives once per machine; a project without it still has the
// canvas, and its comments are read from the folder.
func (r *runner) installDesignPlugin() {
	if !r.hasClient(persona.SelectionOpenCode) || r.home == nil {
		return
	}
	home, err := r.home.Home()
	if err != nil {
		r.lines = append(r.lines, "design plugin: not installed, the home could not be read: "+oneLine(err.Error()))
		return
	}

	path := filepath.Join(home, filepath.FromSlash(DesignPluginRel))
	content := assetsdesign.Plugin()
	exists, err := r.files.Exists(path)
	if err != nil {
		r.lines = append(r.lines, "design plugin: not installed, the path could not be checked: "+oneLine(err.Error()))
		return
	}
	if exists {
		current, err := r.files.Read(path)
		if err == nil && bytes.Equal(current, content) {
			r.lines = append(r.lines, "design plugin: unchanged")
			return
		}
		if _, err := r.files.Backup(path); err != nil {
			r.lines = append(r.lines, "design plugin: not installed, the previous copy could not be backed up: "+oneLine(err.Error()))
			return
		}
	}
	if err := r.files.Write(path, content); err != nil {
		r.lines = append(r.lines, "design plugin: not installed: "+oneLine(err.Error()))
		return
	}
	r.changed = append(r.changed, path)
	r.lines = append(r.lines, "design plugin: "+path+" (delivers canvas comments to the session)")
}

// hasClient reports whether this run selected one client.
func (r *runner) hasClient(agent string) bool {
	for _, client := range r.clients {
		if client.Agent == agent {
			return true
		}
	}
	return false
}
