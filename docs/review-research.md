# Review under pressure: what the field learned, and what our review does with it

Written 2026-11-10, after a session that shipped 0262; updated 2026-11-12 with the
first field reading of the draws (Part 5). The question: code is now
almost free to write, the queue is review, so are our four lenses and our review
system answering the problem the field actually has, or a problem we imagined?

Everything below is either a published measurement with its sample size, or a
number read from this clone. Both are named, because a comparison that mixes them
is a slogan.

## Method

- External evidence: six 2026 sources, listed at the end with dates and samples.
  Where a source sells a product, its own numbers are treated as a claim about the
  field and not as a measurement of its product.
- Internal evidence: the review store at `.git/odachi/reviews/` (16 verdicts
  currently kept), `odd/review.toml`, the four lens prompts in
  `internal/assets/review/`, the floor in `internal/domain/review/floor.go`, and
  the failure record in `docs/review-debt.md` and `docs/next-session.md`.

## Part 1: what the field measured

1. **The bottleneck moved, and the queue grew.** Chen and Stratton (Harvard,
   working paper, Aug 2026; 718 firms, 68.3M Jira issues, 200.9M commits, 49.1M
   pull requests) found agent adoption added ~30% more lines and 23% more pull
   requests per worker-month, while completed issues did not move (0.12 on a 3.67
   baseline, interval excluding more than +12%). Time from PR to merge rose 3.45
   days on a 7.03-day baseline, **+49%**. Change requests per PR nearly doubled
   (13% to ~25%) and comments per PR rose 35%.
2. **AI review tools did not remove it.** In the same dataset, ~80% of firms had
   adopted an AI reviewer by March 2026, yet AI wrote only 23.3% of review
   comments and **10.8% of pull requests received a single AI comment**. Adoption
   is not coverage.
3. **Trust and verification are decoupled.** Sonar's *State of Code* (Jan 2026,
   1,100+ developers): AI is 42% of committed code, expected at 65% by 2027; 96%
   do not fully trust it is correct, but only **48% always verify before
   committing**; 38% say reviewing AI code takes more effort than reviewing a
   colleague's; 53% report code that looks correct but is not reliable. That gap
   has a name now: verification debt.
4. **Throughput and instability rise together.** DORA (2025 report; 2026
   qualitative follow-up on 1,110 Google engineers): higher AI adoption is
   associated with an increase in both delivery throughput and delivery
   instability, and the time saved writing is re-spent auditing. 30% of
   developers report little or no trust in generated code.
5. **AI review comments are themselves unverified output.** Zhong, Noei, Zou and
   Adams (arXiv 2603.15911, 278,790 review conversations across 300 open-source
   projects): AI suggestions are adopted at a significantly lower rate than human
   ones, **over half of unadopted AI suggestions were either wrong or solved
   differently**, adopted AI suggestions increase complexity and code size more,
   and humans exchange 11.8% more rounds when reviewing AI-written code. Humans
   supply the categories agents do not: understanding, testing, knowledge
   transfer.
6. **Determinism is the scaling answer, and it is cheap.** Morgan Stanley and
   Moderne (Code Remix Summit 2026): the constraint is review capacity, the
   proposed fix is risk-based routing (low-risk changes on a fast path, deep
   review reserved for the rest) and deterministic tools doing the execution
   while AI does the planning. Separately, JetBrains reports that **20-25% of AI
   hallucinations are detectable by structural and static analysis** — a class of
   defect that costs nothing once the analyzer runs.

METR's early-2025 RCT (16 experienced developers, 246 tasks) found AI made
experienced developers 19% slower while they believed they were 20% faster; its
2026 follow-up says the newer data suggests a speedup but is biased by selection
effects. The honest reading: workflow-level gains are real for some tasks and
unmeasured for many, while the review queue grows in every dataset that measures
it.

## Part 2: the failure classes

The evidence collapses into eight classes. They are the vocabulary the rest of
this document scores against.

- **C1 Volume.** Change size and change count outgrow the reviewer. Review time
  per change is superlinear in size.
- **C2 Attention.** A queue over capacity converts review into approval: LGTM
  drift, senior bottleneck, no time to reconstruct intent.
- **C3 Intent.** The diff can be correct and still be the wrong change. The
  reviewer cannot see the requirement unless it travels with the candidate.
