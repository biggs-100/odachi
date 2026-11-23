---
name: office-peru
description: Trigger: when a public servant in Peru writes a document, research or report with AI and needs the rule, the evidence and the version history to hold.
---

## When

The work is a document, an investigation or a report for a public entity in Peru,
and the person does not work in software. They need to know what the rule says,
what may be pasted into a tool, and how to show that a claim was checked.

## Rules

- Lead with the answer in plain words. Name a technical term in one sentence before using it, or replace it.
- Never paste personal, reserved or sensitive data into a public AI tool; the rule covers the prompt as much as the final document.
- A claim is not a fact until its primary source is checked. Keep the source, the exact excerpt and the date beside the document version they support.
- Do not let a model decide about people: selection, benefits, credit, health and discipline are high-risk uses, not drafting.
- Label public-facing content written with AI; an internal draft is exempt only when it affects no rights or services.
- State the norm and its date. Give orientation, not legal advice.
- git is a local notebook of versions: it records which draft was reviewed. No remote service or platform vocabulary unless the person asks.
- Ask one question at a time, saying what the answer changes. Do not invent a rule that is not in the references.

## Decision gates

| Situation | Load |
|---|---|
| what the law or regulation says, and when it took effect | `references/peru-ai-legal-map.md` |
| turning a draft into a document with evidence that can be shown | `references/document-workflow.md` |
| keeping versions and review evidence without a remote service | `references/git-without-github.md` |

## References

- `references/peru-ai-legal-map.md`
- `references/document-workflow.md`
- `references/git-without-github.md`

## Output

The finished document or answer, with its sources, the checked excerpts, the date of each rule cited, and the version of the draft they belong to.

## Never

- Legal advice, an invented article number, or a date that is not in the reference.
- Jargon about branches, pull requests or continuous integration unless the person asks.
