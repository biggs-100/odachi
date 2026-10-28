package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// withOdachiOnPath puts a fake `odachi` first on the PATH, so the check that
// matters most ("can a shell run it?") is answerable without installing
// anything. Both spellings are created because Windows resolves `odachi`
// through PATHEXT and POSIX through the exact name.
func withOdachiOnPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, name := range []string{"odachi", "odachi.exe"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

// configuredProject is the state `odachi doctor` must call clean: a repository,
// the rules block with the voice matching the profile, the hook, and the four
// artifacts `start` writes.
//
// The last step is `level --to <profile>` and not `start`. `start` writes the
// profile the binary inferred from what it could see — with --json that is `ci` —
// and the block still carries the voice `init` wrote before that profile existed.
// The two have to agree for doctor to call the project clean, and `level --to` is
// the command that makes them agree: it writes the profile and the voice.
func configuredProject(t *testing.T) string {
	t.Helper()
	withOdachiOnPath(t)

	dir := t.TempDir()
	initRepo(t, dir)

	if code, _, stderr := runDir(t, dir, []string{"init"}, ""); code != 0 {
		t.Fatalf("odachi init: exit %d: %s", code, stderr)
	}
	if code, _, stderr := runDir(t, dir, append([]string{"start"}, allFlags()...), ""); code != 0 {
		t.Fatalf("odachi start: exit %d: %s", code, stderr)
	}
	if code, _, stderr := runDir(t, dir, []string{"level", "--to", readProfileName(t, dir)}, ""); code != 0 {
		t.Fatalf("odachi level --to: exit %d: %s", code, stderr)
	}
	return dir
}

// TestDoctorSaysAConfiguredProjectIsFine is the promise of the verb: a project
// that went through init and start reports eleven lines, no problem, exit 0.
func TestDoctorSaysAConfiguredProjectIsFine(t *testing.T) {
	dir := configuredProject(t)

	out, errOut := output(t, dir, []string{"doctor"})
	if errOut != "" {
		t.Fatalf("a clean run writes nothing to stderr: %q", errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 13 {
		t.Fatalf("twelve checks and a summary, got %d lines:\n%s", len(lines), out)
	}
	for i, line := range lines[:11] {
		if !strings.HasPrefix(line, "ok ") {
			t.Fatalf("line %d is not ok in a configured project: %s", i+1, line)
		}
	}
	if !strings.HasPrefix(lines[12], "all clear") {
		t.Fatalf("the last line is the summary, got %q", lines[12])
	}
	if !strings.Contains(lines[12], "12 checks") {
		t.Fatalf("the summary counts the twelve checks, got %q", lines[12])
	}
	// This is also the guard that init and doctor agree on the hook marker: if
	// init renamed it, check 5 would stop recognising its own hook.
	if !strings.Contains(out, "the pre-commit hook is odachi's: yes") {
		t.Fatalf("doctor must recognise the hook init installs:\n%s", out)
	}
}

// TestDoctorWritesNothing: the verb is read-only. The whole project is compared
// byte for byte, because "writes nothing" is a claim about every file and not about
// the ones a test remembered to look at.
func TestDoctorWritesNothing(t *testing.T) {
	dir := configuredProject(t)
	before := tree(t, dir)

	if code, _, stderr := runDir(t, dir, []string{"doctor"}, ""); code != 0 {
		t.Fatalf("doctor: exit %d: %s", code, stderr)
	}

	after := tree(t, dir)
	if len(after) != len(before) {
		t.Fatalf("doctor changed the number of files: %d -> %d", len(before), len(after))
	}
	for path, content := range before {
		if after[path] != content {
			t.Fatalf("doctor wrote %s", path)
		}
	}
}

// TestDoctorReportsEachBrokenInstall: one break, one problem, one command. The
// message has to say what to run, because that is the whole point of the verb.
func TestDoctorReportsEachBrokenInstall(t *testing.T) {
	cases := []struct {
		name   string
		break_ func(t *testing.T, dir string)
		want   []string
	}{
		{
			name: "AGENTS.md without the block",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "AGENTS.md", "# mi proyecto\n")
			},
			want: []string{"the odachi block is in AGENTS.md", "does not have the odachi block", "odachi init"},
		},
		{
			name: "no AGENTS.md",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, "AGENTS.md")); err != nil {
					t.Fatalf("remove AGENTS.md: %v", err)
				}
			},
			want: []string{"there is no AGENTS.md", "odachi init"},
		},
		{
			name: "no hook",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, ".git", "hooks", "pre-commit")); err != nil {
					t.Fatalf("remove the hook: %v", err)
				}
			},
			want: []string{"there is no .git/hooks/pre-commit", "there is no local gate", "odachi init"},
		},
		{
			name: "the hook belongs to somebody else",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, ".git/hooks/pre-commit", "#!/bin/sh\nnpm test\n")
			},
			want: []string{"is not odachi's", "odachi init"},
		},
		{
			name: "the profile does not resolve",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "odd/profile.toml", "name = \"\"\n")
			},
			want: []string{"does not resolve", "does not say name", "odachi doctor"},
		},
		{
			name: "the profile has a value the tool would never obey",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "odd/profile.toml", "name = \"novato\"\nquestions = \"cuando-pueda\"\n")
			},
			want: []string{"does not resolve", "questions", "odachi doctor"},
		},
		{
			name: "the profile is not TOML",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "odd/profile.toml", "this is not toml\n")
			},
			want: []string{"does not resolve", "odachi doctor"},
		},
		{
			name: "there is no brief",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, "brief.md")); err != nil {
					t.Fatalf("remove brief.md: %v", err)
				}
			},
			want: []string{"there is no brief.md", "without a vision there is nothing to build", "odachi start"},
		},
		{
			name: "the brief is incomplete",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "brief.md", "## Qué tiene que existir\n- algo\n")
			},
			want: []string{"missing from brief.md", "odachi start --force"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := configuredProject(t)
			c.break_(t, dir)

			code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
			if code != 1 {
				t.Fatalf("a broken install must exit 1, got %d:\n%s%s", code, stdout, stderr)
			}
			report := stdout + stderr
			for _, want := range c.want {
				if !strings.Contains(report, want) {
					t.Fatalf("the report must name %q:\n%s", want, report)
				}
			}
			if !strings.Contains(report, "bad ") {
				t.Fatalf("a problem line starts with \"bad\":\n%s", report)
			}
		})
	}
}

