package design

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// HandoffsFile is the third file a design folder carries: what happened to each
// artboard after a person chose it. It is state about the work, not the artboard:
// the folder stays the protocol, and the canvas and the command read the same file.
const HandoffsFile = "handoffs.json"

// The three states a handoff can be in. `sent` means the instruction was handed to
// the agent; the other two are the agent's progress, never the human's guess.
const (
	HandoffSent        = "sent"
	HandoffInProgress  = "in_progress"
	HandoffImplemented = "implemented"
)

// Handoff is one artboard's trip from direction to code. PreviewURL is the running
// app, loopback only: the canvas compares the two, and a public URL would make the
// canvas a browser for the internet.
type Handoff struct {
	Status     string `json:"status"`
	SentAt     string `json:"sentAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
	Queued     bool   `json:"queued,omitempty"`
	Feature    string `json:"feature,omitempty"`
	Note       string `json:"note,omitempty"`
	PreviewURL string `json:"previewUrl,omitempty"`
}

// Handoffs is design/<slug>/handoffs.json. Chosen is the artboard a person picked
// over the others; an empty chosen means the design has no decision yet.
type Handoffs struct {
	Version  int                `json:"version"`
	Chosen   string             `json:"chosen,omitempty"`
	Handoffs map[string]Handoff `json:"handoffs"`
}

// ParseHandoffs reads the file a person or an agent wrote, refusing a status the
// canvas cannot draw and a preview that is not loopback. An unknown version is
// refused by name, like the other two files.
func ParseHandoffs(raw []byte) (Handoffs, error) {
	var h Handoffs
	if err := json.Unmarshal(raw, &h); err != nil {
		return Handoffs{}, fmt.Errorf("design: handoffs.json is not valid JSON: %w", err)
	}
	if h.Version != Version {
		return Handoffs{}, fmt.Errorf("design: handoffs.json says version %d and this binary reads version %d", h.Version, Version)
	}
	if h.Handoffs == nil {
		h.Handoffs = map[string]Handoff{}
	}
	if h.Chosen != "" && !validFile(h.Chosen) {
		return Handoffs{}, fmt.Errorf("design: handoffs.json chooses %q: an artboard is a file inside the design folder", h.Chosen)
	}
	for file, entry := range h.Handoffs {
		if !validFile(file) {
			return Handoffs{}, fmt.Errorf("design: handoffs.json names %q: an artboard is a file inside the design folder", file)
		}
		switch entry.Status {
		case HandoffSent, HandoffInProgress, HandoffImplemented:
		default:
			return Handoffs{}, fmt.Errorf("design: handoff %q is %q, and the states are %s, %s or %s", file, entry.Status, HandoffSent, HandoffInProgress, HandoffImplemented)
		}
		if entry.PreviewURL != "" {
			preview, err := LoopbackURL(entry.PreviewURL)
			if err != nil {
				return Handoffs{}, fmt.Errorf("design: handoff %q: %w", file, err)
			}
			h.Handoffs[file] = withPreview(entry, preview)
		}
	}
	return h, nil
}

// withPreview replaces one entry's preview after normalization; a map value is not
// addressable, and a second copy of the URL would be a second truth.
func withPreview(entry Handoff, preview string) Handoff {
	entry.PreviewURL = preview
	return entry
}

// RenderHandoffs writes the canonical file: two-space JSON, a trailing newline, the
// map ordered by key so two equal states are equal bytes.
func RenderHandoffs(h Handoffs) []byte {
	if h.Version == 0 {
		h.Version = Version
	}
	if h.Handoffs == nil {
		h.Handoffs = map[string]Handoff{}
	}
	out, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		panic("design: a handoffs file a struct cannot encode: " + err.Error())
	}
	return append(out, '\n')
}

// LoopbackURL accepts the running app's address: http or https on localhost,
// 127.0.0.1 or ::1, any port. It returns the normalized URL and names why anything
// else is refused, because the canvas will frame it.
func LoopbackURL(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", fmt.Errorf("the preview URL is empty")
	}
	parsed, err := url.Parse(text)
	if err != nil {
		return "", fmt.Errorf("the preview URL is not a URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("the preview URL is %q: it is http or https", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "localhost", "127.0.0.1", "::1":
	default:
		return "", fmt.Errorf("the preview URL names %q: the app must run on localhost, 127.0.0.1 or [::1]", host)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		// A preview is an address a canvas frames: credentials, a query or a fragment
		// are not part of "the app is here", and a canvas that carries them lies.
		return "", fmt.Errorf("the preview URL carries a user, a query or a fragment: it is a loopback address, nothing else")
	}
	return parsed.String(), nil
}

// validFile is the one rule the manifest and this file share: an artboard is a bare
// file name inside the design folder.
func validFile(name string) bool {
	return name != "" && name == strings.TrimSpace(name) && !strings.ContainsAny(name, `/\`) && name != "." && name != ".."
}
