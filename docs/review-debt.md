# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.odachi/reviews/`) is evidence about one tree and is ignored by git, so it
does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: v0.11.28, closing feature 0056.

## Open

### WARNING · internal/domain/proof/proof.go:137 · reliability

An older proof -written before the `# output` terminator existed- whose command
output begins with the literal line `# output` is parsed as if that line were the
terminator: the line silently disappears from the recorded output. The format
cannot tell the two apart by itself, so the fix is a decision, not a patch: keep
the ambiguity and say so, or version the header.

### WARNING · internal/app/review/review.go:1350 · reliability

The keep-failure path of `refuse` and the two error branches of `keepRefused` have
no test that drives them, and `dropRefused`'s early returns are unreachable with
the current fake. If the guard were lost, a refusal would name an
`--attach <file>` that was never written -exactly the second payment feature 0050
removed- while every test stays green.

### SUGGESTION · internal/app/tasks/tasks.go:124 · reliability

`fewNames` truncates a long list to the first three plus `and N more`, and no test
feeds it more than two names: an off-by-one in the bound or the count would print a
wrong line in `doctor`, `start` and `summary` with the suite green.

### SUGGESTION · internal/app/doctor/doctor.go:635 · reliability

`taskDocs` returns `Broken` when `tasks.Skipped` fails, and no test makes `List`
fail: the repair text of that branch is unpinned.
