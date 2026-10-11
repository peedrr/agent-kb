# Execution Profile — `agent-kb` (Go)

Machine-parsed by the `/execute-plan` coordinator (pi-meta-config). Keep the
four `##` section headers EXACT — the coordinator HALTs if any is missing. Every
GATES command must be exit-code meaningful (non-zero = gate failure).

## GATES
build: just build
vet: go vet ./...
lint: just lint
fmt: test -z "$(gofmt -l .)"
test: go test -count=1 ./...
test-integration: go test -count=1 ./test/ -test.v

## SYMBOL-TOOLS
pi-lens is available and indexed for this project (Go-first):
- Go: use `symbol_search` → `module_report`/`read_symbol`/`read_enclosing` for
  lookup, definition, and reference triage; `lens_diagnostics source=lsp`
  (gopls) for error checks on changed files. Reference lists are ~80%
  name-only resolution — triage-grade, verify before acting on deletions.
- ast-grep (`ast_grep_search`/`ast_grep_replace`, Go/shell/YAML/nix) and
  `lsp_navigation` are situational: activate via `pi_lens_activate_tools`.
- YAML/JSON: full LSP diagnostics (yaml-ls/json-ls) + yamllint. Nix: nixd.
- Shell: live shellcheck diagnostics via bash-ls (nix shellcheck on PATH) +
  shfmt runner.
- Markdown: marksman LSP + typos-lsp spell-check + markdownlint-cli2 all live.
  Vale configured (.vale.ini, Microsoft+alex pruned, akb vocab): `just vale`
  lints tracked Markdown and exits 1 on errors — NOT a gate yet (known
  backlog, cleanup delegated; see .pi/research/vale-style-research.md).
- GitHub Actions: zizmor security aux live (fires on `.github/workflows/`).
- BROKEN on NixOS: opengrep only (pi-lens-managed glibc prebuilt; its
  getToolPath does not fall back to PATH — ignore opengrep gaps in output).
- `lens_diagnostics` outcome "inconclusive" = wait budget elapsed / server is
  silent-on-clean — NOT a failure (that is "unavailable"/"failed").
- No coverage: testscript `.txt`, justfile, go.mod, `.log` — use `rg`.
- Graph stales across commits (tools warn); `pilens_rebuild` refreshes.
- Full tool-availability analysis: `.pi/research/pi-lens-tool-availability.md`.

## TERMINOLOGY
package-unit: package (module: `github.com/peedrr/agent-kb`)
linter: golangci-lint (+ go vet)
dormant-test: "build-tagged / `t.Skip()`ped"

## CONVENTIONS
- Build output ALWAYS `-o bin/akb` (never project root); `just build` handles this.
- YAML: `github.com/goccy/go-yaml` ONLY — never `gopkg.in/yaml.v3` (module
  paths stay lowercase verbatim).
- Integration tests use the testscript framework (`.txt` files in testdata/).
- Managed files: never write `index.md` or `log.md` directly (use `akb index add` /
  `akb log`).
- DB: no concurrent DB operations (max 1 open connection).
- CLI exit codes: 0=success, 1=validation failure, 2=internal error.
- Reuse existing module deps before adding; any dep beyond the plan's enumeration
  needs supervisor approval (canonical module, minimal, approval in commit body).
- Code comments only where the behaviour is not evident from the code itself.
