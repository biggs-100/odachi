# the-block-closes
Base: 0be787bfa04844af83bb24a3f5aa5a0c1aa98924

## Requirements (verbatim, from the user)
R1 · "a session closes by rewriting `docs/next-session.md`, the block says so where the agent reads it, and `doctor` names the file when it went stale"

## Done when
- the block carries a closing paragraph: the phrase a human uses, the file to rewrite, and the check that says it is stale
- `doctor` has a fourteenth check: a next-session file that names an older feature than the newest one is a warning, never a refusal
- the help and the changelog say fourteen

## Steps
- [ ] T1 · the block closes the session: the phrase, the file and the check   R1
- [ ] T2 · doctor has a fourteenth check that warns when the next-session file names an older feature   R1
- [ ] T3 · the help and the changelog say fourteen   R1

## History
L1 · "a session closes by rewriting `docs/next-session.md`, the block says so where the agent reads it, and `doctor` names the file when it went stale"
