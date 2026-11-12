# Delivery and assurance model

Version 0.2.0 (draft), 10 November 2026.

This document defines how the migration early-warning programme is run, governed,
secured and evidenced from discovery to retirement. It is the source for:

- the backlog structure on the backlog.md board
- the workspace the agent bridge deploys into the target IDE
- the order in which the product documents are written

It is written for this programme but is not product-specific, so it can be reused for
other Innovox products.

## 1. Frameworks and what each one answers

| Question | Framework | How it is used |
|---|---|---|
| How do we deliver? | **GOV.UK Service Manual** (agile phases) and the **Service Standard** | Discovery, alpha, beta, live and retirement phases, iterative delivery, user research. The Service Standard is used as a self-assessment, because a supplier product is not formally assessed |
| How do we manage cyber risk? | **Defence Secure by Design (SbD)** and the cross-government **Secure by Design principles** | Continuous, proportionate risk management across the Defence lifecycle stages (CADMID/T), with the risk artefacts a Defence buyer expects |
| How do we build secure software? | **NIST SP 800-218 Secure Software Development Framework (SSDF)** | Development practices in four groups: Prepare the Organization (PO), Protect the Software (PS), Produce Well-Secured Software (PW), Respond to Vulnerabilities (RV). Applied through the Definition of Done and CI gates |
| How do we build secure AI features? | **NIST SP 800-218A** | Additional practices for the generative AI components |
| How do we package evidence for authorisation? | **NIST SP 800-37 Risk Management Framework (RMF)** document names, mapped to SbD | System Security Plan, assessment report and remediation plan, in the form a buyer's accreditor recognises |
| How do we verify application security? | **OWASP ASVS 5.0.0** and **NCSC principles** | Requirement baseline and test criteria |
| What must Innovox meet as a Defence supplier? | **JSP 440 Leaflet 5C**, **Def Stan 05-138 Issue 4** (via DEFCON 658), **Defence Cyber Certification**, Industry Security Notices | Organisation-level cyber controls at the Cyber Risk Profile each contract assigns (section 5) |
| What must Innovox meet for other government buyers? | Cross-government Secure by Design principles, government security standards, **Cyber Essentials Plus** | Baseline for non-Defence buyers such as the Home Office |

The frameworks do not conflict:

- **Service Manual** sets the delivery rhythm.
- **SbD** sets the risk governance that runs through every phase.
- **SSDF** sets how every increment is built.
- **RMF** names the evidence documents.

## 2. Lifecycle

| Service Manual phase | Defence stage | RMF step | Purpose |
|---|---|---|---|
| **Discovery** | Concept | Prepare | Understand the problem, users, buyers, data and constraints. Decide whether to proceed |
| **Alpha** | Assessment | Categorize, Select | Test the riskiest assumptions with prototypes. Choose the architecture and the control baseline |
| **Beta** (private, then buyer pilot) | Demonstration, Manufacture | Implement, Assess, Authorize | Build the product in increments. Prove it with invited buyer users. Obtain risk acceptance |
| **Live** | In-Service | Monitor | Operate, improve, and continuously assure |
| **Retirement** | Disposal or Termination | (lifecycle activity) | Return or destroy data, revoke access, close records |

### 2.1 Discovery

**Objectives**

- Define the problem and who has it.
- Identify buyers and their decisions.
- Establish the data landscape and licences.

**Outputs**

- Problem statement
- User needs and personas
- Buyer hypotheses
- Source register v0
- Security categorisation and DPIA screening
- Initial risk and threat context
- Discovery report with a recommendation

**Exit gate**

- The problem has been validated in conversations with prospective buyers.
- Viable public data exists, with a commercial-use basis.
- The decision to proceed is recorded.

### 2.2 Alpha

**Objectives**

- Prove the core claim with the smallest prototype: a backtest on one corridor that
  beats a baseline.
- Choose the architecture.
- Establish the security baseline and the toolchain.

**Outputs**

