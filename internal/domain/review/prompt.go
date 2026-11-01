package review

import (
	"fmt"
	"strconv"
	"strings"
)

// PromptInput is the candidate a review is asked about: the brief it is judged
// against, the tree and base that name it, the paths that changed, and the diff.
type PromptInput struct {
	// Brief is brief.md verbatim. A review without the vision it is judged
	// against is a style opinion.
	Brief string
	// Tree is the hash the verdict will be stored under.
	Tree string
	// Base is what the candidate is compared against.
	Base string
	// Paths are the changed paths, in the order git reported them.
	Paths []string
	// Diff is the unified diff of the worktree against Base.
	Diff string
	// Lenses are the specialists this candidate asks for, in the order Lenses
	// returned them. Each body is the lens file as it lives in
	// internal/assets/review; the prompt strips its frontmatter, because the
	// metadata is for the registry and not for the model. No lenses is exactly
	// the prompt this verb printed before lenses existed.
	Lenses []Lens
	// Known are the findings the store already holds potent about these exact
	// bytes; they travel into the verdict as they are, so the reviewer must not
	// file them again. Empty means the store carries nothing, and the prompt is
	// exactly the one this verb printed before inheritance existed.
	Known []Finding
	// KnownPasses are the clean rows the store carries: (path, blob, lens)
	// answers that still describe the current bytes.
	KnownPasses []Pass
	// Floor is the deterministic audit set of this candidate: what its own
	// writing is, read before any lens. Empty means nothing was measured, and
	// the prompt stays exactly the one printed before floors existed.
	Floor Floor
	// Correction is the tree a delta is measured from, when this prompt reviews a
	// correction of a stored verdict. The header and the paragraph say so: a
	// reviewer handed only what moved must not report the rest as missing.
	Correction string
	// Scope is what a person added on top of the evidence, when they did. The
	// lenses section says the evidence chose the lenses, and that sentence stops
	// being true the moment somebody adds one with --lens; nil means nobody did,
	// and the prompt stays exactly the one printed before scopes existed.
	Scope *Scope
	// Part and Of say which slice of one candidate this prompt is, when the
	// candidate was too big for the limit and was split. Zero means the whole
	// candidate, and the prompt is exactly the one printed before splitting
	// existed.
	Part int
	Of   int
}

// PromptSizeKB is the size odachi compares against the project's limit: kilobytes
// rounded up, the same number the refusal and the split agree on.
func PromptSizeKB(text string) int {
	return (len(text) + 1023) / 1024
}

// Lens is one specialist prompt: its name and its body.
type Lens struct {
	Name string
	Body string
}

