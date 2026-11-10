# Review under pressure: what the field learned, and what our review does with it

Written 2026-11-10, after a session that shipped 0262. The question: code is now
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

## Part 5: our own numbers

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

## Part 6: what is missing, with the probe that would show it

1. **Nothing measures the reviewer itself.** `docs/skill-evaluation.md` measures
   skills with probes and a control arm. No equivalent exists for the lenses: we
   do not know their recall on a planted defect, or their false-positive rate on
   clean code. Every claim in Part 4 about the lenses is a reading of the prompt,
   not a measurement of the output.
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

## Part 7: the verdict

**The lenses are not the problem.** Their taxonomy is not behind the field: intent
drift, test adequacy, causality of findings, partial failure, and the ban on style
noise are all there, and several of them are rarer in commercial AI reviewers than
here. Changing the four lens files would buy nothing.

**The instrument is the problem.** In the field's terms, odachi solved C7 and C3 —
provenance and intent — better than most production teams, and left C4 and C5
unmeasured, C6 deterministic, and C1 partly. Our review is a well-designed opinion
that has never been graded.

Three things would move it from opinion to instrument, in this order:

1. **Measure the lenses** (C4, C5): a seeded-defect fixture and a clean-code arm,
   the same doctrine `skill-evaluation.md` already uses for skills. Without it,
   every other improvement is an assertion about quality we cannot show.
2. **Run the deterministic tools** (C6, C1): a linter with a real ruleset plus
   `govulncheck` in the gate, and the result as a floor row the review must
   acknowledge. Cheap, reproducible, and it removes a class of defect from the
   model's attention entirely.
3. **Make a standing blocker visible** (C2, C7): a close that carries
   `changes_requested` says so in the gate's own line and needs a recorded human
   waiver to proceed, instead of "ready" with a footnote in a document.

Then, in a second wave: heterogeneous judges for full-tier trees, size-based
escalation of the tier, changed-line coverage as a floor signal, and the verifier
on every full-tier tree rather than only on named risks.

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
