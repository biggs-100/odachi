# verification-is-proportional
Base: 6827dfd2b8c09792beac5c62c1b6b2ed25494d94

## Requirements (verbatim, from the user)
R1 · "verification is proportional — a passive change is verified by reading it, a medium one keeps the writer's own verification, and only a named risk adds an independent verifier — and a red is compared against the base and the environment before it is blamed on the change"

## Done when
- `tests` names the three tiers (passive, medium, high) and who verifies each one: the reader, the writer, `odachi-verify`
- `tests` says a red is run against the base first, and a failure that was already there is named instead of blamed on the change
- `tests` says a red that comes from the machine is named with its condition, and it is never reported as green and never silently skipped
- `slices/references/delegation.md` keeps "a named risk gets a verifier" and sends the reader to the three tiers, so the delegation policy and the test discipline agree
- a test in `skills/embed_test.go` pins the three tier names and the two comparisons, observed red before the text lands
- the body of `tests` stays within the sixty-line cap and `odachi skills --check` is green
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rule
- `go test ./...` is green

## Steps
- [x] T1 · tests teaches the three tiers and the two comparisons, and slices routes to it   R1 · route:inline · tree:0840904532d339a4bb2a088523332d64ce91ab6f
- [x] T2 · the changelog names the rule   R1 · route:inline · tree:84e195629fca0cc2bf05d1db227abba6f6d47522

## History
L1 · "verification is proportional — a passive change is verified by reading it, a medium one keeps the writer's own verification, and only a named risk adds an independent verifier — and a red is compared against the base and the environment before it is blamed on the change"
