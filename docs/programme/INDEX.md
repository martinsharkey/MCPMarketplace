# Innovox programme documents: index

Handover pack for any agent or person picking up this work. Written 12 November 2026.

**Read the design review first:** [marketplace/design-review.md](marketplace/design-review.md).
It reviews the code and design in this repository and sets the next work.

## 1 Start here: the design review

[marketplace/design-review.md](marketplace/design-review.md) reviews the Innovox Assured
MCP/A2A Marketplace in this repository at commit `9f7caf5`: the design spec, the Plan 1
trust chain code and the assurance scheme, against Defence and UK standards.

| Section | What it holds |
| --- | --- |
| 1 Verdict | Three structural issues and the size of the opportunity |
| 2 What is strong | Design choices to keep through any redesign |
| 3 Code review | Findings C1 to C13 on the Plan 1 code, with file locations and fixes |
| 4 Defence and UK standards | JSP 936, JSP 440 Leaflet 5C, Def Stan 05-138, ETSI EN 304 223, NCSC; OWASP MCP Top 10 coverage |
| 5 Security design findings | T1 to T9 |
| 6 Assurance scheme findings | S1 to S6 |
| 7 The larger opportunity | Seven ways to widen the product |
| 8 Recommended changes | 18 changes in priority order |
| 9 Sources | What was opened, and what was not re-checked |

**The six changes the review puts first, before Plan 2 starts:**

1. Re-baseline the gateway on MCP 2026-07-28 (C3, T5).
2. Rebuild the feed trust on the TUF model (C1, T1).
3. Make certificates survive key rotation (C2).
4. Complete the audit record and add signed checkpoints (C4, C8).
5. Have the gateway re-verify signed envelopes; add TLS, mutual TLS and a runtime-read image digest (C5, C6, C7).
6. Decide the independence model, the wording for Factory-built services, and the name (S1, S3).

The review proposes these as a "Plan 1b". Nothing in the review has been implemented yet,
and the owner has not yet accepted or rejected any of its recommendations.

**Findings confirmed by test.** C1, C2, C4 and C5, and the `server/discover` part of C3,
were reproduced with temporary tests that were then removed. Those tests are not in the
repository; write them again as failing tests before fixing each finding.

## 2 How the pieces fit

Three strands of work, all owned by Innovox Limited. Martin Sharkey is product owner,
design authority and approver for all of them.

| Strand | What it is | Status |
| --- | --- | --- |
| **Assured Marketplace** (this repository) | Signed certificates for MCP servers and A2A agents, a customer-side gateway that enforces them, and the assurance scheme behind the mark | Plan 1 (trust chain) built and merged. Design review written; its changes not started |
| **Assurance Platform** | App-agnostic assurance pipeline: standards as data, assessors as MCP tools, human-approved gates, evidence packs | Designed. No code and no repository yet (`innovox-assurance` is proposed) |
| **Migration early-warning** | Public-data dashboard product built on EIS | Designed. Parked as a future reference application |

The review (section 7) proposes that the Marketplace certification pipeline becomes one
profile of the Assurance Platform, with a shared Go trust core used by both.

## 3 Documents in this folder

### Marketplace

| Document | Purpose |
| --- | --- |
| [marketplace/design-review.md](marketplace/design-review.md) | The design review. See section 1 above |
| [../superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md](../superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md) | The design spec that was reviewed (already in this repository) |
| [../superpowers/plans/2026-11-12-plan-1-trust-chain.md](../superpowers/plans/2026-11-12-plan-1-trust-chain.md) | Plan 1, as built (already in this repository) |

### Assurance Platform

| Document | Purpose |
| --- | --- |
| [assurance-platform/high-level-design.md](assurance-platform/high-level-design.md) | Architecture, assessor catalogue, deployment modes, security, decisions AA-01 to AA-14 |
| [assurance-platform/build-plan.md](assurance-platform/build-plan.md) | Phases, gates and dates from discovery (16 November 2026) to the decision point (21 June 2027); epics EP-01 to EP-24 |
| [assurance-platform/assumed-requirements-register.md](assurance-platform/assumed-requirements-register.md) | Requirements R-01 to R-30 and assumptions A1 to A9 |
| [assurance-platform/ide-workspace-design.md](assurance-platform/ide-workspace-design.md) | Agent roles, skills, hooks, risk tiers, the developer reviewer agent and containment for the build workspace |
| [assurance-platform/wireframes/](assurance-platform/wireframes/) | Four portal wireframes (PNG), referenced from the high-level design |

