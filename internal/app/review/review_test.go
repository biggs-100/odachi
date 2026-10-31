package review_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// dir is the project root the run sees. It never touches a real disk: every port
// is a double.
const dir = "/proyecto"

// store is the clone's state, inside the project it describes and outside the
// worktree: the real one is <proyecto>/.git/odachi.
var store = filepath.Join(dir, ".git", "odachi")

// The two candidates. A different tree is a different verdict, and that is the
// property this store exists for.
const (
	currentTree = "4a1c9f7d2e3b5a6c8d9e0f1a2b3c4d5e6f7a8b9c"
	otherTree   = "9f2b7c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b"
)

// now is the instant a stored verdict records, fixed so the file is byte-stable.
var now = time.Date(2026, 10, 21, 20, 31, 0, 0, time.UTC)

// brief is a brief complete enough for a review to have something to judge
// against.
const brief = `# turnos

## Qué tiene que existir
- reservar turnos

## Listo cuando
- reservo un turno y aparece en la agenda
`

// diff is a real-shaped diff: two lines that changed.
const diff = "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n"

// fakeFiles is the Files port in memory. It records every write, because "nothing
// outside the store's reviews/ is written" is the property a test has to be able to
// see.
type fakeFiles struct {
	files    map[string][]byte
	writes   []string
	backups  []string
	err      error
	existErr error
	writeErr error
	// failExists fails Exists for one path only, so a test can reach the refusal
	// without breaking the store checks that run before it.
	failExists func(path string) error
	// failWrite fails Write for one path only, so a test can fail the question's
	// file without failing the answer's.
	failWrite func(path string) error
	// failList counts and can fail a listing, so a test can reach the drop of the
	// refused answers without breaking the read of the potent rows before it.
	failList func(dir string) error
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.files[fullPath(path)] = []byte(content)
	}
	return f
}

// fullPath resolves a project-relative path the way the app does.
func fullPath(rel string) string {
	return filepath.Join(dir, filepath.FromSlash(rel))
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	if f.existErr != nil {
		return false, f.existErr
	}
	if f.failExists != nil {
		if err := f.failExists(path); err != nil {
			return false, err
		}
	}
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	data, ok := f.files[path]
	if !ok {
		return nil, errors.New("missing file: " + path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	if f.err != nil {
		return f.err
	}
	if f.writeErr != nil {
		return f.writeErr
	}
	if f.failWrite != nil {
		if err := f.failWrite(path); err != nil {
			return err
		}
	}
	f.writes = append(f.writes, path)
	f.files[path] = data
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.backups = append(f.backups, path)
	return path + ".bak-test", nil
}

// List answers the verdict-file names the store directory holds, sorted so two
// runs see the same order. Only basenames: the app filters what it reads.
func (f *fakeFiles) List(dir string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.failList != nil {
		if err := f.failList(dir); err != nil {
			return nil, err
		}
	}

	prefix := filepath.ToSlash(dir)
	prefix = strings.TrimSuffix(prefix, "/") + "/"
	var names []string
	for path := range f.files {
		slashed := filepath.ToSlash(path)
		if !strings.HasPrefix(slashed, prefix) {
			continue
		}
		names = append(names, filepath.Base(slashed))
	}
	slices.Sort(names)
	return names, nil
}

// Remove deletes one file, as the store does when a sealed verdict supersedes the
// answers kept for its tree.
func (f *fakeFiles) Remove(path string) error {
	if f.err != nil {
		return f.err
	}
	delete(f.files, path)
	return nil
}

// written is every write as a project-relative path with forward slashes, in the
// order the run made them.
func (f *fakeFiles) written() []string {
	rel := make([]string, 0, len(f.writes))
	for _, path := range f.writes {
		rel = append(rel, relative(path))
	}
	return rel
}

// relative turns an absolute fake path back into the project-relative spelling
// the messages use.
func relative(path string) string {
	slashed := filepath.ToSlash(path)
	return strings.TrimPrefix(slashed, filepath.ToSlash(dir)+"/")
}

// fakeGit is the Git port: fixed facts, and the bases it was asked about.
type fakeGit struct {
	tree    string
	changed []string
	diff    string
	bases   []string
	err     error
	// candidateErr fails only the candidate read, so a test can express "the
	// tree is readable but the base is not".
	candidateErr error

	// ignored is what the repository's rules answer about the store, and
	// ignoredErr is the probe failing. A failed probe must never be read as "the
	// store is already protected".
	ignored    bool
	ignoredErr error
	ignoredAs  []string

	// blobs answers BlobHash per path; blobErr is the port failing for real.
	blobs   map[string]string
	blobErr error

	// candidateCalls counts the one read of the worktree this run asked for.
	candidateCalls int

	// corrections records the trees a correction measured its delta against.
	corrections []string
}

func (g *fakeGit) TreeHash(root string) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.tree, nil
}

// Candidate is the one read: it records every call, because "the verb reads the
// worktree once" is the property this port exists for.
func (g *fakeGit) Candidate(root, base string) (string, []string, string, error) {
	g.candidateCalls++
	g.bases = append(g.bases, base)
	if g.err != nil {
		return "", nil, "", g.err
	}
	if g.candidateErr != nil {
		return "", nil, "", g.candidateErr
	}
	return g.tree, g.changed, g.diff, nil
}

// Correction answers the delta against the tree a correction names: the fake has
// one candidate, so it reports the same bytes and records the tree it was asked
// about, which is what a test asserts.
func (g *fakeGit) Correction(root, tree string) ([]string, string, error) {
	g.corrections = append(g.corrections, tree)
	if g.err != nil {
		return nil, "", g.err
	}
	return g.changed, g.diff, nil
}

func (g *fakeGit) Ignored(root, path string) (bool, error) {
	g.ignoredAs = append(g.ignoredAs, path)
	if g.ignoredErr != nil {
		return false, g.ignoredErr
	}
	return g.ignored, nil
}

// blobs answers BlobHash per path. A path with no entry produces no error and an
// empty receipt, which is the "file is not here" answer, not a git failure.
func (g *fakeGit) BlobHash(root, path string) (string, error) {
	if g.blobErr != nil {
		return "", g.blobErr
	}
	return g.blobs[path], nil
}

// fakeHost is the Host port: the model the session says it is using, or none.
type fakeHost struct {
	model string
	asked bool
	dir   string
}

func (h *fakeHost) Model(dir string) string {
	h.asked = true
	h.dir = dir
	return h.model
}

// fakeShell is the Shell port for --exec.
type fakeShell struct {
	output   string
	stderr   string
	exitCode int
	timedOut bool
	err      error

	asked   bool
	command string
	timeout int
	// commands is every command that ran, in order: one per model group.
	commands []string

	// script is one answer per call, in order: it is how a test makes the
	// second group fail after the first one answered. Without it every call
	// gets the answer above.
	script []fakeAnswer

	// after runs once per call, after the answer: it is how a test makes the
	// reviewer change the worktree it was asked to judge.
	after func()

	// input is the prompt the reviewer was given, once --exec delivers one.
	input string

	// quote is how this fake spells one path for the shell; nil keeps the path.
	quote func(string) string
}

// fakeAnswer is one call's result in a scripted shell.
type fakeAnswer struct {
	output   string
	stderr   string
	exitCode int
	timedOut bool
	err      error
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	answer := fakeAnswer{output: s.output, stderr: s.stderr, exitCode: s.exitCode, timedOut: s.timedOut, err: s.err}
	if len(s.script) > 0 {
		// The last answer repeats: a test says how the first groups behave and
		// lets the rest keep the same shape.
		answer = s.script[min(len(s.commands), len(s.script)-1)]
	}

	s.asked = true
	s.command = command
	s.timeout = timeoutSeconds
	s.commands = append(s.commands, command)
	if s.after != nil {
		s.after()
	}
	return answer.output, answer.stderr, answer.exitCode, answer.timedOut, answer.err
}

// Quote spells a path the way the fake shell wants it; nil is the path itself.
func (s *fakeShell) Quote(path string) string {
	if s.quote != nil {
		return s.quote(path)
	}
	return path
}

func (s *fakeShell) RunInput(dir, command, input string, timeoutSeconds int) (string, string, int, bool, error) {
	s.input = input
	return s.Run(dir, command, timeoutSeconds)
}

// project is one run's wiring: an in-memory project, and the fakes behind it.
type project struct {
	files  *fakeFiles
	git    *fakeGit
	shell  *fakeShell
	lenses *fakeLenses
	host   *fakeHost
}

// newProject seeds a project with a brief and one changed file.
func newProject() *project {
	return &project{
		files: newFakeFiles(map[string]string{"brief.md": brief}),
		git:   &fakeGit{tree: currentTree, changed: []string{"a.go"}, diff: diff},
		shell: &fakeShell{},
		host:  &fakeHost{},
	}
}

// TestEachModelRunsItsOwnReviewer: two lenses, two models, two runs, and one
// verdict that says who judged what. The cheap model carries the cheap lens and
// the sharp one only runs where the human put it.
func TestEachModelRunsItsOwnReviewer(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"risk": "sharp/model"},
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if len(p.shell.commands) != 2 {
		t.Fatalf("commands = %v, want one run per model", p.shell.commands)
	}
	ran := strings.Join(p.shell.commands, " | ")
	if !strings.Contains(ran, "host/default") || !strings.Contains(ran, "sharp/model") {
		t.Errorf("commands = %v, want one run on each model", p.shell.commands)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if stored.Models["risk"] != "sharp/model" || stored.Models["reliability"] != "host/default" {
		t.Errorf("models = %v, want each lens with the model that judged it", stored.Models)
	}
	if stored.Model != "" {
		t.Errorf("Model = %q, want empty when two models judged the tree", stored.Model)
	}
	if !slices.Contains(stored.Lenses, "risk") || !slices.Contains(stored.Lenses, "reliability") {
		t.Errorf("lenses = %v, want both", stored.Lenses)
	}
}

// TestTheSameModelIsOneRun: a lens assigned to the model the session already uses
// is not a second group. Two groups would be two calls paid for one answer.
func TestTheSameModelIsOneRun(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"risk": "host/default"},
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 1 {
		t.Fatalf("commands = %v, want one run: both lenses end on the same model", p.shell.commands)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if stored.Models["risk"] != "host/default" || stored.Models["reliability"] != "host/default" {
		t.Errorf("models = %v, want both lenses on the model that judged them", stored.Models)
	}
}

// TestAFailedGroupKeepsTheAnswerThatLanded: two models, two groups, the second
// fails. The first answer was paid for, so it is on disk; the run says how many
// lenses of how many landed and which one is still owed.
func TestAFailedGroupKeepsTheAnswerThatLanded(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`},
		{output: "boom: the model did not answer", exitCode: 1},
	}

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test asks for a wider scope",
		LensModels: map[string]string{"risk": "sharp/model"},
	})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if len(p.shell.commands) != 2 {
		t.Fatalf("commands = %v, want the two groups", p.shell.commands)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the answer that landed was not stored: %v", err)
	}
	if !slices.Contains(stored.Lenses, "risk") {
		t.Errorf("lenses = %v, want the lens whose reviewer answered", stored.Lenses)
	}
	if slices.Contains(stored.Lenses, "reliability") {
		t.Errorf("lenses = %v, want no lens the failed reviewer never answered", stored.Lenses)
	}
	if !strings.Contains(result.Message, "saved 1 of 2") {
		t.Errorf("message = %q, want how many lenses of how many landed", result.Message)
	}
	if !strings.Contains(result.Message, "missing reliability") {
		t.Errorf("message = %q, want the lens still owed named", result.Message)
	}
	if !strings.Contains(result.Output, "boom") {
		t.Errorf("output = %q, want the reviewer's own words", result.Output)
	}
	if want := domainreview.Dir + "/" + currentTree + ".json"; !slices.Contains(result.Files, want) {
		t.Errorf("files = %v, want the verdict that landed %q", result.Files, want)
	}
}

// TestAReviewerThatMovesTheTreeCannotSplitTheAnswer: the prompt named one
// candidate and every group's answer belongs to those bytes. A reviewer that
// edits the worktree must not move where the answer is filed: one tree, one file,
// every lens that answered.
func TestAReviewerThatMovesTheTreeCannotSplitTheAnswer(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`}, {output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`}}
	calls := 0
	p.shell.after = func() {
		// The second reviewer edits the worktree: from here on, the tree holds
		// bytes the prompt never described.
		calls++
		if calls == 2 {
			p.git.tree = otherTree
		}
	}

	p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test asks for a wider scope",
		LensModels: map[string]string{"risk": "sharp/model"},
	})

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the answer was not filed under the tree the prompt named: %v", err)
	}
	if !slices.Contains(stored.Lenses, "risk") || !slices.Contains(stored.Lenses, "reliability") {
		t.Errorf("lenses = %v, want both groups under the named tree", stored.Lenses)
	}
	if _, ok := p.files.files[storePath(otherTree)]; ok {
		t.Fatalf("the answer was refiled under the tree the reviewer left")
	}
}

// TestAReviewerThatMovesTheTreeStopsTheRun: the answer stays under the tree it
// judged, but the run cannot call the tree reviewed: the bytes in it now are not
// the ones the verdict describes. It stops, names both trees, and says where the
// answer lives.
func TestAReviewerThatMovesTheTreeStopsTheRun(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.script = []fakeAnswer{{output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`}, {output: `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`}}
	calls := 0
	p.shell.after = func() {
		calls++
		if calls == 2 {
			p.git.tree = otherTree
		}
	}

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "risk",
		Because:    "the test asks for a wider scope",
		LensModels: map[string]string{"risk": "sharp/model"},
	})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for a tree the verdict does not describe (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "changed the worktree") {
		t.Errorf("message = %q, want it to say the reviewer moved the tree", result.Message)
	}
	if !strings.Contains(result.Message, otherTree) {
		t.Errorf("message = %q, want the tree the worktree holds now", result.Message)
	}
	if !strings.Contains(result.Message, "2 of 2 lenses") {
		t.Errorf("message = %q, want the answers that landed counted", result.Message)
	}
	if want := domainreview.Dir + "/" + currentTree + ".json"; !slices.Contains(result.Files, want) {
		t.Errorf("files = %v, want the file the answer lives in %q", result.Files, want)
	}
	if _, ok := p.files.files[storePath(otherTree)]; ok {
		t.Fatalf("the answer was refiled under the tree the reviewer left")
	}
}

