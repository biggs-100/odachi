package rules_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// linesOf splits a text into the lines a model reads.
func linesOf(text string) []string {
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// The block is the only text odachi writes that an agent obeys, so it is pinned
// byte for byte: any edit to the voice or to a rule has to pass through this
// test. It is the template that init writes whole and that `level --to`
// rewrites one section of, so the bytes here are the contract both commands
// share.
func TestBlockForNovatoIsPinnedByteForByte(t *testing.T) {
	want := []string{
		"<!-- odachi:begin -->",
		"# Project rules (odachi)",
		"",
		"odachi is a binary and four plain-text artifacts —brief, task doc, proof and",
		"profile—. The binary enforces the gate; the artifacts are the memory. The human",
		"decides the vision and what is non-negotiable; the agent executes with craft. If",
		"you are unsure where the context lives: `brief.md` for the vision, `odd/tasks/` for the plan,",
		"the skills for the craft: `.agents/skills/`, this project's copy.",
		"",
		"Four hard rules. They are not advice: `odachi check` enforces them.",
		"",
		"1. Brief before code. If `brief.md` is not complete, run `odachi start` first.",
		"   Without a written vision, nothing is built.",
		"2. Nothing outside the fence. If something is in \"Not included\", do not touch it:",
		"   ask with `odachi ask \"<question>\"`. `check` enforces the fence.",
		"3. Do not say \"done\" without proof. Done is `odachi check --run \"<command>\"` green,",
		"   with the output saved in `proof/`. Without that, it is not done.",
		"4. Small pieces, one commit each. One task-doc step, one commit.",
		"",
		"One more, about folders. This project is this repository. Another folder is another",
		"project —its own brief, its own fence, its own store, its own memory—, and odachi is",
		"never read or run there, and its files are never touched, unless the human asks for it.",
		"",
		"And one about asking, because the human is in the loop. A question is answered, not",
		"implemented: an answer is not a proposal, and work starts when the human asks for it or",
		"approves one. An irreversible action —push, pull request, merge, tag or release, an",
		"install, another folder— names its exact target and waits for a yes. Ask at the gate,",
		"once: a question that would not change what you do next is noise.",
		"",
		"Skills: the ones in `.agents/skills/`; inside this project, load `drive` before acting: it is the manual.",
		"",
		"## The six verbs",
		"",
		"- `odachi start` — the interview: writes brief, profile, fence and task doc.",
		"- `odachi ask \"<question>\"` — records an open question; with `--answer` it answers it.",
		"- `odachi level` — shows your facts and proposes senior when earned; with `--to <profile>` it applies the profile and this voice.",
		"- `odachi summary` — what is done in each feature.",
		"- `odachi check` — is it ready? brief, fence, open questions and proof; with `--run` it runs the proof.",
		"- `odachi init` — writes this block, the hook, the skills and the profile voice.",
		"",
		"## How you speak",
		"",
		"The goal is not only that it works: it is that the human learns the craft.",
		"These rules are the difference between an agent that executes and one that teaches.",
		"You are the tutor, not the one who does it.",
		"",
		"- Explain the why of every decision in ONE line, at the moment you make it.",
		"- Record every craft decision: odachi ask --craft \"<decision>\" --because \"<reason>\".",
		"- When you stop something, say why in one line: that is the moment that teaches most.",
		"- One in five decisions, ask back \"why did you choose this?\".",
		"- Never a sermon, never a glossary, never two explanations in a row.",
		"- When a feature closes, run odachi summary and tell in three lines what the human",
		"  decided and what concept appeared.",
		"- Speak in the human's language; the artifacts are English.",
		"<!-- odachi:end -->",
	}
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got := linesOf(block)
	if len(got) != len(want) {
		t.Fatalf("the block has %d lines, want %d:\n%s", len(got), len(want), block)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d of the block:\n got: %q\nwant: %q", i+1, got[i], want[i])
		}
	}
}

