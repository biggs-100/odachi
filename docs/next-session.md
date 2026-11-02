# Next session: six features, one release, and the doctrine that names it

Written at the close of 0172, on **v0.11.101** -published, signed, attested and installed on this
machine. `doctor` is green except check 14 until this file was rewritten; `main` is clean and in sync
with `origin`.

## What shipped

- **0167 to 0171** closed the review debt and the light lane's gaps: the lane's three refusals, a
  light feature judged against its own line, `check --run` streaming while it runs, the test that
  proves the stream arrives in time, and a closed light document still naming its line. Five features,
  and only the second and third changed what a person can observe.
- **0172 wrote the release doctrine** in `docs/releases.md`, pointed `drive` at it, and deleted
  `v0.11.100` -a heading whose tag was never pushed, so no bytes anyone could install.
- **v0.11.101 is the release the doctrine produced**: one version for the two observable changes, notes
  copied from the changelog section, the tag on the merge commit, installed from the release's own
  installer.

## The doctrine, in the shortest form

A release is a tag someone can install. Publish when what a person can do or trust changed -never for
tests, CI, docs or a refactor- with `minor` when the contract moves and `patch` otherwise while `0.x`.
The tag is annotated, sits on the release pull request's merge commit, and is never moved. The notes
are the changelog section verbatim. The whole rule is in `docs/releases.md`; do not re-decide it in
chat.

## What was measured today, so nobody re-measures it

- **The suite is ~300-430 s on this machine** (297, 301, 325, 427 s in four clean runs). `check --run`
  still wants `--timeout 1500`.
- **A killed command in Windows leaves children behind** (`go`, `shellcmd.test`) that hold the build
  cache; the next suite prints nothing for minutes and reads as a hang. Look for stray processes
  before believing the suite is slow.
- **The review runs in one command:**
  `odachi review --base <sha> --exec 'opencode run --agent odachi-reviewer --model {model} --file {prompt} "answer with the JSON only"'`.
  PowerShell needs single quotes; `{model}` is required when a model is assigned.
- **A fence can be too wide for its own reason, and narrowing it needs a policy pull request.** CI
  reads the fence from the base ref; a light feature cannot rewrite it at all, which is the lane's own
  signal to leave it (0171, PR #266).
- **`start` rewrites `odd/fence.toml` from the non-goals**; a light start writes only the task doc.
  The architecture test requires a rule over `proof/**` and refuses any rule over
  `internal/domain/proof/**` (0150).
- **The order that survives the hook:** commit the step, then `plan --done T1`. Marking a step done
  first closes the feature and the hook demands a proof; `undo --artifact <doc>` puts the mark back.

## What is open

- `docs/review-debt.md` holds one row: the fence header explains a rule the file does not have.
- **The installer verifies sha256, not the signature.** `checksums.txt` is signed and attested; the
  script does not check the bundle with cosign, so the chain of trust ends at GitHub today.
- **The release body is written by hand** with `gh release edit` after the tag, because goreleaser
  generates notes from commits. It worked once; a step that must be remembered is a step that will be
  forgotten.
- **The change that needs no document** is still only half built: the light lane is the cheap half.
- **A proof that says what it is blocked on** was half paid by 0169: the output streams, and nothing
  names the child process or the lock.

## What is already true, so it is not re-decided

- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The release doctrine lives in `docs/releases.md`** and `drive` names it in one line.
- **A version heading with no tag is not a release**, so it may stop existing.
- **The review gate lives in the clone** (ADR 0042). CI cannot verify a review.
- **Ownership is declared, never inferred** (ADR 0038), and a failed read is not an absence.
- **A fence rule names a file the feature does not touch**, never the package that writes the proofs.
- **Every irreversible action names its target and waits for the human.**
- **A warning earns no review round**: it becomes a row in `docs/review-debt.md`.

## The one decision that waits, and is not polish

**CI cannot enforce the review.** ADR 0042 says what would change it: a trusted job that runs
`odachi review` with its own model and key, or a receipt signed outside the tree.

## Queued

- **the row, paid**: the fence header that describes a rule it does not have.
- **the installer verifies the signature**: `cosign verify-blob` against `checksums.txt.sigstore.json`
  before installing, so the chain ends in the signature and not in GitHub.
- **the release body comes from the changelog**, not from memory: a step of the release flow or a
  verb, so nobody has to remember `gh release edit`.
- **a change resumable from the diff needs no document.** Measure it before building it: `shape`
  first.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has named;
the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says which class
of file grew back.
