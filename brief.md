# the-newest-verdict-is-the-store-s-newest

## What must exist
- the gate reads the newest verdict by when the store wrote it, not by the clock the reviewer declared: a correction written after the verdict it corrects is the one that judged the furthest bytes, even when its `at` says an earlier time, so `check` never names an existing verdict as missing

## Done when
- `lastVerdictTree` compares the verdict file's modification time, the fact `clean` and `review.NewestTrees` already order by, and the reviewer's `at` no longer decides
- a test seeds two verdicts for one base whose `at` goes backwards and whose file times go forward, and asserts the newer file's tree is the one the gate reads; observed red before the change
- a verdict the port cannot stat is skipped, as an unreadable verdict already is, and the answer with no verdict stays the same
- the exit code and the verdict reasons do not change: the defect was a false note, not a block
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the order

## Not included (first version)
- no change to the verdict's shape, its `at` or the store's files: the file time is already there
- no change to `clean` or `review.NewestTrees`: they already ask the same fact
- no change to the review, the doctor or the classifier
- no new verb and no new flag
- no change to the exit code or the verdict reasons
- no change to the brief's format or to `start`
- no change to the stored proofs: never touched
- no change to the skills or to the voice
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow

## Decisions taken
- the store's write time is the order, not the reviewer's clock: a reviewer writes a date where a time was expected, and a correction always lands after the verdict it corrects
- the `at` stays in the file and keeps its meaning for a human reader; it stops deciding which verdict the gate reads
- a verdict that cannot be stat'ed is skipped, the same way one that cannot be parsed is: an unreadable file is not evidence
- the tie breaks on the tree name, the way `NewestTrees` breaks it, so two files written in the same instant do not depend on the order the directory was listed in
