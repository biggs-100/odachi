// Package design implements `odachi design`: the command half of the surface. It
// lists the designs a project keeps, shows the comments a person left, and resolves
// the ones an agent applied.
//
// Every verb here exists so that a client with no tool of its own can drive the same
// surface from a shell: the folder is the protocol, and the command only reads and
// writes it. The canvas -the page a human clicks- is another caller of the same
// files, never a second source of truth.
package design

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	domain "github.com/biggs-100/odachi/internal/domain/design"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// Exit codes follow the rest of the verbs.
const (
	CodeOK      = 0
	CodeBlocked = 1
)

// Actions the command accepts.
const (
	ActionList     = "list"
	ActionComments = "comments"
	ActionOpen     = "open"
	ActionStop     = "stop"
	ActionHandoff  = "handoff"
	ActionSync     = "sync"
	// ActionServe is the detached half of open: the child process that owns the
	// canvas after the command returns. It is internal and never printed in help.
	ActionServe = "serve"
)

// Options is one `odachi design` invocation.
type Options struct {
	Dir string
	// Action is list or comments.
	Action string
	// Slug narrows a report to one design; empty means all of them.
	Slug string
	// Resolve carries the comment ids to mark resolved, with the comments action.
	Resolve []string
	// Status is the handoff state the human or the agent records: sent,
	// in_progress or implemented. Empty means sent.
	Status string
	// Preview is the running app's loopback URL, for Compare.
	Preview string
	// Note is what the agent says about the handoff, optional.
	Note string
	// Sources are the project's design sources, as the command found them: the
	// report names the ones newer than the system it was captured in.
	Sources []Source
	// SystemExists says design/system.md is there; SystemModified is when it was
	// last written, so a source can be compared with it.
	SystemExists   bool
	SystemModified time.Time
	// At is the ISO time a resolution records. The caller owns the clock.
	At string
}

// Files is the disk port: the command reads the folders and rewrites one file.
type Files interface {
	List(dir string) ([]string, error)
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Backup(path string) (string, error)
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code    int
	Message string
	Lines   []string
	Files   []string
	Designs []Design
}

// Design is one folder the command reports.
type Design struct {
	Slug         string
	Manifest     domain.Manifest
	Comments     domain.Comments
	CommentsRead string
	Handoffs     domain.Handoffs
}

// Run performs one `odachi design`.
func Run(opts Options, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}

	switch opts.Action {
	case ActionList:
		return list(opts, files)
	case ActionComments:
		return comments(opts, files)
	case ActionHandoff:
		return handoff(opts, files)
	default:
		return blocked(`unknown action: say odachi design list or odachi design comments list|resolve <id...>`)
	}
}

// list reports every design folder with its artboards and its open comments.
func list(opts Options, files Files) Result {
	designs, problem := readDesigns(opts, files)
	if problem != nil {
		return *problem
	}

	result := Result{Code: CodeOK, Designs: designs}
	system := systemReport(opts)
	if len(designs) == 0 {
		result.Message = "no design in " + domain.Dir + "/: a design starts when an agent writes design/<slug>/"
		result.Lines = append([]string{result.Message}, system...)
		return result
	}

	for _, d := range designs {
		line := fmt.Sprintf("%s · %s · %d artboard%s", d.Slug, d.Manifest.Title, len(d.Manifest.Artboards), plural(len(d.Manifest.Artboards), "", "s"))
		if open := len(d.Comments.Open()); open > 0 {
			line += fmt.Sprintf(" · %d open comment%s", open, plural(open, "", "s"))
		}
		result.Lines = append(result.Lines, line)
		for _, artboard := range d.Manifest.Artboards {
			board := "  " + artboard.File + " · " + artboard.Title + " · " + artboard.Tradeoff
			if strings.TrimSpace(artboard.Screenshot) == "" {
				board += " · no screenshot"
			}
			result.Lines = append(result.Lines, board)
		}
	}
	result.Lines = append(result.Lines, system...)
	result.Message = fmt.Sprintf("%d design%s", len(designs), plural(len(designs), "", "s"))
	return result
}

// systemReport says whether the tokens a design follows are captured and fresh. A
// stale system is a design that will follow values the project no longer has, so the
// report names the sources that moved instead of staying quiet.
func systemReport(opts Options) []string {
	if len(opts.Sources) == 0 {
		return []string{"design sources: none detected"}
	}
	lines := []string{fmt.Sprintf("design sources: %d", len(opts.Sources))}
	if !opts.SystemExists {
		return append(lines, "design system: no design/system.md yet; run odachi design sync before designing")
	}
	stale := []string{}
	for _, source := range opts.Sources {
		if source.Modified.After(opts.SystemModified) {
			stale = append(stale, source.Path)
		}
	}
	if len(stale) == 0 {
		return append(lines, "design system: design/system.md is up to date")
	}
	sort.Strings(stale)
	return append(lines, fmt.Sprintf("design system: design/system.md is older than %d source(s): %s; say so and offer a sync before designing", len(stale), strings.Join(stale, ", ")))
}

