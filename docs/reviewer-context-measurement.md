# The reviewer's context, measured

Written 2026-11-17, closing 0323. The decision this page measures is recorded in
`docs/reviewer-context-research.md`: instead of a diff in the prompt and a reviewer
told to stay quiet, the reviewer reads a frozen copy of the candidate with bounded
tools. That page answered *what* to build. It could not answer whether the change
buys anything here, and said so. This is the answer, and the two runs that had to be
thrown away before it existed.

## The instrument

`odachi review probe` asks the four shipping lens prompts about a fixture and
scores the answers against a written key: found, called low, missed, and the
findings nobody planted, split by what they claim -a claim of a defect is noise, a
suggestion is a note- on both arms.

Two things changed in it for this measurement:

- **The fixture has a base.** `base/auth.go` is the wiring both arms land on: it
  is committed before the candidate, so it is in neither diff. The key plants a
  defect the diff cannot show -the handler reads a context key the middleware
  never writes, so every request answers 401- and only a reviewer that reads
  `auth.go` can see it. Six defects, one of them of that class.
- **A reader is confined by the client, not by its prompt.** Each arm runs in a
  directory odachi owns, with the candidate inside it and an `opencode.json` beside
  it that denies every path outside. The prompt still says "read only this
  directory", because a reviewer that knows the boundary does not waste a call
  probing it; the rule is what enforces one.

The two arms are the question:

| Arm | What the reviewer gets | What it can read |
| --- | --- | --- |
| attached | the prompt, with the frozen candidate named | that candidate, and nothing outside it |
| blind | the same prompt, with no candidate named | an empty directory, and nothing outside it |

The blind arm is a control, not a re-enactment: it is today's prompt minus the
snapshot's paragraph, not the prompt this repository shipped before 0320, which
banned reads outright.

Both arms ran with the same binary, `opencode-go/deepseek-v4.1-flash`, three
draws each, on the tree that confined the reader. The question each answer was
asked is stored beside it in the clone's store.

## The reading

Three draws per arm, on the tree that confined the reader. One row per defect the
key plants, and how many of the three draws named it, at or above its floor:

| defect | lens | where | attached | blind |
| --- | --- | --- | --- | --- |
| `risk-query-built-by-concatenation` | risk | `server.go:37` | 3/3 | 2/3 |
| `reliability-context-key-nobody-writes` | reliability | `server.go:24` | **3/3** | **0/3** |
| `resilience-retry-with-no-cap` | resilience | `cache.go:32` | 3/3 | 3/3 |
| `resilience-read-error-discarded` | resilience | `cache.go:37` | 3/3 | 3/3 |
| `readability-name-does-three-jobs` | readability | `limits.go:11` | 3/3 | 1/3 |
| `reliability-test-copies-the-formula` | reliability | `limits_test.go:8` | 3/3 | 2/3 |

Eighteen defect-draws the arm that could read its candidate answered every time.
The arm without one answered eleven, and the defect the base hides it never named.
It pointed at that line once, and the finding was about missing tests for the
handler, not about the key: the scorer counts a hit by its distance to the line,
and this is the first thing the table cannot tell.

What the tool printed, verbatim, is a per-lens row. It counts a defect once, so
two defects four lines apart can compete for one finding:

```
attached  runs 3 · risk 1/1, readability 1/1, reliability 2/2, resilience 1/2
          · seeded noise 3 (2..5) · seeded notes 0 · clean noise 0 (0..1) · clean notes 0 (0..1)
blind     runs 3 · risk 1/1, readability 0/1, reliability 1/2, resilience 2/2
          · seeded noise 5 (3..6) · seeded notes 0 (0..1) · clean noise 2 (1..2) · clean notes 1
```

The arm with the candidate was quieter on the arm with nothing planted: no claim
in the median, one in a single draw of three, and one suggestion in one. The arm
without it made two claims in two draws of three and one in the third -a median of
two- and a suggestion in every draw. Neither is an invented defect: one names the
missing test for the 4xx branch in `cache.go`, the other a backoff that sleeps
after the last attempt and ignores cancellation. They are improvements a person
would consider, reported at WARNING, and the arm that could read the code did not
raise them.

The canary said what it has to say: `canary: attached and scoped` for the arm with
the candidate, `canary: not attached` for the one without.


## What it says

