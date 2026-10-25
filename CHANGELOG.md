# Changelog

What changed in each released version, in the order a person reads it.
Versions before 0.10.0 were rebuilt from their tags: the repo did not have this
file yet, and a changelog that starts today erases the history of whoever arrives
tomorrow.

## v0.11.15 - 2026-10-25

The review's answer belongs to the tree it judged.

- **Each group is filed under the tree the prompt named**: the run holds the
  candidate it read once, so a reviewer that edits the worktree cannot move the
  answer, split it across two trees or strand the lens that landed first.
- **A moved worktree stops the run**: the message names the file that keeps the
  answer, how many lenses landed, and the tree that exists now - those bytes ask
  for their own review. The paid answer stays as evidence of what was judged.
- **The project's own gate is on**: `.odachi/review.toml` says `on` locally, and
  CI keeps enforcing the fence and the proof with `--no-review`, because a
  verdict is evidence about one tree and lives in a store the tree ignores by
  design.

## v0.11.14 - 2026-10-25

A stopped review keeps what it already paid for.

- **Each model group's verdict is stored as it lands**: a reviewer that fails,
  times out or answers something that is not a verdict no longer throws away the
  lenses the earlier groups answered. The run says `saved 2 of 4 lenses` and names
  the ones still owed, with the command that asks for them.
- **`review --show` calls a half answer partial**: `review of <tree>: BLOCK
  (partial: 1 of 2 lenses)`, and the gate still blocks on the lens nobody judged.
  A partial verdict is not a review.
- **The store never replaces an answered lens**: a test pins the refusal, even
  when the second file also brings lenses the tree still owes.

## v0.11.13 - 2026-10-25

The review knows who judged it, and the human chooses who that is.

- **The verdict cannot come from the model that wrote the tree**: the proof
  records the model behind `--run`, the verdict records the one that judged, and
  `check --require-review` refuses the lenses the builder judged (`review_self`),
  naming the fix. It is disclosure, not identity: a model name can be lied about.
- **Each lens can name its model** in `.odachi/review.toml`: the review runs one
  reviewer per model group, so a cheap model carries the cheap lenses, and
  `review --show` prints the plan before anything is paid for. The verdict grows
  by lens — a second run adds what the tree still owes and never replaces an
  answer already stored.

## v0.11.12 - 2026-10-25

The release says which source built it.

- **Every archive carries a provenance attestation**: after the release is
  published, the workflow attests each archive named in the signed checksums
  with the repository, the commit and the tag that produced it.
  `gh attestation verify <file> --repo biggs-100/odachi` reads it without a key.
- **The chain has three links and the docs say what each proves**: the signature
  says who published, the attestation says which source built it, and the SBOM
  says what is inside. None of them says the source was good.
- **The attestation can be retried without republishing**: it runs after the
  release is public, so a failure there is fixed by running the workflow by hand
  on the tag; that run only attests.

## v0.11.11 - 2026-10-25

The release proves where it came from.

- **The checksums are signed without keys**: the tag workflow signs
  `checksums.txt` with cosign and the OIDC identity of the run, so there is no
  key to store, rotate or leak, and the release publishes one bundle
  (`checksums.txt.sigstore.json`) that holds the signature and its certificate.
  The installer still verifies `sha256` only.
- **Every archive ships an SBOM** (SPDX, by syft): what went into the file is
  published beside it.
- **The README and SECURITY say how to verify and what the chain does not
  prove**: the signature says who published, the SBOM says what is inside, and
  provenance is not published yet.

## v0.11.10 - 2026-10-25

The gate stops being ceremony: a red run says why, and a change cannot loosen the
policy it must obey.

- **A failed run keeps its whole output** in `proof/<feature>.failed.txt` —the
  tree, the command, the exit code and every line, including a timeout—, so the
  one-line verdict sends you to the evidence instead of replacing it. A green run
  removes the stale file, and a file the tool could not write is never named.
- **The fence only loosens with a new vision**: `check --fence-from` compares the
  non-goals of the base brief and every rule, path and dependency the base fence
  had. Dropping one blocks and the gate names it; adding is stricter and passes;
  a fence the tool cannot read blocks either way.
