# cuando-check-run

## Requisitos (textual, del usuario)
R1 · "cuando check --run se queda sin tiempo, el siguiente comando sube --timeout en vez de repetir el mismo"

## Listo cuando
- un timeout de 7s sugiere --timeout 14
- un timeout de 120s sugiere --timeout 240
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [ ] T1 · subir --timeout en el next del timeout   R1

## Historial
L1 · "cuando check --run se queda sin tiempo, el siguiente comando sube --timeout en vez de repetir el mismo"