// comments reports the open comments of a design, or resolves the ids it was given.
func comments(opts Options, files Files) Result {
	slug := strings.TrimSpace(opts.Slug)
	if slug == "" {
		return blocked("say which design: odachi design comments <slug> [--resolve <id...>]")
	}

	designs, problem := readDesigns(Options{Dir: opts.Dir, Slug: slug}, files)
	if problem != nil {
		return *problem
	}
	if len(designs) == 0 {
		return blocked("there is no design " + slug + " in " + domain.Dir + "/")
	}
	d := designs[0]

	if len(opts.Resolve) == 0 {
		result := Result{Code: CodeOK, Designs: designs}
		open := d.Comments.Open()
		if len(open) == 0 {
			result.Message = "no open comment in " + slug
			result.Lines = []string{result.Message}
			return result
		}
		for _, comment := range open {
			result.Lines = append(result.Lines, fmt.Sprintf("%s · %s · %s · %q", comment.ID, comment.File, comment.Selector, comment.Comment))
		}
		result.Message = fmt.Sprintf("%d open comment%s in %s", len(open), plural(len(open), "", "s"), slug)
		return result
	}

	// Resolving is a read-modify-write over a file another writer may touch: the
	// canvas appends a comment while the command resolves one. Reading again before
	// writing, and redoing the work when the bytes moved, is what keeps a click from
	// being swallowed by a shell command that read the same file a moment earlier.
	path := filepath.Join(opts.Dir, filepath.FromSlash(domain.Dir), slug, domain.CommentsFile)
	var grown domain.Comments
	var missing []string
	written := false
	for attempt := 0; attempt < 3 && !written; attempt++ {
		before, err := files.Read(path)
		if err != nil {
			return blocked("could not read " + path + ": " + oneLine(err.Error()))
		}
		current, err := domain.ParseComments(before)
		if err != nil {
			return blocked(slug + ": " + oneLine(err.Error()))
		}

		grown, missing = current.Resolve(opts.Resolve, opts.At)
		if len(missing) == len(opts.Resolve) {
			return blocked("no comment carries " + strings.Join(missing, ", ") + ": nothing was written")
		}

		now, err := files.Read(path)
		if err != nil {
			return blocked("could not read " + path + ": " + oneLine(err.Error()))
		}
		if string(now) != string(before) {
			continue
		}

		if _, err := files.Backup(path); err != nil {
			return blocked("could not back up " + path + " before rewriting it: " + oneLine(err.Error()))
		}
		if err := files.Write(path, domain.RenderComments(grown)); err != nil {
			return blocked("could not write " + path + ": " + oneLine(err.Error()))
		}
		written = true
	}
	if !written {
		return blocked("the comments file kept changing while it was being written: nobody else is writing, run it again")
	}

	result := Result{
		Code:    CodeOK,
		Message: fmt.Sprintf("%d comment%s resolved in %s", len(opts.Resolve)-len(missing), plural(len(opts.Resolve)-len(missing), "", "s"), slug),
		Files:   []string{filepath.ToSlash(path)},
	}
	result.Lines = []string{result.Message}
	if len(missing) > 0 {
		result.Lines = append(result.Lines, "not found: "+strings.Join(missing, ", ")+": no comment carries that id")
	}
	return result
}

