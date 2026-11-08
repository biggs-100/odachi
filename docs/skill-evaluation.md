# The behavioural measurement of a skill

The skills are the product: text an agent reads before it works. `embed_test.go` proves the text
still carries certain sentences; it does not prove that an agent, reading them, works differently.
This document records six attempts at measuring the difference, their results, and what they do not
prove. The measured skill is `tests`, with a neutral agent: first on a one-line fix, then on a new
function whose edges are the task, then on the same function delivered as a task-doc step, the shape
the project actually ships, then on that step with the skill removed, then on the step handed to a
delegated worker, and finally on the parent that hands it.

## The question

Does an agent, given the same task, behave differently when `tests` is available? The rubric reads
four behaviours the skill promises, from the transcript:

1. a test was written before the production code;
2. a red was observed before the fix;
3. the test's name says what breaks if it fails;
4. the fix is at the cause and the suite is green.

The first two probes ask whether the text changes the work when it is available; the third asks
whether the chain the project ships loads it at all; the fourth asks what the text adds when the chain
does; the fifth asks whether the worker receives it; the sixth asks whether the parent sends it.

## The instrument

`odachi bench` cannot do this: its contract says it never executes a command that is not `odachi`.
The instrument is `opencode run --format json`, whose event stream names every tool call in order —
the skill load, the file writes and the commands. That order is the evidence a string test cannot be.

Before measuring `tests`, the instrument was validated with a canary: a throwaway skill that orders
the agent to answer `BANANA` when asked to reply, placed first in the project and then in
`~/.agents/skills`. The transcript showed the load and the answer in both places:

```
TOOL skill  {"id":"canary"}
TEXT        BANANA
```

Two things were learned there: the transcript shows a load, so a missing load is meaningful; and
OpenCode reads the home copy, `~/.agents/skills`, not only the project's. The canary was removed
after the check and the home was verified at its baseline count (see the arms below).

## The first probe: the task

A Go program with one defect, copied fresh for every run:

```
go.mod              module task, go 1.22
main.go             fmt.Println(calc.Sum(2, 3))
calc/sum.go         func Sum(a, b int) int { return a - b }
```

The prompt is neutral and names the symptom, never the method: *"El programa imprime -1 cuando
debería imprimir 5 para 2+3. Arreglalo."*

## The first probe: the arms

Both run the same task, the same model (`opencode-go/deepseek-v4.1-flash`) and the neutral agent
(`--agent build`), in a fresh copy of the fixture, three times each.

```sh
# each arm, in a fresh copy of the fixture
opencode run --format json --agent build --model opencode-go/deepseek-v4.1-flash \
  "El programa imprime -1 cuando debería imprimir 5 para 2+3. Arreglalo."
```

- **With the skill**: the home copy is in place.
- **Without it**: the folder is moved out for the duration and moved back after. A temporary home was
  tried first and did not work: the model stops resolving (`provider.no-route`), so the auth cannot be
  carried into an isolated home. The move is reversible and it was measured at both ends:

```sh
Move-Item ~/.agents/skills/tests $TEMP/opencode/odachi/0232/tests-off   # before the runs
odachi doctor        # ~/.agents/skills: 58 installed; compared: 18 matches in skills/
Move-Item $TEMP/opencode/odachi/0232/tests-off ~/.agents/skills/tests   # after the runs
odachi doctor        # ~/.agents/skills: 59 installed; compared: 19 matches in skills/
```

The baseline before the move is the same 59. An interrupted repeat has the destination written here,
and `odachi doctor` is the check that says it came back.

The event stream is kept in the machine's temp folder; the clone keeps the result and the decisive
events, not the dumps.

## The first probe: the result

The decisive events of one run per arm, in order. Neither shows a `skill` load, and neither writes a
test file: the only write is the edit of `calc/sum.go`.

```
with-2     shell "Get-ChildItem -Force" · read main.go · read go.mod · read calc/sum.go
           edit calc/sum.go "return a - b" -> "return a + b" · shell "go run ."
without-2  shell "ls -Force" · read main.go · read go.mod · read calc/sum.go
           edit calc/sum.go "return a - b" -> "return a + b" · shell "go run ."
```

Six runs, three per arm. Each check is 1 when the transcript shows it and 0 when it does not; `-` is
not applicable because the run wrote no test.

