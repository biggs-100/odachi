# Birth decisions

The twelve answers this project was born with. They are decisions, not
aspirations: changing one is a human decision, and the tool must not quietly
drift away from them.

## 1. Languages

Code, identifiers, comments, docs, commits = English. Text the USER reads (CLI
messages, `brief.md`, task docs) = Spanish.

## 2. Runtime envelope

One static Go binary, offline, no daemon, no network, no telemetry, no cloud.

## 3. Two validations

The WRITE path (profile/fence from human input) is lenient and fails with a
human sentence naming the fix; the READ path (state on disk) is strict and fails
closed. They are separate functions.

## 4. Irreversible effects

Writing user files (`brief.md`, task doc, proof, hook, `AGENTS.md` block). Each
write = backup + expected-hash check + atomic swap (temp file in the same dir +
rename).

## 5. Background work

None. No poller, no worker, no cron. State it explicitly.

## 6. Derived artifacts

The rendered `AGENTS.md` block and `.odachi/state.json` carry source + digest +
version and are swapped atomically; a shape change fails loudly.

## 7. Screens

None. CLI only; no UI harness.

## 8. Human gates (an agent never closes them)

A non-coder reaches "it works" in 10 minutes; an experienced dev uses it for a
week without it getting in the way; review quality verdicts are never asserted
by the tool.

## 9. Review load

Work units of about 400 authored lines; tests and docs travel with the behavior.

## 10. Documents

`brief.md` (vision, durable) + `odd/tasks/NNNN-<slug>.md` (R#/Tasks/Historial).
The residual/unknown is named in the task doc.

## 11. What the platform REFUSES

`check` refuses Ready without proof for the exact tree hash; the profile schema
refuses any field that could disable a gate; the CLI refuses to raise a level
without human consent.

## 12. The one command that proves the project works end to end

`go test ./... && go build ./...` (today); later `odachi check` on a real repo.
