# the-reviewer-runs

## What must exist
- the reviewer runs with the model this session uses, by default, and the human can name another one

## Done when
- {model} in review --exec is replaced with the session's provider/model#variant without the human typing it
- --model overrides the resolved model for one run and is used exactly as given
- a command without {model} runs byte-identical to today
- when {model} is asked for and nothing resolves, review stops naming --model; it never guesses
- resolution order: --model, then ODACHI_MODEL, then the OpenCode session; no database or private file is read

## Not included (first version)
- reading OpenCode's database or log files: only its CLI
- a fallback model when detection fails
- changing the default where the agent itself reviews the printed prompt
- persisting the model in a config file: the override is one run or an env var
- new dependencies

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: no rush.
- Wants to learn: how a child tool learns its parent's context without owning the parent.
