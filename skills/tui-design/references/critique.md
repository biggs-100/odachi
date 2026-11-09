# The critique

## Judge from the frame

Write the design assessment first, from the captured frame, and only then check the measurable
evidence —overflow, contrast, render counts. In the other order the numbers anchor the judgment and the
critique becomes a linter with an opinion.

## The score

Each heuristic scores 0-4: 0 broken, 2 workable, 4 excellent. A heuristic that does not apply is marked
`n/a` and the total is renormalised to 40.

1. **System status** — is live work, progress, mode and permissions visible?
2. **Developer expectations** — does it follow the IDE and terminal conventions the user already has?
3. **Control** — can the person collapse, scroll, escape, undo a view change?
4. **Consistency** — glyphs, roles, section shapes and terms, the same everywhere?
5. **Error prevention** — are dangerous modes visible, and confirmations where they matter?
6. **Recognition over recall** — are hints next to the action, and is the palette discoverable?
7. **Expert efficiency** — keyboard paths, density, no wasted row?
8. **Minimalism** — each fact once, no decoration, no second way to say the same thing?
9. **Recovery** — does a failure name the cause and the next step?
10. **Help** — are the keys and the commands discoverable from the interface?

**Bands**: 36-40 excellent, 28-35 good, 20-27 acceptable, 12-19 poor, under 12 critical.

## Cognitive load

Single focus, chunks of four or fewer, visible grouping, one decision at a time, four options or fewer
per decision, nothing to remember across a screen, progressive disclosure. One failure is low load;
four or more is critical, whatever the score says.

## The three readers

- **The power user** lives on the keyboard: wants density, no repetition, no confirmation for what is
  reversible.
- **The keyboard-only reader** wants contrast, a visible focus and no state that only colour reports.
- **The stress test**: a 12-hour session, fifty agents, two hundred tasks, a 80-column terminal.

## Severity

- **P0** — the task cannot be done, or the interface breaks: a panel disappears, content is unreadable.
- **P1** — significant difficulty or a contrast failure. Tie-breaker: would a user report it? Then P1.
- **P2** — a workaround exists.
- **P3** — polish.

## The report

1. The target, the widths captured, the frame paths.
2. The score table and its total.
3. **The specificity verdict**: could any other TUI ship this screen unchanged? If yes, the design has
   no point of view.
4. What works, two or three things.
5. The priority issues, three to five: `[P?] what / why it matters / the fix / the command`.
6. The red flags per reader.
7. Minor observations.
8. Two to four questions with concrete options, last.

A critique that lists twenty equal findings hides the one that matters. Five findings with severities
beat twenty without.
