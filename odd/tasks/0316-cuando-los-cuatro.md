# cuando-los-cuatro
Base: 729e3375e1980432938bcf3abe43f4d764baf22e

## Requirements (verbatim, from the user)
R1 · "cuando los cuatro lentes son el mismo modelo, doctor lo dice: el quórum es una opinión en cuatro disfraces, y la línea nombra cuántos modelos distintos juzgan"

## Done when
- cuando los cuatro lentes nombran el mismo modelo, doctor lo reporta como aviso nombrando cuántos modelos distintos juzgan y cuál comparten
- cuando los lentes nombran dos modelos distintos, el aviso no aparece
- la línea no cambia el código de salida de doctor
- la línea nombra el arreglo: un modelo por lente en odd/review.toml
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [x] T1 · el quórum se lee: el dominio cuenta los modelos distintos de los lentes y nombra el que comparten   R1 · tree:020a684fed12ab41f005f5ad40fa00e7838dcd61
- [x] T2 · doctor: un check nuevo avisa cuando el quórum es uno y no cambia el código de salida   R1 · tree:b35ef5e790f72516912a42323caecc4a339d79ef
- [x] T3 · las notas dicen 17 checks: el README, la nota de sesión y el changelog   R1 · tree:3c4bb6f255ebfdddb908f60614b768bcd1d7f20e

## History
L1 · "cuando los cuatro lentes son el mismo modelo, doctor lo dice: el quórum es una opinión en cuatro disfraces, y la línea nombra cuántos modelos distintos juzgan"
