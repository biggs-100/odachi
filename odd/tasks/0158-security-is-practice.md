# security-is-practice
Base: 2cbb78634fab671478dccd7939274626d92a43a5

## Requirements (verbatim, from the user)
R1 · "the curriculum teaches the security-relevant practice while the code is written -what arrives from outside, who may do what, a secret, a dependency, a log line, a failure- in one skill a client loads by situation, and one ADR fixes the split: what the gate already enforces and what only a person or an agent can judge"

## Done when
- a change that reads input from outside, decides who may do what, holds a secret, adds a dependency, writes a log or handles a failure loads the skill, and the binary ships it to every client that installs the craft
- the ADR names in one place the enforceable halves -the fence over paths, the proof over a command the human chooses, the risk lens over the diff- and the half that is judgement
- every rule in the skill that claims how an outside system behaves names its source and version, and the skill stays under the sixty-line ceiling

## Steps
- [ ] T1 · the skill says when it loads and every claim about an outside system names its source, under sixty lines   R1
- [ ] T2 · the ADR fixes the split: the fence, the proof and the lens are the enforcement, and judgement is taught   R1
- [ ] T3 · the new skill reaches every client that installs the craft, and a test says so   R1

## History
L1 · "the curriculum teaches the security-relevant practice while the code is written -what arrives from outside, who may do what, a secret, a dependency, a log line, a failure- in one skill a client loads by situation, and one ADR fixes the split: what the gate already enforces and what only a person or an agent can judge"
