# Innovox Assured MCP/A2A Marketplace — Design Spec

Date: 2026-11-11 · Owner: Martin Sharkey (Innovox.uk) · Status: Draft for review

## 1. Purpose and intent

Innovox provides UK Defence customers with a catalogue of MCP servers and A2A agents that are either imported or newly created, each carrying an **Innovox Assured** mark. The mark is Innovox's own independent assurance scheme. Its evidence is mapped to MOD Secure by Design principles, JSP 936, ETSI EN 304 223, the DSIT/NCSC Software Security Code of Practice, and the OWASP MCP and Agentic threat lists.

**Claim wording (mandatory):** "Independently tested by Innovox against the Innovox Assured scheme vX, with signed evidence mapped to MOD and NCSC guidance." Innovox does not claim "SbD certified", "MOD/NCSC approved/accredited" or DEFCON 658 compliance. The customer's SRO remains the risk owner.

**Success for the pilot:** one Defence customer runs a node, consumes certified services, and sees the full lifecycle working: create or import, certify, publish, sync, connect, enforce, and block on change.

**Commercial context (not built in the pilot):** delivered as containers, so the product is cloud-agnostic. Routes to market are hyperscaler marketplaces via private offers, Commercial X, and primes. G-Cloud is not relied on.

## 2. Decisions

| # | Decision |
|---|---|
| D1 | Everything is built by Innovox except protocol SDKs, crypto/signing (Sigstore, TLS, OAuth libraries), OPA, and standard scanners. |
| D2 | Containerised on Kubernetes, and cloud-agnostic. |
| D3 | Separate planes: a central **Assurance Service** run by Innovox, and a **Customer Node** run by each customer. There is no shared tenancy, and customer data and agent traffic never reach Innovox. |
| D4 | Tiered sign-off. Tier 1 is automatic. Tier 2 needs assessor sign-off. Tier 3 needs a cleared assessor plus manual testing. |
| D5 | Go for services and the CLI, TypeScript/React for the portals, Python for generated MCP servers, and PostgreSQL plus an OCI registry for storage. |

## 3. Architecture

### 3.1 Assurance Service (central)
- **intake**: accepts a submission of type `repo | image | remote | openapi`. It validates metadata and checks namespace ownership by DNS/HTTP challenge or GitHub OIDC.
- **orchestrator**: runs the pipeline stages (§4) as isolated jobs and stores the evidence.
- **assessor console**: a review queue, an evidence viewer, sign/reject with reasons, and revoke.
- **signer**: issues certificates as DSSE-wrapped in-toto Statements.
- **publisher**: produces a signed, monotonic feed index, bundle files for air-gap transfer, and a revocation list.

### 3.2 Customer Node (per customer)
- **sync**: pulls the feed or imports a bundle, then verifies it (§5.1).
- **registry**: implements the MCP Registry API v0 (read subset) and an Agent Card registry. It lists only verified items and local entries.
- **gateway**: the single MCP/A2A entry point, enforcing §5.2.
- **runtime**: runs certified images under a gVisor RuntimeClass, with default-deny egress NetworkPolicy.
- **factory**: generates MCP servers from OpenAPI and handles imports (§6).
- **audit**: an append-only, hash-chained log with SIEM export.
- **catalogue portal**: browse, request access, admin approvals and policy.

## 4. Certification pipeline

| Stage | What it does | Pass criterion |
|---|---|---|
| S1 Intake | Validate the schema, check namespace ownership, record the submission | Valid and verified |
| S2 Provenance | Verify signature and SLSA provenance against a policy that pins repo, builder and workflow; rebuild from source with pinned bases | Matches policy; rebuild digest recorded |
| S3 Code and supply chain | Semgrep/CodeQL, secrets scan, Syft SBOM, Grype/OSV, Trivy, licences | No critical/high without accepted VEX |
| S4 MCP static | mcp-scanner plus Innovox rules for poisoning, hidden instructions, shadowing and over-broad capabilities | No high findings |
| S5 Manifest pin | Run in the sandbox, call `tools/list` or fetch the Agent Card, canonicalise with RFC 8785, SHA-256 | Hash recorded |
| S6 Protocol and auth | OAuth 2.1 conformance, audience checks, no token passthrough, resource indicators, schema fuzzing | All required checks pass |
| S7 Behavioural | Scripted calls under gVisor; trace syscalls, file access and egress | No undeclared behaviour |
| S8 AI red-team | garak/promptfoo for injection via outputs, exfiltration, excessive agency | Within tier thresholds |
| S9 Decision | Assemble the evidence pack and sign according to tier | — |

**Tiers:** T1 Assessed = S1–S5 plus automated S6, auto-signed, 12-month expiry. T2 Assured = S1–S8 with assessor sign-off, 12 months. T3 Mission = T2 plus manual pen test and red-team by a cleared assessor, 6 months (out of pilot scope). Remote endpoints are capped at T1 and labelled "Remote – attested", with the manifest hash re-checked daily.

**Recertification triggers:** a new version, a manifest hash change, a new CVE affecting the SBOM, a scheme version bump, or expiry.

## 5. Customer node behaviour