| run | arm | loaded `tests` | 1 test before code | 2 red observed | 3 name says what breaks | 4 fix + suite | score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | with | 0 | 0 | 0 | - | 0 | 0/4 |
| 2 | with | 0 | 0 | 0 | - | 0 | 0/4 |
| 3 | with | 0 | 0 | 0 | - | 0 | 0/4 |
| 1 | without | - | 0 | 0 | - | 0 | 0/4 |
| 2 | without | - | 0 | 0 | - | 0 | 0/4 |
| 3 | without | - | 0 | 0 | - | 0 | 0/4 |

All six fixed the cause (`a + b`, and `go run .` prints 5), but none ran `go test`: the fourth check
asks for the cause and the suite, and only half was observed, so it scores 0. The table and the score
say the same thing: no check was exercised.

## The second probe: the task

The first probe may have sat below the threshold the skill names. The second asks for the moment its
`When` names first: code that does not exist yet. A Go program with one caller and an empty function,
copied fresh for every run:

```
go.mod             module task, go 1.22
main.go            prints config.Parse(os.Args[1]) as <first>-<last>; exits 1 on error
config/config.go   func Parse(s string) (Range, error) { return Range{}, errors.New("not implemented") }
```

The prompt is neutral and names the contract, never the method: two accepted shapes and five rejected
ones, and every rejection has to name what it read.

> El programa `task` recibe un puerto o un rango. `8000` tiene que imprimir `8000-8000` y `8000-8010`
> tiene que imprimir `8000-8010`. Implementá `config.Parse`; `main.go` ya la usa. Cualquier otra
> entrada —vacía, `8010-8000` invertido, `8000-`, `-8010`, `a-b`— tiene que fallar con un mensaje que
> nombre lo que se leyó.

## The second probe: the arms

Same model, same agent, three runs per arm, a fresh copy each time. One method change: the first
layout put the copies under one folder, and the agent listed `..` and read a sibling fixture; each run
now starts in its own `%TEMP%\d<random>` folder, whose parent holds no other run, and the contaminated
run was discarded. The transcript is the only thing kept; the copies die with the run.

The home copy is in place for the with arm and moved out for the without arm, measured at both ends:
`~/.agents/skills` went 59 → 58 → 59, and `odachi doctor` printed the count each time.

The rubric is the first probe's, with one translation. This task has no defect to fix at its cause, so
check 4 asks that the suite is green, and a separate column records whether the contract held —
verified after the runs by building each copy and passing it the seven shapes.

## The second probe: the result

The decisive events of every run, in order. No run shows a `skill` load, none writes a `_test.go`
file, and none runs `go test`. Every run implements the function and then runs the program against the
shapes of the prompt from a shell loop.

```
with-1     ls · read the three files · edit config.go · shell (gofmt, vet, go run over 8 shapes)
with-2     glob · read the three · write config.go · shell (gofmt -l, vet, build) · shell (8 shapes, exit code)
with-3     ls · read the three · write config.go · shell (build, 11 shapes) · shell (gofmt, vet)
without-1  ls · read the three · write config.go · shell (vet, build, 9 shapes)
without-2  find · read the three · write config.go · shell (gofmt, vet, build, 10 shapes) · shell (remove the exe)
without-3  ls · read the three · write config.go · shell (vet, build, 8 shapes)
```

| run | arm | loaded `tests` | 1 test before code | 2 red observed | 3 name says what breaks | 4 suite green | contract |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | with | 0 | 0 | 0 | - | 0 | 1 |
| 2 | with | 0 | 0 | 0 | - | 0 | 1 |
| 3 | with | 0 | 0 | 0 | - | 0 | 1 |
| 1 | without | - | 0 | 0 | - | 0 | 1 |
| 2 | without | - | 0 | 0 | - | 0 | 1 |
| 3 | without | - | 0 | 0 | - | 0 | 1 |

Every run scores 0/4, and the two arms are indistinguishable: read, write the function, list the
shapes by hand, build. The loop covers every shape the prompt names, in every run, so the behaviour
the skill asks for is present —the edges are exercised— while its artifact is not: there is no test
file and no red. The contract holds in all six, verified afterwards.

## The third probe: the task

The first two probes ran in a bare Go folder: no block, no plan, a symptom and a request. That is not
the shape odachi ships. The third runs the same function, inside a minimal odachi project: the block
in `AGENTS.md`, a `brief.md`, one task doc with one step, and the same Go code with its empty
`Parse`. The prompt is the parent's sentence, not the symptom:

