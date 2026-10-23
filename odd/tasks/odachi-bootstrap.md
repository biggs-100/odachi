# odachi · bootstrap

Feature: a CLI that makes the agent do the dirty work with senior discipline while the human decides vision, and that teaches the human until they no longer need the teaching.
Status: cerrada · Branch: main · Owner: human (vision) / agent (execution) · Cerrada en a9b9057, publicada en v0.10.0

Este documento es la **historia y la especificación** de la feature de arranque:
los diez requisitos (S1-S10), las treinta y cuatro tareas (T1-T34) y la bitácora
completa (L1-L84). Se lee para entender por qué odachi es como es; no se le
agregan tareas nuevas.

El próximo trabajo abre su propio documento con la convención del producto:
`odd/tasks/NNNN-<slug>.md`, un archivo por feature, con los headers que escribe
`odachi start`: `## Requisitos (textual, del usuario)`, `## Listo cuando`,
`## Pasos` y `## Historial`. Este archivo es la spec histórica y usa
Specs/Tasks/Log, así que `odachi summary`, `plan` y `level` no lo leen; se queda
donde está porque cinco comentarios del código y el mapa de impacto lo citan por
su ruta.

## Specs

S1 · "cualquier persona sin experiencia pueda usar sus agentes y poder hacer proyectos como un profesional sinior"
S2 · "que sea organico e intuitivo para las personas que deleguen las tareas a los agentes"
S3 · "el agente haga todo el trabajo susio, con toda la disciplina de un sinior, y la persona decida las cosa cognitivas, la vision del proyecto, etc"
S4 · "tambien las personas que ya son experimentadas"
S5 · "dos tipos de perfiles, novato y sinior, e incluso que sea facil de crear nuevos perfiles, para que las personas decidan"
S6 · "si me gustarian los artefactos odd, porque tambien podrian servir como memorias o historial para que el agente entienda rapido"
S7 · "el agente tambien va a servir como tutor para que la persona pueda llegar a ser un senior, para que una persona no se quede para siempre como novato"
S8 · nombre del proyecto: "odachi"
S9 · "que se pueda implementar en cualquier agente si problemas en un futuro"
S10 · "mejoremos nuestro odachi para que no haya problemas al momento de querer usarlo"

## Contract

### A. The four artifacts

```text
brief.md                    vision + non-goals + durable decisions   (human-owned)
odd/tasks/NNNN-<slug>.md    Requirements R# · Tasks · History       (agent writes, human reads)
proof/NNNN-<slug>.txt       real output of the proof command        (agent writes)
.odachi/profile.toml        8 lines, read by the BINARY ONLY         (never reaches the model)
```

`brief.md` canonical headers (exact, in Spanish, because the user reads it):

```markdown
# <name>

## Qué tiene que existir
- ...

## Listo cuando
- ...   (observable, one line each)

## NO incluye (primera versión)
- ...

## Decisiones tomadas
- ...
```

Task doc canonical headers:

```markdown
# <slug>

## Requisitos (textual, del usuario)
R1 · "..."
R2 · "..."

## Listo cuando
- ...   (R1)

## Pasos
- [x] T1 · ...   R1 · 4a1c9f
- [ ] T2 · ...   R2

## Historial
L1 · "..."
```

`proof/NNNN-<slug>.txt` = verbatim stdout+stderr of the proof command for one tree.

`.odachi/profile.toml`:

```toml
name        = "novato"
extends     = ""
brief_source = "interview"   # interview | repo | file
questions    = "one-by-one"  # one-by-one | batched | all
assumptions  = "never"       # never | declare | prefer-safe
output       = "line"        # line | detail | json
proof        = "command"     # command | suite-or-ci
explain      = "why"         # why | none
```

### B. Hard boundary — a profile changes ceremony, never rigor

A profile MAY set: brief source, question pacing, assumption policy, output verbosity, what counts as proof, explanation level.

A profile MAY NEVER set: whether a brief exists, whether a vision decision is recorded, whether proof is mandatory, whether the fence runs, whether `check` returns an exit code. This is enforced structurally: `profile.Profile` has no such field, and a reflection guard test fails the build if one is added.

### C. `check` contract (exit codes, Spanish messages)

| Code | Meaning |
|---:|---|
| 0 | `ready` — proof exists for the exact tree |
| 1 | `needs work` — the agent can act |
| 2 | `needs decision` — only the human can unblock |

Every blocked result carries `Next`: a command that resolves the block. A test asserts `Next` is non-empty and starts with `odachi ` for every blocked reason — no dead-end messages.

### D. Domain API (pure: stdlib + sibling domain packages only)

```go
package fence
type Rule struct{ Not string; Paths []string; Deps []string }
type Violation struct{ Rule, What, Kind string }
const KindPath = "path"; const KindDep = "dep"
func Check(rules []Rule, changedPaths, addedDeps []string) []Violation
// pattern without "/" matches ANY path segment (payment* matches internal/payments/x.go)
// pattern with "/" matches the whole path, where ** matches zero or more segments
// matching is case-insensitive (deterministic on Windows and Linux)
// dep matches when ANY dependency segment equals the pattern (stripe matches github.com/stripe/stripe-go/v76)

package gate
type Code int
const Ready Code = 0; const NeedsWork Code = 1; const NeedsDecision Code = 2
const (
    ReasonReady = "ready"; ReasonNoBrief = "no_brief"; ReasonBriefIncomplete = "brief_incomplete"
    ReasonOutOfScope = "out_of_scope"; ReasonOpenQuestion = "open_question"
    ReasonProofMissing = "proof_missing"; ReasonProofStale = "proof_stale"
)
type State struct {
    HasBrief bool; BriefMissing []string
    TreeHash, ProofTreeHash, ProofOutput string
    OpenQuestions int; Violations []string
}
type Result struct{ Code Code; Reason, Message, Next string }
func Evaluate(s State) Result
// order: no brief -> brief incomplete -> out of scope -> open question -> proof missing -> proof stale -> ready
// messages (Spanish, one line): see the tests in the first work unit

package profile
type Profile struct{ Name, Extends, BriefSource, Questions, Assumptions, Output, Proof, Explain string }
func Validate(p Profile) error
func Resolve(known map[string]Profile, name string) (Profile, error) // child wins over parent; cycles are an error
type Facts struct{ HasBrief, HasRepo, Machine bool }
func Infer(f Facts) string // Machine->ci; HasBrief||HasRepo->senior; else novato
func Builtin() map[string]Profile // novato (extends "") / senior (extends novato) / ci (extends senior)
```

### E. Non-negotiables

- Module path `github.com/biggs-100/odachi`; zero external dependencies in this session.
- Code, identifiers, comments, docs and commits in English. Messages the USER reads are Spanish (see birth decision 1).
- Guard test on import boundaries exists BEFORE domain code.
- Test-first: the three domain test files are written and observed RED before the implementation files.

## Contract II · T3 and T4

### T3 — `internal/domain/brief` and `internal/domain/taskdoc` (pure)

```go
package brief
type Brief struct{ Name string; Outcome, DoneWhen, NonGoals, Decisions []string }
func Parse(md string) Brief       // canonical Spanish headers; unknown sections ignored; never panics
func (b Brief) Missing() []string // Spanish labels of empty required parts, e.g. "qué tiene que existir", "no incluye"
// required: "Qué tiene que existir", "Listo cuando", "NO incluye (primera versión)"

package taskdoc
type Requirement struct{ ID, Text string }
type Step struct{ ID, Text, Requirement, Commit string; Done bool }
type TaskDoc struct{ Title string; Requirements []Requirement; DoneWhen []string; Steps []Step; History []string }
func Parse(md string) TaskDoc
func (d TaskDoc) Next() (Step, bool)                        // first not-done step, in order
func (d TaskDoc) Progress() (done, total int)
func (d TaskDoc) Record(id, commit string) (TaskDoc, error) // marks done, records the commit; unknown id is an error
func (d TaskDoc) Render() string                            // canonical, stable, for round-trip
```

- Round-trip: `Parse(Render(Parse(x)))` equals `Parse(x)` for both `testdata/examples/`.md files.
- Tolerate an agent appending history lines; never panic on malformed input.

### T4 — adapters (may carry dependencies; the domain may not)

```go
package profiletoml // internal/adapters/profiletoml
func Decode(data []byte) (profile.Profile, error) // STRICT: unknown keys are an error naming the key
func Load(path string) (profile.Profile, error)

package fsx // internal/adapters/fsx
func Digest(data []byte) string                                    // sha256 hex
func WriteAtomic(path string, data []byte, perm fs.FileMode) error // temp in same dir + rename
func Backup(path string) (string, error)                           // path + ".bak-<digest8>"

package gitx // internal/adapters/gitx
func TreeHash(root string) (string, error)            // the CURRENT worktree, not just HEAD
func ChangedPaths(root, base string) ([]string, error)
func AddedDeps(root, base string) ([]string, error)   // module paths added to go.mod between base and worktree
```

- `TreeHash` represents the current worktree through a THROWAWAY index: `git read-tree HEAD` (or `read-tree --empty` when there is no commit yet), `git add --all -- .`, `git write-tree`, with `GIT_INDEX_FILE` pointing inside a temp directory removed afterwards. The real index, the worktree and the human's git config are never touched; untracked files are included and git-ignored files are not. (`git stash create` was the first choice and is wrong twice: it silently ignores untracked files, so a new file left the hash unchanged, and it fails before the first commit.)
- `gitx` runs git via `exec.Command` with `-C root`; no network; errors include git stderr.
- `profiletoml` uses `github.com/BurntSushi/toml` and rejects a non-empty `md.Undecoded()`. Only `go.mod` gains the dependency; the domain stays stdlib-only and the boundary guard must stay green.

## Contract III · T5 (`odachi start`)

### Surface

```text
odachi                 == odachi start
odachi start [flags]
  --outcome s      answer 1     --audience s   answer 2: phone|cpu|both
  --done s         answer 3 (repeatable)      --exclude s  answer 4 (repeatable: preset key or free text)
  --sensitive s    answer 5: none|names|health|money
  --deadline s     answer 6: today|week|relaxed
  --learn s        answer 7 (repeatable)
  --brief path     read an existing brief instead of asking (senior/ci)
  --force          overwrite an existing brief.md
  --json           machine output
```

Exit codes: `0` artifacts written · `1` blocked (invalid flag, or `brief.md` exists without `--force`) · `2` needs a human answer (non-interactive without the required flags; the message names the missing flags).

### The seven questions (verbatim, each with the why the user sees)

| # | flag | question | why |
|---|---|---|---|
| 1 | outcome | ¿Qué querés que exista cuando terminemos? (una frase, en tus palabras) | sin esto no hay nada que construir; y en tus palabras es lo que leés después para saber si vamos bien |
| 2 | audience | ¿Quién lo usa y desde dónde? (1 celular · 2 compu · 3 los dos) | cambia todo lo que viene después, y es la decisión más cara de cambiar tarde |
| 3 | done | ¿Cómo vas a saber que funcionó? (te propongo tres; podés escribir la tuya) | esto se llama criterio de aceptación: es lo que separa "anda" de "creo que anda" |
| 4 | exclude | ¿Qué NO querés en esta primera versión? (elegí las que quieras) | es lo que evita 40 cosas a medio hacer; un senior dice no, y acá el cerco lo hace cumplir |
| 5 | sensitive | ¿Hay datos sensibles? (1 ninguno · 2 nombres · 3 salud · 4 plata) | define qué no puede salir del proyecto desde el día uno |
| 6 | deadline | ¿Cuánto tiempo tenés para la primera versión usable? (1 hoy · 2 esta semana · 3 sin apuro) | define cuánto recortamos; un proyecto que no se termina no enseña nada |
| 7 | learn | ¿Qué te gustaría aprender en el camino? (elegí una o dos) | el tutor necesita saber qué explicarte más; esto es lo que te saca de novato |

The why is printed with the question, in the same block, never as a second message. One line, never a paragraph.

### Files written (all through `fsx.WriteAtomic`; existing file → `fsx.Backup` first)

1. `brief.md` — rendered from the answers, canonical headers from Contract A.
2. `.odachi/profile.toml` — the resolved profile, one field per line, `name` always written.
3. `.odachi/fence.toml` — one `[[rule]]` per selected non-goal:

| key | `paths` | `deps` |
|---|---|---|
| pagos | `payment*`, `**/payments/**` | `stripe`, `mercadopago` |
| login | `**/login*`, `**/auth/login*` | — |
| historia-clinica | `**/historia*`, `**/medical-record*` | — |
| recordatorios | `**/reminder*`, `**/whatsapp*` | — |
| (free text) | `<slug>*` | — |

4. `odd/tasks/<NNNN>-<slug>.md` — `NNNN` is one more than the highest existing number (0001 when empty); `<slug>` is the first three words of the outcome, ASCII lowercase joined by `-`. Body: `## Requisitos (textual, del usuario)` with `R1 · "<outcome verbatim>"`, `## Listo cuando` from answer 3, `## Pasos` empty, `## Historial` with `L1 · "<outcome verbatim>"`.

   Rerun rule: when a task document whose filename ends with `-<slug>.md` already exists, reuse it and do **not** rewrite it — its `Pasos` and `Historial` are the memory of work already done. A different outcome gets the next number. One feature is one document.

### Layer and seam

- `internal/app/start` is pure orchestration over ports; it must not import `os` for I/O beyond nothing: reading and writing go through small ports it declares (`Interview`, `Files`), wired in `cmd/odachi`.
- `Interview` seam: `Ask(q Question) (string, error)` where `Question` carries text, why, and options. Test double: `ScriptedInterview` answering from a queue.
- `profile.Infer(Facts{HasBrief, HasRepo, Machine})` decides pacing: `one-by-one` asks each question alone; `batched` asks all, then shows the answers together; `all` never asks and requires the flags or `--brief`, failing with exit 2 that names the missing flags.
- New domain function `brief.Render(b Brief) string` (canonical, `Parse(Render(b))` round-trips). The domain stays stdlib-only: the boundary guard must stay green.

### Tests (test-first, RED captured before implementation)

- `brief.Render` round-trip and header order.
- `app/start`: scripted interview produces exactly the four files with the expected content (assert the brief sections, the profile `name`, one fence rule per selected non-goal, and the task doc `R1` quoting the outcome verbatim).
- Re-running with `--force` backs up the old `brief.md` and writes the same bytes for the same answers (idempotent).
- Re-running without `--force` exits 1 and names `--force`.
- Non-interactive without `--outcome` exits 2 and the message names `--outcome`.
- Empty answer to question 1 is rejected with a Spanish message that asks again, never silently accepted.
- Slug and task-doc numbering: empty dir → `0001-…`; existing `0007-x.md` → `0008-…`.
- `cmd/odachi`: `odachi` with no args equals `odachi start`; unknown command exits 1 with the available commands.

