# the-verifier-runs-focused-commands
Base: e28725a045b526a673b97407d39165aa7cf15f6d

## Requirements (verbatim, from the user)
R1 · "the delegation policy names the verifier's commands: the focused package the change touches, never a second full gate, because a full suite beside the gate competes for the machine and makes the measurement wrong"

## Done when
- `slices/references/delegation.md` says the request carries the focused package or packages, never a second `go test ./...`, and says why
- a test in `skills/embed_test.go` pins the rule, observed red before the text lands
- `odachi skills --check` is green and the slices body stays within the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [ ] T1 · delegation names the focused package, never a second full gate   R1 · route:inline
- [ ] T2 · the changelog names the rule   R1 · route:inline

## History
L1 · "the delegation policy names the verifier's commands: the focused package the change touches, never a second full gate, because a full suite beside the gate competes for the machine and makes the measurement wrong"
