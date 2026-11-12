# Working in this repository

Guidance for any engineer or AI agent. Read this first, then the documents it points to.

## What this is

**Innovox Assured** is an MCP and A2A marketplace for UK Defence, built by Innovox.uk. It has two halves:

- **Assurance service (run by Innovox):** tests MCP servers and A2A agents against the Innovox Assured scheme, issues signed certificates and publishes a signed feed.
- **Marketplace node (run by each customer):** syncs the feed, lists certified services, and enforces at a gateway that only certified, unchanged, unrevoked services reach an agent. It works connected or air-gapped.

The product never claims MOD or NCSC certification. The only permitted claim is the constant `attest.ClaimText`.

## Read in this order

| # | Document | Why |
|---|---|---|
| 1 | `README.md` | One-page overview |
| 2 | `docs/architecture/solution-architecture.md` | The whole target design: context, components (APP-01 to APP-17), data, interfaces, security, physical, roadmap, ADRs |
| 3 | `docs/conformance/self-assessment.md` | What is actually built and proven today, control by control |
| 4 | `scheme/controls.json` and `scheme/sources.json` | The scheme: what "assured" means, mapped to verified clauses of each standard |
| 5 | `docs/superpowers/specs/` and `docs/superpowers/plans/` | Agreed design decisions and the step-by-step build plans |

## Repository map

| Path | Purpose | Plane |
|---|---|---|
| `assurance/` | Innovox-side services. Today: `signer` (issues certificates) and `publisher` (signs the feed). Intake, orchestrator, pipeline runners and assessor console are not built yet. | Assurance |
| `node/` | Customer-node services. Today: `syncer` (verifies and applies feeds), `gateway` (run-time enforcement) and `audit` (hash-chained log). Registry, runtime controller, factory and portal are not built yet. | Node |
| `pkg/` | Shared libraries used by both planes: `manifest` (canonical hashing), `attest` (signed certificates), `trust` (key hierarchy), `feed` (feed and bundle format), `mcpx` (minimal MCP client). | Both |
| `cmd/` | Command-line entry points, one directory per binary. Thin wrappers: logic lives in the packages above. | Both |
| `scheme/` | The Innovox Assured scheme as data (`controls.json`, `sources.json`), our own written assessment (`self-assessment.json`), and the Go loader that validates them. | Assurance |
| `internal/` | Code private to this repo: `conformance` (links tests to controls, builds the report), `testpki` (test key hierarchy), `repocheck` (checks this guide stays accurate). | Tooling |
| `test/` | Cross-package tests: `e2e` (full lifecycle), `adversarial` (malicious MCP behaviours through the gateway), `fixtures/evilmcp` (a deliberately malicious server). | Tooling |
| `docs/` | Architecture, generated conformance report, specs and plans. See `docs/README.md`. | — |

The "plane" matters: code under `assurance/` must never be needed at run time by a node, and code under `node/` must never call Innovox. The signed feed is the only link between them.

## Rules

1. **Fail closed.** Any verification, lookup, policy or audit failure denies. Never add a path that allows on error.
2. **Test first.** Write the failing test, watch it fail, then write the code. No production code without a test.
3. **Tag tests with the controls they prove:** `conformance.Covers(t, "IAS-RT-02")`. A control only counts as met when tagged tests pass.
4. **Never hand-edit `docs/conformance/self-assessment.md`.** Regenerate it (command below) and commit the result whenever tests or `scheme/` change.
5. **Standard library only** for now. Adding a dependency needs an architecture decision recorded in the solution architecture (§15).
6. **Scheme changes need a verified source.** A control's `maps_to` must reference a clause in `sources.json`, and clause text must be checked against the primary document, with the `verified` date updated.
7. **Never commit keys or binaries.** `.gitignore` covers `*.key` and built commands.
8. **Keep this guide accurate.** A new top-level directory must be added to the map above, and every Go package needs a package comment. `internal/repocheck` enforces both.

## Commands

```bash
go test ./...                     # full suite
go vet ./...                      # static checks
go run ./cmd/innovox-conformance -out docs/conformance/self-assessment.md   # measure against the scheme
go test -v -run TestCorpus ./test/adversarial/                              # adversarial corpus, shows known gaps
```

Requires Go 1.24 or later. No network access is needed.

## Status and what to build next

The trust chain and the gateway's enforcement core are built. Most of the service around them is not. `docs/conformance/self-assessment.md` is the source of truth for what is met, partial, not met or not yet testable. The roadmap is in the solution architecture (§14.2): Plan 2 (node service foundation) and Plan 3 (assurance service foundation and certification pipeline) come next.

Known temporary simplifications, each to be replaced: Ed25519 DSSE instead of Sigstore, a grant table instead of OPA, a minimal MCP client instead of the official SDK, files instead of PostgreSQL, static bearer tokens instead of OIDC.
