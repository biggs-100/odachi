package review

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// This file owns the deterministic tools a project declares in odd/review.toml:
// the commands a review runs on the candidate before any lens reads it, and the
// severity a failing one is filed at.
//
// Nothing is built in and nothing is detected. The project names its own tools,
// because a linter with rules the project never chose is a stream of rows nobody
// agreed to pay for, and because odachi does not know what language it is
// standing in. A project that declares none runs none, and costs the same as it
// did before this file existed.

// Check is one deterministic tool the project declared: the name its row and its
// finding carry, the command the review runs, and the severity a failure is
// worth. The severity has no default on purpose -a formatter and a vulnerability
// scanner do not weigh the same, and only the project knows which it declared-.
type Check struct {
	Name     string
	Command  string
	Severity Severity
}

const (
	// DefaultCheckTimeout is how long one check may run when the project did not
	// choose: a tool that hangs must not hang the review, and five minutes is
	// enough for a linter on a candidate of this size.
	DefaultCheckTimeout = 300
	// MaxCheckTimeout is the ceiling a project may ask for. A number nobody
	// reaches is a typo wearing a decision's clothes.
	MaxCheckTimeout = 3600
)

// NormalizeChecks reads the declared tools and refuses each one the review could
// not run: a check without a name has no row to file, without a command has
// nothing to execute, and without a severity cannot be paid at a tier.
//
// The name is the identity, so it arrives trimmed and lowercased and no two
// checks may share one: two tools under one name would pay one row and lose the
// other. The command is the project's own text and travels as written -it is a
// shell line, and normalizing it would be rewriting somebody's tool-.
func NormalizeChecks(raw []Check) ([]Check, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	out := make([]Check, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for i, check := range raw {
		name := strings.ToLower(strings.TrimSpace(check.Name))
		if name == "" {
			return nil, fmt.Errorf("check %d does not say name: the name is the row the review files", i+1)
		}
		if !checkNameLike(name) {
			return nil, fmt.Errorf("check %q has a character a name cannot carry: use letters, digits, dot, dash or underscore", check.Name)
		}
		if seen[name] {
			return nil, fmt.Errorf("the check %q appears twice: two tools under one name would pay one row and lose the other", name)
		}
		seen[name] = true

		command := strings.TrimSpace(check.Command)
		if command == "" {
			return nil, fmt.Errorf("the check %q does not say command: a check with nothing to run measures nothing", name)
		}

		severity, err := NormalizeSeverity(check.Severity)
		if err != nil {
			return nil, fmt.Errorf("the check %q: %w", name, err)
		}

		out = append(out, Check{Name: name, Command: command, Severity: severity})
	}
	return out, nil
}

// NormalizeSeverity reads the severity a check's failure is filed at. Case is not
// the human's problem, so the word is matched in any case and answered in the
// vocabulary the verdict uses; a word outside it is refused naming the four.
func NormalizeSeverity(value Severity) (Severity, error) {
	normalized := Severity(strings.ToUpper(strings.TrimSpace(string(value))))
	if !normalized.Valid() {
		return "", fmt.Errorf("severity %q is not one of BLOCKER, CRITICAL, WARNING, SUGGESTION", strings.TrimSpace(string(value)))
	}
	return normalized, nil
}

// checkNameLike reports whether the text can be a check's name: the alphabet of a
// row that is printed on one line and read by a person.
func checkNameLike(name string) bool {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case r == '.', r == '-', r == '_':
		default:
			return false
		}
	}
	return name != ""
}

// NormalizeCheckTimeout reads the ceiling one check may run inside. Zero is the
// absence of a choice and means the number odachi ships; a negative number or one
// above the ceiling is a typo, because a review that waits an hour for a linter
// is a review nobody runs.
func NormalizeCheckTimeout(value int) (int, error) {
	switch {
	case value < 0:
		return 0, fmt.Errorf("check_timeout_s is %d: a ceiling is zero (the default) or a positive number of seconds", value)
	case value == 0:
		return DefaultCheckTimeout, nil
	case value > MaxCheckTimeout:
		return 0, fmt.Errorf("check_timeout_s is %d: the most a check may run is %d seconds", value, MaxCheckTimeout)
	}
	return value, nil
}

