# the-smallest-change
Base: da3353df799a470576a65b267c7c35871899b432

## Requirements (verbatim, from the user)
R1 · "before the code, the change names what it reuses, what it adds and what it deletes, and the PR carries the same list to the reviewer: an abstraction needs a second caller, a pattern waits for three uses for the same reason, and a wrong abstraction is inlined back before it is extracted again"

## Done when
- `shape` names the smallest change (reuse, add, delete; an abstraction needs a second caller), the rule of three for the same reason, the props test, and the inline-first recovery
- `pull-request` says every PR names its new files, dependencies and abstractions with their reason and the code it reuses, and the reviewer asks what to delete, what was reused and which requirement each line serves
- tests in `skills/embed_test.go` pin both, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rules
- `go test ./...` is green

## Steps
- [ ] T1 · shape asks for the smallest change and the PR carries it   R1 · route:inline
- [ ] T2 · the changelog names the rules   R1 · route:inline

## History
L1 · "before the code, the change names what it reuses, what it adds and what it deletes, and the PR carries the same list to the reviewer: an abstraction needs a second caller, a pattern waits for three uses for the same reason, and a wrong abstraction is inlined back before it is extracted again"
