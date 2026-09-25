#!/usr/bin/env bash
# Release guard: the version lives in VERSION at the repo root, and both build
# paths (justfile, flake.nix) derive from it, so they cannot drift. What CAN
# drift are the human-facing copies — the AGENTS.md Status line, the CHANGELOG
# section — and the git tag. This script checks those against VERSION.
#
# Tag semantics: `just release` tags HEAD, and development then continues past
# the tag — so a tag for the current VERSION pointing at an ANCESTOR of HEAD is
# the normal post-release state, not an error. The anomaly this guards against
# is a tag that is NOT reachable from HEAD: that means the tag was moved, or it
# points at rewritten/divergent history. An absent tag is fine mid-cycle;
# re-tagging at release time is refused by `git tag -a` itself. (In CI the
# checkout fetches no tags, so this check is skipped there.)
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
  if ! git -C "$root_dir" merge-base --is-ancestor "$tag_commit" HEAD; then
    echo "error: version lockstep check failed: tag $tag is not an ancestor of HEAD (moved or stranded tag)" >&2
    printf '  %-18s %s\n' "tag $tag" "$tag_commit" >&2
    printf '  %-18s %s\n' "HEAD" "$(git -C "$root_dir" rev-parse HEAD)" >&2
    fail=1
  fi
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "version lockstep check OK: v$version (VERSION, AGENTS.md, CHANGELOG.md, git tag)"
