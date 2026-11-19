package check

import (
	"strconv"
	"strings"
	"testing"
)

// The counting lives in this package's own test file because authoredLines is
// internal: the gate is its one caller, and exporting it would invite a second
// definition of what a candidate costs.

// TestAuthoredLinesAddsAndDeletesTogether: a candidate that only adds counts
// what it wrote, and one that only removes counts what it took away; the budget
// is the sum of both because the reviewer reads both.
func TestAuthoredLinesAddsAndDeletesTogether(t *testing.T) {
	diff := `diff --git a/internal/app/check/check.go b/internal/app/check/check.go
index 1111111..2222222 100644
--- a/internal/app/check/check.go
+++ b/internal/app/check/check.go
@@ -1,2 +1,3 @@
 package check
-old line
+new line
+another new line
`
	if got := authoredLines(diff); got != 3 {
		t.Fatalf("authoredLines counted %d lines, want 3: one deletion plus two additions", got)
	}
}

// TestAuthoredLinesCountsNoHeaderLine: the file names, the hunk markers and the
// index are scaffolding of the diff format, not lines a person wrote.
func TestAuthoredLinesCountsNoHeaderLine(t *testing.T) {
	diff := `diff --git a/internal/app/check/check.go b/internal/app/check/check.go
index 1111111..2222222 100644
--- a/internal/app/check/check.go
+++ b/internal/app/check/check.go
@@ -1,2 +1,2 @@
`
	if got := authoredLines(diff); got != 0 {
		t.Fatalf("authoredLines counted %d lines in a headers-only diff, want 0", got)
	}
}

// TestAuthoredLinesCountsAHunkByItsOwnLinesOnly: context lines travel with the
// hunk so the reviewer can read it, but they are bytes both trees already had.
func TestAuthoredLinesCountsAHunkByItsOwnLinesOnly(t *testing.T) {
	diff := `diff --git a/docs/change-checklist.md b/docs/change-checklist.md
index 1111111..2222222 100644
--- a/docs/change-checklist.md
+++ b/docs/change-checklist.md
@@ -10,5 +10,5 @@
 context one
 context two
+an added line
-a deleted line
 context three
 context four
`
	if got := authoredLines(diff); got != 2 {
		t.Fatalf("authoredLines counted %d lines, want 2: context belongs to neither tree's author", got)
	}
}

// TestAuthoredLinesCountsNothingUnderProof: the proofs are the tool's own
// evidence, named by the tree that produced them, and they are not the change
// the reviewer is asked to judge.
func TestAuthoredLinesCountsNothingUnderProof(t *testing.T) {
	diff := `diff --git a/proof/0331-un-cierre-que.txt b/proof/0331-un-cierre-que.txt
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/proof/0331-un-cierre-que.txt
@@ -0,0 +1,3 @@
+ok github.com/biggs-100/odachi/internal/app/check
+ok github.com/biggs-100/odachi/internal/domain/gate
+ok github.com/biggs-100/odachi/internal/domain/taskdoc
diff --git a/internal/app/check/check.go b/internal/app/check/check.go
index 1111111..2222222 100644
--- a/internal/app/check/check.go
+++ b/internal/app/check/check.go
@@ -1,1 +1,2 @@
 package check
+// one authored line
`
	if got := authoredLines(diff); got != 1 {
		t.Fatalf("authoredLines counted %d lines, want 1: the proof is not authored work", got)
	}
}

// TestAuthoredLinesCountsARenameAsZero: git reports a rename as a header pair
// with no content, so the budget must not charge for moving a file.
func TestAuthoredLinesCountsARenameAsZero(t *testing.T) {
	diff := `diff --git a/internal/app/check/a.go b/internal/app/check/b.go
similarity index 100%
rename from internal/app/check/a.go
rename to internal/app/check/b.go
`
	if got := authoredLines(diff); got != 0 {
		t.Fatalf("authoredLines counted %d lines for a rename, want 0", got)
	}
}

// TestAuthoredLinesCountsAModeChangeAsZero: making a file executable is a
// permission, not a line, and it is not the reviewer's reading.
func TestAuthoredLinesCountsAModeChangeAsZero(t *testing.T) {
	diff := `diff --git a/scripts/run.sh b/scripts/run.sh
old mode 100644
new mode 100755
`
	if got := authoredLines(diff); got != 0 {
		t.Fatalf("authoredLines counted %d lines for a mode change, want 0", got)
	}
}

// TestAuthoredLinesCountsAnEmptyDiffAsZero: a candidate with nothing in it
// carries no lines, and a feature that only wrote its plan is that candidate.
func TestAuthoredLinesCountsAnEmptyDiffAsZero(t *testing.T) {
	if got := authoredLines(""); got != 0 {
		t.Fatalf("authoredLines counted %d lines in an empty diff, want 0", got)
	}
}

// TestAuthoredLinesCountsNothingInsideABinaryBody: a binary entry encodes its
// bytes, and the encoding can start a line with + or -, so the body is skipped
// by the entry's shape and not by what its lines look like.
func TestAuthoredLinesCountsNothingInsideABinaryBody(t *testing.T) {
	diff := `diff --git a/docs/logo.png b/docs/logo.png
new file mode 100644
index 0000000..1111111
GIT binary patch
literal 4
+LcmZQzU|;|M00
-abcdef
`
	if got := authoredLines(diff); got != 0 {
		t.Fatalf("authoredLines counted %d lines in a binary entry, want 0", got)
	}
}

