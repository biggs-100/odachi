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
- [ ] T4 · el handover corre de verdad: el lente ancho conserva su razon y el conteo nombrado es el minimo que paga una   R1 · route:inline

## History
L1 · "a candidate whose prompt does not fit and is reviewed in parts records one draw per answer, not one per part: `--runs 1` over a split candidate leaves `draws` at 1, and `--runs 3` at 3"
L2 · OFICIO · "La negativa nombra una mas que el conteo mas chico del archivo, no el mas grande" (porque una corrida que paga varias tiradas mientras la frase promete una es una sugerencia que gasta de mas; el minimo mas uno paga exactamente una, para los lentes que menos tienen)
L3 · OFICIO · "El CRITICAL sobre el handover ancho era falso: la razon queda en el registro del arbol y el test lo fija" (porque widenedRefusal acepta un --lens sin --because cuando el registro de alcance del arbol ya existe, y una corrida que aterrizo una respuesta lo escribio; el test corre la linea impresa sin la razon y comprueba que no se rechaza, para que la duda no vuelva)
