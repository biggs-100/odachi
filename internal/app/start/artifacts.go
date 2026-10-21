package start

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/biggz/odachi/internal/domain/brief"
	"github.com/biggz/odachi/internal/domain/fence"
	"github.com/biggz/odachi/internal/domain/profile"
	"github.com/biggz/odachi/internal/domain/taskdoc"
)

// preset is one named non-goal from the Contract III catalog: the key a human
// selects and the fence it becomes.
type preset struct {
	key   string
	paths []string
	deps  []string
}

// presets is the closed catalog. Order is the order the options are shown in.
var presets = []preset{
	{key: "pagos", paths: []string{"payment*", "**/payments/**"}, deps: []string{"stripe", "mercadopago"}},
	{key: "login", paths: []string{"**/login*", "**/auth/login*"}},
	{key: "historia-clinica", paths: []string{"**/historia*", "**/medical-record*"}},
	{key: "recordatorios", paths: []string{"**/reminder*", "**/whatsapp*"}},
}

// presetKeys lists the catalog keys, for the question options.
func presetKeys() []string {
	keys := make([]string, 0, len(presets))
	for _, p := range presets {
		keys = append(keys, p.key)
	}
	return keys
}

// presetFor matches a selection against the catalog, ignoring case, accents and
// punctuation, so "Historia Clínica" is the same non-goal as the preset.
func presetFor(selection string) (preset, bool) {
	key := slugify(selection)
	for _, p := range presets {
		if p.key == key {
			return p, true
		}
	}
	return preset{}, false
}

// rulesFor turns the selected non-goals into fence rules, one per selection in
// the order the human chose them. A preset brings its catalog paths and deps; a
// free-text non-goal becomes "<slug>*".
func rulesFor(selections []string) []fence.Rule {
	rules := []fence.Rule{}
	seen := map[string]bool{}

	for _, selection := range selections {
		text := strings.TrimSpace(selection)
		if text == "" {
			continue
		}

		if p, ok := presetFor(text); ok {
			if seen[p.key] {
				continue
			}
			seen[p.key] = true
			rules = append(rules, fence.Rule{Not: p.key, Paths: p.paths, Deps: p.deps})
			continue
		}

		slug := slugify(text)
		if slug == "" {
			continue
		}
		if seen[slug] {
			continue
		}
		seen[slug] = true
		rules = append(rules, fence.Rule{Not: text, Paths: []string{slug + "*"}})
	}

	return rules
}

// renderFence writes the fence in the canonical TOML shape: one [[rule]] per
// non-goal, one key per line.
func renderFence(rules []fence.Rule) string {
	if len(rules) == 0 {
		// An empty file teaches nothing. The fence is the mechanism that keeps
		// scope small, so when the human chose no exclusion the file explains
		// what it is for and what a rule looks like.
		return "# Vacío a propósito: no elegiste ningún \"NO incluye\".\n" +
			"# El cerco es lo que frena el crecimiento del alcance: sin reglas, nada\n" +
			"# te avisa cuando el trabajo se va de lo que pediste.\n" +
			"#\n" +
			"# Cuando decidas algo que no va en esta primera versión, agregalo acá:\n" +
			"#\n" +
			"# [[rule]]\n" +
			"# not   = \"pagos\"\n" +
			"# paths = [\"payment*\", \"**/payments/**\"]\n" +
			"# deps  = [\"stripe\", \"mercadopago\"]\n"
	}

	var out strings.Builder

	for i, rule := range rules {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("[[rule]]\n")
		fmt.Fprintf(&out, "%-5s = %s\n", "not", strconv.Quote(rule.Not))
		fmt.Fprintf(&out, "%-5s = %s\n", "paths", quoteList(rule.Paths))
		fmt.Fprintf(&out, "%-5s = %s\n", "deps", quoteList(rule.Deps))
	}

	return out.String()
}

