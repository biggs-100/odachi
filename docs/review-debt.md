# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.odachi/reviews/`) is evidence about one tree and is ignored by git, so it
does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after v0.11.30, closing feature 0062.

## Open

### internal/app/check/check.go · the waiver branch of the exported ReviewOwed

`ReviewOwed` answers "nothing owed" for a closed feature whose task doc records
`Review: waived` (check.go:723-730), and no test calls it there: the closed-waiver
test drives `reviewGate`, the report test injects the answer, and the taskdoc test
stops at the parser. If the branch were lost or inverted, every test would stay
green and a waived feature would owe a verdict again. The fix is a test that calls
`check.ReviewOwed` with a closed waived document, and with the same document
without the waiver.

## Accepted

### internal/domain/proof/proof.go · the terminator in an old proof's first output line

An old proof -written before 2026-10-25, when the `# output` terminator arrived
(50cb850)- whose command output starts with the literal line `# output` loses that
line when parsed: with no terminator, the format cannot tell the header from the
command's own line.

Accepted, not fixed. A proof this version writes always carries the terminator, so
the command's line is the second one and survives; a test pins the round trip. The
gate compares the tree, never the text, so no proof is accepted or refused because
of this. The files it can still reach were written by an older binary and not
regenerated since, and the next `check --run` overwrites them. A length-prefixed
header would remove the ambiguity and detect a truncated proof as well; it is a
format change for a case this narrow, and it can wait until the format changes for
another reason.

## Closed by feature 0057

- `internal/app/tasks/tasks.go:124` - the truncation has a test with five names
  and the limit itself.
- `internal/app/doctor/doctor.go:635` - a listing that fails is a problem, not the
  same silence as a folder with no foreign documents.
- `internal/app/review/review.go:1350` - the drop of a refused answer cannot cost
  the verdict; the keep-failure half was already pinned by feature 0051, which the
  store could not know because a test moves a different file's blob.
