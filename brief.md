# security-is-practice

## What must exist
- the curriculum teaches the security-relevant practice while the code is written -what arrives from outside, who may do what, a secret, a dependency, a log line, a failure- in one skill a client loads by situation, and one ADR fixes the split: what the gate already enforces and what only a person or an agent can judge

## Done when
- a change that reads input from outside, decides who may do what, holds a secret, adds a dependency, writes a log or handles a failure loads the skill, and the binary ships it to every client that installs the craft
- the ADR names in one place the enforceable halves -the fence over paths, the proof over a command the human chooses, the risk lens over the diff- and the half that is judgement
- every rule in the skill that claims how an outside system behaves names its source and version, and the skill stays under the sixty-line ceiling

## Not included (first version)
- no scanner and no new doctor check: a tool that guesses is worse than a checklist a person runs
- no rewrite of the review lens: risk already carries the twenty-one patterns a diff is read against
- no change to the seventeen skills that ship today; this feature adds one folder and the index

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why security is decided at trust boundaries and not in a list.
