package skills_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	skillsapp "github.com/biggs-100/odachi/internal/app/skills"
)

const dir = "/proyecto"

// fakeFiles is the Files port in memory. Every write is recorded so a test can
// prove a refused command touched nothing.
type fakeFiles struct {
	files    map[string][]byte
	events   []string
	writeErr error
}

func newFakeFiles(seed map[string]string) *fakeFiles {
	f := &fakeFiles{files: map[string][]byte{}}
	for path, content := range seed {
		f.put(path, content)
	}
	return f
}

func (f *fakeFiles) put(path, content string) {
	f.files[filepath.Join(dir, filepath.FromSlash(path))] = []byte(content)
}

func (f *fakeFiles) abs(path string) string {
	return filepath.Join(dir, filepath.FromSlash(path))
}

func (f *fakeFiles) get(path string) (string, bool) {
	data, ok := f.files[f.abs(path)]
	return string(data), ok
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	// A path whose ancestor is a file is not a missing file: on POSIX it is
	// ENOTDIR. The fake models that on purpose, because treating the two as the
	// same is exactly the cross-platform bug that let a green Windows run hide a
	// red Linux one.
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		if _, ok := f.files[parent]; ok {
			return false, fmt.Errorf("stat %s: not a directory", path)
		}
		if next := filepath.Dir(parent); next == parent {
			break
		}
	}

	if _, ok := f.files[path]; ok {
		return true, nil
	}
	prefix := path + string(filepath.Separator)
	for file := range f.files {
		if strings.HasPrefix(file, prefix) {
			return true, nil
		}
	}
	return false, nil
}

