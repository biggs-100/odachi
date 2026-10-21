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

## Tasks

- T1 · repo, go.mod, boundary guard test, ADR 0001, birth decisions, change checklist, impact map, artifact examples · S1-S8 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T2 · pure domain: fence (globs), gate (exit codes), profile (schema + reflection guard) · S2,S3,S5 · delegated (gentle-ai-worker) · [x] commit: 7aaf628
- T3 · pure domain: brief and task document parsing + progress · S1,S6 · delegated (gentle-ai-worker) · [x] commit: c77ca91
- T4 · adapters: strict TOML profile decoder, filesystem, git · S5 · delegated (gentle-ai-worker) · [x] commit: f53fa1a
- T5 · CLI `start`: interview -> brief + fence + profile · S1,S2 · pending
- T6 · CLI `ask` / `level` / `summary` · S3,S7 · pending
- T7 · CLI `check` + pre-commit hook + AGENTS.md block · S2 · pending
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
