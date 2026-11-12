# Innovox Assurance Platform: IDE Workspace Design

> Exported 12 November 2026 from the live document: https://claude.ai/artifact/VHWUxWgb6Bniokf3QQybAR. The live document is the master copy; 2 diagram(s) appear only there.

Nov 11, 2026 · Martin Sharkey

## 1 Principles

The workspace is where the platform is built: one repository, a backlog.md board, a squad of agents with defined roles, skills that encode how each kind of work is done, and hooks that enforce the rules without anyone remembering them. agent-bridge deploys it, extending what it already ships: the five-column board, ADLC phase labels, owner rules, the worker-command rulebook and the hash-chained audit.

| Framework | What the workspace does with it |
| --- | --- |
| GOV.UK Service Manual | Phases are board milestones (discovery, alpha, beta, decision point); every phase has a gate epic whose children are its exit criteria |
| Agile | Two-week sprints as milestone lanes; the lead agent drafts sprint plans from ready items and runs the sweep; reviews include a synthetic persona review |
| IBM ADLC (already in agent-bridge) | Every task carries one stage label (`adlc-plan` to `adlc-monitor`), so a task's lifecycle and the product's phase are tracked separately |
| Defence Secure by Design | Threat model, risk register and SbD self-assessment are files in the repo, updated by skills and checked at every gate |
| NIST SSDF and SP 800-218A | The Definition of Done, enforced by hooks and CI, with the platform's own assessors running on its own code |
| Owner rules (agent-bridge) | Binding on every agent: native before custom, evidence order, done means proven, least privilege |

**Five rules for building fast:**

1. **Skills, not prompts.** Repeated work (an ADR, an assessor, a gate check) has a skill, so agents do it the same way every time.
2. **Hooks, not reminders.** If a rule can be checked by a machine, a hook or CI gate checks it.
3. **Fast lane in the IDE, slow lane at gates.** Quick checks run on every edit; heavy assessments run at pull request or gate (AA-14).
4. **The platform assures itself.** Its own assessors run in this workspace from the first commit.
5. **Rigour follows risk.** Spikes move fast; critical code gets two independent reviewers and stronger tests (section 10).

## 2 At a glance

Six layers sit on agent-bridge. Agents do the work, skills say how, tools do it, and hooks and gates check it; the owner decides at three points.

> *Diagram not included in this export (workspace layers · 6 layers on agent-bridge, owner decides). See the live document.*

Each layer is a set of files in the repository, so the whole workspace is versioned, reviewed and reproducible like the product itself.

## 3 Repository layout

One repository, `innovox-assurance`, holds the product, its design, its assurance evidence and the workspace configuration. agent-bridge writes the `.agent-bridge/`, `.claude/` and `backlog/` parts; the rest is the product.

```text
innovox-assurance/
├── CLAUDE.md                     workspace rules: points to owner rules, DoD, this design
├── AGENTS.md                     the same rules for non-Claude seats (Zoo, Bob, Codex)
├── .agent-bridge/                bridge policy, seats, mailbox, hash-chained audit
├── .claude/
│   ├── agents/                   subagent definitions (section 5)
│   ├── skills/                   workspace skills (section 6)
│   ├── commands/                 slash commands (section 7)
│   └── settings.json             hooks and permissions (section 8)
├── .mcp.json                     MCP servers for this workspace (section 7)
├── backlog/                      backlog.md board: tasks, milestones, docs, decisions
├── docs/
│   ├── discovery/                assumed requirements register, persona rehearsals
│   ├── design/                   HLD, LLDs, wireframes
│   ├── decisions/                ADRs (AA-nn), from the bridge's ADR template
│   └── delivery/                 delivery and assurance model, build plan
├── security/
│   ├── controls.yaml             control catalogue (single source of truth)
│   ├── threat-model/             Threagile model as code
│   ├── risk-register.md
│   ├── sbd-self-assessment.md
│   └── system-security-plan.md   from beta
├── catalogue/                    OSCAL catalogues, crosswalks, profiles
├── assessors/<name>/             one MCP server per assessor: server, SKILL.md, contract tests
├── orchestrator/                 assessment squad on Agent Server
├── evidence/                     evidence store, run manifest, findings normalisation
├── gates/                        OPA gate policies and their tests
├── reports/                      sector templates
├── portal/                       web UI
├── deploy/                       Helm chart, overlays, air-gap bundle tooling
├── tests/                        integration and end-to-end tests
└── tmp/                          git-ignored scratch, as owner rules require
```

