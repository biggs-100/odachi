# the-reviewer-runs

## Requirements (verbatim, from the user)
R1 · "the reviewer runs with the model this session uses, by default, and the human can name another one"

## Done when
- {model} in review --exec is replaced with the session's provider/model#variant without the human typing it
- --model overrides the resolved model for one run and is used exactly as given
- a command without {model} runs byte-identical to today
- when {model} is asked for and nothing resolves, review stops naming --model; it never guesses
- resolution order: --model, then ODACHI_MODEL, then the OpenCode session; no database or private file is read

## Steps
- [x] T1 · {model} in review --exec resolves to the session's model, or the run stops naming --model   R1 · d8d785f
- [x] T2 · docs: how the reviewer inherits the model, the resolution order and ADR 0018   R1 · 1635e68

## History
L1 · "the reviewer runs with the model this session uses, by default, and the human can name another one"
L2 · OFICIO · "The model travels as a {model} placeholder in the human's command, not as a flag odachi injects" (porque odachi cannot know how each reviewer CLI spells its model flag, and the human keeps writing the command they already know)
L3 · OFICIO · "Resolution order: --model, then ODACHI_MODEL, then opencode session export" (porque the flag is the human's override for one run, the variable is the host contract any agent can set, and the CLI is the last honest source)
L4 · OFICIO · "A placeholder with no model stops the run before the reviewer starts; there is no fallback" (porque a verdict from a model nobody chose is a review nobody asked for, and the receipt would hide it)
L5 · OFICIO · "The probe reads opencode session export, never the OpenCode database" (porque the CLI is a supported contract; a private schema owes odachi nothing and breaks on upgrade)
L6 · OFICIO · "The verdict records the substituted command, not the template" (porque the audit trail has to name the identity that judged, or it is not an audit trail)
L7 · OFICIO · "A --model without {model}, or without --exec, is refused instead of ignored" (porque a flag that changes nothing is a lie about what ran)
