---
name: arch-project
description: Trigger: giving a codebase its shape - the kind of project, the decisions before the first file, the boundary the build enforces. The second door, under bootstrap and above arch-<stack>.
---

## When

A codebase is being shaped: it is starting, or its boundary is not where it should
be. `bootstrap` opens, this is the second door, and `arch-tauri` follows for Tauri v2.

## Rules

- **The shape is a menu, not a ritual**: pick the row of `references/project-shape.md` before the first file, name it in the plan, and drop what the row drops.
- **Dependencies point inward**: the domain core imports the standard library, its own packages and the ports; a framework, a driver or a database lives in an adapter.
- **One owner per business rule**: every rule lives in exactly one domain place and is never copied into a handler, an adapter or the UI.
- **The boundary is enforced mechanically**: the guard of `references/guard-recipes.md` is written before the business code, and a boundary with no test that fails the build is a comment.
- **The twelve decisions of `references/birth-decisions.md` exist before the first file**: an unanswered line is a decision taken by accident.
- **Every non-trivial decision gets an ADR** in `assets/adr-template.md`: context, decision, alternatives, consequences and how it stays enforced, because a decision without its reason is a future bug.
- **A change runs its layers in order**: the focused tests of the touched package, the whole suite, build and lint, then a short runtime probe, and no layer is skipped because the change looked small.
- **A project with a screen is not scaffolded until it is verifiable**: `references/ui-verification.md` is the contract an agent can run, and a screenshot nobody asserts on is not evidence.

## Decision gates

| Situation | What to read |
| --- | --- |
| A project is starting and the first file is next | `references/project-shape.md`: the kind of project, and what each kind drops |
| The decisions that must exist before code | `references/birth-decisions.md`: the twelve, and why none of them can wait |
| The tree the chosen row keeps | `references/layout.md`: the pure core, the ports, the adapters, and the one entrypoint that composes |
| The codebase needs a boundary the build enforces | `references/guard-recipes.md`: the guard per stack, and what does not work as one |
| A change is about to touch a package | `references/stack-commands.md`: the commands of this stack, and the four questions before them |
| The project has a screen | `references/ui-verification.md`: drivable, observable, inventoried, one command |
| A rule feels arbitrary | `references/engineering-lessons.md`: the defect each one was paid for |
| The stack is Go with server-rendered HTML | `references/go-htmx.md`: a stable htmx, templ, and the router inside the adapter |
| The stack is Tauri v2 | Load `arch-tauri` after this one: that stack's boundary is its own door |

## Output

The shape, in this order: the triage row, the tree, the guard test path, the
twelve decisions answered, the first ADR, and the one command that proves the
project end to end. A project with a screen that skipped its UI layer is
reported as scaffolded and not verified.

## Never

- Never restate the rules `drive`, `tests` and `commits` own: this door says what to build, those say how the project keeps it.
- Never pick the shape by habit: a prototype promoted to a product restarts at the product row.
- Never write business code before the guard test of its boundary.
- Never copy a rule into an adapter, a handler or the UI because it was closer.
- Never load this for a one-off script, a single-file utility, or feature work in a project that already has its shape: `drive` is the day-to-day manual.

## References

- `references/project-shape.md` - the triage by project kind, and what each kind drops.
- `references/birth-decisions.md` - the twelve decisions that must exist before the first file.
- `references/layout.md` - the tree every row that keeps it shares, and the mapping per stack.
- `references/guard-recipes.md` - the boundary guard per stack, and what does not work as one.
- `references/stack-commands.md` - the commands of each stack, and the four questions before a change.
- `references/ui-verification.md` - the UI contract an agent can actually run.
- `references/engineering-lessons.md` - the fifteen rules, each with the defect that paid for it.
- `references/go-htmx.md` - a Go project with server-rendered HTML: htmx, templ, the router.

## Assets

- `assets/adr-template.md` - the ADR the decisions and the boundary choices are written with.