// One template, two voices: the block for senior and the block for novato differ
// only in the body of the last section. Anything else drifting means the rules
// depend on who is reading, which is how the two copies odachi used to write
// contradicted each other.
func TestTheTwoVoicesDifferOnlyInTheVoiceSection(t *testing.T) {
	novato, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}

	if novato == senior {
		t.Fatalf("both voices produce the same block")
	}

	// Everything up to and including the heading is the template, and it is the
	// same bytes for both.
	head := strings.Index(novato, rules.Section) + len(rules.Section)
	if head < len(rules.Section) {
		t.Fatalf("the block has no %q section", rules.Section)
	}
	if novato[:head] != senior[:head] {
		t.Fatalf("the block changes before the voice section:\nnovato: %q\nsenior: %q", novato[:head], senior[:head])
	}

	novatoBody, ok := rules.VoiceIn(novato)
	if !ok {
		t.Fatalf("the voice section cannot be read:\n%s", novato)
	}
	seniorBody, ok := rules.VoiceIn(senior)
	if !ok {
		t.Fatalf("the voice section cannot be read:\n%s", senior)
	}
	wantNovato, _ := persona.Voice(persona.Novato)
	wantSenior, _ := persona.Voice(persona.Senior)
	if novatoBody != strings.TrimSpace(wantNovato) {
		t.Errorf("the novato voice of the block is not the persona's:\n%q\nvs\n%q", novatoBody, wantNovato)
	}
	if seniorBody != strings.TrimSpace(wantSenior) {
		t.Errorf("the senior voice of the block is not the persona's:\n%q\nvs\n%q", seniorBody, wantSenior)
	}
	if strings.Contains(seniorBody, "You are the tutor") {
		t.Errorf("the senior voice repeats the tutor's:\n%s", seniorBody)
	}
}

// ci is a profile, not a person: an agent running in CI has nobody to teach, so
// the block it reads carries the senior voice.
func TestCIUsesTheSeniorVoice(t *testing.T) {
	ci, err := rules.Block("ci")
	if err != nil {
		t.Fatalf("Block(ci): %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if ci != senior {
		t.Errorf("the ci block is not the senior one:\n%s", ci)
	}
}

// TestUnknownVoiceIsRefused: nothing is guessed for a name that is not a voice.
func TestUnknownVoiceIsRefused(t *testing.T) {
	if _, err := rules.Block("principiante"); err == nil {
		t.Fatalf("Block(principiante) returned a block")
	} else {
		for _, want := range []string{"principiante", persona.Novato, persona.Senior} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the error does not name %q: %v", want, err)
			}
		}
	}
}

// The first paragraph answers what the agent is working with and who decides
// what, before it is handed the rules. Without it the block talks to an agent
// that already knows.
func TestBlockSaysWhatOdachiIsAndWhereTheVoiceComesFrom(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	flat := strings.Join(strings.Fields(block), " ")
	for _, want := range []string{
		"odachi is a binary and four plain-text artifacts",
		"The binary enforces the gate",
		"the artifacts are the memory",
		"The human decides the vision and what is non-negotiable",
		"the agent executes with craft",
		"`brief.md` for the vision",
		"`odd/tasks/` for the plan",
		"the skills for the craft: `.agents/skills/`, this project's copy.",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("the block does not say %q:\n%s", want, block)
		}
	}

	// The profile file is read by the binary only; that contract is not the
	// block's to publish.
	if strings.Contains(block, "odd/profile.toml") {
		t.Errorf("the block sends the agent to the profile file:\n%s", block)
	}
}

// The voice no longer asks the agent to detect a level: the binary writes the
// voice it read from the profile, and an agent that asks again would be guessing
// what a file already decided.
func TestBlockDoesNotAskTheAgentToGuessTheVoice(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}
	for _, unwanted := range []string{
		"No lo adivines",
		"## Si la persona es novata",
		"¿preferís que te explique el porqué o que vaya al hueso?",
	} {
		if strings.Contains(block, unwanted) {
			t.Errorf("the block still asks the agent to guess the level (%q):\n%s", unwanted, block)
		}
	}

	// The voice is the last section, and it says so where the agent looks for it.
	headings := []string{}
	for _, line := range linesOf(block) {
		if strings.HasPrefix(line, "#") {
			headings = append(headings, line)
		}
	}
	if len(headings) == 0 || headings[len(headings)-1] != rules.Section {
		t.Errorf("the voice must be the block's last section, headings: %v", headings)
	}
}

