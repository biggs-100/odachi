---
name: resilience
description: Trigger: cuando el diff agrega una llamada remota, una escritura, un reintento o un trabajo en background. Busca lo que se rompe cuando algo falla a mitad de camino.
---

Heredás las cinco reglas del review: reportás solo defectos reales y defendibles con
evidencia concreta, el estilo está prohibido, el vocabulario de severidad es cerrado, y
lo que ya estaba es follow-up y no bloqueo. Piso de severidad: WARNING y SUGGESTION nunca bloquean;
solo un BLOCKER o CRITICAL causal frena el candidato.

## Qué mirar

- Una llamada remota sin timeout: si el otro lado no contesta, esto se queda colgado.
- Un retry sin backoff y sin techo de intentos.
- Un retry sobre una operación que no es idempotente: duplica el efecto.
- Un fallo parcial que deja la mitad escrita y la otra mitad no, sin rollback.
- Un error de escritura o de cierre descartado con un guion bajo.
- Un defer de cierre que se ignora y deja el archivo o la conexión abiertos.
- Una transacción que no cubre todas las escrituras del mismo caso de uso.
- Un servicio que arranca y atiende aunque su dependencia no esté lista.
- Un camino de error que contesta 200 con el error adentro del cuerpo.
- Un fallback que devuelve vacío y el llamador no distingue "vacío" de "falló".
- Un trabajo en background que no se puede frenar al apagar el proceso.
- Un lock tomado sin vencimiento: si el proceso muere, nadie puede volver a entrar.
- Un reintento del humano que duplica el efecto: dos turnos, dos cobros, dos mails.
- Una cola o un buffer sin límite: crece hasta que se come la memoria.
- Un log de error sin el dato que permite reproducirlo: id, ruta, instante.
- Una degradación silenciosa: se usa un valor por defecto y nadie se entera.
- Un timeout más corto que el tiempo normal de la operación que envuelve.
- Un contexto que no se propaga a la llamada de abajo: cancelar arriba no cancela abajo.

## Qué NO reportar

- Arquitectura que este candidato no cambió: una cola que faltaría es otro trabajo.
- Preferencias de estilo, de nombres o de estructura, salvo que oculten un defecto.
- Un riesgo teórico sin una ruta concreta que lo dispare en este diff.
- Un retry o un timeout que ya estaban antes del candidato: lo que ya estaba es follow-up.
- Que convendría un circuit breaker, un cache o un servicio aparte.
- Un mensaje de log en otro idioma o con otro formato.
