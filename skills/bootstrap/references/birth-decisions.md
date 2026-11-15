# The decisions that must exist before the first file

Fill every line of this list in the project's own document -`docs/birth-decisions.md`
is a good name for it- and commit it with the first code. An unanswered line is a
decision taken by accident, and the ones taken by accident are the expensive ones.

| # | Decision | Why it must exist now |
| --- | --- | --- |
| 1 | Language of the artifacts: code, commits, UI copy, documents -one answer | Mixing them later costs a migration of every string |
| 2 | Runtime envelope: offline-first? single binary? recurring cost? platform-specific parts? | It decides the architecture, not a feature |
| 3 | The two validations: write path and read path are not the same rule | The split is cheap now and a data migration later |
| 4 | Which effects are irreversible, and how each one is identified | Retrofitting idempotency after a duplicate is a postmortem |
| 5 | Background work: who runs it, when, what the operator sees, what happens if the process closes mid-work | It appears in every project with a poller or a sync |
| 6 | Derived artifacts: their identity, their build location, their atomic swap | Caches and indexes are born early |
| 7 | Screens: the route and state inventory | "We tested the UI" needs a denominator |
| 8 | The human gates, named: real browser, real device, real credential, provider sandbox, taste | They are the only checks no agent closes |
| 9 | Review load: the size metric without generated code, the slice size, the chaining | The scarce resource is the reviewer |
| 10 | Documents: the shape of specs, tasks and log, and where the residual is named | Institutions outlive sessions |
| 11 | What the platform refuses to do, instead of degrading into silence | Silent degradation is worse than an error |
| 12 | The one command that proves the project works end to end | It is the daily loop and the deliverable's receipt |

Rows 4, 5 and 6 come with their own defect record in `engineering-lessons.md`: the
rules exist because something broke.
