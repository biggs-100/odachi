# medida-de-los-lentes
Base: 3833cac9b3a49b955f991e1c8e3407b34ba73877

## Requirements (verbatim, from the user)
R1 · "una medida del revisor: cuántos defectos sembrados encuentra cada lente, y cuánto ruido levanta sobre el mismo código sin defectos"

## Done when
- `odachi review probe --fixture <dir>` sin `--exec` imprime el prompt que pediría, no llama a nadie y sale 0
- con `--exec`, corre un revisor por lente, guarda cada veredicto en el store y puntúa contra el manifiesto
- un defecto cuenta como encontrado si hay un hallazgo del mismo path, con línea dentro de la ventana del manifiesto y severidad igual o mayor al piso
- un hallazgo que no corresponde a ningún defecto cuenta como ruido, y el brazo limpio no tiene ninguno por definición
- un fixture sin manifiesto, un manifiesto con un lente que no existe, o un defecto sin path son errores que nombran lo que falta
- el fixture no entra a la suite: los tests puntúan veredictos escritos a mano, sin red y sin modelo
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la primera medida queda escrita en `docs/review-research.md` con fecha, modelo y número
- Unreleased nombra el verbo y el fixture, patch (0263)

## Steps
- [x] T1 · el fixture clean/ y seeded/ y el manifiesto que nombra cada defecto, con su lector y sus errores   R1 · route:inline · tree:597024fe63b3130d8b1c586cdc57cb0ea4828c37
- [ ] T2 · el puntaje: un veredicto contra el manifiesto da encontrados, perdidos y ruido por lente   R1 · route:inline
- [ ] T3 · review probe corre los cuatro lentes, guarda sus veredictos y escribe la primera medida   R1 · route:inline

## History
L1 · "una medida del revisor: cuántos defectos sembrados encuentra cada lente, y cuánto ruido levanta sobre el mismo código sin defectos"
