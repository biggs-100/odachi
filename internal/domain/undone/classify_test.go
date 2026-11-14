package undone_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/domain/undone"
)

// TestABuildFailureThatNamesTheChangeIsNoticed: the row reverts production files
// and keeps the tests, so a change whose own tests call what it added leaves a copy
// that cannot compile. That failure is not silence: the compiler names the symbol
// the revert took away, and a row that names it has evidence, the way the mutation
// row has evidence when the red names the text it removed (0300).
//
// What this does not prove: that the symbol belongs to the change. The name comes
// from the diff's added side, and a file that mentions it for another reason reads
// the same -which is why the exact name and not the shape of the line is what the
// verdict carries-.
func TestABuildFailureThatNamesTheChangeIsNoticed(t *testing.T) {
	const build = "# github.com/biggs-100/odachi/internal/app/tasks [build failed]\n" +
		"internal/app/tasks/tasks_test.go:40:19: undefined: tasks.Totals\n"

	cases := []struct {
		name     string
		exitCode int
		output   string
		added    []string
		want     undone.Verdict
		wantName string
	}{
		{
			name:     "the compiler names what the change added",
			exitCode: 1,
			output:   build,
			added:    []string{"Totals"},
			want:     undone.Noticed,
			wantName: "Totals",
		},
		{
			name:     "a qualified name is the name",
			exitCode: 1,
			output:   "# github.com/biggs-100/odachi/internal/app/start [build failed]\ninternal/app/start/start_test.go:42:66: too many arguments in call to fencetoml.Render\n\tfencetoml.Render(rules, loose)\n",
			added:    []string{"Render", "loose"},
			want:     undone.Noticed,
			wantName: "Render",
		},
		{
			name:     "a build failure that names nothing from the change",
			exitCode: 1,
			output:   build,
			added:    []string{"Render"},
			want:     undone.Unknown,
			wantName: "",
		},
		{
			name:     "a longer word is not the name",
			exitCode: 1,
			output:   "# github.com/biggs-100/odachi/internal/app/start [build failed]\n\tinternal/app/start/start_test.go:42:19: undefined: Renderer\n",
			added:    []string{"Render"},
			want:     undone.Unknown,
			wantName: "",
		},
		{
			name:     "the suite stayed green",
			exitCode: 0,
			output:   "ok  \tgithub.com/biggs-100/odachi/internal/domain/undone\t0.5s\n",
			added:    []string{"Totals"},
			want:     undone.Missed,
			wantName: "",
		},
		{
			name:     "a red for another reason",
			exitCode: 1,
			output:   "--- FAIL: TestSomething (0.00s)\n",
			added:    []string{"Totals"},
			want:     undone.Noticed,
			wantName: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, name := undone.Classify(tc.exitCode, tc.output, tc.added)
			if got != tc.want {
				t.Errorf("Classify = %q, want %q", got, tc.want)
			}
			if name != tc.wantName {
				t.Errorf("the verdict names %q, want %q", name, tc.wantName)
			}
		})
	}
}
