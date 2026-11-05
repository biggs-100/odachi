# Next session

Written at the close of 0200 and v0.15.0. `main` is clean, the canvas is the one the
release ships, and the OpenCode push plugin is installed on this machine.

## What shipped

- **0200, PR #328**: the design canvas works like gentle-shell's. `design open` is a
  detached child that returns; a second `open` moves the page already open. Every route
  requires a per-server token, a foreign Host is refused, artboards render in sandboxed
  iframes with an injected picker, and only manifest-listed files resolve inside the
  design. `handoffs.json` records `chosen`, `sent`, `in_progress`, `implemented` and a
  loopback `previewUrl`; the canvas marks the chosen artboard and compares it with the
  running app. Focus at 100% with pointer capture, `Shift+Esc`, pins, Collect and Send
  all. The system board comes first. Review approved 4 of 4 after four findings were
  fixed (queue order, focus after a live render, plugin retry, fence rules).
- **v0.15.0, PR #330**: released with 14 signed assets and installed from the release;
  `odachi --version` answers `odachi 0.15.0`.
- **`odachi init --global`**: wrote the push plugin at
  `~/.config/opencode/plugins/odachi-design/index.ts`. The live check: a comment sent
  with `send:true` wrote one queue file, the plugin consumed it, and the comment reached
  the session.

## What was measured

- `go test ./... -count=1` green; CI green on `checks (linux)` and the three `gate`
  jobs. The canvas JavaScript and the plugin were syntax-checked with `node --check`.
- The push was verified by hand in OpenCode: the queue directory was empty after the
  send and the comment's `sentAt` was recorded.
- The base fence of 0199 protected `cmd/odachi/main.go` as "no new flag and no new verb",
  so the two help-text lines that described the new canvas were reverted; they wait for
  a docs feature that changes that rule on `main` first.

## What is already true

- `docs/review-debt.md` has nothing open.
- No release is owed: `## Unreleased` is gone from `CHANGELOG.md`.
- The canvas interaction (focus, pins, Collect) has no browser automation: it was built
  against the documented behavior and needs a human look in the browser. The plugin's
  delivery was verified live; Claude Code stays on the pull path (`comments.json`,
  `handoffs.json`) and the skill says so.
- The next parity slice is the one left out of 0200: automatic design-system sync with
  stale-source detection, the two-way ODD Log link, and pairing the design skill with
  `ui-craft`. Nothing is queued; the next feature is the human's call.
