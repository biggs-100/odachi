# a-run-cannot
Base: 6bc2a8a85acb66741d8679163b91ab69a3015b54

## Requirements (verbatim, from the user)
R1 · "a run cannot widen the scope without saying why: --lens that adds a lens needs --because, and the refusal names the flag and the lenses it would add"

## Done when
- a run whose --lens adds a lens and has no --because is refused before any reviewer starts, and the refusal names --because and the lenses
- the refusal writes nothing: no verdict, no record, no refused answer
- a tree that already has its record keeps working with --lens and no new words, and a run that adds no lens is never asked
- a test drives each answer, and one of them fails without the change

## Steps
- [x] T1 · a run that widens the scope without words is refused before it spends anything   R1 · tree:3d924fe0a3bc5efde09b0679fe77f64b5698fa0b
- [ ] T2 · the debt file and the changelog say the gap closed   R1

## History
L1 · "a run cannot widen the scope without saying why: --lens that adds a lens needs --because, and the refusal names the flag and the lenses it would add"
L2 · OFICIO · "the refusal cannot touch cmd/odachi/main.go, so the --lens help text stays as it is and the refusal carries the rule at the moment the run is refused" (porque the fence in force at this base protects the CLI flag surface, and a help string is not worth loosening it: the message names --lens, --because and the lenses, which is where a caller learns it)