### Delivery model

| Document | Purpose |
| --- | --- |
| [delivery/delivery-and-assurance-model.md](delivery/delivery-and-assurance-model.md) | How work is run and evidenced: GOV.UK Service Manual phases, Defence Secure by Design, NIST SSDF and RMF, backlog labels, Definition of Done. Written for migration early-warning and reused by the Assurance Platform |

### Discovery

| Document | Purpose |
| --- | --- |
| [discovery/rehearsals/2026-11-10-summary.md](discovery/rehearsals/2026-11-10-summary.md) | Patterns and assumptions across the seven rehearsals |
| [discovery/rehearsals/](discovery/rehearsals/) | Seven synthetic persona interview transcripts |
| [discovery/skills/](discovery/skills/) | Copies of the three persona skills: builder, interview, synthesis |

### Migration early-warning

| Document | Purpose |
| --- | --- |
| [migration-early-warning/product-design.md](migration-early-warning/product-design.md) | Product design 0.2.0 |
| [migration-early-warning/source-register.md](migration-early-warning/source-register.md) | Candidate public data sources and licence checks |

## 4 Master copies and things not in this folder

The five Marketplace and Assurance Platform documents are exports. The live documents are
the master copies, and the diagrams in three of them appear only there.

| Document | Live copy |
| --- | --- |
| Marketplace design review | https://claude.ai/artifact/U7uGakV3kBvkWQXHsePNmk |
| Assurance Platform high-level design | https://claude.ai/artifact/MxjzNmL7Nn2qKmjKY5qQFu |
| Build plan | https://claude.ai/artifact/SaJ3Q6MAbEW3z7gAz8yH8A |
| Assumed requirements register | https://claude.ai/artifact/AbNgSmwtnVnNxqA6TXCM6x |
| IDE workspace design | https://claude.ai/artifact/VHWUxWgb6Bniokf3QQybAR |
| DP-AI-001 LangSmith Platform HLD v1.4 (the platform the Assurance Platform runs on; not copied here) | https://claude.ai/artifact/GggvoDFqELnwcqwR6aDRnq |

The live links are private to the owner. An agent that cannot open them should work from
the exports here.

| Material | Where it lives |
| --- | --- |
| Master copies of the delivery model, migration product design and source register | TenderHub repository, branch `design/migration-early-warning`, under `solutions/migration-early-warning/` |
| EIS 0.4.0, the data platform the migration product reuses | TenderHub repository, `eis/` |
| agent-bridge: backlog.md board bundle, owner rules, worker-command rulebook, audit | `agent-bridge` repository |
| Brand tokens, logos and fonts used in the wireframes | `innovox` repository |
| Migration early-warning idea paper (the owner's original) | The owner's "Migration And Borders" project on claude.ai; not copied here |

## 5 Rules for anyone working from these documents

- **Requirements are assumed.** They come from synthetic persona rehearsals. There will be
  no real user interviews before the June 2027 decision point. Never present persona
  output as research evidence.
- **There is no human developer.** The owner approves on evidence, not by reading code.
  The developer reviewer agent in the workspace design stands in for code review, and the
  residual risk is recorded at each gate.
- **Design documents are not named README.** Use a name that says what the document is.
- **The migration early-warning design has nothing to do with tendering.** It reuses EIS
  from the TenderHub repository; do not add tender content to it.
- **Outputs are self-assessment evidence, never certification.** Do not use wording that
  implies NCSC or MOD approval.
- **Check standards claims before relying on them.** Section 9 of the review lists
  sources that were described from general knowledge and not re-checked.

## 6 Open decisions for the owner

| Decision | Where it is discussed |
| --- | --- |
| Whether this repository should be public; the spec in it sets out routes to market and tier design | Review, section 3 |
| Accept, change or reject the 18 recommended changes; whether to run "Plan 1b" before Plan 2 | Review, section 8 |
| Independence model and product name (S1, S3) | Review, section 6 |
| Create the `innovox-assurance` repository and load the backlog | Workspace design, section 14 |
| Add a shared Go trust core, JSP 936 and ETSI EN 304 223 to the Assurance Platform HLD | Review, section 7; not yet in the HLD |
| Update the delivery model's statuses to agent-bridge's five columns | Workspace design, section 14 |
