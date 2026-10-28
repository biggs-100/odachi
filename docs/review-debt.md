# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.git/odachi/reviews/`) is evidence about one tree and lives in the clone,
outside the worktree, so it does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after 0080, the sentinel is pinned end to end.

## Open

### internal/app/review/review.go - the answer file of a failed run is left unnamed

A command that names `{verdict}` gets `answers/<tree>.json`, emptied before each
group, and the success path removes it. The failure returns — bad exit, timeout, a
refutation, a store refusal — leave it behind without naming it, and the next run
on the same tree empties it: an answer somebody paid for disappears silently. The
fix is the one the success path already applies: name the file, or keep it under a
name that says it failed.

### internal/app/review/review.go - --exec does not file the prompt it sends

The printed form writes `reviews/<tree>.prompt.md`. A run that hands the prompt to
a reviewer does not, so the bytes that produced a verdict cannot be read again
without regenerating them, and regenerating needs the same base to build the same
question. Same home, same name, written by both forms.

### the review gate - a verdict with rows missing is accepted when someone pads it

`review --attach` checks one row per (path, lens) and accepts what it is given: it
cannot tell a complete pass list from one a caller completed. In 0082 a reviewer
returned 49 of 52 rows and the caller wrote the other three to get through. The
gate should refuse an incomplete verdict and ask for it again, and the answer
should say which rows were missing - the same idea as gentle-ai's typed
`inspection.paths`, which asserts what was actually read.

### internal/app/doctor/doctor.go - the last branch of personaNote has no test

The branch that answers "the files match and the store records `project_agents =
none`" ships with no test: it is the state right after `init --global
--no-project-agents` over a project whose mirrors were current. One test with
`configured()` intact and a store that records `none`, asserting the store is
named and the "matches" wording is replaced, closes it. Owed by 0081; the next
feature that touches `internal/app/doctor` pays it, the way 0079 paid 0078's.

### odd/tasks/0073-a-professional-persona.md - the doc speaks the old vocabulary

Its Done-when and Steps describe `odachi persona new`, `persona-craft` and
`.odachi/personas/<slug>.md`; the shipped tree has `odachi role`, `role-craft` and
`odd/roles/<slug>.md`. The Requirements are verbatim and untouchable, and the
fence says an old task doc is not rewritten; the divergence is recorded here so
the next reader does not follow a command that does not exist.

## Accepted

### the store · a verdict attached with the wrong base cannot be replaced

`review --attach` seals the base from `--base`; without it the verdict stores
`HEAD` and `check` refuses it for a feature that started elsewhere. The tree then
already has a verdict, so a second attach is refused too, and no verb deletes a
verdict: the only way out is removing the file in the clone's store by hand,
which 0081 did once. Accepted for now: the message names the exact `--base` to
use, so the mistake is loud. A `review --forget <tree>` is a new verb, and 0081's
fence forbade one.

### internal/app/init/init.go · a run with no store adapter records nothing

`writeAgentStore` returns without a line when the caller wired no selection port,
so an explicit `--agents` or `--no-project-agents` from such a caller exits 0 and
records nothing. The binary always wires the adapter, so no user path regresses;
the tests that only care about the project files rely on it. Accepted: a named
refusal is better and costs two lines, but it changes a test-only path and this
feature had its verdict.

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

## Closed by feature 0063

### internal/app/check/check.go · the waiver branch of the exported ReviewOwed

`TestReviewOwedAnswersTheWaiver` calls the exported function with a closed
document that records `Review: waived`, and with the same document without the
line: the branch answers "nothing owed" in the first case and owed at the base in
the second, on the same tree and the same git. Observed red by making the branch
answer owed, so the test is not a comment.

## Closed by feature 0057

- `internal/app/tasks/tasks.go:124` - the truncation has a test with five names
  and the limit itself.
- `internal/app/doctor/doctor.go:635` - a listing that fails is a problem, not the
  same silence as a folder with no foreign documents.
- `internal/app/review/review.go:1350` - the drop of a refused answer cannot cost
  the verdict; the keep-failure half was already pinned by feature 0051, which the
  store could not know because a test moves a different file's blob.