> Trabajá el paso T1 del plan.

## The third probe: the arms

Both arms keep every skill installed and differ in one file: the with arm carries the block in
`AGENTS.md`, the control carries the same project without it. Three runs each, same model, same
agent, a fresh copy per run in its own isolated folder. The first run is both the pilot that
validated the fixture and run 1: nothing in the fixture changed after it.

The rubric is the second probe's, plus one column: whether `drive` and then `tests` were loaded, in
that order, read from the transcript's `skill` calls.

## The third probe: the result

The decisive events of one run per arm, in order.

```
block-1    list · read brief · read the task doc, main.go, config.go, go.mod
           skill drive · skill tests · write config_test.go · go test (RED)
           write config.go · go test (green) · the seven shapes · odachi check · close T1
control-1  glob plan · dir · read brief · read main.go, go.mod, config.go, the task doc
           skill drive · skill tests · write config_test.go · go test (RED)
           write config.go · go test (green) · the seven shapes
```

| run | arm | `drive` → `tests` | 1 test before code | 2 red observed | 3 name says what breaks | 4 suite + contract | score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | block | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 2 | block | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 3 | block | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 1 | control | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 2 | control | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 3 | control | 1 | 1 | 1 | 1 | 1 | 4/4 |

All six load `drive` and then `tests`, write the test before the implementation, watch it fail on the
empty function, and leave a suite that is green and a contract that holds -verified afterwards by
building each copy and passing it the seven shapes. The names state the behaviour, for example
`TestParseRejectsEveryOtherShapeAndNamesWhatItRead` and `TestParseRangeIsClosedInterval`. Run 1 went
further than the rest and closed T1 through the gate: `git init`, commits, the proof and
`plan --done`. That closure is recorded, not scored: the four checks are read from the transcript the
same way in every run.

The block made no difference. The three runs without `AGENTS.md` loaded the same two skills in the
same order, and the block arm did not follow the block's order either: it read the task doc first and
loaded `drive` after. What all six share is the step.

## The marginal probe: the task

The third probe had the skill in both arms, so it proved the chain loads, not that the text adds
anything. The fourth keeps the step and varies the one thing the third held fixed: the home copy of
`tests`. Both arms keep the block, the brief, the task doc and the same prompt, so `drive` still
orders the test in both; the question is what the `tests` text adds to that order.

## The marginal probe: the arms

Three runs per arm, same model, same agent, a fresh copy per run in its own isolated folder. The with
arm runs at the baseline, 59 installed; the without arm moves the home copy out, 58, and restores it
after, 59, measured at both ends with `odachi doctor`. The rubric is the third probe's, with one
reading fixed: check 3 looks at the test's own names, the top-level test and its subtests. Check 2 is
1 only when the failure is observed before the production code; a failure seen after it is not a red
for the check, the way the skill reads "the test first". The craft the skill promises is read from the
test file: whether the five rejected shapes appear and whether the error message is pinned.

## The marginal probe: the result

The decisive events of every run, in order.

```
with-1     skill drive · skill tests · write config_test.go · go test (RED) · write config.go · green
with-2     skill drive · skill tests · skill commits · write test · go test (RED) · write config.go · green
with-3     skill drive · skill tests · skill commits · write test · go test (RED) · write config.go · green
without-1  skill drive · skill commits · write config.go · write test · go test (FAIL, after the code) · fix · green
without-2  skill drive · skill go-tests · write test · go test (RED) · write config.go · green
without-3  no skill · write config.go · write test · go test (green, it never failed)
```

| run | arm | `tests` | 1 test before code | 2 red observed | 3 name says what breaks | 4 suite + contract | score |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | with | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 2 | with | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 3 | with | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 1 | without | 0 | 0 | 0 | 0 | 1 | 1/4 |
| 2 | without | 0 | 1 | 1 | 1 | 1 | 4/4 |
| 3 | without | 0 | 0 | 0 | 0 | 1 | 1/4 |

The with arm scores 12/12 and the without arm 6/12. Every with run wrote the test first and watched it
fail. One without run wrote the test after the implementation and only then saw the failure; one wrote
a test that passed the first time it ran -the case the skill names: a test that never failed proves
nothing- and one did it the right way after loading `go-tests`, whose description says to load
`tests` for the discipline. The edge coverage and the pinned message are the same in both arms: the
requirement names the five shapes, so the test covers them whether or not the skill is read.

