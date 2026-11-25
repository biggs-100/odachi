# workflow-hardening
Base: 61950d193b8a0f442397f1d47fe94da6f6866eaf
Slice: 4000 - workflow hardening spans candidate identity, fail-open evidence, delegation, state, validation and release; each task remains independently reviewable

## Requirements (verbatim, from the user)
R1 · "odachi's workflow fails closed when evidence is ambiguous, judges one consistent candidate snapshot, enforces the delegation and state boundaries it claims, and publishes only verified artifacts"

## Done when
- `check` and `review` resolve one candidate base before reading tree, paths and diff, and distinguish no remote from an unresolved remote default
- a successful `--run` proof names the post-command tree or refuses when candidate bytes changed
- correction tree, paths and diff come from one atomic snapshot
- unreadable inherited evidence blocks inheritance; refutations are tree-bound and carry validated citations
- every declared check is approved and run or represented by a blocking row, and reviewer JSON cannot suppress deterministic rows
- waivers cannot cover later trees or override `--require-review`; paperwork cannot hide an unresolved or incomplete verdict
- verification seals the executor that actually ran and independence follows an explicit policy
- changing cited bytes cannot silently resolve a predecessor blocker
- `delegate` rejects foreign or escaping task documents and empty, escaping or repository-wide surfaces and commands
- skill slugs are safe and authoritative source skill bytes win over stale installed copies
- every step has a valid requirement and route, and delegated prompts name the agent
- `plan --done --commit` accepts only a readable commit in feature ancestry and not already claimed by another step
- branch and hook documentation matches actual enforcement, and linked worktrees receive the hook through the common Git directory
- `ask`, `plan` and review waivers serialize read-modify-write and back up before replacement
- stale-lock handling can never delete a replacement lock
- backups are worktree-scoped and preserve A-to-B-to-A history
- installed-skill comparison reports missing source skills, changed files and obsolete extra files
- unknown fence keys, incomplete briefs and contradictory `--light` inputs are refused
- requirements use sequential unique quoted `R<n>` IDs, every outcome is mapped, and voice and execution use one profile resolver
- `--since` resolves only verdict stems and floor parsing is hunk-aware
- continuations are executable and platform-correct, and JSON preserves typed transitions
- release notes and tag shape are validated before publishing, and release notes advertise only supported installer paths
- installed-binary verification tests the published artifact rather than a fresh local build
- `docs/next-session.md` names the current feature and measured state
- every fixed bypass has a regression that fails on the pre-fix tree and passes after the fix
- `go test ./...` and `odachi check --run "go test ./..."` are green

## Steps
- [x] T1 · resolve one candidate base before check and review snapshot   R1 · route:delegated/odachi-worker · skills:tests,commits · tree:35846d704aceae6e169df9c0c938c942f086eccc
- [x] T2 · bind proof to the post-command tree or refuse candidate mutation   R1 · route:delegated/odachi-worker · skills:tests,commits · tree:b669bff99e499edc9dd614e78bd495f4944cfd4b
- [x] T3 · return one atomic correction snapshot   R1 · route:delegated/odachi-worker · skills:tests,commits · tree:46db5e42907ba469f1272eb95ba4b4f10ca1e680
- [ ] T4 · fail closed on unreadable inherited evidence and bind refutations   R1 · route:delegated/odachi-worker · skills:tests,rdd-defect-workflow
- [ ] T5 · require declared checks to run or produce blocking rows   R1 · route:delegated/odachi-worker · skills:tests,rdd-defect-workflow
- [ ] T6 · scope waivers and validate verdicts before paperwork exemption   R1 · route:delegated/odachi-worker · skills:tests,rdd-defect-workflow
- [ ] T7 · seal and enforce verifier executor identity   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T8 · keep moved blockers causal until re-reviewed or waived   R1 · route:delegated/odachi-worker · skills:tests,rdd-defect-workflow
- [ ] T9 · validate delegation task paths surfaces and commands   R1 · route:delegated/odachi-worker · skills:tests,security,slices
- [ ] T10 · validate skill slugs and prefer authoritative source bytes   R1 · route:delegated/odachi-worker · skills:tests,security,skill-craft
- [ ] T11 · require valid step requirement route and named agent   R1 · route:delegated/odachi-worker · skills:tests,slices
- [ ] T12 · validate done commit ancestry and uniqueness   R1 · route:delegated/odachi-worker · skills:tests,commits
- [ ] T13 · align branch hook documentation and install hooks in the common git directory   R1 · route:delegated/odachi-worker · skills:tests,ship
- [ ] T14 · refuse stale lock takeover instead of deleting a replacement lock   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T15 · lock and back up task document rewrites   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T16 · scope backups to the worktree and preserve content history   R1 · route:delegated/odachi-worker · skills:tests,security
- [ ] T17 · compare installed skills exactly including obsolete files   R1 · route:delegated/odachi-worker · skills:tests,ship
- [ ] T18 · reject unknown fence TOML keys   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T19 · refuse incomplete imported briefs   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T20 · reject contradictory light lane inputs   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T21 · enforce sequential unique quoted requirement IDs   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T22 · use one profile resolver for start voice and execution   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T23 · resolve since only against verdict stems   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T24 · parse floor file headers only outside diff hunks   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T25 · make continuations executable and preserve typed JSON transitions   R1 · route:delegated/odachi-worker · skills:tests,writing
- [ ] T26 · validate release notes and tag shape before publishing   R1 · route:delegated/odachi-worker · skills:tests,ship,security
- [ ] T27 · verify published artifact payload bytes   R1 · route:delegated/odachi-worker · skills:tests,ship
- [ ] T28 · refresh the next-session handover   R1 · route:delegated/odachi-worker · skills:writing

## History
L1 · "odachi's workflow fails closed when evidence is ambiguous, judges one consistent candidate snapshot, enforces the delegation and state boundaries it claims, and publishes only verified artifacts"
L2 · PREGUNTA P1 · "When origin exists but its default ref is not fetched, must check and review block and require an explicit --base, or may they silently omit trailer handling? A genuinely no-origin clone may keep the current closed behavior."
L3 · VOS P1 · "Bloquear: si origin existe pero su default ref no está fetcheado, check y review fallan y piden fetch o --base; un clone sin remote conserva el comportamiento actual."
L4 · OFICIO · "Authorize internal/app/review/review_test.go and internal/app/review/refusals.baseline as T1 surfaces" (porque The shared candidate context requires review tests to exercise merge-base and branch boundaries, and the retired raw-base refusal must leave the pinned baseline.)
L5 · OFICIO · "Accept a size exception for T1 at 796 authored lines" (porque The shared candidate resolver and its real-Git regressions form one cohesive contract fix; splitting tests from behavior would weaken review, and later slices return to the 400-line budget.)
L6 · OFICIO · "Deliver workflow hardening slice 1 in PR #730 and start slice 2 from c80049c" (porque Stacked-to-main keeps each work unit independently reviewable; T2 proof identity depends on the shared candidate context merged in T1.)
L7 · OFICIO · "Deliver workflow hardening slice 2 in PR #731 and start slice 3 from 79774d5" (porque Stacked-to-main keeps proof identity independent; T3 atomic correction snapshots build on the merged candidate context and proof behavior.)
L8 · OFICIO · "Authorize cmd/odachi/files.go as a T3 surface" (porque The atomic Correction signature must be wired through the real Git adapter without breaking the review port boundary.)
