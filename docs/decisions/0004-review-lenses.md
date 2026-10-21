# 0004 · Four lenses, chosen by evidence, and a refuter

Status: accepted · Date: 2026-10 · Session: T21

## Context

T20 gave `review` one generic prompt: the five rules, the brief, the diff, and
"find the defects". A generic reviewer is a generalist on purpose, and a
generalist misses what a specialist looks for: the retry with no ceiling, the
migration without a default, the branch no test reaches.

The tempting answer is to copy a full review machine — persisted ledger,
consent envelopes, refuters voting two of three — because that is what the
system being compared against has. Almost none of that carries power. What
carries power is four things: **specialised prompts**, **selection by evidence**,
**a refuter that must cite**, and **a gate that asks for each lens**.

## Decision

### The four lenses are content, not code

`internal/assets/review/<lens>.md`: frontmatter with `name` and a
`description: Trigger: ...`, then `## Qué mirar` with between 12 and 20 concrete
rules, then `## Qué NO reportar`. They are embedded with `//go:embed`, like the
skills, so a review needs no network and no repository.

The bar is that a rule is actionable or it is not a rule. `un retry sin backoff
ni techo de intentos` is a rule; "consider improving error handling" is noise
wearing a specialist's name. A test counts the rules per file and pins both
sections, the severity floor (`WARNING` and `SUGGESTION` nunca bloquean) and the
ban on style, so a file cannot quietly become vague.

Each lens inherits the five rules of Contract XIII from the prompt, which states
them once. Printing them per lens would let four copies drift apart.

### The tier is decided by evidence, never by volume

`review.Classify(changedPaths)` is pure and takes no diff and no line count, so
volume cannot decide anything: a thousand lines of documentation are `passive`,
five lines under `auth/` are `full`.

| Tier | When | Lenses |
| --- | --- | --- |
| `passive` | every changed path is documentation (`.md`, `.txt`, `README*`, `docs/`) | none |
| `focused` | anything else | `reliability` |
| `full` | any path touching a sensitive signal | the four, in fixed order |

The signals are the ones the fence already knows: `auth`, `login`, `session`,
`token`, `password`, `secret`, `permission`, `role`, `payment`, `billing`,
`migrat`, `schema`, `.env`, `docker`, `ci`, `deploy`, `Makefile`, `go.mod`. They
match the **start of a path segment**, case-insensitively, so `migrat` catches
`migrations/` and `ci` catches `ci.yml` without catching `decision.md`.

Signals are checked **before** the documentation exemption: a document whose path
touches a sensitive signal is `full`. Over-reviewing a document costs one prompt;
under-reviewing auth costs a defect.

`--lens` **adds** and can never remove: the selection is `Add(required, extra)`,
always in the fixed order `risk, readability, reliability, resilience`, so the
same candidate produces the same prompt bytes.

### The verdict records the lenses it covered

An immutable store cannot be repaired: if the gate asked for a lens the stored
verdict does not cover, there would be no way out that is not deleting a file by
hand. So the tool **stamps** the lenses, exactly like the tree and the instant:
the reviewer never declares which lenses it used, and the candidate is what says
which ones were required. The verdict carries `lenses`, `--show` prints
`lentes: … · requeridas: …` plus the ones that are missing, and
`check --require-review` asks for the same set from the same changed paths.

The missing-lens branch therefore covers what the verb cannot stamp: a verdict
written before this unit, or one a human placed by hand. Its message names the
lens, the file, and the two human steps (`borrá <archivo>` and `odachi review`),
because the store is immutable and the tool never deletes.

### The refuter argues with blockers only, and doubt keeps the finding

`odachi review --refute <veredicto>` prints a prompt for one verdict: the causal
`BLOCKER` and `CRITICAL` findings with stable ids (`F1`, `F2`, …, the position in
the file), the diff, and three questions per finding:

1. ¿Es causal del diff, o ya estaba? Citá la línea del diff.
2. ¿Es un defecto real que afecta al usuario, o una preferencia?
3. ¿La evidencia alcanza para defenderlo ante el autor?

The rule is the one that makes a single refuter safe instead of one more
opinion: **no citation, no refutation**. `SealRefutation` applies it
deterministically — a `refuted` with an empty `why` becomes `stands`, and a
blocking finding the refuter did not answer is recorded as `stands`. The default
is always that the finding stands; nothing is ever refuted by default.

The answer is `{"refutations":[{"id","verdict","why"}]}`, stored once at
`.odachi/reviews/<treehash>.refuted.json` and never overwritten. There is no
2-of-3 vote: one refuter, one criterion, and the criterion is evidence.

A refutation is about **one tree**: `--refute` refuses a verdict from another
tree, because the only evidence a refuter can cite is the diff in front of it.

### The gate

`check --require-review` (still opt-in) now requires a verdict for this tree, an
answer for **each lens the candidate requires**, and no causal BLOCKER or
CRITICAL left standing after the refutation is taken into account. A refuted
blocker does not block. A refutation that cannot be read blocks and names the
file: silence would let a broken artifact look like an approval.

## What was taken, and what was not

Taken: lens specialisation, selection by evidence, one refuter with a citation
requirement, and a gate that asks per lens.

Not taken: a persisted ledger, consent envelopes, 2-of-3 votes, correction
budgets, lineages. They exist to make a machine's judgement trustworthy by
adding process; here the trust comes from the file name, the closed vocabulary
and the citation.

## Consequences

- One extra embedded content directory, no new dependency; the domain stays
  stdlib-only and the boundary guard stays green.
- `--attach` and `--exec` store whichever document they are given, verdict or
  refutation. The two share no key and both decode strictly, so `Sniff` decides
  and nothing is guessed from a file name.
- An attached document must be written **outside** the project: a file left in
  the worktree becomes part of the tree hash the verdict is named after.
- `Classify` ignores `.odachi/reviews/**` and `proof/**`: storing a verdict or a
  proof creates a `.gitignore` inside those directories, and counting it as a
  candidate path turned a documentation-only change into a focused review the
  second time it was asked. Everything else under `.odachi/` — the fence, the
  profile — is a human decision and stays in the candidate.
- `--show` reports the stored verdict and, separately, what the refutation did to
  it (`se quedan: 0 · refutados: 1 · ya no bloquea nada`). The stored verdict is
  never rewritten: the file says what the reviewer said, and the refutation says
  what survived.
- Review load grows on purpose: an `auth` change now prints four prompts. That is
  the trade the tier table makes, and it is visible in `--show`.
