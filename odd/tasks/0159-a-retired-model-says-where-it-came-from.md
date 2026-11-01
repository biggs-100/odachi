# a-retired-model-says-where-it-came-from
Base: 37277bbd78dd8a7f666900ef52345165e85c3d89

## Requirements (verbatim, from the user)
R1 · "the project's review names a model the provider serves today, and a reviewer that cannot start says which model it tried and where that model came from, so a retirement is one line to fix instead of an investigation"

## Done when
- odd/review.toml assigns the four lenses a model the provider serves, measured on 2026-11, and a test pins that every entry is a provider/model the decoder accepts
- a failed reviewer names the model, the lens and the source it came from -the flag, odd/review.toml or the session- beside the command it ran, observed with a shell that fails
- the comment the file carries tells the next human what to check when a model is retired

## Steps
- [x] T1 · odd/review.toml names a model the provider serves, and the file says what to check when one is retired   R1 · tree:2c83f07e42475d4ebfdc407400133c262ce74691
- [ ] T2 · a reviewer that cannot start names the model and the source it came from   R1

## History
L1 · "the project's review names a model the provider serves today, and a reviewer that cannot start says which model it tried and where that model came from, so a retirement is one line to fix instead of an investigation"
L2 · OFICIO · "the four lenses are judged by opencode-go/mimo-v2.6-pro" (porque measured 2026-11 against the provider's own list: mimo-v2.6-pro, mimo-v2.5 and mimo-v2.5-pro are served and mimo-v2.6-flash is retired; the pro is the newest of the family and the one that answered the reviews of 0157 and 0158)
L3 · OFICIO · "a failed reviewer names the model, the lens and the line that chose it" (porque an offline tool cannot know what a provider still serves, and it can say which model it used and where that model came from, which turns a retirement into one edit instead of an investigation)
