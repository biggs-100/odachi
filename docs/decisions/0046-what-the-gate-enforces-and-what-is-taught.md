# 0046 - What the gate enforces, and what is taught

Status: accepted - Date: 2026-11 - Session: after v0.11.90

## Context

The question was practical: what must a fullstack developer weigh before shipping so the
project does not end in a security problem -input, who may do what, secrets, dependencies,
logging, personal data, migrations, backups, least privilege, failure modes- and what should
be written down about it.

The research says the same thing in several voices. OWASP Top 10:2025 puts broken access
control first and replaced "vulnerable and outdated components" with software supply chain
failures; its tenth category, mishandling of exceptional conditions, is new, and it is the
one a beginner meets daily and never sees in a scanner. ASVS v5.0.0 (published 2025-05-30)
turns the same ground into verifiable requirements. NIST SSDF SP 800-218 v1.1 organizes it
as four practice groups across the whole lifecycle. And govulncheck, the Go tool, names only
the vulnerabilities a program can actually reach, which is why it is useful and why it cannot
be the whole answer: it sees dependencies, not decisions.

The temptation is to put a scanner in `odachi doctor` and call the project secure. That is
the defect this tool exists to prevent, in another costume: a green that means "a list was
checked" instead of "somebody decided something". A pattern list measures the pattern.

## Decision

- **The gate enforces what a file and a command can prove.** Three halves, and all three
  already exist:
  - **the fence over paths**: a file the change must not touch, or a dependency it must not
    add, read from the base ref so the candidate cannot loosen it (0022, 0027);
  - **the proof over a command the human chooses**: the suite, `govulncheck ./...`, `go vet`
    -whatever the project decides is evidence- tied to the tree it ran on;
  - **the risk lens over the diff**: the review reads the change for the eighteen patterns
    it can see, and a causal `BLOCKER` stops the ship (0043).
- **Everything that is judgement is taught, not enforced.** The trust boundary, who owns the
  object a request names, what a secret may reach, how long personal data lives, whether an
  error should be visible, what failing closed means for this endpoint. The skill `security`
  says it once, in the moment a person writes the code, and names the source and version of
  every claim it makes about an outside system.
- **No scanner and no new doctor check in the first version.** If one lands it is its own
  feature, with its own fence and a test that shows the false positive it accepts; a list
  that fires on fixtures teaches people to ignore it.

## Consequences

- A project that wants the measurable half runs it in the proof, and the proof may name a
  command the tool has never heard of. That is the point: odachi does not curate the list of
  tools, it ties whatever ran to the bytes it judged.
- A project that ignores the skill is still gated by the fence, the proof and the lens. That
  is weaker, and it is honest about which part was enforced.
- The skill carries versions and dates (`ASVS v5.0.0`, `2025-05-30`; `Top 10:2025`;
  `govulncheck`, read 2026-11) because a security claim about the outside world ages like a
  dependency, and an undated one is already false without anyone knowing.
- The review's `risk` lens and the skill are two readers of the same knowledge, in the two
  directions that matter: the skill asks before the code, the lens asks after the diff. They
  must not drift into two lists; the ADR is the place that says which one owns a rule when
  they disagree.
