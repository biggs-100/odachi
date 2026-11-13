package review_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/schema"
)

// blockingFindings is the verdict the refuter is asked about: one blocker, one
// critical, and one warning that must never reach the refuter.
func blockingFindings() []review.Finding {
	return []review.Finding{
		{Path: "internal/auth/session.go", Line: 12, Severity: review.Blocker, Causal: true, Why: "una sesión se acepta sin verificar la firma", Evidence: review.EvidenceCited},
		{Path: "internal/gastos/x.go", Severity: review.Warning, Causal: true, Why: "el nombre miente"},
		{Path: "internal/app/turnos.go", Line: 40, Severity: review.Critical, Causal: true, Why: "dos turnos en el mismo horario", Evidence: review.EvidenceCited},
	}
}

// TestFindingIDsAreStable: the refuter answers about ids, so the ids have to be
// the position of each finding in the verdict the reviewer said to read.
func TestFindingIDsAreStable(t *testing.T) {
	findings := blockingFindings()
	want := []string{"F1", "F2", "F3"}

	if got := review.FindingIDs(findings); !reflect.DeepEqual(got, want) {
		t.Fatalf("FindingIDs = %v, want %v", got, want)
	}
	if got := review.FindingIDs(nil); len(got) != 0 {
		t.Fatalf("FindingIDs(nil) = %v, want none", got)
	}
}

// TestRefutationParsesTheAnswerShape: the refuter returns exactly the JSON the
// prompt asks for, and the domain reads it without guessing.
func TestRefutationParsesTheAnswerShape(t *testing.T) {
	text := `{"refutations":[{"id":"F1","verdict":"refuted","why":"el diff ya validaba la firma en la línea 12"}]}`

	parsed, err := review.ParseRefutation([]byte(text))
	if err != nil {
		t.Fatalf("ParseRefutation = %v", err)
	}
	if len(parsed.Refutations) != 1 {
		t.Fatalf("refutations = %d, want 1", len(parsed.Refutations))
	}
	judgment := parsed.Refutations[0]
	if judgment.ID != "F1" || judgment.Verdict != review.Refuted || !strings.Contains(judgment.Why, "línea 12") {
		t.Fatalf("judgment = %+v", judgment)
	}
}

// TestParseRefutationRefusesWhatItCannotUse: a refutation is an answer, so
// anything but the documented shape is refused before a file is written.
func TestParseRefutationRefusesWhatItCannotUse(t *testing.T) {
	cases := map[string]string{
		"empty":              "",
		"not json":           "no encontré nada que refutar",
		"unknown field":      `{"refutations":[],"resumen":"todo bien"}`,
		"invented verdict":   `{"refutations":[{"id":"F1","verdict":"quizas","why":"no sé"}]}`,
		"no id":              `{"refutations":[{"verdict":"refuted","why":"porque sí"}]}`,
		"two objects":        `{"refutations":[]}{"refutations":[]}`,
		"a verdict instead":  `{"reviewer":"claude","findings":[]}`,
		"refutations as map": `{"refutations":{"F1":"refuted"}}`,
	}

	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := review.ParseRefutation([]byte(text)); err == nil {
				t.Fatalf("ParseRefutation(%q) = nil error, want a refusal", text)
			}
		})
	}
}

// TestRefutationWithoutEvidenceDefaultsToStands is the rule that decides the
// whole verb: a finding that the refuter could not back with evidence is NOT
// refuted. Doubt always votes for keeping the finding.
func TestRefutationWithoutEvidenceDefaultsToStands(t *testing.T) {
	findings := blockingFindings()

	sealed, err := review.SealRefutation(review.Refutation{
		Refutations: []review.Judgment{
			{ID: "F1", Verdict: review.Refuted, Why: "   "},
			{ID: "F3", Verdict: review.Refuted, Why: ""},
		},
	}, findings, tree, now)
	if err != nil {
		t.Fatalf("SealRefutation = %v", err)
	}

	for _, judgment := range sealed.Refutations {
		if judgment.Verdict != review.Stands {
			t.Errorf("judgment %s = %q, want %q: no evidence means the finding stands", judgment.ID, judgment.Verdict, review.Stands)
		}
	}
	if _, blocking := review.BlockingAfter(findings, review.Answers{sealed}); !blocking {
		t.Fatal("a refutation with no evidence lifted the block")
	}
}

