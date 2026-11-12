package review

import (
	"fmt"
	"strings"
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
