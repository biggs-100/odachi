// Package opencode edits the one OpenCode setting odachi owns: which agent a
// session starts with. The file belongs to the human, so the edit inserts a
// member and never re-encodes the document: comments, key order and line
// endings are part of what the human wrote, not formatting odachi may replace.
package opencode

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

const (
	// SchemaURL is what a config created from nothing declares.
	SchemaURL = "https://opencode.ai/config.json"
	// defaultAgentKey is the only key this package writes.
	defaultAgentKey = "default_agent"
)

// Status is what EnsureDefaultAgent found, and what it did.
type Status int

const (
	// Added means the key was inserted.
	Added Status = iota
	// Already means the key already named the agent.
	Already
	// Chosen means another agent is chosen; nothing was written.
	Chosen
	// Unreadable means there is no safe edit; nothing was written.
	Unreadable
)

// Outcome says what happened and why.
type Outcome struct {
	Status Status
	// Agent is the agent already chosen, for Chosen.
	Agent string
	// Reason says why the document is unreadable, for Unreadable.
	Reason string
}

// StripComments replaces JSONC comments with spaces, keeping strings and byte
// positions intact. It is how the editor validates its own result and how a
// caller can read a document that carries comments.
func StripComments(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)

	inString, escaped, line, block := false, false, false, false
	for i := 0; i < len(out); i++ {
		c := out[i]
		switch {
		case line:
			if c == '\n' {
				line = false
			} else {
				out[i] = ' '
			}
		case block:
			if c == '*' && i+1 < len(out) && out[i+1] == '/' {
				out[i], out[i+1] = ' ', ' '
				i++
				block = false
			} else if c != '\n' && c != '\r' {
				out[i] = ' '
			}
		case inString:
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
		default:
			switch {
			case c == '"':
				inString = true
			case c == '/' && i+1 < len(out) && out[i+1] == '/':
				out[i], out[i+1] = ' ', ' '
				i++
				line = true
			case c == '/' && i+1 < len(out) && out[i+1] == '*':
				out[i], out[i+1] = ' ', ' '
				i++
				block = true
			}
		}
	}
	return out
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

	stripped := StripComments(data)
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

	root, ok := rootBrace(stripped)
	if !ok {
		return data, Outcome{Status: Unreadable, Reason: "the document is not a JSON object"}
	}
	updated := insertMember(data, root, agent, isEmptyObject(stripped, root))
	if updated == nil {
		return data, Outcome{Status: Unreadable, Reason: "the document cannot be edited safely"}
	}
	if !validObject(StripComments(updated)) {
		return data, Outcome{Status: Unreadable, Reason: "the edit would not leave valid config"}
	}
	return updated, Outcome{Status: Added}
}

// CurrentAgent answers the default agent the document names. found is false
// when the key is absent; a document with no safe reading is an error. It is
// the read-only half doctor uses: a question never writes.
func CurrentAgent(data []byte) (agent string, found bool, err error) {
	stripped := StripComments(data)
	if len(bytes.TrimSpace(stripped)) == 0 {
		return "", false, nil
	}
	if _, ok := rootBrace(stripped); !ok || !validObject(stripped) {
		return "", false, errors.New("the document is not a JSON object")
	}

	count, current, isString, err := scanDefaultAgent(stripped)
	if err != nil {
		return "", false, err
	}
	switch {
	case count > 1:
		return "", false, errors.New("default_agent appears more than once")
	case count == 1 && !isString:
		return "", false, errors.New("default_agent is not a string")
	case count == 1:
		return current, true, nil
	}
	return "", false, nil
}

// freshDoc is the config of a project that has none: the schema OpenCode's own
// docs publish, and the agent this install makes the default.
func freshDoc(agent string) []byte {
	quoted, _ := json.Marshal(agent)
	return []byte("{\n  \"$schema\": \"" + SchemaURL + "\",\n  \"" + defaultAgentKey + "\": " + string(quoted) + "\n}\n")
}

// rootBrace is the first non-space byte, which has to be the object's `{`.
func rootBrace(stripped []byte) (int, bool) {
	for i := 0; i < len(stripped); i++ {
		switch stripped[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return i, true
		default:
			return 0, false
		}
	}
	return 0, false
}

// validObject is the strict check: comments already removed, the rest has to
// decode as one object. A trailing comma is not repaired; the human's file is
// refused instead of being reformatted.
func validObject(stripped []byte) bool {
	trimmed := bytes.TrimSpace(stripped)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	var doc map[string]json.RawMessage
	return json.Unmarshal(trimmed, &doc) == nil
}

// isEmptyObject says whether the root object has members, so the inserted key
// gets a comma only when it needs one.
func isEmptyObject(stripped []byte, root int) bool {
	depth := 0
	inString, escaped := false, false
	for i := root; i < len(stripped); i++ {
		c := stripped[i]
		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				return len(bytes.TrimSpace(stripped[root+1:i])) == 0
			}
		}
	}
	return false
}

// scanDefaultAgent counts the top-level default_agent keys and reads the value
// of the only one. A duplicate key is ambiguous even though encoding/json would
// silently keep the last: the human's intent cannot be guessed.
func scanDefaultAgent(stripped []byte) (count int, value string, isString bool, err error) {
	depth := 0
	for i := 0; i < len(stripped); {
		c := stripped[i]
		if c == '"' {
			start := i + 1
			i++
			escaped := false
			for i < len(stripped) {
				if escaped {
					escaped = false
				} else if stripped[i] == '\\' {
					escaped = true
				} else if stripped[i] == '"' {
					break
				}
				i++
			}
			if i >= len(stripped) {
				return 0, "", false, errors.New("unterminated string")
			}
			text := string(stripped[start:i])
			i++ // past the closing quote

			if depth == 1 {
				if j := skipSpace(stripped, i); j < len(stripped) && stripped[j] == ':' {
					if text == defaultAgentKey {
						count++
						k := skipSpace(stripped, j+1)
						if k >= len(stripped) || stripped[k] != '"' {
							isString = false
							i = k
							continue
						}
						vstart := k + 1
						k++
						escaped := false
						for k < len(stripped) {
							if escaped {
								escaped = false
							} else if stripped[k] == '\\' {
								escaped = true
							} else if stripped[k] == '"' {
								break
							}
							k++
						}
						if k >= len(stripped) {
							return 0, "", false, errors.New("unterminated default_agent value")
						}
						value = string(stripped[vstart:k])
						isString = true
						i = k + 1
						continue
					}
				}
			}
			continue
		}

		switch c {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
		i++
	}
	return count, value, isString, nil
}

// skipSpace walks whitespace; comments are already spaces by the time this runs.
func skipSpace(data []byte, i int) int {
	for i < len(data) {
		switch data[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

// insertMember puts one member right after the root brace. The original bytes
// stay byte-for-byte; only this line is new, with the file's own line ending.
func insertMember(data []byte, root int, agent string, empty bool) []byte {
	nl := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		nl = "\r\n"
	}
	quoted, _ := json.Marshal(agent)
	line := "  \"" + defaultAgentKey + "\": " + string(quoted)
	insert := nl + line
	if !empty {
		insert += ","
	} else {
		insert += nl
	}

	out := make([]byte, 0, len(data)+len(insert))
	out = append(out, data[:root+1]...)
	out = append(out, []byte(insert)...)
	out = append(out, data[root+1:]...)
	return out
}
