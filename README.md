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

Esas tres líneas funcionan. La última release es `v0.11.0`: seis binarios (Linux,
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
el árbol que juzga. Por eso los dos tienen que quedar fuera de git: si no, la
primera prueba invalidaría el review anterior y el primer review invalidaría la
prueba anterior.

`init` se ocupa de eso por vos, y de dos cosas más que no son artefactos: deja un
bloque marcado en el `.gitignore` del proyecto (`/proof/` y `/.odachi/reviews/`,
versionado, así que sobrevive a un clon) y un bloque de reglas en `AGENTS.md` —lo
que el agente lee antes de tocar nada— además del hook de pre-commit que corre
`odachi check --quiet` sobre la prueba de este árbol. En un proyecto sin `init`,
`check --run` y `review` se protegen solos con su propio `.gitignore` interno, al
precio de mover el árbol una vez.

Antes de reescribir un archivo tuyo, `odachi` guarda la versión anterior para que
`undo` pueda volver. Esos backups viven en `.odachi/backups/`, conservando la ruta
del artefacto, no al lado de `brief.md` ni de los task docs. Los que quedaron
viejos al lado se siguen leyendo, y `odachi undo --collect` los mueve al store sin
reescribirlos: conservan bytes y fecha, y una segunda corrida no hace nada.

## Las skills, en tu agente

Son diez skills, y `odachi init` ya deja nueve en el proyecto y una en tu home.
`odachi pack` copia las diez al lugar donde tu agente las busca, para todo lo
que `init` no hace: otro cliente, otra carpeta, otro scope. Cada una se dispara
con una situación, no con un comando:

| Skill | Cuándo se dispara |
| --- | --- |
| `acceptance` | hay que decidir si algo está terminado |
| `bootstrap` | arrancar un proyecto desde cero, o explicar qué es odachi |
| `commits` | qué entra en un commit, cómo se parte, qué dice el mensaje |
| `data` | guardar, migrar, leer o borrar datos: dueño, vida y respaldo |
| `drive` | trabajás dentro de un proyecto odachi: qué verbo va en cada momento y cómo cerrar una feature |
| `ship` | algo está listo para que otra persona lo use |
| `skill-craft` | escribir una skill nueva o arreglar una que no se usa |
| `slices` | partir un pedido grande, o cuando el código ya no entra en una lectura |
| `tests` | qué prueba un cambio: primero la que falla, después el código |
| `writing` | un mensaje, un PR, un comentario o documentación que alguien va a leer |

`bootstrap` es la de entrada, y es la única que vive en `~/.agents/skills` y no
en el proyecto: enseña a arrancar en una carpeta vacía, y en una carpeta vacía
todavía no hay proyecto donde ponerla. Las otras nueve son el oficio que un
proyecto lleva, y `init` las copia a `./.agents/skills`. Esa carpeta es la
convención que pide el estándar de Agent Skills, así que la encuentran los
clientes que odachi no conoce; también viaja con el repositorio, y por eso el
contenido de un repo ajeno es input no confiable. `bootstrap` solo explica el
flujo: no ejecuta nada.

`odachi init --skills home` deja las diez en `~/.agents/skills` y no escribe copia
en el proyecto; es para una sola copia por máquina. El default sigue siendo
`project`, porque el proyecto versiona el oficio y el catálogo queda más chico.

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

### Subagentes: el worker y el verificador

`init` también escribe dos subagentes nativos por cliente: `odachi-worker`
(implementa un paso acotado) y `odachi-verify` (verifica sin editar). No se eligen
con una tecla: los invoca el agente principal cuando `slices` se dispara. El padre
les pasa el task doc, el paso, las superficies de edición, los comandos y los
paths exactos de las skills; vuelven con evidencia y ≤2k tokens. El padre conserva
git, el gate y la visión; no hay un orquestador separado.

### Tres formas más de decir dónde

- `odachi pack --agents standard`: no un cliente, la carpeta que leen todos;
  `./.agents/skills` en el proyecto, `~/.agents/skills` en el home. Es la que
  `init` ya deja puesta.
- `odachi pack --agents opencode --scope home`: el mismo cliente, en la otra
  raíz (`~/.config/opencode/skills`), para tener una copia para todos tus
  proyectos en vez de una por proyecto.
- `odachi pack --agents claude --dest ./skills-x`: la carpeta que le digas, para
  un cliente que no está en la tabla. `--dest` y `--scope` no van juntos: los
  dos contestan "dónde", y dos respuestas no son ninguna.

`odachi pack --agents opencode` sigue existiendo y escribe en `./.opencode/skills`,
la carpeta nativa de OpenCode, para cuando la querés ahí y no en `.agents/skills`.
`init` no la borra ni la mueve: si la encuentra, la nombra y deja la del proyecto
como la de `.agents/skills`.

### Probar odachi con OpenCode

```sh
mkdir mi-proyecto && cd mi-proyecto
odachi init    # deja el proyecto listo de una sola vez
odachi start   # la entrevista: brief.md, cerco y primer task doc
odachi check   # te dice si falta algo y cómo destrabarlo
```

`odachi init` deja el proyecto listo en un paso: la historia de git si faltaba,
el bloque de reglas que el agente lee, el agente del proyecto y la puerta del home
(los dos se eligen con Tab), el bloque de `.gitignore`, el hook de pre-commit, las
nueve skills en `./.agents/skills`, la skill de arranque en `~/.agents/skills` y la
voz del perfil en el bloque de `AGENTS.md`.

Abrí el proyecto con OpenCode y ya está: lee `.agents/skills` como directorio
externo además del propio, así que ve las nueve sin `pack`. `pack` queda para
llevar las skills a la carpeta nativa de otro cliente, o a donde vos digas.

## La ceremonia y la voz

Dos cosas distintas, con dueños distintos.

El **perfil** es la ceremonia que el binario hace cumplir: cuánto pregunta,
cuánto imprime, qué cuenta como prueba. Vive en `.odachi/profile.toml`, lo
escribe `odachi start` con tus respuestas y lo lee el binario. El agente no lo
lee: no puede aflojar una compuerta ni cambiarla.

Sin archivo, el perfil por defecto es `novato`. `start` no lo sube a `senior`
solo porque haya `.git`: `init` es quien crea ese repo. `senior` se gana con
`odachi level --to`, o se infiere de un brief que la persona ya trajo.

La **voz** es cómo habla el agente: si te explica el porqué de cada decisión o si
va al hueso. Vive en la última sección del bloque de `AGENTS.md`, `## Cómo
hablás`, y en el agente nativo que espeja ese bloque. `odachi init` escribe la voz
de `novato` si todavía no hay perfil, y `odachi level --to <perfil>` cambia el
perfil, esa sección y el espejo juntas —y nada más del bloque—.

El agente nativo tiene dos copias. La del **proyecto** es el mismo bloque en la
carpeta donde tu cliente ya elige agentes: `.opencode/agents/odachi.md` (Tab en
OpenCode) y `.claude/output-styles/odachi.md` (`/config` en Claude Code). Es
generado: si lo editás a mano, `odachi init` lo respalda antes de reescribirlo y
`doctor` te avisa que dejó de coincidir. Medido en OpenCode 2.1.3: un proyecto con
`.git` lista el archivo como agente primario; una carpeta sin `.git` no lista
ninguno.

La copia del **home** es la puerta: `~/.config/opencode/agents/odachi.md` y
`~/.claude/output-styles/odachi.md`. Aparece en cualquier carpeta, no solo en la
que corrió `init`. Si la carpeta tiene el bloque de `odachi`, lo lee y adopta su
voz; si no, carga la skill de arranque y guía para empezar un proyecto. Es corta a
propósito: no repite las reglas, solo dice dónde están. Las dos personas por
perfil de la v0.10.0 quedaron como restos: `init` las nombra y no las borra, son
tuyas.

## El review: cuatro lentes, sin ceremonia

`odachi review` imprime el prompt de revisión del candidato —brief, rutas
cambiadas y diff— y elige las lentes por la evidencia del cambio: ninguna si
sólo se tocaron documentos, `reliability` si hay lógica, las cuatro si el diff
roza seguridad, identidad, dinero, la forma de los datos, un contrato público,
concurrencia o la maquinaria que lo publica. No elegís la ceremonia; el cambio
la pide.

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
refuta queda en pie.

El review es una decisión del humano, no del agente. Por defecto está apagado; el
switch del proyecto vive en `.odachi/review.toml` y se escribe con
`odachi review mode status|on|off`. Con `on`, `odachi check` pide el veredicto de
este árbol; con `off`, lo deja opcional. `--require-review` y `--no-review` fuerzan
una corrida sin cambiar el archivo. El archivo se versiona; cambiarlo mueve el
árbol, así que la prueba se retoma con `odachi check --run "<comando>"`. Un cambio
pasivo —sólo documentos— no pide veredicto ni con el modo prendido: no hay lente
que responder, así que no hay review que fabricar.

## Estado

Dominio puro en `internal/domain/...` con guardia de importaciones, adaptadores en
`internal/adapters/...`, y quince superficies: `start`, `ask`, `plan`, `level`,
`summary`, `undo`, `check`, `review`, `doctor`, `bench`, `init`, `skills`,
`skill new`, `pack` y `--version`. 38 paquetes, en verde en Linux, macOS y
Windows —ubuntu repite la suite interna con el detector de carreras—.

Publicado: `v0.11.0`. Lo que cambió en cada versión está en [CHANGELOG.md](CHANGELOG.md).
La instalación de una línea está probada contra la release
real, no contra un fixture. `install.ps1` se verificó a mano en Windows con la
release publicada: no hay test automatizado porque haría falta un listener HTTP
en CI.

La config de release **no tiene `brews`, firma ni provenance**: no hay tap ni
llaves, y fingir que están sería peor que no tenerlos.

## Licencia

MIT. Ver [LICENSE](LICENSE).
