# 0007 · Delegation: triggers and handoff, not an orchestrator

Status: accepted · Date: 2026-10 · Session: v0.11.0

## Context

`odachi` deliberately has no orchestrator. The agent decides, and the binary
enforces the gate. Delegation was left implicit: the `slices` skill had good
advice, but nothing said which policy wins when another global skill also talks
about subagents. A live OpenCode run loaded `slices` and a global
`hermes-ephemeral-delegation` skill in the same turn and merged their thresholds;
the answer called odachi an "orchestrator" and used file counts that odachi never
had.

The reference implementation, `gentle-ai`, solves this with a dedicated
always-on orchestrator, worker/verifier agents, and a table of mandatory
triggers. The useful part is not the machinery; it is the triggers and the
verification. Copying the machinery would turn odachi into a second gentle-ai,
which is the opposite of the product.

## Decision

- **No orchestrator.** The agent remains the orchestrator of its own work. Odachi
  ships policy, not a router, a budget service, or a lifecycle.
- **`slices` owns delegation.** Within an odachi project its policy wins over any
  other global skill. `drive` — the manual the `AGENTS.md` block makes the agent
  load before acting — sends the agent to `slices` when the request does not fit
  in one reading. That chain is the enforcement odachi can afford today.
- **Delegate to protect the parent context, not to save tokens.** A subagent
  re-reads everything and costs more total tokens. The reason is a cleaner parent,
  never a smaller bill.
- **Triggers, not size.** Delegate an explorer when understanding passes ~3
  readings / ~10k tokens or the session is long. Delegate a writer only by a named
  reason: two or more independent units with disjoint edit surfaces, each heavier
  than starting a subagent; or the ~150k-token context backstop. Never by size,
  file count, a large task alone, or price.
- **One writer per unit.** Parallel writers have disjoint surfaces; the parent owns
  git and the final seam. Reading that prepares a write goes with that write.
- **Handoff by reference.** The subagent receives the task doc path and the step,
  not a paraphrase; its return is bounded (~2k tokens) with `path:line` evidence
  and the files it read. It must record progress with `odachi plan` and must not
  say "listo" without `odachi check --run`.
- **Verify against the brief and the diff, never the summary.** The subagent can
  lie like the principal, with fewer witnesses. The parent owns the decision of
  vision; the subagent executes.

Taken from gentle-ai: the mandatory triggers, the context economy, the bounded
handoff, the per-action distinction (tests/builds may use fresh workers without
changing the implementation route), and the idea of verification proportional to
risk.

Not taken: the 180-line always-on orchestrator, ODD/RDD, feature documents with
spec/task/log and an Engram mirror, consent envelopes, provider-defect handoff,
Judgment Day, a 400-line delivery heuristic, and a fleet of specialized agents.
Those are a product of their own; odachi's job is to be small enough that a
beginner can hold it in their head.

## Consequences

- Delegation is **policy, not a gate**. Nothing stops a parent from skipping
  `slices`, and a competing global skill can still be loaded. The precedence line
  in `slices` and the pointer in `drive` are the mitigation; they are not a
  mechanical guarantee.
- Odachi does not spawn subagents itself; it uses the client's mechanism. A client
  without subagents keeps working inline, exactly like gentle-ai's reduced form.
- Verification of delegated work is still the parent's job. A future version can
  add a worker/verifier agent and a risk tier, but that is a feature with its own
  evidence, not a line in a skill.
- The `slices` body is capped at 60 lines. The policy that does not fit goes to a
  new skill or an ADR, not into an ever-growing manual.
