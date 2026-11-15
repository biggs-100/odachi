---
name: arch-tauri
description: Trigger: a Tauri v2 desktop project, new or restructured (Rust core and JS/TS frontend). The domain never knows Tauri exists, commands are thin, the contract is generated, the UI is verified.
---

## When

When the work is a Tauri v2 desktop app -scaffolding one, or restructuring an
existing single-crate app into a workspace- and the question is where the domain
ends and the framework begins. A project with no Rust core is not this skill's:
load `bootstrap` and nothing here applies.

## Rules

- The domain crate never knows Tauri exists: no `tauri`, `tauri-plugin-*`, `diesel` or `rusqlite` inside it.
- A command is a thin adapter: validate the input, call the domain, map the error, return a DTO. No business rule lives there.
- The frontend reaches the backend through one typed layer under `src/services/`: `invoke()` is never scattered across components.
- The contract is generated from Rust -`tauri-specta` v2 for command bindings, `ts-rs` for shared types- and never hand-mirrored.
- Errors cross the boundary as data: a serializable type of the project's own, never a framework error leaked to the frontend.
- Purity is enforced mechanically: a framework-free domain crate plus a source-scan test as the backstop, written before the business code.
- Adding a file under `src-tauri/permissions/` flips `has_app_acl_manifest` and starts gating every command: declare the permission set and the capability in the same change, or do not add the file.
- The UI is verified in three layers, in order, and none is skipped.
- A business rule changes in the domain crate, never in a command, however small the change looks.

## Decision gates

| Situation | What to do |
| --- | --- |
| A new project | Split at once: the framework-free domain crate plus the app crate under `src-tauri/`; `references/tauri-scaffold.md` is the order |
| An existing single-crate app | Write the source-scan guard first; the workspace split comes later; `references/rust-guard-recipes.md` has every mechanism and what fails as one |
| A command crosses the boundary, or a Rust to TS contract is needed | `references/tauri-bridge.md`: the flow, the thin-command and error templates, the ACL gating |
| The UI has to be verified | `references/ui-verification.md`: the three layers, the platform matrix, the `wdio.conf.ts` fragment |
| Remote or external content in a webview | Commands are gated regardless; configure an explicit `remote` capability |
| A plain web app with no Rust core | Stop: this skill does not apply; `bootstrap` is the door |

## Output

The workspace tree, the guard test path, the generated-bindings path, the service
layer path, and the ordered commands that verify it. A project without its three UI
layers is reported as scaffolded and not verified, never as done.

## Never

- Put a framework import in the domain crate "just for now": that is the boundary this skill exists to keep.
- Scatter `invoke()` across components instead of the one service layer.
- Hand-write the TypeScript bindings beside the generated ones.
- Call a project verified with a UI layer skipped, or with a screenshot nobody asserts on.
- Add a permission file without the capability that grants it in the same change.

## References

- `references/tauri-scaffold.md` - the nine steps of the scaffold, and the order that keeps the boundary.
- `references/tauri-bridge.md` - the command flow, the thin-command and error templates, the typed contract, the ACL gating conditions.
- `references/rust-guard-recipes.md` - the purity mechanisms in order, and what does not work as a guard.
- `references/ui-verification.md` - the three layers, the platform matrix and the `wdio.conf.ts` fragment.

## Assets

- `assets/tauri-layout.txt` - the annotated workspace tree, and the current-to-target mapping for an existing app.
