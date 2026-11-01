package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
	"github.com/biggs-100/odachi/internal/app/start"
)

// runDesign prints the designs a project keeps and the comments a person left on
// them. It writes one file -the comments it resolves- and nothing else: the folder
// is the protocol, and every client reads it the same way.
func runDesign(dir string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		usage(stdout)
		return start.CodeOK
	}

	action := args[0]
	rest := args[1:]

	flags := flag.NewFlagSet("odachi design "+action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var asJSON bool
	var resolve string
	flags.BoolVar(&asJSON, "json", false, "machine output")
	flags.StringVar(&resolve, "resolve", "", "comment ids to mark resolved, comma-separated")

	if err := flags.Parse(rest); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi design -h to see the actions")
		return start.CodeBlocked
	}

	switch action {
	case appdesign.ActionList:
		if flags.NArg() > 1 {
			fmt.Fprintln(stderr, "odachi: odachi design list takes at most one slug")
			return start.CodeBlocked
		}
	case appdesign.ActionComments:
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "odachi: say which design: odachi design comments <slug> [--resolve id1,id2]")
			return start.CodeBlocked
		}
	default:
		fmt.Fprintf(stderr, "odachi: unknown action %q: odachi design list or odachi design comments <slug>\n", action)
		return start.CodeBlocked
	}

	slug := ""
	if flags.NArg() == 1 {
		slug = flags.Arg(0)
	}

	ids := []string{}
	for _, id := range strings.Split(resolve, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}

	result := appdesign.Run(appdesign.Options{
		Dir:     dir,
		Action:  action,
		Slug:    slug,
		Resolve: ids,
		At:      time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}, osFiles{})

	return reportDesign(result, asJSON, stdout, stderr)
}

// reportDesign prints the result: the lines a human reads, or the machine payload
// with the designs and their comments.
func reportDesign(result appdesign.Result, asJSON bool, stdout, stderr io.Writer) int {
	if !asJSON {
		return reportOutcome(outcome{
			Code:    result.Code,
			Message: result.Message,
			Lines:   result.Lines,
			Files:   result.Files,
		}, false, stdout, stderr)
	}

	payload := struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Lines   []string        `json:"lines"`
		Files   []string        `json:"files,omitempty"`
		Designs []designPayload `json:"designs,omitempty"`
	}{Code: result.Code, Message: result.Message, Lines: result.Lines, Files: result.Files}
	for _, d := range result.Designs {
		payload.Designs = append(payload.Designs, designPayload{
			Slug:      d.Slug,
			Title:     d.Manifest.Title,
			Brief:     d.Manifest.Brief,
			Feature:   d.Manifest.Feature,
			Artboards: len(d.Manifest.Artboards),
			Comments:  commentCounts(d),
		})
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, "odachi: could not encode the report: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	fmt.Fprintln(stdout, string(out))
	return result.Code
}

// designPayload is one design in the machine report: what a client needs to draw a
// list without reading the folder itself.
type designPayload struct {
	Slug      string         `json:"slug"`
	Title     string         `json:"title"`
	Brief     string         `json:"brief"`
	Feature   string         `json:"feature,omitempty"`
	Artboards int            `json:"artboards"`
	Comments  commentPayload `json:"comments"`
}

// commentPayload is how many comments are open and how many were resolved.
type commentPayload struct {
	Open     int `json:"open"`
	Resolved int `json:"resolved"`
}

func commentCounts(d appdesign.Design) commentPayload {
	counts := commentPayload{}
	for _, comment := range d.Comments.Comments {
		if comment.Status == "open" {
			counts.Open++
			continue
		}
		counts.Resolved++
	}
	return counts
}
