# una-regla-de
Base: c653ba346b6c9a2551332d585c8bd57eabf0202a

## Requirements (verbatim, from the user)
R1 · "una regla de la cerca protege archivos: lo que no nombra archivos se queda en el brief, y start deja de escribir plantillas comentadas y de avisar una vez por cada una"

## Done when
- start reescribe odd/fence.toml con las reglas activas y nada mas
- start imprime una linea con el numero de reglas y dice que los non-goals sin rutas viven en el brief, en vez de una linea por non-goal
- un non-goal nuevo con rutas se escribe como regla, y uno sin rutas no agrega nada al archivo ni repite su texto
- el ejemplo de una cerca vacia se queda: un proyecto sin non-goals sigue aprendiendo la forma
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0299)

## Steps
- [x] T1 · el render escribe reglas y nada mas, y el ejemplo de la cerca vacia se queda   R1 · route:inline · tree:2f0a04f02bdefc18eb536337648263c4edf3406a
- [x] T2 · start cuenta las plantillas en una linea en vez de una por non-goal   R1 · route:inline · tree:2f0a04f02bdefc18eb536337648263c4edf3406a
- [x] T3 · la cerca de este repo queda sin plantillas: la segunda corrida de start la poda   R1 · route:inline · tree:7b3d4ae2821f27da003b5f7af37206ffdc0c0a6d
- [x] T4 · el changelog nombra 0299 bajo Unreleased, patch   R1 · route:inline · tree:06662bc2e589340ad041915ea46c9cb34f8119b3

## History
L1 · "una regla de la cerca protege archivos: lo que no nombra archivos se queda en el brief, y start deja de escribir plantillas comentadas y de avisar una vez por cada una"
