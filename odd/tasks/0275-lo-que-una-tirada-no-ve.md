# lo-que-una-tirada-no-ve
Base: 507d578d727a913f2d0f610e3de82a15a7726ec2

## Requirements (verbatim, from the user)
R1 · "a verdict can be paid by N draws of the same judge: `odachi review --runs N` asks each required lens N times and files the union of what they answered as one verdict"

## Done when
- `odachi review --runs 3` with `--exec` asks each required lens three times and writes `<tree>.json` whose `draws` says 3 for every lens it asked
- a second `odachi review --runs 3` over the same tree pays only the two missing draws per lens and the file says 3; a run that asks for no new draw is refused, naming the count the lens already has
- two answers naming one defect at one place under one lens pay one row, and the row kept is the first draw's words
- a run whose third draw does not answer stores the two that did, says `2 of 3`, names the draw that failed, and exits blocked with the command that pays the third
- `--runs 1` behaves as today and records one draw per lens; a stored verdict with no `draws` reads as one draw per lens, so every verdict already in a store stays valid
- `--runs 11` is refused naming the ceiling of ten before a store or a reviewer is touched
- `--show` prints the draws beside the lenses, and `## Unreleased` names it, patch (0275)
- `go test ./...` is green and `odachi check --run "go test ./..."` is green

## Steps
- [ ] T1 · el flag tiene borde: --runs va de 1 a 10, el techo se rechaza antes del store, y sin --exec no hay tirada   R1 · route:inline
- [ ] T2 · N tiradas por lente pagan una union: cada lente contesta N veces y dos tiradas del mismo defecto dejan una sola fila   R1 · route:inline
- [ ] T3 · el veredicto dice cuantas tiradas pago cada lente: draws en el archivo, ausente lee uno, y --show lo imprime   R1 · route:inline
- [ ] T4 · una corrida posterior completa la lente: se paga solo la diferencia hasta N, y pedir cero se rechaza con el conteo   R1 · route:inline
- [ ] T5 · una tirada que falla no tira las otras: lo que contesto se guarda, el conteo dice 2 de 3, y la corrida sale bloqueada   R1 · route:inline

## History
L1 · "a verdict can be paid by N draws of the same judge: `odachi review --runs N` asks each required lens N times and files the union of what they answered as one verdict"
L2 · OFICIO · "Una corrida posterior lleva la lente a N en vez de re-contestarla: paga solo la diferencia" (porque la union solo agrega hallazgos, asi que re-tirar nunca puede convertir un changes-requested en aprobado, y nada ya pagado se paga dos veces; el humano lo eligio sobre apostar N antes de ver la primera tirada)
L3 · OFICIO · "Una tirada que no contesta deja el veredicto con las que contestaron, y la corrida sale bloqueada" (porque guardar no es haber hecho lo pedido: el registro no tira plata ya pagada y el codigo de salida dice que la corrida no logro lo que se le pidio)
