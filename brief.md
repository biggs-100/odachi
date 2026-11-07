# the-chain-says-where-it-stands

## What must exist
- a chained PR says where it stands and a commit says what it can be reverted without: the chain's tracker, position, base, dependency and next PR, the `📍` in its diagram, a diff that shows only its unit, a branch named for what it delivers, and a rollback boundary per commit

## Done when
- `pull-request` names the branch prefixes (`feat/`, `fix/`, `docs/`, `chore/`, `refactor/`), the chain context (tracker, position, base, depends on, next PR, leaves out), the draft tracker, the `📍` diagram, and the polluted-diff rule
- `commits` says what a revert would take with it, so each commit can be reverted alone
- tests in `skills/embed_test.go` pin both, observed red before the text lands
- `odachi skills --check` is green and the bodies stay under the sixty-line cap
- `odachi init` leaves the installed copy equal to the tree
- the changelog's Unreleased section names the rules
- `go test ./...` is green

## Not included (first version)
- no issue-first, no GitHub labels and no Conventional Commits: they depend on a repo's CI, and odachi keeps the prose message with the type in the PR title
- no PR template with E2E, CodeRabbit or a contributor checklist
- no new verb and no new flag
- no change to `slices` or to `writing`
- no change to the voice or the interview's questions
- no change to the stored proofs: never touched
- no change to `odachi role` or to the skill it renders
- no change to the design canvas, its token or the push plugin
- no change to the artboard format
- no change to the release workflow
- no change to the skills the feature does not touch: tests, drive, slices, triage, writing, design and ui-craft

## Decisions taken
- the chain context lives in `pull-request` because it is the PR's shape; `slices` owns the split
- a diff that shows another PR's work is a branching bug: retarget or rebase until only this unit shows
- the rollback boundary lives in `commits` because a commit is the unit a revert takes
- branch naming is a rule, not a check: no new verb
- the repo-specific contract (issue-first, labels, conventional commits) is not copied: the skill is installed in other projects, and `pull-request` already says what to do when the repo uses issues
