# la-base-que-cubre

## What must exist
- check no se contradice en el cierre: un veredicto cuya base es anterior a la de la feature la cubre, y un árbol que solo movió documentos después de ese veredicto no vuelve a pedir review
- a verdict whose base is a descendant of the feature's base is still refused, naming the base: the lane must not become a hole
- the two symptoms the human saw are one rule: `ready` next to "the feature is closed and its tree has no verdict", and a `review --base` the store answers with "this tree already has a verdict"

## Done when
- a verdict for this tree whose base is an ancestor of the feature's base asks for nothing, and `check` says which tree and which base it speaks for
- a verdict whose base is a descendant of the feature's base is still refused, naming the base it judged and the base the feature starts from
- after a commit that only moved documents on a closed feature, `check` prints ready with no line about a missing verdict, and says the verdict belongs to the code's tree
- the same tree with no verdict for the code it runs on still names the missing verdict, and the close still names it
- the ancestry question is asked of the repository through one port method, and the fake in the gate's tests answers it
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names the rule, patch (0268)

## Not included (first version)
- one tree, one verdict stays: a second verdict for the same bytes is still refused - back when the store has to hold two judgments of one tree
- what counts as a document is not redefined here: the same classifier the proof and the review already share answers it - back when "documents only" itself is the question
- no verdict is written or repaired: `check` reads the store and reports, `review` stays the only thing that writes - back when a verdict has to be rebuilt from something else

## Decisions taken
- Ancestry, not equality: a base earlier than the feature's judged more than its diff, never less, and treating the two as different erased a review that had happened.
- The port gains one question -is this commit an ancestor of that one- instead of the gate guessing with path lists, because a subset of paths is not a subset of bytes.
