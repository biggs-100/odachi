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

## Instalación

Si no tenés Go, el instalador baja el binario de una release, verifica su
sha256 y lo deja listo. Los tres comandos:

```sh
# linux y macOS
curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.sh | bash

# windows (PowerShell)
irm https://raw.githubusercontent.com/biggs-100/odachi/main/scripts/install.ps1 | iex

# y para confirmar qué versión quedó
odachi --version
```

Honestidad sobre el estado de esto: **el pipeline de release todavía no corrió**.
Este repo no tiene remoto, así que no hay releases publicadas y esas tres líneas
todavía no pueden funcionar. Lo que sí está probado es el instalador: un test
construye un binario real con `-ldflags -X main.version=1.2.3`, arma el `tar.gz`
y el `checksums.txt` como los publicaría goreleaser, corre `scripts/install.sh`
sin red vía `file://` y verifica que el binario instalado diga `odachi 1.2.3`. El
mismo test prueba el caso feo: con un checksum alterado el instalador sale con
error y no deja nada instalado.

Si preferís compilar: `go build ./cmd/odachi` (necesitás Go 1.25 o más nuevo), y
el binario va a decir `odachi dev`, porque la versión la inyecta la release, no
el código.

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

Dominio puro en `internal/domain/...` con guardia de importaciones, adaptadores en
`internal/adapters/...`, y diez superficies funcionando: `start`, `ask`, `plan`,
`level`, `summary`, `check`, `init`, `skills`, `skill new`, `pack`, más
`--version`. 25 paquetes de tests en verde en Linux, macOS y Windows.

Publicado: `v0.2.0`, con seis binarios y `checksums.txt`. La instalación de una
línea está probada contra la release real, no contra un fixture: el script
resuelve la última versión por API, verifica el sha256, extrae e instala; un
checksum que no coincide no deja nada en el destino. `install.ps1` se verificó a
mano en Windows con la release publicada.

La config de release **no tiene `brews`, firma ni provenance**: no hay tap ni
llaves, y fingir que están sería peor que no tenerlos.

## Licencia

MIT. Ver [LICENSE](LICENSE).
