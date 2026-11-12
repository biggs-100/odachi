# govulncheck-corre-en
Base: ae4e2492515a69fb160a5578f854ec9716e177c1

## Requirements (verbatim, from the user)
R1 · "govulncheck corre en cada review como una fila del piso, y la deuda que encuentra se paga: no queda una vulnerabilidad que el codigo llame"

## Done when
- el toolchain sube a la version que arregla las 25: govulncheck ./... sale 0
- odd/review.toml declara govulncheck a CRITICAL, este clon la aprobo, y un review de este arbol dice govulncheck: passed
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0283)

## Steps
- [x] T1 · el toolchain sube a la version que arregla las 25: govulncheck sale 0   R1 · route:inline · tree:8318e074e248c219de4106fe495dabc35431bb4e
- [x] T2 · la fila govulncheck se declara a CRITICAL, este clon la aprueba, y un review la reporta   R1 · route:inline · tree:ea697b82f04907c7a9c456bcd10d000d25632221
- [ ] T3 · la cara: el README dice que el piso corre govulncheck y que necesita la base de vulnerabilidades   R1 · route:inline

## History
L1 · "govulncheck corre en cada review como una fila del piso, y la deuda que encuentra se paga: no queda una vulnerabilidad que el codigo llame"
L2 · OFICIO · "El go directive sube a 1.26.11 y no a 1.26.3: los arreglos de las 25 llegan hasta ahi" (porque 1.26.3 limpio seis de las 25 y dejo 19, asi que la version que pide el Done-when es la ultima del minor; pedir la minima que arregla una parte deja la fila roja y el numero del Done-when seria una mentira)
L3 · OFICIO · "La fila govulncheck se declara a CRITICAL y depende de la red: sin base de vulnerabilidades falla en vez de pasar" (porque un piso que no corre no mide nada y reportarlo como pasado seria el silencio que el piso existe para romper; el precio es que una maquina sin red no puede cerrar, y la salida es una refutacion con razon escrita, que ya existe para las filas de herramienta)