## Contract IV · T6 (`ask`, `level`, `summary`)

### Surface

```text
odachi ask "<pregunta>"                          record an open vision question (the agent does this)
odachi ask --answer "<respuesta>" [--for P<n>]  record the human's answer (needs exactly one open question)
odachi ask --craft "<decisión>" --because "<razón>" record a craft decision the agent took alone
odachi level                                      show the facts and, when earned, propose senior
odachi level --to <profile>                       apply a profile (backup + atomic write), print what changes
odachi summary                                    derived report of every feature: steps, decisions, open questions
```

Exit codes: `0` ok · `1` blocked (no task document, unknown profile, no open question, or two open questions without `--for`) · `2` only when a human answer is required and cannot be inferred. Every blocked message names the command or flag that unblocks it.

### The history record (append-only, inside the task document)

No new state file: the task document's `## Historial` is the log. Three deterministic line shapes, `L<n>` monotonic:

```text
L7 · PREGUNTA P1 · "¿Los turnos se pueden cancelar?"
L8 · VOS P1 · "No en la primera versión"
L9 · OFICIO · "Guardé los turnos en SQLite" (porque dos personas no pueden escribir el mismo archivo)
```

`P<n>` is monotonic per task document. An open question is a `PREGUNTA` line with no matching `VOS` line. `OFICIO` lines are the agent deciding alone and saying why — they are the tutor material, and they are never questions.

### Domain additions (pure, stdlib only)

```go
package taskdoc
// History is append-only: AppendHistory never rewrites an existing line.
func (d TaskDoc) AppendHistory(line string) TaskDoc
func (d TaskDoc) Events() []Event            // parsed, in file order
func (d TaskDoc) NextLine() int              // next L number
func (d TaskDoc) NextQuestion() int          // next P number
func (d TaskDoc) OpenQuestions() []Question  // PREGUNTA without VOS
func (d TaskDoc) CloseQuestion(id, answer string) (TaskDoc, error) // unknown id or already answered is an error
type Event struct{ Line int; Kind string; ID, Text, Because string }  // Kind: question|answer|craft|note
type Question struct{ ID, Text string }

package profile
// Next is the graduation proposal: earned, never automatic.
func Next(current string, e Evidence) (name string, ok bool)
func Apply(current Profile, to Profile) (Profile, []Change)  // Change{Field, From, To} for the message
type Evidence struct{ Features, VisionAnswers, CraftDecisions int }
// next is "senior" when Features >= 3 and VisionAnswers >= 5, and only when current == "novato".
```

### App layer

- New shared helper `internal/app/tasks`: `Current(files, dir)` returns the highest-numbered task document (path + parsed doc) and a human Spanish error when there is none, and `All(files, dir)` returns every task document sorted by number.
- `internal/app/ask`, `internal/app/level`, `internal/app/summary` orchestrate over the existing `Files` port (`Exists`, `Read`, `List`, `Write`, `Backup`) and return the same `Result` shape as `start` so `cmd/odachi` prints one line or JSON.
- `level` reads every task document, counts `Features` (documents with at least one done step), `VisionAnswers` (answer events) and `CraftDecisions` (craft events), and shows: profile, features, preguntas contestadas, decisiones de oficio, open questions. When the proposal is earned it prints exactly what changes and how to accept it. It never rewrites the profile on its own.
- `level --to <profile>` resolves through `profile.Builtin()` plus the current file, backs up `.odachi/profile.toml`, writes it atomically, and prints the field changes in Spanish. A profile that does not exist is blocked with the available names.
- `summary` prints per feature: slug, `R1` verbatim, steps done/total, open questions, and the last three history lines. It ends with one line of totals. It invents no narration: the concepts the week brought are the agent's job (T8), not the binary's.

### Tests

- Domain RED first: history parsing (all three shapes, malformed lines become `note`), `AppendHistory` is append-only, `NextLine`/`NextQuestion` monotonic, `OpenQuestions`/`CloseQuestion` incl. unknown and already-answered, `profile.Next` thresholds and the `current != novato` refusal, `profile.Apply` change list.
- App: `ask` with no task document is blocked naming `odachi start`; two open questions need `--for`; `--answer` closes the right one; `--craft` requires `--because`; `level --to senior` backs up before writing and the file keeps all eight fields; `level` with no evidence prints facts and proposes nothing; `summary` on an empty repo says so in one line instead of failing.
- CLI e2e: `ask` then `level` shows one open question and one answer after closing it; `summary` lists both features after two `start` runs with different outcomes.

### Process

Deliver in three stages, each with its tests green before the next: (1) domain additions, (2) `ask` + `tasks` helper + wiring, (3) `level` + `summary` + wiring. Report each stage separately so the parent can commit them as three work units. No new dependency; the boundary guard must stay green; do not edit `odd/**`, `README.md`, `Contract I-III` sections, or any existing behaviour.

## Contract V · T7 (`check` and `init`)

### Why the proof carries its tree

A proof is only evidence if it says what it was produced from. Every proof file starts with two header lines, written by the tool and never by hand:

```text
# tree <hash del worktree>
# command <el comando que se corrió>
<salida real, stdout+stderr, verbatim>
```

`check --proof <archivo>` refuses a file without the header: it cannot know which tree produced it, and guessing is exactly the defect the tool exists to prevent.

### `odachi check`

```text
odachi check                             evalúa el estado actual
  --run "<comando>"                      corre el comando, captura la salida y escribe proof/<NNNN>-<slug>.txt
  --proof <archivo>                      usa una prueba ya escrita (agentes y CI)
  --base <ref>                           contra qué comparar el cerco (default HEAD; sin commits: todo lo no commiteado)
  --timeout <segundos>                   default 120
  --quiet                                no imprime nada si está listo (para el hook)
  --json
```

Exit codes: los del gate (`0` listo · `1` el agente puede actuar · `2` solo el humano puede destrabarlo). `--run` devuelve el código de `check`, **nunca el del comando**: si el comando falla, `check` informa `proof_failed` con la última línea de salida, porque lo que el humano necesita saber es "la prueba falló y esto dijo", no un exit code ajeno.

Los cuatro gates, en el orden de `gate.Evaluate` ya existente: brief (existe y está completo), cerco (`fence.Check` con `gitx.ChangedPaths` y `gitx.AddedDeps`), preguntas abiertas (`taskdoc.OpenQuestions` del feature actual), y prueba (header presente, hash igual al del árbol actual, y el comando no falló). Sin task doc, las preguntas abiertas son cero: el proyecto puede verificarse antes de tener features.

`check` no escribe nada salvo el `proof/` que crea `--run`. Nunca toca `brief.md`, `fence.toml` ni `profile.toml`.

### Cambio de dominio necesario

`gate` gana un reason: `ReasonProofFailed = "proof_failed"` (`NeedsWork`), mensaje `la prueba falló: <última línea de salida no vacía>` y `Next` = `odachi check --run "<el mismo comando>"` — correrlo resuelve el bloqueo una vez que el humano o el agente arregla lo que falló. Los siete reasons existentes no cambian y el test que exige `Next` en todo bloqueo sigue valiendo.

### `--run` y el shell

`--run` ejecuta a través de un adaptador nuevo `internal/adapters/shellcmd`, con un archivo por plataforma: `sh -c` en Unix, `cmd /c` en Windows, y una implementación que **rechaza** en plataformas sin shell conocido (regla de nacimiento 11: negarse en vez de degradar). La salida es stdout+stderr combinados. Timeout vencido: `proof_failed` con el mensaje `la prueba no terminó en <n>s`.

### `odachi init`

```text
odachi init [--json]
```

1. Escribe en `AGENTS.md` el bloque de reglas (~40 líneas) entre `<!-- odachi:begin -->` y `<!-- odachi:end -->`, con backup previo y escritura atómica. Idempotente: re-correr reemplaza el bloque y conserva todo lo demás del archivo. Si no existe `AGENTS.md`, lo crea. Contenido del bloque (fijo, pineado por test): las cuatro reglas duras — brief antes de código, nada fuera del cerco (`ask`), no decir "listo" sin `check --run`, pedazos chicos con un commit — y los seis verbos con su uso en una línea.
2. Instala `.git/hooks/pre-commit` con `#!/bin/sh` + `odachi check --quiet || exit 1` y un comentario marcador. Si ya existe un hook distinto, **no lo pisa**: imprime las tres líneas para pegarlo a mano y sale con `1` nombrando el archivo. Re-correr sobre un hook propio es idempotente.

### Tests

- Dominio: el reason nuevo en la tabla y en el test de `Next` obligatorio.
- `proof`: escribir con header, leer rechazando un archivo sin header, hash que coincide y que no coincide.
- App `check`: tabla completa de los ocho resultados (listo, sin brief, brief incompleto, fuera de alcance, pregunta abierta, falta prueba, prueba vieja, prueba fallida) usando dobles de `Files`, `Git` y `Shell`.
- CLI e2e: `--run "echo ok"` → listo y `proof/` escrito con header; `--run "exit 1"` → exit 1 con la última línea; `--proof` sin header → bloqueado nombrando `--run`; dos corridas seguidas con el mismo árbol → mismas bytes de prueba; repo sin commits → no rompe.
- `init`: bloque creado, idempotente, conserva el resto del archivo, backup previo, hook instalado y no pisado cuando ya existe uno ajeno.

### Proceso

Dos etapas, cada una verde antes de la siguiente y reportada aparte: (1) `check` (dominio + adaptador de shell + app + CLI), (2) `init` (AGENTS.md + hook). Sin dependencias nuevas. El guard de fronteras debe seguir verde.

## Contract VI · T9 (distribución)

Objetivo: que alguien sin Go pueda instalarlo y verificar que anda. Nada más.

### Superficie

```text
odachi --version    imprime "odachi <version>" y sale con 0
odachi version      lo mismo (el humano lo va a tipear)
```

`version` es una variable de paquete en `cmd/odachi`, default `"dev"`, inyectada por ldflags. Sin dependencias nuevas.

### `.goreleaser.yaml` (v2)

- `builds`: `./cmd/odachi`, `CGO_ENABLED=0`, `-trimpath`, `ldflags: -s -w -X main.version={{ .Version }}`.
- `goos`: linux, darwin, windows · `goarch`: amd64, arm64.
- `archives`: `tar.gz` con nombre `odachi_{{ .Version }}_{{ .Os }}_{{ .Arch }}`; windows en `zip` vía `format_overrides`. Incluir `README.md` y `LICENSE` **solo si existen** (todavía no hay LICENSE: no inventar una).
- `checksum`: sha256, nombre `checksums.txt`.
- Snapshot con `dev` en el nombre.
- Sin `brews`, sin firma, sin provenance: un comentario al inicio dice que no están y por qué (no hay tap, no hay claves). No fingir que están.

### `.github/workflows/ci.yml`

- `push` y `pull_request`.
- `test`: matriz `ubuntu-latest`, `macos-latest`, `windows-latest` → `go test ./...`; en ubuntu además `go vet ./...` y `gofmt -l .` (falla si imprime algo).
- `race`: ubuntu → `go test -race ./internal/...`.
- `build`: `go build ./...`.
- `release-check`: ubuntu → `goreleaser/goreleaser-action` con `args: check` (valida la config, no publica).
- `permissions: contents: read`. Sin secretos.

### `scripts/install.sh`

`set -euo pipefail`. Requiere `curl` y `tar`; sin ellos, error claro.

1. Detecta OS (`linux`, `darwin`) y arquitectura (`amd64` desde `x86_64`, `arm64` desde `aarch64`/`arm64`). Plataforma no soportada: exit 1 diciéndolo.
2. Versión: `$ODACHI_VERSION`, o la última release vía `api.github.com/repos/biggs-100/odachi/releases/latest`.
3. Base de descarga configurable: `ODACHI_BASE_URL` (default la URL de releases de GitHub para esa versión). Es lo que permite testearlo sin red.
4. Descarga el archivo y `checksums.txt`, verifica sha256 (sin `sha256sum` usar `shasum -a 256`; si no hay ninguno de los dos, error claro en vez de saltear la verificación).
5. Instala en `${ODACHI_BIN_DIR:-$HOME/.local/bin}`, `chmod +x`. Si no puede escribir ahí, instala en `./odachi` y lo dice.
6. Imprime `odachi <versión> instalado en <ruta>` y, si la ruta no está en `PATH`, el hint exacto a pegar.

### `scripts/install.ps1`

Equivalente para Windows: detecta `AMD64`/`ARM64`, descarga el `zip`, verifica sha256 con `Get-FileHash`, instala en `$env:LOCALAPPDATA\odachi\bin`, agrega esa carpeta al `PATH` del usuario si falta, imprime la versión. Mismos errores claros.

### Tests (sin red, obligatorios)

- `cmd/odachi`: `--version` y `version` imprimen `odachi dev` en un build sin ldflags y salen con 0; un comando desconocido sigue saliendo con 1.
- `cmd/odachi/install_test.go` (package main, usa `bash`): construye un binario real con `-ldflags -X main.version=1.2.3` en `t.TempDir()`, arma `odachi_1.2.3_<os>_amd64.tar.gz` + `checksums.txt` en un directorio de release, y corre `bash scripts/install.sh` con `ODACHI_BASE_URL=file://<release> ODACHI_VERSION=1.2.3 ODACHI_BIN_DIR=<bin>`. Asserta: exit 0, el binario existe, y `<bin>/odachi --version` dice `odachi 1.2.3`. Segundo caso: checksums.txt alterado → exit != 0 y **el binario no queda instalado**. Si `bash` no está disponible, `t.Skip` con la razón.
- Nada de tests que dependan de la red ni de una release real.

### Proceso

Unidad única. No tocar el comportamiento de los seis verbos. No agregar dependencias Go. Si `LICENSE` no existe, los archivos de release no lo mencionan.

## Contract VII · T12 (skills: registro y creador)

### Surface

```text
odachi skills                lista las skills del proyecto con su trigger (humano)
odachi skills --json
odachi skills --check        valida; exit 1 si algo no cumple (lo corre CI)
odachi skills --write        regenera SKILLS.md desde los frontmatter
odachi skill new <slug>      crea skills/<slug>/SKILL.md desde la plantilla
```

### Formato (única fuente de verdad)

```markdown
---
name: acceptance
description: Trigger: cuando hay que decidir si algo está terminado. Convierte un pedido en criterios observables.
---

## Cuándo
...

## Reglas
...

## Nunca
...
```

### Validaciones de `--check` (una razón por violación, en español)

