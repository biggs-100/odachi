package init

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// SkillDest is one directory the skills half of `init` wrote in, or found already
// right. Label is how a human reads it: `./.agents/skills` in the project and
// `~/.agents/skills` in the home, never an absolute path.
type SkillDest struct {
	Label     string
	Installed int
	// Replaced is a skill whose bytes differ from the binary's: a real run backs
	// it up and writes it, a rehearsal says it would. Without the count a
	// rehearsal of an upgraded binary reported "nothing to write" while every
	// changed skill awaited a rewrite.
	Replaced  int
	Unchanged int
	Files     []string
}

// SkillReport is what the skills half did, in the words the human will read.
type SkillReport struct {
	// Dests is one entry per destination, in the order they were installed.
	Dests []SkillDest
	// Notice is a line about this project that deserves saying even though
	// nothing was written here: a second copy of the curriculum that an explicit
	// `odachi pack` put in another directory. init says it instead of pretending
	// it is not there.
	Notice []string
	// Blocked names a destination that could not be written, with the reason.
	// Empty when every destination is in place.
	Blocked string
	// Next is the command that unblocks it.
	Next string
	// Fatal says the block is inside the project. The home is the only
	// destination outside it, and a home that cannot be written never fails the
	// run that fixed the project; a project that did not get its skills is not
	// ready, because the agent would work without them.
	Fatal bool
}

// Skills is the half of the job that belongs to `odachi pack`. init decides that
// two destinations get skills — the project's cross-client directory and the
// human's home — and where this run puts them, and pack owns how a skill is
// installed: the destination table, the backup before a differing file, the
// identical bytes skipped. The scope is an argument instead of a field because
// init is the only one that can read the store: a caller that fixed the scope in
// advance could not honour the choice the project recorded.
//
// The two are joined in cmd/odachi, the only layer that already wires pack's
// payload and adapters; an app importing another app would inherit its error
// strings and its table for a job that is not its own.
type Skills interface {
	Install(scope string, dryRun bool) SkillReport
}

// installSkills is the last step of a start: the project is not ready until the
// agent has the skills it reads. A destination that cannot be written is reported
// with the command that fixes it, and never turns a finished project into a
// failed run — the home is the only part of the job outside the project.
func (r *runner) installSkills() {
	if r.skills == nil {
		return
	}
	if line := r.skillsChoice(); line != "" {
		r.lines = append(r.lines, line)
	}
	report := r.skills.Install(r.skillsScope, r.dryRun)

	for _, dest := range report.Dests {
		r.skillDests = append(r.skillDests, dest)
		r.skillLabels = append(r.skillLabels, dest.Label)
		r.skillFiles = append(r.skillFiles, dest.Files...)
		if line := skillLine(dest, r.dryRun); line != "" {
			r.lines = append(r.lines, line)
		}
	}
	r.lines = append(r.lines, report.Notice...)

	if report.Blocked != "" {
		r.skillBlocked = report.Blocked
		r.skillNext = report.Next
		r.skillFatal = report.Fatal
		r.lines = append(r.lines, "pending: "+report.Blocked)
	}
}

// skillsChoice names where the craft goes and who decided it, in the words that
// explain the next step. It comes before the destinations because "where" is the
// decision; "how many" is its consequence. The source matters: a human who reads
// "from the store" learns the flag is remembered, and one who reads "the default"
// learns nothing was ever chosen.
func (r *runner) skillsChoice() string {
	switch r.skillsSource {
	case skillsFromFlag:
		return "skills: " + r.skillsScope + ", from --skills; the store keeps the choice, and odachi init --skills " + other(r.skillsScope) + " changes it"
	case skillsFromStore:
		return "skills: " + r.skillsScope + ", from the store (" + persona.SelectionPath + "); odachi init --skills " + other(r.skillsScope) + " changes it"
	case skillsFromAdoption:
		return "skills: " + r.skillsScope + "; this repository already lives, so the craft stays in the machine (odachi setup puts it there); odachi init --skills project puts this project's own copy here"
	case skillsFromSetup:
		return "skills: " + r.skillsScope + "; setup keeps the craft once per machine, and every project reads it"
	default:
		return "skills: " + r.skillsScope + " (the default); odachi init --skills " + persona.SkillsHome + " keeps one copy per machine"
	}
}

// other names the scope this run did not use.
func other(scope string) string {
	if scope == persona.SkillsHome {
		return persona.SkillsProject
	}
	return persona.SkillsHome
}

// skillLine is one destination in one line: what arrived and what was already
// there. A rerun says `unchanged` instead of claiming work it did not do. A
// destination that got nothing and had nothing is not worth a line. A rehearsal
// says what it would install, in the same words the run would use.
func skillLine(dest SkillDest, dry bool) string {
	parts := make([]string, 0, 3)
	if dest.Installed > 0 {
		if dry {
			parts = append(parts, "would install "+count(dest.Installed, "skill", "skills"))
		} else {
			parts = append(parts, count(dest.Installed, "installed skill", "installed skills"))
		}
	}
	if dest.Replaced > 0 {
		if dry {
			parts = append(parts, "would replace "+count(dest.Replaced, "skill", "skills"))
		} else {
			parts = append(parts, count(dest.Replaced, "replaced skill", "replaced skills"))
		}
	}
	if dest.Unchanged > 0 {
		parts = append(parts, count(dest.Unchanged, "unchanged", "unchanged"))
	}
	if len(parts) == 0 {
		return ""
	}
	return dest.Label + ": " + strings.Join(parts, ", ")
}

// count renders "1 installed skill" and "8 installed skills".
func count(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
