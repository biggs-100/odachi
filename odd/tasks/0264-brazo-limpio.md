# brazo-limpio
Base: 2557f118035a4be5ffae203281ea19435f8ebe05

## Requirements (verbatim, from the user)
R1 · "el brazo limpio del fixture deja de tener defectos: los hallazgos que un lente reporta sobre él se pagan en el código, o se registran como gusto y se dejan"

## Done when
- el código de `clean/` responde a los diez hallazgos de la primera medida: el status del remoto, el error de lectura y el de escritura, el timeout del cliente, el 404 que tapa un timeout de base, el límite que no acota negativos, y los tests que faltan del handler y del cache
- una corrida del probe con revisor devuelve `clean 0` y los cinco sembrados encontrados
- el manifiesto no se mueve: los cinco defectos, sus paths, sus líneas y sus pisos son los mismos
- `go test ./...` es verde y `odachi check --run "go test ./..."` es verde
- la medida nueva queda en `docs/review-research.md` con fecha y modelo, y reemplaza a la anterior como lectura vigente
- Unreleased nombra el fixture limpio, patch (0264)

## Steps
- [x] T1 · limpiar el codigo del brazo clean: los diez hallazgos de la primera medida, sin tocar el manifiesto   R1 · route:delegated/odachi-worker · skills:tests · tree:f8ef58e2e427d671743b1ec4ee80a5d0d9e3a8fe
- [ ] T2 · medir de nuevo y escribir la lectura vigente en docs/review-research.md con fecha y modelo   R1 · route:inline

## History
L1 · "el brazo limpio del fixture deja de tener defectos: los hallazgos que un lente reporta sobre él se pagan en el código, o se registran como gusto y se dejan"
