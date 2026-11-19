# The failures

Every row is a defect a person paid for once. The last column is the one that
matters: what would have caught it **before** it happened. If a column is empty
for you, you have a rule that only works after the incident.

| The mistake | What you see | The fix | What catches it earlier | Source |
|---|---|---|---|---|
| `json-file` logs with no rotation | the host fills its disk, evictions, a daemon that stops answering | `local` driver, or `max-size`/`max-file` in `daemon.json` | a disk and inode alert on `/var/lib/docker`, and a log-rate metric | Docker logging, 2026-11 |
| a daemon restart during an upgrade | every container dies at once, not only the one being updated | `live-restore` for patch upgrades, restart policies, or a scheduler | a change process that knows the daemon is a single point of failure | Docker `live-restore`, 2026-11 |
| `docker system prune -a` at 3 a.m. | the warm build cache is gone, the next build is ten times slower, images you did not name vanished | a scoped `image prune`/`builder prune --filter until=` | `--dry-run` first, and knowing that prune always takes the build cache, cache mounts included | Docker prune, 2026-11 |
| untagged manifests accumulate | registry storage grows with no new tags to explain it | scheduled GC with `--delete-untagged` and a retention policy | a storage-growth alert and a manifest count | Distribution GC, 2026-11 |
| the registry is down at deploy time | `ImagePullBackOff`, backoff to five minutes, rollouts stall | a mirror or pre-pulled digests on the nodes, more than one region | pull-error rate and pull latency, and the mirror hit ratio | Kubernetes images, 2026-11 |
| a container in the wrong timezone | logs off by hours, token `nbf`/`exp` rejections, cron at the wrong hour | mount the host time, or set `TZ` **and** install tzdata | a skew check between host and container timestamps | distroless base, 2026 |
| limits set once, never revisited | OOMKill at peak, or throttling nobody sees | revisit from the working-set p95; memory limit is not the request | `container_oom_events_total` and CPU-throttled seconds | Docker resource constraints, 2026-11 |
| a fix applied with `docker exec` | it works until the next deploy, then the bug returns | reproduce it, change the Dockerfile, rebuild, redeploy | a rule that no production change exists without a commit, and drift detection | practice, not a doc |
| a rebuild per environment | "it works in staging", and the digest there is not the one in production | build once, promote the same digest | provenance compared across environments | SLSA provenance v1, 2026 |
| a secret passed as `ARG` or `ENV` | it appears in `docker history`, in `mode=max` provenance, or in `docker inspect` | `RUN --mount=type=secret`, or a file the process reads | a reviewer reading the Dockerfile, and a scan of the history | Docker build secrets, 2026-11 |
| shell-form `CMD` | every stop waits ten seconds and ends in SIGKILL | exec form, and `--init` when anything forks | a test that stops the container and asserts the process received SIGTERM | `pid_namespaces(7)` and tini, 2026-11 |
| `-p 8080:80` on a public host | the service is reachable from the internet, and the firewall never blocked it | bind `127.0.0.1` and put a proxy in front, or filter in `DOCKER-USER` | an external port scan of the host | Docker packet filtering, 2026-11 |
| a healthcheck believed to be an action | an unhealthy container keeps serving traffic | Podman's `--health-on-failure`, or the orchestrator's readiness probe | a readiness signal that actually gates traffic | Podman `run(1)`, 2026 |
| `--memory` without `--memory-swap` | swap thrash, then an OOM kill at a size nobody budgeted | set both, with `--memory-swap` equal to `--memory` | an alert on memory plus swap against the budget | Docker resource constraints, 2026-11 |
| a glibc binary on Alpine | "not found", or DNS that resolves differently than in staging | match the libc, or use a base that ships glibc | a test that runs the built image, not only that it builds | musl docs, 2026-11 |
| a cache mount trusted in CI | fast on the laptop, cold on every CI runner | a registry cache with `--cache-to`/`--cache-from` | build time tracked in CI as a number, not a feeling | Dockerfile reference, 2026-11 |
| stateful data on an anonymous volume | it survives `down`, then a recreate silently mounts an empty one | a named volume, declared and backed up | a restore you have actually performed once | Compose `down`, 2026-11 |
| a tag that moved | a mix of old and new instances behind one service | digest pinning with an update bot | an admission policy that only accepts digests, and drift detection | Kubernetes images, 2026-11 |

## The documents behind this page

Everything above is documented behaviour, checked 2026-11-18:

- Docker: engine resource constraints, logging configuration, packet filtering and
  firewalls, storage, the run and compose references, build best practices,
  `live-restore`, the distribution GC recipe.
- Kubernetes: images, Secrets, admission controllers.
- Podman: `podman-run(1)`, `podman-systemd.unit(5)`, rootless notes.
- Standards: CIS Docker Benchmark v1.9.0 (2026), NIST SP 800-190 (Sept 2017),
  OWASP Docker and Kubernetes cheat sheets (post-Jan-2024), SLSA v1.2 (2026),
  CycloneDX 1.7 (2025-10-21), SPDX 3.0 (2025-08), Notation v1.1.0 (2024).

A claim in this table is true for the version named, and the engine's default is
not the other engine's default. When a row surprises you, check the version before
you check the person.
