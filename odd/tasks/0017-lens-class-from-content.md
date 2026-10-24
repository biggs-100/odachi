# lens-class-from-content

## Requirements (verbatim, from the user)
R1 · "the deterministic floor may raise the lens class from the diff's content — never lower it — and review and check ask the same lenses"

## Done when
- a change to ordinary code whose added lines spawn concurrency (go, chan, sync., atomic.) or touch secrets asks for the four lenses, and review and check agree on the answer
- a candidate whose path class is passive is never raised by prose, and a class is never lowered by content
- the human-owned review switch and the four-lens vocabulary are untouched
- go test ./... and odachi check are green

## Steps
- [ ] T1 · Assess: the floor raises the class from added lines (concurrency, secrets), never lowers it   R1
- [ ] T2 · review selects and stamps with Assess, and the prompt's class is the assessed one   R1
- [ ] T3 · check demands the same lenses: Diff in the port and content-aware Need per file   R1
- [ ] T4 · ADR 0017, CHANGELOG, impact map and dogfood of the raised class   R1

## History
L1 · "the deterministic floor may raise the lens class from the diff's content — never lower it — and review and check ask the same lenses"
L2 · OFICIO · "Assess raises only Focused to Full, from the added lines' concurrency or secrets; Passive prose never raises and nothing lowers" (porque Concurrency and secrets are the two classes a path cannot see; a document's words are not code, and over-reviewing code costs one prompt while under-reviewing auth costs a defect)
L3 · OFICIO · "Need and Assess share one assessClass, so the review selection and the gate's demand can never disagree" (porque ADR 0010's lesson was one authority for per-candidate lens demand; two implementations would drift the day one changes)
