---
name: security
description: Trigger: when the change reads input from outside, decides who may do what, holds a secret, adds a dependency, writes a log or handles a failure. Name the trust boundary before the code.
---

## When

When a request, form, file, URL, header or token enters the system; when a route
decides who may call it; when a secret, a session or a key appears; when a
dependency is added or upgraded; when an error is answered; before a release.

## Rules

- Name the boundary before the code: what arrives from outside, who may call it, what it can reach (OWASP A06:2025).
- Authorize on the server, on every request, against the object's owner - not because a button is hidden (A01:2025, the first risk of the 2025 list).
- A value from outside is hostile until parsed into the type you expect: validate at the edge, encode for the sink it lands in (A05:2025).
- Build queries, shell commands and paths from parameters, never concatenation: placeholders, an argv, a base path checked after the join (ASVS v5.0.0 chapter 1; 1.2.5 names the OS call).
- Use a library for sessions and passwords, argon2id or bcrypt for the hash, and never a scheme you invented (A04, A07:2025).
- A secret lives in the environment or the secret manager: not the repo, the fixture, the log or an error message (A02:2025).
- Every dependency needs a reason in the diff, a pinned version and its lockfile line; run `govulncheck ./...` and read what it reports, because it names only what your code can reach (A03:2025, new in this edition; go.dev/security/vuln).
- Fail closed: an error in a permission, a parse or a signature is "no", and the client's message never carries the internals (A10:2025, new in this edition).
- Log the event -who, what, when, which id- never the token, the password or the personal data (A09:2025).
- Bound what you accept: body size, page size, timeouts, retries with a cap. An input that never ends is an outage.
- Personal data: the minimum, a stated lifetime and a delete with a date - load `data` for the storage half.
- The proof is where the measurable half runs -the suite plus `govulncheck ./...`- and the fence is where a file that must never be touched lives. The rest of this skill is judgement.
- Before a release, read the diff once asking what it can leak, escalate or break; the `risk` lens asks a reviewer the same question.

## Never

- Trust a check the client can skip: a hidden field, a frontend validation, a role the request names.
- Copy a secret into a prompt, a log, a test or a commit body.
- Ship with `govulncheck` red and no decision written down.
- Return a driver error, a stack trace or a version number to the user.

## Where this comes from

- OWASP Top 10:2025, owasp.org/Top10 (read 2026-11): eight data-ranked categories and two voted in.
- OWASP ASVS v5.0.0, published 2025-05-30; ids read as `v<version>-<chapter>.<section>.<requirement>`.
- NIST SSDF SP 800-218 v1.1 (the four practice groups) and govulncheck, go.dev/security/vuln (read 2026-11).