- High-Level Design (HLD) v1 and the first architecture decision records (ADRs)
- Interface control documents (ICDs) for phase-1 sources
- Threat model v1, risk register v1 and SbD self-assessment v1
- Control baseline: SbD principles, SSDF tasks and ASVS level
- CI pipeline with SSDF gates
- Prototype and backtest report
- Alpha report

**Exit gate**

- The backtest beats its baseline out of time.
- The HLD is baselined.
- Top risks have owners and treatments.
- CI gates are running.
- The decision to enter beta is recorded.

### 2.3 Beta

**Objectives**

- Deliver the product in increments, using the delivery phases in the HLD.
- Run a buyer pilot with invited users.
- Obtain the buyer's risk acceptance.

**Outputs**

- Low-Level Designs (LLDs) per component
- System Security Plan
- Test and security evidence per release
- Independent IT health check (penetration test) and remediation
- Security assessment report and remediation plan
- Operational runbooks and incident response plan
- Accessibility statement
- Pilot evaluation

**Exit gate**

- The IT health check findings are remediated or accepted.
- The buyer risk owner has accepted the residual risk.
- Operational readiness is confirmed.
- The decision to go live is recorded.

### 2.4 Live

- Continuous assurance:
  - SbD self-assessment and risk register refreshed every release
  - Scheduled DAST and an annual IT health check
  - Vulnerability management under SSDF RV
- Change control, accuracy scorecard monitoring and feedback-driven iteration.

### 2.5 Retirement

- Buyer data is returned and its destruction verified.
- Credentials and keys are revoked.
- The archive is disposed of as agreed.
- Records are retained for the agreed period.

## 3. Governance and roles

| Role | Accountability | Held by |
|---|---|---|
| Product owner | Vision, priorities, backlog ordering, phase gate decisions | Martin Sharkey |
| Design authority | Architecture, HLD and ADR approval | Martin Sharkey, until delegated |
| Security lead | Threat model, risk register, SbD self-assessment, SSDF compliance | Named before alpha |
| Delivery lead | Cadence, flow and phase gate readiness | Product owner until delegated |
| Buyer risk owner | Acceptance of residual risk for their deployment | Each buyer |
| Agents (agentic squad) | Drafting, coding, testing, evidence collection and review assistance | Agents act under a named human; they never approve gates or accept risk |

Every phase gate decision, risk acceptance and ADR is recorded with the person who made
it and the date.

## 4. Ways of working

- **Cadence:** two-week sprints. Review and retrospective at sprint end. Phase gates
  are held when exit criteria are met, not on a calendar.
- **Branching:** trunk-based, with short-lived branches and pull requests. The default
  branch is protected.
- **Increments:** every sprint ends with something demonstrable and its evidence
  updated.
- **User research:** in discovery and alpha, at least one buyer or user conversation
  per sprint. The findings are recorded.

### 4.1 Definition of Ready

A work item is ready when:

1. It traces to a user need, requirement, risk or control.
2. Its acceptance criteria are testable.
3. Its security considerations are noted, or recorded as "no change to the threat
   model".
4. Its dependencies are identified.
5. It is small enough for one sprint.

### 4.2 Definition of Done

A work item is done when:

1. Its acceptance criteria pass, with automated tests where practical.
2. Its code has been reviewed by a human or an approved reviewer agent, with human
   sign-off.
3. **The pipeline is green (section 6):** all security and compliance tests pass, and
   every control touched by the change has a passing test or a valid waiver. Any
   waiver records an owner, a reason and an expiry.
4. The threat model, risk register and affected documents are updated if the design
   changed.
5. Accessibility checks pass for UI changes.
6. Evidence is attached or linked to the item.
7. The change is deployed to the test environment.

## 5. Compliance layers

Standards apply at three layers. Each layer has its own evidence and owner.

