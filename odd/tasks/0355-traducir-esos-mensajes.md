# traducir-esos-mensajes
Base: 1f566c168663610a329edd337206aa21db88acba

## Requirements (verbatim, from the user)
R1 · "traducir esos mensajes al lenguaje del perfil"

## Done when
- with the office-peru profile, odachi check prints the observed gate messages in plain Spanish without developer jargon
- with novato, senior or ci the human output stays byte for byte as it is today
- the --json payload keeps the same keys and values for every profile
- tests pin both the office-peru rendering and the unchanged default rendering
- the changelog names the feature and the proof is green on the closing tree

## Steps
- [x] T1 · el traductor de lineas de check con pruebas de ida y vuelta   R1 · route:inline · tree:566216a51569ccd608cd8ca63ec5e6f80728e8af
- [x] T2 · runCheck carga el perfil y pasa la voz sin tocar el JSON   R1 · route:inline · tree:566216a51569ccd608cd8ca63ec5e6f80728e8af
- [x] T3 · el changelog nombra la feature y la investigacion su limite   R1 · route:inline · tree:2b367576abdd19e0e48f584c087d098fbfb5f038
- [x] T4 · go test ./... y check --run verdes en el arbol que cierra   R1 · route:inline

## History
L1 · "traducir esos mensajes al lenguaje del perfil"
