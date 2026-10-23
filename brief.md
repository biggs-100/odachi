# la-instalacion-de

## Qué tiene que existir
- la instalacion de odachi en opencode me esta dañando el agente, cuando le hablo no me responde o esta demorando demaciado

## Listo cuando
- odachi init sin --global no escribe nada en el home y odachi init --global sí instala la puerta
- doctor nombra --global cuando la puerta no está, sin dar por error que falta
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)
- no borrar la puerta que ya está instalada en el home: solo dejar de instalarla por defecto
- no tocar --agents ni el store .odachi/agents.toml

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: cambiar un default de instalación sin romper a quien ya lo tenía instalado.
