package review

import "strings"

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
}

// Lens is one specialist prompt: its name and its body.
type Lens struct {
	Name string
	Body string
}

// Prompt renders the review prompt. The content is fixed and pinned by tests: a
// prompt that drifts produces reviews that drift, and the five rules are the
// whole reason this verb exists.
//
// The rules are not invented here. They are what a review system learns the hard
// way: report only real, user-impacting defects you can defend with evidence;
// when in doubt say nothing; style and preference are banned unless they hide a
// defect; the severity vocabulary is closed; an empty review is exactly one
// sentence; a finding counts only if the candidate caused it; and one exhaustive
// pass per lens, with no loop until it dries.
func Prompt(in PromptInput) string {
	var out strings.Builder

	out.WriteString("# odachi review · árbol " + in.Tree + " · base " + in.Base)
	if names := lensNames(in.Lenses); len(names) > 0 {
		out.WriteString(" · lentes " + strings.Join(names, ", "))
	}
	out.WriteString("\n\n")
	out.WriteString("Sos el reviewer de diseño de este candidato. Revisás diseño, no estilo:\n")
	out.WriteString("un review con ruido se ignora, y un review ignorado no protege a nadie.\n\n")

	out.WriteString("## El brief (lo que el humano pidió)\n")
	// The diff is written raw, without a fence: a diff of a Markdown file
	// contains fences, and a fence inside a fence is a prompt that leaks.
	out.WriteString(section(in.Brief))
	out.WriteString("\n")

	out.WriteString("## Las rutas cambiadas\n")
	for _, path := range in.Paths {
		out.WriteString("- " + path + "\n")
	}
	out.WriteString("\n")

	out.WriteString("## El diff (el candidato completo, contra la base)\n")
	out.WriteString(section(in.Diff))
	out.WriteString("\n")

	out.WriteString("## Las cinco reglas\n")
	out.WriteString("1. Reportá un finding solo si es un defecto real que afecta al usuario y que podés\n")
	out.WriteString("   defender con evidencia concreta. Ante la duda, callate.\n")
	out.WriteString("   Estilo y preferencia están prohibidos salvo que oculten un defecto.\n")
	out.WriteString("2. La severidad es exactamente una de BLOCKER, CRITICAL, WARNING, SUGGESTION.\n")
	out.WriteString("3. Si no hay nada que reportar, no inventes nada para tener algo que decir: el JSON\n")
	out.WriteString("   lleva \"findings\": [] y, si contestás en prosa, la respuesta es exactamente:\n")
	out.WriteString("   " + NoFindings + "\n")
	out.WriteString("4. Un finding cuenta solo si el candidato lo introdujo, lo activó o lo empeoró. Lo que\n")
	out.WriteString("   ya estaba es un follow-up, no un bloqueo: marcalo \"causal\": false.\n")
	out.WriteString("5. Un solo barrido por lente. No vuelvas a escanear hasta que no quede nada.\n")
	out.WriteString("\n")

	out.WriteString(lensSection(in.Lenses))

	out.WriteString("## Qué devolver\n")
	out.WriteString(schema(in))
	out.WriteString("\n")
	out.WriteString("El veredicto se deriva de los findings: si hay un finding con \"causal\": true y\n")
	out.WriteString("severidad BLOCKER o CRITICAL, el veredicto es " + ChangesRequested + "; si no, es " + Approved + ".\n")
	out.WriteString("No lo decidas vos: lo decide la regla.\n")
	if len(in.Lenses) > 0 {
		out.WriteString("Las lentes que cubriste las anota la herramienta: no las escribas en el JSON.\n")
	}
	out.WriteString("\n")

	out.WriteString("Contestá solo con el JSON, sin texto alrededor.\n")

	return out.String()
}

// lensSection is the block that turns one generic prompt into a specialist's,
// without repeating the five rules: they are stated once, above, and every lens
// inherits them.
func lensSection(lenses []Lens) string {
	if len(lenses) == 0 {
		return ""
	}

	var out strings.Builder
	out.WriteString("## Las lentes de este árbol\n")
	out.WriteString("Las lentes las eligió la evidencia del diff, no el gusto de nadie. Hacé un barrido por lente,\n")
	out.WriteString("en el orden de esta lista, y reportá cada finding una sola vez.\n\n")

	for _, lens := range lenses {
		out.WriteString("### " + lens.Name + "\n")
		out.WriteString(strings.TrimRight(stripFrontmatter(lens.Body), " \n") + "\n\n")
	}
	return out.String()
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

// schema is the expected answer, with the keys the domain parses. The tree is
// stamped by the tool when it is empty, so a reviewer that cannot compute it
// still produces a usable verdict.
func schema(in PromptInput) string {
	var out strings.Builder

	out.WriteString("{\n")
	out.WriteString(`  "tree": "` + in.Tree + `"` + ",\n")
	out.WriteString(`  "base": "` + in.Base + `"` + ",\n")
	out.WriteString("  \"reviewer\": \"<el comando o tu nombre>\",\n")
	out.WriteString("  \"at\": \"<RFC3339, por ejemplo 2026-10-21T20:31:00Z>\",\n")
	out.WriteString("  \"findings\": [\n")
	out.WriteString("    {\n")
	out.WriteString("      \"path\": \"<archivo>\",\n")
	out.WriteString("      \"line\": 0,\n")
	out.WriteString("      \"severity\": \"BLOCKER | CRITICAL | WARNING | SUGGESTION\",\n")
	out.WriteString("      \"causal\": true,\n")
	out.WriteString("      \"why\": \"<el defecto y la evidencia concreta>\"\n")
	out.WriteString("    }\n")
	out.WriteString("  ],\n")
	out.WriteString("  \"verdict\": \"" + Approved + " | " + ChangesRequested + "\"\n")
	out.WriteString("}\n")

	return out.String()
}

// section is one block of the prompt. Content that arrives without a trailing
// newline still ends its own block, so the next heading starts on its own line.
func section(text string) string {
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		return "(vacío)\n"
	}
	return trimmed + "\n"
}
