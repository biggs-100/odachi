# tui-instalacion
Base: 8bd0450ee8ccb2d0300efa9c5f882c35bd49b1a7

## Requirements (verbatim, from the user)
R1 · "una tui de instalación y configuración al igual que gentle-ai"

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

## Steps
- [x] T1 · the configuration state and the ordered operations that apply it, with tests   R1 · route:inline · tree:8415d336856092ec8bd5dd31a78b37aec868c60e
- [x] T2 · the odachi tui verb, Bubbletea, the no-TTY refusal and the detection screen   R1 · route:inline · tree:6339255e56b51cae5a9c1db6ad0b83d8f9f8ac5d
- [ ] T3 · the client, skills-scope and voice screens, and cancel writes nothing   R1 · route:inline
- [ ] T4 · the review screen: on or off and the model of each lens, preserving the file   R1 · route:inline
- [ ] T5 · the apply screen: one line per operation, a failure stops and names what remains   R1 · route:inline
- [ ] T6 · the ADR that supersedes 0001, the help, the changelog and the impact map   R1 · route:inline
- [ ] T7 · the screen tests with teatest and the closing proof   R1 · route:inline

## History
L1 · "una tui de instalación y configuración al igual que gentle-ai"
