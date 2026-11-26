# fresh-runtime-grants
Base: aa5c5f6c7ed366193ab162c80e7f436f4ec520ea
Slice: 2400 - grant authority, one thin OpenCode adapter, managed delegate/review/verify, receipt readers and runtime controls are independently reviewable

## Requirements (verbatim, from the user)
R1 · "odachi issues one-shot runtime grants for managed delegation, review and verification, binds each grant to one task or immutable candidate and one fresh runtime session, and refuses replay or resume in code"

## Done when
- a managed invocation receives an opaque Go-issued grant claimed atomically once, and replay after success, failure, timeout or crash is refused
- each grant binds canonical worktree, role, runtime, model, task-document path/digest/step/requirement/agent/surfaces/commands or review candidate tree/base/lenses/draw
- the OpenCode managed adapter returns a fresh session identity and observed model, and Go binds both to the grant
- missing, malformed, duplicate or foreign session identities fail closed before evidence is stored
- managed adapters expose no resume operation and never pass session, continue or fork flags
- every worker attempt, reviewer process/unit/draw and verifier run uses a distinct grant and session; corrections use new sessions
- a tool-derived runtime receipt records grant, session, runtime, model, role, candidate/output trees, snapshot digest, timestamps and outcome
- verification schema 2 stamps the runtime receipt and ignores model-supplied executor claims
- review reuses its frozen archive and verify freezes the same candidate tree; snapshot mismatch before or after execution stores no evidence
- legacy `--exec` remains explicitly unmanaged and receives no freshness claim
- opt-in OpenCode runtime tests prove one spawn per grant, replay refusal without a second spawn, rejected resume input, distinct actor sessions and candidate-mutation refusal

## Steps
- [ ] T1 · issue and atomically claim one-shot runtime grants   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T2 · add a fresh-only OpenCode runtime adapter   R1 · route:delegated/odachi-worker · skills:tests,security,use-modern-go
- [ ] T3 · bind managed delegation to one fresh worker session   R1 · route:delegated/odachi-worker · skills:tests,security,slices
- [ ] T4 · give every managed review unit and draw a fresh session on one frozen tree   R1 · route:delegated/odachi-worker · skills:tests,security,rdd-defect-workflow
- [ ] T5 · freeze and seal the managed verifier executor   R1 · route:delegated/odachi-worker · skills:tests,security,rdd-defect-workflow
- [ ] T6 · validate and export tool-sealed runtime receipts   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T7 · prove OpenCode freshness with opt-in runtime controls   R1 · route:delegated/odachi-worker · skills:tests,ship,security

## History
L1 · "odachi issues one-shot runtime grants for managed delegation, review and verification, binds each grant to one task or immutable candidate and one fresh runtime session, and refuses replay or resume in code"
