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
