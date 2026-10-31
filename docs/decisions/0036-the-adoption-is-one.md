# 0036 - The adoption is one inference, and three verbs ask it

Status: accepted · Date: 2026-10 · Session: after v0.11.73

## Context

`odachi` infers where the craft and the native agent live when a repository already
has a commit: the project keeps the rules and the hook, and the craft and the agent
stay in the machine (0124). Three verbs live on that inference:

- `init` writes the project and decides the footprint from it;
- `role` writes a role's skill into the folder the craft lives in (0126);
- `doctor` reads the shape the adoption leaves and says whether it is finished.

Each one gathered the facts itself. `init` had `derivesMinimal` and
`projectAgentPresent`, `role` re-derived the craft half in the CLI, and `doctor`
read the result without knowing what it was: check 11 saw no agent file and no
`project_agents` key and called a correct adoption "an unfinished one", advising
`odachi init --global` to record a choice nobody made.

The duplication was not academic. The review of 0126 spent a round on a suspicion
that the two verbs read different facts, and answering it cost a test and a
dismissal: a reader of the code could believe it because the two predicates were
written twice.

## Decision

- **The facts and their meaning live in `internal/app/adoption`.** `Facts` carries
  the four answers (a repository that lives, the craft here, the agent here, the
  store), and `Craft()`, `Agent()` and `Scope(flag)` are the only definitions of
  what they mean together. `Read` asks them in order and stops at the first one that
  fails, because a read that failed is not a fact that is false (0125).
- **Each verb keeps its own ports and its own words.** The facts arrive as
  closures: `init` reads them through its `Files` and the store it already loaded,
  `role` through `os` and the CLI's git adapter, `doctor` through its read-only
  `Files`, its `Agents` port and a new `Git` port that answers one question. An
  error is the caller's, so a refusal names the path the verb asked about.
- **`doctor` reads the inference instead of describing its leftovers.** When the
  shape is an adoption it says so -the rules and the hook here, the craft and the
  agent in the machine- and names `odachi setup` when the door is missing. A store
  that keeps no choice is the design, not a gap, and the note says that out loud.
- **A note that cannot tell does not claim.** A repository git cannot read answers
  "no commit" to the doctor, because check 2 already names a broken repository and
  two checks blaming one cause is noise. The note falls back to what it sees.

## Alternatives rejected

- **Let `doctor` infer on its own.** Four more lines of the same rule, and the
  third place to keep in step: the defect this feature removes.
- **Only change the note.** The advice stops naming `init --global`, but the
  doctor still cannot tell an adoption from a project a half-install left behind,
  and the next verb writes the predicate a fourth time.
- **A `Git` port that can also write.** The doctor is read-only by construction,
  and its port has no write method on purpose; it asks one question.
- **Passing `Facts` from the CLI.** `init` infers inside its own run, before it
  writes, and its refusal vocabulary is its own: the CLI would have to know which
  verb is asking to phrase the failure.

## Consequences

- One more package under `internal/app`, in the shape of `internal/app/tasks`: pure
  answers, ports supplied by the caller.
- `doctor` gains a read-only `Git` port; a doctor wired without it never claims an
  adoption, which is the safe default for a test that does not care.
- A future verb that has to know where the craft lives asks `adoption` instead of
  reading `.agents/skills` itself.
