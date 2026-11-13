# el-ratchet-llega
Base: 28fd6cc40863425e698756001fc31a814a2a8ee8

## Requirements (verbatim, from the user)
R1 · "el ratchet llega a review: sus negativas sin salida quedan congeladas y el mecanismo sigue viviendo en un solo lugar"

## Done when
- internal/refusal es el unico lugar del walker y de las dos direcciones: review lo consume sin copiar la logica
- review congela sus negativas sin salida y sus llamadas ilegibles en su propio refusals.baseline
- una negativa nueva en review sin salida falla la suite
- check y start conservan su baseline byte por byte: la prueba sigue midiendo lo mismo
- go test ./... is green and odachi check --run "go test ./..." is green

## Steps
- [ ] T1 · review lee sus negativas por el walker compartido y las congela en su baseline   R1 · route:inline

## History
L1 · "el ratchet llega a review: sus negativas sin salida quedan congeladas y el mecanismo sigue viviendo en un solo lugar"
