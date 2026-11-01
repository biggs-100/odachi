---
name: design
description: Trigger: design a screen, mockup, landing page or UI directions; iterate artboards; what a person asked on one; design system sync. HTML under design/<slug>/.
---

# Skill: design

## When

The human asks, in plain words, for a screen, a mockup, a landing page or a few visual
directions, or iterates on artboards under `design/`. Also when they ask what a reviewer
wrote on an artboard, or to capture the project's design system.

Not for implementing production code in the project's stack: that is the handoff, and the
handoff is a step of its own.

## Rules

- **A design is a folder.** `design/<slug>/` holds `NN-<name>.html` artboards and
  `design.json`. Never overwrite another design: a new brief is a new slug.
- **Every artboard is one self-contained HTML file**: inline CSS, JS and SVG, system font
  stacks only. Never an external URL, CDN, remote font, image or script. It has to open
  offline in any browser.
- **Directions differ in strategy, not in colour.** Three by default, each with the tradeoff
  it makes written in the manifest. Three options that make the same tradeoff are one option
  painted three times.
- **Realistic content**, for the audience named in the brief. No lorem ipsum.
- **Accessible markup**: semantic elements and landmarks, WCAG AA contrast, visible focus,
  `prefers-reduced-motion` respected.
- **The manifest is the truth**: `{"title", "brief", "feature"?, "artboards":[{"file",
  "title", "tradeoff", "width", "height"}]}`, kept in sync with the files on every change.
  `feature` is the task document this came from, empty for untracked work.
- **Ask at most one clarifying question**, and only when the brief lacks the goal or the
  content.
- **Iterating edits only the targeted artboard.** Never rewrite the others.
- **Design system first when it exists**: read `design/system.md` and use only its tokens and
  component rules. Before finishing any artboard, list the colours and font families it uses,
  compare them with the system, and fix every deviation.
- **Capture the system by reading, never by writing**: the sources are CSS custom properties
  (`:root`, `@theme`, `[data-theme]`), the Tailwind config or v4 `@theme`, component folders,
  token files and the fonts they declare. Write `design/system.md` naming each entry's source
  path and `design/system.html` rendering the tokens. When there are no sources, say so and
  propose a starter system; never present invented values as project tokens.
- **Comments come from the folder, not from a tool**: `design/<slug>/comments.json` is a
  versioned file (`{"version": 1, "comments": [...]}`) with `id`, `file`, `selector`,
  `excerpt`, `comment`, optional `screenshot`, `createdAt`, `status` (`open` | `resolved`),
  `sentAt`, `resolvedAt`. Apply each open comment by editing only its artboard, using
  `selector` and `excerpt` to find the element; look at `screenshot` first when there is one.
  Mark what you applied resolved with `odachi design comments <slug> --resolve <id...>`, or by
  setting `status` and `resolvedAt` when the command is not there. Never delete or reorder an
  entry, and leave what you did not apply open, saying why.
- **The handoff is a decision, not a rewrite**: when an artboard is chosen, name it, name the
  other artboards as the alternatives they were, and implement against the project's real
  stack. Report `in_progress` when you start and `implemented` when the app runs.

## Never

- Never reference a remote asset: an artboard that needs the network is not a mockup, it is a
  page that breaks tomorrow.
- Never edit the project's design sources during a sync.
- Never mark a comment resolved that you did not apply.
- Never create a task document just to link a design: for small untracked work, leave
  `feature` out.
- Never present a design as implemented: the artboard is a direction, the code is the proof.
