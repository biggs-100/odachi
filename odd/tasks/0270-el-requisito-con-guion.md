# el-requisito-con-guion
Base: b6975a725f2303ee553ccaa9d83cc3cc19e67d17

## Requirements (verbatim, from the user)
R1 · "un requisito que empieza con guion se puede pasar: `start --light -- "--runs 0 …"` lo escribe tal cual, y sin el `--` el error dice cómo pasarlo"

## Done when
- `odachi start --light -- "--runs 0 se rechaza"` writes a task document whose requirement and step carry that sentence verbatim
- `odachi start --light "--runs 0 se rechaza"` fails with a message that names `odachi start --light -- "<the line>"`
- `odachi plan "- [ ] T1 · algo"` is still refused, and the message says a step does not start with a bullet
- a step text that starts with a single dash is accepted, because the hazard is the bullet and not the dash
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0270)

## Steps
- [ ] T1 · un paso puede empezar con guion: se rechaza la viñeta, y el mensaje dice viñeta   R1 · route:inline
- [ ] T2 · start --light sin -- ensena la forma: el error nombra el -- y la linea entre comillas   R1 · route:inline

## History
L1 · "un requisito que empieza con guion se puede pasar: `start --light -- "--runs 0 …"` lo escribe tal cual, y sin el `--` el error dice cómo pasarlo"
