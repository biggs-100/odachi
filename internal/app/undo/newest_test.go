package undo_test

import (
	"testing"

	"github.com/biggs-100/odachi/internal/app/undo"
)

// TestTheNewestBackupWinsWithinOneArtifact keeps the "newest" rule honest after
// the refusal was added: it still decides, but only among the backups of the
// artifact the human named. The choice must not depend on directory order.
func TestTheNewestBackupWinsWithinOneArtifact(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))
	files.put(briefRel+".bak-"+digest("brief medio"), "brief medio", at(20, 12, 0))

	res := undo.Run(undo.Options{Dir: dir, File: "brief"}, files)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if got := files.content(t, briefRel); got != "brief medio" {
		t.Errorf("%s = %q, want the newest backup of the brief", briefRel, got)
	}
}
