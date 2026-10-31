package adoption_test

import (
	"errors"
	"testing"

	"github.com/biggs-100/odachi/internal/app/adoption"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestTheCraftStaysInTheMachineOnlyForARepositoryThatLives: the craft half of the
// inference, with the four facts that move it. One definition is what makes init
// and role answer the same about the same project.
func TestTheCraftStaysInTheMachineOnlyForARepositoryThatLives(t *testing.T) {
	cases := []struct {
		name  string
		facts adoption.Facts
		want  bool
	}{
		{"a repository that lives, nobody chose, no copy here", adoption.Facts{Lives: true}, true},
		{"a folder with no history is a project starting now", adoption.Facts{}, false},
		{"a project that chose the project keeps it", adoption.Facts{Lives: true, Stored: persona.Store{Skills: persona.SkillsProject}}, false},
		{"a project that already carries the copy keeps it", adoption.Facts{Lives: true, CraftHere: true}, false},
	}
	for _, c := range cases {
		if got := c.facts.Craft(); got != c.want {
			t.Errorf("%s: Craft() = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestTheAgentHalfHasItsOwnFacts: the agent is inferred separately from the craft.
// A project that keeps its copy can still keep the agent in the machine, and vice
// versa -folding the two would make one answer move the other.
func TestTheAgentHalfHasItsOwnFacts(t *testing.T) {
	if !(adoption.Facts{Lives: true, CraftHere: true}).Agent() {
		t.Error("a project that carries the craft lost the agent inference")
	}
	if (adoption.Facts{Lives: true, Stored: persona.Store{ProjectAgents: persona.ProjectAgentsWrite}}).Agent() {
		t.Error("a recorded choice to write the mirrors was overridden by the inference")
	}
	if (adoption.Facts{Lives: true, AgentHere: true}).Agent() {
		t.Error("a project that already carries an agent file was called one without it")
	}
	if (adoption.Facts{}).Agent() {
		t.Error("a folder with no history was called a project that could adopt")
	}
}

// TestScopeIsTheOneAnswerBothVerbsAsk: the scope keeps coming from persona, and the
// adoption only feeds the one fact persona cannot see.
func TestScopeIsTheOneAnswerBothVerbsAsk(t *testing.T) {
	adopted := adoption.Facts{Lives: true}
	if scope, source := adopted.Scope(""); scope != persona.SkillsHome || source != persona.SkillsFromAdoption {
		t.Errorf("Scope() = (%q, %q), want the adoption", scope, source)
	}
	if scope, source := adopted.Scope(persona.SkillsProject); scope != persona.SkillsProject || source != persona.SkillsFromFlag {
		t.Errorf("Scope() = (%q, %q), want the flag", scope, source)
	}
}

// TestReadStopsAtTheFirstFailedQuestion: a read that fails is not a fact that is
// false. The read answers nothing and hands the failure back with its words.
func TestReadStopsAtTheFirstFailedQuestion(t *testing.T) {
	want := errors.New("permission denied")
	asked := 0

	facts, err := adoption.Read(adoption.Reader{
		Lives: func() (bool, error) { asked++; return true, nil },
		CraftHere: func() (bool, error) {
			asked++
			return false, want
		},
		AgentHere: func() (bool, error) {
			t.Error("a failed question stopped the read: the next one was still asked")
			return false, nil
		},
	})

	if !errors.Is(err, want) {
		t.Fatalf("Read() error = %v, want %v", err, want)
	}
	if facts != (adoption.Facts{}) {
		t.Errorf("Read() = %+v, want no facts at all on a failure", facts)
	}
	if asked != 2 {
		t.Errorf("asked %d questions, want it to stop at the second", asked)
	}
}

// TestReadAsksOnlyWhatItIsGiven: a verb wires what its job needs. A nil question
// is false, not a failure: role never asks about the agent half.
func TestReadAsksOnlyWhatItIsGiven(t *testing.T) {
	facts, err := adoption.Read(adoption.Reader{
		Lives: func() (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !facts.Lives || facts.CraftHere || facts.AgentHere {
		t.Errorf("Read() = %+v, want only the question that was asked", facts)
	}
}