## 4 How an item flows

An item starts in Backlog, where the lead authorises it once it has acceptance criteria and R-numbers; from there it follows the same eight steps every time.

> *Diagram not included in this export (item flow · 8 steps, 2 loops back). See the live document.*

Hooks catch most problems during step 2. CI and the platform's own assessors catch the rest at step 3; a failure goes back to the builder with the failing lines, never patched by hand. The reviewer runs on a different model class, and only the owner approves a merge. At the end of a phase, the evidence clerk runs `/gate-check` over everything Done, and the owner decides the gate.

## 5 Agents

The roles follow the bridge's board model (owner, lead, worker, reviewer), with specialist subagents added for design and assurance. Each is a definition in `.claude/agents/` with its own tools and model; workers on other seats (Zoo, Bob, Codex) take the same roles through AGENTS.md.

| Agent | Role | May work on | Tools | Must never |
| --- | --- | --- | --- | --- |
| Owner (Martin) | Approver and design authority; approves outcomes on evidence, not by reading code | Everything; every gate and risk acceptance | All | — |
| Lead | Triage, sprint planning, sweep, merge after review and approval | Moves items to To Do and Done; sets and clears flags | Board, bridge, GitHub, read-only repo | Approve its own `needs-approval` items or any gate |
| Builder | Implements one claimed item, test first, on its own branch and worktree | `task`, `feature`, `bug`, `chore`; `adlc-build` | Edit, tests, fast-lane assessors | Move to Done; edit `security/controls.yaml` without a `control` item; push to main |
| Assessor builder | Scaffolds and implements assessor MCP servers to the contract | Items labelled `component:assessors` | Edit, assessor scaffold skill, contract tests | Add a tool not rated by the supply chain and sovereignty assessors |
| Developer reviewer | Stands in for a senior engineer: reviews every non-spike change against standards and vendor docs (section 11) | Any item in In Review | Read-only repo, local RAG library, code analysis MCP servers | Change code or status; review work from its own model class; approve without citations |
| Security reviewer | Threat model and controls review for design changes and critical code | Items with `sbd:`, `ssdf:` or `tier:critical` labels | Threat model skill, Threagile, read-only repo, RAG | Accept risk |
| Designer | Drafts HLD sections, LLDs and ADRs | `docs`, `adr`, `spike` | Design skills, C4 model, read-only code | Accept an ADR (owner only) |
| Evidence clerk | Assembles gate evidence, coverage, the traceability matrix and the residual-risk statement | Gate epics, `evidence` items | Gate check, evidence pack, control catalogue skills | Mark evidence verified that it did not observe |
| Persona panel | Synthetic stakeholder review at sprint review and gates | Sprint and gate reviews | Persona interview skill | Present output as research; always labelled synthetic |

**Independence.** The developer reviewer and the security reviewer run on a different model class from the builder; critical code needs two developer reviewers from different providers. An item reaches Done only after the reviewers and the owner have each signed off, the owner on the evidence.

## 6 Skills

Twenty-four skills, grouped by the kind of work. Nine already exist (marked *existing*) and are reused as they are; fifteen are new workspace skills in `.claude/skills/`, each with the standard it serves and the artefact it produces.

