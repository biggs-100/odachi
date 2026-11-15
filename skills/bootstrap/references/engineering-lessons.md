# Engineering lessons that hold for any project

Each rule below was paid for by a real defect in a real project (Go + SQLite + HTMX +
SUNAT e-invoicing). They are stack-agnostic: the mechanics change, the failure does not.
Read this when scaffolding, and re-read it when a unit feels "obviously fine".

## 1. The artefact beats the test

**Rule.** For any format, protocol or dataset owned by a THIRD PARTY, put one **real sample**
in the repo before designing the parser, and run the real thing **once** before calling the
unit done. Fixtures written from documentation are a hypothesis.

*Evidence.* A padrón parser passed its fixtures and died on the real file at row 15.9 M: the
format escapes a pipe inside a value (`\|`) **and** a backslash (`\\`). Reading the raw bytes
of two anomalous rows settled a design question that no amount of test-writing could. The
same project shipped a wrong XML five times while the whole suite was green; what caught it
each time was comparing against a real accepted document.

**Apply.** One real artifact in the repo, one run against the real source, and the unresolved
details named. The fixture that matters is the ugly real row, verbatim.

## 2. What a value must be to be STORED is not what it must be to be LOADED

**Rule.** A validator that runs on the read path as well as the write path turns every new
rule into a possible **data lockout**. Either split the two (`Validate` / `ValidateStored`),
or record an ADR that says the double duty is intentional and list the rules that may enter it.

*Evidence.* A rule that reads naturally ("a factura needs a RUC acquirer") would have made
every factura stored before it — with a DNI or no acquirer — unreadable, because the
repository validates each row it loads. It had to live outside `Validate`, with a test pinning
that exemption so a later "fix" cannot move it back.

**Apply.** Ask of every new rule: "can this refuse a row that is already in the database?"
If yes, it belongs to the write path only, or the read path needs a weaker structural check.

## 3. An irreversible external effect never repeats

**Rule.** Any action that cannot be undone (a payment, a submission to a tax authority, an
email, a deploy, a file transfer) gets a **state machine, an idempotency key and a recorded
external identifier** before it gets a retry.

