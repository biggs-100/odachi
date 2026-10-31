# 0038 - Clean removes only what odachi wrote

Status: accepted — Date: 2026-10 — Session: after v0.11.72

## Context

A clone accumulates four kinds of file and no verb said whose they were. Measured
in this repository on 2026-10-31: `proof/` 366 KB (133 files), the clone's store
`.git/odachi/reviews` 14.7 MB (240 files: 146 verdicts, 89 prompts, 4 dismissals
and one scope record), `.git/odachi/backups` 792 KB (751 files), and 82 files of
notes an agent wrote loose in the git directory because the tree had to stay clean
and nothing named a folder for them. Nothing in the tool declared an owner or a
lifetime, so every class grew forever.

The obvious verb -"delete what is not needed"- is the wrong question for `.git/`.
Ownership cannot be inferred: `objects/` is git's, `gentle-ai/` is another tool's,
`engram-project-identity.json` is another tool's, `.COMMIT_EDITMSG.swp` is an
editor's, and the 82 notes are an agent's. A tool that guesses will one day delete
a file it did not write, and a clone is exactly where that mistake costs the most.

## Decision

- **Every class is named with its owner and its size, and one class is one rule.**
  The report is the verb: it reads, measures and prints, and writes nothing.
- **Ownership is declared, never inferred.** odachi removes only the four classes
  it wrote: review records, review scaffolding, undo backups and the agent's
  scratch folder. `proof/`, `odd/tasks`, the client choice and the lock, and
  everything else in `.git/` are named and left alone, whatever the lifetimes say.
- **Each class carries one lifetime, declared where the code lives.** A review
  record is one tree -a verdict, its prompt, the dismissals and the scope record
  beside it- and the newest ten survive. A backup is one version of one artifact
  and the newest five survive. The scratch folder is not history at all: it is
  emptied. The numbers are this release's default; a project that wants another
  one asks for a knob, which is a later feature (`keep_verdicts`, `keep_prompts`
  in `odd/review.toml`).
- **A failed read is not an absence.** A class the run maintains that it could not
  read stops `--apply` before the first removal, and a date it could not read
  keeps the file out of the class instead of making it look old. The report prints
  every unreadable path; the run never deletes on a partial view.
- **Another writer's name is never a version.** A file in `backups/` that does not
  carry the `.bak-<8 hex>` shape `undo.SplitBackup` accepts never joins a group,
  so it is never counted as old and never removed.
- **`--legacy` names the agent's old habit and is never implied.** The notes an
  older agent left loose in `.git/` are listed by every run and removed only by
  `--apply --legacy`; `--legacy` alone is refused, because a flag that removes
  nothing is a flag the human will misread.
- **The block teaches the folder.** The rule that an ad-hoc note lives in
  `.git/odachi/scratch/` is in the AGENTS.md block, not in the `drive` skill: the
  block is the door every agent reads in every project, and the fence in force at
  this feature protects `skills/**`, so the skill would have cost a policy change
  first and reached fewer readers.

## Alternatives rejected

- **Pruning `proof/` too.** It is 366 KB of the gate's evidence, tied to a tree;
  deleting it makes `check` ask for a run that already happened. The report shows
  its size so the human can decide to delete `proof/` by hand, which is a
  decision about evidence and not a retention rule.
- **Retention by age (30 days, 90 days).** A date is easy to explain and wrong in
  both directions: a long-lived branch reviewed three weeks ago would lose the
  verdict its close needs, and a slow month would keep bytes nobody is going to
  read. A count per class is the shape the two things actually have: N records,
  N versions.
- **Inferring owners from git's known file list.** The list drifts with git, and
  it turns a new file into a file odachi believes it may delete. Declared classes
  fail closed: an unknown path is never in a class that removes.
- **Deleting on `--apply` and re-scanning to print the survivor.** The report a
  human reads is the one that justifies the deletion; it shows what was there and
  what went, and the message adds the freed size.

## Consequences

- `odachi clean` reports before it acts, and the numbers in the report are the
  numbers `--apply` uses, because both read the same map.
- A clone can now be emptied back to a small store in one command, and the only
  files it loses are ones odachi wrote, past their declared lifetime.
- The store still copies inherited rows into every verdict, which is the reason a
  verdict grew from 0.7 KB to 125 KB in eight days; retention bounds the count, not
  the size. The next feature makes a verdict reference one memory keyed by blob
  instead of rewriting it, and the one after it puts the numbers in
  `odd/review.toml`.
- A project that keeps its evidence elsewhere -another clone, a CI artifact- still
  pays for the copies here until it runs `clean`; that is the price of a store that
  never travels.
