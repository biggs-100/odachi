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
	ctx, cancel := context.WithTimeout(context.Background(), ExportTimeoutSeconds*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "opencode", "session", "export", id)
	cmd.Dir = dir
	return cmd.Output()
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
