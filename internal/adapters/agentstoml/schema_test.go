package agentstoml_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/agentstoml"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// TestRenderWritesTheSchemaItWasWrittenWith: the store says which odachi wrote
// it, so the next key it gains arrives as an upgrade instead of a typo.
func TestRenderWritesTheSchemaItWasWrittenWith(t *testing.T) {
	data, err := agentstoml.Render(persona.Store{Agents: persona.SelectionOpenCode})
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	if !strings.Contains(string(data), "schema = "+strconv.Itoa(schema.Current)) {
		t.Fatalf("Render = %q, want the schema written", data)
	}

	if _, err := agentstoml.Decode(data); err != nil {
		t.Fatalf("Decode(Render) = %v", err)
	}
}

// TestDecodeOfANewerStoreAsksForAnUpgrade: the version decides, not the keys that
// came with it. This is the file that cost a close when project_agents arrived:
// the installed binary called a good store broken.
func TestDecodeOfANewerStoreAsksForAnUpgrade(t *testing.T) {
	future := schema.Current + 1
	_, err := agentstoml.Decode([]byte("schema = " + strconv.Itoa(future) + "\nagents = \"opencode\"\nfuture_key = true\n"))
	if err == nil {
		t.Fatal("Decode accepted a store from a newer odachi")
	}
	message := err.Error()
	if !strings.Contains(message, "upgrade") {
		t.Errorf("the refusal %q does not name the upgrade", message)
	}
	if strings.Contains(message, "unknown keys") {
		t.Errorf("the refusal %q names unknown keys: a newer store is refused by its version", message)
	}
}

// TestDecodeOfAStoreWithoutASchemaStillWorks: every store written before the key
// existed carries no version and must keep choosing the same client.
func TestDecodeOfAStoreWithoutASchemaStillWorks(t *testing.T) {
	got, err := agentstoml.Decode([]byte("agents = \"opencode\"\nproject_agents = \"none\"\n"))
	if err != nil {
		t.Fatalf("Decode = %v", err)
	}
	if got.Agents != persona.SelectionOpenCode || got.ProjectAgents != persona.ProjectAgentsNone {
		t.Errorf("Decode = %+v, want the store read as it always was", got)
	}
}

// TestTheSkillsKeyMovedTheSchema: a store this binary writes carries a version an
// older binary refuses as newer, so the key it never heard of arrives as "upgrade"
// and not as a typo for the human to fix.
func TestTheSkillsKeyMovedTheSchema(t *testing.T) {
	if schema.Current <= schema.First {
		t.Fatal("the store gained a key and its schema did not move")
	}

	data, err := agentstoml.Render(persona.Store{Agents: persona.SelectionBoth, Skills: persona.SkillsHome})
	if err != nil {
		t.Fatalf("Render = %v", err)
	}
	if !strings.Contains(string(data), "schema = "+strconv.Itoa(schema.Current)) {
		t.Fatalf("Render = %q, want schema %d", data, schema.Current)
	}
}
