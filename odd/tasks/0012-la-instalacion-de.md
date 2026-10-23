# la-instalacion-de

## Requisitos (textual, del usuario)
R1 · "la instalacion de odachi en opencode me esta dañando el agente, cuando le hablo no me responde o esta demorando demaciado"

## Listo cuando
- odachi init sin --global no escribe nada en el home y odachi init --global sí instala la puerta
- doctor nombra --global cuando la puerta no está, sin dar por error que falta
- go test ./... queda en verde
- odachi check queda listo

## Pasos
- [x] T1 · el flag --global de init decide si se escribe la puerta del home, y sin el no se toca el home   R1 · 7d18bef
- [x] T2 · doctor nombra --global cuando la puerta no esta y no la da por faltante   R1 · 805a0c5
- [x] T3 · la ayuda, el README y el CHANGELOG dicen que la puerta es opt-in   R1 · 2e98f86
- [x] T4 · go test ./... y odachi check en verde, y el repo se re-init sin --global   R1 · 746e94c

## Historial
L1 · "la instalacion de odachi en opencode me esta dañando el agente, cuando le hablo no me responde o esta demorando demaciado"
L2 · OFICIO · "La puerta del home es opt-in con --global" (porque un agente primary en toda carpeta es una decision del humano; instalarlo como efecto de init secuestra el selector de clientes que no lo pidieron)