- **One workflow judges a tree**: `.github/workflows/gate.yml` runs on Linux,
  macOS and Windows, builds odachi from the base ref and uploads `proof/` even
  when it fails. `ci.yml` is gone, so no two files can disagree about the same
  tree.
- **Review pays for itself**: `review --show` answers whether this tree asks for a
  verdict and why, a documentation-only tree is not asked for ceremony, and the
  rule is one review, one correction, one re-review.
- **main only takes a gate that passed**: `.github/main-ruleset.json` requires the
  four checks, needs no approval from a second human and has no bypass list; the
  policy is a file in the repo, so a settings change can be reviewed.
- **A stranger can go from zero to a merged pull request**: CONTRIBUTING with the
  loop and the exact local command, issue forms, a PR template that asks for the
  proof, SECURITY and private vulnerability reporting, and a ruleset that does
  not demand a second person to merge.
- **The plan lands before the work**: `start` ends by telling the agent to commit
  the plan —brief, fence, profile and task doc with its steps—, and the drive
  skill makes that commit step two of the cycle. ADR 0028 records the three
  resets that paid for the rule.
- **The suite is deterministic**: no test reads state another test writes, the
  version line is pure, and the ldflags seam is pinned by reading the source
  instead of mutating the package.

## v0.11.9 - 2026-10-24

The gate can run in CI without trusting the change.

- **`check --fence-from <ref>`**: the fence is read from a git ref, not the
  worktree, so a change cannot loosen the policy it must obey. A missing file at
  a valid ref is no fence; a ref that does not exist blocks. The proof names the
  ref it used.
- **The CI recipe builds the gate from the base ref**
  (`.github/workflows/gate.yml`) and uploads `proof/` even when the gate fails.
- **`ODACHI_NO_SETUP=1` means binary only**: both installers skip `setup`, and
  `setup` and `init --global` refuse with the variable named, so a shared runner
  never gets a home door or a default agent.

## v0.11.8 — 2026-10-24

Installing odachi is enough.

- **`odachi setup` leaves the machine ready**: the bootstrap skill in
  `~/.agents/skills`, the home door for the agent clients found, and OpenCode's
  default agent. It runs from any folder and writes nothing in it.
- **The installer runs it**, so the user only installs and talks. The door and
  the bootstrap skill now say the agent runs `odachi init` and `odachi start`
  itself, asking permission before writing in a repository that is not the
  human's, instead of asking the human to type them.

## v0.11.7 — 2026-10-24

The global door is the default agent.

- **`odachi init --global` makes odachi OpenCode's default agent**: it writes
  `default_agent` in `~/.config/opencode/opencode.json` (or the `.jsonc` that
  already exists), creating the file with the published schema when there is
  none. A default the human picked before is named and left alone; a config that
  cannot be read safely is refused with the exact member to add by hand.
- **The edit never rewrites the file**: comments, key order and CRLF survive, and
  `doctor` notes when the door is installed but the config does not choose it.

## v0.11.6 — 2026-10-24

The gates read the worktree once.

- **`check` and `review` measure the candidate once**: the tree, the changed
  paths and the diff come from one git read instead of three, and the proof, the
  floor, the lens class, the prompt and the receipt reuse that value. A review
  drops from about 13 git processes to 6; `review --show` goes from 1029ms to
  351ms warm, and `check` from about 1009ms to 340ms.
- **No weaker gate and no new dependency**: the fence, the lenses and the
  coverage gate read the same bytes as before; a reviewer that edits the tree is
  still measured again before the verdict is filed. The suite wall is
  machine-bound: 23.8-24.6s in a quiet window, 25.6-36s under background load.
  The work removed is the honest signal; see ADR 0019.

## v0.11.5 — 2026-10-24

The reviewer inherits the session's model; review answers come from stdout, and
the proof keeps both streams.

- **`{model}` in `review --exec` is the model the session is using**: the
  reviewer judges with the same model you are talking to, so the verdict and the
  change come from the same eyes. The order is `--model`, then `ODACHI_MODEL`
  —the contract any host can set—, then the OpenCode session through
  `opencode session export`. A command without the placeholder runs exactly as
  before.
- **No fallback, and a flag that changes nothing is refused**: when `{model}` is
  asked for and no model resolves, the review stops before starting the reviewer
  and names `--model`; `--model` without `{model}`, or without `--exec`, is
  named instead of silently dropped. The receipt records the command that really
  ran, never the template.
