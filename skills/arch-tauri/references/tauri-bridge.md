# Tauri v2 bridge: frontend ↔ Rust contract

Tauri v2. This document owns the Tauri-specific bridge rules. Layout, ADRs and the
layered change checklist belong to `bootstrap` and `drive`, which every project in
this catalog carries; nothing here repeats them.

## Flow

```
React component
  └─ src/services/<feature>.ts          # THE single typed layer (Hard Rule 3)
       └─ generated bindings (src/bindings/)  # types + command names generated from Rust (Hard Rule 4)
            └─ #[tauri::command] adapter # validate → domain → map error → DTO (Hard Rule 2)
                 └─ domain crate         # pure business rules; knows no framework (Hard Rule 1)
```

Components never call `invoke()` directly. They call `src/services/`, which wraps the
generated bindings. That is the only place the wire format is visible.

## Thin-command template

```rust
// src-tauri/src/commands/projects.rs
use serde::Serialize;

use crate::error::AppError;

#[derive(Serialize)]
pub struct ProjectDto {
    pub id: String,
    pub name: String,
}

#[derive(serde::Deserialize)]
pub struct CreateProjectInput {
    pub workspace_id: String,
    pub name: String,
}

#[tauri::command]
pub fn create_project(input: CreateProjectInput) -> Result<ProjectDto, AppError> {
    // 1. validate SHAPE only — business rules live in the domain
    let name = input.name.trim().to_string();
    if name.is_empty() {
        return Err(AppError::validation("name must not be blank"));
    }

    // 2. call the domain; `?` converts the domain error via From
    let project = domain::service::create_project(&input.workspace_id, &name)?;

    // 3. map errors and return a DTO — never a domain type, never a framework type
    Ok(ProjectDto::from(project))
}
```

The command contains no rule, no query, and no persistence. Everything it does with
meaning lives in `domain::service`, reached through a port implemented in
`src-tauri/src/adapters/`.

## Serializable error pattern

Errors cross the boundary as data. `thiserror` builds the error type; a manual
`serde::Serialize` fixes the wire shape:

```rust
// src-tauri/src/error.rs
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("validation: {0}")]
    Validation(String),
    #[error("not found: {0}")]
    NotFound(String),
    #[error("conflict: {0}")]
    Conflict(String),
    /// Infrastructure detail that must NEVER cross the IPC boundary.
    #[error("internal: {detail}")]
    Internal { id: String, detail: String },
}

impl AppError {
    pub fn validation(message: impl Into<String>) -> Self {
        Self::Validation(message.into())
    }

    /// Records the detail locally and returns only what the frontend may see.
    /// This is the single place the detail is written; it is never serialized.
    pub fn internal(detail: impl std::fmt::Display) -> Self {
        let id = uuid::Uuid::new_v4().to_string();
        eprintln!("app error [{id}]: {detail}"); // route this to your logger
        Self::Internal { id, detail: detail.to_string() }
    }
}

impl From<domain::DomainError> for AppError {
    fn from(error: domain::DomainError) -> Self {
        match error {
            domain::DomainError::NotFound(m) => Self::NotFound(m),
            domain::DomainError::Conflict(m) => Self::Conflict(m),
            domain::DomainError::Invalid(m) => Self::Validation(m),
        }
    }
}

impl serde::Serialize for AppError {
    fn serialize<S: serde::Serializer>(&self, serializer: S) -> Result<S::Ok, S::Error> {
        use serde::ser::SerializeStruct;

        let (kind, message, correlation_id) = match self {
            AppError::Validation(m) => ("validation", m.as_str(), None),
            AppError::NotFound(m) => ("not_found", m.as_str(), None),
            AppError::Conflict(m) => ("conflict", m.as_str(), None),
            // The detail stays server-side. Only the id crosses the boundary, so the
            // user can report it and you can find the logged line.
            AppError::Internal { id, .. } => ("internal", "internal error", Some(id.as_str())),
        };

        let mut state = serializer.serialize_struct("AppError", 3)?;
        state.serialize_field("kind", kind)?;
        state.serialize_field("message", message)?;
        state.serialize_field("correlationId", &correlation_id)?;
        state.end()
    }
}
```

Commands return `Result<T, AppError>`. Why framework errors must not leak:

- A framework error type is not a stable contract — it changes with the framework and
  couples every frontend caller to internals it should not know.
- Its display text can leak paths and internal detail; a tagged `kind` gives the UI a
  stable thing to branch on.
- Internal failures are the sharpest case: a raw message can carry SQL, absolute paths and
  infrastructure detail into the renderer. Serialize a generic message plus a correlation
  id, and keep the detail in your logs. `AppError::internal` above is the only sanctioned
  way to build one.
- The frontend can exhaustively handle `kind`; it cannot exhaustively handle an opaque
  framework error.
- Never `impl serde::Serialize` for a framework error type to "make it work". Map it into
  `AppError::internal(...)` at the adapter boundary and keep the framework type private.

## Typed contract (Hard Rule 4)

Generate TypeScript from Rust; do not hand-mirror DTOs.

