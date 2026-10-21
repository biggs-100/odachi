# Odachi

**Dos manos: vos decidís, el agente construye. Y aprendés el oficio.**

## Qué es

- Una CLI que hace que el agente haga el trabajo sucio con disciplina de senior,
  mientras vos tomás las decisiones que importan: la visión, el alcance, lo que
  no se negocia.
- Un tutor: cada pregunta explica por qué existe, para que la persona deje de ser
  novata y no dependa del agente para siempre.
- Cuatro artefactos en texto plano que sirven de memoria: `brief.md`,
  `odd/tasks/NNNN-<slug>.md`, `proof/NNNN-<slug>.txt` y `.odachi/profile.toml`.

## Qué NO es

- No es un instalador de agentes ni un adaptador por cada harness.
- No es una TUI ni una aplicación con pantallas.
- No es una memoria ni una base de conocimiento propietaria: los artefactos son
  archivos que podés leer, editar y versionar.
- No manda telemetría, no abre la red, no corre en la nube.

## Los cuatro artefactos

| Archivo | Dueño | Para qué |
| --- | --- | --- |
| `brief.md` | humano | visión, no-objetivos, decisiones durables |
| `odd/tasks/NNNN-<slug>.md` | agente | requisitos R#, pasos e historial |
| `proof/NNNN-<slug>.txt` | agente | salida real del comando de prueba |
| `.odachi/profile.toml` | humano | 8 líneas de ceremonia, nunca de rigor |

Ejemplos completos en `testdata/examples/`.

## Estado

Bootstrap en progreso: dominio puro (`internal/domain/...`) con guardia de
importaciones, sin dependencias externas. Sin commits todavía.
