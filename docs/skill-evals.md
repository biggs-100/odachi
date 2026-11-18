# The loop that measures a skill

The skills are the product: text an agent reads before it works. `embed_test.go`
proves the text still carries certain sentences; it does not prove that an agent,
reading them, works differently. That question was asked once, by hand, and
`docs/skill-evaluation.md` records nine probes and what each one showed. What it
does not have is a second run: a skill that gets worse today looks exactly like
one that never moved, because repeating the measurement meant repeating the nine
probes and moving the home copy of the skill out and back.

This is the loop: one command, two arms, a score, and a record.

## What it runs

```
odachi skills eval tests --exec "<the command that runs your agent>" --runs 3
```

Without `--exec` the verb prints the prompt and runs nobody. With it, each draw
builds a project from the exam that ships with this binary -a task, a small Go
module, a task document- and runs the command in it. `{prompt}` in the command is
replaced by the path of the prompt file.

The two arms get the same prompt:

| arm | what it has |
| --- | --- |
| `with` | the skill's files in the project's `.agents/skills/<name>`, where the client discovers them |
| `without` | the same project, and no such folder |

The arm without the skill is the folder absent, not a stub: the client reads
`~/.agents/skills` and its configuration only adds paths, so there is no switch to
leave a folder out. That is why the run needs `--move-home`: it takes the home copy
aside for the length of the run and puts it back before it answers. The move writes
a journal in the clone's store before it happens, so a run that dies leaves a mark
the next one reads; `--recover` puts back what a dead run left out, and measures
nothing.

## What it decides

The event stream the client writes names every tool call in order. From that order
the exam decides two of the rubric's four behaviours:

1. **the test was written before the code** - the first write to a `_test.go`
   before the first write to a `.go`;
2. **the suite ran between the two** - a `go test` command after the test and
   before the code, which is the red a person observes.

Each behaviour carries the event that decided it, so a reader can check the score
without trusting the number. The other two -whether the test's name says what
breaks, and whether the fix is at the cause- need a reader, and the exam hands the
reader the transcript instead of guessing.

## What it does not

- **It does not run in CI or in the gate.** N draws per arm times two arms is a
  paid measurement, and the fence keeps a model call out of `.github/`.
- **It does not judge the two behaviours that need judgment.** They are printed and
  kept; a person reads them.
- **It does not measure the description of a skill or the `AGENTS.md` block.** The
  nine probes gave the description no support on their corpus.
- **It measures one skill.** `tests` is the exam that ships; a second one is the
  question the form has to answer next.
- **It does not rewrite the old page.** `docs/skill-evaluation.md` is frozen: it is
  the record of the nine probes, and this is the instrument that repeats them.

## The record

A run files what it measured in the clone's store, `.git/odachi/evals/`, under the
identity of the bytes: the skill's files by path and the task they are asked about.
The same text over the same exam is refused instead of paid again; `--again` is the
human saying the second run is worth it. A skill that changes answers a different
identity, and that is what makes the record a comparison instead of a repetition.

`odachi clean` knows the class and keeps the newest ten records.

## What this page does not carry yet

A number. The loop has not been run against a skill for a record: the first paid
run will leave the first one, and this page will quote it with the model, the draws
and the transcript's decisive events -the way the reviewer's page quotes its own.

Two things the loop found while it was built, and both are pinned by tests:

- **A nested `go.mod` takes its whole folder out of the embed, silently.** The
  exam's project ships its module as `go.mod.exam`; the run writes it back under
  the name the project needs, and a test reads the tree the binary carries.
- **A copied home does not authenticate.** Without its database the client answers
  401; with the database linked it keeps the real state, so the copy isolates
  nothing. That is why the control arm takes the folder out, with permission,
  instead of copying the home.