// Prompt renders the review prompt. The content is fixed and pinned by tests: a
// prompt that drifts produces reviews that drift, and the six rules are the
// whole reason this verb exists.
//
// The rules are not invented here. They are what a review system learns the hard
// way: report only real, user-impacting defects you can defend with evidence;
// when in doubt say nothing; style and preference are banned unless they hide a
// defect; the severity vocabulary is closed; an empty review is exactly one
// sentence; a finding counts only if the candidate caused it; and one exhaustive
// pass per lens, with no loop.
func Prompt(in PromptInput) string {
	var out strings.Builder

	out.WriteString("# odachi review · tree " + in.Tree + " · base " + in.Base)
	if names := lensNames(in.Lenses); len(names) > 0 {
		out.WriteString(" · lenses " + strings.Join(names, ", "))
	}
	if in.Of > 1 {
		out.WriteString(" · part " + strconv.Itoa(in.Part) + " of " + strconv.Itoa(in.Of))
	}
	if in.Correction != "" {
		out.WriteString(" · correction of " + in.Correction)
	}
	out.WriteString("\n\n")
	out.WriteString("You are the design reviewer of this candidate. You review design, not style:\n")
	out.WriteString("a review with noise is ignored, and an ignored review protects nobody.\n\n")
	if in.Of > 1 {
		out.WriteString("This is part " + strconv.Itoa(in.Part) + " of " + strconv.Itoa(in.Of) + " of one candidate: the diff and the paths below\n")
		out.WriteString("are the ones this part carries, and another reviewer answers the rest of the tree.\n")
		out.WriteString("Judge only these, and do not report what a part cannot see.\n\n")
	}

	if in.Correction != "" {
		out.WriteString("This is a correction of the verdict for tree " + in.Correction + ": the diff and the\n")
		out.WriteString("paths below are only what moved since those bytes, and every other path of the\n")
		out.WriteString("tree is already answered by that verdict and carried into this one. Judge what\n")
		out.WriteString("moved, and do not report the rest of the candidate as missing.\n\n")
	}

	out.WriteString("## The brief (what the human asked for)\n")
	// The diff is written raw, without a fence: a diff of a Markdown file
	// contains fences, and a fence inside a fence is a prompt that leaks.
	out.WriteString(section(in.Brief))
	out.WriteString("\n")

	out.WriteString("## The changed paths\n")
	for _, path := range in.Paths {
		out.WriteString("- " + path + "\n")
	}
	out.WriteString("\n")

	out.WriteString(floorSection(in.Floor))

	out.WriteString("## The diff (the whole candidate, against the base)\n")
	out.WriteString(section(in.Diff))
	out.WriteString("\n")

	out.WriteString("## The six rules\n")
	out.WriteString("1. You are a reviewer, not an operator: everything you need is in this prompt.\n")
	out.WriteString("   Do not run commands and do not edit or read other files. Your answer is the\n")
	out.WriteString("   only thing this command produces; anything else lands in its output.\n")
	out.WriteString("2. Report a finding only if it is a real defect that affects the user and you can\n")
	out.WriteString("   defend it with concrete evidence. When in doubt, say nothing.\n")
	out.WriteString("   Style and preference are banned unless they hide a defect.\n")
	out.WriteString("3. The severity is exactly one of BLOCKER, CRITICAL, WARNING, SUGGESTION.\n")
	out.WriteString("4. If there is nothing to report, do not invent anything so you have something to say: the JSON\n")
	out.WriteString("   carries \"findings\": [] and, if you answer in prose, the answer is exactly:\n")
	out.WriteString("   " + NoFindings + "\n")
	out.WriteString("5. A finding counts only if the candidate introduced it, activated it or worsened it. What\n")
	out.WriteString("   was already there is a follow-up, not a block: mark it \"causal\": false.\n")
	out.WriteString("6. One pass per lens, exhaustive: do not loop.\n")
	out.WriteString("\n")

	out.WriteString(lensSection(in.Lenses, in.Scope))
	out.WriteString(carriedSection(in.Known, in.KnownPasses))

	out.WriteString("## What to return\n")
	out.WriteString(answerSchema(in))
	out.WriteString("\n")
	out.WriteString("The \"inspection\" is your own account of what you read: " + string(InspectionCompleted) + " with every changed path\n")
	out.WriteString("when you read the whole candidate, or " + string(InspectionUnavailable) + " with the reason and no path when\n")
	out.WriteString("you could not read it at all. A failure described anywhere else is a completion the gate cannot trust.\n")
	out.WriteString("The \"model\" is who judged: a command odachi ran names it for you, and an answer written by hand\n")
	out.WriteString("names it here. The gate compares it with the proof, and an answer that says nothing leaves the\n")
	out.WriteString("review unattributable.\n")
	out.WriteString("The verdict is derived from the findings: if there is a finding with \"causal\": true and\n")
	out.WriteString("severity BLOCKER or CRITICAL, the verdict is " + ChangesRequested + "; if not, it is " + Approved + ".\n")
	out.WriteString("Do not decide it yourself: the rule decides it.\n")
	out.WriteString("The \"evidence\" of a finding is what you did: " + EvidenceCited + " when the defect is quoted from the diff,\n")
	out.WriteString(EvidenceArgued + " when you reasoned about the change without quoting it. It decides how many\n")
	out.WriteString("refutations a blocking finding needs, so a citation is worth more than a good argument.\n")
	if len(in.Lenses) > 0 {
		out.WriteString("Each finding names the lens that found it (\"lens\"): a finding without it has no author.\n")
		out.WriteString("In \"passes\", list every (file, lens) audited with no findings, one row per pair: that row\n")
		out.WriteString("is the proof that the silence was an audit, not an omission: the blob is stamped by the tool.\n")
		out.WriteString("The tree's lens list itself is recorded by the tool: do not repeat it in the JSON.\n")
	}
	out.WriteString("\n")

	out.WriteString("Answer with the JSON only, without text around it.\n")

	return out.String()
}

