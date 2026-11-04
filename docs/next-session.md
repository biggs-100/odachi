# Next session: the model a project can afford, and the gate that still waits

Written at the close of 0192. `main` is clean at aac0435, the four checks run on
`gate (macos-15-intel)`, and the local proof is the one taken with `-count=1`.

## What shipped

- **0189, PR #312**: `scripts/install.ps1` opens with `#Requires` and no BOM, so the
  documented `irm | iex` prints no error; the direct-file path pays ANSI accents under
  PowerShell 5.1.
- **0190, PR #313**: the two failure branches of `prunePrompts` are pinned - a `ModTime`
  that fails keeps every prompt, a mid-sweep `Remove` reports what went and what stayed -
  with `failModTime` / `failRemove` on the fake.
- **0191, PR #315**: `check --run` streams only to a terminal, so a pipe nobody drains can
  no longer block the command, and a timeout kills the whole process tree (`taskkill /T`
  on Windows, the process group on Unix).
- **0192, PR #314**: the macOS gate runs on `macos-15-intel`; the ruleset requires
  `gate (macos-15-intel)`. GitHub's arm64 pool never assigned the old job.

No release since v0.14.7: 0189-0192 are on `main`, unreleased.

## The lessons this session paid for

**A gate cannot stream to a writer nobody drains.** The live copy went to stderr, and a
caller that captures stderr without reading it filled the pipe: the command blocked and
the gate waited for its own timeout. The proof keeps both streams, so the live copy is a
courtesy a terminal can accept and a pipe cannot.

**A timeout that kills the parent is not a timeout.** `cmd`/`sh` died and `go test`
survived, writing to a closed pipe; the orphan then poisoned every retry. The cancel now
ends the tree.

**A required check that never gets a runner is not a gate.** `macos-latest` (arm64) was
never assigned; GitHub cancelled the job after 15 minutes on every attempt. The check
was required, so every pull request stayed blocked by infrastructure, not by code.

## What was measured, so nobody re-measures it

- The CI on the new label: macOS 2m44s-3m50s, Windows 2m45s-5m3s, Ubuntu 1m18s-1m36s,
  `checks (linux)` 1m0s-1m20s. The push gate on aac0435 is green.
- The local proof for `main` is `proof/0192-the-macos-gate.txt`, command
  `go test ./... -count=1`. The same suite with the plain `go test ./...` under `cmd /c`
  hangs locally; the CI does not reproduce it. `-count=1` finishes in about three minutes.
- The store is 39.2 MB; `odachi clean --apply` would free 38.2 MB in 14 files. The switch
  is the human's.
- The configured reviewer (`mimo-v2.6-flash`) stalled twice - on 0190's and 0192's review -
  and DeepSeek answered both in seconds. The project's model decision is still open; see
  the queue.

## Queued

1. **The fence correction, then the model feature.** The fence from 0191 protects
   `odd/review.toml` and `internal/domain/review/cleanup.go` under "no change to the
   retention numbers or which trees survive". The review's model policy lives in that
   same file, so a pull request that changes `[lenses]` is refused by the base fence. A
   normal `start` (not light) whose brief does not carry the retention non-goal
   regenerates the fence without that rule and changes only `brief.md`, `fence.toml` and
   its task doc; it lands first. Then the model feature can name the model it can afford:
   DeepSeek V4.1 Flash (the agent's own) or a free model available on the host, never a
   model switch without the human's yes, the same model allowed to review when the
   project owns one (the `review waive` lane, ADR 0048), and a failure that names the
   model, the lens and the free/configured options before anything else runs.
2. **The review debt 0191 left.** `docs/review-debt.md` has three open rows: `killTree`
   discards taskkill's result, so a missing or blocked taskkill leaves the grandchild and
   reports a clean timeout; the comment above the renamed JSON test describes the stream
   the gate no longer sends; and the positive branch of `streamTarget` has no test. A
   WARNING earned no round; the next feature that touches the shell or the gate pays them.
3. **The local no-count hang.** `go test ./...` through `cmd /c` hangs locally while
   `-count=1` finishes and the CI is green. Measure whether it is the test cache, the
   wrapper or a test that runs only with the cache, and pin it or name it in the ledger.
4. **A release for 0189-0192.** The installer fix and the gate fix are the user-facing
   ones; the release is a tag for a person's change, and it is the human's call.

## What is already true, so it is not re-decided

- `main` is protected by four checks; the macOS one is `gate (macos-15-intel)`; the
  ruleset is applied from `.github/main-ruleset.json` by `scripts/protect-main.ps1`.
- The gate streams only to a terminal, and a timeout kills the tree (0191).
- The local proof for `main` is the `-count=1` one; the plain command is the CI's.
- The 0192 feature's tree has no verdict at its base: the merge brought 0190 and 0191.
  The local check names it; the CI does not check the review (ADR 0043).
- `odd/review.toml` is `mode = "on"` and names `mimo-v2.6-flash` for the four lenses.
  Changing it is the project's decision and a change of its own.
- v0.14.7 is the last release; 0189-0192 are unreleased.
- The store frees 38.2 MB with `odachi clean --apply`; the switch is the human's.