// TestSealRefutationCompletesTheRecord: a blocking finding the refuter did not
// answer stands, and the stored file says so instead of leaving the question
// open. The record is then readable a month later.
func TestSealRefutationCompletesTheRecord(t *testing.T) {
	findings := blockingFindings()

	sealed, err := review.SealRefutation(review.Refutation{
		Refuter:     "claude",
		Refutations: []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: "el diff no toca la validación, ya estaba en la línea 12"}},
	}, findings, tree, now)
	if err != nil {
		t.Fatalf("SealRefutation = %v", err)
	}

	if sealed.Tree != tree {
		t.Errorf("tree = %q, want %q", sealed.Tree, tree)
	}
	if sealed.At != now.Format("2006-01-02T15:04:05Z07:00") {
		t.Errorf("at = %q, want the stamped instant", sealed.At)
	}
	if len(sealed.Refutations) != 2 {
		t.Fatalf("refutations = %v, want one per blocking finding", sealed.Refutations)
	}
	if sealed.Refutations[0].ID != "F1" || sealed.Refutations[0].Verdict != review.Refuted {
		t.Errorf("first judgment = %+v, want the refuted one first", sealed.Refutations[0])
	}
	if sealed.Refutations[1].ID != "F3" || sealed.Refutations[1].Verdict != review.Stands {
		t.Errorf("second judgment = %+v, want F3 standing unanswered", sealed.Refutations[1])
	}
}

// TestSealRefutationRefusesAnOffListFinding: the refuter may only answer about
// the findings it was shown. A warning sneaking in would let a refutation lift
// something that never blocked.
func TestSealRefutationRefusesAnOffListFinding(t *testing.T) {
	cases := map[string]review.Judgment{
		"a warning":            {ID: "F2", Verdict: review.Stands, Why: "no importa"},
		"an invented id":       {ID: "F9", Verdict: review.Refuted, Why: "porque sí"},
		"an empty id":          {ID: " ", Verdict: review.Stands, Why: "porque sí"},
		"a non causal blocker": {ID: "F4", Verdict: review.Stands, Why: "ya estaba"},
	}

	withNonCausal := append(blockingFindings(), review.Finding{
		Path: "internal/viejo.go", Severity: review.Blocker, Causal: false, Why: "lo trajo otro cambio",
	})

	for name, judgment := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := review.SealRefutation(review.Refutation{Refutations: []review.Judgment{judgment}}, withNonCausal, tree, now)
			if err == nil {
				t.Fatalf("SealRefutation(%+v) = nil error, want a refusal", judgment)
			}
			if !strings.Contains(err.Error(), judgment.ID) && judgment.ID != " " {
				t.Errorf("error %q does not name the offending id", err)
			}
		})
	}
}

// TestSealRefutationNeedsATree: a refutation is about one tree, exactly like a
// verdict. Without the hash it cannot be stored anywhere.
func TestSealRefutationNeedsATree(t *testing.T) {
	if _, err := review.SealRefutation(review.Refutation{}, blockingFindings(), "", now); err == nil {
		t.Fatal("SealRefutation without a tree = nil error, want a refusal")
	}
}

// TestRefutesNeedsAReason: the single predicate every other rule is built on.
func TestRefutesNeedsAReason(t *testing.T) {
	refutation := review.Refutation{Refutations: []review.Judgment{
		{ID: "F1", Verdict: review.Refuted, Why: "el diff de la línea 12 ya validaba la firma"},
		{ID: "F2", Verdict: review.Refuted, Why: "  "},
		{ID: "F3", Verdict: review.Stands, Why: "no pude citar la línea"},
	}}

	if !refutation.Refutes("F1") {
		t.Error("F1 with evidence does not refute")
	}
	if refutation.Refutes("F2") {
		t.Error("F2 without evidence refutes: no evidence must mean stands")
	}
	if refutation.Refutes("F3") {
		t.Error("F3 says stands and refutes")
	}
	if refutation.Refutes("F9") {
		t.Error("a finding with no judgment refutes: the default is stands")
	}
	if (review.Refutation{}).Refutes("F1") {
		t.Error("an empty refutation refutes something")
	}
}

