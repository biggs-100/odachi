package doctor

import (
	"strings"
	"testing"
)

// TestEveryCheckHasATitle pins the numbering Contract XV fixes. The number is
// how a report and a test name the same check, so it cannot drift; a new check is
// appended after the ones already numbered, because inserting one in the middle
// would silently rename every check after it.
func TestEveryCheckHasATitle(t *testing.T) {
	for number := CheckVersion; number <= CheckLeftovers; number++ {
		if title(number) == "" {
			t.Errorf("check %d has no title", number)
		}
	}
	if CheckLeftovers != CheckVersion+CheckCount-1 {
		t.Fatalf("the thirteen checks are numbered %d..%d, got %d", CheckVersion, CheckLeftovers, CheckVersion+CheckCount-1)
	}
	if CheckCount != 13 {
		t.Fatalf("ten from Contract XV plus the voice, the task docs and the leftovers one, got %d", CheckCount)
	}
}

// TestTitleOutOfRange: a number no check owns has no name, and inventing one
// would let a report name a check that does not exist.
func TestTitleOutOfRange(t *testing.T) {
	if title(0) != "" || title(99) != "" {
		t.Fatalf("unknown check numbers must have no title")
	}
}

// TestLineIsOneEnglishLine: the whole report is these lines, one per check, so
// the shape is the contract a human reads.
func TestLineIsOneEnglishLine(t *testing.T) {
	cases := []struct {
		name  string
		check Check
		want  string
	}{
		{
			name:  "ok",
			check: OK(CheckVersion, "odachi 0.5.0"),
			want:  "ok 1 · the binary version: odachi 0.5.0",
		},
		{
			name:  "problem names the command that fixes it",
			check: Broken(CheckHook, "there is no pre-commit hook: there is no local gate", "run odachi init"),
			want:  "bad 5 · the pre-commit hook is odachi's: there is no pre-commit hook: there is no local gate; run odachi init",
		},
		{
			name:  "warning is said and never carries a fix",
			check: Warning(CheckIgnored, "proof/ is empty"),
			want:  "warning 10 · the review store and proof/ are ignored: proof/ is empty",
		},
		{
			name:  "ok without detail",
			check: OK(CheckGit, ""),
			want:  "ok 3 · there is .git",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Line(c.check); got != c.want {
				t.Fatalf("Line() = %q, want %q", got, c.want)
			}
			if strings.Count(Line(c.check), "\n") != 0 {
				t.Fatalf("a check is one line: %q", Line(c.check))
			}
		})
	}
}

// TestNotAProjectNamesBothCommandsInOrder pins the one situation a folder that is
// not a project yet gets. It is not one of the checks and has no number:
// numbering it would rename every check after it. The two commands appear in the
// order a person runs them, and each says what it leaves behind.
func TestNotAProjectNamesBothCommandsInOrder(t *testing.T) {
	message, lines := NotAProject()

	if message != NotAProjectMessage {
		t.Fatalf("message = %q, want %q", message, NotAProjectMessage)
	}
	if len(lines) != 2 || lines[0] != message || lines[1] != NotAProjectFix {
		t.Fatalf("lines = %q, want the situation and the one step", lines)
	}
	for _, line := range lines {
		if strings.Count(line, "\n") != 0 {
			t.Fatalf("a line is one line: %q", line)
		}
	}

	init := strings.Index(NotAProjectFix, "odachi init")
	start := strings.Index(NotAProjectFix, "odachi start")
	if init < 0 || start < 0 || init > start {
		t.Fatalf("the fix must name odachi init before odachi start: %q", NotAProjectFix)
	}
	for _, want := range []string{"rules", "interview"} {
		if !strings.Contains(NotAProjectFix, want) {
			t.Errorf("the fix must say what each command leaves behind, missing %q: %q", want, NotAProjectFix)
		}
	}
}

// TestTallyAndCode: only a problem fails the run. A warning is said and does
// not change the exit code.
func TestTallyAndCode(t *testing.T) {
	cases := []struct {
		name     string
		checks   []Check
		problems int
		warnings int
		code     int
	}{
		{"all ok", []Check{OK(CheckVersion, "x"), OK(CheckGit, "yes")}, 0, 0, 0},
		{"warning only", []Check{OK(CheckVersion, "x"), Warning(CheckIgnored, "nothing yet")}, 0, 1, 0},
		{"one problem", []Check{OK(CheckVersion, "x"), Broken(CheckGit, "there is no .git", "run git init")}, 1, 0, 1},
		{"problem and warning", []Check{Broken(CheckGit, "there is no .git", "run git init"), Warning(CheckIgnored, "nothing")}, 1, 1, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			problems, warnings := Tally(c.checks)
			if problems != c.problems || warnings != c.warnings {
				t.Fatalf("Tally() = %d problems, %d warnings; want %d, %d", problems, warnings, c.problems, c.warnings)
			}
			if got := Code(c.checks); got != c.code {
				t.Fatalf("Code() = %d, want %d", got, c.code)
			}
		})
	}
}