// TestExecOnADocumentsOnlyChangeSaysNothingIsDue: a candidate with no lens asks
// for no reviewer, and the message must not claim a verdict exists: the store has
// none, and a lie sends the human looking for a file nobody wrote.
func TestExecOnADocumentsOnlyChangeSaysNothingIsDue(t *testing.T) {
	p := newProject()
	p.git.changed = []string{"README.md"}
	p.git.diff = "diff --git a/README.md b/README.md\n--- a/README.md\n+++ b/README.md\n@@ -1 +1 @@\n-viejo\n+nuevo\n"

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for a change with nothing to review (%s)", result.Code, result.Message)
	}
	if strings.Contains(result.Message, "already has a verdict") {
		t.Errorf("message = %q, want no claim that a verdict exists", result.Message)
	}
	if !strings.Contains(result.Message, "asks for no lens") || !strings.Contains(result.Message, "documents only") {
		t.Errorf("message = %q, want the reason and no lens due", result.Message)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "--lens all") {
		t.Errorf("lines = %v, want the command that asks anyway", result.Lines)
	}
	if p.shell.asked {
		t.Error("a reviewer ran for a change with no lens")
	}
	if writes := p.verdictWrites(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none", writes)
	}
}

// TestASecondRunCompletesTheVerdict: the store grows by lens. A run that finds a
// verdict covering reliability judges risk and adds it, keeping the rows it did
// not pay for again.
func TestASecondRunCompletesTheVerdict(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"risk"}]}`

	first, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"reliability"},
		Passes:   []domainreview.Pass{{Path: "a.go", Blob: gitBlob, Lens: "reliability"}},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(first)

	result := p.run(review.Options{Exec: "revisor --model {model}", Lens: "risk", Because: "the test asks for a wider scope"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "2 of 2 lenses") {
		t.Errorf("message = %q, want the completed count: the run says how many lenses the verdict holds", result.Message)
	}
	if len(p.shell.commands) != 1 {
		t.Fatalf("commands = %v, want only the lens the verdict still owed", p.shell.commands)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the completed verdict does not parse: %v", err)
	}
	if !slices.Contains(stored.Lenses, "reliability") || !slices.Contains(stored.Lenses, "risk") {
		t.Fatalf("lenses = %v, want the old answer and the new one", stored.Lenses)
	}
	var kept bool
	for _, pass := range stored.Passes {
		if pass.Lens == "reliability" && pass.Path == "a.go" {
			kept = true
		}
	}
	if !kept {
		t.Errorf("the second run dropped the rows the first one paid for: %v", stored.Passes)
	}
	if !strings.Contains(stored.Reviewer, "first") {
		t.Errorf("reviewer = %q, want the earlier reviewer named", stored.Reviewer)
	}
	if stored.Verdict == "" {
		t.Errorf("the completed verdict has no derived verdict field: %+v", stored)
	}
}

// TestASecondGroupCannotRideOnTheFirstDeclaration: the store lends a verdict its
// rows, never its declaration. A second lens group inherits the first group's
// inspection the moment the merge runs, so checking the merged verdict would let
// an answer that declares nothing be sealed under somebody else's reading. The
// answer is judged first, and a group that says nothing is refused.
func TestASecondGroupCannotRideOnTheFirstDeclaration(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"findings":[],"passes":[{"path":"a.go","lens":"risk"}]}`

	first, err := domainreview.Seal(domainreview.Verdict{
		Reviewer:   "first",
		Lenses:     []string{"reliability"},
		Inspection: &domainreview.Inspection{Status: domainreview.InspectionCompleted, Paths: []string{"a.go"}},
		Passes:     []domainreview.Pass{{Path: "a.go", Blob: gitBlob, Lens: "reliability"}},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(first)

	result := p.run(review.Options{Exec: "revisor --model {model}", Lens: "risk", Because: "the test asks for a wider scope"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1: the second answer declared nothing (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "inspection") {
		t.Errorf("message = %q, want the field the answer is missing", result.Message)
	}
	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the seeded verdict no longer parses: %v", err)
	}
	if slices.Contains(stored.Lenses, "risk") {
		t.Errorf("lenses = %v: an answer that declared nothing was sealed", stored.Lenses)
	}
}

// TestAPartialRunStillStores: coverage is asked about the lenses this run
// claimed, not about the tree. A sensitive candidate asks for four lenses, and
// when two models split them the first group's answer - complete for the two
// lenses it was asked - is stored before the second group even starts. Refusing
// it for the rows the other model owns would make the split impossible and throw
// away an answer that was already paid for.
func TestAPartialRunStillStores(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.git.changed = []string{"auth/login.go"}
	p.git.diff = "diff --git a/auth/login.go b/auth/login.go\n--- a/auth/login.go\n+++ b/auth/login.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n"
	p.git.blobs = map[string]string{"auth/login.go": gitBlob}
	p.shell.script = []fakeAnswer{
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["auth/login.go"]},"passes":[{"path":"auth/login.go","lens":"risk"},{"path":"auth/login.go","lens":"readability"}]}`},
		{output: `{"findings":[],"inspection":{"status":"completed","paths":["auth/login.go"]},"passes":[{"path":"auth/login.go","lens":"reliability"},{"path":"auth/login.go","lens":"resilience"}]}`},
	}

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		LensModels: map[string]string{"risk": "sharp/model", "readability": "sharp/model"},
	})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(stored.Lenses) != 4 {
		t.Errorf("lenses = %v, want the answer of each group", stored.Lenses)
	}
}

// TestAFullAnswerThatCoversEverythingIsUntouched: the command gains a promise,
// not a line. An answer with one row per file and lens stores exactly as before,
// and the human hears the same message.
func TestAFullAnswerThatCoversEverythingIsUntouched(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, line := range result.Lines {
		if strings.Contains(line, "rows missing") || strings.Contains(line, "does not cover") {
			t.Errorf("a complete answer gained a refusal line: %q", line)
		}
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Fatal("a complete answer was not stored")
	}
}

// TestARunDoesNotAnswerTheSameLensTwice: a verdict that already covers what the
// evidence asks for is not rewritten, and no reviewer is started.
func TestARunDoesNotAnswerTheSameLensTwice(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"

	stored, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"reliability"},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(stored)

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code == review.CodeOK {
		t.Fatalf("the run rewrote a verdict that already answered its lens: %s", result.Message)
	}
	if !strings.Contains(result.Message, "already has a verdict") {
		t.Errorf("message = %q, want the tree said answered, which is true here", result.Message)
	}
	if len(p.shell.commands) != 0 {
		t.Errorf("the reviewer ran for a lens that was already answered: %v", p.shell.commands)
	}
}

// --show names the model that will judge each lens before anything runs: that is
// the plan the human pays for. A lens with no assignment says so instead of
// pretending the tool chooses.
func TestShowNamesTheModelOfEachLens(t *testing.T) {
	p := newProject()

	got := p.run(review.Options{Show: true, LensModels: map[string]string{"reliability": "cheap/model"}})

	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, "model reliability: cheap/model") {
		t.Errorf("--show does not name the assigned model:\n%s", lines)
	}
	if strings.Contains(lines, "model reliability: the reviewer command decides") {
		t.Errorf("--show ignores the assignment:\n%s", lines)
	}
	// Only the lenses that will run are in the plan: a line for a lens the
	// evidence did not ask for would be a cost nobody approved.
	if strings.Contains(lines, "model risk:") {
		t.Errorf("--show plans a lens the candidate does not require:\n%s", lines)
	}
}

// TestShowCallsAPartialVerdictPartial: a verdict that answered one of the two
// lenses the tree requires is half an answer. --show says partial, counts what
// the file holds, and names the lens nobody judged, so the gate cannot be read as
// satisfied.
func TestShowCallsAPartialVerdictPartial(t *testing.T) {
	p := newProject()
	partial, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"risk"},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(partial)

	got := p.run(review.Options{Show: true, Lens: "risk"})

	if !strings.Contains(got.Message, "partial") {
		t.Errorf("message = %q, want it to call the verdict partial", got.Message)
	}
	if !strings.Contains(got.Message, "1 of 2 lenses") {
		t.Errorf("message = %q, want how much of the tree it covers", got.Message)
	}
	lines := strings.Join(got.Lines, "\n")
	if !strings.Contains(lines, "missing lenses: reliability") {
		t.Errorf("--show does not name the lens nobody judged:\n%s", lines)
	}

	complete, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"risk", "reliability"},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the complete verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(complete)

	done := p.run(review.Options{Show: true, Lens: "risk"})
	if strings.Contains(done.Message, "partial") {
		t.Errorf("message = %q, want no partial for a verdict that answered both", done.Message)
	}
}

// run performs one `odachi review` with the fixed clock.
func (p *project) run(opts review.Options) review.Result {
	opts.Dir = dir
	opts.Store = store
	return p.runWith(opts)
}

// runWithoutStore is the same run in a folder that has no repository: the CLI
// leaves the store empty there, and the verb has to refuse it.
func (p *project) runWithoutStore(opts review.Options) review.Result {
	opts.Dir = dir
	opts.Store = ""
	return p.runWith(opts)
}

// runWith performs one run with the ports the project already holds.
func (p *project) runWith(opts review.Options) review.Result {
	if opts.Now.IsZero() {
		opts.Now = now
	}
	if p.lenses == nil {
		p.lenses = &fakeLenses{}
	}
	if p.host == nil {
		p.host = &fakeHost{}
	}
	return review.Run(opts, p.files, p.git, p.shell, p.lenses, p.host)
}

// verdictWrites is what a run wrote that claims something. The store's own
// ignore file is plumbing, not an answer: --exec prepares the store before the
// reviewer starts, because the prompt names the tree the verdict will be filed
// under and creating that file changes the tree.
func (p *project) verdictWrites() []string {
	var claims []string
	for _, written := range p.files.written() {
		if strings.HasSuffix(filepath.ToSlash(written), ".gitignore") {
			continue
		}
		// The question a verdict answered is filed beside it (0088): it is the
		// record of what the reviewer was handed, not a claim about the tree.
		if strings.Contains(filepath.Base(written), ".prompt") {
			continue
		}
		claims = append(claims, written)
	}
	return claims
}

// storePath is where a verdict for one tree lives, in the spelling the app uses.
func storePath(tree string) string {
	return storeFile(domainreview.Dir + "/" + tree + ".json")
}

// storeFile is the absolute path of one entry of the clone's store.
func storeFile(rel string) string {
	return filepath.Join(store, filepath.FromSlash(rel))
}

// storeRel is the project-relative spelling of a store entry, the shape the
// messages use.
func storeRel(rel string) string {
	return relative(storeFile(rel))
}

// TestExecReadsTheAnswerFromTheVerdictFile: a command that names {verdict} writes
// its answer there, and odachi reads that file instead of stdout. A reviewer that
// prints a banner, a spinner or its own logs keeps its answer readable, which is
// the whole reason the token exists.
func TestExecReadsTheAnswerFromTheVerdictFile(t *testing.T) {
	p := newProject()
	answer := `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}],"verdict":"approved"}`
	// stdout is everything an agent CLI prints around the answer.
	p.shell.output = "> thinking...\n> tool call\n"
	p.shell.after = func() {
		path := verdictPathIn(t, p.shell.command)
		p.files.files[path] = []byte(answer)
	}

	result := p.run(review.Options{Exec: "agent --write {verdict} --model {model}", Model: "m/x"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (message %q, output %q)", result.Code, result.Message, result.Output)
	}
	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatal("the answer was not stored as a verdict")
	}
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored answer does not parse: %v", err)
	}
	if parsed.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want approved", parsed.Verdict)
	}
	if _, ok := p.files.files[storeFile(domainreview.AnswerDir+"/"+currentTree+".json")]; ok {
		t.Error("the answer file was left in the store after it was read")
	}
}

// TestExecRefusesAVerdictFileLeftEmpty: naming {verdict} is a promise to write
// there. An empty file is not an empty verdict, and stdout is not a fallback: a
// reviewer that answered somewhere else answered nowhere.
func TestExecRefusesAVerdictFileLeftEmpty(t *testing.T) {
	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "agent {verdict}"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want blocked (message %q)", result.Code, result.Message)
	}
	for _, want := range []string{"{verdict}", domainreview.AnswerDir + "/" + currentTree + ".json"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("the message does not name %q: %q", want, result.Message)
		}
	}
	if _, ok := p.files.files[storePath(currentTree)]; ok {
		t.Error("an empty answer was stored as a verdict")
	}
}

// TestTheAnswerPathIsQuotedForTheShell: the app substitutes the path, so it asks
// the shell how to spell it. {prompt} is quoted inside the adapter for the same
// reason: a clone under a directory with a space is an ordinary Windows profile,
// and an unquoted redirect target ends at the space.
func TestTheAnswerPathIsQuotedForTheShell(t *testing.T) {
	p := newProject()
	p.shell.quote = func(path string) string { return "<" + path + ">" }
	p.shell.after = func() {
		path := verdictPathIn(t, p.shell.command)
		p.files.files[path] = []byte(`{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}],"verdict":"approved"}`)
	}

	result := p.run(review.Options{Exec: "agent --write {verdict}"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(p.shell.command, "<") || !strings.Contains(p.shell.command, ".json>") {
		t.Errorf("the command carries the path unquoted: %q", p.shell.command)
	}
}

// TestEachGroupGetsAFreshAnswerFile: a command that names {verdict} runs once per
// model group, and the file is emptied before each one. A second group that writes
// nothing must fail, not inherit the first group's answer: that is what the
// empty-before-run write is for.
func TestEachGroupGetsAFreshAnswerFile(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	answers := 0
	p.shell.after = func() {
		answers++
		if answers > 1 {
			return // the second group writes nothing
		}
		path := verdictPathIn(t, p.shell.command)
		p.files.files[path] = []byte(`{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"risk"}],"verdict":"approved"}`)
	}

	result := p.run(review.Options{
		Exec:       "agent --write {verdict} --model {model}",
		Lens:       "risk,reliability",
		Because:    "the test splits the run by model",
		LensModels: map[string]string{"reliability": "sharp/model"},
	})

	if result.Code == review.CodeOK {
		t.Fatalf("code = 0: the second group inherited the first answer (message %q)", result.Message)
	}
	if !strings.Contains(result.Message, "wrote nothing") {
		t.Errorf("the message does not name the empty answer: %q", result.Message)
	}
}

// verdictPathIn is the path a command names where the reviewer writes its answer.
func verdictPathIn(t *testing.T, command string) string {
	t.Helper()
	for _, word := range strings.Fields(command) {
		word = strings.Trim(word, `"'<>`)
		if strings.Contains(word, domainreview.AnswerDir) && strings.HasSuffix(word, ".json") {
			return word
		}
	}
	t.Fatalf("the command names no {verdict} path: %q", command)
	return ""
}

