# UI/UX verification an agent can actually do

Read this before scaffolding any project with a screen. It answers one question: **how does
an AI agent verify that the UI/UX is correct**, given it has no eyes, no hands and no taste.

The stack packs supply the mechanical ladder (`jsdom → Playwright → WebdriverIO`, `CDP`,
`axe-core`). **This file states the contract those tools must satisfy**; a ladder that does
not meet it produces screenshots nobody asserts on, which is the most common false comfort
in AI-driven frontend work.

## Why agents fail at this (the diagnosis, not the excuse)

1. **The UI is the only layer without a typed contract.** A function has a signature; a
   screen has "the money must be visible without scrolling at 1366×768". Agents test what
   they can assert, so the unasserted layer stays unverified.
2. **The interesting states are unreachable.** The empty state, the refusal, the 60-character
   name, the four-digit price. If the running app cannot be *driven* into a state by data,
   nobody — human or agent — can verify it, and the screenshots only ever show the happy path.
3. **Coverage has no denominator.** "We tested the UI" without a route inventory is a claim,
   not a measurement. The inventory is the denominator.
4. **Screenshots are treated as evidence.** A screenshot nobody asserts on is a picture. The
   evidence is the packet: the image **plus** the machine-readable facts beside it.

## The four requirements (a harness that misses any of them does not count)

1. **Drivable.** One command boots the real app against a seeded, disposable data store, and
   reaches every screen in every interesting state. No mocks at this level: the point is to
   see the real composition, the real database and the real rendering.
2. **Observable, in machine terms.** Each capture yields, beside the image: the rendered DOM,
   the accessibility tree (roles, landmarks, the outline), the computed geometry of the
   elements that matter, the console log, and the network failures. The agent reads these
   fields; the human reads the image.
3. **Inventoried.** A file lists every route **and the data that makes each state appear**
   (empty, refusal, long text, big numbers, permission, error). The run reports
   `routes × states × widths` in the packet, so coverage is a number, not a feeling.
4. **One command, one exit code.** `./verify-ui --routes all --widths 1366x768,390x844`
   exits non-zero on a violation and writes a packet. The agent runs it exactly like a test;
   nothing requires judgment to know whether it passed.

## What IS assertable (and what is not)

| Assertion | Mechanism | Example from a real project |
| --- | --- | --- |
| The exact copy in a refusal, an empty state or a confirmation | read the served HTML | "La factura necesita un cliente con RUC…" asserted literally in the packet |
| Structure and landmarks | a11y role counts + outline | exactly one `aria-current="page"`; one alert with three list items; a table with a header row |
| Accessible names and labels | axe-core (`label`, `aria-allowed-role`, `listitem`) | a `role="alert"` on a `<ul>` silently orphaned every `listitem`; axe caught what no unit test could |
| Contrast, target size | axe-core rules | 314 `target-size` warnings in one packet: measured, visible, not a failure |
| Geometry of the critical element | computed styles at the target viewport | a sticky money band at the bottom **with a page taller than the viewport** |
| No horizontal scroll / no overlap | scroll width vs client width; rect intersection | a 390 px capture that was really a 500 px layout cropped |
| The thing that must be visible first | bounding rect inside the first viewport | totals, cash, change and the pay button above the fold on a counter screen |
| Keyboard reachability | focus order + `tab` drive | the primary action of a ticket reachable without a mouse |
| State after an action | re-read the store or the screen | a refused form left the database and the counters unchanged |
| Console and network silence | console log + failed requests | the class of defect no Go/TS test sees |

**Not assertable, do not pretend otherwise:** beauty, hierarchy, "does this feel right",
which of two valid designs is better. Those go to the human — with a packet that makes the
review cheap.

## The capture packet (the only acceptable evidence)

One directory per run, timestamped and reproducible, containing per route and width:

- `*.png` — the full page, at the real target width (see the traps below);
- `*.html` / text snapshot — the rendered markup;
- `*-ax.json` — role counts and the filtered outline (links, headings, landmarks);
- `*.json` — the geometry and the computed styles the assertions read;
- `console.log` / `network.json` — errors and failed requests;
- `summary.json` — routes × states × widths run, violations, warnings, and the exact
  command line that produced it.

A run that cannot be reproduced from the packet is not evidence. Pin the seed, freeze the
clock, record the build id.

## The agent's loop

1. Run the harness. Read `summary.json` first — it is the verdict.
2. For each violation, read the machine fields (DOM node, rule, message), not the prose.
3. Look at the screenshots **as images** for the judgment calls the machine fields cannot
   make: overlap, truncation, an empty hole, a number that reads wrong.
4. Fix, re-run, and attach the packet to the work unit. The diff is the change; the packet
   is the proof it renders.

## The human gate (do not automate it away)

Automated checks catch structure, copy, contrast, geometry and silence — not taste. So the
deliverable of a UI unit is **a small packet a human can review in minutes**, not a claim
that "the UI is verified". The agent's job ends at "0 violations, N warnings, here is the
packet"; the human's begins at "does this read well for someone standing at a counter".

A warnings-only packet is normal (target size, color contrast on decorative text). Say the
count; never call it clean.

## Traps, each measured

- **The viewport lies before the layout does.** A browser with a minimum window width renders
  500 px when asked for 390; the "mobile" screenshot is a crop. Verify the real width (an
  iframe of 390 px, or the device metrics override) and assert the measured viewport in the
  packet.
- **A sticky element cannot be verified in a short page.** With content that fits, the sticky
  band sits in its natural place and proves nothing. Force a page taller than the viewport.
- **`axe-core` absent = 0 failures, not a clean run.** Vendor it on disk (e.g.
  `.verify-ui/axe.min.js` + an env var) so the a11y layer is never silently off. Say in the
  summary whether axe ran.
- **A green run without the error states walked is not coverage.** The shared error list was
  only ever rendered in one route; two error-state routes had to be added to the inventory.
- **The rendered page and the form submission disagree in edge cases.** A `selected disabled`
  option is displayed but NOT submitted (measured in Chrome: `FormData` carries no entry).
  When a control's value drives a server decision, assert what the *form posts*, not what the
  screen shows.
- **Generated markup churns diffs.** A template edit produces hundreds of changed lines in the
  generated file; never count them as authored size, and never review them line by line.
- **Non-ASCII through argv, `%` through URLs.** Encoding bugs in the harness look like app
  bugs. Pass bytes through files, and assert the digest of what the server actually received.

## Scaffold deliverable

A project with a screen is not scaffolded until it has: the route/state inventory, the
harness with its one command, the packet format, and the axe layer wired. Put all four in
`docs/change-checklist.md` so every later change runs them.
