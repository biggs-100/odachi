package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// TestAFailedRunKeepsTheAnswerItPaidFor: a reviewer that writes its answer into
// the {verdict} file and then fails was paid for those bytes. The file is
// scaffolding - the next run on this tree empties it - so the run copies it under
// a name its content decides and names it in the refusal.
func TestAFailedRunKeepsTheAnswerItPaidFor(t *testing.T) {
	p := newProject()
	answer := `{"reviewer":"agente","findings":[]}`
	p.shell.after = func() {
		path := verdictPathIn(t, p.shell.command)
		p.files.files[path] = []byte(answer)
	}
	p.shell.exitCode = 1

	result := p.run(review.Options{Exec: "agent --write {verdict}"})

	if result.Code != review.CodeBlocked {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	kept := keptAnswers(t, p)
	if len(kept) != 1 {
		t.Fatalf("kept answers = %v, want the failed one kept", kept)
	}
	rel := relative(kept[0])
	if !strings.HasPrefix(rel, storeRel(domainreview.Dir)+"/"+currentTree+".refused-") {
		t.Errorf("kept file %q is not named by the tree and the answer", rel)
	}
	if got := string(p.files.files[kept[0]]); !strings.Contains(got, `"reviewer":"agente"`) {
		t.Errorf("kept bytes = %q, want the reviewer's answer", got)
	}
	lines := strings.Join(result.Lines, "\n")
	if !strings.Contains(lines, rel) {
		t.Errorf("the refusal does not name where the answer was kept:\n%s", lines)
	}
}

// TestTheScaffoldingIsNotTheKeptAnswer: the kept name is not the scaffolding name,
// so the next run cannot empty it. Keeping the answer under answers/<tree>.json
// would be keeping it in the one file every run rewrites.
func TestTheScaffoldingIsNotTheKeptAnswer(t *testing.T) {
	p := newProject()
	p.shell.after = func() {
		path := verdictPathIn(t, p.shell.command)
		p.files.files[path] = []byte(`{"reviewer":"agente","findings":[]}`)
	}
	p.shell.exitCode = 1

	p.run(review.Options{Exec: "agent --write {verdict}"})

	scaffolding, err := domainreview.AnswerPath(currentTree)
	if err != nil {
		t.Fatalf("AnswerPath: %v", err)
	}
	if _, kept := p.files.files[storeFile(scaffolding)]; kept {
		t.Fatalf("the answer was kept in the scaffolding file %s", scaffolding)
	}
}
