# Innovox Assured MCP/A2A Marketplace

Go module `innovox.uk/assured`. Owned by Innovox Limited.

## Read these first

1. **Design review:** [docs/programme/marketplace/design-review.md](docs/programme/marketplace/design-review.md).
   A review of this repository's design and Plan 1 code against Defence and UK standards.
   Section 8 lists 18 recommended changes in priority order; the first six come before Plan 2.
   None is implemented yet, and the owner has not yet accepted them.
2. **Document index:** [docs/programme/INDEX.md](docs/programme/INDEX.md).
   Where every design, plan, requirement and discovery document is, how the three strands
   of work fit together, the working rules, and the decisions still open.
3. **Design spec:** [docs/superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md](docs/superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md).
4. **Plan 1 (built):** [docs/superpowers/plans/2026-11-12-plan-1-trust-chain.md](docs/superpowers/plans/2026-11-12-plan-1-trust-chain.md).

## Code layout

| Path | Contents |
| --- | --- |
| `pkg/` | Shared libraries: `manifest`, `attest`, `trust`, `feed`, `mcpx` |
| `assurance/` | Central Assurance Service: `signer`, `publisher` |
| `node/` | Customer Node: `syncer`, `gateway`, `audit` |
| `cmd/` | Command-line entry points |
| `test/e2e/`, `test/fixtures/evilmcp/` | End-to-end test and the malicious test server |

## Checks

```sh
go test ./...
go test -race ./...
go vet ./...
```

All three passed at commit `9f7caf5`. There is no CI yet.

## Working rules

- Do not start a review finding without the owner's go-ahead on that change.
- Write a failing test that reproduces a finding before fixing it.
- Do not name design documents README.
- The remaining rules are in section 5 of the index.
