# la-hora-de
Base: c9e8958a411978f0b9f1edd6967cd503a9894fd6

## Requirements (verbatim, from the user)
R1 · "la hora de un veredicto la escribe el reloj al sellarlo, y lo que declaro el modelo queda al lado como reclamo, no como hecho"

## Done when
- un veredicto cuyo JSON dice 00:00Z se sella con la hora del reloj y la declarada queda en declared_at
- un veredicto sin hora declarada no gana el campo, y uno que declara la misma hora tampoco
- odachi review --show dice la hora del sello y, cuando discrepa, la que escribio el revisor
- el esquema sube a 7: un binario viejo dice actualiza
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, minor (0286)

## Steps
- [ ] T1 · el sello escribe la hora del reloj y guarda el reclamo aparte   R1 · route:inline
- [ ] T2 · la cara: --show y la linea del review dicen la hora y el reclamo cuando discrepa   R1 · route:inline
- [ ] T3 · el esquema sube a 7 y el changelog lo dice   R1 · route:inline

## History
L1 · "la hora de un veredicto la escribe el reloj al sellarlo, y lo que declaro el modelo queda al lado como reclamo, no como hecho"
