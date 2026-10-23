---
name: skill-improve
description: Trigger: when a skill is not used, grew too long or the agent does not load it. Audit the trigger, cut the body and validate the format again.
---

## When

When a skill exists but the agent does not load it, when its body passed 60
lines, when it repeats another skill, or when a rule stopped being observable.

## Rules

- Start with the trigger: if it does not say the human's situation, the agent
  will not load it. Rewrite it as "when..." and not as a topic.
- One rule per line. If a rule needs two, they are two rules; if it cannot be
  violated observably, it is not a rule: it is an intention.
- What the model already knows does not go in. What the project decided does.
  What can be verified with a test or a fence glob goes to the test or the fence
  before the skill.
- If the body passed 60 lines, move the detail to a document and leave the
  trigger and the hard rules in the skill. Do not split it into two skills
  without a distinct loading reason.
- If it repeats another skill, pick one owner and delete the copy. A rule with
  two owners drifts.
- After editing: `odachi skills --check` and `odachi skills --write` if the
  trigger or the name changed. The registry is generated, not edited by hand.
- If the skill is not used, test the trigger before adding more content: the
  problem is usually when it loads, not what it says.
- Measure before and after: lines, rules, and whether the agent loaded it in a
  real session.
- If the rule is one-off, it is not a skill: it goes to `brief.md`, the fence or
  a test.

## Never

- Add rules "just in case".
- Copy a whole skill from another project without adapting the trigger.
- Leave a skill with two owners or an empty section.
- Pass 60 lines and call it "documentation".
- Touch `SKILLS.md` by hand.
