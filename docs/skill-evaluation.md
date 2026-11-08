# The first behavioural measurement of a skill

The skills are the product: text an agent reads before it works. `embed_test.go` proves the text
still carries certain sentences; it does not prove that an agent, reading them, works differently.
This document records the first attempt at measuring the difference, its result, and what it does
not prove. The measured skill is `tests`, on a one-line fix, with a neutral agent.

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

## The task

A Go program with one defect, copied fresh for every run:

```
go.mod              module task, go 1.22
main.go             fmt.Println(calc.Sum(2, 3))
calc/sum.go         func Sum(a, b int) int { return a - b }
```

The prompt is neutral and names the symptom, never the method: *"El programa imprime -1 cuando
debería imprimir 5 para 2+3. Arreglalo."*

## The two arms

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

## The result

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

## What it proves, and what it does not

- **It proves the instrument works.** The canary shows a load in the transcript, so the absence of a
  load in these six runs is a fact, not a gap in the reading.
- **It proves that on this task the skill changes nothing**, because it is never consulted: a skill
  the agent does not load cannot change the work.
- **It does not prove the skill is useless.** The task may be below the threshold the skill itself
  names —"verification is proportional"— and the model judged a one-line fix not worth a test. The
  result is about a one-line fix and a neutral agent, nothing more.
- The rubric assumes the skill is loaded. When it is not, a 0 means "not exercised", not "the rule
  failed".

## The next probe

A task where a test is clearly owed, and where the model cannot finish without deciding: a new
function with edges (empty, one, many, an error), or a defect whose shape hides in an edge. If the
skill is still not loaded there, the finding is about the trigger and the description, and that is
the next thing to edit. It reopens when a task like that exists or when the measurement is repeated
enough to automate; a model judge replaces the person reading only if it agrees with the person on
this corpus first.
