# The comment

A comment lands in the middle of somebody's work. It has one job: move the work
one step, without costing them a meeting.

## The formula

```text
<the observation or the request>

<why it matters, only if it does>

<the next action, concrete>
```

One to three short paragraphs, or a tight list. Start with the actionable point:
do not recap the whole change before the feedback.

## The tone

- A thoughtful teammate, not a corporate bot: direct about the work, warm about
  the person.
- Explain the why when you ask for a change. The reason is what turns a request
  into something the other person can decide, instead of a preference.
- Comment on the highest-value issue, not on every small preference. A pile-on
  costs more than the nits it fixes.
- The language of the thread: a Spanish thread gets Spanish, an English thread
  gets English, a mixed context gets the language of the message. If the human
  asked for a language or a tone, that wins. A comment is a message to a person,
  not an artifact: the artifacts stay English.
- No em dashes: commas, periods or parentheses.
- Do not sign with a tool, a model or a title nobody asked for.

## Examples

Request a change:

```markdown
Good approach overall. I would split the validation from the UI wiring into its
own commit: it keeps the review narrow and the rollback clean.

The tests for the validation can travel with it, so the split does not leave a
red step behind.
```

Approve with a note:

```markdown
Approved. The scope is clear and the change is contained.

For the next one, link the previous and the next PR so the chain stays
navigable.
```

Ask for a split:

```markdown
This change passes the 400-line budget, so it needs the split: foundation and
tests first, then the integration, then the docs. Each one reads on its own,
and the first one can land while the others are reviewed.
```
