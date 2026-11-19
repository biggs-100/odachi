# The run, and the decisions it takes for you

The body gives the shape of the decision. This page is the run itself: what the engine does
by default, where the default is a decision nobody took, and the version that makes each
line current. Read a line as a question about the command in front of you.

## Restart is a bet on the exit code

- **`on-failure` watches the exit code and nothing else**: a non-zero exit restarts, and it does not re-arm after a daemon restart, where `always` does (Docker docs, 2026-11).
- **A policy arms only after ten seconds up**, and backoff doubles from 100 ms to a one-minute cap and resets after a ten-second run (Docker docs, 2026-11).
- **`on-failure:10` counts consecutive failures**, so an app that exits 0 when it has failed never triggers the policy at all (Docker docs, 2026-11).
- **`always` is the policy for a container that must survive a daemon restart**, and neither policy is a substitute for the exit code that caused the loop (Docker docs, 2026-11).

## Health is a fact the engine reports, not an action it takes

- **`HEALTHCHECK` sets starting, healthy or unhealthy and emits a `health_status` event; it restarts nothing** (Dockerfile reference, 2026-11).
- **0 and 1 are the whole vocabulary and 2 is reserved**, so a probe that leaks an interpreter's exit 2 is using a code that means nothing (Dockerfile reference, 2026-11).
- **The defaults you inherit are `--interval=30s` and `--retries=3`**, and a timed-out probe is SIGKILLed, so a hung probe reads as a failure with no output (Dockerfile reference, 2026-11).
- **`--start-period` is the flag that exists for a boot slower than the probe interval**, and naming it is the difference between a status and a guess (Dockerfile reference, 2026-11).
- **Nothing in the engine restarts, drains or gates traffic on `unhealthy`**: the restart, the removal from the load balancer and the page are three things you write (Dockerfile reference, 2026-11).
- **Podman makes the missing behaviour explicit**: `--health-on-failure` defaults to `none` and you pick `kill`, `restart` or `stop`; Docker has no equivalent, so the choice lives in your entrypoint or your orchestrator (podman-run(1), 2026-11).

## `depends_on` orders a start, not a recovery

- **The default waits for running, not for ready**, so the first request can still beat the database's first accept (Compose docs, 2026-11).
- **`condition: service_healthy` gates on the healthcheck and `service_completed_successfully` on exit 0**: both are contracts you wrote, not ones Compose infers (Compose docs, 2026-11).
- **It applies at `up` and not to a later restart**, because by then the dependency is already running and nothing re-checks it (Compose docs, 2026-11).
- **"Retry until the database accepts connections" cannot be expressed in the file**, so it is application code with a backoff and an attempt cap (Compose docs, 2026-11).

## A stop is three knobs and one handler

- **`STOPSIGNAL` chooses the signal, `--stop-timeout` the wait, and PID 1's handler decides whether anything happens**: three independent knobs, and the third is the one nobody sets (Dockerfile reference; docker run reference, 2026-11).
- **`--stop-timeout` defaults to 10 s on Linux and 30 s on Windows, then SIGKILL**, and `-1` waits forever, which is how a deploy never finishes (docker run reference, 2026-11).
- **The kernel delivers only the signals namespace init handles**: SIGKILL and SIGSTOP from an ancestor are the exception, which is why neither can be trapped (pid_namespaces(7), 2026-11).
- **`STOPSIGNAL` does not change what Ctrl+C sends**, so a container that stops cleanly under `docker stop` can still die mid-write in a terminal (Dockerfile reference, 2026-11).

## PID 1 is also the reaper

- **A PID 1 that never `wait()`s accumulates orphans**, and a busy enough container starves the host of PIDs (docker run reference, 2026-11).
- **`docker run --init` puts docker-init, which is tini, in front of the process as PID 1**, and that is the whole fix for a program that only spawns (docker run reference, 2026-11).
- **When namespace init dies the kernel SIGKILLs every process in the namespace**, so "the container died when its shell exited" is the kernel finishing the job (pid_namespaces(7), 2026-11).

## Limits: two budgets and two meanings

- **`--memory=300m` with no `--memory-swap` buys 300m of memory plus 300m of swap**, and the belief that it caps the total is the common error (Docker docs, 2026-11).
- **`--memory-swap` equal to `--memory` is how you say "no swap", and `-1` means unlimited**, so the flag has no off switch other than a number you compute (Docker docs, 2026-11).
- **With no constraints the OOM killer chooses among everything on the host**, Docker and other applications included (Docker docs, 2026-11).
- **`--cpus=1.5` is `--cpu-period=100000 --cpu-quota=150000`**: a ceiling on the cycles this container can use, and nothing about priority (Docker docs, 2026-11).
- **`--cpu-shares` is a soft weight enforced only when cycles are constrained**: it is priority under contention and a reservation of nothing (Docker docs, 2026-11).

