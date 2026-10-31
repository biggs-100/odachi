package review_test

import (
	"strings"
	"testing"

	"github.com/biggs-100/odachi/internal/domain/review"
)

// candidate is a prompt input with n files, each carrying a diff of roughly
// sizeKB kilobytes, so a test can drive the split with sizes it chose.
func candidate(n, sizeKB int) review.PromptInput {
	var diff strings.Builder
	paths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		path := "internal/p" + string(rune('a'+i)) + ".go"
		paths = append(paths, path)
		diff.WriteString("diff --git a/" + path + " b/" + path + "\n")
		diff.WriteString("index 1111111..2222222 100644\n")
		diff.WriteString("--- a/" + path + "\n")
		diff.WriteString("+++ b/" + path + "\n")
		diff.WriteString("@@ -1,3 +1,3 @@\n")
		for line := 0; line < sizeKB; line++ {
			diff.WriteString("+var somethingThatTakesSpace = \"" + strings.Repeat("x", 1000) + "\"\n")
		}
	}
	return review.PromptInput{
		Brief: "# brief\n\nThe human asked for this.\n",
		Tree:  strings.Repeat("a", 40),
		Base:  strings.Repeat("b", 40),
		Paths: paths,
		Diff:  diff.String(),
	}
}

// A candidate that fits is one part, and the prompt is the one it was: the split
// changes nothing a project already paid for.
func TestAPromptThatFitsIsOnePart(t *testing.T) {
	in := candidate(2, 10)

	parts, err := review.Split(in, 96)
	if err != nil {
		t.Fatalf("Split = %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("parts = %d, want the whole candidate", len(parts))
	}
	if got, want := review.Prompt(parts[0]), review.Prompt(in); got != want {
		t.Errorf("the single part is not the prompt the candidate was:\n%s", got)
	}
}

// A prompt over the limit becomes parts that fit: every part under the limit, every
// changed path in exactly one part, and each part's diff carrying only its own
// hunks -a reviewer of part 2 must not be handed part 1's bytes-.
func TestAPromptOverTheLimitIsSplitIntoPartsThatFit(t *testing.T) {
	in := candidate(4, 3)
	full := review.PromptSizeKB(review.Prompt(in))
	limit := full / 2

	parts, err := review.Split(in, limit)
	if err != nil {
		t.Fatalf("Split = %v", err)
	}
	if len(parts) < 2 {
		t.Fatalf("parts = %d, want more than one for a prompt of %d KB under a %d KB limit", len(parts), full, limit)
	}

	seen := map[string]int{}
	for i, part := range parts {
		if size := review.PromptSizeKB(review.Prompt(part)); size > limit {
			t.Errorf("part %d is %d KB under a %d KB limit", i+1, size, limit)
		}
		if part.Part != i+1 || part.Of != len(parts) {
			t.Errorf("part %d says %d of %d", i+1, part.Part, part.Of)
		}
		if !strings.Contains(review.Prompt(part), "part "+itoa(part.Part)+" of "+itoa(part.Of)) {
			t.Errorf("the prompt of part %d does not say which part it is", i+1)
		}
		for _, path := range part.Paths {
			seen[path]++
			for _, other := range in.Paths {
				carries := strings.Contains(part.Diff, " b/"+other+"\n")
				isMine := other == path
				if carries != isMine {
					t.Errorf("part %d carries the hunks of %s = %v for path %s", i+1, other, carries, path)
				}
			}
		}
	}
	for _, path := range in.Paths {
		if seen[path] != 1 {
			t.Errorf("%s appears in %d parts, want exactly one", path, seen[path])
		}
	}
}

// A path that cannot fit by itself is refused naming it: a reviewer handed half a
// file is not reviewing it.
func TestAPathThatCannotFitAloneIsRefused(t *testing.T) {
	in := candidate(3, 40)

	_, err := review.Split(in, 20)
	if err == nil {
		t.Fatal("Split accepted a limit no single file fits under")
	}
	if !strings.Contains(err.Error(), "internal/p") {
		t.Errorf("error = %q, want the path named", err)
	}
}

// A project with no limit is never split: zero is the absence of a size the human
// measured, and the parts would be a decision nobody made.
func TestALimitOfZeroNeverSplits(t *testing.T) {
	parts, err := review.Split(candidate(4, 30), 0)
	if err != nil {
		t.Fatalf("Split = %v", err)
	}
	if len(parts) != 1 {
		t.Errorf("parts = %d, want one", len(parts))
	}
}

// A diff odachi cannot attribute to a path is not split: the parts would be slices
// nobody can name.
func TestADiffThatNamesNoFileIsRefused(t *testing.T) {
	in := candidate(2, 3)
	in.Diff = strings.Repeat("+no header here\n", 1000)

	if _, err := review.Split(in, 2); err == nil {
		t.Fatal("Split accepted a diff it cannot attribute")
	}
}

// itoa is strconv.Itoa without the import, for the message a test builds.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
