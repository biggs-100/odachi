package review_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// TestNeedAsksPerFile: coverage is answered per file, so the demand is per file
// too. Each path answers the lenses its own class requires, and a path whose
// class asks for nothing never appears: asking a document for a specialist's
// row is ceremony, not evidence.
func TestNeedAsksPerFile(t *testing.T) {
	need := review.Need([]string{
		"internal/app/engine/turno.go",
		"internal/auth/session.go",
		"docs/notas.md",
	})

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

// TestClassifyIgnoresTheStoresTheToolWrites: writing a verdict creates
// .odachi/reviews/.gitignore, and writing a proof creates proof/.gitignore. Both
// are plumbing this tool owns, not changes the candidate made: counting them
// turned a documentation-only change into a focused review the second time it was
// asked, and it is the only reason a docs-only candidate ever asked for a lens.
func TestClassifyIgnoresTheStoresTheToolWrites(t *testing.T) {
	cases := map[string][]string{
		"docs plus the store's ignore file": {"docs/notas.md", ".odachi/reviews/.gitignore"},
		"only the store's ignore file":      {".odachi/reviews/.gitignore"},
		"docs plus a proof":                 {"docs/notas.md", "proof/0001-x.txt", "proof/.gitignore"},
	}

	for name, paths := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(paths); tier != review.Passive {
				t.Errorf("Classify(%v) = %q, want %q: the tool's own stores are not the candidate", paths, tier, review.Passive)
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
		"code next to the store":  {[]string{".odachi/reviews/.gitignore", "internal/gastos/x.go"}, review.Focused},
		"auth next to the store":  {[]string{".odachi/reviews/.gitignore", "internal/auth/session.go"}, review.Full},
		"the fence is a decision": {[]string{".odachi/fence.toml"}, review.Focused},
		"the profile is a config": {[]string{".odachi/profile.toml"}, review.Focused},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tier := review.Classify(tc.paths); tier != tc.want {
				t.Errorf("Classify(%v) = %q, want %q", tc.paths, tier, tc.want)
			}
		})
	}
}
