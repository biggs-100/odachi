# obligatorio-review-y
Base: 7f930d163ee12560aa95960005bec32af6b782d2
Slice: 3200 - one cohesive policy change; the additional lines make the hook read staged project artifacts and policy while keeping proof, review and verification evidence on disk, with regressions for each boundary

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
- [x] T4 · make CI local-only review contract explicit without running a model   R1 · route:delegated/odachi-worker · skills:tests · tree:97b53eceeb67d68f00f2516f812e38e4786aa4b5
- [x] T5 · teach the mandatory contract in generated rules, skills, help, docs and ADRs   R1 · route:delegated/odachi-worker · skills:writing · tree:5be80a4dcc7915d7de475f1f16fb29b5beae7edc

## History
L1 · "obligatorio, review y verificasion"
L2 · PREGUNTA P1 · "Do you accept Slice: 1900 - one cohesive policy change spanning gate, hook, evidence, CI and generated instructions; splitting would separate the contract from its enforcement and docs?"
L3 · VOS P1 · "Slice: 1900 - one cohesive policy change spanning gate, hook, evidence, CI and generated instructions; splitting would separate the contract from its enforcement and docs"
L4 · OFICIO · "Pay the deterministic lint floor inside the closing candidate instead of opening T6" (porque The closing review found QF1001 and an unused function; plan refuses a new step because the accepted feature already exceeds 400 authored lines, and these mechanical fixes belong to the existing close.)
L5 · OFICIO · "Unreadable current risk evidence blocks even when readable history is valid" (porque The stored review found that discarding the current message read error lets a closing commit hide a verifier-requiring risk; readable history may still contribute known lenses, but it cannot replace unreadable current evidence.)
L6 · OFICIO · "An unchanged closing tree still validates its stored verdict" (porque Independent verification reproduced a fail-open where an empty delta since a parseable but incomplete verdict skipped requireReview; empty means the verdict names this exact tree, so the gate must still validate its lenses and coverage.)
L7 · PREGUNTA P2 · "Do you accept Slice: 2000 - one cohesive policy change; the additional lines are the regression and fix for a review-validation bypass found by independent verification?"
L8 · VOS P2 · "Slice: 2000 - one cohesive policy change; the additional lines are the regression and fix for a review-validation bypass found by independent verification"
L9 · OFICIO · "Commit-message checks snapshot the staged index, not the worktree" (porque Independent verification proved that an unstaged deletion could hide staged code from the commit-msg gate; candidate bytes, coverage blobs and dependency fences must all describe what Git will commit.)
L10 · PREGUNTA P3 · "Do you accept Slice: 2400 - one cohesive policy change; the additional lines make the commit hook read staged candidate bytes, coverage blobs and dependencies, with regressions for each bypass?"
L11 · VOS P3 · "Slice: 2400 - one cohesive policy change; the additional lines make the commit hook read staged candidate bytes, coverage blobs and dependencies, with regressions for each bypass"
L12 · OFICIO · "The commit hook reads project bytes and policy from the index while evidence stays on disk" (porque Independent review proved that staging only the Git candidate still let worktree task docs, brief, fence and review policy contradict the commit; project reads now share the staged snapshot, while proof, verdict and verification records remain clone-local evidence.)
L13 · PREGUNTA P4 · "Do you accept Slice: 3200 - one cohesive policy change; the additional lines make the hook read staged project artifacts and policy while keeping proof, review and verification evidence on disk, with regressions for each boundary?"
L14 · VOS P4 · "Slice: 3200 - one cohesive policy change; the additional lines make the hook read staged project artifacts and policy while keeping proof, review and verification evidence on disk, with regressions for each boundary"
L15 · OFICIO · "Explicit proof paths are real evidence even inside the project" (porque The raw review found that --proof manual.txt was still routed through the staged index; explicit proof, like proof/ and clone-store evidence, must be read from the real filesystem during a commit-message check.)
