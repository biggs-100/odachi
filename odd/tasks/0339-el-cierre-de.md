# el-cierre-de
Base: 2add5afa9750c248c25077c87530ebea1be8fcb5

## Requirements (verbatim, from the user)
R1 · "el cierre de cada paso nombra las lineas autoradas acumuladas desde la base y el limite que el gate aplica al cerrar, sin negarse a cerrar"

## Done when
- plan --done NOMBRE el acumulado y el limite, y no inventa un numero cuando el documento no tiene base o el repositorio no esta conectado
- el unico cierre que se niega sigue siendo el de la feature, y la ADR dice por que el paso solo deja constancia
- go test ./... is green and odachi check --run is green and odachi skills --check is green

## Steps
- [ ] T1 · cerrar un paso nombra el acumulado desde la base y el limite que el cierre de la feature aplica   R1 · route:delegated/odachi-worker
- [ ] T2 · la ADR dice por que el paso deja constancia y el cierre de la feature se niega, con lo que hace gentle-shell   R1 · route:inline
- [ ] T3 · la skill de commits ensena el numero en el momento en que se puede partir, no solo al final   R1 · route:inline

## History
L1 · "el cierre de cada paso nombra las lineas autoradas acumuladas desde la base y el limite que el gate aplica al cerrar, sin negarse a cerrar"
