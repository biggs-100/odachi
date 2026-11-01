# a-retired-model-says-where-it-came-from
Base: 37277bbd78dd8a7f666900ef52345165e85c3d89

## Requirements (verbatim, from the user)
R1 · "the project's review names a model the provider serves today, and a reviewer that cannot start says which model it tried and where that model came from, so a retirement is one line to fix instead of an investigation"

## Done when
- odd/review.toml assigns the four lenses a model the provider serves, measured on 2026-11, and a test pins that every entry is a provider/model the decoder accepts
- a failed reviewer names the model, the lens and the source it came from -the flag, odd/review.toml or the session- beside the command it ran, observed with a shell that fails
- the comment the file carries tells the next human what to check when a model is retired

## Steps
- [ ] T1 · odd/review.toml names a model the provider serves, and the file says what to check when one is retired   R1
- [ ] T2 · a reviewer that cannot start names the model and the source it came from   R1

## History
L1 · "the project's review names a model the provider serves today, and a reviewer that cannot start says which model it tried and where that model came from, so a retirement is one line to fix instead of an investigation"
