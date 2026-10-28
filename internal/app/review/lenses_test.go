package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// fakeLenses is the lens port in memory. It records what was asked, because
// "which lens did this run load" is the property the selection exists for.
type fakeLenses struct {
	asked []string
	err   error
}

func (l *fakeLenses) Lens(name string) (string, error) {
	l.asked = append(l.asked, name)
	if l.err != nil {
		return "", l.err
	}
	return "---\nname: " + name + "\ndescription: Trigger: test.\n---\n\n" +
		"Severity floor: WARNING and SUGGESTION never block.\n\n" +
		"## What to look at\n\n- actionable " + name + " rule\n\n" +
		"## What NOT to report\n\n- style\n", nil
}

// newLensProject is a project whose change is only code: focused, one lens.
func newLensProject() *project {
	p := newProject()
	p.lenses = &fakeLenses{}
	return p
}

// TestNoFlagsUsesTheLensesTheEvidenceAsksFor: the bare verb is unchanged except
// for this — a code change is reviewed with the reliability lens, and the human
// does not have to know that.
func TestNoFlagsUsesTheLensesTheEvidenceAsksFor(t *testing.T) {
	p := newLensProject()

	result := p.run(review.Options{})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if strings.Join(p.lenses.asked, ",") != domainreview.Reliability {
		t.Fatalf("lenses loaded = %v, want only reliability", p.lenses.asked)
	}
	if !strings.Contains(result.Prompt, "actionable reliability rule") {
		t.Errorf("the prompt does not carry the reliability rules:\n%s", result.Prompt)
	}
	if strings.Join(result.Lenses, ",") != domainreview.Reliability {
		t.Errorf("result lenses = %v, want [reliability]", result.Lenses)
	}
}

// TestDocumentationOnlyAsksForNoLens: a docs-only change is passive. The verb
// prints the same prompt it printed before lenses existed.
func TestDocumentationOnlyAsksForNoLens(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"docs/guia.md", "README.md"}

	result := p.run(review.Options{})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(p.lenses.asked) != 0 {
		t.Fatalf("lenses loaded = %v, want none: documentation asks for no lens", p.lenses.asked)
	}
	if strings.Contains(result.Prompt, "## The lenses") {
		t.Errorf("a passive candidate got a lens section:\n%s", result.Prompt)
	}
	if len(result.Lenses) != 0 {
		t.Errorf("result lenses = %v, want none", result.Lenses)
	}
}

// TestAnAuthChangeAsksForTheFour: the sensitive signals are the difference
// between one lens and four.
func TestAnAuthChangeAsksForTheFour(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"internal/gastos/x.go", "internal/auth/session.go"}

	result := p.run(review.Options{})

	want := domainreview.Lenses(domainreview.Full)
	if strings.Join(result.Lenses, ",") != strings.Join(want, ",") {
		t.Fatalf("result lenses = %v, want %v", result.Lenses, want)
	}
	for _, lens := range want {
		if !strings.Contains(result.Prompt, "actionable "+lens+" rule") {
			t.Errorf("the prompt does not carry the %s lens", lens)
		}
	}
}

// TestLensFlagAddsAndNeverRemoves: --lens risk on a code change keeps
// reliability, because the evidence asked for it first.
func TestLensFlagAddsAndNeverRemoves(t *testing.T) {
	p := newLensProject()

	result := p.run(review.Options{Lens: "risk"})

	want := "risk,reliability"
	if strings.Join(result.Lenses, ",") != want {
		t.Fatalf("result lenses = %v, want %s", result.Lenses, want)
	}
	if !strings.Contains(result.Prompt, "actionable risk rule") {
		t.Error("the prompt does not carry the risk rules")
	}
}

// TestLensFlagAll: `--lens all` is the four, in the fixed order, even for a
// change the evidence would have called passive.
func TestLensFlagAll(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"docs/guia.md"}

	result := p.run(review.Options{Lens: "all"})

	want := domainreview.Lenses(domainreview.Full)
	if strings.Join(result.Lenses, ",") != strings.Join(want, ",") {
		t.Fatalf("result lenses = %v, want %v", result.Lenses, want)
	}
}

