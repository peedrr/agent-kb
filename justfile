# Task runner for agent-kb. Install via `nix develop` (just is in the devShell).
#
# The version is derived from the VERSION file at the repo root — the same
# file flake.nix reads — so the two build paths cannot drift. Human-facing
# copies (AGENTS.md, CHANGELOG.md, the git tag) are guarded by
# scripts/check-version-lockstep.sh, which `just release` runs before tagging.

version := `cat VERSION`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
# Untracked files count as dirty here too, so the marker means the same thing
# as the clean-tree guard in scripts/check-release-preconditions.sh.
dirty := `test -z "$(git status --porcelain 2>/dev/null)" || echo "-dirty"`
ldflags := "-X main.version=" + version + "-" + commit + dirty

# List available recipes
default:
    @just --list

# Build with the version injected (NEVER output to the project root)
build:
    go build -ldflags "{{ldflags}}" -o bin/akb ./cmd/akb/

# Unit tests
test:
    go test ./...

# Integration tests (testscript)
test-integration:
    go test ./test/ -test.v

# Lint
lint:
    golangci-lint run ./...

# Remove build output
clean:
    rm -rf bin/

# Every guard runs first, so a failed check aborts before any tag is created.
# Release: clean-tree + branch guard, version lockstep check, then tag HEAD as v<VERSION>
release:
    scripts/check-release-preconditions.sh
    scripts/check-version-lockstep.sh
    git tag -a v{{version}} -m "akb v{{version}}"
    @echo "tag v{{version}} created. Publish with: git push --follow-tags"
    @echo "(the tag push triggers .github/workflows/release.yml: test, build, GitHub Release)"
