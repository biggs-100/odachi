# the-canvas-works-like-gentle-shell

## What must exist
- odachi design works like gentle-shell's canvas: a detached local canvas with focus, compare, live handoff state, and comments that reach the OpenCode session through an installed plugin.

## Done when
- `odachi design open` returns at once and leaves the canvas running; `odachi design stop` closes it; a second `open` shows the requested design on the page already open instead of opening a new tab.
- Navigation and focus: in navigation the wheel and drag always pan or zoom the canvas, wherever the pointer is; clicking an artboard enters focus at exactly 100% with the mouse and keyboard captured and the artboard's own scroll, hovers and clicks; `Shift+Esc` or a visible "Back to canvas" returns to the previous view; comment mode picks an element instead of entering focus.
- `design/<slug>/handoffs.json` records per artboard `sent`, `in_progress`, `implemented`, the design's `chosen` artboard and an optional loopback `previewUrl`; the canvas shows the state live, marks the chosen artboard over its alternatives, and Compare shows the artboard beside the running app, with an "Open in new tab" fallback when the app refuses to be framed.
- Comments: the canvas pins each commented element, offers "Send now" and "Collect" with "Send all", and writes `design/<slug>/comments.json`; an OpenCode plugin installed by odachi delivers each comment (slug, artboard, selector, excerpt, screenshot) as a session message and marks the queue entry delivered; Claude Code has no push and reads the file.
- Security: every canvas route requires a per-server random token; the server rejects a foreign Host and a POST without its own Origin and a JSON body; artboards render in sandboxed iframes without `allow-same-origin`; only manifest-listed files are served, with no traversal and no symlink escape; comment length is bounded.
- The canvas shows `design/system.html` as the first board when it exists.
- `odachi init` installs the OpenCode plugin, and the design skill tells the agent when to open or refresh the canvas, read the comments and update the handoff; a plugin or canvas failure is named, never silent.
- `go test ./...` is green; the queue contract and every canvas rule above are pinned by Go tests; the plugin's delivery is verified by hand in OpenCode and recorded in the proof.

## Not included (first version)
- no automatic design-system sync and no stale-source warning: that is the next feature
- no two-way ODD Log link and no pairing with the ui-craft skill: that is the next feature
- no push for Claude Code: it reads `comments.json` and `handoffs.json`; the skill documents the difference
- no native `gentle_design` tool: the skill and the `odachi design` commands are the interface
- no change to the gate, the review, the proof format or the artboard format: `design.json` stays as it is and `handoffs.json` is the only new file
- no Node test runner and no CI step for the plugin: the queue is tested in Go and the plugin is thin and verified by hand

## Decisions taken
- It is used from the computer.
- There is no sensitive data in the first version.
- First usable version: this week.
- Wants to learn: how a local server pushes into a session without a provider; how a canvas and a session agree without a race.
