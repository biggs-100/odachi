# el-requisito-con-guion

## What must exist
- un requisito que empieza con guion se puede pasar: `start --light -- "--runs 0 …"` lo escribe tal cual, y sin el `--` el error dice cómo pasarlo
- the refusal that a step cannot start with a dash becomes the narrower rule it meant: a markdown bullet (`- `) doubles the checkbox, and `--runs` duplicates nothing
- the message names the form instead of the flag package's complaint, because "invalid flag" does not tell a human what to type

## Done when
- `odachi start --light -- "--runs 0 se rechaza"` writes a task document whose requirement and step carry that sentence verbatim
- `odachi start --light "--runs 0 se rechaza"` fails with a message that names `odachi start --light -- "<the line>"`
- `odachi plan "- [ ] T1 · algo"` is still refused, and the message says a step does not start with a bullet
- a step text that starts with a single dash is accepted, because the hazard is the bullet and not the dash
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0270)

## Not included (first version)
- the verb keeps its shape: `--light` still takes one quoted line, and `--` stays Go's own separator instead of a new flag - back when the flag set itself is the question
- a step that starts with a bullet and a space is never accepted: that is the one that doubles the checkbox - back when the renderer stops writing the bullet
- nothing about the other verbs changes: `plan` and `ask` keep their messages, and only `start --light` learns the form - back when another verb takes a free line

## Decisions taken
- The hazard is the bullet, not the dash: `- [ ] T1` pasted whole is what breaks the file, and `--runs 0` protects nothing.
- The error teaches the form it wants: a human reads the message, not the flag package's vocabulary.

## 0271 · el-brief-acumula

## What must exist
- el brief deja de reemplazarse: cada start agrega la seccion de su feature, y lo que estaba escrito antes queda
- el parser lee todas las secciones, asi que outcomes, done-when, non-goals y decisiones pasan a ser la union
- la lista de non-goals con la que se arma el fence es la union, igual que las reglas que ya se acumulan

## Done when
- after a second `start` with a new requirement, `brief.md` holds the first block and the second, each under its own `## NNNN · slug` header
- `brief.Parse` returns both outcomes and both non-goals, and the gate still calls the brief complete
- the fence the second run writes still carries the first feature's rule with its sentence
- running the same requirement twice does not append a second block
- the first run of a project writes the file as it does today: a `# name` and one block
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0271)

## Not included (first version)
- the shape of a block does not change: it repeats the canonical headers, which is what the parser already reads - back when the block shape itself is the question
- bullets are not merged inside a section: a repeated outcome is ignored by its exact text, never folded into a similar one - back when a human wants two spellings unified
- the past is never rewritten: a block already written is not corrected, reordered or trimmed - back when the vision itself has to be edited

## Decisions taken
- A block repeats the canonical headers instead of inventing a nested grammar: `brief.Parse` already appends every section it finds, and a second grammar would be a parser to keep in step forever.
- A feature is named by its number and slug, the same pair its task document carries, so the brief and the plan are easy to read side by side.

## 0272 · el-prefijo-que-resuelve

## What must exist
- `--since` acepta el árbol que un humano copia: un prefijo del veredicto guardado, y resuelve al árbol completo antes de medir el delta
- the refusal tells the two cases apart: nothing in the store starts with that prefix, or several verdicts do and more characters are needed
- the full forty-character tree keeps working exactly as it does today

## Done when
- `--since <prefix>` of a single stored verdict measures the delta from its full tree, and the run says which tree it completed
- a prefix that matches two verdicts is refused, naming both and asking for more characters
- a prefix that matches none is refused with a message that no stored verdict starts with it, and names `odachi review --all`
- a full tree keeps working without a resolver in the way
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0272)

## Not included (first version)
- no other verb learns the prefix: `--show` and `--attach` keep asking for the exact name or a path - back when they need the same courtesy
- an ambiguous prefix never picks one: it asks for more characters instead of guessing - back when the store keeps aliases
- nothing is suggested or ordered like a search: the correction lane names the candidates it found, it does not rank them - back when finding a tree becomes a question of its own

