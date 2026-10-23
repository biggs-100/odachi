# los-backups-de

## Qué tiene que existir
- los backups de odachi viven en .odachi/backups y undo los lee junto con los viejos, sin perder la posibilidad de deshacer

## Listo cuando
- antes de reescribir un archivo del proyecto, el backup va a .odachi/backups con su ruta relativa
- undo descubre y restaura backups del store y de los viejos al lado
- odachi undo --collect mueve los backups viejos al store; una segunda corrida no hace nada
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: undo y diseno de stores.