// quoteList renders a TOML array of strings.
func quoteList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, strconv.Quote(item))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// RenderProfile writes the resolved profile, one field per line, name first and
// always present, in the key order of Contract A. It is exported because
// `odachi level --to` writes the same file, and the format has one owner.
func RenderProfile(p profile.Profile) string {
	fields := []struct {
		key   string
		value string
	}{
		{"name", p.Name},
		{"extends", p.Extends},
		{"brief_source", p.BriefSource},
		{"questions", p.Questions},
		{"assumptions", p.Assumptions},
		{"output", p.Output},
		{"proof", p.Proof},
		{"explain", p.Explain},
	}

	var out strings.Builder
	for _, field := range fields {
		fmt.Fprintf(&out, "%-12s = %s\n", field.key, strconv.Quote(field.value))
	}
	return out.String()
}

// artifactInput is what changes between the interview path and the --brief
// path: the project title, the non-goals to fence, and whether the human's
// brief.md may be written.
type artifactInput struct {
	title      string
	nonGoals   []string
	writeBrief bool
}

// writeArtifacts renders and writes the four artifacts, backing up any file it
// is about to replace.
func (r *runner) writeArtifacts(in artifactInput) Result {
	slug := slugify(in.title)
	if slug == "" {
		return blocked("no pude armar el nombre del proyecto: el outcome no tiene letras ni números; volvé a correr odachi start con una frase distinta")
	}

	if in.writeBrief {
		content := brief.Brief{
			Name:      slug,
			Outcome:   []string{r.answers.outcome},
			DoneWhen:  r.answers.done,
			NonGoals:  in.nonGoals,
			Decisions: r.answers.decisions(),
		}
		if problem := r.put(briefFile, []byte(brief.Render(content))); problem != nil {
			return *problem
		}
	}

	if problem := r.put(profileFile, []byte(RenderProfile(r.profile))); problem != nil {
		return *problem
	}
	if problem := r.put(fenceFile, []byte(renderFence(rulesFor(in.nonGoals)))); problem != nil {
		return *problem
	}

	taskPath, existing, problem := r.taskPathFor(slug)
	if problem != nil {
		return *problem
	}

	// An existing task document is the memory of work already done: its steps
	// and its history are never rewritten by starting again. The brief is what
	// a rerun corrects; the log of the work is not.
	if !existing {
		doc := taskdoc.TaskDoc{
			Title:        slug,
			Requirements: []taskdoc.Requirement{{ID: "R1", Text: r.answers.outcome}},
			DoneWhen:     r.answers.done,
			History:      []string{`L1 · "` + r.answers.outcome + `"`},
		}
		if problem := r.put(taskPath, []byte(doc.Render())); problem != nil {
			return *problem
		}
	}

	return Result{
		Code:    CodeOK,
		Message: "listo: escribí " + humanList(r.written) + "; revisá brief.md y el task doc antes de seguir",
		Profile: r.profile.Name,
		Task:    filepath.ToSlash(taskPath),
		Files:   r.written,
	}
}

