# Next session: five features on main, and a tag that waits

Written at the close of 0171, on `main` at `376e872` -not on a published release, because the
publish is the decision that waits. `doctor` is green except check 14 until this file was rewritten;
`main` is clean and in sync with `origin`.

## What 0167, 0168, 0169 and 0170 shipped

- **0167 - the light lane's refusals are pinned.** One test per refusal: `--light` with `--brief`,
  a name that is already open, and a second positional argument.
- **0168 - a light feature is reviewed against its own line.** The review prompt read `brief.md`,
  which a one-line change never writes, so 0167 was judged against 0166's done-when. It now reads
  the task document whose base is the candidate's, quotes the requirement it carries and names the
  file. The domain is untouched: the lane still never chooses lenses.
- **0169 - `check --run` shows the command's output while it runs.** The proof is written when the
  process ends, so a run blocked on a lock printed nothing for seventeen minutes. The stream goes
  to stderr; `--json` on stdout stays one document.
- **0170 - one test proves the stream arrives while the command runs.** Its child is the test binary
  re-run: it prints to both streams and blocks until the parent releases it, so the assertion lands
  while the run is still open.
- **0171 - a closed light feature is judged against its line.** The last row of the debt: a test for
  a light document whose steps are all checked. It left the fence narrower than 0169 wrote it, and
  that narrowing had to land alone (below).

## What was measured today, so nobody re-measures it

- **The suite is ~300 s on this machine** (297 s and 301 s in two clean runs; ~94 s with
  `-count=1` and a warm build cache). `check --run` still wants `--timeout 1500` to be safe.
- **A killed command in Windows leaves children behind, and they poison the next suite.** Symptom:
  `go test` prints nothing for minutes while nothing seems to be running; the tool then reports a
  timeout with empty output. Look for stray `go`, `shellcmd.test` or `odachi` processes and kill
  them before believing the suite is slow. 0170 bounded its own child so that this file's test can
  no longer be the cause, but the generic friction is still open.
- **The review runs in one command now:**
  `odachi review --base <sha> --exec 'opencode run --agent odachi-reviewer --model {model} --file {prompt} "answer with the JSON only"'`.
  `{model}` is required when a model is assigned, and PowerShell needs single quotes around the
  whole command. No subagent, no pasted prompt.
- **The coverage gate refuses a verdict that misses a row.** A finding on one path does not cover
  another: every changed path needs a finding or a `passes` row under the lens.
- **The order that survives the hook:** commit the step, then `odachi plan --done T1` writes the
  tree while the worktree is clean. Marking a step done before committing closes the feature and the
  hook demands a proof that does not exist yet; `odachi undo --artifact <doc>` puts the mark back.
- **`start` rewrites `odd/fence.toml` from the non-goals; a light start writes only the task doc.**
  The architecture test requires a rule over `proof/**` and refuses any rule over
  `internal/domain/proof/**` (0150): the fence guards the evidence, never the writer.
- **A `Risk:` line of `contract` plus a diff that touches concurrency earns four lenses**
  (risk, readability, reliability, resilience). All four ran in one `--exec` call on
  `opencode-go/mimo-v2.6-flash`.
- **A fence can be too wide for its own reason, and narrowing it needs a policy PR.** 0169 guarded
  "how the reviewer runs" with a glob over `internal/app/review/**`, which also froze the tests that
  pin the lane. CI reads the fence from the base ref, so the feature that needed the narrower path
  was refused on all three platforms; `policy/0171` landed first, alone (PR #266), and the feature
  rebased on top. A light feature cannot do this at all: the lane never rewrites the fence, and the
  hook refuses its commit. That is the lane's own signal to leave it.
- **The local hook needs a proof of the branch's own tree.** A policy-only commit on top of `main`
  is refused until `check --run` runs there too: the proof directory belongs to the clone, and the
  newest document of `main` is not the feature being built.

## What is open

- **The review debt is empty**: `docs/review-debt.md` has no open row, for the first time since 0154.
- **The publish.** `v0.11.100` is a commit on `main` (`99a1f75`) whose tag was never pushed, and it
  is what publishes 0167's changelog section. 0168 to 0171 wait under `## Unreleased` for the next
  one. The human decides when.
- **The change that needs no document** is still only half built: the light lane is the cheap half,
  and the half without a task doc has never been measured.
- **A proof that says what it is blocked on** was half paid by 0169: the output streams now, so a
  lock is visible, and nothing names the child process or the lock itself.

## What is already true, so it is not re-decided

- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **What a parser cannot read, it refuses.** Tolerating a line is not the same as guessing it.
- **A fence rule names a file the feature does not touch**, and never the package that writes the
  proofs (0150). A phrase is not a rule.
- **Every irreversible action names its target and waits for the human.**
- **A warning earns no review round**: it becomes a row in `docs/review-debt.md`, paid by the next
  feature that touches those bytes. 0170 is that mechanism working end to end.
- **A test-only change opens no changelog bullet.** It changes nothing a person can observe.

## The one decision that waits, and is not polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that runs
`odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **a change resumable from the diff needs no document.** Measure it before building it: `shape`
  first, and nothing is written until the shape says what the artifact would be.
- **publish v0.11.101** (and v0.11.100 first, if its tag is worth pushing at all), then install from
  the signed bytes instead of a local build.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has
named; the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says
which class of file grew back.