- **`review --exec` reads the verdict from stdout**: a reviewer CLI's banner,
  warnings or ANSI on stderr no longer poisons the JSON answer. When the
  reviewer fails, the message shows stdout and stderr together.
- **The shell ports report the streams separately**: `check --run` still writes
  both into the proof — a failing test's stdout and its build errors are both
  evidence — while the review reads only the answer. This fixes a real failure
  found by dogfooding `opencode run`: the verdict was correct and the ANSI
  banner made it unparseable.

## v0.11.4 — 2026-10-23

The lens class can be raised by the diff's content: a concurrent or
credential-touching change to an ordinary file now asks for the four lenses.

- **`Assess` raises, never lowers**: the path class is the floor, and added
  lines that spawn concurrency (`go func`, `chan`, `sync.`, `atomic.`, `mutex`,
  `<-`) or touch secrets (`password`, `secret`, `token`, `api_key`,
  `private_key`) raise a focused candidate to full. A passive candidate (a
  document) is immune to prose, and no content lowers a class.
- **`review` and `check` share the assessment**: the prompt's lens list, the
  stored verdict's lenses, and the gate's per-file coverage demand all come from
  the same deterministic reading of the diff. The check port gained `Diff`.
- **Still deterministic and opt-in**: no model decides the class, and the
  human-owned review switch is untouched. See ADR 0017.

## v0.11.3 — 2026-10-23

Review receipts: a verdict now answers per file and per byte, and a corrected candidate pays only for what it changed.

- **Per-finding content receipts**: every finding and every clean pass carries
  the git blob id of the file it judged (path, blob, lens). The tool derives the
  blob from git; a reviewer-declared hash is overwritten. A finding whose file
  git cannot hash (deleted, or invented by the reviewer) is stored without a
  receipt, and the gate treats it as answering nothing.
- **Coverage is the gate's arithmetic**: with `--require-review`, every changed
  file whose bytes git can measure must answer each lens its path class
  requires — a finding, a clean pass, or a dismissal. A stale receipt, a
  missing row, or a verdict written before receipts existed blocks with exit 2
  naming `odachi review`; an approval the tool cannot prove is not used.
- **Potency is per blob**: storing a verdict inherits the standing rows of
  earlier verdicts whose blob still matches, so the fix loop re-reviews the two
  files you touched and inherits the rest. The prompt lists what is already
  carried, and a repeated finding is filed once.
- **Dismissals are first-class**: a finding refuted with evidence travels as a
  dismissed row with its reason and a fingerprint (path, line, lens,
  evidence). A re-raise of the same fingerprint stays dismissed; the verdict
  says what it discarded, and `--show` prints the reason. Dismissed rows answer
  coverage and never block.
- **A deterministic floor runs before any lens**: the diff names the added
  lines a finding may cite and the categories they touch (concurrency, secrets,
  network, failure, time). A causal BLOCKER or CRITICAL that cannot cite them
  is downgraded one tier before the verdict is stored — blocking is earned. The
  floor reports the path class; it does not change ADR 0010's lens selection.
- **Candidate diffs use the merge-base**: `Diff` and `ChangedPaths` compare
  against the fork point instead of the base tip, so another branch's commits
  no longer appear as deletions in the candidate. This is a behavior change for
  scripts that pinned the old two-dot diff.

## v0.11.2 · 2026-10-23

The version that adds a method for attacking a backlog, cutting delivery into reviewable PRs, writing Go tests, auditing skills, and speaking English in everything that is not chat.

- **English artifacts, docs and CLI**: the brief, task doc, `AGENTS.md` block,
  skills and CLI messages are English; the agent chat stays in the human's
  language. The old Spanish headers remain readable as aliases, so closed task
  docs and old briefs still parse, and a pre-commit hook installed before the
  migration is still odachi's: `init` backs it up and rewrites it in place.
  This is a breaking change for custom skills and scripts that pinned the old
  `Cuándo / Reglas / Nunca` sections or the Spanish CLI messages.
- **`triage`**: classifies each report by root cause, groups the ones that share a
  cause into one step, reproduces before fixing, and closes with the test that
  fails. It does not copy GitHub labels or the ODD grammar: it uses `odachi ask`,
  `odachi plan`, `tests` and `check --run`.
