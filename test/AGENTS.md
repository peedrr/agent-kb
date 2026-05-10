# test/

**Parent:** `./AGENTS.md`

## OVERVIEW

Integration tests using the testscript framework. 41 `.txt` script files define end-to-end CLI scenarios.

## FILES

| File | Purpose |
|------|---------|
| `integration_test.go` | TestMain builds akb binary; Test runs testscript with isolated HOME per test |
| `testdata/*.txt` | 41 testscript scenarios covering full workflows, edge cases, and command permutations |

## TESTSCRIPT CONVENTIONS

- **Input files**: `-- filename.md --` ... `-- filename.md --` blocks
- **Commands**: `exec akb <cmd>`, `cd <dir>`, `stdout 'pattern'`, `stderr 'pattern'`, `exists <path>`
- **Negation**: `!` prefix for negative assertions (`! stdout`, `! exists`)
- **Binary**: Built once in TestMain with `CGO_ENABLED=0`; reused across all tests
- **Isolation**: Each test gets unique `HOME` directory to prevent registry conflicts

## TEST CATEGORIES

| Category | Files | Coverage |
|----------|-------|----------|
| Core workflow | `full_workflow.txt`, `round_trip.txt`, `core_io_verification.txt` | init→write→read→append→delete |
| Search/links | `search_tests.txt`, `link_graph_round_trip.txt`, `link_resolution.txt`, `wikilink_round_trip.txt` | FTS5, wikilinks, backlinks |
| Lint | `lint_tests.txt` | All 9 checkers |
| Raw | `raw_workflow.txt`, `raw_status_drift.txt`, `raw_sync.txt`, `raw_delete.txt` | SHA-256 manifest, drift detection |
| Git | `git_commits.txt`, `git_edge_cases.txt`, `no_commit.txt` | Auto-commit, merge conflicts |
| Index/log | `index_add_remove.txt`, `index_rebuild.txt`, `log_show_append.txt` | Index management, append-only log |
| Registry | `registry_tests.txt`, `multi_kb.txt` | Multi-KB registry, `akb use` |
| Approval | `approve_tests.txt`, `isdraft_tests.txt` | Draft approval, provenance stripping |
| CEL validation | `cel_validation.txt`, `cel_lint.txt`, `cel_old_format.txt` | Write-time CEL rules, lint-time CEL rules, old-format rejection |
| Validation | `validation.txt`, `path_guards.txt`, `write_errors.txt` | Input validation, error paths |

## RUNNING TESTS

```bash
go test ./...                   # All tests
go test ./test/ -test.v         # Integration tests only
```

## NOTES

- Tests run from `test/` directory; project root is one level up
- TestMain builds `akb` binary to a temp directory and adds it to PATH
- Each `.txt` file is an independent scenario; no shared state between files
- Use `! exec akb <cmd>` to assert command failure
