# The surface a model is inside

Every product now has a model in the loop, and the interfaces that work are not chat
panels bolted to the side. A side panel teaches people that the model is a second
product; the pattern that ships puts it inside the workflow and shows its work. The
states and their sentences belong to `states-and-copy.md`; the markup and the keyboard
to `accessible-and-adaptive.md`; this page is what a model adds.

## Four things a model does, and what each owes the screen

| what it does | what the screen owes | the failure it replaces |
| --- | --- | --- |
| answers, token by token | a stable frame, the text arriving in the space it will occupy, an end state | a spinner over the whole page, then a jump when the text lands |
| runs a tool | the tool's name and the object it touches, one line per step, openable | a black box between the question and the answer |
| writes into something | the change it made, visible in place, and an undo | an edit nobody saw happen |
| takes an irreversible action | a preview of the exact effect, a confirmation that names it, an undo | an agent that pays the invoice |

## Show the work, at a budget

- One line per step, not a diary: the person needs to know what happened to their
  data and what is still running, not how the model reasoned.
- A tool call says which tool and which object, and opens to its arguments on demand.
  "Checking your calendar" is a line; the request it made is a disclosure.
- Waiting has three zones: below a tenth of a second nothing is shown because nothing
  was noticed; around a second the work keeps its frame and the person keeps their
  place; past ten seconds attention is gone, so the wait needs a name, a phase and
  the ability to leave and come back.
- Never a generic spinner over content that is already on the screen. The region that
  is working says so, and the rest of the page stays usable.
- A failure says what was not done, not that something failed: "The draft was not
  saved, the connection dropped" leaves the person with a state they can act on.

## Generative UI, in the only three forms that exist

- **Soft**: the surface adapts to the role or the history -the empty state names this
  person's next step, the dashboard opens on the question their role asks first.
- **Medium**: an input produces a working interface instead of a paragraph -a form
  built from "file a bug report with a screenshot", a filtered view from a sentence.
- **Strong**: the whole surface is composed per session from primitives.

All three are designed as **primitives and the rules that assemble them**, not as
screens, because a screen assembled at runtime was never mocked. And the honest bound:
a table of records does not need to be generated. Reserve composition for the moment
the person's intent is genuinely open, and ship a fixed surface where the shape of the
job is known.

## Inline, and still a draft

- The output lands where the work is, editable, with the input still visible. An
  answer that replaces the question costs the person their own words.
- Anything generated is a draft until a person accepts it, and the accept is a normal
  action of the product, not a second product's button.
- The prompt is not the interface: no screen is a text box with an example and no way
  to do the same thing by hand.

## Approval is a design, not a dialog

- The preview names the effect in the person's units: the amount and the recipient,
  the file and the folder, the pull request and its base branch.
- "Irreversible" is the product's own list -pay, send, delete, publish, deploy- and it
  is the same list every screen reads. A confirmation on a reversible step trains
  people to click through the one that matters.
- The undo survives the session and says what it will restore.
- An action taken on somebody's behalf says whose, because a model acting as the
  person is still acting on a person's data.

## What this page does not decide

The sentence each state says is `states-and-copy.md`'s; the contrast, the targets and
the keyboard are `accessible-and-adaptive.md` and `the-numbers.md`; the density and the
hierarchy are `grouping-and-emphasis.md`.
