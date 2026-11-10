// Package lensprobe implements `odachi review probe`: the instrument that
// measures the four lenses against a fixture with defects planted in it.
//
// The measurement has one shape. A fixture carries two arms of the same small
// project, a clean one and one with defects, and a manifest that names each
// defect by file, line, lens and severity floor. The probe writes the clean arm
// into a throwaway repository, commits it, overlays the seeded arm, and asks the
// shipping prompt -the same one `review` builds- one question per lens. It then
// scores every answer against the key: found, called low, missed, and noise.
//
// Two properties are what make the number worth reading:
//
//   - the prompt is composed by internal/domain/review, not by this package, so
//     the probe measures the lenses that ship and not a copy of them;
//   - the answers are written under the clone's store, in a folder of their own,
//     because a probe's answer is not a verdict about a tree and must never be
//     read as one by `check`.
//
// Nothing here calls a model by itself: without --exec the verb prints the
// prompts, and with it the human's command answers, exactly like `review`.
package lensprobe

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/probe"
	"github.com/biggs-100/odachi/internal/domain/review"
	assets "github.com/biggs-100/odachi/internal/probe"
)

// Brief is what the fixture's candidate says it does. It is the intent every
// lens is judged against, and it is fixed here because the fixture is fixed: a
// reviewer that looks for what nobody asked for measures nothing.
const Brief = `The candidate adds a small HTTP service: one handler that reads a user,
one in-memory cache with a remote fetch, and one function that caps the size of
an answer, with their tests. Nothing else changed.`

// TimeoutSeconds is what one reviewer may take. It matches `review --exec`: a
// model that stalls at this size does not answer late, it does not answer.
const TimeoutSeconds = 600

// CodeOK is a probe that ran and produced a number, whatever the number is: the
// measurement does not pass or fail, it reports.
const CodeOK = 0

// CodeBlocked is a probe that could not run: no manifest, no reviewer, or an
// answer nobody could read.
const CodeBlocked = 1

// Files is the port the probe writes the fixture through.
type Files interface {
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Exists(path string) (bool, error)
}

// Git is the port that names the candidate: one read of the worktree, the same
// one every gate uses.
type Git interface {
	Candidate(root, base string) (tree string, paths []string, diff string, err error)
}

// Shell is the port that runs the fixture's git and the reviewer's command.
type Shell interface {
	Run(dir, command string, timeoutSeconds int) (stdout, stderr string, exitCode int, timedOut bool, err error)
	Quote(path string) string
}

// Lenses is the port for the four specialist prompts, verbatim.
type Lenses interface {
	Lens(name string) (string, error)
}

// Options is one `odachi review probe` invocation.
type Options struct {
	// Dir is the project directory: the answers are stored under it.
	Dir string
	// Fixture is a directory that holds clean/, seeded/ and manifest.json, or
	// empty for the fixture that ships with the binary.
	Fixture string
	// Exec is the reviewer's command. Empty prints the prompts and runs nothing.
	Exec string
	// Model replaces {model} in Exec.
	Model string
}

// Result is what the CLI prints and exits with.
type Result struct {
	Code int
	// Lines is the report, and Prompt holds the prompts when no reviewer ran.
	Lines  []string
	Prompt string
	// Summary is the last line: what the measurement says.
	Summary string
}

// Run measures the fixture.
func Run(opts Options, files Files, git Git, shell Shell, lenses Lenses) Result {
	manifest, fixture, problem := load(opts)
	if problem != nil {
		return *problem
	}

	work, err := os.MkdirTemp("", "odachi-probe-")
	if err != nil {
		return *blocked("could not make a throwaway directory: " + oneLine(err.Error()))
	}
	defer os.RemoveAll(work)

	tree, paths, diff, problem := materialize(work, fixture, files, shell, git)
	if problem != nil {
		return *problem
	}

	names := []string{review.Risk, review.Readability, review.Reliability, review.Resilience}
	prompts := make(map[string]string, len(names))
	for _, name := range names {
		body, err := lenses.Lens(name)
		if err != nil {
			return *blocked(err.Error())
		}
		prompts[name] = review.Prompt(review.PromptInput{
			Brief:  Brief,
			Tree:   tree,
			Base:   "HEAD",
			Paths:  paths,
			Diff:   diff,
			Lenses: []review.Lens{{Name: name, Body: body}},
		})
	}

	// The key names the lenses this probe owes: a fixture with no readability
	// defect is measured on three lenses, not on the four the review can run.
	owed := make([]string, 0, len(names))
	for _, name := range names {
		for _, defect := range manifest.Defects {
			if defect.Lens == name {
				owed = append(owed, name)
				break
			}
		}
	}

	if strings.TrimSpace(opts.Exec) == "" {
		var out strings.Builder
		for _, name := range owed {
			out.WriteString("# ---- " + name + " ----\n")
			out.WriteString(prompts[name])
			out.WriteString("\n")
		}
		return Result{
			Code:    CodeOK,
			Prompt:  out.String(),
			Summary: "printed " + strconv.Itoa(len(owed)) + " prompts and ran nothing: add --exec to measure",
		}
	}

	verdicts := make(map[string]review.Verdict, len(owed))
	lines := make([]string, 0, len(owed)+2)
	for _, name := range owed {
		answer, problem := ask(opts, files, shell, work, name, prompts[name], tree)
		if problem != nil {
			return *problem
		}
		verdict, err := review.Parse([]byte(answer))
		if err != nil {
			return *blocked("the " + name + " reviewer answered something that is not a verdict: " + oneLine(err.Error()))
		}
		verdicts[name] = verdict
		lines = append(lines, "stored: "+storePath(opts.Dir, name, tree))
	}

	score := probe.ScoreOne(manifest, verdicts)
	lines = append(lines, render(score)...)
	summary := score.Summary()
	lines = append(lines, summary)
	return Result{Code: CodeOK, Lines: lines, Summary: summary}
}

