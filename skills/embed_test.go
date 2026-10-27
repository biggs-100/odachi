package skills_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/skills"
)

// bundled is every skill directory in this package's directory, read from disk
// instead of listed here: adding a skill must not require editing this file.
var bundled = func() []string {
	entries, err := os.ReadDir(".")
	if err != nil {
		panic("read skills directory: " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return names
}()

// TestAllCarriesEveryStarterSkill: the binary knows them by name, sorted, so
// `pack` writes the same set on every machine and in the same order.
func TestAllCarriesEveryStarterSkill(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.Name)
		if len(s.Content) == 0 {
			t.Errorf("%s came back empty", s.Name)
		}
	}
	if !slices.Equal(names, bundled) {
		t.Fatalf("All() = %v, want %v", names, bundled)
	}
	if got := skills.Names(); !slices.Equal(got, bundled) {
		t.Errorf("Names() = %v, want %v", got, bundled)
	}
}

// TestEmbeddedBytesAreTheFilesInTheRepo: the embed is the source, not a copy of
// it. If a skill is edited on disk and the bytes here do not match, the binary
// is shipping a version nobody reviewed.
func TestEmbeddedBytesAreTheFilesInTheRepo(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	for _, s := range all {
		onDisk, err := os.ReadFile(s.Name + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s/SKILL.md: %v", s.Name, err)
		}
		if string(s.Content) != string(onDisk) {
			t.Errorf("%s: the embedded bytes differ from %s/SKILL.md", s.Name, s.Name)
		}
	}
}

// TestDelegationPolicyLivesInSlicesAndDrive: the policy that decides when a
// subagent is used is content, and content drifts. `slices` owns the triggers and
// the handoff; `drive` is what the block loads before acting, so it is the one
// that has to send the agent to `slices` when the request does not fit.
func TestDelegationPolicyLivesInSlicesAndDrive(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	for _, want := range []string{
		"this one wins",
		"Never for size",
		"task doc path",
		"edit surfaces",
		"Skills to load before work",
		"odachi-worker",
		"odachi-verify",
		"=2k tokens",
		"Never delegate the vision decision",
	} {
		if !strings.Contains(byName["slices"], want) {
			t.Errorf("slices no dice %q: la política de delegación se perdió", want)
		}
	}
	if !strings.Contains(byName["drive"], "load `slices` before splitting") {
		t.Errorf("drive no manda a cargar slices cuando el pedido no entra en una lectura")
	}
}

// TestShapeTeachesExplorationBeforeTheBrief: the front of the cycle is a skill,
// not a verb. It explores before the brief, ends in `start --brief`, keeps the
// brief in the human's hands, and `drive` is what loads it when the request does
// not fit in one reading.
func TestShapeTeachesExplorationBeforeTheBrief(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["shape"]
	if !ok {
		t.Fatal("el payload no tiene la skill shape: el frente del ciclo no existe")
	}
	for _, want := range []string{
		"Explore before you declare",
		"odachi start --brief",
		"Exploration is not a verb",
		"the human owns",
		"odachi ask",
		"the least information",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("shape no dice %q: explorar antes de declarar se perdio", want)
		}
	}
	if !strings.Contains(byName["drive"], "load `shape`") {
		t.Errorf("drive no manda a cargar shape cuando el pedido no entra en una lectura")
	}
}

// TestCommitsTeachesRiskByBlastRadius: the six Risk: names are a vocabulary, and
// a vocabulary without consequences is decoration. Each one names its lens and
// its verifier, and none of them buys a waiver.
func TestCommitsTeachesRiskByBlastRadius(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["commits"]
	if !ok {
		t.Fatal("el payload no tiene la skill commits: el riesgo no tiene vocabulario")
	}
	if !strings.Contains(content, "Risk by blast radius") {
		t.Error("commits no tiene la tabla de riesgo por radio de impacto")
	}
	for _, item := range []string{"data", "security", "contract", "concurrency", "delivery", "no-test"} {
		line := ""
		for _, candidate := range strings.Split(content, "\n") {
			if strings.HasPrefix(candidate, "| "+item+" |") {
				line = candidate
				break
			}
		}
		if line == "" {
			t.Errorf("commits no tiene la fila %q: el item no dice que fuerza", item)
			continue
		}
		if !strings.Contains(line, "| yes |") {
			t.Errorf("la fila %q no pide verifier: %s", item, line)
		}
	}
	for _, want := range []string{"odachi review --lens", "odachi-verify", "waiver"} {
		if !strings.Contains(content, want) {
			t.Errorf("commits no dice %q: la ceremonia se perdio", want)
		}
	}
}

// TestDriveTeachesTheRequirementChangeRule: the plan grows and the vision does
// not. A requirement the human changes goes to `ask --change`, which reopens the
// steps that claimed it, and the brief only moves when the vision does.
func TestDriveTeachesTheRequirementChangeRule(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	for _, want := range []string{
		"odachi ask --change",
		"--req R1",
		"steps reopen",
		"`brief.md` does not move",
	} {
		if !strings.Contains(byName["drive"], want) {
			t.Errorf("drive no dice %q: la regla del cambio de requisito se perdio", want)
		}
	}
}

// TestDriveTeachesTheWorkflowDiscipline: ODD-lite is content, and content
// drifts. `drive` names the spec rule, the RED before GREEN rule, the delegation
// record and the risk line; `commits` owns the risk vocabulary.
func TestDriveTeachesTheWorkflowDiscipline(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	for _, want := range []string{
		"Requirements quoted **verbatim**",
		"observe RED",
		"delegated step",
		"Risk:",
	} {
		if !strings.Contains(byName["drive"], want) {
			t.Errorf("drive no dice %q: ODD-lite se perdió", want)
		}
	}
	for _, want := range []string{
		"Risk: none",
		"data|security|contract|concurrency|delivery|no-test",
	} {
		if !strings.Contains(byName["commits"], want) {
			t.Errorf("commits no dice %q: la línea de riesgo se perdió", want)
		}
	}
}

