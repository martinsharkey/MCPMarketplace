# Innovox Assured Marketplace: Design Review

> Exported 12 November 2026 from the live document: https://claude.ai/artifact/U7uGakV3kBvkWQXHsePNmk. The live document is the master copy.

Nov 12, 2026 · Martin Sharkey

## 1 Verdict

The design is sound and unusually disciplined, and the opportunity is larger than a marketplace. Three structural issues need fixing before more code is built on the current base.

1. **The gateway is built to a superseded protocol and protects less than the claim implies.** It speaks the MCP revision before 2026-07-28, and it pins tool descriptions only. Tool results, the main route for prompt injection, pass through unchecked.
2. **The trust chain is home-made where proven designs exist.** A stolen online signing key can undo revocations and cannot itself be revoked. Certificates stop working when a signing key is rotated. There is no post-quantum path.
3. **Defence alignment is named, not yet designed in.** The spec lists the standards it maps to but does not yet shape the certificate, the evidence or the company around them. The largest gap is independence: Innovox builds, consults on and certifies the same services.

**The opportunity.** JSP 936 requires Defence suppliers who use open-source components in AI to "develop compelling arguments" that confidence has not been undermined (para 201). A signed certificate with a structured evidence pack is exactly that argument, packaged. Every Defence AI programme and every prime needs it, which makes this trust infrastructure for Defence agents, with the catalogue as one front end.

**What I reviewed.** The design spec in full, all production Go code (about 1,700 of 2,910 lines; the rest are tests), the Plan 1 summary, and the end-to-end test. I ran the test suite, the race detector and `go vet`: all pass. I confirmed four findings with temporary tests, then removed them. I did not read Plan 1's task-by-task detail or most unit tests.

## 2 What is strong

These choices are right and should survive any redesign.

| Strength | Why it matters |
| --- | --- |
| Separate planes: a central Assurance Service and a Customer Node, with no customer traffic reaching Innovox | The only model a Defence customer can accept, and it makes air-gapped use natural |
| Fail closed everywhere: a failed check, lookup or parse is a denial | Matches NCSC's advice to enforce the boundary outside the model |
| The gateway forwards what was hashed, not what the server sent | Tool lists are rebuilt from the five certified fields, so an unhashed field can never reach the model. Server `instructions` are stripped for the same reason |
| Canonical hashing that never normalises Unicode | A zero-width character or homoglyph in a description changes the hash. This is the subtle case most pinning tools miss |
| Strict call parameters | Duplicate or unexpected keys are refused, so gateway and server cannot disagree about which tool is called |
| Audit before action | The intent to call is logged before the call, and an allow that cannot be logged becomes a denial |
| Replay and freeze protection | Feeds must have a rising sequence number, future-dated feeds are refused, and a stale feed denies everything after the grace period |
| Honest claim wording, enforced in code | The certificate can only emit the approved sentence. Over-claiming is the fastest way to lose a Defence buyer |
| A malicious test server and an end-to-end test | The "rug pull" (a tool that changes after certification) is proven blocked, not assumed |
| Review focus written into the plan | Five named edge cases, each pinned to a test. This is how a senior engineer would brief the work |

## 3 Code review: the Plan 1 trust chain

The code is clean, small and well tested: 2,910 lines of Go on the standard library only, with every package's tests, the race detector and `go vet` passing. The findings below are design consequences rather than careless bugs. "Confirmed" means I reproduced it with a temporary test.

