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
- Verification is proportional to the blast radius. A passive change -a comment, a
  document, a rename with no call site- is verified by reading the diff: it owes no
  test.
- A medium change alters behavior and carries its test: the writer's own
  verification, the RED observed and the suite green, is enough.
- A high change names one of the six risks in `commits` and adds an independent verifier:
  `odachi-verify` reads the whole task doc and answers per requirement.
- Before a red is blamed on the change, run the same test on the base: stash the
  work or check out the base tree. A test that already failed there is named as
  pre-existing, with its output, and the change is judged on its own red.
- A red that comes from the machine is named, not hidden: say the condition (the
  platform, a missing tool, a `PATH` the doctor measures) and what it broke. It is
  never reported as green and never turned into a `t.Skip` without the reason and
  the condition that unblocks it.

## Output

The failing test first, its name saying what breaks if it fails, and the edges the change covers; then the fix and the suites that were run. The report names the tier the change owed and who verified it. A test that cannot be written is named with the reason, and the report says which packages ran.

## Never

- Adjust the test to make it pass without changing the cause of the failure.
- Test private details that prevent refactoring tomorrow.
- Tests that depend on execution order, time or the network.
- A `t.Skip` without saying why and when it unblocks.
- Say "the tests pass" when you only ran the package you touched.
