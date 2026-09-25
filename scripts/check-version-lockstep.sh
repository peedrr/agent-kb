#!/usr/bin/env bash
# Release guard: the version lives in VERSION at the repo root, and both build
# paths (justfile, flake.nix) derive from it, so they cannot drift. What CAN
# drift are the human-facing copies — the AGENTS.md Status line, the CHANGELOG
# section — and the git tag. This script checks those against VERSION.
#
# Tag semantics: before a release the tag must not exist; after `just release`
# it points at HEAD. A tag for the current VERSION that points anywhere else
# means VERSION was bumped without retagging, or the tag was moved — both are
# errors. An absent tag is fine mid-cycle (checked again at release time by
# `git tag -a`, which refuses to overwrite).
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version="$(tr -d '[:space:]' < "$root_dir/VERSION")"

if [ -z "$version" ]; then
  echo "error: version lockstep check failed: VERSION is empty" >&2
  exit 1
fi

fail=0

status_line="$(sed -n 's/^\*\*Status:\*\*[[:space:]]*//p' "$root_dir/AGENTS.md" | head -n 1 || true)"
case "$status_line" in
  "v$version "*) ;;
  *)
    echo "error: version lockstep check failed: AGENTS.md Status line does not start with v$version" >&2
    printf '  %-18s %s\n' "AGENTS.md Status" "${status_line:-<missing>}" >&2
    fail=1
    ;;
esac

if ! grep -q "^## \[$version\]" "$root_dir/CHANGELOG.md"; then
  echo "error: version lockstep check failed: CHANGELOG.md has no [$version] section" >&2
  fail=1
fi

tag="v$version"
if git -C "$root_dir" rev-parse -q --verify "refs/tags/$tag" >/dev/null 2>&1; then
  tag_commit="$(git -C "$root_dir" rev-parse "$tag^{commit}")"
  head_commit="$(git -C "$root_dir" rev-parse HEAD)"
  if [ "$tag_commit" != "$head_commit" ]; then
    echo "error: version lockstep check failed: tag $tag does not point at HEAD" >&2
    printf '  %-18s %s\n' "tag $tag" "$tag_commit" >&2
    printf '  %-18s %s\n' "HEAD" "$head_commit" >&2
    fail=1
  fi
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "version lockstep check OK: v$version (VERSION, AGENTS.md, CHANGELOG.md, git tag)"
