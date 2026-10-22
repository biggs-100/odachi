# 0003 · Review verdicts: one file per tree

Status: accepted · Date: 2026-10 · Session: T20

## Context

The gate answers "is it proven". It never answered "is it well designed", and
the honest way to answer that today is to have an agent read the diff. A review
system can be a machine — revisions, correction budgets, consent envelopes,
refutation rounds — or it can be a piece of paper with the tree's name on it.
The machine is what makes people stop reading reviews.

The one thing an LLM reviewer reliably does wrong is produce noise: style
opinions, preferences, "consider extracting a helper". Noise is not harmless;
it is how a review becomes something people scroll past.

## Decision

- **The verdict is derived, not administered.** A finding counts only if the
  candidate `introduced, activated or worsened` it (`causal: true`) and its
  severity is `BLOCKER` or `CRITICAL`. Anything else is recorded and never
  blocks. There is no correction budget, no lifecycle, no second round: the
  rule is four lines of code and it cannot be argued with.
- **Storage is content-addressed.** A verdict lives at
  `.odachi/reviews/<treehash>.json`, one file per tree. Approval cannot float
  because the approval *is* the file name. A second verdict for the same tree is
  reported (`ya hay un veredicto para este árbol`) and dropped, never applied.
- **The store keeps itself out of the tree.** The verdict is written inside the
  candidate it describes, so `.odachi/reviews/.gitignore` (with `*` and
  `!.gitignore`) is written *before* the tree is hashed. Without that order,
  storing a verdict would change the tree that names it and every verdict would
  invalidate itself. This is the same trick `check --run` uses for `proof/`.
- **The prompt is fixed content, pinned by tests.** The brief, the changed
  paths, the diff, the five rules and the JSON schema. Without a diff there is
  no review: `no hay nada que revisar en este árbol`, exit 1. The five rules are
  the whole point: report only real user-impacting defects you can defend with
  evidence; when in doubt say nothing; style and preference are banned unless
  they hide a defect; severity is exactly `BLOCKER | CRITICAL | WARNING |
  SUGGESTION`; an empty review is exactly `No findings.`; and one exhaustive
  pass per lens, with no loop until it dries.
- **`check --require-review` is opt-in.** Without the flag a review never
  influences the gate: a beginner must not be blocked by a verb they have never
  heard of. With it, a tree without a verdict is exit 2 and `odachi review`; a
  `changes_requested` verdict is exit 2 naming the first causal finding; an
  `approved` one lets the gate continue. The review gate runs *after* the
  existing gates, because everything before it is something the agent can act
  on, and sending an agent to `odachi review` while the proof gate is failing
  would be a dead end that resolves nothing.
- **Nothing outside `.odachi/reviews/` is written**, and `gate` is untouched:
  the review reasons live in `internal/app/check`, not in the domain gate.

## What was taken from a review system, and what was not

Taken: the precision gate, the severity vocabulary, the canonical empty answer,
causality, and one pass per lens.

Not taken: a ledger in a memory service, refuters with 2-of-3 votes, consent
envelopes, lineages, revisions. Every one of those exists to make a machine's
judgement trustworthy by adding process. A file named after the tree does not
need the process: it can only be about one tree, and it can only say what it
says.

## Consequences

- **The project ignores its own evidence, and `init` owns that rule.** The store
  and `proof/` live inside the worktree they certify, so their ignore rule cannot
  arrive one run late: written lazily, each first use moved the tree and
  invalidated the evidence taken before it. `init` now keeps a marked block in
  the project's `.gitignore` —`/proof/` and `/.odachi/reviews/`— which survives a
  clone, is readable by a person, and is refreshed in place on every run.
  Verified end to end: `init`, `check --run`, `review --attach`, `check
  --require-review` passes with the proof byte-identical, and no second
  `.gitignore` exists anywhere in the project.
- **A project that never ran `init` still protects itself.** `check --run` and
  `review` ask the repository first (`git check-ignore`), and only when nothing
  ignores the path do they write the old per-directory `.gitignore` —before
  hashing, as always—. That is the legacy mechanism, kept for the projects that
  have not migrated and for the case where git cannot answer at all: evidence
  must never be allowed to move the tree because a probe failed. The price stays
  the old one for those projects: one retake each, once.
- **A hand-mangled verdict is reported, not ignored.** The store is a plain file
  a human can edit, so both `review --show` and `check --require-review` say the
  file cannot be read and name `odachi review --attach`.
- **`--attach` stamps what the file cannot know.** The tree, the base and the
  instant come from the tool; the reviewer and the findings come from the file. A
  declared tree that disagrees with the current one is refused instead of being
  rewritten, and a declared verdict that contradicts the rule is refused instead
  of being silently replaced: the author has to see why.
- **`--exec` gives the reviewer the prompt it prints.** The command is started
  with the same text `odachi review` prints — one builder, so the reviewer cannot
  be asked a different question than the gate asked — delivered two ways from one
  temporary file: the file is the command's stdin, and every `{prompt}` in the
  command is replaced by its quoted path. Both because neither is enough alone:
  a command that reads stdin needs no quoting and no knowledge of the shell, and
  an agent CLI that takes a file needs a path, while `$(cat)` does not exist in
  `cmd /c`. stdin is fed from a file rather than a pipe, so a reviewer that
  ignores its input cannot block the run by leaving the buffer full. Its answer
  is read from stdout, and when the answer omits `reviewer` the command itself is
  stamped as the reviewer — it is the one fact the tool knows about an answer
  that came out of a shell.

  Accepted limit: the prompt names the tree the verdict will be filed under, so a
  reviewer that writes into the worktree shifts that tree. The verdict is then
  filed under the shifted tree and the gate asks for a fresh proof, which is the
  same self-healing path any other change takes. A reviewer that writes is not
  reviewing; nothing is silently approved.

## Alternatives rejected

- **A revision counter or a verdict history.** That is the state machine this
  unit exists to avoid. One tree, one verdict: to review again, change the code.
- **A verdict that floats (`.odachi/review.json`).** The file would outlive the
  tree it judged, and the first stale approval would end the feature's
  credibility.
- **Letting an author write `verdict` by hand.** Two sources of truth for the
  same sentence. The findings are the facts; the verdict is arithmetic.
- **Requiring the reviewer to compute the tree hash.** It would make every
  hand-written verdict a copy-paste job and invite the one mistake that matters:
  a verdict filed under the wrong tree.
- **Blocking `review` on a missing brief.** The brief is `check`'s gate. A design
  defect is still a design defect without one, and the prompt says the brief is
  missing instead of the verb refusing to work.