- **C4 Trust calibration.** Confident text is the default output; a reviewer
  cannot tell plausible from correct, and neither can an AI reviewer of an AI.
- **C5 Test adequacy.** Generated tests pass, assert little, and copy the
  implementation. Coverage goes up; behaviour stays unpinned.
- **C6 Security and supply chain.** Injection, secrets, permissions, and new
  dependencies with nobody's name on them.
- **C7 Provenance.** Which bytes were judged, by what, when, and on whose
  authority. Compliance asks this question; most pipelines cannot answer it.
- **C8 Comprehension.** The humans who must own, debug and extend the code did
  not write it. The knowledge did not transfer.

## Part 3: what odachi answers, mechanically

Not the marketing version: the four artifacts and the code path that enforces
each.

- **Intent before code (C3).** `brief.md` is written by the human before a step
  exists, quoted verbatim into the task doc, with "Done when" observable and
  non-goals that the fence can turn into paths. The review prompt carries the
  brief verbatim: `## The brief (what the human asked for)`. Most review tools in
  the field see the diff and nothing else.
- **The fence (C2, C6).** Paths the feature must not touch, read from the base
  ref by the gate in CI, so a candidate cannot loosen the policy it must obey.
  Scope creep becomes a check failure, not a reviewer's opinion.
- **The proof (C7).** `check --run` writes the real output, named by the tree it
  ran on. A proof for another tree is refused. This is stronger than most of what
  the field calls "CI green": the gate says *which bytes* passed, not that a
  branch did.
- **The verdict is bound to one tree (C7).** `review --attach` refuses a verdict
  for another tree; a correction measures only the delta (`--since`); findings
  travel with their receipts and cannot be re-filed.
- **The floor (C6, C1).** Before any model runs, the diff is read for the added
  lines a finding may cite, five signal categories, and a path class. A causal
  BLOCKER that cannot cite the candidate's own lines loses a tier. Silence is
  cheaper than noise, mechanically.
- **Risk tiering (C1).** Path classes decide the tier: `passive` (docs, no lens),
  `focused` (one lens), `full` (four). Content signals can raise the class.
  Scrutiny follows evidence, not volume — the same principle Morgan Stanley
  proposed from production experience.
- **Noise control (C2, C4).** Four rules out of six are about not reporting:
  style is banned unless it hides a defect, what was already there is a
  follow-up, an empty review is exactly one sentence, and one exhaustive pass per
  lens with no loop.
- **The second reader (C4).** `odachi-reviewer` is prompt-only and cannot run
  anything; `odachi-verify` is a separate agent that reads the repository, runs
  probes and answers per requirement. Named risks owe the verifier.
- **The ledger (C2).** `docs/review-debt.md` keeps potent findings against bytes
  that did not change, so a warning the human chose not to pay is not forgotten.

## Part 4: the coverage matrix

| Class | Mechanism | State |
| --- | --- | --- |
| C1 Volume | one step per commit, risk tiering, prompt cap + split | **partial**: no size escalation, split loses cross-part reasoning |
| C2 Attention | fence, causality rule, tiering, review debt | **covered** for one agent and one human; unmeasured |
| C3 Intent | brief carried into the prompt, verbatim requirements | **covered**, ahead of the field |
| C4 Calibration | severity vocabulary, one pass, refutation, verifier | **partial**: one model judges all four lenses, no recall measurement |
| C5 Test adequacy | reliability lens asks the right questions | **partial**: asserted by reading, never measured |
| C6 Security | risk lens, floor signals, fence | **partial**: no analyzer runs; the floor greps five token lists |
| C7 Provenance | proof and verdict named by tree, gate from the base ref | **covered**, ahead of the field, with one hole (below) |
| C8 Comprehension | the human reads briefs and verdicts, not code | **partial by design**; no artifact that transfers the change |

## Part 5: the first measurement of the lenses

`odachi review probe` (feature 0263) asks the four shipping lens prompts about a
fixture with defects planted in it and scores the answers against a written key:
found, called low, missed, and findings nobody planted. The fixture is a small
HTTP service; the key names five defects -SQL built by concatenation (risk), a
retry with no cap and a discarded body-read error (resilience, twice), a function
named `data` that reads a file, keeps a package global and returns a capped size
(readability), and a test that copies the implementation's formula (reliability).
The clean arm is the same service with none of them.

