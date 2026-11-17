# una-reescritura-del
Base: 4081d66fe132a4fa204b354fd7c5ddab72e8d6c7

## Requirements (verbatim, from the user)
R1 · "una reescritura del mismo tamaño no puede borrar una regla de los lentes sin que la puerta lo nombre: cada linea normativa queda medida y cualquier cambio pide una edicion deliberada de la baseline"

## Done when
- la baseline nombra cada linea normativa de los cuatro lentes
- borrar una regla y poner otra del mismo largo hace fallar la prueba nombrando la linea perdida
- agregar una linea normativa sin escribirla en la baseline hace fallar la prueba nombrando la linea nueva
- la baseline cambiada con su razon deja la suite verde
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [x] T1 · la baseline nombra cada linea normativa de los cuatro lentes   R1 · route:inline · tree:714d34d0731101ab01b9db64109d49b8e91e4315
- [x] T2 · el barrido de union falla en los dos sentidos y se prueba mutandolo   R1 · route:inline · tree:0f01f58d769a52378e5016abb02a78fbd978a30f
- [x] T3 · el changelog y la nota de cierre   R1 · route:inline · tree:55f63f3b27e1e30a9ee630d9f16f15cc8750c957

## History
L1 · "una reescritura del mismo tamaño no puede borrar una regla de los lentes sin que la puerta lo nombre: cada linea normativa queda medida y cualquier cambio pide una edicion deliberada de la baseline"
L2 · OFICIO · "Abro el barrido de unión sobre los cuatro lentes con esta frase como requisito" (porque El humano dijo ok, continua despues de que ofreciera el segundo aprendizaje de la comparacion con gentle-shell; los lentes son la unica superficie que el revisor recibe y el pin de bytes no distingue una regla borrada de una reescritura del mismo largo)
