package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDBCraftTakesTheRulesThatCostIncidents: the door for a schema and a migration
// exists, and its rules are the ones whose absence is an outage: the lock timeout, the
// expand and contract, the constraint in two steps and the types that are always chosen
// wrong. Pinned because a page that lost them would still read like advice (0311).
func TestDBCraftTakesTheRulesThatCostIncidents(t *testing.T) {
	skill := mustSkill(t, "db-craft")
	body := string(skill.Main())
	flat := strings.Join(strings.Fields(body), " ")

	assertEveryReferenceIsRouted(t, skill, body)

	for _, want := range []string{
		"lock_timeout",
		"expand the new shape",
		"NOT VALID",
		"numeric",
		"timestamptz",
		"retries the whole transaction",
		"`data`'s door",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("db-craft does not say %q", want)
		}
	}
}

// TestDBCraftCarriesItsDepth: every reference is reachable from a gate and listed, and
// the two pages that break production say the thing that makes them worth reading -the
// lock table and the four types-. A reference nobody routes is a file nobody reads
// (0311).
func TestDBCraftCarriesItsDepth(t *testing.T) {
	skill := mustSkill(t, "db-craft")

	want := map[string][]string{
		"references/migrations.md":   {"lock_timeout", "ACCESS EXCLUSIVE", "NOT VALID", "expand", "pg_blocking_pids"},
		"references/schema.md":       {"numeric", "timestamptz", "jsonb", "Soft delete", "snake_case"},
		"references/transactions.md": {"Read Committed", "retry", "SKIP LOCKED", "outbox", "idempotency"},
		"references/indexes.md":      {"the order is the index", "INCLUDE", "GIN", "EXPLAIN", "idx_scan"},
		"references/queries.md":      {"N+1", "keyset", "OFFSET", "NOT IN", "p95"},
		"references/access.md":       {"FORCE ROW LEVEL SECURITY", "SET LOCAL", "WITH CHECK", "BYPASSRLS", "sslmode"},
		"references/operations.md":   {"restore", "pg_stat_statements", "RTO", "transaction mode", "autovacuum"},
		"references/the-failures.md": {"lock_timeout", "N+1", "FORCE", "outbox", "the page"},
	}
	for path, phrases := range want {
		var content string
		for _, file := range skill.Files {
			if file.Path == path {
				content = strings.Join(strings.Fields(string(file.Content)), " ")
			}
		}
		if content == "" {
			t.Fatalf("db-craft does not carry %s", path)
		}
		for _, phrase := range phrases {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(phrase)) {
				t.Errorf("%s does not say %q", path, phrase)
			}
		}
	}
}

// TestDBCraftCountsWhatItShips: the impact map says how many references the skill
// carries, and a count that disagrees with the folder is the line a maintainer trusts
// while it lies. The number in the prose is pinned against the payload (0311).
func TestDBCraftCountsWhatItShips(t *testing.T) {
	skill := mustSkill(t, "db-craft")

	references := 0
	for _, file := range skill.Files {
		if strings.HasPrefix(file.Path, "references/") {
			references++
		}
	}
	if references != 8 {
		t.Fatalf("db-craft carries %d references, and the map and the changelog say eight", references)
	}

	// The two sentences the review corrected, pinned so a rewrite cannot put the
	// plausible version back: a read-only transaction is not a plan hint, and
	// `row_security = off` fails a query instead of bypassing a policy (0311).
	transactions := fileContent(t, "references/transactions.md")
	if !strings.Contains(transactions, "not a performance hint") {
		t.Error("transactions.md claims a read-only transaction changes the plan")
	}
	access := fileContent(t, "references/access.md")
	if !strings.Contains(access, "**3. A policy broader than the tenant.**") {
		t.Error("access.md does not state the third hazard as a policy broader than the tenant")
	}
	if !strings.Contains(access, "makes a query that would apply a policy fail") {
		t.Error("access.md does not say what row_security = off does")
	}
	if strings.Contains(access, "row_security = off` setting ignore") {
		t.Error("access.md lists row_security = off among the doors around the policy again")
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "impact-map.md"))
	if err != nil {
		t.Fatalf("read docs/impact-map.md: %v", err)
	}
	if !strings.Contains(string(data), "`db-craft` ships eight references") {
		t.Error("the impact map does not say db-craft ships eight references")
	}
}

// fileContent is one reference of a skill, flattened to a single line so a phrase can
// be searched across the wrapping.
func fileContent(t *testing.T, path string) string {
	t.Helper()
	for _, file := range mustSkill(t, "db-craft").Files {
		if file.Path == path {
			return strings.Join(strings.Fields(string(file.Content)), " ")
		}
	}
	t.Fatalf("db-craft does not carry %s", path)
	return ""
}

// TestDataSendsTheSchemaToItsOwner: `data` decides what is kept; the schema and the
// migration belong to db-craft. A trigger that claims both leaves the agent with two
// doors for one situation, which is the defect the catalog keeps paying for (0311).
func TestDataSendsTheSchemaToItsOwner(t *testing.T) {
	data := mustSkill(t, "data")
	flat := strings.Join(strings.Fields(string(data.Main())), " ")

	if !strings.Contains(flat, "`db-craft`'s door") {
		t.Error("data does not send the schema and the migration to db-craft")
	}
	if strings.Contains(flat, "a migration appears") {
		t.Error("data still claims the migration as its own situation")
	}
}
