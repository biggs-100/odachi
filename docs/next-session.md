# Next session: the file bump, and finding the next real work

Written at the close of 0134, by the human's request: the plan and the method live
here, because the chat is gone tomorrow.

## The three features the human planned are shipped

- **0132 `odachi clean`** (v0.11.73): every class of file a clone accumulates, with
  its owner, its size and what the run would do; `--apply` keeps a lifetime only on
  the classes odachi wrote, `--legacy` names the notes an older agent left loose in
  `.git/`. ADR 0038.
- **0133 the store without copies** (v0.11.74): a verdict carries the rows of the
  paths its candidate judged, plus a blocker that still stands, so it follows the
  candidate and not the store. 126 KB became 2.4 KB on the feature's own review.
  ADR 0039.
- **0134 the project decides** (v0.11.75): `keep_verdicts` and `keep_prompts` in
  `odd/review.toml`, prompts as their own class, the schema at 3. ADR 0040.

## What 0134 left open, and 0135 closed

This repository's `odd/review.toml` carries `schema = 3`, `keep_verdicts = 10` and
`keep_prompts = 0` since 0135, in a change of its own. The numbers are the ones it
already obeyed, so nothing moved but the file telling the truth. The two rows the
review of 0134 left are closed by the same feature and left `docs/review-debt.md`.

## What is measured, so nobody re-measures it

- The store after the first real `clean --apply` on 2026-10-31: 15.6 MB and 1004
  files became 2.2 MB and 557. Before it: 240 review records (14.7 MB), 751
  versions (792 KB), 82 notes in `.git/` and 366 KB of proofs.
- After 0133: a verdict is 2-16 KB instead of 126 KB. The ten old ones still weigh
  1.1 MB and will age out through `keep_verdicts`.
- Prompts: 597 KB in eleven files, the largest 114 KB. `keep_prompts = 0` empties
  them; the report says so before it does.
- `mimo-v2.6-flash` closes a 59 KB prompt in about a minute and does **not** close
  114 KB in 600s. `odd/review.toml` carries both measurements.

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038). `proof/`, `odd/tasks`, the
  client choice and the rest of `.git/` are named and never removed.
- **A failed read is not an absence.** `clean --apply` refuses before the first file
  when a class it maintains cannot be read; the review refuses a store it cannot
  read instead of borrowing the old answer.
- **The bound lives in `potentRows`** (ADR 0039). A future lens that asks about an
  untouched path moves the rule with it.
- **The fence in force protects** `internal/domain/proof/**`, `proof/**` and
  `skills/**`. A feature that needs one of those lands the policy correction first,
  in its own change (0120).
- **Numbers follow the landing order**, and `check` reads the highest task document
  as the current feature.
- **Every irreversible action names its target and waits for the human.**

## How to find the next real work

`docs/review-debt.md` holds the accepted rows; `odachi doctor` and `odachi bench`
still find friction nobody has named; the hygiene trio (`go vet`, `go mod tidy`,
`gofmt -l .`) is cheap. And run `odachi clean` itself: a class that grew back is a
lifetime that does not fit the way the project works, which is a feature and not a
cleanup.
