package review

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	domain "github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
)

// This file runs the deterministic tools the project declared in odd/review.toml,
// once per review, before any lens reads the candidate.
//
// The tools are the project's, never odachi's: a linter whose ruleset nobody chose
// is a stream of rows nobody agreed to pay for, and this binary does not know what
// language it is standing in. A project that declares none pays nothing and reads
// the prompt it always read.
//
// The declaration travels with the repository and the agreement to execute it does
// not: a cloned review.toml must not be able to run anything on the machine that
// cloned it. `odachi review checks approve` writes that agreement once, in the
// clone's store, about the exact list that is declared; a list that changes is a
// list nobody approved and the review runs none until the human says so again.

// approvalFiles is the smallest port the approval needs: whether the record is
// there and, when it is, what it says. Both the review's Files and the switch's
// ModeFiles answer it.
type approvalFiles interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
}

// plural is the word for a count: one check, two checks.
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// CheckOptions is what the two small verbs need: the project, its store, and -for
// the approval- who agreed and when.
type CheckOptions struct {
	Dir   string
	Store string
	By    string
	Now   time.Time
}

// Approve records this clone's agreement to run the tools the project declares. It
// refuses a project that declares none (nothing to approve) and a list the domain
// refuses (a tool the review could not run).
func Approve(opts CheckOptions, files ModeFiles, codec ModeCodec) Result {
	if strings.TrimSpace(opts.Store) == "" {
		return blocked("this folder has no clone store, so there is nowhere to record the approval: run odachi review inside a git repository")
	}

	config, problem := projectChecks(opts.Dir, files, codec)
	if problem != nil {
		return *problem
	}
	if len(config.Checks) == 0 {
		return blocked("odd/review.toml declares no checks: there is nothing to approve, and a project that declares none runs none")
	}
	if strings.TrimSpace(opts.By) == "" {
		return blocked("the approval does not say who agreed: run it from the clone that holds your account")
	}

	when := opts.Now
	if when.IsZero() {
		when = time.Now()
	}
	approval := domain.NewCheckApproval(config.Checks, opts.By, when.UTC().Format(time.RFC3339))

	path := filepath.Join(opts.Store, store.Checks)
	if err := files.Write(path, domain.RenderCheckApproval(approval)); err != nil {
		return blocked(fmt.Sprintf("could not write %s: %s", filepath.ToSlash(store.Checks), oneLine(err.Error())))
	}

	names := make([]string, 0, len(config.Checks))
	for _, check := range config.Checks {
		names = append(names, check.Name)
	}
	return Result{
		Code: CodeOK,
		Message: fmt.Sprintf("approved %d %s in this clone: %s; the next review runs them before any lens",
			len(names), plural(len(names), "check", "checks"), strings.Join(names, ", ")),
	}
}

// CheckStatus answers what the project declared and whether this clone agreed to
// run it, without running anything.
func CheckStatus(opts CheckOptions, files ModeFiles, codec ModeCodec) Result {
	config, problem := projectChecks(opts.Dir, files, codec)
	if problem != nil {
		return *problem
	}
	if len(config.Checks) == 0 {
		return Result{
			Code:    CodeOK,
			Message: "odd/review.toml declares no checks: this review runs no tools before the lenses",
		}
	}

	lines := make([]string, 0, len(config.Checks)+2)
	for _, check := range config.Checks {
		lines = append(lines, fmt.Sprintf("%s · %s · %s", check.Name, check.Severity, check.Command))
	}

	approval, exists, problem := readApproval(opts.Store, files)
	if problem != nil {
		return *problem
	}
	header := "not approved:"
	switch {
	case exists && approval.Covers(config.Checks):
		header = fmt.Sprintf("approved: %d %s on %s, by %s",
			len(config.Checks), plural(len(config.Checks), "check", "checks"), approval.At, approval.By)
	case exists:
		header = "not approved: this clone approved a different list"
	}
	lines = append([]string{header, "the review runs them before any lens"}, lines...)
	if !exists || !approval.Covers(config.Checks) {
		lines = append(lines, "odachi review checks approve")
	}
	return Result{Code: CodeOK, Message: header, Lines: lines}
}

