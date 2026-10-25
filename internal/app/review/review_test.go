package review_test

import (
	"errors"
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
// outside .odachi/reviews/ is written" is the property a test has to be able to
// see.
type fakeFiles struct {
	files   map[string][]byte
	writes  []string
	backups []string
	err     error
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

	// input is the prompt the reviewer was given, once --exec delivers one.
	input string
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, string, int, bool, error) {
	s.asked = true
	s.command = command
	s.timeout = timeoutSeconds
	s.commands = append(s.commands, command)
	return s.output, s.stderr, s.exitCode, s.timedOut, s.err
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
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
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
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{
		Exec:       "revisor --model {model}",
		Lens:       "reliability,risk",
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

// TestASecondRunCompletesTheVerdict: the store grows by lens. A run that finds a
// verdict covering reliability judges risk and adds it, keeping the rows it did
// not pay for again.
func TestASecondRunCompletesTheVerdict(t *testing.T) {
	p := newProject()
	p.host.model = "host/default"
	p.shell.output = `{"findings":[],"passes":[{"path":"a.go","lens":"risk"}]}`

	first, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "first",
		Lenses:   []string{"reliability"},
		Passes:   []domainreview.Pass{{Path: "a.go", Blob: "b", Lens: "reliability"}},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("seeding the verdict: %v", err)
	}
	p.files.files[storePath(currentTree)] = domainreview.Render(first)

	result := p.run(review.Options{Exec: "revisor --model {model}", Lens: "risk"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
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

// run performs one `odachi review` with the fixed clock.
func (p *project) run(opts review.Options) review.Result {
	opts.Dir = dir
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
		claims = append(claims, written)
	}
	return claims
}

// storePath is where a verdict for one tree lives, in the spelling the app uses.
func storePath(tree string) string {
	return filepath.Join(dir, filepath.FromSlash(domainreview.Dir), tree+".json")
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

// TestPromptKeepsTheStoreOutOfTheTree: the verdict lives inside the candidate it
// describes, so the store's directory is ignored BEFORE the tree is hashed.
// Without that order, storing a verdict changes the tree that names it.
func TestPromptKeepsTheStoreOutOfTheTree(t *testing.T) {
	p := newProject()

	result := p.run(review.Options{})

	want := domainreview.Dir + "/" + domainreview.IgnoreName
	written := p.files.written()
	if !slices.Contains(written, want) {
		t.Fatalf("writes = %v, want the store's %s", written, want)
	}
	if len(written) != 1 {
		t.Fatalf("writes = %v, want exactly the ignore file", written)
	}
	if slices.Contains(result.Files, want) == false {
		t.Errorf("result.Files = %v, want it to report %s", result.Files, want)
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
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","findings":[]}`)

	result := p.run(review.Options{Base: "HEAD", Attach: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .odachi/reviews/%s.json", currentTree)
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
		t.Fatalf("no verdict stored at .odachi/reviews/%s.json", currentTree)
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
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","findings":[]}`)

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
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{Exec: "claude -p"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored, ok := p.files.files[storePath(currentTree)]
	if !ok {
		t.Fatalf("no verdict stored at .odachi/reviews/%s.json", currentTree)
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
	p.files.files[fullPath(refRel)] = domainreview.RenderRefutation(sealedRefutation)
	return p
}

// TestARefutedFindingTravelsDismissed: a finding the refuter killed with
// evidence does not die with its verdict. It travels as a dismissal — reason
// and fingerprint included — so the next candidate can defend its silence
// instead of re-litigating it.
func TestARefutedFindingTravelsDismissed(t *testing.T) {
	p := refutedProject(t)
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[{"path":"a.go","line":3,"severity":"CRITICAL","lens":"reliability","causal":true,"why":"pierde el turno"}]}`

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
	p.shell.output = `{"findings":[]}`

	if result := p.run(review.Options{Exec: "claude -p"}); result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	parsed, err := domainreview.Parse(p.files.files[storePath(currentTree)])
	if err != nil {
		t.Fatalf("the stored verdict does not parse: %v", err)
	}
	if len(parsed.Findings) != 0 || len(parsed.Passes) != 0 {
		t.Errorf("findings = %d, passes = %d, want none: stale receipts never travel", len(parsed.Findings), len(parsed.Passes))
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
	for _, want := range []string{"the store already carries", "a.go:3", "Do not re-report"} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the prompt does not name what is carried: missing %q", want)
		}
	}

	fresh := newProject().run(review.Options{})
	if strings.Contains(fresh.Prompt, "the store already carries") {
		t.Errorf("an empty store must not name carried rows: %s", fresh.Prompt)
	}
}

// TestARepeatedFindingIsFiledOnce: the reviewer may re-report a finding the
// store already holds potent; the verdict keeps one copy — the fresh one, right
// after the stamps.
func TestARepeatedFindingIsFiledOnce(t *testing.T) {
	p := potentStore(t)
	p.shell.output = `{"findings":[{"path":"a.go","line":3,"severity":"WARNING","lens":"reliability","causal":true,"why":"pierde el turno"}]}`

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
	p.files.files[fullPath("primero.json")] = []byte(`{"reviewer":"humano","findings":[]}`)
	p.files.files[fullPath("segundo.json")] = []byte(`{"reviewer":"otro","findings":[]}`)

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
// names both trees and no verdict is stored.
//
// The store's ignore file may be created by this attempt: the verdict is checked
// against the tree hash, and the hash can only be taken once the store is out of
// the tree. That file belongs to the tool, is idempotent, and is inside
// .odachi/reviews/; the verdict is not written.
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
	p.files.files[shared] = []byte(`{"reviewer":"humano","findings":[]}`)

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
	first.files.files[fullPath("v.json")] = []byte(`{"reviewer":"humano","findings":[]}`)
	if result := first.run(review.Options{Attach: "v.json"}); result.Code != review.CodeOK {
		t.Fatalf("first attach failed: %s", result.Message)
	}

	second := newProject()
	second.git.tree = otherTree
	second.files.files[fullPath("v.json")] = []byte(`{"reviewer":"humano","findings":[]}`)
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
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"claude","findings":[
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
	p.shell.output = `{"findings":[]}`
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
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[]}`

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

// TestTheExecReadsAgainAfterTheReviewer: the reviewer is arbitrary code. The
// verdict must describe the tree it left, not the one the prompt named.
func TestTheExecReadsAgainAfterTheReviewer(t *testing.T) {
	p := newProject()
	p.git.ignored = true
	p.shell.output = `{"findings":[]}`

	result := p.run(review.Options{Exec: "revisor"})
	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if p.git.candidateCalls != 2 {
		t.Errorf("candidate reads = %d, want 2: one for the prompt, one after the reviewer", p.git.candidateCalls)
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
	p.shell.output = `{"findings":[]}`

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
	p.shell.output = `{"findings":[]}`

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
// file, and the command's own output shown.
func TestExecFailureWritesNothing(t *testing.T) {
	cases := map[string]*fakeShell{
		"exit code":    {output: "boom: could not read the brief", exitCode: 1},
		"timeout":      {output: "starting…", timedOut: true},
		"does not run": {err: errors.New("no known shell")},
		"not JSON":     {output: "I think it is quite good."},
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

	result := review.Run(review.Options{Dir: dir, Now: now}, p.files, git, p.shell, &fakeLenses{}, p.host)

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

func (g *flakyGit) Ignored(root, path string) (bool, error) { return g.inner.Ignored(root, path) }

func (g *flakyGit) BlobHash(root, path string) (string, error) { return g.inner.BlobHash(root, path) }

// TestNothingOutsideTheStoreIsWritten: the whole surface of this verb is
// .odachi/reviews/. A review that touched the brief or the code would be a
// different tool.
func TestNothingOutsideTheStoreIsWritten(t *testing.T) {
	store := filepath.ToSlash(filepath.Join(dir, filepath.FromSlash(domainreview.Dir))) + "/"

	for _, opts := range []review.Options{
		{},
		{Show: true},
		{Exec: "revisor"},
		{Attach: "veredicto.json"},
	} {
		p := newProject()
		p.shell.output = `{"findings":[]}`
		p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","findings":[]}`)

		p.run(opts)

		for _, path := range p.files.writes {
			if !strings.HasPrefix(filepath.ToSlash(path), store) {
				t.Errorf("opts %+v wrote outside the store: %s", opts, path)
			}
		}
	}
}
