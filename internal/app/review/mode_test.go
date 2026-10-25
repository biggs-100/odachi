package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	"github.com/biggs-100/odachi/internal/app/review"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// modeCodec is the real TOML adapter: the app tests exercise the contract the
// CLI wires, not a different dialect.
type modeCodec struct{}

func (modeCodec) Decode(data []byte) (domainreview.ProjectConfig, error) {
	return reviewtoml.Decode(data)
}
func (modeCodec) Render(config domainreview.ProjectConfig) ([]byte, error) {
	return reviewtoml.Render(config)
}

// TestModeStatusWithoutAFileIsOff: not writing the switch is a decision, not a
// missing setting. Status says so without creating a file.
func TestModeStatusWithoutAFileIsOff(t *testing.T) {
	files := newFakeFiles(nil)

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeStatus}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if !strings.Contains(result.Message, "review: off") || !strings.Contains(result.Message, "the default is off") {
		t.Errorf("message = %q, want the default named", result.Message)
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: status writes nothing", files.written())
	}
}

// TestModeOnWritesTheSwitch: on is a file the project carries, and the gate can
// read it without a flag.
// Flipping the switch is one field of the file: the human's lens table is the
// other one, and `review mode off` must not silently delete it.
func TestModeKeepsTheLensTable(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: "mode = \"on\"\n\n[lenses]\nrisk = \"sharp/model\"\nreliability = \"cheap/model\"\n",
	})

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeOff}, files, modeCodec{})
	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}

	stored := string(files.files[fullPath(domainreview.ModePath)])
	for _, want := range []string{`mode = "off"`, `risk = "sharp/model"`, `reliability = "cheap/model"`} {
		if !strings.Contains(stored, want) {
			t.Errorf("the switch rewrite lost %q:\n%s", want, stored)
		}
	}
}

func TestModeOnWritesTheSwitch(t *testing.T) {
	files := newFakeFiles(nil)

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeOn}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if got := files.written(); len(got) != 1 || got[0] != domainreview.ModePath {
		t.Fatalf("writes = %v, want [%s]", got, domainreview.ModePath)
	}
	stored := string(files.files[fullPath(domainreview.ModePath)])
	if !strings.Contains(stored, `mode = "on"`) {
		t.Errorf("stored switch = %q, want mode = \"on\"", stored)
	}
	if len(files.backups) != 0 {
		t.Fatalf("backups = %v, want none: there was no file to back up", files.backups)
	}
}

// TestModeOffBacksUpTheExistingSwitch: switching a human's file changes it, so
// the old answer survives as a backup.
func TestModeOffBacksUpTheExistingSwitch(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: `# mine
mode = "on"
`,
	})

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeOff}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(files.backups) != 1 {
		t.Fatalf("backups = %v, want one", files.backups)
	}
	if got := string(files.files[fullPath(domainreview.ModePath)]); !strings.Contains(got, `mode = "off"`) {
		t.Errorf("stored switch = %q, want mode = \"off\"", got)
	}
}

// TestModeOnWhenAlreadyOnDoesNotRewrite: saying the same thing twice is not a
// change, and the file's comments are not ours to replace for free.
func TestModeOnWhenAlreadyOnDoesNotRewrite(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: "mode = \"on\"\n",
	})

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeOn}, files, modeCodec{})

	if result.Code != 0 {
		t.Fatalf("code = %d, want 0 (%s)", result.Code, result.Message)
	}
	if len(files.writes) != 0 || len(files.backups) != 0 {
		t.Fatalf("writes = %v, backups = %v; want none (unchanged)", files.written(), files.backups)
	}
	if !strings.Contains(result.Message, "unchanged") {
		t.Errorf("message = %q, want unchanged", result.Message)
	}
}

// TestModeRejectsAMangledSwitch: a typo must block, never quietly mean off.
func TestModeRejectsAMangledSwitch(t *testing.T) {
	files := newFakeFiles(map[string]string{
		domainreview.ModePath: `mode = "maybe"` + "\n",
	})

	result := review.Mode(review.ModeOptions{Dir: dir, Action: review.ModeStatus}, files, modeCodec{})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1 (%s)", result.Code, result.Message)
	}
	for _, want := range []string{domainreview.ModePath, "on or off"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("message = %q, want %q", result.Message, want)
		}
	}
	if len(files.writes) != 0 {
		t.Fatalf("writes = %v, want none: a malformed switch is not overwritten", files.written())
	}
}

// TestModeRejectsAnUnknownAction: the command has three answers, and a fourth
// is a human error worth naming.
func TestModeRejectsAnUnknownAction(t *testing.T) {
	result := review.Mode(review.ModeOptions{Dir: dir, Action: "toggle"}, newFakeFiles(nil), modeCodec{})

	if result.Code != 1 {
		t.Fatalf("code = %d, want 1", result.Code)
	}
	for _, want := range []string{"status", "on", "off"} {
		if !strings.Contains(result.Message, want) {
			t.Errorf("message = %q, want %q", result.Message, want)
		}
	}
}

// TestEnabledReadsTheSwitchAndTreatsMissingAsOff is what `check` uses: the
// absence of the file is the opt-in default, and on is the only true.
func TestEnabledReadsTheSwitchAndTreatsMissingAsOff(t *testing.T) {
	files := newFakeFiles(nil)
	enabled, err := review.Enabled(dir, files, modeCodec{})
	if err != nil {
		t.Fatalf("Enabled(no file) = %v", err)
	}
	if enabled {
		t.Fatal("Enabled(no file) = true, want false")
	}

	files.files[fullPath(domainreview.ModePath)] = []byte("mode = \"on\"\n")
	enabled, err = review.Enabled(dir, files, modeCodec{})
	if err != nil {
		t.Fatalf("Enabled(on) = %v", err)
	}
	if !enabled {
		t.Fatal("Enabled(on) = false, want true")
	}
}