- `name` == nombre del directorio, `^[a-z][a-z0-9-]{2,40}$`.
- `description` empieza con `Trigger:` y tiene ≤ 200 caracteres.
- Cuerpo ≤ 60 líneas; sin secciones vacías; sin `TODO`.
- Nombres únicos.
- Toda skill del disco aparece en `SKILLS.md`, y `SKILLS.md` tiene que ser **exactamente** lo que `--write` generaría: si alguien lo editó a mano, `--check` falla y dice que corra `--write`.

### SKILLS.md (generado)

Primera línea: `<!-- generado por odachi skills --write; no lo edites a mano -->`. Después una tabla `Skill | Trigger | Path`, ordenada por nombre. Nada más.

### `skill new`

Crea el directorio y el `SKILL.md` con la plantilla (frontmatter con `name` y una `description` de `Trigger:` vacía para completar, y las tres secciones). Si ya existe: exit 1, sin tocar. Al terminar corre el equivalente de `--write` para que el registro quede al día, e imprime el path y qué sigue.

### Integración

- `odachi init` **no** escribe skills: no las posee. Su bloque gana **una** línea: `Skills: ver SKILLS.md; cargá la que matchee el trigger antes de escribir código.`
- `pack` (unidad siguiente) es quien copia las skills incluidas al proyecto; esta unidad no lo implementa.
- Las 5 skills iniciales se entregan dentro de este repo (`skills/`), dogfooding: `acceptance`, `slices`, `tests`, `data`, `ship`, cada una ≤ 60 líneas con `## Cuándo / ## Reglas / ## Nunca`.

### Capas

- Dominio puro: parseo de frontmatter, validación, y generación de la tabla. Sin `os`.
- La app lee y escribe con el puerto `Files` que ya existe; el CLI cablea.
- Sin dependencias nuevas.

### Tests

- Dominio: frontmatter válido, sin frontmatter, CRLF, `description` sin `Trigger:`, nombre que no coincide con el directorio, cuerpo de 61 líneas, `TODO`, tabla generada y ordenada.
- App/CLI: `--check` verde con las 5 skills; una violación por test (cada una con su mensaje); `--write` idempotente (dos veces, mismos bytes); `--check` falla si `SKILLS.md` se editó a mano; `skill new` crea y no pisa; repo sin `skills/` sale 0 con una línea que dice cómo empezar.
- CI: `go test ./...` ya lo cubre; no hace falta un job nuevo.

## Contract VIII · T13 (`odachi pack`)

Objetivo: que las skills lleguen al agente. Hoy viven en el repo; para que el agente las cargue tienen que estar donde él las busca.

### Paths verificados (no supuestos)

Sacados de los adaptadores de gentle-ai (`SkillsDir(homeDir)`), todos home-relativos:

| Agente | Destino |
|---|---|
| `claude` | `~/.claude/skills` |
| `codex` | `~/.codex/skills` |
| `cursor` | `~/.cursor/skills` |
| `gemini` | `~/.gemini/skills` |
| `copilot` | `~/.copilot/skills` |
| `qwen` | `~/.qwen/skills` |
| `kiro` | `~/.kiro/skills` |

**Pi no entra**: su adaptador devuelve `SkillsDir == ""` porque no usa un directorio de skills. Se rechaza con esa razón, no se salta en silencio. Opencode queda afuera de esta unidad: su path depende de `ConfigPath` (varía por plataforma) y no voy a escribir en una ruta que no verifiqué.

### Surface

```text
odachi pack                         instala las skills incluidas en cada agente detectado
odachi pack --agents claude,codex    solo esos
odachi pack --dry-run               qué haría, sin escribir
odachi pack --json
```

### Comportamiento

- Fuente: las skills incluidas en el binario (`skills/*/SKILL.md` embebidas con `//go:embed`). Nada de red.
- Escribe `<home>/<destino>/<name>/SKILL.md` con `fsx.WriteAtomic`; si el archivo existe y difiere, backup antes. Si es idéntico: `sin cambios`, sin backup.
- **Nunca borra ni mueve nada.** Un archivo que sobra se reporta, no se toca.
- Agente desconocido: exit 1 nombrando los soportados. `pi`: exit 1 con "Pi no usa un directorio de skills".
- `--dry-run` no escribe nada y lista las acciones igual.
- `pack` **no** toca `skills/` del proyecto ni `SKILLS.md`: eso es de `odachi skills --write`.
- Resumen en español, una línea por agente: `claude · 5 instaladas · ~/.claude/skills`.

### Capas

- `internal/assets/skills` (o equivalente) embebe las cinco con `//go:embed`; el paquete del repo raíz `skills/` es la fuente y no se duplica.
- `internal/app/pack` sobre los puertos `Files` (ya existe) y uno nuevo `Home` para resolver el home del usuario. Sin `os` en la app.
- Tabla de agentes como datos en el dominio, no como ramas de código.

### Tests

- Tabla: cada agente soportado resuelve su path; `pi` y un nombre inventado se rechazan con su mensaje.
- Escribe los cinco archivos; segunda corrida idéntica no escribe ni hace backup; un archivo local distinto se respalda antes de pisarse.
- `--dry-run` no toca el filesystem (se assertea que el doble no recibió escrituras).
- Un fallo en un agente no aborta los demás y el mensaje nombra agente y path.
- Nada del proyecto (`skills/`, `SKILLS.md`) cambia.

## Contract IX · T15 (`plan`)

Hoy el agente agrega pasos editando Markdown a mano. Si escribe una sección que el render canónico no conoce, el guard de `ask` le bloquea toda escritura posterior y el proyecto queda trabado. `plan` cierra ese hueco.

### Surface

```text
odachi plan "<texto>" [--req R1]           agrega el próximo paso al task doc actual
odachi plan --done T2 [--commit abc1234]   marca un paso hecho y registra el commit
odachi plan --json
```

### Reglas

- Task doc: el de mayor número, la misma resolución que usa `ask`. Sin task doc: exit 1 nombrando `odachi start`.
- `<texto>` obligatorio, una linea, ≤ 120 caracteres, sin guiones iniciales ni saltos de linea.
- `--req R<n>` tiene que existir en el documento; si no, error que lista los requisitos. Sin `--req` el paso va sin vínculo, y el render lo refleja.
- El número `T` es `max+1`: monótono, nunca se reutiliza.
- `--done` con id desconocido: error que lista los pasos abiertos. Un paso ya hecho: no-op, exit 0, `ya estaba hecho`.
- `--commit` sin `--done`: error. `--commit` con un paso que no se está marcando: error.
- La escritura pasa por **el mismo guard que `ask`**: si el render canónico perdería una linea que el humano escribió, se bloquea y el mensaje dice dónde moverla. Ese guard se extrae a `internal/app/tasks` y lo usan los dos, para que haya un solo lugar que decide si se puede escribir.
- Backup antes de escribir, escritura atómica, y `plan` no toca `brief.md`, `profile.toml` ni `fence.toml`.
- Mensaje: `paso T3 agregado a odd/tasks/0001-x.md` / `paso T3 hecho · abc1234`.

### Tests

- Agrega con y sin `--req`; `--req` inexistente lista los requisitos.
- Texto vacío, multilínea, con guión inicial y de 121 caracteres: rechazados antes de escribir.
- Dos `plan` seguidos: `T1` y `T2`, y el documento queda canónico (round-trip).
- `--done` marca y registra el commit; desconocido lista los abiertos; ya hecho es no-op.
- `--commit` sin `--done` rechazado.
- Documento con una sección extra: `plan` se bloquea con el mensaje que nombra `## Historial`, **y no escribe nada**.
- Sin task doc: exit 1 nombrando `odachi start`.
- `--json` con el mismo contenido que el mensaje humano.

## Contract X · T16 (`undo`)

`undo` deshace **lo que odachi escribió**, no tu código. El código es tuyo y de git; si la herramienta empieza a elegir commits para deshacer, se convierte en lo que critica.

### Surface

```text
odachi undo                  restaura el artefacto que odachi escribió último
odachi undo --list           qué se puede deshacer, con fecha
odachi undo --file brief      solo ese: brief | profile | fence | task
odachi undo --json
```

### Artefactos y backups

Se consideran solo los archivos que odachi escribe: `brief.md`, `.odachi/profile.toml`, `.odachi/fence.toml` y `odd/tasks/*.md`. Un backup es `<archivo>.bak-<8 hex>` al lado del archivo (el formato que ya produce `fsx.Backup`). **Nada fuera de esa lista se lee ni se toca.**

### La regla que hace `undo` seguro

Antes de restaurar, **respalda el estado actual** con el mismo mecanismo. Después copia el backup elegido sobre el archivo. El resultado se imprime con las dos rutas:

```text
restauré brief.md desde brief.md.bak-466dffe6; lo que había quedado en brief.md.bak-a317e32c
```

Así `undo` es a su vez deshacible y una edición humana posterior a la última escritura del agente no se pierde.

### Selección

- El backup **más reciente por `mtime`** entre los artefactos permitidos. `--file` filtra a uno. **Un solo candidato restaura; varios se rechazan** listándolos y pidiendo `--file`: "el más reciente" entre artefactos distintos lo decide el orden interno de escritura, no lo que el humano quiso. Si los bytes actuales ya son los del backup, dice `no hubo cambios` y no escribe.
- `--list`: una línea por artefacto con backup, de más nuevo a más viejo: `brief.md · 2 backups · último 2026-10-21 20:31`.
- Sin backups: exit 1, una línea que dice qué escribe odachi y sugiere `odachi start`.
- `--file` con un nombre fuera de la lista: exit 1 con los cuatro permitidos.
- El puerto `Files` gana `ModTime(path) (time.Time, error)`; el dominio elige, la app escribe.

### Lo que `undo` NO hace

No toca `git`, no revierte commits, no borra backups. Si el humano quiere deshacer código, el mensaje le dice que eso es `git revert <sha>` — nombrado, no ejecutado.

### Tests

- Restaura el backup más reciente y deja el estado previo como backup nuevo (se assertea el orden: backup → copia).
- `--list` ordena por fecha y muestra el conteo; sin backups, exit 1 con la línea que nombra `odachi start`.
- `--file brief` ignora backups más nuevos de otros artefactos; `--file inventado` lista los cuatro.
- Un `.bak-*` de un archivo ajeno al proyecto no se considera.
- `--json` con el mismo contenido que el mensaje humano.
- Ningún test ni ejecución toca archivos fuera de los cuatro artefactos (se assertea sobre el doble).

## Contract XI · T17 (lock de escritura)

Es el último riesgo real del diseño. Los archivos se escriben de forma atómica uno por uno, así que no hay corrupción: lo que puede pasar es que dos corridas se intercalen y dejen un conjunto **coherente por archivo e incoherente entre sí** (el `brief.md` de una corrida con el `fence.toml` de la otra).

### Alcance

Solo los verbos que escriben **más de un archivo**: `start` y `undo`. `plan`, `ask` y `level --to` escriben un solo archivo de forma atómica; no necesitan lock y no lo toman (un lock que no protege nada es fricción).

### Mecánica

- Archivo `.odachi/lock`, creado en modo exclusivo (`O_CREATE|O_EXCL`). Contenido: una línea con el verbo, el pid y el instante. Se libera al terminar, incluso si el verbo falla.
- Si ya existe: se lee.
  - Más viejo que **10 minutos** → se considera abandonado: se toma, se avisa en una línea (`había un lock viejo de <verbo>, hace <X>; lo tomé`) y sigue.
  - Más nuevo → exit 1: `otra corrida de odachi está escribiendo en este proyecto (<verbo>, pid <n>, hace <X>s); esperá y volvé a intentar`. Nombra el archivo y el comando que lo libera a mano (borrarlo), nunca lo borra.
- Un lock ilegible o con contenido inesperado **no bloquea para siempre**: se trata como viejo si su mtime lo es, y si no, se bloquea con el mensaje que nombra el archivo.
- Nunca se borra un lock ajeno vivo. Nunca se ignora un lock presente sin decirlo.

### Capas

- Puerto `Lock` nuevo en `internal/app/lock`: `Acquire(path, owner string) (held bool, stale string, err error)` y `Release()`. Sin `os` en la app; el adaptador lo implementa con `O_EXCL`, `ModTime` y `Remove`.
- El dominio no participa: no hay regla de negocio, solo exclusión.
- `start` y `undo` toman el lock **antes de la primera lectura** que decide algo y lo liberan con `defer`.

### Tests

- Segundo `start` mientras el primero tiene el lock: exit 1, mensaje con verbo/pid/antigüedad, y **ningún archivo tocado** (se assertea sobre el doble).
- Lock viejo (mtime > 10 min): se toma, se avisa, la corrida completa y el lock queda liberado.
- Lock liberado tras un fallo del verbo (se assertea que el release se llamó).
- `plan`/`ask`/`level` no crean `.odachi/lock`.
- Un lock con contenido basura y mtime reciente: bloquea con el mensaje que nombra el archivo.
- `--json` reporta el mismo mensaje.

## Contract XII · T18 (OpenCode, y destinos por alcance)

OpenCode se rechazaba porque su path "dependía de la plataforma". Es falso para el path que importa: **`.opencode/skills/<name>/SKILL.md` es relativo al proyecto**. Verificado en la documentación y en el código de OpenCode (`OPENCODE_SKILL_PATTERN = "{skill,skills}/**/SKILL.md"`; además lee `.claude/skills/` y `.agents/skills/` como externos compatibles). El global, `~/.config/opencode/skills`, existe pero no es el que queremos.

### El cambio de modelo

La tabla de agentes gana **alcance**. Hasta ahora todo era `home`; ahora hay dos:

| Alcance | Dónde escribe | Quién |
|---|---|---|
| `home` | `<home>/<Dest>` | claude, codex, cursor, gemini, copilot, qwen, kiro |
| `project` | `<project>/<Dest>` | **opencode** → `.opencode/skills` |

`opencode` sale de la lista de rechazos. `pi` se queda, con su razón verificada.

### Reglas

- `pack` resuelve `project` contra el directorio del proyecto (el mismo `--dir` de los demás verbos, default `.`), y `home` contra el home del usuario. Nada más cambia: escritura atómica, backup si difiere, `sin cambios` si es igual, nunca borra.
- En los mensajes, `home` se muestra como `~/...` y `project` como `./.opencode/skills`.
- **Detección**: un agente `project` se detecta si existe su directorio raíz en el proyecto (`.opencode/`). Un agente `home` como antes (su directorio de config en el home). `pack` sin `--agents` en un proyecto sin nada detectado sigue saliendo 0 y nombrando `--agents`.
- `pack --agents opencode` funciona en cualquier proyecto, exista o no `.opencode/`.

