# the-newest-verdict-is-the-store-s-newest
Base: d2e099c19ee80c5011d766337d57cd164e33482f

## Requirements (verbatim, from the user)
R1 · "the gate reads the newest verdict by when the store wrote it, not by the clock the reviewer declared: a correction written after the verdict it corrects is the one that judged the furthest bytes, even when its `at` says an earlier time, so `check` never names an existing verdict as missing"

## Done when
- `lastVerdictTree` compares the verdict file's modification time, the fact `clean` and `review.NewestTrees` already order by, and the reviewer's `at` no longer decides
- a test seeds two verdicts for one base whose `at` goes backwards and whose file times go forward, and asserts the newer file's tree is the one the gate reads; observed red before the change
- a verdict the port cannot stat is skipped, as an unreadable verdict already is, and the answer with no verdict stays the same
- the exit code and the verdict reasons do not change: the defect was a false note, not a block
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- the changelog's Unreleased section names the order

## Steps
- [x] T1 · the gate reads the newest verdict by the store's write time   R1 · route:inline · tree:12744088211d82ac1505f9d0c0919a71c29e8171
- [x] T2 · the changelog names the order   R1 · route:inline · tree:b078f0fe30998b6ca7210ee06d3a1b301c424533

## History
L1 · "the gate reads the newest verdict by when the store wrote it, not by the clock the reviewer declared: a correction written after the verdict it corrects is the one that judged the furthest bytes, even when its `at` says an earlier time, so `check` never names an existing verdict as missing"
