---
name: container-craft
description: Trigger: a Dockerfile, a compose file, a `docker run` or `podman` command, an image tag or a registry, a container that died at 3 a.m. Decisions a senior makes, and the version behind each rule.
---

## When

An image is built or tagged, a container is run, a compose file is written, or a
container misbehaves. What a value is and who owns it is `data`; how it is stored
is `db-craft`; how it reaches a host is this door.

## Rules

- **A tag is a pointer and a digest is the bytes**: `:1.4` moves under you, `@sha256:` cannot, and pinning with no update bot trades CVEs for determinism.
- **A secret never enters a layer**: `ARG` shows in `docker history` and in `mode=max` provenance, `ENV` survives `unset`; it is a mount or a file the process reads.
- **The last stage carries artifacts, not toolchains**: multi-stage, and what crosses `COPY --from` is the only thing that ships.
- **The dependency layer comes before the source**: `COPY . .` ahead of the install reinstalls everything on a one-line change.
- **The base you can operate beats the base you can build**: distroless has no shell, musl resolves DNS unlike glibc, `scratch` has no CA certificates.
- **`exec` form, and `--init` when anything forks**: shell-form `CMD` leaves `/bin/sh` at PID 1, which ignores SIGTERM, so each stop costs 10 seconds and a SIGKILL.
- **A restart policy is not a health signal**: the Docker engine's health status changes nothing by itself -Podman's `--health-on-failure` and a compose `service_healthy` are where it becomes an action-, and `on-failure` never fires on exit 0.
- **Memory has two budgets and CPU has two meanings**: `--memory` without `--memory-swap` allows the same again in swap; `--cpus` is a ceiling, `--cpu-shares` a weight.
- **Publish to `127.0.0.1` unless you argued the exception**: `-p 8080:80` binds `0.0.0.0` and writes a `nat` rule the host firewall never sees.
- **Data on a named volume, source and config on a bind mount**: `down -v` deletes volumes, and an anonymous one is silently replaced by an empty one.
- **The default capability set is fourteen caps, not zero**: `--privileged` gives the host; drop all and add back what the process proves it needs.
- **Logs fill the disk by default**: `json-file` does not rotate, and a daemon setting only reaches containers created after it.
- **Build once and promote the digest**: rebuilding per environment means staging tested an artifact production does not run.
- **`docker exec` is how a fix becomes folklore**: if it matters it goes in the Dockerfile, through a commit and a pipeline.

## Decision gates

| Situation | Load |
|---|---|
| a Dockerfile, a cache, a base image or an image size | `references/images-and-build.md` |
| `docker run`, compose, signals, limits, volumes or network | `references/running-containers.md` |
| Docker against Podman, rootless, Quadlet or a compose stand-in | `references/docker-and-podman.md` |
| a tag, a registry, a scanner, an SBOM, signing or provenance | `references/supply-chain.md` |
| capabilities, secrets, an audit baseline or a hardened host | `references/hardening.md` |
| CI, a build cache, multi-arch or a deploy that pulls | `references/pipelines.md` |
| something already broke | `references/the-failures.md`: the mistake, its symptom, what would have caught it |

## Output

The Dockerfile, the run command or the compose service; the failure it prevents; the version that makes the rule true. A base-image choice returns what it removes -shell, libc, CA certificates, tzdata- not a size number.

## Never

- Approve `:latest` without a digest, `--privileged`, a secret in `ARG`/`ENV`, or the docker socket inside a container.
- Recommend a smaller base without naming what it takes away.
- Repeat a container claim without its version: half of what everyone knows is the other engine's default.
- Write a recipe for one stack: the rule is the decision, the stack is the example.

## References

- `references/images-and-build.md` - multi-stage, cache, base images, size, multi-arch.
- `references/running-containers.md` - lifecycle, signals, limits, storage, network, compose.
- `references/docker-and-podman.md` - what changes when the engine is Podman: rootless, Quadlet, networking.
- `references/supply-chain.md` - tags, registries, scanners, SBOM, signing and provenance.
- `references/hardening.md` - capabilities, secrets, rootless and the audit baselines.
- `references/pipelines.md` - build once, promote a digest, caches and multi-arch in CI.
- `references/the-failures.md` - the catalogue: the mistake, what you see, what catches it earlier.