// TestPromptCarriesTheCandidate: with no flags the verb prints the review prompt,
// built from the brief and the diff, and stores nothing.
func TestPromptCarriesTheCandidate(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"reservar turnos", "-viejo", "+nuevo", "a.go", currentTree, "When in doubt, say nothing"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not contain %q", want)
		}
	}
	if result.Verdict != nil {
		t.Errorf("printing the prompt stored a verdict: %+v", result.Verdict)
	}
	if slices.Contains(p.files.written(), relative(storePath(currentTree))) {
		t.Error("printing the prompt wrote a verdict file")
	}
}

// TestPromptHasAHome: the printed form also writes the prompt into the clone's
// store, beside the verdict and under the same tree, so an agent that needs it on
// disk hands the reviewer that path instead of inventing one - 0081 invented
// three of them inside proof/, where proof means the output of check --run.
func TestPromptHasAHome(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	rel := domainreview.Dir + "/" + currentTree + ".prompt.md"
	content, ok := p.files.files[storeFile(rel)]
	if !ok {
		t.Fatalf("the prompt was not written to %s", storeRel(rel))
	}
	if string(content) != result.Prompt {
		t.Error("the file and the printed prompt are not the same bytes")
	}
	if !slices.Contains(result.Files, storeRel(rel)) {
		t.Errorf("Files = %v, want it to name %s", result.Files, storeRel(rel))
	}
	if _, ok := p.files.files[storePath(currentTree)]; ok {
		t.Error("printing the prompt wrote a verdict")
	}
}

// TestPromptWritesOnlyThePrompt: the store lives outside the worktree, so the
// printed form has nothing to prepare and nothing to ignore. The one write is the
// prompt itself, filed under the tree it names: a verdict file here would be an
// approval nobody gave.
func TestPromptWritesOnlyThePrompt(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{})

	want := storeRel(domainreview.Dir + "/" + currentTree + ".prompt.md")
	if written := p.files.written(); len(written) != 1 || written[0] != want {
		t.Fatalf("writes = %v, want only %s", written, want)
	}
	if len(result.Files) != 1 || result.Files[0] != want {
		t.Errorf("result.Files = %v, want [%s]", result.Files, want)
	}
}

// TestNothingChangedMeansNoReview: without a diff there is no candidate, so there
// is no prompt, and the run writes nothing at all.
func TestNothingChangedMeansNoReview(t *testing.T) {
	p := newProject()
	p.git.changed = nil

	result := p.run(review.Options{})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if result.Message != "there is nothing to review in this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
	if result.Prompt != "" {
		t.Errorf("a prompt was printed without a diff: %q", result.Prompt)
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none", writes)
	}
}

// TestAttachStoresTheVerdictUnderTheTree: one file per tree, named by the hash,
// carrying what the tool stamped and what the author said.
func TestAttachStoresTheVerdictUnderTheTree(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .git/odachi/reviews/%s.json", currentTree)
	}

	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Tree != currentTree {
		t.Errorf("tree = %q, want %q", parsed.Tree, currentTree)
	}
	if parsed.Base != "HEAD" {
		t.Errorf("base = %q, want HEAD", parsed.Base)
	}
	if parsed.Reviewer != "humano" {
		t.Errorf("reviewer = %q, want the author's", parsed.Reviewer)
	}
	if parsed.At != "2026-10-21T20:31:00Z" {
		t.Errorf("at = %q, want the stamped instant", parsed.At)
	}
	if parsed.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want approved", parsed.Verdict)
	}
}

