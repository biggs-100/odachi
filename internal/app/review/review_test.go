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

	// ignored is what the repository's rules answer about the store, and
	// ignoredErr is the probe failing. A failed probe must never be read as "the
	// store is already protected".
	ignored    bool
	ignoredErr error
	ignoredAs  []string
}

func (g *fakeGit) TreeHash(root string) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.tree, nil
}

func (g *fakeGit) ChangedPaths(root, base string) ([]string, error) {
	g.bases = append(g.bases, base)
	if g.err != nil {
		return nil, g.err
	}
	return g.changed, nil
}

func (g *fakeGit) Diff(root, base string) (string, error) {
	g.bases = append(g.bases, base)
	if g.err != nil {
		return "", g.err
	}
	return g.diff, nil
}

func (g *fakeGit) Ignored(root, path string) (bool, error) {
	g.ignoredAs = append(g.ignoredAs, path)
	if g.ignoredErr != nil {
		return false, g.ignoredErr
	}
	return g.ignored, nil
}

// fakeShell is the Shell port for --exec.
type fakeShell struct {
	output   string
	exitCode int
	timedOut bool
	err      error

	asked   bool
	command string
	timeout int

	// input is the prompt the reviewer was given, once --exec delivers one.
	input string
}

func (s *fakeShell) Run(dir, command string, timeoutSeconds int) (string, int, bool, error) {
	s.asked = true
	s.command = command
	s.timeout = timeoutSeconds
	return s.output, s.exitCode, s.timedOut, s.err
}

func (s *fakeShell) RunInput(dir, command, input string, timeoutSeconds int) (string, int, bool, error) {
	s.input = input
	return s.Run(dir, command, timeoutSeconds)
}

// project is one run's wiring: an in-memory project, and the fakes behind it.
type project struct {
	files  *fakeFiles
	git    *fakeGit
	shell  *fakeShell
	lenses *fakeLenses
}

// newProject seeds a project with a brief and one changed file.
func newProject() *project {
	return &project{
		files: newFakeFiles(map[string]string{"brief.md": brief}),
		git:   &fakeGit{tree: currentTree, changed: []string{"a.go"}, diff: diff},
		shell: &fakeShell{},
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
	return review.Run(opts, p.files, p.git, p.shell, p.lenses)
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

	result := review.Run(review.Options{Dir: dir, Now: now}, p.files, git, p.shell, &fakeLenses{})

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

func (g *flakyGit) ChangedPaths(root, base string) ([]string, error) {
	g.bases = append(g.bases, base)
	if base == review.DefaultBase && !g.failed {
		g.failed = true
		return nil, errors.New("fatal: ambiguous argument 'HEAD'")
	}
	if base == review.EmptyTree {
		g.sawEmptyTree = true
	}
	return g.inner.changed, nil
}

func (g *flakyGit) Diff(root, base string) (string, error) { return g.inner.Diff(root, base) }

func (g *flakyGit) Ignored(root, path string) (bool, error) { return g.inner.Ignored(root, path) }

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
