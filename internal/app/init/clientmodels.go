package init

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/jsonmember"
	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// modelClients is the two clients a model can be written for, in the order the
// report names them. A model is not the --agents selection: the flag names one
// client per occurrence, and that is the file the value goes to.
var modelClients = []string{"opencode", "claude"}

// ParseModels turns the repeated `--model <client>=<model>` values into the map
// a run writes. Every refusal names the form the human has to type, because the
// flag is the only place the grammar is visible.
func ParseModels(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	raw := make(map[string]string, len(values))
	for _, value := range values {
		client, model, found := strings.Cut(value, "=")
		if !found {
			return nil, fmt.Errorf("--model takes client=model, and %q has no =; the clients are opencode and claude", value)
		}
		raw[client] = model
	}
	return NormalizeModels(raw)
}

// NormalizeModels checks the map a run was handed, whether it came from the flag
// or from a caller. An unknown client is refused naming the two; an empty value
// and a value that is not an identifier are refused because the value ends up in
// a file a client parses.
func NormalizeModels(models map[string]string) (map[string]string, error) {
	if len(models) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(models))
	for client, model := range models {
		name := strings.ToLower(strings.TrimSpace(client))
		if name != "opencode" && name != "claude" {
			return nil, fmt.Errorf("%q is not a client: --model writes opencode or claude", client)
		}
		if _, twice := out[name]; twice {
			return nil, fmt.Errorf("the client %q appears twice: write it once", name)
		}
		checked, err := jsonmember.NormalizeIdentifier(model)
		if err != nil {
			return nil, fmt.Errorf("--model %s: %w", name, err)
		}
		out[name] = checked
	}
	return out, nil
}

// installClientModels writes the model each client was told to use, in the file
// that client reads. The work happens after the door because the paths are the
// same shape, but it is not gated on the door: naming a model is a decision of
// its own, and init honours it even when --agents did not name that client.
func (r *runner) installClientModels(models map[string]string) {
	if len(models) == 0 || r.home == nil {
		return
	}
	home, err := r.home.Home()
	if err != nil || strings.TrimSpace(home) == "" {
		r.pendGlobal("could not resolve the home to write the model; run odachi init again")
		return
	}

	for _, client := range modelClients {
		model, asked := models[client]
		if !asked {
			continue
		}
		switch client {
		case "opencode":
			r.installOpenCodeModel(home, model)
		case "claude":
			r.installClaudeModel(home, model)
		}
	}
}

// installOpenCodeModel edits the config OpenCode reads: opencode.json, or the
// .jsonc that already exists beside it. A different model is replaced in place;
// the rest of the human's file is never re-encoded.
func (r *runner) installOpenCodeModel(home, model string) {
	target, display, exists, err := openCodeConfigPath(home, r.files)
	if err != nil {
		r.pendModel(display, model, err.Error())
		return
	}

	var data []byte
	if exists {
		data, err = r.files.Read(target)
		if err != nil {
			r.pendModel(display, model, "could not read it: "+err.Error())
			return
		}
	}

	updated, outcome := domainopencode.EnsureModel(data, model)
	r.writeModel(display, target, exists, updated, jsonmember.Outcome{
		Status:   modelStatus(outcome.Status),
		Previous: outcome.Previous,
		Reason:   outcome.Reason,
	}, model)
}

// installClaudeModel edits the file Claude Code reads. A file that does not
// exist yet is created as the smallest document that client reads, one line and
// a trailing newline; there is no human layout to keep in a file nobody wrote.
func (r *runner) installClaudeModel(home, model string) {
	target, display := claudeSettingsPath(home)

	exists, err := r.files.Exists(target)
	if err != nil {
		r.pendModel(display, model, "could not check it: "+err.Error())
		return
	}
	if !exists {
		// The value already passed the identifier rule, so it carries nothing a
		// JSON string has to escape.
		fresh := []byte(`{"model": "` + model + `"}` + "\n")
		r.writeModel(display, target, false, fresh, jsonmember.Outcome{Status: jsonmember.Added}, model)
		return
	}

	data, err := r.files.Read(target)
	if err != nil {
		r.pendModel(display, model, "could not read it: "+err.Error())
		return
	}
	updated, outcome := jsonmember.Ensure(data, "model", model)
	r.writeModel(display, target, true, updated, outcome, model)
}

// openCodeConfigPath resolves the config OpenCode reads. The edit goes where the
// file already lives: an existing .jsonc is the human's choice and is not
// shadowed by a second .json. The writer and the reader call this one function,
// so the value the screen shows and the value a run writes cannot name two
// different files.
func openCodeConfigPath(home string, files Files) (target, display string, exists bool, err error) {
	rel := path.Join(".config", "opencode", "opencode.json")
	if door, ok := persona.GlobalFilesFor("opencode"); ok {
		// The config lives one level above the agents folder.
		rel = path.Join(path.Dir(path.Dir(door.Path)), "opencode.json")
	}
	jsoncRel := strings.TrimSuffix(rel, ".json") + ".jsonc"

	target = filepath.Join(home, filepath.FromSlash(rel))
	display = "~/" + rel
	exists, err = files.Exists(target)
	if err != nil || exists {
		return target, display, exists, err
	}

	jsonc := filepath.Join(home, filepath.FromSlash(jsoncRel))
	jsoncExists, err := files.Exists(jsonc)
	if err != nil {
		return target, display, false, err
	}
	if jsoncExists {
		return jsonc, "~/" + jsoncRel, true, nil
	}
	return target, display, false, nil
}