- **`pull-request`**: one work unit per PR, a 400-line budget, a single chain
  strategy (stacked or feature branch chain), an approved issue or `Refs`, and
  `odachi check --run` before review. The slug is `pull-request` because the skill
  format requires names of three characters or more.
- **`go-tests`**: complements `tests` with the Go pattern: a table with `t.Run`,
  `t.TempDir`, `testing.Short`, `teatest`, golden files, mocks at the edge and the
  focal package before the suite. It does not repeat the discipline of `tests`.
- **`skill-improve`**: audits a skill that is not used or grew. It starts with the
  trigger, requires observable rules, cuts at 60 lines, leaves one owner per rule,
  and sends you to `skills --check`/`--write`. The format is still validated by the
  binary.

## v0.11.1 · 2026-10-23

La versión que protege el contexto del agente principal y termina de limpiar el arranque.

- **`init --agents`**: `odachi init --agents opencode|claude|both` elige qué
  clientes reciben el agente nativo y la puerta del home. La elección queda en
  `.odachi/agents.toml`; `start`, `level` y `doctor` la respetan, y un `init`
  posterior sin flag la conserva. El default sigue siendo `both`, y los archivos
  del cliente que no elegís no se borran: sólo dejan de escribirse.
- **Subagentes nativos**: `init` escribe `.opencode/agents/odachi-worker.md` y
  `.opencode/agents/odachi-verify.md`, y sus equivalentes en `.claude/agents/`.
  El worker implementa un paso acotado; el verifier es read-only. `slices` manda
  a pasarles el task doc, el paso, las superficies de edición, los comandos y los
  paths exactos de las skills; vuelven con evidencia y ≤2k tokens. El padre
  conserva git, el gate y la visión.
- **`init --skills home`**: todas las skills en `~/.agents/skills`, sin copia en
  el proyecto. El default sigue siendo `project`. `doctor` cuenta y compara las
  dos ubicaciones, y no afirma que las skills ajenas del home coincidan.
- **`*.bak-*` ignorados**: `plan`, `ask --craft`, `start` y `level` dejan backups
  para `undo`; el bloque de `.gitignore` los mantiene fuera de git.
- **`start` sin TTY**: si el terminal muere en la primera pregunta, el mensaje
  nombra los flags que faltan en vez de "no pude leer la respuesta".
- **Review proporcional**: `check --require-review` ya no pide veredicto cuando
  el cambio es sólo documentación; el candidato pasivo no tiene lentes que
  responder. El clasificador de riesgo suma contratos públicos (`api`,
  `openapi`, `proto`, `graphql`, `sdk`), concurrencia (`lock`, `mutex`, `race`,
  `async`, `worker`, `queue`) y entrega (`.github`, `release`, `goreleaser`,
  `terraform`, `helm`, `k8s`, `pipeline`).
- **Modo de review del proyecto**: `odachi review mode status|on|off` escribe
  `.odachi/review.toml`; `check` lo lee, y `--require-review`/`--no-review` lo
  override por una corrida. Sin archivo, el default sigue siendo `off`.
- **Backups fuera de la vista**: las versiones que `undo` guarda antes de
  reescribir un artefacto viven en `.odachi/backups/`, con la ruta relativa del
  archivo. `undo` sigue leyendo los backups viejos al lado, y `undo --collect`
  los mueve con un rename que conserva bytes y fecha; la segunda corrida no hace
  nada. Los archivos fuera del proyecto conservan el backup al lado.
- **Timeout accionable**: un `check --run` que no termina sugiere duplicar
  `--timeout` en vez de repetir el mismo comando.
- **`init --global`**: la puerta del home es opt-in. `odachi init` ya no escribe
  `~/.config/opencode/agents/odachi.md` ni `~/.claude/output-styles/odachi.md`:
  instalar un agente `primary` en el selector de cualquier carpeta es una decisión
  del humano, y como efecto de `init` podía reemplazar el agente que la persona
  ya había elegido sin que lo pidiera. `odachi init --global` la instala para los
  clientes elegidos, y nada se borra: lo ya instalado sigue donde está. `doctor`
  nombra `--global` cuando la puerta falta o no coincide.