The first two checks agree in every row, and that is a property of this corpus: they can differ only
when a test is written before the code and not run until after it, and no run did that. The second
column is not a second measurement of the same act; it catches the case the first cannot, a test
written first that still never fails before the code.

## The marginal probe: what it adds

- **The text has a marginal effect on this corpus.** Same step, same requirement, same model: the
  checks the skill names -the test first, the red observed, the name- move from 1/3 to 3/3 when the
  skill is installed. That is the first evidence the library's text changes the work, not only that
  it is read.
- **The effect is on the order and the naming, not on the edge coverage.** The prompt already names
  the five rejected shapes, and both arms test them. A task that did not name them would be the next
  place to look for that difference.
- **The chain without the skill is not deterministic.** One run loaded no skill at all, one
  substituted `go-tests`, one did it by hand. Three runs show the spread, not its distribution.

## The delegated probe: the task

The worker's own prompt says two things this probe can measure: load the skills the parent lists in
`## Skills to load before work`, by exact path, before touching code, and report `skills_loaded` when
it returns. The fifth probe hands it the same step as the fourth, in the same minimal project, with
and without that section. The parent's prompt is the only thing that changes.

## The delegated probe: the arms

`--agent odachi-worker`, same model, three runs per arm, a fresh copy per run in its own isolated
folder. The with arm's prompt carries:

```
## Skills to load before work

Read these exact files before reading, writing or testing:

- C:\Users\USER\.agents\skills\drive\SKILL.md
- C:\Users\USER\.agents\skills\tests\SKILL.md
```

The control is the same prompt without the section. Every skill stays installed in both arms. A load
is counted either way the worker can do it: the `skill` call or a `read` of the `SKILL.md` path, and
the transcript decides, not the report.

## The delegated probe: the result

The decisive events of one run per arm, in order.

```
with-1     read drive/SKILL.md · read tests/SKILL.md · read the task doc
           write config_test.go · go test (RED) · write config.go · green · skills_loaded: both paths
without-1  read the task doc · write config_test.go · go test (RED) · write config.go · green
           skills_loaded: none (the task doc has no section)
```

| run | arm | paths read | before the first edit | 1 test before code | 2 red observed | 3 name says what breaks | 4 suite + contract | report matches | score |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | with | 1 | 1 | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 2 | with | 1 | 1 | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 3 | with | 1 | 1 | 1 | 1 | 1 | 1 | 1 | 4/4 |
| 1 | without | 0 | - | 1 | 1 | 0 | 1 | 1 | 3/4 |
| 2 | without | 0 | - | 1 | 1 | 0 | 1 | 1 | 3/4 |
| 3 | without | 0 | - | 1 | 1 | 1 | 1 | 1 | 4/4 |

The contract holds on the worker's side: with the section, the worker read both paths before its first
edit in 3 of 3; without it, it read no `SKILL.md` at all in 3 of 3, though every skill was installed
and listed. It loads by the parent's paths, not by the client's list.

The order does not move: the worker's own prompt says "test-first", and all six wrote the test first
and watched it fail. What moves is the name: the with arm names the behaviour
(`TestParseRejectsShapesAndNamesTheInput`), the without arm keeps `TestParseAccept` and
`TestParseReject`, with one exception. The report is truthful in all six; in the without arm it says
"none", and places the missing section in the task doc, not in the parent's prompt.

## The parent-side probe: the task

The fifth probe proved the worker obeys the list; it did not prove the list is written. The sixth runs
the door -`--agent odachi`- on the same step, in the same project, with the step declared
`route:delegated` and `agent:odachi-worker`, and reads the call the parent makes.

## The parent-side probe: the arms

There is one arm. Three runs, same model, a fresh copy each, its own isolated folder; the prompt is
the parent's sentence: *Trabajá el paso T1 del plan.* The measurement is the `subagent` call in the
parent's transcript: which agent it names, whether the prompt carries `## Skills to load before work`,
and which paths it contains.

## The parent-side probe: the result

```
parent-1  read the task doc · read drive from the shell · skill tests · subagent odachi-worker: no section
parent-2  read the task doc · skill drive · skill tests · subagent ×2: neither carries the section
parent-3  read the task doc · skill drive · skill tests · skill commits · subagent: no section
```