| ID | Severity | Finding | Where | Fix |
| --- | --- | --- | --- | --- |
| C1 | High | **The online key can undo a revocation, and cannot itself be revoked.** A feed is a full snapshot signed by any delegated key. A newer feed that omits a revocation reinstates the service. Delegations have an expiry but no revocation, so a stolen online key stays valid until it expires. Confirmed | `pkg/trust`, `node/syncer` | Root-signed revocation of delegations; revocations that stay revoked unless the root reinstates them; separate keys for signing certificates and signing feeds (section 5, T1) |
| C2 | High | **Certificates die when their signing key is rotated.** A certificate is verified against keys valid today, not at issue. A 12-month certificate signed by a 3-month key is silently dropped after rotation. Confirmed: listings fell to zero | `node/syncer` | Verify against the key's validity at a trusted issue time, or use a long-lived certificate key distinct from a short-lived feed key |
| C3 | High | **The gateway speaks a superseded MCP revision.** It handles `initialize` and `ping`, which 2026-07-28 removed, and denies `server/discover`, which servers must now implement. Confirmed. Rebuilt tool lists omit the now-required `resultType`, `ttlMs` and `cacheScope` | `node/gateway`, `pkg/mcpx` | Bring the SDK swap forward from Plan 5; decide a gateway policy for every method and result type in the current revision |
| C4 | Medium | **The audit record does not say which tool was called.** An allowed call logs the service and `tools/call`, but not the tool name, an arguments digest, the certificate or the outcome. Confirmed | `node/gateway` | Log tool, arguments SHA-256, certificate digest, manifest hash, upstream status and duration |
| C5 | Medium | **The gateway trusts its state file.** Signatures are checked once at sync, then discarded. A listing signed by an untrusted key, placed in `state.json`, was served. Confirmed | `node/syncer`, `node/gateway` | Keep the signed envelopes and re-verify against the pinned root when the gateway loads state |
| C6 | Medium | **The running image digest is a configuration value.** The gateway compares the certificate to a digest typed into `gateway.json`, not to what the cluster is running | `cmd/innovox-gateway` | Read the digest from the runtime, and add admission control so an uncertified image cannot start |
| C7 | Medium | **No transport security yet.** The gateway listens over plain HTTP with no timeouts, uses static bearer tokens and calls services over HTTP. OIDC is in Plan 5; TLS and mutual TLS are in no plan | `cmd/innovox-gateway` | TLS at the gateway, mutual TLS to services, server timeouts |
| C8 | Medium | **The audit chain is not anchored.** Hash chaining catches a partial edit. It does not catch a full rewrite or a truncated tail, and the whole file is re-verified at every start | `node/audit` | Signed checkpoints exported to the customer's SIEM; rotation |
| C9 | Medium | **Tools only.** Prompts, resources and every other method are denied, so servers that rely on them cannot be used. This is safe, but it is an undeclared scope limit | `node/gateway` | State it in the scheme; extend the manifest to prompts and resources later |
| C10 | Low | Every tool call re-reads the state file and makes an extra `tools/list` call upstream | `node/gateway` | Cache within the server's `ttlMs`; subscribe to list changes |
| C11 | Low | Two certificates for the same name and version: whichever file sorts last wins | `node/syncer` | A supersession rule in the feed |
| C12 | Low | An envelope is accepted on the first valid signature, so a rule such as "two signatures" or "classical plus post-quantum" cannot be expressed | `pkg/attest` | Threshold verification |
| C13 | Low | Keys are hex files on disk | `cmd/*` | Hardware-backed root key and a key ceremony before any customer pilot |

**Repository hygiene.** There is no CI, no fuzz test for the canonicaliser, no licence file and an empty README. The repository is public, and the spec in it sets out routes to market and tier design; decide whether that is intended.

## 4 Defence and UK standards

The spec maps evidence to five sources but leaves three Defence obligations out entirely: the node's own Secure by Design case, Innovox's supplier obligations, and classification. The table shows what each standard asks, what the design does today, and the change.

