# the-first-behavioural-measurement

## What must exist
- a skill's effect is measured once and the method can be repeated: `docs/skill-evaluation.md` records a task, an agent run with the `tests` skill and one without it, an observable rubric read from both transcripts, and the result, so the library's first claim -that the text changes the work- is evidence instead of belief

## Done when
- the document carries the task, the commands of both arms, the rubric and the result table, with three runs per arm
- the rubric's four checks are readable from the transcript: the test before the code, the red observed, the name saying what breaks, the fix at the cause with the suite green
- the without arm hides the home copy of `tests` for its runs and the document proves the restore: `odachi doctor` names the installed count again
- the document names what the result does and does not prove: one model, three runs, one task
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- no changelog entry: nothing a person installs changed

## Not included (first version)
- no change to `odachi bench` and no new verb: it cannot run an agent by contract; it reopens when the measurement is repeated often enough to automate
- no change to the skills: the probe measures `tests` as it is; it reopens when the result asks for an edit
- no CI and no test that runs an agent: it spends tokens and is not deterministic; it reopens when a model judge replaces the person reading
- no change to the person's home beyond the temporary move, restored and checked; it reopens if a run leaves it moved
- no change to the brief's format or to `start`
- no change to the review, the doctor or the classifier
- no change to the stored proofs: never touched
- no change to the design canvas, its token or the installers
- no change to the release workflow

## Decisions taken
- the baseline is a neutral agent (`build`) with and without the skill: the odachi agent already carries the method, and that run would measure the marginal effect instead of the teaching
- the task is a Go program with one defect and a neutral request: enough for a test to be the natural first move, small enough for the fix to fit one line
- the rubric is read from the `--format json` transcript: it shows the skill call, the file writes and the commands in order, which is the evidence a string test cannot be
- three runs per arm: one run is the model's mood, not the skill's effect
- the home copy is moved out and back for the without arm: a temporary home breaks the model auth (measured), and the move is reversible with a restore check
- the document keeps the decisive events quoted, not the whole transcripts: the commands reproduce them, and the clone keeps evidence, not dumps
