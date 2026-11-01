# a-retired-model-says-where-it-came-from

## What must exist
- the project's review names a model the provider serves today, and a reviewer that cannot start says which model it tried and where that model came from, so a retirement is one line to fix instead of an investigation

## Done when
- odd/review.toml assigns the four lenses a model the provider serves, measured on 2026-11, and a test pins that every entry is a provider/model the decoder accepts
- a failed reviewer names the model, the lens and the source it came from -the flag, odd/review.toml or the session- beside the command it ran, observed with a shell that fails
- the comment the file carries tells the next human what to check when a model is retired

## Not included (first version)
- no network and no provider query: the binary cannot know what a provider serves, and a gate that calls out to answer is not a gate
- no automatic retry with another model: a verdict from a model nobody chose is a review nobody asked for
- the four lens prompts are not edited; this is the model that runs them, not what they ask

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: today.
- Wants to learn: why an offline tool cannot know which models exist and still owes an honest answer about the one it used.