// TestWithBlockReplacesInPlaceAndKeepsTheRest: init owns the block and nothing
// outside it. A human's lines above and below survive every run.
func TestWithBlockReplacesInPlaceAndKeepsTheRest(t *testing.T) {
	block, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got, err := rules.WithBlock("# mío\n\n"+block+"\nmás mío\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if got != "# mío\n\n"+senior+"\nmás mío\n" {
		t.Errorf("the block was not replaced in place:\n%q", got)
	}

	// A file with no block gains one at the end, after one blank line.
	appended, err := rules.WithBlock("# mío\n", persona.Novato)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	if appended != "# mío\n\n"+block {
		t.Errorf("the block was not appended at the end:\n%q", appended)
	}

	// A truncated block is repaired instead of duplicated.
	truncated, err := rules.WithBlock("antes\n<!-- odachi:begin -->\n# Reglas viejas\n", persona.Novato)
	if err != nil {
		t.Fatalf("WithBlock: %v", err)
	}
	if truncated != "antes\n"+block {
		t.Errorf("the truncated block was not repaired:\n%q", truncated)
	}
}

// TestWithVoiceRewritesOnlyTheSection: what `level --to` owns. Every other byte
// of the file is the human's and survives, including text the agent wrote
// outside the block.
func TestWithVoiceRewritesOnlyTheSection(t *testing.T) {
	novato, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block(novato): %v", err)
	}

	got, changed, err := rules.WithVoice("mío arriba\n\n"+novato+"\nmío abajo\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if !changed {
		t.Fatalf("WithVoice said nothing changed")
	}
	senior, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	if got != "mío arriba\n\n"+senior+"\nmío abajo\n" {
		t.Errorf("only one section changed and more changed:\n%q", got)
	}

	// Identical bytes are never a change, so a rerun writes nothing.
	again, changed, err := rules.WithVoice(got, persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if changed || again != got {
		t.Errorf("WithVoice rewrote an identical voice")
	}
}

// A block written before the voice moved into it has no voice section. It gains
// one as the last section, and every byte that was already there stays.
func TestWithVoiceAddsTheSectionToAnOldBlock(t *testing.T) {
	old := "<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas viejas\n<!-- odachi:end -->\n"

	got, changed, err := rules.WithVoice("mío\n\n"+old, persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if !changed {
		t.Fatalf("WithVoice did not add the section")
	}

	want := "mío\n\n<!-- odachi:begin -->\n# Reglas del proyecto (odachi)\n\nreglas viejas\n\n" +
		rules.Section + "\n\n" + mustVoice(t, persona.Senior) + "<!-- odachi:end -->\n"
	if got != want {
		t.Errorf("the section was not added at the end of the block:\n got %q\nwant %q", got, want)
	}

	// A file with no block is not level's to invent.
	untouched, changed, err := rules.WithVoice("solo mío\n", persona.Senior)
	if err != nil {
		t.Fatalf("WithVoice: %v", err)
	}
	if changed || untouched != "solo mío\n" {
		t.Errorf("WithVoice wrote to a file without a block: %q", untouched)
	}
	if !rules.HasBlock(old) {
		t.Errorf("HasBlock did not recognize a complete block")
	}
	if rules.HasBlock("solo mío\n") {
		t.Errorf("HasBlock said yes to a file without a block")
	}
	if _, ok := rules.VoiceIn(old); ok {
		t.Errorf("VoiceIn invented a voice for a block without a section")
	}
}

// TestBodyIsTheBlockWithoutItsMarkers: the native agent mirrors exactly what the
// agent reads, and the markers are plumbing for in-place edits, not rules.
func TestBodyIsTheBlockWithoutItsMarkers(t *testing.T) {
	block, err := rules.Block(persona.Senior)
	if err != nil {
		t.Fatalf("Block(senior): %v", err)
	}
	body, err := rules.Body(persona.Senior)
	if err != nil {
		t.Fatalf("Body(senior): %v", err)
	}

	if strings.Contains(body, rules.Begin) || strings.Contains(body, rules.End) {
		t.Errorf("the block body carries the markers:\n%s", body)
	}
	if strings.HasPrefix(body, "\n") {
		t.Errorf("the body starts with a newline: %q", body[:10])
	}
	if !strings.HasSuffix(body, "\n") || strings.HasSuffix(body, "\n\n") {
		t.Errorf("the body does not end in exactly one newline: %q", body[len(body)-10:])
	}
	want := rules.Begin + "\n" + body + rules.End + "\n"
	if block != want {
		t.Errorf("Block is not Begin + Body + End:\n%q\nvs\n%q", block, want)
	}

	got, ok := rules.BodyIn(block)
	if !ok {
		t.Fatalf("BodyIn did not recognize the block")
	}
	if got != body {
		t.Errorf("BodyIn does not return the body:\n%q\nvs\n%q", got, body)
	}
	for _, content := range []string{"", "sin bloque\n", rules.Begin + "\ntruncado\n"} {
		if _, ok := rules.BodyIn(content); ok {
			t.Errorf("BodyIn invented a body for %q", content)
		}
	}
}

// mustVoice is the voice text, or a failed test: the template tests compare
// against the persona text, so a missing one is not a test to skip.
func mustVoice(t *testing.T, name string) string {
	t.Helper()
	voice, err := persona.Voice(name)
	if err != nil {
		t.Fatalf("persona.Voice(%q): %v", name, err)
	}
	return voice
}

// TestACRLFBlockReadsLikeAnLFBlock: a checkout is not an edit. A Windows clone with
// core.autocrlf=true gives every line a CR, and the readers parse by line; folding the
// endings once, here, keeps doctor from calling a fine block hand-edited - the repair it
// names is for a different problem.
func TestACRLFBlockReadsLikeAnLFBlock(t *testing.T) {
	lf, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("Block: %v", err)
	}
	crlf := strings.ReplaceAll(lf, "\n", "\r\n")

	wantBody, ok := rules.BodyIn(lf)
	if !ok {
		t.Fatal("the LF block has no body")
	}
	gotBody, ok := rules.BodyIn(crlf)
	if !ok {
		t.Fatal("the CRLF block has no body")
	}
	if gotBody != wantBody {
		t.Errorf("BodyIn differs by line endings:\n%q\nvs\n%q", gotBody, wantBody)
	}

	if !rules.HasBlock(crlf) {
		t.Error("HasBlock says the CRLF block has no block")
	}
	wantVoice, ok := rules.VoiceIn(lf)
	if !ok {
		t.Fatal("the LF block has no voice")
	}
	gotVoice, ok := rules.VoiceIn(crlf)
	if !ok {
		t.Fatalf("VoiceIn does not read the CRLF block's voice")
	}
	if gotVoice != wantVoice {
		t.Errorf("VoiceIn = %q, want %q", gotVoice, wantVoice)
	}
}