### 5.1 Sync
1. Verify the feed signature against the Innovox root key pinned at install. Reject any feed whose sequence number is not greater than the last accepted one.
2. For each certificate: verify the signature, check it has not expired, and check it is not revoked. Mirror the image by digest into the local OCI registry.
3. An admin accepts items into the catalogue. Local policy may be stricter (for example, minimum tier).
4. On revocation: flag the listing, and the gateway denies it immediately.

### 5.2 Gateway, per connection and call
1. Authenticate the caller's OIDC token (human or workload).
2. Authorise with OPA using the caller, service, tool, arguments, classification label, AccessGrants and PolicyBindings.
3. Verify the certificate: valid, unexpired, unrevoked, and its image digest equals the running digest.
4. Manifest check: on connect and on every `notifications/tools/list_changed`, re-hash the manifest. A mismatch means block plus alert. Tools are filtered to those the caller is allowed.
5. Token exchange (RFC 8693) to a short-lived, audience-scoped downstream token. The caller's token is never forwarded.
6. Validate arguments against the schema. Inspect outputs for injection markers and DLP, with warn or block set by policy.
7. Hold calls to high-impact tools until a human approves.
8. Append every decision to the audit log.
9. Kill switch per service, per agent, or global.

For A2A, steps 1–3 and 8–9 apply, and the Agent Card hash plus the Innovox countersignature are verified before a task is accepted.

**Fail-closed rules:** if the policy engine, revocation data or verification is unavailable or fails, the gateway denies. When offline, the last verified feed stays valid for a configurable grace period (default 7 days). After that, all listings are denied.

## 6. Factory
- **Create:** takes an OpenAPI 3.x spec plus an allow-list of operationIds. It generates a Python MCP server with one tool per allowed operation, carrying the input schema derived from the operation's parameters and body. Backend auth uses a token-exchanged credential. It emits a Dockerfile and build provenance.
- **Import:** takes an MCP Registry entry or repo URL, mirrors the source at a pinned commit, and rebuilds it.
- All factory output is submitted to the pipeline. Nothing is listed without a certificate.

## 7. Data model

**Central records:**
- Submission(id, publisher, source_type, source_ref, requested_tier, status)
- ServiceVersion(name `uk.example/svc`, version, kind mcp|a2a, transport, image_digest | remote_url)
- ToolManifest(service_version, canonical_json, sha256)
- Evidence(submission, stage, tool, result, report_uri, sha256)
- Certificate(id, subject digests, manifest_sha256, tier, scheme_version, issued_at, expires_at, assessor, envelope)
- Revocation(cert_id, reason, revoked_at)
- Bundle(seq, created_at, index, signature)

**Node records:**
- Listing(cert_id, accepted_by, status)
- AccessGrant(principal, service, scopes, expires_at)
- PolicyBinding(service, policy_id)
- AuditEvent(seq, prev_hash, hash, ts, actor, action, decision, detail)

**Certificate predicate** (`predicateType: https://innovox.uk/attestation/certificate/v1`):
- `scheme_version`, `tier`, `service {name, version, kind}`, `manifest_sha256`
- `evidence[] {stage, result, sha256}`, `issued_at`, `expires_at`, `assessor`, `claims_mapping_ref`

The subjects are the OCI image digest. For remote endpoints, the subject is `remote:` followed by the URL, with digest equal to the manifest hash.

## 8. Testing
- Unit tests written test-first for every package.
- Golden tests: canonicalisation and hashing are deterministic and insensitive to key order and tool order.
- A malicious fixture server exhibiting poisoned descriptions, a post-certification tool change, and token passthrough. The pipeline must fail S4/S6 and the gateway must block the rug-pull.
- End to end in kind: submit, certify, publish, sync, connect, mutate tools, blocked, revoke, denied.

## 9. Pilot scope
**In:** T1 and T2; MCP over streamable HTTP; A2A Agent Card verification only; online feed and air-gap bundle; OpenAPI create; repo import; a single customer node; basic portals.
**Out:** T3; full A2A task proxy; billing; marketplace listing packaging; multi-node fleet; UKAS (ISO/IEC 17065) processes.

## 10. Repository layout
```
assurance/{intake,orchestrator,signer,publisher}   Go services
node/{sync,registry,gateway,factory,audit}         Go services
pkg/{manifest,attest,feed,policy,mcpx}             shared Go libraries
factory-templates/python                           generated server template
web/{catalogue,assessor}                           React
scheme/                                            scheme text + control mapping (YAML)
deploy/helm                                        charts
test/fixtures/evil-mcp                             malicious test server
```

## 11. Risks
- Conflict of interest. Mitigations: assessors are separate from any Innovox consulting on the same service, decisions are logged, and an independent accreditation route is planned.
- Upstream spec churn in MCP (2026-07-28: stateless, `server/discover`) and A2A v1.0. Mitigation: isolate protocol handling in `pkg/mcpx`.
- Compromise of the root key. Mitigations: offline root, online intermediate signing key, rotation via feed.
- Over-claiming. Mitigation: claim wording (§1) is enforced in the portal UI and evidence packs.
