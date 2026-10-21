---
name: risk
description: Trigger: cuando el diff toca seguridad, permisos, datos sensibles, dinero o dependencias. Busca lo que puede filtrar, escalar o romper.
---

Heredás las cinco reglas del review: reportás solo defectos reales y defendibles con
evidencia concreta, el estilo está prohibido, el vocabulario de severidad es cerrado, y
lo que ya estaba es follow-up y no bloqueo. Piso de severidad: WARNING y SUGGESTION nunca bloquean;
solo un BLOCKER o CRITICAL causal frena el candidato.

## Qué mirar

- Una consulta a la base armada concatenando entrada del request en vez de parámetros.
- Un endpoint nuevo sin verificación de permiso en el servidor.
- Un chequeo de permiso que solo existe en el frontend.
- Un id que llega del cliente y decide de quién es el dato que se lee o se escribe.
- Un secreto, token o contraseña leídos de un archivo que el repositorio versiona.
- Un log o un mensaje de error que imprime un token, una contraseña o un dato personal.
- Una respuesta que devuelve más campos de los que la pantalla usa: hash, DNI, teléfono.
- Un token sin verificar la firma, o con expiración infinita.
- Un HTML armado con datos del usuario sin escapar.
- Un archivo subido que se guarda con el nombre que mandó el cliente.
- Una redirección a una URL que llega por parámetro.
- Un comando de shell armado con entrada del usuario.
- Un mensaje del driver devuelto tal cual en la respuesta HTTP.
- Una dependencia nueva sin justificación, o que pide ejecutar un script al instalarse.
- Un permiso de archivo más abierto que lo necesario: 0777, mundo escribible.
- Una migración que agrega una columna obligatoria sin default sobre una tabla con filas.
- Un rol comparado contra un string que viene del request del usuario.
- Un `.env` de ejemplo con valores reales adentro.

## Qué NO reportar

- Preferencias de estilo, de nombres o de estructura: si no ocultan un defecto, no van.
- Endurecimiento genérico sin un ataque concreto: "podría ser más seguro" no es un finding.
- Un problema que ya estaba antes del diff: lo que ya estaba es follow-up, no bloqueo.
- Falta de HTTPS, de rate limit o de WAF en algo que corre en la máquina de una persona.
- Dependencias que ya estaban declaradas antes de este candidato.
- Un secreto inventado dentro de un test, que no puede llegar a producción.
