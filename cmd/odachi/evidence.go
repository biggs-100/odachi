package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/biggs-100/odachi/internal/app/evidence"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// runEvidence parses the two words of `odachi evidence` and dispatches the one
// action that exists. Flags may follow the tree, as a human types them.
func runEvidence(dir string, args []string, stdout, stderr io.Writer) int {
	const usageLine = "usage: odachi evidence export <tree> --out <path>"

	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "odachi: evidence needs an action: "+usageLine)
		return start.CodeBlocked
	}
	if args[0] != "export" {
		_, _ = fmt.Fprintf(stderr, "odachi: evidence %s is not an action; %s\n", args[0], usageLine)
		return start.CodeBlocked
	}

	tree := ""
	out := ""
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h", arg == "--help":
			_, _ = fmt.Fprintln(stdout, usageLine)
			return start.CodeOK
		case arg == "--out":
			if i+1 >= len(args) {
				_, _ = fmt.Fprintln(stderr, "odachi: --out needs a path: "+usageLine)
				return start.CodeBlocked
			}
			i++
			out = args[i]
		case strings.HasPrefix(arg, "--out="):
			out = strings.TrimPrefix(arg, "--out=")
		case strings.HasPrefix(arg, "-"):
			_, _ = fmt.Fprintf(stderr, "odachi: %s is not a flag of evidence export; %s\n", arg, usageLine)
			return start.CodeBlocked
		case tree == "":
			tree = arg
		default:
			_, _ = fmt.Fprintf(stderr, "odachi: extra argument: %s; %s\n", arg, usageLine)
			return start.CodeBlocked
		}
	}
	if tree == "" {
		_, _ = fmt.Fprintln(stderr, "odachi: evidence export needs a tree: "+usageLine)
		return start.CodeBlocked
	}
	if strings.TrimSpace(out) == "" {
		_, _ = fmt.Fprintln(stderr, "odachi: evidence export needs --out <path>: the bundle must have a place to go")
		return start.CodeBlocked
	}

	root, err := store.Root(dir, gitRepo{})
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error())+"; odachi evidence export needs the clone's store, so run it inside a git repository")
		return start.CodeBlocked
	}

	result := evidence.Run(evidence.Options{Dir: dir, Store: root, Tree: tree, Out: out}, evidenceDisk{disk(dir)}, gitRepo{})
	for _, line := range result.Lines {
		_, _ = fmt.Fprintln(stdout, line)
	}
	if result.Code != evidence.CodeOK {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(result.Message))
	}
	return result.Code
}

// evidenceDisk adds the one method osFiles keeps private: MkdirAll. The app
// calls it only for the output's parent, so a bundle cannot plant a directory
// anywhere the human did not name.
type evidenceDisk struct{ osFiles }

func (evidenceDisk) MkdirAll(dir string) error {
	return os.MkdirAll(dir, 0o755)
}