### Tests

- `Resolve("opencode")` devuelve alcance `project` y destino `.opencode/skills`; los siete anteriores siguen en `home` con sus paths.
- `pack --agents opencode` escribe los cinco `SKILL.md` bajo `<project>/.opencode/skills/`, y **no toca el home** (se assertea sobre el doble: cero escrituras fuera del proyecto).
- `pack` en un proyecto con `.opencode/` detecta opencode sin `--agents`.
- `pack --agents opencode` en un proyecto sin `.opencode/` funciona igual.
- Los mensajes muestran `./.opencode/skills` y no una ruta absoluta.
- `pi` sigue rechazado con su razón; un nombre inventado sigue listando los soportados, ahora con opencode incluido.

### Documentación

README: la lista de agentes soportados incluye opencode con `--agents opencode`, y el camino de prueba con OpenCode (init + pack + start + check).

## Contract XIII · T20 (`review`)

Lo que odachi todavía no contesta: *¿está bien diseñado?* El gate dice si está probado. Esto agrega un veredicto sobre el diseño, sin heredar la máquina de estados de gentle-ai: **un archivo por árbol, content-addressed, inmutable.**

### De gentle-ai se toma esto (y solo esto)

1. **Compuerta de precisión**: reportar solo defectos reales que afecten al usuario y que se puedan defender con evidencia concreta; ante la duda, callarse. Estilo y preferencia prohibidos salvo que oculten un defecto.
2. **Vocabulario de severidad**: `BLOCKER | CRITICAL | WARNING | SUGGESTION`.
3. **Resultado vacío canónico**: si está limpio, exactamente `No findings.`
4. **Causalidad**: un finding cuenta solo si el candidato lo **introdujo, activó o empeoró**. Lo que ya estaba es follow-up, no bloqueo.
5. **Un barrido por lente, sin loop hasta que se seque.**

No se toma: ledger en Engram, refuters con voto 2-de-3, envelopes de consentimiento, lineages.

### Surface

```text
odachi review                     imprime el prompt de review del candidato actual
odachi review --exec "<comando>"   lo corre, lee el veredicto JSON y lo guarda
odachi review --attach <archivo>   guarda un veredicto ya escrito (agente o humano)
odachi review --show              qué veredicto hay para este árbol, o que no hay
odachi review --json
```

### El candidato y el almacenamiento

- Candidato = `TreeHash` actual + `ChangedPaths` contra `--base` (default `HEAD`) + el `brief.md`. Sin lineage, sin revisiones, sin autoridad.
- Almacenamiento: `.odachi/reviews/<treehash>.json`, un archivo por árbol. Inmutable: si ya existe, `--exec`/`--attach` lo respetan y avisan (`ya hay un veredicto para este árbol`); `--show` lo lee.
- Un árbol distinto es un veredicto distinto. La aprobación no puede flotar: es literalmente el nombre del archivo.
- El veredicto guardado lleva: `tree`, `base`, `reviewer` (texto libre: el comando o el humano), `at` (RFC3339), `findings[]` con `{path, line, severity, causal, why}` y `verdict`: `approved | changes_requested`.
- **Regla derivada, no administrada**: si algún finding tiene `causal: true` y severidad `BLOCKER` o `CRITICAL`, el veredicto es `changes_requested`. El resto es `approved`. No hay presupuesto de corrección ni estados.

### El prompt que imprime

Contenido fijo (contenido, no código): el brief, las rutas cambiadas, el diff, las cinco reglas tomadas de gentle-ai, el schema JSON esperado y la instrucción de contestar solo con el JSON. Sin el diff no hay review: si no cambió nada, exit 1 (`no hay nada que revisar en este árbol`).

### El gate

`check` gana un flag opcional `--require-review`: apagado por defecto (un novato no debe quedar bloqueado por no saber qué es un review). Cuando está encendido y no hay veredicto para el árbol actual: exit 2, mensaje `no hay review para este árbol` y `Next` = `odachi review`. Si hay veredicto `changes_requested`: exit 2 con la primera finding causal. Si `approved`: sigue.

### Tests

- Veredicto: escritura del archivo con el nombre del árbol; segundo intento avisa y no pisa; `--show` lee; sin veredicto dice que no hay.
- Derivación: BLOCKER causal → `changes_requested`; CRITICAL no causal → `approved`; solo WARNING → `approved`.
- Un árbol distinto no ve el veredicto del anterior (se assertea el nombre del archivo).
- `--attach` rechaza un JSON inválido con un mensaje en español y no escribe nada.
- `--exec` con un comando que falla: exit 1, no escribe, y muestra la salida del comando.
- El prompt contiene el brief, las reglas y el schema; y sin diff no hay prompt (exit 1).
- `check --require-review` en las tres situaciones; sin el flag, el review no influye.
- Nada fuera de `.odachi/reviews/` se escribe.

## Contract XIV · T21 (lentes y refutador)

Objetivo: paridad de potencia con el 4R de gentle-ai sin heredar su máquina de estados. Potencia = cuatro prompts especializados + selección por evidencia + refutación de blockers. Lo que no se copia no aporta potencia: ledger persistido, consentimiento, voto 2-de-3.

### Surface

```text
odachi review                       usa las lentes que pide la evidencia del diff
odachi review --lens risk            una lente explícita (repetible o separada por comas)
odachi review --lens all             las cuatro
odachi review --refute <verdict>     imprime el prompt del refutador para ese veredicto
odachi review --lens --show          qué lentes hay para este árbol y cuáles faltan
```

### Las cuatro lentes (contenido embebido, `internal/assets/review/<lente>.md`)

Cada archivo es el prompt de un especialista. Frontmatter `name` y `description: Trigger: ...`; cuerpo con `## Qué mirar` (entre 12 y 20 reglas **concretas y accionables**, del estilo "una consulta armada concatenando entrada del request" o "un retry sin backoff ni techo") y `## Qué NO reportar`. Nada de adjetivos sueltos.

| Lente | Territorio |
|---|---|
| `risk` | seguridad, límites de privilegio, exposición de datos, dependencias, sinks peligrosos |
| `readability` | nombres, complejidad, duplicación, código muerto, tamaño de revisión, intención |
| `reliability` | tests por comportamiento, casos borde, determinismo, contratos, regresiones |
| `resilience` | fallbacks, retry/backoff, degradación, observabilidad, carga, rollback |

Todas heredan las cinco reglas de Contract XIII y agregan el piso de severidad: `WARNING` y `SUGGESTION` nunca bloquean.

### Selección por evidencia (determinista, pura)

`internal/domain/review/tier.go`:

```go
type Tier string // "passive" | "focused" | "full"
func Classify(changedPaths []string) Tier
func Lenses(t Tier) []string // passive: ninguna · focused: [reliability] · full: las cuatro
```

- `passive`: **solo** documentación (`.md`, `.txt`, `docs/**`, `README*`) y ningún archivo de código ni configuración de ejecución.
- `full`: cualquier ruta o diff que toque señales sensibles — `auth`, `login`, `session`, `token`, `password`, `secret`, `permission`, `role`, `payment`, `billing`, `migrat`, `schema`, `.env`, `docker`, `ci`, `deploy`, `Makefile`, `go.mod`. Es la misma lista de señales que ya usa el cerco, no una segunda taxonomía inventada.
- `focused`: todo lo demás → una sola lente, `reliability`.
- **El volumen no decide**: mil líneas de documentación siguen siendo `passive`, y cinco líneas de `auth` son `full`. Un test por cada regla.
- El agente o el humano pueden **agregar** lentes con `--lens`; no pueden quitar las que la evidencia pide: `check --require-review` verifica que exista veredicto para **cada lente requerida**.

### Refutador (`--refute <verdict>`)

Imprime un prompt que recibe el veredicto y le pide refutar **solo los BLOCKER y CRITICAL**, con tres preguntas por finding:

1. ¿Es causal del diff, o ya estaba? (cita la línea del diff)
2. ¿Es un defecto real que afecta al usuario, o una preferencia?
3. ¿La evidencia alcanza para defenderlo ante el autor?

Devuelve un JSON `{"refutations":[{"id":"...","verdict":"stands|refuted","why":"..."}]}`. Regla: **si el refutador no puede citar la evidencia, el finding no se refuta** (default `stands`, nunca al revés). Se guarda en `.odachi/reviews/<treehash>.refuted.json` y el veredicto derivado lo tiene en cuenta: un blocker refutado deja de bloquear, y el archivo de refutación se lista en `--show`.

### El gate

`check --require-review` ahora exige: veredicto para el árbol **y para cada lente requerida**; bloquea si cualquier finding causal BLOCKER/CRITICAL **no refutado** existe en cualquier lente. Un blocker refutado no bloquea. Sigue siendo opt-in.

### Tests

- `Classify`: docs-only → passive; `auth/session.go` → full; `internal/gastos/x.go` → focused; docs grandes → passive (volumen no decide); cada señal de la lista, una por test.
- `Lenses(t)`: passive vacío, focused `[reliability]`, full las cuatro en orden fijo.
- Lentes: cada archivo embebido existe, tiene frontmatter válido, entre 12 y 20 reglas, y `## Qué NO reportar`. Un test que cuenta las reglas.
- `--lens` explícito genera el prompt de esa lente; `--lens all` las cuatro; `--show` dice cuáles faltan.
- Refutación: refutado deja de bloquear; `stands` sigue bloqueando; refutador sin evidencia → `stands`; el archivo `.refuted.json` se escribe una vez y no se pisa.
- Gate: falta una lente requerida → exit 2 nombrando cuál; blocker refutado → listo.

## Contract XV · T22 (`doctor`) y T23 (`bench`)

Dos instrumentos. Ninguno escribe: `doctor` lee el proyecto, `bench` corre el binario en directorios descartables.

### `odachi doctor [--json]`

Diez chequeos, **de solo lectura**, cada uno con una línea en español y, si algo está mal, el comando que lo arregla:

| # | Chequeo | Problema |
|---|---|---|
| 1 | la versión del binario | — |
| 2 | `odachi` resuelve desde un shell | el agente no puede ejecutar las reglas |
| 3 | hay `.git` | no hay hook posible |
| 4 | el bloque de odachi está en `AGENTS.md` (marcadores) | el agente no lee las reglas |
| 5 | el hook de pre-commit existe y es nuestro | no hay gate local |
| 6 | el binario que nombra el hook existe | el hook falla en cada commit |
| 7 | `.odachi/profile.toml` resuelve a un perfil válido | ceremonia rota |
| 8 | `brief.md` existe y no le falta nada | sin visión no hay nada que construir |
| 9 | skills: cuántas y si `SKILLS.md` está al día | el catálogo se desvió |
| 10 | el store de review y `proof/` están ignorados | cada review invalidaría la prueba |

Exit 0 si no hay problemas (los avisos no fallan), 1 si hay al menos un problema. `--json` con el mismo contenido.

### `odachi bench [--json] [--journeys <nombres>] [--binary <ruta>]`

El instrumento que mide cómo se traba una persona. Cada recorrido corre en un directorio temporal con `HOME` controlado.

- Un recorrido es una secuencia de invocaciones con su exit esperado. Una salida ≠ 0 es un **bloqueo**.
- Cada bloqueo se clasifica por su mensaje de salida:
  - **`in_band`**: nombra un comando concreto y, corrido en el mismo sandbox, sale 0.
  - **`out_of_band`**: no nombra ningún comando.
  - **`dead_end`**: nombra un comando que, corrido, **no** resuelve el bloqueo.
  - **`templated`**: el comando trae un `<placeholder>`; no se puede correr solo. Se cuenta aparte y no falla el bench.
- **Exit 1 si hay al menos un `dead_end`.** Es el ratchet: el bench falla el build cuando un mensaje manda a un lugar sin salida.
- Recorridos mínimos: repo nuevo con `start`; `start` dos veces; `check` sin prueba; `check --run` verde; archivo prohibido por el cerco; `plan` + `undo`; `review` de un cambio solo de docs; repo sin commits.
- Reporte: total de recorridos, total de bloqueos, los tres conteos, y la lista de dead ends con el comando que no resolvió y el mensaje que lo nombró. `--json` con lo mismo.

### Tests

- `doctor`: en un proyecto recién configurado da todo ok y exit 0; sin `AGENTS.md`, sin hook, sin `brief.md`, con perfil inválido y con `odachi` fuera del PATH (puerto de PATH falso) produce el problema correcto con el comando correcto; nunca escribe nada (se assertea sobre el doble).
- `bench`: los tres conteos sobre un corpus inyectado (uno que resuelve, uno sin comando, uno que no resuelve, uno templated); un dead end hace fallar el bench; `--journeys` filtra; los directorios temporales se limpian.

## Tasks

