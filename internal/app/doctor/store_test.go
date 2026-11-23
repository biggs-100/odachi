package doctor

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/review"
	"github.com/biggs-100/odachi/internal/domain/store"
	"github.com/biggs-100/odachi/internal/domain/verify"
)

// evidenceStore is the store the fake git points at: the clone's common dir in
// the fake project, plus odachi's folder inside it.
func evidenceStore() string {
	return store.Path(at(".git"), store.DirName)
}

// TestUnreadableEvidenceStoreIsOneBrokenLineWithAFix: a store whose verdicts
// folder cannot be read is the only broken line, and it names the path and the
// repair. A verdict that cannot be read is not "no evidence": the human has to
// be told where to look.
func TestUnreadableEvidenceStoreIsOneBrokenLineWithAFix(t *testing.T) {
	files := configured()
	reviews := store.Path(evidenceStore(), review.Dir)
	files.errOn = reviews

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	got := checkOf(t, result, domaindoctor.CheckStore)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("an unreadable store is a problem, got %v: %q", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, store.DirName) || !strings.Contains(got.Detail, review.Dir) {
		t.Errorf("the detail must name the store path %q, got %q", reviews, got.Detail)
	}
	if strings.TrimSpace(got.Fix) == "" {
		t.Error("a broken store line must carry a repair")
	}
	broken := 0
	for _, c := range result.Checks {
		if c.Status == domaindoctor.StatusProblem {
			broken++
		}
	}
	if broken != 1 {
		t.Errorf("exactly one check is broken, got %d: %s", broken, strings.Join(result.Lines, " / "))
	}
}

// TestReadableEvidenceStoreIsOK: a store that holds verdicts and verifications
// is readable, so check 18 says where it looked and stays green.
func TestReadableEvidenceStoreIsOK(t *testing.T) {
	files := configured()
	files.put(filepath.Join(".git", store.DirName, review.Dir, "tree.json"), `{}`)
	files.put(filepath.Join(".git", store.DirName, verify.Dir, "tree.json"), `{}`)

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	got := checkOf(t, result, domaindoctor.CheckStore)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("a readable store is ok, got %v: %q", got.Status, got.Detail)
	}
	if !strings.Contains(got.Detail, store.DirName) {
		t.Errorf("an ok line still says where it looked, got %q", got.Detail)
	}
}

// TestAbsentEvidenceStoreIsOK: a clone with no store yet is not broken. Reading
// a store creates nothing, so the check must treat a missing folder as a state.
func TestAbsentEvidenceStoreIsOK(t *testing.T) {
	files := configured()

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", fakeGit{})

	got := checkOf(t, result, domaindoctor.CheckStore)
	if got.Status != domaindoctor.StatusOK {
		t.Fatalf("an absent store is ok, got %v: %q", got.Status, got.Detail)
	}
}

// TestGitCannotLocateTheStoreIsBroken: when git cannot answer where the common
// directory is, the store cannot be read; the line says so and how to repair it.
func TestGitCannotLocateTheStoreIsBroken(t *testing.T) {
	files := configured()
	git := fakeGit{commonDirErr: errors.New("git directory unavailable")}

	result := runInARepository(files, cleanLookup(), cleanProfiles(), "/home/ana/bin/odachi", git)

	got := checkOf(t, result, domaindoctor.CheckStore)
	if got.Status != domaindoctor.StatusProblem {
		t.Fatalf("a store that cannot be located is a problem, got %v: %q", got.Status, got.Detail)
	}
	if strings.TrimSpace(got.Fix) == "" {
		t.Error("a store the tool cannot locate must carry a repair")
	}
}