// readDesigns reads the folders under design/, one per slug, in name order. A
// folder without a manifest is not a design yet and is named instead of dropped:
// an agent that writes the artboards first hears what is missing.
func readDesigns(opts Options, files Files) ([]Design, *Result) {
	root := filepath.Join(opts.Dir, filepath.FromSlash(domain.Dir))
	names, err := files.List(root)
	if err != nil {
		return nil, problem("could not read " + domain.Dir + "/: " + oneLine(err.Error()))
	}

	wanted := strings.TrimSpace(opts.Slug)
	designs := []Design{}
	sort.Strings(names)
	for _, name := range names {
		if !isSlug(name) || (wanted != "" && name != wanted) {
			continue
		}
		dir := filepath.Join(root, name)
		exists, err := files.Exists(filepath.Join(dir, domain.ManifestFile))
		if err != nil {
			return nil, problem("could not read " + name + ": " + oneLine(err.Error()))
		}
		if !exists {
			continue
		}

		raw, err := files.Read(filepath.Join(dir, domain.ManifestFile))
		if err != nil {
			return nil, problem("could not read " + name + "/" + domain.ManifestFile + ": " + oneLine(err.Error()))
		}
		manifest, err := domain.ParseManifest(raw)
		if err != nil {
			return nil, problem(name + ": " + oneLine(err.Error()))
		}

		d := Design{Slug: name, Manifest: manifest, Comments: domain.Comments{Version: domain.Version}}
		commentsPath := filepath.Join(dir, domain.CommentsFile)
		hasComments, err := files.Exists(commentsPath)
		if err != nil {
			return nil, problem("could not read " + name + "/" + domain.CommentsFile + ": " + oneLine(err.Error()))
		}
		if hasComments {
			raw, err := files.Read(commentsPath)
			if err != nil {
				return nil, problem("could not read " + name + "/" + domain.CommentsFile + ": " + oneLine(err.Error()))
			}
			if len(strings.TrimSpace(string(raw))) > 0 {
				parsed, err := domain.ParseComments(raw)
				if err != nil {
					return nil, problem(name + ": " + oneLine(err.Error()))
				}
				d.Comments = parsed
			}
		}

		d.Handoffs = domain.Handoffs{Version: domain.Version, Handoffs: map[string]domain.Handoff{}}
		handoffsPath := filepath.Join(dir, domain.HandoffsFile)
		hasHandoffs, err := files.Exists(handoffsPath)
		if err != nil {
			return nil, problem("could not read " + name + "/" + domain.HandoffsFile + ": " + oneLine(err.Error()))
		}
		if hasHandoffs {
			raw, err := files.Read(handoffsPath)
			if err != nil {
				return nil, problem("could not read " + name + "/" + domain.HandoffsFile + ": " + oneLine(err.Error()))
			}
			if len(strings.TrimSpace(string(raw))) > 0 {
				parsed, err := domain.ParseHandoffs(raw)
				if err != nil {
					return nil, problem(name + ": " + oneLine(err.Error()))
				}
				d.Handoffs = parsed
			}
		}
		designs = append(designs, d)
	}
	return designs, nil
}

// validFeatureDoc is the one shape a design's feature link takes: a relative path
// under odd/tasks/ ending in .md, with no traversal. The manifest may name a path;
// the handoff is the one that writes to it, so it is the one that checks it.
func validFeatureDoc(feature string) bool {
	if !strings.HasPrefix(feature, "odd/tasks/") || !strings.HasSuffix(feature, ".md") {
		return false
	}
	return path.Clean(feature) == feature && !strings.Contains(feature, "..")
}

// isSlug reports whether a folder name may be a design: the manifest names files,
// not paths, and a design is a folder like 02-pricing-page.
func isSlug(name string) bool {
	return name != "" && name == filepath.Base(name) && !strings.HasPrefix(name, ".")
}

func problem(message string) *Result {
	result := blocked(message)
	return &result
}

func blocked(message string) Result {
	return Result{Code: CodeBlocked, Message: oneLine(message), Lines: []string{oneLine(message)}}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}

