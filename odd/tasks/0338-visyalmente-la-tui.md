# visyalmente-la-tui
Base: d3c90447992e60ffdab6574e47be741101dfd2f3

## Requirements (verbatim, from the user)
R1 · "visyalmente la tui se ve fea, cuando quiero cambiar un modelo o otra opciones, las tengo que escribir, seria mejor que detectara los modelos disponibles del agente para poder seleccionar y no tener que escribirlos, asi como en gentle-ai"

## Done when
- elegir un modelo abre una lista navegable con el valor en vigor marcado, y enter lo deja sin escribir nada
- una deteccion vacia o fallida se explica en la pantalla y deja la escritura como salida, sin bloquear el verbo
- ninguna pantalla compone filas sueltas: cabecera, fila seleccionada, valor, aviso y pie de teclas salen de una capa con nombres
- go test ./... is green and odachi check --run is green and odachi skills --check is green

## Steps
- [x] T1 · la tui estrena una capa visual con nombres y cada pantalla la usa en vez de componer texto suelto   R1 · route:delegated/odachi-worker · tree:93897538a9ec33a50d53788ff095906828995834
- [x] T2 · una lista larga se recorre con ventana y dice que hay mas arriba o abajo, en vez de crecer sin fin   R1 · route:delegated/odachi-worker · tree:2b053fc9faac1ae69520ccda604d4b26b8f1b02a
- [x] T3 · el modelo se elige de una lista que el host responde, con enter, y la escritura queda como salida   R1 · route:delegated/odachi-worker · tree:6c8106f92d37772dd3a6714dddea8d9644c0c817
- [x] T4 · la lista explica una deteccion vacia, fallida o ausente -claude- y ninguna de las tres bloquea   R1 · route:delegated/odachi-worker · tree:b78c6b2925610d9861dab174506ced58f2caeb91

## History
L1 · "visyalmente la tui se ve fea, cuando quiero cambiar un modelo o otra opciones, las tengo que escribir, seria mejor que detectara los modelos disponibles del agente para poder seleccionar y no tener que escribirlos, asi como en gentle-ai"
