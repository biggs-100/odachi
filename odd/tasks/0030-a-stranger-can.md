# a-stranger-can

## Requirements (verbatim, from the user)
R1 · "a stranger can go from zero to a merged pull request by reading the repository"

## Done when
- CONTRIBUTING names the path and the one command that reproduces the gate locally
- AI_POLICY says what must be disclosed and who owns the change; SECURITY gives a private channel
- the issue forms ask for version and doctor; the pull request template asks for the proof, the risk and the rollback
- a test fails if a file disappears or the pull request template stops asking for the proof
- the README links the surface in one place

## Steps
- [x] T1 · CONTRIBUTING and AI_POLICY: the loop, the size, disclosure   R1 · c6f63a4
- [x] T2 · issue forms, pull request template and SECURITY   R1 · 230658c
- [x] T3 · README links and a test that the surface stays there   R1 · e8372ce

## History
L1 · "a stranger can go from zero to a merged pull request by reading the repository"
L2 · OFICIO · "Contributors never edit brief.md or the fence" (porque the maintainer owns the vision and the non-goals; a change cannot decide what it is judged against)
L3 · OFICIO · "The pull request template asks for the proof, the risk and the rollback" (porque those are the three things a reviewer cannot guess from the diff)
L4 · OFICIO · "The local command carries --fence-from and --base" (porque without them a contributor runs a weaker gate than CI and finds out at the merge)
L5 · OFICIO · "Security reports go through GitHub advisories" (porque a public issue is a disclosure; private reporting is enabled on the repository)
L6 · OFICIO · "The ruleset turns off the extra approval for unattributed changes" (porque without it an outside pull request could never merge in a one-person repository)
