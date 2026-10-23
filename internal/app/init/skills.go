package init

import (
	"fmt"
	"strings"
)

// SkillDest is one directory the skills half of `init` wrote in, or found already
// right. Label is how a human reads it: `./.agents/skills` in the project and
// `~/.agents/skills` in the home, never an absolute path.
type SkillDest struct {
	Label     string
	Installed int
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
// human's home — and pack owns how a skill is installed: the destination table,
// the backup before a differing file, the identical bytes skipped. The two are
// joined in cmd/odachi, the only layer that already wires pack's payload and
// adapters; an app importing another app would inherit its error strings and its
// table for a job that is not its own.
type Skills interface {
	Install() SkillReport
}

// installSkills is the last step of a start: the project is not ready until the
// agent has the skills it reads. A destination that cannot be written is reported
// with the command that fixes it, and never turns a finished project into a
// failed run — the home is the only part of the job outside the project.
func (r *runner) installSkills() {
	if r.skills == nil {
		return
	}
	report := r.skills.Install()

	for _, dest := range report.Dests {
		r.skillDests = append(r.skillDests, dest)
		r.skillLabels = append(r.skillLabels, dest.Label)
		r.skillFiles = append(r.skillFiles, dest.Files...)
		if line := skillLine(dest); line != "" {
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

// skillLine is one destination in one line: what arrived and what was already
// there. A rerun says `unchanged` instead of claiming work it did not do. A
// destination that got nothing and had nothing is not worth a line.
func skillLine(dest SkillDest) string {
	parts := make([]string, 0, 2)
	if dest.Installed > 0 {
		parts = append(parts, count(dest.Installed, "installed skill", "installed skills"))
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