One run, model `opencode-go/deepseek-v4.1-flash`, 2026-11-10: seeded tree
`fb234044`, clean tree `37c46aac`.

| Arm | risk | readability | reliability | resilience | not in the key |
| --- | --- | --- | --- | --- | --- |
| seeded | 1/1 | 1/1 | 1/1 | 2/2 | 4 |
| clean | - | - | - | - | 10 |

What this says:

1. **Recall on the planted defects was 5 of 5.** Every defect the key names was
   found, at its place, at or above its floor, by the lens that owed it. On this
   fixture the four prompts do what they claim.
2. **The clean arm is not clean.** Ten findings on code with nothing planted, and
   reading them one by one, every one names a real weakness: the handler answers
   every database error -including a timeout and a refused connection- with 404,
   it never checks the response status from the remote, the response write error is
   discarded, `limit` clamps the upper side and not a negative, the handler and the
   cache have no tests. The lenses were right and the fixture was sloppy. Until the
   clean arm is genuinely clean, the number that reads as *noise* is really *a
   person's taste written into code the reviewer had no key for*.
3. **The key is the boundary of the measurement, not a verdict on the reviewer.**
   Four seeded findings fell outside it, and they were real too: a package global
   written from several goroutines without a lock, the request error discarded, the
   read error returned as success from another lens. A defect nobody planted is
   called noise by the score and would be a finding a person keeps.

Two earlier runs of the same fixture were made with a probe that never asked the
clean arm and with comments in the fixture that told the reviewer it was an exam.
Their numbers -3/4 and 4/5, a different defect missed each time- are not kept here:
the instrument that produced them could not report the half it was silent about,
and the fixture leaked its own answer. They are the reason both fixes exist.

The clean arm, three draws (0264). The same clean code, the same model, three
runs: 10 findings, then 1, then 4, and the defect that one run called a warning
the next one did not see. Ten real defects were paid between the first and the
second draw; the number still moved up on the third. So the reading is not "the
lenses are noisy" and not "the arm is clean": it is that **one run is one draw**,
and a precision number needs repetition before it means anything. That is the next
feature on the instrument, not on the fixture.

The repeated measurement (0265). `--runs 3` on the same fixture, same model,
2026-11-11: risk, readability and reliability found their defect in all three
draws; resilience found both of its own in two draws and one in the third; the
findings nobody planted moved from 3 to 7, and the clean arm from 0 to 5.

Read together with the three earlier draws, this is the shape of the instrument:
**what the key plants is found reliably, and what nobody planted is where the
variance lives**. A single precision number would hide that, and a single recall
number would flatter it. The reading is a median with a range, and the range is
the finding.
What it does not say: five defects in a 120-line fixture is not a population, one
run of one model is a snapshot, and a key written by the same person who wrote the
fixture inherits that person's blind spots. The instrument is what this part adds;
the sample is one tree, and the next measurement is the same fixture across models
and repeated runs.

## Part 5b: the draws on real work

The instrument left the fixture when `--runs` shipped in `v0.39.0`: one answer per lens
had always been one sample, and the tool can now buy more. Its first field reading came
from its own candidate, four reviews of one afternoon and two draws on the last of them.

- **`dfca16e5`, the full candidate read once.** One finding: a causal WARNING that a
  lens already paid to N could gain extra draws when it shared a call with a lens that
  still owed one. Approved, because a warning does not block.
- **`425dcaac`, the same candidate read once again after the merge.** Three findings,
  one of them a CRITICAL: a prompt split into parts counted one draw per part, so
  `--runs 1` recorded two. The first read had approved this code.
- **`721b02cc`, the fix read with two draws.** Two findings: one real -the refusal named
  one more than the largest count and so spent several draws while promising one- and
  one false, a CRITICAL about the widened handover that a test closed.
- **`bd559098`, the correction that closed the feature, two draws over the delta.**
  Approved, and the false row retired by a recorded dismissal.

What this says: on this candidate the second read did not find a better sentence for
what the first said -it found a different mechanism of the class the first had landed
on (the count), plus two defects the first never mentioned. The probe measured the same
shape on a fixture, where the key's defects are found every time and the variance lives
in what nobody planted; here the unplanted class was the counter itself, written by the
same agent that had just approved it.

