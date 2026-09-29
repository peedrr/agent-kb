#!/usr/bin/env bash
# Release precondition guard: refuse to release unless the working tree is
# clean and the expected release branch is checked out. Runs before the
# version-lockstep check and before any tag is created, so a failed check
# leaves the repo untouched.
set -euo pipefail

# Change this single value when the release branch moves.
expected_branch="main"

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

# Test gate: run the suite the way CI runs it — no system or global git
# configuration and no identity environment — so machine-dependent scenarios
# fail here, before the tag exists, instead of after it is published.
# (v0.22.0 shipped a red suite this way: the NixOS host's system-level git
# config supplied an identity the CI runners lacked, and testscripts that
# shell out to raw `git commit` passed locally and failed in CI.)
echo "running test suite (CI-simulated git environment)..."
(
  cd "$root_dir"
  env -u GIT_AUTHOR_NAME -u GIT_AUTHOR_EMAIL \
      -u GIT_COMMITTER_NAME -u GIT_COMMITTER_EMAIL \
      -u AKB_AUTHOR_NAME -u AKB_AUTHOR_EMAIL \
      GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null \
      CGO_ENABLED=0 go test -count=1 ./...
# -count=1: the go test cache does not key on GIT_CONFIG_* (only the git
# subprocess reads them), so a cached green from the host's own environment
# would otherwise satisfy this gate without running anything.
)

echo "release precondition check OK: clean tree on branch '$expected_branch', test suite green"
