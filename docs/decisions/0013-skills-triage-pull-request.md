# 0013 — Triage and pull-request skills

Status: accepted · Date: 2026-10 · Session: v0.11.2

## Context

The comparison with gentle-ai showed two real gaps in odachi's catalog: attacking
a bug backlog and cutting a delivery into reviewable PRs. Gentle-ai solves them
with `systemic-issue-triage`, `issue-root-resolution`, `branch-pr`, `chained-pr`
and `work-unit-commits`. Odachi had `commits` for the message and `tests` for the
reproduction, but no skill that decided **what one fix is** or **how a PR is
split**.

Copying those skills whole does not fit: odachi's format requires `When / Rules /
Never`, a `Trigger:` in the description, a 60-line body and a name of at least
three characters. Nor does it make sense to copy ODD's grammar, GitHub labels or
RDD machinery: odachi decided not to have them (ADR 0008).

## Decision

- **`triage`**: classifies each report by root (A–E), joins those sharing a cause
  into one step, reproduces the symptom before accepting the diagnosis and closes
  with the failing test. A report outside the fence goes to `odachi ask`; several
  roots are several `odachi plan` steps.
- **`pull-request`**: one work unit per PR, a 400-changed-line budget, one chain
  strategy (stacked or feature branch chain), approved issue or `Refs`, and
  `odachi check --run` before review. The slug is `pull-request` because the
  format requires three characters; the trigger still says "PR".
- **Adapt the method, not the grammar.** There is no `Specs/Tasks/Log`, no
  `status:approved` labels and no RDD receipts. Both skills use the verbs that
  already exist: `ask`, `plan`, `tests`, `check --run`, `commits`.
- **Install them like the rest of the curriculum.** `init` copies them to
  `./.agents/skills`; `bootstrap` stays in `~/.agents/skills`; `doctor` counts the
  copies and `skills --check` validates the format.
- **Each one has a content test** in `skills/embed_test.go`: if someone deletes
  the root rule or the 400-line budget, the test fails.

## Not taken

- ODD's grammar and its Engram mirror.
- GitHub labels and issue-first as an obligation: the skill says what to do if the
  repo uses issues; it does not invent an approval.
- RDD, judgment-day and receipts: odachi already has `review` with lenses and a
  refuter.
- `collab-perfect`, `cognitive-doc-design`, `comment-writer`, `go-testing` and
  `gentle-ai-bench`: they are specific to gentle-ai or cover a flow odachi does
  not have yet. If they become a need, they get their own feature.

## Consequences

- The catalog goes from ten to twelve skills; the project installs eleven and the
  home one.
- A backlog and a delivery have an explicit method without adding verbs or flags.
- The context cost grows a little: the two skills add 78 lines and stay under the
  60-line cap each.
- The comparison with gentle-ai is documented: the method was taken, not the
  machinery. The next skill that does not fit the cap goes to a document, not to a
  longer body.
