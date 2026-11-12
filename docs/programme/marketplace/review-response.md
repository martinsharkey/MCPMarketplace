# Design review: response and status

Tracks every finding in [design-review.md](design-review.md). Update this file whenever a finding's status changes.

- **Review baseline:** commit `9f7caf5` (Plan 1 only). The review could not see later work.
- **Owner approval:** on 12 November 2026 the owner approved recommended changes 1 to 5 ("Plan 1b"). Change 6 and changes 7 to 18 are not yet approved.
- **Status values:** Fixed, Partly fixed, Open, Owner decision.

## Code findings

| ID | Finding | Status | Where it stands |
|---|---|---|---|
| C1 | Online key can undo a revocation and cannot be revoked | Partly fixed | The root can revoke a signer key, and nodes keep both key and certificate revocations for good (IAS-TRUST-03, IAS-TRUST-06). Still open: one key signs both certificates and feeds, and there are no thresholds or root rotation (IAS-TRUST-09). |
| C2 | Certificates die when their signing key is rotated | Fixed | A certificate is checked against the key's delegation at the time it was issued (IAS-TRUST-07). |
| C3 | Gateway speaks a superseded MCP revision | Open | Approved. Planned as Plan 1b part A. Control IAS-RT-14. |
| C4 | Audit record does not say which tool was called | Fixed | Records carry the tool, an arguments digest, certificate digest, manifest hash, outcome and duration (IAS-AUD-05). |
| C5 | Gateway trusts its state file | Fixed | Listings are rebuilt from the signed feed against the pinned root on every load (IAS-TRUST-08). Swapping the whole file for an older valid one is not detected. |
| C6 | Running image digest is a configuration value | Open | Approved. Already recorded as partial under IAS-ISO-02. Needs the runtime controller. |
| C7 | No transport security | Partly fixed | Server timeouts added. TLS, mutual TLS and OIDC are open. Planned as Plan 1b part C. Control IAS-RT-15. |
| C8 | Audit chain is not anchored | Partly fixed | Signed checkpoints and an export sink are built (IAS-AUD-03). Log rotation and avoiding a full re-verify at every start are open. |
| C9 | Tools only: prompts, resources and other methods are denied | Open | Not yet approved. The scheme should state the scope limit. |
| C10 | Every call re-reads state and re-lists tools upstream | Open | Not yet approved. |
| C11 | Two certificates for one name and version: last file wins | Open | Not yet approved. |
| C12 | First valid signature wins: no threshold or hybrid rule | Open | Not yet approved. |
| C13 | Keys are hex files on disk | Open | Not yet approved. `cmd/innovox-root` states it is for development and pilots only. |
| Hygiene | No CI, no fuzz test, no licence, empty README; repository is public | Partly fixed | README written. CI, fuzzing and a licence are open. Whether the repository stays public is the owner's decision. |

## Security design findings

| ID | Finding | Status | Where it stands |
|---|---|---|---|
| T1 | Feed re-invents part of The Update Framework | Open | Approved (change 2). Planned as Plan 1b part B. Control IAS-TRUST-09. |
| T2 | No transparency log | Open | Not yet approved. |
| T3 | No post-quantum path | Open | Not yet approved. |
| T4 | Pinning protects descriptions, not outputs | Partly fixed | Pattern-based output inspection and size limits are built (IAS-RT-09). Output-schema enforcement, link handling and taint tracking are open. |
| T5 | New protocol revision adds undecided surfaces | Open | Approved with C3 (change 1). |
| T6 | Certified behaviour is observed once, not enforced | Open | Not yet approved. |
| T7 | Pipeline thresholds are undefined | Open | Not yet approved. |
| T8 | Fail-closed is also an availability risk | Partly fixed | The grace period is set per node in the gateway configuration. Break-glass and early warning are open. |
| T9 | Agent Cards identify an agent but do not assure it | Open | Not yet approved. |

## Assurance scheme findings

| ID | Finding | Status | Where it stands |
|---|---|---|---|
| S1 | Innovox builds, advises on and certifies the same services | Owner decision | Change 6. The claim text says "independently tested", which is not true of Factory output. |
| S2 | Accreditation route points at the wrong standard | Owner decision | The review recommends NCSC Principles Based Assurance through ISO/IEC 17020 test facilities, not ISO/IEC 17065. The review marks the 17020 detail as not re-checked. |
| S3 | The name "Assured" may collide with NCSC's schemes | Owner decision | Change 6. |
| S4 | Tier 1 is automatic and lasts 12 months | Open | Not yet approved. |
| S5 | Liability and scheme rules are not written | Open | Not yet approved. |
| S6 | Assessor is named in a signed, published certificate | Open | Not yet approved. |

## Work that postdates the review

Built after commit `9f7caf5`, so the review does not mention it:

- The scheme as code: 59 controls mapped to verified clauses (`scheme/`).
- A conformance harness and generated self-assessment (`cmd/innovox-conformance`, `docs/conformance/`).
- An adversarial corpus of malicious MCP behaviours (`test/adversarial`).
- Kill switch, rate limits, output inspection, signer-key revocation, audit export and checkpoints.
- A fix for a redirect-following flaw in the MCP client that allowed server-side request forgery.
