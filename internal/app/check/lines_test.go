package check

import "testing"

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
