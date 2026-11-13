# la-procedencia-de
Base: e6aad27427eb54aa98fbe99499909c4c5910f442

## Requirements (verbatim, from the user)
R1 · "la procedencia de un veredicto nombra a cada autor una sola vez"

## Done when
- un veredicto que fusiona dos tiradas y una correccion nombra a cada autor una sola vez
- un autor que solo aparece en una de las dos partes sigue en la fusion
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0284)

## Steps
- [x] T1 · la fusion nombra a cada autor una sola vez y no pierde uno que solo trae una parte   R1 · route:inline · tree:5ccbb012801814da6b2882b36190ec98a7adfb47
- [ ] T2 · la cara: el changelog dice que la procedencia deja de acumular   R1 · route:inline

## History
L1 · "la procedencia de un veredicto nombra a cada autor una sola vez"
L2 · OFICIO · "La union se hace separando por ' + ', que es la grafia que la app ya escribe, en vez de cambiar el campo a una lista" (porque el esquema no sube y un binario viejo sigue leyendo el registro; una lista seria mas honesta estructuralmente, pero cuesta un contrato nuevo y una release minor por un campo que hoy se lee mal, y el humano eligio pagar el defecto sin mover la forma)