// CheckStatus is what one declared tool answered when the review ran it. Three of
// the four are a failure the verdict has to carry: a tool that could not start is
// not a tool that passed.
type CheckStatus string

const (
	// CheckPassed is the tool ran and exited zero.
	CheckPassed CheckStatus = "passed"
	// CheckFailed is the tool ran and exited non-zero.
	CheckFailed CheckStatus = "failed"
	// CheckTimedOut is the tool was still running at the ceiling and was stopped.
	CheckTimedOut CheckStatus = "timed out"
	// CheckBroken is the tool could not start at all: the command does not exist,
	// or the operating system refused it.
	CheckBroken CheckStatus = "could not run"
)

// CheckResult is one declared tool's answer about one candidate: which tool, what
// it said, and how it ended. Output is clipped, never summarized: the tool's own
// words are the evidence, and a summary would be odachi's opinion about them.
type CheckResult struct {
	Name    string
	Command string
	// Severity is the severity the project declared: what a failure is worth is
	// its decision, and the row the verdict files carries it.
	Severity Severity
	Status   CheckStatus
	// Exit is the tool's own exit code, zero when it did not produce one.
	Exit int
	// Output is what the tool printed, clipped by ClipCheckOutput.
	Output string
	// Note is why the tool could not start, when it could not. Empty otherwise.
	Note string
}

// Failed reports whether this answer is a failure the verdict must carry. Only a
// check that ran and exited zero is a pass.
func (c CheckResult) Failed() bool {
	return c.Status != CheckPassed
}

const (
	// CheckOutputLines is how many lines of a tool's output the prompt and the
	// verdict keep. A linter over a package can print thousands, and a prompt that
	// pays for them is a prompt the reviewer stops reading.
	CheckOutputLines = 40
	// CheckOutputBytes is the ceiling for one tool's output when its lines are
	// enormous: one minified line is not forty lines, and the clip is charged in
	// the currency that was spent.
	CheckOutputBytes = 4096
)

// ClipCheckOutput keeps the head of what a tool printed and says how much it cut,
// in lines and, when one line is enormous, in bytes. What is cut is counted and
// never hidden: a row that silently drops the defect the reviewer needed is worse
// than a row that is long.
func ClipCheckOutput(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	trimmed := strings.TrimRight(normalized, "\n")
	if trimmed == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	cut := 0
	if len(lines) > CheckOutputLines {
		cut = len(lines) - CheckOutputLines
		lines = lines[:CheckOutputLines]
	}

	out := strings.Join(lines, "\n") + "\n"
	if cut > 0 {
		out += fmt.Sprintf("(%d more lines)\n", cut)
	}

	if len(out) > CheckOutputBytes {
		head := validPrefix(out, CheckOutputBytes)
		out = head + fmt.Sprintf("\n(%d bytes cut)\n", len(out)-len(head))
	}
	return out
}

// validPrefix is the longest prefix of text that fits in bytes and still ends on a
// rune boundary: cutting inside a rune produces a file the gate refuses as not
// UTF-8, and a clipped row must never be the reason a store is unreadable.
func validPrefix(text string, bytes int) string {
	if len(text) <= bytes {
		return text
	}
	head := text[:bytes]
	for len(head) > 0 && !utf8.ValidString(head) {
		head = head[:len(head)-1]
	}
	return head
}

// CheckApproval is the clone's record that its human agreed to run the tools a
// project declares. It lives in the store, never in the tree: the declaration is
// the team's and travels with the repository, and the agreement to execute it is
// the clone's. A repository's review.toml must not be able to run anything by
// being cloned.
//
// Fingerprint is the list itself -the name, the command and the severity of each
// tool- so an approval covers the tools that were approved and nothing else: one
// word changed in a command is another list, and the human is asked again.
type CheckApproval struct {
	Fingerprint string
	At          string
	By          string
	Count       int
}

