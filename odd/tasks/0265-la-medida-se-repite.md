# la-medida-se-repite
Base: 532720b1055ddc23a5c30e25187cae9203f5f68e

## Requirements (verbatim, from the user)
R1 · "el probe deja de reportar una tirada: corre el fixture N veces y reporta, por lente y por brazo, una mediana y el rango en que se movió"

## Done when
- sin `--runs`, el comando se comporta igual que hoy y lo dice en el resumen
- con `--runs N`, se ejecutan N corridas por cada brazo y cada lente, y el informe imprime mediana y rango de encontrados, ruido y limpio
- un `--runs 0` o negativo se rechaza nombrando el mínimo, y un `--runs` mayor que 10 se rechaza nombrando el costo
- los tests puntúan tres corridas distintas con respuestas escritas a mano y pinchan mediana y rango, sin red y sin modelo
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la nueva lectura de la precisión queda en `docs/review-research.md` con fecha, modelo y la dispersión observada
- Unreleased nombra la repetición, patch (0265)

## Steps
- [x] T1 · runs N: correr el fixture N veces por brazo y lente, guardar cada respuesta y agregar mediana y rango   R1 · route:inline · tree:bca26aa31442d6f5f4ed8bac4f5854f1caa80337
- [ ] T2 · medir con tres corridas y escribir la dispersion en docs/review-research.md   R1 · route:inline

## History
L1 · "el probe deja de reportar una tirada: corre el fixture N veces y reporta, por lente y por brazo, una mediana y el rango en que se movió"
