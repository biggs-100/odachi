# odachi-init-agents

## Qué tiene que existir
- odachi init --agents opencode|claude|both elige que clientes reciben el agente nativo y la puerta del home, y la eleccion persiste para que .claude no vuelva a crearse

## Listo cuando
- odachi init --agents opencode no escribe nada en .claude/ y una segunda corrida conserva la eleccion
- doctor revisa solo los clientes elegidos y lo dice
- go test ./... queda en verde
- odachi check queda listo

## NO incluye (primera versión)
- no borrar archivos de clientes que ya existen: solo dejar de escribirlos
- el default de un proyecto sin archivo de eleccion sigue siendo los dos clientes

## Decisiones tomadas
- Se usa desde la compu.
- No hay datos sensibles en la primera versión.
- Primera versión usable: esta semana.
- Quiere aprender: como se guarda una decision de configuracion por proyecto.
