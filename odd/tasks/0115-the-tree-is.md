# the-tree-is
Base: 6957fbc57856138e49f6160ab4b19e2f4775fc0a

## Requirements (verbatim, from the user)
R1 · "the tree is gofmt-clean and a test keeps it so: every Go file in the repository is the bytes gofmt would write, checked in process so a CRLF or a missing final newline cannot land again"

## Done when
- gofmt -l . prints nothing, and the file that was red is canonical
- a test walks every .go file and fails when format.Source would rewrite it, with the path and the first difference
- the test is fast: it parses the tree in one pass, without spawning a process per file
- the changelog says one line

## Steps
- [x] T1 · the file is canonical, and a test walks every .go file with format.Source   R1
- [x] T2 · the changelog says one line   R1 · tree:e5cdc59cde1539624811e9715a311fda63cb965c

## History
L1 · "the tree is gofmt-clean and a test keeps it so: every Go file in the repository is the bytes gofmt would write, checked in process so a CRLF or a missing final newline cannot land again"
L2 · OFICIO · "the plan said the blob carried CRLF and it did not: .gitattributes normalizes *.go to LF, the blob was canonical, and the CRLF lived in the worktree this machine wrote" (porque a plan written from a worktree reading must say which bytes it measured; the defect that survives is that nothing checked formatting in the suite)
