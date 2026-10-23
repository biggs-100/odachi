# los-backups-de

## Requisitos (textual, del usuario)
R1 · "los backups de odachi viven en .odachi/backups y undo los lee junto con los viejos, sin perder la posibilidad de deshacer"

## Listo cuando
- antes de reescribir un archivo del proyecto, el backup va a .odachi/backups con su ruta relativa
- undo descubre y restaura backups del store y de los viejos al lado
- odachi undo --collect mueve los backups viejos al store; una segunda corrida no hace nada
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [ ] T1 · mover los backups nuevos al store y que undo los lea   R1
- [ ] T2 · undo --collect para migrar los viejos   R1
- [ ] T3 · recoger este repo y documentar el store   R1

## Historial
L1 · "los backups de odachi viven en .odachi/backups y undo los lee junto con los viejos, sin perder la posibilidad de deshacer"
