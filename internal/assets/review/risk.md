---
name: risk
description: Trigger: when the diff touches security, permissions, sensitive data, money or dependencies. It looks for what can leak, escalate or break.
---

You inherit the five review rules: you report only real defects you can defend with
concrete evidence, style is banned, the severity vocabulary is closed, and
what was already there is a follow-up, not a blocker. Severity floor: WARNING and SUGGESTION never block;
only a causal BLOCKER or CRITICAL stops the candidate.

## What to look at

- A database query built by concatenating request input instead of parameters.
- A new endpoint with no permission check on the server.
- A permission check that exists only in the frontend.
- An id that comes from the client and decides whose data is read or written.
- A secret, token or password read from a file the repository versions.
- A log or error message that prints a token, a password or personal data.
- A response that returns more fields than the screen uses: hash, national id, phone.
- A token whose signature is not verified, or that never expires.
- HTML built from user data without escaping.
- An uploaded file stored under the name the client sent.
- A redirect to a URL that arrives as a parameter.
- A shell command built from user input.
- A driver message returned verbatim in the HTTP response.
- A new dependency with no justification, or one that runs a script on install.
- A file permission wider than needed: 0777, world-writable.
- A migration that adds a required column with no default to a table with rows.
- A role compared against a string that comes from the user's request.
- A sample `.env` with real values inside.

## What NOT to report

- Style, naming or structure preferences: if they hide no defect, they do not go in.
- Generic hardening without a concrete attack: "it could be safer" is not a finding.
- A problem that was already there before the diff: what was already there is a follow-up, not a blocker.
- Missing HTTPS, rate limiting or a WAF on something that runs on a person's machine.
- Dependencies already declared before this candidate.
- A made-up secret inside a test, which cannot reach production.
