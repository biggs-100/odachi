# the-review-refuses
Base: 8e9d2c10e2571915ff6e0e3d6b88a53bffc3e080

## Requirements (verbatim, from the user)
R1 · "the review refuses to start a reviewer on a prompt bigger than the project allows, and says the size, the limit and the ways out"

## Done when
- odd/review.toml carries the size limit, and a file without it means no limit
- review --exec refuses before starting anything when the prompt is over the limit, naming the size, the limit and the two ways out
- the printed form still prints, because it spends nothing, and says the size on stderr
- a test drives both answers, observed red first

## Steps
- [x] T1 · odd/review.toml carries the size limit and review --exec refuses before starting a reviewer   R1
- [x] T2 · the printed form prints anyway and says the size on stderr   R1

## History
L1 · "the review refuses to start a reviewer on a prompt bigger than the project allows, and says the size, the limit and the ways out"
L2 · OFICIO · "the size limit is the project's, not the model's: the tool cannot know what a model reads, so a human measures it once and writes it in odd/review.toml" (porque a per-model table would be a guess dressed as configuration; the measurement in this project's file says 89 KB closed in 46s and 122 KB stalled, and only the human can decide which model they pay for)
L3 · OFICIO · "the printed form never refuses, because it spends nothing" (porque the human's debugging tool cannot be blocked by a limit that exists to protect a paid call; it prints and says the size where the run's notes go)