// projectChecks is the declared list of this project, read through the same codec
// the review uses: a file the verb cannot read blocks with its path, never as
// "no tools declared".
func projectChecks(dir string, files ModeFiles, codec ModeCodec) (domain.ProjectConfig, *Result) {
	config, err := Config(dir, files, codec)
	if err != nil {
		problem := blocked(oneLine(err.Error()))
		return domain.ProjectConfig{}, &problem
	}
	return config, nil
}

// readApproval reads the clone's agreement. A record that cannot be read blocks
// instead of reading as absence: a clone that believes its tools run while they do
// not is the failure this file exists to prevent.
func readApproval(storeRoot string, files approvalFiles) (domain.CheckApproval, bool, *Result) {
	if strings.TrimSpace(storeRoot) == "" {
		return domain.CheckApproval{}, false, nil
	}

	path := filepath.Join(storeRoot, store.Checks)
	exists, err := files.Exists(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not check %s: %s", filepath.ToSlash(store.Checks), oneLine(err.Error())))
		return domain.CheckApproval{}, false, &problem
	}
	if !exists {
		return domain.CheckApproval{}, false, nil
	}

	data, err := files.Read(path)
	if err != nil {
		problem := blocked(fmt.Sprintf("could not read %s: %s", filepath.ToSlash(store.Checks), oneLine(err.Error())))
		return domain.CheckApproval{}, false, &problem
	}

	approval, err := domain.ParseCheckApproval(data)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; remove " + filepath.ToSlash(store.Checks) + " and approve the list again")
		return domain.CheckApproval{}, false, &problem
	}
	return approval, true, nil
}

// ranChecks runs every declared check on the candidate and answers what each one
// said, plus the reason nothing ran when nothing did. It runs once per invocation:
// the prompt of every model group and every part of a split candidate reads this
// same value, because a tool that ran twice could answer twice.
//
// The answer is never a refusal. A tool that exits non-zero, times out or cannot
// start is a row the verdict carries; blocking the run would let one broken
// command hide everything the other tools said.
func (r *runner) ranChecks() ([]domain.CheckResult, string, *Result) {
	if r.ran != nil {
		return *r.ran, r.off, nil
	}

	timeout, err := domain.NormalizeCheckTimeout(r.opts.CheckTimeout)
	if err != nil {
		problem := blocked(oneLine(err.Error()) + "; fix odd/review.toml and run odachi review again")
		return nil, "", &problem
	}

	results := make([]domain.CheckResult, 0, len(r.opts.Checks))
	if len(r.opts.Checks) > 0 {
		approval, exists, problem := readApproval(r.opts.Store, r.files)
		if problem != nil {
			return nil, "", problem
		}
		if !exists || !approval.Covers(r.opts.Checks) {
			// Nothing runs, and the prompt says so: a review that silently skipped
			// the tools would report on a tree nobody measured.
			r.off = offReason(len(r.opts.Checks), exists)
			r.ran = &results
			return results, r.off, nil
		}
		for _, check := range r.opts.Checks {
			results = append(results, r.runCheck(check, timeout))
		}
	}

	r.ran = &results
	return results, r.off, nil
}

// offReason is why the declared tools did not run: either this clone never agreed
// to them, or it agreed to a list the project has changed since.
func offReason(declared int, approved bool) string {
	if approved {
		return fmt.Sprintf("the project declares %d %s and this clone approved a different list: run `odachi review checks approve` to run them",
			declared, plural(declared, "check", "checks"))
	}
	return fmt.Sprintf("the project declares %d %s and this clone has not approved them: run `odachi review checks approve` to run them",
		declared, plural(declared, "check", "checks"))
}

