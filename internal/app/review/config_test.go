package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// seeded is a review file with every field the wizard does not touch.
const seeded = `mode = "on"
max_prompt_kb = 120
keep_verdicts = 4
keep_prompts = 2

[lenses]
risk = "old/model"
`

// TestSetConfigKeepsTheNumbersTheWizardDoesNotTouch: the prompt limit and the
// two retention lifetimes survive a write of the switch and the lens table.
func TestSetConfigKeepsTheNumbersTheWizardDoesNotTouch(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	result := review.SetConfig(review.ConfigOptions{
		Dir:    dir,
		Mode:   review.ModeOff,
		Models: map[string]string{"reliability": "new/model"},
	}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	for _, want := range []string{
		`mode = "off"`, `max_prompt_kb = 120`, `keep_verdicts = 4`, `keep_prompts = 2`,
		`reliability = "new/model"`,
	} {
		if !strings.Contains(stored, want) {
			t.Errorf("the write lost %q:\n%s", want, stored)
		}
	}
	if strings.Contains(stored, "old/model") {
		t.Errorf("the old lens table survived a replacement:\n%s", stored)
	}
}

// TestSetConfigKeepsTheLensTableWhenModelsIsNil: a write that only flips the
// switch must not delete the table.
func TestSetConfigKeepsTheLensTableWhenModelsIsNil(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Mode: review.ModeOn}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	if !strings.Contains(stored, `risk = "old/model"`) {
		t.Errorf("the switch rewrite dropped the lens table:\n%s", stored)
	}
}

// TestSetConfigClearsTheTableWithAnEmptyMap: the wizard can remove every model,
// and the file then has no table.
func TestSetConfigClearsTheTableWithAnEmptyMap(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	result := review.SetConfig(review.ConfigOptions{
		Dir:    dir,
		Models: map[string]string{},
	}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	if strings.Contains(stored, "old/model") || strings.Contains(stored, "[lenses]") {
		t.Errorf("the empty table was not cleared:\n%s", stored)
	}
}

// TestSetConfigRefusesAModelThatIsNotAnIdentifier: nothing is written when the
// model would break the reviewer command.
func TestSetConfigRefusesAModelThatIsNotAnIdentifier(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	result := review.SetConfig(review.ConfigOptions{
		Dir:    dir,
		Models: map[string]string{"risk": "a b"},
	}, files, modeCodec{})

	if result.Code == 0 {
		t.Fatalf("code = 0, want blocked")
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Fatalf("writes = %v, backups = %v; want none", files.written(), files.backups)
	}
}

// TestSetConfigWritesAMissingFile: the wizard can create the file, and there is
// nothing to back up.
func TestSetConfigWritesAMissingFile(t *testing.T) {
	files := newFakeFiles(nil)

	result := review.SetConfig(review.ConfigOptions{
		Dir:    dir,
		Mode:   review.ModeOn,
		Models: map[string]string{"reliability": "opencode-go/deepseek-v4.1-flash"},
	}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if got := files.written(); len(got) != 1 || got[0] != domainreview.ModePath {
		t.Fatalf("writes = %v, want [%s]", got, domainreview.ModePath)
	}
	if len(files.backups) != 0 {
		t.Fatalf("backups = %v, want none: there was no file", files.backups)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	for _, want := range []string{`mode = "on"`, `opencode-go/deepseek-v4.1-flash`} {
		if !strings.Contains(stored, want) {
			t.Errorf("the new file lost %q:\n%s", want, stored)
		}
	}
}

// TestSetConfigWritesTheVerifierModel: the wizard can name who answers, and the
// lens table and the numbers it does not touch survive the write (0337).
func TestSetConfigWritesTheVerifierModel(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})
	verify := "cheap/model#fast"

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Verify: &verify}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	for _, want := range []string{
		`verify = "cheap/model#fast"`, `mode = "on"`, `max_prompt_kb = 120`,
		`keep_verdicts = 4`, `keep_prompts = 2`, `risk = "old/model"`,
	} {
		if !strings.Contains(stored, want) {
			t.Errorf("the write lost %q:\n%s", want, stored)
		}
	}
}

// TestSetConfigKeepsTheVerifierWhenItIsNil: a write that only changes the lens
// table must not delete the project's verifier.
func TestSetConfigKeepsTheVerifierWhenItIsNil(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: "mode = \"on\"\nverify = \"cheap/model\"\n",
	})

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Models: map[string]string{}}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if stored := string(files.files[fullPath(domainreview.ModePath)]); !strings.Contains(stored, `verify = "cheap/model"`) {
		t.Errorf("the verifier line was dropped:\n%s", stored)
	}
}