// TestARefusedAnswerThatCannotBeKeptIsReported: the promise of this feature is
// that a refused answer is kept and the refusal names it. When the store cannot
// be read or written, the refusal must say that instead of naming a file nobody
// wrote - an attach command pointing at nothing is the second payment this
// feature exists to prevent.
func TestARefusedAnswerThatCannotBeKeptIsReported(t *testing.T) {
	answer := []byte(`{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`)

	t.Run("the store cannot be checked", func(t *testing.T) {
		p := newProject()
		p.git.blobs = map[string]string{"a.go": gitBlob}
		p.files.files[fullPath("veredicto.json")] = answer
		p.git.ignored = true
		p.files.failExists = func(path string) error {
			if strings.Contains(path, ".refused-") {
				return errors.New("permiso denegado")
			}
			return nil
		}

		result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

		assertKeepFailure(t, result)
	})

	t.Run("the store cannot be written", func(t *testing.T) {
		p := newProject()
		p.git.blobs = map[string]string{"a.go": gitBlob}
		p.files.files[fullPath("veredicto.json")] = answer
		p.git.ignored = true
		p.files.writeErr = errors.New("disco lleno")

		result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

		assertKeepFailure(t, result)
	})
}

// assertKeepFailure pins the two halves of the refusal when keeping fails: it is
// blocked, it says so, and it never names a file that was not written.
func assertKeepFailure(t *testing.T, result review.Result) {
	t.Helper()

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "could not keep it") {
		t.Errorf("message = %q, want the failure to keep the answer", result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if strings.Contains(lines, "odachi review --attach") {
		t.Errorf("the refusal names an attach command for a file it did not write:\n%s", lines)
	}
	if strings.Contains(lines, "the answer is kept in") {
		t.Errorf("the refusal claims the answer is kept when it is not:\n%s", lines)
	}
}

// TestARefusedAnswerIsKeptOnceAndDroppedWhenOneLands: the name is the content, so
// the same answer twice is one file; two different answers for one tree coexist,
// because each is evidence of what the reviewer said. The answer that lands is the
// explicit act that ends them: the sealed verdict is the record, and the leftovers
// in the ignored store go away.
func TestARefusedAnswerIsKeptOnceAndDroppedWhenOneLands(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	// Two different refusals, then the same one again.
	for _, answer := range []string{
		`{"reviewer":"uno","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`,
		`{"reviewer":"otro","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`,
		`{"reviewer":"uno","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`,
	} {
		p.files.files[fullPath("veredicto.json")] = []byte(answer)
		if result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"}); result.Code != review.CodeBlocked {
			t.Fatalf("answer %q: code = %d, want 1 (%s)", answer, result.Code, result.Message)
		}
	}
	if kept := keptAnswers(t, p); len(kept) != 2 {
		t.Fatalf("kept answers = %d (%v), want two: one per distinct answer", len(kept), kept)
	}

	// The answer that covers the row lands, and the refused ones go.
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[{"path":"a.go","lens":"reliability"}]}`)
	if result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"}); result.Code != review.CodeOK {
		t.Fatalf("the covering answer was refused: %s", result.Message)
	}
	if kept := keptAnswers(t, p); len(kept) != 0 {
		t.Errorf("kept answers = %v, want none after one landed", kept)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Error("the covering answer was not sealed")
	}
}

// TestAVerdictStillLandsWhenTheRefusedAnswersCannotBeDropped: dropping the
// refused answers is best effort by design -the verdict is already written, and
// failing a stored review for litter would be worse than the litter-. A store that
// cannot be listed at that moment must not cost the verdict the run just paid for,
// and the litter stays where it is.
func TestAVerdictStillLandsWhenTheRefusedAnswersCannotBeDropped(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	// An answer an earlier run refused, and a listing that answers the read of the
	// potent rows and fails the drop that comes after the write.
	keptRel := domainreview.Dir + "/" + currentTree + ".refused-deadbeef.json"
	p.files.files[storeFile(keptRel)] = []byte(`{"reviewer":"antes","findings":[],"passes":[]}`)
	calls := 0
	p.files.failList = func(string) error {
		calls++
		if calls > 1 {
			return errors.New("store unavailable")
		}
		return nil
	}

	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[{"path":"a.go","lens":"reliability"}]}`)
	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("the verdict was lost to the drop: %d (%s)", result.Code, result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Error("no verdict was sealed")
	}
	if _, ok := p.files.files[storeFile(keptRel)]; !ok {
		t.Error("the refused answer was dropped even though the listing failed")
	}
	if calls != 2 {
		t.Errorf("the store was listed %d time(s), want the potent rows and the drop", calls)
	}
}

// TestTheRefusalAsksTheReviewerForTheRows: the rows a reviewer did not write are
// the reviewer's to write. A refusal that says "fix the rows" teaches whoever
// holds the file to invent them - an agent did exactly that, twice, rather than
// paying for the missing rows again.
func TestTheRefusalAsksTheReviewerForTheRows(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "ask the reviewer") {
		t.Errorf("the refusal does not send the gap back to the reviewer:\n%s", lines)
	}
	if !strings.Contains(lines, "not edited") {
		t.Errorf("the refusal does not say the answer is not edited to pass:\n%s", lines)
	}
	if strings.Contains(lines, "fix the rows") {
		t.Errorf("the refusal still teaches the caller to fill the rows:\n%s", lines)
	}
	if !strings.Contains(lines, "rows missing: a.go (reliability)") {
		t.Errorf("the refusal lost the rows it names:\n%s", lines)
	}
	if !strings.Contains(lines, "next: odachi review --attach") {
		t.Errorf("the refusal lost the command that attaches it again:\n%s", lines)
	}
}

// TestARefusedAnswerIsKeptAndNamed: coverage is a promise of the command, not a
// discovery of the gate. The answer that would have left a hole is not sealed, and
// it is not thrown away either: the reviewer was paid for it, so it is kept under
// a name its own bytes decide and the refusal names the file. The reviewer writes
// the rows it left out; the caller attaches the fuller answer, it does not fill it.
func TestARefusedAnswerIsKeptAndNamed(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for an answer that answered nothing (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "a.go") || !strings.Contains(result.Message, "reliability") {
		t.Errorf("message = %q, want the file and the lens that never answered", result.Message)
	}
	if !strings.Contains(result.Message, "no verdict was written") {
		t.Errorf("message = %q, want the promise that nothing was sealed", result.Message)
	}

	// The answer is kept: one file, named with this tree and a digest of these
	// bytes, in the store the project already ignores.
	kept := keptAnswers(t, p)
	if len(kept) != 1 {
		t.Fatalf("kept answers = %v, want exactly one", kept)
	}
	rel := relative(kept[0])
	if !strings.HasPrefix(rel, storeRel(domainreview.Dir)+"/"+currentTree+".refused-") || !strings.HasSuffix(rel, ".json") {
		t.Errorf("kept file %q is not named by the tree and the answer", rel)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, rel) {
		t.Errorf("the refusal does not name the kept answer %q:\n%s", rel, lines)
	}
	if !strings.Contains(lines, "odachi review --attach "+rel) {
		t.Errorf("the refusal does not say how to attach it again:\n%s", lines)
	}
	if _, sealed := p.files.files[storePath(currentTree)]; sealed {
		t.Error("a verdict was sealed for an answer that never covered the tree")
	}

	stored, err := domainreview.Parse(p.files.files[fullPath(rel)])
	if err != nil {
		t.Fatalf("the kept answer does not parse: %v", err)
	}
	if stored.Tree != currentTree || len(stored.Findings) != 0 {
		t.Errorf("kept answer = %+v, want the answer as it arrived, under this tree", stored)
	}
}

// TestAnUnreadableAnswerIsKeptAndNamed: a reviewer that answers with a verdict the
// tool cannot parse -an ANSI footer after the JSON is the case that started this-
// still answered, and the human paid for those bytes. The run seals nothing for the
// tree, keeps the answer exactly as it arrived under a name its bytes decide, and
// names the file and the command that attaches it again.
func TestAnUnreadableAnswerIsKeptAndNamed(t *testing.T) {
	answers := map[string]string{
		"prose":                    "I think it is quite good.",
		"JSON with an ANSI footer": `{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[],"passes":[{"path":"a.go","lens":"reliability"}]}` + "\x1b[0m\n",
	}

	for name, answer := range answers {
		t.Run(name, func(t *testing.T) {
			p := newProject()
			p.git.blobs = map[string]string{"a.go": gitBlob}
			p.shell.output = answer

			result := p.run(review.Options{Exec: "revisor"})

			if result.Code != review.CodeBlocked {
				t.Fatalf("code = %d, want 1 for an answer that cannot be read (%s)", result.Code, result.Message)
			}
			if _, sealed := p.files.files[storePath(currentTree)]; sealed {
				t.Error("a verdict was sealed for an answer the tool cannot read")
			}

			kept := keptAnswers(t, p)
			if len(kept) != 1 {
				t.Fatalf("kept answers = %v, want exactly one", kept)
			}
			rel := relative(kept[0])
			if !strings.HasPrefix(rel, storeRel(domainreview.Dir)+"/"+currentTree+".refused-") || !strings.HasSuffix(rel, ".json") {
				t.Errorf("kept file %q is not named by the tree and the answer", rel)
			}
			if got := string(p.files.files[fullPath(rel)]); got != answer {
				t.Errorf("kept bytes = %q, want the answer as it arrived %q", got, answer)
			}

			lines := strings.Join(result.Lines, "\n")
			if !strings.Contains(lines, "the answer is kept in "+rel) {
				t.Errorf("the blockage does not name the kept answer %q:\n%s", rel, lines)
			}
			if !strings.Contains(lines, "odachi review --attach "+rel) {
				t.Errorf("the blockage does not say how to attach it again:\n%s", lines)
			}
			if strings.Contains(lines, "I wrote nothing") {
				t.Errorf("the blockage claims nothing was written while it kept the answer:\n%s", lines)
			}
			if !strings.Contains(result.Output, answer) {
				t.Errorf("the output does not carry the reviewer's answer: %q", result.Output)
			}
		})
	}
}

