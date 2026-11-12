# el-conteo-cuenta-respuestas
Base: b08b5c0339595371ea5d1796adc4c096c9935df2

## Requirements (verbatim, from the user)
R1 · "a candidate whose prompt does not fit and is reviewed in parts records one draw per answer, not one per part: `--runs 1` over a split candidate leaves `draws` at 1, and `--runs 3` at 3"

## Done when
- a split candidate, forced with a small `max_prompt_kb`, records `draws` of 1 for a run of one and 3 for a run of three
- the next line of a failed draw is a command `odachi review` accepts: it names `--exec` with a placeholder and the count, and it keeps the lens the run was told to judge
- a tree whose lens holds 5 draws, asked for 3, is refused with `--runs 6` in the next line
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0276)

## Steps
- [x] T1 · una tirada es una respuesta: las partes de un prompt partido no suman tiradas   R1 · route:inline · tree:9499221117180d0775a57e4000b6388a223cdf56
- [x] T2 · el handover se puede correr: lleva --exec con marcador y el lente que se pidio   R1 · route:inline · tree:7f9a211520327fecd451f030e45f866ffee9a585
- [x] T3 · la negativa nombra el conteo del archivo mas uno, no el de la corrida rechazada   R1 · route:inline · tree:ace7fe70f511d087b8f015b031de4b5d044417cb

## History
L1 · "a candidate whose prompt does not fit and is reviewed in parts records one draw per answer, not one per part: `--runs 1` over a split candidate leaves `draws` at 1, and `--runs 3` at 3"
