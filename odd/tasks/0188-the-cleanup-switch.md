# the-cleanup-switch
Base: 01c5577094e8d3d7265204a49ff393cbcced7481

## Requirements (verbatim, from the user)
R1 · "the cleanup switch cannot be wedged by the shape of a folder: a path beyond the walk names itself as outside the shape and does not block --apply, a class --apply empties is removed as a tree, and the report never promises bytes the switch refuses to free"

## Done when
- a folder deeper than the walk is counted as outside the shape, not as unreadable, and it does not block --apply
- a real read failure still blocks every removal and names the path
- apply empties scratch and scaffolding as folders, so a subtree the walk never enumerated goes with them
- a lifetime class -records, prompts, backups- is still removed file by file
- the clean report says the switch is blocked when it is, instead of promising bytes it will refuse to free
- a close store line and the report agree: only a real read failure blocks
- odachi clean --apply empties this clone store and frees the bytes it promised, and go test ./... is green

## Steps
- [ ] T1 · a folder beyond the walk is outside the shape, counted and not a failure; a real read error still blocks   R1 · route:inline
- [ ] T2 · apply empties scratch and scaffolding as folders, so a subtree the walk never enumerated goes with them   R1 · route:inline
- [ ] T3 · the report says the switch is blocked when it is, instead of promising bytes it will refuse to free   R1 · route:inline
- [ ] T4 · the close and the report agree: only a real read failure blocks, and both name the same store   R1 · route:inline

## History
L1 · "the cleanup switch cannot be wedged by the shape of a folder: a path beyond the walk names itself as outside the shape and does not block --apply, a class --apply empties is removed as a tree, and the report never promises bytes the switch refuses to free"