| Group | Skill | Produces | Serves |
| --- | --- | --- | --- |
| Delivery | `phase-gate` | Gate epic status: exit criteria, evidence links, open items, approval request | Service Manual phases; RMF |
| Delivery | `sprint-plan` | Draft sprint from ready items within the WIP limit, for owner approval | Agile |
| Delivery | `sprint-review` | Demo notes, done items with evidence, persona review, retrospective data | Agile; Service Standard |
| Delivery | `writing-plans`, `executing-plans` *existing* | Implementation plan per item, then execution against it | Agile; owner rules §1 |
| Design | `adr` | ADR from the bridge template, numbered AA-nn, status Proposed | Owner rules; SbD p6 |
| Design | `hld-section` | An HLD or LLD section in the house structure, with traceability | Delivery model |
| Design | `threat-model` | Updated Threagile model, generated threats and risk register entries | SbD p3; SSDF PW.1 |
| Build | `assessor-scaffold` | New assessor: MCP server, SKILL.md, contract tests, runtime class, catalogue entry | AA-02, AA-03, AA-14 |
| Build | `oscal-profile` | Catalogue entries, crosswalks or a profile in OSCAL, validated | AA-01 |
| Build | `test-driven-development` *existing* | Failing test first, then code | Owner rules §6; SSDF PW.8 |
| Build | `systematic-debugging` *existing* | Root cause before fix | SSDF RV.3 |
| Assurance | `control-change` | Edits to `security/controls.yaml` with mapping, verification and owner | Delivery model 5.1 |
| Assurance | `gate-check` | Runs the platform's gate on the workspace itself; coverage and verdict | AA-06; SbD p9 |
| Assurance | `evidence-pack` | Evidence bundle, traceability matrix and SbD self-assessment update | RMF Assess; SbD p9 |
| Assurance | `requesting-code-review`, `verification-before-completion` *existing* | Independent review; proof before claiming done | Owner rules §6; SSDF PW.7 |
| Assurance | `developer-review` | Rubric review grounded in the RAG library and analysis tools, with cited findings and a verdict | Section 11; owner rules §4 and §6 |
| Assurance | `seeded-defects` | Refreshes the seeded-defect benchmark and measures the reviewer's catch and false-alarm rates | Section 11.4; R-28 |
| Assurance | `rag-ingest` | Adds a source to the RAG library with version, licence, hash and a proving query | Owner rules §5 |
| Discovery | `discovery-persona-builder` *existing* | Refreshed persona from public sources | Service Manual (synthetic) |
| Discovery | `discovery-persona-interview` *existing* | Synthetic rehearsal against a question or screen | Service Manual (synthetic) |
| Discovery | `discovery-synthesis` *existing* | Register updates from rehearsals | Assumed requirements register |
| Discovery | `requirements-trace` | Links an item to its R-numbers, design elements, controls and tests | Delivery model 8.5 |

**`assessor-scaffold` is the one that most speeds the build.** Every assessor follows one contract, so after the first one the scaffold produces the server, skill, tests and catalogue entry, and the builder only writes the tool-specific parsing.

## 7 Commands, MCP servers and tools

**Slash commands** are the fast way in: each starts the right skill with the right agent for a common job.

| Command | What it does | Runs |
| --- | --- | --- |
| `/item-start <id>` | Claims a ready item, creates its branch and worktree, writes the plan, traces it to its R-numbers | Lead, then builder; `writing-plans`, `requirements-trace` |
| `/adr <title>` | Drafts the next AA-nn decision and links it to the item | Designer; `adr` |
| `/threat-model` | Updates the threat model for the current change and raises risk entries | Security reviewer; `threat-model` |
| `/new-assessor <tool>` | Scaffolds a new assessor to the contract | Assessor builder; `assessor-scaffold` |
| `/review` | Independent review of the current branch | Reviewer; `requesting-code-review` |
| `/gate-check [phase]` | Runs the platform's own gate on the workspace and reports coverage and verdict | Evidence clerk; `gate-check` |
| `/evidence-pack` | Builds the evidence bundle and traceability matrix for a gate | Evidence clerk; `evidence-pack` |
| `/sprint-plan`, `/sprint-review` | Drafts the next sprint; runs the review with persona feedback | Lead, persona panel |

**MCP servers** in `.mcp.json`:

