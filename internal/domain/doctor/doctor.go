// Package doctor owns the rules of `odachi doctor`: the twelve health checks in
// order, how the pre-commit hook and the AGENTS.md block written by
// `odachi init` are read back, and the exact phrase that repairs each problem.
//
// It is pure: no filesystem, no os, no network. The app layer gathers the facts
// one check at a time and hands them here, so the same project produces the same
// report on every machine and a test can pin every line without touching a disk.
package doctor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/biggs-100/odachi/internal/domain/rules"
)

// The checks Contract XV asks for, in the order a human reads them, plus the one
// that checks the block's voice against the profile the binary reads. The number is the line's number, and
// it is how a report and a test name the same check: adding one in the middle
// would silently rename all the others, so the order is fixed here and nowhere
// else, and a new check is appended.
const (
	// CheckVersion is the version this binary reports.
	CheckVersion = iota + 1
	// CheckOnPath asks whether a shell can run `odachi` by name. It is the most
	// valuable of the twelve: when it fails, the agent cannot execute a single
	// rule and spends its budget looking for the binary.
	CheckOnPath
	// CheckGit asks whether there is a repository: without one there is no hook.
	CheckGit
	// CheckAgents asks whether the rules block is in AGENTS.md.
	CheckAgents
	// CheckHook asks whether the pre-commit hook exists and is odachi's.
	CheckHook
	// CheckHookExe asks whether the binary the hook names is still there.
	CheckHookExe
	// CheckProfile asks whether .odachi/profile.toml resolves to a valid profile.
	CheckProfile
	// CheckBrief asks whether brief.md exists and nothing is missing from it.
	CheckBrief
	// CheckSkills asks how many skills there are and whether SKILLS.md is up to
	// date, and how many are installed where the clients read them: two facts and
	// two answers on one line, because a project that just ran `odachi init` has
	// no skills/ registry and nine installed skills, and merging them into one
	// number is what made the line unreadable.
	CheckSkills
	// CheckIgnored asks whether the review store and proof/ are ignored: without
	// that, every review would invalidate the proof and the other way round.
	CheckIgnored
	// CheckVoice asks whether the last section of the AGENTS.md block talks the
	// way the profile in .odachi/profile.toml says it should, and whether the
	// native agent still mirrors that block. The binary writes both, so the
	// halves disagreeing is a project that prints one register and talks in
	// another. It is appended after the ten Contract XV named, so no existing
	// check changes its number.
	CheckVoice
	// CheckTaskDocs asks which documents in odd/tasks odachi can read. It is not a
	// problem: a project may keep another shape - ODD names a feature without a
	// number - and the two memories can live in the same folder. Saying it is the
	// difference between "the tool ignores my history" and "the tool told me it
	// does".
	CheckTaskDocs
)

// CheckCount is how many checks a complete report has. A test asserts it, so a
// report can never quietly grow or shrink.
const CheckCount = 12

// Status is what one check found.
type Status int

const (
	// StatusOK means there is nothing to fix.
	StatusOK Status = iota
	// StatusProblem means the install is broken in a way a command repairs.
	// One problem is enough to make `odachi doctor` exit 1.
	StatusProblem
	// StatusWarning is worth saying and not worth failing for: nothing is broken
	// yet. A warning never changes the exit code.
	StatusWarning
)

// String is the word a human reads at the start of the line.
func (s Status) String() string {
	switch s {
	case StatusProblem:
		return "bad"
	case StatusWarning:
		return "warning"
	default:
		return "ok"
	}
}

// Check is one line of the report: what was looked at, what was found, and —
// when something is wrong — how to fix it.
type Check struct {
	Number int
	Title  string
	Status Status
	Detail string
	// Fix is the phrase a human acts on. It is a phrase and not only a command
	// because one of the problems (a name a shell cannot resolve) is a shell
	// setting, and the honest repair is the exact line to paste.
	Fix string
}

// OK is a check that found nothing to fix. The detail says what it found, so a
// green line is still information and not a tick.
func OK(number int, detail string) Check {
	return Check{Number: number, Title: title(number), Status: StatusOK, Detail: detail}
}

// Broken is a check that found a problem, with the phrase that repairs it.
func Broken(number int, detail, fix string) Check {
	return Check{Number: number, Title: title(number), Status: StatusProblem, Detail: detail, Fix: fix}
}

// Warning is a check that found something worth saying and nothing broken.
func Warning(number int, detail string) Check {
	return Check{Number: number, Title: title(number), Status: StatusWarning, Detail: detail}
}

// Line renders one check as the single line a human reads.
func Line(c Check) string {
	line := fmt.Sprintf("%s %d · %s", c.Status, c.Number, c.Title)
	if c.Detail != "" {
		line += ": " + c.Detail
	}
	if c.Status == StatusProblem && c.Fix != "" {
		line += "; " + c.Fix
	}
	return line
}

// Tally counts what the run found. The two numbers are separate because they
// mean different things: a problem fails the run, a warning does not.
func Tally(checks []Check) (problems, warnings int) {
	for _, c := range checks {
		switch c.Status {
		case StatusProblem:
			problems++
		case StatusWarning:
			warnings++
		}
	}
	return problems, warnings
}

