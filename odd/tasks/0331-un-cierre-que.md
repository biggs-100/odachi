# un-cierre-que
Base: 6c0e355a91e50bf8233ad2372866c5d232297f25

## Requirements (verbatim, from the user)
R1 · "un cierre que le debe un verifier mide su candidato antes de gastarlo: si el diff contra la base de la feature pasa de 400 lineas autoradas, el gate bloquea y nombra la rebanada"

## Done when
- un cierre con un riesgo que pide verifier y 401 lineas autoradas bloquea nombrando el numero y la rebanada
- el mismo cierre con 399 lineas, o con un riesgo que no pide verifier, pasa como hoy
- la cuenta excluye los archivos generados y el proof, y su definicion queda escrita para el humano
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [ ] T1 · la cuenta mide las lineas autoradas del diff contra la base y excluye los archivos generados y el proof   R1 · route:delegated/odachi-worker · skills:tests
- [ ] T2 · un cierre que le debe un verifier y pasa de 400 lineas bloquea nombrando el numero y la rebanada   R1 · route:delegated/odachi-worker · skills:tests

## History
L1 · "un cierre que le debe un verifier mide su candidato antes de gastarlo: si el diff contra la base de la feature pasa de 400 lineas autoradas, el gate bloquea y nombra la rebanada"
L2 · OFICIO · "el tope es 400 lineas autoradas, contadas como suma y resta del diff contra la base" (porque es el numero que el metodo ya lleva escrito en la heuristica por tarea y el que gentle-shell usa en commit-slice-signal.ts; un tope nuevo seria un segundo numero que nadie recuerda)