| Server | Used for | Notes |
| --- | --- | --- |
| agent-bridge | Messaging between seats; audit | Already deployed by the bridge |
| backlog.md | Reading and updating board tasks from agents | Confirm the MCP interface at the installed version before relying on it, per owner rules §2 |
| Platform assessors (static analysis, secrets, components, configuration) | Fast-lane checks during work, using the platform's own MCP servers | The workspace dogfoods the product from alpha sprint 1 |
| Workspace RAG | Vendored docs and standards, searched first, per owner rules §4 | The MCP-RAG repo is a candidate; to confirm |
| GitHub | Branches, pull requests, checks | `gh` CLI where an MCP server is not needed |
| LangSmith | Traces and evaluations of the orchestrator and LLM review assessor | Through DP-AI-001 |

**Command-line tools** the agents may call, under the worker-command rulebook: `semgrep`, `gitleaks`, `syft`, `grype` or `trivy`, `checkov`, `conftest` and `opa`, `threagile`, `pytest`, `ruff`, `oscal-cli` or `compliance-trestle`, `cosign`, `backlog`, `gh`.

## 8 Hooks and guardrails

Hooks enforce in the IDE what CI enforces at the pull request, so most problems are caught seconds after they are made. Claude Code hooks live in `.claude/settings.json`; other seats get the same rules through their own rule files and the CI gate.

| Hook | Checks | On failure | Serves |
| --- | --- | --- | --- |
| SessionStart | Loads owner rules, the claimed item, its R-numbers and the bridge inbox | — (context only) | Owner rules §1 |
| PreToolUse: file writes | Path is inside the workspace and the item's worktree; `security/controls.yaml` and `catalogue/` only under a matching item type | Blocks the write | SbD p7; owner rules §7 |
| PreToolUse: shell | Command is on the worker-command rulebook; no `--no-verify`, `--force`, `git add -A`, push to main, or reads of `.env` | Blocks the command | SSDF PS.1; owner rules §7 |
| PreToolUse: dependencies | A new package or tool has an approved item and a licence record | Blocks and raises `needs-approval` | SbD p2; SSDF PW.4 |
| PostToolUse: file writes | Fast lane: formatter, linter, gitleaks and fast Semgrep rules on the changed file | Returns findings to the agent to fix now | SSDF PW.7; AA-14 |
| Stop | The item's tests ran and passed, the board item has an evidence link, nothing is claimed done without proof | Blocks stopping and says what is missing | Definition of Done; owner rules §6 |
| SubagentStop (reviewer) | Review findings were recorded on the item with author and model | Blocks hand-over | Independent review |
| CI on pull request | Full SSDF stages 1 to 4 from the delivery model, plus the platform's own assessors and `gate-check` | Blocks merge | Delivery model 6 |

**Hooks are configuration under review.** Changing a hook, a permission or the rulebook needs its own item and the owner's approval, and every change is recorded in the bridge audit.

## 9 Agile cadence

Two-week sprints, run mostly by the lead agent, with the owner deciding at three points: the sprint plan, merges, and gates.

| Ceremony | When | Run by | Output | Owner's part |
| --- | --- | --- | --- | --- |
| Sprint planning | First day | Lead with `sprint-plan` | Sprint milestone with ready items, within one package and one item per worker in progress | Approves the plan |
| Daily sweep | Each working day | Lead | Accepted work moved to Done; blocked and flagged items commented; one message listing what waits on the owner | Clears the approval queue |
| Backlog refinement | Mid-sprint | Lead, designer | Next items made ready: acceptance criteria, R-numbers, security notes | Orders the backlog |
| Sprint review | Last day | Lead with `sprint-review`; persona panel | Demonstration, persona feedback (synthetic), register updates | Accepts the increment |
| Retrospective | Last day | Lead | Cycle time, hook blocks, gate failures, review findings, what to change in skills or hooks | Agrees changes |
| Phase gate | When exit criteria are met | Evidence clerk with `phase-gate` and `/gate-check` | Gate epic complete, evidence pack, persona review, verdict | Approves or rejects the gate |