// TestAnInventedLensIsRefusedBeforeAnyWrite: a typo must not print a prompt that
// reviewed nothing, and it must not leave a file behind.
func TestAnInventedLensIsRefusedBeforeAnyWrite(t *testing.T) {
	p := newLensProject()

	result := p.run(review.Options{Lens: "seguridad"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if !strings.Contains(result.Message, domainreview.Risk) {
		t.Errorf("message %q does not name the valid lenses", result.Message)
	}
	if len(p.files.writes) != 0 {
		t.Fatalf("writes = %v, want none", p.files.written())
	}
}

// TestAnUnloadableLensIsABlock: the selection and the embedded content are two
// lists, so the verb has to say which one failed instead of printing half a
// prompt.
func TestAnUnloadableLensIsABlock(t *testing.T) {
	p := newLensProject()
	p.lenses.err = errors.New("could not read the risk lens")

	result := p.run(review.Options{Lens: "risk"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "risk") {
		t.Errorf("message %q does not name the lens", result.Message)
	}
}

// TestShowSaysWhichLensesAreMissing: --show answers the question the gate will
// ask, so the human sees it before the gate does.
func TestShowSaysWhichLensesAreMissing(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"internal/auth/session.go"}
	p.files.files[storePath(currentTree)] = []byte(storedVerdict(t, "claude", []string{domainreview.Risk}))

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"risk", "readability", "reliability", "resilience", "missing lenses"} {
		if !strings.Contains(lines, want) {
			t.Errorf("--show does not say %q:\n%s", want, lines)
		}
	}
}

// TestShowWithNoVerdictSaysWhyTheTreeAsks: the gate is invisible until it runs,
// and --show is where a human decides whether the review pays before paying.
func TestShowWithNoVerdictSaysWhyTheTreeAsks(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"internal/app/x.go"}
	p.git.diff = ""

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d: without a verdict --show is blocked (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"required:", "reliability", "ordinary code", "next: odachi review"} {
		if !strings.Contains(lines, want) {
			t.Errorf("--show does not say %q:\n%s", want, lines)
		}
	}
}

// TestShowWithNoVerdictAndDocumentsOnlySaysNoneIsDue: a passive tree pays for
// no lens, and --show says that instead of sending the human to a review that
// cannot find anything.
func TestShowWithNoVerdictAndDocumentsOnlySaysNoneIsDue(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"README.md"}
	p.git.diff = ""

	result := p.run(review.Options{Show: true})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want %d: none is due, and --show still answers the verdict question (%s)", result.Code, review.CodeBlocked, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	for _, want := range []string{"none is due", "documents only"} {
		if !strings.Contains(lines, want) {
			t.Errorf("--show does not say %q:\n%s", want, lines)
		}
	}
}

// TestShowWithAnUnreadCandidateDoesNotGuessLenses: when the base cannot be
// compared, --show still finds a verdict by tree, but it does not invent lenses
// it could not measure.
func TestShowWithAnUnreadCandidateDoesNotGuessLenses(t *testing.T) {
	p := newLensProject()
	p.git.candidateErr = errors.New("unknown revision")

	result := p.run(review.Options{Show: true, Base: "nosuchref"})

	lines := strings.Join(result.Lines, "\n")
	if strings.Contains(lines, "required:") || strings.Contains(lines, "none is due") {
		t.Errorf("--show guessed the lenses of a candidate it could not read:\n%s", lines)
	}
	if !strings.Contains(lines, "there is no review for this tree") {
		t.Errorf("--show did not answer the verdict question:\n%s", lines)
	}
}

