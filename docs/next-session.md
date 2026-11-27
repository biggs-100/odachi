# Next session

Feature 0364 is closed locally on `feat/0364-grant-domain`. The final reviewed tree is
`1633db9ae44a88d1324e291c085c1559ba7216c4`; T11 committed at `8d69c50` and the task
document close committed at `86b8228`. Nothing in this handover authorizes push,
pull request, merge, tag, release or install.

## Current locations

- Authorized worktree: `C:\Users\USER\odachi-worktrees\0364-grant-slices`.
- Original checkout: `C:\Users\USER\Desktop\dev-biggz\odachi` still preserves the
  earlier aggregate grant diff; it is not authoritative for 0364.
- Final proof: `proof/0364-fresh-runtime-grants.txt`.
- Final review: approved on tree `1633db9ae44a88d1324e291c085c1559ba7216c4`
  against base `aa5c5f6c7ed366193ab162c80e7f436f4ec520ea`, with risk,
  readability, reliability and resilience lenses.
- Final independent verification: verified on the same tree.
- Latest published release remains `v0.65.0`; no release or install was changed.

## What 0364 delivered

- One-shot grant domain identities and validated lifecycle transitions.
- Durable atomic grant-record storage with permanent claim markers.
- Exactly-once claim under concurrency and terminal-dominant reconstruction.
- Session binding and deterministic terminal-race linearization.
- A fresh-only OpenCode adapter with no resume, continue or fork input.
- Managed delegation, review and verifier orchestration with distinct grants
  and fresh sessions per attempt, unit and draw.
- Tool-derived sealed runtime receipts, verification schema 2 and receipt-aware
  evidence export.
- Domain/check schema-2 reading and deterministic frozen-directory snapshots.
- An opt-in deterministic OpenCode runtime-control harness proving one spawn per
  grant, replay refusal without a second spawn, no resume argv, distinct actor
  sessions and candidate-mutation refusal.

## Decisions taken

- The human selected DeepSeek V4.1 Flash for fresh workers.
- T7 required a correction after independent verification found non-zero worker
  exits were recorded as successful.
- T10 required a correction after independent verification found model-supplied
  `by` identity could enter sealed evidence.
- T11 was re-sliced: schema 2/check and a deterministic runtime harness stay in
  0364; production actor transport, model binding and review mutation policy
  move to a future feature.
- The human accepted `Slice: 12000` because T1-T11 were independently reviewed
  and verified; the final aggregate reached 11495 authored lines.

## What is not yet delivered

T12 is deferred and must start as a new feature because the 400-line gate refuses
another step in 0364. It owns:

- production CLI actor transport for managed delegation, review and verification;
- selection and durable binding of the observed model;
- candidate-mutation refusal for managed review;
- live OpenCode/model E2E, separate from the deterministic control harness;
- cross-checking sealed receipts against the durable grant journal before stamping.

The current harness uses a deterministic control executable, not the installed
OpenCode binary or a live model. No user-facing managed CLI path is wired yet.

## Next actions

1. Decide whether to deliver 0364 first: push the branch and open the planned
   task-sized PR chain. This requires explicit human authorization.
2. Start the T12 follow-up feature for production runtime transport and policy.
3. Then resume 0363 T6: scope waivers and validate verdicts before paperwork
   exemption. 0363 T1-T5 are already merged; T6-T28 remain open.

## Verification and limits

- `odachi check --run "go test ./..." --timeout 1800` is green on the reviewed
  closing tree.
- The opt-in harness passes with `ODACHI_OPENCODE_RUNTIME_TEST=1` and skips in the
  ordinary suite.
- Non-blocking review warnings remain: duplicate review units, validation order
  before freeze, some refusal-branch coverage, stale staging recovery, and the
  managed-result comment. They are recorded in the final verdict, not hidden.
- No live model call, push, PR, merge, tag, release or install was performed.

The tool counts this line for the tree it describes, copied from `odachi summary`:

total: 362 features · 908 steps done of 931 · 0 open questions
