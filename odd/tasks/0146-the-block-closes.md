# the-block-closes
Base: 0be787bfa04844af83bb24a3f5aa5a0c1aa98924

## Requirements (verbatim, from the user)
R1 · "a session closes by rewriting `docs/next-session.md`, the block says so where the agent reads it, and `doctor` names the file when it went stale"

## Done when
- the block carries a closing paragraph: the phrase a human uses, the file to rewrite, and the check that says it is stale
- `doctor` has a fourteenth check: a next-session file that names an older feature than the newest one is a warning, never a refusal
- the help and the changelog say fourteen

## Steps
- [x] T1 · the block closes the session: the phrase, the file and the check   R1 · tree:7dfe1bd4bfc5ae0e5d8dcfcaf696c8328d5e3fd2
- [ ] T2 · doctor has a fourteenth check that warns when the next-session file names an older feature   R1
- [ ] T3 · the help and the changelog say fourteen   R1

## History
L1 · "a session closes by rewriting `docs/next-session.md`, the block says so where the agent reads it, and `doctor` names the file when it went stale"
L2 · OFICIO · "T2 and T3 land in one commit: the check, the count, the usage sentence and the changelog" (porque the count, the usage string and the changelog are one fact, and splitting them ships a commit where the summary says fourteen while the help still says thirteen)
