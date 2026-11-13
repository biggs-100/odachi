# la-hora-de
Base: 8ee4f9bbe82cbb5fb98e151ba714dcd7baa044c8

## Requirements (verbatim, from the user)
R1 · "la hora de una refutacion tambien la escribe el reloj al sellarla, y lo que declara el refutador queda al lado como reclamo, no como hecho"

## Done when
- una refutacion que declara su hora se sella con la hora del reloj y lo declarado queda en declared_at
- una refutacion sin hora declarada no gana el campo, y una que declara la misma hora tampoco
- review --show dice la hora del sello y, cuando discrepa, la que escribio el refutador
- el esquema sube a 8 y un binario anterior a la refutacion nueva dice actualiza, no que el archivo esta roto
- las refutaciones ya guardadas no se reescriben: conservan la hora que traen
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, minor (0291)

## Steps
- [x] T1 · el sello: el reloj escribe la hora, lo declarado queda en declared_at, y la refutacion declara su esquema   R1 · route:inline · tree:69a97ee8747c8632acb0b23f72d2ad0877682e3b
- [x] T2 · la cara: review --show y la linea del guardado dicen la hora del sello y, cuando discrepa, la que escribio el refutador   R1 · route:inline · tree:9a7e8158caf6c5b6e1b86f58672bdafdf9865f22
- [x] T3 · el changelog nombra 0291 bajo Unreleased   R1 · route:inline · tree:79c2bd5e835266ec99e2cb647d468494c6ca61ae

## History
L1 · "la hora de una refutacion tambien la escribe el reloj al sellarla, y lo que declara el refutador queda al lado como reclamo, no como hecho"
L2 · OFICIO · "La linea de un testigo anterior al numero dice que su hora no es atribuible" (porque el parser marca esos archivos con el primer esquema y el comentario del campo dice que nada distingue la pluma; llamarla sello inventaria una atribucion para cada refutacion que ya vive en un store, que es justo la confusion que esta feature quita (0291))
L3 · OFICIO · "El reclamo se preserva por el campo guardado, no por el vacio" (porque en un segundo sello el at es la hora que escribio la herramienta, no la palabra del refutador; guardar el vacio como condicion hacia que la misma entrada dependiera de cuanto cayeron separados los dos sellos, y descartaba el reclamo real (0291))
