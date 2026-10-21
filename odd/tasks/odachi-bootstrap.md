# odachi · bootstrap

Feature: a CLI that makes the agent do the dirty work with senior discipline while the human decides vision, and that teaches the human until they no longer need the teaching.
Status: in progress · Branch: main · Owner: human (vision) / agent (execution) · Session: 1

## Specs

S1 · "cualquier persona sin experiencia pueda usar sus agentes y poder hacer proyectos como un profesional sinior"
S2 · "que sea organico e intuitivo para las personas que deleguen las tareas a los agentes"
S3 · "el agente haga todo el trabajo susio, con toda la disciplina de un sinior, y la persona decida las cosa cognitivas, la vision del proyecto, etc"
S4 · "tambien las personas que ya son experimentadas"
S5 · "dos tipos de perfiles, novato y sinior, e incluso que sea facil de crear nuevos perfiles, para que las personas decidan"
S6 · "si me gustarian los artefactos odd, porque tambien podrian servir como memorias o historial para que el agente entienda rapido"
S7 · "el agente tambien va a servir como tutor para que la persona pueda llegar a ser un senior, para que una persona no se quede para siempre como novato"
S8 · nombre del proyecto: "odachi"

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

- Module path `github.com/biggz/odachi`; zero external dependencies in this session.
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

## Tasks

- T1 · repo, go.mod, boundary guard test, ADR 0001, birth decisions, change checklist, impact map, artifact examples · S1-S8 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T2 · pure domain: fence (globs), gate (exit codes), profile (schema + reflection guard) · S2,S3,S5 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T3 · pure domain: brief and task document parsing + progress · S1,S6 · delegated (gentle-ai-worker) · [x] commit: c77ca91
- T4 · adapters: strict TOML profile decoder, filesystem, git · S5 · delegated (gentle-ai-worker) · [x] commit: f53fa1a
- T5 · CLI `start`: interview -> brief + fence + profile · S1,S2 · delegated (gentle-ai-worker) · [x] commit: 96ccaec
- T6 · CLI `ask` / `level` / `summary` · S3,S7 · delegated (gentle-ai-worker), three stages · [x] commits: bdb5a6b (domain), 1669768 (app), e113f94 (CLI)
- T7 · CLI `check` + pre-commit hook + AGENTS.md block · S2 · delegated (gentle-ai-worker), two stages · [ ]
- T8 · tutor content: why each question exists, explanation rules · S7 · pending

## Log

L1 · "analiza gentle-ai y dime como crearias algo mucho mejor y que no sea complicado ni demoroso de hacer" — path C:\Users\USER\Desktop\herramientas\gentle-ai
L2 · nombre elegido: odachi. Descartados: gentle-*, oficio, wright, koder-ai, dev-ai
L3 · se descarta `decisions.md` como archivo aparte: lo durable va a `brief.md`, lo cronológico al `Historial` del task doc. Una sola fuente de verdad.
L4 · el humano autoriza T1 y T2: "ok, comencemos". Path C:\Users\USER\Desktop\dev-biggz\odachi
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
L20 · El humano delegó la decisión de T7 en el agente ("lo que tú me recomiendes para que el proyecto salga mejor"). Decidido: **B** — `check --run "<comando>"` ejecuta y captura, y `--proof <archivo>` queda para agentes y CI. Regla agregada: el archivo de prueba lleva el hash del árbol en un header; sin header se rechaza, porque una prueba que no dice de qué árbol salió no prueba nada. `--run` devuelve el exit de `check`, nunca el del comando. Contract V escrito. Siguiente decisión pendiente: L15(a), el `fence.toml` vacío cuando no se elige ningún "NO incluye".
