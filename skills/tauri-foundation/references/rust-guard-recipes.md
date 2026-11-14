# Rust domain-purity guard recipes

Enforce this boundary for Rust/Tauri v2: the domain imports ONLY stdlib, sibling
domain code, and ports. Pick the recipe by situation; do not
claim more coverage than each mechanism actually has.

## Recipe A — workspace split (primary, compiler-enforced)

Move the domain into its own workspace crate with no framework deps in its `Cargo.toml`:

```toml
# domain/Cargo.toml — no tauri, no tauri-plugin-*, no diesel, no rusqlite
[package]
name = "domain"
edition = "2021"
```

`src-tauri/Cargo.toml` depends on it (`domain = { path = "../domain" }`) and keeps every
framework dependency. A framework import inside the domain crate then fails to compile
(E0432/E0433: unresolved crate/import).

**Limitation — crate-level only.** The compiler cannot stop `domain::model` from
importing `domain::infra` inside that same crate; both are the same crate. It also cannot
stop `[dev-dependencies]` or feature-gated dependencies from pulling a framework in
without breaking the "no framework deps" claim. Recipe A narrows the hole; Recipe B
closes the intra-crate part.

**tauri-cli compatibility.** A workspace split does NOT conflict with tauri-cli
provided the app crate remains the crate that (i) runs `tauri_build::build()` in its own
`build.rs`, (ii) hosts `generate_handler!`, and (iii) contains `capabilities/**/*` and
`permissions/**/*`. See `assets/tauri-layout.txt`.

UNCONFIRMED: whether tauri-cli v2 mis-discovers the built binary when a workspace hoists
`target/` to the workspace root.

## Recipe B — source-scan `#[test]` (module-level backstop)

The ONLY option that catches module-level violations inside a single crate. Use it as the
first guard when an existing app is still a single crate (e.g. ANTIGRAVITY today), and
keep it after a split to cover the intra-crate holes Recipe A cannot see.

```rust
// domain/tests/domain_purity.rs
// Backstop for violations INSIDE one crate: walks src/domain/**/*.rs and fails on
// any framework reference. Regex/string scanning is a backstop, never a proof.
use std::{
    fs,
    path::{Path, PathBuf},
};

const FORBIDDEN_CRATES: &[&str] = &["tauri", "tauri_plugin", "diesel", "rusqlite"];

fn collect_rs_files(dir: &Path, out: &mut Vec<PathBuf>) {
    for entry in fs::read_dir(dir).expect("read_dir") {
        let path = entry.expect("entry").path();
        if path.is_dir() {
            collect_rs_files(&path, out);
        } else if path.extension().map_or(false, |ext| ext == "rs") {
            out.push(path);
        }
    }
}

#[test]
fn domain_sources_do_not_reference_frameworks() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("src/domain");
    let mut files = Vec::new();
    collect_rs_files(&root, &mut files);
    assert!(
        !files.is_empty(),
        "no domain sources found at {} — fix the scan root",
        root.display()
    );

    let mut violations: Vec<String> = Vec::new();
    for file in files {
        let source = fs::read_to_string(&file).expect("read source");
        for (index, raw_line) in source.lines().enumerate() {
            let line = raw_line.trim_start();
            if line.starts_with("//") {
                continue; // naive: only whole-line comments are skipped
            }
            // Scan BARE PATH TOKENS, not just `use` statements, so fully-qualified
            // calls such as `tauri::Manager::...` are also caught.
            for krate in FORBIDDEN_CRATES {
                let token = format!("{krate}::");
                if line.contains(&token) {
                    violations.push(format!(
                        "{}:{}: {}",
                        file.display(),
                        index + 1,
                        raw_line.trim()
                    ));
                }
            }
        }
    }

    assert!(
        violations.is_empty(),
        "domain purity violations:\n{}",
        violations.join("\n")
    );
}
```

Adjust the scan root: `src/domain` inside a single crate, `src` when the test lives in
the dedicated domain crate.

### Known holes (document them; do not pretend they are closed)

- **Fully-qualified calls with no `use`** (e.g. `tauri::Manager::...`). Mitigate by
  scanning for bare `tauri::` / `diesel::` / `rusqlite::` path tokens, not just `use`
  statements — the sketch above already does this.
- **`#[path]` module redirection** — a file pulled in from outside the scanned glob is
  never read.
- **`macro_rules!`-generated imports** — the framework tokens never appear literally in
  the scanned file.
- **Re-exports through a facade module** — `pub use tauri as framework;` followed by
  `framework::...` hides the crate name from the scan.

## Recipe C — dependency backstop (`cargo-machete`, `cargo-deny [bans]`)

Prove the core crate declares no framework dependency and that none enters its resolved
graph:

- `cargo-machete` — flags unused manifest dependencies. Useful only as a backstop that a
  framework dep is not declared in the core crate.
- `cargo-deny` `[bans]` — bans a crate from the resolved dependency graph.

Both operate on the resolved crate graph (`Cargo.lock` / `cargo metadata`). They are a
cheap second signal, not the guard.

## What does NOT work

- **`cargo-deny` `[bans]`, `cargo-tree`, `cargo-machete`** — crate-graph only. They are
  structurally incapable of seeing `use` statements and CANNOT catch a module-level
  violation inside a single crate.
- **`cargo-modules`** (`cargo modules dependencies --with-uses`) — DISPLAYS per-module
  crate usage; advisory output, not a build/CI failure by itself. It becomes a gate only
  if wrapped in a script that asserts on its output.
- **A rustc/clippy lint for "module X must not depend on crate Y"** — no stable lint
  exists. Do not wait for one; use Recipe B.
