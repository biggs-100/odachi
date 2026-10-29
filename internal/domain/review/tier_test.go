package review_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestTheFilesInitInstallsAreNotACandidate: the agent files and the skills odachi
// installs talk about secrets and commands because that is their job. A tree that
// only changes them asks for no lens: the tool does not review its own text, and
// adopting it must not cost a four-lens review before the first line of code.
func TestTheFilesInitInstallsAreNotACandidate(t *testing.T) {
	installed := []string{
		"AGENTS.md",
		".agents/skills/data/SKILL.md",
		".opencode/agents/odachi.md",
		".opencode/agents/odachi-worker.md",
		".opencode/agents/odachi-reviewer.md",
		".claude/agents/odachi-worker.md",
		".claude/agents/odachi-verify.md",
		".claude/output-styles/odachi.md",
	}

	if got := review.Classify(installed); got != review.Passive {
		t.Errorf("Classify(installed files) = %q, want passive", got)
	}
	if lenses := review.Lenses(review.Classify(installed)); len(lenses) != 0 {
		t.Errorf("lenses = %v, want none", lenses)
	}

	// The directory is not plumbing: a settings file that grants a tool, a config
	// with a token or a human's agent stays in the candidate, because agent
	// privileges and secrets live exactly there.
	for _, path := range []string{
		".claude/settings.json",
		".opencode/config.json",
		".opencode/agents/odachi-helper.md", // the prefix is not the rule: this one is the human's
		".claude/agents/odachi-tools.md",
		".opencode/agents/helper.md", // nor is the extension: an agent file grants tools
	} {
		if got := review.Classify([]string{path}); got == review.Passive {
			t.Errorf("Classify(%s) = passive, want the file reviewed", path)
		}
	}

	// An agent file the human wrote asks for the lens that reads code, not for
	// silence: it decides what the agent may do.
	if got := review.Lenses(review.Classify([]string{".opencode/agents/odachi-helper.md"})); !slices.Equal(got, []string{review.Reliability}) {
		t.Errorf("lenses(.opencode/agents/odachi-helper.md) = %v, want [reliability]", got)
	}

	// The fence and .gitignore stay in the candidate: they are human decisions and
	// an ignore rule changes the bytes the tree covers.
	if got := review.Classify([]string{".gitignore", "odd/fence.toml"}); got == review.Passive {
		t.Errorf("Classify(.gitignore, fence) = passive, want ordinary code to review")
	}

	// A nested AGENTS.md is a document the human wrote inside the tree; only the
	// root block init owns is plumbing.
	if got := review.Classify([]string{"docs/AGENTS.md"}); got != review.Passive {
		t.Errorf("Classify(docs/AGENTS.md) = %q, want passive as a document", got)
	}
}

// TestTheTextInitInstallsDoesNotRaiseTheClass: a skill that says "never hardcode
// a secret" is the tool's own text, not a security change in this repo. The
// content signals read the candidate's lines, so the plumbing rows must not be
// read: adopting odachi used to demand all four lenses on the plan commit
// because a skill mentioned secrets.
func TestTheTextInitInstallsDoesNotRaiseTheClass(t *testing.T) {
	diff := "diff --git a/.agents/skills/data/SKILL.md b/.agents/skills/data/SKILL.md\n" +
		"--- a/.agents/skills/data/SKILL.md\n+++ b/.agents/skills/data/SKILL.md\n" +
		"@@ -1 +1,2 @@\n+never hardcode a secret or a token\n"
	paths := []string{".gitignore", ".agents/skills/data/SKILL.md"}

	if got := review.Assess(paths, diff); got != review.Focused {
		t.Errorf("Assess = %q, want focused: the tool's own text is not a security change", got)
	}
	if reason := review.Reason(paths, diff); strings.Contains(reason, "secrets") {
		t.Errorf("Reason = %q, want no content signal from the installed files", reason)
	}
}

