# el-revisor-esta
Base: 5ebc302664c3c77fb834e59d0fa293ea984534d3

## Requirements (verbatim, from the user)
R1 · "el revisor esta acotado por el cliente y no por el prompt: la corrida recibe una configuracion que niega cualquier ruta fuera del candidato, el snapshot vive en un directorio que odachi posee, y una marca afuera no se puede leer aunque el prompt lo pida, en la sonda y en el review real"

## Done when
- un revisor que intenta leer una ruta fuera del candidato recibe Permission denied: external_directory, y el mismo revisor lee el candidato sin problema
- glob y grep dentro del candidato no encuentran una marca escrita afuera
- el snapshot que review --exec le entrega al revisor tiene un padre que odachi creo, y ahi vive la configuracion que lo acota
- la sonda corre sus dos brazos bajo la misma configuracion, y el canario lo demuestra pidiendo una lectura de afuera que se rechaza
- go test ./... is green and odachi check --run go test ./... is green

## Steps
- [x] T1 · review --exec le da al revisor un snapshot con padre propio, y la configuracion que lo acota vive ahi   R1 · tree:2f63142a0b1ce87dd77051491268e6babc3c045a
- [x] T2 · la sonda escribe la misma configuracion en su mundo y su canario pide una lectura de afuera que se rechaza   R1 · tree:dcb8182a6133ff754224440e4cecf37f003aa20f
- [x] T3 · el acotamiento se prueba con una corrida grabada: sin configuracion la marca de afuera se lee, con ella se rechaza   R1 · tree:24d3046ebac1c32760f92cd04a2da1408fef92dd

## History
L1 · "el revisor esta acotado por el cliente y no por el prompt: la corrida recibe una configuracion que niega cualquier ruta fuera del candidato, el snapshot vive en un directorio que odachi posee, y una marca afuera no se puede leer aunque el prompt lo pida, en la sonda y en el review real"