// TestTriageSkillTeachesRootCause: the skill that closes the backlog gap must
// classify by root, not by issue, reproduce before fixing, and send an
// out-of-fence report to ask instead of patching it.
func TestTriageSkillTeachesRootCause(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["triage"]
	if !ok {
		t.Fatal("el payload no tiene la skill triage: el backlog sigue sin método")
	}
	for _, want := range []string{
		"by **root**",
		"one fix",
		"Reproduce the symptom",
		"odachi ask",
		"odachi plan",
		"failing test",
		"outside the fence",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("triage no dice %q: el método de raíz se perdió", want)
		}
	}
}

// TestPullRequestSkillTeachesReviewSlices: the delivery gap must keep one work
// unit per PR, a 400-line budget, one chain strategy, and the tree's proof
// before review.
func TestPullRequestSkillTeachesReviewSlices(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["pull-request"]
	if !ok {
		t.Fatal("el payload no tiene la skill pull-request: la entrega sigue sin corte")
	}
	for _, want := range []string{
		"400 changed lines",
		"one work unit",
		"stacked",
		"feature branch chain",
		"Closes/Fixes/Resolves",
		"odachi check --run",
		"left out",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("pr no dice %q: el corte de entrega se perdió", want)
		}
	}
}

// TestGoTestsSkillTeachesGoPatterns: the Go skill chooses the concrete pattern
// without repeating the language-agnostic discipline owned by tests.
func TestGoTestsSkillTeachesGoPatterns(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["go-tests"]
	if !ok {
		t.Fatal("el payload no tiene la skill go-tests: Go sigue sin patrón")
	}
	for _, want := range []string{
		"Load `tests` first",
		"Table for several cases",
		"t.Run",
		"t.TempDir",
		"testing.Short",
		"teatest",
		"golden",
		"focal package",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("go-tests no dice %q: el patrón de Go se perdió", want)
		}
	}
}

// TestSkillImproveTeachesTheAudit: the skill that keeps the catalog healthy must
// start at the trigger, cut at 60 lines, keep one owner per rule, and revalidate
// the generated registry.
func TestSkillImproveTeachesTheAudit(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["skill-improve"]
	if !ok {
		t.Fatal("el payload no tiene la skill skill-improve: el catálogo no se audita")
	}
	for _, want := range []string{
		"Start with the trigger",
		"60 lines",
		"observable",
		"odachi skills --check",
		"SKILLS.md",
		"one owner",
		"before and after",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("skill-improve no dice %q: la auditoría se perdió", want)
		}
	}
}

// TestBootstrapIsTheEntryPoint: the payload has exactly one skill that teaches
// how to start a project, and it is the one `odachi init` installs in the
// human's home. Splitting it out of the curriculum is the whole reason the
// helper exists: a project carries craft, the home carries how to arrive.
func TestBootstrapIsTheEntryPoint(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatalf("el payload no tiene la skill %q: un agente no puede aprender a arrancar", skills.BootstrapName)
	}
	if bootstrap.Name != skills.BootstrapName {
		t.Errorf("Bootstrap() = %q, want %q", bootstrap.Name, skills.BootstrapName)
	}
	if len(bootstrap.Content) == 0 {
		t.Error("la skill de arranque vino vacía")
	}
}

// TestCurriculumIsEverySkillButTheEntryPoint: the project copy is the craft and
// nothing else, so init never installs the same two skills in both roots.
func TestCurriculumIsEverySkillButTheEntryPoint(t *testing.T) {
	curriculum, err := skills.Curriculum()
	if err != nil {
		t.Fatalf("Curriculum(): %v", err)
	}

	want := make([]string, 0, len(bundled))
	for _, name := range bundled {
		if name != skills.BootstrapName {
			want = append(want, name)
		}
	}
	got := make([]string, 0, len(curriculum))
	for _, s := range curriculum {
		got = append(got, s.Name)
		if s.Name == skills.BootstrapName {
			t.Errorf("el currículum del proyecto incluye la skill de arranque")
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Curriculum() = %v, want %v", got, want)
	}
}

// TestBootstrapSaysTheAgentRunsTheCommands: the human installs odachi and talks.
// The skill is what the agent reads, so it says who runs the three commands.
func TestBootstrapSaysTheAgentRunsTheCommands(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatal("the payload has no bootstrap skill")
	}
	if !strings.Contains(string(bootstrap.Content), "You run them; the human never types them") {
		t.Errorf("the bootstrap skill does not say the agent runs the commands:\n%s", bootstrap.Content)
	}
}

// TestRoleCraftTeachesTheInterviewAndTheResearch: a role is written from an
// interview and evidence, not from adjectives. The skill names the questions, the
// sources, the probe and the command that writes the source.
func TestRoleCraftTeachesTheInterviewAndTheResearch(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Content)
	}

	content, ok := byName["role-craft"]
	if !ok {
		t.Fatal("el payload no tiene la skill role-craft: un profesional no se puede crear")
	}
	for _, want := range []string{
		"Ask for the scars",
		"Years are not a field",
		"three to five sources with a date",
		"Write the probe",
		"The human owns the words",
		"odachi role new <slug>",
		"60 lines",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("role-craft no dice %q: la disciplina de la entrevista se perdio", want)
		}
	}
}
