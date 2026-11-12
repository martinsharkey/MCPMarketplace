# Innovox Assurance Platform: Assumed Requirements Register

> Exported 12 November 2026 from the live document: https://claude.ai/artifact/AbNgSmwtnVnNxqA6TXCM6x. The live document is the master copy.

Nov 10, 2026 · Martin Sharkey

Thirty assumed requirements, drawn from seven synthetic persona rehearsals and traced to the HLD. Twenty are Must for the build to the 21 June 2027 decision point. None has been validated with real users.

## 1 How to read the register

Each requirement is written as a user need, with one acceptance criterion the product owner can test on Innovox's own repositories.

| Field | Meaning |
| --- | --- |
| ID | R-01 to R-30, stable once baselined; retired IDs are never reused |
| Requirement | As a persona role, I need something, so that an outcome follows |
| Priority | MoSCoW for the build to the decision point: Must (20), Should (6), Could (2, designed for but deferred), Won't (2, after the decision point) |
| Acceptance | The test that shows it is met |
| Source | The synthetic persona rehearsal it came from (10 November 2026) |
| Traces to | HLD decisions (AA-nn), HLD sections and build plan epics (EP-nn) |

**Status.** Every requirement is *assumed*. The register is baselined at the discovery gate on 27 November and changed only by recorded decision. Persona reviews at each sprint may add, change or retire requirements.

## 2 Evidence and reproducibility

| ID | Requirement | Priority | Acceptance | Source | Traces to |
| --- | --- | --- | --- | --- | --- |
| R-01 | As a cyber security assessor, I need every finding to show the tool, its version, the target hash, the time and the vulnerability-data date, so that I can judge whether evidence is genuine and current | Must | Every finding in every report carries all five fields; a finding missing any is rejected before storage | Cyber security assessor | AA-04, AA-05, HLD 5 |
| R-02 | As a cyber security assessor, I need to reproduce an assessment from its manifest alone on a separate instance, so that supplier evidence is not self-marking | Must | A re-run from the signed manifest on a second instance gives the same findings for every deterministic assessor | Cyber security assessor | AA-12, EP-21 |
| R-03 | As a supplier security lead, I need every finding linked to the requirements it bears on, so that I can show traceability to controls | Must | Each finding maps to at least one requirement ID in the profile; a traceability matrix is generated per run | Supplier security lead | AA-01, HLD 6 |
| R-04 | As a supplier delivery lead, I need a current evidence pack to exist at any time, so that evidence requests never trigger an assembly exercise | Must | The pack for the latest run is produced on demand with no manual steps | Supplier delivery lead | HLD 7, EP-17 |
| R-05 | As a platform engineer, I need each evidence pack tied to the commit it assessed, so that I never re-run scans against an old release | Must | Packs for any past assessed commit can be retrieved unchanged | Platform or DevSecOps engineer | AA-05, AA-12 |

## 3 Coverage and honesty

| ID | Requirement | Priority | Acceptance | Source | Traces to |
| --- | --- | --- | --- | --- | --- |
| R-06 | As a supplier security lead, I need every report to show which requirements were not assessed and to separate automated, manual and external controls, so that nobody mistakes coverage for compliance | Must | Every report lists unassessed requirements and shows the three control types separately for each standard | Supplier security lead; cyber security assessor | AA-06, HLD 1.2 |
| R-07 | As a supplier security lead, I need organisational controls such as Def Stan 05-138 tracked as evidence I supply, with an owner and an expiry, so that scanning never claims to test them | Must | Organisational controls appear only as manual or external, never as passed by an automated assessor | Supplier security lead | HLD 6 (governance assessor) |
| R-08 | As a supplier security lead, I need findings de-duplicated across tools and a known false-positive rate, so that the team does not waste time arguing with noise | Should | The same issue from several tools appears once; the false-positive rate is measured against the product owner's review sample | Supplier security lead; platform engineer | AA-04 |
| R-09 | As a cyber security assessor, I need no report to state or imply certification, so that self-assessment is never mistaken for accreditation | Must | Every report carries a self-assessment statement; no report template uses certification language | Cyber security assessor | HLD 11, executive summary |

