package scope_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/scope"
)

// TestTheRuleDeniesEveryExternalDirectory: these bytes are the boundary. A rule
// that stops naming external_directory confines nothing, and a run that writes it
// while believing otherwise is the defect 0323's field run found.
func TestTheRuleDeniesEveryExternalDirectory(t *testing.T) {
	t.Parallel()

	var conf struct {
		Permissions []struct {
			Action   string `json:"action"`
			Resource string `json:"resource"`
			Effect   string `json:"effect"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(scope.Render(), &conf); err != nil {
		t.Fatalf("the rule is not JSON the client can read: %v\n%s", err, scope.Render())
	}

	denied := false
	for _, rule := range conf.Permissions {
		if rule.Action != "external_directory" {
			continue
		}
		if rule.Resource != "*" {
			t.Errorf("the external directory rule names %q, want every path", rule.Resource)
		}
		if rule.Effect == "deny" {
			denied = true
		}
	}
	if !denied {
		t.Errorf("nothing denies an external directory, so the reviewer's tools are unconfined:\n%s", scope.Render())
	}
}

// TestTheRuleIsPinnedByItsBytes: a file a client reads on every run is a shape
// somebody else consumes. Changing it is a deliberate act with a reason, not a
// reformat that lands with something else.
func TestTheRuleIsPinnedByItsBytes(t *testing.T) {
	t.Parallel()

	const pinned = `{
  "$schema": "https://opencode.ai/config.json",
  "permissions": [
    { "action": "external_directory", "resource": "*", "effect": "deny" }
  ]
}
`
	if got := string(scope.Render()); got != pinned {
		t.Errorf("the rule changed:\n%s\nwant:\n%s", got, pinned)
	}
	if !strings.HasSuffix(scope.FileName, ".json") {
		t.Errorf("FileName = %q, and the client discovers a JSON document", scope.FileName)
	}
}
