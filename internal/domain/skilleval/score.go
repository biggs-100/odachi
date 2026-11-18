package skilleval

import (
	"path/filepath"
	"strings"
)

// Event is one tool call the transcript named, in the order it happened. Path is
// the file a write touched, Command the line a shell ran, and Name the skill a
// load asked for: three shapes of one stream, and the exam reads only these.
type Event struct {
	Tool    string
	Path    string
	Command string
	Name    string
}

// Behaviour is one thing the exam decided about the run: whether it happened and
// the event that decided it, so a reader can check the answer without trusting
// the score.
type Behaviour struct {
	ID    string
	Label string
	Met   bool
	Why   string
}

// The two behaviours the event stream decides, in the order the exam reports
// them. The pages of the measurement name them as the rubric's first two: a test
// written before the production code, and the red observed before the fix.
const (
	TestBeforeCode = "test-before-code"
	RedBeforeFix   = "red-before-fix"
)

// Score reads a stream in order and answers the two behaviours. It decides only
// what the order of the tool calls can show: whether the test file was written
// before the code file, and whether a test command ran between the two. What the
// test is worth, and whether the name says what breaks, needs a reader, and the
// exam hands that reader the transcript instead of guessing.
func Score(events []Event) []Behaviour {
	testAt := -1
	codeAt := -1
	for i, event := range events {
		path := filepath.ToSlash(event.Path)
		switch {
		case path == "":
			continue
		case strings.HasSuffix(path, "_test.go"):
			if testAt < 0 {
				testAt = i
			}
		case strings.HasSuffix(path, ".go"):
			if codeAt < 0 {
				codeAt = i
			}
		}
		if testAt >= 0 && codeAt >= 0 {
			break
		}
	}

	before := Behaviour{ID: TestBeforeCode, Label: "the test was written before the code"}
	switch {
	case testAt < 0:
		before.Why = "no test file was written in the run"
	case codeAt < 0:
		before.Met = true
		before.Why = "the test file was written and no code file was"
	case testAt < codeAt:
		before.Met = true
		before.Why = eventAt(events, testAt, "came before") + " the first code file"
	default:
		before.Why = eventAt(events, testAt, "came after") + " the first code file"
	}

	red := Behaviour{ID: RedBeforeFix, Label: "the suite ran between the test and the code"}
	switch {
	case !before.Met:
		red.Why = "the test did not come first, so there was nothing to see red"
	default:
		for i := testAt + 1; i < len(events); i++ {
			if !runsTheSuite(events[i]) {
				continue
			}
			if codeAt < 0 || i < codeAt {
				red.Met = true
				red.Why = "the suite ran between the test and the code: " + oneLine(events[i].Command)
			}
			break
		}
		if !red.Met && red.Why == "" {
			red.Why = "the suite never ran between the test and the code"
		}
	}

	return []Behaviour{before, red}
}

// Met counts the behaviours met, out of the ones the exam asks for.
func Met(behaviours []Behaviour) int {
	met := 0
	for _, behaviour := range behaviours {
		if behaviour.Met {
			met++
		}
	}
	return met
}

// runsTheSuite answers whether one event is the project's test command. It reads
// the line the shell ran and asks for "go test": a different command that happens
// to write files is not the red this behaviour is about.
func runsTheSuite(event Event) bool {
	if event.Tool != "bash" && event.Tool != "shell" && event.Tool != "command" {
		return false
	}
	fields := strings.Fields(event.Command)
	for i, field := range fields {
		if field != "go" {
			continue
		}
		if i+1 < len(fields) && strings.TrimSpace(fields[i+1]) == "test" {
			return true
		}
	}
	return false
}

func eventAt(events []Event, at int, verb string) string {
	return "the test file " + verb + ": " + oneLine(events[at].Path)
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
