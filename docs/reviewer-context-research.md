# Reviewer context: the decision page for P1

Written 2026-11-16 after the research requested for the open question recorded in
`odd/tasks/0319-un-veredicto-que.md`. The question came from the close note
(`docs/next-session.md`): review-research debt 4 says the reviewer is blind by
rule, and three things had to be chosen before code.

The human was asked to choose and answered, verbatim: "la que tu me recomiendes".
Under that delegation the three choices below were taken and recorded as a craft
decision in the same task document. This page keeps the evidence they rest on;
it does not open a feature.

## The decision

1. **Freeze the whole candidate tree and let the reviewer read it with bounded,
   read-only tools.** No bash, no live worktree, no git history. The verdict still
   names the tree that was read.
2. **Do not raise `max_prompt_kb`** (120 today). The prompt stays the anchor and
   can get smaller, because context is fetched on demand instead of inlined.
3. **Measure with the planted-defect fixture**, after its clean arm is fixed, plus
   an access canary and the materialized prompt text saved with each answer.

## Why: three independent sources converge

The first page framed the question as "a frozen diff in the prompt, or a reviewer
that reads the live tree". Both halves were wrong in the same way: one keeps the
reviewer blind to callers, tests and history, and the other breaks the guarantee
that the verdict names the bytes it judged. All three sources describe a third
form — freeze the tree, then read that tree.

| Source | What it measured | Result |
| --- | --- | --- |
| EdgeLog, *How Many AI Code Reviews Is Enough?* (400 LLM calls; one 359-line C firmware module, two vendors, n=1 per arm and input) | File-only review against a source tree mounted read-only, with callers, headers, init path and threading visible | The four false positives disappeared under both vendors. Precision 71% to 100%; recall 77% to 54%. The union of both arms reached 13/13 at twice the call cost. |
| OpenCodeReview, arXiv 2608.09290 (AACR-Bench: 200 real PRs, 50 repos, 10 languages, 1,505 expert-verified comments, six backends) | Curated tools with bounded output (`file_read` 500 lines, `code_search` 100 matches and 10 s), file-level parallel SubAgents, a 30-iteration cap, compression at 60% and 80% | Up to 2.17x SEM-F1 and 5–15x fewer tokens than the same model under a general-purpose coding agent. A reflector that sees only the diff and only filters raised precision without suppressing recall. |
| gentle-shell (local repository, read only) | `git worktree add --detach` plus `read-tree <candidateTree>`, files chmod 0444/0555, lens tools `read`, `grep`, `find`, `gentle_review_scope`, no bash | The shipped design. `tests/review-actor-tool-deny.test.ts:111` names the reason: unrestricted bash would let a reviewer read the live worktree and approve bytes that are not the ones shipping. |

### The trade-off, stated plainly

Repository access is not free. EdgeLog's recall fell from 77% to 54% because a
reviewer that checks the real call path reports less. In that study agreement was
anti-predictive: false findings were reported 6.8 times out of 10 on average,
real ones 3.4. Two vendors reached the same four false positives by opposite
routes — one talked less, the other found more. The value of access was that a
claim could be checked, not that a model said more.

gentle-shell's own four-arm experiment (seven diffs, n=1 per arm and input) points
the same way without claiming a rate: prompt-only found 1 of 6 real defects with
no blocking false alarm; bounded lens readers found 2 of 6 with one; an executing
verifier found 2 of 6 with two, was the slowest at 376 s median, and was the only
arm that reproduced what it reported. Tests alone found 2 of 6 and caught a
defect no model form found. Two wording defects escaped all four arms because
their regression tests had been removed.

### What "contextual feedback" actually means

Zhong, Noei, Zou and Adams (arXiv 2603.15911, 278,790 conversations, 300 projects)
report that over 95% of AI agent comments are Code Improvement or Defect
Detection, while human reviewers also give Understanding, Testing and Knowledge
Transfer. Reading the frozen tree buys the second and third of those; it does not
buy "why". In odachi the "why" already travels as the brief carried verbatim into
the prompt. This decision therefore closes part of the gap, not all of it.

