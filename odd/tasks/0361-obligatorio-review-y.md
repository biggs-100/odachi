# obligatorio-review-y
Base: 7f930d163ee12560aa95960005bec32af6b782d2

## Requirements (verbatim, from the user)
R1 · "obligatorio, review y verificasion"

## Done when
- with review mode on, a non-passive close without a valid verdict blocks instead of printing a note
- with review mode on, a named risk without an independent verification blocks
- a strict non-passive close declares a valid risk, and Risk none carries a written reason
- the commit hook evaluates review and verification for the closing commit
- a clean close reads lenses and coverage from the feature diff, not an empty worktree
- executable and test files under docs or root design are not passive documentation
- CI states its local-only review contract with no model call
- the generated rules, skills, help, README and ADRs teach the same contract

## Steps
- [x] T1 · parse a written reason for Risk none and classify code under docs or design as code   R1 · route:delegated/odachi-worker · skills:tests · tree:d6e013a4c10634420055f204452a282be35067a6
- [x] T2 · make review mode block missing review and required verification, including the commit hook   R1 · route:delegated/odachi-worker · skills:tests · tree:eed6c39be3fd600ef07c374412a5fe85796edb71
- [x] T3 · read lenses and coverage from the feature diff and refuse invalid risk declarations   R1 · route:delegated/odachi-worker · skills:tests · tree:8c76c3e76c4c5230ed6e08c589506faee0cabbaa
- [ ] T4 · make CI local-only review contract explicit without running a model   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T5 · teach the mandatory contract in generated rules, skills, help, docs and ADRs   R1 · route:delegated/odachi-worker · skills:writing

## History
L1 · "obligatorio, review y verificasion"