// TestNeedAsksPerFile: coverage is answered per file, so the demand is per file
// too. Each path answers the lenses its own class requires, and a path whose
// class asks for nothing never appears: asking a document for a specialist's
// row is ceremony, not evidence.
func TestNeedAsksPerFile(t *testing.T) {
	need := review.Need([]string{
		"internal/app/engine/turno.go",
		"internal/auth/session.go",
		"docs/notas.md",
	}, "")

	if got := need["internal/app/engine/turno.go"]; !slices.Equal(got, []string{review.Reliability}) {
		t.Errorf("ordinary code asks %v, want reliability only", got)
	}
	if got := need["internal/auth/session.go"]; len(got) != 4 {
		t.Errorf("code that can hurt asks %v, want all four lenses", got)
	}
	if _, asks := need["docs/notas.md"]; asks {
		t.Errorf("a plain document never asks: %v", need["docs/notas.md"])
	}
}

// TestAssessRaisesTheClassFromContent: the path class is the floor and the
// added lines can raise it — never lower it. Concurrency and secrets are the
// two categories a path cannot see; prose never turns a document into code.
func TestAssessRaisesTheClassFromContent(t *testing.T) {
	concurrent := "diff --git a/internal/engine/engine.go b/internal/engine/engine.go\n--- a/internal/engine/engine.go\n+++ b/internal/engine/engine.go\n@@ -1 +1,2 @@\n package engine\n+go func() { sync.Mutex{} }()\n"
	if got := review.Assess([]string{"internal/engine/engine.go"}, concurrent); got != review.Full {
		t.Errorf("content with a goroutine and a mutex asks %q, want full", got)
	}

	secret := "diff --git a/internal/engine/config.go b/internal/engine/config.go\n--- a/internal/engine/config.go\n+++ b/internal/engine/config.go\n@@ -1 +1,2 @@\n package config\n+token := os.Getenv(\"TOKEN\")\n"
	if got := review.Assess([]string{"internal/engine/config.go"}, secret); got != review.Full {
		t.Errorf("content that reads a token asks %q, want full", got)
	}

	plain := "diff --git a/internal/engine/engine.go b/internal/engine/engine.go\n--- a/internal/engine/engine.go\n+++ b/internal/engine/engine.go\n@@ -1 +1,2 @@\n package engine\n+func Nueva() string { return \"x\" }\n"
	if got := review.Assess([]string{"internal/engine/engine.go"}, plain); got != review.Focused {
		t.Errorf("plain code asks %q, want focused: content adds review, it does not invent it", got)
	}

	doc := "diff --git a/docs/notas.md b/docs/notas.md\n--- a/docs/notas.md\n+++ b/docs/notas.md\n@@ -1 +1,2 @@\n # notas\n+el token y el mutex se explican aqui\n"
	if got := review.Assess([]string{"docs/notas.md"}, doc); got != review.Passive {
		t.Errorf("prose asks %q, want passive: a document's words are not code", got)
	}

	auth := "diff --git a/internal/auth/session.go b/internal/auth/session.go\n--- a/internal/auth/session.go\n+++ b/internal/auth/session.go\n@@ -1 +1,2 @@\n package auth\n+func Nueva() {}\n"
	if got := review.Assess([]string{"internal/auth/session.go"}, auth); got != review.Full {
		t.Errorf("an auth path asks %q, want full: content never lowers the path class", got)
	}
}

// TestNeedAsksPerFileWithContent: coverage follows the same assessment. A file
// whose own added lines touch concurrency owes the four rows; its neighbour
// keeps its path class; a document owes nothing.
func TestNeedAsksPerFileWithContent(t *testing.T) {
	diff := "diff --git a/internal/engine/engine.go b/internal/engine/engine.go\n--- a/internal/engine/engine.go\n+++ b/internal/engine/engine.go\n@@ -1 +1,2 @@\n package engine\n+go func() { sync.Mutex{} }()\n"
	need := review.Need([]string{"internal/engine/engine.go", "internal/app/turno.go", "docs/notas.md"}, diff)

	if got := need["internal/engine/engine.go"]; len(got) != 4 {
		t.Errorf("the concurrent file asks %v, want the four lenses", got)
	}
	if got := need["internal/app/turno.go"]; !slices.Equal(got, []string{review.Reliability}) {
		t.Errorf("the plain file asks %v, want reliability only", got)
	}
	if _, asks := need["docs/notas.md"]; asks {
		t.Errorf("the document never asks: %v", need["docs/notas.md"])
	}
}

