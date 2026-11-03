# el-folder-de
Base: 5232bad7b4fb02d8b48a3d7edf45d113d96aa706

## Requirements (verbatim, from the user)
R1 · "el folder de diseño exige una captura por artboard; design list y el canvas nombran las que faltan"

## Done when
- design.json names a screenshot per artboard, and a screenshot that is not a file inside the design folder is refused by name
- odachi design list names every artboard without a screenshot, and the canvas marks it
- the design skill says a design is not done until each artboard has its screenshot
- tests pin the present, the missing and the refused case

## Steps
- [x] T1 · the manifest names a screenshot per artboard and refuses a name outside the design folder   R1 · route:inline · tree:d82ac66847bfced4fc411092b8b0b265739fb6e3
- [x] T2 · design list and the canvas name the artboards without a screenshot   R1 · route:inline · tree:f5fa46a79f4f4d29b4e0f50af5ac73a12a3b2614
- [ ] T3 · the design skill says an artboard is not done until its screenshot is beside the manifest   R1 · route:inline

## History
L1 · "el folder de diseño exige una captura por artboard; design list y el canvas nombran las que faltan"
L2 · OFICIO · "design.json accepts an artboard without a screenshot, and the gap is named where a reader sees it" (porque a design written before the field keeps its bytes; the rule is enforced by the list and the canvas naming the gap, not by refusing the past (0094))