// TestDeriveAfterTakesRefutationsIntoAccount: a refuted blocker stops blocking,
// a standing one keeps blocking, and everything else approves as before.
func TestDeriveAfterTakesRefutationsIntoAccount(t *testing.T) {
	blockerOnly := []review.Finding{
		{Path: "internal/auth/session.go", Severity: review.Blocker, Causal: true, Why: "sin verificar la firma", Evidence: review.EvidenceCited},
	}

	cases := map[string]struct {
		findings []review.Finding
		refuted  bool
		want     string
	}{
		"blocker standing":           {blockerOnly, false, review.ChangesRequested},
		"blocker refuted":            {blockerOnly, true, review.Approved},
		"no findings":                {nil, false, review.Approved},
		"warning causal is not a no": {blockingFindings()[1:2], false, review.Approved},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			refutation := review.Refutation{}
			if tc.refuted {
				refutation.Refutations = []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: "está en el diff, ya estaba"}}
			}

			if got := review.DeriveAfter(tc.findings, review.Answers{refutation}); got != tc.want {
				t.Errorf("DeriveAfter = %q, want %q", got, tc.want)
			}
			if got := review.Derive(tc.findings); got != review.DeriveAfter(tc.findings, review.Answers{}) {
				t.Errorf("Derive = %q and DeriveAfter with no refutation = %q: they must agree", got, review.DeriveAfter(tc.findings, review.Answers{}))
			}
		})
	}
}

// TestBlockingAfterSkipsRefutedFindings: the gate names the first finding that
// still blocks, in file order, so the message is the same on every run.
func TestBlockingAfterSkipsRefutedFindings(t *testing.T) {
	findings := blockingFindings()
	refutation := review.Refutation{Refutations: []review.Judgment{
		{ID: "F1", Verdict: review.Refuted, Why: "el diff la línea 12 muestra que ya estaba"},
	}}

	finding, blocking := review.BlockingAfter(findings, review.Answers{refutation})
	if !blocking {
		t.Fatal("the critical finding F3 stopped blocking")
	}
	if finding.Why != findings[2].Why {
		t.Errorf("blocking finding = %+v, want F3", finding)
	}
}

// TestRefutationPathNamesTheTree: the refutation lives beside the verdict, named
// after the same tree, so one tree has one verdict and one refutation.
func TestRefutationPathNamesTheTree(t *testing.T) {
	name, err := review.RefutationName(tree, 1)
	if err != nil {
		t.Fatalf("RefutationName = %v", err)
	}
	if name != tree+".refuted.json" {
		t.Errorf("name = %q", name)
	}

	path, err := review.RefutationPath(tree, 1)
	if err != nil {
		t.Fatalf("RefutationPath = %v", err)
	}
	if path != review.Dir+"/"+tree+".refuted.json" {
		t.Errorf("path = %q", path)
	}

	if _, err := review.RefutationPath("no es un hash", 1); err == nil {
		t.Error("RefutationPath accepted a name that is not a tree")
	}
}

// TestRenderRefutationIsStable: the file is written once and never overwritten,
// so it has to be byte-stable and parse back to the same document.
func TestRenderRefutationIsStable(t *testing.T) {
	sealed, err := review.SealRefutation(review.Refutation{
		Refuter: "claude",
		Refutations: []review.Judgment{
			{ID: "F1", Verdict: review.Stands, Why: "la línea 12 no alcanza para defenderlo"},
		},
	}, blockingFindings(), tree, now)
	if err != nil {
		t.Fatalf("SealRefutation = %v", err)
	}

	first := review.RenderRefutation(sealed)
	second := review.RenderRefutation(sealed)
	if string(first) != string(second) {
		t.Fatal("RenderRefutation is not stable")
	}
	if !strings.HasSuffix(string(first), "\n") {
		t.Errorf("rendered refutation does not end with a newline: %q", first)
	}

	back, err := review.ParseRefutation(first)
	if err != nil {
		t.Fatalf("ParseRefutation(Render) = %v", err)
	}
	if back.Tree != tree || len(back.Refutations) != 2 {
		t.Fatalf("round trip = %+v", back)
	}
}