// put backs up an existing file and then replaces it atomically. Backup runs
// first, always: the human's file is never the price of a rewrite.
func (r *runner) put(rel string, data []byte) *Result {
	path := filepath.Join(r.opts.Dir, rel)

	exists, err := r.files.Exists(path)
	if err != nil {
		return failure(blocked("no pude revisar " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}
	if exists {
		if _, err := r.files.Backup(path); err != nil {
			return failure(blocked("no pude respaldar " + rel + " antes de reescribirlo: " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
		}
	}
	if err := r.files.Write(path, data); err != nil {
		return failure(blocked("no pude escribir " + rel + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}

	// Reported paths are slash-separated on every OS: the message and the JSON
	// output are read by a human and by a machine, not by the filesystem.
	r.written = append(r.written, filepath.ToSlash(rel))
	return nil
}

// taskPathFor returns the task document for this feature: the existing one
// whose slug matches (reported as existing), or a new one with the next number.
// Reusing by slug is what keeps one feature to one document, so a beginner who
// fixes the brief by running start again does not leave the agent two documents
// to choose from.
func (r *runner) taskPathFor(slug string) (path string, existing bool, problem *Result) {
	names, err := r.files.List(filepath.Join(r.opts.Dir, taskDir))
	if err != nil {
		return "", false, failure(blocked("no pude leer " + taskDir + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}

	suffix := "-" + slug + ".md"
	for _, name := range names {
		if strings.HasSuffix(name, suffix) {
			return filepath.Join(taskDir, name), true, nil
		}
	}

	number, problem := r.nextNumber()
	if problem != nil {
		return "", false, problem
	}
	return filepath.Join(taskDir, fmt.Sprintf("%04d-%s.md", number, slug)), false, nil
}

// nextNumber is one more than the highest task-document number on disk, or 1
// when the directory has no task document yet. Backup files and notes do not
// count.
func (r *runner) nextNumber() (int, *Result) {
	names, err := r.files.List(filepath.Join(r.opts.Dir, taskDir))
	if err != nil {
		return 0, failure(blocked("no pude leer " + taskDir + ": " + oneLine(err.Error()) + "; revisá los permisos y volvé a correr odachi start"))
	}

	highest := 0
	for _, name := range names {
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		prefix, _, ok := strings.Cut(name, "-")
		if !ok {
			continue
		}
		number, err := strconv.Atoi(prefix)
		if err != nil || number <= 0 {
			continue
		}
		if number > highest {
			highest = number
		}
	}
	return highest + 1, nil
}

// humanList joins paths the way a Spanish sentence does: "a, b y c".
func humanList(items []string) string {
	switch len(items) {
	case 0:
		return "nada"
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " y " + items[len(items)-1]
	}
}

// taskSlug is the first three words of the outcome, ASCII lowercase, joined by
// "-". The task document and the brief share it, so the two files agree.
func taskSlug(outcome string) string {
	words := slugWords(outcome)
	if len(words) > 3 {
		words = words[:3]
	}
	return strings.Join(words, "-")
}

// slugify is every word of the text, ASCII lowercase, joined by "-".
func slugify(text string) string {
	return strings.Join(slugWords(text), "-")
}

// slugWords splits text into ASCII lowercase words: accents are folded and
// everything that is not a letter or a digit separates words.
func slugWords(text string) []string {
	var words []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	for _, r := range text {
		if folded, ok := foldAccent(r); ok {
			r = folded
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			current.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			current.WriteRune(r + ('a' - 'A'))
		default:
			flush()
		}
	}
	flush()

	return words
}

// foldAccents maps one accented rune to its ASCII base. Spanish is the only
// language the messages guarantee, and the slug has to survive Windows and
// Linux alike.
func foldAccents(value string) string {
	var out strings.Builder
	for _, r := range value {
		if folded, ok := foldAccent(r); ok {
			out.WriteRune(folded)
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func foldAccent(r rune) (rune, bool) {
	switch r {
	case 'á', 'à', 'ä', 'â', 'ã', 'Á', 'À', 'Ä', 'Â', 'Ã':
		return 'a', true
	case 'é', 'è', 'ë', 'ê', 'É', 'È', 'Ë', 'Ê':
		return 'e', true
	case 'í', 'ì', 'ï', 'î', 'Í', 'Ì', 'Ï', 'Î':
		return 'i', true
	case 'ó', 'ò', 'ö', 'ô', 'õ', 'Ó', 'Ò', 'Ö', 'Ô', 'Õ':
		return 'o', true
	case 'ú', 'ù', 'ü', 'û', 'Ú', 'Ù', 'Ü', 'Û':
		return 'u', true
	case 'ñ', 'Ñ':
		return 'n', true
	case 'ç', 'Ç':
		return 'c', true
	default:
		return 0, false
	}
}