- **`doctor` ve la puerta global**: el chequeo 11 también compara
  `~/.config/opencode/agents/odachi.md` y `~/.claude/output-styles/odachi.md`.

## v0.11.0 · 2026-10-22

La voz vive en el bloque, el agente nativo la espeja, y el doctor ve lo que `init` dejó.

- **La voz vive en el bloque**: `AGENTS.md` termina con `## Cómo hablás`, y esa
  sección la escribe el binario. `odachi init` pone la voz de `novato` si todavía
  no hay perfil; `level --to <perfil>` cambia el perfil y la voz juntas, y `start`
  también, porque es el otro comando que decide el perfil.
- **El agente nativo es un espejo del bloque**: `init` escribe
  `.opencode/agents/odachi.md` y `.claude/output-styles/odachi.md` con las mismas
  reglas y la misma voz, para elegir odachi con Tab (OpenCode) o `/config` (Claude
  Code). Medido en OpenCode 2.1.3: la carpeta se lista como agente primario cuando
  el proyecto tiene `.git`; `mode/` también funciona, pero la documentada es
  `agents/`. `level --to` y `start` refrescan el espejo si ya existe; `doctor`
  avisa si dejó de coincidir. Las dos personas por perfil de la v0.10.0 quedan
  como restos: `init` las nombra y no las borra.
- **La puerta del home**: el mismo `init` deja `odachi` en
  `~/.config/opencode/agents/odachi.md` y `~/.claude/output-styles/odachi.md`, así
  que aparece en **cualquier** carpeta, no solo en la que corrió `init`. Es un
  despachador corto, no una copia de las reglas: si la carpeta tiene el bloque de
  `AGENTS.md`, lo lee y adopta su voz; si no, carga la skill `bootstrap` y guía
  para arrancar con `odachi init` y `odachi start`.
- **`doctor` pasa a once chequeos**: el 9 dice las dos verdades de las skills (el
  registro del proyecto y lo instalado en `.agents/skills`), y el 11 compara la voz
  del bloque con el perfil, y el agente nativo con el bloque.
- **Un proyecto nuevo arranca en `novato`**: `start` ya no infiere `senior` porque
  haya `.git`; `init` es justamente quien crea ese repo. Sin perfil escrito y sin
  brief, el default es `novato`. `senior` se gana con `odachi level --to`, o se
  infiere de un brief que la persona ya trajo.
- **La delegación tiene una política y un dueño**: `slices` gana sobre otras
  skills globales; `drive` manda a cargarla cuando el pedido no entra en una
  lectura; un escritor se delega solo por paralelismo real (unidades disjuntas) o
  por el context backstop, nunca por tamaño o cantidad de archivos. Se tomaron
  los triggers y el handoff acotado de gentle-ai, no su orquestador.
- **ODD-lite, sin cambiar el formato del task doc**: `drive` exige requisitos
  citados verbatim y un "Listo cuando" observable, `tests` con RED antes del
  código, el registro del paso delegado con `ask --craft` y la línea de riesgo de
  `commits`. No se adopta la gramática `Specs`/`Tasks`/`Log` de gentle-ai; la
  decisión está en el ADR 0008.
- **El instalador no depende de `Get-FileHash`**: si el cmdlet no existe —pasa en
  Windows PowerShell 5.1—, verifica el sha256 con `certutil`. Probado con las dos
  herramientas sobre la release real.

## v0.10.0 · 2026-10-22

Listo en un comando, y el agente sabe manejarlo.

- **`odachi init` deja el proyecto listo de una vez**: historial de git si faltaba,
  bloque de reglas en `AGENTS.md`, bloque de `.gitignore`, hook de pre-commit, las
  skills del proyecto y las dos personas. Ya no son dos comandos.
- **Skills en la ruta estándar**: `.agents/skills` en el proyecto y `~/.agents/skills`
  en el home, la convención cross-client que recomienda el estándar Agent Skills.
  `pack` gana `--scope project|home`, `--dest <carpeta>` y el pseudo-agente `standard`.
- **Personas generadas**: `novato` y `senior` en `.opencode/agents/` (se eligen con
  Tab) y en `.claude/output-styles/` (desde `/config`). La ceremonia la sigue
  decidiendo `.odachi/profile.toml`; la persona decide solo cómo habla el agente.
