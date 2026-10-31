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
- [x] T1 · odd/review.toml carries keep_verdicts and keep_prompts, and the schema moves so an old binary says upgrade   R1 · tree:0df7135838c41235590fe037d19539041a1a05e5
- [x] T2 · clean reads the numbers from the file, defaults when absent, refuses a file it cannot read, and says which it used   R1 · tree:0df7135838c41235590fe037d19539041a1a05e5
- [x] T3 · the changelog and the ADR say what the project decides, and review mode keeps the numbers   R1 · tree:1f88da6d7f6f31378c5704870264e900052bb6be

## History
L1 · "the project decides how much review evidence survives, in odd/review.toml"
L2 · OFICIO · "the schema moves to 3 for two keys added to odd/review.toml" (porque one counter serves the three strict files, and 0094 accepted that cost: an older binary now refuses a verdict whose shape did not change, and says upgrade, which is still true)
L3 · OFICIO · "this repository's own odd/review.toml stays at schema 2 until a release reads 3" (porque CI builds the gate from the base ref, and that binary reads schema 2: a candidate file that says 3 would be refused before the suite ran, so the file that carries the keys follows the release that understands them)
L4 · OFICIO · "T2's ledger mark names T1's tree (0df7135); its commit is f7482d8" (porque the mark was written by a command that ran after a commit the hook refused, and plan --done cannot be re-run for a step already done: the close of the feature is judged by the worktree tree, so the gate is not affected)
L5 · OFICIO · "the review's one suggestion earns no round and goes to docs/review-debt.md" (porque the tree's behaviour is right and the finding is a misplaced comment; fixing it now would change the bytes the verdict judged and pay for a second review of the same work)
