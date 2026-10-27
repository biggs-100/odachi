# 0008 · ODD-lite: the discipline without the grammar

Status: accepted · Date: 2026-10 · Session: v0.11.0 · Name superseded by 0031 (the decisions stand)

## Context

`gentle-ai` runs ODD (Organic Driven Development): a feature document with
`## Specs` / `## Tasks` / `## Log`, task-size classification, a risk line,
test-first and verification tiers. Odachi does not. Its task document has four
different headers (`## Requisitos`, `## Listo cuando`, `## Pasos`,
`## Historial`), and `summary`/`plan`/`level` only recognize
`odd/tasks/NNNN-<slug>.md`. A gentle-ai feature document is invisible to the
verbs.

The comparison left three options: (A) teach the parser the ODD grammar, (B) keep
the grammar and add the discipline, or (C) accept read-only compatibility. The
user asked for B now.

## Decision

- **The grammar stays.** No `## Specs` / `## Tasks` / `## Log` parser, no second
  document format, no migration.
- **The discipline goes into the manual the agent already loads.** `drive` gains
  four rules:
  - requirements quote the user **verbatim** and every `Listo cuando` is
    observable; a missing criterion is one `ask`, not an invention;
  - before a step, load `tests` and observe RED; `check --run` runs with GREEN;
  - a delegated step is recorded with
    `ask --craft "paso delegado: ..." --because "..."`;
  - the commit closes with the `Risk:` line owned by `commits`.
- **`commits` owns the risk vocabulary**: `Risk: none` or one of
  `datos | seguridad | contrato | concurrencia | entrega | sin-test`. A named risk
  asks for `review` before the commit.
- **Nothing new to parse.** The risk line lives in the commit body; the route
  record lives in `## Historial` through `ask --craft`; the spec and the
  acceptance criteria are the existing `## Requisitos` and `## Listo cuando`.

## Not taken

- The ODD feature-document grammar and its Engram mirror.
- Task-size classification and a route column per task: the task doc has no field
  for them, and adding one is a grammar change, not content.
- Independent verifier tiers and the delivery/PR chain: they need evidence the
  current artifacts do not carry.
- A second workflow next to odachi's. One cycle, with more discipline.

## Consequences

- The task document stays readable, the gate is untouched, and no existing
  document breaks.
- The risk line is **policy, not a parsed field**: `summary` and `check` do not
  enforce it. The parent owns it, and the commit is the record.
- Interop with a gentle-ai ODD document is still read-only. If that becomes a
  requirement, the parser compatibility is its own feature, with a migration and
  the `Unaccounted` guard tests.
- The `drive` body is capped at 60 lines; the next discipline that does not fit
  goes to a new skill, not into a growing manual.
