# Next session: the installer's first line, and a reviewer that stalled

Written at the close of 0188. `main` is clean, v0.14.7 is published and installed, and `doctor` is
14/14 once this note lands.

## What shipped

- **0188, v0.14.7**: a folder past the walk is the class's shape - counted and named, not a read
  failure -, a class `--apply` empties is removed as a folder so the subtree the walk never reached
  goes with it, and the report says the switch is blocked instead of promising bytes it will refuse
  to free. The real run emptied this clone's store: 35.0 MB in 1376 files became 944 KB in 782, and
  34.1 MB in 595 files were freed, including the eight folders beyond the walk.

## The lessons this session paid for

**A safety rule that blocks the switch it protects has stopped being safe.** The depth bound and
"a run that cannot see does not delete" were each right alone; together they made the only command
that could clean the scratch folder refuse to run. The fix separated two facts one word was hiding:
a read that failed, and a folder outside the declared shape.

**Deleting a tree it could not measure is the honest half of a switch that deletes trees.** The
unknown part is named - "8 folders beyond the shape went with them" - instead of counted as
removed.

**A model that answers a probe can stall on the real prompt.** `mimo-v2.6-flash` answered a small
message in seconds and timed out three times at 600s on the 40 KB review prompt; `claude-haiku-5-5`
answered it in seconds. The verdict records the one-run `--model` override, and the project's
configured model is unchanged.

## What was measured, so nobody re-measures it

- `go test ./...` green on the closing tree; the proof is `proof/0188-the-cleanup-switch.txt`.
- The store before and after: 35.0 MB in 1376 files, then 944 KB in 782; `clean --apply` emptied
  one folder and freed 34.1 MB in 595 files.
- The reviewer stall: three 600s timeouts on `mimo-v2.6-flash`, then an answer from
  `claude-haiku-5-5` for the same tree.
- The documented `irm ... | iex` install still prints `?#Requires` because of the script's BOM;
  installing the downloaded file directly is silent. v0.14.7 was installed that way.

## Queued

1. **`install.ps1` starts with an error when piped.** The documented `irm | iex` reads the UTF-8
   BOM as a character, so the first line is `?#Requires`. A project's documented install teaching
   the reader to ignore errors is a defect of teaching; strip the BOM or serve the script without
   it, and pin the documented command.
2. **The review debt row.** `prunePrompts` has two user-visible failure branches no test exercises,
   and the fake's single `err` field cannot express "the listing answers and the removal fails";
   the next change to the sweep pays it with `failRemove` and the two tests.
3. **The reviewer model.** `odd/review.toml` still names `mimo-v2.6-flash` for the four lenses; the
   stall may be transient, and the model was chosen for cost. Changing it is the project's decision
   and a change of its own - a feature must not edit that file (0188's fence said so).

## What is already true, so it is not re-decided

- A stored verdict obeys `keep_prompts` at write time; the close names the store's size; `proof/`
  and `odd/tasks` are never bounded (0187, ADR 0049).
- Depth is the class's shape; a class `--apply` empties goes as a folder; a real read failure still
  blocks every removal; the report and the close ask one list (0188).
- The store is small and the switch works: `odachi clean --apply` runs on this clone.
- `main` is protected: four checks, everything through a pull request.
- **v0.14.7 is published, signed and installed**; the home carries the agents and the skills.

## How to find the next real work

`docs/review-debt.md` has one open row (the two untested failure branches of the sweep). The
installer's BOM is the smallest real defect on the board.