## The network decides who finds whom

- **The default bridge has no DNS**: containers there resolve each other by IP, which is an address that changes the next time one is created (Docker docs, 2026-11).
- **A user-defined network gives name resolution**, so the name written in the config is the name that resolves (Docker docs, 2026-11).

## Publishing is a firewall decision

- **`-p 8080:80` binds `0.0.0.0`**, and Docker writes its own rules in the `nat` table, ignoring the host firewall configuration in practice (docker run reference; Docker docs, 2026-11).
- **`-p 127.0.0.1:8080:80` is the form that keeps the port on the host**, and real filtering belongs in the `DOCKER-USER` chain, not in the host's default policy (Docker docs, 2026-11).
- **Before Docker 28.0.0 another host on the same L2 could reach a port published to `127.0.0.1`** (moby#45610): on an older daemon, localhost is not localhost (docker run reference, 2026-11).

## Storage is an ownership question

- **A volume is Docker's to own and a bind mount is yours**: the volume can be pre-populated and migrated, the bind carries host ownership and permissions verbatim (Docker docs, 2026-11).
- **A bind mount is writable by default, and `ro` is the flag you add when the container only reads**: a writable config mount is a config the process can edit under you (Docker docs, 2026-11).
- **On SELinux `:z` and `:Z` relabel the host path**, so `-v /home:/x:Z` relabels `/home` and "renders your host machine inoperable" (Docker docs, 2026-11).
- **`--mount` cannot set an SELinux label**, so on a relabeling host the long syntax loses the option the short one had (Docker docs, 2026-11).

## `down -v` is the data-loss command

- **`down` removes containers and networks and leaves volumes alone; `-v` removes them**, so the flag is the distance between a restart and a restore (Compose docs, 2026-11).
- **An anonymous volume survives `down` and is not reattached on the next `up`**, because it has no stable name, so the recreated container silently gets an empty one (Compose docs, 2026-11).

## Logs are a disk decision

- **`json-file` without rotation is the default**, and Docker's own docs name it as a cause of disk exhaustion (Docker docs, 2026-11).
- **`local` rotates by default, and a daemon setting only reaches containers created after it**, so a fix is retroactive only after a recreate (Docker docs, 2026-11).

## Compose: the file remembers what the command does not

- **The top-level `version` key is obsolete and only warns**: the schema it selected is now assumed (compose-spec, 2026-11).
- **A service under `profiles` does not start until you activate it**, so a core service carrying a profile is one `up` skips (Compose docs, 2026-11).
- **Compose `secrets` mount at `/run/secrets/<name>`, and the docs say declaring one "does not encrypt these files"**: it is file placement, not a secret store (Compose docs, 2026-11).
- **`up --wait` waits for running or healthy and means nothing without healthchecks**, so it promises exactly what you probed (Compose docs, 2026-11).
- **`env_file` is interpolated by the Compose CLI and not by `docker run --env-file`**, so the same file read two ways yields two different values (Compose docs, 2026-11).

## What is never approved

| the default you are accepting | what it costs | the decision that replaces it | source |
| --- | --- | --- | --- |
| `--privileged` | the container is the host with a namespace in front | the capability the process proves it needs, added back | Docker docs, 2026-11 |
| the docker socket inside a container | the API of a privileged process, and read-only is a filesystem flag | a narrow proxy, or nothing | `docker-and-podman.md`, OWASP Rule #1 |
| stateful data in an anonymous volume | the next `up` attaches a fresh empty one | a named volume, or a bind you own | Compose docs, 2026-11 |
| `restart: always` on a crash loop | one restart per minute, forever, and no page | `on-failure` with a cap, and an alert that reads it | Docker docs, 2026-11 |
| ignoring the exit code | the 0 that means failure is the policy that never fires | a non-zero exit, and a healthcheck that reads it | Docker docs, 2026-11 |
| `json-file` with no rotation | a full disk and the engine that writes to it | `local`, or `max-size` with `max-file` | Docker docs, 2026-11 |
| a `docker exec` fix with no record | it disappears at the next recreate, and it is already folklore | a commit, a build, a pipeline | docker run reference, 2026-11 |
| `-p 8080:80` on a public host | a port the host firewall never saw | `-p 127.0.0.1:` unless the exception was argued | docker run reference, 2026-11 |

Each row is a default that answered a question nobody asked: the policy by restarting, the
volume by staying anonymous, the publish by binding everywhere. The last column is the
decision, and a decision needs an owner, so name one before the next deploy.
