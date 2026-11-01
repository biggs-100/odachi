package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	designserver "github.com/biggs-100/odachi/internal/adapters/designserver"
	appdesign "github.com/biggs-100/odachi/internal/app/design"
	"github.com/biggs-100/odachi/internal/app/start"
	assetsdesign "github.com/biggs-100/odachi/internal/assets/design"
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

	// The canvas takes its own flags -a port and a browser- so it is dispatched
	// before the flags every other action shares.
	if action == appdesign.ActionOpen || action == appdesign.ActionStop {
		return runCanvas(action, dir, rest, stdout, stderr)
	}
	if action == appdesign.ActionSync {
		return runDesignSync(dir, stdout, stderr)
	}

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
	case appdesign.ActionHandoff:
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "odachi: say which artboard: odachi design handoff <slug>/<artboard>")
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

// runCanvas starts or stops the local surface a person clicks on: the artboards at
// their size, and the comments written into the folder every agent reads. It is a
// command and not a tool, so every client that can run a shell can open it.
func runCanvas(action, dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi design "+action, flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var port int
	var noBrowser bool
	flags.IntVar(&port, "port", 0, "loopback port; 0 lets the kernel choose")
	flags.BoolVar(&noBrowser, "no-browser", false, "print the URL and do not open a browser")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	path := appdesign.CanvasPath(dir)

	if action == appdesign.ActionStop {
		canvas, ok := appdesign.ReadCanvas(path)
		if !ok {
			fmt.Fprintln(stdout, "no canvas is running for this clone")
			return start.CodeOK
		}
		if err := appdesign.Stop(nil, path, canvas); err != nil {
			// A canvas that died with its terminal leaves an address behind. The
			// command forgets it and says so: a dead server must not make `stop`
			// fail forever.
			if !appdesign.Running(nil, canvas) {
				_ = appdesign.ForgetCanvas(path)
				fmt.Fprintln(stdout, "no canvas is running at "+canvas.Addr+": the address is forgotten")
				return start.CodeOK
			}
			fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, "canvas stopped")
		return start.CodeOK
	}

	if flags.NArg() > 1 {
		fmt.Fprintln(stderr, "odachi: odachi design open takes at most one slug")
		return start.CodeBlocked
	}
	slug := ""
	if flags.NArg() == 1 {
		slug = flags.Arg(0)
	}

	// A canvas already up is refreshed, never doubled: the address in the store is
	// the page a person has open, and a second server would be a second truth.
	if running, ok := appdesign.ReadCanvas(path); ok && appdesign.Running(nil, running) {
		fmt.Fprintln(stdout, "canvas already at "+running.URL)
		fmt.Fprintln(stdout, "odachi design stop closes it")
		if !noBrowser {
			openBrowser(running.URL)
		}
		return start.CodeOK
	}

	server := designserver.New(designserver.Options{Dir: dir, Slug: slug, Canvas: assetsdesign.Canvas()})
	addr, url, err := server.Listen(port)
	if err != nil {
		fmt.Fprintln(stderr, "odachi: could not open the canvas: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	if err := appdesign.WriteCanvas(path, appdesign.Canvas{
		Addr:      addr,
		URL:       url,
		Slug:      slug,
		PID:       os.Getpid(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		_ = server.Close()
		fmt.Fprintln(stderr, "odachi: the canvas started but its address could not be recorded in "+path+": "+oneLine(err.Error()))
		fmt.Fprintln(stderr, "odachi: nothing is serving; fix the store and run odachi design open again")
		return start.CodeBlocked
	}

	fmt.Fprintln(stdout, "canvas at "+url)
	fmt.Fprintln(stdout, "odachi design stop closes it; Ctrl-C quits here")
	if !noBrowser {
		openBrowser(url)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	defer signal.Stop(stop)
	select {
	case <-server.Done():
	case <-stop:
	}

	_ = server.Close()
	_ = appdesign.ForgetCanvas(path)
	fmt.Fprintln(stdout, "canvas stopped")
	return start.CodeOK
}

// openBrowser is best effort: a machine with no browser is a machine where the URL
// is the answer, and failing to open one must not fail the canvas.
func openBrowser(url string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	_ = command.Start()
}

// runDesignSync reads the project's stylesheets and writes the system the skill
// follows. It reads every .css under the project -skipping the tool's own folders-
// and never writes to one: a sync that edits its sources is a sync nobody trusts.
func runDesignSync(dir string, stdout, stderr io.Writer) int {
	sources := []appdesign.Source{}

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "dist", "vendor", ".next", "target":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".css") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			rel = path
		}
		sources = append(sources, appdesign.Source{Path: filepath.ToSlash(rel), Content: string(raw)})
		return nil
	})
	if err != nil {
		fmt.Fprintln(stderr, "odachi: could not read the project's stylesheets: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	return reportDesign(appdesign.Sync(sources, osFiles{}, dir), false, stdout, stderr)
}
