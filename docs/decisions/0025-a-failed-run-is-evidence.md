# 0025 - A failed run is evidence, not a proof

Status: accepted · Date: 2026-10 · Session: after v0.11.9

## Context

`check --run` captured both streams and kept the output only when the command
exited 0. A failed run returned the last non-empty line, and the rest was
discarded: a red Windows job said `FAIL` and nothing else. The proof artifact
that CI uploads "even when the gate fails" did not exist on red, because odachi
wrote no file before exiting nonzero. The wording in the workflow and the ADR
said the log was the evidence, and the log was one line.

A proof is the real output of a green run, and that definition is worth keeping:
an artifact that says "this failed" is not a proof of anything passing. But a
failure is still evidence, and the person fixing it needs the failing test, the
assertion and the build error, not the last line of them.

## Decision

- **A failed run writes `proof/NNNN-<slug>.failed.txt`.** Same tree and command
  headers, an optional fence ref, then `# exit <code>` or `# timed out`, then the
  output verbatim, stdout and stderr together.
- **The name and the third line say what it is.** The green proof keeps its name,
  its format and its bytes; the parser still reads only `<slug>.txt`, so nothing
  can mistake a failure for a proof.
- **The gate prints the last line and the path.** The message stays one line (a
  terminal is one line), the path follows it, and `next:` stays last. The exit
  code is unchanged: the verdict is the gate's, not the command's.
- **Best effort, never a false receipt.** If the file cannot be written, the path
  is not printed. Silence about a missing artifact is better than a path that
  points at nothing.
- **CI needs no change.** The gate workflow uploads `proof/` on `always()`, so the
  failed file appears in the artifact of the red run.

## Alternatives rejected

- **Print the whole output to the terminal.** It buries the one line that says
  what broke, and CI terminals are read backwards from the end.
- **Truncate or diff the output.** A truncated failure hides the assertion; a
  diff is a second format to read and to keep.
- **Write the failure under the proof name.** The proof would stop meaning "a
  green run", and the gate would have to learn to tell two things apart that
  today cannot be confused.
- **Exit with the command's code.** The gate's exit codes are the contract with
  the hook and CI (0 ready, 1 act, 2 human); a test's code is not one of them.

## Consequences

- A red CI job now has an artifact with the whole failure; the workflow already
  uploaded `proof/`, so the only change is that there is something to upload.
- `proof/` holds at most one failed file per failed run: the next run of the same
  feature overwrites the name, and the file carries the tree, so a stale one is
  never mistaken for the current failure by the tool.
- The green proof bytes are untouched, and a green run writes no failed file.
- The local terminal still shows one line plus the path; a human who wants the
  detail opens the file, and an agent can read it without rerunning the command.

## Sources

- Feature 0026's review: the workflow claimed a proof on red that odachi never
  wrote.
- Feature 0029: the failed file and the line that names it.
