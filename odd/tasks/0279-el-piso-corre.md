# el-piso-corre
Base: 3486e65cb842e9852492b6b222a531c2df496082

## Requirements (verbatim, from the user)
R1 · "el piso corre de verdad: golangci-lint corre en cada review y la deuda que encuentra se paga"

## Done when
- el ruleset queda fijado en .golangci.yml y golangci-lint run ./... sale cero en este arbol
- los 46 hallazgos se arreglan o se borran: nada queda silenciado con //nolint sin la razon escrita
- odd/review.toml declara la comprobacion a CRITICAL, este clon la aprobo, y un review dice lint: passed
- un arbol con un hallazgo nuevo deja la fila check:lint en el veredicto y pide cambios
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0279)

## Steps
- [x] T1 · el ruleset queda fijado en .golangci.yml: cinco linters nombrados y una version nueva no mueve el piso   R1 · route:inline · tree:0b58026d7789f7a1fb516a26ef11baec2097147e
- [x] T2 · los 38 errores ignorados se contestan: errcheck sale cero y cada Close y Release dice que hace con su error   R1 · route:inline · tree:c58448f3d128f0557617b118374f313a5e9cbd01
- [x] T3 · lo que el linter llama muerto se borra: unused, ineffassign y staticcheck salen cero   R1 · route:inline · tree:f0800e08091b41c336e90d1404bc608a77f17d67
- [ ] T4 · el piso se declara a CRITICAL, este clon lo aprueba, y un review de este arbol dice lint: passed   R1 · route:inline
- [ ] T5 · la cara: el README dice como se declara un piso y el changelog nombra la feature   R1 · route:inline

## History
L1 · "el piso corre de verdad: golangci-lint corre en cada review y la deuda que encuentra se paga"
L2 · OFICIO · "El ruleset se fija en .golangci.yml con los cinco linters nombrados, y el piso no hereda los defaults de la version instalada" (porque los defaults de golangci-lint cambian entre versiones: un piso que se mueve solo convierte un upgrade en hallazgos nuevos sin que nadie lo decidiera, y una comprobacion CRITICAL que aparece sola frena arboles que nadie toco)
L3 · OFICIO · "Se declara a CRITICAL y por eso el proyecto paga los 46 antes de declararla" (porque una fila que solo informa es ruido con fecha; declararla antes de pagar la deuda deja el arbol en changes_requested por defectos que ya existian, y la deuda existente se paga una vez, no se hereda como bloqueo permanente)
L4 · OFICIO · "Nada se silencia sin razon: un //nolint lleva el porque y la fila que contesta, y un linter entra al ruleset como decision, no como paso" (porque el silencio es como una comprobacion que reporta sin correr: el piso existe para que se vea, y un //nolint sin razon es el mismo defecto escrito en el codigo)
L5 · OFICIO · "El ruleset fijado mide 54, no 46: los ocho que el default escondia se pagan igual" (porque al nombrar los linters en el archivo aparecen tres SA5011 en tests, cuatro QF (Fprintf y tagged switch) y un campo sin usar que el default de la version instalada no reportaba; el numero del Done-when era una medicion, y la vara es que la corrida salga cero)
L6 · OFICIO · "El fence de 0278 deja de proteger internal/app/lensprobe: la regla sigue cuidando el fixture y el puntaje, y el arbol de 0279 puede limpiar la sonda donde el linter la nombra" (porque los dos hallazgos del probe son un defer identico en semantica y una funcion que nadie llama; la regla existia para que el review de 0278 no ajustara la medicion, y el humano decidio que la deuda del linter no se paga excluyendo el paquete que mide los lentes)
