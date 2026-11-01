# the-design-surface-any-agent-can-use
Base: 18014fa22dba18105e3512ade0a1ecb7bd9bfef2

## Requirements (verbatim, from the user)
R1 · "construye todo lo que tiene gentle-design, para que funcione igual en nuestros agentes"

## Done when
- a skill named design ships with odachi and travels to every client through setup and pack
- a design is a folder with self-contained artboards, a manifest and a comments file, and the formats are pinned by a versioned fixture
- the canvas is a command any agent can run -list, open, stop- and writes comments the agent reads from disk
- the design system sync reads the project tokens and never edits them

## Steps
- [x] T1 · the fence names the sixteen skills, so a new one is a feature and not an edit   R1 · tree:5142543e219d597e2c5d0064088431ce32cb506e
- [x] T2 · the design folder has a format, a manifest and comments, pinned by a versioned fixture   R1 · tree:df93fd53f06f4336e0703db397083eef9517d5b3
- [x] T3 · an agent lists designs and comments with a command, and resolves them   R1 · tree:4594717825938c3be62b4cb99eb6d4e47418dc5f
- [ ] T4 · the design skill ships with odachi and travels to every client   R1
- [ ] T5 · the canvas opens, serves the artboards and writes comments any agent can read   R1
- [ ] T6 · the design system sync reads the project tokens and never edits them   R1
- [ ] T7 · the handoff turns an artboard into the implementation instruction   R1

## History
L1 · "construye todo lo que tiene gentle-design, para que funcione igual en nuestros agentes"
L2 · OFICIO · "T1 landed in the plan commit: the fence and the brief changed before the first step was marked" (porque the policy correction has to exist before any file under skills/ is written, and the plan commit is the tree that carries it)
