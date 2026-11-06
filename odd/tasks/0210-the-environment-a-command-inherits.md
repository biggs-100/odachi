# the-environment-a-command-inherits
Base: fd32dc1d539ad7c0d69efeb2e68024c98acccbae

## Requirements (verbatim, from the user)
R1 · "`odachi doctor` names the PATH this process hands to a command when it passes what Windows can pass to a child, with the length, the limit and the repair, so a proof that fails for an inherited environment is not read as a broken suite"

## Done when
- check 15 exists, appended, `CheckCount` says 15, and the usage text and `docs/impact-map.md` say fifteen
- on Windows with a PATH of 8192 characters or more the check is a warning that names the length, the 8191 limit and the repair; at 8191 or less it is ok
- on a platform that does not cut the environment, the check is ok and says which platform it measured
- nothing trims, rewrites or persists the human's PATH, and the command a proof runs receives the same environment it received before
- `go test ./...` is green
- the changelog's Unreleased section names the check

## Steps
- [ ] T1 · doctor measures the PATH a child inherits and warns above the Windows limit   R1 · route:inline
- [ ] T2 · the report says fifteen, and the impact map names the new check   R1 · route:inline
- [ ] T3 · the changelog names the check   R1 · route:inline

## History
L1 · "`odachi doctor` names the PATH this process hands to a command when it passes what Windows can pass to a child, with the length, the limit and the repair, so a proof that fails for an inherited environment is not read as a broken suite"
