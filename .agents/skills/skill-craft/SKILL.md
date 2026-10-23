---
name: skill-craft
description: Trigger: cuando hay que escribir una skill nueva, corregir una que no se está usando, o decidir si algo merece ser skill.
---

## Cuándo

Cuando una regla se repite en cada sesión, cuando una skill existe pero el agente no
la carga, o cuando dudás entre escribir una skill, un test o un glob del cerco.

## Reglas

- Una regla accionable se puede **violar de forma observable**: "un retry sin backoff
  ni techo de intentos" se ve en el diff; "revisá la resiliencia" no.
- El trigger es lo único que decide si se carga. Escribilo como la situación del
  usuario —"cuando el diff toca permisos"—, no como el tema —"seguridad"—.
- Una regla por línea. Si no entra en una línea, son dos reglas.
- Lo que el modelo ya sabe (sintaxis, frameworks, patrones de moda) no va: ocupa
  contexto y se saltea.
- Lo que no se puede adivinar —tus convenciones, tus decisiones, lo prohibido— va al
  `brief.md` o al cerco, no a una skill.
- Si la regla se puede verificar, mejor un glob del cerco o un test **antes** que una
  skill: el cerco lo hace cumplir el binario, la skill la puede ignorar el modelo.
- Para crearla: `odachi skill new <slug>` deja la plantilla, y `odachi skills --check`
  dice si quedó bien formada.

## Nunca

- Glosarios, enciclopedias ni historia de la tecnología.
- Reglas que solo entiende alguien que ya sabe.
- Copiar una skill entera sin adaptarla a este proyecto.
- Pasar de 60 líneas: si no entra, es un documento, no una skill.
