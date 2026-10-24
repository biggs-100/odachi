package doctor

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

// homeAt is the fake's spelling of a path in the human's home.
func homeAt(rel string) string {
	return filepath.Join(homeRoot, filepath.FromSlash(rel))
}

// TestTheDoorThatIsNotTheDefaultIsANote: doctor reads, never writes, and says
// whose default agent is in charge so the human can decide.
func TestTheDoorThatIsNotTheDefaultIsANote(t *testing.T) {
	door, ok := persona.GlobalFilesFor("opencode")
	if !ok {
		t.Fatal("opencode has no global door")
	}
	files := newFakeFiles()
	files.files[homeAt(door.Path)] = "the door"
	files.files[homeAt(".config/opencode/opencode.json")] = `{"default_agent": "build"}`

	r := &runner{opts: Options{Home: homeRoot}, files: files}
	note := r.opencodeDefaultNote([]persona.Client{{Agent: "opencode"}})

	if !strings.Contains(note, "build") || !strings.Contains(note, "opencode.json") {
		t.Errorf("note = %q, want the agent and the file named", note)
	}
}

// TestTheDoorThatIsTheDefaultIsSilent: nothing to fix, nothing to say.
func TestTheDoorThatIsTheDefaultIsSilent(t *testing.T) {
	door, ok := persona.GlobalFilesFor("opencode")
	if !ok {
		t.Fatal("opencode has no global door")
	}
	files := newFakeFiles()
	files.files[homeAt(door.Path)] = "the door"
	files.files[homeAt(".config/opencode/opencode.json")] = `{"default_agent": "odachi"}`

	r := &runner{opts: Options{Home: homeRoot}, files: files}
	if note := r.opencodeDefaultNote([]persona.Client{{Agent: "opencode"}}); note != "" {
		t.Errorf("note = %q, want silence", note)
	}
}
