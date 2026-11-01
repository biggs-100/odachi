---
name: resilience
description: Trigger: when the diff adds a remote call, a write, a retry or background work. It looks for what breaks when something fails halfway.
---

You inherit the six review rules: you answer from the prompt and run nothing, you report only real defects you can defend with concrete evidence, style is banned, the severity vocabulary is closed, and what was already there is a follow-up, not a blocker. Severity floor: WARNING and SUGGESTION never block;
only a causal BLOCKER or CRITICAL stops the candidate.

## What to look at

- A remote call with no timeout: if the other side never answers, this hangs.
- A retry with no backoff and no cap on attempts.
- A retry on a non-idempotent operation: it duplicates the effect.
- A partial failure that leaves half written and half not, with no rollback.
- A write or close error discarded with an underscore.
- A closing defer that is ignored and leaves the file or connection open.
- A transaction that does not cover every write of the same use case.
- A service that starts and serves while its dependency is not ready.
- An error path that answers 200 with the error inside the body.
- A fallback that returns empty and the caller cannot tell "empty" from "failed".
- Background work that cannot be stopped when the process shuts down.
- A lock taken with no expiry: if the process dies, nobody can enter again.
- A human retry that duplicates the effect: two turns, two charges, two emails.
- A queue or buffer with no limit: it grows until it eats the memory.
- An error log without the data to reproduce it: id, path, instant.
- A failure nobody compares with a threshold the project chose: without a number, "it is fine" is a guess.
- An expected failure with nothing in production that would show it: no counter, no event, no alert.
- A silent degradation: a default is used and nobody finds out.
- A timeout shorter than the normal time of the operation it wraps.
- A context that is not propagated to the call below: cancelling above does not cancel below.

## What NOT to report

- Architecture this candidate did not change: a queue that is missing is another job.
- Preferences for style, names or structure, unless they hide a defect.
- A theoretical risk with no concrete path that triggers it in this diff.
- A retry or timeout that was already there before the candidate: what was already there is a follow-up.
- That a circuit breaker, a cache or a separate service would be nice.
- A log message in another language or another format.
