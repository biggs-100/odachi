// Package opencode edits the two OpenCode settings odachi owns: which agent a
// session starts with, and the model it runs. The file belongs to the human, so
// every edit goes through jsonmember, which inserts or replaces one member and
// never re-encodes the document: comments, key order and line endings are part
// of what the human wrote, not formatting odachi may replace.
package opencode

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/jsonmember"
)

const (
	// SchemaURL is what a config created from nothing declares.
	SchemaURL = "https://opencode.ai/config.json"
	// defaultAgentKey is the member that names the agent a session starts with.
	defaultAgentKey = "default_agent"
	// modelKey is the root member OpenCode reads the session model from.
	modelKey = "model"
)

// Status is what an edit found, and what it did.
type Status int

const (
	// Added means the key was inserted.
	Added Status = iota
	// Already means the key already named the value.
	Already
	// Chosen means another agent is chosen; nothing was written.
	Chosen
	// Changed means a different model was written in place.
	Changed
	// Unreadable means there is no safe edit; nothing was written.
	Unreadable
)

// Outcome says what happened and why.
type Outcome struct {
	Status Status
	// Agent is the agent already chosen, for Chosen.
	Agent string
	// Previous is the model that was there before a Changed edit.
	Previous string
	// Reason says why the document is unreadable, for Unreadable.
	Reason string
}

// StripComments replaces JSONC comments with spaces, keeping strings and byte
// positions intact. It is how the editor validates its own result and how a
// caller can read a document that carries comments.
func StripComments(data []byte) []byte {
	return jsonmember.StripComments(data)
}

// EnsureDefaultAgent returns the document with default_agent set to agent when
// it does not choose one. It never rewrites the file: it inserts one member
// after the root brace, so comments, order and line endings survive.
//
// When another agent is already chosen the document is returned untouched and
// the outcome says which one: odachi installs its door, it does not hijack a
// default the human picked.
func EnsureDefaultAgent(data []byte, agent string) ([]byte, Outcome) {
	agent = strings.TrimSpace(agent)
	if agent == "" {
		return data, Outcome{Status: Unreadable, Reason: "no agent name"}
	}

	stripped := jsonmember.StripComments(data)
	if len(bytes.TrimSpace(stripped)) == 0 {
		return freshDoc(agent), Outcome{Status: Added}
	}

	current, found, err := CurrentAgent(data)
	if err != nil {
		return data, Outcome{Status: Unreadable, Reason: err.Error()}
	}
	if found {
		if current == agent {
			return data, Outcome{Status: Already}
		}
		return data, Outcome{Status: Chosen, Agent: current}
	}

	// The key is absent: Ensure is the insertion path. It reads and inserts, and
	// the read above is what decided that no human default is being replaced.
	updated, outcome := jsonmember.Ensure(data, defaultAgentKey, agent)
	return updated, Outcome{Status: statusOf(outcome), Reason: outcome.Reason}
}

// CurrentAgent answers the default agent the document names. found is false
// when the key is absent; a document with no safe reading is an error. It is
// the read-only half doctor uses: a question never writes.
func CurrentAgent(data []byte) (agent string, found bool, err error) {
	return jsonmember.Read(data, defaultAgentKey)
}

// EnsureModel returns the document with the root model set to model. The flag
// that calls this names the value explicitly, so a different one is replaced in
// place: only the string's content changes, and every byte around it survives.
// Same value, nothing to do; a document with no safe reading is never written.
func EnsureModel(data []byte, model string) ([]byte, Outcome) {
	checked, err := jsonmember.NormalizeIdentifier(model)
	if err != nil {
		return data, Outcome{Status: Unreadable, Reason: err.Error()}
	}

	updated, outcome := jsonmember.Ensure(data, modelKey, checked)
	mapped := Outcome{Status: statusOf(outcome), Reason: outcome.Reason}
	if outcome.Status == jsonmember.Replaced {
		mapped.Status = Changed
		mapped.Previous = outcome.Previous
		return updated, mapped
	}
	if outcome.Status == jsonmember.Already {
		// An already-right value keeps the original bytes, the way the agent does.
		return data, mapped
	}
	return updated, mapped
}

// statusOf maps the shared editor's vocabulary onto this package's, so the two
// halves of a config report the same four outcomes they always had. Replaced is
// the caller's decision: only EnsureModel accepts it as a write.
func statusOf(outcome jsonmember.Outcome) Status {
	switch outcome.Status {
	case jsonmember.Added:
		return Added
	case jsonmember.Already:
		return Already
	case jsonmember.Replaced:
		return Changed
	default:
		return Unreadable
	}
}

// freshDoc is the config of a project that has none: the schema OpenCode's own
// docs publish, and the agent this install makes the default.
func freshDoc(agent string) []byte {
	quoted, _ := json.Marshal(agent)
	return []byte("{\n  \"$schema\": \"" + SchemaURL + "\",\n  \"" + defaultAgentKey + "\": " + string(quoted) + "\n}\n")
}
