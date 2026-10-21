---
name: readability
description: Trigger: cuando el diff es difícil de leer, tiene nombres que no dicen nada o creció de más. Busca lo que hace que el próximo cambio salga mal.
---

Heredás las cinco reglas del review: reportás solo defectos reales y defendibles con
evidencia concreta, el estilo está prohibido, el vocabulario de severidad es cerrado, y
lo que ya estaba es follow-up y no bloqueo. Piso de severidad: WARNING y SUGGESTION nunca bloquean;
solo un BLOCKER o CRITICAL causal frena el candidato.

## Qué mirar

- Un número mágico que debería ser una constante con nombre.
- Código muerto: un bloque comentado, una rama que nunca se alcanza.
- Una función que hace dos cosas y el nombre dice una.
- Un nombre que miente: `data`, `temp`, `aux`, `handle2`.
- Un parámetro booleano que cambia el comportamiento: `guardar(x, true)`.
- Un comentario que repite lo que el código ya dice.
- Un comentario que quedó viejo respecto del código que acompaña.
- Un bloque duplicado que ya existe en el mismo paquete.
- Una función de más de cincuenta líneas o anidada más de tres niveles.
- Un error que se devuelve sin decir qué falló.
- Una variable declarada lejos de donde se usa.
- Un `else` después de un `return`: una rama que se puede aplanar.
- Un archivo que mezcla capas: HTTP, SQL y regla de negocio en el mismo lugar.
- Un test cuyo nombre no dice qué comportamiento fija.
- Un cambio tan grande que no se puede leer por partes.
- Un pendiente anotado sin dueño y sin condición de cierre.
- Una abreviatura local en un nombre que se usa desde otros paquetes.
- Una constante exportada que nadie usa fuera del paquete.

## Qué NO reportar

- Lo que un formateador resuelve solo: espacios, comillas, orden de imports.
- Preferencias de nombres, de idioma de los comentarios o de estilo personal.
- "Esto es feo" sin un defecto concreto que el desorden esconda.
- Tamaño de una línea o de un archivo por sí mismo.
- Prosa sobre la arquitectura de todo el proyecto.
- Un problema de legibilidad que ya estaba antes del diff: es follow-up, no bloqueo.
