# the-subagents-live
Base: d4b1e4d7eff345bf59d9b469691168c5d40d798d

## Requirements (verbatim, from the user)
R1 · "the subagents live in the home too, so a project that never writes .opencode keeps the worker, the verifier and the reviewer"

## Done when
- init --global and odachi setup write odachi-worker, odachi-verify and odachi-reviewer in the home of each chosen client, next to the door
- without --global the home gets nothing, and a second run reports the subagents unchanged
- doctor compares the home copies and names a stale one
- the tests pin the paths, the write and the comparison, and the map and README record it

## Steps
- [x] T1 · the domain renders the global subagents and --global writes them, with the tests   R1 · tree:638d9f59c3540ec129c26e894562e428011bc753
- [x] T2 · doctor compares the home copies, and the map and the README record it   R1 · tree:1781dc2f63f8fabc672672fef5ab83ac908de586

## History
L1 · "the subagents live in the home too, so a project that never writes .opencode keeps the worker, the verifier and the reviewer"
L2 · OFICIO · "the global subagents reuse the project wrapper and prompt, only the root changes" (porque two prompts that have to agree would drift; the same bytes in two roots are still one thing)
