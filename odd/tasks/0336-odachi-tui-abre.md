# odachi-tui-abre
Base: e1a819998db97fd120e6f7dd7b23df19000ae2d8

## Requirements (verbatim, from the user)
R1 · "odachi tui abre en un hub con menu y router declarativo, y desde ahi el humano edita los seis ejes del perfil, no solo su nombre"

## Done when
- odachi tui abre en un menu con las acciones de instalar y perfil, y esc vuelve al menu desde cualquier seccion
- elegir un eje muestra el valor en vigor y permite cambiarlo entre los que el perfil declara, y aplicar escribe odd/profile.toml con ese valor y un backup
- un perfil que no se toco queda byte a byte identico, y cancelar antes de aplicar no escribe nada
- la navegacion es una tabla declarativa: una pantalla nueva es una fila, y una prueba recorre la tabla
- go test ./... is green and odachi check --run "go test ./..." is green and odachi skills --check is green

## Steps
- [x] T1 · el perfil gana un escritor de un eje: level escribe el campo pedido y conserva los demas   R1 · tree:e71ea2561bd69301cc01174d68e738892b509678
- [x] T2 · la navegacion del tui es una tabla declarativa y una prueba la recorre   R1 · tree:1a85ac208fd96faaa3bbfb07d918e1d5271a3e34
- [x] T3 · el hub abre en un menu con instalar y perfil, y esc vuelve al menu desde cualquier seccion   R1 · tree:681bd9ddc390044a0340052ecdfff330ffbbba6c
- [ ] T4 · el editor de perfil cambia un eje y el aplicar escribe el archivo con su backup   R1

## History
L1 · "odachi tui abre en un hub con menu y router declarativo, y desde ahi el humano edita los seis ejes del perfil, no solo su nombre"
