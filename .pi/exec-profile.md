# Execution Profile — agent-kb (Go)

Consumed mechanically by the `/execute-plan` coordinator (pi-meta-config). Keep the
four `##` section headers EXACT — the coordinator HALTs if any is missing. Every
GATES command must be exit-code meaningful (non-zero = gate failure).

## GATES
build: make build
vet: go vet ./...
lint: make lint
fmt: test -z "$(gofmt -l .)"
test: go test ./...
test-integration: go test ./test/ -test.v

## SYMBOL-TOOLS
No symbol-navigation MCP is configured for this project — use `rg`/`grep`
(`--type go`) for symbol lookup, definition, and reference checks.

## TERMINOLOGY
package-unit: package (module: github.com/peedrr/agent-kb)
linter: golangci-lint (+ go vet)
dormant-test: "build-tagged / t.Skip()ped"

## CONVENTIONS
- Build output ALWAYS `-o bin/akb` (never project root); `make build` handles this.
- YAML: `github.com/goccy/go-yaml` ONLY — never `gopkg.in/yaml.v3`.
- Integration tests use the testscript framework (`.txt` files in testdata/).
- Managed files: never write `index.md` or `log.md` directly (use `akb index add` /
  `akb log`).
- DB: no concurrent DB operations (max 1 open connection).
- CLI exit codes: 0=success, 1=validation failure, 2=internal error.
- Reuse existing module deps before adding; any dep beyond the plan's enumeration
  needs supervisor approval (canonical module, minimal, approval in commit body).
- Code comments describe non-obvious behaviour only.