| Standard | What it asks that matters here | Design today | Change |
| --- | --- | --- | --- |
| **JSP 936 Part 1** (Dependable AI in Defence) | Externally acquired AI must reach the same confidence as internal AI (para 198). Suppliers using open-source components must give "compelling arguments" (201). Provenance assured and documented (202). Through-life support and access (203). Export controls assessed so MOD keeps "Freedom of Action" (204). Models come with assumptions, known weaknesses and an operational design domain (165). Regression testing after updates (172). In-service monitoring (179 to 180) | Named as a mapping target; `claims_mapping_ref` field is empty in the code | Give every certificate a **service card**: intended use, operating limits, known weaknesses, data touched, model and component dependencies, export-control and sovereignty statement. Structure the evidence pack as claims, argument and evidence against these paragraphs |
| **JSP 440 Leaflet 5C** (Secure by Design) | Any capability handling Defence data needs continuous risk management, a named risk owner and evidence through life | Treated as a mapping for certified services | The Customer Node is itself a Defence system. Ship it with its own Secure by Design pack: threat model, risk register, security case, patching and exit plan. It will also need to meet the customer's network joining rules |
| **Def Stan 05-138 Issue 4, DEFCON 658, Defence Cyber Certification** | Supplier-level cyber controls at the Cyber Risk Profile a contract assigns | Spec rightly says the mark must not claim this | Separate the two: the mark never claims it, but Innovox as a supplier will need it, plus Cyber Essentials Plus, to contract at all |
| **ETSI EN 304 223** (V2.1.1, December 2025; 13 principles) | Audit trails, asset inventory, versioning and authentication of AI assets, supply chain security, timely updates, operational monitoring, controlled decommissioning, end-user communication | Strong on versioning, authentication and supply chain (stages S2, S3, S5). Silent on decommissioning and on communicating with users | Add a decommission step (credentials revoked, data removed, listing retired) and a customer notice process for revocations and vulnerabilities |
| **Software Security Code of Practice** (DSIT and NCSC, May 2025) | 14 principles in four themes. NCSC publishes Assurance Principles and Claims (APCs) for it, and points to Cyber Resilience Test Facilities for independent audit | Named as a mapping target | Write the Innovox scheme in the same APC form, so evidence arrives in NCSC's own structure. Apply the Code to Innovox's own software and publish the self-assessment |
| **NCSC interim advice on agentic AI** (August 2026) | Boundary outside the model, deny-by-default network, brokered short-lived credentials, protected logs, a tested emergency stop | Close match: gateway, token exchange, default-deny egress, kill switch | Protect and anchor the logs (C8); test the kill switch on a schedule |
| **NCSC post-quantum timelines** | Discovery and plan by 2028, priority migration by 2031, complete by 2035 | Ed25519 only | Hybrid signatures now, while there is one node to migrate (section 5, T3) |
| **Classification** (Government Security Classifications) | Data must not flow to a lower tier; imports to higher tiers go through an accredited route | One mention: a "classification label" as a policy input | Each service declares the highest classification it may handle; the gateway blocks flows from a higher-rated tool to a lower-rated one; treat the air-gap bundle as a cross-domain import under NCSC's safe import pattern |

**OWASP MCP Top 10 coverage.** The design covers five, partly covers three and is silent on two.

| Risk | Coverage | Note |
| --- | --- | --- |
| MCP01 Token mismanagement | Designed, not built | Token exchange, never forwarding the caller's token |
| MCP02 Scope creep | Partial | Grants today; OPA later. Certified scopes are not yet in the certificate |
| MCP03 Tool poisoning | Partial | Descriptions pinned; poisoned outputs are not covered |
| MCP04 Supply chain | Strong | Provenance, rebuild, SBOM, scanning |
| MCP05 Command injection | Designed | Static analysis and behavioural stage |
| MCP06 Prompt injection via context | Weak | Output inspection is one line in the spec |
| MCP07 Authentication and authorisation | Designed, not built | OIDC and OPA in Plan 5 |
| MCP08 Audit and telemetry | Built, with gaps | C4 and C8 |
| MCP09 Shadow MCP servers | Not addressed | An opportunity (section 7) |
| MCP10 Context over-sharing | Not addressed | Needs classification flow rules and per-task isolation |

## 5 Security design findings

Nine design-level findings, beyond the code. The first three concern the trust chain; the rest concern what the gateway and pipeline can and cannot promise.

### T1 The feed re-invents part of The Update Framework

The signed feed solves replay and freeze attacks well, but not key compromise. The Update Framework (TUF) is the established design for exactly this problem, and decision D1 already allows standard signing libraries. Adopt its model:

- **Separate roles:** root, certificate signing, snapshot and timestamp, each with its own key.
- **Thresholds:** root actions need more than one key holder.
- **Root rotation as a signed chain,** so a node can move to a new root without reinstalling.
- **Revocable delegations,** so a stolen online key is cut off by the root in the next feed.

### T2 Nobody can see what Innovox has certified

A customer sees only its own feed. Nothing stops a certificate being issued quietly, or backdated. A public, append-only transparency log of every certificate and revocation fixes this, and inclusion proofs can travel inside the air-gap bundle. It is also the strongest answer to the conflict-of-interest risk the spec lists: anyone can audit what was certified, when, and by whom.