// Code is the exit code: 1 as soon as one check found a problem, 0 otherwise.
func Code(checks []Check) int {
	if problems, _ := Tally(checks); problems > 0 {
		return 1
	}
	return 0
}

// Summary is the closing line: how many checks ran and what they found, so a
// human reads the verdict without counting lines.
func Summary(checks []Check) string {
	problems, warnings := Tally(checks)
	switch {
	case problems == 0 && warnings == 0:
		return fmt.Sprintf("all clear: %d checks, no problems", len(checks))
	case problems == 0:
		return fmt.Sprintf("all clear: %d checks, %s, no problems", len(checks), plural(warnings, "warning", "warnings"))
	default:
		return fmt.Sprintf("%s of %d checks: run the command each line names", plural(problems, "problem", "problems"), len(checks))
	}
}

// The one situation a folder that is not a project yet gets. It is not one of the
// checks — those all assume a project — and it has no number, because the
// number is the line's number and numbering it would rename every check after it.
const (
	// NotAProjectMessage says the folder was never started. It is a fact and not
	// a verdict: nothing is broken here yet.
	NotAProjectMessage = "this folder is not an odachi project yet"
	// NotAProjectFix is the one step: the two commands in the order a person runs
	// them, each saying what it leaves behind. In an empty folder the checks named
	// three different commands, and that is what a beginner cannot use.
	NotAProjectFix = "run odachi init (it leaves the rules the agent reads) and then odachi start (the interview)"
)

// NotAProject is that report: the situation, and the one step that follows it.
// Two lines, because the whole point is that a beginner reads one step.
func NotAProject() (message string, lines []string) {
	return NotAProjectMessage, []string{NotAProjectMessage, NotAProjectFix}
}

// title is the English name of one check. Unknown numbers have no name: a
// report may not invent a check that does not exist.
func title(number int) string {
	switch number {
	case CheckVersion:
		return "the binary version"
	case CheckOnPath:
		return "odachi resolves from a shell"
	case CheckGit:
		return "there is .git"
	case CheckAgents:
		return "the odachi block is in AGENTS.md"
	case CheckHook:
		return "the pre-commit hook is odachi's"
	case CheckHookExe:
		return "the binary the hook names exists"
	case CheckProfile:
		return ".odachi/profile.toml resolves to a valid profile"
	case CheckBrief:
		return "brief.md exists and nothing is missing from it"
	case CheckSkills:
		return "skills and SKILLS.md"
	case CheckIgnored:
		return "the review store and proof/ are ignored"
	case CheckVoice:
		return "the block's voice matches the profile"
	case CheckTaskDocs:
		return "the task documents odachi reads"
	default:
		return ""
	}
}

// HookMarker starts the line that names a pre-commit hook as odachi's.
// `odachi init` writes it right after the shebang and never replaces a hook
// without it; doctor reads the same string the writer writes, and the CLI test
// asserts the two agree by running `init` and then `doctor`.
const HookMarker = "# odachi:"

// The two comments that delimit the rules block in AGENTS.md. The agent reads
// the rules only when both are there: one marker alone is a truncated block,
// and a truncated block is not a block.
//
// The renderer owns the strings: `odachi init` writes them and doctor reads them
// back, so a copy here would be a second owner of the one text an agent obeys.
const (
	AgentsBegin = rules.Begin
	AgentsEnd   = rules.End
)

// IsOurHook reports whether this hook is the one `odachi init` installs.
func IsOurHook(hook string) bool {
	return strings.Contains(hook, HookMarker)
}

// HookExe reads the binary the hook runs, from the `ODACHI='...'` line the
// generator writes. Empty means the hook names no binary at all, and the caller
// falls back to `odachi` from the PATH.
func HookExe(hook string) string {
	const prefix = "ODACHI="

	for _, raw := range strings.Split(strings.ReplaceAll(hook, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		value = strings.Trim(value, `'"`)
		return strings.TrimSpace(value)
	}
	return ""
}

// HasBlock reports whether AGENTS.md carries the rules block, by the renderer's
// own rule: a marker that does not own its own line is not a block this tool
// wrote, and asking that question twice with two answers is how one check says
// yes and the next one says no about the same file.
func HasBlock(agents string) bool { return rules.HasBlock(agents) }

// PathFix is the exact line a human pastes to make `odachi` reachable from a
// shell. It names the folder of the binary they already have: "put it in the
// PATH" is the advice that leaves a beginner exactly where they were, and a real
// session spent seven shell calls looking for the binary instead of doing any
// work.
func PathFix(exe string) string {
	exe = strings.TrimSpace(exe)
	dir := strings.TrimSpace(filepath.Dir(exe))
	if exe == "" || dir == "." || dir == string(filepath.Separator) || dir == "" {
		return `add the folder where you installed odachi to the PATH (in sh: export PATH="<folder>:$PATH")`
	}
	return `add that folder to the PATH (in sh: export PATH="` + filepath.ToSlash(dir) + `:$PATH")`
}

// plural is how an English sentence counts things.
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
