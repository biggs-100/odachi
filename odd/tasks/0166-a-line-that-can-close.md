# a-line-that-can-close
Base: e9de1aa37fb32db56bb7f4c7b237305963743816

## Requirements (verbatim, from the user)
R1 · "a change that fits one line is written as one line: `odachi start --light "<requirement>"` opens a feature with a task doc that carries the quoted requirement and one step with the same words, no brief, and the project's fence and voice untouched"

## Done when
- `odachi start --light "<requirement>"` writes one task doc with the line `Light: yes`, R1 quoted and T1 named with the same words, and writes no brief
- the same run does not rewrite `odd/fence.toml` or `odd/profile.toml`
- `odachi check` on a light feature does not ask for a brief, and still refuses a fence violation, an open question and a missing or stale proof
- a light feature closes like any other: one step done means the proof is owed and the review is asked for its tree
- `odachi plan "<text>"` on a light feature is refused, saying that a second step is the signal to leave the lane
- a document without `Light:` behaves exactly as today, and `Light: yes` outside the preamble is prose the gate never reads
- `drive` and the changelog name the lane and say when not to use it

## Steps
- [x] T1 · the task doc carries the lane it was written for, in its preamble   R1 · tree:992b9abd0e55f5e715d5c5683415a65ee377b90a
- [x] T2 · start --light writes one line and leaves the fence and the voice alone   R1 · tree:81febecd21ba48465804026c6fd5da187244a24b
- [x] T3 · check does not ask a light feature for a brief, and asks the same of everything else   R1 · tree:8da267d5e5e457100b4914e53f00b60183d48e8a
- [x] T4 · plan refuses a second step on a light feature   R1 · tree:449c26f10a85683c358e306d6a69cc52fe561a66
- [x] T5 · drive and the changelog name the lane and when not to use it   R1 · tree:d51569e20001248e71118709085c87bdf25daef4

## History
L1 · "a change that fits one line is written as one line: `odachi start --light "<requirement>"` opens a feature with a task doc that carries the quoted requirement and one step with the same words, no brief, and the project's fence and voice untouched"
L2 · OFICIO · "two refusals of the light lane and the two-argument branch have no test: the --brief conflict, the already-open name, and the extra positional (review WARNING and SUGGESTION on start.go:189 and main.go:186); the next feature pays them" (porque the verdict approved this tree and a warning earns no review round; the tests are three lines each and land with the next change to those verbs)