// TestDoctorJSONSaysTheSameThing: --json carries the same content, so a machine
// never learns less than a human reading the block.
func TestDoctorJSONSaysTheSameThing(t *testing.T) {
	dir := configuredProject(t)

	_, human, _ := runDir(t, dir, []string{"doctor"}, "")
	code, machine, stderr := runDir(t, dir, []string{"doctor", "--json"}, "")
	if stderr != "" {
		t.Fatalf("--json writes nothing to stderr: %q", stderr)
	}
	if code != 0 {
		t.Fatalf("--json must exit 0 for a clean project, got %d", code)
	}

	var payload struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Problems int    `json:"problems"`
		Warnings int    `json:"warnings"`
		Checks   []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Status string `json:"status"`
			Detail string `json:"detail"`
			Fix    string `json:"fix"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(machine), &payload); err != nil {
		t.Fatalf("--json must be valid JSON: %v\n%s", err, machine)
	}

	if len(payload.Checks) != 12 {
		t.Fatalf("--json carries twelve checks, got %d", len(payload.Checks))
	}
	lines := strings.Split(strings.TrimRight(human, "\n"), "\n")
	if payload.Message != lines[len(lines)-1] {
		t.Fatalf("the JSON message must be the summary line: %q vs %q", payload.Message, lines[len(lines)-1])
	}
	for i, check := range payload.Checks {
		line := lines[i]
		for _, want := range []string{check.Title, check.Status} {
			if !strings.Contains(line, want) {
				t.Fatalf("line %d (%q) must carry %q", i+1, line, want)
			}
		}
		if check.Detail != "" && !strings.Contains(line, check.Detail) {
			t.Fatalf("line %d (%q) must carry the detail %q", i+1, line, check.Detail)
		}
	}
}

// TestDoctorRejectsArguments: the verb takes flags and nothing else, and saying
// so is cheaper than ignoring it.
func TestDoctorRejectsArguments(t *testing.T) {
	dir := t.TempDir()
	code, _, stderr := runDir(t, dir, []string{"doctor", "extra"}, "")
	if code != 1 || !strings.Contains(stderr, "extra argument") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

// TestDoctorOnAFolderThatIsNotAProject is the beginner case, and the reason the
// pre-flight exists: in an empty folder the checks named three different
// commands (git init, odachi init twice, odachi start) and never said the only
// thing that person needs — this folder is not a project yet, start here.
func TestDoctorOnAFolderThatIsNotAProject(t *testing.T) {
	dir := t.TempDir()

	code, stdout, stderr := runDir(t, dir, []string{"doctor"}, "")
	if code != 1 {
		t.Fatalf("a folder without a project is not ready yet: exit %d, want 1", code)
	}
	if stdout != "" {
		t.Fatalf("a blocked report goes to stderr, stdout = %q", stdout)
	}

	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("a beginner reads one situation and one step, got %d lines:\n%s", len(lines), stderr)
	}
	if lines[0] != "this folder is not an odachi project yet" {
		t.Fatalf("first line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "odachi init") || !strings.Contains(lines[1], "odachi start") {
		t.Fatalf("the step must name both commands in order: %q", lines[1])
	}
	if strings.Index(lines[1], "odachi init") > strings.Index(lines[1], "odachi start") {
		t.Fatalf("the rule comes before the interview: %q", lines[1])
	}
	if strings.Contains(stderr, "checks") {
		t.Fatalf("the tally is what a beginner cannot act on:\n%s", stderr)
	}

	// --json stays valid and carries the same lines, so a machine never learns
	// less than the human reading the block.
	jsonCode, machine, jsonErr := runDir(t, dir, []string{"doctor", "--json"}, "")
	if jsonErr != "" {
		t.Fatalf("--json writes nothing to stderr: %q", jsonErr)
	}
	if jsonCode != 1 {
		t.Fatalf("--json exits with the same code as the block, got %d", jsonCode)
	}

	var payload struct {
		Code     int               `json:"code"`
		Message  string            `json:"message"`
		Problems int               `json:"problems"`
		Lines    []string          `json:"lines"`
		Checks   []json.RawMessage `json:"checks"`
	}
	if err := json.Unmarshal([]byte(machine), &payload); err != nil {
		t.Fatalf("--json must be valid JSON: %v\n%s", err, machine)
	}
	if payload.Code != 1 {
		t.Fatalf("payload code = %d, want 1", payload.Code)
	}
	if payload.Message != lines[0] {
		t.Fatalf("payload message = %q, want the situation", payload.Message)
	}
	if !slices.Equal(payload.Lines, lines) {
		t.Fatalf("--json lines = %q, want the block's lines %q", payload.Lines, lines)
	}
	if len(payload.Checks) != 0 {
		t.Fatalf("--json must not invent checks for a folder that is not a project: %s", machine)
	}

	// Read-only: the pre-flight runs before anything exists, and it creates nothing.
	if after := tree(t, dir); len(after) != 0 {
		t.Fatalf("doctor wrote %v", after)
	}
}

// output runs one command and returns stdout and stderr separately.
func output(t *testing.T, dir string, args []string) (stdout, stderr string) {
	t.Helper()
	_, stdout, stderr = runDir(t, dir, args, "")
	return stdout, stderr
}
