package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/shellcmd"
	reviewapp "github.com/biggs-100/odachi/internal/app/review"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/app/tasks"
	verifyapp "github.com/biggs-100/odachi/internal/app/verify"
	"github.com/biggs-100/odachi/internal/domain/change"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
)

// runVerify runs the independent verifier of the current tree and stores the
// record it answers. It composes the question from what the project already
// wrote -the requirement, the promises, the risk the commits declared- so the
// verifier reads the human's words and not a summary of them.
func runVerify(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi verify", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var (
		execFlag  string
		modelFlag string
		timeout   int
	)

	flags.StringVar(&execFlag, "exec", "", "run this command with the prompt on stdin (and {prompt} as a path to it) and store the record it answers")
	flags.StringVar(&modelFlag, "model", "", "model that replaces {model} in --exec (default: the verify key in odd/review.toml, or the session's)")
	flags.IntVar(&timeout, "timeout", verifyapp.DefaultTimeout, "seconds the verifier may take")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		_, _ = fmt.Fprintln(stderr, "odachi: invalid flag: "+oneLine(err.Error())+"; run odachi verify -h to see the flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "odachi: extra argument: %s; odachi verify takes flags only (run odachi verify -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts := verifyapp.Options{
		Dir:            dir,
		Store:          storeRoot(dir),
		Exec:           execFlag,
		TimeoutSeconds: timeout,
		// The silence budgets are the adapter's own numbers, read from the
		// constants its clock uses: the sentence the verb prints and the clock
		// that stopped the run cannot drift apart.
		FirstOutput: shellcmd.FirstOutputTimeout,
		Idle:        shellcmd.IdleTimeout,
	}

	// The verifier's model is resolved here, where every other choice an option
	// carries is read: the flag, then the project's file, then the session. The
	// verb receives one identity and never asks which of the three answered.
	model, err := verifyModel(dir, execFlag, modelFlag)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	opts.Model = model

	// The feature's own words. A project with no task document has nothing to
	// verify, and the verb says so instead of composing a prompt about nothing.
	current, err := tasks.Current(osFiles{}, dir)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	opts.Requirement = firstRequirement(current.Doc)
	opts.DoneWhen = current.Doc.DoneWhen
	opts.Risks = declaredRisks(dir, current.Doc.Base)

	result := verifyapp.Run(opts, osFiles{}, gitRepo{}, osShell{})

	for _, line := range result.Lines {
		_, _ = fmt.Fprintln(stdout, line)
	}
	if result.Code != verifyapp.CodeOK {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(result.Message))
	}
	return result.Code
}

// verifyModel resolves the identity the verifier command runs with: the flag
// wins, the project's file comes next, and the session's model is the last resort.
// The bare form asks nobody, so it pays no probe and reads no file; only a typed
// --model is kept, because the verb refuses a model that would change nothing.
func verifyModel(dir, execFlag, name string) (string, error) {
	model := strings.TrimSpace(name)
	if strings.TrimSpace(execFlag) == "" {
		return model, nil
	}
	if model == "" {
		config, err := reviewapp.Config(dir, osFiles{}, reviewModeCodec{})
		if err != nil {
			return "", err
		}
		model = strings.TrimSpace(config.Verify)
	}
	if model == "" && strings.Contains(execFlag, verifyapp.ModelToken) {
		model = hostModel{}.Model(dir)
	}
	return model, nil
}

// firstRequirement is R1 when the document carries it, and the first requirement
// otherwise: the verifier reads the human's sentence, never a rewrite of it.
func firstRequirement(doc taskdoc.TaskDoc) string {
	for _, requirement := range doc.Requirements {
		if strings.EqualFold(strings.TrimSpace(requirement.ID), "R1") {
			return requirement.Text
		}
	}
	if len(doc.Requirements) > 0 {
		return doc.Requirements[0].Text
	}
	return ""
}

// declaredRisks is what the feature's commits said they can hurt. A repository
// that cannot answer, or a range with no readable declaration, asks for no
// verification: the ready note is where an unreadable line is named, and this verb
// does not judge it a second time.
func declaredRisks(dir, base string) []string {
	if strings.TrimSpace(base) == "" {
		return nil
	}
	messages, err := gitRepo{}.Messages(dir, base)
	if err != nil {
		return nil
	}
	risk, err := change.ParseRisk(messages)
	if err != nil {
		return nil
	}
	return risk.Named
}
