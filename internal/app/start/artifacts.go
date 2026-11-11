package start

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/fencetoml"
	"github.com/biggs-100/odachi/internal/app/tasks"
	"github.com/biggs-100/odachi/internal/domain/brief"
	"github.com/biggs-100/odachi/internal/domain/fence"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/profile"
	"github.com/biggs-100/odachi/internal/domain/rules"
	"github.com/biggs-100/odachi/internal/domain/taskdoc"
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

// rulesFor turns the selected non-goals into fence rules, one per selection that
// carries paths, in the order the human chose them. A preset brings its catalog
// paths and deps. A phrase brings nothing: a rule protects files, and a glob of the
// phrase's own words matches files named like the phrase and protects nothing -the
// fence's own header says so, and `renderFence` carries the phrase as a template
// for the human to complete instead.
func rulesFor(selections []string) []fence.Rule {
	rules := []fence.Rule{}
	seen := map[string]bool{}

	for _, selection := range selections {
		text := strings.TrimSpace(selection)
		if text == "" {
			continue
		}

		p, ok := presetFor(text)
		if !ok {
			continue
		}
		if seen[p.key] {
			continue
		}
		seen[p.key] = true
		rules = append(rules, fence.Rule{Not: p.key, Paths: p.paths, Deps: p.deps})
	}

	return rules
}

