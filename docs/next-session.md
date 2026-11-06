# Next session

Written at the close of 0219, `the-design-pair-covers-product-ux`. The design pair now covers
collections and checks every artboard before it is reported.

## What shipped

- **`ui-craft` carries `browse-and-compare`.** The representation, the search scope, the filters,
  pagination/load more/virtualization/infinite scroll, and the return with page, scroll and focus
  restored. The body routes it in `## Decision gates`.
- **`design` checks every artboard.** Each finished artboard goes through `ui-craft`'s `the-read`
  (and `adaptive-interfaces` when it has states); another approach on the same brief is the next
  `NN` artboard, never a new slug; and a comment's `screenshot` is read first, because it shows the
  transient state the static file cannot.
- **The gate.** `TestUICraftTeachesBrowseAndCompare` and `TestDesignChecksEveryArtboard` were red
  before the text landed. Proof `proof/0219-the-design-pair-covers-product-ux.txt`; the review
  approved the tree on the reliability lens with no findings. Released as v0.24.0 and installed.

## What was measured

- **The payload golden list catches a new file, not just a new skill.** `skills/router_test.go` pins
  the exact file list of `ui-craft`; the new reference broke it after the focused test was green.
  The T1 commit was made before the full package ran; it was reset locally and re-recorded with the
  green tree. Run the whole package before committing a new file, not only the focused test.
- **The plan can ride the feature PR when the fence allows it.** The 0218 fence did not protect
  `design` or `ui-craft`, so the plan commit and the feature shared PR #389; the two-PR split is for
  a path the base fence protects, not a ritual.
- **The install was clean this time.** No gate held the binary; the release installer replaced two
  skills and left 19 matched.

## The next feature

- **The canvas tuning.** Sliders for spacing, colour and type scale, plus a "Copy as prompt" control,
  so a person can tune values in the artboard and hand the chosen numbers back as text. It is code
  (`internal/assets/design/canvas.html` and its tests), so the fence has to let the canvas change.
- **The installer names the process that holds the binary**, still open from 0218.
- **Pin `all:*`** with a repo test, still open from 0216.
- One writer per worktree; one heavy gate per machine.

## The commands that work

```sh
odachi review --base ab82e8b66afb51a1221cf49cdf77d3fe71f79a1d \
  --exec 'C:\Users\USER\AppData\Roaming\npm\node_modules\@opencode\cli\bin\opencode.exe run --agent odachi-reviewer --model {model} --file {prompt} "Answer the review prompt in the attached file."'
```

`--base` is the base the task doc names, not `main`.

```sh
$env:PATH = "C:\Windows\System32;C:\Windows;C:\Windows\System32\Wbem;C:\Program Files\Go\bin;C:\Users\USER\go\bin;C:\Users\USER\AppData\Local\odachi\bin;C:\Users\USER\AppData\Roaming\npm"; odachi check --run "go test ./..."
```

No `--timeout` is needed; the default is 900. A build goes to
`%TEMP%\opencode\odachi\<feature>\`, never to the clone.