## Decisions taken
- The prefix is read against the store's own listing, which is what `--all` prints: one place decides what a tree is called, and the correction lane does not grow a second index.
- The message says which tree the correction completed, because the whole point of the lane is that the human does not have to remember a forty-character hash.

## 0273 · la-cara-dice-la-verdad

## What must exist
- atacá todo eso: que el help liste todos los verbos, que el README no afirme lo que el repo desmiente (provenance, surfaces, packages, install.ps1), y los topics y la descripción del repo
- the usage block cannot drift from the dispatcher again: a test reads the verbs the binary answers and finds each one where a person looks
- the README stops measuring what it cannot measure: a count in prose is a claim nobody repeats, and the command that answers it is one line away

## Done when
- every verb the dispatcher handles appears in `odachi --help`, and a test fails naming the one that is missing (observed red with `design` today)
- `review probe`, `review mode` and `review waive` are visible where a person looking for review finds them
- the README no longer claims a package count, a surface count, that the release has no provenance, or that `install.ps1` has no automated test; each place names the command or keeps no number
- the README opens with a short table of contents that links the sections it has
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0273)

## Not included (first version)
- no new verb and no new flag: the change is what the binary says about itself - back when a surface is actually missing
- the prose of the README beyond the claims that are false or rot: no rewrite of the tutorial - back when the story itself has to change
- the repository's metadata (topics and description) is set with `gh repo edit`, outside the gate, because it is not a file in the tree - back when the tree has to carry it

## Decisions taken
- The test reads the dispatcher's own cases instead of a third list: a hand-kept list of verbs is the same drift with one more place to forget.
- The README loses the numbers instead of updating them: "62 packages" is true today and a lie in a month, and the project already says a number in prose is a claim nobody measures.

## 0274 · lo-que-la-compuerta-ve

## What must exist
- un artefacto que no decodifica como UTF-8, o que trae las secuencias de un UTF-8 leído como CP850, se rechaza en la compuerta nombrando el archivo, para que la doble codificación se vea sin modelo y sin review
- the gate compares what the brief says with the requirement its task document carries: two copies of one sentence that no longer match is a record that was mangled, and no heuristic is needed to see it
- a file saved with a BOM is refused too, because its first line stops being the line the parser reads

## Done when
- a `brief.md` whose bytes are not UTF-8 blocks the gate, naming the file and how to save it
- a `brief.md` with a BOM blocks, naming the BOM
- a clean tree stays ready, so the guards do not tax every run
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0274)

## Not included (first version)
- nothing is repaired: the gate reads and refuses, it never writes the good file back - back when a repair command is the question
- nothing is guessed: the gate names the file and the sequence it found, it never decides which of two versions was meant - back when the tool has to choose
- the task documents are never scanned for the markers: they are prose, and one of them may have to name the sequence to explain this defect - back when every artifact may be scanned

## Decisions taken
- The marker list is the rule and the task documents are the exception: the brief and the TOMLs are rendered by the tool, so a sequence there is a double encoding, and prose is the one place that may name it.
- The gate refuses and names; the repair is a human editor or `odachi start`, never a silent rewrite of the record.
- A cross-check against the task documents was tried and dropped: it blocked a brief before its first document and a block whose document was removed, and the record accumulates.

## 0275 · lo-que-una-tirada-no-ve

## What must exist
- a verdict can be paid by N draws of the same judge: `odachi review --runs N` asks each required lens N times and files the union of what they answered as one verdict
- the verdict says how many draws paid for each lens, so "approved" stops meaning "approved once" without saying so
- a lens already paid by one draw can be brought to N by a later run, paying only the draws that are missing: the union grows, it never shrinks
- two draws that name the same defect -same path, blob, lens, line and severity- pay one row, so repetition does not become noise
- a draw that does not answer leaves the draws that did: the verdict's count says what answered, and the run names the draw that failed and the command that pays the rest

