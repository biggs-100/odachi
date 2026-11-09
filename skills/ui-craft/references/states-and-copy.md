# States and copy

## The six states, with what to say

A state is not a blank panel with a spinner. Each one has a sentence, and the sentence is the design.

| the state | show | say | never |
| --- | --- | --- | --- |
| empty collection | what belongs here and the first action | `No bookings yet` + `Create booking` | "No results" when nothing was searched |
| no results | the query or filters in force, and a way to broaden | `No bookings match "lathe"` + `Clear search` | erasing what the person typed |
| loading | what is loading, in the space it will occupy | `Loading reservations…` | a blank screen, or a spinner over the whole page |
| partial or stale | what is incomplete, or when it was read | `Updated 3 minutes ago · 40 of 120 shown` | an old total presented as current |
| error | what failed, whether it was saved, what to do | `Reservation not saved. Try again.` | `Something went wrong` |
| zero | the real number | `0 bookings this week` | inventing a value, or reading a missing one as zero |

- On a failed save, keep the input. Say whether a retry is safe, because a retry can double the charge.
- Never claim success before the result is known: the optimistic update is a design, and it needs the
  sentence for the case where it was wrong.
- A disabled control says why, and where the path to enable it is. A tooltip is not where a person looks.
- Announce a status change to assistive technology once, in context. Repeating an unchanged message is
  noise that gets muted.
- The empty state is a designed screen: a heading, one sentence, one action. It is the first screen a new
  project sees, and it is the one most often left to a default.

## The copy is part of the state

- The sentence names the object and the outcome: "Reservation not saved", not "Error".
- Two sentences at most. The first says what happened, the second what to do.
- No exclamation marks, no "oops", no jokes on a failure: the person's data is at stake.
- Numbers carry their unit and their time: "$4.2M, Q3" and "3 minutes ago", not "$4.2M" and "recently".
- In a list, the count is part of the state: "40 of 120" answers the question an empty panel creates.

## What this file does not decide

The layout of a list and the choice of a loading pattern are in `browse-and-compare.md`; the hierarchy and
the measured contrast in `the-numbers.md`; the content itself in `task-and-content.md`.
