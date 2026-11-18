package skilleval_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skilleval"
)

// stream is the three shapes the exam reads, written the way the client writes
// them: a write with a filePath, a shell with a command, a skill with an id, and
// narration the exam ignores.
const stream = `{"type":"step_start","part":{"type":"step-start"}}
{"type":"tool_use","part":{"type":"tool","tool":"write","state":{"input":{"filePath":"C:\\work\\config_test.go","content":"package probe"}}}}
{"type":"text","part":{"type":"text","text":"writing the test first"}}
{"type":"tool_use","part":{"type":"tool","tool":"bash","state":{"input":{"command":"go test ./..."}}}}
{"type":"tool_use","part":{"type":"tool","tool":"edit","state":{"input":{"path":"C:\\work\\config.go"}}}}
{"type":"tool_use","part":{"type":"tool","tool":"skill","state":{"input":{"id":"tests"}}}}
`

// TestTheStreamKeepsItsToolCallsInOrder: the wave of events a run writes down is
// read as calls, in the order they happened, and the narration between them is
// not an event.
func TestTheStreamKeepsItsToolCallsInOrder(t *testing.T) {
	t.Parallel()

	got, err := skilleval.ParseStream([]byte(stream))
	if err != nil {
		t.Fatalf("ParseStream: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("ParseStream = %d events, want 4: %+v", len(got), got)
	}
	if got[0].Tool != "write" || !strings.HasSuffix(got[0].Path, "config_test.go") {
		t.Errorf("the first event is not the test write: %+v", got[0])
	}
	if got[1].Tool != "bash" || got[1].Command != "go test ./..." {
		t.Errorf("the second event is not the suite: %+v", got[1])
	}
	if got[2].Tool != "edit" || !strings.HasSuffix(got[2].Path, "config.go") {
		t.Errorf("the third event is not the code edit: %+v", got[2])
	}
	if got[3].Tool != "skill" || got[3].Name != "tests" {
		t.Errorf("the fourth event is not the skill load: %+v", got[3])
	}

	behaviours := skilleval.Score(got)
	if !behaviours[0].Met || !behaviours[1].Met {
		t.Errorf("the stream that shows the exam's run scores %+v, want both", behaviours)
	}
}

// TestALineThatIsNotJSONIsRefused: a stream that cannot be read is a run whose
// answer nobody can score, and the refusal names the line instead of scoring the
// half it understood.
func TestALineThatIsNotJSONIsRefused(t *testing.T) {
	t.Parallel()

	_, err := skilleval.ParseStream([]byte("{\"type\":\"tool_use\"}\nnot json\n"))
	if err == nil {
		t.Fatal("ParseStream accepted a line that is not JSON")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("the refusal does not name the line: %v", err)
	}
}

// TestAToolCallWithNoArgumentsIsStillAnEvent: a call that carries nothing is
// still a call, and the score ignores it because it names no file and no command.
func TestAToolCallWithNoArgumentsIsStillAnEvent(t *testing.T) {
	t.Parallel()

	got, err := skilleval.ParseStream([]byte(`{"type":"tool_use","part":{"tool":"todo"}}`))
	if err != nil {
		t.Fatalf("ParseStream: %v", err)
	}
	if len(got) != 1 || got[0].Tool != "todo" {
		t.Fatalf("events = %+v, want one todo", got)
	}
}

// TestAnEmptyStreamIsNoEvents: a run that wrote nothing is not a broken stream,
// and the caller says so with a score of zero rather than an error.
func TestAnEmptyStreamIsNoEvents(t *testing.T) {
	t.Parallel()

	got, err := skilleval.ParseStream(nil)
	if err != nil {
		t.Fatalf("ParseStream(nil): %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ParseStream(nil) = %+v, want none", got)
	}
}