// load reads the manifest and roots the fixture: the one the human named, or the
// one that ships with the binary.
func load(opts Options) (probe.Manifest, fs.FS, *Result) {
	fixture := assets.Fixture()
	root := assets.Root
	if strings.TrimSpace(opts.Fixture) != "" {
		fixture = os.DirFS(opts.Fixture)
		root = "."
	}

	data, err := fs.ReadFile(fixture, slash(root, "manifest.json"))
	if err != nil {
		return probe.Manifest{}, nil, blocked("the fixture has no manifest.json: " + oneLine(err.Error()))
	}
	manifest, err := probe.Parse(data)
	if err != nil {
		return probe.Manifest{}, nil, blocked(err.Error())
	}

	if root == "." {
		return manifest, fixture, nil
	}
	sub, err := fs.Sub(fixture, root)
	if err != nil {
		return probe.Manifest{}, nil, blocked("could not read the bundled fixture: " + oneLine(err.Error()))
	}
	return manifest, sub, nil
}

// materialize writes the clean arm, commits it, overlays the seeded arm and
// reads the candidate. The candidate is the one every gate reads, so the diff
// the reviewer sees here is shaped like the diff it sees in a feature.
func materialize(work string, fixture fs.FS, files Files, shell Shell, git Git) (string, []string, string, *Result) {
	if err := writeArm(work, assets.Clean, fixture, files); err != nil {
		return "", nil, "", blocked("could not write the clean arm: " + oneLine(err.Error()))
	}
	for _, command := range []string{"git init -q", "git add -A", "git commit -qm clean"} {
		stdout, stderr, code, _, err := shell.Run(work, command, 120)
		if err != nil || code != 0 {
			return "", nil, "", blocked("the fixture's git failed: " + oneLine(strings.TrimSpace(stdout+stderr)))
		}
	}
	if err := writeArm(work, assets.Seeded, fixture, files); err != nil {
		return "", nil, "", blocked("could not write the seeded arm: " + oneLine(err.Error()))
	}

	tree, paths, diff, err := git.Candidate(work, "HEAD")
	if err != nil {
		return "", nil, "", blocked("could not read the candidate: " + oneLine(err.Error()))
	}
	return tree, paths, diff, nil
}

// writeArm writes one arm's files under work, keeping their relative paths.
func writeArm(work, arm string, fixture fs.FS, files Files) error {
	return fs.WalkDir(fixture, arm, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(fixture, name)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(name, arm+"/")
		return files.Write(filepath.Join(work, filepath.FromSlash(rel)), data)
	})
}

// ask runs one reviewer and returns its answer, after writing the prompt where
// the command can read it and the answer where a person can find it.
func ask(opts Options, files Files, shell Shell, work, lens, prompt, tree string) (string, *Result) {
	promptPath := filepath.Join(work, "prompt-"+lens+".txt")
	if err := files.Write(promptPath, []byte(prompt)); err != nil {
		return "", blocked("could not write the " + lens + " prompt: " + oneLine(err.Error()))
	}

	command := strings.ReplaceAll(opts.Exec, "{prompt}", shell.Quote(promptPath))
	command = strings.ReplaceAll(command, "{model}", strings.TrimSpace(opts.Model))

	stdout, stderr, code, timedOut, err := shell.Run(work, command, TimeoutSeconds)
	if err != nil {
		return "", blocked("the " + lens + " reviewer could not be run: " + oneLine(err.Error()))
	}
	if timedOut {
		return "", blocked("the " + lens + " reviewer did not answer in " + strconv.Itoa(TimeoutSeconds) + "s")
	}
	if code != 0 {
		return "", blocked("the " + lens + " reviewer exited " + strconv.Itoa(code) + ": " + oneLine(strings.TrimSpace(stderr)))
	}

	if err := files.Write(storePath(opts.Dir, lens, tree), []byte(stdout)); err != nil {
		return "", blocked("could not store the " + lens + " answer: " + oneLine(err.Error()))
	}
	return stdout, nil
}

// storePath is where a probe stores one answer. It is not the review store: a
// probe's answer is about a fixture, and `check` must never read one as a
// verdict for the tree this clone builds.
func storePath(dir, lens, tree string) string {
	return filepath.Join(dir, ".git", "odachi", "probe", tree+"-"+lens+".json")
}

// render prints the score as a table a person reads: one row per lens, then the
// findings nobody planted, then the line that summarises it.
func render(score probe.Score) []string {
	lines := make([]string, 0, len(score.Lenses)+len(score.Noise))
	lines = append(lines, fmt.Sprintf("%-12s %5s %5s %6s", "lens", "found", "low", "missed"))
	for _, row := range score.Lenses {
		lines = append(lines, fmt.Sprintf("%-12s %2d/%-2d %5d %6d", row.Lens, row.Found, row.Total, row.Low, row.Missed))
		if len(row.MissedIDs) > 0 {
			lines = append(lines, "             missed: "+strings.Join(row.MissedIDs, ", "))
		}
	}
	for _, noise := range score.Noise {
		lines = append(lines, fmt.Sprintf("noise        %s:%d (%s, %s)", noise.Finding.Path, noise.Finding.Line, noise.Lens, noise.Finding.Severity))
	}
	return lines
}

// blocked is a probe that could not run, with the one line the human reads.
func blocked(message string) *Result {
	return &Result{Code: CodeBlocked, Summary: message}
}

// slash joins a fixture root and a file in the form an fs.FS uses.
func slash(root, name string) string {
	if root == "." {
		return name
	}
	return root + "/" + name
}

// oneLine collapses a message to one line: a broken fixture reports one reason.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
