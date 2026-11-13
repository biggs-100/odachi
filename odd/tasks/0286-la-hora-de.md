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
- [x] T1 · el sello escribe la hora del reloj y guarda el reclamo aparte   R1 · route:inline · tree:edd488883037be15c15f4214f38e8aa3c71fe4e6
- [x] T2 · la cara: --show y la linea del review dicen la hora y el reclamo cuando discrepa   R1 · route:inline · tree:762b03a036183f1ca8cc06a5b9f1a78715bbb2da
- [ ] T3 · el esquema sube a 7 y el changelog lo dice   R1 · route:inline

## History
L1 · "la hora de un veredicto la escribe el reloj al sellarlo, y lo que declaro el modelo queda al lado como reclamo, no como hecho"
L2 · OFICIO · "El esquema sube con el campo (T1), no con la cara (T3)" (porque un archivo que trae un campo nuevo bajo el numero viejo es justo lo que el numero existe para anunciar: un binario anterior leeria el campo como desconocido y diria que el archivo esta roto, que es la mentira que el esquema evita)
L3 · OFICIO · "Las refutaciones no cambian en esta feature" (porque una refutacion tambien declara su hora con la pluma del que la escribe, pero el requisito es sobre el veredicto y el mismo arreglo ahi es otra superficie; queda nombrado, no escondido)
