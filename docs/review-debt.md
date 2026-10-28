# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.git/odachi/reviews/`) is evidence about one tree and lives in the clone,
outside the worktree, so it does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after 0078, the store leaves the project.

## Open

### cmd/odachi/files.go — `agentSelection.Target` cannot tell "no repository" from "git failed"

`Target` and `Written` swallow the error from `store.Root` and answer "the
project never chose". When git cannot name the common directory (not installed,
corrupt repo, permissions), the verbs that read the selection fall back to `both`
and refresh the agent files of the client the person excluded, with no line
saying the store was unreadable; `init` then writes with an empty path and the
message names an empty file. Named by two reviews, on trees 4fb1049 and the one
before it. The fix is a sentinel: an error the caller can tell apart from the
empty answer.

### internal/app/review/review.go — `storePath` joins an empty store

`storePath` joins `r.opts.Store` with `rel` without checking for the empty string.
A folder that is not a git repository has no store, and the path becomes relative
to the process directory instead of being refused. `check` and `progress` treat
an empty store as "no verdict on file"; `review` should refuse the same way.

### internal/adapters/gitx — `TreeOf`, `Has` and `CommitWithTree` have no adapter test

The ledger depends on them: `Has` maps `git merge-base --is-ancestor` exit 1 to
"no", and `CommitWithTree` keeps the *oldest* matching commit. The app tests drive
both through fakes, so a change that keeps the newest match would pass every test
and name the wrong commit for a step. One test against a real repository pins the
resolution.

### odd/tasks/0073-a-professional-persona.md — the doc speaks the old vocabulary

Its Done-when and Steps describe `odachi persona new`, `persona-craft` and
`.odachi/personas/<slug>.md`; the shipped tree has `odachi role`, `role-craft` and
`odd/roles/<slug>.md`. The Requirements are verbatim and untouchable, and the
fence says an old task doc is not rewritten; the divergence is recorded here so
the next reader does not follow a command that does not exist.

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
