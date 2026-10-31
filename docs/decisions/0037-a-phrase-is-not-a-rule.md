# 0037 - A phrase is not a rule

Status: accepted · Date: 2026-10 · Session: after v0.11.72

## Context

`odachi start` writes the fence from the brief's "NOT included" list, and it wrote
one rule per non-goal. A non-goal that arrives as a preset carries the catalog's real
paths and dependencies. One that arrives as free text -"datos de pacientes", "no
tocar la base de datos"- carried nothing, so the run invented a glob out of the
phrase's own words: `paths = ["datos-de-pacientes*"]`.

That glob matches a file named after the phrase, which no repository has. The rule
protects nothing, and `check` will never refuse anything because of it. The file the
project trusts as its policy carried a promise that could not be kept, and the run
said so out loud -the message named the invented glob and asked the human to write
the paths- while the file kept the lie.

0120 wrote the rule the code was breaking: a fence rule protects a file, never a
reason. The header of `odd/fence.toml` says it in the fence's own words: *a rule
names paths, not a phrase: a glob of the phrase's own words matches files named like
the phrase and protects nothing.* The advice shipped; the behaviour did not.

## Decision

- **A rule is written only when somebody gave it paths.** A preset carries them from
  the catalog, so it becomes a rule. A phrase carries none, so it becomes nothing
  the fence obeys.
- **The phrase is carried as a template.** In its place the fence holds a commented
  block: the human's words, `# paths = []`, and the line that says to fill the paths
  and remove the `#`. The question survives in the artifact the human edits; it just
  stops pretending to be an answer.
- **The run says what it did.** One line per phrase, naming the file and the
  template, instead of claiming the fence wrote a glob. The line that teaches how to
  prove a rule works stays, because it is true once the paths exist.
- **The empty case does not change.** A project with no non-goals still gets the
  file that explains what a rule is and shows one, commented.

## Alternatives rejected

- **Refusing the run until every non-goal has paths.** The interview would demand
  knowledge the human does not have at the start: the paths of code that does not
  exist yet. The fence is a decision recorded now and completed later.
- **Inventing a better glob.** `**/pagos/**`, `*pago*`: a guess that still protects
  nothing and looks even more like a rule.
- **Writing `paths = []` as a real rule.** A rule with no paths violates the shape
  `fence.Parse` accepts, and a validator that has to special-case an empty rule is a
  hole in the gate.
- **Dropping the phrase from the fence entirely.** The brief keeps it, and the human
  would have to reread the brief to know what they excluded. The template puts the
  question where the answer goes.

## Consequences

- `odd/fence.toml` can hold zero rules after a `start` that only heard phrases, and
  that is honest: the file says what is still a question.
- A repository that adopted odachi in an older version keeps the glob rules it
  already has; this feature changes what `start` writes, and never rewrites a fence
  that exists without a human asking.
- The next preset added to the catalog is one row and keeps working, because a rule's
  paths come from the catalog and never from a phrase.
- 0120's rule now has an enforcer: a fence rule without paths cannot be produced by
  the verb that writes fences.