1. **The change bought the defect the diff cannot show.** The arm with the
   candidate named it in three draws of three, citing `auth.go` lines it could
   read; the arm without one named it in none. This is the case 0320 was argued
   for, and it is the one the field can see: a handler that reads a key nobody
   writes looks correct in a diff and is a dead endpoint in the tree.
2. **It bought quiet on code with nothing planted.** No claim in the median from
   the arm that could read -one in a single draw of three- against a median of two
   from the arm that could not. Both of the blind arm's were improvements rather
   than inventions, which is the distinction `0282` added and the reason the number
   is reported per draw.
3. **On the defects inside the diff it changed little.** Fifteen in-diff
   defect-draws: fifteen for the arm that could read, eleven for the one that could
   not. The gap is in the softer defects -the name that does three jobs, the id
   concatenated into the query- where a reader that checks a caller is more
   careful, not more informed.
4. **The price is the call.** The arm with the candidate spent the same eight
   answers per draw; what it spent more of was the model's time, reading files the
   blind arm never opened.

## The boundary, and the run that proved it

The confinement is one rule the client enforces, not a sentence the model obeys.
Two recorded runs, on a copy of the shape odachi builds -a directory with the
candidate inside it and the rule beside it:

```
Read C:/.../candidate/main.go        →  Read file ..., lines 1-3  (the candidate)
Read C:/.../secret/marker.txt        →  Permission denied: external_directory
Glob "**/marker.txt"                 →  0 matches
Grep "SECRET-MARKER-OUTSIDE"         →  1 match, inside the candidate: the prompt file
```

Without that rule, the same tools searched the whole machine: `Glob "**/auth.go"
in C:/Users/USER/AppData/Local/Temp · 4 matches`. That is the difference between a
page that measures a reader and one that measures the filesystem.
## The runs that do not count

**The first run measured one arm.** The probe composed the review prompt without
the frozen candidate: it never passed the directory, so the reviewer received the
diff and no tree. That is the arm *before* 0320, measured by accident, and it is
why the first number the note carried said nothing about the change.

**The second run proved the arms were not separated.** The blind arm answered with
a CRITICAL at `server.go:24` citing `auth.go` lines it was never given, and the
tool log showed how:

```
Glob "**/auth.go" in C:/Users/USER/AppData/Local/Temp · 4 matches
```

The reviewer ran in an empty directory and its tools searched the whole machine
from there. A prompt that says "read only this tree" was the only boundary there
was, and a model that wants a file argues with a sentence. Both runs were void,
and the store keeps their answers as the evidence for why.

## What it does not show

- **One model, one fixture, three draws.** Six defects in about 160 lines is not a
  population, and one vendor's habits are not the lenses'. `--runs 5` and a second
  family are what would say more.
- **The fixture changed with this feature.** The base and the context defect were
  added to answer 0320's question, so this number is not comparable with the one
  `docs/review-research.md` carries: it is a different exam.
- **A lucky guess scores.** The key counts a hit at a line, and a reviewer that
  reasons its way to "the middleware probably uses another key" earns the same
  mark as one that read the file. The `evidence` field separates a citation from
  an argument, and only the arm that can read `auth.go` can cite it; the page
  reports which one it was.
- **The scorer counts a line, not a claim.** The blind arm's only finding within
  the window of the context defect is about missing tests for the same handler, and
  it scores as that defect. The same rule claims each finding once, so the two
  resilience defects four lines apart compete for one finding and the per-lens row
  under-counts them: `resilience 1/2` is that artefact, and the per-defect table is
  the reading. The next pass moves those two defects apart.
- **The confinement is opencode's.** A project that runs its review through
  another client, or through a command that ignores the configuration, keeps the
  prompt's sentence and nothing more. `claude` is not confined in this version.
- **The candidate can carry the rule that beats odachi's.** The client reads a
  configuration from the directory it runs in and every ancestor, and the nearest
  one wins. The reviewer runs in the candidate, so an `opencode.json` committed at
  the candidate's root is discovered closer than the deny odachi writes beside it.
  No test and no run in this page exercises that case; a project that ships one
  voids the boundary, silently. Either odachi refuses such a candidate or it
  merges the two rules; until then, the guarantee holds for candidates with no
  client configuration of their own.
- **It measures review, not delivery.** Nothing here says a defect a lens reports
  is a defect a reviewer would have caught in a pull request, and nothing says the
  findings are worth the calls they cost.
