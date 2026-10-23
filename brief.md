# cuando-check-run

## Qué tiene que existir
- cuando check --run se queda sin tiempo, el siguiente comando sube --timeout en vez de repetir el mismo

## Listo cuando
- un timeout de 7s sugiere --timeout 14
- un timeout de 120s sugiere --timeout 240
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: mensajes de error.
