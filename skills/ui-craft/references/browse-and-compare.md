# Browse and compare

## The representation exposes the choice

The shape of the list is a decision about what the person compares.

| The need | The default | The exception |
| --- | --- | --- |
| compare several attributes | compact aligned rows or a table | cards, when the records differ in structure |
| judge appearance | previews with names and selection state | rows, when a thumbnail is enough |
| scan mixed summaries | cards with one shared structure | never a card per field |

Keep the identifier, the status and the action visible. An attribute that decides the choice does
not live behind hover, and nobody opens every record to compare it.

## Search and filters say what they cover

- Name the scope: `Search titles and authors`, not a bare magnifier.
- Show the sort and the count; expose active filters as removable chips and offer a reset.
- No results keeps the query and offers a way to broaden it. An empty collection is another state:
  show how to add the first item when the person can.

## Loading is a decision

| The pattern | It fits | It costs |
| --- | --- | --- |
| pagination | returning to a known portion | a visible page and the position restored |
| load more | sequential exploration | the list stays and the way back is offered |
| a virtualized list | many rows, rendering cost | keyboard, focus and item semantics kept |
| infinite scroll | a continuous feed | the position restored and the footer reachable |

Bound what is rendered; the data source is a separate decision. Choose with the real architecture,
never for a promise of speed.

## The return restores the place

- Encode the query, the filters, the sort and the page in the URL when the product can.
- Back from a detail restores the page, the scroll and the focus. If the item vanished, focus the
  results heading and say what changed.
- On a narrow screen: search first, filters behind a labeled control with a count, and the secondary
  fields below the name. The status and the action stay visible without hover.