// runCheck runs one tool in the candidate the prompt names and turns its ending
// into a status. The command runs in the project directory, exactly the bytes the
// tree hash describes: a check that edits the candidate is out of contract, and the
// verdict is filed under the tree read before it.
func (r *runner) runCheck(check domain.Check, timeout int) domain.CheckResult {
	result := domain.CheckResult{Name: check.Name, Command: check.Command, Severity: check.Severity}

	stdout, stderr, exitCode, timedOut, err := r.shell.Run(r.opts.Dir, check.Command, timeout)
	result.Exit = exitCode

	switch {
	case err != nil:
		// The command did not start: a tool nobody installed, or a shell that
		// refused the line. It is a row, and the note says which of the two.
		result.Status = domain.CheckBroken
		result.Note = oneLine(err.Error())
	case timedOut:
		result.Status = domain.CheckTimedOut
		result.Note = fmt.Sprintf("stopped after %d s", timeout)
	case exitCode != 0:
		result.Status = domain.CheckFailed
	default:
		result.Status = domain.CheckPassed
	}

	result.Output = domain.ClipCheckOutput(joinOutput(stdout, stderr))
	return result
}

// joinOutput is what the tool printed, in the order a reader expects it: its
// answer and then its complaints. A failing tool usually says why on stderr, and
// dropping either half would hide the line the human needs.
func joinOutput(stdout, stderr string) string {
	out := strings.TrimRight(stdout, "\n")
	said := strings.TrimRight(stderr, "\n")
	switch {
	case out == "":
		return said
	case said == "":
		return out
	}
	return out + "\n" + said
}

// checkFindings is what the tools said, as the rows the verdict carries: one per
// failing check, at the severity the project declared, with the tool's own words
// as the evidence. A tool that answered clean filed nothing, and no tool filed
// anything when this clone has not approved the list.
//
// The row is the reviewer's peer, not its pupil: it blocks by the same rule, and a
// person dismisses it with a reason like any other.
func (r *runner) checkFindings() ([]domain.Finding, *Result) {
	results, off, problem := r.ranChecks()
	if problem != nil {
		return nil, problem
	}
	if off != "" {
		return nil, nil
	}

	rows := make([]domain.Finding, 0, len(results))
	for _, result := range results {
		if !result.Failed() {
			continue
		}
		rows = append(rows, domain.Finding{
			Check:    result.Name,
			Severity: result.Severity,
			// The tool measured these bytes, and causal is what makes it able to
			// block: what a linter reports on the candidate is the candidate's.
			Causal:   true,
			Evidence: domain.EvidenceCited,
			Why:      checkWhy(result),
		})
	}
	return rows, nil
}

// checkWhy is the row's evidence: how the tool ended and what it said. The first
// line stands alone so a person reading the verdict sees the reason before the
// noise.
func checkWhy(result domain.CheckResult) string {
	var out strings.Builder
	switch result.Status {
	case domain.CheckFailed:
		fmt.Fprintf(&out, "the check %s exited %d", result.Name, result.Exit)
	case domain.CheckTimedOut:
		out.WriteString("the check " + result.Name + " was stopped by the ceiling")
		if result.Note != "" {
			out.WriteString(" (" + result.Note + ")")
		}
	default:
		out.WriteString("the check " + result.Name + " could not run")
		if result.Note != "" {
			out.WriteString(": " + result.Note)
		}
	}
	if output := strings.TrimSpace(result.Output); output != "" {
		out.WriteString("\n" + output)
	}
	return out.String()
}

// appendMissing adds the rows a verdict does not already carry, by the same
// identity the store files every finding under: a second part of one run, or a
// draw that paid again, cannot file the same check twice.
func appendMissing(findings, extra []domain.Finding) []domain.Finding {
	seen := make(map[string]bool, len(findings))
	for _, finding := range findings {
		seen[findingKey(finding)] = true
	}
	for _, row := range extra {
		if seen[findingKey(row)] {
			continue
		}
		findings = append(findings, row)
	}
	return findings
}
