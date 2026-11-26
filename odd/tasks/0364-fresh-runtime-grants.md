# fresh-runtime-grants
Base: aa5c5f6c7ed366193ab162c80e7f436f4ec520ea
Slice: 4800 - five independently reviewable grant slices plus six OpenCode runtime tasks; the aggregate T1 exception is superseded by the re-sliced plan

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
- [x] T1 · define the one-shot grant domain contract   R1 · route:delegated/odachi-worker · skills:tests,security · tree:7625e2f3d11c754cee9230daccffc749bd1306e3
- [x] T2 · make grant storage durable and publish complete records atomically   R1 · route:delegated/odachi-worker · skills:tests,security,use-modern-go · tree:f50b91bff24e84fbd6878231d1ff8cb15b3e93c0
- [x] T3 · persist issued grants and claim each one once   R1 · route:delegated/odachi-worker · skills:tests,security · tree:bf1dd443817cd1ee572fc088e3d69fb4b4955fb8
- [x] T4 · reconstruct durable lifecycle with terminal dominance   R1 · route:delegated/odachi-worker · skills:tests,security · tree:0ee2021831526f3b7c3b8cf2e4ea1a9ba2572eb7
- [x] T5 · bind one session and linearize terminal races   R1 · route:delegated/odachi-worker · skills:tests,security · tree:5398fdeb95ffc5685f2fcf7c6fd9c9198fc02a3b
- [x] T6 · add a fresh-only OpenCode runtime adapter   R1 · route:delegated/odachi-worker · skills:tests,security,use-modern-go · tree:7e43667ad826d6a3e801d89807261c347e118ba5
- [x] T7 · bind managed delegation to one fresh worker session   R1 · route:delegated/odachi-worker · skills:tests,security,slices · tree:465ead7dc737e43f9c06d2f671c975a649b7fa9e
- [x] T8 · give every managed review unit and draw a fresh session on one frozen tree   R1 · route:delegated/odachi-worker · skills:tests,security,rdd-defect-workflow · tree:00cce6d0df9d3cb47ae81f4b9e259793b3f71f39
- [x] T9 · freeze and seal the managed verifier executor   R1 · route:delegated/odachi-worker · skills:tests,security,rdd-defect-workflow · tree:ef5fcd7c1f4de5a1d8f3d4d8e2d1498ffbc97417
- [ ] T10 · validate and export tool-sealed runtime receipts   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T11 · prove OpenCode freshness with opt-in runtime controls   R1 · route:delegated/odachi-worker · skills:tests,ship,security

## History
L1 · "odachi issues one-shot runtime grants for managed delegation, review and verification, binds each grant to one task or immutable candidate and one fresh runtime session, and refuses replay or resume in code"
L2 · OFICIO · "Accept a size exception for 0364 T1 at 1174 authored lines" (porque Grant identity, exclusive-create storage and concurrency/replay tests form one security invariant and must land together.)
L3 · OFICIO · "Update the 0364 T1 size exception to 1305 authored lines" (porque The final correction adds the deterministic BindSession/Fail race regression and post-create terminal linearization required by independent review.)
L4 · OFICIO · "Reopen 0364 T1 for durable and terminal-dominant grant storage" (porque Review warnings found missing parent-directory fsync, a terminal sampling window in Load, and Fail returning a stale session; crash and replay guarantees require a new candidate.)
L5 · OFICIO · "Update the 0364 T1 size exception to 1672 authored lines" (porque Durable directory sync, stable terminal-dominant reconstruction and Fail session preservation require platform-specific code and race regressions in one invariant.)
L6 · VOS · "Rebanear T1 (Recommended)" (porque The aggregate grant implementation reached 2363 authored lines and consumed almost the entire feature forecast.)
L7 · VOS · "Aceptar Slice 4800 (Recommended)" (porque Five grant slices plus six OpenCode runtime tasks require a realistic feature fence.)
L8 · OFICIO · "Renumber the runtime tasks and split aggregate T1 into five steps" (porque Whole-file slicing was insufficient; domain, publication, claim, lifecycle and session races have distinct review boundaries.)
L9 · OFICIO · "Preserve the re-slicing snapshot at C:\Users\USER\AppData\Local\Temp\opencode\0364-t1-snapshot with SHA-256 hashes" (porque The aggregate candidate must remain recoverable after the rewrite.)
L10 · OFICIO · "Use C:\Users\USER\odachi-worktrees\0364-grant-slices on feat/0364-grant-domain" (porque The original checkout preserves a different aggregate diff and is not authoritative.)
L11 · OFICIO · "Map old T1 to T1-T5 and old T2-T7 to T6-T11" (porque The re-slice separates domain, publication, claim, lifecycle and session-race review boundaries.)
L12 · OFICIO · "Forecast T1 403, T2 600, T3 500, T4 340, T5 470, T6 340, T7 320, T8 400, T9 380, T10 280 and T11 300 authored lines" (porque The realistic feature total is 4400-4600 authored lines.)
L13 · OFICIO · "Close T1 at commit 9eba382 after approved risk and reliability review plus independent verification" (porque The corrected candidate tree 7625e2f3d11c754cee9230daccffc749bd1306e3 has green proof and fixes both refutation gaps.)
