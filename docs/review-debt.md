# Review debt

What the review store still holds potent about bytes that have not changed. The
store (`.git/odachi/reviews/`) is evidence about one tree and lives in the clone,
outside the worktree, so it does not travel with the repository. This file does.

- **Written at close**, from the rows the last verdict carried.
- **A row leaves** when the bytes it names change or the fix lands with its test.
- **It is not a gate**: `check` does not read it, and a row here never blocks a
  merge. It is the list a person reads before choosing the next feature.

Last written: after 0098, which closed the scope row.


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

Measured again on 2026-10-29, against this project's own store instead of one
afternoon's experiments: `ls .git/odachi/reviews/*.prompt.md` holds 34 prompts, the
median is 71.5 KB, the biggest is 125,759 B, **11 run within 10 KB of the limit**
(≥ 88 KB) and **3 are over it** - all three from before v0.11.44 shipped
`max_prompt_kb`, so the project can no longer collect evidence above 96 KB.

The three that are over the limit are the interesting part: 122,672 B and 123,679 B
left no verdict, and 125,759 B has one. Same size, opposite outcomes, so the wall is
not a line a size crosses; it is a probability a run meets. That is worth knowing
before the choice the row leaves to the human: raising `max_prompt_kb` would buy
attempts, not certainty, and splitting by paths is the shape that removes the bet.

What is left is a decision, not a measurement: this project's prompts already run
within 10 KB of its own limit, so the next large feature meets the wall. The two ways
out cost different things - a bigger limit risks a stall the tool cannot undo, and
splitting by paths costs one more prompt per piece and a merge of their answers - and
only the human knows which price this project should pay.

Measured again on 2026-10-29, with the cut of 0107 in place and the binary that carries it
(v0.11.58): the prompts it prints today for the last four candidates of this project are
**52,938, 53,014, 35,046 and 70,397 B**, a median of 53,014 B with **7.4 to 8 KB of carried
memory and 10 to 20 rows**. Before the cut the same comparison was a median of 71,489 B
with 41,388 B of memory and 632 rows. The memory went from 58 % of the prompt to about
14 %, and the largest prompt is now the one whose diff is largest (0103), not the one with
the most history.

So the wall is no longer a memory wall: it is a candidate-size wall, and this project's
candidates run at 35 to 70 KB with 26 KB of headroom. The two ways out the row named stay
available and stay the human's - a bigger `max_prompt_kb` or splitting by paths - but
neither is pressing, and the next time one is needed it deserves a fresh row with the
candidate that crosses the line, not this one.

### the review pipeline - the configured reviewer missed a test it was shown

On 0087 the reviewer reported "the CLI's propagation of the size limit has no test"
twice, the second time about a tree whose prompt carried
`TestTheReviewRefusesWhatTheProjectWillNotPayFor` at line 139 and whose diff added
it. The finding was approved as a WARNING and verified false by reading the prompt
it judged. At 74 KB with 21 changed paths, a reviewer is not reading every line it
is handed; a tool that trusts its silence as coverage (ADR 0033) should know that a
reviewer can miss a row it declared, and the next feature that touches the review
pipeline should measure this instead of assuming the answer is no.

Measured on 2026-10-29 on a prompt this store already held: tree 56641e43's, 94,678 B,
eleven changed paths, the same model the project configures
(`opencode-go/mimo-v2.6-flash`), asked through `opencode run --model ... --file
<prompt> "..."` with no agent override and three factual questions any reader of the
prompt can answer:

| probe | what it asked | answer | time |
| --- | --- | --- | --- |
| 1 | how many paths under "## The changed paths", and which ones the diff creates | **correct**: 11, and all five new files | 13 s |
| 2 | every test function the diff adds in `internal/app/review/scope_test.go` | **correct**: the five tests and the one helper | 11 s |
| 3 | what the prompt says about `inspection` and where it must appear | **correct**: the field, `completed`/`unavailable`, before `findings` | 9 s |