// TestAnUnreadableRefutationIsKeptAndNamed: a refuter's answer is an answer too.
// When it lands as a refutation the tool cannot parse, the bytes are kept the same
// way, because the refuter was paid for them just as the reviewer was.
func TestAnUnreadableRefutationIsKeptAndNamed(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"refutations":"not a list"}`

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for a refutation that cannot be read (%s)", result.Code, result.Message)
	}
	kept := keptAnswers(t, p)
	if len(kept) != 1 {
		t.Fatalf("kept answers = %v, want exactly one", kept)
	}
	rel := relative(kept[0])
	if got := string(p.files.files[fullPath(rel)]); got != p.shell.output {
		t.Errorf("kept bytes = %q, want the answer as it arrived %q", got, p.shell.output)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "odachi review --attach "+rel) {
		t.Errorf("the blockage does not say how to attach it again:\n%s", lines)
	}
}

// TestAnAnswerThatFailsValidationIsKeptAndNamed: parse and validation are two
// different gates, and both can refuse an answer the reviewer was paid for. The
// declared verdict that contradicts its own findings is the case: the bytes are
// kept, because repairing them is cheaper than another call.
func TestAnAnswerThatFailsValidationIsKeptAndNamed(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.shell.output = `{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"verdict":"approved","findings":[{"path":"a.go","line":1,"severity":"BLOCKER","causal":true,"lens":"reliability","why":"pierde el turno"}]}`

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for an answer that contradicts itself (%s)", result.Code, result.Message)
	}
	kept := keptAnswers(t, p)
	if len(kept) != 1 {
		t.Fatalf("kept answers = %v, want exactly one", kept)
	}
	rel := relative(kept[0])
	if got := string(p.files.files[fullPath(rel)]); got != p.shell.output {
		t.Errorf("kept bytes = %q, want the answer as it arrived %q", got, p.shell.output)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "odachi review --attach "+rel) {
		t.Errorf("the blockage does not say how to attach it again:\n%s", lines)
	}
}

// TestAnUnreadableAnswerThatCannotBeKeptSaysSo: keeping is a promise, and a
// promise the store cannot keep must not name a file nobody wrote - an attach
// command pointing at nothing is the second payment this feature exists to prevent.
func TestAnUnreadableAnswerThatCannotBeKeptSaysSo(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.git.ignored = true
	p.shell.output = `{"reviewer":"agente","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}` + "\x1b[0m"
	p.files.failExists = func(path string) error {
		if strings.Contains(path, ".refused-") {
			return errors.New("permiso denegado")
		}
		return nil
	}

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "could not keep it") {
		t.Errorf("message = %q, want the failure to keep the answer", result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if strings.Contains(lines, "odachi review --attach") {
		t.Errorf("the blockage names an attach command for a file it did not write:\n%s", lines)
	}
	if strings.Contains(lines, "the answer is kept in") {
		t.Errorf("the blockage claims the answer is kept when it is not:\n%s", lines)
	}
}

// TestAnUnreadableAnswerIsKeptOnceAndDroppedWhenOneLands: the name is the content,
// so the same unreadable answer twice is one file; two different ones coexist for
// one tree. The answer that lands is the explicit act that ends them. One rule for
// every answer odachi would not seal is why they share the refused namespace.
func TestAnUnreadableAnswerIsKeptOnceAndDroppedWhenOneLands(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	for _, answer := range []string{
		`{"reviewer":"uno","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}` + "\x1b[0m",
		`{"reviewer":"otro","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}` + "\x1b[0m",
		`{"reviewer":"uno","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}` + "\x1b[0m",
	} {
		p.shell.output = answer
		if result := p.run(review.Options{Exec: "revisor"}); result.Code != review.CodeBlocked {
			t.Fatalf("answer %q: code = %d, want 1 (%s)", answer, result.Code, result.Message)
		}
	}
	if kept := keptAnswers(t, p); len(kept) != 2 {
		t.Fatalf("kept answers = %d (%v), want two: one per distinct answer", len(kept), kept)
	}

	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`
	if result := p.run(review.Options{Exec: "revisor"}); result.Code != review.CodeOK {
		t.Fatalf("the readable answer was refused: %s", result.Message)
	}
	if kept := keptAnswers(t, p); len(kept) != 0 {
		t.Errorf("kept answers = %v, want none after one landed", kept)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Error("the readable answer was not sealed")
	}
}

// keptAnswers is every refused answer the store holds, in path spelling.
func keptAnswers(t *testing.T, p *project) []string {
	t.Helper()
	var kept []string
	for path := range p.files.files {
		if strings.Contains(filepath.ToSlash(path), ".refused-") {
			kept = append(kept, path)
		}
	}
	slices.Sort(kept)
	return kept
}

// gitBlob is what the git port answers for a.go: the blob id of the bytes as
// they stand in the worktree, computed from the content, not from anyone's word.
const gitBlob = "c1a0f9e8d7b6a5c4d3e2f1a0b9c8d7e6f5a4b3c2"

// TestTheVerdictStampsTheBlobGitComputes: a finding's content receipt is derived
// from the repository. The reviewer may write a blob, and it is overwritten — a
// hash is a fact about bytes, and facts are measured, not declared.
func TestTheVerdictStampsTheBlobGitComputes(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "agente",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {
      "path": "a.go",
      "line": 1,
      "severity": "WARNING",
      "lens": "reliability",
      "causal": true,
      "why": "defecto de ejemplo",
      "blob": "1111111111111111111111111111111111111111"
    }
  ]
}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .git/odachi/reviews/%s.json", currentTree)
	}
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}

	got := parsed.Findings[0].Blob
	if got != gitBlob {
		t.Errorf("blob = %q, want the port's %q: a verdict must never keep the reviewer's hash", got, gitBlob)
	}
	if parsed.Findings[0].Lens != "reliability" {
		t.Errorf("lens = %q, want the reviewer's attribution kept", parsed.Findings[0].Lens)
	}
}

// TestThePromptCarriesTheCandidateSFloor: the same deterministic reading the
// store applies is handed to the reviewer before any lens runs.
func TestThePromptCarriesTheCandidateSFloor(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"## The floor", "- a.go: added lines 1"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not carry the floor: missing %q", want)
		}
	}
}

// TestTheStoreDowngradesTheUncitedFinding: a CRITICAL finding that cannot cite
// the candidate's own writing loses one tier before it is stored — blocking is
// earned, and the tool applies the rule, not the reviewer's confidence.
func TestTheStoreDowngradesTheUncitedFinding(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "agente",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {
      "path": "a.go",
      "line": 50,
      "severity": "CRITICAL",
      "causal": true,
      "why": "no cita la escritura del candidato"
    }
  ]
}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if got := parsed.Findings[0].Severity; got != domainreview.Warning {
		t.Errorf("severity = %q, want WARNING: the finding cannot cite the floor", got)
	}
	if parsed.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want approved: without the block nothing requests changes", parsed.Verdict)
	}
	if !slices.ContainsFunc(result.Lines, func(line string) bool {
		return strings.Contains(line, "earned blocking")
	}) {
		t.Errorf("lines %v do not report the downgrade", result.Lines)
	}
}

// concurrentDiff is a real-shaped diff whose added line spawns concurrency:
// the content the path cannot see, and the reason this feature exists.
const concurrentDiff = "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1,2 @@\n-viejo\n+nuevo\n+go func() { sync.Mutex{} }()\n"

// TestTheSelectionReadsTheDiffContent: ordinary code whose added lines spawn a
// goroutine asks for the four lenses, not just the path class's one.
func TestTheSelectionReadsTheDiffContent(t *testing.T) {
	p := newProject()
	p.git.diff = concurrentDiff

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	want := []string{"risk", "readability", "reliability", "resilience"}
	if !slices.Equal(result.Lenses, want) {
		t.Errorf("lenses = %v, want the four: the added goroutine raises the class", result.Lenses)
	}
}

// TestTheStoredVerdictKeepsTheRaisedClass: the verdict records the class it was
// asked under, computed from the same assessment the prompt used.
func TestTheStoredVerdictKeepsTheRaisedClass(t *testing.T) {
	p := newProject()
	p.git.diff = concurrentDiff
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	if result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Lenses) != 4 {
		t.Errorf("lenses = %v, want the four: the stored verdict answers the raised class", parsed.Lenses)
	}
}

// TestAPathGitCannotHashCarriesNoBlob: a file that is not in the worktree
// (deleted by the candidate, or invented by the reviewer) gets no receipt. The
// stamping is not a git failure, so the verdict is still stored — and the gate
// later has no implicit coverage to pretend.
func TestAPathGitCannotHashCarriesNoBlob(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "agente",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {
      "path": "deleted.go",
      "line": 0,
      "severity": "WARNING",
      "causal": true,
      "why": "defecto de ejemplo"
    }
  ]
}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := p.files.files[storePath(currentTree)]
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if got := parsed.Findings[0].Blob; got != "" {
		t.Errorf("blob = %q, want none: a file that is not here is nobody's content receipt", got)
	}
}

// potentStore is a project whose store already holds one verdict for a previous
// tree, with a potent finding and a potent pass about a.go. The bytes a.go has
// now are the blob the git port answers, so both rows describe content this
// candidate does not change.
func potentStore(t *testing.T) *project {
	t.Helper()
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	prev := domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.Approved,
		Findings: []domainreview.Finding{
			{Path: "a.go", Line: 3, Severity: domainreview.Warning, Causal: true, Why: "pierde el turno", Blob: gitBlob, Lens: "reliability"},
		},
		Passes: []domainreview.Pass{
			{Path: "a.go", Blob: gitBlob, Lens: "readability"},
		},
	}
	sealed, err := domainreview.Seal(prev, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the previous verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)
	return p
}

// TestTheStoreLendsItsPotentRows: findings that stand on unchanged bytes and
// passes about unchanged bytes travel into the new verdict. The next candidate
// re-reviews only what changed; it never re-litigates rows whose receipts still
// describe the exact bytes on disk.
func TestTheStoreLendsItsPotentRows(t *testing.T) {
	p := potentStore(t)
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "claude -p"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .git/odachi/reviews/%s.json", currentTree)
	}
	parsed, err := domainreview.Parse(stored)
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 1 || parsed.Findings[0].Line != 3 || parsed.Findings[0].Blob != gitBlob {
		t.Errorf("findings = %+v, want the potent one carried with its receipt", parsed.Findings)
	}
	if len(parsed.Passes) != 1 || parsed.Passes[0].Lens != "readability" || parsed.Passes[0].Blob != gitBlob {
		t.Errorf("passes = %+v, want the potent clean row carried", parsed.Passes)
	}
	if !slices.ContainsFunc(result.Lines, func(line string) bool {
		return strings.Contains(line, "carried from the store")
	}) {
		t.Errorf("lines %v do not say what the verdict carried", result.Lines)
	}
}

// TestAVerdictCarriesThePathsOfItsCandidate: the store lends this candidate the
// rows it can act on. A clean row about a path the diff never touches is evidence
// for a review that already happened, and a warning about one is information
// nobody can act on now; carrying them makes every verdict the size of the store
// -769 rows where 15 mattered, measured on 0132-.
func TestAVerdictCarriesThePathsOfItsCandidate(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob, "otro.go": gitBlob}
	prev := domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.Approved,
		Passes: []domainreview.Pass{
			{Path: "a.go", Blob: gitBlob, Lens: "reliability"},
			{Path: "otro.go", Blob: gitBlob, Lens: "readability"},
		},
		Findings: []domainreview.Finding{
			{Path: "otro.go", Line: 9, Severity: domainreview.Warning, Causal: true, Why: "un aviso", Blob: gitBlob, Lens: "reliability"},
		},
	}
	sealed, err := domainreview.Seal(prev, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the previous verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)

	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`
	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	for _, pass := range stored.Passes {
		if pass.Path == "otro.go" {
			t.Errorf("the verdict carries a clean row about otro.go, a path this candidate never touched")
		}
	}
	if !slices.ContainsFunc(stored.Passes, func(pass domainreview.Pass) bool { return pass.Path == "a.go" }) {
		t.Errorf("passes = %+v, want the row about the path the candidate judged", stored.Passes)
	}
	for _, finding := range stored.Findings {
		if finding.Path == "otro.go" {
			t.Errorf("the verdict carries a warning about otro.go, a path this candidate never touched: %+v", finding)
		}
	}
}

// TestABlockerTravelsWhereverItStands: a causal BLOCKER about a file this
// candidate never touches still travels. It is not a row about bytes somebody
// read; it is the reason not to ship, and dropping it would make the gate forget
// a defect nobody fixed.
func TestABlockerTravelsWhereverItStands(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob, "otro.go": gitBlob}
	prev := domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.ChangesRequested,
		Passes: []domainreview.Pass{
			{Path: "a.go", Blob: gitBlob, Lens: "reliability"},
		},
		Findings: []domainreview.Finding{
			{Path: "otro.go", Line: 4, Severity: domainreview.Critical, Causal: true, Why: "pierde el turno", Blob: gitBlob, Lens: "reliability"},
		},
	}
	sealed, err := domainreview.Seal(prev, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the previous verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)

	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`
	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if !slices.ContainsFunc(stored.Findings, func(finding domainreview.Finding) bool {
		return finding.Path == "otro.go" && finding.Severity == domainreview.Critical
	}) {
		t.Errorf("findings = %+v, want the standing blocker carried", stored.Findings)
	}
}

// TestThreeReviewsInARowLeaveVerdictsThatDoNotGrow: the bound is what turns the
// store's growth from the square of the reviews into the size of each candidate.
// The store starts with a verdict about two hundred files nobody will touch, and
// three candidates change one file each: every verdict stays the size of its own
// rows, and the gate answered every changed path -a run that stored is a run the
// coverage check let through.
func TestThreeReviewsInARowLeaveVerdictsThatDoNotGrow(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}

	seed := domainreview.Verdict{Reviewer: "claude", Verdict: domainreview.Approved}
	for i := 0; i < 200; i++ {
		path := fmt.Sprintf("viejo/%03d.go", i)
		p.git.blobs[path] = gitBlob
		seed.Passes = append(seed.Passes, domainreview.Pass{Path: path, Blob: gitBlob, Lens: "readability"})
	}
	sealed, err := domainreview.Seal(seed, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the seed verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)
	if len(p.files.files[storePath(otherTree)]) < 20*1024 {
		t.Fatalf("the seed verdict is %d bytes; the test needs a store that would have grown the old copy", len(p.files.files[storePath(otherTree)]))
	}

	p.shell.output = `{"findings":[],"passes":[{"path":"a.go","lens":"reliability"}],"inspection":{"status":"completed","paths":["a.go"]}}`

	sizes := []int{}
	trees := []string{strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)}
	blobs := []string{strings.Repeat("1", 40), strings.Repeat("2", 40), strings.Repeat("3", 40)}
	for i, tree := range trees {
		p.git.tree = tree
		p.git.blobs["a.go"] = blobs[i]
		if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
			t.Fatalf("review %d: code = %d, want 0 (%s)", i+1, result.Code, result.Message)
		}
		stored, ok := p.files.files[storePath(tree)]
		if !ok {
			t.Fatalf("review %d stored no verdict for %s", i+1, tree)
		}
		sizes = append(sizes, len(stored))

		parsed, err := domainreview.Parse(stored)
		if err != nil {
			t.Fatalf("verdict %d does not parse: %v", i+1, err)
		}
		for _, pass := range parsed.Passes {
			if strings.HasPrefix(pass.Path, "viejo/") {
				t.Fatalf("verdict %d carries %s, a path its candidate never touched", i+1, pass.Path)
			}
		}
	}

	for i, size := range sizes {
		if size >= 4096 {
			t.Errorf("verdict %d is %d bytes; the bound keeps it the size of its own rows, not the store's", i+1, size)
		}
	}
	if sizes[2] > sizes[0] {
		t.Errorf("verdict sizes %v: the third grew over the first, so the store still copies itself forward", sizes)
	}
}

// refutedProject is a candidate whose previous verdict filed one blocking
// finding, and a refuter killed it with evidence.
func refutedProject(t *testing.T) *project {
	t.Helper()
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	blocking := domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.ChangesRequested,
		Findings: []domainreview.Finding{
			{Path: "a.go", Line: 3, Severity: domainreview.Critical, Causal: true, Why: "pierde el turno", Blob: gitBlob, Lens: "reliability"},
		},
	}
	sealed, err := domainreview.Seal(blocking, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the previous verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)

	answer := domainreview.Refutation{
		Refuter:     "refuter",
		Refutations: []domainreview.Judgment{{ID: "F1", Verdict: domainreview.Refuted, Why: "la transaccion guarda el turno, la cita lo prueba"}},
	}
	sealedRefutation, err := domainreview.SealRefutation(answer, blocking.Findings, otherTree, now)
	if err != nil {
		t.Fatalf("sealing the refutation: %v", err)
	}
	refRel, err := domainreview.RefutationPath(otherTree)
	if err != nil {
		t.Fatalf("the refutation path: %v", err)
	}
	p.files.files[storeFile(refRel)] = domainreview.RenderRefutation(sealedRefutation)
	return p
}

// TestARefutedFindingTravelsDismissed: a finding the refuter killed with
// evidence does not die with its verdict. It travels as a dismissal — reason
// and fingerprint included — so the next candidate can defend its silence
// instead of re-litigating it.
func TestARefutedFindingTravelsDismissed(t *testing.T) {
	p := refutedProject(t)
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "claude -p"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 1 {
		t.Fatalf("findings = %d, want 1: the dismissal travels", len(parsed.Findings))
	}
	dismissed := parsed.Findings[0]
	if !domainreview.IsDismissed(dismissed) {
		t.Errorf("status = %q, want dismissed", dismissed.Status)
	}
	if dismissed.Reason != "la transaccion guarda el turno, la cita lo prueba" {
		t.Errorf("reason = %q, want the refuter's evidence", dismissed.Reason)
	}
	if dismissed.Fingerprint == "" {
		t.Error("the dismissal travels without its fingerprint")
	}
	if parsed.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want approved: a dismissal never blocks", parsed.Verdict)
	}
}

// TestAReRaisedDismissalStaysDismissed: the reviewer may re-report a defect the
// store already dismissed. The tool keeps the dismissal — the reason and
// fingerprint win over the fresh claim, deterministically.
func TestAReRaisedDismissalStaysDismissed(t *testing.T) {
	p := refutedProject(t)
	p.shell.output = `{"inspection":{"status":"completed","paths":["a.go"]},"findings":[{"path":"a.go","line":3,"severity":"CRITICAL","lens":"reliability","causal":true,"why":"pierde el turno"}]}`

	result := p.run(review.Options{Exec: "claude -p"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 1 {
		t.Fatalf("findings = %d, want one row: the re-raise is the same defect", len(parsed.Findings))
	}
	if !domainreview.IsDismissed(parsed.Findings[0]) {
		t.Errorf("status = %q, want dismissed: a re-raise never undoes a judgment", parsed.Findings[0].Status)
	}
	if !strings.Contains(parsed.Findings[0].Reason, "transaccion") {
		t.Errorf("reason = %q, want the original evidence kept", parsed.Findings[0].Reason)
	}
}

// TestThePromptNamesTheDismissals: the reviewer is told what was judged and why
// before it reports anything, or it will re-raise a defect nobody wants again.
func TestThePromptNamesTheDismissals(t *testing.T) {
	p := refutedProject(t)

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"dismissed", "la transaccion guarda el turno", "re-raise"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not name the dismissal: missing %q", want)
		}
	}
}

// TestShowNamesADismissal: the stored verdict says what it discarded and why,
// and --show prints it without making the human open the JSON.
func TestShowNamesADismissal(t *testing.T) {
	p := newProject()
	dismissed := domainreview.Finding{
		Path: "a.go", Line: 3, Severity: domainreview.Warning, Causal: true,
		Why: "pierde el turno", Lens: "reliability",
		Status: domainreview.Dismissed, Reason: "la transaccion lo cubre",
	}
	dismissed.Fingerprint = domainreview.Fingerprint(dismissed)
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.Approved,
		Lenses:   []string{"reliability"},
		Findings: []domainreview.Finding{dismissed},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the stored verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(sealed)

	result := p.run(review.Options{Show: true})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "dismissed: la transaccion lo cubre") {
		t.Errorf("lines %v do not print the dismissal's reason", result.Lines)
	}
}

// TestStaleBytesDoNotTravel: a receipt that describes other bytes is dead by
// definition, so the next verdict stays free of it.
func TestStaleBytesDoNotTravel(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	stale := domainreview.Pass{
		Path: "a.go", Blob: "7d3c2b1a9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b", Lens: "reliability",
	}
	prev := domainreview.Verdict{
		Reviewer: "claude",
		Verdict:  domainreview.Approved,
		Findings: []domainreview.Finding{
			{Path: "a.go", Line: 3, Severity: domainreview.Warning, Causal: true, Why: "pierde el turno", Blob: "7d3c2b1a9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b", Lens: "reliability"},
		},
		Passes: []domainreview.Pass{stale},
	}
	sealed, err := domainreview.Seal(prev, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the previous verdict: %v", err)
	}
	p.files.files[storePath(otherTree)] = domainreview.Render(sealed)
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]},"passes":[{"path":"a.go","lens":"reliability"}]}`

	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 0 {
		t.Errorf("findings = %d, want none: stale receipts never travel", len(parsed.Findings))
	}
	for _, pass := range parsed.Passes {
		if pass.Blob == stale.Blob {
			t.Errorf("passes = %v, want no row about the old bytes", parsed.Passes)
		}
	}
	// The run answered the row itself, about the bytes that stand now: the fresh
	// pass is the run's answer, not the stale one carried over.
	if len(parsed.Passes) != 1 || parsed.Passes[0].Blob != gitBlob {
		t.Errorf("passes = %v, want the run's own answer about the current bytes", parsed.Passes)
	}
}

