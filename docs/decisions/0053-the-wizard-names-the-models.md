# 0053 - The wizard names the models, the clients keep them

Status: accepted - Date: 2026-11 - Session: after v0.54.0

## Context

ADR 0051 bounded the first version of the wizard and left the per-client model pickers out, on the argument that a client's configuration belongs to that client. The human asked for them twice, and the second time after the hub shipped: the hour the wizard exists to remove is one where a person configures odachi in a screen and then edits two files it does not know, each with its own grammar, to say which model runs.

The verifier had the same hole from the other side. `odachi verify --exec` carried `{prompt}` and no `{model}`, so which model answered was free text the verifier wrote about itself in the record, and nothing in the project could name it.

## Decision

- **ADR 0051's "per-client model pickers, backups and uninstall stay out" is superseded for the model pickers.** The rest of 0051 stands: one verb, an adapter over the verbs that already exist, Bubbletea, the bounded scope.
- **One section of the wizard names the models**: the verifier's, and one per installed client. The rows read the file that stands, and an unchanged row writes nothing.
- **The wizard is still a caller, never a second writer.** `odd/review.toml` is written by `review.SetConfig`, and each client's file by `init`/`setup --model client=model`. The CLI keeps the machine surface: `odachi verify --model id` and `--model client=model` do by hand what the section does by key.
- **Another product's file is edited, never re-encoded.** `internal/domain/jsonmember` inserts or replaces one member and leaves comments, key order, indentation and line endings as they were, the promise the opencode editor already kept for `default_agent`. A value that is already there is named, not rewritten; a real change is backed up first.
- **The verifier's model is declared in `odd/review.toml`** as `verify`, beside `[lenses]`, because that file is already where a project says who answers what. The schema moves to 11: a key that arrived after the installed binary has to make that binary say upgrade, not unknown key.
- **The verifier resolves its model in three tiers**, the way the reviewer does: `--model`, then the `verify` key, then the session's model. A command that writes `{model}` with no model known is refused instead of started, and a model that would replace nothing is refused for the same reason.

## Alternatives rejected

- **Keep 0051.** The file does belong to the client, and that is exactly why the wizard has to write it: the human's hour is the product, and a wizard that configures half of it sends them back to two grammars.
- **One odachi-owned file listing the client models.** It configures nothing a client reads, so the wizard would show a decision nobody honours.
- **Write the model into the agent files odachi installs.** The chat model is not the agent's, it is the client's; and only some clients accept a per-agent model, so the section would mean two different things in two rows.
- **A CLI verb per client file.** Two verbs to keep in step with two formats; the flag on the verb that already installs is the same surface the wizard calls.

## Consequences

- Two files another product parses now carry `model`. A client that renames the key changes what the human runs, so the write is reported per file, backed up, and a failure pends one sentence naming the member to add by hand instead of failing the run.
- A client's model can be set and changed from the wizard but not removed: no verb removes that key, so the section refuses an empty value and names where the line lives, instead of showing a row that is empty while the file keeps its model.
- The verifier can be a different model from the reviewer, which is what "a second instrument" asks for, and the run names which model answered.
- The wizard gains a section to maintain. ADR 0001's cost stands: screens are a second product, and the bound is now a decision per file, not per version.
- A project that declares no `verify` renders the file it always rendered: the key is written only when a value is chosen.
