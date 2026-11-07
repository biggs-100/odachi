# a-non-goal-says-when-it-comes-back

## What must exist
- a non-goal says when it comes back: every entry in a brief's "Not included" names the condition that reopens it -a measurement, a second caller, a version-, and `shape` asks for that condition when it drafts the brief, so the list of what waits does not become a list nobody can date

## Done when
- `shape` says every non-goal names the condition that reopens it, with the three shapes (a measurement, a second caller, a version) and one example
- `shape`'s Output names the non-goals with the path each one protects and the condition that reopens it
- tests in `skills/embed_test.go` pin the text, observed red before it lands
- `odachi skills --check` is green and the body stays under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the convention
- `go test ./...` is green
- `odachi check --run "go test ./..."` is green

## Not included (first version)
- no change to the brief's format: the headers, the parse and the render stay as they are
- no change to the interview or to `start`: the convention is asked where an agent drafts a brief, and the flags carry one answer per non-goal
- no new verb and no new flag
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the skills other than `shape`, and none to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- the convention lives in `shape`, not in the brief's renderer: a brief is free text, and a line the tool injects is a reminder; a skill asks the question at the moment somebody can answer it
- the interview keeps one answer per non-goal: a second question per entry needs a new flag, and the fence presets already need the first word
- the condition is a sentence, not a date: a measurement or a second caller can be checked; a date passes without the reason