// TestSetConfigRemovesTheVerifierWithAnEmptyValue: clearing the row means the
// project names no verifier, and the line goes away while the lens table stays.
func TestSetConfigRemovesTheVerifierWithAnEmptyValue(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: "mode = \"on\"\nverify = \"cheap/model\"\n\n[lenses]\nrisk = \"sharp/model\"\n",
	})
	empty := ""

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Verify: &empty}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	if strings.Contains(stored, "verify") {
		t.Errorf("the verifier line survived a cleared row:\n%s", stored)
	}
	if !strings.Contains(stored, `risk = "sharp/model"`) {
		t.Errorf("clearing the verifier dropped the lens table:\n%s", stored)
	}
}

// TestSetConfigRefusesAVerifierThatIsNotAnIdentifier: the value replaces {model}
// inside a command, so it obeys the model rule and nothing is written when it
// does not.
func TestSetConfigRefusesAVerifierThatIsNotAnIdentifier(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})
	bad := "a b"

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Verify: &bad}, files, modeCodec{})

	if result.Code == 0 {
		t.Fatalf("code = 0, want blocked")
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Fatalf("writes = %v, backups = %v; want none", files.written(), files.backups)
	}
}

// TestSetConfigRendersAnUntouchedVerifierByteForByte: writing a field the wizard
// did not change leaves the file exactly as it was, verifier included (0337).
func TestSetConfigRendersAnUntouchedVerifierByteForByte(t *testing.T) {
	canonical, err := reviewtoml.Render(domainreview.ProjectConfig{
		Mode:         domainreview.ModeOn,
		Verify:       "cheap/model#fast",
		Lenses:       map[string]string{"risk": "sharp/model"},
		MaxPromptKB:  120,
		KeepVerdicts: 4,
		KeepPrompts:  2,
	})
	if err != nil {
		t.Fatalf("reviewtoml.Render: %v", err)
	}
	files := newFakeFiles(map[string]string{domainreview.ModePath: string(canonical)})

	result := review.SetConfig(review.ConfigOptions{Dir: dir, Mode: review.ModeOn}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if stored := string(files.files[fullPath(domainreview.ModePath)]); stored != string(canonical) {
		t.Errorf("an untouched file moved:\n--- want ---\n%s\n--- got ---\n%s", canonical, stored)
	}
}

// TestSetConfigBacksUpBeforeReplacing: the previous file survives the write.
func TestSetConfigBacksUpBeforeReplacing(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	review.SetConfig(review.ConfigOptions{Dir: dir, Mode: review.ModeOff}, files, modeCodec{})

	if len(files.backups) != 1 || files.backups[0] != fullPath(domainreview.ModePath) {
		t.Fatalf("backups = %v, want the file before the write", files.backups)
	}
}

// seededWithChecks is a canonical file with every scalar the wizard edits and one
// deterministic tool declared at CRITICAL. The check is the project's gate
// severity: the new section sets four scalars, not the array that owns it (0342).
func seededWithChecks(t *testing.T) string {
	t.Helper()
	canonical, err := reviewtoml.Render(domainreview.ProjectConfig{
		Mode:         domainreview.ModeOn,
		Verify:       "cheap/model#fast",
		Lenses:       map[string]string{"risk": "old/model"},
		MaxPromptKB:  120,
		KeepVerdicts: 4,
		KeepPrompts:  2,
		CheckTimeout: 45,
		Checks: []domainreview.Check{
			{Name: "lint", Command: "go vet ./...", Severity: domainreview.Critical, Remedy: "fix it"},
		},
	})
	if err != nil {
		t.Fatalf("reviewtoml.Render: %v", err)
	}
	return string(canonical)
}

// checkBlock is the [[check]] array of a file, from its header to the end. It is
// the project's gate policy, so a write of the four scalars must not move a byte
// of it (0342).
func checkBlock(file string) string {
	at := strings.Index(file, "[[check]]")
	if at < 0 {
		return ""
	}
	return file[at:]
}

// TestSetConfigSetsEachLimitAndKeepsTheRest: the wizard can set each of the four
// scalars on its own, and the other three, the switch, the lens table, the
// verifier and the whole check array survive the write byte for byte (0342).
func TestSetConfigSetsEachLimitAndKeepsTheRest(t *testing.T) {
	cases := []struct {
		key  string
		want string
		set  func(*int) review.ConfigOptions
	}{
		{"max_prompt_kb", "max_prompt_kb = 200", func(v *int) review.ConfigOptions { return review.ConfigOptions{MaxPromptKB: v} }},
		{"keep_verdicts", "keep_verdicts = 7", func(v *int) review.ConfigOptions { return review.ConfigOptions{KeepVerdicts: v} }},
		{"keep_prompts", "keep_prompts = 3", func(v *int) review.ConfigOptions { return review.ConfigOptions{KeepPrompts: v} }},
		{"check_timeout_s", "check_timeout_s = 90", func(v *int) review.ConfigOptions { return review.ConfigOptions{CheckTimeout: v} }},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			canonical := seededWithChecks(t)
			files := newFakeFiles(map[string]string{domainreview.ModePath: canonical})
			value := 200
			switch tc.key {
			case "keep_verdicts":
				value = 7
			case "keep_prompts":
				value = 3
			case "check_timeout_s":
				value = 90
			}
			opts := tc.set(&value)
			opts.Dir = dir

			result := review.SetConfig(opts, files, modeCodec{})

			if result.Code != 0 {
				t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
			}
			stored := string(files.files[fullPath(domainreview.ModePath)])
			if !strings.Contains(stored, tc.want) {
				t.Errorf("the write lost %q:\n%s", tc.want, stored)
			}
			for _, keep := range []string{
				`mode = "on"`, `verify = "cheap/model#fast"`, `risk = "old/model"`,
				"max_prompt_kb = 120", "keep_verdicts = 4", "keep_prompts = 2", "check_timeout_s = 45",
			} {
				if strings.HasPrefix(keep, tc.key) {
					continue
				}
				if !strings.Contains(stored, keep) {
					t.Errorf("the write lost %q:\n%s", keep, stored)
				}
			}
			if got := checkBlock(stored); got != checkBlock(canonical) {
				t.Errorf("the CRITICAL check moved:\n--- before ---\n%s\n--- after ---\n%s", checkBlock(canonical), got)
			}
		})
	}
}

