# Images and build

What the final stage carries, what the builder forgets, and what each base takes
away. The evidence behind the body's rules, not the rules again.

## Multi-stage

- The final stage ships only what `COPY --from` brings across; every install in an earlier stage stops at that boundary. (Docker multi-stage docs, 2026-11)
- Stages that do not depend on each other build in parallel, so splitting a stage costs wall-clock only where the graph joins. (Docker multi-stage docs, 2026-11)

## The cache is per instruction

- Invalidation is per instruction and its inputs, not per file touched: one changed byte under `COPY . .` re-runs that step and every step after it. (Docker cache optimization docs, 2026-11)
- `COPY . .` ahead of the dependency install makes a one-line source edit reinstall everything; copy the manifest, install, then copy the source. (Docker cache optimization docs, 2026-11)
- `.dockerignore` shrinks the context the builder receives and applies to the whole context, not to one `COPY`. (Docker cache optimization docs, 2026-11)
- A fat context is paid on every build, cached or not, because the context is sent before the first instruction runs. (Docker cache optimization docs, 2026-11)

## Tags and digests

- A tag is a pointer that can move; a digest pins the exact bytes. (Docker best practices, 2026-11)
- Pinning a digest is, in the docs' words, "opting out of automated security fixes", so a digest with no update bot is a CVE with a fixed date. (Docker best practices, 2026-11)
- In a build, `--pull` is the only thing that re-resolves the base tag your Dockerfile names; a digest cannot be refreshed, only replaced, and a pulled tag is refreshed by pulling it. (Docker best practices, 2026-11)

## Non-root

- OWASP's Rule #2 is a non-root user; set the UID/GID explicitly, because a rebuild assigns nondeterministic ones. (OWASP Docker Cheat Sheet, 2026-11)
- `useradd --no-log-init`: without it the Go `archive/tar` sparse-file bug lets a build user's log file fill `/var/log/faillog` and the layer with it. (Docker best practices, 2026-11)

## Entrypoint and shell

- `ENTRYPOINT` is the command and `CMD` its default arguments; a `CMD` alone is replaced whole by `docker run <image> <args>`. (Docker best practices, 2026-11)
- A wrapper script ends in `exec "$@"`, so the app replaces the shell as PID 1 and receives the signal. (Docker best practices, 2026-11)
- Exec form is mandatory where there is no shell: distroless images "by default do not contain a shell", so a shell-form `CMD` fails at start. (distroless README, 2026-11)

## Arguments, environment, secrets

- `ARG` leaks twice: in `docker history` and in `mode=max` provenance. (Dockerfile reference, 2026-11)
- `ENV` survives `unset`: clearing it in a later instruction leaves the value in the layer's bytes. (Dockerfile reference, 2026-11)
- A secret is a mount: `RUN --mount=type=secret,id=token` places it at `/run/secrets/token` for that one instruction and it is gone after. (Docker build secrets, 2026-11)
- Secret mounts never appear in provenance, which is why a token the build needs is a mount and not an argument. (Docker build secrets, 2026-11)
- Cache mounts are local only: `--cache-to` does not export them and `--cache-from` does not restore them, so a fresh CI runner starts empty. (Dockerfile reference, 2026-11)
- `apt` behind a shared cache mount needs `sharing=locked`. (Dockerfile reference, 2026-11)

## Layers and failures

- `apt-get update` in its own `RUN` is a cache bug: the lists layer is reused while the index behind it moves, so the image installs last month's package. (Docker best practices, 2026-11)
- `update`, `install` and `rm -rf /var/lib/apt/lists/*` are one `RUN`, one layer. (Docker best practices, 2026-11)
- A pipe swallows failures: `/bin/sh -c` returns the last command's status, so `curl … | sh` succeeds when `curl` did not. (Docker best practices, 2026-11)
- `set -o pipefail` at the top of a `RUN`, or the step is green for the wrong reason. (Docker best practices, 2026-11)
- `rm` in a later layer frees nothing: the bytes stay in the earlier layer and are still retrievable. (OWASP Docker Cheat Sheet, 2026-11)

## Choosing the base

Size is not the decision; what the base removes is.

| base | size of the Debian 13 tag, 2026 |
|---|---|
| `gcr.io/distroless/static` | ≈2 MiB |
| `alpine` | ≈5 MiB |
| `debian` | ≈124 MiB |

- `scratch` carries no CA certificates, no tzdata, no passwd entry and no shell, so the first HTTPS call has nothing to verify against. (Docker base images docs, 2026-11)
- `distroless/static` is the shortest honest answer to `scratch`: it adds ca-certificates, `/etc/passwd`, `/tmp` and tzdata. (distroless base README, 2026-11)
- musl is a different libc, not a smaller glibc: its resolver queries all nameservers in parallel and takes the first answer. (musl functional differences, 2026-11)
- `ndots` and `search` handling also differ from glibc, so a name resolves by other rules than in the Debian image you tested. (musl functional differences, 2026-11)

## Reproducibility, provenance, SBOM

- Pinning is not reproducibility: the docs' own `mode=min` example reports `"reproducible": false`. (Docker SLSA provenance docs, 2026-11)
- Two builds of one commit are two digests unless timestamps are controlled; `SOURCE_DATE_EPOCH` is how they are. (reproducible-builds.org, 2026-11)
- `--sbom=true` scans the final stage only, because it covers what that stage carries. (Docker SBOM attestations, 2026-11)
- `FROM scratch` hides build-time dependencies from that scan unless `BUILDKIT_SBOM_SCAN_STAGE=true`. (Docker SBOM attestations, 2026-11)
- Provenance takes `mode=min|max` and `version=v0.2|v1`; v0.2 is the docs' default as of 2026-11. (Docker SLSA provenance docs, 2026-11)
- `mode=max` "exposes the values of build arguments", so an `ARG` in a `max` build is published. (Docker SLSA provenance docs, 2026-11)

## Anti-patterns

| what is wrong | why it bites | instead |
|---|---|---|
| `:latest` in production | a rollback redeploys the new code; k8s pulls `Always` for `:latest` | a digest, refreshed by an update bot |
| a secret in `ARG` or `ENV` | it is in `docker history` and in `mode=max` provenance | `RUN --mount=type=secret` |
| `COPY .` then `rm` the leftovers | the bytes stay retrievable in the earlier layer | `.dockerignore`, so they never enter |
| `apt-get update` alone in a layer | the cached lists outlive the index behind them | one `RUN`: update, install, `rm -rf /var/lib/apt/lists/*` |
| `COPY . .` before the dependency install | every source edit reinstalls the world | copy the manifest first |
| build tools in the final image | they ship | multi-stage, `COPY --from` |
| root | the process holds everything the image can reach, so a compromise does too | `USER` with an explicit UID/GID |
| a mutable base consumed by CI | in a build, `--pull` is the only refresh, and a rebuild is a new image | a digest plus an update bot |
| a cache mount treated as cache | `--cache-to` does not export it | expect it empty on a fresh runner |
| believing `rm` shrinks the image | the layer keeps the bytes | it does not; build small in a stage |