**Measures the retrospective tracks:** cycle time per item, share of items passing review first time, hook blocks by rule, CI gate failures by stage, and time from claim to evidence linked. These show where a skill or hook should change to speed the build up.

## 10 Risk tiers

Following NCSC's principle that different code deserves different care, every item carries one tier label. The tier sets the review, tests and approval, so spikes move fast and critical code moves slowly.

| Tier | Covers | Review | Tests | Owner approves on |
| --- | --- | --- | --- | --- |
| `tier:spike` | Prototypes, experiments, throwaway code; never merged to main | Fast-lane hooks only | As useful | Findings, in the item notes |
| `tier:standard` | Most product code: portal views, report templates, catalogue data, assessor parsing | Developer reviewer agent; CI and the platform's own assessors | Unit and integration tests, written first | Evidence: tests, review findings, scan results |
| `tier:critical` | Gate engine and policies; evidence store, hashing and run manifest; sandboxing and tool execution; authentication, sessions and roles; cryptography and signing; secrets handling; anything that sends data outside the boundary | Developer reviewer agent on two model classes from different providers, both required; security reviewer; threat model update | Unit, integration, property-based and negative tests; mutation score above an agreed threshold; fuzzing for parsers | Evidence, plus a recorded residual-risk statement for each change |

**Critical paths are fixed in `CODEOWNERS`.** The paths above are listed there, so a change to any of them is tagged `tier:critical` automatically, whatever the item says. The `CODEOWNERS` file is itself critical.

## 11 Developer reviewer agent

With no human developer on the team, the developer reviewer agent stands in for the senior engineer who would review every change. It does not rely on what a model remembers: every judgement is grounded in a local library of standards and vendor documentation, backed by analysis tools, and checked against a measured accuracy. This narrows the gap left by having no human reviewer; it does not close it, so the residual risk is recorded at every gate.

### 11.1 Local RAG library

The library follows the owner rules' evidence order: workspace sources first, cited research second, model recall last and never as evidence. Every source is held at the version the code uses.

| Collection | Contents | Notes |
| --- | --- | --- |
| Security standards | OWASP ASVS 5.0.0; OWASP Top 10, Top 10 for LLM applications and Top 10 for agentic applications; CWE Top 25; OpenSSF guide for AI code assistant instructions; NCSC secure development and deployment guidance; NIST SSDF and SP 800-218A | Where a licence forbids copying, a read log with identifiers and citations is held instead |
| Coding standards | PEP 8, PEP 257, Python typing documentation; TypeScript handbook; the workspace's own coding rules | The workspace rules win where stricter |
| Vendor documentation | Python standard library, FastAPI, Pydantic, MCP specification and Python SDK, LangGraph, LangSmith and Agent Server at DP-AI-001's versions, OPA and Rego, OSCAL, Semgrep, Trivy, Syft, cosign, Keycloak, PostgreSQL, Helm | Pinned to installed versions; refreshed when a dependency changes |
| Project knowledge | HLD, LLDs, ADRs, requirements register, threat model, control catalogue, past review findings and escaped defects | Lets reviews check the change against the design, not only the standards |

Ingest follows owner rules sections 4 and 5: each source records its version, licence and sha256, is searchable by meaning and by exact text, and has a test query that must return it. A small local embedding model keeps the library usable in the air-gapped tier.

### 11.2 Code analysis tools

Each tool is wrapped as an MCP server to the platform's own assessor contract, so the workspace uses the product it is building.

| Purpose | Tools |
| --- | --- |
| Security analysis | Semgrep with OWASP and CWE rule sets, CodeQL, Bandit |
| Correctness and quality | ruff, mypy or pyright, radon for complexity, vulture for dead code |
| Test strength | pytest with coverage, mutmut for mutation testing, Hypothesis for property-based tests, Schemathesis for API fuzzing |
| Dependencies | pip-audit, Syft and Grype, ScanCode for licences, and a package-existence check against the registry to catch hallucinated package names |
| Secrets and policy | gitleaks; `opa test` and conftest for Rego |

### 11.3 Review protocol

