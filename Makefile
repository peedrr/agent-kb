.PHONY: build test lint clean release

VERSION := 0.19.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Untracked files count as dirty here too, so the marker means the same thing
# as the clean-tree guard in scripts/check-release-preconditions.sh.
DIRTY := $(shell test -z "$$(git status --porcelain 2>/dev/null)" || echo "-dirty")
LDFLAGS := -X main.version=$(VERSION)-$(COMMIT)$(DIRTY)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/akb ./cmd/akb/

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

# Release: refuse unless the repo is releasable (clean tree on the expected
# branch), verify the version string agrees across Makefile, flake.nix and
# AGENTS.md, then tag the current commit. Every guard runs first, so a failed
# check aborts before any tag is created.
release:
	scripts/check-release-preconditions.sh
	scripts/check-version-lockstep.sh
	git tag -a v$(VERSION) -m "akb v$(VERSION)"
