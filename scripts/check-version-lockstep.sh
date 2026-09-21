#!/usr/bin/env bash
# Release guard: the akb version lives in three files and they must all agree.
# Exits non-zero, listing every disagreement, before a release mutates the repo.
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

makefile_version="$(sed -n 's/^VERSION[[:space:]]*:=[[:space:]]*\(.*\)$/\1/p' "$root_dir/Makefile" | head -n 1 | tr -d '[:space:]')"
flake_version="$(sed -n 's/^[[:space:]]*akbVersion[[:space:]]*=[[:space:]]*"\([^"]*\)".*$/\1/p' "$root_dir/flake.nix" | head -n 1)"
agents_version="$(sed -n 's/^\*\*Status:\*\*[[:space:]]*v\([^[:space:]]*\).*$/\1/p' "$root_dir/AGENTS.md" | head -n 1)"

report() {
  printf '  %-24s %s\n' "$1" "${2:-<not found>}" >&2
}

if [ -z "$makefile_version" ] || [ -z "$flake_version" ] || [ -z "$agents_version" ]; then
  echo "error: version lockstep check failed: missing version string" >&2
  report "Makefile VERSION" "$makefile_version"
  report "flake.nix akbVersion" "$flake_version"
  report "AGENTS.md Status" "$agents_version"
  exit 1
fi

if [ "$makefile_version" != "$flake_version" ] || [ "$makefile_version" != "$agents_version" ]; then
  echo "error: version lockstep check failed: versions disagree" >&2
  report "Makefile VERSION" "$makefile_version"
  report "flake.nix akbVersion" "$flake_version"
  report "AGENTS.md Status" "$agents_version"
  exit 1
fi

echo "version lockstep check OK: v$makefile_version (Makefile, flake.nix, AGENTS.md)"
