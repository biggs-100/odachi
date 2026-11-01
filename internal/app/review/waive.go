package review

import (
	"fmt"
	"strings"

	"github.com/biggs-100/odachi/internal/app/tasks"
)

// WaiveOptions is one `odachi review waive` invocation.
type WaiveOptions struct {
	Dir string
	// Because is the human's reason. It is required: a waiver without one is
	// silence wearing a decision's clothes.
	Because string
}

// Waive records the human's decision that the current feature lands without a review
// verdict: the preamble line the gate reads at the close, and a history line with
// the reason.
//
// It exists for a project that owns one model. The gate can name a self-review
// instead of refusing it, but a project that asks for the strong guarantee -or a
// human who simply decided- needs the decision on the document rather than on a
// flag that dies with the shell. The waiver removes the verdict and nothing else:
// the proof and the fence still apply, and a finding the review already made still
// stops the ship.
func Waive(opts WaiveOptions, files Files) Result {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	r := &runner{opts: Options{Dir: opts.Dir, Because: opts.Because}, files: files}
	return r.waive()
}

// waive is the work once the ports are wired.
func (r *runner) waive() Result {
	reason := strings.TrimSpace(r.opts.Because)
	if reason == "" {
		return blocked(`a waiver without a reason is silence: run odachi review waive --because "<why this feature lands without a verdict>"`)
	}

	current, err := tasks.Current(r.files, r.opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; a waiver is a decision about one feature, and there is no open one to name")
	}

	doc := current.Doc
	if doc.ReviewWaived {
		return Result{
			Code:    CodeOK,
			Message: "the review of " + current.Rel + " is already waived",
			Lines:   []string{"the review of " + current.Rel + " is already waived: the document carries Review: waived"},
			Files:   []string{current.Rel},
		}
	}

	doc.ReviewWaived = true
	line := fmt.Sprintf("L%d · %q", doc.NextLine(), "waived the review of this feature: "+reason)
	grown := doc.AppendHistory(line)
	if err := tasks.Save(r.files, current, grown); err != nil {
		return blocked("could not write the waiver: " + oneLine(err.Error()))
	}

	message := "the review of " + current.Rel + " is waived: " + reason
	return Result{
		Code:    CodeOK,
		Message: message,
		Lines: []string{
			message,
			"the proof and the fence still apply, and a finding the review already made still stops the ship",
			"next: odachi check",
		},
		Files: []string{current.Rel},
	}
}