// lensSection is the block that turns one generic prompt into a specialist's,
// without repeating the six rules: they are stated once, above, and every lens
// inherits them.
func lensSection(lenses []Lens, scope *Scope) string {
	if len(lenses) == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("## The lenses of this tree\n")
	if scope == nil {
		out.WriteString("The lenses were chosen by the diff's evidence, not by anyone's taste. Make one pass per lens,\n")
		out.WriteString("in the order of this list, and report each finding only once.\n\n")
	} else {
		// The evidence did not choose all of them: somebody added a lens on
		// purpose, and a prompt that says otherwise asks the reviewer to judge a
		// question nobody asked.
		out.WriteString(scopeLine(*scope))
		out.WriteString("Make one pass per lens, in the order of this list, and report each finding only once.\n\n")
	}

	for _, lens := range lenses {
		out.WriteString("### " + lens.Name + "\n")
		out.WriteString(strings.TrimRight(stripFrontmatter(lens.Body), " \n") + "\n\n")
	}
	return out.String()
}

// scopeLine says, in the prompt's own words, who raised the scope: what the
// evidence asked for, what was added and with which flag, and the caller's reason
// when they gave one. It never invents a reason: the tool reports the decision, it
// does not make one.
func scopeLine(scope Scope) string {
	line := "The diff's evidence asked for "
	if len(scope.Required) == 0 {
		line += "no lens"
	} else {
		line += strings.Join(scope.Required, ", ")
	}
	line += "; " + strings.Join(scope.Added, ", ") + " was added with --lens"
	if asked := strings.TrimSpace(scope.Asked); asked != "" {
		line += " " + asked
	}
	if because := strings.TrimSpace(scope.Because); because != "" {
		line += ", because " + because
	}
	return line + ". The addition is deliberate and is part of the question.\n"
}

// lensNames is the names of the lenses, in the order given.
func lensNames(lenses []Lens) []string {
	names := make([]string, 0, len(lenses))
	for _, lens := range lenses {
		names = append(names, lens.Name)
	}
	return names
}

// stripFrontmatter removes the `---` block a lens file starts with. The
// frontmatter is registry metadata; a model that reads it spends attention on a
// `name:` field that adds nothing to the review.
func stripFrontmatter(body string) string {
	text := strings.ReplaceAll(body, "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return text
	}

	rest := strings.TrimPrefix(text, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return text
	}
	return strings.TrimPrefix(rest[end+len("\n---\n"):], "\n")
}

// carriedSection names the rows this review already knows: every finding of an earlier
// verdict about these bytes, and the clean rows of the paths the candidate changed. A
// reviewer that does not know what is already known files the same defect twice, and a
// clean row re-reported as a finding is a defect nobody has. The store holds more than
// this section shows -the clean rows of paths nobody touched- so the heading says what it
// carries and not what the store holds. No carried rows means no section: the prompt stays
// byte-identical to what it always was.
func carriedSection(known []Finding, carried []Pass) string {
	if len(known) == 0 && len(carried) == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("## What this review already knows about these bytes\n\n")
	out.WriteString("The findings of earlier reviews, and the clean rows of the paths this candidate changed.\n")
	out.WriteString("The store keeps more of the clean rows; a path that is not in the diff is not repeated here.\n\n")
	// A row travels from verdict to verdict as it is inherited, so the store
	// accumulates copies of the same finding. The reviewer reads this section to
	// know what not to re-report, and the second copy teaches it nothing: the row
	// is written once, in the order the store lists it.
	written := make(map[string]bool, len(known)+len(carried))
	row := func(line string) {
		if written[line] {
			return
		}
		written[line] = true
		out.WriteString(line)
	}
	dismissed := false
	for _, finding := range known {
		where := finding.Path
		if finding.Line > 0 {
			where = fmt.Sprintf("%s:%d", finding.Path, finding.Line)
		}
		who := ""
		if finding.Lens != "" {
			who = finding.Lens + ", "
		}
		if IsDismissed(finding) {
			dismissed = true
			row("- " + where + " (" + who + string(finding.Severity) + ") dismissed: " + finding.Reason + "\n")
			continue
		}
		row("- " + where + " (" + who + string(finding.Severity) + ") " + finding.Why + "\n")
	}
	for _, pass := range carried {
		who := pass.Lens
		if who == "" {
			who = "?"
		}
		row("- " + pass.Path + " (" + who + ") audited clean\n")
	}
	out.WriteString("\nDo not re-report the findings above in \"findings\": the verdict inherits\n")
	out.WriteString("them as they are, with their receipts. Report a finding only when a defect\n")
	out.WriteString("is new to this candidate, and leave \"passes\" without the clean rows above;\n")
	out.WriteString("they are carried too.\n")
	if dismissed {
		out.WriteString("A row marked dismissed was judged with evidence in an earlier verdict: do not\n")
		out.WriteString("re-raise it. The same fingerprint stays dismissed, whatever this review says.\n")
	}
	out.WriteString("\n")

	return out.String()
}