Three out of three, at 94.7 KB and in seconds, so this model does read what it is
handed at the size where 0087's miss happened (74 KB). The miss was in the reasoning:
the reviewer concluded a test was absent from a prompt that carried it, which is a
different failure from not reading - and the difference decides what the tool should do
about it (nothing about size).

What that leaves: the tool cannot promise a reviewer's reasoning, only its reading and
its declaration. The declaration and the coverage check stay what they are - a reader
that names a path it left out is refused - and the promise "silence is coverage" is now
a promise about reading, which is measured, not about judgment, which is not. A future
miss of the same shape is a data point about the model, not about the prompt's size.

## Closed by feature 0107

### the review prompt - the memory is bigger than the question

The prompt this project pays for is mostly what the store already carries. Measured on
2026-10-29 over the 37 prompts in `.git/odachi/reviews`: total 2,882,450 B, of which
1,553,334 B (54 %) is the section "What the store already carries for these exact
bytes"; the median prompt is 71,489 B with 41,388 B of memory in it (58 %), the worst is
83 %, and the biggest carries 44 KB. That section is **23,498 rows**: 168 findings and
23,330 "audited clean" passes, one per (path, lens).

The rows are not about the candidate. **23,269 of the 23,330 passes (99.7 %) name a path
the candidate never touched** - they are the repo's memory, carried into a review of a
few hundred lines. That is why eleven of these prompts run within 10 KB of the 96 KB
limit while their diffs are small: the wall the row above measures is pressed by memory,
not by the change.

What three ways of pruning would save, measured on the same 37 prompts:

| pruning | rows kept of 23,498 | median prompt | what it takes away |
| --- | --- | --- | --- |
| only the candidate's paths | 64 (0.3 %) | ~30 KB | the reviewer no longer sees that an unchanged path already carries a known defect, so it can only re-raise it as a follow-up (`causal: false`), never as a new causal finding |
| findings always, passes only for the candidate's paths | 232 (1 %) | ~45 KB | keeps every known defect and drops only the "audited clean" memory of paths nobody touched: the reviewer may re-audit a clean path, and nothing it can cite is lost |
| findings without passes | 168 (0.7 %) | ~44 KB | the same as the row above plus the clean rows of the paths that changed, which are the ones this run could have used |
| a byte budget (8 KB of rows) | 2,878 (12 %) | ~38 KB | an arbitrary cut: the rows are ordered by path, so the reviewer gets the alphabet's first eighth and is not told |

The four do not cost the same in what they drop. The candidate's paths keep only what the
review can act on; the two findings-first variants keep the shape of the verdict (what is
known) and differ in whether the clean rows of the *changed* paths survive, which are the
ones this run could have used; the budget keeps neither and hides what it dropped, which
is why it is the only pruning with no argument behind it beyond a number.

What the numbers say about the row above: the two ways out it names -a bigger
`max_prompt_kb` or splitting a candidate by paths- are the wrong levers while 99 % of the
prompt is memory. Splitting would pay one prompt per piece to fit a prompt whose size the
pieces barely touch, and a bigger limit would buy attempts against a wall this measurement
can move. The cheapest honest change is the hybrid: keep every finding, keep the clean rows
of the paths that changed, drop the rest.

What it does not prove: that a reviewer behaves the same with less memory. That is a
before-and-after on a real candidate, with the probe method 0101 used, and it belongs to
the feature that prunes - not to this row. The decision stays the human's, and it now has
the numbers it was missing.

Measured on 2026-10-29, after the human chose the hybrid: tree 56641e43's prompt cut by
the rule (every finding, plus the clean rows of the eleven changed paths) goes from
**94,678 B to 56,915 B (40 % less)**, keeping 3 rows of 645. The same model was asked
the three questions 0101 asked before, by the same plain `opencode run --file`:

| probe | full prompt (0101) | pruned prompt | time |
| --- | --- | --- | --- |
| 1. changed paths and new files | 11 and all 5 - correct | **one miss in two runs** (4 of 5 files), then correct | 12 s / 5 s |
| 2. every test added in `scope_test.go` | correct, 5 tests and the helper | correct, the same six names | 8 s |
| 3. what the prompt says about `inspection` | correct | correct, and sharper: "a top-level field, a sibling of findings" | 7 s |

A control run of the full prompt on the same question answered correctly (9 s), and the
miss did not repeat with the same pruned prompt, so it is the variance this model has
shown at full size - 0087's miss was the same shape - not the cut. What this shows is
narrow and worth writing exactly: a 40 % cut did not make the prompt unreadable to this
model; it does not prove that two arms are equal on four runs.

None of them costs the gate anything, and that is the fact that makes any of them possible:
the prompt is the reviewer's context, while the verdict's inheritance and the coverage
check read the store. `inherit` and `--show` do not consult the prompt, so pruning it
changes what the reviewer reads and nothing about what the tool proves.

The cut ships here: promptFor keeps every finding and filters the clean rows to the candidate's paths, while potentRows still reads the store whole and the verdict still inherits every row.

## Closed by feature 0103

### the gate - a committed change after the close is invisible to a local check

Found by the smoke test of 0102: commit code after a close, run `odachi check` with its
default base, and the gate answered ready - the candidate it measures is the worktree
against HEAD, which is empty once the change is committed. CI caught it; the lie was
local, which is the worst kind for this method.

The gate now anchors on the verdict that judged the close: a feature's verdicts share the
document's base, so the newest one names the furthest bytes the review looked at, and
`Between(that tree, the tree that stands)` says what arrived after. Anything there that
is not a document is a change nobody reviewed, committed or not. Nothing guesses the
close commit; the store wrote the answer.

A close with no verdict in the store keeps the answer it gave before, `--require-review`
keeps its meaning, and the release note after a paid close stays paperwork.

## Closed by feature 0102

### the gate - a feature that writes only documents cannot close

0096 gave the proof a twin so paperwork after a close owes no suite; the review had
none, and 0101 hit the hole: a docs-only feature whose closing tree was asked for a
verdict that `review` refuses to make ("this change asks for no lens"), with `--lens
all` as the only way out.

`theCloseOwes` now asks the feature's own question: the diff from the document's base to
the tree that stands, classified by the same function the review uses. A diff that asks
for no lens has nothing for a verdict to judge, so the close is ready - the two verbs
cannot disagree, because they ask one function.

What it does not change, named: code after the close still owes, an explicit
`--require-review` still forces the verdict, and a feature that carries code still owes
it at the document's base. It measures the feature's range and never the closing tree
alone: that tree only marks the last step, and asking it by itself would answer "no
code" about a feature that wrote plenty.

## Closed by feature 0098

### the review record · the scope a run judged is not written down

`--lens` widened what a run reviewed and nothing recorded that the human or the agent
asked for it. The row said the record belonged in the verdict; it could not land
there -the fence in force protects `internal/domain/review/review.go` under "no change
to the verdict shape", and CI reads the base fence, which is the wall 0096 hit- and a
new flag could not land either, because `cmd/odachi/main.go` is protected too.

So the record lives beside the verdict, the way 0091 put the human's dismissal there:
`reviews/<tree>.scope.json`, written by the run that added the lens, once, and read by
tree. It names the lenses the evidence asked for, the ones added, the flag as the
caller wrote it, and the caller's words and person when they gave them -the tool never
invents a reason. `--show` prints it and the prompt states it, because the lenses
section used to claim the evidence chose every lens and that sentence stopped being
true the moment somebody added one.

What it still cannot say: nothing, as of 0100. `--because` is required when `--lens`
adds a lens, and the words already on file count, so a raised scope either carries its
reason or does not happen.

The first version of the file had a defect 0099 paid: it was written before the
declaration and coverage refusals, so a refused `--attach` answer left a record beside
a verdict that never existed. The record now follows the verdict, which is the one
order gentle-ai allows -a refusal creates no authority and writes nothing.