// refuteInput is a candidate with two blocking findings and one warning.
func refuteInput() review.RefuteInput {
	return review.RefuteInput{
		Tree:     tree,
		Base:     "HEAD",
		Reviewer: "claude",
		Findings: blockingFindings(),
		Diff:     "diff --git a/internal/auth/session.go b/internal/auth/session.go\n@@ -12 +12 @@\n-viejo\n+nuevo\n",
	}
}

// TestRefutePromptAsksTheThreeQuestions: the refuter is not asked to agree with
// the review. It is asked three questions per blocking finding, and it has to
// answer them with the diff in hand.
func TestRefutePromptAsksTheThreeQuestions(t *testing.T) {
	prompt := review.RefutePrompt(refuteInput())

	for _, want := range []string{
		"Is it causal of the diff, or was it already there?",
		"Is it a real defect that affects the user, or a preference?",
		"Is the evidence enough to defend it before the author?",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the refuter prompt does not ask %q", want)
		}
	}
}

// TestRefutePromptCarriesTheEvidenceAndTheIds: the refuter can only cite what it
// is given, so the diff and the findings with their ids are in the prompt.
func TestRefutePromptCarriesTheEvidenceAndTheIds(t *testing.T) {
	prompt := review.RefutePrompt(refuteInput())

	for _, want := range []string{
		tree,
		"HEAD",
		"-viejo",
		"+nuevo",
		"F1",
		"F3",
		"internal/auth/session.go",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the refuter prompt does not contain %q", want)
		}
	}
	if strings.Contains(prompt, "el nombre miente") {
		t.Error("the refuter prompt carries the warning F2: only BLOCKER and CRITICAL are refutable")
	}
}

// TestRefutePromptDefaultsToStands: the instruction that carries the rule, and
// the answer shape the domain parses.
func TestRefutePromptDefaultsToStands(t *testing.T) {
	prompt := review.RefutePrompt(refuteInput())

	for _, want := range []string{
		"stands",
		"refuted",
		"Cite the diff line",
		`"refutations"`,
		"When in doubt",
		"without text around it",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the refuter prompt does not contain %q", want)
		}
	}
}

// TestRefutePromptSaysWhatIsNotRefutable: the refuter must not spend its budget
// on warnings, and it must not invent findings of its own.
func TestRefutePromptSaysWhatIsNotRefutable(t *testing.T) {
	prompt := review.RefutePrompt(refuteInput())

	for _, want := range []string{"WARNING", "SUGGESTION", "do not add new findings"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the refuter prompt does not say %q", want)
		}
	}
}

// TestSealRefutationStampsTheClockAndKeepsTheClaim: the refuter writes `at`, and
// the tool kept it, the same way it kept a reviewer's until 0286. The seal stamps
// the instant the clone wrote the file and keeps the declared time beside it as a
// claim, so the record says both and a reader sees they disagree (0291).
func TestSealRefutationStampsTheClockAndKeepsTheClaim(t *testing.T) {
	t.Parallel()

	stamp := time.Date(2026, 11, 13, 21, 30, 12, 0, time.UTC)
	sealed, err := review.SealRefutation(review.Refutation{
		Refuter:     "agent",
		At:          "2026-11-13T00:00:00Z",
		Refutations: []review.Judgment{{ID: "F1", Verdict: review.Refuted, Why: "el diff no toca la validacion, ya estaba"}},
	}, blockingFindings(), tree, stamp)
	if err != nil {
		t.Fatalf("SealRefutation: %v", err)
	}
	if sealed.At != "2026-11-13T21:30:12Z" {
		t.Errorf("At = %q, want the instant the tool sealed it", sealed.At)
	}
	if sealed.DeclaredAt != "2026-11-13T00:00:00Z" {
		t.Errorf("DeclaredAt = %q, want the time the refuter wrote", sealed.DeclaredAt)
	}
}

