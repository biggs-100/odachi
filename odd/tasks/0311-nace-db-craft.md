# nace-db-craft
Base: 65967dcd522956950adb23229ae74d0f790c4c6c

## Requirements (verbatim, from the user)
R1 · "nace db-craft: el oficio de trabajar con una base de datos, con migraciones sin downtime, transacciones, indices, consultas, acceso y operacion, y el catalogo de errores que cuestan incidentes"

## Done when
- db-craft existe con cuerpo de 60 lineas o menos, cada reference con su fila en ## Decision gates y su bullet
- las siete paginas cubren esquema, migraciones, transacciones, indices, consultas, acceso y operacion
- the-failures.md es el catalogo de errores con su sintoma y su arreglo, y cada fila apunta a la pagina duena
- data deja de ser duena de las migraciones y las manda a db-craft, sin perder lo que ya dice
- Postgres primera persona, con la nota de que cambia en MySQL y SQLite
- una prueba fija el ruteo de cada reference y las cadenas que no deben faltar
- docs/impact-map.md y el changelog lo nombran, y SKILLS.md se regenera
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0311)

## Steps
- [x] T1 · la puerta nace con migraciones y esquema, las dos paginas que rompen produccion   R1 · tree:ea59ff9808337bfaff7d15b36154c5a0f9e71855
- [x] T2 · transacciones e idempotencia: aislamiento, reintento completo y outbox   R1 · tree:4e41f7efb3aec9b0140d41d1950c2d5f669c144c
- [x] T3 · indices y consultas: orden de columnas, covering, keyset y N+1   R1 · tree:967c8524bb4f57909145be21e4e48d12a675d97a
- [x] T4 · acceso y operacion: minimo privilegio, RLS, PITR y drill de restore   R1 · tree:a426baf57aabd6e2d19acad71228f1b3f95d30d7
- [x] T5 · el catalogo de fallos, los registros y la prueba del ruteo   R1 · tree:1317ae0611748a3c3dfed84cac600a0e26ea9ee2

## History
L1 · "nace db-craft: el oficio de trabajar con una base de datos, con migraciones sin downtime, transacciones, indices, consultas, acceso y operacion, y el catalogo de errores que cuestan incidentes"
L2 · OFICIO · "La base se describe en Postgres y no en SQL generico" (porque los numeros que cambian una decision -lock_timeout, NOT VALID, CREATE INDEX CONCURRENTLY, niveles de lock- son de Postgres; un SQL generico queda en consejos que no cambian nada, y la portabilidad se dice en una nota por pagina)
L3 · OFICIO · "db-craft nace como puerta nueva y data delega las migraciones" (porque son dos situaciones distintas: donde vive el dato y cuanto tiempo, contra escribir el esquema, la migracion o la consulta; una sola puerta con las dos se queda sin disparador y sin sitio en 60 lineas)
