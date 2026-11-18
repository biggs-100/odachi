package skilleval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ParseStream reads the event stream a client writes when it runs an agent: one
// JSON object per line, where a tool call names the tool in `part.tool` and its
// arguments in `part.state.input`. Everything else -the narration, the step
// markers, the token counts- is not an event, and the exam does not read it.
//
// The keys are the client's, not ours: a path is `filePath` in a write, `path` in
// a read, and a skill load names an `id`. The exam answers what the stream can
// show, so a shape this parser cannot recognise is a line that scores nothing
// rather than a run that fails.
func ParseStream(data []byte) ([]Event, error) {
	var events []Event
	for number, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(string(raw))
		if line == "" {
			continue
		}

		var entry struct {
			Type string `json:"type"`
			Part struct {
				Tool  string `json:"tool"`
				State struct {
					Input map[string]any `json:"input"`
				} `json:"state"`
			} `json:"part"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("line %d of the stream is not JSON: %w", number+1, err)
		}
		if entry.Type != "tool_use" || entry.Part.Tool == "" {
			continue
		}

		events = append(events, Event{
			Tool:    entry.Part.Tool,
			Path:    firstString(entry.Part.State.Input, "filePath", "path", "file"),
			Command: firstString(entry.Part.State.Input, "command"),
			Name:    firstString(entry.Part.State.Input, "id"),
		})
	}
	return events, nil
}

// firstString answers the first of the keys the input carries as a string. The
// clients spell the same argument differently across versions, and the exam reads
// the argument, not the spelling.
func firstString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := input[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}
