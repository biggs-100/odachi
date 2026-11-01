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
	"path/filepath"
	"sort"
	"strings"

	domain "github.com/biggs-100/odachi/internal/domain/design"
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
	if len(designs) == 0 {
		result.Message = "no design in " + domain.Dir + "/: a design starts when an agent writes design/<slug>/"
		result.Lines = []string{result.Message}
		return result
	}

	for _, d := range designs {
		line := fmt.Sprintf("%s · %s · %d artboard%s", d.Slug, d.Manifest.Title, len(d.Manifest.Artboards), plural(len(d.Manifest.Artboards), "", "s"))
		if open := len(d.Comments.Open()); open > 0 {
			line += fmt.Sprintf(" · %d open comment%s", open, plural(open, "", "s"))
		}
		result.Lines = append(result.Lines, line)
		for _, artboard := range d.Manifest.Artboards {
			result.Lines = append(result.Lines, "  "+artboard.File+" · "+artboard.Title+" · "+artboard.Tradeoff)
		}
	}
	result.Message = fmt.Sprintf("%d design%s", len(designs), plural(len(designs), "", "s"))
	return result
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

	grown, missing := d.Comments.Resolve(opts.Resolve, opts.At)
	path := filepath.Join(opts.Dir, filepath.FromSlash(domain.Dir), slug, domain.CommentsFile)
	if _, err := files.Backup(path); err != nil {
		return blocked("could not back up " + path + " before rewriting it: " + oneLine(err.Error()))
	}
	if err := files.Write(path, domain.RenderComments(grown)); err != nil {
		return blocked("could not write " + path + ": " + oneLine(err.Error()))
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
		designs = append(designs, d)
	}
	return designs, nil
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