- T1 · repo, go.mod, boundary guard test, ADR 0001, birth decisions, change checklist, impact map, artifact examples · S1-S8 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T2 · pure domain: fence (globs), gate (exit codes), profile (schema + reflection guard) · S2,S3,S5 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T3 · pure domain: brief and task document parsing + progress · S1,S6 · delegated (gentle-ai-worker) · [x] commit: c77ca91
- T4 · adapters: strict TOML profile decoder, filesystem, git · S5 · delegated (gentle-ai-worker) · [x] commit: f53fa1a
- T5 · CLI `start`: interview -> brief + fence + profile · S1,S2 · delegated (gentle-ai-worker) · [x] commit: 96ccaec
- T6 · CLI `ask` / `level` / `summary` · S3,S7 · delegated (gentle-ai-worker), three stages · [x] commits: bdb5a6b (domain), 1669768 (app), e113f94 (CLI)
- T7 · CLI `check` + pre-commit hook + AGENTS.md block · S2 · delegated (gentle-ai-worker), two stages · [x] commits: 0520b83 (domain), 64d51e6 (check), 56ebf87 (init + wiring)
- T8 · tutor content: why each question exists, explanation rules · S7 · inline · [x] commits: 98290c3 (brief), 829b633 (tutor), a9b9057 (hook PATH)
- T9 · distribución: `--version`, goreleaser, CI, install.sh y install.ps1 · S1,S8 · delegated (gentle-ai-worker) + verificación independiente · [x] commit: 0e459f0
- T10 · CI verde en Windows: el test del instalador dependía de un shim de uname · S1 · inline · [x] commit: 85694e2
- T11 · publicación: licencia MIT, tag v0.1.0 y release verificada de punta a punta · S1,S8 · inline · [x] commit: 4d92358 · release run: 39657703108
- T12 · skills: registro generado, validador, creador y las 5 skills iniciales · S1,S7 · delegated (gentle-ai-worker) · [x] commits: 58282d4 (domain), f409072 (cli), 208e606 (content)
- T13 · `pack`: llevar las skills al directorio nativo de cada agente · S1,S3 · delegated (gentle-ai-worker) · [x] commits: 511cc7c (domain), c0516e6 (embed), e036438 (app + cli)
- T14 · `IsDir` en el escáner de skills: la clase de error de `stat` no decide qué es una entrada · S2,S4 · inline · [x] commit: 3c9f2ae
- T15 · `plan`: agregar y cerrar pasos sin editar Markdown a mano · S3,S6 · delegated (gentle-ai-worker) · [x] commits: 6dbecfc (domain), 1130ee9 (guard), 658e900 (plan), 3597bed (cli + readme)
- T16 · `undo`: restaurar el último artefacto que odachi escribió, respaldando lo actual · S2 · delegated (gentle-ai-worker) + corrección del padre · [x] commits: 58fe726 (domain), d41b3c7 (app), 28a2791 (cli)
- T17 · lock de escritura para `start` y `undo`: exclusión entre corridas concurrentes · S2,S4 · delegated (gentle-ai-worker) · [x] commits: a9559ac (lock), da8392c (start), 00bc89a (undo), d2889bb (cli + adr)
- T18 · OpenCode: destinos por alcance (proyecto vs home) y `opencode` soportado · S1,S4 · delegated (gentle-ai-worker) · [x] commits: ed56b2f (domain), 911ec0d (app), 57e3669 (cli + readme)
- T19 · `init` avisa cuando `odachi` no resuelve desde un shell: una regla que el agente no puede ejecutar es peor que ninguna · S1,S2 · inline · [x] commit: 400eee8
- T20 · `review`: veredicto de diseño por árbol, content-addressed, y `check --require-review` opcional · S3,S4 · delegated (gentle-ai-worker) · [x] commits: 9e5c3b8 (domain), cf1bcde (app + check), ed6bdc9 (cli + adr)
- T21 · lentes 4R con selección por evidencia + refutador de blockers · S3,S4 · delegated (gentle-ai-worker) · [x] commits: e254f19 (assets + tier), 4f5f36b (domain), 466e940 (app + check), 74cf7c7 (cli + adr)
- T22 · `doctor`: diez chequeos de salud, de solo lectura, cada uno con su arreglo · S1,S2 · delegated (gentle-ai-worker) · [x] commits: b0e99be, c2d9bc9
- T23 · `bench`: medir fricción (in-band / out-of-band / dead-end) con ratchet sobre los dead ends · S2 · delegated (gentle-ai-worker) · [x] commits: 7d09160, c2d9bc9 · T24: correcciones del instrumento y del mensaje del cerco
- T25 · `review --exec`: el revisor recibe el prompt (stdin desde un archivo temporal y `{prompt}` sustituido por su ruta) · S3,S4 · inline · [x] commits: f51913e (shellcmd), 73d684d (review + docs)
- T26 · README al día: ocho skills, quince superficies, la release real y la sección de review que faltaba · S1,S2 · inline · [x] commit: 38e66a5
- T27 · `gitx.Diff`: el adaptador de diff sale de `cmd/` y el binario queda con un solo adaptador de git · S4 · delegado (gentle-ai-worker) + verificación del padre · [x] commit: 9db2c4b
- T28 · la evidencia deja de mover el árbol: `init` mantiene un bloque marcado en el `.gitignore` del proyecto y `check`/`review` preguntan antes de escribir su archivo interno · S2,S6 · delegado (gentle-ai-worker) + verificación del padre · [x] commits: 690d9bd (gitx), ab52fce (init), bbe7505 (check+review), c4e5ab5 (doctor), 7c619bc (docs)
- T29 · identidad del producto y del rol en el bloque de `AGENTS.md`, y `doctor` que reconoce una carpeta vacía con un solo paso · S2,S7,S10 · delegado (gentle-ai-worker) + verificación del padre · [x] commits: 091a753 (bloque), 8a30cce (doctor)
- T30 · `pack` portable: `.agents/skills` como destino estándar, tabla nativa + estándar por cliente, `--scope` y `--dest` para cualquier agente futuro · S1,S9 · delegado (gentle-ai-worker) + verificación del padre · [x] commits: fa61787 (tabla), ffcc676 (pack + cli)
- T31 · skill de arranque en `~/.agents/skills` e `init` que deja el proyecto listo de una sola vez · S1,S9 · delegado (gentle-ai-worker) + verificación del padre · [x] commits: 9c86e26 (skill), 4a79df2 (pack), b506d43 (init + cli)
- T32 · personas generadas donde el cliente las tiene (agentes de OpenCode, output styles de Claude Code), con la línea de no-divergencia · S4,S7,S9 · delegado (gentle-ai-worker) + verificación del padre · [x] commits: 4237a91 (dominio), d37347c (init), 3160784 (cli)
- T33 · la skill que enseña a manejar odachi desde adentro (`drive`), y la línea del bloque que la manda a cargar · S3,S7,S10 · inline · [x] commit: c42aaf1
- T34 · el bloque nombraba `SKILLS.md`, que no existe en un proyecto de usuario; encontrado por un agente real · S2,S3 · inline · [x] commit: 185e8c0

## Log

