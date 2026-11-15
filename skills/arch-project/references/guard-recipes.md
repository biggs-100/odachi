# Architecture Guard Tests by Stack

Enforce the domain boundary mechanically. Pick the recipe for the project stack; the guard MUST fail the build/CI on illegal imports. The rule is always the same: domain imports ONLY stdlib, sibling domain packages, and ports.

## Go

Scan with `go list` in a test (pattern: `internal/domain/arch_test.go` in agua-potable):

```go
func TestDomainPurity(t *testing.T) {
    out, err := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./...").Output()
    // allow: stdlib, module domain siblings, module ports; anything else fails
}
```

Language bonus: `internal/` visibility is compiler-enforced outside the module.

## Java / Kotlin (JVM)

ArchUnit (JUnit 5):

```java
@AnalyzeClasses(packages = "com.acme.domain")
class DomainPurityTest {
    @ArchTest
    static final ArchRule domain_pure = noClasses()
        .that().resideInAPackage("..domain..")
        .should().dependOnClassesThat()
        .resideInAnyPackage("..adapters..", "..server..", "com.acme.infra..");
}
```

## C#

ArchUnitNET:

```csharp
[Fact]
public void Domain_should_not_depend_on_adapters() {
    var arch = new ArchLoader().LoadNamespacesWithinAssembly(typeof(Domain.Money).Assembly);
    var rule = Types().That().ResideInNamespace("Domain").Should()
        .NotDependOnAny(Types().That().ResideInNamespace("Adapters"));
    CheckRule(arch, rule);
}
```

## TypeScript

dependency-cruiser, forbidden rule (fail in CI: `depcruise src`):

```json
{
  "forbidden": [{
    "name": "domain-purity",
    "from": { "path": "^src/domain" },
    "to": { "path": "^(src/(adapters|server|infra))|(node_modules)" }
  }]
}
```

Alternative: ESLint `import/no-restricted-paths`.

## Python

import-linter (`.importlinter` config, run `lint-imports` in CI):

```ini
[importlinter]
root_package = myapp

[importlinter:contract:domain-purity]
name = Domain is pure
type = forbidden
source_modules = myapp.domain
forbidden_modules = myapp.adapters, myapp.server
```

## No dedicated tooling available

Fall back to folder-layout convention + CI script that greps imports (rg) and fails on forbidden prefixes. Still mechanical: a human running CI is not a check.
