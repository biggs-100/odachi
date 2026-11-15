# The shape of a project, by kind

The discipline is a menu, not a ritual. Pick the row before the first file, say
which row you are on in the plan, and drop what the row drops. A prototype
promoted to a product restarts at the product row.

| Project | Keep | Drop |
| --- | --- | --- |
| **Product app** (users, data, screens, money) | everything below | - |
| **Internal tool / CLI** | the boundary guard test, layered verification, one-page ADRs | the full hexagonal skeleton when there is no domain to isolate; the UI harness when there is no UI |
| **Library / package** | a public-API contract test, semver and ADR discipline, a guard on the dependencies | the composition root, the UI harness |
| **Pipeline / batch** | determinism (frozen inputs), artifact identity and atomic swap, loud schema drift | the UI harness; per-request concerns |
| **Prototype / spike** | one ADR saying "this is a prototype", the verification command | an ADR per decision, guard tests, review-load rules |

## What the shape is made of

When the row keeps it, the layout is the one this project already carries: a pure
core, ports as interfaces, adapters as implementations, and one entrypoint that
composes. The rules that keep it that way -dependencies point inward, one owner per
business rule, a boundary test that fails the build- live in `drive` and `tests`,
which every project reads: this file says what to build, those say how the project
keeps it.
