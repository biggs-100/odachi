# 0009 · Native subagents: the orchestrator's effect without the orchestrator

Status: accepted · Date: 2026-10 · Session: v0.11.1

## Context

`gentle-ai` keeps the main agent's context clean with a dedicated orchestrator:
it classifies the task, delegates reading and writing to bounded workers, passes
exact skill paths, and verifies the result with an independent actor. The value is
not the router; it is that the parent never carries the module, the diff or the
exploration.

Odachi deliberately has no orchestrator (ADR 0007). It had the policy in `slices`
but not the mechanism: a generic client subagent could still be used, but there
was no worker contract, no verifier, and no handoff that named the task doc,
surfaces, commands and skill paths.

## Decision

- **The parent remains the orchestrator.** The always-on layer is still the
  `AGENTS.md` block plus `drive`; `slices` is the delegation policy.
- **`init` writes two native subagents per client:** `odachi-worker` and
  `odachi-verify`.
  - The worker is the bounded writer: it reads the task doc and the exact skill
    paths, edits only the allowed surfaces, runs the exact verification commands,
    and returns a small evidence report. It cannot commit, delegate or decide
    vision.
  - The verifier is read-only: it derives probes from the task doc, runs them in
    a scratch copy when they mutate state, and returns a verdict with `path:line`
    evidence. It cannot edit or delegate.
- **`slices` owns the handoff:** task doc path, step, allowed surfaces,
  verification commands, and the exact `SKILL.md` paths under
  `## Skills to load before work`. The return is bounded to ~2k tokens.
- **`doctor` compares the subagents with what the binary generates**, the same
  way it compares the primary mirror and the global door. `init` is the only
  writer; `level` and `start` do not touch them.
- **No registry and no always-on router.** The client's native skill catalog is
  still the index; the subagent files are generated only when the parent invokes
  them.

## Consequences

- The parent's context holds the task doc, the handoff and a ≤2k return, not the
  module or the full diff. The heavy reading and writing happen in the worker's
  context; the verification happens in a third one.
- The mechanism is native per client: OpenCode gets `mode: subagent` and a
  permission block; Claude Code gets `name`, `description` and, for the verifier,
  a tools allowlist without Edit or Write.
- Delegation is still **policy, not a gate**: the model has to follow `slices`.
  There is no token telemetry and no mechanical enforcement. A competing global
  skill can still be loaded; the precedence line in `slices` is the mitigation.
- The worker can lie like the parent; the parent verifies against `brief.md`, the
  diff and `check --run`, never against the worker's summary.
- A client without native subagents keeps working inline; odachi does not invent a
  subagent mechanism for it.
