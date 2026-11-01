# the-lenses-name-what-they-missed
Base: eb1c1dcb00e25dede43d845eac0fd8496204c500

## Requirements (verbatim, from the user)
R1 · "the four lens prompts name the defects the comparison with gentle-ai found missing: a suite that got smaller, a test at the wrong level, a failure with no threshold and a failure nobody can see, and behaviour the brief does not explain"

## Done when
- reliability names a focused or skipped test that makes CI pass by shrinking the suite, and a test that proves a function through a whole flow
- resilience names a failure with no threshold the project chose, and a failure with nothing in production that would show it
- readability names behaviour the diff changes and the brief never asked for
- every lens keeps between twelve and twenty rules of the length the format enforces, and the asset tests stay green

## Steps
- [x] T1 · reliability names a suite that got smaller and a test at the wrong level   R1 · tree:c43307ec5c00b43de865f49c94232a11a12b40f1
- [x] T2 · resilience asks for a threshold the project chose and something in production that would show the failure   R1 · tree:792637ffaa6a1e92a273f48b9d21c7f98c23b922
- [x] T3 · readability names behaviour the diff changes and the brief never asked for   R1 · tree:d0eaf0a827331925f007d58e7462171cabe8346b

## History
L1 · "the four lens prompts name the defects the comparison with gentle-ai found missing: a suite that got smaller, a test at the wrong level, a failure with no threshold and a failure nobody can see, and behaviour the brief does not explain"
