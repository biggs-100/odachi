package init_test

import (
	"errors"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
	"github.com/biggs-100/odachi/internal/domain/persona"
)

// TestInitNamesTheStoreItCouldNotFind: the client choice lives in the clone's
// store, and a store the adapter cannot resolve is a blocked run that names it.
// Writing with an empty path would put the choice relative to the process
// directory - the one place it must never land - and a silent fallback would
// rewrite the files of a client the person excluded.
func TestInitNamesTheStoreItCouldNotFind(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	selection := &fakeAgentSelection{targetErr: errors.New("fatal: detected dubious ownership in repository")}

	got := runAgents(files, persona.SelectionOpenCode, selection, "")

	if got.Code != initapp.CodeBlocked {
		t.Fatalf("Code = %d, want blocked (message: %s)", got.Code, got.Message)
	}
	if !strings.Contains(got.Message, "clone's store") {
		t.Errorf("Message = %q, want it to name the store it could not find", got.Message)
	}
	for path := range files.files {
		if strings.HasSuffix(path, "agents.toml") {
			t.Errorf("the run wrote %s with no store to write it in", path)
		}
	}
}