### T3 Cryptography has no post-quantum path

Everything rests on Ed25519, which a large quantum computer would break. A root key pinned at install is exactly the long-lived asset NCSC's migration timeline targets. DSSE envelopes already carry several signatures, so add a second, post-quantum signature (ML-DSA, FIPS 204) to roots, delegations, feeds and certificates, and require both to verify. Put the root in hardware with a recorded key ceremony.

### T4 Pinning protects descriptions, not behaviour or outputs

The manifest hash proves the tool list is unchanged. It says nothing about what a tool returns, and returned content is how most injection reaches a model. The spec covers this in one line: "inspect outputs for injection markers and DLP". This is the hardest control in the product and needs its own design:

- Enforce the certified `outputSchema` on every result; refuse free text where structure was certified.
- Limit result size and content types; strip or neutralise links.
- Track taint: after an agent reads untrusted content, policy can require human approval before a high-impact tool is called.
- Say plainly in the scheme that output inspection reduces risk and does not eliminate it.

### T5 The new protocol revision adds surfaces the design has not decided on

MCP 2026-07-28 changes what a gateway must police. Each of these needs an explicit allow or deny rule in the gateway and a check in the pipeline:

- **Input requests.** A tool result can now be `input_required`, asking the user for information or sending them to a URL. A malicious server can use this to phish.
- **Server identity in `_meta`** on every result, replacing the handshake.
- **Looser schemas.** Input and output schemas may use any JSON Schema keyword, including `$ref`. A reference to outside content is not covered by the hash and should fail certification.
- **Custom headers** from tool parameters (`x-mcp-header`), **extensions** and the **tasks** extension.

The revision also helps: tool lists no longer vary per connection, which makes pinning more robust.

### T6 Certified behaviour is observed once, not enforced

Stage S7 watches system calls, file access and network egress during certification and stores the result as evidence, but the node does not enforce it. A server can behave during the test and misbehave later. Turn the observation into an enforced envelope: put the permitted egress hosts, file paths and system-call profile in the certificate, and have the node generate the network policy and sandbox profile from it. A certified service then cannot do what it was not certified to do. This closes most of the evasion gap and is a real differentiator.

### T7 Pipeline thresholds and judgement calls are undefined

- Stage S8 passes "within tier thresholds" that the spec does not set.
- AI red-team results are probabilistic and need calibration against human review before they count.
- "No critical or high without accepted VEX" does not say who may accept.
- Rebuild from source assumes reproducible builds; say what happens when a build is not reproducible.

### T8 Fail-closed is also an availability risk

After the 7-day grace period every service is denied. For a deployed or disconnected unit that is an operational outage caused by a missed file transfer. Make the grace period a policy the customer's risk owner sets per node, add an audited break-glass, and warn well before expiry.

### T9 Agent Cards identify an agent; they do not assure it

For A2A the design pins the Agent Card. A2A 1.0 now supports signed cards, which prove a card is intact but not that the agent is trustworthy, so an Innovox countersignature in the native format has value. But an agent's behaviour depends on its model, prompts and tools. An agent certificate needs to pin those too, and carry evaluation evidence across its intended use, which is what JSP 936 asks for.

## 6 Assurance scheme findings

The scheme, not the software, is what a Defence buyer is trusting. Six points need settling.