## 4 Gates, approval and integrity

| ID | Requirement | Priority | Acceptance | Source | Traces to |
| --- | --- | --- | --- | --- | --- |
| R-10 | As a Senior Responsible Owner, I need only a named, MFA-authenticated person to pass a gate or accept a risk, so that accountability is clear | Must | No agent or service account can close a gate; every decision records the approver, time and evidence set hash | SRO; supplier security lead | AA-07, HLD 9 |
| R-11 | As a Senior Responsible Owner, I need a record of who accepted which risk, on what evidence and when, so that my decisions are defensible later | Should | A decision record view lists every gate decision and risk acceptance with links to its evidence | SRO | EP-24 |
| R-12 | As a cyber security assessor, I need an LLM finding never to pass a gate on its own, so that probabilistic judgement cannot decide compliance | Must | Gate policy tests show no combination of LLM findings alone yields a pass | Cyber security assessor | AA-08 |
| R-13 | As a supplier delivery lead, I need failed or rejected gates to raise backlog items automatically, so that fixes enter the normal flow of work | Should | Each failing finding or rejection creates a labelled backlog item linked to its evidence | Supplier delivery lead | HLD 7, EP-18 |
| R-14 | As an MOD delivery team security lead, I need every profile change and every disabled or waived check recorded and shown in every report, so that suppliers cannot quietly tune results | Must | Reports list profile version and all changes, disabled checks and waivers since the baseline | MOD delivery team security lead | AA-13, EP-22 |
| R-15 | As an MOD delivery team security lead, I need the same profile to produce the same shape of evidence for every supplier, so that I can compare them | Should | Two workspaces on one profile produce reports with identical structure | MOD delivery team security lead | AA-01, EP-17 |
| R-16 | As an MOD delivery team security lead, I need a read-only view across several suppliers' gates and coverage, so that I see programme risk in one place | Could | Designed for; built only after a go at the decision point | MOD delivery team security lead; SRO | HLD 5 (designed for, built later) |
| R-17 | As a cyber security assessor, I need to verify a supplier's assessment on my own instance, so that I do not rely on the supplier's word | Could | Designed for; built only after a go at the decision point | Cyber security assessor | AA-12, HLD 5 |

## 5 Pipeline and engineering

| ID | Requirement | Priority | Acceptance | Source | Traces to |
| --- | --- | --- | --- | --- | --- |
| R-18 | As a platform engineer, I need the platform to wrap the scanners I already run, so that I am not asked to replace trusted tools | Must | Every code, supply chain and configuration assessor wraps an established tool named in the assessor catalogue | Platform or DevSecOps engineer | AA-02, HLD 6 |
| R-19 | As a platform engineer, I need fast checks on every pull request and heavy assessments at gates or on a schedule, so that the pipeline stays quick | Must | Each assessor declares a runtime class; the pull request lane stays within a time budget set in alpha | Platform or DevSecOps engineer | AA-14, EP-23 |
| R-20 | As a platform engineer, I need agents running tools to add no attack path into my pipeline, so that the platform does not weaken what it assesses | Must | Assessors run sandboxed, with no network by default, read-only target mounts and scoped tokens; tests prove each limit | Platform or DevSecOps engineer | HLD 9 |
| R-21 | As a supplier delivery lead, I need the team to do nothing beyond approvals and supplying manual-control evidence, so that the platform is not another thing to feed | Should | After set-up, a sprint's runs need no manual step other than approvals and manual-control evidence | Supplier delivery lead | HLD 7, EP-18 |

## 6 Data, sovereignty, reporting and commercial