// TestSealRefutationKeepsNoClaimWhenThereIsNothingToDisagreeWith: a refutation
// that declares no time, or the instant the tool stamped, gains no claim. Nothing
// parses is refused the same way a verdict refuses it (0291).
func TestSealRefutationKeepsNoClaimWhenThereIsNothingToDisagreeWith(t *testing.T) {
	t.Parallel()

	stamp := time.Date(2026, 11, 13, 21, 30, 12, 0, time.UTC)
	for name, at := range map[string]string{
		"none":  "",
		"equal": "2026-11-13T21:30:12Z",
	} {
		sealed, err := review.SealRefutation(review.Refutation{Refuter: "agent", At: at}, blockingFindings(), tree, stamp)
		if err != nil {
			t.Fatalf("%s: SealRefutation: %v", name, err)
		}
		if sealed.DeclaredAt != "" {
			t.Errorf("%s: DeclaredAt = %q, want no claim", name, sealed.DeclaredAt)
		}
		if sealed.At != "2026-11-13T21:30:12Z" {
			t.Errorf("%s: At = %q, want the stamp", name, sealed.At)
		}
	}
}

// TestARefutationWithAClaimRendersAndParses: the claim travels in the file, so the
// next reader sees the disagreement the seal recorded instead of one time it
// cannot attribute.
func TestARefutationWithAClaimRendersAndParses(t *testing.T) {
	t.Parallel()

	sealed, err := review.SealRefutation(review.Refutation{Refuter: "agent", At: "2026-11-13T00:00:00Z"},
		blockingFindings(), tree, time.Date(2026, 11, 13, 21, 30, 12, 0, time.UTC))
	if err != nil {
		t.Fatalf("SealRefutation: %v", err)
	}
	parsed, err := review.ParseRefutation(review.RenderRefutation(sealed))
	if err != nil {
		t.Fatalf("ParseRefutation: %v", err)
	}
	if parsed.At != "2026-11-13T21:30:12Z" {
		t.Errorf("At = %q, want the stamp", parsed.At)
	}
	if parsed.DeclaredAt != "2026-11-13T00:00:00Z" {
		t.Errorf("DeclaredAt = %q after a round trip, want the claim", parsed.DeclaredAt)
	}
}

// TestARefutationWrittenTodayCarriesItsSchema: the refutation was the fourth
// strict decoder and the only one with no number to announce its shape. Without
// one, the next field it gains reads as an unknown key in every older binary, and
// the words send the human to fix a file that is not broken (0291).
func TestARefutationWrittenTodayCarriesItsSchema(t *testing.T) {
	sealed, err := review.SealRefutation(review.Refutation{Refuter: "agent"}, blockingFindings(), tree, now)
	if err != nil {
		t.Fatalf("SealRefutation: %v", err)
	}

	rendered := string(review.RenderRefutation(sealed))
	if !strings.Contains(rendered, `"schema": `+strconv.Itoa(schema.Current)) {
		t.Fatalf("the stored refutation does not carry schema %d:\n%s", schema.Current, rendered)
	}

	back, err := review.ParseRefutation(review.RenderRefutation(sealed))
	if err != nil {
		t.Fatalf("ParseRefutation: %v", err)
	}
	if back.Schema != schema.Current {
		t.Errorf("Schema = %d, want %d", back.Schema, schema.Current)
	}
}

