package skills_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/skill"
	"github.com/biggs-100/odachi/skills"
)

// declaredReference and declaredAsset mirror the grammar's backticked paths so
// the payload-level manifest check can name a file the folder does not carry.
var (
	declaredReference = regexp.MustCompile("`(references/[A-Za-z0-9._/-]+\\.md)`")
	declaredAsset     = regexp.MustCompile("`(assets/[A-Za-z0-9._-]+)`")
)

// structuralProblems is the one structural check for an embedded payload: it
// parses SKILL.md through the domain grammar with the folder contents attached,
// then adds the manifest checks the grammar cannot see from text alone.
func structuralProblems(payload skills.Skill) []skill.Problem {
	problems := []skill.Problem{}
	add := func(path, format string, args ...any) {
		problems = append(problems, skill.Problem{
			File:   payload.Name + "/" + path,
			Reason: fmt.Sprintf(format, args...),
		})
	}

	references := []string{}
	assets := []string{}
	present := map[string]bool{}
	hasMain := false
	for _, file := range payload.Files {
		present[file.Path] = true
		switch {
		case file.Path == skills.SkillFile:
			hasMain = true
		case strings.HasPrefix(file.Path, "references/") && strings.HasSuffix(file.Path, ".md"):
			references = append(references, file.Path)
		case strings.HasPrefix(file.Path, "assets/"):
			assets = append(assets, file.Path)
		default:
			add(file.Path, "the file path %q is neither SKILL.md, a references/*.md file nor an assets/ file", file.Path)
		}
	}
	if !hasMain {
		add(skills.SkillFile, "the payload has no SKILL.md")
	}

	body := string(payload.Main())
	for _, path := range declaredPaths(body, declaredReference) {
		if !present[path] {
			add(path, "the declared reference %q is not a file in the payload", path)
		}
	}
	for _, path := range declaredPaths(body, declaredAsset) {
		if !present[path] {
			add(path, "the declared asset %q is not a file in the payload", path)
		}
	}

	parsed := skill.Parse(payload.Name, payload.Name+"/"+skills.SkillFile, body)
	parsed.References = references
	parsed.Assets = assets
	problems = append(problems, parsed.Problems()...)
	return problems
}

// declaredPaths returns the distinct backticked paths of one kind in text.
func declaredPaths(text string, pattern *regexp.Regexp) []string {
	seen := map[string]bool{}
	paths := []string{}
	for _, match := range pattern.FindAllStringSubmatch(text, -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			paths = append(paths, match[1])
		}
	}
	return paths
}

// formatProblems renders problems for a failure message, one per line.
func formatProblems(problems []skill.Problem) string {
	parts := make([]string, 0, len(problems))
	for _, p := range problems {
		parts = append(parts, p.File+": "+p.Reason)
	}
	return strings.Join(parts, "\n")
}

// problemsMention reports whether any problem names text.
func problemsMention(problems []skill.Problem, text string) bool {
	for _, p := range problems {
		if strings.Contains(p.Reason, text) || strings.Contains(p.File, text) {
			return true
		}
	}
	return false
}

// TestStructuralCheckCatchesABrokenReference: a body that names a reference the
// folder does not carry is a dead link, and a vacuous structural check would let
// it through. This control fails while the helper returns no problems.
func TestStructuralCheckCatchesABrokenReference(t *testing.T) {
	const text = "---\n" +
		"name: probe\n" +
		"description: Trigger: when the structural check must catch a dead reference\n" +
		"---\n\n" +
		"## When\n\nwhen probing the structural check.\n\n" +
		"## Rules\n\n- name one reference the folder does not carry.\n\n" +
		"## Output\n\na list of problems.\n\n" +
		"## References\n\n- `references/missing.md` — the file that is not there.\n\n" +
		"## Never\n\n- claim the payload is complete.\n"

	payload := skills.Skill{
		Name:  "probe",
		Files: []skills.File{{Path: skills.SkillFile, Content: []byte(text)}},
	}

	problems := structuralProblems(payload)
	if !problemsMention(problems, "references/missing.md") {
		t.Fatalf("the structural check missed the dead reference; problems: %s", formatProblems(problems))
	}
}

// TestEmbeddedSkillsAreStructurallyValid: every skill the binary installs must
// satisfy the same grammar `--check` applies on disk, and its manifest must be
// complete. A payload nobody parses is a payload nobody has checked.
func TestEmbeddedSkillsAreStructurallyValid(t *testing.T) {
	all, err := skills.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	if len(all) == 0 {
		t.Fatal("All() returned no skills: an empty payload checks nothing")
	}

	for _, payload := range all {
		problems := structuralProblems(payload)
		if len(problems) != 0 {
			t.Errorf("skill %q is not structurally valid:\n%s", payload.Name, formatProblems(problems))
		}
	}
}
