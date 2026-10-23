---
name: go-tests
description: Trigger: when writing or reviewing Go tests, golden files or Bubbletea flows. Load `tests` for the discipline and this skill for the concrete pattern.
---

## When

When the change is Go and you have to decide which test to write, how to name it,
when to use `t.TempDir`, `testing.Short`, `teatest` or a golden file.

## Rules

- Load `tests` first: one test per behavior and watch it fail for the right
  reason. This skill does not repeat that discipline; it chooses the pattern.
- Table for several cases: `for _, tt := range cases { t.Run(tt.name, ...) }`.
  The case name says the scenario, not the input.
- Test behavior and state, not private details. If the test breaks on a refactor
  that does not change behavior, it is in the wrong place.
- Files and permissions: `t.TempDir()`; never the real home. A test that writes to
  `$HOME` fails on another machine.
- Slow integration or external commands: `testing.Short()` and a `t.Skip` with
  the reason and when it unblocks.
- Bubbletea: `Model.Update()` directly for state; `teatest` only for the
  interactive flow.
- Golden files: deterministic, updated through the repo's `-update` flag and run
  again without `-update` to prove the diff stayed.
- Mocks at the system boundary (command, clock, network), not in the middle of
  the logic.
- Run the focal package first and the full suite after; do not say "they pass"
  with one package.
- If the test fails, fix the cause. Do not adjust the test to make it pass.

## Never

- A test that depends on order, time or the network.
- `t.Skip` without a reason and an exit condition.
- Golden files updated by hand without reading the diff.
- Test getters, setters or private details.
- Invent data to make a case pass.
