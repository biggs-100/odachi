# a-feature-can
Base: 68f1c040c717dbaa3dd399a06d4320a218314f9e

## Requirements (verbatim, from the user)
R1 · "a feature can be chosen by its number, not only the highest"

## Done when
- plan, ask, check and progress accept --feature NNNN and act on that feature
- without the flag the highest-numbered feature is used, exactly as today
- an unknown number is blocked and names the features that exist

## Steps
- [x] T1 · the selector resolves a number to its feature, and plan and ask act on it   R1 · eeb2fff
- [x] T2 · check and progress judge the selected feature, and an unknown number names the ones that exist   R1 · ebffb96

## History
L1 · "a feature can be chosen by its number, not only the highest"
L2 · OFICIO · "the first review of 0060 did not seal: the reviewer answered the JSON and then a footer, and the parse refused it" (porque the findings were still evidence and were fixed before the second run: check selector on a project with no docs, and a negative number folding into the default)
L3 · OFICIO · "check delegates its selector to tasks.Feature instead of repeating the lookup" (porque a second copy of the rule answered nothing on an empty project; the gate must refuse an unknown number like every other verb)
