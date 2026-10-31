# Next session: nothing is owed, and where the next work comes from

Written at the close of 0138, on v0.11.79. The chat is gone tomorrow; this file is
the state.

## What the last four features shipped

- **0135 (v0.11.76)**: the two rows the review of 0134 left, and this repository's
  `odd/review.toml` declaring `schema = 3`, `keep_verdicts = 10`, `keep_prompts = 0`
  -the numbers it already obeyed, now that the binary that reads them is released-.
- **0136 (v0.11.77, ADR 0041)**: a candidate whose prompt does not fit
  `max_prompt_kb` is sent as parts that do, one reviewer per part with the group's
  lenses and model, and the store merges the answers into the tree's one verdict.
  The only refusal left is a file that cannot fit by itself.
- **0137 (v0.11.78)**: `storeIn` takes the lenses this run has already landed and
  judges the collision per lens, so a part of the same run adds its rows and a second
  run over the same oversized candidate starts no reviewer.
- **0138 (v0.11.79)**: the merge decision is one predicate, `collision`, which returns
  the first lens an answer shares with the stored verdict that this run did not land
  itself; `storeIn` asks it and the refusal names the lens already paid for. A table
  test pins six answers, and putting the candidate's shape back now fails two tests.

## What is open

- **`docs/review-debt.md` has no rows.** The first time since 0124. Every row the
  reviews of 0136 and 0137 left is closed with its reason.
- No open questions in `odd/tasks`, and every feature is closed.

## What is measured, so nobody re-measures it

- The clone after `odachi clean --apply` on 2026-10-31: 601 files and 1.5 MB, minus
  27 files and 598 KB the run removed (4 prompts, 3 verdicts past the tenth tree, 10
  undo backups, 10 notes of scratch). The store keeps what `odd/review.toml` declares.
- **`opencode run` as `--exec` stalls to the 600s timeout on this machine**, twice, at
  59 KB and at 29 KB. The review path that works: `odachi review --base <sha>` to
  print the prompt into `.git/odachi/scratch/`, a subagent answers it, and
  `odachi review --base <sha> --attach <answer.json>` files the verdict. The verdict
  names the model that judged; `check` compares it with the proof's model.
- `mimo-v2.6-flash` closes 59 KB in about a minute and does **not** close 114 KB in
  600s; `odd/review.toml` carries both.
- `check --run` needs `--timeout 400` for this suite on this machine (it passes at
  120s but with no room).

## What is already true, so it is not re-decided

- **Ownership is declared, never inferred** (ADR 0038).
- **A failed read is not an absence.** `clean --apply` refuses before the first file
  when a class it maintains cannot be read; the review refuses a store it cannot read.
- **The bound lives in `potentRows`** (ADR 0039): a row travels when its path is
  judged or when it still blocks.
- **The fence in force protects** `internal/domain/proof/**`, `proof/**` and
  `skills/**`. A feature that needs one of those lands the policy correction first.
- **A relaxed key stays at the schema the released binary reads.** CI builds the gate
  from the base ref.
- **The changelog goes back to `## Unreleased`**, the release PR renames it to
  `## vX - date`, and the tag publishes. 0137 wrote a versioned heading and v0.11.78
  needed no release commit; do not repeat that.
- **Numbers follow the landing order**, and `check` reads the highest task document as
  the current feature.
- **Every irreversible action names its target and waits for the human.**

## The one decision that was waiting, and 0139 took

This repository kept the craft twice: `.agents/skills` (15 skills, 15 tracked files,
`bootstrap` deliberately absent) and `~/.agents/skills` (56, `bootstrap` included). The
block named the project's copy because the copy existed, and `setup` reported the
machine's because that is what it maintains: nothing contradicted, but one of the two
was refreshed by hand.

0139 deleted the copy and `.opencode/opencode.json`, and the block now names
`~/.agents/skills/`, the copy the machine keeps. The doctor learned the fact it was
missing -the scope `init` resolves- so a block that names the other folder is a warning
naming both folders and `odachi init`; the state this repository was in cannot come back
without one.

The rule that stays: **a repository that already lives and carries no copy keeps the
craft in the machine**; one that chose the project's copy keeps it, and the store's
choice wins over the folder's presence.

## How to find the next real work

`docs/review-debt.md` holds the accepted rows (none today); `odachi doctor` and
`odachi bench` find friction nobody has named; the hygiene trio (`go vet`,
`go mod tidy -diff`, `gofmt -l .`) is cheap; and `odachi clean` says which class grew
back, which is a lifetime that does not fit the way the project works.

Measured on 2026-10-31 after the release: `doctor` 13/13, `bench` 8 journeys and 20
steps with no dead ends, and the hygiene trio silent.
