# odachi-clean-reports
Base: dfb18c7fe1eeea310bf895448b9047cbd7ae697b

## Requirements (verbatim, from the user)
R1 · "odachi clean reports every class of artifact in a clone with its owner, its size and what it would do, and applies retention only to the classes odachi itself wrote"

## Done when
- odachi clean prints every class with its owner, its file count and its size, and names a path it could not read instead of counting it as empty
- odachi clean --apply leaves the newest review records and the newest backups and deletes nothing outside the store: git's files, another tool's files, proof/ and odd/ stay byte-identical
- odachi clean --apply --legacy removes the old agent scratch files in .git/ after naming them, and --apply without --legacy leaves them
- the block says where an ad-hoc note goes, and odachi clean empties that folder

## Steps
- [x] T1 · clean names every class with its owner, its files, its size and what it would do, and a failed read is named   R1 · tree:a915e647fdef8bd778c2d7c957008f8bad178693
- [x] T2 · clean --apply keeps the newest review records and the newest backups, empties scratch, and deletes nothing else   R1 · tree:b94067307b0b143e9377570155301c555f831866
- [x] T3 · clean --apply --legacy removes the old agent scratch files in .git/ after naming them   R1 · tree:b94067307b0b143e9377570155301c555f831866
- [x] T4 · the block says where an ad-hoc note goes, and the folder is the one clean empties   R1 · tree:f3ba85023a4dbc80d0021874a9d6e82cc233eda6

## History
L1 · "odachi clean reports every class of artifact in a clone with its owner, its size and what it would do, and applies retention only to the classes odachi itself wrote"
L2 · OFICIO · "the fence carries its two rules with real paths, written by hand after start" (porque the binary installed here is v0.11.72, older than 0130, so the run wrote a glob of the phrase's own words; a rule protects paths and the human's words did not change)
L3 · OFICIO · "the scratch rule goes in the AGENTS.md block and not in the drive skill" (porque the block is the door every agent reads in any project, and the fence in force protects skills/**, so drive would cost a policy change first)
L4 · OFICIO · "T2 and T3 landed in one commit (0aca2f4)" (porque the legacy class was already in the report and the removal is the same loop over the plan; separating them would have meant writing the class twice, so the step is closed by the commit that carries its tests)
