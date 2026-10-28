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

### internal/adapters/gitx - the ledger's reads run only through fakes

`TreeOf`, `Has` and `CommitWithTree` are the adapter functions the ledger depends
on: `Has` maps `git merge-base --is-ancestor` exit 1 to `(false, nil)` and any
other nonzero exit to an error, and `CommitWithTree` keeps the *oldest* matching
line to name the commit that introduced the bytes. No adapter-level test drives
them against a real repository - `gitx_test.go` covers only `HeadTree` - so the
app-layer tests exercise `fakeGit` answers instead of git's. A change that kept the
newest match instead of the oldest would name the wrong commit for a step whose
bytes returned later, and every test would stay green.

### internal/app/role - the target list is built twice in one package

`Render` assembles the skill plus one `persona.RoleFile` per client, and `matches`
assembles the identical list to decide stale versus up to date. Nothing couples the
two, so adding a target to one and not the other makes `odachi role` report a state
that is wrong, with the tests green because they drive both halves today. One
`targets(...)` helper makes the two answers one definition.

### cmd/odachi - pack's installed bytes are never compared with their source

The end-to-end contract is that a skill created with `odachi skill new` reaches
`.agents/skills` with its bytes, and the test checks membership and a stdout
substring only. The one path that feeds that file - `skill.Parse`'s `Raw` joined
into `pack.Skill{Content: s.Raw}` - is exercised by no test that compares content,
so dropping `Raw` would install an empty `SKILL.md` with every test green.

### internal/app/start/artifacts.go - an unreadable task document blocks the run, untested

`documentOpen` returns a blocked result when a document whose slug matches cannot
be read. Before 0083 `start` never read that file, so this is a new way for a run
to stop — after `brief.md`, `profile.toml` and `fence.toml` are already written, so
the retry needs `--force` — and the fake cannot inject a read error, so no test
reaches the branch. A read-error seam in `fakeFiles` and one assertion on the
message close it.

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
question. Same home, same name, written by both forms. 0087 paid for it: a
reviewer's finding could only be judged false by regenerating the prompt by hand
and assuming the tree had not moved since.

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


