# Innovox Assurance Platform: Build Plan

> Exported 12 November 2026 from the live document: https://claude.ai/artifact/SaJ3Q6MAbEW3z7gAz8yH8A. The live document is the master copy; 1 diagram(s) appear only there.

Nov 10, 2026 · Martin Sharkey

Innovox builds the platform to prove it works on its own repositories, in four phases from discovery on 16 November 2026 to a decision point on 21 June 2027. Requirements are assumed from synthetic persona rehearsals, and every phase ends at a gate the platform itself evidences.

## 1 Approach

The plan follows the delivery and assurance model already agreed: GOV.UK Service Manual phases for delivery, Defence Secure by Design for risk, NIST SSDF for how each increment is built, and RMF names for the evidence. It builds on the [Assurance Platform HLD](https://claude.ai/artifact/MxjzNmL7Nn2qKmjKY5qQFu) and the DP-AI-001 LangSmith platform.

| Framework | What it sets in this plan |
| --- | --- |
| GOV.UK Service Manual and Service Standard | Phases (discovery, alpha, beta, live), synthetic persona review every sprint in place of user research, a Service Standard self-assessment at each gate |
| Agile delivery | Two-week sprints, a review and retrospective at each sprint end, a demonstrable increment every sprint, gates held when exit criteria are met rather than on a calendar date |
| Defence Secure by Design | Threat model, risk register and SbD self-assessment started in discovery and refreshed every release; the product owner accepts residual risk at each gate |
| NIST SSDF and SP 800-218A | The Definition of Done and the pipeline gates on every pull request; AI-specific practices for the orchestrator and LLM review assessor |
| NIST RMF evidence names | System Security Plan, assessment report and remediation plan produced in beta |

**The platform assures itself.** From alpha sprint 1, every pull request to the platform runs through the first assessors the platform provides. As assessors are added, the platform's own gate gets stricter. By the decision point, the platform's own evidence pack has been produced by the platform for months.

**Thin slices, not layers.** Each phase delivers one end-to-end slice (request to assessor to evidence to gate to report) and then widens it, rather than building each layer in turn. The first slice is deliberately small: one profile, three assessors, one report.

**Assumed requirements, self-validated.** There is no external user research. Requirements come from the seven synthetic persona rehearsals and are held in an assumed requirements register, each traced to the transcript it came from. The same personas review every sprint's increment and every gate. Proof is the platform working on Innovox's own repositories against measurable criteria: the platform itself, TenderHub, the migration early-warning product and agent-bridge. The Service Manual phases, Secure by Design and SSDF still apply in full, because proving the method is the point.

## 2 Roadmap

Alpha is the phase to watch: it proves the thin slice, that an assessment can be reproduced from its manifest, and that the platform can pass its own gate.

> *Diagram not included in this export (roadmap · 4 phases, 3 gates, not to scale). See the live document.*

Preparation for alpha runs 30 November to 18 December, and alpha starts after the holiday break. Each gate is held when its criteria in section 3 are met; a later gate moves the following dates rather than shortening the next phase.

## 3 Phase plans

Each phase lists what it must learn or build, the assurance outputs it produces, and the exit gate. Dates are proposals that assume the team in section 6.

### 3.1 Discovery: 16 to 27 November 2026 (1 sprint)

The seven persona rehearsals ran on 10 November and are the requirement source.

Three skills support this: `discovery-persona-builder` builds and refreshes the personas from public role sources, `discovery-persona-interview` rehearses each persona against new questions or screens, and `discovery-synthesis` turns transcripts into the register. Persona output is labelled synthetic throughout.

**Objectives**

- Baseline the assumed requirements register from the rehearsals, each requirement traced to its transcript.
- Fix the first profile: Defence Secure by Design for suppliers.
- Confirm every tool in the first slice is licensed for the intended use.
- Choose the proof repositories: the platform itself, TenderHub, migration early-warning and agent-bridge.

**Outputs**

- Assumed requirements register and problem statement
- Source and tool register with licences
- Security categorisation, initial threat context and risk register v0
- Backlog loaded on the board with the alpha gate epic

**Exit gate (27 November)**

- Register baselined by the product owner.
- First-slice tools licensed.
- Proof repositories chosen.

### 3.2 Alpha: 4 January to 12 March 2027 (5 sprints)

Preparation runs 30 November to 18 December: environment on DP-AI-001 in the connected tier, repository, CI pipeline skeleton and Keycloak realm.

**Objectives**

- Prove the riskiest assumptions with the thinnest end-to-end slice:
  - an assessor contract agents can use safely through MCP and skills
  - evidence that is hashed and reproducible from a signed run manifest
  - a gate engine whose verdict the product owner agrees with
  - LLM review findings calibrated against the product owner's own labelled review set

**Outputs**

- HLD v1 baselined and ADRs AA-01 to AA-14 accepted or revised
- Low-level design for the assessor contract and run manifest
- Working slice: first profile (subset), three assessors, evidence store, run manifest, gate engine, one report
- Threat model v1, risk register v1, SbD self-assessment v1, control catalogue v1
- Pipeline with SSDF gates, running the platform's own assessors on itself
- Calibration report for the first LLM review rubrics
- Alpha report, Service Standard self-assessment and persona review

**Exit gate (12 March)**

- The slice assesses TenderHub end to end and the product owner can act on the report.
- An assessment is reproduced from its manifest alone on a separate instance, with the same result.
- At least one LLM rubric meets the agreed precision against the labelled review set.
- CI gates run on the platform itself; top risks have owners and treatments.

### 3.3 Beta: 15 March to 18 June 2027 (7 sprints)

**Objectives**

- Widen to the full first profile and all ten assessor categories.
- Add the sovereignty and post-quantum assessors, sector report templates and the workspace bridge.
- Add tamper-evident profiles, fast and slow assessor lanes, and the decision record view.
- Run continuously on all four proof repositories through agent-bridge workspaces.
- Prove customer-tenancy packaging on one cloud, and the air-gapped bundle process on an isolated test network.

**Outputs**

- Low-level designs per component
- System Security Plan, self-assessment report and remediation plan, produced by the platform
- Operational runbooks and incident response plan
- Persona review at each release; beta report

**Exit gate (18 June)**

- All four repositories assessed on every change through the last three sprints of beta.
- On a sampled set, the platform's findings agree with the product owner's own review at the agreed rate.
- Reproduction from the manifest holds across releases.
- Customer-tenancy install proven on one cloud.

### 3.4 Decision point: 21 June 2027

The product owner decides whether the platform works well enough to take further, on the beta evidence: the platform's own evidence pack, agreement with manual review, run times and reproducibility. If the answer is yes, the next steps are planned then: real user research, an independent IT health check, a buyer pilot and a G-Cloud listing. Assessor verification mode and the multi-supplier programme view are designed for in the HLD and built only after a yes.

## 4 Backlog

Twenty-four epics, each tagged with the phase milestone, component and the standards it evidences, using the label families from the delivery model. Each phase also has one gate epic whose items are its exit criteria, so gate readiness reads straight off the board.

| ID | Epic | Phase | Component | Main labels |
| --- | --- | --- | --- | --- |
| EP-01 | Assumed requirements register from persona rehearsals | discovery | product | `gate:discovery` |
| EP-02 | Tool and standards licence register | discovery | catalogue | `sbd:p2` `layer:organisation` |
| EP-03 | Security categorisation, threat context, DPIA screening | discovery | platform | `rmf:prepare` `sbd:p3` |
| EP-04 | Platform foundation on DP-AI-001: environment, Keycloak realm, CI | alpha | platform | `ssdf:po.3` `ssdf:po.5` |
| EP-05 | Assessor contract and secure MCP library | alpha | assessors | `ssdf:pw.1` `sbd:p6` `sbd:p7` |
| EP-06 | First three assessors: static analysis, components and licences, secrets | alpha | assessors | `ssdf:pw.7` `ssdf:pw.4` `ssdf:ps.1` |
| EP-07 | Standards catalogue and first profile in OSCAL | alpha | catalogue | `rmf:select` |
| EP-08 | Evidence store and findings normalisation | alpha | evidence | `ssdf:ps.3` `sbd:p9` |
| EP-09 | Gate engine and human approval | alpha | gates | `sbd:p1` `rmf:authorize` |
| EP-10 | Orchestrator squad on Agent Server | alpha | orchestrator | `ssdf:pw.1` `agent:lead` |
| EP-11 | LLM review assessor and calibration | alpha | assessors | `ssdf:pw.8` SP 800-218A |
| EP-12 | Portal: new assessment, run and gate screens | alpha | portal | Service Standard, `asvs:` |
| EP-13 | Platform self-assurance pipeline | alpha | platform | `ssdf:pw.7` `sbd:p9` `sbd:p10` |
| EP-14 | Remaining assessor categories: design, configuration, running service, accessibility, governance | beta | assessors | `sbd:p7` `sbd:p8` |
| EP-15 | Sovereignty assessor | beta | assessors | `sbd:p2` `layer:contract` |
| EP-16 | Post-quantum assessor and CBOM | beta | assessors | NCSC PQC timelines |
| EP-17 | Sector report templates and OSCAL outputs | beta | reports | `rmf:assess` |
| EP-18 | Workspace bridge integration with agent-bridge | beta | bridge | `agent:` routing |
| EP-19 | Customer-tenancy and air-gapped packaging | beta | hosting | `sbd:p6` AA-09 AA-11 |
| EP-20 | Beta assurance: System Security Plan, self-assessment report, runbooks | beta | platform | `rmf:assess` `sbd:p5` |
| EP-21 | Signed run manifest and reproduction | alpha | evidence | ssdf:ps.3 sbd:p9 AA-12 |
| EP-22 | Tamper-evident profiles | beta | catalogue | sbd:p8 AA-13 |
| EP-23 | Fast and slow assessor lanes | beta | orchestrator | AA-14 |
| EP-24 | Decision record view | beta | portal | sbd:p1 rmf:authorize |

Work after the decision point is added to the board only if the decision is to go further.

## 5 Alpha sprint plan

Five two-week sprints. Each ends with a demonstration reviewed by the synthetic personas, and the platform's own gate gets stricter as its assessors arrive.

| Sprint | Dates | Goal | Demonstration | Epics |
| --- | --- | --- | --- | --- |
| A1 | 4 to 15 Jan 2027 | One assessor runs through MCP with its skill and writes hashed evidence | An agent runs the static analysis assessor on a sample repository; evidence verifies | EP-05, EP-06, EP-08, EP-13 |
| A2 | 18 to 29 Jan | First profile in OSCAL; three assessors; findings normalised | Same repository assessed against a profile subset; coverage shown | EP-06, EP-07, EP-08 |
| A3 | 1 to 12 Feb | Gate engine and human approval close the loop; signed run manifest | A run reaches "needs judgement"; a named approver decides; failures become backlog items | EP-09, EP-10, EP-21 |
| A4 | 15 to 26 Feb | New assessment and gate screens; first report | A user starts an assessment in the portal and receives the evidence pack | EP-12, EP-17 (first template) |
| A5 | 1 to 12 Mar | LLM review rubrics calibrated; platform passes its own gate | Calibration report; an assessment reproduced from its manifest on a separate instance; the platform passes its own alpha gate | EP-11, EP-13, alpha gate epic |

**Every sprint:**

- synthetic persona review of the increment, findings added to the assumed requirements register
- threat model updated for any design change
- Definition of Done enforced by the pipeline
- sprint review, retrospective and backlog refinement

## 6 Team and roles

One accountable person, working with an agent squad that does most of the drafting, coding, testing and evidence collection, and synthetic personas that stand in for stakeholders. Agents never approve a gate or accept a risk.

| Role | Held by | Accountable for |
| --- | --- | --- |
| Product owner, design authority and approver | Martin Sharkey | Vision, backlog order, ADRs, every gate decision and risk acceptance |
| Security lead | Martin Sharkey, supported by the security reviewer agent | Threat model, risk register, SbD self-assessment, control catalogue |
| Engineering | Martin Sharkey with the agent squad | Assessor contract, orchestrator, pipeline; signs off every agent-written change |
| Synthetic stakeholders | The seven persona skills | Source of assumed requirements; review of each sprint and gate; never evidence of real user need |
| Agent squad | Agents on DP-AI-001 and in the IDE workspace via agent-bridge | Drafting designs and ADRs, writing code and tests, running assessors, assembling evidence, raising backlog items |
| Independent tester | Only after the decision point, if the decision is to go further | IT health check of the platform |

**Agent permissions** follow the delivery model: each agent has the item types and labels it may work (`agent:` routing), its tools are limited to the MCP servers its role allows, and every action is traced in LangSmith. Human sign-off is required on every merge and every gate.

## 7 Risks, dependencies and open decisions

| Type | Item | Response |
| --- | --- | --- |
| Risk | Assumed requirements differ from what real users would need | Held as an assumption register; real research is a decision-point option |
| Risk | The approver is also the builder, so gates lack independent challenge | Reviewer agents on a different model class challenge each gate; reproduction from the manifest on a separate instance |
| Risk | LLM rubrics do not reach the precision needed to count toward gates | Keep them advisory; the gate still works on deterministic assessors alone |
| Risk | Scope widens before the first profile is proven | Alpha gate requires the thin slice only |
| Risk | A wrapped tool's licence does not allow the intended use | Licence check in discovery (EP-02); substitute before alpha |
| Dependency | DP-AI-001 platform available in the connected tier by 30 November | Fallback: a minimal Agent Server and LangSmith install for alpha only |
| Dependency | Isolated test network for the air-gapped process | Local isolated environment in beta |
| Open | Labelled review set for LLM calibration | Product owner labels findings from the proof repositories in alpha |
| Open | Hosting for alpha and beta | UK region; decided in discovery |
