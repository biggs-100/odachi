# the-reviewer-runs

## Requirements (verbatim, from the user)
R1 · "the reviewer runs with the model this session uses, by default, and the human can name another one"

## Done when
- {model} in review --exec is replaced with the session's provider/model#variant without the human typing it
- --model overrides the resolved model for one run and is used exactly as given
- a command without {model} runs byte-identical to today
- when {model} is asked for and nothing resolves, review stops naming --model; it never guesses
- resolution order: --model, then ODACHI_MODEL, then the OpenCode session; no database or private file is read

## Steps
- [ ] T1 · {model} in review --exec resolves to the session's model, or the run stops naming --model   R1
- [ ] T2 · docs: how the reviewer inherits the model, the resolution order and ADR 0018   R1

## History
L1 · "the reviewer runs with the model this session uses, by default, and the human can name another one"