// handoff turns a chosen artboard into the instruction to implement it: the file,
// what it traded, and the alternatives that were not chosen. It writes nothing and
// starts nothing -the implementation is the agent's step, and the artboard is a
// direction, not a proof.
func handoff(opts Options, files Files) Result {
	target := strings.TrimSpace(opts.Slug)
	slug, file, ok := strings.Cut(target, "/")
	if !ok || slug == "" || file == "" {
		return blocked("say which artboard: odachi design handoff <slug>/<artboard>")
	}

	designs, problem := readDesigns(Options{Dir: opts.Dir, Slug: slug}, files)
	if problem != nil {
		return *problem
	}
	if len(designs) == 0 {
		return blocked("there is no design " + slug + " in " + domain.Dir + "/")
	}
	d := designs[0]

	chosen := -1
	for i, artboard := range d.Manifest.Artboards {
		if artboard.File == file {
			chosen = i
			break
		}
	}
	if chosen < 0 {
		names := make([]string, 0, len(d.Manifest.Artboards))
		for _, artboard := range d.Manifest.Artboards {
			names = append(names, artboard.File)
		}
		return blocked(file + " is not an artboard of " + slug + ": the artboards are " + strings.Join(names, ", "))
	}

	feature := strings.TrimSpace(d.Manifest.Feature)
	featurePath := ""
	if feature != "" {
		if !validFeatureDoc(feature) {
			return blocked("the design links " + feature + ": a feature is a task document under odd/tasks/, a relative .md path")
		}
		featurePath = filepath.Join(opts.Dir, filepath.FromSlash(feature))
		if _, err := files.Read(featurePath); err != nil {
			return blocked("the design links " + feature + ", and that document cannot be read: " + oneLine(err.Error()))
		}
	}

	alternatives := []string{}
	for i, artboard := range d.Manifest.Artboards {
		if i == chosen {
			continue
		}
		alternatives = append(alternatives, artboard.File+" · "+artboard.Title)
	}

	path := domain.Dir + "/" + slug + "/" + file
	lines := []string{
		path + " · " + d.Manifest.Artboards[chosen].Title + " · " + d.Manifest.Artboards[chosen].Tradeoff,
		"brief: " + d.Manifest.Brief,
	}
	if d.Manifest.Feature != "" {
		lines = append(lines, "feature: "+d.Manifest.Feature)
	}
	lines = append(lines, "implement this artboard in the project's real stack.")
	if len(alternatives) > 0 {
		lines = append(lines, "the alternatives it was chosen over: "+strings.Join(alternatives, "; "))
	}
	lines = append(lines, "the artboard is a direction; the code is the proof.")

	status := strings.TrimSpace(opts.Status)
	if status == "" {
		status = domain.HandoffSent
	}
	switch status {
	case domain.HandoffSent, domain.HandoffInProgress, domain.HandoffImplemented:
	default:
		return blocked("a handoff is " + domain.HandoffSent + ", " + domain.HandoffInProgress + " or " + domain.HandoffImplemented + ": " + status + " is none of them")
	}

	preview := ""
	if strings.TrimSpace(opts.Preview) != "" {
		normalized, err := domain.LoopbackURL(opts.Preview)
		if err != nil {
			return blocked(err.Error())
		}
		preview = normalized
	}

	entry := d.Handoffs.Handoffs[file]
	if entry.SentAt == "" {
		entry.SentAt = opts.At
	}
	entry.Status = status
	entry.UpdatedAt = opts.At
	entry.Feature = d.Manifest.Feature
	entry.Note = strings.TrimSpace(opts.Note)
	if preview != "" {
		entry.PreviewURL = preview
	}
	d.Handoffs.Version = domain.Version
	d.Handoffs.Chosen = file
	d.Handoffs.Handoffs[file] = entry

	statePath := filepath.Join(opts.Dir, filepath.FromSlash(domain.Dir), slug, domain.HandoffsFile)
	if exists, err := files.Exists(statePath); err != nil {
		return blocked("could not read " + statePath + ": " + oneLine(err.Error()))
	} else if exists {
		if _, err := files.Backup(statePath); err != nil {
			return blocked("could not back up " + statePath + " before rewriting it: " + oneLine(err.Error()))
		}
	}
	if err := files.Write(statePath, domain.RenderHandoffs(d.Handoffs)); err != nil {
		return blocked("could not write " + statePath + ": " + oneLine(err.Error()))
	}

	lines = append(lines, "state: "+status+" (chosen "+file+")")
	if preview != "" {
		lines = append(lines, "preview: "+preview+" (the canvas compares it)")
	}
	if feature != "" {
		raw, err := files.Read(featurePath)
		if err != nil {
			return blocked("the handoff is recorded, and the feature document cannot be read: " + oneLine(err.Error()))
		}
		altFiles := []string{}
		for i, artboard := range d.Manifest.Artboards {
			if i != chosen {
				altFiles = append(altFiles, artboard.File)
			}
		}
		text := "design/" + slug + "/" + file + " chosen for this feature"
		if len(altFiles) > 0 {
			text += "; the alternatives are " + strings.Join(altFiles, ", ")
		} else {
			text += "; it is the only artboard"
		}
		doc := taskdoc.Parse(string(raw))
		doc = doc.AppendHistory(fmt.Sprintf("L%d · %q", doc.NextLine(), text))
		if _, err := files.Backup(featurePath); err != nil {
			return blocked("the handoff is recorded, and the feature document could not be backed up: " + oneLine(err.Error()))
		}
		if err := files.Write(featurePath, []byte(doc.Render())); err != nil {
			return blocked("the handoff is recorded, and the feature document could not be written: " + oneLine(err.Error()))
		}
		lines = append(lines, "logged in "+feature)
	}

	return Result{Code: CodeOK, Message: "handoff " + path, Lines: lines, Files: []string{filepath.ToSlash(statePath)}, Designs: designs}
}
