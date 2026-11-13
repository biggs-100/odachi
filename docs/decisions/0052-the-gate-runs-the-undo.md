# 0052 - The gate runs the undo, and nothing else

Status: accepted - Date: 2026-11 - Session: after 0289

## Context

0290 asks the gate to measure whether a test can fail, not only whether a command
passed: if the suite stays green with the change undone, the tree does not close.
That row needs the gate to run one command of its own, and until today a fence
rule said the opposite in one sentence: *the gate does not run the checks, it
reads the verdict they produced*.

The rule was right for what it protected. The gate reading verdicts is what keeps
the close offline, cheap and repeatable: lint and `govulncheck` run in the review
and their exit status travels as a row, so the gate never installs, never reaches
the network, and never becomes a second runner with its own opinions. The undo row
is the one case where reading is not enough, because the evidence it needs is a
command run against a tree that does not exist yet.

## Decision

- **The undo reverts the production Go files of the feature's diff, and only
  those.** `*.go` that is not `_test.go`, no `testdata/`, no `go.mod` or `go.sum`.
  Reverting the tests removes the witness and the suite goes green for the wrong
  reason; reverting the module turns a behavior into a dependency.
- **A build failure is not a test that noticed, and it does not block.** The row
  separates `--- FAIL` from `build failed`. A copy that does not build is reported
  by name as a note -never silently, never as a green- and the tree may close.
  Mutation testing settled the first half: a mutant that does not compile is
  discarded, not killed. The second half is ours: blocking on it would stop every
  change that adds a symbol its tests call, which is most of them, and a gate that
  blocks most closes teaches people to waive. What the row blocks is the one thing
  the requirement names: a suite that stayed green.
- **The undo runs the same command the proof ran.** One command, one source. Two
  commands would let the actor pick a strong one for the proof and a weak one for
  the undo, and the row would measure nothing.
- **The second run happens on a copy outside the worktree, in parallel when the
  machine allows it.** Serial, the close goes from about twelve minutes to
  twenty-four on this machine, and a gate nobody runs is a gate nobody obeys.
- **The evidence has its own file and its own first line.** `# undone <tree>`,
  beside the command and the output verbatim. `Parse` refuses anything that does
  not start with `# tree`, so a stored undo can never be read as a proof, the way
  `# failed` already cannot (0025).
- **A change with nothing to undo says so by name.** A documents-only or
  tests-only tree reports that it adds no production behavior; the row is not
  silently absent, because silence is not exculpation (0033).
- **A refactor is exempted by the human, not inferred.** `start --waive-undo
  --because "<why>"` records the declaration in the task doc, visible in the diff,
  the way `--waive-review` already does. A suite that stays green because the
  behavior did not change is indistinguishable from a suite that does not notice,
  and the tool does not guess which one it is: the human says so, with a reason a
  reviewer reads.
- **The rule that closed `internal/domain/gate/**` retires with this feature.**
  The boundary it protected is now narrower and is written here instead: the gate
  runs the undo and nothing else -no linter, no lens, no vulnerability check-.

## Alternatives rejected

- **Revert the whole diff.** The tests the feature added disappear with the code,
  and a suite with fewer tests passes more easily. The row would reward deleting
  the witness.
- **Count a build failure as the test noticing.** It is the cheapest way to a
  green and it proves only that a test names a symbol the old tree lacks.
- **Run a narrower command the second time.** It saves minutes and hands the
  actor the choice of instrument, which is the defect the row exists to close.
- **Keep the fence rule and add an exception.** A rule with a hole in it is a
  comment with a parser; the fence protects paths or it protects nothing (0027).
- **Make the undo a review lens.** A lens answers from the prompt and runs
  nothing; this row is a command and its exit status, and it belongs where the
  proof lives.

## Consequences

- The close costs a second suite run. The parallel copy keeps wall time near one
  run; on a machine that cannot run two, the cost is named instead of hidden.
- The row promises sensitivity, not correctness, and only when it can see. A test
  that fails on the old tree only because it asserts a new symbol exists passes
  it, and a change whose reverted copy does not build is a note, not a proof. The
  reliability lens is the second reader, and this paragraph is the limit.
- `proof/` gains a third kind of file and no second grammar: the header is the
  same three fields, and the first line is what tells the kinds apart.
- `internal/domain/gate/**` is no longer closed by a rule. The next change to it
  is judged by the brief's non-goals like any other package, and the sentence that
  used to guard it now lives here, where a person reads it.
