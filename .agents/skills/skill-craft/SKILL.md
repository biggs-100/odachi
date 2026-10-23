---
name: skill-craft
description: Trigger: when writing a new skill, fixing one that is not being used, or deciding whether something deserves a skill.
---

## When

When a rule repeats in every session, when a skill exists but the agent does not
load it, or when you are unsure whether to write a skill, a test or a fence glob.

## Rules

- An actionable rule can be **violated observably**: "a retry without backoff or
  an attempt cap" shows in the diff; "check resilience" does not.
- The trigger is the only thing that decides whether it loads. Write it as the
  human's situation -"when the diff touches permissions"- not as the topic
  -"security".
- One rule per line. If it does not fit in one line, it is two rules.
- What the model already knows (syntax, frameworks, fashionable patterns) does
  not go in: it takes context and gets skipped.
- What cannot be guessed -your conventions, your decisions, what is forbidden-
  goes in `brief.md` or the fence, not in a skill.
- If the rule can be verified, prefer a fence glob or a test **before** a skill:
  the fence is enforced by the binary, the skill can be ignored by the model.
- To create one: `odachi skill new <slug>` leaves the template, and
  `odachi skills --check` says whether it is well formed.

## Never

- Glossaries, encyclopedias or technology history.
- Rules only someone who already knows can understand.
- Copy a whole skill without adapting it to this project.
- Pass 60 lines: if it does not fit, it is a document, not a skill.