// TestSummary: the closing line says how many checks ran and what they found,
// so a human can read the verdict without counting lines.
func TestSummary(t *testing.T) {
	cases := []struct {
		name   string
		checks []Check
		want   string
	}{
		{"clean", []Check{OK(CheckVersion, "x"), OK(CheckGit, "yes")}, "all clear: 2 checks, no problems"},
		{"warning", []Check{OK(CheckVersion, "x"), Warning(CheckGit, "nothing")}, "all clear: 2 checks, 1 warning, no problems"},
		{"problem", []Check{OK(CheckVersion, "x"), Broken(CheckGit, "there is no .git", "run git init")}, "1 problem of 2 checks: run the command each line names"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Summary(c.checks); got != c.want {
				t.Fatalf("Summary() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestIsOurHook: a hook is odachi's when it carries the marker `odachi init`
// writes. Any other hook is somebody else's work, and `init` refuses to
// replace it.
func TestIsOurHook(t *testing.T) {
	ours := "#!/bin/sh\n" + HookMarker + " no se commitea sin prueba de este árbol\n" +
		"ODACHI='odachi'\n[ -x \"$ODACHI\" ] || ODACHI=odachi\n\"$ODACHI\" check --quiet || exit 1\n"
	if !IsOurHook(ours) {
		t.Fatalf("the hook init writes must be recognised: %q", ours)
	}
	if IsOurHook("#!/bin/sh\nnpm test\n") {
		t.Fatalf("a foreign hook is not ours")
	}
	if IsOurHook("") {
		t.Fatalf("an empty hook is not ours")
	}
}

// TestHookExe: the hook reaches the binary through the ODACHI= line, which is
// the only record of what it runs.
func TestHookExe(t *testing.T) {
	cases := []struct {
		name string
		hook string
		want string
	}{
		{"absolute", "ODACHI='/usr/local/bin/odachi'\n", "/usr/local/bin/odachi"},
		{"windows", "ODACHI='C:/Users/ana/bin/odachi.exe'\n", "C:/Users/ana/bin/odachi.exe"},
		{"bare name", "ODACHI='odachi'\n", "odachi"},
		{"not quoted", "ODACHI=/opt/odachi\n", "/opt/odachi"},
		{"missing", "#!/bin/sh\nexit 0\n", ""},
		{"empty value", "ODACHI=''\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := HookExe(c.hook); got != c.want {
				t.Fatalf("HookExe() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestHasBlock: the agent reads the rules only if both markers are there. One
// marker alone is a truncated block, and a truncated block is not a block.
func TestHasBlock(t *testing.T) {
	cases := []struct {
		name   string
		agents string
		want   bool
	}{
		{"both", "# proyecto\n\n" + AgentsBegin + "\nreglas\n" + AgentsEnd + "\n", true},
		{"only begin", AgentsBegin + "\nreglas\n", false},
		{"only end", "reglas\n" + AgentsEnd + "\n", false},
		{"none", "# proyecto\n", false},
		{"empty", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := HasBlock(c.agents); got != c.want {
				t.Fatalf("HasBlock() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestPathFix names the folder of the binary the human already has: the advice
// that leaves a beginner where they were is the advice that costs nothing to
// give, and this tool does not give it.
func TestPathFix(t *testing.T) {
	fix := PathFix("/home/ana/.local/bin/odachi")
	if !strings.Contains(fix, "/home/ana/.local/bin") {
		t.Fatalf("PathFix must name the folder of the running binary, got %q", fix)
	}
	if !strings.Contains(fix, "PATH") {
		t.Fatalf("PathFix must name the PATH, got %q", fix)
	}

	// Without a path there is no folder to name, and inventing one would be
	// worse than asking.
	generic := PathFix("")
	if strings.Contains(generic, "  ") || !strings.Contains(generic, "PATH") {
		t.Fatalf("PathFix(\"\") must stay a sentence about the PATH, got %q", generic)
	}
}
