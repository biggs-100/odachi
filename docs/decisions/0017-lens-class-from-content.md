# 0017 — The lens class from content: upgrade only

Status: accepted · Date: 2026-10 · Session: v0.11.3

## Context

ADR 0010 made the lens selection path-based and proportional, and ADR 0016 added
a deterministic floor that reads the candidate's diff before any lens. 0016 left
one half undone on purpose: its fence forbade touching the selection rules, so
the floor reported the path class instead of raising it. The human reopened that
fence for this feature, with one condition that is the whole design: **content
may add review, it may never remove it**.

The failure the reopening fixes is concrete: a change to an ordinary file that
adds a goroutine, a mutex or a credential read is reviewed by the reliability
lens alone, because no path segment says `auth`, `worker` or `race`. The
specialist that would have caught the race is never asked.

## Decision

- **The path class is the floor.** `Assess(paths, diff)` starts from
  `Classify(paths)`. Only a `focused` candidate has room to be raised: a
  `passive` candidate (documents) stays passive, and a `full` candidate already
  asks for everything. A class is never lowered by content.
- **Two categories raise, and only two.** Added lines that spawn concurrency
  (`go func`, `chan`, `sync.`, `atomic.`, `mutex`, `<-`) or touch secrets
  (`password`, `secret`, `token`, `api_key`, `private_key`) raise `focused` to
  `full`. They are the two categories a path cannot see and a defect can hurt
  from. Network, failure and time signals stay in the prompt as evidence, but
  they do not buy four lenses: the reliability lens already covers them.
- **Prose is not code.** A document's added lines can carry the same words and
  change nothing: the passive class is immune to content. Otherwise a changelog
  that says "secrets" would demand a review verdict.
- **One assessment, two verbs.** `review` selects with `Assess`; `check` demands
  with the same function through `Need(paths, diff)`. The gate reads the diff
  from its own Git port. Two implementations would drift the day one changes —
  ADR 0010's lesson.
- **Deterministic, no model.** The assessment is the floor's fixed vocabulary,
  milliseconds, reproducible. No LLM decides how much review a candidate earns.

## Alternatives rejected

- **An LLM classifier for risk.** It would decide how much review a candidate
  earns with a call that can fail, cannot be reproduced, and has its own
  incentive to please. The deterministic floor is cheaper and auditable.
- **Let content lower a class.** A path that says `auth` reviewed as ordinary
  code because the diff looks calm is the defect the path signals exist to catch.
  Content can only add.
- **Always full.** Over-reviewing every candidate turns the review into a tax
  and the human turns it off; proportionality is what keeps the switch on.

## Consequences

- A concurrent or credential-touching change to an ordinary file asks for the
  four lenses, and the gate demands their rows per file — a verdict that covers
  only reliability is `review_lens_missing` or `review_coverage`.
- The reviewer sees the same class the gate asks under: the prompt's floor line
  says `full` and names the signals that raised it.
- Documents stay free: a documentation-only candidate still asks for no review.
- `Assess` and `Need` are the single authority. If the vocabulary grows, it grows
  in `floorSignals` and the two answers with it.

## Sources

- ADR 0010 (proportional review and the human switch), ADR 0016 (the floor).
- The floor's own signals: the categories are the lenses' vocabulary, not a
  second taxonomy invented here.