L1 · "analiza gentle-ai y dime como crearias algo mucho mejor y que no sea complicado ni demoroso de hacer" — path C:\Users\USER\Desktop\herramientas\gentle-ai
L2 · nombre elegido: odachi. Descartados: gentle-*, oficio, wright, koder-ai, dev-ai
L3 · se descarta `decisions.md` como archivo aparte: lo durable va a `brief.md`, lo cronológico al `Historial` del task doc. Una sola fuente de verdad.
L4 · el humano autoriza T1 y T2: "ok, comencemos" (path local del repo omitido a propósito en un archivo que se publica).
L5 · T1+T2 verificados por el padre, no por el reporte del worker: `go test ./...` (archtest + fence + gate + profile ok), `go build ./...`, `go vet ./...`, `gofmt -l` vacío; 1.189 líneas de Go; el guard confirma que el dominio solo importa stdlib.
L6 · Riesgos que el worker marcó, resueltos como decisiones: (a) out_of_scope = NeedsDecision(2), porque la pregunta real es "¿esto entra al proyecto?" y esa es del humano; (b) prueba faltante = ProofTreeHash vacío o ProofOutput vacío, prueba vieja = hashes distintos; (c) no_brief y brief_incomplete también son NeedsDecision(2): el brief lo define el humano.
L7 · Corrección del padre sobre la entrega: `senior.Proof` pasa de heredar "command" a "suite-or-ci", que es lo acordado en la tabla de perfiles.
L8 · T3+T4 verificados por el padre, no por el reporte: 9 paquetes ok, `go build`, `go vet`, `gofmt -l` vacío. El fixture real del task doc tiene 3 requisitos y 3 pasos; mi brief de delegación decía 2/2 por error mío: se corrige la expectativa, no el artefacto.
L9 · Defecto real encontrado en la entrega, con causa en mi especificación: `TreeHash` vía `git stash create` ignoraba los archivos sin trackear (una prueba vieja seguía pareciendo válida) y fallaba antes del primer commit. RED observado en `treehash_test.go` antes del arreglo. Ahora usa un índice descartable: incluye sin trackear, respeta `.gitignore`, no toca el índice real, funciona sin commits. Contract II corregido.
L10 · Diferido a T5: `profiletoml.Decode` de un archivo sin `name` devuelve un perfil vacío en vez de error; T5 lo carga y ahí puede fallar con un mensaje humano.
L11 · T5 no se delega con contexto recargado: la entrevista es el corazón del producto. Antes de escribirla hay que fijar las preguntas exactas y el "por qué" de cada una (adelanta parte de T8).
L12 · El humano aprueba las 7 preguntas propuestas sin cambios: "continua". Quedan escritas verbatim en Contract III, con su "por qué" al lado. Se delega T5.
L13 · T5 verificado por el padre (11 paquetes verdes, build, vet, gofmt) y con un run end-to-end propio, no el del reporte. Defecto encontrado: re-correr `start` con el mismo outcome dejaba dos task docs (0001 y 0002), y reescribir el existente habría borrado los `Pasos` ya hechos. Corregido: se reutiliza el doc cuyo slug coincide y no se reescribe nunca. RED observado en `cmd/odachi/reuse_test.go` antes del arreglo; dos tests del worker que fijaban la conducta vieja se actualizaron a la nueva regla.
L14 · Riesgo de carga de revisión, registrado y no oculto: el commit 96ccaec tiene 2.948 líneas (1.450 de tests), muy por encima de las ~400 que pide la disciplina de unidades. Causa: se delegó el paquete entero. De T6 en adelante, cada verbo se entrega como unidad propia.
L15 · Observado y diferido: (a) si no se elige ningún "NO incluye", `.odachi/fence.toml` queda vacío y el cerco no protege nada — decidir en T6/T8 si la entrevista exige al menos una exclusión o si el archivo vacío lleva un comentario que lo explique; (b) pulido de redacción para que el brief suene a la persona y no al sistema: "Quiere aprender: tests." → "Aprender en el camino: tests.", "Se usa desde el celular." → "Lo uso desde el celular."
L16 · Decisión del humano sobre cómo se gradúa: opción B con D como disparador ("continua"). Se registra cada escalación en el `Historial` del task doc —sin archivo nuevo, una sola fuente de verdad— y el agente propone en el hito; la persona acepta con `level --to`. Umbral: 3 features con pasos hechos y 5 respuestas de visión. Contract IV escrito.
L17 · T6 verificado por el padre: 15 paquetes verdes con `-count=1`, build, vet, gofmt, y end-to-end propio (start → ask ×3 → level → summary). Se commiteó en **tres unidades por capa** —dominio bdb5a6b, app 1669768, CLI e113f94— y cada una compila sola, que es lo que faltó en T5 (L14).
L18 · Inconsistencia encontrada verificando, corregida antes de commitear: `level` decía "features: 0" mientras `summary` decía "1 feature". Misma palabra, dos significados en la misma herramienta. `level` ahora dice "features con pasos" y el umbral sigue exigiendo features con al menos un paso hecho.
L19 · Límites aceptados de T6: (a) `ask`/`summary` solo reconocen `NNNN-<slug>.md`, así que este repo los ve como un proyecto vacío — deliberado, para no reescribir el documento escrito a mano; (b) un documento con secciones extra no se puede anexar hasta mover esas notas a `## Historial`, y el bloqueo lo dice; (c) `level` no cuenta un feature hasta que tenga un paso hecho, por eso un proyecto recién iniciado informa 0.
L20 · El humano delegó la decisión de T7 en el agente ("lo que tú me recomiendes para que el proyecto salga mejor"). Decidido: **B** — `check --run "<comando>"` ejecuta y captura, y `--proof <archivo>` queda para agentes y CI. Regla agregada: el archivo de prueba lleva el hash del árbol en un header; sin header se rechaza, porque una prueba que no dice de qué árbol salió no prueba nada. `--run` devuelve el exit de `check`, nunca el del comando. Contract V escrito.
L21 · T7 verificado por el padre: 19 paquetes verdes con `-count=1`, build, vet, gofmt, compilación cruzada de `shellcmd` (linux, darwin, windows, wasm, plan9) y un end-to-end propio de ocho pasos: start → check (falta) → `--run true` → editar y ver `stale` → `--run "echo ok"` → archivo prohibido (exit 2) → init → hook ejecutable. Tres commits por capa: 0520b83, 64d51e6, 56ebf87.
L22 · Defecto encontrado verificando, corregido test-first: `check --run "true"` informaba "falta la prueba" y mandaba a `--proof`, un dead end que no resuelve nada. Causa: la regla de L6 ("salida vacía = falta la prueba") era anterior al header de árbol. Con header, un silencio exitoso **es** prueba válida — `go vet`, `tsc --noEmit` y `gofmt -l` callan cuando están bien. Ahora "falta" significa solo "no hay prueba con árbol". RED en `internal/domain/gate/silent_test.go`.
L23 · Decisión de implementación que vale registrar: `--run` escribe `proof/.gitignore` **antes** de hashear, porque la prueba vive dentro del worktree y escribirla cambiaría el hash que certifica (se autoinvalidaría). Las pruebas son generadas; el `.gitignore` se versiona. Riesgo aceptado: una corrida fallida no escribe prueba, así que una prueba buena anterior del mismo árbol sigue valiendo — el árbol es byte-idéntico, así que la evidencia es correcta, y la falla ya se informó en esa misma corrida.
L24 · Sigue abierto L15(a). Decisión tomada para T8: si no se elige ningún "NO incluye", `fence.toml` se escribe con un comentario que explica que está vacío y cómo agregar una regla, en vez de un archivo vacío que no explica nada.
L25 · T8 se hizo **inline, no delegado**: la voz del tutor es oficio de escritura, no implementación. Se agregó al bloque de `AGENTS.md` la sección `## Si el perfil es novato` con seis reglas: porqué en una línea por decisión, `ask --craft` con su razón, decir por qué cuando se frena algo, una pregunta de vuelta de cada cinco, nunca un sermón ni un glosario, y `summary` narrado en tres líneas al cerrar una feature.
L26 · Dead end encontrado en el e2e: sin exclusiones, `check` decía "el brief está incompleto: falta no incluye" y mandaba a `odachi start`, que con las mismas respuestas reproduce lo mismo. Causa: `brief.Missing()` exigía "no incluye" mientras la entrevista lo deja opcional. Resuelto haciendo opcional esa parte: forzar una exclusión inventa una regla que la persona no quiso, y una violación falsa del cerco enseña mal. `Missing()` exige solo "qué tiene que existir" y "listo cuando". RED en `nongoals_test.go`.
L27 · Defecto encontrado en el e2e del hook: sin `odachi` en el PATH, el hook terminaba en `command not found` y bloqueaba cada commit sin explicar nada — el caso normal de alguien que bajó el binario. `init` ahora escribe la ruta absoluta del ejecutable (`Options.Exe`, resuelta por `hookExe()` en el CLI) con fallback a `odachi` del PATH, y la última línea sigue siendo la regla. Verificado: bloquea sin prueba nombrando la salida, pasa con prueba, y un commit real entra.
L28 · Límite conocido y aceptado: el bloque de `AGENTS.md` solo está fijado por aserciones `contains`, así que un cambio futuro podría borrar la sección del tutor sin que ningún test falle. Es contenido, no comportamiento; si vuelve a pasar, se pinea el bloque completo.
L29 · T9 verificado por el padre y por un verificador independiente (`gentle-ai-verify`): 6 de 7 propiedades PASS. La séptima figuraba como FAIL porque `odd/**` estaba modificado — era mi propia edición del spec, no trabajo del worker. 19 paquetes verdes, build, vet, gofmt, `bash -n`.
L30 · Validado con artefactos REALES, no con fixtures: `goreleaser release --snapshot --skip=publish` construyó los 6 targets, las 6 archives y `checksums.txt`; el glob `LICENSE*` vacío **solo avisa**, no rompe la release. `scripts/install.sh` instaló desde `dist/` con `file://` (con un shim de `uname`, porque el script soporta linux/darwin y esta máquina es Windows) y con el checksum alterado salió 1 sin dejar **nada** en el destino. `install.ps1` instaló `odachi.exe 0.0.1-dev` desde un HTTP local y el binario reportó su versión.
L31 · Tres correcciones nacidas de la verificación, no del brief: (a) `install.sh` copia a un nombre temporal y hace `mv` en el mismo directorio, para que un corte no deje un binario truncado donde había uno bueno; (b) `install.ps1` fija `[Console]::OutputEncoding` en UTF-8 — medido: el archivo ya era UTF-8 válido y lo que estaba mal era la salida (`82` en vez de `c3a9`), comprobado en PowerShell 7.6 y 5.1; (c) `--version` se atiende antes de `os.Getwd()`, así que un directorio borrado ya no impide preguntar la versión (probado con el cwd eliminado).
L32 · Lo que sigue SIN verificar, y se dice: la rama "última release" de `install.sh` (los tests siempre pasan `ODACHI_VERSION`), el fallback `shasum` (en esta máquina hay `sha256sum`), la escritura del PATH de usuario en `install.ps1` (probé una copia con esa línea neutralizada: no toco el registro del humano) y `go test` en los runners de GitHub.
L33 · Deuda declarada del verificador: `install.ps1` no tiene test automatizado. Se cubrió a mano en Windows con un artefacto real; automatizarlo requiere un listener HTTP en CI. Siguiente decisión pendiente: L15(a), el `fence.toml` vacío cuando no se elige ningún "NO incluye".
L34 · La primera CI en la nube encontró un defecto que esta máquina no podía encontrar: en windows-latest el shim de uname no ganaba en el PATH, el instalador se negaba correctamente a instalar en MINGW64, y el test tomaba esa negativa como una instalación verificada. Arreglado en la causa: install.sh acepta ODACHI_OS y ODACHI_ARCH validados contra la misma lista que los valores detectados, y el test declara la plataforma. Run 39657162034: los seis jobs en verde, windows-latest incluido.
L35 · Publicado v0.1.0 con licencia MIT (sin LICENSE el repo era visible pero no reutilizable: leer no es permiso). El workflow de tag corrió los tests antes de publicar y subió 6 archives + checksums.txt. Después de publicar se verificó la cadena completa contra la release REAL, no contra fixtures: (a) `install.sh` sin `ODACHI_VERSION` resolvió la última release desde la API y instaló 0.1.0 — la rama que L32 marcaba como nunca ejercitada; (b) `install.ps1` instaló odachi.exe y el binario respondió `odachi 0.1.0`; (c) el comando del README, `curl -fsSL .../install.sh | bash`, instaló 0.1.0 tal cual está escrito. El job de macOS pasó el test del instalador, así que la verificación de sha256 también funciona donde `sha256sum` no existe.
L36 · T12 verificado por el padre: 21 paquetes verdes, y las pruebas que importan hechas a mano. Rompí SKILLS.md agregando una fila → `--check` exit 1 con el mensaje que manda a `--write`; `--write` restauró byte a byte. Borré el artefacto que dejó el e2e del worker (`skills/ejemplo/`) y regeneré el registro: 5 skills, 31–33 líneas cada una. El registro es un **comando**, no una skill: gentle-ai mantiene su tabla con un prompt y audita el drift con un test ratchet; acá el archivo es derivado y comparado byte a byte.
L37 · T13 verificado por el padre: 24 paquetes verdes y un e2e propio con home falso. `pack --agents claude,codex` instaló los 10 archivos y `cmp` confirmó bytes idénticos a la fuente; la segunda corrida no escribió nada ni dejó backups; un archivo editado a mano se respaldó antes de pisarse, con el texto del humano intacto en el `.bak`; `pi` e `inventado` se rechazaron con su razón. El proyecto quedó intacto: `pack` no toca `skills/` ni `SKILLS.md`. Decisión: el paquete que embebe las skills vive en `skills/` mismo (`skills/embed.go`) porque `//go:embed` no puede leer fuera de su propio directorio, y duplicar los archivos habría creado dos fuentes de verdad.
L38 · Mantenimiento: las tres acciones de CI iban tres majors atrás (checkout v4, setup-go v5, goreleaser-action v6) y el runner las forzaba a Node 24. Subidas a v7, v7 y v7; la suite lo verifica en el mismo push que las lleva.
L39 · CI roja en ubuntu y macOS, verde en Windows, para el mismo commit: el escáner de skills deducía "no es una skill" de la clase de error (`ENOTDIR` en POSIX, "no existe" en Windows), y `skills/embed.go` —un archivo dentro de `skills/`, puesto ahí a propósito para que `//go:embed` no duplique los archivos— rompía en Linux. Arreglado en la causa: el puerto gana `IsDir` y la entrada responde qué es. El doble de test ahora modela POSIX, así que el defecto se reproduce en cualquier máquina. RED observado antes del arreglo. Run 39667468391: 24 paquetes y 6 jobs verdes. Regla nueva del proyecto: cuando un test puede pasar por la clase de error de una plataforma, el doble modela la plataforma difícil.
L40 · T15 verificado por el padre: 25 paquetes verdes y un e2e propio sobre un proyecto nuevo. `plan` agregó T1 con `--req R1` y T2 sin vínculo; `--done T1 --commit deadbee` marcó y registró; repetirlo fue no-op (`ya estaba hecho`); `--done T9` listó el paso abierto. Con una sección `## Notas` agregada a mano, `plan` se negó y el sha256 del documento quedó idéntico: no escribió nada y no dejó backup. El worker encontró un defecto real en el camino: el guard comparaba líneas y bloqueaba marcar un paso, porque su propia línea canónica cambia. La respuesta fue `taskdoc.Unaccounted`: solo cuenta como pérdida lo que la gramática no puede representar.
L41 · Mi e2e encontró un defecto de UX que el reporte del worker había tapado usando `--file`: `undo` sin filtro restauraba `fence.toml`, porque "el más reciente por mtime" queda dominado por el orden interno de escritura de `start`. Ahora un solo candidato restaura, varios se rechazan listándolos y pidiendo `--file`; y si los bytes actuales ya son los del backup dice `no hubo cambios` sin escribir (antes imprimía el mismo backup como origen y destino). Dos tests fijaban la regla vieja y se actualizaron: **es un cambio de contrato, no un defecto del worker**, y quedó anotado en Contract X.
L42 · T17 verificado por el padre con el binario: un lock fresco se rechaza con verbo, pid y antigüedad y el `brief.md` queda intacto (sha256); un lock de 20 minutos se toma con un aviso de una línea y se libera; `plan` no crea lock; un `start` que falla no deja lock. Decisión del worker que vale registrar: `Release` compara los bytes que escribió, así una corrida que sobrevive diez minutos no borra el lock del que la reemplazó. Agregué `.odachi/lock` al `.gitignore`: una corrida matada lo deja a propósito y sin eso aparecería como sucio.
L43 · T18 verificado por el padre con el binario y un home vacío: `pack --agents opencode` escribió las 5 skills en `.opencode/skills/` del proyecto con bytes idénticos a la fuente y el home quedó **vacío**; la segunda corrida no escribió ni dejó backups; `pack` sin `--agents` detectó opencode por `.opencode/`; `pi` sigue rechazado y la lista de soportados ahora incluye opencode. Corrección de un argumento mío: rechacé opencode por "path dependiente de la plataforma", y eso solo era cierto del path global. El del proyecto es relativo, viaja con el repo, y es el correcto para una herramienta de proyecto. Fuente: la documentación de OpenCode y su código (`{skill,skills}/**/SKILL.md`).
L44 · Primera corrida real con OpenCode como agente, en un proyecto con brief, cerco y skills. Con `odachi` **fuera** del PATH: 41 shell calls, 7 buscando el binario (incluida una recursión de PowerShell sobre `C:\`) y **cero** verbos ejecutados. Con `odachi` en `~/bin`: 12 shell calls y **sí** usó `odachi check`, `--help` y `--version`. Las skills de odachi (`acceptance`, `data`, `tests`) se cargaron en las dos por trigger. Conclusión: el contenido funciona cuando la herramienta es alcanzable; lo que fallaba no era la obediencia sino la instalación. Ambas corridas murieron con `503 Endpoint is unavailable` del proveedor antes de terminar la implementación, así que el veredicto sobre `plan` y `craft` sigue pendiente. Segundo hallazgo: el prompt completo de gentle-ai está **embebido en `opencode.jsonc`** (148 KB), así que desinstalarlo de OpenCode no es mover archivos; para aislar odachi hace falta un HOME con solo `auth.json` y una config mínima.
L45 · Tercera corrida, la que vale: HOME aislado con solo `auth.json` (sin `opencode.jsonc`, sin skills globales), `--standalone`, mismo proyecto y mismo prompt. **exit 0, 54 shell calls, cero errores.** Resultado: 3 pasos agregados con `odachi plan` y cerrados con `plan --done` registrando su commit; 6 decisiones `OFICIO` con su `porque`, cada una citando una línea del brief; `check --run "go test ./..."` corrido varias veces con `proof/0001-…txt` atado al hash del árbol; 4 commits, uno por paso; **cero archivos de pagos** (el cerco se respetó sin que nadie lo mirara). El agente además descubrió `--help`, `--json`, `skills --check` y `check --base ""` por su cuenta. Veredicto: el contenido se obedece cuando (a) el binario es alcanzable y (b) no hay instrucciones compitiendo. Dos observaciones del entorno, no del producto: los mensajes de commit salieron con acentos rotos por la code page CP1252 de PowerShell (el mismo defecto que arreglamos en `install.ps1`), y el agente se puso `git config user.name "odachi"` en vez de usar la identidad del humano.
L46 · Consulta a gentle-ai para diseñar el review. Se toma: la compuerta de precisión ("si dudás, callate; estilo y preferencia prohibidos salvo que oculten un defecto"), el vocabulario `BLOCKER|CRITICAL|WARNING|SUGGESTION`, la regla de que sin findings la respuesta es exactamente `No findings.`, la causalidad (`introducido/activado/empeorado` por el candidato) y un barrido por lente sin loop hasta secarse. No se toma: ledger en Engram, refuters con voto 2-de-3, envelopes de consentimiento ni lineages. El review de odachi es un archivo por árbol (` .odachi/reviews/<treehash>.json`): la aprobación no puede flotar porque es literalmente el nombre del archivo.
L47 · T20 verificado por el padre con el binario: el prompt lleva el brief verbatim y las 5 reglas (6/6 marcas); `--attach` con `approved` y `check --require-review` da listo; el segundo attach al mismo árbol se rechaza con `siguiente: odachi review --show`; un árbol nuevo con BLOCKER causal da `changes_requested` y el gate sale 2 nombrando **esa** finding; y un attach inválido sale 1 **sin escribir** (2 archivos de review antes y después). Deudas declaradas: (a) el primer `review` agrega `.odachi/reviews/.gitignore`, lo que invalida la prueba por única vez — recomiendo que ese archivo lo cree `start`; (b) el diff vive en `cmd/odachi/reviewgit.go` y debería mudarse a `gitx.Diff`; (c) `--exec` no puede recibir el prompt por stdin porque `shellcmd` no tiene ese puerto.
L48 · Intenté arreglar el papercut del store de review haciendo que `start` creara `.odachi/reviews/.gitignore` y **lo revertí**: dos tests fijaban la invariante de "cuatro artefactos" y tenían razón — un quinto archivo en el primer `start` la rompe, y además queda versionado algo que es plumbing. El fix correcto es `.git/info/exclude`: está fuera del worktree, así que no cambia el hash del árbol ni agrega un artefacto. Queda como deuda con nombre y forma. Consecuencia mientras tanto: el primer `review` en un proyecto invalida la prueba una vez y hay que repetir `check --run`.
L49 · T21 verificado por el padre: `docs` → passive sin lentes; código bajo `internal/` → solo `reliability`; tocar `auth/` → las cuatro en orden fijo; `--lens seguridad` rechazada. El loop del refutador completo: adjuntar un veredicto con un BLOCKER causal **no mueve el árbol** (el veredicto está ignorado por el `.gitignore` del store, que es lo que hace usable el flujo), el gate sale 2 nombrando ese blocker, adjuntar una refutación con cita del diff lo destraba (`listo`), y una segunda refutación se rechaza. 31 paquetes verdes.
L50 · Casi reporté un bug que era mío: el primer e2e del refutador dijo "la prueba es de otro estado del código" y estuve a punto de acusar al producto. La causa era mi propio directorio `fuera/` con los JSON de prueba **dentro del proyecto**: el candidato es el worktree, así que un archivo de scratch adentro es parte del candidato. Regla para mí: los archivos de prueba van fuera. Lo que sí quedó confirmado es la propiedad importante: el archivo de veredicto está ignorado por el store, así que adjuntarlo no cambia el hash.
L51 · `doctor` verificado: 10 chequeos en verde en un proyecto recién configurado (exit 0), y cada rotura produce el problema correcto con el comando correcto: sin `AGENTS.md`, sin hook, perfil inválido, brief incompleto. La huella del proyecto fue idéntica antes y después: no escribe.
L52 · Primera lectura del bench: 8 recorridos, 5 bloqueos, **3 dead ends**. Analicé los tres y **solo uno era defecto del producto**: `start` y `start --force` salen 2 porque piden respuestas, y por contrato el exit 2 es "solo el humano puede destrabar" — el instrumento no distinguía eso de un callejón. El tercero sí era real: el cerco nombraba `odachi ask` pelado, que imprime el uso incluso para una persona. Arreglé las dos cosas: el bench gana el tipo `interactive` (se cuenta y se lista, nunca se esconde) y el cerco nombra la forma del comando. Regla que queda: un instrumento que llama callejón sin salida a un handoff humano está midiendo su propia ceguera.
L53 · Segunda lectura, después del arreglo: `in_band 1 · out_of_band 0 · templated 2 · interactive 2 · dead_end 0`, exit 0. El ratchet sigue mordiendo (hay un test de app donde un dead end hace fallar la corrida), así que el verde no es permisividad: es que ya no hay mensajes que manden a un lugar sin salida.
L54 · Las dos deudas del doctor, arregladas y fijadas con tests. Worktree: `.git` es un archivo y ahora dice `sí: worktree (el .git real está en el repo principal)`; el check del hook manda a correr `init` **desde el repo principal** en vez de "corré git init", que era mentira y callejón. Perfil: `name = "inventado"` sin `extends` ahora es un problema que nombra los tres perfiles y el `extends` que lo arreglaría, y `doctor` delega en `loadProfile` para que la decisión tenga un solo dueño. Dos errores míos en el camino, ambos corregidos por tests: puse la regla en `profile.Validate` (donde "vacío = heredado" es correcto) y en la delegación devolví `inferred = true` también cuando el archivo existe, así que un perfil válido se reportaba como inexistente.
L55 · Odachi no tiene orquestador, y es un hueco real: el orquestador de gentle-ai (routing, triggers de delegación, budget de evidencia, handoff) es una capa de política que odachi no tiene. Decisión: no crear esa capa. Su sustancia cabe en la skill que ya se dispara con "partir un pedido grande": delegar por razón, pasar el task doc **por path** y no un resumen, exigirle `plan` y `check`, verificar su salida contra el brief y no contra su reporte, y no delegar nunca la decisión de visión. Contenido que se carga cuando hace falta, en vez de cuatro páginas inyectadas en cada sesión. `slices` quedó en 49 líneas, validado por `skills --check`.
L56 · Corrección del humano, aceptada: delegar **sí** protege el contexto del padre, y mi versión de 49 líneas se había quedado sin umbrales. Lo que no es cierto es que ahorre tokens totales: un subagente arranca sin caché y para algo chico sale más caro. La regla quedó: **delegá lo que se lee mucho, no lo que se escribe poco**, con presupuesto medible (~3 lecturas o ~10k tokens, o sesión larga) y retorno acotado (≤2k tokens, evidencia `path:line`, lista de archivos leídos) — porque el retorno acotado es lo que mantiene limpio el contexto. `slices` quedó en 54 líneas y el trigger se amplió a "entender el código ya no entra en una lectura".
L57 · Comparación de bibliotecas, medida: gentle-ai 15 skills / 1.148 líneas (la mayor, `branch-pr`, 201); odachi 5 / 182. Pero la diferencia no es el tamaño: **8 de sus 15 son para mantener su propia herramienta** (branch-pr, issue-creation, skill-creator, work-unit-commits, chained-pr, judgment-day, skill-improver/registry, rdd-defect-workflow) y ninguna enseña a construir y entregar software. Las 5 de odachi son el currículum del que construye. El humano pidió cerrar los tres huecos que el uso real pide: `commits` (32 líneas), `writing` (31) y `skill-craft` (32). Quedan 8 skills, 277 líneas, todas bajo el cap. Efecto colateral útil: agregar skills rompió seis tests que fijaban "cinco" a mano; ahora el conteo se pregunta al payload y al directorio, porque un cambio de contenido no debe obligar a editar un test.
L58 · El hueco más grande para el público elegido, encontrado al preguntar "¿y si no sabe qué es git?": `odachi init` se negaba con "corré git init" a alguien que no sabe qué es git, y `check` necesita el historial. Cambio: **`init` crea el historial** y lo dice en castellano llano, nombrando `git init` entre paréntesis junto a para qué sirve ("es lo que me deja probar que una prueba corresponde a este estado del código"). Sin git en la máquina sigue bloqueando, con una frase que dice qué instalar. El test que fijaba la negativa se reescribió para fijar la creación. Queda pendiente de redacción: la línea de cierre de `init` dice "commiteá", que es vocabulario de git para el humano — el que commitea es el agente.
L59 · Cerrado el pendiente de redacción de L58: la línea de cierre de `init` decía "commiteá", que asumía que el humano commitea. Ahora dice "cuando odachi check --quiet esté en verde, el agente guarda el avance". Un test fija las dos mitades: la palabra no está y el reemplazo sí.
L60 · Retomado el trabajo con tres deudas registradas (L47): (a) el prompt de `--exec`, (b) `gitx.Diff`, (c) `.git/info/exclude` para el store. Se toma (a) primero porque es el único de los tres que es un defecto visible: `--exec` corría el revisor **sin prompt**, y el único modo de dárselo era `odachi review --exec 'agent "$(odachi review)"'`, que no existe en `cmd /c` y calcula el diff dos veces. Decisión: el prompt viaja en un archivo temporal que es el stdin del revisor **y** cuya ruta sustituye a `{prompt}` en el comando. Las dos formas salen del mismo archivo: stdin porque es la tubería natural, y la ruta porque el destino real (un agente como OpenCode) no lee stdin por defecto y `$(cat)` no es portable. El archivo se abre como stdin, no por pipe, así un revisor que ignora stdin no puede bloquear a odachi llenando el buffer.
L61 · T25 hecho inline y en dos commits por capa. Lo que encontré al abrirlo era más grande que la deuda: `--exec` no solo no pasaba el prompt, sino que **en Windows cualquier comando con comillas llegaba roto**. Causa medida, no supuesta: `cmd.exe` no es un programa de argv, parsea su propia línea y no entiende el escape con barras invertidas que Go aplica a un argumento con comillas, así que `--run "go test -run \"TestX\""` llegaba como error de sintaxis o con las barras literales. Arreglado con `/s /c "<comando>"`, la forma documentada de entregarle a cmd un string tal cual. Sin ese arreglo, `{prompt}` no podía funcionar nunca: una ruta entre comillas era justo el caso roto.
L62 · Decisiones de T25 que vale registrar. (a) Un solo constructor de prompt (`promptText`) para las dos formas, porque dos constructores se separan y un revisor que contesta otra pregunta que la que hizo la compuerta es peor que ningún review. (b) El archivo temporal se abre como stdin, no se escribe por pipe: un revisor que ignora su entrada no puede bloquear la corrida con el buffer lleno (test con 780 KB). (c) La ruta se cita **siempre**, no solo cuando tiene espacios: el adaptador no puede saber cómo se llama el temporal en otra máquina. (d) Sin prompt, `{prompt}` queda literal: no se inventa un archivo que no existe.
L63 · Consecuencia aceptada: como armar el prompt prepara el store, un `--exec` que falla deja `.odachi/reviews/.gitignore`. No afirma nada, y el prompt tiene que nombrar el árbol donde se archivará el veredicto, que es lo que ese archivo cambia. El test que exigía "no escribe nada" pasó a exigir "no escribe ningún veredicto", con la razón escrita al lado. Otra consecuencia, declarada en el ADR: un revisor que escribe en el worktree corre el árbol, el veredicto se archiva bajo el árbol corrido y la compuerta pide una prueba nueva — el mismo camino de auto-curación que cualquier cambio.
L64 · Verificación de T25: 35 paquetes verdes con `-count=1`, `go vet`, `gofmt`, build cruzado (linux, darwin, wasm, plan9) y `go vet` cruzado con CGO_ENABLED=0. Del lado del binario real, en un proyecto con brief: `review` imprimió el prompt, un revisor que lee stdin guardó su veredicto, un segundo árbol con `{prompt}` guardó el suyo, `--show` leyó los dos, `check --require-review` dijo "listo" y no quedó ningún temporal. Límite declarado: **los tests de unix no se ejecutaron en esta máquina** (no hay Go en WSL y el demonio de Docker no está corriendo); están compilados y veteados para linux/darwin, y los corre la CI en el push. Los de Windows, que es donde vivía el defecto, sí corrieron.
L65 · Deuda nueva, vista al verificar: el README quedó atrás (dice cinco skills, v0.2.0 y diez superficies; van ocho skills, v0.9.1 y catorce superficies, y `review`, `doctor`, `bench` y `undo` no figuran). No se toca en T25 para no mezclar documentación de estado con un cambio de comportamiento.
L66 · T26: README puesto al día con hechos medidos, no con memoria. Dice ocho skills (la tabla sale de `SKILLS.md`), quince superficies (`odachi --help`), 35 paquetes (`go list`), `v0.9.1` con trece tags (`gh release list`), CI en tres sistemas con carreras sólo en ubuntu (el workflow), y las señales que suben el tier de review (los `signals` de `tier.go`, no una lista inventada). Se fue la frase "el pipeline de release todavía no corrió / este repo no tiene remoto", que era cierta hace ocho versiones y hoy es falsa. Se agregó la sección de review, que era el hueco más grande: la herramienta ya lo tenía y el README no lo contaba. También quedó escrito el precio de la evidencia generada: el primer `check --run` invalida un review y el primer `review` invalida una prueba, una vez cada uno — es la deuda de L48 dicha en voz alta para el que la va a sufrir.
L67 · T27 delegado al worker (deuda de L47b) y verificado por el padre, no por su reporte: `gitx.Diff` vive en el adaptador, reusa el runner de gitx en vez de una tercera copia del pin de entorno, se lleva la constante del árbol vacío y la regla de que `""` y ese hash significan lo mismo, `reviewGit` desapareció y `gitRepo` es el único adaptador de git del binario. Seis tests nuevos donde antes no había ninguno: archivo sin trackear incluido (la razón de ser del índice descartable), ignorado excluido, repo sin commits, `""` igual al hash del árbol vacío byte a byte, índice y `git status` intactos después, y base irresoluble con el stderr de git. Mi propia corrida con el binario: el prompt lista el archivo modificado y el que no está trackeado, y `check --run` sigue diciendo listo. Corrección mía sobre la entrega: los mensajes de falla de los tests venían en castellano mientras el resto de los tests del repo los escribe en inglés; unificados. Queda la duplicación de la constante del árbol vacío en `check` y `review` — el worker no la tocó porque estaba fuera de su superficie, y unificarla es otra unidad.
L68 · T28, la última de las tres deudas de L47: el humano eligió, entre tres opciones, que `init` mantenga un bloque marcado y versionado en el `.gitignore` del proyecto (`/proof/` y `/.odachi/reviews/`) en vez de usar `.git/info/exclude` (invisible y no sobrevive a un clon) o dejar el retake. Delegado al worker en cinco commits por capa, cada uno verde en aislamiento (worktree temporal por commit, `go build` + suite completa). La verificación del padre con el binario real: `init` escribe el bloque, `check --run` deja la prueba, `review --attach` guarda el veredicto, `check --require-review` sale 0 con la prueba byte a byte igual y **sin un solo `.gitignore` interno**, `doctor` 10/10, y un segundo `init` no escribe nada. En un proyecto sin `init` el respaldo interno sigue apareciendo, que es la otra mitad.
L69 · Decisiones del camino, todas medidas. (a) `git check-ignore -q proof` sale 1 cuando la carpeta todavía no existe y `/proof/` es una regla de directorio —el patrón solo matchea directorios y git no puede saber que ese nombre lo será—, así que el adaptador pregunta dos veces, con y sin barra final; el comentario y el test lo fijan. (b) Un bloque truncado (marcador de inicio sin fin) se repara en vez de duplicarse. (c) `init` no respalda el `.gitignore`: la edición solo agrega y preserva cada byte, un `.bak` sería ruido. (d) `doctor` decide con un matcher puro sobre el `.gitignore` del proyecto y el archivo interno, no con git: no tiene puerto de git y agregarlo tocaba su cableado; el costo es un falso negativo si alguien ignora la evidencia por `.git/info/exclude`, y en ese caso el arreglo que nombra (`odachi init`) sigue siendo correcto y barato. (e) `review` dejó de reportar siempre el archivo interno del store en su salida de máquina: ahora reporta solo lo que realmente escribió.
L70 · Límite de migración, declarado: un proyecto viejo conserva su `proof/.gitignore` versionado y ese archivo sigue entrando al árbol (git no ignora archivos ya trackeados), así que el retake único no desaparece para él hasta que lo saque con `git rm`. No se borra nada automáticamente: es un archivo del humano, en su índice.
L71 · El humano señala el hueco de fondo, con estas palabras: "si el usuario no sabe y el agente tampoco sabe usar odachi, como el agente va a iniciar un proyecto?". Medido en una carpeta vacía: `doctor` informa "4 problemas" con tres comandos distintos (`git init`, `odachi init` ×2, `odachi start`), y `odachi` pelado pide cinco flags de entrevista. Ninguno dice lo único que hace falta: "esta carpeta todavía no es un proyecto; empezá por acá". Es un problema de orden: el canal que enseña al agente (`AGENTS.md`, las skills) lo escribe odachi, y para escribir odachi hay que saber que existe.
L72 · Investigación de los estándares (pedida por el humano: "investiga en internet como se deberia hacer correctamente, y que se pueda implementar en cualquier agente si problemas en un futuro"). `AGENTS.md` es el estándar de reglas de proyecto (agents.md, Agentic AI Foundation, 60k+ repos, 16 de 19 clientes documentados lo leen) y el archivo más cercano gana; **no existe un AGENTS.md global estándar**: cada cliente tiene su ruta (`~/.config/opencode/AGENTS.md`, `~/.codex/AGENTS.md`, `~/.claude/CLAUDE.md`, `~/.gemini/GEMINI.md`, `~/.copilot/copilot-instructions.md`). `Agent Skills` sí tiene ruta cross-client: la guía oficial del estándar pide escanear, por scope, la carpeta nativa del cliente **y** `.agents/skills/` — proyecto `<proyecto>/.agents/skills/`, usuario `~/.agents/skills/` — y dice que esa convención "se consolidó como ampliamente adoptada". Codex la documenta como scope USER; OpenCode escanea `.agents/skills` como directorio externo; precedencia estándar: proyecto gana sobre usuario. La carga es por divulgación progresiva: catálogo de ~50-100 tokens por skill, cuerpo (<5000 tokens recomendados) recién al activarse.
L73 · Y el cliente ya tiene selector de persona, que odachi estaba ignorando: el humano lo dice así: "en opencode es facil elegir un perfil con shift mas tab". OpenCode cicla **agentes primarios** con Tab / `switch_agent` y los define en `.opencode/agents/<nombre>.md` (frontmatter + body como system prompt); Claude Code tiene **output styles** en `.claude/output-styles/<nombre>.md` elegibles con `/output-style`; Codex tiene **perfiles** pero son capas de config, no persona. Diagnóstico del error de diseño: odachi mete dos ejes distintos en la palabra "perfil" — la ceremonia que ejecuta el binario (`brief_source`, `questions`, `proof`, `output`) y la persona del agente (`explain = why|none`) — y solo honra el primero. Ningún agente de OpenCode puede cambiar lo que hace el binario, y el binario no puede cambiar cómo habla el agente.
L74 · Decisión: tres capas, de abajo hacia arriba, y la de abajo nunca falla. (1) **Invariante**: `AGENTS.md` del proyecto con reglas + identidad. (2) **Portable**: skills en `.agents/skills` (proyecto) y `~/.agents/skills` (arranque global, on-demand, ~100 tokens en reposo). (3) **Nativo**: personas generadas desde el perfil donde el cliente las tiene, con una primera línea que impide la doble fuente de verdad ("la ceremonia la manda `.odachi/profile.toml`; esto es solo cómo hablás; si no coincide, pedí `odachi level --to <perfil>`"). Se descarta escribir un AGENTS.md global: no hay ruta estándar y una compuerta global le mete reglas falsas a carpetas que no son proyectos de odachi. Un comando: `init` deja el proyecto listo (reglas, hook, `.gitignore`, skills del proyecto, skill de arranque) y dice el siguiente paso. Costo declarado: `.agents/skills` viaja con el repo, así que un repo ajeno es input no confiable — el propio estándar pide pedir confianza antes de cargar skills de proyecto; por eso la skill de arranque no ejecuta nada, solo explica el flujo.
L75 · Orden de trabajo: T29 (identidad + doctor de carpeta vacía), T30 (`pack` portable), T31 (skill de arranque + `init` de un comando), T32 (personas generadas). La dependencia manda: T31 y T32 necesitan los destinos estándar de T30.
L76 · T29 verificado por el padre con el binario. El bloque de `AGENTS.md` abre ahora diciendo qué es odachi, quién decide y dónde está el contexto, y el tutor ya no depende de un archivo que el modelo no lee: si no sabe el nivel, pregunta una vez. El bloque quedó **fijado byte a byte**, con lo que se cierra el límite aceptado en L28 (antes solo `contains`, y la sección del tutor podía desaparecer sin que fallara nada). `doctor` gana un pre-vuelo: carpeta sin `.git` (directorio o archivo), sin `brief.md`, sin `AGENTS.md` y sin `.odachi/` → una situación y un paso (`init`, después `start`), exit 1, sin escribir nada, y `--json` lleva las mismas dos líneas. Un error de lectura no es veredicto: cae en los diez chequeos en vez de mentir.
L77 · T30: `pack` deja de tener un destino fijo por agente. La tabla gana `In(scope)` con overrides donde la ruta cambia de forma (OpenCode en el home es `~/.config/opencode/skills`, no `~/.opencode/skills`), el pseudo-agente `standard` (`.agents/skills` en el proyecto, `~/.agents/skills` en el home, la convención cross-client del estándar) y refusals que nombran las dos salidas. `--scope`, `--dest`, y `--dest` + `--scope` rechazados por contradictorios. Los resultados se agrupan por destino: dos agentes en la misma carpeta son una copia y una línea. Deuda de disciplina, registrada: el cambio se pensó en tres commits por capa y **se squasheó a dos** (ffcc676) porque el intermedio no compilaba —el CLI todavía leía la forma vieja del resultado—; un commit que no compila rompe el bisect y no se deja pasar. Costo aceptado: `pack --json` cambia de forma (filas por destino, `agents: []` adentro).
L78 · T31: la skill de arranque (`skills/bootstrap/SKILL.md`, 39 líneas) vive en `~/.agents/skills` porque es la única que tiene que existir **antes** del proyecto; su trigger la enciende para "arrancar un proyecto", "qué es odachi" o "no sé por dónde empiezo", y su sección "Nunca" prohíbe correr compuertas en carpetas que no son de odachi y `init` en repositorios ajenos. `init` instala las 8 de currículum en `.agents/skills` y la de arranque en el home, con `pack` como único dueño de cómo se instala una skill: la unión son dos `pack.Run` en el CLI, no una app importando otra. Semántica de fallas decidida de paso: un bloqueo en la copia **del proyecto** falla la corrida (el proyecto quedaría sin lo que el agente lee), un bloqueo en el **home** deja una línea `pendiente:` y el proyecto queda listo. Un proyecto que ya tenía copia de `pack --agents opencode` se reconoce en una línea y no se toca. Los tests del CLI ganan un `TestMain` que apunta `homeDir` a un temporal, para que ningún test toque el home real.
L79 · T32: dos ejes, dos dueños. `internal/domain/persona` renderiza las dos voces con un solo template y sin leer el perfil (hay un test que camina sus imports para que siga así); `novato` reusa palabra por palabra los bullets del tutor del bloque, y otro test lo fija para que la herramienta no se contradiga. `init` escribe **las dos** siempre —no puede saber quién sos, y derivarlas del perfil devolvería la decisión al TOML— en `.opencode/agents/` (`mode: primary`, se eligen con Tab) y `.claude/output-styles/` (`keep-coding-instructions: true`); el reporte nombra la tecla de cada cliente. `level --to` no las toca (verificado con md5). Evidencia de los clientes, medida: OpenCode 2.1.2 escanea `{agent,agents}/**/*.md` (strings del binario), la doc publica `.opencode/agents/`, y `opencode debug agents` lista las dos personas como primary; en Claude Code `/output-style` está deprecado a favor de `/config`.
L80 · Correcciones y límites de este bloque, dichos y no escondidos. (a) L74 estaba mal en dos cosas: las personas **no** se derivan del perfil y la línea de no-divergencia es la segunda del **cuerpo**, no la primera del archivo (cada cliente pone frontmatter primero). Se corrige acá, no se reescribió la entrada. (b) El bloque de `AGENTS.md` **no** lleva la línea de no-divergencia: la llevan las personas, que son el archivo que el humano elige; el bloque se usa cuando no hay persona elegida. (c) El chequeo 9 de `doctor` lee el registro `skills/` del proyecto, no lo instalado en `.agents/skills`: quien corrió `init` ve "ninguna: no hay skills/" y no es un error, es otra pregunta. (d) `pack --json` cambió de forma. (e) `init` escribe en el home; si no puede, deja `pendiente:` y no falla. (f) `odachi start` necesita terminal: sin TTY sale 2 tras imprimir la primera pregunta (el README documenta el camino humano, no ese). (g) El repo pasa a 36 paquetes: entró `internal/domain/persona`.
L81 · El humano señala el hueco que quedaba, con el contraste a la vista: "el que deberia saber usar todo eso es el agente asi como en gentle-ai". Medido en su repo (`C:\Users\USER\Desktop\herramientas\gentle-ai`): el agente de gentle-ai sabe manejar la herramienta porque su contrato viaja **siempre encendido**, espejado por cliente — `internal/assets/<cliente>/orchestrator.md`, 168 a 188 líneas por cliente (opencode, claude, codex, cursor, qwen, windsurf, antigravity) — más personas propias (`persona-gentleman.md`, `output-style-gentleman.md`), agentes especialistas (`review-risk`, `review-refuter`, `jd-judge-a/b`) y 12 skills de 1.119 líneas. Total: ~3.4k líneas de prompts y 33k líneas de Markdown. Odachi, en cambio, tenía identidad (bloque de ~50 líneas), arranque (`bootstrap`) y oficio (8 skills), pero **nadie le decía al agente cómo se maneja la herramienta por dentro**: qué verbo va en cada momento, qué significa cada exit, qué pide un bloqueo.
L82 · Decisión: la paridad no se compra espejando 180 líneas por cliente, se compra en la capa que ya está diseñada para cargarse cuando hace falta. Nueva skill `drive` (54 líneas): las cuatro piezas y su dueño, el ciclo de una feature en orden, los verbos en un párrafo, qué significan 0/1/2 (y que un 2 no se reintenta), el review por árbol y cinco "nunca" — el primero, escribir una prueba a mano. El bloque de `AGENTS.md` gana una línea que obliga a cargarla ("dentro de este proyecto, cargá `drive` antes de actuar") porque un modelo que tiene que inferir que debe leer el manual no lo lee; y `bootstrap` gana la línea que hace el traspaso: con el proyecto armado, el manual del día a día es `drive`. Efecto medido: `init` instala 9 skills en `.agents/skills` (entra `drive`) y 1 en el home; 36 paquetes verdes, `skills --check` verde, el commit verde en aislamiento.
L83 · Primera verificación con un agente real, no con fixtures: OpenCode 2.1.2 headless (`--standalone`) sobre un proyecto recién inicializado, con HOME aislado (solo `auth.json`, sin el `opencode.jsonc` de 148 KB de gentle-ai), para que lo único que el modelo pueda leer sea lo que odachi escribió. Tres corridas. (1) "¿qué es odachi y cómo se arranca?": contestó las dos cosas bien y citó las cinco piezas del ciclo, leyendo **solo `AGENTS.md`** — y avisó que en la raíz no había `SKILLS.md`. (2) "¿qué skills tenés y cuál cargarías primero?": listó las nueve de `.agents/skills/` y dijo que cargaría `drive` primero **porque `AGENTS.md` lo pide antes de actuar**, y detectó solo que el proyecto no tenía brief, así que el primer paso era `odachi start`. (3) Después del arreglo: leyó `AGENTS.md` y `.agents/skills/drive/SKILL.md`, no inventó archivos y nombró el arranque y el ciclo correctos.
L84 · Defecto que encontró la corrida (1) y que ningún test podía encontrar: el bloque mandaba a `SKILLS.md` "para el oficio", y `SKILLS.md` es el registro del repo de odachi, no un archivo que `init` escriba en el proyecto del humano. En toda instalación real esa línea apuntaba a un archivo inexistente. Arreglado en las dos referencias (el párrafo de identidad y la línea de skills) a `.agents/skills/`, que sí existe; la corrida (3) confirma que el agente encuentra `drive` ahí. Regla que queda: el artefacto que más se lee no puede nombrar rutas que dependen de en qué repo estás.
L85 · Corrección medida de la v0.11.0: la premisa de retirar las personas era falsa. OpenCode 2.1.3 lista `.opencode/agents/<n>.md` y `.opencode/mode/<n>.md` como `mode: primary` en 5 corridas sobre un repo git temporal y 3 sobre `odachi-demo`; una carpeta **sin** `.git` devuelve `[]` (ni los agentes built-in), que es el estado más probable de la medición anterior. El agente nativo vuelve, pero como **espejo del bloque**, no como segunda voz: `persona.Files(rules.Body(voice))` envuelve el bloque sin marcadores en el frontmatter de cada cliente (`.opencode/agents/odachi.md`, `.claude/output-styles/odachi.md`), `init` lo instala, `level --to` y `start` lo refrescan si existe, y el chequeo 11 de `doctor` compara bytes y nombra `odachi init` cuando falta o fue editado. Las personas por perfil de la v0.10.0 quedan como restos nombrados y nunca borrados. Tests nuevos: espejo por cliente, Tab primario, backup de una edición a mano, refresco en level/start, nota del doctor. Docs corregidas: CHANGELOG, README y ADR 0006 (segunda enmienda, con la medición y el costo de ~3 KB por cliente).
L86 · El humano abrió OpenCode en otra carpeta y no encontró la persona: el agente del proyecto solo se lista dentro de ese proyecto, y eso no es lo que significa «abrir mi agente y empezar a trabajar». Se agregó la **puerta del home**: `init` escribe también `~/.config/opencode/agents/odachi.md` y `~/.claude/output-styles/odachi.md` (las rutas globales documentadas por cada cliente), con el mismo nombre `odachi` que el del proyecto. La copia del proyecto gana dentro del proyecto; la del home es un despachador corto, no una copia del bloque: si hay bloque de `AGENTS.md` lo lee y adopta su voz, si no hay lo manda a la skill `bootstrap` para arrancar con `init` y `start`. Un home que no se puede escribir deja `pendiente:` y no falla el proyecto. Tests: la puerta, el nombre compartido, el home aislado con exactamente bootstrap + las dos puertas, y el proyecto intacto cuando el home falla.
L87 · Defecto encontrado al probar el arranque en una carpeta vacía, no en un test: `init` crea `.git`, y `start` infería `senior` de `HasRepo`, así que el primer proyecto de un principiante terminaba con perfil y voz senior, exactamente al revés del público elegido. `profile.Facts` pierde `HasRepo`; `Infer` solo promueve a senior cuando ya hay un brief (o ci en una máquina), y el resto arranca en novato. `start`, `currentProfile` y el doctor dejan de leer `.git` para decidir el perfil. Test nuevo: un repo sin brief arranca en novato. La spec original (Contract, línea 145) queda como historia; esta entrada dice por qué cambió.
L88 · Análisis de gentle-ai para la delegación, a pedido del humano. Lo que se copia: triggers obligatorios (inline ≤3 llamadas / ~10k tokens; >~5 búsquedas → un explorador read-only con handoff ≤2k y un solo spot check; escritor solo por paralelismo de unidades disjuntas o backstop de ~150k tokens; tests/builds pueden usar workers frescos sin cambiar la ruta; verificación por riesgo). Lo que NO se copia: el orquestador siempre encendido, ODD/RDD, feature document con espejo en Engram, envelopes de consentimiento, Judgment Day y la flota de agentes. Cambios en odachi: `slices` declara que su política gana sobre otras skills globales y corrige el writer trigger (nunca por tamaño ni cantidad de archivos); `drive` manda a cargar `slices` cuando el pedido no entra en una lectura; test `TestDelegationPolicyLivesInSlicesAndDrive`; ADR 0007. Motivo inmediato: una corrida real cargó `slices` y `hermes-ephemeral-delegation` y mezcló los umbrales.
L89 · Corrección de una afirmación falsa en la cabecera de este documento: decía que el próximo feature usaría Specs/Tasks/Log, «el mismo formato que escribe odachi start». `start` escribe `## Requisitos (textual, del usuario)`, `## Listo cuando`, `## Pasos` y `## Historial`; `taskdoc.Parse` solo conoce esos headers. Este documento histórico usa Specs/Tasks/Log y por eso `odachi summary`/`plan`/`level` no lo ven (L19). No se migra: cinco comentarios y el mapa de impacto lo citan por su ruta. La comparación con ODD de gentle-ai queda como está: odachi tiene el mismo espinazo, no su gramática ni su protocolo.
L90 · ODD-lite implementado a pedido del humano. Odachi no usa ODD de gentle-ai y no se copia su gramática (Specs/Tasks/Log, feature document, espejo Engram, task size, tiers de verificación, delivery chain). Se agrega la disciplina que sí cabe sin cambiar el task doc: `drive` exige requisitos citados verbatim y `Listo cuando` observable, `tests` con RED antes del código, registro del paso delegado con `ask --craft` y la línea de riesgo de `commits` (`none` o datos/seguridad/contrato/concurrencia/entrega/sin-test; un ítem pide `review`). `commits` es la dueña del vocabulario. Test `TestDriveTeachesTheWorkflowDiscipline`. ADR 0008. La compatibilidad con un feature doc ODD real sigue siendo read-only: `summary`/`plan`/`level` no lo leen; si se necesita, es un parser propio con migración y `Unaccounted`. Siguiente: dogfoodear odachi en su propio repo, con init/start y el agente de OpenCode.
