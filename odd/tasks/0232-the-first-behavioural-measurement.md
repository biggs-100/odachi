# the-first-behavioural-measurement
Base: f363f630730745b617142ff858a4faf03b4eff77

## Requirements (verbatim, from the user)
R1 · "a skill's effect is measured once and the method can be repeated: `docs/skill-evaluation.md` records a task, an agent run with the `tests` skill and one without it, an observable rubric read from both transcripts, and the result, so the library's first claim -that the text changes the work- is evidence instead of belief"

## Done when
- the document carries the task, the commands of both arms, the rubric and the result table, with three runs per arm
- the rubric's four checks are readable from the transcript: the test before the code, the red observed, the name saying what breaks, the fix at the cause with the suite green
- the without arm hides the home copy of `tests` for its runs and the document proves the restore: `odachi doctor` names the installed count again
- the document names what the result does and does not prove: one model, three runs, one task
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- no changelog entry: nothing a person installs changed

## Steps
- [ ] T1 · the first measurement of tests is recorded   R1 · route:inline

## History
L1 · "a skill's effect is measured once and the method can be repeated: `docs/skill-evaluation.md` records a task, an agent run with the `tests` skill and one without it, an observable rubric read from both transcripts, and the result, so the library's first claim -that the text changes the work- is evidence instead of belief"