The reviewer works from a rubric and the tools, not a persona prompt, which OpenSSF found can make results worse.

1. Read the item, its R-numbers, its tier, the relevant LLD and ADRs, and the diff.
2. Run the tools for the changed files and collect their findings.
3. Retrieve the standards and vendor sections that apply to each changed area.
4. Review against the rubric: meets the acceptance criteria; secure against the mapped ASVS requirements; correct use of each vendor API at the pinned version; conforms to the design; tests include negative cases and kill mutants; no dead code; simplest workable design; documented.
5. Record each finding with severity, file and line, the standard or vendor source and version it relies on, and a suggested fix. A finding with no citation is discarded.
6. Return a verdict: approve, changes needed, or escalate. Escalate when confidence is low or, for critical code, when the two reviewers disagree; the owner then sees the disagreement and the residual risk.

### 11.4 Measuring the reviewer

The reviewer's own accuracy is measured, the same way LLM rubrics are calibrated (R-28).

- **Seeded defects.** A benchmark of known-vulnerable samples and deliberately injected defects in the project's own code, refreshed each sprint. The reviewer's catch rate and false-alarm rate are measured per category: injection, access control, cryptography, error handling, concurrency, test weakness.
- **Gate eligibility.** The reviewer's verdict counts toward a gate only in categories where it meets the agreed catch rate. Elsewhere it is advisory, and the item needs the extra tests in the critical tier.
- **Escaped defects.** Every bug found after merge is traced back to the review that missed it, and the retrospective adds a rubric rule, a Semgrep rule or library content so it is caught next time.
- **Residual risk.** Each phase gate records that the code has had no human developer review, with the reviewer's measured accuracy, so the decision to accept that risk is explicit.

## 12 Containment and workspace threat model

NCSC's agentic AI advice is to enforce the boundary outside the model and its harness. Hooks (section 8) run inside the agent's harness, so they are the fast feedback layer; the controls below are the boundary.

| Control | How | NCSC principle |
| --- | --- | --- |
| Isolated runtime | Agents run in a dev container or VM per workspace, never on the host directly | Compute isolation matched to consequence |
| Network deny by default | Outbound traffic only to an allowlist: package registries, GitHub, the model gateway and the RAG store, through a proxy | Network isolation |
| Agent identities | Each seat has its own GitHub identity and short-lived token, scoped to its branches; no agent can push to main or change branch protection | Credentials: least privilege, short lifetime |
| Server-side protection | GitHub branch protection: main needs passing checks, the developer reviewer's approval and the owner's approval; `CODEOWNERS` enforced | Human approval at high-impact actions |
| Secrets out of reach | No secrets in the workspace; the proxy injects credentials the agent never sees | Credentials |
| Protected logs | The bridge's hash-chained audit and the container's tool, process and network logs are written where agents cannot alter them | Observability |
| Kill switch | One command, independent of the agents, revokes their tokens and stops the container; tested once a sprint | Emergency stop |

**Workspace threat model.** Kept in `security/threat-model/` alongside the product's, and built from the OWASP Top 10 for agentic applications:

| Threat | Example in this workspace | Main controls |
| --- | --- | --- |
| Prompt injection | Instructions hidden in a dependency's README, a board item, a fetched web page or an assessor's output | Tool output schema validation; content treated as data; network allowlist; reviewer on a different model |
| Tool or MCP server poisoning | A malicious MCP server or a changed tool description | MCP servers pinned by digest and vetted by the platform's own supply chain and sovereignty assessors before use |
| Excessive agency | An agent deletes files, rewrites history or approves its own work | Roles and hooks; agent identities; branch protection; no approval rights for any agent |
| Supply chain | A hallucinated or typo-squatted package | Package-existence check; approval for new dependencies; SBOM and scanning |
| Data leakage | Code or secrets sent to an unapproved service | Network deny by default; secrets out of reach; model access only through the DP-AI-001 gateway |
| Context and memory poisoning | Bad content added to the RAG library or project knowledge | RAG ingest records source, version and hash; additions go through review like code |

