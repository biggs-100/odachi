# odachi-init-agents

## Requisitos (textual, del usuario)
R1 · "odachi init --agents opencode|claude|both elige que clientes reciben el agente nativo y la puerta del home, y la eleccion persiste para que .claude no vuelva a crearse"

## Listo cuando
- odachi init --agents opencode no escribe nada en .claude/ y una segunda corrida conserva la eleccion
- doctor revisa solo los clientes elegidos y lo dice
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · el vocabulario de clientes y el store .odachi/agents.toml leen y escriben both|opencode|claude   R1 · 08df48d
- [x] T2 · odachi init --agents escribe solo los clientes elegidos y guarda la eleccion   R1 · 29ede85
- [x] T3 · start y level refrescan el agente nativo solo de los clientes elegidos   R1 · 0ec3466
- [x] T4 · doctor revisa solo los clientes elegidos y lo dice en el chequeo 11   R1 · a380f30
- [x] T5 · recoger este repo con --agents opencode, documentar y verificar   R1 · 7002387

## Historial
L1 · "odachi init --agents opencode|claude|both elige que clientes reciben el agente nativo y la puerta del home, y la eleccion persiste para que .claude no vuelva a crearse"
L2 · OFICIO · "El default de --agents es both y la ausencia de .odachi/agents.toml tambien" (porque un proyecto que nunca eligio no puede cambiar de comportamiento ni ganar un archivo nuevo)
L3 · OFICIO · "La eleccion filtra el espejo del proyecto, la puerta del home y lo que refrescan start, level y doctor" (porque el cliente elegido es una sola decision; si cada verbo la interpretara distinto, el agente nativo volveria a aparecer donde el humano lo saco)
L4 · OFICIO · "Un store roto no se adivina: init, start y level se frenan y doctor lo reporta como nota" (porque elegir un cliente es una decision del humano; completarla con un default silencioso es lo que el flag existe para evitar)