| # | Finding | Why it matters | Recommendation |
| --- | --- | --- | --- |
| S1 | **Innovox builds, advises on and certifies the same services.** The Factory generates MCP servers, then the pipeline certifies them | The claim says "independently tested", which is not true of something Innovox built. ISO/IEC 17020, the standard NCSC's test facilities are accredited to, grades inspection bodies by independence; a body that designs what it inspects cannot hold the most independent grade | Two marks or two wordings: "Innovox Assured" for third-party services, "Innovox Built" (first-party declaration) for Factory output. Or have a partner test facility assess Factory output. Keep assessors organisationally separate from consulting |
| S2 | **The accreditation route points at the wrong standard.** The spec plans ISO/IEC 17065 product certification later | NCSC's current route for product assurance is Principles Based Assurance, delivered by Cyber Resilience Test Facilities accredited to ISO/IEC 17020 | Write the scheme as an APC set now. Then choose: partner with a test facility for Tier 2 and 3, or later become one (training in the method, a trial evaluation, UKAS accreditation and an NCSC contract) |
| S3 | **The name may collide with NCSC's.** NCSC brands its own schemes as "Assured" | A buyer could read "Innovox Assured" as NCSC-backed, which the claim wording works hard to avoid | Brand and legal check before launch; consider "Innovox Verified" or "Innovox Tested" |
| S4 | **Tier 1 is automatic and lasts 12 months** | A year is long for unattended certification of third-party code. Only known CVEs trigger recertification | Short-lived Tier 1 certificates, 30 to 90 days, renewed automatically while checks still pass. This also makes revocation and the grace period far less critical |
| S5 | **Liability and scheme rules are not written.** No certificate terms, disclosure process or customer notice commitment | A mark creates a duty of care. If a certified service causes harm, the first question is what the mark promised | Scheme rules, certificate terms, a vulnerability disclosure policy, a revocation notice commitment, and professional indemnity cover |
| S6 | **The assessor is named inside a signed, published certificate** | A personal name in an immutable public record is a data protection and personal security issue, more so for cleared staff | Record an assessor identifier; keep the mapping to a person internal |

**Recertification triggers** are well chosen. Add three: a change of upstream maintainer or ownership, a licence change, and a change to any model the service depends on.

## 7 The larger opportunity

The durable asset is the scheme, the trust chain and the gateway. A catalogue of MCP servers is one use of them. Seven ways to make it larger, each with a first step.

| Opportunity | Why it is bigger | First step |
| --- | --- | --- |
| **Sell the "compelling argument" to suppliers, not only MOD** | JSP 936 para 201 puts the burden on every supplier using open-source AI components. Primes and SMEs bidding into Defence need the evidence as much as the end customer does | A certify-as-a-service offer: a supplier submits a component and receives the certificate and service card for its own assurance case |
| **Enforced behaviour envelope** (T6) | Docker's catalogue, the nearest comparator, describes commit pinning, AI-assisted review and signed images. Its published approach does not describe enforcing certified behaviour at run time, mapping to Defence policy or an air-gap feed | Put egress, file and system-call limits in the certificate; generate policy from it on the node |
| **Shadow MCP discovery** | OWASP lists unapproved MCP servers as a top-ten risk and the design does not touch it. A node that shows a customer every MCP endpoint in use, and which are uncertified, is useful on day one, before anything is certified | A discovery report from gateway and egress logs |
| **Beyond MCP servers** | The same pipeline can assure agent skills, prompts, model files and datasets. One scheme, several profiles | Define the certificate subject generically now, so new subject types are data |
| **Own the standard** | There is no published NCSC assurance principle set for agent tools. A scheme written in NCSC's form, with an open certificate format anyone can verify, can become the reference | Publish the certificate predicate and the APC draft; share them with NCSC and Dstl |
| **Allied reuse** | JSP 936 aims at interoperability with NATO and allies, and ETSI EN 304 223 is a European standard | Keep sector mappings as data so a NATO or Five Eyes profile is an added mapping, not a fork |
| **One engine with the Assurance Platform** | The pipeline here is one assurance profile of the platform already designed; the signed certificate and feed are what that platform's evidence store needs | A shared Go trust core used by both products |

**What not to chase yet:** billing, multi-node fleets and a public marketplace front end. None of them is where the value or the risk lies.

## 8 Recommended changes, in priority order

Six changes come first because later plans build on the parts they alter. The rest fit into Plans 2 to 5 or follow the pilot.

