---
name: tests
description: Trigger: when writing, fixing or deciding what a change tests. The failing test first, the code after.
---

## When

When you are about to write new code, when a test fails and you do not know why,
or when someone asks "is this tested?".

## Rules

- The test first, and watch it fail for the right reason. A test that never failed
  proves nothing: it proves the code was already written.
- One test per behavior. The name says what breaks if it fails, not which function
  it calls.
- Test the edge that hurts: empty, one, many, the error, the half-written file,
  CRLF, permission denied. That is where defects live.
- When the message the human reads is the contract (an exit code, a line in
  Spanish), the test pins that message.
- Domain tests do not touch disk, network or clock. If they do, the test is slow
  and starts failing for reasons unrelated to the change.
- A failing test is fixed at the cause. If the test was badly written, say so in
  the report and explain why.

## Never

- Adjust the test to make it pass without changing the cause of the failure.
- Test private details that prevent refactoring tomorrow.
- Tests that depend on execution order, time or the network.
- A `t.Skip` without saying why and when it unblocks.
- Say "the tests pass" when you only ran the package you touched.
