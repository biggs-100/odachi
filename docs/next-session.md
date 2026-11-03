# Next session: a release that broke, and the patch that proved the chain

Written at the close of 0175. **v0.12.1 is published**, signed and attested, and this machine runs
**v0.12.0** -the same binary, installed by the fixed script against the real release. `doctor` is
green except check 14 until this file was rewritten; `main` is clean and in sync with `origin`.

## What shipped

- **0175 fixed what v0.12.0 broke.** The PowerShell installer aborted exactly when the signature
  verified: cosign prints `Verified OK` on stderr, and PowerShell 5.1 with `$ErrorActionPreference
  'Stop'` turns a native program's stderr into a terminating error. Nobody with cosign could install
  the release. The preference is now relaxed for the call and restored; the exit code is the answer.
- **v0.12.1 is the patch**, and its notes say what happened. The install that proved it ran against
  the real release: cosign verified, sha256 passed, the binary landed.

## The lesson this pair of releases paid for

**A test double that does not behave like the original proves nothing.** The shim answered "success"
without writing a byte, so every test passed while the real installer failed on its own success
message. The shim is now a `.cmd` -a native program for PowerShell- that prints to stderr like cosign
does, and the red reproduces the production failure exactly.

The rule that follows, and belongs beside "observe red": when the double stands in for a tool that
talks, the double must talk on the same channel.

## What was measured, so nobody re-measures it

- **The suite is ~410 s** with the installer tests; `check --run` wants `--timeout 1500`.
- **`opencode run` stalls at ~40 KB on the installer diff**; a subagent that reads the prompt file
  answers in seconds. Use that when the one-command review stalls.
- **`install.ps1` is UTF-8 with BOM on purpose.** Windows PowerShell 5.1 decodes a BOM-less script
  with the ANSI code page and every accent becomes mojibake; ASCII-only assertions do not notice.
- **The pre-commit hook refuses a commit whose proof is stale:** stage the correction, `check --run`
  on that tree, then commit.
- **cosign is installed on this machine** (`~/go/bin/cosign.exe`, on PATH) and v0.12.0 verified
  against its real bundle before installing.
- **The agent's shell is pwsh 7.6.9, but a script that ships to Windows is proven on `powershell`
  5.1**, which is what a user has: in 7 a native program's stderr is not a terminating error, so the
  same installer passes there and fails on the machine that matters. And file rewrites belong to the
  editor tool, never to `-replace` in a shell: a global replace ate the changelog header once.

## What is open

- `docs/review-debt.md` holds two rows: the Windows branch that refuses without cosign has no test,
  and the fence header quotes a decision the brief never took.
- **The release body is still written by hand** with `gh release edit` after the tag. Two releases
  today, twice by hand.
- **The change that needs no document** is still only half built.
- **ADR 0042**: CI cannot enforce the review. A decision, not work.
- The installed binary says `0.12.0` while the newest release is `0.12.1`; the bytes are the same,
  and reinstalling is one command if the number matters.

## What is already true, so it is not re-decided

- **A release is a tag someone can install** (`docs/releases.md`); a fix someone suffered is a patch,
  published as soon as it is green.
- **A skill of this project ships complete and stands alone.**
- **The installer refuses bytes it cannot verify**, and the trust ends in the signature.
- **A test double must behave like what it replaces**, or the test measures the double.
- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The review gate lives in the clone** (ADR 0042); **ownership is declared** (ADR 0038).
- **Every irreversible action names its target and waits for the human.**
- **A warning earns no review round**: it becomes a row in `docs/review-debt.md`.

## Queued

- **the two rows, paid**: the Windows refusal without cosign, and the fence header.
- **the release body comes from the changelog**, not from memory: a step of the workflow, so nobody
  writes it twice by hand.
- **a change resumable from the diff needs no document**: `shape` it before building it.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has named;
the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says which class
of file grew back.