// IsDir reports whether path is a directory in the fake: it is one when a file
// lives under it.
func (f *fakeFiles) IsDir(path string) (bool, error) {
	prefix := path + string(filepath.Separator)
	for file := range f.files {
		if strings.HasPrefix(file, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeFiles) List(listDir string) ([]string, error) {
	seen := map[string]bool{}
	var names []string
	prefix := listDir + string(filepath.Separator)
	for file := range f.files {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		rest := strings.TrimPrefix(file, prefix)
		name, _, _ := strings.Cut(rest, string(filepath.Separator))
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.events = append(f.events, "write "+filepath.ToSlash(path))
	f.files[path] = data
	return nil
}

// skillText is a complete skill in the canonical form.
func skillText(name, trigger string) string {
	return "---\nname: " + name + "\ndescription: Trigger: " + trigger +
		"\n---\n\n## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n"
}

// registryOf renders what --write must produce for the given skills.
func registryOf(names ...string) string {
	var out strings.Builder
	out.WriteString("<!-- generado por odachi skills --write; no lo edites a mano -->\n")
	out.WriteString("| Skill | Trigger | Path |\n")
	out.WriteString("| --- | --- | --- |\n")
	slices.Sort(names)
	for _, name := range names {
		out.WriteString("| " + name + " | trigger de " + name + " | skills/" + name + "/SKILL.md |\n")
	}
	return out.String()
}

// seededProject is a project whose skills and registry agree.
func seededProject() *fakeFiles {
	return newFakeFiles(map[string]string{
		"skills/acceptance/SKILL.md": skillText("acceptance", "trigger de acceptance"),
		"skills/ship/SKILL.md":       skillText("ship", "trigger de ship"),
		"SKILLS.md":                  registryOf("acceptance", "ship"),
	})
}

// A repository without skills/ is not a failure: it is a starting point, and the
// line says how to start.
func TestWithoutSkillsDirectorySaysHowToStart(t *testing.T) {
	for _, opts := range []skillsapp.Options{
		{Dir: dir},
		{Dir: dir, Check: true},
		{Dir: dir, Write: true},
	} {
		got := skillsapp.Run(opts, newFakeFiles(nil))

		if got.Code != 0 {
			t.Errorf("%+v: Code = %d, want 0", opts, got.Code)
		}
		if !strings.Contains(got.Message, "odachi skill new") {
			t.Errorf("%+v: Message = %q, want it to name odachi skill new", opts, got.Message)
		}
	}
}

// The listing is one line per skill, in name order, with its trigger and path.
func TestListShowsTheTriggerAndThePath(t *testing.T) {
	got := skillsapp.Run(skillsapp.Options{Dir: dir}, seededProject())

	if got.Code != 0 {
		t.Fatalf("Code = %d (message %q)", got.Code, got.Message)
	}
	want := []string{
		"acceptance · trigger de acceptance · skills/acceptance/SKILL.md",
		"ship · trigger de ship · skills/ship/SKILL.md",
	}
	if !slices.Equal(got.Lines, want) {
		t.Errorf("Lines = %q, want %q", got.Lines, want)
	}
	if len(got.Skills) != 2 {
		t.Errorf("Skills = %d, want 2", len(got.Skills))
	}
}

// A valid project has nothing to report.
func TestCheckIsGreen(t *testing.T) {
	got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, seededProject())

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
	if len(got.Lines) != 0 {
		t.Errorf("Lines = %q, want nothing on a green check", got.Lines)
	}
	if !strings.Contains(got.Message, "SKILLS.md") {
		t.Errorf("Message = %q, want it to mention the registry", got.Message)
	}
}

// Every violation is one line that names the file and what is wrong.
func TestCheckReportsEveryViolation(t *testing.T) {
	cases := []struct {
		name   string
		seed   map[string]string
		file   string
		reason string
	}{
		{
			name:   "sin frontmatter",
			seed:   map[string]string{"skills/suelta/SKILL.md": "## When\n\ncuando\n\n## Rules\n\nregla\n\n## Never\n\nnunca\n"},
			file:   "skills/suelta/SKILL.md",
			reason: "frontmatter",
		},
		{
			name:   "sin Trigger",
			seed:   map[string]string{"skills/aceptar/SKILL.md": strings.Replace(skillText("aceptar", "x"), "description: Trigger: x", "description: aceptar cosas", 1)},
			file:   "skills/aceptar/SKILL.md",
			reason: `no empieza con "Trigger:"`,
		},
		{
			name:   "nombre que no coincide",
			seed:   map[string]string{"skills/aceptar/SKILL.md": skillText("otro", "x")},
			file:   "skills/aceptar/SKILL.md",
			reason: "no coincide con el directorio",
		},
		{
			name:   "TODO",
			seed:   map[string]string{"skills/aceptar/SKILL.md": strings.Replace(skillText("aceptar", "x"), "regla", "TODO regla", 1)},
			file:   "skills/aceptar/SKILL.md",
			reason: "TODO",
		},
		{
			name:   "registro desactualizado",
			seed:   map[string]string{"skills/aceptar/SKILL.md": skillText("aceptar", "x"), "SKILLS.md": "| a mano |\n"},
			file:   "SKILLS.md",
			reason: "--write",
		},
		{
			name:   "registro ausente",
			seed:   map[string]string{"skills/aceptar/SKILL.md": skillText("aceptar", "x")},
			file:   "SKILLS.md",
			reason: "falta SKILLS.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, newFakeFiles(tc.seed))

			if got.Code != 1 {
				t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
			}
			joined := strings.Join(got.Lines, "\n")
			if !strings.Contains(joined, tc.file) {
				t.Errorf("lines %q do not name %s", joined, tc.file)
			}
			if !strings.Contains(joined, tc.reason) {
				t.Errorf("lines %q do not contain %q", joined, tc.reason)
			}
		})
	}
}

// Writing twice produces the same bytes, and the second run says so instead of
// claiming work it did not do.
func TestWriteIsIdempotent(t *testing.T) {
	files := newFakeFiles(map[string]string{
		"skills/ship/SKILL.md":       skillText("ship", "trigger de ship"),
		"skills/acceptance/SKILL.md": skillText("acceptance", "trigger de acceptance"),
	})

	first := skillsapp.Run(skillsapp.Options{Dir: dir, Write: true}, files)
	if first.Code != 0 {
		t.Fatalf("first Code = %d (message %q)", first.Code, first.Message)
	}
	written, ok := files.get("SKILLS.md")
	if !ok {
		t.Fatalf("SKILLS.md was not written; events %v", files.events)
	}
	if want := registryOf("acceptance", "ship"); written != want {
		t.Errorf("SKILLS.md =\n%s\nwant\n%s", written, want)
	}

	files.events = nil
	second := skillsapp.Run(skillsapp.Options{Dir: dir, Write: true}, files)
	if second.Code != 0 {
		t.Fatalf("second Code = %d (message %q)", second.Code, second.Message)
	}
	again, _ := files.get("SKILLS.md")
	if again != written {
		t.Errorf("the second run changed the bytes:\n%s\nvs\n%s", again, written)
	}
	if len(files.events) != 0 {
		t.Errorf("the second run wrote %v, want nothing", files.events)
	}
	if !strings.Contains(second.Message, "sin cambios") {
		t.Errorf("second Message = %q, want it to say nothing changed", second.Message)
	}
}

// A hand-edited registry is replaced by --write, and the check that failed on it
// passes afterwards.
func TestWriteRestoresAHandEditedRegistry(t *testing.T) {
	files := seededProject()
	files.put("SKILLS.md", "| acceptance | robado a mano | skills/acceptance/SKILL.md |\n")

	if got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files); got.Code != 1 {
		t.Fatalf("check on a hand-edited registry = %d, want 1", got.Code)
	}
	if got := skillsapp.Run(skillsapp.Options{Dir: dir, Write: true}, files); got.Code != 0 {
		t.Fatalf("write Code = %d (message %q)", got.Code, got.Message)
	}
	if got := skillsapp.Run(skillsapp.Options{Dir: dir, Check: true}, files); got.Code != 0 {
		t.Errorf("check after write = %d, want 0 (lines %q)", got.Code, got.Lines)
	}
}

