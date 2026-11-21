package undone_test

import (
	"slices"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/undone"
)

// TestAddedNamesReadsWhatTheDiffDeclares: the row that reverts a change needs the
// names the change added, because the compiler is what says them when the copy does
// not build. Only the added side counts: a name the diff removed is not something a
// reverted copy can be missing.
func TestAddedNamesReadsWhatTheDiffDeclares(t *testing.T) {
	diff := "" +
		"diff --git a/internal/app/tasks/tasks.go b/internal/app/tasks/tasks.go\n" +
		"--- a/internal/app/tasks/tasks.go\n" +
		"+++ b/internal/app/tasks/tasks.go\n" +
		"@@ -230,3 +230,9 @@ func Totals(docs []Document) string {\n" +
		"-func Render(rules []fence.Rule, loose []string) string {\n" +
		"+func Render(rules []fence.Rule) string {\n" +
		"+func (r *runner) nextSession() domaindoctor.Check {\n" +
		"+type Evidence struct {\n" +
		"+const buildFailed = \"[build failed]\"\n" +
		"+var seen map[string]bool\n" +
		"-func Removed(a int) {}\n" +
		"+// a comment names nobody\n" +
		"+\treturn out.String()\n"

	cases := []struct {
		name string
		diff string
		want []string
	}{
		{
			name: "declarations of the added side",
			diff: diff,
			want: []string{"Render", "nextSession", "Evidence", "buildFailed", "seen"},
		},
		{
			name: "no added line declares anything",
			diff: "+++ b/x.go\n+\tfmt.Println(\"hola\")\n+\n",
			want: nil,
		},
		{
			name: "a name is taken once",
			diff: "+++ b/x.go\n+func Same() {}\n+func Same() {}\n",
			want: []string{"Same"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := undone.AddedNames(c.diff); !slices.Equal(got, c.want) {
				t.Errorf("AddedNames = %v, want %v", got, c.want)
			}
		})
	}
}

// TestAFieldAddedToAStructIsNamedByTheUndo: reverting a field leaves the tests
// that read it unable to build, and the compiler names the field. The row can
// only answer noticed if the diff says the field was added. Git writes the
// enclosing declaration in the hunk header, so that is the scope the row reads;
// a Name Type line anywhere else is a guess the row must not make (0341).
func TestAFieldAddedToAStructIsNamedByTheUndo(t *testing.T) {
	const structDiff = "diff --git a/internal/domain/gate/gate.go b/internal/domain/gate/gate.go\n" +
		"--- a/internal/domain/gate/gate.go\n" +
		"+++ b/internal/domain/gate/gate.go\n" +
		"@@ -194,6 +201,13 @@ type State struct {\n" +
		" \t// VerificationLines is the authored count (0339).\n" +
		" \tVerificationLines int\n" +
		"+\t// SliceAccepted is how many authored lines the human accepts (0340).\n" +
		"+\tSliceAccepted int\n" +
		"+\tSliceReason   string\n" +
		"+\tthing = value\n" +
		"+\tio.Reader\n" +
		" \t// VerificationRefuted holds the notes (0300).\n"

	cases := []struct {
		name string
		diff string
		want []string
	}{
		{
			name: "the fields a struct hunk added",
			diff: structDiff,
			want: []string{"SliceAccepted", "SliceReason"},
		},
		{
			name: "the same line outside a struct hunk names nobody",
			diff: "+++ b/x.go\n@@ -1,2 +1,3 @@ func Totals(docs []Document) string {\n+SliceAccepted int\n",
			want: nil,
		},
		{
			name: "a new struct names its fields too",
			diff: "+++ b/x.go\n@@ -0,0 +1,3 @@\n+type State struct {\n+\tSliceAccepted int\n+}\n",
			want: []string{"State", "SliceAccepted"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := undone.AddedNames(c.diff); !slices.Equal(got, c.want) {
				t.Errorf("AddedNames = %v, want %v", got, c.want)
			}
		})
	}
}