// claudeSettingsPath is the file Claude Code reads, and the label a human sees.
// It is a function for the same reason as openCodeConfigPath: one path for the
// writer and the reader (0337).
func claudeSettingsPath(home string) (target, display string) {
	rel := ".claude/settings.json"
	return filepath.Join(home, filepath.FromSlash(rel)), "~/" + rel
}

// ReadClientModels answers the model each client reads from its own file today,
// which is what the models section shows before anything is written. It is a
// read: no file is created and nothing is changed. A client whose config or key
// is absent simply has no entry, and one whose file exists but cannot be read is
// named in the second map instead of failing the call: "no model" is a fact about
// the file, "could not read it" is another, and neither is a broken project.
func ReadClientModels(home string, files Files) (map[string]string, map[string]string, error) {
	if strings.TrimSpace(home) == "" {
		return nil, nil, fmt.Errorf("there is no home to read the clients' config from")
	}

	out := map[string]string{}
	// One file that cannot be read is a row the wizard names, not a reason to
	// refuse the whole verb: the section is a convenience over files the human
	// owns, and a JSONC document this reader cannot parse is not a broken project
	// (0337, the review's warning).
	unreadable := map[string]string{}

	target, display, exists, err := openCodeConfigPath(home, files)
	if err != nil {
		unreadable["opencode"] = "no se pudo comprobar " + display + ": " + err.Error()
	} else if exists {
		if model, found, err := readModel(files, target, display); err != nil {
			unreadable["opencode"] = err.Error()
		} else if found {
			out["opencode"] = model
		}
	}

	target, display = claudeSettingsPath(home)
	exists, err = files.Exists(target)
	if err != nil {
		unreadable["claude"] = "no se pudo comprobar " + display + ": " + err.Error()
	} else if exists {
		if model, found, err := readModel(files, target, display); err != nil {
			unreadable["claude"] = err.Error()
		} else if found {
			out["claude"] = model
		}
	}

	if len(out) == 0 {
		out = nil
	}
	if len(unreadable) == 0 {
		unreadable = nil
	}
	return out, unreadable, nil
}

// readModel is the read-only half the writer already owns: the same member the
// editor replaces, read where it stands. Comments are stripped by the reader
// itself, so an opencode.json with them is read without a second parser.
func readModel(files Files, target, display string) (string, bool, error) {
	data, err := files.Read(target)
	if err != nil {
		return "", false, fmt.Errorf("could not read %s: %w", display, err)
	}
	model, found, err := jsonmember.Read(data, "model")
	if err != nil {
		return "", false, fmt.Errorf("%s: %w", display, err)
	}
	return model, found, nil
}

// writeModel applies one editor outcome to the disk and the report. Only a real
// change is backed up and written; a value that is already there is named, and a
// rehearsal says what it would write without a byte moving.
func (r *runner) writeModel(display, target string, exists bool, updated []byte, outcome jsonmember.Outcome, model string) {
	switch outcome.Status {
	case jsonmember.Added, jsonmember.Replaced:
		if exists {
			if _, err := r.files.Backup(target); err != nil {
				r.pendModel(display, model, "could not back it up: "+err.Error())
				return
			}
		}
		if err := r.files.Write(target, updated); err != nil {
			r.pendModel(display, model, "could not write it: "+err.Error())
			return
		}
		r.changed = append(r.changed, display)
		if r.dryRun {
			if outcome.Status == jsonmember.Added {
				r.lines = append(r.lines, display+": would write "+model+" as the model")
			} else {
				r.lines = append(r.lines, display+": would replace "+outcome.Previous+" with "+model)
			}
			return
		}
		if outcome.Status == jsonmember.Added {
			r.lines = append(r.lines, display+": "+model+" is the model")
		} else {
			r.lines = append(r.lines, display+": the model is now "+model+", it was "+outcome.Previous)
		}
	case jsonmember.Already:
		r.lines = append(r.lines, display+": "+model+" is already the model")
	case jsonmember.Unreadable:
		r.pendModel(display, model, outcome.Reason)
	}
}

// pendModel is the one pending sentence for a model that could not be written:
// the file is named, the member to add is spelled, and the run still finishes.
func (r *runner) pendModel(display, model, reason string) {
	r.pendGlobal("could not set the model in " + display + ": " + reason + "; add \"model\": \"" + model + "\" by hand")
}

// modelStatus maps the opencode vocabulary onto the shared one the writer reads.
func modelStatus(status domainopencode.Status) jsonmember.Status {
	switch status {
	case domainopencode.Added:
		return jsonmember.Added
	case domainopencode.Already:
		return jsonmember.Already
	case domainopencode.Changed:
		return jsonmember.Replaced
	default:
		return jsonmember.Unreadable
	}
}
