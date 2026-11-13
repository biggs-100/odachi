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
- [ ] T1 · la prueba falla con una negativa nueva sin salida y avisa cuando una desaparece   R1 · route:inline
- [ ] T2 · el gate queda fijado: cada razon construye un next no vacio   R1 · route:inline

## History
L1 · "una negativa nueva que no diga como salir no entra, y el numero de negativas sin salida solo baja"
