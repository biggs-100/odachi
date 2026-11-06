---
name: design
description: Trigger: design a screen, mockup, landing page or UI directions; iterate artboards; what a person asked on one; design system sync. HTML under design/<slug>/.
---

# Skill: design

## When

The human asks for a screen, mockup, landing page or visual directions, iterates on artboards, asks what a person wrote on one, or captures the design system. Not for production code: that is the handoff.

## Rules

- **A design is a folder.** `design/<slug>/` holds the artboards, their screenshots, `design.json` and, when the canvas collected them, `comments.json`; `references/files.md` has every format. Never overwrite another design: a new brief is a new slug.
- **Every artboard is one self-contained HTML file**: inline CSS, JS and SVG, system font stacks only. Never an external URL, CDN, remote font, image or script; it has to open offline.
- **Every artboard carries the screenshot it was read against**: a file named in the manifest's `screenshot`, taken after the last change. A screen nobody looked at is a screen nobody checked; `odachi design list` names the ones without it.
- **Before designing, ask the folder**: `odachi design list` says whether `design/system.md` is missing, stale or fresh; sync before the first design and name a stale source before following old tokens.
- **Open the canvas after changes**: `odachi design open <slug>` returns at once and leaves a detached canvas; a second call moves the open page. Report its URL unless the human said not to.
- **Comments arrive as session messages when the OpenCode plugin is installed** (`odachi init --global` writes it); without it, or on Claude Code, read `design/<slug>/comments.json` and apply the `open` entries, saying which path you took.
- **The handoff records what happened**: `odachi design handoff <slug>/<artboard>` marks the chosen artboard `sent` and logs it in the linked feature; `--status in_progress` and then `--status implemented --preview http://localhost:<port>` (loopback: the canvas frames it).
- **The handoff is a decision, not a rewrite**: name the chosen artboard, name the others as the alternatives they were, and implement against the project's real stack.
- **Link a tracked feature both ways**: when the design comes from `odd/tasks/<feature>.md`, set `feature` in the manifest and append one `L` entry to that document's log naming `design/<slug>/` and its artboards; untracked work leaves `feature` out.
- **Directions differ in strategy, not in colour.** Three by default, each with the tradeoff it makes; three that make the same tradeoff are one option painted three times.
- **Pair with `ui-craft` before choosing directions**: it owns the hierarchy, the numbers and the tells; this skill owns the delivery. Read only the parts the brief needs.
- **Before reporting an artboard, read it with `ui-craft`**: `the-read` for every artboard, and `adaptive-interfaces` when it has states; fix every gap before saying it is done.
- **The manifest is the truth**: `design.json` names every artboard and its screenshot, and it is kept in sync with the folder on every change.
- **Real content and accessible markup**: the audience's copy, no lorem ipsum; landmarks, WCAG AA, visible focus, `prefers-reduced-motion`.
- **Ask at most one clarifying question**, and only when the brief lacks the goal or the content.
- **Iterating edits only the targeted artboard.** Never rewrite the others.
- **Another approach on the same brief is another artboard**: add the next `NN` and keep the others; a new slug is a new brief, never a new direction.
- **When the human wants to tune values, put the controls in the artboard**:
  sliders bound to its CSS custom properties for spacing, colour and type scale,
  plus a `Copy as prompt` button that copies the chosen values as a prompt. The
  button copies text; the artboard never calls a network.
- **Design system first when it exists**: read `design/system.md` and use only its tokens and component rules; before finishing an artboard, compare its colours and families with it.
- **Capture the system by reading, never by writing**: `odachi design sync` reads CSS custom properties, the Tailwind config, components, token files and fonts, and writes `design/system.md` and `design/system.html` with each entry's source. No sources: propose a starter, never invent.
- **Comments come from the folder, not a tool**: apply each open entry of `design/<slug>/comments.json` to its artboard with `selector` and `excerpt`, and resolve it with `odachi design comments <slug> --resolve <id...>`; never delete or reorder, leave the rest open.
- **A comment's screenshot is read first**: when the entry carries `screenshot`, it shows the transient state the static artboard does not.

## Output

The slug, each artboard with its tradeoff and the read it survived, the files changed in this turn, the comments resolved, and any open question. The canvas URL when it was opened, and the chosen artboard when it was handed off.

## References

- `references/files.md` — the design folder, the manifest and the comments format.

## Never

- Never reference a remote asset: an artboard that needs the network is not a mockup, it is a page that breaks tomorrow.
- Never edit the project's design sources during a sync.
- Never mark a comment resolved that you did not apply.
- Never present a design as implemented: the artboard is a direction, the code is the proof.