// TestClassifyDocumentationOnlyIsPassive: a change that only touches documents
// has nothing for a specialist lens to look at.
func TestClassifyDocumentationOnlyIsPassive(t *testing.T) {
	cases := map[string][]string{
		"readme":             {"README.md"},
		"docs directory":     {"docs/guia.md", "docs/decisiones/0001-x.md"},
		"task documents":     {"odd/tasks/0001-turnos.md", "brief.md"},
		"plain text":         {"notas.txt"},
		"image inside docs":  {"docs/diagrama.png"},
		"readme without ext": {"README"},
		"nothing changed":    nil,
	}

	for name, paths := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(paths); tier != review.Passive {
				t.Errorf("Classify(%v) = %q, want %q", paths, tier, review.Passive)
			}
		})
	}
}

// TestClassifyVolumeDoesNotDecide: a thousand lines of documentation are still
// documentation. A volume rule would turn a typo fix in a big README into a
// four-lens review, and a five-line auth change into nothing.
func TestClassifyVolumeDoesNotDecide(t *testing.T) {
	docs := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		docs = append(docs, fmt.Sprintf("docs/capitulo-%03d.md", i))
	}

	if tier := review.Classify(docs); tier != review.Passive {
		t.Errorf("Classify(1000 docs) = %q, want %q: volume never decides", tier, review.Passive)
	}

	// Five lines of auth are one path: the tier is about what the paths touch,
	// never about how much text they carry.
	if tier := review.Classify([]string{"internal/auth/session.go"}); tier != review.Full {
		t.Errorf("Classify(auth) = %q, want %q", tier, review.Full)
	}
}

// TestClassifySensitiveSignalsAreFull is the table Contract XIV fixes: one test
// per signal, so a signal removed from the list is a failing test and not a
// silent weakening.
func TestClassifySensitiveSignalsAreFull(t *testing.T) {
	signals := map[string]string{
		"auth":        "internal/auth/session.go",
		"login":       "web/login.tsx",
		"session":     "internal/session/store.go",
		"token":       "internal/tokens.go",
		"password":    "internal/passwords/hash.go",
		"secret":      "internal/secrets/loader.go",
		"permission":  "internal/permissions/policy.go",
		"role":        "internal/roles/rbac.go",
		"payment":     "internal/payments/stripe.go",
		"billing":     "internal/billing/invoice.go",
		"migrat":      "db/migrations/0007-turnos.sql",
		"schema":      "db/schema.sql",
		"api":         "internal/api/handler.go",
		"contract":    "contracts/turnos.md",
		"openapi":     "openapi.yaml",
		"proto":       "proto/turnos.proto",
		"graphql":     "graphql/schema.graphql",
		"sdk":         "sdk/client.go",
		"concurrency": "internal/concurrency/limiter.go",
		"race":        "internal/race/detector.go",
		"mutex":       "internal/mutex/lock.go",
		"lock":        "internal/lock/store.go",
		"async":       "internal/async/pool.go",
		"worker":      "internal/worker/pool.go",
		"queue":       "internal/queue/turns.go",
		".env":        ".env",
		"docker":      "Dockerfile",
		"ci":          ".github/workflows/ci.yml",
		".github":     ".github/CODEOWNERS",
		"deploy":      "deploy/produccion.yaml",
		"release":     ".github/workflows/release.yml",
		"goreleaser":  "goreleaser.yaml",
		".goreleaser": ".goreleaser.yaml",
		"Makefile":    "Makefile",
		"go.mod":      "go.mod",
		"terraform":   "infra/terraform/main.tf",
		"helm":        "helm/values.yaml",
		"k8s":         "k8s/deployment.yaml",
		"pipeline":    ".github/workflows/pipeline.yml",
	}

	for signal, path := range signals {
		t.Run(signal, func(t *testing.T) {
			if tier := review.Classify([]string{path}); tier != review.Full {
				t.Errorf("Classify(%q) = %q, want %q: %q is a sensitive signal", path, tier, review.Full, signal)
			}
		})
	}
}