| Layer | Applies to | Standards | How compliance is shown |
|---|---|---|---|
| **Product** | The software and its hosted service | Secure by Design (JSP 440 Leaflet 5C for Defence; cross-government principles otherwise), NIST SSDF and SP 800-218A, OWASP ASVS 5.0.0, NCSC principles, WCAG 2.2 AA | Pipeline tests and gates (section 6), threat model, risk register, SbD self-assessment, System Security Plan, IT health check |
| **Organisation** | Innovox as a supplier | Def Stan 05-138 Issue 4 at the assigned Cyber Risk Profile (level 0 to 3), Defence Cyber Certification, Cyber Essentials Plus, ISN obligations (including ISN 2023/09 and ISN 2026/01) | Certification, the Supplier Assurance Questionnaire, a Cyber Implementation Plan where needed, and organisational policies and records |
| **Contract** | Each buyer engagement | Security Aspects Letter, data classification, DEFCON 658 where it applies, buyer-specific conditions | Per-contract control set, added to the catalogue with a `contract:` tag |

**Rules**

- Def Stan 05-138 does not replace system-level security. Product controls come from
  Secure by Design and the contract's Security Aspects Letter.
- Defence requirements are treated as the default superset, so the product is ready for
  MOD buyers without rework. Non-Defence buyers receive the same evidence.
- Only official control identifiers are used: Def Stan 05-138 controls from the Issue 4
  control set, ASVS requirement numbers, SSDF task IDs and SbD principle numbers.
  Mappings are reviewed by the security lead.

### 5.1 Control catalogue

One machine-readable catalogue, `security/controls.yaml`, is the single source of truth
for what we comply with and how it is verified.

```yaml
- id: CTL-AUTH-003
  title: MFA enforced for privileged roles
  layer: product
  maps_to:
    sbd: [p4, p8]
    ssdf: [PW.1]
    asvs: []              # ASVS 5.0.0 requirement IDs, added once verified
    defstan_05_138: []    # product control; organisation equivalents mapped separately
  applies_when: [all]
  verification:
    method: automated     # automated | manual | external
    tests: [policy/test_keycloak_realm.py::test_mfa_required_for_privileged_roles]
  evidence: pipeline
  owner: security-lead
  status: implemented     # planned | implemented | verified | waived
```

- **Automated controls** name the tests that verify them.
- **Manual controls** name the evidence document and review frequency.
- **External controls** (such as certification or the IT health check) name the
  certificate or report and its expiry.
- **Waivers** carry an owner, reason, compensating control and expiry. An expired waiver
  fails the pipeline.

## 6. Test and assurance pipeline

Every test is tagged with the control IDs it verifies. The pipeline reads the control
catalogue, runs the tests, and fails if any applicable control has no passing test,
valid evidence or valid waiver.

| Stage | Runs on | Checks | Standards evidenced |
|---|---|---|---|
| **1. Pre-commit** | Developer machine and agents | Secret scanning, formatting, linting | SSDF PS.1, PW.5 |
| **2. Build and unit** | Every pull request | Unit tests; SAST (Bandit for Python, ESLint security rules for TypeScript); SCA (`pip-audit`, `npm audit`, OWASP Dependency-Check); dependency licence check; SBOM (CycloneDX) | SSDF PW.4, PW.5, PW.6, PW.7, PW.8, PS.3; SbD p2 |
| **3. Configuration and policy** | Every pull request | Container image scan; infrastructure and compose configuration scan; **compliance-as-code tests**: only the proxy publishes ports, no Docker socket mounts, Keycloak realm policy (MFA, no self-registration, session limits, brute-force detection), security headers and CSP, cookie flags, TLS settings, least-privilege database grants, row-level security enabled, logging and audit enabled | SbD p7, p8, p4, p5; ASVS; NCSC cloud principles; product equivalents of Def Stan 05-138 technical controls |
| **4. Data and ethics rules** | Every pull request | Schema tests that no individual-level fields exist; small-cell suppression; licence attribution present for every displayed source; no restricted source enabled without a licence record | Responsible design; licence register; DPIA commitments |
| **5. Integration and security** | Merge to main, deployed to the test environment | Integration tests; authenticated ZAP DAST against the BFF; automated ASVS test suite; accessibility (axe, Playwright); AI evaluation: prompt-injection tests on enrichment and explain, citation-grounding checks | SSDF PW.8; ASVS; WCAG 2.2 AA; SP 800-218A; SbD p9 |
| **6. Release** | Release candidate | Image signing and provenance attestation; evidence bundle (all reports with tool versions, targets and hashes); control coverage report; traceability matrix generated from the board | SSDF PS.2, PS.3, PO.4; SbD p10; RMF Assess |
| **7. Scheduled** | Weekly and on new vulnerability data | Re-scan dependencies and images against current vulnerability data; scheduled DAST; waiver expiry check; certificate and certification expiry check | SSDF RV.1, RV.2; SbD p9 |
| **8. Periodic (outside the pipeline)** | Annually or as contracts require | IT health check; Cyber Essentials Plus and Defence Cyber Certification assessments; SbD self-assessment and risk register review; tabletop incident exercise | Organisation layer; SbD p1, p3, p9; RMF Monitor |

