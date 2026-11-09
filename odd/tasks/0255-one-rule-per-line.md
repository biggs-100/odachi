# one-rule-per-line
Base: 7de10ee899341ce2e1f5f3fb6d87914981699ff6

## Requirements (verbatim, from the user)
R1 · "a rule is one bullet on one line: `skills --check` refuses prose or a wrapped rule in `## Rules`"

## Done when
- `internal/domain/skill` refuses a `## Rules` line that is not a bullet, and a rule over 140 characters, each with its test observed red first
- every bundled skill carries `## Rules` as bullets, none over 140 characters, and none loses a rule the old text had
- `odachi skills --check` is green and `odachi skills --write` leaves SKILLS.md unchanged
- the tests that pin a skill's words still pass, or their expectation moves with the rule it pins
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change for skills, minor (0255)

## Steps
- [x] T1 · the check refuses prose and a rule over 140 characters, tests red first   R1 · route:inline · tree:272d32a5c277e13135e7948f4210c332fb4799f3
- [x] T2 · the rewrite, first batch: the prose-heavy skills   R1 · route:inline · tree:825ff08fac2d65e1cd456da9db2d9e0ba827603d
- [ ] T3 · the rewrite, the rest, the registry and Unreleased   R1 · route:inline

## History
L1 · "a rule is one bullet on one line: `skills --check` refuses prose or a wrapped rule in `## Rules`"
