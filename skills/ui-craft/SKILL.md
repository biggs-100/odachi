---
name: ui-craft
description: Trigger: build or review a screen, landing page, dashboard, form or component; when it looks generic or AI-made; which numbers and which motion to use.
---

# Skill: ui-craft

## When

A screen, a landing, a dashboard, a form or a component is to be built or reviewed, or something looks generic. The rules stand alone: a deeper animation skill, if the machine has one, is optional and never required.

## Rules

**The order.** Content first: the real copy, the longest name, the biggest number. Then one thing wins the screen, and size, weight, colour and space carry the hierarchy in that order of cheapness. Layout from the numbers; breakpoints by content, never by device. Type: a scale, two families at most, one display size that earns its weight. Colour: one accent and a neutral ramp tinted with it, contrast measured and never estimated. Detail: one radius per level, one shadow per elevation, a 1px border at rest, one icon family. States: empty, loading, error, disabled and focus for every control, before the happy path. Remove: cut until it breaks, then put back one thing.

**The catalogue: the tell, and what replaces it.**

| the tell | replace it with |
|---|---|
| Everything centred: hero, three equal cards, section after section | One thing dominates; deliberate asymmetry |
| Purple or blue gradient, gradient text, glass everywhere | One accent, tinted neutrals, no gradient on text |
| Inter at every size, no scale, 120-character lines | The scale and the measure below |
| Spacing by guess: 13, 17, 22px | The 4/8 scale below |
| Shadow on everything, border plus shadow plus gradient | One elevation system; a border at rest |
| Emoji as icons, mixed sets, icons in coloured circles | One family, one stroke, 16/20/24 |
| "Supercharge your workflow", "seamless", "empower" | Verb and real object: the task, the user, a number |
| Hover-only affordance, no focus ring, modal without Escape | 24x24 minimum, keyboard, Escape, :focus-visible |
| A kicker above the heading, pills everywhere, "AI-powered" badges, fake charts | The heading carries its own weight; real data or an honest empty state |

**The numbers.**

| what | value |
|---|---|
| scale | 1.2: 12, 14, 16, 20, 24, 32, 40, 56 |
| body | 16-18px, leading 1.5, measure 45-90 characters (65 is ideal) |
| display | leading 1.1-1.2, tracking -0.02em at 32px and up |
| space | 4/8: 4, 8, 12, 16, 24, 32, 48, 64, 96 |
| radius | one per level; a nested one is the outer minus the padding |
| contrast | 4.5:1 for text, 3:1 for large text and UI edges |
| pointer | 24x24 minimum, 44x44 recommended |
| layout | 12 columns, 24 gutter, content 1200-1280, text 60-75ch |

**Motion.** First the question: feedback, continuity, hierarchy or story; if it explains none, it does not go. 100ms feedback, 150-200 exit, 200-300 enter, 300-500 for a large transition; ease-out to enter, ease-in to exit, ease-in-out to move, never linear except progress and loops. Nothing over 500ms without a reason a person would say out loud. Gestures get springs with 1:1 tracking and velocity handoff, and every animation is interruptible. One lead element, the rest staggered 20-50ms, at most two properties; layout uses FLIP and pages use View Transitions. Transform and opacity only; respect prefers-reduced-motion, and no parallax, scroll-jacking or autoplay carousel.

## Before you call it done

Build it, screenshot it, and read it against this page: the winner, the five states, the measured contrast, the tokens, the motion, the worst content.

## Never

- Ship a screen you have not looked at, or use motion to decorate a layout that does not work.
- Depend on a skill the project does not ship: this page is complete on its own.
