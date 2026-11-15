# The tree, and what each folder answers

The shape is one row of `references/project-shape.md`, and the tree is the same for
every row that keeps it: a pure core, ports as interfaces, adapters as
implementations, and one entrypoint that composes.

```
<project>/
├── cmd/<app>/                # entrypoints only (main), thin - rename per stack (main.ts, Application.kt, ...)
├── internal/
│   ├── domain/               # pure core: entities, value objects, the rules. Standard library only.
│   │                         # Go's internal/ is compiler-enforced; other stacks get it from the guard.
│   ├── ports/                # interfaces: repositories, transactions, clock, pdf, numbering...
│   ├── adapters/             # implementations: sqlite, http, pdf, ...
│   └── server/ (or app/)     # composition shell: wiring and handlers. No business rules.
├── docs/
│   ├── decisions/            # ADRs: 0001-*.md, 0002-*.md, ...
│   └── impact-map.md         # who depends on what, and which tests cover it
├── migrations/               # forward-only schema changes
├── frontend/ (optional)      # UI assets and templates
└── testdata/                 # goldens and fixtures
```

Stack mapping: Go's `internal/` = TypeScript `src/` with dependency-cruiser rules =
Java package `..domain..` with ArchUnit = Python package with import-linter. The
names adapt; the boundary does not.

Why this tree: the folder that holds a rule is the folder that can be tested without
a database, a browser or a network. A core that needs the framework to compile cannot
be proven, and the guard of `references/guard-recipes.md` is what keeps it that way.
