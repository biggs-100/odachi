# The behavioural measurement of a skill

The skills are the product: text an agent reads before it works. `embed_test.go` proves the text
still carries certain sentences; it does not prove that an agent, reading them, works differently.
This document records two attempts at measuring the difference, their results, and what they do not
prove. The measured skill is `tests`, with a neutral agent: first on a one-line fix, then on a new
function whose edges are the task.

## The question

Does an agent, given the same task, behave differently when `tests` is available? The rubric reads
four behaviours the skill promises, from the transcript:

1. a test was written before the production code;
2. a red was observed before the fix;
3. the test's name says what breaks if it fails;
4. the fix is at the cause and the suite is green.

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

## What it proves, and what it does not

- **The instrument repeats.** The canary already showed that a load is visible; in these twelve runs
  no load is visible, and the absence is a fact read from the transcript, not a gap.
- **The null is not the one-line fix.** The second task is new code with a contract on the edge, which
  is the moment `tests` names first, and the skill is not consulted either. On these two tasks the
  text cannot change the work because it is never read.
- **It does not prove the skill is useless, or that its rules are wrong.** A skill that is not loaded
  cannot be judged by the work. The finding is upstream of the rules: the trigger.
- **It does not prove which trigger would match.** The description is the only line an agent sees
  before it loads; the `When` section, which here says "when you are about to write new code", is
  invisible until after the load. That the model would load the skill under a different sentence is a
  hypothesis this corpus cannot test yet.
- Three runs per arm, one model, one task each: the claim is about this corpus, and a probe that runs
  an agent is evidence, not a proof.

## The next thing to edit: the trigger

The description is `Trigger: when writing, fixing or deciding what a change tests`. Its frame is a
test decision: `fixing` and `deciding what a change tests` carry it, and `writing` reads as one more
verb of the same list. Six transcripts cannot say whether that frame failed to match the task or
whether the model never considered a skill on work this narrow —both predict no load— so the edit
varies the one thing the document can name: it removes the testing frame from the trigger and leaves
the craft in the body.

> Trigger: when writing or changing code, before the first edit.

`skills/**` is fenced for this probe —it measures `tests` as it is— so the edit is its own feature,
and it reopens the question with the same two arms and the same rubric. A load under the new line
confirms the frame as the cause; no load under it says the description was not the lever, and the
finding moves out of the text: it is about how the client selects a skill at all, and the next probe
measures that instead. A model judge replaces the person reading only if it agrees with the person
on this corpus first.
