---
name: ship
description: Trigger: cuando algo está listo para que otra persona lo use. Verifica el camino del que instala, no el del que desarrolla.
---

## Cuándo

Cuando vas a decir "listo", publicar una versión, entregar a alguien que no tiene
tu entorno, o cerrar una feature.

## Reglas

- "Listo" es `odachi check` en verde sobre el árbol actual, con la salida real
  guardada en `proof/`. Todo lo demás es "creo que está listo".
- Probá el camino del que instala, no el del que desarrolla: el binario
  descargado, sin Go, sin tu PATH, en otra máquina si se puede.
- Verificá contra el artefacto real, no contra fixtures. Un instalador que solo
  anduvo con un archivo de mentira no está verificado.
- Decí qué NO se verificó y por qué. Ese párrafo es parte de la entrega, no una
  confesión.
- El changelog habla de lo que cambió para el humano, no de los archivos que
  tocaste.
- Si cambia el comportamiento de algo que ya existía, decilo primero y con todas
  las letras.

## Nunca

- Publicar con tests rojos, sin versión o sin checksum.
- Decir "listo" cuando solo compila.
- Cambiar un verbo existente sin avisar en el mismo release.
- Firmar como verificado algo que se probó a mano una sola vez.
- Etiquetar una versión desde un árbol que no es el que verificaste.
