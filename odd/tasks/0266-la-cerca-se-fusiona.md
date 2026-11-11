# la-cerca-se-fusiona
Base: dac1e1bfaf2c6a3c39418ff689700a1fadef6adf

## Requirements (verbatim, from the user)
R1 · "start deja de reescribir odd/fence.toml: conserva las reglas que ya están y solo agrega las nuevas, y una regla con paths nunca se pierde"

## Done when
- `odachi start` over a fence that already protects `proof/**` ends with that rule and its `not` unchanged
- a new non-goal adds its rule and `check` reads the result: the round trip through the adapter returns the same rules
- a non-goal that repeats an existing rule leaves one `[[rule]]` with that `not`
- with no fence file the run writes today's file, the commented template for a non-goal without paths included
- the run says how many rules it kept
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names the merge, patch (0266)

## Steps
- [x] T1 · one owner for the file: fencetoml decodes and renders it, and check and start both go through it   R1 · route:inline · tree:65bf43f7848a0827d8f713cb2a21f10a0b28cfcc
- [ ] T2 · start merges: it keeps every rule already in the fence, adds only the new ones, and says how many it kept   R1 · route:inline

## History
L1 · "start deja de reescribir odd/fence.toml: conserva las reglas que ya están y solo agrega las nuevas, y una regla con paths nunca se pierde"
L2 · OFICIO · "the fence grammar moves from check to internal/adapters/fencetoml" (porque two apps need it now, the domain may not import BurntSushi (archtest), and a merge that renders what the gate cannot decode is the same defect in a new coat)
L3 · OFICIO · "opening 0266 with the current binary dropped the fence again, and this run put it back by hand" (porque start still rewrites odd/fence.toml until T2 lands; the restore keeps the four rules and adds this run's four templates, and T2 deletes the workaround it proves)
