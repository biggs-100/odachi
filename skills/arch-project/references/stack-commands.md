# The commands per stack, and the four questions before one

The layers belong to `drive` and `tests`: RED first, the change in the right place,
one commit per work unit. This page answers what a stack decides - what to run for
it, and what to ask before touching one of its packages.

## The commands, by stack

| Stack | The package | The suite | Build and lint |
| --- | --- | --- | --- |
| Go | `go test ./<pkg>/...` | `go test ./...` | `go build ./...` · `go vet ./...` · `gofmt -l .` |
| TypeScript | `npx vitest run <dir>` | `npm test` | `tsc --noEmit` · `eslint .` |
| Java / Kotlin | `mvn test -pl <module>` | `mvn test` | `mvn compile` · spotless or checkstyle |
| Python | `pytest <dir>` | `pytest` | `ruff check .` |
| C# | `dotnet test <proj>` | `dotnet test` | `dotnet build` · `dotnet format` |

## The four questions before a change

1. What do I touch: a package, a route, a template, a migration.
2. Who depends on it: the impact map first, then the stack's own scan - `go list -deps ./...`, `depcruise src`, `mvn dependency:tree` and a search for the symbol.
3. Which tests cover it: the suites of that area, and only those, before the whole suite pays for them.
4. Domain or adapter: the domain runs the whole suite, an adapter runs its own and the port's, and the frontend runs the tests plus a walkthrough, because a test does not see the page.

## A change that spans both

Branch by abstraction: the port lands first, the system compiles and runs at every
step, and the implementation is replaced afterwards. A branch that only compiles at
its end is integration paid all at once, on the day nobody has time for it.
