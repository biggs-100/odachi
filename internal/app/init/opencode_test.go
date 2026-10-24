package init_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	domainopencode "github.com/biggs-100/odachi/internal/domain/opencode"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// openCodeConfig is the file the global half owns. Every test names it through
// this helper so Windows and Unix agree on the path the app writes.
func openCodeConfig(home, name string) string {
	return filepath.Join(home, filepath.FromSlash(".config/opencode/"+name))
}

// defaultAgent reads the agent the config names, stripping comments the way
// OpenCode does.
func defaultAgent(t *testing.T, data []byte) any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(domainopencode.StripComments(data), &doc); err != nil {
		t.Fatalf("the config is not valid: %v\n%s", err, data)
	}
	return doc["default_agent"]
}

// TestInitGlobalMakesTheDoorTheDefault: installing the door with --global also
// answers the question the door asks: who starts the session.
func TestInitGlobalMakesTheDoorTheDefault(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})
	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	data, ok := files.files[openCodeConfig("hogar", "opencode.json")]
	if !ok {
		t.Fatal("the global config was not written")
	}
	if agent := defaultAgent(t, data); agent != persona.AgentName {
		t.Errorf("default_agent = %v, want %q", agent, persona.AgentName)
	}
	if !strings.Contains(strings.Join(got.Lines, "\n"), "default agent") {
		t.Errorf("the report does not say what it did:\n%s", strings.Join(got.Lines, "\n"))
	}
}

// TestInitGlobalKeepsAnotherDefault: the human's choice wins. odachi installs
// its door and says which agent is in charge, in the file they can edit.
func TestInitGlobalKeepsAnotherDefault(t *testing.T) {
	config := openCodeConfig("hogar", "opencode.json")
	original := `{"default_agent": "build"}`
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[config] = []byte(original)

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})

	if string(files.files[config]) != original {
		t.Errorf("the config changed:\n%s", files.files[config])
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "build") {
		t.Errorf("the report does not name the default it kept:\n%s", joined)
	}
	for _, event := range files.events {
		if strings.Contains(event, "backup") && strings.Contains(event, "opencode.json") {
			t.Errorf("a file that was not modified was backed up: %s", event)
		}
	}
}

// TestInitWithoutGlobalLeavesTheConfigAlone: the global config is written by
// --global and by nothing else.
func TestInitWithoutGlobalLeavesTheConfigAlone(t *testing.T) {
	config := openCodeConfig("hogar", "opencode.json")
	original := `{}`
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[config] = []byte(original)

	initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}}, files, &fakeSkills{})

	if string(files.files[config]) != original {
		t.Errorf("init without --global wrote the global config:\n%s", files.files[config])
	}
}

// TestInitGlobalRefusesAMalformedConfig: a config the tool cannot read is not
// rewritten. The line names the file and the exact member to add.
func TestInitGlobalRefusesAMalformedConfig(t *testing.T) {
	config := openCodeConfig("hogar", "opencode.json")
	original := `{"mcp": {`
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[config] = []byte(original)

	got := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})

	if string(files.files[config]) != original {
		t.Errorf("a malformed config was rewritten:\n%s", files.files[config])
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "by hand") {
		t.Errorf("the report does not tell the human what to add:\n%s", joined)
	}
}

// TestInitGlobalEditsTheJSONCWhenThatIsTheFile: OpenCode accepts .jsonc. The
// edit preserves its comments and does not create a second config beside it.
func TestInitGlobalEditsTheJSONCWhenThatIsTheFile(t *testing.T) {
	config := openCodeConfig("hogar", "opencode.jsonc")
	original := "{\n  // keep me\n  \"mcp\": {}\n}\n"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[config] = []byte(original)

	initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})

	updated := string(files.files[config])
	if !strings.Contains(updated, "// keep me") {
		t.Errorf("the comment was lost:\n%s", updated)
	}
	if agent := defaultAgent(t, files.files[config]); agent != persona.AgentName {
		t.Errorf("default_agent = %v, want %q", agent, persona.AgentName)
	}
	if _, exists := files.files[openCodeConfig("hogar", "opencode.json")]; exists {
		t.Error("a second config was created beside the .jsonc")
	}
}

// TestInitGlobalIsIdempotent: a second run says the agent is already the
// default and writes nothing.
func TestInitGlobalIsIdempotent(t *testing.T) {
	config := openCodeConfig("hogar", "opencode.json")
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	first := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})
	if first.Code != 0 {
		t.Fatalf("first init: Code = %d", first.Code)
	}
	written := string(files.files[config])
	files.events = nil

	second := initapp.Run(initapp.Options{Dir: dir, Home: fakeHome{dir: "hogar"}, Global: true}, files, &fakeSkills{})

	if string(files.files[config]) != written {
		t.Errorf("the second run rewrote the config:\n%s", files.files[config])
	}
	if !strings.Contains(strings.Join(second.Lines, "\n"), "already the default") {
		t.Errorf("the second run does not say the work is done:\n%s", strings.Join(second.Lines, "\n"))
	}
}
