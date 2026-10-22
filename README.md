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

Esas tres líneas funcionan. La última release es `v0.9.1`: seis binarios (Linux,
macOS y Windows, en amd64 y arm64) y un `checksums.txt`. El instalador resuelve la
última versión por la API, verifica el sha256 y recién entonces escribe; un
checksum que no coincide sale con error y no deja **nada** en el destino. La
cadena completa está probada contra la release real, no contra un fixture: un
test arma el `tar.gz` y los checksums como los publica goreleaser, corre
`install.sh` sin red vía `file://`, y el binario instalado responde su versión.

Si preferís compilar: `go build ./cmd/odachi` (necesitás Go 1.25 o más
nuevo), y el binario va a decir `odachi dev`, porque la versión la inyecta la
release, no el código.

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

`proof/` y `.odachi/reviews/` son **evidencia generada**, no se versiona: la
prueba lleva el hash del árbol que la produjo y el veredicto se llama igual que
el árbol que juzga. Cada uno escribe su propio `.gitignore` la primera vez que lo
necesita, y eso tiene un precio que conviene saber: el primer `check --run` de un
proyecto invalida un review anterior y el primer `review` invalida una prueba
anterior, una sola vez cada uno.

`init` agrega dos cosas que no son artefactos: un bloque de reglas en `AGENTS.md`
—lo que el agente lee antes de tocar nada— y un hook de pre-commit que corre
`odachi check --quiet` sobre la prueba de este árbol.

## Las skills, en tu agente

`odachi pack` copia las ocho skills incluidas en el binario al lugar donde tu
agente las busca. Cada una se dispara con una situación, no con un comando:

| Skill | Cuándo se dispara |
| --- | --- |
| `acceptance` | hay que decidir si algo está terminado |
| `commits` | qué entra en un commit, cómo se parte, qué dice el mensaje |
| `data` | guardar, migrar, leer o borrar datos: dueño, vida y respaldo |
| `ship` | algo está listo para que otra persona lo use |
| `skill-craft` | escribir una skill nueva o arreglar una que no se usa |
| `slices` | partir un pedido grande, o cuando el código ya no entra en una lectura |
| `tests` | qué prueba un cambio: primero la que falla, después el código |
| `writing` | un mensaje, un PR, un comentario o documentación que alguien va a leer |

Los agentes soportados:

| Agente | `--agents` | Destino |
| --- | --- | --- |
| Claude Code | `claude` | `~/.claude/skills` |
| Codex | `codex` | `~/.codex/skills` |
| Cursor | `cursor` | `~/.cursor/skills` |
| Gemini | `gemini` | `~/.gemini/skills` |
| Copilot | `copilot` | `~/.copilot/skills` |
| Qwen | `qwen` | `~/.qwen/skills` |
| Kiro | `kiro` | `~/.kiro/skills` |
| **OpenCode** | `opencode` | `./.opencode/skills` (en el proyecto) |

Sin `--agents` instala en los agentes que detecta: los que tienen su carpeta de
configuración en tu home, y OpenCode si el proyecto tiene `.opencode/`. Nunca
borra nada, respalda lo que difiere y no repite lo que ya está igual.

### Probar odachi con OpenCode

```sh
mkdir mi-proyecto && cd mi-proyecto
odachi init                   # reglas del proyecto en AGENTS.md + hook pre-commit
odachi pack --agents opencode # las ocho skills en ./.opencode/skills/
odachi start                  # la entrevista: brief.md, cerco y primer task doc
odachi check                  # te dice si falta algo y cómo destrabarlo
```

Después de `pack`, abrí el proyecto con OpenCode: las skills están en
`.opencode/skills/<nombre>/SKILL.md`, que es donde las busca. `pack --agents
opencode` funciona igual en un proyecto que todavía no tiene `.opencode/`;
esa carpeta la crea el comando.

## El review: cuatro lentes, sin ceremonia

`odachi review` imprime el prompt de revisión del candidato —brief, rutas
cambiadas y diff— y elige las lentes por la evidencia del cambio: ninguna si
sólo se tocaron documentos, `reliability` si hay lógica, las cuatro si el diff
roza seguridad, identidad, dinero, la forma de los datos o la maquinaria que
los publica. No elegís la ceremonia; el cambio la pide.

```sh
odachi review                            # el prompt, para pegarlo en tu agente
odachi review --exec "agente {prompt}"   # o que lo corra: prompt por stdin y en {prompt}
odachi review --attach veredicto.json    # o guardá lo que contestó
odachi review --show                     # qué veredicto hay para este árbol
```

El veredicto vive en `.odachi/reviews/<hash-del-árbol>.json`, así que la
aprobación no puede flotar: el nombre del archivo *es* el estado del código que
se aprobó. Un `BLOCKER` causal bloquea, y para destrabarlo está
`odachi review --refute`: el refutador tiene que citar la línea del diff que
desmiente el finding, porque lo que no se puede citar no se refuta y lo que no se
refuta queda en pie. Para que el review sea obligatorio antes de commitear,
`odachi check --require-review`.

## Estado

Dominio puro en `internal/domain/...` con guardia de importaciones, adaptadores en
`internal/adapters/...`, y quince superficies: `start`, `ask`, `plan`, `level`,
`summary`, `undo`, `check`, `review`, `doctor`, `bench`, `init`, `skills`,
`skill new`, `pack` y `--version`. 35 paquetes, en verde en Linux, macOS y
Windows —ubuntu repite la suite interna con el detector de carreras—.

Publicado: `v0.9.1`. La instalación de una línea está probada contra la release
real, no contra un fixture. `install.ps1` se verificó a mano en Windows con la
release publicada: no hay test automatizado porque haría falta un listener HTTP
en CI.

La config de release **no tiene `brews`, firma ni provenance**: no hay tap ni
llaves, y fingir que están sería peor que no tenerlos.

## Licencia

MIT. Ver [LICENSE](LICENSE).