// looseNonGoals is the non-goals that arrived as free text. A rule protects files,
// so a phrase becomes a commented template with the paths left to fill: writing it
// as a glob of its own words would be a rule that protects nothing and passes for a
// promise.
func looseNonGoals(selections []string) []string {
	var loose []string
	for _, selection := range selections {
		text := strings.TrimSpace(selection)
		if text == "" {
			continue
		}
		if _, ok := presetFor(text); ok {
			continue
		}
		loose = append(loose, text)
	}
	return loose
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
// path: the project title, the non-goals to fence, whether the human's brief.md
// may be written, and -when --brief read a draft- the brief to record.
type artifactInput struct {
	title      string
	nonGoals   []string
	writeBrief bool
	// briefDoc is the parsed brief a draft carried. When it is set, that is what
	// brief.md holds: the run records the vision it read, not a second rendering
	// of the answers. Nil means the interview built it from the answers.
	briefDoc *brief.Brief
}

// writeArtifacts renders and writes the four artifacts, backing up any file it
// is about to replace.
func (r *runner) writeArtifacts(in artifactInput) Result {
	slug := slugify(in.title)
	if slug == "" {
		return blocked("could not build the project name: the outcome has no letters or digits; run odachi start again with a different sentence")
	}

	// The fence is read before anything is written: it is the memory of what the
	// project refuses, and a run that cannot merge it must block with the human's
	// brief, profile and voice untouched. Every rule already there is kept as it
	// is, and only the non-goals this run brings are new (0266). A non-goal that
	// arrives as a phrase is not a rule -it protects no path, so the file carries
	// it as a template and the run says what to write- and a rule nobody can
	// trust is worse than a rule that is still a question.
	// The fence is built from every non-goal the brief carries, not only this run's:
	// the brief accumulates since 0271, and a rule whose sentence survives while the
	// template it came from disappears would be half a memory. The union is read
	// before the first write, so a fence or a brief that cannot be read still blocks
	// with the human's files untouched.
	fenceGoals, problem := r.briefAndRunNonGoals(in.nonGoals)
	if problem != nil {
		return *problem
	}
	rules, loose, kept, hadFence, problem := r.mergeFence(fenceGoals)
	if problem != nil {
		return *problem
	}

	if in.writeBrief {
		content := brief.Brief{
			Name:      slug,
			Outcome:   []string{r.answers.outcome},
			DoneWhen:  r.answers.done,
			NonGoals:  in.nonGoals,
			Decisions: r.answers.decisions(),
		}
		if in.briefDoc != nil {
			content = *in.briefDoc
		}
		number, problem := r.nextNumber()
		if problem != nil {
			return *problem
		}
		if problem := r.writeBriefBlock(number, slug, content); problem != nil {
			return *problem
		}
	}

	if problem := r.put(profileFile, []byte(RenderProfile(r.profile))); problem != nil {
		return *problem
	}
	// The profile decides the voice, and this run is what decided the profile:
	// leaving the block at `novato` while the profile says `senior` is an agent
	// that talks differently from what the project promises.
	if problem := r.refreshVoice(); problem != nil {
		return *problem
	}
	if problem := r.put(fenceFile, []byte(fencetoml.Render(rules, loose))); problem != nil {
		return *problem
	}

	taskPath, existing, problem := r.taskPathFor(slug)
	if problem != nil {
		return *problem
	}

	// An existing task document is the memory of work already done: its steps
	// and its history are never rewritten by starting again. The brief is what
	// a rerun corrects; the log of the work is not.
	reused := ""
	if existing {
		reused = filepath.ToSlash(taskPath) + ": kept, the steps and history it carries are the memory of the work"
	} else {
		doc := taskdoc.TaskDoc{
			Title:        slug,
			Base:         r.opts.Base,
			ReviewWaived: r.opts.WaiveReview,
			Requirements: []taskdoc.Requirement{{ID: "R1", Text: r.answers.outcome}},
			DoneWhen:     r.answers.done,
			History:      []string{`L1 · "` + r.answers.outcome + `"`},
		}
		if problem := r.put(taskPath, []byte(doc.Render())); problem != nil {
			return *problem
		}
	}

	// The documents odachi does not read are named once: a project that keeps
	// another shape (ODD names a feature without a number) must hear what the
	// tool skips instead of wondering why summary counts one.
	var lines []string
	if reused != "" {
		lines = append(lines, reused)
	}
	if kept > 0 || hadFence {
		line := fmt.Sprintf("the fence kept %d rules it already had", kept)
		if added := len(rules) - kept; added > 0 {
			line += fmt.Sprintf(" and added %d", added)
		}
		lines = append(lines, line)
	}
	for _, text := range loose {
		lines = append(lines, fmt.Sprintf(
			"the non-goal %q came as free text: a rule protects files, and a phrase names none, so %s carries it as a template; write the paths it must keep out and remove the '#' in front of the block",
			text, filepath.ToSlash(fenceFile)))
	}
	if len(loose) > 0 {
		lines = append(lines, "to prove a rule works: touch a path it protects and run odachi check - it must refuse")
	}
	if skipped, err := tasks.Skipped(r.files, r.opts.Dir); err == nil {
		if line := tasks.SkippedLine(skipped); line != "" {
			lines = append(lines, line)
		}
	}

	return Result{
		Code:    CodeOK,
		Message: "done: wrote " + humanList(r.written) + "; commit what start wrote (the plan) before you build",
		Lines:   lines,
		Profile: r.profile.Name,
		Task:    filepath.ToSlash(taskPath),
		Files:   r.written,
	}
}

// briefAndRunNonGoals is the non-goal list the fence is built from: the existing
// brief's, in the order they were written, and then this run's, without repeating a
// sentence. A brief that cannot be read is not an empty one: it blocks, because
// building the fence from half a vision is worse than stopping (0271).
func (r *runner) briefAndRunNonGoals(run []string) ([]string, *Result) {
	goals := make([]string, 0, len(run)+1)

	path := filepath.Join(r.opts.Dir, briefFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("could not check " + briefFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again")
		return nil, &problem
	}
	if exists {
		data, err := r.files.Read(path)
		if err != nil {
			problem := blocked("could not read " + briefFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again")
			return nil, &problem
		}
		goals = append(goals, brief.Parse(string(data)).NonGoals...)
	}

	for _, goal := range run {
		if !slices.Contains(goals, strings.TrimSpace(goal)) {
			goals = append(goals, goal)
		}
	}
	return goals, nil
}

// writeBriefBlock records this run's vision without erasing the ones before it. A
// brief that replaces itself loses the sentence behind every fence rule that
// outlived it, and the fence has been append-only since 0266: the block repeats the
// canonical headers Parse already reads, so the union costs no second grammar
// (0271). A requirement the file already carries is the same feature, and nothing
// is appended twice.
func (r *runner) writeBriefBlock(number int, title string, content brief.Brief) *Result {
	path := filepath.Join(r.opts.Dir, briefFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		return failure(blocked("could not check " + briefFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}
	if !exists {
		return r.put(briefFile, []byte(brief.Render(content)))
	}

	data, err := r.files.Read(path)
	if err != nil {
		return failure(blocked("could not read " + briefFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}

	// The same requirement is the same feature: a run that repeats it appends
	// nothing, because the brief is a record of decisions and not a log of attempts.
	existing := brief.Parse(string(data))
	fresh := make([]string, 0, len(content.Outcome))
	for _, outcome := range content.Outcome {
		if !slices.Contains(existing.Outcome, strings.TrimSpace(outcome)) {
			fresh = append(fresh, outcome)
		}
	}
	if len(fresh) == 0 {
		// Nothing new, and the file is still written: --force is the human asking
		// for the artifacts again, its backup is what undo reads, and rewriting the
		// same bytes adds no block.
		return r.put(briefFile, data)
	}
	content.Outcome = fresh

	block := fmt.Sprintf("## %04d · %s\n\n%s", number, title, brief.RenderFeature(content))
	merged := strings.TrimRight(string(data), "\n") + "\n\n" + block
	return r.put(briefFile, []byte(merged))
}

// mergeFence reads the fence the worktree already has and returns it with the
// non-goals this run brings appended; kept is how many rules were already there
// and hadFence is whether there was a file at all, which is what the run reports.
// A rule already written wins over the same non-goal arriving again, because the
// rule's name is its identity and a run is not the owner of a policy it found.
//
// The rules are read through fencetoml, the same grammar check reads them with:
// a merge that writes what the gate cannot parse is the defect this replaces. A
// file it cannot read blocks instead of being overwritten, because a policy
// nobody could understand is not this run's to delete.
func (r *runner) mergeFence(nonGoals []string) (rules []fence.Rule, loose []string, kept int, hadFence bool, problem *Result) {
	existing, hadFence, problem := r.fenceRulesOnDisk()
	if problem != nil {
		return nil, nil, 0, false, problem
	}

	rules = append(rules, existing...)
	kept = len(existing)
	seen := make(map[string]bool, len(existing)+len(nonGoals))
	for _, rule := range existing {
		seen[rule.Not] = true
	}
	for _, rule := range rulesFor(nonGoals) {
		if seen[rule.Not] {
			continue
		}
		seen[rule.Not] = true
		rules = append(rules, rule)
	}
	for _, text := range looseNonGoals(nonGoals) {
		if seen[text] {
			continue
		}
		seen[text] = true
		loose = append(loose, text)
	}
	return rules, loose, kept, hadFence, nil
}

// fenceRulesOnDisk is the fence this run is about to replace. A missing file is
// an empty policy; a file that exists and cannot be read or parsed is a policy
// the run cannot merge, and the block names the file.
func (r *runner) fenceRulesOnDisk() ([]fence.Rule, bool, *Result) {
	path := filepath.Join(r.opts.Dir, fenceFile)
	exists, err := r.files.Exists(path)
	if err != nil {
		problem := blocked("could not check " + fenceFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again")
		return nil, false, &problem
	}
	if !exists {
		return nil, false, nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		problem := blocked("could not read " + fenceFile + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again")
		return nil, false, &problem
	}

	rules, err := fencetoml.Decode(data)
	if err != nil {
		problem := blocked("could not read " + fenceFile + ": " + oneLine(err.Error()) + "; fix the file and run odachi start again")
		return nil, false, &problem
	}
	return rules, true, nil
}

// refreshVoice keeps the block's voice in step with the profile this run wrote.
//
// The block belongs to `init`, so a project without one is left alone: creating
// it here would make two commands own the same artifact. A profile with no voice
// of its own (a custom name that extends another) changes nothing about how the
// agent talks, so it is not an error here.
//
// The native agent lives in the home since 0242: a project file with that name is
// a leftover an older release wrote, and start does not refresh it. Level learned
// the same rule; a profile change only rewrites the block.
func (r *runner) refreshVoice() *Result {
	voice, err := persona.VoiceName(r.profile.Name)
	if err != nil {
		return nil
	}

	path := filepath.Join(r.opts.Dir, rules.File)
	exists, err := r.files.Exists(path)
	if err != nil {
		return failure(blocked("could not check " + rules.File + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}
	if !exists {
		return nil
	}

	data, err := r.files.Read(path)
	if err != nil {
		return failure(blocked("could not read " + rules.File + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}

	updated, changed, err := rules.WithVoice(string(data), voice)
	if err != nil {
		return failure(blocked("could not read the voice of profile " + r.profile.Name + ": " + oneLine(err.Error()) + "; run odachi init again"))
	}
	if changed {
		if problem := r.put(rules.File, []byte(updated)); problem != nil {
			return problem
		}
	}
	return nil
}

// put backs up an existing file and then replaces it atomically. Backup runs
// first, always: the human's file is never the price of a rewrite.
func (r *runner) put(rel string, data []byte) *Result {
	path := filepath.Join(r.opts.Dir, rel)

	exists, err := r.files.Exists(path)
	if err != nil {
		return failure(blocked("could not check " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}
	if exists {
		if _, err := r.files.Backup(path); err != nil {
			return failure(blocked("could not back up " + rel + " before rewriting it: " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
		}
	}
	if err := r.files.Write(path, data); err != nil {
		return failure(blocked("could not write " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}

	// Reported paths are slash-separated on every OS: the message and the JSON
	// output are read by a human and by a machine, not by the filesystem.
	r.written = append(r.written, filepath.ToSlash(rel))
	return nil
}

// taskPathFor returns the task document for this feature: the existing one whose
// slug matches and still has work (reported as existing), or a new one with the
// next number. Reusing by slug is what keeps one feature to one document, so a
// beginner who fixes the brief by running start again does not leave the agent two
// documents to choose from. A document with every step done is another feature's
// memory: its name is the same three words, and it is not this feature's to fill.
func (r *runner) taskPathFor(slug string) (path string, existing bool, problem *Result) {
	names, err := r.files.List(filepath.Join(r.opts.Dir, taskDir))
	if err != nil {
		return "", false, failure(blocked("could not read " + taskDir + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
	}

	suffix := "-" + slug + ".md"
	for _, name := range names {
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		rel := filepath.Join(taskDir, name)
		open, problem := r.documentOpen(rel)
		if problem != nil {
			return "", false, problem
		}
		if open {
			return rel, true, nil
		}
	}

	number, problem := r.nextNumber()
	if problem != nil {
		return "", false, problem
	}
	return filepath.Join(taskDir, fmt.Sprintf("%04d-%s.md", number, slug)), false, nil
}

// documentOpen reports whether a document still has work: a step not done, or no
// steps yet, which is what a fresh document looks like. Reading it is the only way
// to tell a re-run of this feature from a new feature that repeats its first three
// words.
func (r *runner) documentOpen(rel string) (bool, *Result) {
	data, err := r.files.Read(filepath.Join(r.opts.Dir, rel))
	if err != nil {
		problem := failure(blocked("could not read " + rel + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
		return false, problem
	}
	done, total := taskdoc.Parse(string(data)).Progress()
	return total == 0 || done < total, nil
}

// nextNumber is one more than the highest task-document number on disk, or 1
// when the directory has no task document yet. Backup files and notes do not
// count.
func (r *runner) nextNumber() (int, *Result) {
	names, err := r.files.List(filepath.Join(r.opts.Dir, taskDir))
	if err != nil {
		return 0, failure(blocked("could not read " + taskDir + ": " + oneLine(err.Error()) + "; check the permissions and run odachi start again"))
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

// humanList joins paths the way an English sentence does: "a, b and c".
func humanList(items []string) string {
	switch len(items) {
	case 0:
		return "nothing"
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
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

// foldAccents maps one accented rune to its ASCII base. A human may type
// accented words in their own language, and the slug has to survive Windows and
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
