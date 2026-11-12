package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	delegateapp "github.com/biggs-100/odachi/internal/app/delegate"
	"github.com/biggs-100/odachi/internal/app/start"
)

// craftResolver answers the delegation's two questions with the same order every
// client reads: the project's copy first, then the one the machine keeps.
type craftResolver struct{ dir string }

func (r craftResolver) Path(name string) (string, bool) { return craftPath(r.dir, name) }
func (r craftResolver) Names() []string                 { return craftNames(r.dir) }

// runDelegate composes the prompt a parent hands a delegated worker. The task
// doc and the step are the input; the skills are the ones the plan recorded, so
// the parent cannot drop the section by writing a shorter paragraph.
func runDelegate(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi delegate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		step     string
		surfaces stringList
		commands stringList
	)

	flags.StringVar(&step, "step", "", "T<n> of the delegated step to hand over")
	flags.Var(&surfaces, "surface", "path the worker may edit (repeatable)")
	flags.Var(&commands, "command", "verification command the worker runs (repeatable)")

	// The documented shape starts with the task doc, so a leading argument is
	// the document and the flags may come after it.
	taskDoc := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		taskDoc = args[0]
		args = args[1:]
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		_, _ = fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi delegate -h to see the flags")
		return start.CodeBlocked
	}

	rest := flags.Args()
	if taskDoc == "" && len(rest) == 1 {
		taskDoc = rest[0]
		rest = nil
	}
	switch {
	case taskDoc == "":
		_, _ = fmt.Fprintln(stderr, `odachi: missing the task doc: write odachi delegate odd/tasks/<file>.md --step T1`)
		return start.CodeBlocked
	case len(rest) > 0:
		_, _ = fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi delegate takes one task doc\n", rest[0])
		return start.CodeBlocked
	}

	result := delegateapp.Run(delegateapp.Options{
		Dir:      dir,
		TaskDoc:  taskDoc,
		Step:     step,
		Surfaces: surfaces,
		Commands: commands,
		Resolver: craftResolver{dir: dir},
	}, disk(dir))

	if result.Code != delegateapp.CodeOK {
		_, _ = fmt.Fprintln(stderr, "odachi: "+result.Message)
		return result.Code
	}
	_, _ = fmt.Fprint(stdout, result.Prompt)
	return start.CodeOK
}