// TestARefutationFromANewerOdachiAsksForAnUpgrade: the strict decoder meets a
// field it does not know and a version above its own, and must name the version
// first. Naming the field sends the human to fix a file that is not broken.
func TestARefutationFromANewerOdachiAsksForAnUpgrade(t *testing.T) {
	future := schema.Current + 1
	text := `{"schema":` + strconv.Itoa(future) + `,"tree":"` + tree +
		`","at":"2026-11-13T00:00:00Z","refutations":[],"a_field_nobody_knows":true}`

	_, err := review.ParseRefutation([]byte(text))
	if err == nil {
		t.Fatalf("ParseRefutation(schema %d) = no error, want a refusal", future)
	}
	message := err.Error()
	for _, want := range []string{"upgrade", "schema " + strconv.Itoa(future), "schema " + strconv.Itoa(schema.Current)} {
		if !strings.Contains(message, want) {
			t.Errorf("the refusal %q does not name %q", message, want)
		}
	}
	if strings.Contains(message, "unknown") {
		t.Errorf("the refusal %q names an unknown key, which sends the human to fix a file that is not broken", message)
	}
}

// TestARefutationStoredBeforeTheNumberStillReadsAndStillLifts: the oldest files
// carry no schema and no claim, and an absent number is the first schema forever.
// A refutation written before 0291 keeps the power it had (0291).
func TestARefutationStoredBeforeTheNumberStillReadsAndStillLifts(t *testing.T) {
	text := `{"tree":"` + tree + `","refuter":"agent","at":"2026-11-13T00:00:00Z","refutations":[` +
		`{"id":"F1","verdict":"refuted","why":"el diff no toca la validacion, ya estaba en la linea 12"}]}`

	parsed, err := review.ParseRefutation([]byte(text))
	if err != nil {
		t.Fatalf("ParseRefutation: %v", err)
	}
	if parsed.Schema != schema.First {
		t.Errorf("Schema = %d, want the first schema %d", parsed.Schema, schema.First)
	}
	if parsed.DeclaredAt != "" {
		t.Errorf("DeclaredAt = %q, want nothing: the file never had the field", parsed.DeclaredAt)
	}
	if _, blocking := review.BlockingAfter(blockingFindings(), review.Answers{parsed}); !blocking {
		t.Error("a refutation stored before the number stopped lifting its finding")
	}
}

// TestASecondSealKeepsTheClaimItsFirstSealRead: attaching a refutation this tool
// stored before seals it again, and the `at` it carries is the stamp the tool wrote,
// not the refuter's word. Reading it as a claim would print "the refuter wrote" a
// time the clone wrote and would drop the claim the first seal read (0291).
func TestASecondSealKeepsTheClaimItsFirstSealRead(t *testing.T) {
	t.Parallel()

	stored, err := review.SealRefutation(review.Refutation{Refuter: "agent", At: "2026-11-13T00:00:00Z"},
		blockingFindings(), tree, time.Date(2026, 11, 13, 21, 30, 12, 0, time.UTC))
	if err != nil {
		t.Fatalf("SealRefutation: %v", err)
	}

	again, err := review.SealRefutation(stored, blockingFindings(), tree, time.Date(2026, 11, 13, 22, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SealRefutation of a stored refutation: %v", err)
	}
	if again.DeclaredAt != "2026-11-13T00:00:00Z" {
		t.Errorf("DeclaredAt = %q, want the claim the first seal read", again.DeclaredAt)
	}
	if again.At != "2026-11-13T22:00:00Z" {
		t.Errorf("At = %q, want this seal's stamp", again.At)
	}
}

// TestASecondSealInTheSameSecondIsTheSameRecord: the guard that keeps a claim must
// not depend on the clock. Two seals in one second read the same input, and one
// record comes out (0291).
func TestASecondSealInTheSameSecondIsTheSameRecord(t *testing.T) {
	t.Parallel()

	stamp := time.Date(2026, 11, 13, 21, 30, 12, 0, time.UTC)
	stored, err := review.SealRefutation(review.Refutation{Refuter: "agent", At: "2026-11-13T00:00:00Z"},
		blockingFindings(), tree, stamp)
	if err != nil {
		t.Fatalf("SealRefutation: %v", err)
	}

	again, err := review.SealRefutation(stored, blockingFindings(), tree, stamp)
	if err != nil {
		t.Fatalf("SealRefutation of a stored refutation: %v", err)
	}
	if again.DeclaredAt != "2026-11-13T00:00:00Z" {
		t.Errorf("DeclaredAt = %q, want the claim, not a second reading of the clock", again.DeclaredAt)
	}
}
