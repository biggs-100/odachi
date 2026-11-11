# el-requisito-con-guion

## What must exist
- un requisito que empieza con guion se puede pasar: `start --light -- "--runs 0 ÔÇª"` lo escribe tal cual, y sin el `--` el error dice c├│mo pasarlo
- the refusal that a step cannot start with a dash becomes the narrower rule it meant: a markdown bullet (`- `) doubles the checkbox, and `--runs` duplicates nothing
- the message names the form instead of the flag package's complaint, because "invalid flag" does not tell a human what to type

## Done when
- `odachi start --light -- "--runs 0 se rechaza"` writes a task document whose requirement and step carry that sentence verbatim
- `odachi start --light "--runs 0 se rechaza"` fails with a message that names `odachi start --light -- "<the line>"`
- `odachi plan "- [ ] T1 ┬À algo"` is still refused, and the message says a step does not start with a bullet
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
