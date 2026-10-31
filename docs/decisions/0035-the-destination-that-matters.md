# 0035 - The destination the scope chose is the one that must be written

Status: accepted · Date: 2026-10 · Session: after v0.11.72

## Context

`odachi init` writes the craft to one of two roots: the project's `.agents/skills`
or, with `skills = home`, the machine's `~/.agents/skills`. When a destination
refused the write, `initSkills.report` decided whether the run failed by looking at
the folder that failed -the project's copy was fatal, the home's was reported as
`pending:` with exit 0- and that rule was written when a project always had a copy
of its own: a broken home still left the agent with the curriculum in the tree.

0124 changed the premise. A repository that already lives keeps the craft in the
machine, so for those projects the home is the only destination. The block naming
`~/.agents/skills` is written before the skills half runs, so a home that refused
the write left a run that exited 0 while the folder it had just named did not
exist -exactly the state 0123 removed, one verb over.

## Decision

- **Fatality is the scope the run asked for, not the folder that failed.** Every
  pack run is paired with the scope it was asked for; the result whose scope is the
  scope the project chose must be written, and a failure there fails the run. Any
  other destination stays `pending:` with exit 0, because the project still has
  what it needs.
- **The line says which one it is.** `failed:` for a fatal destination, `pending:`
  for the other. A green run that names a folder nothing created is the defect;
  a word that hides a failure behind "pending" is the same defect in miniature.
- **The refusal names the verb that ran.** `blockedSkills` hardcoded `odachi init`,
  so a failed `setup` -whose only destination is the home and whose job is the
  machine- told the human to run a command that would not retry it. `setup` now
  fails with its own name, and its closing line no longer claims the bootstrap
  skill was installed when the write failed.
- **`setup` inherits it.** Its only destination is the home, so its failure is
  fatal by the same rule; before, it exited 0 and left the machine half-installed.

## Alternatives rejected

- **Keep the folder rule and special-case the adoption.** It works for the two
  scopes the tool has today and breaks the moment a verb asks for one destination,
  which is `setup` already.
- **Make every failure fatal.** A project-scope run whose home fails still has its
  curriculum; failing it would refuse a project that is ready, and the bootstrap
  skill in the home is a convenience, not what the agent reads.
- **A warning on a fatal destination.** A run that did not do its job is not a
  warning: exit codes are what scripts and the hook read.

## Consequences

- `init --skills home`, `init` in an adopted project, and `setup` now exit 1 where
  they exited 0 with a `pending:` line. The message names the file, the reason and
  the verb.
- A project-scope run keeps its old behaviour for the home half, pinned by
  `TestInitKeepsTheProjectWhenTheHomeIsNotWritable` and
  `TestInitReportsABlockedHomeAndKeepsTheProject`.
- A future verb that installs into the home inherits the rule from the pairing:
  it lists the destinations it asked for, and the chosen one must be written.
