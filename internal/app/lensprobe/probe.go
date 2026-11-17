// Package lensprobe implements `odachi review probe`: the instrument that
// measures the four lenses against a fixture with defects planted in it.
//
// The measurement has one shape. A fixture carries two arms of the same small
// project, a clean one and one with defects, and a manifest that names each
// defect by file, line, lens and severity floor. The probe writes the clean arm
// into a throwaway repository, commits it, overlays the seeded arm, and asks the
// shipping prompt -the same one `review` builds- one question per lens. It then
// scores every answer against the key: found, called low, missed, and the
// findings nobody planted, split by what they claim -a claim of a defect is
// noise, a suggestion is a note, on both arms.
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
const Brief = `The candidate adds a small HTTP service to the wiring the repository already
has: one handler that reads the authenticated caller's record, one in-memory
cache with a remote fetch, and one function that caps the size of an answer,
with their tests. The authentication middleware is already there and does not
change.`

// TimeoutSeconds is what one reviewer may take. It matches `review --exec`: a
// model that stalls at this size does not answer late, it does not answer.
const TimeoutSeconds = 600

// MaxRuns is how many draws one measurement may hold. Past ten it is an
// experiment, and an experiment is designed somewhere else.
const MaxRuns = 10

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

// Git is the port that names the candidate and freezes it: one read of the
// worktree, the same one every gate uses, and one export of the tree into a
// directory the reviewer may read.
type Git interface {
	Candidate(root, base string) (tree string, paths []string, diff string, err error)
	ArchiveInto(root, tree, dir string) error
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
	// Runs is how many draws the measurement holds, from one to MaxRuns. A run
	// that asks for zero is refused and told the minimum: one draw is a number,
	// and zero draws is a question nobody measured.
	Runs int
	// Blind runs the arm before the change: the reviewer gets the prompt and no
	// directory, and the run happens in an empty one. It is the other half of the
	// comparison, and the only way to say what the snapshot bought.
	Blind bool
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
	defer func() { _ = os.RemoveAll(work) }()
	// readers are the directories a reviewer may read: each arm gets its own, at
	// the temp root, because a snapshot that sits beside the other arm is not a
	// snapshot: the exam is next door. Their parent is the system's temp root, the
	// same place the copy `review --exec` hands its reviewer lives in.
	var readers []string
	defer func() {
		for _, dir := range readers {
			_ = os.RemoveAll(dir)
		}
	}()
	// The two arms are two candidates: the clean one is a change that carries no
	// defect, and its findings are what the lenses raise on code nobody broke.
	// They are read separately because a reviewer handed both at once knows which
	// half is the exam, and each one is split by what a finding claims: a claim is
	// noise, a suggestion is a note.
	clean, problem := build(work, "clean", fixture, assets.Clean, files, shell, git, opts.Blind)
	if problem != nil {
		return *problem
	}
	readers = appendReader(readers, clean.reader)
	seeded, problem := build(work, "seeded", fixture, assets.Seeded, files, shell, git, opts.Blind)
	if problem != nil {
		return *problem
	}
	readers = appendReader(readers, seeded.reader)

	// The blind arm has no candidate to read, so its reviewer runs in a directory
	// that holds neither arm: a prompt that bans reads is not isolation, and the
	// canary is what says the difference.
	blank := ""
	if opts.Blind {
		root, err := os.MkdirTemp("", "odachi-probe-blank-")
		if err != nil {
			return *blocked("could not make the blind arm's empty directory: " + oneLine(err.Error()))
		}
		readers = append(readers, root)
		blank = filepath.Join(root, "blank")
		if err := os.MkdirAll(blank, 0o755); err != nil {
			return *blocked("could not make the blind arm's empty directory: " + oneLine(err.Error()))
		}
	}

	names := []string{review.Risk, review.Readability, review.Reliability, review.Resilience}
	bodies := make(map[string]string, len(names))
	for _, name := range names {
		body, err := lenses.Lens(name)
		if err != nil {
			return *blocked(err.Error())
		}
		bodies[name] = body
	}

	prompts := make(map[string]map[string]string, 2)
	for _, one := range []candidate{clean, seeded} {
		prompts[one.name] = make(map[string]string, len(names))
		for _, name := range names {
			prompts[one.name][name] = review.Prompt(review.PromptInput{
				Brief:      Brief,
				Tree:       one.tree,
				Base:       "HEAD",
				Paths:      one.paths,
				Diff:       one.diff,
				Lenses:     []review.Lens{{Name: name, Body: bodies[name]}},
				FrozenTree: one.snapshot,
			})
		}
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

	draws := opts.Runs
	if draws < 1 || draws > MaxRuns {
		return *blocked("--runs " + strconv.Itoa(opts.Runs) + ": the measurement holds between 1 and " + strconv.Itoa(MaxRuns) +
			" draws, and each one is a reviewer per lens and arm")
	}

	if strings.TrimSpace(opts.Exec) == "" {
		var out strings.Builder
		for _, one := range []candidate{clean, seeded} {
			for _, name := range names {
				out.WriteString("# ---- " + one.name + " / " + name + " ----\n")
				out.WriteString(prompts[one.name][name])
				out.WriteString("\n")
			}
		}
		return Result{
			Code:    CodeOK,
			Prompt:  out.String(),
			Summary: "printed " + strconv.Itoa(2*len(names)) + " prompts and ran nothing: add --exec to measure",
		}
	}

	scores := make([]probe.Score, 0, draws)
	cleanTallies := make([]probe.Tally, 0, draws)
	lines := make([]string, 0, draws*(len(owed)+len(names))+8)

	for draw := 1; draw <= draws; draw++ {
		seededVerdicts := make(map[string]review.Verdict, len(owed))
		for _, name := range owed {
			answer, stored, problem := ask(opts, files, shell, work, runDir(seeded, blank), "seeded", name, prompts[seeded.name][name], seeded.tree, draw)
			if problem != nil {
				return *problem
			}
			verdict, err := review.Parse([]byte(answerObject(answer)))
			if err != nil {
				return *blocked("the seeded " + name + " reviewer answered something that is not a verdict: " + oneLine(err.Error()))
			}
			seededVerdicts[name] = verdict
			lines = append(lines, "stored: "+stored)
		}

		cleanFindings := make([]probe.Noise, 0)
		for _, name := range names {
			answer, stored, problem := ask(opts, files, shell, work, runDir(clean, blank), "clean", name, prompts[clean.name][name], clean.tree, draw)
			if problem != nil {
				return *problem
			}
			verdict, err := review.Parse([]byte(answerObject(answer)))
			if err != nil {
				return *blocked("the clean " + name + " reviewer answered something that is not a verdict: " + oneLine(err.Error()))
			}
			lines = append(lines, "stored: "+stored)
			for _, finding := range verdict.Findings {
				cleanFindings = append(cleanFindings, probe.Noise{Lens: name, Finding: finding})
			}
		}

		scores = append(scores, probe.ScoreOne(manifest, seededVerdicts))
		cleanTallies = append(cleanTallies, probe.TallyOf(cleanFindings))
	}

	runs := Summarize(scores, cleanTallies)
	// The marker inside the candidate goes where the reviewer could read it: the
	// frozen clean arm. The blind arm has no snapshot, so it goes into the arm's
	// own directory, outside the empty one its command runs in.
	inside := clean.snapshot
	if inside == "" {
		inside = filepath.Join(work, "clean")
	}
	// The marker a scoped reader cannot reach sits beside the directory it was
	// given: its parent, which belongs to this run and holds nothing else.
	dir := runDir(clean, blank)
	outside := filepath.Join(filepath.Dir(dir), "canary-outside.txt")
	canaryLine, problem := runCanary(opts, files, shell, work, inside, outside, dir, opts.Blind)
	if problem != nil {
		return *problem
	}
	lines = append(lines, canaryLine)
	lines = append(lines, "reader: "+readerLine(opts, clean))
	lines = append(lines, renderRuns(runs)...)
	summary := runs.Summary()
	return Result{Code: CodeOK, Lines: lines, Summary: summary}
}

// candidate is one arm materialized into a throwaway repository and read by the
// same port every gate reads.
type candidate struct {
	name string
	tree string
	// reader is the directory this arm's reviewer owns: the candidate and the
	// marker beside it, and nothing else. It exists so the parent of the reviewer's
	// working directory cannot hold the other arm.
	reader string
	// snapshot is the directory the reviewer may read: an export of this arm's
	// tree, or empty in the blind arm, where the reviewer gets no directory.
	snapshot string
	paths    []string
	diff     string
}

// runDir is where the reviewer's command runs: inside the frozen candidate when
// it has one, as `review --exec` does, and in the empty directory otherwise.
func runDir(one candidate, blank string) string {
	if one.snapshot != "" {
		return one.snapshot
	}
	return blank
}

// build writes one arm into its own directory and reads its candidate. The base
// is committed first -the repository the change lands in, which neither arm's
// diff contains- and the arm is the worktree on top of it: the clean arm is the
// same change with no defect in it, and whatever a lens reports there is noise
// nobody planted.
func build(work, name string, fixture fs.FS, arm string, files Files, shell Shell, git Git, blind bool) (candidate, *Result) {
	dir := filepath.Join(work, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return candidate{}, blocked("could not make the " + name + " arm: " + oneLine(err.Error()))
	}

	if err := writeArm(dir, assets.Base, fixture, files); err != nil {
		return candidate{}, blocked("could not write the base of the " + name + " arm: " + oneLine(err.Error()))
	}
	for _, command := range []string{"git init -q", "git add -A", "git commit -q -m base"} {
		stdout, stderr, code, _, err := shell.Run(dir, command, 120)
		if err != nil || code != 0 {
			return candidate{}, blocked("the " + name + " arm's base commit failed: " + oneLine(strings.TrimSpace(stdout+stderr)))
		}
	}

	if err := writeArm(dir, assets.Clean, fixture, files); err != nil {
		return candidate{}, blocked("could not write the clean files of the " + name + " arm: " + oneLine(err.Error()))
	}
	if arm == assets.Seeded {
		if err := writeArm(dir, assets.Seeded, fixture, files); err != nil {
			return candidate{}, blocked("could not write the seeded arm: " + oneLine(err.Error()))
		}
	}

	// The candidate is the worktree against the base commit: nothing is committed
	// on top, so the diff is exactly what the change adds or rewrites.
	stdout, stderr, code, _, err := shell.Run(dir, "git add -A", 120)
	if err != nil || code != 0 {
		return candidate{}, blocked("the " + name + " arm's git failed: " + oneLine(strings.TrimSpace(stdout+stderr)))
	}

	tree, paths, diff, err := git.Candidate(dir, "HEAD")
	if err != nil {
		return candidate{}, blocked("could not read the " + name + " candidate: " + oneLine(err.Error()))
	}

	// The reviewer reads the frozen arm, not the directory the arms are built in:
	// a command that ran beside the seeded arm could read the exam it is graded on.
	reader, snapshot := "", ""
	if !blind {
		reader, err = os.MkdirTemp("", "odachi-probe-"+name+"-")
		if err != nil {
			return candidate{}, blocked("could not make the " + name + " arm's reader: " + oneLine(err.Error()))
		}
		snapshot = filepath.Join(reader, "candidate")
		if err := git.ArchiveInto(dir, tree, snapshot); err != nil {
			_ = os.RemoveAll(reader)
			return candidate{}, blocked("could not freeze the " + name + " arm: " + oneLine(err.Error()))
		}
	}
	return candidate{name: name, tree: tree, reader: reader, snapshot: snapshot, paths: paths, diff: diff}, nil
}

// appendReader keeps a directory the run has to remove, skipping the empty one a
// blind arm carries.
func appendReader(readers []string, dir string) []string {
	if dir == "" {
		return readers
	}
	return append(readers, dir)
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
func ask(opts Options, files Files, shell Shell, work, dir, arm, lens, prompt, tree string, draw int) (string, string, *Result) {
	promptPath := filepath.Join(work, "prompt-"+arm+"-"+lens+".txt")
	if err := files.Write(promptPath, []byte(prompt)); err != nil {
		return "", "", blocked("could not write the " + lens + " prompt: " + oneLine(err.Error()))
	}

	command := strings.ReplaceAll(opts.Exec, "{prompt}", shell.Quote(promptPath))
	command = strings.ReplaceAll(command, "{model}", strings.TrimSpace(opts.Model))

	stdout, stderr, code, timedOut, err := shell.Run(dir, command, TimeoutSeconds)
	if err != nil {
		return "", "", blocked("the " + lens + " reviewer could not be run: " + oneLine(err.Error()))
	}
	if timedOut {
		return "", "", blocked("the " + lens + " reviewer did not answer in " + strconv.Itoa(TimeoutSeconds) + "s")
	}
	if code != 0 {
		return "", "", blocked("the " + lens + " reviewer exited " + strconv.Itoa(code) + ": " + oneLine(strings.TrimSpace(stderr)))
	}

	stored := storePath(opts, arm, lens, tree, draw, files)
	if err := files.Write(stored, []byte(stdout)); err != nil {
		return "", "", blocked("could not store the " + lens + " answer: " + oneLine(err.Error()))
	}
	// The question is kept with the answer: a measurement whose prompt was not
	// recorded cannot show what the reviewer actually saw, so "same or better"
	// cannot be established across two runs. The name shares the answer's stem,
	// so the first-free-number rule keeps the pairs together.
	question := strings.TrimSuffix(stored, ".json") + ".prompt.md"
	if err := files.Write(question, []byte(prompt)); err != nil {
		return "", "", blocked("could not store the " + lens + " prompt: " + oneLine(err.Error()))
	}
	return stdout, stored, nil
}

// runCanary asks one extra question of the same reviewer: whether it can read a
// marker inside the frozen candidate and cannot read a marker beside it. The
// answer is filed with its prompt like any other probe answer, so a person can
// check the two legs later. A canary that did not answer is reported as such,
// never read as isolation.
//
// The blind arm has nothing to attach to, so both markers sit outside the empty
// directory its reviewer runs in and any quote is "too wide": a reviewer that
// found one went looking, and its number would not be the arm's.
func runCanary(opts Options, files Files, shell Shell, work, inside, outside, dir string, blind bool) (string, *Result) {
	const insideMark = "ODACHI-CANARY-INSIDE"
	const outsideMark = "ODACHI-CANARY-OUTSIDE"
	for path, mark := range map[string]string{
		filepath.Join(inside, "canary-inside.txt"): insideMark,
		outside: outsideMark,
	} {
		if err := files.Write(path, []byte(mark+"\n")); err != nil {
			return "", blocked("could not write the access canary: " + oneLine(err.Error()))
		}
	}

	canary := Canary{Inside: insideMark, Outside: outsideMark}
	prompt := canary.Prompt()
	promptPath := filepath.Join(work, "canary-prompt.txt")
	if err := files.Write(promptPath, []byte(prompt)); err != nil {
		return "", blocked("could not write the access canary prompt: " + oneLine(err.Error()))
	}
	command := strings.ReplaceAll(opts.Exec, "{prompt}", shell.Quote(promptPath))
	command = strings.ReplaceAll(command, "{model}", strings.TrimSpace(opts.Model))
	stdout, stderr, code, timedOut, err := shell.Run(dir, command, TimeoutSeconds)
	if err != nil || timedOut || code != 0 {
		return "canary: no answer", nil
	}

	stored := storePath(opts, "canary", "canary", "fixture", 1, files)
	if err := files.Write(stored, []byte(stdout)); err != nil {
		return "", blocked("could not store the access canary answer: " + oneLine(err.Error()))
	}
	if err := files.Write(strings.TrimSuffix(stored, ".json")+".prompt.md", []byte(prompt)); err != nil {
		return "", blocked("could not store the access canary prompt: " + oneLine(err.Error()))
	}
	if blind {
		return canary.ReadBlind(stdout + stderr), nil
	}
	return canary.Read(stdout + stderr), nil
}

// answerObject is the JSON object a reviewer answered, without the narration a
// model writes around it: "I'll read the frozen candidate" is a status line, and
// the verb asked for one object. The stored file keeps the whole answer, so the
// sentence a model wrote is still there for a person to read.
func answerObject(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return text
	}
	return text[start : end+1]
}

// readerLine says which arm the number belongs to: a person comparing two runs
// has to know whether the reviewer could read the candidate, and a report that
// leaves it to memory is a report about an unknown question.
func readerLine(opts Options, clean candidate) string {
	if opts.Blind {
		return "no directory (--blind): the prompt is the whole candidate"
	}
	return "the frozen candidate at " + clean.snapshot
}

// storePath is where a probe stores one answer. It is not the review store: a
// probe's answer is about a fixture, and `check` must never read one as a
// verdict for the tree this clone builds.
//
// The fixture's tree is the same on every run, so the tree alone would make the
// second measurement overwrite the first: the name carries the first free
// number, and the store keeps the runs a person compares. The blind arm is a
// different question about the same bytes, so its answers carry the arm in the
// name too: evidence answers only its own question.
func storePath(opts Options, arm, lens, tree string, draw int, files Files) string {
	if opts.Blind {
		arm += "-blind"
	}
	base := filepath.Join(opts.Dir, ".git", "odachi", "probe", tree+"-"+arm+"-"+lens)
	if draw > 1 {
		base = base + "-r" + strconv.Itoa(draw)
	}
	path := base + ".json"
	for i := 2; ; i++ {
		if exists, err := files.Exists(path); err != nil || !exists {
			return path
		}
		path = base + "-" + strconv.Itoa(i) + ".json"
	}
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
