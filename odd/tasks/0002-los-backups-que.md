# los-backups-que

## Requisitos (textual, del usuario)
R1 · "los backups que generan plan y ask no ensucian el arbol: el bloque de .gitignore ignora *.bak-*"

## Listo cuando
- despues de correr plan y ask --craft, git status queda limpio
- un proyecto existente recibe el bloque actualizado al correr odachi init
- go test ./... queda en verde

## Pasos
- [x] T1 · agregar *.bak-* al bloque de .gitignore   R1 · ccaafe5

## Historial
L1 · "los backups que generan plan y ask no ensucian el arbol: el bloque de .gitignore ignora *.bak-*"
L2 · OFICIO · "El bloque de .gitignore es del proyecto, no del store: *.bak-* cubre task docs, brief, perfil y fence, y undo sigue leyendo los backups" (porque el backup es material de undo, no contenido del proyecto; ignorarlo mantiene el arbol limpio sin mover la maquinaria de undo)