## Closed by feature 0097

### the classifier · a Markdown the suite reads is called a document

The row 0096 opened, paid where it belongs. `review.ShippedText` is the one list of
what travels inside the binary or beside it -`skills/`, `internal/assets/` and
`.agents/`- and `isDocumentation` asks it: shipped text is code, so a skill or a
review prompt that changed asks for a lens, and a tree whose only change is one of
them is focused instead of passive. The proof lane deleted its copy and asks the same
function, so the two gates cannot drift again.

The installed copy stays plumbing for the lenses: 0043 decided that adopting odachi
must not cost a review before the first line of code, and a change under
`.agents/skills/` still asks for none. It owes the suite all the same, because the
pack test compares those bytes with their source - the reason the lane asks the list
at all, and the reason `TestAReleaseNoteIsPaperwork` was already watching it.

## Closed by feature 0096

### the proof · the tree that writes a release note owes the suite

The proof's twin of `theCloseOwes` landed: after a close, a tree whose paths that
moved since the proof ask for no lens -the same classifier the review uses- is
ready, and the report says which tree the proof belongs to and what moved. The lane
asks the close's own question -the document at HEAD already carried the close- so
the tree that closes a feature still owes the proof, a missing proof is never
paperwork, and a command this run watched fail still blocks.

The classifier half did not land here. It lives in `internal/domain/review/tier.go`,
and the fence in force when this feature started -the one at its base, which CI
reads on purpose- protects that domain under "no change to the verdict or the proof
format". The gate caught it on the pull request, and the change was reverted instead
of argued. The refusal the brief demands lives in the lane for now
(`internal/app/check/shippedText`): shipped text is never paperwork for this gate,
while the review's classifier still calls it a document, which is the open row below
and the reason that row exists. Closed by 0097, above: the list moved into the
classifier and the lane's copy is gone.

## Closed by feature 0095

### the gate · a plan commit that writes a real fence owes a review before it lands

The row 0094 opened. `featureReview` answered "not a feature" while the document
had no steps (`total == 0`), so the review gate fell through to `requireReview("")`
and asked the plan commit for a verdict -and the plan commit writes
`odd/fence.toml`, ordinary code on purpose. The fix is the rule 0093 had already
written three lines above for the proof: a document with a base and no steps is
open, the plan lands before the work. An open feature owes the verdict at its
close, the closing tree still owes it at the document's base, and a waived close
still owes nothing. `ReviewOwed` asks the same function, so `review --show` and
the gate cannot answer differently.

## Closed by feature 0094

### the review store · a new field makes an old binary call a good file broken

The verdict, `.git/odachi/agents.toml` and `odd/review.toml` say which schema wrote
them, and their decoders ask the version before naming a key. A value above the
binary's is refused as "written by a newer odachi: upgrade", never as `unknown keys`;
a file with no value reads as the first schema and keeps its bytes until something
rewrites it. One definition (`internal/domain/schema`) is asked by the three readers,
so the day a fourth file joins, absence means the past and not the present.

What it cannot do, named on purpose: a binary that predates the field still calls a
newer file broken, because it cannot read a key it has never heard of. The next field
this tool adds is the one that arrives as an upgrade — which is what every close from
now on gets for free.

## Closed by feature 0091

### the review store · a false warning is carried forever

The row named two ways out of a false warning, and the second one landed:
`review --dismiss <path>:<line> --because "<why>"` retires a row with a reason, bound
to the fingerprint of the bytes it judged, and the verdict is never rewritten. The
first way out —letting a refutation cite a file outside the diff— is still unbuilt
and no longer needed.

The rows themselves are the human's to retire, not this file's: three false ones were
still carried when 0094 was written (`internal/adapters/gitx/gitx.go:190`,
`internal/app/doctor/doctor.go:845`, `cmd/odachi/main.go:578`), each with its reason
already written in its task document.

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