It also says what a correction is not. The verdict that closed `0275` (`bbe8d2ff`) read
only the delta since `dfca16e5` and found nothing; the next full read of the merged tree
found a CRITICAL. ADR 0048's title was already right -a second reading is not a second
opinion- and this is the first field evidence that a *full* second reading is one.

What it does not prove: four verdicts, one model, one candidate, one afternoon. These
are counts of findings, not of defects that would have hurt, and a finding nobody fixed
would be the only thing that could measure the difference. Four draw counts in one
store is a start, not a rate.

## Part 5c: two models on the same fixture

Part 8 asked for the probe to be repeated across models, and the first reading of that
question came from the two cheapest the provider serves: `deepseek-v4.1-flash` and
`muse-spark-1.3-contributor`, one draw each on the same bundled fixture (two arms, five
planted defects, four lenses), on 2026-11-12.

The first table had one `noise` column and one `clean` column, and that was the defect
this part ended up naming: both counted a `SUGGESTION` the same way they counted an
invented `CRITICAL`, so "this reviewer talks" and "this reviewer invents" printed as the
same number. `0282` split it -a suggestion is a note, a warning or above is noise- and
the same stored answers, each one replayed through the new reading instead of asked
again, read as two numbers per arm:

| model | risk | readability | reliability | resilience | seeded noise | seeded notes | clean noise | clean notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| deepseek-v4.1-flash | 1/1 | 1/1 | 1/1 | 2/2 | 5 | 0 | 0 | 6 |
| muse-spark-1.3 | 1/1 | 0/1 | 1/1 | 1/2 | 1 | 0 | 0 | 0 |

The replay reproduces the first reading exactly -`noise 5 · clean 6` and `noise 1 ·
clean 0`-, so the split adds a meaning, not a measurement.

- **The planted class is not the question.** Both found the SQL injection the key calls
  `risk-query-built-by-concatenation`, both found the reliability defect, both found the
  retry defect in `cache.go`. What they disagree about is the softer defects: muse did not
  report the name that does three jobs (readability) nor the body error discarded as a
  success (resilience).
- **Neither model invented a defect on the arm with nothing planted.** Both `clean noise`
  columns are zero. The six findings deepseek raised there are six suggestions, and a model
  that fabricated a blocker on clean code would be the only one moving that column -which
  is the number the arm exists for.
- **The clean arm was never dirty.** Read one by one, the six findings are real
  observations about the fixture: a `time.Sleep` backoff that never selects on `ctx.Done()`,
  permanent errors retried three times, untested branches. All six are `SUGGESTION`, none
  invents a defect, and a person reading them would agree with several. "clean 6" was six
  improvements the arm could take, and the old column printed them as if they were six
  false alarms.
- **The seeded noise is small and it is made of claims.** Five findings at places no
  planted defect sits (deepseek) and one (muse), and neither model wrote a `SUGGESTION` on
  the arm that carries the defects: everything they said there either names a defect or
  opens with "something is wrong". That is the noise this probe exists to count.
- **The two are opposite instruments, and the split says where.** Deepseek finds every
  planted defect, makes five claims nobody planted and six suggestions; muse finds three of
  five, makes one claim and no suggestions. The old reading ranked them by the suggestions
  deepseek adds. The split ranks them by the claims they both make, 5 to 1. Neither is the
  better instrument, and a project that pays for one gets one half of each.
- **The next change this part named, the split, is `0282`.** It said the arm did not need
  to move before the reading did, and the re-read above is what the reading says once it
  can tell an opinion from an alarm.

What it does not prove: one draw per model, two models, one fixture, one afternoon. A
`--runs 5` reading would say which of these disagreements are the model's opinion and
which are the draw's, and a third model with a different family would say whether
"suggests on clean code" is one vendor's habit or a general property of a lens prompt that
asks for real defects only.

## Part 5d: the third model could not be measured

Part 5c left one question open: whether "suggests on clean code" is one vendor's habit or a
property of the lens prompt. The answer needs a third family and several draws. This is what
happened when it was tried on 2026-11-12, and it is written down because the attempt is
evidence about the instrument, not about the models.

