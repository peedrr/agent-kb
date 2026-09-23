# test/

**Parent:** `./AGENTS.md`

## OVERVIEW

Integration tests using the testscript framework. 54 `.txt` script files define end-to-end CLI scenarios.

## FILES

| File | Purpose |
|------|---------|
| `integration_test.go` | TestMain builds akb binary; Test runs testscript with `HOME` set to the scenario work dir and `AKB_KB=.` |
| `testdata/*.txt` | 54 testscript scenarios covering full workflows, edge cases, and command permutations |

## TESTSCRIPT CONVENTIONS

- **Script first**: testscript reads a scenario as a txtar archive, so the commands are the comment section before the first `-- file --` marker and every script is written script-first/files-last; text after a marker belongs to that file
- **Input files**: `-- filename.md --` ... `-- filename.md --` blocks create files in the scenario work dir
- **Commands**: `exec akb <cmd>`, `cd <dir>`, `stdout 'pattern'`, `stderr 'pattern'`, `exists <path>`
- **Negation**: `!` prefix for negative assertions (`! stdout`, `! exists`)
- **Binary**: Built once in TestMain with `CGO_ENABLED=0`; reused across all tests
- **Isolation**: `HOME` is the scenario work dir itself, so a scenario cannot read the developer's configuration and a discovery walk from inside the scenario is bounded there; the harness also sets `AKB_KB=.`, so scenarios run their commands from the KB root they created, and a scenario with no KB clears the variable
- **Empty neighborhood**: a discovery scenario that needs a directory outside every KB creates its own subdirectory (e.g. `mkdir $WORK/empty`, as `discover_empty.txt` does); the harness provides no `home` subdir

## TEST CATEGORIES

| Category | Files | Coverage |
|----------|-------|----------|
| Core workflow | `full_workflow.txt`, `round_trip.txt`, `core_io_verification.txt` | init→write→read→append→delete |
| Search/links | `search_tests.txt`, `link_graph_round_trip.txt`, `link_resolution.txt`, `wikilink_round_trip.txt` | FTS5, wikilinks, backlinks |
| Lint | `lint_tests.txt` | All 9 checkers |
| Raw | `raw_workflow.txt`, `raw_status_drift.txt`, `raw_sync.txt`, `raw_delete.txt` | SHA-256 manifest, drift detection |
| Git | `git_commits.txt`, `git_edge_cases.txt`, `no_commit.txt` | Auto-commit, merge conflicts |
| Index/log | `index_add_remove.txt`, `index_rebuild.txt`, `log_show_append.txt` | Index management, append-only log |
| Addressing | `registry_tests.txt`, `multi_kb.txt`, `kb_addressing.txt`, `discover.txt`, `discover_empty.txt` | Base selection from inside and outside a KB, the removed `use`/`registry` commands rejected as unknown, two bases kept isolated via `--kb`/`AKB_KB`, and the discovery scan; neither `registry_tests.txt` nor `multi_kb.txt` carries a skip marker |
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