**Gate rules**

- Stages 1 to 4 block merging. Stage 5 blocks promotion beyond test. Stage 6 blocks a
  release.
- A failed, empty or unauthenticated scan is a failure, never a pass. This rule is
  carried over from the EIS security gates.
- Findings at high or critical severity block. Lower findings need triage within a set
  time and are tracked as `defect` items.
- Reports are retained as machine-readable pipeline artifacts with tool version,
  vulnerability-data date, target, scope, artifact hash and result.

**What the pipeline cannot prove:** most Def Stan 05-138 controls are organisational
(governance, people, suppliers and physical security). The pipeline verifies their
technical equivalents in the product and tracks the rest through the control catalogue
as manual or external controls with expiry dates. The coverage report shows the two
separately, so nobody mistakes a green pipeline for certification.

## 7. Document set

| Document | Path | Created | Maintained |
|---|---|---|---|
| Delivery and assurance model (this document) | `delivery-model.md` | Discovery | When the model changes |
| Problem statement | `discovery/problem-statement.md` | Discovery | Through alpha |
| User needs and personas | `discovery/user-needs.md` | Discovery | Every phase |
| Discovery report | `discovery/discovery-report.md` | Discovery end | Baselined |
| Product brief | `product/product-brief.md` | Discovery end | Every phase |
| Source register | `product/source-register.md` | Discovery | Continuously |
| High-Level Design | `design/hld.md` | Alpha | Every release |
| Architecture decision records | `design/decisions/adr-NNN-*.md` | Alpha | Append-only |
| Interface control documents | `design/interfaces/icd-*.md` | Alpha onwards | Per source change |
| Low-Level Designs | `design/lld/*.md` | Beta | Per component release |
| Threat model | `security/threat-model.md` | Discovery (context), alpha (v1) | Every design change |
| Risk register | `security/risk-register.md` | Discovery | Every release |
| SbD self-assessment | `security/sbd-self-assessment.md` | Alpha | Every release |
| System Security Plan | `security/system-security-plan.md` | Beta | Every release |
| SSDF practice evidence | `security/ssdf-evidence.md` and CI artifacts | Alpha | Every release |
| Control catalogue | `security/controls.yaml` | Alpha | Every change to a control |
| Compliance coverage report and traceability matrix | Generated by the pipeline at release | Alpha | Every release |
| Organisation compliance record (Def Stan 05-138, Defence Cyber Certification, Cyber Essentials Plus) | `governance/organisation-compliance.md` | Discovery | On certification and contract changes |
| Security assessment report and remediation plan | `security/assessment/*.md` | Beta | Per assessment |
| DPIA and ethics statement | `governance/dpia.md`, `governance/ethics.md` | Discovery (screening), alpha (full) | Every phase |
| Phase reports | `governance/phase-reports/*.md` | At each gate | Baselined |
| Runbooks and incident response plan | `operations/*.md` | Beta | Continuously |
| README | `README.md` | Discovery | Repository landing page only |

**Document status:** each document carries a status (`draft`, `reviewed` or
`baselined`), a version, an owner and a date. Baselined documents change only through
a recorded decision.

## 8. Backlog model (input to the backlog.md board)

### 6.1 Hierarchy

