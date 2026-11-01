# the-lenses-name-what-they-missed

## What must exist
- the four lens prompts name the defects the comparison with gentle-ai found missing: a suite that got smaller, a test at the wrong level, a failure with no threshold and a failure nobody can see, and behaviour the brief does not explain

## Done when
- reliability names a focused or skipped test that makes CI pass by shrinking the suite, and a test that proves a function through a whole flow
- resilience names a failure with no threshold the project chose, and a failure with nothing in production that would show it
- readability names behaviour the diff changes and the brief never asked for
- every lens keeps between twelve and twenty rules of the length the format enforces, and the asset tests stay green

## Not included (first version)
- no change to the review engine, the prompt builder or the severity vocabulary: this feature edits four prompts and nothing else
- no borrowed numbers: a threshold is what the project chose, and another project's figure is not a rule here
- no fifth lens: the four stay, and each one grows by what it was missing
- no change to the quorum or the refutation rule

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why a lens is a list of defects and not a list of virtues.
