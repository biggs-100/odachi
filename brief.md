# la-cerca-se-fusiona

## What must exist
- start deja de reescribir odd/fence.toml: conserva las reglas que ya están y solo agrega las nuevas, y una regla con paths nunca se pierde
- the merge reads the file it is about to replace with the same grammar the gate reads it with, so nothing start writes is a file check cannot parse
- a non-goal this run repeats does not add a second rule with the same name

## Done when
- `odachi start` over a fence that already protects `proof/**` ends with that rule and its `not` unchanged
- a new non-goal adds its rule and `check` reads the result: the round trip through the adapter returns the same rules
- a non-goal that repeats an existing rule leaves one `[[rule]]` with that `not`
- with no fence file the run writes today's file, the commented template for a non-goal without paths included
- the run says how many rules it kept
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names the merge, patch (0266)

## Not included (first version)
- the file keeps its shape: canonical TOML, one key per line, and the commented template for a non-goal without paths - back when the shape itself is the question
- one rule is never folded into another: the same `not` keeps the rule as written, and its paths and deps do not grow across runs - back when a rule itself has to change
- a commented template does not outlive the brief that wrote it: only rules with paths survive a new brief - back when a human wants their own notes kept
- no migration of fences an older odachi wrote: a file check already reads is merged as it is - back when the old shape has to be repaired

## Decisions taken
- Decode and render live in one adapter, fencetoml: a merge that writes a file the gate cannot read is this same defect in a new coat.
- The rule's `not` is its identity: a run that repeats a non-goal must find the rule already there, not add a twin.
