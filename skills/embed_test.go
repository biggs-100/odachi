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
		if len(s.Main()) == 0 {
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

// TestBootstrapIsTheFirstCycle: the entry point is the checklist a new project
// follows, and it says the things the probe measured: init creates the repository,
// the interview's answers are the human's, and the first proof closes the setup.
func TestBootstrapIsTheFirstCycle(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatal("the binary has no bootstrap skill")
	}

	text := string(bootstrap.Main())
	for _, want := range []string{
		"creates the git repository",
		"odachi start",
		"--outcome",
		"Never invent the interview's answers",
		"odachi check --run",
		"first proof",
		"drive",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the bootstrap does not say %q", want)
		}
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
		if string(s.Main()) != string(onDisk) {
			t.Errorf("%s: the embedded bytes differ from %s/SKILL.md", s.Name, s.Name)
		}
	}
}

// TestDelegationPolicyLivesInSlicesAndDrive: the policy that decides when a
// subagent is used is content, and content drifts. `slices` owns the triggers and
// the handoff; `drive` is what the block loads before acting, so it is the one
// that has to send the agent to `slices` when the request does not fit.
//
// The test asks the folder, not the body: the policy moved to
// references/delegation.md when the body met the cap, and a skill is its folder
// (0205, 0206).
func TestDelegationPolicyLivesInSlicesAndDrive(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
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

// TestTheVerifierRunsFocusedCommands: the handoff names which commands the
// verifier receives. A second full suite beside the gate competes for the machine
// and makes the gate's own measurement wrong (0217), so the request names the
// focused package, not the world.
func TestTheVerifierRunsFocusedCommands(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"focused package",
		"never a second",
		"go test ./...",
		"measurement",
	} {
		if !strings.Contains(byName["slices"], want) {
			t.Errorf("slices no dice %q: el verificador puede volver a correr la suite entera", want)
		}
	}
}

// TestUICraftTeachesBrowseAndCompare: a collection is a product decision, not a
// card grid. The reference names search scope, filters, the loading pattern and
// the return, and the body routes it.
func TestUICraftTeachesBrowseAndCompare(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"browse-and-compare.md",
		"Search",
		"filters",
		"pagination",
		"focus",
	} {
		if !strings.Contains(byName["ui-craft"], want) {
			t.Errorf("ui-craft no dice %q: buscar y comparar no tiene guía", want)
		}
	}
}

// TestDesignChecksEveryArtboard: the delivery skill sends every artboard through
// the UX read before reporting it, treats another approach as another artboard,
// and looks at a comment's screenshot before the element.
func TestDesignChecksEveryArtboard(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"the-read",
		"adaptive-interfaces",
		"next `NN`",
		"screenshot",
		"transient state",
	} {
		if !strings.Contains(byName["design"], want) {
			t.Errorf("design no dice %q: el artboard puede reportarse sin leerlo", want)
		}
	}
}

// TestDesignLetsTheHumanTune: a person tuning values in the canvas needs the
// numbers back as text, not as a screenshot. The artboard carries the sliders and
// the button, and the rule keeps it self-contained.
func TestDesignLetsTheHumanTune(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"sliders",
		"CSS custom properties",
		"Copy as prompt",
		"copies the chosen values",
		"never calls a network",
	} {
		if !strings.Contains(byName["design"], want) {
			t.Errorf("design no dice %q: los valores no vuelven como texto", want)
		}
	}
}

// TestWritingCoversTheReviewPath: the document gives the reviewer a path and a
// command that measures what moved, and the skill loads when a doc feels dense.
func TestWritingCoversTheReviewPath(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"git diff --name-only",
		"out of scope",
		"previous and the next",
		"acceptance criteria",
		"hard to scan",
	} {
		if !strings.Contains(byName["writing"], want) {
			t.Errorf("writing no dice %q: el camino de review se perdió", want)
		}
	}
}

// TestPullRequestMeasuresTheBudget: the 400-line budget is measured with a
// command, not estimated by eye.
func TestPullRequestMeasuresTheBudget(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	for _, want := range []string{
		"gh pr view",
		"additions,deletions,changedFiles",
	} {
		if !strings.Contains(byName["pull-request"], want) {
			t.Errorf("pull-request no dice %q: el presupuesto se estima a ojo", want)
		}
	}
}

