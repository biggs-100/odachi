---
name: commits
description: Trigger: cuando hay que decidir qué entra en un commit, cómo se parte el trabajo, o qué dice el mensaje y la descripción del PR.
---

## Cuándo

Cuando el cambio ya está y hay que commitearlo, cuando un diff no se puede revisar
de una sentada, o cuando el historial ya no se lee en orden.

## Reglas

- Un commit es un paso del task doc y una cosa que se puede revertir sola. Si el
  mensaje necesita un "y", son dos commits.
- El mensaje dice qué cambió **para quien usa el proyecto**, en una línea, en
  imperativo. El cómo y el por qué van en el cuerpo, no en el título.
- Los tests y la documentación del cambio viajan en el mismo commit. No existe un
  commit posterior llamado "tests" o "docs".
- Un refactor y un cambio de comportamiento no comparten commit: si algo se rompe,
  no sabés cuál de los dos fue.
- Antes de commitear, la prueba de este árbol: `odachi check --run "<comando>"`. El
  hook lo va a exigir igual, y es mejor enterarse antes.
- Un PR es un conjunto de commits que se leen **en orden**. Si no se lee en veinte
  minutos, se parte; si los commits se pisan entre sí, el orden está mal elegido.

## Nunca

- Commitear con tests rojos "para no perder el trabajo".
- Mezclar formato, reordenamiento o renombres con lógica.
- Un mensaje que no dice qué cambió: "fix", "wip", "cambios", "update".
- Commitear el trabajo de otra persona que ya estaba en el árbol.
- Un commit gigante "porque ya está todo junto".