// TestSetConfigWritesAllFourLimitsWithoutTouchingChecks: setting every scalar in
// one write still leaves the switch, the models, the verifier and the check array
// exactly as the file had them (0342).
func TestSetConfigWritesAllFourLimitsWithoutTouchingChecks(t *testing.T) {
	canonical := seededWithChecks(t)
	files := newFakeFiles(map[string]string{domainreview.ModePath: canonical})
	limit, verdicts, prompts, timeout := 200, 7, 3, 90

	result := review.SetConfig(review.ConfigOptions{
		Dir:          dir,
		MaxPromptKB:  &limit,
		KeepVerdicts: &verdicts,
		KeepPrompts:  &prompts,
		CheckTimeout: &timeout,
	}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	for _, want := range []string{
		"max_prompt_kb = 200", "keep_verdicts = 7", "keep_prompts = 3", "check_timeout_s = 90",
		`mode = "on"`, `verify = "cheap/model#fast"`, `risk = "old/model"`, `severity = "CRITICAL"`,
	} {
		if !strings.Contains(stored, want) {
			t.Errorf("the write lost %q:\n%s", want, stored)
		}
	}
	if got := checkBlock(stored); got != checkBlock(canonical) {
		t.Errorf("the CRITICAL check moved:\n--- before ---\n%s\n--- after ---\n%s", checkBlock(canonical), got)
	}
}

// TestSetConfigRefusesAnInvalidLimit: a scalar the domain cannot accept writes
// nothing at all, not even a backup, so a typo cannot half-change the gate (0342).
func TestSetConfigRefusesAnInvalidLimit(t *testing.T) {
	canonical := seededWithChecks(t)
	files := newFakeFiles(map[string]string{domainreview.ModePath: canonical})
	bad := -5

	result := review.SetConfig(review.ConfigOptions{Dir: dir, KeepVerdicts: &bad}, files, modeCodec{})

	if result.Code == 0 {
		t.Fatalf("code = 0, want blocked")
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Fatalf("writes = %v, backups = %v; want none", files.written(), files.backups)
	}
	if stored := string(files.files[fullPath(domainreview.ModePath)]); stored != canonical {
		t.Errorf("the file moved on a refused write:\n--- before ---\n%s\n--- after ---\n%s", canonical, stored)
	}
}
