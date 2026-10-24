# 0018 — The reviewer inherits the model of the session

Status: accepted · Date: 2026-10 · Session: v0.11.5

## Context

The human asked for one thing: *"the reviewer model should be the same model the
user is using, by default, unless the user decides to change it."* The earlier
investigation concluded it was impossible: no `OPENCODE_MODEL` variable exists,
`opencode.json` carries no model, and a child process cannot see which model its
parent session chose. That conclusion was half right. The session model is not in
the environment, but OpenCode exposes it through a supported command:
`opencode session export <id>` prints the session's `info.model` as
`providerID`, `id` and `variant`, and the host exports `OPENCODE_SESSION_ID` to
every child.

The gap was never the default path. When `odachi review` prints its prompt and
the agent answers it, the reviewer *is* the session's model by construction. The
gap is `--exec`: a human who writes `opencode run --model ...` or any other
reviewer CLI has to name the model by hand, and a human who forgets gets
whatever default that CLI picked. A verdict from a model nobody chose is a review
nobody asked for.

## Decision

- **The default does not move.** The bare form still prints the prompt for the
  agent in the session to answer, and that agent is the model the human is
  talking to. Nothing about it needs detecting.
- **`{model}` is a placeholder in the command, not a flag odachi invents.** The
  human keeps writing the reviewer command they know; where the identity goes,
  they write `{model}`: `odachi review --exec "opencode run --model {model}"`.
  odachi does not need to know how each CLI spells its model flag.
- **Resolution order: `--model`, then `ODACHI_MODEL`, then the host.** `--model`
  is the human's override for one run. `ODACHI_MODEL` is the host contract: any
  agent, plugin or shell can set a name without odachi knowing which client it
  is. Only when both are empty does odachi ask OpenCode, and only through
  `opencode session export` — a supported command, never its database or logs.
- **No fallback. A placeholder with no model stops the run.** The verb exits
  blocked, names `--model`, and does not start the reviewer. A substituted
  default would put a model in the receipt that the human never chose, which is
  exactly the defect this decision exists to prevent.
- **A flag that changes nothing is refused.** `--model` without `{model}`, or
  without `--exec`, is named instead of accepted and silently dropped.
- **The receipt records what ran.** The verdict's `reviewer` field is the
  command after substitution, not the template: the audit trail must name the
  identity that judged, or it is not an audit trail.

## Alternatives rejected

- **Reading the OpenCode SQLite database.** It holds the answer today and owes
  odachi nothing tomorrow. A private schema turns an upgrade into a broken
  binary; the CLI is a contract.
- **A fallback model when detection fails.** The reviewer would run with an
  identity the human did not choose, and the receipt would hide it. Failing with
  the remedy named is cheaper than a verdict of unknown provenance.
- **Injecting `--model` into the command automatically.** odachi would have to
  know every reviewer CLI's spelling, and the day one changes, the command
  breaks. The placeholder keeps the knowledge in the human's command.
- **Persisting the model in `.odachi/review.toml`.** The session model changes
  with the session; a stored name goes stale and becomes a lie. The override is
  one run (`--model`) or the environment (`ODACHI_MODEL`), never a file.
- **Detecting through `opencode models` or config parsing.** They list what is
  available or configured, not what this session is using.

## Consequences

- `odachi review --exec "opencode run --model {model} ..."` judges with the
  session's model without the human typing it, and the same holds for any CLI
  that takes a model flag.
- A command without `{model}` runs byte-identical to before, so every existing
  reviewer keeps working.
- When the host cannot answer (another agent, a detached shell, a broken probe),
  the message names the two ways out: `--model` and `ODACHI_MODEL`.
- The resolution lives in two places by design: the order and the refusal in
  `internal/app/review`, the host probe in `internal/adapters/hostmodel`. The
  domain stays stdlib-only and knows nothing about models.
- No new dependency: the probe is `os/exec` over the host's own CLI.

## Sources

- The human's requirement, quoted above.
- ADR 0010 (review as a human decision), ADR 0016 (the verdict's receipts).
- `opencode session export <id>`: `info.model` with `providerID`, `id`, `variant`;
  `OPENCODE_SESSION_ID` in the child's environment.
