# the-chain-says-where-it-stands
Base: 0cddda8d773faf10248340217369abb4fa9b7a3c

## Requirements (verbatim, from the user)
R1 · "a chained PR says where it stands and a commit says what it can be reverted without: the chain's tracker, position, base, dependency and next PR, the `📍` in its diagram, a diff that shows only its unit, a branch named for what it delivers, and a rollback boundary per commit"

## Done when
- `pull-request` names the branch prefixes (`feat/`, `fix/`, `docs/`, `chore/`, `refactor/`), the chain context (tracker, position, base, depends on, next PR, leaves out), the draft tracker, the `📍` diagram, and the polluted-diff rule
- `commits` says what a revert would take with it, so each commit can be reverted alone
- tests in `skills/embed_test.go` pin both, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rules
- `go test ./...` is green

## Steps
- [ ] T1 · the chain says where it stands and the commit says its rollback   R1 · route:inline
- [ ] T2 · the changelog names the rules   R1 · route:inline

## History
L1 · "a chained PR says where it stands and a commit says what it can be reverted without: the chain's tracker, position, base, dependency and next PR, the `📍` in its diagram, a diff that shows only its unit, a branch named for what it delivers, and a rollback boundary per commit"