- **Skill `drive`**: el manual que el agente carga para manejar la herramienta —los
  verbos, el ciclo de una feature, qué significa cada exit— y la línea del bloque
  que lo obliga a cargarla.
- **Skill `bootstrap`** en el home: enseña a arrancar un proyecto donde todavía no
  hay nada, que es justo cuando no hay otro archivo que leer.
- **`doctor` con dos formas**: en una carpeta que no es un proyecto responde una
  situación y un paso (`init`, después `start`); en el resto, los diez chequeos.
- **El bloque de `AGENTS.md` dice qué es odachi**, quién decide qué y dónde está el
  contexto; y quedó fijado byte a byte por un test.
- `pack --json` cambia de forma: las filas son por destino e incluyen `agents: []`.
- `doctor --json` gana `lines`.

## v0.9.1 · 2026-10-21

- La línea de cierre de `init` deja de asumir que el humano commitea: el que guarda
  el avance es el agente.

## v0.9.0 · 2026-10-21

- **`init` crea el historial de git** cuando falta, y lo explica en castellano llano.
  Antes se negaba con «corré git init» a alguien que no sabe qué es git.
- Sin git en la máquina sigue bloqueando, con una frase que dice qué instalar.

## v0.8.0 · 2026-10-21

- Tres skills nuevas para cerrar el currículum del que construye: `commits`,
  `writing` y `skill-craft`. Ocho skills, todas bajo el cap de 60 líneas.
- El registro de skills pasa a preguntarse al payload y al directorio: cambiar el
  contenido ya no obliga a editar un test.

## v0.7.1 · 2026-10-21

- La regla de delegación queda con presupuesto medible (~3 lecturas o ~10k tokens)
  y retorno acotado (≤2k tokens, evidencia `path:line`), y el trigger se amplía.

## v0.7.0 · 2026-10-21

- La mitad de delegación de un orquestador vive en la skill que ya parte pedidos:
  delegar por razón, pasar el task doc por path, verificar contra el brief y no
  contra el resumen, y no delegar nunca la decisión de visión.

## v0.6.1 · 2026-10-21

- El doctor deja de mentir en dos casos medidos: un worktree (`.git` es un archivo)
  ya no se reporta como «no hay repositorio», y un perfil inventado sin `extends`
  ya no valida en verde.

## v0.6.0 · 2026-10-21

- **`doctor`**: diez chequeos de salud, de solo lectura, cada uno con el comando que
  lo arregla.
- **`bench`**: ocho recorridos que clasifican cada bloqueo (in-band, out-of-band,
  templated, interactive, dead end) con ratchet: un dead end hace fallar la corrida.

## v0.5.0 · 2026-10-21

- **Las cuatro lentes 4R** con selección por evidencia: sin lentes para documentos,
  `reliability` para lógica, las cuatro cuando el cambio roza algo que puede lastimar.
- **Refutador**: un `BLOCKER` causal solo se destraba citando la línea del diff que
  lo desmiente. Lo que no se puede citar, no se refuta.

## v0.4.0 · 2026-10-21

- **`review`**: veredicto de diseño por árbol, content-addressed en
  `.odachi/reviews/<hash>.json`. La aprobación no puede flotar porque el nombre del
  archivo *es* el estado del código.
- `check --require-review` opcional: el review como compuerta.

## v0.3.1 · 2026-10-21

- `init` avisa cuando `odachi` no resuelve desde un shell: una regla que el agente no
  puede ejecutar es peor que ninguna.

## v0.3.0 · 2026-10-21

- **OpenCode** como destino de las skills, con destinos por alcance (proyecto vs home)
  y la lista de agentes soportados.
- `undo` respalda lo que reemplaza y se niega a elegir por vos cuando hay más de un
  candidato; `plan` agrega y cierra pasos sin editar Markdown a mano.

## v0.2.0 · 2026-10-21

- **`pack`**: las skills llegan al directorio nativo de cada agente, sin borrar nada,
  con respaldo de lo que difiere.
- Registro generado de skills, validador y creador.

## v0.1.0 · 2026-10-21

- Primera versión que se instala: `start`, `ask`, `level`, `summary`, `check`, `init`,
  `--version`, hook de pre-commit, instaladores `sh` y `ps1`, y seis binarios.
