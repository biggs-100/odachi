# quiero-que-el
Base: e8a05d5bec428091219666fada2c4e16b80e3c41

## Requirements (verbatim, from the user)
R1 · "quiero que el usuario en la tui interactiva pueda seleccionar que perfil por default usar en el global y que cuando quiera trabajar con un perfil deferente solo en un proyecto, le pueda decir al agente con que perfil trabajar solo para ese proyecto, entonces seria mas facil que el usuario solo facil sin tanto problema"

## Done when
- the TUI voice screen offers machine or project scope
- machine scope writes the global default and leaves odd/profile.toml untouched
- project scope writes odd/profile.toml as it does today
- resolution order is project file, then AGENTS block voice, then global default, then inference
- doctor and the verbs use one resolver and agree on the profile
- tests pin the storage, resolution and TUI scope
- the changelog names the feature and the proof is green

## Steps
- [ ] T1 · guardar el default global y unificar la resolucion del perfil   R1 · route:inline
- [ ] T2 · la TUI elige maquina o proyecto al aplicar la voz   R1 · route:inline
- [ ] T3 · el changelog y el handover nombran la feature   R1 · route:inline
- [ ] T4 · go test ./... y check --run verdes en el arbol que cierra   R1 · route:inline

## History
L1 · "quiero que el usuario en la tui interactiva pueda seleccionar que perfil por default usar en el global y que cuando quiera trabajar con un perfil deferente solo en un proyecto, le pueda decir al agente con que perfil trabajar solo para ese proyecto, entonces seria mas facil que el usuario solo facil sin tanto problema"
