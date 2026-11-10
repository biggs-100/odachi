package review_test

import (
	"strings"
	"testing"

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

// TestSetConfigBacksUpBeforeReplacing: the previous file survives the write.
func TestSetConfigBacksUpBeforeReplacing(t *testing.T) {
	files := newFakeFiles(map[string]string{domainreview.ModePath: seeded})

	review.SetConfig(review.ConfigOptions{Dir: dir, Mode: review.ModeOff}, files, modeCodec{})

	if len(files.backups) != 1 || files.backups[0] != fullPath(domainreview.ModePath) {
		t.Fatalf("backups = %v, want the file before the write", files.backups)
	}
}