## 13 Lean start, and when to bring in a human developer

The full workspace is the target, not the starting point. Long, unfocused instructions make agents perform worse, so it starts small and grows only when the retrospective shows a need.

| Stage | Agents | Skills | Controls |
| --- | --- | --- | --- |
| Alpha preparation | Lead, builder, developer reviewer | `assessor-scaffold`, `adr`, `threat-model`, `gate-check`, `requirements-trace` | Containment layer, branch protection, hooks in warning mode, RAG with security standards and the first vendor docs |
| Alpha sprint 2 onward | Add security reviewer and evidence clerk | Add `control-change`, `evidence-pack`, `phase-gate` | Hooks enforcing; reviewer accuracy measured on seeded defects |
| Beta | Add designer, assessor builder and persona panel | The rest, as needed | Second reviewer model for critical code; full RAG library |

**Triggers for a human developer.** Bring one in, even part-time, when the first of these happens:

- the decision point on 21 June 2027 says to take the platform further;
- the platform is about to handle anyone else's code or data;
- the developer reviewer misses a critical-tier defect that is found later;
- the reviewer cannot reach its catch-rate threshold for a critical category after two sprints.

Until then, the developer reviewer agent, the critical-tier tests and the recorded residual risk stand in for that review.

## 14 Rollout through agent-bridge, and open decisions

The workspace ships as a new agent-bridge bundle, `integrations/innovox-assurance/1.0`, built the same way as the existing backlog.md and creative-writing bundles: an `install.json` naming what goes where, and assets copied into the target workspace by `agent-bridge init`.

| Bundle part | Source | Change from today's bridge |
| --- | --- | --- |
| Board config | backlog.md 1.55 bundle | Keeps the five columns, flags and ADLC labels; adds the delivery model's label families (`phase:`, `component:`, `sbd:`, `ssdf:`, `ctl:`, `asvs:`, `defstan:`, `layer:`, `rmf:`, `agent:`), item types (`control`, `evidence`, `icd`, `adr`), and the SSDF items in the Definition of Done |
| Board documents | backlog.md bundle | Adds a phase gate guide and the gate epic templates for discovery, alpha and beta |
| Rules | `library/owner-rules.md`, `rules/worker-commands.json` | Unchanged; CLAUDE.md and AGENTS.md point to them and add the delivery model's Definition of Done |
| Agents, skills, commands, hooks | New | Sections 5 to 8 of this design |
| MCP config | Bridge host registry | Adds the servers in section 7 to `.mcp.json` |

**One reconciliation.** The delivery model lists six statuses (To Do, Ready, In Progress, In Review, Done, Blocked). The workspace adopts the bridge's model instead: Backlog, To Do (meaning ready), In Progress, In Review and Done, with `blocked` and `needs-approval` as flags. The delivery model will be updated to match.

**Order of work**

1. Create the `innovox-assurance` repository (owner), with branch protection, CODEOWNERS and an agent identity per seat.
2. Build the bundle in agent-bridge and run `agent-bridge init` on the new repository.
3. Build the containment layer and the first RAG library (security standards and the first vendor docs), then write developer-review, rag-ingest and the five skills used in alpha sprint 1: `assessor-scaffold`, `adr`, `threat-model`, `gate-check`, `requirements-trace`.
4. Load the backlog: 24 epics, gate epics and the alpha sprint 1 items.
5. Turn on hooks in warning mode for one sprint, then enforcing mode.

**Open decisions**

| Decision | Options |
| --- | --- |
| Which seats build: Claude Code only, or Claude Code with Zoo, Bob or Codex as workers | Start with Claude Code; add a second seat for review independence if model-class separation through subagents is not enough |
| Reviewer model class | Two providers through DP-AI-001's router, required for critical code; choose the two by measured catch rate on the seeded defects |
| backlog.md MCP interface | Confirm at the installed version; otherwise agents use the `backlog` command line |
| Workspace RAG | MCP-RAG or another store, holding vendored standards and tool documentation |
| Hook enforcement start | After one sprint in warning mode, or from the first commit |