- **`mimo-v2.6-flash` answered five of the six calls of one draw.** Each answer took between
  two and eight minutes, and the sixth -the clean arm's readability lens- stalled at the 600 s
  ceiling the probe allows, which is the same defect `odd/review.toml` recorded in 0190 and
  0192, now on the probe's small prompts instead of a full review.
- **What it did answer was valid.** On the seeded arm its risk lens named the SQL blocker
  (`server.go:25`) and the IDOR beside it (`server.go:21`); its readability lens raised five
  findings; and on the clean arm's risk lens it said nothing, with `approved`.
- **The free models are not a way around the cost, and the reason is the size of a probe
  prompt.** The same CLI -`opencode v2.2.7`, the binary the `PATH` wrapper calls, with no
  second installation- answers a small prompt with `opencode/flint-40b-free` and
  `opencode/glm-5.3-flash-free` in two seconds, and keeps answering when the calls are 45 and
  90 seconds apart. The 8.7 KB prompt this probe sends is a different question: both models
  answered `Free usage limit reached` after 45 and 90 seconds of waiting, and six calls back
  to back spent the budget from the first one. So the free tier is a small budget measured by
  burst and by size -a conversation fits in it, a probe call does not-, and a draw needs eight
  of those calls.
- **The cheap paid siblings were declined** for this session. The estimate was under ten
  cents for forty calls of prompts this size, which is a fact for the next decision and not a
  claim about their quality.

What this part proves: **availability is part of a judge's fitness.** A probe draw needs nine
answers -four lenses on the seeded arm, four on the clean one- and a score with a missing arm
is not a score. A model that takes eight minutes per call and stalls one in six cannot sit
behind a close either: `review` would pay the 600 s ceiling per stalled lens and return
nothing. The two models Part 5c measured answered every call in seconds, which is why their
numbers exist at all.

What it does not prove: anything about noise, notes or the clean arm for a third family. The
question stays where Part 5c left it, and it is answered the day a free tier allows a call this
size or a cheap model answers at this size.

## Part 6: our own numbers

From `.git/odachi/reviews/` in this clone, 16 verdicts kept:

- **15 approved, 1 changes_requested.** 13 findings in total: 7 WARNING, 5
  SUGGESTION, 1 BLOCKER. Nine came from the reliability lens, four from
  readability. Two models appear across the sample; every lens in a given verdict
  used the same one.
- A **94% approval rate** is exactly the signature the field warns about. It has
  two explanations and the repository cannot currently tell them apart:
  (a) the work is small, tested, and pre-verified — the verifier and the human
  already filtered it; (b) the lens review is a formality. Both fit the data.
- The record says (a) is at least partly true: the review caught the brief saying
  140 characters while the code measured 300 (0255), and found four real gaps in
  0258 that were paid before the merge. It also says (a) is not the whole story:
  in 0259 the **verifier** found the missing setup/init screen, the voseo, the
  untidy module and an unrecorded step, and the review that followed approved
  with one warning.
- One provenance hole, found while counting: the `at` field of every verdict in
  the sample is a round hour, because the reviewer writes it. The gate compares
  the model but not the time, so "when it was judged" is self-reported. C7 is
  covered for bytes and incomplete for time.

## Part 7: what is missing, with the probe that would show it

1. **The fixture's clean arm is sloppy** (0263 closed the gap that no measurement
   existed): ten findings on code with nothing planted, all of them real. The
   instrument reported them as noise because the key says a finding there matches
   nothing; `0282` split that count, so the same findings are read by what they
   claim -a note that suggests an improvement, or a claim that something is wrong-
   instead of all landing in one column. Precision is still not measured until the
   clean arm is code a careful person would approve. That is the next job on the
   fixture, not on the lenses.
2. **No deterministic analyzer runs.** The gate runs `go vet`, `go test -race`
   and the suite. It does not run `govulncheck`, a linter with a real ruleset, a
   secret scanner or a dependency review. The field's cheapest answer to C6, and
   the one class of defect static analysis catches better than a model, is absent
   from both the gate and the floor.
3. **A verdict cannot stop a close.** `check` passes with `changes_requested` on
   the tree; the mode asks for *a* verdict, not an approving one. That is a
   deliberate choice — the human decides — but the artifact says "ready" while a
   blocker stands, and the only thing preventing a merge is the agent's judgment
   on the day.