// TestClassifySignalsAreCaseInsensitive: the same tree classifies the same on
// Windows and Linux, where a path can arrive spelled either way.
func TestClassifySignalsAreCaseInsensitive(t *testing.T) {
	for _, path := range []string{"AUTH/session.go", "Internal/Payments/x.go", "MakeFile", ".ENV"} {
		if tier := review.Classify([]string{path}); tier != review.Full {
			t.Errorf("Classify(%q) = %q, want %q", path, tier, review.Full)
		}
	}
}

// TestClassifySignalBeatsDocumentation: a document whose path touches a
// sensitive signal is still full. Over-reviewing a document costs a prompt;
// under-reviewing auth costs a defect. The signals are checked first on purpose.
func TestClassifySignalBeatsDocumentation(t *testing.T) {
	if tier := review.Classify([]string{"docs/auth.md"}); tier != review.Full {
		t.Errorf("Classify(docs/auth.md) = %q, want %q", tier, review.Full)
	}
	if tier := review.Classify([]string{"docs/ci.md", "README.md"}); tier != review.Full {
		t.Errorf("Classify(docs/ci.md) = %q, want %q", tier, review.Full)
	}
}

// TestClassifyCodeIsFocused: everything that is not documentation and touches no
// signal gets one lens, not four.
func TestClassifyCodeIsFocused(t *testing.T) {
	cases := map[string][]string{
		"a package":            {"internal/gastos/x.go"},
		"the binary":           {"cmd/odachi/main.go"},
		"a test":               {"internal/domain/gate/gate_test.go"},
		"mixed with docs":      {"internal/gastos/x.go", "README.md"},
		"tooling config":       {"tsconfig.json"},
		"a directory of code":  {"web/app/page.tsx"},
		"a repository root go": {"main.go"},
	}

	for name, paths := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(paths); tier != review.Focused {
				t.Errorf("Classify(%v) = %q, want %q", paths, tier, review.Focused)
			}
		})
	}
}

// TestClassifyOneSignalInManyPathsIsEnough: five files of code and one auth file
// is a full review, because the auth file is the one that can hurt.
func TestClassifyOneSignalInManyPathsIsEnough(t *testing.T) {
	paths := []string{
		"internal/gastos/x.go",
		"internal/gastos/y.go",
		"README.md",
		"docs/guia.md",
		"internal/auth/session.go",
	}

	if tier := review.Classify(paths); tier != review.Full {
		t.Errorf("Classify(%v) = %q, want %q", paths, tier, review.Full)
	}
}

// TestLensesPerTier fixes the selection: passive asks for nothing, focused asks
// for one lens, full asks for all four in a fixed order.
func TestLensesPerTier(t *testing.T) {
	cases := map[string]struct {
		tier review.Tier
		want []string
	}{
		"passive asks for nothing": {review.Passive, nil},
		"focused asks reliability": {review.Focused, []string{"reliability"}},
		"full asks the four":       {review.Full, []string{"risk", "readability", "reliability", "resilience"}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := review.Lenses(tc.tier)
			if len(got) != len(tc.want) {
				t.Fatalf("Lenses(%q) = %v, want %v", tc.tier, got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("Lenses(%q) = %v, want %v", tc.tier, got, tc.want)
				}
			}
		})
	}
}

// TestLensesIsACopy: a caller that sorts the answer must not corrupt the table
// for the next caller. The order is the product here.
func TestLensesIsACopy(t *testing.T) {
	first := review.Lenses(review.Full)
	first[0] = "inventada"

	second := review.Lenses(review.Full)
	if second[0] != "risk" {
		t.Fatalf("Lenses(Full)[0] = %q after a caller mutated the slice, want %q", second[0], "risk")
	}
}

