# el-veredicto-juzga
Base: 56fd6d94caa868d2e62fee933ec951df1fe7903d

## Requirements (verbatim, from the user)
R1 · "el veredicto juzga el candidato y solo el candidato, y dice quien lo juzgo: un candidato que trae su propia configuracion de cliente se rechaza en vez de revisarse, --exec y --attach leen las mismas rutas del mismo candidato, y el modelo que queda escrito es el que respondio"

## Done when
- un candidato que trae su propio opencode.json se niega a revisarse con un mensaje que nombra el archivo, y el mismo candidato sin ese archivo se revisa
- el mismo candidato declarado por --exec y por --attach produce el mismo conjunto de rutas cambiadas, con un caso que hoy discrepa
- un veredicto guardado nombra el modelo que respondio aunque la corrida se haya adjuntado con --attach
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [x] T1 · un candidato que trae su propia configuracion de cliente se niega a revisarse, y el mensaje nombra el archivo   R1 · route:inline · tree:1a61e58daa54ab5f6eb06a544632cd2d08bfd06a
- [x] T2 · el candidato declarado por --exec y por --attach produce las mismas rutas, y una prueba fija el caso que hoy discrepa   R1 · route:inline · tree:6cbe5fb774a7a7accd7653a5a258c8e5701b5519
- [x] T3 · el veredicto guardado nombra el modelo que respondio cuando la corrida se adjunta con --attach   R1 · route:inline · tree:63646a77ac9e938aa78947e5813a57752d413f53

## History
L1 · "el veredicto juzga el candidato y solo el candidato, y dice quien lo juzgo: un candidato que trae su propia configuracion de cliente se rechaza en vez de revisarse, --exec y --attach leen las mismas rutas del mismo candidato, y el modelo que queda escrito es el que respondio"
