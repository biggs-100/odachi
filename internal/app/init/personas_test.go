package init_test

import (
	"errors"
	"path/filepath"
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

// TestInitWritesTheNativeAgent: one file per client, the same block the agent
// reads in AGENTS.md, and the report says how the human picks it. The old
// v0.10.0 files are not rewritten under their old names.
func TestInitWritesTheNativeAgent(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	block, err := rules.Body(persona.Novato)
	if err != nil {
		t.Fatalf("rules.Body(novato): %v", err)
	}
	for _, client := range persona.Clients() {
		file, ok := persona.FilesFor(client.Agent, block)
		if !ok {
			t.Fatalf("persona.FilesFor(%q) no existe", client.Agent)
		}
		content, ok := files.get(file.Path)
		if !ok {
			t.Errorf("%s no se escribió", file.Path)
			continue
		}
		if content != file.Content {
			t.Errorf("%s no es el espejo del bloque:\n%q\nvs\n%q", file.Path, content, file.Content)
		}
		if !strings.Contains(strings.Join(got.Lines, "\n"), client.Dir) {
			t.Errorf("las líneas no nombran %s:\n%s", client.Dir, strings.Join(got.Lines, "\n"))
		}
	}
	for _, file := range persona.Deprecated() {
		if _, ok := files.get(file); ok {
			t.Errorf("%s se escribió otra vez: el agente de esta versión es %s", file, persona.FileName)
		}
	}
}

// TestInitBacksUpAHandEditedAgent: the file is generated, so a human edit is
// lost work. init keeps it in a .bak before replacing it and says it wrote.
func TestInitBacksUpAHandEditedAgent(t *testing.T) {
	path := ".opencode/agents/odachi.md"
	files := newFakeFiles(map[string]string{
		".git/HEAD": "ref: refs/heads/main\n",
		path:        "editado a mano\n",
	})

	got := run(files)

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	backup, wrote := -1, -1
	for i, event := range files.events {
		switch {
		case event == "backup "+path:
			backup = i
		case event == "write "+path:
			wrote = i
		}
	}
	if backup < 0 || wrote < 0 || backup > wrote {
		t.Fatalf("el respaldo tiene que ir antes de la escritura: %v", files.events)
	}
	if content, _ := files.get(path); content == "editado a mano\n" {
		t.Errorf("el agente quedó con la edición a mano en vez del espejo")
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), persona.FileName) {
		t.Errorf("las líneas no dicen qué agente se escribió:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// fakeHome is the Home port: where the human lives, or the reason there is no
// answer. init must not reach os to find it.
type fakeHome struct {
	dir string
	err error
}

func (h fakeHome) Home() (string, error) { return h.dir, h.err }

// TestInitInstallsTheGlobalDoor: one command leaves odachi available in every
// folder, not only in the project. The door is generated, so a second run keeps
// it and says so.
func TestInitInstallsTheGlobalDoor(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	runHome := func() initapp.Result {
		return initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: home}}, files, &fakeSkills{})
	}

	got := runHome()

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, file := range persona.GlobalFiles() {
		abs := filepath.Join(home, filepath.FromSlash(file.Path))
		content, ok := files.files[abs]
		if !ok {
			t.Errorf("%s no se instaló en el home", file.Path)
			continue
		}
		if string(content) != file.Content {
			t.Errorf("%s no es la puerta:\n%q\nvs\n%q", file.Path, content, file.Content)
		}
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "~/"+persona.Clients()[0].HomeDir) {
		t.Errorf("las líneas no nombran el destino del home:\n%s", strings.Join(got.Lines, "\n"))
	}

	files.events = nil
	again := runHome()
	if len(files.events) != 0 {
		t.Errorf("la segunda corrida tocó el home: %v", files.events)
	}
	if !strings.Contains(strings.Join(again.Lines, "\n"), "sin cambios") {
		t.Errorf("la segunda corrida no dice que la puerta ya estaba:\n%s", strings.Join(again.Lines, "\n"))
	}
}

// TestInitKeepsTheProjectWhenTheHomeIsNotWritable: the project half is done and
// the home half is pending. Exit 0, because the project is ready; the line names
// the command that retries it.
func TestInitKeepsTheProjectWhenTheHomeIsNotWritable(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{err: errors.New("sin home")}}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for _, path := range []string{"AGENTS.md", ".opencode/agents/odachi.md", ".claude/output-styles/odachi.md"} {
		if _, ok := files.get(path); !ok {
			t.Errorf("%s no se escribió: el proyecto quedó a medias", path)
		}
	}
	joined := strings.Join(got.Lines, "\n")
	for _, want := range []string{"pendiente:", "home", "odachi init"} {
		if !strings.Contains(joined, want) {
			t.Errorf("las líneas no dicen %q:\n%s", want, joined)
		}
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
	for _, want := range []string{"v0.10.0", ".opencode/agents", rules.File, "borralas"} {
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
