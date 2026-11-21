# cuando-el-cambio
Base: 64b610a1441c93e7e683a34344fb43b899a05848
Light: yes

## Requirements (verbatim, from the user)
R1 · "cuando el cambio agrega un campo de struct, el undo nombra el campo y responde noticed en vez de unknown"

## Done when

## Steps
- [ ] T1 · cuando el cambio agrega un campo de struct, el undo nombra el campo y responde noticed en vez de unknown   R1

## History
L1 · "cuando el cambio agrega un campo de struct, el undo nombra el campo y responde noticed en vez de unknown"
L2 · OFICIO · "el undo reconoce los campos de struct siguiendo el ambito type ... struct, no adivinando desde una linea suelta" (porque el comment de AddedNames advierte que adivinar convertiria cualquier x = y en declaracion; un nombre falso mentiria en la nota del undo)
