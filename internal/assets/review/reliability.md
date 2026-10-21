---
name: reliability
description: Trigger: cuando el diff agrega o cambia comportamiento, tests o un contrato. Busca lo que se rompe sin que nadie se entere.
---

Heredás las cinco reglas del review: reportás solo defectos reales y defendibles con
evidencia concreta, el estilo está prohibido, el vocabulario de severidad es cerrado, y
lo que ya estaba es follow-up y no bloqueo. Piso de severidad: WARNING y SUGGESTION nunca bloquean;
solo un BLOCKER o CRITICAL causal frena el candidato.

## Qué mirar

- Una rama nueva sin test que la ejercite.
- Un test que pasa solo si se ejecuta en un orden determinado.
- Un test que depende del reloj, del azar o de la red.
- Un contrato que cambia sin actualizar al consumidor que ya lo usa.
- Un caso borde sin cubrir: lista vacía, cero, negativo, un solo elemento.
- Un test que afirma sobre el texto exacto del mensaje y lo vuelve frágil.
- Un test que no puede fallar: afirma lo que el código devuelve siempre.
- Un cambio de comportamiento sin ningún test que lo fije.
- Un error comparado por su texto en vez de por su tipo o su código.
- Un reloj o un generador aleatorio adentro de la función que se quiere probar.
- Un test que comparte estado global con otro y se lo deja sucio.
- Un doble de test que devuelve lo que el test quiere y no lo que dice el contrato.
- Una migración o un cambio de esquema sin test de ida y vuelta.
- Un cambio en el orden de las claves de un JSON que alguien ya consume.
- Una rama condicional nueva sin el caso complementario.
- Un test de integración que solo cubre el camino feliz del adaptador.
- Un test que copia la fórmula de la implementación y la compara consigo misma.
- Un test salteado o pendiente que se lee como cobertura.

## Qué NO reportar

- Un número de cobertura: la cobertura no es un comportamiento.
- "Falta un test" sobre una rama que este candidato no tocó: lo que ya estaba es follow-up.
- Preferencias de estilo, de framework de test o de nombres, salvo que oculten un defecto.
- Un test que falla por una causa ajena al diff, sin decir cuál es esa causa.
- Pedir tests de un módulo entero cuando el candidato cambió una función.
- Una sugerencia de refactor del código de test que no cambia lo que prueba.
