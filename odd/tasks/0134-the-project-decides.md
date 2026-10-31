# the-project-decides
Base: e46f8f81aecbe77fad9513f8e744023572f225ca

## Requirements (verbatim, from the user)
R1 · "the project decides how much review evidence survives, in odd/review.toml"

## Done when
- keep_verdicts and keep_prompts live in odd/review.toml, and a project that never writes them keeps today's numbers: ten records and no prompt
- odachi clean reads them from the file, uses the defaults when the file does not name them, refuses a file it cannot read by naming it, and says which numbers it used
- an older odachi refuses the new file as newer, not as an unknown key
- odachi review mode on rewrites the file and keeps the two numbers

## Steps
- [ ] T1 · odd/review.toml carries keep_verdicts and keep_prompts, and the schema moves so an old binary says upgrade   R1
- [ ] T2 · clean reads the numbers from the file, defaults when absent, refuses a file it cannot read, and says which it used   R1
- [ ] T3 · the changelog and the ADR say what the project decides, and review mode keeps the numbers   R1

## History
L1 · "the project decides how much review evidence survives, in odd/review.toml"
L2 · OFICIO · "the schema moves to 3 for two keys added to odd/review.toml" (porque one counter serves the three strict files, and 0094 accepted that cost: an older binary now refuses a verdict whose shape did not change, and says upgrade, which is still true)