// TestVerificationIsProportional: the ceremony a change owes is decided by its
// blast radius, not by habit. `tests` names the three tiers and the two
// comparisons a red owes, and `slices` routes the delegation policy to the same
// ladder: a curriculum that says "always verify" is ignored, and one that says
// "never verify" is dangerous.
func TestVerificationIsProportional(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		var folder strings.Builder
		for _, file := range s.Files {
			folder.Write(file.Content)
		}
		byName[s.Name] = folder.String()
	}

	for _, want := range []string{
		"A passive change",
		"A medium change",
		"A high change",
		"writer's own",
		"odachi-verify",
		"on the base",
		"already failed",
		"comes from the machine",
		"never reported as green",
	} {
		if !strings.Contains(byName["tests"], want) {
			t.Errorf("tests no dice %q: la verificación proporcional se perdió", want)
		}
	}
	for _, want := range []string{"proportional", "passive change", "odachi-verify", "six risks"} {
		if !strings.Contains(byName["slices"], want) {
			t.Errorf("slices no dice %q: la escalera no llega a la política de delegación", want)
		}
	}
}

// TestPublishingIsReadBack: a create command is not done at the URL it prints.
// The delivery owns the command and the fields a PR has to match, the writing
// owns the comparison of any published text, and the two cannot disagree.
func TestPublishingIsReadBack(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	for _, want := range []string{
		"Read the result back",
		"gh pr view",
		"baseRefName",
		"headRefName",
		"before you ask for review",
	} {
		if !strings.Contains(byName["pull-request"], want) {
			t.Errorf("pull-request no dice %q: el readback del PR se perdió", want)
		}
	}
	for _, want := range []string{
		"not a text that arrived",
		"the file you wrote",
		"pull-request",
	} {
		if !strings.Contains(byName["writing"], want) {
			t.Errorf("writing no dice %q: la comparación del texto publicado se perdió", want)
		}
	}
}

// TestWritingShipsTheDocumentTemplate: the document shape was prose; the asset is
// the copyable skeleton that makes it an artifact. The body declares it under
// ## Assets and sends the writer to copy it.
func TestWritingShipsTheDocumentTemplate(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}

	var writing skills.Skill
	for _, s := range all {
		if s.Name == "writing" {
			writing = s
			break
		}
	}
	if writing.Name == "" {
		t.Fatal("the payload has no writing skill")
	}

	var template []byte
	for _, file := range writing.Files {
		if file.Path == "assets/doc.md" {
			template = file.Content
		}
	}
	if len(template) == 0 {
		t.Fatal("writing carries no assets/doc.md: the copyable skeleton is missing")
	}
	for _, want := range []string{"## Quick path", "## Details", "## Checklist", "## Next step"} {
		if !strings.Contains(string(template), want) {
			t.Errorf("the template does not have %q", want)
		}
	}

	body := string(writing.Main())
	for _, want := range []string{"## Assets", "`assets/doc.md`", "copy"} {
		if !strings.Contains(body, want) {
			t.Errorf("writing does not say %q: the asset has no route", want)
		}
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
		byName[s.Name] = string(s.Main())
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
		byName[s.Name] = string(s.Main())
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
		byName[s.Name] = string(s.Main())
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
// drifts. `drive` names the spec rule, the RED before GREEN rule, the route a
// delegated step declares and the risk line; `commits` owns the risk vocabulary.
func TestDriveTeachesTheWorkflowDiscipline(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	for _, want := range []string{
		"Requirements quoted **verbatim**",
		"observe RED",
		"--route delegated",
		"--light",
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
		byName[s.Name] = string(s.Main())
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
		byName[s.Name] = string(s.Main())
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
		byName[s.Name] = string(s.Main())
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
		byName[s.Name] = string(s.Main())
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
	if len(bootstrap.Main()) == 0 {
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
// The skill is what the agent reads, so it says who runs the cycle's commands.
func TestBootstrapSaysTheAgentRunsTheCommands(t *testing.T) {
	bootstrap, ok := skills.Bootstrap()
	if !ok {
		t.Fatal("the payload has no bootstrap skill")
	}
	if !strings.Contains(string(bootstrap.Main()), "You run them; the human never types them") {
		t.Errorf("the bootstrap skill does not say the agent runs the commands:\n%s", bootstrap.Main())
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
		byName[s.Name] = string(s.Main())
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

// TestSkillCraftTeachesExplorationAndResearch: a rule about the outside world
// ages the day the system changes, so the creator explores the code, researches
// the current docs, and puts the version and the date in the rule.
func TestSkillCraftTeachesExplorationAndResearch(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	byName := map[string]string{}
	for _, s := range all {
		byName[s.Name] = string(s.Main())
	}

	content, ok := byName["skill-craft"]
	if !ok {
		t.Fatal("el payload no tiene la skill skill-craft: el creador de skills no existe")
	}
	for _, want := range []string{
		"Explore before writing",
		"version and the date",
		"Research facts, not fashions",
		"odachi skill new",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("skill-craft no dice %q: explorar e investigar se perdio", want)
		}
	}
}
