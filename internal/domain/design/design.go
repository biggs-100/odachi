// Package design owns the format of a design surface: the folder a person reads,
// the manifest that lists its artboards and the comments a human leaves on them.
//
// It is pure: no filesystem, no network, no clock. The caller finds the folders,
// reads and writes the files, and the package decides what a valid design is. The
// format is the contract every client shares -the skill writes it and an agent
// reads it with a text tool when it has no command- so it is versioned and pinned
// by a fixture, and it never depends on the model that produced it.
package design

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
)

// Version is the format this package reads and writes. A file that names another
// version is refused by name: a format that silently accepts the future is a
// format nobody can migrate.
const Version = 1

// Dir is the folder a project keeps its designs in, next to brief.md.
const Dir = "design"

// ManifestFile and CommentsFile are the two files a design folder carries.
const (
	ManifestFile = "design.json"
	CommentsFile = "comments.json"
)

// Artboard is one direction: a self-contained HTML file with the tradeoff it makes.
type Artboard struct {
	File     string `json:"file"`
	Title    string `json:"title"`
	Tradeoff string `json:"tradeoff"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// Manifest is design/<slug>/design.json: what the design is and which artboards it
// has. Feature is the task document it came from, empty for work nobody tracked.
type Manifest struct {
	Title     string     `json:"title"`
	Brief     string     `json:"brief"`
	Feature   string     `json:"feature,omitempty"`
	Artboards []Artboard `json:"artboards"`
}

// Comment is one note a human left on an element of an artboard.
type Comment struct {
	ID         string `json:"id"`
	File       string `json:"file"`
	Selector   string `json:"selector"`
	Excerpt    string `json:"excerpt"`
	Comment    string `json:"comment"`
	Screenshot string `json:"screenshot,omitempty"`
	CreatedAt  string `json:"createdAt"`
	Status     string `json:"status"`
	SentAt     *string `json:"sentAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

// Comments is design/<slug>/comments.json.
type Comments struct {
	Version  int       `json:"version"`
	Comments []Comment `json:"comments"`
}

// The two states a comment can be in. A comment the canvas dropped because its
// element is gone is still a comment: the file says so, never the absence of it.
const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

// ParseManifest reads a manifest and refuses the shapes the canvas cannot draw: a
// file outside the folder, a size that is not a viewport, a manifest with no
// artboard. An artboard without a tradeoff is refused too -three options that make
// the same tradeoff are one option painted three times.
func ParseManifest(raw []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("design: the manifest is not valid JSON: %w", err)
	}
	m.Title = strings.TrimSpace(m.Title)
	m.Brief = strings.TrimSpace(m.Brief)
	m.Feature = strings.TrimSpace(m.Feature)

	if m.Title == "" {
		return Manifest{}, fmt.Errorf("design: the manifest has no title")
	}
	if len(m.Artboards) == 0 {
		return Manifest{}, fmt.Errorf("design: the manifest lists no artboard")
	}
	seen := make(map[string]bool, len(m.Artboards))
	for i, artboard := range m.Artboards {
		file := strings.TrimSpace(artboard.File)
		switch {
		case file == "":
			return Manifest{}, fmt.Errorf("design: artboard %d has no file", i+1)
		case file != path.Base(file) || strings.Contains(file, `\`):
			return Manifest{}, fmt.Errorf("design: artboard %d names %q: a file inside the design folder has no path", i+1, file)
		case seen[file]:
			return Manifest{}, fmt.Errorf("design: artboard %q is listed twice", file)
		case artboard.Width <= 0 || artboard.Height <= 0:
			return Manifest{}, fmt.Errorf("design: artboard %q has no viewport: width and height are CSS pixels", file)
		case strings.TrimSpace(artboard.Tradeoff) == "":
			return Manifest{}, fmt.Errorf("design: artboard %q does not say the tradeoff it makes", file)
		}
		seen[file] = true
		m.Artboards[i].File = file
	}
	return m, nil
}

// RenderManifest writes the canonical manifest: two-space JSON and a trailing
// newline, so two designs with the same content are the same bytes.
func RenderManifest(m Manifest) []byte {
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		// The struct is ours and holds no value json cannot encode.
		panic("design: a manifest a struct cannot encode: " + err.Error())
	}
	return append(out, '\n')
}

// ParseComments reads comments.json, refusing a version this package does not
// know by name and a comment that cannot be applied: no file, no selector, no text.
func ParseComments(raw []byte) (Comments, error) {
	var c Comments
	if err := json.Unmarshal(raw, &c); err != nil {
		return Comments{}, fmt.Errorf("design: comments.json is not valid JSON: %w", err)
	}
	if c.Version != Version {
		return Comments{}, fmt.Errorf("design: comments.json says version %d and this binary reads version %d", c.Version, Version)
	}
	for i, comment := range c.Comments {
		if strings.TrimSpace(comment.ID) == "" {
			return Comments{}, fmt.Errorf("design: comment %d has no id", i+1)
		}
		if strings.TrimSpace(comment.File) == "" {
			return Comments{}, fmt.Errorf("design: comment %q names no artboard", comment.ID)
		}
		if strings.TrimSpace(comment.Comment) == "" {
			return Comments{}, fmt.Errorf("design: comment %q has no text", comment.ID)
		}
		if comment.Status != StatusOpen && comment.Status != StatusResolved {
			return Comments{}, fmt.Errorf("design: comment %q is %q, and the states are %s or %s", comment.ID, comment.Status, StatusOpen, StatusResolved)
		}
	}
	return c, nil
}

// RenderComments writes the canonical comments file.
func RenderComments(c Comments) []byte {
	if c.Version == 0 {
		c.Version = Version
	}
	if c.Comments == nil {
		c.Comments = []Comment{}
	}
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic("design: a comments file a struct cannot encode: " + err.Error())
	}
	return append(out, '\n')
}

// Open is the comments a person left and nobody applied yet, oldest first, and the
// list the agent reads when it has no command: the file is the protocol.
func (c Comments) Open() []Comment {
	open := make([]Comment, 0, len(c.Comments))
	for _, comment := range c.Comments {
		if comment.Status == StatusOpen {
			open = append(open, comment)
		}
	}
	return open
}

// Resolve marks the comments with those ids resolved at the given moment, and
// reports the ids it did not find. It never drops or reorders an entry: the file
// is the history of what a person asked for.
func (c Comments) Resolve(ids []string, at string) (Comments, []string) {
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			wanted[id] = true
		}
	}

	found := make(map[string]bool, len(wanted))
	out := c
	out.Comments = make([]Comment, len(c.Comments))
	copy(out.Comments, c.Comments)
	for i, comment := range out.Comments {
		if !wanted[comment.ID] {
			continue
		}
		found[comment.ID] = true
		out.Comments[i].Status = StatusResolved
		resolved := at
		out.Comments[i].ResolvedAt = &resolved
	}

	missing := make([]string, 0, len(wanted))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && !found[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	return out, missing
}
