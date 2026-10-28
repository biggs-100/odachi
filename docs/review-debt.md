# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.git/odachi/reviews/`) is evidence about one tree and lives in the clone,
outside the worktree, so it does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after 0089, one row was verified false and three were closed.


## Open

### the review store - a new field makes an old binary call a good file broken

`verdicts`, `agents.toml` and `odd/review.toml` decode strictly on purpose: a typo
must be named instead of silently half-understood. What the strictness cannot tell is
a typo from a *newer writer*: a binary that predates a field reports `unknown keys` /
`unknown field` about a file that is perfectly well formed, and the message sends the
human to fix a file that is not broken. It cost three closes in one session
(`project_agents` in 0081, `inspection` in 0085, `max_prompt_kb` in 0087), each time
solved by building the tree locally and swapping the installed binary. gentle-ai
solves it from the other side: an older binary reports the authority as
`applicability: corrupted` with action `repair_authority`, and the repair section
says which release can continue the lineage. The odachi shape is a `schema` field
written by every new verdict, agent selection and review config, so a decoder that
meets an unknown value can say "this file was written by a newer odachi: upgrade or
use the binary that wrote it" instead of naming keys.

### the review record - the scope a run judged is not written down

`--lens` widens or narrows what a run reviews, and nothing records that the human (or
the agent) asked for it: the verdict stamps the lenses it answered, but not why this
candidate was judged by one lens and that one by four. gentle-ai freezes an
`agent_escalation` - an item from the shared high-risk list plus a reason - inside the
authority, so a later reader sees that the scope was raised deliberately and a replay
cannot drop it. The odachi shape is a line in the verdict: who widened the scope, to
which lenses and in whose words, beside the declaration that landed in 0085.

### the review store - a false warning is carried forever

A potent row leaves the store only when the bytes it names change, and a warning
cannot be refuted: `--refute` takes causal `BLOCKER` and `CRITICAL` rows, and
"warnings are never refutable". So a finding the reviewer got wrong about a file
nobody touches again is re-reported by every verdict that inherits it. Evidence: the
row `cmd/odachi/main.go:573 - the propagation of max_prompt_kb is pinned by no test`
was filed by 0087's review, is false (the test is
`cmd/odachi/feature_test.go` `TestTheReviewRefusesWhatTheProjectWillNotPayFor`, and
it fails when the line is removed), and has been carried through 0087, 0088 and
0089. Two ways out, both for a feature of their own: let a refutation cite a file
outside the diff for a non-blocking row, or let the human dismiss a row with a
reason the way a finding is dismissed.
### internal/app/review - `opencode run` as the `--exec` reviewer, measured

Measured on 2026-10-28: `opencode run --agent odachi-reviewer --model {model}
--file <prompt> "..."` with a **tiny** file answers in seconds; 69 KB and 74 KB
close in 46 seconds and one run; 89 KB closes in 46 seconds; 0085's **122 KB**
prompt prints its banner and produces nothing until the 600s timeout, twice. So
the wall sits between 89 and 122 KB and it is not slowness: a reviewer that stalls
at a size does not answer late. 0087 turned that into `max_prompt_kb`, so the tool
now refuses before spending - what is left is capability, not safety: a candidate
whose prompt passes the limit cannot be reviewed by this model at all, and the
choice between a model that reads more and splitting the candidate by paths is the
human's. Nothing in the tool splits a prompt today.

### the review pipeline - the configured reviewer missed a test it was shown

On 0087 the reviewer reported "the CLI's propagation of the size limit has no test"
twice, the second time about a tree whose prompt carried
`TestTheReviewRefusesWhatTheProjectWillNotPayFor` at line 139 and whose diff added
it. The finding was approved as a WARNING and verified false by reading the prompt
it judged. At 74 KB with 21 changed paths, a reviewer is not reading every line it
is handed; a tool that trusts its silence as coverage (ADR 0033) should know that a
reviewer can miss a row it declared, and the next feature that touches the review
pipeline should measure this instead of assuming the answer is no.

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

## Closed by feature 0090

- `internal/app/start/artifacts.go` - `documentOpen`'s read error is reachable: the
  fake fails one path, and the test drives the block that lands after the brief and
  the fence are written.
- `internal/app/doctor/doctor.go` - the last branch of `personaNote` has its test:
  mirrors matching while the store records `project_agents = none`, asserting the
  store is named and the files are called leftovers.
- `odd/tasks/0073-a-professional-persona.md` - the Done-when keeps the words the
  human accepted and gains a note at the top mapping them to `odachi role new`,
  `odd/roles/` and `role-craft`. Changing the bytes also retired the reviewer's row.

## Closed by feature 0089

- `internal/adapters/gitx` - **the row was false**. `internal/adapters/gitx/ledger_test.go`
  already drives `TreeOf`, `Has` (exit 1 as the answer, an unresolvable name as the
  error) and `CommitWithTree` (the oldest match, through an empty commit that
  repeats the tree) against a real repository. The reviewer looked at
  `gitx_test.go` and missed that file; the row is closed as verified, not paid.
- `internal/app/role` - the target list is one definition: `targets(...)` returns the
  skill plus one persona file per client and the client that has none, and both
  `Render` and `matches` use it.
- `cmd/odachi` - pack's installed bytes are compared with their source: dropping the
  `skill.Parse` → `pack.Skill` join now fails
  `TestPackInstallsTheProjectSkillEndToEnd`.

## Closed by feature 0088

- `internal/app/review/review.go` - the answer file of a failed run: an answer left
  behind by a bad exit or a timeout is copied under a name its content decides,
  named in the refusal, and the scaffolding is removed, so the next run cannot
  empty it. A store that cannot file the prompt says so and keeps the answer.
- `internal/app/review/review.go` - `--exec` files the prompt it sends: once sent,
  under the canonical name for one group and one file per group for a split run.
  0087's false positive can now be checked against the bytes that were sent.

## Closed by feature 0086

### the review gate - a verdict that names no model is a review the gate cannot attribute

`check` now refuses a verdict that names no model for a required lens while the
proof names one, naming the lens, the verdict's file and the model the proof
carries. The way out is a decision on the record: re-review with
`--model <provider/model>`, or the human's `start --waive-review`, which the help
now lists. The prompt asks the answer to name its model, so an attached answer can
still be attributed.

## Closed by feature 0085

### the review gate - a complete set of rows cannot say who wrote them

An answer to a lens-bearing candidate with no `inspection` is refused, and the
declaration is compared with the changed paths in both directions before the
store lends it anything. A padded row is now a declared path or a named gap.
ADR 0033 says why, and what the gate still cannot prove (that the reader read).

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