// TestThePromptNamesWhatTheStoreCarries: a re-reviewer must know which rows are
// already held, or it files the same defect twice. The printed prompt carries
// them; a store with nothing potent prints the same prompt as always.
func TestThePromptNamesWhatTheStoreCarries(t *testing.T) {
	p := potentStore(t)

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"already knows about these bytes", "The store keeps more of the clean rows", "a.go:3", "Do not re-report"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not name what is carried: missing %q", want)
		}
	}

	fresh := newProject().run(review.Options{})
	if strings.Contains(fresh.Prompt, "already knows about these bytes") {
		t.Errorf("an empty store must not name carried rows: %s", fresh.Prompt)
	}
}

// TestARepeatedFindingIsFiledOnce: the reviewer may re-report a finding the
// store already holds potent; the verdict keeps one copy — the fresh one, right
// after the stamps.
func TestARepeatedFindingIsFiledOnce(t *testing.T) {
	p := potentStore(t)
	p.shell.output = `{"inspection":{"status":"completed","paths":["a.go"]},"findings":[{"path":"a.go","line":3,"severity":"WARNING","lens":"reliability","causal":true,"why":"pierde el turno"}]}`

	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 1 {
		t.Errorf("findings = %d, want 1: the same (path, blob, lens, line, severity) is one defect", len(parsed.Findings))
	}
}