// TestShowWithLensAsksAboutThatLens: `--lens --show` is the documented spelling
// for "what does this tree need and what is already answered".
func TestShowWithLensAsksAboutThatLens(t *testing.T) {
	p := newLensProject()
	p.git.changed = []string{"internal/gastos/x.go"}
	p.files.files[storePath(currentTree)] = []byte(storedVerdict(t, "claude", nil))

	result := p.run(review.Options{Show: true, Lens: "risk"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, "risk") {
		t.Errorf("--show with --lens risk does not mention risk:\n%s", lines)
	}
	if !strings.Contains(lines, "missing lenses") {
		t.Errorf("--show does not list the missing lenses:\n%s", lines)
	}
}

// TestStoredVerdictRecordsTheLensesTheToolAsked: the reviewer never declares its
// own lenses; the tool stamps what the candidate required, exactly like the tree
// and the instant.
func TestStoredVerdictRecordsTheLensesTheToolAsked(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(`{"reviewer":"humano","inspection":{"status":"completed","paths":["a.go"]},"findings":[]}`)

	result := p.run(review.Options{Attach: "veredicto.json", Lens: "risk"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(p.files.files[storePath(currentTree)])
	if !strings.Contains(stored, `"risk"`) || !strings.Contains(stored, `"reliability"`) {
		t.Fatalf("the stored verdict does not record the lenses it covered:\n%s", stored)
	}
}

// blockedVerdict is a verdict with one causal blocker, rendered for the store.
func blockedVerdict(t *testing.T) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Lenses:   []string{domainreview.Risk},
		Findings: []domainreview.Finding{{
			Path:     "internal/auth/session.go",
			Line:     12,
			Severity: domainreview.Blocker,
			Causal:   true,
			Why:      "la sesión se acepta sin verificar la firma",
		}},
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the blocked verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}

// TestRefutePrintsTheRefuterPrompt: --refute takes the verdict file and prints
// the prompt that asks the three questions about its blockers.
func TestRefutePrintsTheRefuterPrompt(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(blockedVerdict(t))

	result := p.run(review.Options{Refute: "veredicto.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{
		"Is it causal of the diff, or was it already there?",
		"F1",
		"internal/auth/session.go:12",
		"+nuevo",
		"When in doubt",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Errorf("the refuter prompt does not contain %q\n%s", want, result.Prompt)
		}
	}
	if len(p.files.writes) != 0 {
		t.Fatalf("writes = %v, want none: printing a prompt writes nothing", p.files.written())
	}
}

// TestRefuteWithoutBlockersSaysThereIsNothingToRefute: a verdict that approves
// has nothing to refute, and inventing work is worse than saying so.
func TestRefuteWithoutBlockersSaysThereIsNothingToRefute(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("veredicto.json")] = []byte(storedVerdict(t, "claude", []string{domainreview.Reliability}))

	result := p.run(review.Options{Refute: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	if !strings.Contains(result.Message, "nothing to refute") {
		t.Errorf("message %q does not say there is nothing to refute", result.Message)
	}
	if len(p.files.writes) != 0 {
		t.Fatalf("writes = %v, want none", p.files.written())
	}
}

// TestRefuteOnlyWorksOnThisTree: the evidence is the diff in front of you. A
// verdict from another tree cannot be refuted with it.
func TestRefuteOnlyWorksOnThisTree(t *testing.T) {
	p := newLensProject()
	other, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: "claude",
		Findings: []domainreview.Finding{{Path: "a.go", Severity: domainreview.Blocker, Causal: true, Why: "pierde el turno"}},
	}, otherTree, "HEAD", now)
	if err != nil {
		t.Fatalf("Seal = %v", err)
	}
	p.files.files[fullPath("veredicto.json")] = domainreview.Render(other)

	result := p.run(review.Options{Refute: "veredicto.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, otherTree) {
		t.Errorf("message %q does not name the other tree", result.Message)
	}
}

// TestAttachStoresARefutationBesideTheVerdict: --attach stores either document,
// and the refutation is named after the same tree, once.
func TestAttachStoresARefutationBesideTheVerdict(t *testing.T) {
	p := newLensProject()
	p.files.files[storePath(currentTree)] = []byte(blockedVerdict(t))
	p.files.files[fullPath("refutacion.json")] = []byte(`{"refuter":"humano","refutations":[{"id":"F1","verdict":"refuted","why":"el diff +nuevo muestra que la firma ya se validaba en la línea 12"}]}`)

	result := p.run(review.Options{Attach: "refutacion.json"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	refutedPath := storeFile(domainreview.Dir + "/" + currentTree + ".refuted.json")
	stored, ok := p.files.files[refutedPath]
	if !ok {
		t.Fatalf("the refutation was not stored; writes = %v", p.files.written())
	}
	if !strings.Contains(string(stored), domainreview.Refuted) {
		t.Fatalf("the stored refutation does not carry the verdict:\n%s", stored)
	}
	if result.Refutation == nil {
		t.Error("the result does not carry the refutation it stored")
	}

	second := p.run(review.Options{Attach: "refutacion.json"})
	if second.Code != review.CodeBlocked {
		t.Fatalf("second attach code = %d, want 1: the refutation is written once", second.Code)
	}
	if !strings.Contains(second.Message, "already has a refutation") {
		t.Errorf("second attach message = %q", second.Message)
	}
}

// TestAttachRefusesARefutationAboutFindingsThatDoNotBlock: the refuter may only
// judge what blocks, and anything else leaves no file behind.
func TestAttachRefusesARefutationAboutFindingsThatDoNotBlock(t *testing.T) {
	p := newLensProject()
	p.files.files[storePath(currentTree)] = []byte(blockedVerdict(t))
	p.files.files[fullPath("refutacion.json")] = []byte(`{"refutations":[{"id":"F7","verdict":"refuted","why":"porque sí"}]}`)

	result := p.run(review.Options{Attach: "refutacion.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "F7") {
		t.Errorf("message %q does not name the offending id", result.Message)
	}
	if len(p.files.written()) != 0 {
		t.Fatalf("writes = %v, want none", p.files.written())
	}
}

// TestAttachRefutationNeedsAVerdict: refuting findings that were never filed is
// a misunderstanding, not a document.
func TestAttachRefutationNeedsAVerdict(t *testing.T) {
	p := newLensProject()
	p.files.files[fullPath("refutacion.json")] = []byte(`{"refutations":[{"id":"F1","verdict":"refuted","why":"porque sí"}]}`)

	result := p.run(review.Options{Attach: "refutacion.json"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "odachi review") {
		t.Errorf("message %q does not name the command that produces a verdict", result.Message)
	}
	if len(p.files.written()) != 0 {
		t.Fatalf("writes = %v, want none", p.files.written())
	}
}

// TestExecStoresARefutationAnswer: --exec stores whichever document the reviewer
// answered, and a refutation is not a verdict.
func TestExecStoresARefutationAnswer(t *testing.T) {
	p := newLensProject()
	p.files.files[storePath(currentTree)] = []byte(blockedVerdict(t))
	p.shell.output = `{"refutations":[{"id":"F1","verdict":"stands","why":"la línea 12 no alcanza para defenderlo"}]}`

	result := p.run(review.Options{Exec: "revisor"})

	if result.Code != review.CodeOK {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Output)
	}
	refutedPath := storeFile(domainreview.Dir + "/" + currentTree + ".refuted.json")
	stored, ok := p.files.files[refutedPath]
	if !ok {
		t.Fatalf("the refutation was not stored; writes = %v", p.files.written())
	}
	if !strings.Contains(string(stored), domainreview.Stands) {
		t.Fatalf("the stored refutation does not carry the answer:\n%s", stored)
	}
	if !strings.Contains(string(stored), "revisor") {
		t.Errorf("the refutation does not record the command that answered:\n%s", stored)
	}
}

// TestShowListsTheRefutation: a refutation is part of what happened to this
// tree, so --show mentions it.
func TestShowListsTheRefutation(t *testing.T) {
	p := newLensProject()
	p.files.files[storePath(currentTree)] = []byte(blockedVerdict(t))
	refutation, err := domainreview.SealRefutation(domainreview.Refutation{
		Refutations: []domainreview.Judgment{{ID: "F1", Verdict: domainreview.Refuted, Why: "el diff +nuevo muestra que ya estaba"}},
	}, []domainreview.Finding{{
		Path: "internal/auth/session.go", Line: 12, Severity: domainreview.Blocker, Causal: true,
		Why: "la sesión se acepta sin verificar la firma",
	}}, currentTree, now)
	if err != nil {
		t.Fatalf("SealRefutation = %v", err)
	}
	refutedPath := storeFile(domainreview.Dir + "/" + currentTree + ".refuted.json")
	p.files.files[refutedPath] = domainreview.RenderRefutation(refutation)

	result := p.run(review.Options{Show: true})

	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, ".refuted.json") {
		t.Errorf("--show does not list the refutation file:\n%s", lines)
	}
}

// storedVerdict is a verdict as it sits in the store: sealed and rendered, with
// the lenses it covered. A verdict with no lenses is the one a project carries
// from before lenses existed.
func storedVerdict(t *testing.T, reviewer string, lenses []string) string {
	t.Helper()
	sealed, err := domainreview.Seal(domainreview.Verdict{
		Reviewer: reviewer,
		Lenses:   lenses,
	}, currentTree, "HEAD", now)
	if err != nil {
		t.Fatalf("sealing the stored verdict: %v", err)
	}
	return string(domainreview.Render(sealed))
}