## Done when
- `odachi review --runs 3` with `--exec` asks each required lens three times and writes `<tree>.json` whose `draws` says 3 for every lens it asked
- a second `odachi review --runs 3` over the same tree pays only the two missing draws per lens and the file says 3; a run that asks for no new draw is refused, naming the count the lens already has
- two answers naming one defect at one place under one lens pay one row, and the row kept is the first draw's words
- a run whose third draw does not answer stores the two that did, says `2 of 3`, names the draw that failed, and exits blocked with the command that pays the third
- `--runs 1` behaves as today and records one draw per lens; a stored verdict with no `draws` reads as one draw per lens, so every verdict already in a store stays valid
- `--runs 11` is refused naming the ceiling of ten before a store or a reviewer is touched
- `--show` prints the draws beside the lenses, and `## Unreleased` names it, patch (0275)
- `go test ./...` is green and `odachi check --run "go test ./..."` is green

## Not included (first version)
- no project default and no tier table: the count is typed on the run that pays for it, because a default spends money nobody asked to spend - back when a project asks for a floor by tier and the gate can name it
- the gate does not refuse a one-draw verdict and `--require-review` keeps its meaning: the record says what the verdict cost and the human weighs it (0043, 0048) - back when a project asks for `--min-draws` and accepts the refusal
- a draw is the same judge a second time, not a second model: independence is a different instrument, and two lenses answered by the same weights are one opinion in four costumes (0048) - back when heterogeneous draws are designed and measured
- no per-row count of the draws that saw it: a row found by one draw of three stands as the row it is, because silence is not exculpation (0033) - back when a row's instability becomes something the review can measure
- `review probe` and its fixture are untouched: the probe measures the lenses against a key, and this changes what a verdict about a tree costs - back when the probe's clean arm earns its name

## Decisions taken
- The count lives per lens in the verdict, because a lens is the unit that answers, and one draw can fail alone.
- An absent `draws` means one draw: the field arrives as a schema bump (current 4 to 5), and every verdict already written keeps its old meaning instead of being reread as a claim it never made.
- Two draws that find one defect pay one row: `findingKey` is already the identity a defect is filed under, and repetition must not become a second finding.
- A later run brings a lens to N instead of re-answering it: the union can only add findings, so re-drawing can never turn a changes-requested into an approved, and nothing already paid is paid twice.
- A draw that fails does not throw away the draws that answered: the count is written truthfully and the exit is blocked, because the verb did not do what it was asked even though the store kept what it paid for.
- N draws buy variance, not independence: three draws of a judge that always answers the same sentence are one draw paid three times, and the count says how much was paid, never how much was seen.

## 0276 · el-conteo-cuenta-respuestas

## What must exist
- a candidate whose prompt does not fit and is reviewed in parts records one draw per answer, not one per part: `--runs 1` over a split candidate leaves `draws` at 1, and `--runs 3` at 3
- the command a failed or a complete run hands over can be run as printed: it carries the placeholder the flag needs, so pasting it pays what the run says it pays
- the refusal that says a lens already holds enough names one more than the file counts, not one more than the refused run asked for

## Done when
- a split candidate, forced with a small `max_prompt_kb`, records `draws` of 1 for a run of one and 3 for a run of three
- the next line of a failed draw is a command `odachi review` accepts: it names `--exec` with a placeholder and the count, and it keeps the lens the run was told to judge
- a tree whose lens holds 5 draws, asked for 3, is refused with `--runs 6` in the next line
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- `## Unreleased` names it, patch (0276)

## Not included (first version)
- no change to what a split part may cite, or to how the parts of one answer merge their rows: the count is the only thing that is wrong - back when the parts themselves have to change
- the handover never guesses the human's reviewer command and never remembers it: it carries a placeholder, because which model judges is the human's decision at the keyboard - back when odachi stores the command it ran
- no new flag and no new verb: the fix is what the two existing lines say and what one counter counts - back when a surface is actually missing

## Decisions taken
- A draw is one answer to one lens over the whole candidate: the parts of one answer are one round, so the count is per answer and never per call.
- The hint carries `--exec "<your command>"` because the flag refuses without it, and a hint a person cannot run is not a hint.
- The number a refusal names is the one the file holds plus one: the count is the fact, and the run that was refused is not.