| Priority | Change | Addresses |
| --- | --- | --- |
| **Now**, before Plan 2 | 1. Re-baseline the gateway on MCP 2026-07-28, keeping the earlier revision for older servers, with an explicit rule for every method and result type | C3, T5 |
|  | 2. Rebuild the feed trust on the TUF model: separate roles, revocable delegations, root rotation, revocations that persist | C1, T1 |
|  | 3. Make certificates survive key rotation | C2 |
|  | 4. Complete the audit record and add signed checkpoints | C4, C8 |
|  | 5. Have the gateway re-verify signed envelopes; add TLS, mutual TLS and a runtime-read image digest | C5, C6, C7 |
|  | 6. Decide the independence model, the wording for Factory-built services, and the name | S1, S3 |
| **Next**, within Plans 2 to 5 | 7. Service card and claims-argument-evidence pack mapped to JSP 936 paragraphs, in NCSC's APC form | Section 4, S2 |
|  | 8. Enforced behaviour envelope from stage S7 | T6 |
|  | 9. A full design for tool-result handling and input requests | T4, T5 |
|  | 10. Hybrid post-quantum signatures, hardware root key, key ceremony | T3, C12, C13 |
|  | 11. Short-lived Tier 1 certificates; defined thresholds for stage S8 | S4, T7 |
|  | 12. Classification rules in the certificate and gateway; customer-set grace period with break-glass | Section 4, T8 |
|  | 13. CI, fuzzing of the canonicaliser, a licence, and signed releases with an SBOM for Innovox's own code | Section 3 |
| **Before a customer pilot** | 14. Secure by Design pack for the Customer Node; Cyber Essentials Plus and the Def Stan 05-138 position for Innovox | Section 4 |
|  | 15. Scheme rules, certificate terms, disclosure policy, indemnity cover | S5, S6 |
| **Later** | 16. Transparency log | T2 |
|  | 17. Test facility partnership and an APC proposal to NCSC | S2, section 7 |
|  | 18. Shadow MCP discovery; further subject types; shared core with the Assurance Platform | Section 7 |

**Effect on the plan.** Changes 1 to 5 rework parts of Plan 1 that already pass their tests, so they are cheapest now, while the code is 2,900 lines. I would insert them as a "Plan 1b" before starting the certification pipeline.

## 9 Sources

**Reviewed:** the [MCPMarketplace repository](https://github.com/martinsharkey/MCPMarketplace) at commit `9f7caf5`: design spec (11 November 2026), Plan 1 summary, and all production code.

**Opened for this review, 12 November 2026:**

- [MCP specification 2026-07-28: key changes](https://modelcontextprotocol.io/specification/2026-07-28/changelog)
- [JSP 936 Part 1: Dependable AI in Defence (MOD)](https://assets.publishing.service.gov.uk/media/6735fc89f6920bfb5abc7b62/JSP936_Part1.pdf)
- [Implementing dependable AI in Defence: key directives from JSP 936 (Burges Salmon)](https://www.burges-salmon.com/articles/102jpv4/implementing-dependable-ai-in-defence-key-directives-from-jsp-936)
- [ETSI EN 304 223 summary (arc42)](https://quality.arc42.org/standards/etsi-en-304-223) and [news report (Help Net Security)](https://www.helpnetsecurity.com/2026/01/19/etsi-european-standard-ai-security/); the standard itself was not opened
- [NCSC: Software Security Code of Practice, Assurance Principles and Claims](https://www.ncsc.gov.uk/guidance/software-security-code-of-practice-assurance-principles-claims)
- [NCSC: Cyber Resilience Test Facilities, assurance principles and claims](https://www.ncsc.gov.uk/schemes/cyber-resilience-test-facilities/assurance-principles-claims) and [information for facilities](https://www.ncsc.gov.uk/schemes/cyber-resilience-test-facilities/information-for-crtfs)
- [OWASP MCP Top 10](https://owasp.org/www-project-mcp-top-10/) (2025 list, beta phase)
- [Docker: Securing the Docker MCP Catalog](https://www.docker.com/blog/enhancing-mcp-trust-with-the-docker-mcp-catalog/)

**Opened earlier in this work:**

- [NCSC: Timelines for migration to post-quantum cryptography](https://ncsc.gov.uk/guidance/pqc-migration-timelines)
- [Summary of NCSC "Managing the cyber risk of agentic AI"](https://aicybr.com/blog/ncsc-agentic-ai-security-sandbox-credentials-observability); NCSC's own page was not opened
- [MOD: What is Secure by Design?](https://www.digital.mod.uk/secure-by-design/what-is-sbd)
- [Def Stan 05-138 (GOV.UK)](https://www.gov.uk/government/publications/cyber-security-for-defence-suppliers-def-stan-05-138)

**Not re-checked.** The Update Framework, ISO/IEC 17020 independence grades, FIPS 204 and NCSC's safe import pattern are described from general knowledge. A2A signed Agent Cards and NCSC's "Assured" scheme names come from search results I did not open. Confirm these before relying on them.
