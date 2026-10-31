# Next session: the store stops copying, and the numbers become the project's

Written at the close of 0132, by the human's request: the plan and the method live
here, because the chat is gone tomorrow.

## What 0132 left, measured

`odachi clean` names every class a clone accumulates and removes only the four
odachi wrote, each with a declared lifetime (ADR 0038). Measured in this clone
before the first `--apply`: 14.7 MB of review records (240 files), 792 KB of undo
backups (751 files), 366 KB of proofs, and 82 notes an agent left loose in `.git/`,
now named by `--apply --legacy` and prevented by the scratch folder the block
teaches.

Retention bounds the **count**, not the size, and the size is the bigger problem:
60,246 pass rows are written across all verdicts while the newest verdict carries
789, and a verdict went from 0.7 KB to 125 KB in eight days. Every verdict copies
the whole inherited memory.

## F1 (next number) - the store without copies

A verdict stops writing the inherited rows. One memory keyed by the blob, referenced
by every verdict that judged it, turns the growth from O(n²) into O(n).

Acceptance, as it should be tested:

- a new verdict's size does not grow with the number of earlier verdicts in the
  store (make three reviews in a row over one file and compare the file sizes);
- the gate answers the same questions: the coverage arithmetic, the carried rows
  and the prompt still carry what 0085, 0104 and 0107 pinned, now read through the
  memory instead of parsed from each verdict;
- a store written by the previous release is read unchanged: the verdict's schema
  is the way out, and an old verdict must not need a migration to be believed.

Design care already decided elsewhere: the store leaves the project (0078), the
verdict's tree is its name (immutability by where it lives), and the scope record
is written only when the verdict it explains is written (0099).

## F2 (after F1) - the numbers belong to the project

`keep_verdicts` and `keep_prompts` in `odd/review.toml`, with today's numbers as
the default (10 records, no prompt kept by itself), so a project decides how much
evidence it pays for. `odd/review.toml` is a strict file with a schema counter
(0094): a new key is a release-bound change, and an older binary must say "upgrade"
instead of naming an unknown key.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038). `proof/`, `odd/tasks`, the
  client choice and the rest of `.git/` are named and never removed.
- **A failed read is not an absence.** `clean --apply` refuses before the first file
  when a class it maintains cannot be read; a date it cannot read keeps the file out
  of the class.
- **The fence at 0132's close protects** `internal/domain/proof/**`, `proof/**` and
  `skills/**`. A feature that needs one of those lands the policy correction first,
  in its own change (0120).
- **Numbers follow the landing order**, and `check` reads the highest task document
  as the current feature. If a correction lands ahead of the feature that needed it,
  renumber and move the recorded base.
- **Every irreversible action names its target and waits for the human**: push, pull
  request, merge, tag or release, an install, another folder.

## How to find the next real work

`docs/review-debt.md` holds the accepted rows; `odachi doctor` and `odachi bench`
still find friction nobody has named, and the hygiene trio (`go vet`, `go mod tidy`,
`gofmt -l .`) is cheap. Run `odachi clean` itself: a class that grew back is a
lifetime that does not fit the way the project works, which is a feature, not a
cleanup.
