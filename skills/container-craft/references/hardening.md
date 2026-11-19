# The host is what a container can still reach

Capabilities, secrets and the baselines that audit them. The engine's defaults are
permissive on purpose, so every rule here is the exception you argue for and then
write down; the rules about running are in `running-containers.md`.

## The fourteen capabilities a container already has

- **The default set is fourteen caps, not zero**: CHOWN, DAC_OVERRIDE, FSETID, FOWNER, MKNOD, NET_RAW, SETGID, SETUID, SETFCAP, SETPCAP, NET_BIND_SERVICE, SYS_CHROOT, KILL, AUDIT_WRITE (moby v27.3.1, `oci/caps/defaults.go`).
- **`CAP_NET_RAW` and `CAP_DAC_OVERRIDE` are the two that bite**: raw sockets for a spoofed packet, and file permission checks bypassed for everything the process can see.
- **So "it runs without privileges" is a claim about fourteen**, and the review question is which of the fourteen the process used.
- **Observe before you drop**: run with the default set, record what the process needs, then add back only that (docs.docker.com/engine/security).
- **A drop with no observation is a guess**, and it fails in the code path nobody exercised before the release.

## `--privileged` is the host, spelled as a flag

- **All capabilities, no seccomp, no AppArmor, no SELinux, every device, `/sys` read-write**, and it "can get a root shell on the host" (docker run reference).
- **It is not "more permissions than it needed"**: it is the removal of every boundary this page is about, so the exception is a workload that is itself a container runtime.
- **When you take it, the host is the blast radius**, and that host is either dedicated or the exception is wrong.
- **A narrower grant exists for almost every reason `--privileged` gets typed**: one device, one capability, one path, named separately.

## Two flags that cost nothing

- **`no-new-privileges` blocks the setuid escalation path**, and a read-only rootfs blocks the write that completes most of them.
- **The OWASP Docker Security Cheat Sheet (post-Jan-2024, it cites CVE-2024-21626)** lists 14 rules including rootless mode and build-time secret leaks (owasp.org cheat sheets).
- **Read it as a diff against your compose file**, because every rule it lists is one decision you either took or defaulted.

## Rootless changes the failure modes, not just the owner

- **Rootless Docker runs the daemon and its containers in a user namespace**, needs `newuidmap`/`newgidmap`, and needs at least 65536 subuids (docs.docker.com/engine/security/rootless).
- **Rootless Podman hides one user's containers from root's Podman**, stores images in `~/.local/share/containers/storage`, and needs overlayfs on kernel ≥5.12.9 or `fuse-overlayfs` (podman.1).
- **It does not support NFS graphroots**, which is the storage decision that decides whether it is an option on that host at all.
- **It needs pasta to create a network device**, so a host without it has a container that starts and cannot route.
- **Count the operational cost before the security win**: the same image, a different uid mapping, and every mount now has to be readable by that mapping.

## The baselines, and what each one does not cover

- **CIS Docker Benchmark v1.9.0 (2026)** scores host, daemon and container configuration on Linux (cisecurity.org/benchmark/docker).
- **It says nothing about dependency hygiene in your Dockerfile**, so a perfect score is compatible with a base image full of CVEs.
- **NIST SP 800-190, final Sept 2017**, is still the risk taxonomy: image, registry, orchestrator, container (csrc.nist.gov/pubs/sp/800/190/final).
- **It still names clear-text secrets in images**, which is why the 2017 document still decides reviews in 2026.
- **Choose the baseline by what it measures**: CIS measures the host you configured, NIST names the risk you mitigate, and neither scans your dependencies.

## Three different things are called a secret

- **Swarm secrets are tmpfs, mTLS and encrypted raft** (docs.docker.com), so the value is not on disk in the manager.
- **Compose `secrets:` is a bind mount**, and the docs say declaring one "does not encrypt these files".
- **Kubernetes Secrets are unencrypted in etcd by default**, and "anyone who is authorized to create a Pod in a namespace can use that access to read any Secret in that namespace" (kubernetes.io/docs/concepts/configuration/secret).
- **So name which of the three you are relying on**, because only the first two of those three claims are about encryption.

## Environment variables leak by construction

- **`/proc/<pid>/environ` is readable at the same privilege tier as ptrace**, inherited by every child, and present in core dumps.
- **`docker inspect` returns `Config.Env` verbatim**, so the secret is in the output of a command people run in incidents.
- **Build args and build env "persist in the final image"** (docs.docker.com/build/building/secrets), which is why a build-time secret is a shipped secret.
- **So the choice is a mounted path or an inherited variable**, and only the first one has a rotation story that does not need a restart.

## Rotation is a deploy

- **A secret mounted as a volume updates eventually consistent**, so the process may read the old value for a while after the rotation.
- **A `subPath` mount never updates**, which is the one that looks mounted and is frozen at start.
- **A secret read from an environment variable needs a restart**, by construction, because the env is set when the process is created.
- **Plan the rotation as a rollout with its own verification**, or the rotation is a change nobody watched and the old credential stays live.