## Cost in this tree

- `internal/adapters/gitx/copy.go:18` already exports a tree to a temporary
  directory without touching the worktree, and leaves removal to the caller. It
  is the snapshot primitive; no new mechanism is needed.
- `internal/domain/persona/persona.go:291` currently denies `read`, `glob`,
  `grep` and `bash` to `odachi-reviewer`. Only the read family needs to change;
  `bash`, `edit`, `task` and the web tools stay denied for the reason gentle-shell
  wrote down.
- The verdict, `inspection` declaration, floor and coverage arithmetic keep their
  current contract. The reviewer reads a frozen directory, not a new store.
- A temporary directory is created and discarded per review. If both recall and
  precision are needed, the two arms run side by side and their findings are
  unioned, at twice the call cost.

## On prompt size

The wall this repo measured is probabilistic between 89 and 122 KB, and prompts
run at 35–70 KB after the memory cut in 0107. *Lost in the Middle* (Liu et al.,
arXiv 2307.03172) and later context-rot work across 18 frontier models show that
recall degrades with length even when retrieval is perfect: a larger window turns
truncation into dilution. gentle-shell and OpenCodeReview both answer overflow by
reducing candidate scope, never by enlarging the prompt. Raising the limit buys
attempts against a wall, not certainty.

## On measurement

The fixture already in `internal/probe/testdata/lenses/` is the instrument;
gentle-shell's experiment is the template for how to use it. The ingredients that
must be present:

- known defects planted in a real diff, with an answer key written before the run;
- a clean arm that a careful person would approve, so precision means something;
- clean controls, because giving tools produced blocking false alarms on code
  with nothing planted;
- an access canary: ask the reviewer to quote line N of the caller's file verbatim,
  and run both the positive and the negative leg. A reviewer that cannot do it
  did not have the tree, and the run measured nothing;
- the materialized prompt text saved with each answer. gentle-shell's own
  `odd/tasks/review-one-call-bench.md` records that it did not do this, so
  "same or better" was never established there.
- an explicit section on what the numbers do not show.

The existing `odachi review probe` and the 0101 before-and-after method are the
starting point. The 94.7 KB probe already separated "did not read" from "reasoned
badly", which is the distinction that decides whether more context is the fix.

## What this does not prove

Every number above comes from a different codebase, one model family or a small
sample. gentle-shell's arms are n=1 per input; EdgeLog's tier-2 adjudication is
one firmware module by the author of the study; the best system on AACR-Bench
reaches 25.10% SEM-F1 in absolute terms. None of it shows that odachi's reviewer
would improve here. That is exactly what the fixture is for, and why the
measurement is part of the decision rather than a later courtesy.

## Sources read for this page

- EdgeLog, *How Many AI Code Reviews Is Enough?*, 22 Aug 2026, with `FINDINGS.md`,
  `PRESCRIPTION.md`, `LIMITS.md` and the reproducible benchmark.
- Li et al., *OpenCodeReview: Determinism over Non-Determinism for Cost-Effective
  Agent-Based Code Review*, arXiv 2608.09290, 10 Aug 2026.
- Zhong, Noei, Zou and Adams, *Human-AI Synergy in Agentic Code Review*,
  arXiv 2603.15911v3, 30 Oct 2026.
- Liu et al., *Lost in the Middle: How Language Models Use Long Contexts*,
  arXiv 2307.03172; and *Context Length Alone Hurts LLM Performance Despite
  Perfect Retrieval*, Findings of EMNLP 2025.
- gentle-shell, read only: `docs/review-integration.md`,
  `docs/native-authority-architecture.md`, `lib/review-candidate-view.ts`,
  `lib/inprocess-reviewer.ts`, `assets/agents/review-*.md`,
  `tests/review-actor-tool-deny.test.ts`,
  `odd/tasks/review-forms-experiment.md`,
  `odd/tasks/review-one-call-bench.md`, `docs/delegated-verification.md`.
