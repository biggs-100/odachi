# Next session: the craft skill, two releases, and the loop that catches what rules do not

Written at the close of 0173, on **v0.11.102** -published, signed, attested and installed on this
machine. `doctor` is green except check 14 until this file was rewritten; `main` is clean and in sync
with `origin`.

## What shipped

- **0173 added `ui-craft`** to the curriculum: the decision order a designer follows, nine tells of a
  generic interface with the move that replaces each, the numbers that make a screen look deliberate
  (scale, measure, contrast, targets), the motion rules with their durations and easings, a read-back,
  and the rule that a skill of this project stands alone.
- **The skill was tested before it was published**: a landing for odachi, built with it and looked at
  four times. It is not in the tree: it lives in `.git/odachi/scratch/design-test/` as evidence. If it
  should exist, it is a design of its own under `design/<slug>/`.
- **v0.11.102 is that feature**, released by the doctrine: one bullet, the tag on the merge, the notes
  copied from the changelog, installed from the release.

## The two decisions the research settled

**A release is a tag someone can install.** Publish when what a person can do or trust changed -never
for tests, CI, docs or a refactor- with `minor` when the contract moves and `patch` otherwise while
`0.x`. The whole rule is `docs/releases.md`; do not re-decide it in chat.

**A skill ships complete.** `ui-craft` names no skill odachi does not carry: a user who installs the
binary gets every rule. Where depth does not fit, the answer is another small skill of this project,
never a pointer to something the machine happens to have.

## What the design test measured

- **The rules set the floor; the loop finds the rest.** Four rounds, four defects no rule would have
  caught: auto-placement dropped the terminal 300px below the headline, the terminal's dim text sat
  at ~3.5:1, control borders did not reach 3:1 against paper, and one empty grid row left a 90px hole.
- **A screenshot is one command away.** `msedge --headless=new --hide-scrollbars --window-size=1440,1400 --screenshot=<png> file:///<page>`
  works on this machine, with no dependency to install. Look at the image, not at the CSS.
- **System fonts are the ceiling of an artboard.** Type read correctly at 1.2 scale, 17/1.55 and a
  55-character measure, and still had no character: a real landing needs a chosen typeface, which the
  artboard rule (self-contained, no remote assets) does not allow.

## What was measured before, still true

- **The suite is ~300-430 s on this machine.** `check --run` wants `--timeout 1500`.
- **A killed command in Windows leaves children** (`go`, `shellcmd.test`) that hold the build cache;
  the next suite prints nothing for minutes and reads as a hang. Look for stray processes first.
- **The review runs in one command:**
  `odachi review --base <sha> --exec 'opencode run --agent odachi-reviewer --model {model} --file {prompt} "answer with the JSON only"'`.
- **A fence can be too wide for its own reason, and narrowing it needs a policy pull request.** CI
  reads the fence from the base ref; a light feature cannot rewrite it at all.
- **`gate (macos-latest)` sometimes cancels at 15m02s**; a re-run passes in ~1m20s. Runner flake.
- **The order that survives the hook:** commit the step, then `plan --done T1`.

## What is open

- `docs/review-debt.md` holds one row: the fence header explains a rule the file does not have.
- **The installer verifies sha256, not the signature.** `checksums.txt` is signed and attested; the
  script does not check the bundle with cosign.
- **The release body is written by hand** with `gh release edit` after the tag, because goreleaser
  generates notes from commits.
- **The change that needs no document** is still only half built: the light lane is the cheap half.
- **A proof that says what it is blocked on** was half paid by 0169: the stream exists, the child's
  name and the lock do not.

## What is already true, so it is not re-decided

- **Four artifacts are the memory**: brief, task doc, proof and profile.
- **The release doctrine lives in `docs/releases.md`** and `drive` names it in one line.
- **A version heading with no tag is not a release**, so it may stop existing.
- **A skill of this project ships complete and stands alone**; it may name a deeper skill as optional.
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
  before installing.
- **the release body comes from the changelog**, not from memory.
- **a change resumable from the diff needs no document.** Measure it before building it: `shape`
  first.
- **the design loop as a practice**: the four defects the screenshot found are the argument for a
  read-back step in every UI feature, and the artboard rule (system fonts only) is the thing to
  revisit if a landing ever needs a real typeface.

## How to find the next real work

`docs/review-debt.md` names what a verdict left; `doctor` and `bench` find friction nobody has named;
the hygiene trio (`go vet`, `go mod tidy -diff`, `gofmt -l .`) is cheap; and `clean` says which class
of file grew back.
