# 0049 - The cleanup policy: disposable at write, named at close, never bounded

Status: accepted - Date: 2026-11 - Session: after v0.14.5

## Context

`odachi clean` was written with an owner per class and a lifetime for the classes odachi wrote
(0132). It reads the store, says what each class costs, and removes only with `--apply`; `proof/`
and `odd/tasks` carry an owner and the words "never touched". That answer was honest and quiet: the
store only spoke when somebody remembered to ask, and the class that grew fastest -the prompt a
reviewer was paid to read, 60 to 120 KB each- was kept while `keep_prompts = 0` already said it was
disposable. A clone measured on 2026-11-03 held 34.9 MB in 1352 files, 34.0 MB of it recoverable.

The comparison with gentle-ai and gentle-shell raised the question this ADR answers: both prune at
write time, and neither asks a human to run a command for the files it already knows are
disposable. Odachi's other rule is older and stricter: evidence is never deleted by a verb that is
not the human's hand (0150, the proof is immutable).

Two kinds of file were being treated as one. A verdict, a proof and the plan are **evidence**: a
different reader must be able to check a past decision against the tree it names. A prompt, a
refused answer and a backup are **scaffolding**: they exist for the run that wrote them, and no
reader of the record needs them afterwards.

## Decision

- **Disposable means nobody reads it, not "a machine can rebuild it byte for byte".** A prompt is
  the question the run asked, and a prompt rebuilt later is a different question: it carries the
  rows the store held at that moment. It is disposable because the verdict is the answer that was
  paid for, and the gate reads the answer.
- **The writer obeys its own policy, now.** A stored verdict removes the prompts its
  `keep_prompts` stopped keeping, and the answer scaffolding kept for that tree. The rule is one
  domain answer, the same one `odachi clean` reports, so the verb and the write cannot disagree
  about which trees survive. A review that gets no verdict removes nothing, and a store that cannot
  be read prunes nothing and says so: a leftover costs a file, a guess costs a paid answer.
- **The close names the growth; the human keeps the switch.** The tree that closes a feature prints
  one line with the store's size and what `odachi clean --apply` would free, or that the switch is
  blocked by the paths it could not read. The line deletes nothing and changes no answer or exit
  code. A close is where a person decides what comes next, and it is the only place this number is
  worth a line.
- **Odachi never bounds the evidence.** `proof/` and `odd/tasks` keep their owner and their "never
  touched" action. Bounding them would be a decision about which history is forgotten, and that
  belongs to the human who owns the record -not to a lifetime constant shipped in a binary. A
  project that one day wants it writes that decision down and names the command, the way every
  other irreversible act is named.
- **A new class earns a lifetime by answering what reads it.** Before a file class enters
  `--apply`'s reach, the change says who reads it and what is lost when it goes. A class that
  cannot answer stays behind an owner line.

## Consequences

- A clone's store stops growing with every review. A project that wants the questions back writes
  `keep_prompts` and gets them: the number is the policy, not the verb.
- The close shows the cost where the human already reads, and CI logs it. The measurement runs once
  per feature, not on every commit: an open tree does not ask.
- A store with unreadable paths prints "cannot free anything" instead of a number `--apply` would
  refuse to act on: the line is a measurement, and a failed read is not an empty store.
- The line between disposable and evidence is now a test and a sentence. If a future class is
  ambiguous, the question to ask is the one this ADR asks: who reads it after the run that wrote it?