| run | delegations | to `odachi-worker` | section | paths | the parent loaded |
| --- | --- | --- | --- | --- | --- |
| 1 | 1 | 1 | 0 | 0 | `drive` (read), `tests` (skill) |
| 2 | 2 | 2 | 0 | 0 | `drive` (skill), `tests` (skill) |
| 3 | 1 | 1 | 0 | 0 | `drive` (skill), `tests` (skill), `commits` (skill) |

The column counts both ways a parent can have the text: the `skill` call and a read of the skill's
`SKILL.md`.

Every delegation goes to the right worker, and none carries the section or a single `SKILL.md` path,
though the parent had loaded `drive` and `tests` itself in all three -by the `skill` call or by reading
the file. The rest of the policy's cargo is there: the task doc path, the requirement verbatim, the
surfaces and the method, down to the shape of the test to write.

## The parent-side probe: what it adds

- **The chain breaks at the parent.** The worker obeys the list, 3 of 3; the parent does not write it,
  none of 4 delegations, even after reading `drive`, which says to pass the skill paths.
- **The cost lands on the craft, not the order.** The worker's own prompt still orders test-first, so
  without the paths it writes the test and loses the naming (the fifth probe's 1 of 3).
- **The instructions are not missing.** `drive` and `slices/references/delegation.md` both say to pass
  the exact paths; the gap is between the text and the moment the `subagent` call is composed.
- One model, three runs, one fixture: the claim is about this corpus; a different step or model may
  behave differently.

## What it proves, and what it does not

- **The instrument repeats, and the chain fires on a step.** Twelve runs on a bare symptom never
  loaded anything (six with the skill installed, six without). Twelve runs on a task-doc step loaded
  `tests` in every run where it was installed, 9 of 9, and `drive` in 11 of 12; the one run that
  loaded nothing had no `tests` either. The absence in the first two probes was real, and it was
  about the task, not the text.
- **The block is not the lever.** The control arm has no `AGENTS.md` and behaves the same. The block's
  instruction to load `drive` was not even the order the block arm followed.
- **The task's shape is the lever the corpus names.** What the step runs share is the brief, the task
  doc and a prompt that names the step; what the twelve bare runs lack is all three at once.
- **It does not isolate the ingredient.** The step, the brief, the task doc and the word *plan* arrive
  together; which one matched is not measured.
- **It does not prove the test is better code.** The marginal probe shows the artifact arriving in the
  order the skill names; whether that test finds a defect the other would miss is a different
  question, and no run measured a defect found.
- **The worker honors the contract it is given.** With the section, the two paths were read before
  the first edit in 3 of 3; without it, no `SKILL.md` was read in 3 of 3, though every skill was
  installed and listed. The parent's list is the door; the client's list is not consulted by the
  worker.
- **The report matches the transcript.** `skills_loaded` names the two paths when they are read and
  `none` when they are not, in all six. When the section is missing it says the task doc has no such
  section, which places the gap one level down from where the parent writes it.
- **The order is the worker's, the craft is the skill's.** The worker's own prompt orders test-first,
  and all six obey; the name moves with the skill (3 of 3 against 1 of 3).
- **The parent does not pass the paths.** Four delegations to the right worker in three runs, none with
  the section or a path, though the parent had loaded `drive` and `tests` itself. The instruction is in
  the text; it does not survive into the call.
- One model, three runs per arm, one task: the claim is about this corpus.

## The next thing to edit: the moment of delegation

The instruction exists in `drive` and in the delegation policy, and the parent read it. What fails is
the moment the `subagent` call is written: the prompt is composed from the step, and the skills
section is dropped. Two candidates, and the human decides:

- **The worker makes the omission visible.** It returns `status: blocked` when the section is absent
  and names it; the parent then either passes the paths or writes a section that says there are none.
  It closes the gap in the same turn, and it adds a line to every delegation that owes no skill.
- **The door carries the section.** A line in the parent's own prompt makes the section part of every
  `subagent` call. It costs the worker nothing and repeats an instruction the parent already ignored
  once.

Either one reopens this probe with the same three runs. The description edit is dropped: this corpus
gives it no support. The trigger fired when the work looked like a step, and the sentence was never
the variable.
