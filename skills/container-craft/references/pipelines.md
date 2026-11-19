# Pipelines

One artifact, promoted. What CI has to hold still so that staging and production
ran the same bytes, and what fails when it does not.

## Build once, promote the digest

- Promote a digest, not a rebuild: with `@sha256:` staging and production pull the same bytes, and a rebuild per environment is a second artifact wearing the same tag. (Docker best practices, 2026-11)
- What makes the claim checkable is provenance: each build attaches a record, so comparing what staging pulled with what production pulled is evidence instead of a tag name. (Docker SLSA provenance docs, 2026-11)

## Multi-arch builds

- A multi-arch tag is a manifest list: one tag, one image per platform, selected at pull time. (Docker multi-platform docs, 2026-11)
- QEMU under `buildx` needs no Dockerfile change and "can be much slower… especially for compute-heavy tasks". (Docker multi-platform docs, 2026-11)
- A native node per architecture is the docs' answer where you need "better support… that QEMU can't handle". (Docker multi-platform docs, 2026-11)
- Cross-compilation is the third: `FROM --platform=$BUILDPLATFORM` builds on the builder, and `TARGETPLATFORM`/`TARGETARCH` choose the output. (Docker multi-platform docs, 2026-11)

## The image store

- Manifest lists need the containerd image store: Docker Desktop and Engine 29.0+ default to it. (Docker multi-platform docs, 2026-11)
- On older classic storage the build needs the `docker-container` driver, and those images are not loaded into the local engine automatically. (Docker multi-platform docs, 2026-11)

## What breaks silently under multi-arch

- Anything that probes the machine at build time answers about the builder: `uname`, `nproc`, CPU features. (Docker multi-platform docs, 2026-11)
- A prebuilt native module is fetched for the builder's architecture, so the image carries a binary the target cannot load. (Docker multi-platform docs, 2026-11)

## CI caches

- A CI cache is a code-execution surface: GitHub's caches are shared by branch and tag. (GitHub dependency caching docs, 2026-11)
- "Any run that can read a cache restores its contents as-is", and a poisoned cache "can lead to code execution in trusted workflows". (GitHub dependency caching docs, 2026-11)
- So an untrusted run never writes a cache that a release workflow reads. (GitHub dependency caching docs, 2026-11)

## The registry at deploy time

- `ImagePullBackOff` retries with backoff up to 5 minutes, so a registry outage at deploy is a slow failure. (Kubernetes images docs, 2026-11)
- `imagePullPolicy` defaults to `IfNotPresent` for any tag that is not `:latest`, so a node that already holds the digest survives the outage. (Kubernetes images docs, 2026-11)
- A tag rename is a deploy you did not test: the docs' own warning is "a mix of Pods running the old and new code". (Kubernetes images docs, 2026-11)

## Garbage collection

- Registry GC is stop-the-world: run it with the registry read-only or stopped, or in-flight pushes lose layers. (Distribution GC docs, 2026-11)
- `--delete-untagged` is opt-in, so untagged manifests accumulate until someone asks for them gone. (Distribution GC docs, 2026-11)

## Pull-through caches

- One upstream per mirror: a pull-through cache serves the registry it was configured for and nothing else. (Distribution mirror recipe, 2026-11)
- Tag pulls revalidate against the upstream, so a cache is never a pin. (Distribution mirror recipe, 2026-11)
- Credentials on the mirror expose that private content to everyone who can read the mirror. (Distribution mirror recipe, 2026-11)
- `delete` must be enabled for the mirror to shrink; without it the cache only grows. (Distribution mirror recipe, 2026-11)
- The daemon's `registry-mirrors` is Docker Hub only, so every other registry goes in the image reference. (Distribution mirror recipe, 2026-11)

## The failures

| failure | symptom | what catches it earlier |
|---|---|---|
| a rebuild per environment | staging passed, production fails on the same commit | promote the digest and compare provenance |
| QEMU on a compute-heavy build | the pipeline times out on one platform only | native nodes, or cross-compilation |
| build-time `uname`/`nproc` | the image is wrong for the target, silently | build from `BUILDPLATFORM` and never probe the machine |
| a prebuilt native module | the container exits at start on one platform only | fetch it per `TARGETARCH` |
| a poisoned CI cache | untrusted input executes in a trusted workflow | scope cache writes to trusted branches |
| a registry outage at deploy | `ImagePullBackOff`, backing off to 5 minutes | `IfNotPresent` and nodes that already hold the digest |
| GC while pushes are in flight | pushed layers missing, image unreadable | read-only or stopped GC |
| a tag moved under a running deployment | old and new Pods in the same ReplicaSet | deploy by digest |