| ID | Requirement | Priority | Acceptance | Source | Traces to |
| --- | --- | --- | --- | --- | --- |
| R-22 | As an MOD delivery team security lead, I need targets and evidence to stay inside the customer's boundary, so that data never sits where we cannot see or control it | Must | In customer-tenancy and air-gapped modes, no target content, finding or evidence leaves the deployment; egress tests prove it | MOD delivery team security lead; SRO | AA-09, HLD 8 |
| R-23 | As a Senior Responsible Owner, I need the platform's own sovereignty and supply chain assessed and published with each release, so that the tool meets the standard it checks | Must | Each release ships a sovereignty statement, SBOM and Secure by Design self-assessment produced by the platform | SRO | AA-10, EP-13 |
| R-24 | As a commercial officer, I need all evidence exportable in open formats at any time, and returned on exit, so that there is no lock-in | Must | Full export in OSCAL, SARIF and CycloneDX succeeds for any workspace and is readable by a third-party tool | Commercial officer | AA-01, AA-04 |
| R-25 | As a commercial officer, I need a compliant route to buy the platform, so that buyers can adopt it | Won't | After the decision point: G-Cloud and marketplace listings | Commercial officer | Build plan 3.4 |
| R-26 | As a commercial officer, I need model contract wording for requiring suppliers to use the platform, so that a mandate is easy to write | Won't | After the decision point | Commercial officer; MOD delivery team security lead | Build plan 3.4 |
| R-27 | As a cyber security assessor, I need each LLM finding to cite its file and line or document section, with its confidence and calibration status, so that I can check it | Must | Findings without a valid citation are discarded; every LLM finding shows confidence and whether its rubric is calibrated | Cyber security assessor | AA-08, HLD 6.1 |
| R-28 | As a supplier security lead, I need LLM rubrics calibrated against a labelled review set before they count toward a gate, so that their error rate is known | Must | Only rubrics at or above the agreed precision are marked gate-eligible; the calibration report is kept as evidence | Supplier security lead | AA-08, EP-11 |
| R-29 | As an MOD delivery team security lead, I need reports that map to Secure by Design as assessors understand it, so that evidence needs no translation | Must | The Defence report template organises evidence by Secure by Design principle and lifecycle stage | MOD delivery team security lead; cyber security assessor | EP-17 |
| R-30 | As a supplier security lead, I need one evidence set to produce several report formats, so that I stop re-mapping evidence for each customer | Should | One run produces the Defence pack, OSCAL results and a remediation plan without re-running assessors | Supplier security lead | EP-17 |

## 7 Assumptions behind the register

The nine assumptions from the rehearsal summary sit under the requirements. The build can test the technical half of most of them; the half about what other people would value is recorded for the decision point.

| # | Assumption | Requirements | What the build tests | Left for the decision point |
| --- | --- | --- | --- | --- |
| A1 | Re-mapping evidence for each customer is the largest manual cost | R-30, R-03 | Time to produce three report formats from one run, against doing it by hand for TenderHub | Whether others share the cost |
| A2 | Stale evidence causes more work than the first pack | R-04, R-05 | Packs stay current and tied to commits across beta | Whether buyers' repeat requests fall |
| A3 | Security leads hold an effective veto over adoption | R-06, R-08 | None | Who decides adoption |
| A4 | Assessors would trust evidence they can reproduce | R-02, R-17 | Reproduction from the manifest on a second instance | Whether assessors accept it |
| A5 | MOD-side leads would push for a common tool in contracts | R-15, R-16, R-26 | Identical report structure across two workspaces | Whether a mandate would happen |
| A6 | A G-Cloud listing is a precondition for government buyers | R-25 | None | The route to market |
| A7 | Engineers accept only wrappers of their existing scanners | R-18, R-19 | Pull request lane time and noise on four repositories | Whether other teams adopt it |
| A8 | SROs value a defensible decision record most | R-11 | The record is complete for every gate in beta | Whether SROs use it |
| A9 | Showing what was not assessed increases trust | R-06, R-09 | Every report shows it | Whether it raises or lowers trust |

**Sources.** All requirements come from the synthetic rehearsals saved in the Migration And Borders project (`claude/discovery/2026-11-10-discovery-rehearsals.md`) and trace to the [HLD](https://claude.ai/artifact/MxjzNmL7Nn2qKmjKY5qQFu) and [build plan](https://claude.ai/artifact/SaJ3Q6MAbEW3z7gAz8yH8A).
