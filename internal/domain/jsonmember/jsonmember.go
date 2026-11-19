// Package jsonmember edits one member of a JSON document that belongs to the
// human. The edit inserts or replaces a single value and never re-encodes the
// document: comments, key order and line endings are part of what the human
// wrote, not formatting odachi may replace. The promise lives here, once, so
// every client config odachi touches keeps it instead of re-implementing it.
package jsonmember

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// identifierAlphabet is what model ids use beside letters and digits:
// provider/name#variant, with dots, colons, plus and dashes.
const identifierAlphabet = "._/#@:+-"

// Status is what Ensure found, and what it did.
type Status int

const (
	// Added means the member was inserted.
	Added Status = iota
	// Already means the key already named the value.
	Already
	// Replaced means a different value was written in place.
	Replaced
	// Unreadable means there is no safe edit; nothing was written.
	Unreadable
)

// Outcome says what happened and why.
type Outcome struct {
	Status Status
	// Previous is the value that was there before a Replaced edit, unquoted.
	Previous string
	// Reason says why the document is unreadable, for Unreadable.
	Reason string
}

// NormalizeIdentifier checks a value that ends up inside a file a client parses.
// It is the same alphabet the review domain accepts for a model, repeated rather
// than imported: this package is the editor of every client config, and it has no
// business importing one domain's rules to write another's file. Empty is refused
// because the flag grammar and this check have to agree that a model names
// something.
func NormalizeIdentifier(value string) (string, error) {
	identifier := strings.TrimSpace(value)
	if identifier == "" {
		return "", errors.New("the value is empty")
	}
	for _, r := range identifier {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case strings.ContainsRune(identifierAlphabet, r):
		default:
			return "", fmt.Errorf("%q contains %q: it has to be an identifier (provider/name#variant), because it ends up inside a file a client parses", value, r)
		}
	}
	return identifier, nil
}

// StripComments replaces JSONC comments with spaces, keeping strings and byte
// positions intact. It is how the editor validates its own result and how a
// caller can read a document that carries comments. Positions matter: a lookup
// runs on the stripped copy and the edit is applied to the original bytes.
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

// Read answers the root member named key. found is false when the member is
// absent; a document with no safe reading is an error. It is the read-only half
// a caller asks before deciding whether an edit is a write or a keeping.
func Read(data []byte, key string) (value string, found bool, err error) {
	stripped := StripComments(data)
	if len(bytes.TrimSpace(stripped)) == 0 {
		return "", false, nil
	}
	if _, ok := rootBrace(stripped); !ok || !validObject(stripped) {
		return "", false, errors.New("the document is not a JSON object")
	}

	member, err := lookup(stripped, key)
	if err != nil {
		return "", false, err
	}
	switch {
	case member.count > 1:
		return "", false, fmt.Errorf("%s appears more than once", key)
	case member.count == 1 && !member.isString:
		return "", false, fmt.Errorf("%s is not a string", key)
	case member.count == 1:
		return member.value, true, nil
	}
	return "", false, nil
}

// Ensure returns the document with key set to value: absent, it inserts the
// member; already the same, it returns the bytes untouched; a different value is
// replaced in place. Only the string's content changes on a replacement, so
// every comment, key and line ending around it survives.
func Ensure(data []byte, key, value string) ([]byte, Outcome) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	switch {
	case key == "":
		return data, Outcome{Status: Unreadable, Reason: "no member name"}
	case value == "":
		return data, Outcome{Status: Unreadable, Reason: "no value"}
	}

	stripped := StripComments(data)
	if len(bytes.TrimSpace(stripped)) == 0 {
		// A document with no bytes has no layout to keep: the member lands in the
		// smallest object, with the platform-independent newline of a new file.
		return insertMember([]byte("{}"), 0, key, value, true), Outcome{Status: Added}
	}
	if _, ok := rootBrace(stripped); !ok || !validObject(stripped) {
		return data, Outcome{Status: Unreadable, Reason: "the document is not a JSON object"}
	}

	member, err := lookup(stripped, key)
	if err != nil {
		return data, Outcome{Status: Unreadable, Reason: err.Error()}
	}
	switch {
	case member.count > 1:
		return data, Outcome{Status: Unreadable, Reason: key + " appears more than once"}
	case member.count == 1 && !member.isString:
		return data, Outcome{Status: Unreadable, Reason: key + " is not a string"}
	case member.count == 1 && member.value == value:
		return data, Outcome{Status: Already}
	case member.count == 1:
		updated := replaceValue(data, member, value)
		if !validObject(StripComments(updated)) {
			return data, Outcome{Status: Unreadable, Reason: "the edit would not leave valid config"}
		}
		return updated, Outcome{Status: Replaced, Previous: member.value}
	}

	root, _ := rootBrace(stripped)
	updated := insertMember(data, root, key, value, isEmptyObject(stripped, root))
	if !validObject(StripComments(updated)) {
		return data, Outcome{Status: Unreadable, Reason: "the edit would not leave valid config"}
	}
	return updated, Outcome{Status: Added}
}

// member is one root member found by lookup. start and end are the byte span of
// the string's content in the unstripped document, because stripping preserves
// byte positions: the replacement edits the original bytes, not a copy.
type member struct {
	count    int
	value    string
	isString bool
	start    int
	end      int
}

// lookup counts the root members named key and reads the value of the only one.
// A duplicate key is ambiguous even though encoding/json would silently keep the
// last: the human's intent cannot be guessed.
func lookup(stripped []byte, key string) (member, error) {
	var found member
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
				return member{}, errors.New("unterminated string")
			}
			text := string(stripped[start:i])
			i++ // past the closing quote

			if depth == 1 {
				if j := skipSpace(stripped, i); j < len(stripped) && stripped[j] == ':' {
					if text == key {
						found.count++
						k := skipSpace(stripped, j+1)
						if k >= len(stripped) || stripped[k] != '"' {
							found.isString = false
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
							return member{}, fmt.Errorf("unterminated %s value", key)
						}
						found.value = string(stripped[vstart:k])
						found.isString = true
						found.start, found.end = vstart, k
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
	return found, nil
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
func insertMember(data []byte, root int, key, value string, empty bool) []byte {
	nl := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		nl = "\r\n"
	}
	quotedKey, _ := json.Marshal(key)
	quotedValue, _ := json.Marshal(value)
	line := "  " + string(quotedKey) + ": " + string(quotedValue)
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

// replaceValue swaps the content of one string literal and keeps every other
// byte. The span comes from the stripped copy, whose positions match the
// original because comments became spaces instead of disappearing.
func replaceValue(data []byte, member member, value string) []byte {
	quoted, _ := json.Marshal(value)
	content := quoted[1 : len(quoted)-1]

	out := make([]byte, 0, len(data)+len(content))
	out = append(out, data[:member.start]...)
	out = append(out, content...)
	out = append(out, data[member.end:]...)
	return out
}
