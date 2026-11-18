package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/biggs-100/odachi/internal/adapters/skillhome"
	skillevalapp "github.com/biggs-100/odachi/internal/app/skilleval"
	"github.com/biggs-100/odachi/internal/app/start"
	"github.com/biggs-100/odachi/internal/domain/store"
	corpus "github.com/biggs-100/odachi/internal/skilleval"
	"github.com/biggs-100/odachi/skills"
)

const skillsEvalUsage = `usage: odachi skills eval <skill> [--exec "<command>"] [--runs N] [--timeout N] [--move-home] [--recover]`

// runSkillsEval wires the exam: the corpus that ships with the binary, the disk
// the arms are built on, the shell the human named, and the home the control arm
// has to take the skill out of.
func runSkillsEval(dir string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi skills eval", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var opts skillevalapp.Options
	flags.StringVar(&opts.Exec, "exec", "", "run this command once per draw with {prompt} as the file to read; it answers the exam")
	flags.IntVar(&opts.Runs, "runs", 3, "draws each arm answers")
	flags.IntVar(&opts.Timeout, "timeout", 600, "seconds one draw may take")
	flags.BoolVar(&opts.MoveHome, "move-home", false, "take the skill out of ~/.agents/skills for the run, and put it back")
	flags.BoolVar(&opts.Recover, "recover", false, "put back what a run left out, and measure nothing")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprintln(stdout, skillsEvalUsage)
			return start.CodeOK
		}
		_, _ = fmt.Fprintf(stderr, "odachi: invalid flag: %s; %s\n", oneLine(err.Error()), skillsEvalUsage)
		return start.CodeBlocked
	}
	if flags.NArg() != 1 {
		_, _ = fmt.Fprintln(stderr, "odachi: the exam names one skill; "+skillsEvalUsage)
		return start.CodeBlocked
	}
	name := flags.Arg(0)

	exam, err := corpus.Load(name)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	opts.Exam = exam
	opts.Skill, err = bundledSkill(name)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: "+oneLine(err.Error()))
		return start.CodeBlocked
	}

	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: I cannot tell where the home is: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	mover := skillhome.Mover{Skills: filepath.Join(home, ".agents", "skills")}
	// The journal lives in the clone's store, and a run that only prints the
	// prompt never writes one: a clone is not owed a file for a question.
	if opts.Exec != "" || opts.Recover {
		root, err := store.Root(dir, gitRepo{})
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "odachi: the exam keeps its journal in the clone's store, and this folder has none: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		mover.Journal = filepath.Join(root, "evals", "home.json")
	}

	scratch, err := os.MkdirTemp("", "odachi-eval-")
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "odachi: could not make room for the arms: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	opts.Scratch = scratch

	result := skillevalapp.Run(opts, evalDisk{}, osShell{}, mover)
	for _, line := range result.Lines {
		_, _ = fmt.Fprintln(stdout, line)
	}
	if result.Code != start.CodeOK {
		_, _ = fmt.Fprintln(stderr, "odachi: "+result.Message)
	}
	return result.Code
}

// bundledSkill answers the files of one shipping skill, which are what the arm
// with the skill writes into its project. The skill is the product: a name that
// does not ship has no text to measure.
func bundledSkill(name string) ([]corpus.File, error) {
	all, err := skills.All()
	if err != nil {
		return nil, fmt.Errorf("the bundled skills cannot be read: %w", err)
	}
	for _, skill := range all {
		if skill.Name != name {
			continue
		}
		files := make([]corpus.File, 0, len(skill.Files))
		for _, file := range skill.Files {
			files = append(files, corpus.File{Path: file.Path, Content: file.Content})
		}
		return files, nil
	}
	return nil, fmt.Errorf("the skill %q does not ship with this binary: there is no text to measure", name)
}

// evalDisk is the disk one eval run writes on: the arms are folders this run made,
// and it makes them itself.
type evalDisk struct{ osFiles }

func (evalDisk) MkdirAll(path string) error { return os.MkdirAll(path, 0o755) }
