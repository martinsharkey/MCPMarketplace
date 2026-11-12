# Innovox Assured: MCP and A2A Marketplace for UK Defence

Innovox Assured lets UK Defence teams adopt AI-agent tools (MCP servers and A2A agents) with evidence their risk owner can use. Run-time enforcement ensures that only what was tested actually runs.

- **Assurance service (Innovox):** tests each service against the Innovox Assured scheme, issues a signed certificate that binds the exact image and tool definitions to tiered evidence mapped to MOD and NCSC guidance, and publishes a signed feed.
- **Marketplace node (each customer, in their own environment, connected or air-gapped):** syncs the feed, lists certified services, creates or imports new ones, and enforces at the gateway that nothing uncertified, changed or revoked reaches an agent.

> Independently tested by Innovox against the Innovox Assured scheme, with signed evidence mapped to MOD and NCSC guidance. This is not MOD accreditation: the customer's SRO remains the risk owner.

## Documentation

| Document | What it covers |
|---|---|
| [Solution architecture](docs/architecture/solution-architecture.md) | Goals, business, context, logical, data, interface, security and physical design, flows, NFRs, roadmap, ADRs |
| [Design spec](docs/superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md) | Agreed design decisions for the pilot |
| [Plan 1: trust chain](docs/superpowers/plans/2026-11-12-plan-1-trust-chain.md) | Implementation plan for the first build increment |

## Repository layout

```
assurance/   Innovox-side services (signer, publisher; intake, orchestrator to come)
node/        Customer-node services (sync, gateway, audit; registry, runtime, factory to come)
pkg/         Shared libraries (manifest hashing, attestations, trust, feed, MCP client)
cmd/         Command-line entry points
test/        End-to-end tests and malicious fixtures
docs/        Architecture, specs and plans
```

## Build and test

```bash
go test ./...
go vet ./...
```

Requires Go 1.24 or later and no external modules.