// TestWhenOneLensRunsItOwnsEveryFinding: with exactly one lens answering, a
// finding's author is not in doubt — the reviewer does not have to name what
// only one specialist could have found.
func TestWhenOneLensRunsItOwnsEveryFinding(t *testing.T) {
	p := newProject()
	p.git.blobs = map[string]string{"a.go": gitBlob}
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "agente",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {
      "path": "a.go",
      "line": 1,
      "severity": "WARNING",
      "causal": true,
      "why": "defecto de ejemplo"
    }
  ]
}`)

	if result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if got := parsed.Findings[0].Lens; got != "reliability" {
		t.Errorf("lens = %q, want reliability: one lens ran, and the finding is its answer", got)
	}
}

// TestAHashFailureBlocksTheStoring: when git cannot tell why, the store keeps
// quiet rather than filing a verdict whose receipts are guesses.
func TestAHashFailureBlocksTheStoring(t *testing.T) {
	p := newProject()
	p.git.blobErr = errors.New("fatal: bad object")
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "agente",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {
      "path": "a.go",
      "line": 1,
      "severity": "WARNING",
      "causal": true,
      "why": "defecto de ejemplo"
    }
  ]
}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if writes := p.verdictWrites(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none: a block leaves no file behind", writes)
	}
}

// TestAttachDerivesChangesRequested: a causal BLOCKER is enough. Nobody
// administers the verdict; the finding decides it.
func TestAttachDerivesChangesRequested(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{
  "reviewer": "claude",
  "inspection": {"status": "completed", "paths": ["a.go"]},
  "findings": [
    {"path": "internal/app/check/check.go", "line": 42, "severity": "BLOCKER", "causal": true, "why": "un fallo de lectura se informa como listo"}
  ]
}`)

	result := p.run(review.Options{Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Verdict != domainreview.ChangesRequested {
		t.Fatalf("verdict = %q, want changes_requested", parsed.Verdict)
	}
}

// TestASecondVerdictIsReportedNotApplied: the store is immutable. A second
// attempt says so and leaves the first verdict's bytes untouched.
func TestASecondVerdictIsReportedNotApplied(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("primero.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)
	p.files.files[fullPath("segundo.json")] = []byte(`{"reviewer":"otro","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	first := p.run(review.Options{Attach: "primero.json"})
	if first.Code != review.CodeOK {
		t.Fatalf("the first attach failed: %s", first.Message)
	}
	before := string(p.files.files[storePath(currentTree)])

	second := p.run(review.Options{Attach: "segundo.json"})

	if second.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for a second verdict", second.Code)
	}
	if !strings.Contains(second.Message, "this tree already has a verdict") {
		t.Errorf("message = %q, want %q", second.Message, "this tree already has a verdict")
	}
	if after := string(p.files.files[storePath(currentTree)]); after != before {
		t.Fatalf("the stored verdict was overwritten:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestTheStoreRefusesToReplaceOneAnsweredLens: the file is the answer that was
// paid for. A second verdict that claims an answered lens is refused even when it
// also brings lenses the tree still owes: the store grows by lens through a run,
// never by overwriting what a run already judged.
func TestTheStoreRefusesToReplaceOneAnsweredLens(t *testing.T) {
	p := newProject()
	p.git.changed = []string{"auth/session.go"}
	p.git.diff = "diff --git a/auth/session.go b/auth/session.go\n--- a/auth/session.go\n+++ b/auth/session.go\n@@ -1 +1 @@\n-viejo\n+nuevo\n"
	p.files.files[fullPath("segundo.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	first, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"risk"},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(first)
	before := string(p.files.files[storePath(currentTree)])

	got := p.run(review.Options{Attach: "segundo.json"})

	if got.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 for a verdict that replaces an answered lens (%s)", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, "already has a verdict") {
		t.Errorf("message = %q, want it to say the answered lens is not replaced", got.Message)
	}
	if after := string(p.files.files[storePath(currentTree)]); after != before {
		t.Fatalf("the stored verdict was overwritten:\nbefore: %s\nafter:  %s", before, after)
	}
}

// TestAttachRefusesABrokenVerdictWithoutWriting: an invalid or incomplete JSON is
// one English line and zero writes. A half-read verdict stored under a tree's
// name would lie about what was reviewed.
func TestAttachRefusesABrokenVerdictWithoutWriting(t *testing.T) {
	cases := map[string]string{
		"not JSON":                     `done, approved`,
		"no reviewer":                  `{"findings":[]}`,
		"finding without why":          `{"reviewer":"x","findings":[{"path":"a.go","severity":"WARNING"}]}`,
		"invented severity":            `{"reviewer":"x","findings":[{"path":"a.go","severity":"MAJOR","why":"something"}]}`,
		"verdict that does not follow": `{"reviewer":"x","findings":[{"path":"a.go","severity":"BLOCKER","causal":true,"why":"breaks"}],"verdict":"approved"}`,
	}

	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject()
			p.files.files[fullPath("veredicto.json")] = []byte(content)

			result := p.run(review.Options{Attach: "veredicto.json"})

			if result.Code != review.CodeBlocked {
				t.Fatalf("code = %d, want 1", result.Code)
			}
			if strings.Contains(result.Message, "\n") {
				t.Errorf("message %q is more than one line", result.Message)
			}
			if writes := p.verdictWrites(); len(writes) != 0 {
				t.Fatalf("a failed reviewer wrote %v, want no verdict", writes)
			}
			if _, ok := p.files.files[storePath(currentTree)]; ok {
				t.Fatal("a verdict was stored from an invalid file")
			}
		})
	}
}

// TestAttachRefusesAVerdictForAnotherTree: approval cannot float. The refusal
// names both trees and no verdict is stored. Nothing is written at all: an
// attached document for another tree is refused before the store is touched.
func TestAttachRefusesAVerdictForAnotherTree(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"tree":"` + otherTree + `","reviewer":"humano","findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if !strings.Contains(result.Message, otherTree) || !strings.Contains(result.Message, currentTree) {
		t.Errorf("message = %q, want both trees named", result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; ok {
		t.Fatal("a verdict for another tree was stored under this tree")
	}
	for _, path := range p.files.written() {
		if strings.HasSuffix(path, ".json") {
			t.Fatalf("writes = %v, want no verdict", p.files.written())
		}
	}
}

// TestAttachTakesAnAbsolutePath: a human or an agent can pass the path of the
// verdict file as the shell gave it, and an absolute path must not be resolved
// inside the project — that turns a real file into an unreadable name.
func TestAttachTakesAnAbsolutePath(t *testing.T) {
	// Absolute on this platform: the shape a shell actually hands over.
	shared, err := filepath.Abs(filepath.Join("veredicto.json"))
	if err != nil {
		t.Fatalf("resolving the absolute path: %v", err)
	}

	p := newProject()
	p.files.files[shared] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: shared})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Fatal("an absolute --attach path was not stored")
	}
}

// TestAnotherTreeIsAnotherFile: the same project, one character changed, is a
// different verdict with a different name. That is the whole content-addressing
// argument, asserted on the file name.
func TestAnotherTreeIsAnotherFile(t *testing.T) {
	first := newProject()
	first.files.files[fullPath("v.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)
	if result := first.run(review.Options{Attach: "v.json"}); result.Code != review.CodeOK {
		t.Fatalf("first attach failed: %s", result.Message)
	}

	second := newProject()
	second.git.tree = otherTree
	second.files.files[fullPath("v.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)
	if result := second.run(review.Options{Attach: "v.json"}); result.Code != review.CodeOK {
		t.Fatalf("second attach failed: %s", result.Message)
	}

	if _, ok := first.files.files[storePath(currentTree)]; !ok {
		t.Fatal("the first tree has no verdict")
	}
	if _, ok := second.files.files[storePath(otherTree)]; !ok {
		t.Fatal("the second tree has no verdict under its own name")
	}
	if _, ok := second.files.files[storePath(currentTree)]; ok {
		t.Fatal("the second tree read or wrote the first tree's verdict")
	}
}

// TestShowWithoutAVerdictSaysSo: asking what a tree's review is, when there is
// none, is a blocked answer with the documented sentence, and no write.
func TestShowWithoutAVerdictSaysSo(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if result.Message != "there is no review for this tree" {
		t.Errorf("message = %q, want the documented sentence", result.Message)
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none", writes)
	}
}

// TestShowReadsThisTreesVerdict: what is stored is what is shown, findings
// included.
func TestShowReadsThisTreesVerdict(t *testing.T) {
	p := newProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"claude","inspection":{"status":"completed","paths":["a.go"]},"findings":[
		{"path":"a.go","line":3,"severity":"CRITICAL","causal":true,"why":"pierde el turno"}
	]}`)
	if result := p.run(review.Options{Attach: "veredicto.json"}); result.Code != review.CodeOK {
		t.Fatalf("attach failed: %s", result.Message)
	}

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	shown := strings.Join(result.Lines, "\n")
	for _, want := range []string{domainreview.ChangesRequested, "claude", "a.go:3", "pierde el turno"} {
		if !strings.Contains(shown, want) {
			t.Errorf("the shown verdict does not contain %q (got %q)", want, shown)
		}
	}
}

// TestShowRefusesAVerdictThatCannotBeRead: the store is a plain file a human can
// edit, so a mangled one is reported instead of crashing or being ignored.
func TestShowRefusesAVerdictThatCannotBeRead(t *testing.T) {
	p := newProject()
	p.files.files[storePath(currentTree)] = []byte("basura")

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if !strings.Contains(result.Message, currentTree) {
		t.Errorf("message = %q, want the tree named", result.Message)
	}
}

// TestTheVerdictComesFromStdout: a banner and ANSI on stderr are the CLI's
// decoration, not the answer. The verdict must come from stdout alone.
func TestTheVerdictComesFromStdout(t *testing.T) {
	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`
	p.shell.stderr = "\x1b[31m> reviewing...\x1b[0m\n"

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if _, ok := p.files.files[storePath(currentTree)]; !ok {
		t.Fatal("no verdict was stored: stderr was read as part of the answer")
	}
}

// TestTheSessionModelFillsThePlaceholder: the human writes {model} once; the
// reviewer receives the model this session is using, and the receipt names the
// command that really ran - not the template.
func TestTheSessionModelFillsThePlaceholder(t *testing.T) {
	p := newProject()
	p.host.model = "opencode-go/deepseek-v4.1-flash#max"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	const want = "revisor --model opencode-go/deepseek-v4.1-flash#max"
	if p.shell.command != want {
		t.Errorf("command = %q, want %q", p.shell.command, want)
	}
	if !p.host.asked {
		t.Error("the host was not asked for the session's model")
	}
	if p.host.dir != dir {
		t.Errorf("host asked about %q, want the project %q", p.host.dir, dir)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Reviewer != want {
		t.Errorf("reviewer = %q, want what ran: %q", parsed.Reviewer, want)
	}
	// The model is a field, not something a later gate has to dig out of a
	// command string: the proof carries the builder's, and this is the one that
	// judged.
	if parsed.Model != "opencode-go/deepseek-v4.1-flash#max" {
		t.Errorf("model = %q, want the identity that judged", parsed.Model)
	}
}

// TestTheExplicitModelWinsOverTheSession: --model is the override the human
// asked for, so the host is not even consulted.
func TestTheExplicitModelWinsOverTheSession(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor --model {model}", Model: "human/model"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if want := "revisor --model human/model"; p.shell.command != want {
		t.Errorf("command = %q, want %q", p.shell.command, want)
	}
	if p.host.asked {
		t.Error("the host was asked even though --model named one")
	}
}

// TestACommandWithoutThePlaceholderIsUntouched: the default stays
// byte-identical to before this feature, even when the host knows its model.
func TestACommandWithoutThePlaceholderIsUntouched(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "claude -p"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if p.shell.command != "claude -p" {
		t.Errorf("command = %q, want it untouched", p.shell.command)
	}
}

// TestAModelThatDoesNotResolveStopsBeforeTheReviewer: a placeholder with no
// model is an unknown reviewer, not a reviewer with a default. Starting one
// would file a verdict whose model nobody can name.
func TestAModelThatDoesNotResolveStopsBeforeTheReviewer(t *testing.T) {
	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor --model {model}"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{"{model}", "--model"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("message %q does not name %q", result.Message, want)
		}
	}
	if p.shell.asked {
		t.Error("the reviewer was started without the model it asked for")
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Errorf("writes = %v, want none", writes)
	}
}

// TestAModelFlagWithoutThePlaceholderIsRefused: a flag that changes nothing is
// a lie about what ran, so it is named instead of ignored.
func TestAModelFlagWithoutThePlaceholderIsRefused(t *testing.T) {
	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "claude -p", Model: "human/model"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "{model}") {
		t.Errorf("message %q does not name the placeholder", result.Message)
	}
	if p.shell.asked {
		t.Error("the reviewer ran with a model that had nowhere to go")
	}
}

// TestThePromptReadsTheWorktreeOnce: the prompt, the lens selection and the
// tree it names come from one read. A project that ran init is the normal case:
// the store is already ignored, so nothing is written and nothing is re-read.
func TestThePromptReadsTheWorktreeOnce(t *testing.T) {
	p := newProject()
	p.git.ignored = true

	result := p.run(review.Options{})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if p.git.candidateCalls != 1 {
		t.Errorf("candidate reads = %d, want 1: the prompt asks the worktree once", p.git.candidateCalls)
	}
}

// TestTheExecReadsTheCandidateOnce: the answer belongs to the candidate the
// prompt named, so the run does not read the worktree again after the reviewer.
// The tree is held, not re-measured; whether it moved is a cheaper question, and
// T2's.
func TestTheExecReadsTheCandidateOnce(t *testing.T) {
	p := newProject()
	p.git.ignored = true
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if p.git.candidateCalls != 1 {
		t.Errorf("candidate reads = %d, want 1: the prompt and the file are the same candidate", p.git.candidateCalls)
	}
}

// TestAModelWithoutExecIsRefused: --model outside --exec has nothing to
// replace, so it is named instead of accepted and silently dropped.
func TestAModelWithoutExecIsRefused(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{Show: true, Model: "human/model"})
	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "--exec") {
		t.Errorf("message %q does not name the form that takes it", result.Message)
	}
}

// TestAFailingReviewerShowsBothStreams: when the reviewer fails, the human needs
// the whole picture — the answer it started and the error it wrote.
func TestAFailingReviewerShowsBothStreams(t *testing.T) {
	p := newProject()
	p.shell.output = "the answer half"
	p.shell.stderr = "the diagnosis half"
	p.shell.exitCode = 2

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code == review.CodeOK {
		t.Fatal("a failed reviewer must not read as a stored verdict")
	}
	for _, want := range []string{"the answer half", "the diagnosis half"} {
		if !strings.Contains(result.Output, want) {
			t.Errorf("output %q does not show %q", result.Output, want)
		}
	}
}

// TestExecStoresWhatTheCommandAnswered: --exec runs the reviewer and stores its
// answer, stamped with the command as the reviewer when the answer omits one.
func TestExecStoresWhatTheCommandAnswered(t *testing.T) {
	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	result := p.run(review.Options{Exec: "claude -p"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !p.shell.asked || p.shell.command != "claude -p" {
		t.Errorf("the shell was asked %v with %q", p.shell.asked, p.shell.command)
	}
	if p.shell.timeout != review.ExecTimeoutSeconds {
		t.Errorf("timeout = %d, want %d", p.shell.timeout, review.ExecTimeoutSeconds)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if parsed.Reviewer != "claude -p" {
		t.Errorf("reviewer = %q, want the command", parsed.Reviewer)
	}
	if parsed.Verdict != domainreview.Approved {
		t.Errorf("verdict = %q, want approved", parsed.Verdict)
	}
}

// TestExecGivesTheReviewerThePrompt: --exec starts a reviewer, so the command has
// to receive the same prompt `odachi review` prints. Until this rule it received
// nothing, and the only way to give it the prompt was a POSIX-only
// `agent "$(odachi review)"` that computed the diff twice.
func TestExecGivesTheReviewerThePrompt(t *testing.T) {
	printed := newProject().run(review.Options{})
	if printed.Prompt == "" {
		t.Fatalf("the printed form has no prompt, so comparing against it proves nothing")
	}

	p := newProject()
	p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`

	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	if p.shell.input != printed.Prompt {
		t.Errorf("the reviewer was given a different prompt than `odachi review` prints")
		t.Logf("what the reviewer got: %s", p.shell.input)
		t.Logf("what the verb prints: %s", printed.Prompt)
	}
}

// TestExecFailureWritesNothing: a reviewer that fails is not a review. Exit 1, no
// file, and the command's own output shown. An answer that arrived and cannot be
// read is not a failed reviewer: it is kept and named, which
// TestAnUnreadableAnswerIsKeptAndNamed pins on its own.
func TestExecFailureWritesNothing(t *testing.T) {
	cases := map[string]*fakeShell{
		"exit code":    {output: "boom: could not read the brief", exitCode: 1},
		"timeout":      {output: "starting…", timedOut: true},
		"does not run": {err: errors.New("no known shell")},
	}

	for name, shell := range cases {
		t.Run(name, func(t *testing.T) {
			p := newProject()
			p.shell = shell

			result := p.run(review.Options{Exec: "revisor"})

			if result.Code != review.CodeBlocked {
				t.Fatalf("code = %d, want 1", result.Code)
			}
			if writes := p.verdictWrites(); len(writes) != 0 {
				t.Fatalf("a failed reviewer wrote %v, want no verdict", writes)
			}
			if !strings.Contains(result.Output, shell.output) {
				t.Errorf("output %q does not carry the command's %q", result.Output, shell.output)
			}
		})
	}
}

// TestExecWithNothingToReview: --exec needs a candidate too. No diff, no run: the
// command is not even started.
func TestExecWithNothingToReview(t *testing.T) {
	p := newProject()
	p.git.changed = nil

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if p.shell.asked {
		t.Error("the reviewer command ran without a diff")
	}
	if writes := p.files.written(); len(writes) != 0 {
		t.Fatalf("writes = %v, want none", writes)
	}
}

// TestBaseFallsBackToTheEmptyTreeWithoutCommits: a project with no commit has no
// HEAD to compare against, and everything uncommitted is the comparison. That is
// the same rule `check` uses, so the two verbs never disagree.
func TestBaseFallsBackToTheEmptyTreeWithoutCommits(t *testing.T) {
	p := newProject()
	git := &flakyGit{inner: p.git}

	result := review.Run(review.Options{Dir: dir, Store: store, Now: now}, p.files, git, p.shell, &fakeLenses{}, p.host)

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !git.sawEmptyTree {
		t.Fatalf("bases asked = %v, want the empty tree after HEAD failed", git.bases)
	}
}

// flakyGit fails the first comparison against HEAD, which is what a repository
// without a single commit does.
// flakyGit fails the first comparison against HEAD, which is what a repository
// without a single commit does.
type flakyGit struct {
	inner        *fakeGit
	failed       bool
	bases        []string
	sawEmptyTree bool
}

func (g *flakyGit) TreeHash(root string) (string, error) { return g.inner.TreeHash(root) }

func (g *flakyGit) Candidate(root, base string) (string, []string, string, error) {
	g.bases = append(g.bases, base)
	if base == review.DefaultBase && !g.failed {
		g.failed = true
		return "", nil, "", errors.New("fatal: ambiguous argument 'HEAD'")
	}
	if base == review.EmptyTree {
		g.sawEmptyTree = true
	}
	return g.inner.tree, g.inner.changed, g.inner.diff, nil
}

func (g *flakyGit) Correction(root, tree string) ([]string, string, error) {
	return g.inner.Correction(root, tree)
}

func (g *flakyGit) Ignored(root, path string) (bool, error) { return g.inner.Ignored(root, path) }

func (g *flakyGit) BlobHash(root, path string) (string, error) { return g.inner.BlobHash(root, path) }

// TestNothingOutsideTheStoreIsWritten: the whole surface of this verb is
// the store's reviews/. A review that touched the brief or the code would be a
// different tool.
func TestNothingOutsideTheStoreIsWritten(t *testing.T) {
	storePrefix := filepath.ToSlash(store) + "/"

	for _, opts := range []review.Options{
		{},
		{Show: true},
		{Exec: "revisor"},
		{Attach: "veredicto.json"},
	} {
		p := newProject()
		p.shell.output = `{"findings":[],"inspection":{"status":"completed","paths":["a.go"]}}`
		p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

		p.run(opts)

		for _, path := range p.files.writes {
			if !strings.HasPrefix(filepath.ToSlash(path), storePrefix) {
				t.Errorf("opts %+v wrote outside the store: %s", opts, path)
			}
		}
	}
}