// TestAuthoredLinesCountsAHunkLineThatLooksLikeAHeader: inside a hunk the first
// character is the whole grammar. A deleted line whose text begins with `-- `
// renders as `--- ...` and an added one beginning with `++ ` renders as `+++ ...`;
// they are the author's lines, and reading them as file headers would both drop
// them from the count and rename the entry mid-hunk (0331).
func TestAuthoredLinesCountsAHunkLineThatLooksLikeAHeader(t *testing.T) {
	diff := `diff --git a/docs/notes.md b/docs/notes.md
index 1111111..2222222 100644
--- a/docs/notes.md
+++ b/docs/notes.md
@@ -1,3 +1,3 @@
--- a note the author deleted
--another deleted line
+++ an added line that begins with plus signs
+and one more
`
	if got := authoredLines(diff); got != 4 {
		t.Fatalf("authoredLines counted %d lines, want 4: hunk lines that look like headers are still authored work", got)
	}
}

// TestAuthoredLinesCountsNothingWhenADeletedProofOpensWithDashes: a proof is
// excluded by its path, and a deleted proof whose first line begins with `-- `
// used to rewrite the path with its own text, so every line after it looked like
// authored work. The entry keeps its header path for the whole hunk (0331).
func TestAuthoredLinesCountsNothingWhenADeletedProofOpensWithDashes(t *testing.T) {
	diff := `diff --git a/proof/0331-un-cierre-que.txt b/proof/0331-un-cierre-que.txt
deleted file mode 100644
index 1111111..0000000
--- a/proof/0331-un-cierre-que.txt
+++ /dev/null
@@ -1,3 +0,0 @@
--- ok github.com/biggs-100/odachi/internal/app/check
--ok github.com/biggs-100/odachi/internal/domain/gate
--ok github.com/biggs-100/odachi/internal/domain/taskdoc
`
	if got := authoredLines(diff); got != 0 {
		t.Fatalf("authoredLines counted %d lines for a deleted proof, want 0: the path was rewritten mid-hunk", got)
	}
}

// addedFileDiff is a candidate that creates path with n authored lines: the shape
// a generator leaves behind when it rewrites a file nobody typed (0331).
func addedFileDiff(path string, n int) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + path + " b/" + path + "\n")
	b.WriteString("new file mode 100644\n")
	b.WriteString("index 0000000..1111111\n")
	b.WriteString("--- /dev/null\n")
	b.WriteString("+++ b/" + path + "\n")
	b.WriteString("@@ -0,0 +1," + strconv.Itoa(n) + " @@\n")
	for i := 1; i <= n; i++ {
		b.WriteString("+linea " + strconv.Itoa(i) + "\n")
	}
	return b.String()
}

// TestAuthoredLinesCountsNothingForAGeneratedFile: a generated file is written by
// a command, and a feature that only reruns the generator pays for bytes nobody
// wrote. The list is closed, so the test walks it: every entry the count forgives
// is measured here, and an empty list fails instead of passing unmeasured (0331).
func TestAuthoredLinesCountsNothingForAGeneratedFile(t *testing.T) {
	if len(generatedPaths) == 0 {
		t.Fatal("the closed list is empty: this test would pass without measuring anything")
	}
	for _, entry := range generatedPaths {
		t.Run(entry.path, func(t *testing.T) {
			if got := authoredLines(addedFileDiff(entry.path, 3)); got != 0 {
				t.Fatalf("authoredLines counted %d lines in the generated %s, want 0: %s wrote them", got, entry.path, entry.owner)
			}
		})
	}
}

// TestEveryGeneratedPathNamesItsOwner: an entry is a claim about the tool, so it
// says which command writes the file and what a hand edit costs. A path added
// without them is a hole in the count nobody can check (0331).
func TestEveryGeneratedPathNamesItsOwner(t *testing.T) {
	for _, entry := range generatedPaths {
		if strings.TrimSpace(entry.owner) == "" {
			t.Errorf("%s is forgiven and no command owns it: name the writer", entry.path)
		}
		if strings.TrimSpace(entry.edit) == "" {
			t.Errorf("%s is forgiven and says nothing about a hand edit: name what happens to it", entry.path)
		}
	}
}

// TestAuthoredLinesStillChargesWhatAPersonWrote: the exclusion is two exact files
// and not a name that ends in SKILLS.md, nor a folder: a document, a registry
// spelled in another directory and the human's own skill beside the generated one
// are authored work (0331).
func TestAuthoredLinesStillChargesWhatAPersonWrote(t *testing.T) {
	for _, path := range []string{"docs/next-session.md", "docs/SKILLS.md", ".agents/skills/acceptance/SKILL.md"} {
		t.Run(path, func(t *testing.T) {
			if got := authoredLines(addedFileDiff(path, 2)); got != 2 {
				t.Fatalf("authoredLines counted %d lines in %s, want 2: a person wrote it", got, path)
			}
		})
	}
}

// TestAuthoredLinesCountsADeletedFileOnce: a deletion names the file where it
// is gone, `+++ /dev/null`, so the count has to read the path from the old
// side or it would miss every removed file.
func TestAuthoredLinesCountsADeletedFileOnce(t *testing.T) {
	diff := `diff --git a/internal/app/check/old.go b/internal/app/check/old.go
deleted file mode 100644
index 1111111..0000000
--- a/internal/app/check/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package check
-// gone
`
	if got := authoredLines(diff); got != 2 {
		t.Fatalf("authoredLines counted %d lines for a deleted file, want 2", got)
	}
}
