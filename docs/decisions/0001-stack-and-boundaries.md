# 0001 · Stack and boundaries

Status: accepted · Date: 2026-01 · Session: 1

## Context

The agent does the work; the human decides the vision. That split only survives
if the tool can run in CI, in a terminal, and inside any agent harness without
asking the human to install, configure, or babysit anything. The human we serve
may never have written software, and the one we serve next may have twenty years
of it — the same binary has to be tolerable for both.

Anything that requires a daemon, a language runtime, a package manager, or a
network call at run time is a way for the tool to become the work instead of
removing it.

## Decision

- One static Go binary. `go build` produces the whole product; there is no
  runtime to install and no process to keep alive.
- Pure domain + ports + adapters. The domain (`internal/domain/...`) imports
  only the standard library and sibling domain packages; everything that touches
  the filesystem, git, or the network lives in an adapter behind a port.
- Profiles and prompts are data, not code. `.odachi/profile.toml` changes
  ceremony; the four artifact formats are the contract.
- Messages the user reads are Spanish. Code, identifiers, comments, docs and
  commits are English.

## Alternatives rejected

- **Stateful review-authority ledger.** A stored "who may approve" record makes
  the tool a system of record with migrations, sync and recovery paths. The
  artifacts on disk already are the record.
- **17 hand-written agent adapters.** One per harness means seventeen things to
  keep current and seventeen ways to drift. A rendered `AGENTS.md` block plus a
  pre-commit hook covers the same ground with files the human can read.
- **Bundling a reviewer or model providers.** Shipping a model call would put a
  key, a bill and a network dependency inside the thing that must work offline.
- **A TUI.** Screens are a second product and a second set of bugs. The CLI plus
  the artifacts is the whole interface.

## Consequences

- The domain is testable without a filesystem or a network: pure functions in,
  conclusions out.
- The profile schema becomes the only extension point, so it is guarded
  structurally: `profile.Profile` cannot express a switch that disables a gate,
  and a reflection test fails the build if a field is added that could.
- The domain keeps zero dependencies, which the boundary guard test enforces
  before any domain code lands.
- Anything the user reads is Spanish and must name a command that resolves the
  block; every blocked result carries `Next`.
