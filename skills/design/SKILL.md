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
  nobody checked, and `odachi design list` names the ones without it.
- **Open the canvas after writing or changing artboards**: `odachi design open <slug>` returns at
  once and leaves a detached canvas; a second call moves the page already open. Report the URL;
  skip it only when the human said not to.
- **Comments arrive as session messages when the OpenCode plugin is installed** (`odachi init
  --global` writes it); without it, or on Claude Code, read `design/<slug>/comments.json` and apply
  the `open` entries, saying which path you took.
- **The handoff records what happened**: `odachi design handoff <slug>/<artboard>` writes the chosen
  artboard as `sent`; run it with `--status in_progress` when you start and `--status implemented
  --preview http://localhost:<port>` when the app runs. The preview is loopback only: the canvas
  frames it and compares it with the artboard.
- **The handoff is a decision, not a rewrite**: name the chosen artboard, name the others as the
  alternatives they were, and implement against the project's real stack.
- **Directions differ in strategy, not in colour.** Three by default, each with the tradeoff it
  makes written in the manifest. Three options that make the same tradeoff are one option painted
  three times.
- **Realistic content and accessible markup**: the audience's real copy, no lorem ipsum; semantic
  elements and landmarks, WCAG AA contrast, visible focus, `prefers-reduced-motion`.
- **The manifest is the truth**: `{"title", "brief", "feature"?, "artboards":[...]}` with `file`,
  `title`, `tradeoff`, `width`, `height` and `screenshot` per artboard, kept in sync on every
  change. `feature` is the task document it came from, empty for untracked work.
- **Ask at most one clarifying question**, and only when the brief lacks the goal or the content.
- **Iterating edits only the targeted artboard.** Never rewrite the others.
- **Design system first when it exists**: read `design/system.md` and use only its tokens and
  component rules. Before finishing an artboard, list its colours and font families, compare them
  with the system, and fix every deviation.
- **Capture the system by reading, never by writing**: `odachi design sync` reads CSS custom
  properties, the Tailwind config, component folders, token files and their fonts, and writes
  `design/system.md` and `design/system.html` with each entry's source path. With no sources, say
  so and propose a starter system; never invent project tokens.
- **Comments come from the folder, not from a tool**: `design/<slug>/comments.json` is versioned
  (`{"version": 1, "comments": [...]}`) with `id`, `file`, `selector`, `excerpt`, `comment`,
  optional `screenshot`, `createdAt`, `status` (`open` | `resolved`), `sentAt`, `resolvedAt`. Apply
  each open comment by editing only its artboard, using `selector` and `excerpt`; look at
  `screenshot` first. Resolve what you applied with `odachi design comments <slug> --resolve
  <id...>`. Never delete or reorder an entry; leave what you did not apply open, saying why.

## Never

- Never reference a remote asset: an artboard that needs the network is not a mockup, it is a
  page that breaks tomorrow.
- Never edit the project's design sources during a sync.
- Never mark a comment resolved that you did not apply.
- Never create a task document just to link a design: for small untracked work, leave `feature` out.
- Never present a design as implemented: the artboard is a direction, the code is the proof.
