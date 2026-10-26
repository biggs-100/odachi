# the-questions-arrive

## What must exist
- the questions arrive in the report with their id and their words, not only their count

## Done when
- progress --json publishes each open question with its id and its words
- the human report names them from the same state, and a feature with none prints what it printed before
- the contract test pins the names id and text, and an empty list instead of null

## Not included (first version)
- the gate domain rule, which keeps counting
- answering a question from progress: it only reports
- the answer command in the state: the id is enough
- the plugin itself

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: how a contract changes a day before its first reader freezes it.
