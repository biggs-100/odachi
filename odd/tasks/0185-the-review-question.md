# the-review-question
Base: 87a49b6362d68e9cb2a4df24df3b445e3ef2982a

## Requirements (verbatim, from the user)
R1 · "the review question anchors on the feature's diff: a close committed before its verdict is named instead of read as answered"

## Done when
- a closed feature whose range carries code and has no verdict is named by the mode, whether the closing document is in the worktree or already committed
- a feature whose range is documents only owes nothing
- a tree after a verdict where only documents moved stays paperwork
- an unreadable feature range blocks with the reason, and go test ./... is green

## Steps
- [x] T1 · the close owes its verdict from the diff, not from when the document was committed   R1 · route:inline · tree:2098fabc70b0d8162a3aa29663848be3c82b6c57

## History
L1 · "the review question anchors on the feature's diff: a close committed before its verdict is named instead of read as answered"
L2 · OFICIO · "the review question anchors on the feature's diff, not on when the document was committed; closedBefore keeps its meaning in the proof lane only" (porque the order of commits is not something the gate can demand: a close committed before its verdict was silently read as answered, and --require-review was the only thing that caught it)
