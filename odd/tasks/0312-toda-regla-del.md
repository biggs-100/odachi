# toda-regla-del
Base: e149c092bd2cb4503d8f0e7eb3252164f47a0c28

## Requirements (verbatim, from the user)
R1 · "toda regla del catálogo que se puede hacer cumplir nombra su enforcer, y doctor cuenta como deuda la que no lo tiene; cada fila de check nombra el comando que arregla su rojo"

## Done when
- odd/rules.toml empareja cada regla con su enforcer, y un archivo ausente no es deuda
- odachi doctor imprime el conteo y nombra cada regla sin enforcer con su skill
- cada fila de check de review.toml nombra el comando que arregla su rojo desde el schema 10, y la tabla de este repo lo hace en la feature que sigue al release
- el backfill cubre las reglas del catálogo que ya tienen enforcement, y las que no quedan en none con su razón
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · la tabla y su lectura: el dominio puro clasifica el enforcer y el adaptador lee odd/rules.toml, con sus tests   R1 · route:inline · tree:2802000eaf3a5783d5cbe4fa7329d0e8713b704f
- [x] T2 · doctor cuenta las reglas sin enforcer y nombra cada una con su skill, y un enforcer que no resuelve es un problema   R1 · route:inline · tree:0c00e27c6b2b229c82047eae9d499e7f9b49154d
- [x] T3 · cada fila de check de review.toml nombra el comando que arregla su rojo, y una fila sin remedio no valida   R1 · route:inline · tree:7337bb5d29379df77bf1abcc851d832ce43a37e6
- [x] T4 · el backfill de las reglas que hoy tienen enforcement y el changelog   R1 · route:inline · tree:f282ee435a3d0fde06f7d7c3541e1ba608cc8552

## History
L1 · "toda regla del catálogo que se puede hacer cumplir nombra su enforcer, y doctor cuenta como deuda la que no lo tiene; cada fila de check nombra el comando que arregla su rojo"
L2 · OFICIO · "un cambio de schema se commitea con el binario del arbol adelante en el PATH, compilado en temp: el hook resuelve odachi del PATH, y un binario instalado viejo no puede leer el archivo nuevo" (porque instalar el binario cambia la maquina y necesita el si del humano; compilar en temp y prepender el PATH corre la puerta con el codigo del arbol, que es el emparejamiento correcto, y no salta nada)
L3 · CAMBIO R1 · "cada fila de check nombra el comando que arregla su rojo desde el schema 10, y la tabla de este repo lo hace en la feature que sigue al release"
