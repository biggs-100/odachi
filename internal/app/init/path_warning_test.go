package init_test

import (
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

// The rules block tells the agent to run `odachi ...`. If a shell cannot resolve
// that name, the agent cannot obey a single rule: in a real OpenCode run it spent
// seven shell calls hunting for the binary, including a recursive scan of the
// whole drive. The hook still works, because it uses the absolute path, so the
// install is half broken and init has to say so.
func TestInitWarnsWhenTheBinaryIsNotOnPath(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	// An OS-appropriate path: asserting a POSIX one makes the test fail on
	// Windows for a reason that has nothing to do with the behaviour.
	exe := filepath.Join("opt", "odachi", "bin", "odachi")

	got := initapp.Run(initapp.Options{
		Dir:         dir,
		Exe:         exe,
		PathMissing: true,
	}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0: the artifacts were written (lines %q)", got.Code, got.Lines)
	}

	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "PATH") {
		t.Fatalf("las líneas no mencionan el PATH: %q", got.Lines)
	}
	if !strings.Contains(joined, filepath.Dir(exe)) {
		t.Fatalf("las líneas no dicen qué carpeta agregar: %q", got.Lines)
	}
	if !strings.Contains(joined, "hook") {
		t.Fatalf("las líneas no aclaran que el hook sí funciona: %q", got.Lines)
	}
}

// When the name resolves, saying anything about the PATH is noise.
func TestInitSaysNothingAboutThePathWhenTheBinaryResolves(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Exe: "odachi"}, files, &fakeSkills{})

	for _, line := range got.Lines {
		if strings.Contains(line, "PATH") {
			t.Fatalf("una línea habla del PATH sin necesidad: %q", line)
		}
	}
}