// TestAddNeverRemovesWhatTheEvidenceAsksFor: extra lenses add, they never take
// away. The result is always in the fixed order and without duplicates.
func TestAddNeverRemovesWhatTheEvidenceAsksFor(t *testing.T) {
	cases := map[string]struct {
		required []string
		extra    []string
		want     []string
	}{
		"nothing added": {
			required: []string{"reliability"},
			want:     []string{"reliability"},
		},
		"one added": {
			required: []string{"reliability"},
			extra:    []string{"risk"},
			want:     []string{"risk", "reliability"},
		},
		"duplicates collapse": {
			required: []string{"reliability", "risk"},
			extra:    []string{"risk", "risk"},
			want:     []string{"risk", "reliability"},
		},
		"passive plus one": {
			extra: []string{"readability"},
			want:  []string{"readability"},
		},
		"the four in order": {
			extra: []string{"resilience", "risk", "reliability", "readability"},
			want:  []string{"risk", "readability", "reliability", "resilience"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := review.Add(tc.required, tc.extra)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("Add(%v, %v) = %v, want %v", tc.required, tc.extra, got, tc.want)
			}
		})
	}
}

// TestParseLensesReadsTheFlagValue: the flag takes a name, a comma list, or
// `all`, and refuses a name that is not a lens by naming the four.
func TestParseLensesReadsTheFlagValue(t *testing.T) {
	cases := map[string]struct {
		spec string
		want []string
	}{
		"one":               {"risk", []string{"risk"}},
		"one with spaces":   {" risk ", []string{"risk"}},
		"all":               {"all", []string{"risk", "readability", "reliability", "resilience"}},
		"a comma list":      {"risk,resilience", []string{"risk", "resilience"}},
		"list out of order": {"resilience,risk", []string{"risk", "resilience"}},
		"repeated":          {"risk,risk", []string{"risk"}},
		"empty":             {"", nil},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := review.ParseLenses(tc.spec)
			if err != nil {
				t.Fatalf("ParseLenses(%q) = %v", tc.spec, err)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("ParseLenses(%q) = %v, want %v", tc.spec, got, tc.want)
			}
		})
	}
}

// TestParseLensesRefusesAnInventedLens: a typo must not quietly review nothing.
func TestParseLensesRefusesAnInventedLens(t *testing.T) {
	for _, spec := range []string{"seguridad", "risk,security", "all,risk"} {
		got, err := review.ParseLenses(spec)
		if err == nil {
			t.Errorf("ParseLenses(%q) = %v, want a refusal", spec, got)
			continue
		}
		if !strings.Contains(err.Error(), review.Risk) {
			t.Errorf("ParseLenses(%q) error %q does not name the valid lenses", spec, err)
		}
	}
}

// TestTierIsAStringOnPurpose: the tier travels to the prompt and to a message
// without a translation layer, so its spelling is the contract.
func TestTierIsAStringOnPurpose(t *testing.T) {
	if string(review.Passive) != "passive" || string(review.Focused) != "focused" || string(review.Full) != "full" {
		t.Fatalf("tiers = %q %q %q, want passive focused full", review.Passive, review.Focused, review.Full)
	}
}

// TestClassifyIgnoresTheStoreTheToolWrites: writing a proof creates
// proof/.gitignore, which is plumbing this tool owns and not a change the candidate
// made: counting it turned a documentation-only change into a focused review the
// second time it was asked. The verdicts are not here at all: they live outside the
// worktree, so they never appear in a candidate.
func TestClassifyIgnoresTheStoreTheToolWrites(t *testing.T) {
	cases := map[string][]string{
		"docs plus a proof": {"docs/notas.md", "proof/0001-x.txt", "proof/.gitignore"},
	}

	for name, paths := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(paths); tier != review.Passive {
				t.Errorf("Classify(%v) = %q, want %q: the tool's own stores are not the candidate", paths, tier, review.Passive)
			}
		})
	}
}