// `skill new` creates the file, brings the registry up to date, and tells the
// human what is missing.
func TestNewCreatesTheSkillAndUpdatesTheRegistry(t *testing.T) {
	files := seededProject()

	got := skillsapp.New(skillsapp.Options{Dir: dir, Slug: "ejemplo"}, files)

	if got.Code != 0 {
		t.Fatalf("Code = %d (message %q)", got.Code, got.Message)
	}
	content, ok := files.get("skills/ejemplo/SKILL.md")
	if !ok {
		t.Fatalf("the skill was not written; events %v", files.events)
	}
	if !strings.Contains(content, "name: ejemplo") {
		t.Errorf("SKILL.md does not declare its name:\n%s", content)
	}
	registry, _ := files.get("SKILLS.md")
	if !strings.Contains(registry, "| ejemplo |") {
		t.Errorf("SKILLS.md does not list the new skill:\n%s", registry)
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "skills/ejemplo/SKILL.md") || !strings.Contains(joined, "SKILLS.md") {
		t.Errorf("Lines = %q, want the skill path and the registry", joined)
	}
	if !strings.Contains(joined, "odachi skills --check") {
		t.Errorf("Lines = %q, want the next command", joined)
	}
	if len(got.Files) != 2 {
		t.Errorf("Files = %v, want the skill and the registry", got.Files)
	}
}

// An existing slug is refused with exit 1 and nothing is touched.
func TestNewRefusesAnExistingSlug(t *testing.T) {
	files := seededProject()
	before, _ := files.get("skills/ship/SKILL.md")
	registryBefore, _ := files.get("SKILLS.md")
	files.events = nil

	got := skillsapp.New(skillsapp.Options{Dir: dir, Slug: "ship"}, files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("a refused slug wrote %v, want nothing", files.events)
	}
	if after, _ := files.get("skills/ship/SKILL.md"); after != before {
		t.Errorf("the existing skill changed:\n%s\nvs\n%s", after, before)
	}
	if after, _ := files.get("SKILLS.md"); after != registryBefore {
		t.Errorf("the registry changed on a refused slug:\n%s\nvs\n%s", after, registryBefore)
	}
}

// A slug that could never be a valid name is rejected before any write.
func TestNewRejectsABadSlugBeforeWriting(t *testing.T) {
	files := seededProject()
	files.events = nil

	got := skillsapp.New(skillsapp.Options{Dir: dir, Slug: "Slug Invalido"}, files)

	if got.Code != 1 {
		t.Fatalf("Code = %d, want 1 (message %q)", got.Code, got.Message)
	}
	if len(files.events) != 0 {
		t.Errorf("an invalid slug wrote %v, want nothing", files.events)
	}
	if !strings.Contains(got.Message, "^[a-z][a-z0-9-]{2,40}$") {
		t.Errorf("Message = %q, want the slug shape", got.Message)
	}
}
