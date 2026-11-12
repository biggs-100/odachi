package review

import (
	"fmt"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/app/tasks"
)

// WaiveOptions is one `odachi review waive` invocation.
type WaiveOptions struct {
	Dir string
	// Because is the human's reason. It is required: a waiver without one is
	// silence wearing a decision's clothes.
	Because string
	// Row is the fingerprint of one blocking row the human answers. Empty means
	// the whole feature lands without a verdict, which is the older decision.
	Row string
	// By is who decided, when the caller knows it. A decision nobody signed is a
	// sentence, and this file exists to hold the person's word.
	By string
	// Now fixes the instant the record carries; zero means the wall clock.
	Now time.Time
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
	r := &runner{opts: Options{Dir: opts.Dir, Because: opts.Because, By: opts.By, Now: opts.Now}, files: files}
	if row := strings.TrimSpace(opts.Row); row != "" {
		return r.waiveRow(row)
	}
	return r.waive()
}

// waiveRow records the human's decision about one blocking row of this feature's
// verdict. The row is named by its fingerprint -the same identity the gate printed
// and the verdict files it under- so the decision answers exactly what the human
// read: another row is another question, and a tree that changes reopens it.
func (r *runner) waiveRow(fingerprint string) Result {
	reason := strings.TrimSpace(r.opts.Because)
	if reason == "" {
		return blocked(`a waiver without a reason is silence: run odachi review waive --row <fingerprint> --because "<why the tree ships with this row open>"`)
	}
	if !rowLike(fingerprint) {
		return blocked(fmt.Sprintf("--row says %q, which is not a row: use the fingerprint the gate prints, sixteen hex characters", fingerprint))
	}

	current, err := tasks.Current(r.files, r.opts.Dir)
	if err != nil {
		return blocked(oneLine(err.Error()) + "; a waiver is a decision about one feature, and there is no open one to name")
	}

	when := r.opts.Now
	if when.IsZero() {
		when = time.Now()
	}
	signed := "by " + oneLine(r.opts.By) + " at " + when.UTC().Format(time.RFC3339)
	if strings.TrimSpace(r.opts.By) == "" {
		signed = "at " + when.UTC().Format(time.RFC3339)
	}
	text := "waived the row " + strings.ToLower(fingerprint) + ": " + reason + ", " + signed

	doc := current.Doc
	if doc.WaivedRow(fingerprint) {
		return Result{
			Code:    CodeOK,
			Message: "the row " + fingerprint + " is already waived in " + current.Rel,
			Files:   []string{current.Rel},
		}
	}
	grown := doc.AppendHistory(fmt.Sprintf("L%d · %q", doc.NextLine(), text))
	if err := tasks.Save(r.files, current, grown); err != nil {
		return blocked("could not write the waiver: " + oneLine(err.Error()))
	}

	message := "the row " + strings.ToLower(fingerprint) + " is waived: " + reason
	return Result{
		Code:    CodeOK,
		Message: message,
		Lines: []string{
			message,
			"kept in " + current.Rel + ": the decision travels with the repository, the verdict stays in the clone",
			"the gate keeps saying what the review asked for, and that the human answered it",
			"next: odachi check",
		},
		Files: []string{current.Rel},
	}
}

// rowLike reports whether the text can be a fingerprint: the sixteen hex characters
// the verdict prints. A word that is not one names no row, and a waiver about
// nothing is the silent gate this feature exists to close.
func rowLike(fingerprint string) bool {
	if len(fingerprint) != 16 {
		return false
	}
	for _, r := range fingerprint {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
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