// NewCheckApproval reads a list and records the clone's agreement to run it.
func NewCheckApproval(checks []Check, by, at string) CheckApproval {
	count := len(checks)
	if normalized, err := NormalizeChecks(checks); err == nil {
		count = len(normalized)
	}
	return CheckApproval{
		Fingerprint: ChecksFingerprint(checks),
		At:          strings.TrimSpace(at),
		By:          strings.TrimSpace(by),
		Count:       count,
	}
}

// Covers reports whether this approval is about exactly this list, in this order.
// An approval with no fingerprint covers nothing, and a list with no tools needs
// no approval: the absence of a declaration runs nothing.
func (a CheckApproval) Covers(checks []Check) bool {
	if len(checks) == 0 {
		return true
	}
	if strings.TrimSpace(a.Fingerprint) == "" {
		return false
	}
	return a.Fingerprint == ChecksFingerprint(checks)
}

// ChecksFingerprint is the identity of one declared list: the tools in the order
// the file writes them, each with its name, its command and the severity it is
// paid at. Nothing is hashed twice, so two runs of the same file agree.
//
// Every field is written with its length in front, because the command is a
// human's shell line and can carry any character at all: without the length, one
// check whose command contains the delimiter would hash the same as two checks,
// and an approval would cover a list nobody wrote.
func ChecksFingerprint(checks []Check) string {
	if len(checks) == 0 {
		return ""
	}

	normalized, err := NormalizeChecks(checks)
	if err != nil {
		normalized = checks
	}

	var lines strings.Builder
	for _, check := range normalized {
		writeField(&lines, string(check.Name))
		writeField(&lines, string(check.Command))
		writeField(&lines, string(check.Severity))
		lines.WriteString("\n")
	}

	sum := sha256.Sum256([]byte(lines.String()))
	return hex.EncodeToString(sum[:])
}

// writeField writes one field with its length in front: the length is what makes
// two different lists impossible to spell the same way.
func writeField(out *strings.Builder, value string) {
	out.WriteString(strconv.Itoa(len(value)))
	out.WriteString(":")
	out.WriteString(value)
}

// RenderCheckApproval writes the record the clone keeps. It is the plain header
// shape the proofs use, because it is a handful of facts about one run and a
// grammar would be more surface than the fact.
func RenderCheckApproval(approval CheckApproval) []byte {
	var out strings.Builder
	out.WriteString("# odachi: the deterministic tools this clone agreed to run; the fingerprint is the list, so changing it asks again\n")
	out.WriteString("# fingerprint " + approval.Fingerprint + "\n")
	out.WriteString("# count " + strconv.Itoa(approval.Count) + "\n")
	out.WriteString("# at " + approval.At + "\n")
	out.WriteString("# by " + approval.By + "\n")
	return []byte(out.String())
}

// ParseCheckApproval reads the record back. A file this binary cannot read is
// refused instead of read as absence: a clone that believes its tools run while
// they do not is the one failure this record exists to prevent.
func ParseCheckApproval(data []byte) (CheckApproval, error) {
	var approval CheckApproval
	counted := false

	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		field, value, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		switch strings.TrimSpace(field) {
		case "fingerprint":
			approval.Fingerprint = strings.TrimSpace(value)
		case "count":
			count, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return CheckApproval{}, fmt.Errorf("the check approval says count %q: a count is a number", strings.TrimSpace(value))
			}
			approval.Count = count
			counted = true
		case "at":
			approval.At = strings.TrimSpace(value)
		case "by":
			approval.By = strings.TrimSpace(value)
		}
	}

	if strings.TrimSpace(approval.Fingerprint) == "" {
		return CheckApproval{}, fmt.Errorf("the check approval does not say fingerprint: without the list it covers nothing")
	}
	if !counted {
		return CheckApproval{}, fmt.Errorf("the check approval does not say count: the record says how many tools it agreed to run")
	}
	if approval.Count < 0 {
		return CheckApproval{}, fmt.Errorf("the check approval says count %d: a list holds zero or more tools", approval.Count)
	}
	return approval, nil
}
