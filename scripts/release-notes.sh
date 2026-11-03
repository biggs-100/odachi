#!/usr/bin/env bash
# Prints the body of a release: the changelog section for the tag, verbatim, and
# the install block for that tag. One source, so the notes cannot drift.
#
# ODACHI_CHANGELOG names the file; the tests point it at a fixture. A tag with no
# section is refused: a release note nobody wrote is a release nobody can read.
set -euo pipefail

tag="${1:-}"
if [ -z "$tag" ]; then
  echo "usage: scripts/release-notes.sh <tag>" >&2
  exit 2
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
changelog="${ODACHI_CHANGELOG:-$root/CHANGELOG.md}"

if [ ! -f "$changelog" ]; then
  echo "no $changelog: the notes come from it" >&2
  exit 1
fi

section="$(awk -v tag="$tag" '
  index($0, "## " tag " ") == 1 { found = 1; print; next }
  found && /^## / { exit }
  found { print }
' "$changelog")"

if [ -z "$(printf '%s' "$section" | tr -d '[:space:]')" ]; then
  echo "no section \"## $tag \" in $changelog: write the notes before the tag" >&2
  exit 1
fi

printf '%s\n' "$section"

sed -e "s|@TAG@|$tag|g" -e "s|@VERSION|@$tag|g" <<'EOF'

## Install

**This release requires [cosign](https://docs.sigstore.dev/cosign/system_config/installation/) on your PATH.**

- **Windows:** `powershell -ExecutionPolicy Bypass -File (irm https://raw.githubusercontent.com/biggs-100/odachi/@TAG@/scripts/install.ps1 -OutFile install.ps1; './install.ps1')`
- **macOS / Linux:** `curl -fsSL https://raw.githubusercontent.com/biggs-100/odachi/@TAG@/scripts/install.sh | bash`
- **Go:** `go install github.com/biggs-100/odachi/cmd/odachi@VERSION`
EOF