4. **The reviewer is blind outside the prompt, by rule.** It cannot read callers,
   tests, the rest of the file, or history. The verifier can, but only for named
   risks, and it runs before the review rather than as part of it. The gap the
   research names — contextual feedback — is precisely the part we outsourced to a
   different agent that is not always called.
5. **One judge for four lenses.** `odd/review.toml` names the same model for all
   four. The format already supports one model per lens, so heterogeneity is a
   configuration change, not a feature: two lenses answering with the same
   weights is one opinion in four costumes.
6. **Test adequacy has no evidence.** The reliability lens can only read the
   test. It cannot know whether the added lines are exercised, because nothing
   measures changed-line coverage or mutation on the candidate.

## Part 8: the verdict

**The lenses are not the problem.** Their taxonomy is not behind the field: intent
drift, test adequacy, causality of findings, partial failure, and the ban on style
noise are all there, and several of them are rarer in commercial AI reviewers than
here. Changing the four lens files would buy nothing.

**The instrument is the problem.** In the field's terms, odachi solved C7 and C3 —
provenance and intent — better than most production teams, and left C4 and C5
unmeasured, C6 deterministic, and C1 partly. Our review was a well-designed
opinion that had never been graded; 0263 graded it, and the grade says the lenses
find what a key plants, on a fixture whose clean arm still needs to earn its name.

The order that follows from the measurement:

1. **Run the deterministic tools** (C6, C1): a linter with a real ruleset plus
   `govulncheck` in the gate, and the result as a floor row the review must
   acknowledge. Cheap, reproducible, and it removes a class of defect from the
   model's attention entirely.
2. **Make a standing blocker visible** (C2, C7): a close that carries
   `changes_requested` says so in the gate's own line and needs a recorded human
   waiver to proceed, instead of "ready" with a footnote in a document.
3. **Make the clean arm clean** (C4): the probe's precision number is only worth
   reading once the arm with nothing planted is code a person would approve, and
   the first run says it is not. Then repeat the run across models.
   *The last clause is started*: a verdict stopped being one draw in `v0.39.0`
   (`--runs`), the first field reading is in Part 5b -a full second read found a
   CRITICAL the first had approved-, and Part 5c repeated the probe across two
   models, one draw each, which the split of `0282` made comparable. A third model
   with `--runs 5` is still open.

Then, in a second wave: heterogeneous judges for full-tier trees, size-based
escalation of the tier, changed-line coverage as a floor signal, and the fixture
grown across models and repeated runs.

## What this does not prove

- The external studies measure firms and open-source projects with more than one
  engineer, a review queue and a release cadence. This repository has one human,
  one agent and a proof gate; the volume problem (C1) is self-imposed here, and
  the attention problem (C2) appears as independence, not as queue pressure.
- The 16 verdicts are the ones retention kept (`keep_verdicts = 10` plus recent
  corrections), not the whole history. The 94% approval rate is a sample of the
  newest trees, not of the project.
- The comparison scores mechanisms, not outcomes. No claim here says a change
  reviewed by odachi has fewer defects in production; nothing measures that yet,
  and that measurement is item 1 above.

## Sources

- Chen, F. and Stratton, J., *AI in the Firm: Bottlenecks in Software
  Production*, Harvard working paper, Aug 2026 (718 firms; 68.3M issues, 200.9M
  commits, 49.1M PRs).
- Sonar, *State of Code Developer Survey*, January 2026 (1,100+ developers).
- DORA, *2025 State of AI-assisted Software Development* and *Balancing AI
  tensions*, March 2026 (1,110 qualitative responses).
- Zhong, S., Noei, S., Zou, Y., Adams, B., *Human-AI Synergy in Agentic Code
  Review*, arXiv:2603.15911, March 2026, revised October 2026 (278,790
  conversations, 300 projects).
- METR, *Measuring the Impact of Early-2025 AI on Experienced Open-Source
  Developer Productivity*, July 2025 (16 developers, 246 tasks), and its February
  2026 follow-up.
- Katz, D. and Elsawaf, K. (Morgan Stanley), Code Remix Summit 2026, as reported
  by Moderne, July 2026.
- JetBrains, *Stop Sending IDE-Catchable AI Code Errors to Review*, May 2026.