- **`tauri-specta` v2** targets Tauri v2 — its specta v2 line is the supported pairing
  with Tauri v2 (specta v1 paired with Tauri v1). It exports TypeScript bindings for
  commands: <https://github.com/specta-rs/tauri-specta>.
- **`ts-rs`** is the lighter alternative when you only need to mirror shared types.

Put the generated output on the FRONTEND side (for example `src/bindings/`, next to
`src/services/`) and never hand-edit it. Do not place it in `src-tauri/gen/`: Tauri
already generates its own artifacts there (`src-tauri/gen/schemas`), so mixing two
generators in one directory conflates ownership and Tauri's default `.gitignore`
commonly excludes that path. `src/services/` is the only consumer surface.

## ACL: when commands are actually gated (LOAD-BEARING)

Verified against `tauri-2.11.2/src/webview/mod.rs` (~line 1823):

```rust
if (plugin_command.is_some() || has_app_acl_manifest || !is_local)
  && request.cmd != crate::ipc::channel::FETCH_CHANNEL_DATA_COMMAND
  && invoke.acl.is_none()
{ /* reject */ }
```

In-source comment: "Check ACL on plugin commands, when the app defined its ACL manifest,
or when the request comes from a non-local (remote) origin. This ensures remote content
can never reach custom commands unless an explicit `remote` capability has been
configured for them."

Consequences:

- App-defined commands (`#[tauri::command]` in the app crate) are callable from the
  frontend with NO capability when the app defines no app-level ACL manifest and the
  origin is local.
- Plugin and core commands are always gated.
- Any command invoked from a non-local (remote) origin is always gated.

### THE TRAP: one permission file gates every app command

`tauri-build-2.6.2/src/acl.rs` (~line 408):

```rust
let has_app_manifest = app_acl.manifest.default_permission.is_some()
  || !app_acl.manifest.permission_sets.is_empty()
  || !app_acl.manifest.permissions.is_empty();
if has_app_manifest { acl_manifests.insert(APP_ACL_KEY.into(), app_acl.manifest); }
```

Creating ANY app permission file under `src-tauri/permissions/` flips
`has_app_acl_manifest` and starts gating EVERY app command. Previously working `invoke()`
calls then fail unless a capability grants them. This is a silent, surprising behavior
change.

**Rule: declare a default permission set AND a capability in the SAME change, or do not
add the file.**

Verified details:

- `tauri-build-2.6.2/src/acl.rs` reads `./permissions/**/*` and `./capabilities/**/*`
  relative to `env::current_dir()` — the APP crate directory
  (`parse_capabilities("./capabilities/**/*")` ~line 428; `env::current_dir()?.join("permissions")`
  ~line 308). This is why a workspace split is safe as long as the app crate keeps
  `capabilities/` and `permissions/`.
- App permissions are referenced BARE in a capability (`"allow-my-command"`), while
  core/plugin ones are prefixed (`"core:default"`, `"dialog:allow-open"`). Resolution:
  `let key = permission_id.get_prefix().unwrap_or(APP_ACL_KEY);`
  (`tauri-build-2.6.2/src/acl.rs` ~line 356).
- NAME COLLISION TRAP: the core permission set `core:app:default` / `app:default` is
  Tauri's own `app` plugin and is NOT your commands. Do not reach for it as your app's
  default.
- Auto-generation of app permission files is OPT-IN, only via
  `tauri_build::Attributes::new().app_manifest(AppManifest::new().commands(&[...]))` in
  `build.rs` (`tauri-build-2.6.2/src/lib.rs` ~line 411). A default `build.rs` with plain
  `tauri_build::build()` does NOT create one.
- Verified example shape (`tauri-utils-2.9.2/src/acl/build.rs::autogenerate_command_permissions`),
  `src-tauri/permissions/<anything>.toml`:

  ```toml
  [[permission]]
  identifier = "allow-my-command"
  description = "Enables the my_command command without any pre-configured scope."
  commands.allow = ["my_command"]
  ```

  and a set file `src-tauri/permissions/default.toml` with `[default]` +
  `permissions = ["allow-my-command"]`, consumed by a capability as
  `"permissions": ["allow-my-command"]`.

UNCONFIRMED: which v2 minor introduced the `has_app_acl_manifest` branch; older v2
releases gated plugin/core commands only.

## Anti-patterns

- `invoke("...")` inside a component instead of `src/services/` (Hard Rule 3).
- Business rules, queries, or persistence inside a `#[tauri::command]` (Hard Rule 2).
- Hand-written TS mirrors of Rust DTOs instead of generated bindings (Hard Rule 4).
- Returning domain types or framework error types across IPC (Hard Rule 5).
- Adding a file under `src-tauri/permissions/` without a granting capability in the same
  change (the gating trap).
- Treating `app:default` / `core:app:default` as your app's default permission set.
- Depending on `tauri`, `tauri-plugin-*`, `diesel`, or `rusqlite` from the domain crate
  (Hard Rule 1).
- Placing `capabilities/` or `permissions/` outside the app crate.
