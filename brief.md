# one-rule-per-line

## What must exist
- a rule is one bullet on one line: `skills --check` refuses prose or a wrapped rule in `## Rules`

## Done when
- `internal/domain/skill` refuses a `## Rules` line that is not a bullet, and a rule over 140 characters, each with its test observed red first
- every bundled skill carries `## Rules` as bullets, none over 140 characters, and none loses a rule the old text had
- `odachi skills --check` is green and `odachi skills --write` leaves SKILLS.md unchanged
- the tests that pin a skill's words still pass, or their expectation moves with the rule it pins
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the change: a contract change for skills, minor (0255)

## Not included (first version)
- no change to the 60-line body cap: the count becomes honest, the ceiling stays - back when a skill needs more than 60 lines of rules
- no token counter and no frontmatter field: a line is a rule - back when a budget per skill is measured to matter
- no change to `Never`, the gates or the references: this feature judges `## Rules` - back when another section hides prose the same way
- no change to the installer, the workflow, the canvas or the proofs - back when one of those paths changes

## Decisions taken
- One physical line per rule: the model follows a list better than a paragraph, and a line count is gameable because one line can hide twelve rules.
- 140 characters is the limit; a rule that needs more is two rules.
- The rewrite keeps every rule: a loss is a regression, not a simplification.
- The reach is the whole catalog, because the skill that hides the most rules is the one in use every day.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: whether a list of short rules changes what the agent does, or only how the file reads.
