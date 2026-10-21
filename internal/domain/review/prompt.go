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

	out.WriteString("# odachi review · árbol " + in.Tree + " · base " + in.Base + "\n\n")
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

	out.WriteString("## Qué devolver\n")
	out.WriteString(schema(in))
	out.WriteString("\n")
	out.WriteString("El veredicto se deriva de los findings: si hay un finding con \"causal\": true y\n")
	out.WriteString("severidad BLOCKER o CRITICAL, el veredicto es " + ChangesRequested + "; si no, es " + Approved + ".\n")
	out.WriteString("No lo decidas vos: lo decide la regla.\n\n")
	out.WriteString("Contestá solo con el JSON, sin texto alrededor.\n")

	return out.String()
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