```text
Programme
└── Phase (milestone): discovery, alpha, beta, live, retirement
    └── Epic: a capability or assurance outcome
        └── Item: story, task, security control, evidence, spike, ICD, ADR or defect
```

### 6.2 Item types

| Type | Used for | Done means |
|---|---|---|
| `story` | User-facing capability, written from a user need | Acceptance criteria pass and are demonstrated |
| `task` | Technical work with no direct user outcome | Merged under the Definition of Done |
| `control` | Implementing a control from the catalogue (any layer) | Control implemented, catalogue updated, and verified by a test or recorded evidence |
| `evidence` | Producing or refreshing an assurance artefact | Artefact updated, reviewed and linked |
| `spike` | Time-boxed research | Findings recorded and a decision taken |
| `icd` | Interface control document for a source | Checklist complete and licence confirmed |
| `adr` | Architecture decision | Decision recorded with its context and consequences |
| `defect` | A fault, including a vulnerability | Fixed, tested and, for vulnerabilities, handled under the RV process |

### 6.3 Labels

| Label family | Values | Purpose |
|---|---|---|
| `phase:` | `discovery`, `alpha`, `beta`, `live`, `retirement` | Phase milestone |
| `component:` | `acquisition`, `model`, `portal`, `platform`, `identity`, `hosting` | Ownership |
| `sbd:` | `p1` to `p10` | Secure by Design principle addressed |
| `ssdf:` | Practice IDs, for example `po.1`, `ps.2`, `pw.4`, `rv.1` | SSDF practice evidenced |
| `ctl:` | Control catalogue IDs, for example `ctl-auth-003` | Control implemented or verified |
| `asvs:` | ASVS requirement numbers | Application security requirement |
| `defstan:` | Def Stan 05-138 control references | Organisation or product equivalent control |
| `layer:` | `product`, `organisation`, `contract` | Compliance layer |
| `rmf:` | `prepare`, `categorize`, `select`, `implement`, `assess`, `authorize`, `monitor` | Evidence placement |
| `gate:` | The gate that needs the item | Phase gate readiness |
| `agent:` | Which agent may work the item | Agent routing through the bridge |

### 6.4 Statuses

| Status | Meaning |
|---|---|
| `To Do` | Not yet started |
| `Ready` | Meets the Definition of Ready |
| `In Progress` | Being worked |
| `In Review` | Awaiting human or reviewer-agent approval |
| `Done` | Meets the Definition of Done |
| `Blocked` | Cannot proceed; the reason is stated on the item |

### 6.5 Traceability

Every item links upward and downward along this chain:

```text
user need → requirement → design element (HLD/LLD/ADR) → control → test → evidence
```

The board can therefore generate:

- A traceability matrix for any phase gate
- The SbD principle and SSDF practice coverage for the evidence pack
- A list of open items blocking the next gate

### 6.6 Phase-gate epics

Each phase has one gate epic. It lists the exit criteria from section 2 as items, so
gate readiness can be read straight off the board.

## 9. Agent bridge deployment

The agent bridge deploys the following into the target IDE workspace:

1. The backlog.md board, initialised with the phases, labels, statuses and item
   templates from section 8, and the discovery and alpha gate epics.
2. The document skeletons from section 7, each with its status header.
3. Agent definitions with their permitted item types and labels (`agent:` routing),
   and the rule that agents never approve gates or accept risk.
4. The control catalogue skeleton and the pipeline stages from section 6, with the gate
   rules enforced.
5. A pointer to this document as the governing model.

## 10. Current position and next steps

**The programme is in discovery.** The existing design draft is ahead of the problem
definition and is treated as alpha input, not a baselined HLD.

1. Agree this model.
2. Build the control catalogue and pipeline gates from sections 5 and 6.
3. Define the backlog.md board and the agent-bridge deployment from sections 8 and 9.
4. Write the problem statement and user needs, evidenced from public sources.
   Material from previous client work is used only where its owner permits it.
5. Complete the DPIA screening.
6. Restructure the current design draft into the product brief and HLD, using the
   document set in section 7.
