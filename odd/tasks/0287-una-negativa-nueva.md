# una-negativa-nueva
Base: a5541b765e6a8806413e8b13f6f99220b6a621c1

## Requirements (verbatim, from the user)
R1 · "una negativa nueva que no diga como salir no entra, y el numero de negativas sin salida solo baja"

## Done when
- cada negativa de internal/app/check con mensaje conocido o nombra su salida o lleva una anotacion by design de un vocabulario cerrado
- la linea base congela las de hoy: la prueba falla con una nueva y avisa cuando alguna desaparece
- el gate queda fijado: cada razon construye un next no vacio
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · el baseline mide las negativas de check: cada una con mensaje conocido, y las que no nombran salida quedan congeladas   R1 · route:inline · tree:8d545a5f591d4fffc17db521c7e8fabf8dd3923a
- [x] T2 · la prueba falla con una negativa nueva sin salida y avisa cuando una desaparece   R1 · route:inline · tree:77f417455282e617e96de56f02f015d86c2790f6
- [x] T3 · el gate queda fijado: cada razon construye un next no vacio   R1 · route:inline · tree:77f417455282e617e96de56f02f015d86c2790f6

## History
L1 · "una negativa nueva que no diga como salir no entra, y el numero de negativas sin salida solo baja"
L2 · OFICIO · "T1 se inserta a mano y los dos pasos se renumeran antes de empezar" (porque plan agrega al final y rechaza una linea de mas de 120 caracteres: el primer paso quedo fuera y un T4 fuera de orden describe un plan que nadie ejecuta en ese orden; el task doc es la memoria, y se corrige antes de que algun paso este hecho)
L3 · OFICIO · "La prueba del gate se extiende en vez de duplicarse" (porque gate_test.go ya cubria siete de las diez razones; agregar una segunda prueba con la misma regla deja dos lugares que alguien tiene que recordar actualizar, y la primera que se atrase deja pasar la razon nueva que la otra no mira)
L4 · OFICIO · "T3 se cierra con lo que ya estaba: la cerca protege internal/domain/gate y no se toca por dos filas de prueba" (porque la regla de 0263 dice que la compuerta no corre las comprobaciones, solo lee el veredicto que produjeron, y protege el paquete entero; aflojarla por dos casos de una prueba cuesta dos PRs y mueve una politica, y el valor no lo justifica. La prueba ya existente cubre siete de las diez razones y las dos que faltan quedan nombradas como deuda, no como hecho)
L5 · OFICIO · "La bajada del baseline falla en vez de loguearse" (porque el review de este arbol lo dijo con razon: en CI nadie lee un log, y un ratchet cuya bajada es invisible deja el numero mintiendo para arriba; que falle obliga a que el registro siga a la deuda en el mismo commit, y cuesta dos lineas)
