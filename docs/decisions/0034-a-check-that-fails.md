# 0034 - A check that fails is not an answer

Status: accepted · Date: 2026-10 · Session: after v0.11.71

## Context

`odachi init` decides a project's footprint by looking at it: a repository that
already has a commit and no project copy of the craft gets the minimal install
(0124), and everything else gets the full one. The same shape appears in
`odachi role`, which mirrors the block the project already carries.

Those reads are the decision. When one of them failed, the code asked the
question "did the read find it?" and treated a failure as "no":

- `derivesMinimal` used `err == nil && found`, so an `.agents/skills` it could not
  stat read as a project without the craft - and the whole curriculum went into a
  project that keeps it in the machine;
- `projectAgentPresent` did the same for the client's agent file, so a project
  that had one could be told it had none;
- `runInit` turned a failed `git rev-parse HEAD` into "this repository has no
  commit", which chose the full install for a repository git could not read;
- `projectBlock` fell back to the default block when `AGENTS.md` could not be
  read, so `odachi role` wrote a role over a block it never saw.

The convention this skipped was already in the same file: every other `Exists` in
`Run` answers a failure with `blocked("could not check ...; check the
permissions and run odachi init again")`. A check that fails is not a "no", and
the rest of the verb already knew it.

## Decision

- **An error is never an absence.** A read that fails is a fact about the run,
  not about the disk: it is reported, by the path it could not read and the
  reason the adapter gave, and the run stops.
- **The footprint is decided before the first byte.** Both halves of the
  adoption inference -where the craft lives and where the agent lives- are
  resolved by one function that runs before any writer, so a refusal leaves the
  project exactly as it was instead of half installed.
- **A state and a failure are different answers.** "This repository has no commit
  yet" is a state, and it keeps the full install; "git cannot use this repository"
  is a failure, and it stops the run. The CLI asks `Head` first and `CommonDir`
  second, because the second one is what tells a first day from a broken clone.
- **The same rule reaches `role`.** A block it cannot read is refused; a project
  with no `AGENTS.md`, or with one whose block the parser cannot find, still gets
  the block for its voice - those are states, not failures.

## Alternatives rejected

- **Treat a failed check as "present".** The other default: assume the project
  already has everything and touch nothing. It hides the failure just as well and
  leaves a project that never gets the craft, with a green run.
- **Warn and continue.** A footprint is a write of dozens of files. An inference
  drawn from a failed read is a decision the human did not make, and a warning
  scrolls past.
- **Parse git's message in `Head`.** Distinguishing "no commits" from "not a
  repository" inside the adapter would put a classification of English text in
  the layer that only runs commands; asking the second question is cheaper and
  says what it knows.

## Consequences

- `init` can now exit 1 on a machine where it used to exit 0 - a project whose
  `.agents/skills` or client folder it cannot read, or one whose `.git` is broken.
  That is the point: those runs installed the wrong footprint in silence.
- `derivesMinimal` and `projectAgentPresent` return the refusal with the answer,
  so the one place that decides before writing is also the place that reports.
- A future inference added to this verb inherits the rule from the signature: it
  cannot answer a bool without also answering the failure.
