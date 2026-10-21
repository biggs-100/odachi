package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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
// the rules block, the hook, and the four artifacts `start` writes.
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
	return dir
}

// TestDoctorSaysAConfiguredProjectIsFine is the promise of the verb: a project
// that went through init and start reports ten lines, no problem, exit 0.
func TestDoctorSaysAConfiguredProjectIsFine(t *testing.T) {
	dir := configuredProject(t)

	out, errOut := output(t, dir, []string{"doctor"})
	if errOut != "" {
		t.Fatalf("a clean run writes nothing to stderr: %q", errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 11 {
		t.Fatalf("ten checks and a summary, got %d lines:\n%s", len(lines), out)
	}
	for i, line := range lines[:10] {
		if !strings.HasPrefix(line, "ok ") {
			t.Fatalf("line %d is not ok in a configured project: %s", i+1, line)
		}
	}
	if !strings.HasPrefix(lines[10], "todo en orden") {
		t.Fatalf("the last line is the summary, got %q", lines[10])
	}
	// This is also the guard that init and doctor agree on the hook marker: if
	// init renamed it, check 5 would stop recognising its own hook.
	if !strings.Contains(out, "el hook de pre-commit es de odachi: sí") {
		t.Fatalf("doctor must recognise the hook init installs:\n%s", out)
	}
}

// TestDoctorWritesNothing: the verb is read-only. The whole project is compared
// byte for byte, because "no escribe" is a claim about every file and not about
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
			want: []string{"el bloque de odachi está en AGENTS.md", "no tiene el bloque", "odachi init"},
		},
		{
			name: "no AGENTS.md",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, "AGENTS.md")); err != nil {
					t.Fatalf("remove AGENTS.md: %v", err)
				}
			},
			want: []string{"no hay AGENTS.md", "odachi init"},
		},
		{
			name: "no hook",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, ".git", "hooks", "pre-commit")); err != nil {
					t.Fatalf("remove the hook: %v", err)
				}
			},
			want: []string{"no hay .git/hooks/pre-commit", "no hay gate local", "odachi init"},
		},
		{
			name: "the hook belongs to somebody else",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, ".git/hooks/pre-commit", "#!/bin/sh\nnpm test\n")
			},
			want: []string{"no es de odachi", "odachi init"},
		},
		{
			name: "the profile does not resolve",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, ".odachi/profile.toml", "name = \"\"\n")
			},
			want: []string{"no resuelve", "no dice name", "odachi doctor"},
		},
		{
			name: "the profile has a value the tool would never obey",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, ".odachi/profile.toml", "name = \"novato\"\nquestions = \"cuando-pueda\"\n")
			},
			want: []string{"no resuelve", "questions", "odachi doctor"},
		},
		{
			name: "the profile is not TOML",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, ".odachi/profile.toml", "esto no es toml\n")
			},
			want: []string{"no resuelve", "odachi doctor"},
		},
		{
			name: "there is no brief",
			break_: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, "brief.md")); err != nil {
					t.Fatalf("remove brief.md: %v", err)
				}
			},
			want: []string{"no hay brief.md", "sin visión no hay nada que construir", "odachi start"},
		},
		{
			name: "the brief is incomplete",
			break_: func(t *testing.T, dir string) {
				writeFile(t, dir, "brief.md", "## Qué tiene que existir\n- algo\n")
			},
			want: []string{"le falta", "odachi start --force"},
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
			if !strings.Contains(report, "mal ") {
				t.Fatalf("a problem line starts with «mal»:\n%s", report)
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

	if len(payload.Checks) != 10 {
		t.Fatalf("--json carries ten checks, got %d", len(payload.Checks))
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
	if code != 1 || !strings.Contains(stderr, "sobra un argumento") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

// output runs one command and returns stdout and stderr separately.
func output(t *testing.T, dir string, args []string) (stdout, stderr string) {
	t.Helper()
	_, stdout, stderr = runDir(t, dir, args, "")
	return stdout, stderr
}
