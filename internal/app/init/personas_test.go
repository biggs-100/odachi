package init_test

import (
	"errors"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/domain/persona"
	"github.com/biggs-100/odachi/internal/domain/rules"
)

// fakeProfiles is the profile port: the profile the human wrote down, or none.
// init must not infer one, so the double has no way to say "inferred".
type fakeProfiles struct {
	name string
	err  error
}

func (p fakeProfiles) Written(string) (string, error) { return p.name, p.err }

// runProfiled performs one `odachi init` with a profile already written down.
func runProfiled(files *fakeFiles, profiles fakeProfiles) initapp.Result {
	return initapp.Run(initapp.Options{Dir: dir, Profiles: profiles}, files, &fakeSkills{})
}

// TestInitWritesNoPersonaFiles: the voice lives in AGENTS.md now, and the client
// directories odachi used to fill are not touched at all. A file nobody's
// selector lists is worse than no file: it looks installed.
func TestInitWritesNoPersonaFiles(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, client := range []string{".opencode", ".claude"} {
		for _, path := range files.events {
			if strings.Contains(path, client) {
				t.Errorf("init tocó %s: %s", client, path)
			}
		}
	}
	for _, file := range persona.Deprecated() {
		if _, ok := files.get(file); ok {
			t.Errorf("%s se escribió otra vez", file)
		}
	}
	if joined := strings.Join(got.Lines, "\n"); strings.Contains(joined, "personas") && !strings.Contains(joined, "v0.10.0") {
		t.Errorf("las líneas todavía hablan de escribir personas:\n%s", joined)
	}
}

// TestInitWritesTheVoiceOfTheProfileAtHand: the block carries the voice of the
// profile the human accepted, and novato when there is no file yet — init runs
// before the interview and cannot know the person.
func TestInitWritesTheVoiceOfTheProfileAtHand(t *testing.T) {
	cases := []struct {
		name     string
		profiles fakeProfiles
		want     string
	}{
		{"sin perfil es novato", fakeProfiles{}, persona.Novato},
		{"novato escrito", fakeProfiles{name: "novato"}, persona.Novato},
		{"senior escrito", fakeProfiles{name: "senior"}, persona.Senior},
		{"ci usa la voz de senior", fakeProfiles{name: "ci"}, persona.Senior},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

			got := runProfiled(files, c.profiles)

			if got.Code != 0 {
				t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
			}
			want, err := rules.Block(c.want)
			if err != nil {
				t.Fatalf("rules.Block(%q): %v", c.want, err)
			}
			content, _ := files.get("AGENTS.md")
			if content != want {
				t.Fatalf("el bloque no lleva la voz %s:\n%q\nvs\n%q", c.want, content, want)
			}
			if !strings.Contains(strings.Join(got.Lines, "\n"), c.want) {
				t.Errorf("las líneas no dicen con qué voz se escribió el bloque:\n%s", strings.Join(got.Lines, "\n"))
			}
		})
	}
}

// A profile that cannot be read is not a failed install: the hook and the skills
// are what the project needs, and the voice falls back to novato with the file
// named and the action that fixes it.
func TestInitNamesABrokenProfileAndKeepsGoing(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runProfiled(files, fakeProfiles{err: errors.New("perfil: TOML inválido")})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	want, err := rules.Block(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Block(novato): %v", err)
	}
	if content, _ := files.get("AGENTS.md"); content != want {
		t.Errorf("el bloque no salió con la voz de novato:\n%q", content)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{".odachi/profile.toml", "odachi init"} {
		if !strings.Contains(joined, want) {
			t.Errorf("las líneas no nombran %q:\n%s", want, joined)
		}
	}
}

// A profile name odachi has no voice for is the same case: fall back, say it, and
// name the three names that have a voice.
func TestInitNamesAProfileWithoutAVoice(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := runProfiled(files, fakeProfiles{name: "mio"})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"mio", "novato", "senior", "ci"} {
		if !strings.Contains(joined, want) {
			t.Errorf("las líneas no nombran %q:\n%s", want, joined)
		}
	}
}

// TestInitReportsTheV0100PersonaFilesAndTouchesNothing: the human's disk is not
// odachi's to clean. One line says what those files are and that they can go.
func TestInitReportsTheV0100PersonaFilesAndTouchesNothing(t *testing.T) {
	leftover := ".opencode/agents/odachi-senior.md"
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		leftover:    "la versión vieja\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"v0.10.0", ".opencode/agents", rules.File, "borralos"} {
		if !strings.Contains(joined, want) {
			t.Errorf("las líneas no dicen %q:\n%s", want, joined)
		}
	}
	if content, _ := files.get(leftover); content != "la versión vieja\n" {
		t.Errorf("init tocó el archivo viejo: %q", content)
	}
	for _, event := range files.events {
		if strings.Contains(event, leftover) {
			t.Errorf("init escribió o respaldó el archivo viejo: %s", event)
		}
	}
}

// TestInitRerunStillWritesNothing: the block is generated, so a second run leaves
// no backup behind and says it changed nothing.
func TestInitRerunStillWritesNothing(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	if code := run(files).Code; code != 0 {
		t.Fatalf("first run code = %d", code)
	}

	files.events = nil
	got := run(files)

	if got.Code != 0 {
		t.Fatalf("second run code = %d (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("second run touched %v, want no writes and no backups", files.events)
	}
	if len(got.Files) != 0 {
		t.Errorf("Files = %v, want none on a rerun that wrote nothing", got.Files)
	}
	if !strings.Contains(got.Message, "sin cambios") {
		t.Errorf("Message = %q, want it to say nothing changed", got.Message)
	}
}
