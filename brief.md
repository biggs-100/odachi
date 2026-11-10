# tui-instalacion

## What must exist
- una tui de instalación y configuración al igual que gentle-ai
- `odachi tui` opens a terminal wizard that configures the clients, the skills scope, the voice and the review, and applies each choice with the verbs that already exist

## Done when
- `odachi tui` opens the wizard when there is a terminal; without one it refuses with a line that names `odachi setup` and prints no screen output
- the first screen says the version, whether the folder is an odachi project, the clients chosen or detected, the skills scope and the profile in force
- when the folder is not a project, the wizard offers to configure the machine alone (`setup`) or to initialize the project (`init`), and never creates a repository without the human choosing it
- the wizard chooses at least one client (opencode, claude or both); the choice is stored in `.git/odachi/agents.toml` and filters what `init` and `setup` write
- the wizard chooses the skills scope (home or project) and applies it
- the wizard chooses the voice (novato, senior or ci) and applies it through the same path as `level --to`, without copying the profile table
- the wizard turns the review on or off and can set the model of each lens; the resulting `odd/review.toml` keeps the `max_prompt_kb` and the `keep_*` it already had
- applying shows each operation with its result; one that fails is named and stops the rest, leaving no half configuration in silence
- cancelling before applying writes nothing: the tree is byte for byte the same
- the domain gains no dependency and knows nothing about the screens; the TUI lives in `cmd/odachi` or in an adapter and calls the existing operations
- the tests cover each screen's transition and the state it applies, with `teatest`; the no-TTY start has its test
- `go test ./...` is green and `odachi check --run "go test ./..."` is green
- Unreleased names the new verb and the dependency, patch (0259)

## Not included (first version)
- no per-client model picker (opencode's or claude's internal model): only the review lens models - back when the client's own model is what has to be chosen
- no backups, uninstall or community tools: that is a second product - back when the TUI manages more than the setup
- the binary installation (download and replace) stays in the scripts: the TUI configures the installed binary - back when the installer itself needs a screen
- `max_prompt_kb` and the store retention are not touched from the TUI - back when someone asks to tune the store's memory from the wizard
- no editing `brief.md`, the task doc, or opening a feature from the TUI: it is setup, not the whole product - back when the wizard becomes the default entry
- no mouse, no themes and no animation beyond the spinner - back when the screens are used daily and the feel matters
- no `--json` for the TUI: the flags remain the surface for machines - back when a script needs the same configuration

## Decisions taken
- The TUI is an adapter over the operations that already exist; the domain stays pure and dependency-free.
- Bubbletea + Lipgloss, the same stack gentle-ai uses: the binary stays static and makes no network call.
- A new `odachi tui` verb; the flags and `setup` keep their behavior.
- ADR 0001 ("no TUI") is superseded by a new ADR: it said screens are a second product; this one says installing and configuring cannot demand knowing seven flags. The CLI stays the interface for machines.
- The first version is installation and configuration, bounded: clients, skills, voice and review.
- The copy the human reads is Spanish; code, docs and commits are English.
- A review file is edited through the format that already exists, preserving what the TUI does not touch.
- The apply screen reports each operation; silence is not success.
- The screens follow `tui-design`: the terminal floor and the capture at 180x50 and 80x24.
