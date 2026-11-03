# Next session: the installer stops trusting the host, and the first minor the doctrine produced

Written at the close of 0174. **v0.12.0 is published**, signed and attested; the binary installed on
this machine is still **v0.11.102**, because v0.12.0 requires cosign to install and this machine does
not have it yet. `doctor` is green except check 14 until this file was rewritten; `main` is clean and
in sync with `origin`.

## What shipped

- **0174 made both installers verify the release signature before they read a hash.** They download
  `checksums.txt.sigstore.json`, ask cosign to check the checksum file against this repository's
  release workflow and the tag, and refuse when cosign is missing -a check that runs only sometimes
  is no check, and this project already refuses instead of degrading.
- **v0.12.0 is the first `minor` the release doctrine produced**, because installing now requires a
  step it did not before: cosign on the PATH. The changelog says it, the release notes say it twice.
- **The review caught a regression this branch introduced**: the edit that added the signature check
  to `install.ps1` dropped the file's UTF-8 BOM, and Windows PowerShell 5.1 decodes a BOM-less script
  with the ANSI code page, so every Spanish accent would have printed as mojibake. The BOM is back,
  and nothing in the tests would have noticed without the reviewer.

## What was measured, so nobody re-measures it

- **`opencode run` stalls at ~40 KB on the installer diff.** Two runs hit the 600 s limit with no
  answer, twice on the same prompt; a subagent reading the prompt file answered in seconds. When the
  one-command review stalls, hand the prompt path to a subagent before believing the model is broken.
- **A file's encoding is part of its bytes.** `install.ps1` is UTF-8 with BOM on purpose; an editor or
  a scripted rewrite that drops the BOM breaks every accent on Windows PowerShell 5.1, and the tests
  see nothing because they match ASCII.
- **The pre-commit hook refuses a commit whose proof is stale.** The order that survives it: stage
  the correction, run `check --run` on that tree, then commit.
- **PowerShell 5.1 and `pwsh` are different shells** and the runner has both; the Windows installer
  test prefers `powershell`, which is the one users have.
- The suite is ~400 s with the new installer tests; `check --run` still wants `--timeout 1500`.

## What is open

- **This machine cannot install v0.12.0 until cosign is installed.** `go install
  github.com/sigstore/cosign/v2/cmd/cosign@latest` or the sigstore release binary; then
  `ODACHI_VERSION=v0.12.0 scripts/install.ps1` proves the whole chain on the real release.
- `docs/review-debt.md` holds two rows: the Windows branch that refuses without cosign has no test,
  and the fence header quotes a decision the brief never took.
- **The release body is still written by hand** with `gh release edit` after the tag.
- **The change that needs no document** is still only half built.
- **ADR 0042**: CI cannot enforce the review. A decision, not work.

## What is already true, so it is not re-decided

- **A release is a tag someone can install** (`docs/releases.md`): publish when a person's experience
  changed, `minor` when a documented workflow step or a stored shape moves.
- **A skill of this project ships complete and stands alone.**
- **The installer refuses bytes it cannot verify**, and now the trust ends in the signature.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The review gate lives in the clone** (ADR 0042); **ownership is declared** (ADR 0038).
- **Every irreversible action names its target and waits for the human.**
- **A warning earns no review round**: it becomes a row in `docs/review-debt.md`.

## Queued

- **install cosign and run v0.12.0's installer on this machine**, so the binary in use is the signed
  release and the new check is exercised for real.
- **the two rows, paid**: the Windows refusal without cosign, and the fence header.
- **the release body comes from the changelog**, not from memory.
- **a change resumable from the diff needs no document**: `shape` it before building it.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has named;
the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says which class
of file grew back.
