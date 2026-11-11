# la-base-que-cubre
Base: 0d9293c44e5b46cbc94737c958342f8e32fda31a

## Requirements (verbatim, from the user)
R1 · "check no se contradice en el cierre: un veredicto cuya base es anterior a la de la feature la cubre, y un árbol que solo movió documentos después de ese veredicto no vuelve a pedir review"

## Done when
- a verdict for this tree whose base is an ancestor of the feature's base asks for nothing, and `check` says which tree and which base it speaks for
- a verdict whose base is a descendant of the feature's base is still refused, naming the base it judged and the base the feature starts from
- after a commit that only moved documents on a closed feature, `check` prints ready with no line about a missing verdict, and says the verdict belongs to the code's tree
- the same tree with no verdict for the code it runs on still names the missing verdict, and the close still names it
- the ancestry question is asked of the repository through one port method, and the fake in the gate's tests answers it
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names the rule, patch (0268)

## Steps
- [x] T1 · a verdict with an earlier base covers the feature: the gate asks git instead of refusing a wider diff   R1 · route:inline · tree:70835b18ed640a353257b54a857f609c2dd265b3
- [ ] T2 · a close that only moved documents names the verdict it has instead of asking for one   R1 · route:inline

## History
L1 · "check no se contradice en el cierre: un veredicto cuya base es anterior a la de la feature la cubre, y un árbol que solo movió documentos después de ese veredicto no vuelve a pedir review"
L2 · OFICIO · "T1's recorded tree (70835b1) predates a gofmt fix and the step's commit is now 765686c" (porque the new adapter test was written without its final newline and gofmt -l flagged it; plan --done does not rewrite a finished step, so the record says which tree the step actually left and what changed after it)
