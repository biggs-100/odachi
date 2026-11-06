---
name: ui-craft
description: Trigger: build or review a screen, landing page, dashboard, form or component; when it looks generic or AI-made; which numbers and which motion to use.
---

# Skill: ui-craft

## When

A screen, a landing, a dashboard, a form or a component is to be built or reviewed, or something looks generic. The four references ship with this skill: load the one the situation names, and only that one.

## Rules

**The order.** Content first: the real copy, the longest name, the biggest number. Then one thing wins the screen, and size, weight, colour and space carry the hierarchy in that order of cheapness. Layout from the numbers; breakpoints by content, never by device. Type: a scale, two families at most, one display size that earns its weight. Colour: one accent and a neutral ramp tinted with it, contrast measured and never estimated. Detail: one radius per level, one shadow per elevation, a 1px border at rest, one icon family. States: empty, loading, error, disabled and focus for every control, before the happy path. Remove: cut until it breaks, then put back one thing.

## Decision gates

| Situation | Load |
|---|---|
| the screen looks generic and you need the tell and its replacement | `references/the-tells.md` |
| choosing a scale, spacing, radius, contrast or layout | `references/the-numbers.md` |
| search, filters, a large collection or a comparison | `references/browse-and-compare.md` |
| a transition, a gesture or anything that moves | `references/motion.md` |
| the last pass before calling it done | `references/the-read.md` |

## Output

The winner and why, the files changed, the five states covered, the contrast measured, the motion chosen or dropped, and what is left open. For a collection, the representation and the loading pattern chosen. A review returns one finding per defect with the fix it needs.

## References

- `references/the-tells.md` — the catalogue: what reads as AI-made, and what replaces it.
- `references/the-numbers.md` — the scale, spacing, radius, contrast and layout numbers.
- `references/browse-and-compare.md` — search, filters, the loading pattern and the return to the place.
- `references/motion.md` — when motion earns its place, and the timings that keep it calm.
- `references/the-read.md` — the last pass: the winner, the states, the measured contrast, the worst content.

## Never

- Ship a screen you have not looked at, or use motion to decorate a layout that does not work.
- Depend on a skill the project does not ship: this page and its four references are complete on their own.
