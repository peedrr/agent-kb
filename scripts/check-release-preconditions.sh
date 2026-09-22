#!/usr/bin/env bash
# Release precondition guard: refuse to release unless the working tree is
# clean and the expected release branch is checked out. Runs before the
# version-lockstep check and before any tag is created, so a failed check
# leaves the repo untouched.
set -euo pipefail

# Change this single value when the release branch moves.
expected_branch="master"

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

status="$(git -C "$root_dir" status --porcelain)"
if [ -n "$status" ]; then
  echo "error: release precondition failed: working tree is not clean" >&2
  printf '%s\n' "$status" >&2
  exit 1
fi

branch="$(git -C "$root_dir" rev-parse --abbrev-ref HEAD)"
if [ "$branch" = "HEAD" ]; then
  echo "error: release precondition failed: HEAD is detached (expected branch '$expected_branch')" >&2
  exit 1
fi

if [ "$branch" != "$expected_branch" ]; then
  echo "error: release precondition failed: on branch '$branch', expected '$expected_branch'" >&2
  exit 1
fi

echo "release precondition check OK: clean tree on branch '$expected_branch'"
