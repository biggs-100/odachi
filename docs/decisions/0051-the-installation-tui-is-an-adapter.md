# 0051 - The installation TUI is an adapter

Status: accepted - Date: 2026-11 - Session: after v0.38.2

## Context

ADR 0001 rejected a TUI with one sentence: "Screens are a second product and a second set of bugs." The sentence holds for the product. It does not answer the person who has to install and configure odachi for the first time: that hour is seven flags spread across `init`, `setup`, `level`, `review mode` and the two TOML files, and the only record of what was chosen is the files themselves. The human asked for a terminal wizard like the one gentle-ai ships.

The study of gentle-ai -read-only, never copied- found a multi-screen Bubbletea wizard: detection, per-client model pickers, install progress, completion, backups and uninstall. odachi's configuration surface is smaller: two clients, one skills scope, one voice and the review file.

## Decision

- **One verb, an adapter over the verbs that already exist.** `odachi tui` validates the choices and runs them in order through `init` or `setup`, `level --to`, `review mode` and the review config writer. The domain gains no dependency and knows nothing about screens.
- **Bubbletea and Lipgloss.** The same family gentle-ai uses: cross-platform keys and rendering, and `teatest` for the screens. The binary stays static and makes no network call.
- **The first version is installation and configuration, bounded.** Clients, skills scope, voice, the review switch and the lens models. Per-client model pickers, backups, uninstall, community tools, the binary download, the prompt limit and the retention stay out.
- **The CLI keeps the machine surface.** The flags and `setup` keep their behavior; the TUI refuses without a terminal and names `odachi setup`. No script changes because a person got a screen.
- **The review file is written by its own format.** `review.SetConfig` reads the file, changes the switch and the lens table, and renders it back with the prompt limit and the retention it did not touch.
- **ADR 0001's "A TUI" is superseded for this surface only.** Screens are still a second product; the scope is the price of having one where the first hour is impossible without it.

## Alternatives rejected

- **A web UI or a local server.** A daemon, a port and a browser, against one static binary that works offline.
- **`setup` opening the wizard when a TTY is present.** It changes the behavior scripts and CI already depend on; a verb is the explicit door.
- **A hand-rolled ANSI TUI.** More code, fragile keys and resize on Windows, and no screen test harness.

## Consequences

- The release gains a dependency family. The build stays `CGO_ENABLED=0`, the binary stays static, and the domain's boundary test still forbids imports.
- The apply screen is the same code the verbs run: the TUI cannot configure something the CLI cannot, and a failure names the operation and what remains pending.
- A person can configure odachi without reading the flags; a script keeps the flags. The two surfaces share the code and the files.
- The wizard is a surface to maintain. The bounded first version is the answer to ADR 0001's cost: the next screen is a decision, not a reflex.
