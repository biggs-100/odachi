# Next session

Session closed on 2026-11-26. The repository is between delivered workflow-hardening
slices and the first re-sliced runtime-grant slice. Nothing in this handover authorizes
push, pull request, merge, tag, release or install.

## Current locations

- Authorized worktree: `C:\Users\USER\odachi-worktrees\0364-grant-slices`.
- Worktree branch: `feat/0364-grant-domain` at `a00f70ca8c276ddea41d21a917ea214123fab0ff`.
- Original checkout: `C:\Users\USER\Desktop\dev-biggz\odachi`, branch
  `feat/0364-fresh-runtime-grants`, with the preserved aggregate grant diff still
  uncommitted.
- Local `main` is `e41645cc74cba5403d30bd4fa64283a8b4fe9a99`; `origin/main` is
  `aa5c5f6c7ed366193ab162c80e7f436f4ec520ea`. The difference is the unpushed 0364
  opening commit.
- Full aggregate T1 snapshot, including SHA-256 hashes:
  `C:\Users\USER\AppData\Local\Temp\opencode\0364-t1-snapshot`.
- Current development binary used for odachi commands:
  `C:\Users\USER\AppData\Local\Temp\opencode\odachi-dev.exe`.

## Delivered this session

- v0.65.0 remains the latest published release; its tag and installed binary were not
  changed.
- Workflow-hardening feature 0363 delivered T1-T5:
  - PR #730 / `c80049c`: one candidate base before check/review snapshot.
  - PR #731 / `79774d5`: post-command proof tree identity.
  - PR #732 / `0847c5d`: atomic correction snapshot.
  - PR #733 / `2a0a3c4`: fail-closed refutation evidence and shared citation validator.
  - PR #734 / `aa5c5f6`: measured declared-check evidence and `EffectiveVerdict` for
    every stored-verdict reader.
- `main` is clean at the delivered 0363 boundary before the 0364 opening commit.
- A read-only study of `gentle-ai` at `50151955a82926fc62c95190cd5c5a7fb4b855f6`
  confirmed the direction: Go owns authority, adapters are thin, runtime freshness is
  enforced in code, and real-runtime negative controls are separate from mocks.

## Open features

### 0363 workflow-hardening

- T1-T5 are merged. T6-T28 remain open.
- Next unfinished task is T6: scope waivers and validate verdicts before paperwork
  exemption.
- Do not resume 0363 until the fresh-runtime-grant lifecycle work is delivered.

### 0364 fresh-runtime-grants

- Opened on local `main` as `e41645c`; task doc is
  `odd/tasks/0364-fresh-runtime-grants.md`.
- Human accepted `Slice: 4800` and re-sliced aggregate T1 into five steps.
- Current steps are T1-T11:
  - T1 domain grant contract.
  - T2 durable atomic publication primitives.
  - T3 issued grants and one-time claim.
  - T4 durable lifecycle reconstruction.
  - T5 session binding and terminal races.
  - T6 fresh-only OpenCode adapter.
  - T7 managed delegation.
  - T8 managed review.
  - T9 managed verifier.
  - T10 runtime receipts.
  - T11 opt-in runtime controls.
- Legacy `--exec` remains unmanaged. Claude/Codex are deferred. Live model E2E needs
  separate authorization.

## Exact stopping point

The first re-sliced candidate is T1, domain-only:

- Files are present but uncommitted in the authorized worktree:
  `internal/domain/grant/grant.go` and `grant_test.go`.
- They are byte-identical to the preserved snapshot.
- Review tree `d3a100a43302a33ac0ce575fb7c8cf041ebf66fd` is `approved` with lenses
  `risk` and `reliability`, zero findings.
- Independent verification is `refuted` and stored under the shared clone store.
- Full `go test ./...` passed on this candidate.

The refutation found two concrete gaps:

1. `BindSession` accepts Unicode Cf format controls U+200B, U+202E, U+2066 and U+FEFF.
   Restrict session identity to a strict safe ASCII grammar or explicitly reject Cf.
2. `Claim` checks only `StateIssued`; an invalid grant can transition and remain
   invalid. Validate before/after transitions and add direct tests for `Validate` and
   `State.String`.

No correction worker was started before session close.

## Next actions

1. Stay in `C:\Users\USER\odachi-worktrees\0364-grant-slices`.
2. Start a fresh `odachi-worker` for the T1 correction; do not reuse any previous
   worker, reviewer or verifier session.
3. Add RED tests for Unicode Cf/session grammar and invalid-grant Claim, implement the
   smallest domain fix, then run focused tests, lint and diff checks.
4. Use fresh reviewer and verifier sessions for the corrected candidate. One correction
   per immutable candidate; another blocker means a new candidate and human decision.
5. Commit T1 only after review, independent verification and full suite are green.
6. Deliver T1 before starting T2. Continue T2-T5, then T6-T11, then resume 0363 T6.

## Process rule learned the hard way

Subagents are fresh per task and per immutable candidate. Reusing a worker/reviewer/
verifier session across corrections contaminated context and produced repeated missed
readers and race paths. The final T5 candidate used fresh sessions and the fresh
verifier found the last omitted reader. Do not regress to session reuse.

## Limits

- The aggregate grant diff in the original checkout is preserved, not authoritative.
  Use the snapshot and the re-sliced worktree.
- Real power-loss behavior was not injected; durability relies on OS/filesystem
  guarantees and platform-specific tests.
- `.atl/skill-registry.md` remains stale and belongs to another tool.
- No push, PR, merge, tag, release or install is authorized for 0364.
