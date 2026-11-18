package skilleval_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// write and run are the two shapes of an event the score cares about, named once
// so each case reads as the story its order tells.
func write(tool, path string) skilleval.Event { return skilleval.Event{Tool: tool, Path: path} }
func run(command string) skilleval.Event      { return skilleval.Event{Tool: "bash", Command: command} }

// TestASuiteBetweenTheTestAndTheCodeMeetsBoth: the run the exam is asking for -
// test first, red, then the fix - is the run that scores two.
func TestASuiteBetweenTheTestAndTheCodeMeetsBoth(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", `C:\work\config_test.go`),
		run("go test ./..."),
		write("write", `C:\work\config.go`),
	})

	if !got[0].Met || !got[1].Met {
		t.Fatalf("score = %+v, want both behaviours met", got)
	}
	if !strings.Contains(got[1].Why, "go test") {
		t.Errorf("the second behaviour does not name the command that decided it: %q", got[1].Why)
	}
}

// TestACodeFileFirstFailsBoth: a run that writes the implementation and then a
// test is a run where no red could have been seen, and both behaviours say so.
func TestACodeFileFirstFailsBoth(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", "config.go"),
		write("write", "config_test.go"),
		run("go test ./..."),
	})

	if got[0].Met {
		t.Errorf("the test counted as written before the code: %+v", got[0])
	}
	if got[1].Met {
		t.Errorf("the red counted for a test that came second: %+v", got[1])
	}
}

// TestATestWithNoSuiteRunFailsTheRed: the order can be right and the red never
// observed, which is the pair the exam exists to tell apart.
func TestATestWithNoSuiteRunFailsTheRed(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", "config_test.go"),
		write("write", "config.go"),
		run("go test ./..."),
	})

	if !got[0].Met {
		t.Errorf("the test came first and the score says it did not: %+v", got[0])
	}
	if got[1].Met {
		t.Errorf("a suite that ran after the code counted as the red: %+v", got[1])
	}
}

// TestARunThatWroteNoTestFailsBoth: nothing was written to a test file, so there
// is no order to read and no red to have seen.
func TestARunThatWroteNoTestFailsBoth(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", "config.go"),
		run("go test ./..."),
	})

	if got[0].Met || got[1].Met {
		t.Fatalf("score = %+v, want neither behaviour met", got)
	}
	if !strings.Contains(got[0].Why, "no test file") {
		t.Errorf("the first behaviour does not say why: %q", got[0].Why)
	}
}

// TestAnEditCountsAsAWrite: the clients name the tool differently and the exam
// reads the file, not the verb, so an edit to a test file is a test written.
func TestAnEditCountsAsAWrite(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("edit", "config_test.go"),
		write("patch", "config.go"),
	})

	if !got[0].Met {
		t.Errorf("an edit to the test file did not count: %+v", got[0])
	}
}

// TestOnlyAGoTestCommandIsTheSuite: a formatter or a build is not the red this
// behaviour is about, however often it runs between the two writes.
func TestOnlyAGoTestCommandIsTheSuite(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", "config_test.go"),
		run("gofmt -l ."),
		run("go build ./..."),
		run("go vet ./..."),
		write("write", "config.go"),
	})

	if got[1].Met {
		t.Errorf("a command that is not go test counted as the suite: %+v", got[1])
	}
}

// TestTheSuiteInsideALongerLineCounts: the shell line the agent writes carries a
// directory and flags; the behaviour asks whether go test ran, not whether it was
// alone.
func TestTheSuiteInsideALongerLineCounts(t *testing.T) {
	t.Parallel()

	got := skilleval.Score([]skilleval.Event{
		write("write", "config_test.go"),
		run("cd . && go test ./... -run TestParse"),
		write("write", "config.go"),
	})

	if !got[1].Met {
		t.Errorf("go test inside a longer command did not count: %+v", got[1])
	}
}

// TestMetCountsWhatTheExamAsksFor is the number the table prints: two here, and
// zero when nothing fits.
func TestMetCountsWhatTheExamAsksFor(t *testing.T) {
	t.Parallel()

	if n := skilleval.Met(skilleval.Score([]skilleval.Event{
		write("write", "a_test.go"), run("go test ./..."), write("write", "a.go"),
	})); n != 2 {
		t.Errorf("Met = %d, want 2", n)
	}
	if n := skilleval.Met(skilleval.Score(nil)); n != 0 {
		t.Errorf("Met on an empty stream = %d, want 0", n)
	}
}
