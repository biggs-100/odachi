// Package hostmodel answers which model the agent driving odachi is using.
//
// It asks the host instead of guessing. A model is an identity, and the one
// odachi invented would be a lie in the receipt: the reviewer field would name
// a model that never ran. So the answer comes from the host's own variable when
// it exists, from the OpenCode session when odachi runs under OpenCode, and
// from nowhere else. No database, no log, no default.
package hostmodel

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// EnvModel is the variable any host can set. It wins over probing because it is
// the host speaking directly; the name is odachi's, so it does not tie the tool
// to one client.
const EnvModel = "ODACHI_MODEL"

// SessionEnv is how OpenCode names the session a command runs in. It is set for
// the whole process tree, which is why a child can read it at all.
const SessionEnv = "OPENCODE_SESSION_ID"

// ExportTimeoutSeconds bounds the probe. It reads local state; a CLI that does
// not answer in this window will not answer before the reviewer starts.
const ExportTimeoutSeconds = 5

// Model answers the model of the session, or "" when this host cannot say.
// It never guesses and never fails the verb: a missing model is only fatal when
// the reviewer command asks for one.
func Model(dir string) string {
	if name := strings.TrimSpace(os.Getenv(EnvModel)); name != "" {
		return name
	}
	id := strings.TrimSpace(os.Getenv(SessionEnv))
	if id == "" {
		return ""
	}
	out, err := export(id, dir)
	if err != nil {
		return ""
	}
	return ParseSession(out)
}

// export runs the supported command that prints a session as JSON. Reading the
// CLI is a contract; reading its database would be a copy of a schema that owes
// odachi nothing.
func export(id, dir string) ([]byte, error) {
	return run(dir, "session", "export", id)
}

// Models answers the identities the host serves, or false when it cannot say.
// The catalog is the host's own list; a missing CLI or an empty answer is not a
// host that serves nothing, and the caller must be able to tell them apart.
func Models(dir string) ([]string, bool) {
	out, err := run(dir, "models")
	if err != nil {
		return nil, false
	}
	models := ParseModels(out)
	if len(models) == 0 {
		return nil, false
	}
	return models, true
}

// run asks the host CLI one question with a bounded wait. Both the session and
// the catalog are local reads; a CLI that does not answer in this window will
// not answer before the reviewer starts.
func run(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ExportTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Dir = dir
	return cmd.Output()
}

// ParseModels reads `opencode models`: one provider/name per line. A line
// without both halves is not an identity, a duplicate is the same model, and the
// answer is sorted so a message reads the same on every run.
func ParseModels(data []byte) []string {
	seen := make(map[string]bool)
	var models []string

	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		provider, name, ok := strings.Cut(line, "/")
		provider, name = strings.TrimSpace(provider), strings.TrimSpace(name)
		if !ok || provider == "" || name == "" {
			continue
		}
		model := provider + "/" + name
		if seen[model] {
			continue
		}
		seen[model] = true
		models = append(models, model)
	}

	sort.Strings(models)
	return models
}

// exported is the part of a session export this package reads. Decoding only
// the info block keeps the answer small and the coupling to one field.
type exported struct {
	Info struct {
		Model struct {
			ID         string `json:"id"`
			ProviderID string `json:"providerID"`
			Variant    string `json:"variant"`
		} `json:"model"`
	} `json:"info"`
}

// ParseSession reads the model out of a `session export` document, in the shape
// a reviewer CLI takes: provider/name, and #variant when the session used one.
// A document without both halves is no model at all.
func ParseSession(data []byte) string {
	var doc exported
	if err := json.Unmarshal(data, &doc); err != nil {
		return ""
	}

	provider := strings.TrimSpace(doc.Info.Model.ProviderID)
	name := strings.TrimSpace(doc.Info.Model.ID)
	if provider == "" || name == "" {
		return ""
	}

	model := provider + "/" + name
	if variant := strings.TrimSpace(doc.Info.Model.Variant); variant != "" {
		model += "#" + variant
	}
	return model
}
