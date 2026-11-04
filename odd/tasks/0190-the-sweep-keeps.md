# the-sweep-keeps
Base: 35092d9ffbe63a8c317c8a898d04c3c4d05817bd
Light: yes

## Requirements (verbatim, from the user)
R1 · "the sweep keeps every prompt when a ModTime fails, and names the count when a Remove fails"

## Done when

## Steps
- [x] T1 · the sweep keeps every prompt when a ModTime fails, and names the count when a Remove fails   R1 · tree:a5b07a8d2d11d01b7393ca4711e8ceb2b2297d80

## History
L1 · "the sweep keeps every prompt when a ModTime fails, and names the count when a Remove fails"
L2 · OFICIO · "the review's WARNING on the ModTime test is verified false: the run's own prompt is the second, readable, prunable prompt, so disabling the ModTime branch left only the unreadable one and the test failed on the skip-and-prune case" (porque the mutation output was prompts = one unreadable file, which is exactly the case the warning claimed the test could not see; a second stale prompt would strengthen the fixture but would change the tree the verdict judged, and a WARNING earns no review round)
