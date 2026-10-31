# the-adoption-is
Base: 747dfe6487ad270b35d5a287bea69d2fc172f2b6

## Requirements (verbatim, from the user)
R1 · "the doctor calls a correct adoption an unfinished one: an adopted project has no agent file and no project_agents key -by design, because the inference is never stored- and check 11 reports exactly that state as a gap, advising odachi init --global to record a choice. Either the note recognises the adopted state or the advice stops calling the missing key a defect"

## Done when
- the adoption's three facts -does this repository live, is the craft here, is an agent file here- are read in one place, and init, role and doctor ask it
- a project that adopted odachi is not reported as unfinished, and the note names odachi setup when the door is missing
- a project that never ran odachi is not called adopted: the note names what it sees and the two commands that would change it
- init and role keep their behaviour: their suites stay green without an edited assertion

## Steps
- [x] T1 · the adoption's three facts are read in one package, and init and role ask it   R1 · tree:f3e8afae7aa73e572bb4cc56724afebcb020ae6d
- [ ] T2 · the doctor recognises the adopted state and stops calling a silent store a defect   R1
- [ ] T3 · the ADR and the changelog say the adoption is one inference   R1

## History
L1 · "the doctor calls a correct adoption an unfinished one: an adopted project has no agent file and no project_agents key -by design, because the inference is never stored- and check 11 reports exactly that state as a gap, advising odachi init --global to record a choice. Either the note recognises the adopted state or the advice stops calling the missing key a defect"
L2 · OFICIO · "the doctor reads the adoption through the package that owns it, and calls a silent store a shape and not a gap" (porque an adopted project has no agent file and no project_agents key by design, and advising init --global recorded a decision nobody made)
