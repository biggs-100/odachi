# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.odachi/reviews/`) is evidence about one tree and is ignored by git, so it
does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after v0.11.28, closing feature 0057.

## Open

### WARNING · internal/domain/proof/proof.go:137 · reliability

An older proof -written before the `# output` terminator existed- whose command
output begins with the literal line `# output` is parsed as if that line were the
terminator: the line silently disappears from the recorded output. The format
cannot tell the two apart by itself, so the fix is a decision, not a patch: keep
the ambiguity and say so, or version the header.

## Closed by feature 0057

- `internal/app/tasks/tasks.go:124` - the truncation has a test with five names
  and the limit itself.
- `internal/app/doctor/doctor.go:635` - a listing that fails is a problem, not the
  same silence as a folder with no foreign documents.
- `internal/app/review/review.go:1350` - the drop of a refused answer cannot cost
  the verdict; the keep-failure half was already pinned by feature 0051, which the
  store could not know because a test moves a different file's blob.
