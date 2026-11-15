# The scaffold, in the order that keeps the boundary

Nine steps, in this order. The shared half -which kind of project this is, the tree,
the ADRs- belongs to `arch-project`; this page is what changes when the stack is
Tauri v2. A step skipped is not a shortcut: the guard comes before the business code
because a boundary written afterwards is a boundary nobody can measure.

1. **Scaffold the workspace** from `assets/tauri-layout.txt`: a framework-free domain
   crate plus the app crate under `src-tauri/`, which keeps `build.rs`,
   `generate_handler!`, `capabilities/**/*` and `permissions/**/*` where Tauri put
   them. Move files, do not recreate them: a capability that stops being declared is
   every command it granted.
2. **Write the domain-purity guard before the business code**, from
   `references/rust-guard-recipes.md`. Recipe A is the workspace split and it is
   compiler-enforced; the source scan is the backstop for what the compiler cannot see.
3. **Define the serializable error type** of `references/tauri-bridge.md` before the
   first command: a command that maps errors after the fact leaks the framework one.
4. **Write one thin command as the template**: validate the input, call the domain,
   map the error, return a DTO. The next ten commands copy this shape, so it is the
   one place to get it right.
5. **Generate the TypeScript bindings from Rust** with `tauri-specta` v2, and `ts-rs`
   only for shared types. A hand-written binding beside a generated one is two
   contracts that drift in silence.
6. **Create the single service layer** under `src/services/` over the generated
   bindings. Components call the service; `invoke()` appears once per command, there.
7. **Wire the three verification levels** of `references/ui-verification.md`, and
   replace the scaffold's stub test script (`"test": "echo \"Error: no test specified\"
   && exit 1"`): a project that ships that line has a test step that cannot fail.
8. **Write the first ADR** with `arch-project`'s `assets/adr-template.md`: the split,
   the guard and the reason. The verification notes section names the guard test, so
   the decision is enforced and not only remembered.
9. **If a file landed under `src-tauri/permissions/`, the granting capability lands in
   the same change.** That file flips `has_app_acl_manifest` and starts gating every
   app command, so a half of the pair is a build that refuses its own frontend.

## What this page assumes

`arch-project` is already loaded: the triage row, the tree, the twelve decisions and
the ADR discipline are its own. `drive` owns the day-to-day order -one step, RED
first, one commit- and `tests` owns the mutation that proves a guard is real. A
project with no Rust core never reads this page.
