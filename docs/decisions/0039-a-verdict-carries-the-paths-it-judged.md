# 0039 - A verdict carries the paths it judged

Status: accepted — Date: 2026-10 — Session: after v0.11.73

## Context

A verdict copies the store's potence into itself. `potentRows` weighs every other
verdict against the working tree -it keeps each row whose blob still equals the
blob git measures now- and `inherit` merges all of them into the verdict being
written. The gate then reads one file, which is why the copy existed: the answer
for every row had to travel with the tree it was given for.

The bill, measured on 0132's verdict: **126.3 KB, 769 clean rows, and 15 of them
name a file the candidate changed**. The other 754 describe files the diff never
touches. A verdict went from 0.7 KB to 125 KB in eight days, and the store grew
with the square of the reviews: the sixth day's verdict carries the first day's
rows.

The prompt had already paid this bill once: 0107 cuts the carried rows to the
changed paths, because 99.7 % of them were noise for a reviewer reading a diff.
The cut lived in the prompt only -`knownCleanRows`- and the verdict kept copying.

## Decision

- **The bound is the read's, not the prompt's.** `potentRows` takes the candidate's
  paths and keeps a clean row only when the diff touches its path; `inherit` takes
  the same paths from the worktree. One rule, one place, and the prompt's own
  filter disappears.
- **Before hashing.** A row about an untouched path is dropped before its file is
  hashed, so the review also stops paying one git call per row nobody asked about.
- **A finding is not that kind of row.** A causal BLOCKER or CRITICAL about a file
  the candidate never touches still travels while its bytes stand. It is not an
  answer about bytes somebody read; it is the reason not to ship, and dropping it
  would make the gate forget a defect nobody fixed. `Finding.Blocks` is now the one
  definition of that rule -the floor asks it too instead of spelling the same three
  conditions.
- **Old verdicts are never rewritten.** A store written by the previous release
  keeps its copied rows; the bound applies to what is written from now on, and
  `clean`'s retention is what retires the old ones.

## Alternatives rejected

- **A memory keyed by blob, referenced by every verdict.** The shape this feature
  was first planned with. It dedupes rows *across* verdicts, but it needs a new
  file format with its own schema, a reader in the gate, a rule for who merges a
  row first, and a lifetime for the folder -and it dedupes a set that the bound
  already keeps at the size of one candidate. The same measured goal, ten times the
  artifact.
- **Bounding findings the same way.** Ten of the eleven findings in 0132's verdict
  name untouched files; they are standing warnings a reviewer should not re-raise.
  Bounding them would have been simpler, and it would silently drop a standing
  blocker the day one exists.
- **Pruning the store harder.** Retention bounds the count of records, not the size
  of one: ten 126 KB verdicts are still 1.2 MB, and the next one would carry them.
- **Reading every verdict at gate time** (the store as the memory). Sound, and it
  moves the cost from the write to every check the pre-commit hook runs. The bound
  keeps the gate reading one file, which is what 0085 leaned on.

## Consequences

- A verdict's size follows its candidate: the same 0132 verdict, re-rendered with
  its bounded rows, is **about 4.8 KB** instead of 126.3 KB.
- Growth is linear in the number of reviews, and the store stops holding the same
  answer once per tree that outlived it.
- The gate's promise is unchanged and one file still answers it, because the gate
  only asks about the changed paths. A future lens that asks about an untouched
  path moves the bound with it -the rule lives in `potentRows` and nowhere else-.
- `review --dismiss` keeps working on carried rows: they are still in the tree's
  verdict when the candidate judged their path, which is the only case the command
  was ever about.
- A project that re-reviews the same file across many trees keeps one row per blob
  per lens in each verdict; the rows dedupe within the verdict as before.
