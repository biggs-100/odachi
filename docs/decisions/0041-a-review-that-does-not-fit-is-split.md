# 0041 - A review that does not fit is split

Status: accepted — Date: 2026-10 — Session: after v0.11.76

## Context

The review prompt is the whole diff. On 0132 -1,873 lines- it measured **114.4 KB**,
and the attempt before it, 113.7 KB, did not come back: `mimo-v2.6-flash` was still
running at the 600s odachi gives it. The project's limit existed -`max_prompt_kb`,
96 KB measured on 2026-10-28 with 89 KB closing in 46s and 122 KB stalling, raised to
120 KB by the human for that feature- and the tool did the honest thing: it refused
before spending.

The cost of that honesty is not paid by the tool. A feature big enough to be worth a
review is exactly the one the project's reviewer cannot take, so the human either
raises a limit they measured as unsafe or pays another model -and 0132's verdict was
paid by the model that built the tree, which is the review this tool exists to
prevent-.

The tool already ran more than one reviewer and merged: a project that assigns a
model per lens gets one reviewer per group and one verdict. Nothing in that path knew
about size.

## Decision

- **The split happens at the read, on the paths git reported.** `Split` hands the
  candidate out in parts, greedily and deterministically, each carrying its own
  paths, its own hunks and its number in the header. The limit is the project's;
  zero means no limit and never splits.
- **A part is a reviewer of its own, with the group's lenses and model**, and the
  store merges the answers into the tree's one verdict. Each part's inspection
  declaration is judged against the paths it read, and its coverage against those
  rows, so a part cannot vouch for bytes another part judged.
- **A part of the same run may answer lenses the store already holds.** `storeIn`
  takes `repeat` for exactly that case: it adds rows to the verdict it just wrote
  instead of refusing it as a double payment.
- **Each part's question is filed under its own name**,
  `<tree>.prompt-<lenses>-part-NofM.md`: one file holding only the last question is a
  lie about the others.
- **The one refusal left is a file that cannot fit by itself**, and it names the
  file, the size it costs and the limit. Splitting cannot help there, and a reviewer
  handed half a file is not reviewing it.
- **The prompt says it is a part** and tells the reviewer to judge only what it
  carries, so a part does not report the rest of the tree as missing.

## Alternatives rejected

- **Refusing, as before.** Honest and useless for the features that need a review
  most: the human's only ways out are a limit they measured as unsafe or a model
  they did not choose.
- **Asking the reviewer not to stall.** The timeout is not a suggestion the prompt
  can make; 600s is what the tool waits and 122 KB is where the model it pays for
  stops answering.
- **Splitting in the tool by a chunk size** (the first N kilobytes). A prompt cut
  mid-hunk asks about half a function; the diff's file boundaries are the only
  honest cut, and git already drew them.
- **One verdict per part.** The gate asks about a tree, and four verdicts for one
  tree is four answers to "did anybody review this" -the merge is what keeps one
  question with one answer.
- **A new key recording the split in the verdict.** The verdict's shape is the
  schema's, and a bump costs every project an upgrade for a fact a reader can see in
  the filed prompts. If it ever earns the key, it earns the bump then.

## Consequences

- A feature of any size is reviewed by the model the project chose, in parts that
  fit the limit it measured.
- A reviewer sees less than the whole candidate, and can miss an interaction between
  two parts. That is the price of the split, it is written in the prompt, and it is
  paid only when the size demands it.
- Four parts cost four calls to the same model: the same tokens roughly, and one
  verdict.
- The parts are visible in the store -each question under its own name- so a human
  can read what each reviewer was given.
