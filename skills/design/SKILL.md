---
name: design
description: Trigger: design a screen, mockup, landing page or UI directions; iterate artboards; what a person asked on one; design system sync. HTML under design/<slug>/.
---

# Skill: design

## When

The human asks for a screen, mockup, landing page or visual directions, iterates on artboards, asks
what a person wrote on one, or captures the design system. Not for production code: that is the handoff.

## Rules

- **A design is a folder.** `design/<slug>/` holds `NN-<name>.html` artboards and `design.json`.
  Never overwrite another design: a new brief is a new slug.
- **Every artboard is one self-contained HTML file**: inline CSS, JS and SVG, system font stacks
  only. Never an external URL, CDN, remote font, image or script; it has to open offline.
- **Every artboard carries the screenshot it was read against**: a file beside it named in the
  manifest's `screenshot`, taken after the last change. A screen nobody looked at is a screen
  nobody checked; `odachi design list` names the ones without it.
- **Before designing, ask the folder**: `odachi design list` says whether `design/system.md` is
  missing, stale or fresh; sync before the first design and name a stale source before following
  old tokens.
- **Open the canvas after changes**: `odachi design open <slug>` returns at once and leaves a
  detached canvas; a second call moves the open page. Report its URL unless the human said not to.
- **Comments arrive as session messages when the OpenCode plugin is installed** (`odachi init
  --global` writes it); without it, or on Claude Code, read `design/<slug>/comments.json` and apply
  the `open` entries, saying which path you took.
- **The handoff records what happened**: `odachi design handoff <slug>/<artboard>` marks the chosen
  artboard `sent` and logs it in the linked feature; `--status in_progress` and then
  `--status implemented --preview http://localhost:<port>` (loopback: the canvas frames it).
- **The handoff is a decision, not a rewrite**: name the chosen artboard, name the others as the
  alternatives they were, and implement against the project's real stack.
- **Link a tracked feature both ways**: when the design comes from `odd/tasks/<feature>.md`, set
  `feature` in the manifest and append one `L` entry to that document's log naming
  `design/<slug>/` and its artboards; untracked work leaves `feature` out.
- **Directions differ in strategy, not in colour.** Three by default, each with the tradeoff it
  makes; three that make the same tradeoff are one option painted three times.
- **Pair with `ui-craft` before choosing directions**: it owns the hierarchy, the numbers and the
  tells; this skill owns the delivery. Read only the parts the brief needs.
- **The manifest is the truth**: `{"title", "brief", "feature"?, "artboards":[...]}` with `file`,
  `title`, `tradeoff`, `width`, `height`, `screenshot` per artboard, kept in sync.
- **Real content and accessible markup**: the audience's copy, no lorem ipsum; landmarks, WCAG AA,
  visible focus, `prefers-reduced-motion`.
- **Ask at most one clarifying question**, and only when the brief lacks the goal or the content.
- **Iterating edits only the targeted artboard.** Never rewrite the others.
- **Design system first when it exists**: read `design/system.md` and use only its tokens and
  component rules; before finishing an artboard, compare its colours and families with it.
- **Capture the system by reading, never by writing**: `odachi design sync` reads CSS custom
  properties, the Tailwind config, components, token files and fonts, and writes `design/system.md`
  and `design/system.html` with each entry's source. No sources: propose a starter, never invent.
- **Comments come from the folder, not a tool**: `design/<slug>/comments.json` carries `id`, `file`,
  `selector`, `excerpt`, `comment`, optional `screenshot`, `createdAt`, `status`, `sentAt`,
  `resolvedAt`. Apply each open comment to its artboard with `selector` and `excerpt`, and resolve it
  with `odachi design comments <slug> --resolve <id...>`; never delete or reorder, leave the rest open.

## Never

- Never reference a remote asset: an artboard that needs the network is not a mockup, it is a
  page that breaks tomorrow.
- Never edit the project's design sources during a sync.
- Never mark a comment resolved that you did not apply.
- Never present a design as implemented: the artboard is a direction, the code is the proof.