*Evidence.* The tax authority answers a repeated document submission with a fault ("this file
was already sent"). Retrying after a reply had actually arrived duplicates a filing. The fix
was the state machine: a row with no submission is submitted; one with a ticket is polled;
one that already carries its receipt is finished; a crash between reply and write is a named
window, not a silent one.

**Apply.** Write the state set and the transitions first; make "already sent" a state, not a
hope; store the provider's own identifier the moment it answers.

## 4. Background work has a written contract

**Rule.** Every worker — drainer, poller, indexer, refresher — is documented in one page:
who runs it, when, how it is visible to the operator, what happens if the app closes mid-work,
how it is kept off the critical path, and what it does when the resource (disk, network, CPU)
is not there.

*Evidence.* A refresh that pulls ~390 MB and builds a ~2.3 GB index in ~23 minutes cannot be a
comment. The design that survived review: download anytime, ingest yielding to live traffic,
atomic swap so a failure never degrades the current data, progress on screen, and the cost
stated in the UI.

**Apply.** If a job can outlive a request, it needs a contract. If it cannot be paused or
resumed, say so and bound it.

## 5. A derived artifact carries its own identity

**Rule.** Anything derived from an external source (an index, a cache, a generated document)
records **source, digest, date, size and count**, is built in a temp location and swapped in
atomically, and **fails loudly** when the upstream shape changed instead of writing shifted
data.

*Evidence.* An index built from a daily-published dataset must answer "which snapshot is this
answer from?"; without it, a stale answer looks exactly like a fresh one. The loud schema check
is what refused to index a row it could not map — and the atomic swap is what kept the previous
2 GB index serving through a failed rebuild.

**Apply.** Metadata table/key inside the derived artifact; temp + rename; a unit test that a
failed rebuild leaves the old artifact answering; a test that a changed upstream shape is
refused by name.

## 6. Measure in your version; never inherit a performance recipe

**Rule.** A performance claim from documentation or a blog is a hypothesis for your runtime.
Read the plan (or the profile) in the version you ship. Write the plan assertion as a test so
the regression cannot return in silence.

*Evidence.* "A `LIKE 'x%'` cannot use a case-insensitive index; use a range comparison" — in
the SQLite version in play, the `LIKE` **did** use the index and the hand-written range did a
full scan. The recipe was inherited; the measurement said otherwise.

**Apply.** One plan/profile assertion per hot query, and the measured number in the comment
that justifies the shape.

## 7. An assertion is not coverage until a mutation makes it RED

**Rule.** For every new assertion that claims to protect something, break the code
deliberately, watch it fail, restore the file **byte-exactly**, and verify the hash. A mutation
only counts if it reaches the layer under test.

*Evidence.* Three consecutive units had tests that could not fail (asserting through the
constant they were testing; comparing a value with itself; observing the first of two reads).
And the inverse trap: a mutation run with a wrong test name printed "no tests to run" and
looked like a pass.

**Apply.** Record the mutation, the failing line and the restored hash next to the test.
"Observed RED" without the line is a claim.

## 8. Declare the review load, and cut by behaviour

**Rule.** A change is reviewed in slices small enough to read in one sitting. Declare the
size **excluding generated code**, split by behaviour (not by file), keep the tests and docs
inside their slice, and make the fixes of a review travel inside the unit they belong to.

*Evidence.* Two units declared 1 393 and 2 207 changed lines; in the first, **64 %** was
template-generator output, which inflates the number without adding review load. The
declaration rule that survived: total **and** reviewable, with the generated share named.

**Apply.** `git diff` with the generated paths excluded; chained slices when one behaviour
does not depend on the next; never open a pull request in a state already known defective.

## 9. Name the human gates on day one

**Rule.** List, in the plan, the checks **no agent can close**: the real browser walkthrough,
the real printer, the real device, the real credential, the provider's sandbox, the taste
call. They do not disappear by being omitted; they disappear by being owned.

*Evidence.* A UI packet with zero failures and 314 warnings still required a human to look at
the screens; a payment/tax integration still required a real credential and a real sandbox
before anyone could say "it works".

**Apply.** A "human gates" section in the feature document, updated per unit, never marked
done by an agent.

## 10. Name the residual; never leave a silent gap

**Rule.** What was deliberately not done — and why — is written next to what was done. A gap
that is not named is indistinguishable from an oversight, and both reviewers and future you
pay for it.

*Evidence.* The most valuable line in several review rounds was "this is deliberately not
here: X, because Y", and the most expensive defect class was the opposite: a claim whose
authority was a comment.

**Apply.** Each unit ends with "deliberately NOT done: …" and "open: …", with a name on every
item.

## 11. One owner per rule — check across layers, not only inside one

**Rule.** A business rule has exactly one home. When an outer layer needs the same rule for
its own shape, it must CALL the owner, not restate it.

*Evidence.* "A factura needs a RUC acquirer" ended up stated twice: as a domain predicate and
as an adapter check. Both were defensible in isolation; together they are drift waiting to
happen.

**Apply.** Grep the vocabulary of each rule after writing it; if it appears twice, one of the
two must call the other.

## 12. A green suite and a correct artefact are different questions

**Rule.** Tests protect the rules you thought of; the artefact carries the truth you did not.
Keep the two checks separate and run both: the suite for logic, the artefact comparison for
correctness.

*Evidence.* Five times in one feature the whole suite was green while the produced document was
wrong — attributes that did not match the accepted example, a namespace invented, a boundary
no test touched (`>` mutated to `>=` with the suite still green).

**Apply.** Keep a real accepted artifact in the repo, compare against it in a test, and anchor
every boundary with a case exactly ON it.

## 13. A refuter, a reviewer or a model is an opinion; the measurement decides

**Rule.** Adversarial review is an outside view, not authority. When a claim and a measurement
disagree, the measurement wins — and the claim gets re-checked with a different method.

*Evidence.* A reviewer's CRITICAL was "refuted" by the provider's own refuter; a real browser
measurement showed the effect was genuine (a disabled selected option is not submitted, so the
form silently stored a different document). In the same session the opposite happened too: a
reviewer's finding about a test being unable to fail was right and led to a better test.

**Apply.** For any severe claim: reproduce the mechanism, then decide. Record which method
settled it.

## 14. Reproducible evidence, or it is not evidence

**Rule.** Every claim in a report carries the command, the digest, the count or the path that
produces it. Sizes and timings are measured for the exact slice they describe; a measurement
made for another cut is not a measurement.

*Evidence.* Declared sizes that had never been measured produced silent exceptions; a probe
that printed sixteen bytes of a file and concluded was repeated by counting bytes to prove a
checkout rewrote it.

**Apply.** Paste the number **and** the command that produced it; keep the raw output when it
is cheap; never restate a size you did not measure for that slice.

## 15. Generated code is not authored

**Rule.** Pin the generator's version, commit its output, never count it as authored review
load, and never hand-edit it — regenerate.

*Evidence.* One template edit renumbered hundreds of lines in the generated file; the diff
looked enormous and the change was three attributes. A second generator has a shape the
compiler refuses: a brace moved outside its branch fails with a line and column, and no test
would have seen it.

**Apply.** The generator command in the checklist, its version pinned, and the generated paths
named in the size rule.
