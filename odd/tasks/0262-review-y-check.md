# review-y-check
Base: 0357fc02ed5c57eb3039f93ce4e9fdd1afa5ece0
Light: yes

## Requirements (verbatim, from the user)
R1 · "review y check nombran el mismo árbol para el mismo worktree limpio, y una carpeta ignorada no mueve ese id"

## Done when

## Steps
- [x] T1 · review y check nombran el mismo árbol para el mismo worktree limpio, y una carpeta ignorada no mueve ese id   R1 · tree:099ac096998a5ea6b6fad42255c82b07d942f016

## History
L1 · "review y check nombran el mismo árbol para el mismo worktree limpio, y una carpeta ignorada no mueve ese id"
L2 · OFICIO · "The index the candidate is read from starts at HEAD, not at the base: one helper, stageWorktree, now feeds TreeHash, Candidate and Diff" (porque read-tree base keeps the entry of a path a later commit untracked, and add --all stages it again even though git ignores it, so review named a tree check never asks about; HEAD is the only parent the three verbs agree on)
L3 · OFICIO · "A correction measures its delta through the same HEAD-rooted index, so a path the judged tree tracked and HEAD has untracked reads as a deletion there" (porque the delta is the worktree as check names it; two readers disagreeing about what the candidate contains is the bug this feature removed, and reporting the removal as a deletion is what git says)
