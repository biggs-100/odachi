# AI policy

odachi is built by an agent following a written plan, and it is meant to be used
that way. AI-assisted contributions are welcome; the rules are about evidence and
ownership, not about the tool you use.

## What is required

- **Disclose it.** The pull request template asks how the change was produced.
  Say "with <agent or model>" or "by hand". There is no penalty for either, and
  no prize for hiding one.
- **A human owns it.** A person reads the diff, the proof and the review
  findings, and answers them. "The model wrote it" is not an answer to a finding,
  and an unreviewed agent output is not a contribution.
- **Evidence is real.** Never write a proof, a verdict or a command log by hand.
  The proof carries the hash of the tree that produced it, `--proof` refuses one
  that does not, and CI runs the command again. A fabricated green is the one
  thing this project exists to make impossible.
- **A test must be able to fail.** A test generated together with the code it
  tests proves nothing until it has been seen red for the right reason. If you
  cannot show the red, say so in the pull request.
- **No secrets in a prompt.** The repository, its issues and its proofs are
  public. An agent that reads the tree reads it as public, and a secret pasted
  into a prompt is a leaked secret.

## What is not allowed

- Presenting an agent's review as an independent human review.
- Fabricated command output, logs, issue reports or proofs.
- An agent pushing to `main` directly. The ruleset refuses it; the change goes
  through a pull request like any other.
- Using an agent to open issues or pull requests at a volume a person cannot
  read. The maintainer answers one at a time.

## What this buys you

The same thing it buys the maintainer: the agent does the writing, the human
does the deciding, and the gate keeps both honest. A contribution that follows
this policy is reviewed on the same terms as one written by hand: the proof, the
tests and the reasoning, in that order.
