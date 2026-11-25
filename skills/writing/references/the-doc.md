# The document

A document is read to act: a guide, an RFC, a README, an onboarding page. The
reader did not live the conversation that produced it, so the page is the whole
context.

## The shape

1. A title that names the outcome, not the topic: "How a feature closes", not
   "Process".
2. One paragraph: what it is for, who reads it, and why it matters.
3. The quick path: the shortest sequence that works, numbered, with the command
   or the link that proves each step.
4. The details: one section or one table per decision, never one long story.
5. A checklist the reader can confirm without you.
6. The next step: the link or the command that continues.

## The rules behind the shape

- Answer first: the decision, the action or the result, before the context.
- Progressive disclosure: the happy path, then the edge cases, then the
  references. A reader who stops early still knows what to do.
- Chunking: one topic per section, and short lists stay short.
- Signposting: headings that say the content, tables for comparisons,
  checklists for acceptance.
- Recognition over recall: a table or a template beats prose the reader has to
  remember.
- Review empathy: a document that goes to a reviewer says what to read first,
  what is out of scope, and which decision each section carries.

## The review path

When the document goes to a reviewer:

- Say what to review first and what is intentionally out of scope.
- Link the previous and the next PR when the work is chained.
- One decision per section, and a checklist for the acceptance criteria.

## Verify the artifact

A source file is not the delivered artifact. Walk the ladder and record each result:

1. **Source correctness:** spelling, links, commands, code samples and claims match the
   reviewed source.
2. **Rendered or exported artifact:** open the actual render or export; check formatting,
   navigation and embedded media.
3. **Real destination:** open the published URL, delivered file or installed copy; confirm
   the intended version and real destination behavior.
4. **Applicable behavior:** exercise the states, controls and keyboard paths the artifact
   exposes; measure contrast and zoom for a screen.
5. **Readback:** compare the delivered result with the source and record what was observed,
   not what the build printed.

If a rung cannot be run, report `not verified` and why; never turn it into a pass.

## Provenance

For quoted, external or generated content and assets, record provenance with these fields:

| field | record |
| --- | --- |
| creator or source | person, organization or tool that made it |
| URL or local path | exact source locator and the retrieved or generated copy |
| usage rights | license, permission or restriction for this use |
| version or date | version, revision or retrieval date |
| status | original, generated or demonstration/demo |

## Commands

```sh
git diff --name-only -- '*.md'
```

## Never

- A wall of prose with the point in the third paragraph.
- A tutorial where the reader needed a reference.
- "See the code": a document that sends the reader elsewhere has no reader.
- A checklist that only the author can confirm.
