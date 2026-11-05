# the-canvas-works-like-gentle-shell
Base: 878e8cd6ca7eb751590b67bfeeb695e133f6928b

## Requirements (verbatim, from the user)
R1 · "odachi design works like gentle-shell's canvas: a detached local canvas with focus, compare, live handoff state, and comments that reach the OpenCode session through an installed plugin."

## Done when
- `odachi design open` returns at once and leaves the canvas running; `odachi design stop` closes it; a second `open` shows the requested design on the page already open instead of opening a new tab.
- Navigation and focus: in navigation the wheel and drag always pan or zoom the canvas, wherever the pointer is; clicking an artboard enters focus at exactly 100% with the mouse and keyboard captured and the artboard's own scroll, hovers and clicks; `Shift+Esc` or a visible "Back to canvas" returns to the previous view; comment mode picks an element instead of entering focus.
- `design/<slug>/handoffs.json` records per artboard `sent`, `in_progress`, `implemented`, the design's `chosen` artboard and an optional loopback `previewUrl`; the canvas shows the state live, marks the chosen artboard over its alternatives, and Compare shows the artboard beside the running app, with an "Open in new tab" fallback when the app refuses to be framed.
- Comments: the canvas pins each commented element, offers "Send now" and "Collect" with "Send all", and writes `design/<slug>/comments.json`; an OpenCode plugin installed by odachi delivers each comment (slug, artboard, selector, excerpt, screenshot) as a session message and marks the queue entry delivered; Claude Code has no push and reads the file.
- Security: every canvas route requires a per-server random token; the server rejects a foreign Host and a POST without its own Origin and a JSON body; artboards render in sandboxed iframes without `allow-same-origin`; only manifest-listed files are served, with no traversal and no symlink escape; comment length is bounded.
- The canvas shows `design/system.html` as the first board when it exists.
- `odachi init` installs the OpenCode plugin, and the design skill tells the agent when to open or refresh the canvas, read the comments and update the handoff; a plugin or canvas failure is named, never silent.
- `go test ./...` is green; the queue contract and every canvas rule above are pinned by Go tests; the plugin's delivery is verified by hand in OpenCode and recorded in the proof.

## Steps
- [x] T1 · odachi design open leaves the canvas running and returns, and a second open refreshes the page already open   R1 · route:inline · tree:9f2dc27004640d0261a7b7b56cf6afa8d045d005
- [x] T2 · the canvas requires its token, refuses a foreign Host and a cross-origin POST, and sandboxes the artboards   R1 · route:inline · tree:f4d374ab3b7206632f2b9156e16d048229bb9d3a
- [x] T3 · handoffs.json records sent, in_progress, implemented, chosen and previewUrl, and the canvas marks and compares   R1 · route:inline · tree:8ee18653a4122e2a305bcec5813f1c46ddc03bca
- [ ] T4 · the canvas focuses an artboard at 100% with pointer capture, Shift+Esc returns, comments pin and Collect batches   R1 · route:inline
- [ ] T5 · odachi init installs the OpenCode plugin, and the canvas queue delivers a comment to the session and marks it delivered   R1 · route:inline
- [ ] T6 · the design skill opens the canvas, reads comments and updates the handoff, and the docs name the Claude difference   R1 · route:inline
- [ ] T7 · the canvas shows the system board first when design/system.html exists   R1 · route:inline

## History
L1 · "odachi design works like gentle-shell's canvas: a detached local canvas with focus, compare, live handoff state, and comments that reach the OpenCode session through an installed plugin."
