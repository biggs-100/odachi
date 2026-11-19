# The engine owns the defaults, and it is not the same engine

Which sentence is true depends on which binary you type: Podman's defaults are not
Docker's, and a Compose file is not a deploy strategy on either. Decide the engine
first, then the rules below follow from that one choice.

## What changes when the engine is Podman

- **Daemonless**: there is no socket to mount and no daemon restart that kills the containers, because the process tree is the lifecycle.
- **So the lifecycle owner is systemd, not a daemon**: a container that must survive a reboot is a unit that was enabled, not a container that was started.
- **Rootless by default on many distributions**, which moves the uid mapping, the storage path and the failure modes at once.
- **One user's containers are invisible to another user's**, including root's, so "where is the container" has a user as part of the answer.
- **Images live under `~/.local/share/containers/storage`**, per user, which makes image provisioning a per-user step rather than a host step.
- **Decide the engine per host and write it down**, because half of the advice in circulation is the other engine's default.

## Health behaviour is a flag, not an inference

- **`--health-on-failure none|kill|restart|stop`, default `none`** (podman.1): the healthcheck runs, records the status, and the default does nothing about it.
- **`HealthOnFailure=kill` "integrates best with systemd"**, because the unit restarts what the engine killed, under the policy the unit declares.
- **Name the policy where the healthcheck is defined**, or the check is a metric with no consequence.
- **`restart` inside a unit and `Restart=` on the unit are two owners of one decision**: pick one, and let the other stay out.

## Quadlet is the production story

- **`.container`, `.volume`, `.network`, `.pod` and `.kube` files** are read by a systemd generator at boot and at `daemon-reload` (docs.podman.io, `podman-systemd.unit.5`).
- **The generator produces real `.service` units**, so `systemctl status`, `journalctl -u` and `Restart=` are the operational surface, not `podman ps`.
- **The generated `Restart` differs by type**: pods default to `Restart=on-failure`.
- **`ExitPolicy=stop` means a clean pod exit does not restart**, which is correct for a batch pod and wrong for a service that exits 0 on a dependency error.
- **So a pod you expect to come back needs `ExitPolicy` named**, not assumed; the default is the one that leaves it down.
- **Changing a unit file needs a `daemon-reload` before the unit sees it**: edit, reload, restart, in that order.

## `podman-compose` is not Compose

- **It is rootless and daemonless**, so the file is compatible and the guarantees behind it are not the same ones.
- **Name resolution needs netavark plus aardvark-dns** (or dnsname), or a service reaches its neighbour by container name in one engine and not the other.
- **Its own README points production-like single-machine users to k3s**, which is the maintainers answering the question directly.
- **Treat it as a dev-loop tool**, and if the host must survive a reboot, that is the sentence the maintainers left you.

## Docker Compose in production is a single-host supervisor

- **Docker's docs endorse it for one server, with a separate production file**, and both halves of that sentence matter (docs.docker.com).
- **It has no scheduler, no rolling deploy and no multi-node**: `up` recreates the service, and that recreation is downtime.
- **Both statements are true and the contest is scope**: endorsed for one host, insufficient for the fleet, so name which one you have.
- **Decide what you give up before the first deploy**: rolling replacement, host failure and placement are the three, and each one is a tool you are not using.

## `host` networking is an exception you argue

- **It deletes the network namespace and with it port isolation**, so two containers cannot both want 8080 and the host's policy applies to both.
- **It is defensible for performance-sensitive workloads and for port-range-heavy ones**, where the per-packet path or the range is the requirement.
- **It is also the answer for rootless Podman without pasta**, where there is no network device to create.
- **An exception is a written reason, not a default that spread**: the compose file should say why, or the next reader removes the wrong thing.

## The docker socket in a container is host root

- **Mounting it grants the API, and the API is root**: `:ro` "does not replace API authorization" (docs.docker.com).
- **So `-v /var/run/docker.sock:/var/run/docker.sock:ro` is host root with extra steps**, and it is the sentence to quote when the security review asks.
- **A scoped socket proxy is the only defensible form**, with the endpoints it exposes listed and the exception documented beside it.
- **Ask whether the container needs the API or one endpoint**, because that answer decides between a proxy and a redesign.

## `live-restore` is an upgrade convenience

- **Docker restricts it to patch releases** (docs.docker.com/engine/daemon/live-restore), so it covers the restart you do monthly and not the one you do for a feature.
- **It keeps containers running while the daemon is down**, which is not the same as keeping them healthy while a node is.
- **A scheduler with health checks is the real answer**, at a much higher cost: a control plane, a second host, and something that owns placement.
- **Choose by what you are buying**: minutes of upgrade downtime against an order of magnitude in operational surface.
