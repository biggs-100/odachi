# una-medicion-es
Base: 1733e6c7a2485cc66b782233272fc5c9ba1cb82a

## Requirements (verbatim, from the user)
R1 · "una medicion es un documento con forma escrita antes de correr (criterio de paso, tabla de veredictos, comandos exactos, corridas excluidas y lo que el dato no muestra) y el traspaso cita la linea que el tool cuenta en vez de afirmar un numero"

## Done when
- odachi skills lista measure y el skill nombra las cinco secciones con lo que cada una prueba y su limite
- SKILLS.md nombra measure con su trigger, y la prueba de embed encuentra las cinco secciones en el texto
- odachi summary imprime la misma linea de totales que hoy, y una sola funcion la rinde para los dos lectores
- odachi doctor compara la linea de totales que docs/next-session.md cita contra la que cuenta, y nombra las dos cuando difieren
- si el archivo no cita una linea de totales, doctor avisa en vez de pasar
- docs/next-session.md cita la linea del tool y doctor queda en 15 checks sin problemas
- go test ./... is green and odachi check --run "go test ./..." is green
- ## Unreleased names it, patch (0298)

## Steps
- [x] T1 · el skill measure: las cinco secciones, su limite y el registro al dia   R1 · route:inline · tree:d85e1d7992d54dc161d559a6603505738c62bd7a
- [x] T2 · una sola funcion rinde la linea de totales, y summary la imprime igual   R1 · route:inline · tree:82a58482bd64c4939e959df6b0f58174cf9e61d2
- [ ] T3 · check 14 recalcula la linea que el archivo cita y nombra las dos cuando difieren   R1 · route:inline
- [ ] T4 · docs/next-session.md cita la linea del tool y doctor queda limpio   R1 · route:inline
- [ ] T5 · el changelog nombra 0298 bajo Unreleased, patch   R1 · route:inline

## History
L1 · "una medicion es un documento con forma escrita antes de correr (criterio de paso, tabla de veredictos, comandos exactos, corridas excluidas y lo que el dato no muestra) y el traspaso cita la linea que el tool cuenta en vez de afirmar un numero"
L2 · OFICIO · "el skill measure se escribe a mano en skills/ y no con odachi skill new" (porque skill new escribe .agents/skills, que es la copia que un cliente lee en un proyecto adoptado; el payload de odachi vive en skills/<name>/SKILL.md y es el que embebe el binario, asi que crearlo en el otro lado dejaria el arbol con un skill que nadie embebe)