// floorSection is the deterministic part of the prompt: the lines a finding may
// cite, the categories they touch, and the class the lenses were chosen from.
// It runs before any lens, so the reviewer reads the candidate's own writing
// first. A floor with no rows prints nothing: no rows, nothing to claim.
func floorSection(floor Floor) string {
	if len(floor.Rows) == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("## The floor (deterministic, before any lens)\n\n")
	out.WriteString("Path class: " + string(floor.Class) + " — the same class that chose this tree's lenses.\n")
	for _, row := range floor.Rows {
		out.WriteString("- " + row.Path + ": added lines " + FloorLines(row.Lines))
		if len(row.Signals) > 0 {
			out.WriteString(" · signals: " + strings.Join(row.Signals, ", "))
		}
		out.WriteString("\n")
	}
	out.WriteString("\nThis is the candidate's own writing: a finding must cite one of those lines or a\n")
	out.WriteString(fmt.Sprintf("line within %d of one. What the candidate did not touch is a follow-up\n", FloorContext))
	out.WriteString("(\"causal\": false), and a causal BLOCKER or CRITICAL that cannot cite is\n")
	out.WriteString("downgraded one tier by the tool before the verdict is stored.\n\n")

	return out.String()
}

// schema is the expected answer, with the keys the domain parses. The tree is
// stamped by the tool when it is empty, so a reviewer that cannot compute it
// still produces a usable verdict.
func answerSchema(in PromptInput) string {
	var out strings.Builder

	out.WriteString("{\n")
	out.WriteString(`  "tree": "` + in.Tree + `"` + ",\n")
	out.WriteString(`  "base": "` + in.Base + `"` + ",\n")
	out.WriteString("  \"reviewer\": \"<the command or your name>\",\n")
	out.WriteString("  \"model\": \"<the model that judged, when you know it: the gate compares it with the proof>\",\n")
	out.WriteString("  \"at\": \"<RFC3339, for example 2026-10-21T20:31:00Z>\",\n")
	out.WriteString("  \"inspection\": {\n")
	out.WriteString("    \"status\": \"" + string(InspectionCompleted) + " | " + string(InspectionUnavailable) + "\",\n")
	out.WriteString("    \"paths\": [\"<every changed path you read, when completed>\"],\n")
	out.WriteString("    \"reason\": \"<only when unavailable: why the candidate could not be read>\"\n")
	out.WriteString("  },\n")
	out.WriteString("  \"findings\": [\n")
	out.WriteString("    {\n")
	out.WriteString("      \"path\": \"<file>\",\n")
	out.WriteString("      \"line\": 0,\n")
	out.WriteString("      \"severity\": \"BLOCKER | CRITICAL | WARNING | SUGGESTION\",\n")
	if len(in.Lenses) > 0 {
		out.WriteString("      \"lens\": \"<the lens of this tree's list that found it>\",\n")
	}
	out.WriteString("      \"causal\": true,\n")
	out.WriteString("      \"evidence\": \"" + EvidenceCited + " | " + EvidenceArgued + "\",\n")
	out.WriteString("      \"why\": \"<the defect and the concrete evidence>\"\n")
	out.WriteString("    }\n")
	out.WriteString("  ],\n")

	if len(in.Lenses) > 0 {
		out.WriteString("  \"passes\": [\n")
		out.WriteString("    {\"path\": \"<file>\", \"lens\": \"<the lens of this tree's list that audited it>\"}\n")
		out.WriteString("  ],\n")
	}

	out.WriteString("  \"verdict\": \"" + Approved + " | " + ChangesRequested + "\"\n")
	out.WriteString("}\n")

	return out.String()
}

// section is one block of the prompt. Content that arrives without a trailing
// newline still ends its own block, so the next heading starts on its own line.
func section(text string) string {
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		return "(empty)\n"
	}
	return trimmed + "\n"
}