// TestTheBlockNamesTheFolderTheScopeUses: the block names one folder, the one the
// run resolved. Naming both and letting the reader choose was tried and it lies
// when a stale copy survives; naming one for the wrong scope lies too. Each scope
// names its own, and neither names the other.
func TestTheBlockNamesTheFolderTheScopeUses(t *testing.T) {
	cases := map[string]struct{ want, not string }{
		persona.SkillsProject: {want: "`.agents/skills/`", not: "`~/.agents/skills/`"},
		persona.SkillsHome:    {want: "`~/.agents/skills/`", not: "`.agents/skills/`"},
	}
	for scope, c := range cases {
		body, err := rules.BodyFor(persona.Senior, scope)
		if err != nil {
			t.Fatalf("BodyFor(%s): %v", scope, err)
		}
		if !strings.Contains(body, c.want) {
			t.Errorf("the block for %s does not name %q:\n%s", scope, c.want, body)
		}
		if strings.Contains(body, c.not) {
			t.Errorf("the block for %s names the other scope's folder %q:\n%s", scope, c.not, body)
		}
	}
}

// TestBlockForHomeIsPinnedByteForByte: the home scope is what every project that
// adopts odachi now gets in AGENTS.md, so it is pinned the same way as the
// project one. A body that changes by a space is a different block to the doctor
// and a different sentence to the agent.
func TestBlockForHomeIsPinnedByteForByte(t *testing.T) {
	block, err := rules.BlockFor(persona.Novato, persona.SkillsHome)
	if err != nil {
		t.Fatalf("BlockFor(home): %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	if len(lines) == 0 || lines[0] != "<!-- odachi:begin -->" {
		t.Fatalf("the block does not open with the marker: %v", lines)
	}
	flat := strings.Join(strings.Fields(block), " ")
	for _, want := range []string{
		"the skills for the craft: `~/.agents/skills/`, the copy this machine keeps.",
		"Skills: the ones in `~/.agents/skills/`, the copy this machine keeps; inside this project, load `drive` before acting: it is the manual.",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("the home block does not say %q:\n%s", want, block)
		}
	}
	if strings.Contains(block, "`.agents/skills/`") {
		t.Errorf("the home block names the project's copy:\n%s", block)
	}
}