// TestReasonNamesWhatMakesTheTreeAsk: the class is the gate, and the reason is
// the line that lets a human decide whether the review pays before paying for
// it. It names the signal and the path, never "high risk" in the abstract.
func TestReasonNamesWhatMakesTheTreeAsk(t *testing.T) {
	concurrent := "diff --git a/internal/gastos/x.go b/internal/gastos/x.go\n--- a/internal/gastos/x.go\n+++ b/internal/gastos/x.go\n@@ -1 +1,2 @@\n package gastos\n+go func() { sync.Mutex{} }()\n"
	secret := "diff --git a/internal/gastos/x.go b/internal/gastos/x.go\n--- a/internal/gastos/x.go\n+++ b/internal/gastos/x.go\n@@ -1 +1,2 @@\n package gastos\n+const password = \"hunter2\"\n"

	cases := map[string]struct {
		paths []string
		diff  string
		want  string
	}{
		"nothing changed":          {nil, "", "nothing changed"},
		"only the tool's store":    {[]string{"proof/.gitignore"}, "", "nothing changed"},
		"documents only":           {[]string{"README.md", "docs/usage.md"}, "", "documents only"},
		"ordinary code":            {[]string{"internal/gastos/x.go"}, "", "ordinary code"},
		"a path the fence knows":   {[]string{"internal/auth/session.go"}, "", "touches auth (internal/auth/session.go)"},
		"content adds concurrency": {[]string{"internal/gastos/x.go"}, concurrent, "added lines touch concurrency"},
		"content adds secrets":     {[]string{"internal/gastos/x.go"}, secret, "added lines touch secrets"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := review.Reason(tc.paths, tc.diff); got != tc.want {
				t.Errorf("Reason(%v) = %q, want %q", tc.paths, got, tc.want)
			}
		})
	}
}

// TestClassifyStillSeesTheCandidatesOwnGeneratedFilesFiltered: filtering the
// stores must not filter the code next to them.
func TestClassifyStillSeesTheCandidatesOwnFiles(t *testing.T) {
	cases := map[string]struct {
		paths []string
		want  review.Tier
	}{
		"code next to the store":  {[]string{"proof/.gitignore", "internal/gastos/x.go"}, review.Focused},
		"auth next to the store":  {[]string{"proof/.gitignore", "internal/auth/session.go"}, review.Full},
		"the fence is a decision": {[]string{"odd/fence.toml"}, review.Focused},
		"the profile is a config": {[]string{"odd/profile.toml"}, review.Focused},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(tc.paths); tier != tc.want {
				t.Errorf("Classify(%v) = %q, want %q", tc.paths, tier, tc.want)
			}
		})
	}
}

// TestClassifyShippedTextIsCode: a Markdown the binary packs is not a document.
// A skill is instructions the agent obeys and a review prompt decides what a
// reviewer looks for; both are embedded (skills/, internal/assets/) and both are
// read by the suite, so a tree that only edits one is ordinary code and reaches a
// lens. Calling them documents is how a tree that changed the reviewer's own
// prompt asked for no review at all.
func TestClassifyShippedTextIsCode(t *testing.T) {
	for _, path := range []string{
		"skills/drive/SKILL.md",
		"internal/assets/review/risk.md",
	} {
		if got := review.Classify([]string{path}); got != review.Focused {
			t.Errorf("Classify(%s) = %q, want focused", path, got)
		}
		if lenses := review.Need([]string{path}, ""); !slices.Equal(lenses[path], []string{review.Reliability}) {
			t.Errorf("Need(%s) = %v, want [%s]", path, lenses[path], review.Reliability)
		}
	}
}

// TestShippedTextNamesWhatTheBinaryPacks: one list answers "does this travel
// inside the binary or beside it". It is asked twice - by the classifier, for the
// lenses, and by the proof lane, for the suite - so it names the installed copy
// too, even though that copy is plumbing and never reaches a lens.
func TestShippedTextNamesWhatTheBinaryPacks(t *testing.T) {
	for _, path := range []string{
		"skills/drive/SKILL.md",
		"internal/assets/review/risk.md",
		".agents/skills/drive/SKILL.md",
	} {
		if !review.ShippedText([]string{path}) {
			t.Errorf("ShippedText(%s) = false, want true", path)
		}
	}
	for _, path := range []string{"docs/decisions/0001-a.md", "README.md", "internal/app/check/check.go"} {
		if review.ShippedText([]string{path}) {
			t.Errorf("ShippedText(%s) = true, want false", path)
		}
	}

	// The installed copy is plumbing: it is shipped text for the proof, and no
	// lens for the review (0043, the adoption costs nothing).
	if got := review.Classify([]string{".agents/skills/drive/SKILL.md"}); got != review.Passive {
		t.Errorf("Classify(installed skill) = %q, want passive", got)
	}
}
