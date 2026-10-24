# Security policy

## Reporting a vulnerability

Report it privately through GitHub Security Advisories:

<https://github.com/biggs-100/odachi/security/advisories/new>

Do not open a public issue, and do not paste a proof of concept in a pull
request. If GitHub advisories are not available to you, open an issue that says
only "I need a private channel for a security report" and nothing else; the
maintainer will answer with one.

## What is in scope

- **The binary and its artifacts:** `brief.md`, the task doc, `proof/`, the
  fence, `.odachi/reviews/`, and the ruleset script.
- **The gate's trust model:** the fence is read from the base ref, the gate is
  built from the base ref, `main` only takes a pull request whose four checks
  passed, and a candidate cannot loosen the fence it must obey.
- **The workflows** in `.github/workflows` and the ruleset in
  `.github/main-ruleset.json`.

The interesting class is a change that gets a green gate it did not earn: a
proof that is not from this tree, a fence that is obeyed from the worktree, a
check that reports without running, a rule that looks enforced and is not. That
is the defect this project exists to make impossible; a report of it is always
welcome.

## What is not in scope

- The third-party agents and models that run odachi. Report those to their
  vendor; a model answering badly is not a vulnerability of this tool.
- A repository that removed the ruleset, edited `CODEOWNERS` or no longer runs
  the gate: the protection is the configuration, and a fork can drop it.
- The content of a project's own brief, fence, task doc or proofs.
- Denial of service by running an expensive command in `--run`: the command is
  the user's.

## What to expect

An acknowledgement as soon as the maintainer reads it, and a target of seven
days for a first answer. A fix ships with a test that fails without it, and the
advisory names the versions affected. Credit is given unless you ask otherwise.
