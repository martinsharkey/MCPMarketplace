# Migration early-warning — product design 0.2.0 (draft)

Draft for review, 10 November 2026. Supersedes 0.1.0. Built on EIS 0.4.0.

**Changes from 0.1.0**

- Repositioned as a commercial product for border, policing and humanitarian buyers.
- Added the pressure and release model layers, covariates, lag definitions and
  pooled chaining.
- Added alerts and email notifications.
- Added the rewind demonstration, 14-day crossing outlook, explainer, morning brief,
  accuracy scorecard and scenarios.
- Added statistics panels, a route heat map, and a readiness and capacity view.
- Added the security assurance model: Defence Secure by Design for risk governance and
  the NIST Secure Software Development Framework for development (section 14).
- Recorded responses to the external review (section 16).

## 1. Product and boundary

An operational early-warning product for irregular migration toward the UK. It shows
where pressure is building upstream, when it is likely to be released across the
Channel, and what readiness that implies. It proves its own foresight by replaying past
surges using only the data that was published at the time.

**Buyers:** UK border, maritime and policing organisations; search-and-rescue and
reception planners; and, later, European partners, humanitarian organisations and
analysts.

| In scope | Out of scope |
|---|---|
| Aggregate published counts by place, route, nationality, age band, sex and period | Individual-level data, tracking, profiling or identification of people |
| Early warning of the migration *response* to events and conditions | Predicting the triggering events themselves |
| Probabilities with intervals, lags and full audit trails | Point forecasts presented as certain |
| UK-side readiness and capacity planning across interception, rescue and reception | Recommendations for assets in another state's jurisdiction |
| Public sources licensed for commercial use | Restricted or official-sensitive data in the product baseline |

"Trails" and heat maps show **aggregate flows** whose intensity represents volume or
probability. They never represent individual journeys.

## 2. Signature capabilities

| Capability | What the buyer sees | Why it matters |
|---|---|---|
| **Rewind** | Scrub to a past surge (for example the late-2022 rise in crossings). Upstream nodes light up, the warning fires and the email that would have been sent appears with its lead time. The system only sees what had been published by each date. | Proof of foresight, not a claim of it |
| **14-day crossing outlook** | Daily expected-arrivals range for the next fortnight, combining upstream pressure, weather windows, tides and suppressed backlog | Immediately actionable for staffing, rescue cover and reception |
| **Live alerts** | Corridor and node status lights; pressure and release alerts; email and digest notifications | Turns analysis into an operational system |
| **Explain** | A plain-English question ("why is risk up this week?") answered from the provenance graph, citing nodes, links and receipts | Explainability that holds up under scrutiny |
| **Morning brief** | One page: changes since yesterday, active alerts, outlook, confidence | Fits existing operational routines |
| **Accuracy scorecard** | Every forecast and alert scored against outcomes, with the Brier score and calibration | Honest track record, which earns analysts' trust |
| **Scenarios** | Sliders such as "French prevention rate +20%" or "upstream surge on route X" | Policy and planning conversations |
| **Statistics and heat map** | Landed since 2018, attempts and success rate, deaths, returns, age and sex bands, routes by intensity | The complete picture in one place |
| **Readiness view** | Implied demand on UK interception, rescue and reception capacity by sector and day | Converts forecasts into resourcing |

## 3. Platform reuse

| Element | Treatment | Notes |
|---|---|---|
| Native EIS 0.4.0 (ledger, NiFi, Camel, Kafka, Kafka Connect/Debezium, archive) | **Reused unchanged** | `nativeCoreModified: false`. Separate namespace `migration-acquisition` |
| EIS solution overlay structure (`solution.json`, `connectors/`, `scopes/`, `artifacts.json`, interface control documents) | **Standard EIS pattern** | Connectors stay disabled until approved |
| Generic NiFi flow pack, Camel HTTP component, `anonymous-public-read`, `reviewed-public-sources`, `raw-durable` | **Reused unchanged** | New endpoints and paced retry policies |
| EIS post-receipt processor phases and review findings | **Standard EIS pattern** | New format extractors; review status also covers LLM event tags |
| EIS ledger plus committed outbox | **Standard EIS pattern** | Also used for alerts and notifications (section 9) |
| Downstream consumer: independent Kafka consumer group, offsets acknowledged after committed projection, rebuildable read model | **Standard EIS pattern** | Implemented new as the migration projector |
| Backend for frontend: FastAPI, server-side sessions, OIDC with PKCE, CSRF, audit, row-level security | **New build on a standard stack** | Read, alert and subscription endpoints |
| Web: React, TypeScript, TanStack Query, server-sent events | **New build on a standard stack** | Map, timeline, statistics, alerts and readiness views |
| Keycloak | **Hosted** | Section 11 |
| Compose generation tooling | **New, generalised** | Development and hosted profiles |
| EIS NCSC/OWASP assurance gates, axe and Playwright accessibility checks | **Reused** | Extended to new views and email templates |

## 4. Logical architecture

```text
Public sources: UK arrivals, preventions, returns · Frontex · UNHCR · Eurostat · IOM ·
                Italy MoI · weather and sea state · tides · calendars · UCDP · GDELT
        │  read-only, paced, anonymous
NiFi schedule → Camel HTTP → archive (immutable version, SHA-256) → ledger receipt
        │
Kafka  document.received  (unchanged EIS envelope, opaque body)
        │
migration-acquisition processors: extract → harmonise → enrich events
        │
        ├──────────────────────────────────────────┐
migration-model (debounced after projection)   migration-portal projector
  pressure layer:  baselines → spikes →          observations, events,
                   lags → pooled links           nodes, edges, freshness
  release layer:   weather, tide, backlog,                │
                   calendar → 14-day outlook              │
  readiness, scenarios, backtests, scoring                │
        │ model run + alerts (same transaction) + outbox  │
        └──────────────► PostgreSQL + PostGIS ◄───────────┘
                               │                     │
               notification worker (email)   FastAPI BFF (Keycloak) ─ explain (LLM)
                                                     │
          React web: map · heat map · timeline · alerts · statistics · readiness ·
                     rewind · scorecard · scenarios · brief
```

The model never polls sources or reads EIS tables. The portal never runs the model.
Every forecast and alert references its model run, which references the exact receipt
versions it used.

## 5. Acquisition overlay

One connector per source, read-only, `payloadMode: opaque-entity-bytes`,
`archiveBeforeInterpretation: true`, disabled until approved. Full candidate list,
cadence and licence status are in the [source register](source-register.md).

| Group | Connectors | Native forms |
|---|---|---|
| UK spine | `uk-small-boats-daily`, `uk-preventions-weekly`, `uk-illegal-entry-quarterly`, `uk-returns` | ODS/CSV/xlsx on GOV.UK publication pages |
| Upstream flows | `frontex-detections`, `unhcr-odp`, `eurostat-asylum`, `iom-dtm`, `italy-moi-daily` | xlsx, REST JSON, SDMX/JSON, files, PDF |
| Humanitarian | `iom-missing-migrants` | CSV |
| Release covariates | `uk-crossing-conditions`, `marine-reanalysis`, `marine-forecast`, `tides-channel`, `calendars` | xlsx, gridded NetCDF/GRIB subsets, tide predictions, curated calendar files |
| Events | `ucdp-ged`, `gdelt-events`, `policy-markers` | REST JSON, zipped CSV, curated and reviewed file |

Standard EIS overlay rules apply:

- Scopes freeze a native window; limits are incomplete-work budgets, never truncation.
- A file source is complete when its listed file is landed and hashed, not when it parses.
- Parsing failures produce review findings and never remove a receipt.

**Publications are vintages.** Provisional daily UK figures are superseded by quarterly
releases, and other sources revise retrospectively. Every receipt is kept as its own
version, so rewind, backtests and scoring use exactly what was known on each date.

## 6. Harmonised read model

PostgreSQL with PostGIS, in a separate `migration` database with row-level security by
namespace.

| Table | Key fields | Notes |
|---|---|---|
| `node` | id, kind (country, crossing, sea area, coastal sector, port, UK), name, geometry, parent, jurisdiction | Curated and versioned in Git; French coastal sectors are coarse, never individual beaches |
| `edge` | from_node, to_node, mode, route family, geometry | Generated sea paths and simplified land paths, reviewed |
| `measure` | id, definition, unit, source, comparability notes | Detections, arrivals, preventions, attempts, applications, deaths and returns are distinct and never summed |
| `observation` | node, measure, nationality, age band, sex, period start and end, granularity, value, provisional flag, reference period, published_at, receipt and version | One row per published value per vintage |
| `covariate` | node or sea area, variable (navigability, wave height, wind, tide range, holiday flag), period, value, forecast or observed, issued_at, receipt | Forecast covariates keep their issue time |
| `event` | id, source, occurred_at, location, category, intensity, enrichment status, receipt | Links and derived tags only, no article text |
| `policy_marker` | id, jurisdiction, effective date, category, description, source link, review status | Breaks on the timeline, not model covariates |
| `spike` | node, measure, nationality, period, magnitude, model run | |
| `link_estimate` | from_node, to_node, nationality, transit lag distribution, publication lag, conditional probability, base rate, interval, sample size, status | Status: `estimated`, `insufficient_history` or `failed_backtest` |
| `forecast` | target, nationality or aggregate, layer (pressure or release), horizon, probability, expected range, interval, contributing links and covariates, model run | |
| `readiness` | sector, day, implied interceptions, rescue demand, reception demand, interval, model run | UK-side assets only |
| `alert` | id, type, level, target, opened_at, state, evidence, model run, outcome, scored_at | Lifecycle in section 9 |
| `subscription` | user, target filter, minimum level, channel, digest preference | Keyed to Keycloak subject |
| `notification` | alert, user, channel, idempotency key, sent_at, status | Written via outbox |
| `model_run` | id, code version, parameters, vintage cut-off, input receipt set hash, trigger, created_at | Every output references one |

Nationality codes are ISO 3166-1 alpha-3, with source groupings such as "Other"
preserved as published. Age and sex are stored only in the bands each source publishes.

## 7. Modelling method

### 7.1 Two layers

- **Pressure** (weeks to months): how much movement is building upstream and in
  northern France.
- **Release** (days): which days that pressure actually crosses, governed by sea state,
  wind, tides, backlog and calendar.

Upstream signal decides how many. Local conditions decide when.

### 7.2 Lag definitions

Each link records two separate lags:

- **Transit lag** (Δt_ref): the time between reference periods at A and B, meaning when
  movement happened.
- **Publication lag** (Δt_pub): the time between the end of A's reference period and
  when its data landed in the ledger.

The usable lead time of a warning is transit lag minus publication lag. Estimators
condition on the vintages observable at the forecast date, so real-time lead times are
reported honestly. A monthly source published four weeks late cannot warn of a
three-week transit.

### 7.3 Pressure layer

1. **Baseline:** seasonal decomposition per series. For the UK node, use arrivals per
   navigable day, or include navigability as an exogenous covariate, so calm spells are
   not mistaken for corridor spikes.
2. **Spikes:** robust residual thresholds with a minimum duration, set as parameters on
   the model run.
3. **Links:** P(spike at B within L | spike at A) against the base rate, with
   Beta-binomial intervals and the observed transit-lag distribution.
4. **Combining links:** a single path is explained by its chain of links. When several
   upstream links fire together for the same nationality, combine them by log-odds
   pooling (Noisy-OR, or L2-penalised logistic regression on lagged spike indicators)
   calibrated in the backtest. Never sum or naively multiply probabilities across
   correlated paths. Share-reaching-UK terms cover Mediterranean and Schengen legs, and
   inferred links are marked as inferred.
5. **Nationality at the UK node:** daily UK data has no nationality. Forecasts therefore
   run in two modes:
   - **Operational:** nationality-specific upstream pressure is weighted into an
     aggregate daily UK pressure band.
   - **Retrospective:** calibrated each quarter by nationality against the quarterly
     release.

   Temporal disaggregation (for example Chow-Lin) is deferred until the simpler
   approach is shown to be insufficient.
6. **Sample guard:** links below a minimum number of upstream spikes are marked
   insufficient history and excluded.

### 7.4 Release layer and 14-day outlook

- **Covariates:**
  - Daily navigability ratings (the Home Office crossing-conditions series)
  - Wind and wave reanalysis for history, and marine forecasts for the outlook
  - Spring/neap phase and the timing of tidal windows at sector level
  - Calendar flags: public holidays in France and the UK, Ramadan and Eid, school
    holidays
- **Backlog state:** latent pressure accumulates on non-navigable days and is released
  on navigable ones (the "coiled spring" effect).
- **Model:** a count model (negative binomial or state-space) for daily arrivals, with
  pressure as an input and weather, tide and calendar as covariates. The outlook uses
  each covariate's forecast as issued, never later observations.
- Calendar and tide terms are **hypotheses**: each is kept only if it improves
  out-of-time skill.

### 7.5 Policy markers

Changes to returns arrangements, enforcement funding, asylum support rules and
bilateral agreements are recorded as reviewed **markers**. They are not model
covariates, because changes are too rare to estimate statistically. The interface
shows them as breaks, and the backtest reports whether link behaviour shifted around
them. Wording on the markers is neutral and descriptive.

### 7.6 Validation and scoring

- **Out of time:** fit up to a vintage cut-off, then score later periods with the Brier
  score, calibration and interval coverage, against baselines that ignore upstream and
  covariate data.
- **Promotion gate:** a link, covariate or alert rule reaches the default view only if
  it beats its baseline.
- **Live scoring:** every expired forecast and alert is scored automatically, feeding
  the public accuracy scorecard.

### 7.7 Readiness

Converts the outlook into implied UK-side demand by sector and day: interceptions,
search-and-rescue cover and reception throughput, each with intervals. Capacity
figures are buyer-supplied configuration, not inferred. The view shows demand against
configured capacity. It does not recommend assets in French jurisdiction and does not
predict individual launch points.

### 7.8 Events and explanation

- **Events:** categorised and geolocated by an LLM enrichment step, with review status
  and unreviewed tags labelled. Events are not a forecast input until a backtest shows
  they add skill.
- **Explain:** answers questions using only retrieved read-model records (nodes,
  links, forecasts, receipts) and cites every claim back to provenance. It has no
  free-text access to the web.

### 7.9 Runs

- A containerised Python batch job (pandas, statsmodels, NetworkX, a probabilistic
  programming library for the release layer).
- Triggered after projection with a 15-minute debounce, plus a weekly scheduled run.
- Fully reproducible from a model run record.

## 8. User experience

| View | Content |
|---|---|
| **Operations** (landing) | Corridor status light, active alerts, 14-day outlook strip, readiness against capacity, freshness of each feed |
| **Map** | Nodes sized by volume and coloured by status; routes; animated aggregate flows; inferred links dashed; forecast flows faded |
| **Heat map** | Route intensity by nationality and period, time-lapse on the shared timeline, coloured by warning level |
| **Timeline** (shared cursor) | History, now and forecast zones; event ticks; policy markers; play, scrub and **rewind mode** |
| **Statistics** | Landed since 2018; attempts (arrivals plus preventions) and success rate where preventions data exists; returns; age and sex bands; all with source and vintage |
| **Lives lost** | A dedicated, respectful panel: deaths and missing on Channel and upstream routes, sex and child/adult status where known, unknowns shown explicitly |
| **Node detail** | Series with baseline and spikes, nationality breakdown, freshness, links with transit and publication lags, provenance trail |
| **Alerts** | Open, acknowledged and resolved alerts with evidence, history and outcome |
| **Forecasts and scenarios** | UK-facing warnings ("this upstream rise has preceded a UK increase in X of Y cases within Z weeks") and scenario sliders |
| **Scorecard** | Calibration, Brier score, hit and false-alarm rates, by alert type and horizon |
| **Explain** | Question box with cited answers |
| **Morning brief** | Generated daily, viewable and emailable |
| **Interfaces and platform** | Source status, last successful window, bounded operator controls and native console links |

Every node and panel carries a **freshness badge** showing its last publication date
and granularity. "Now" means the latest publication.

**Rewind mode** sets a vintage cut-off on every request. The BFF serves only receipts,
model runs and alerts available at that date, and the interface shows the historical
alerts and notifications as they fired.

**Front-end additions**

| Need | Library |
|---|---|
| Base map | MapLibre GL JS with a self-hosted Protomaps PMTiles basemap, served by the reverse proxy with HTTP range requests |
| Layers | deck.gl: `ScatterplotLayer` or `H3HexagonLayer`, `PathLayer`, `ArcLayer`, `TripsLayer`, `HeatmapLayer` |
| Timeline and charts | Observable Plot |
| State | TanStack Query, server-sent events, one shared time-cursor and vintage store |

**Accessibility:** every map state has a table equivalent, animation can be paused and
respects `prefers-reduced-motion`, and status is never conveyed by colour alone.

## 9. Alerts and notifications

**Types**

- **Pressure:** the probability of a UK increase within the 4–12 week horizon crosses a
  threshold. Only links that passed the backtest can raise one.
- **Release:** a navigable window opens in the next 72 hours while backlog is high.
- **Data:** a feed is overdue, its layout has changed, or it was revised materially.
  Data alerts go to operators only.

**Levels:** green, amber and red per node and for the UK corridor.

**Lifecycle**

1. Raised.
2. Acknowledged, by whom and when.
3. Updated, escalated or de-escalated.
4. Resolved or expired.
5. Scored against the outcome.

**Fatigue controls**

- Hysteresis: for example, raise at 70% and clear at 55%.
- Notify only on new, escalated or resolved alerts.
- Cooldown and deduplication per target.
- Minimum lead time.
- Daily digest option.

**Delivery**

- The model run writes alerts and outbox rows in one transaction. A notification worker
  sends with an idempotency key, so alerts cannot be lost or emailed twice.
- **Email** goes through a UK-region provider. It contains the headline, what changed
  and a link to the evidence behind Keycloak login, with no detailed data in the
  message body.
- **Subscriptions** are per user, by target, nationality, minimum level and channel.
  Teams and Slack webhooks follow later.

## 10. BFF API outline

All read endpoints require an authenticated session and accept an optional `asOf`
vintage cut-off. Every response carries its model run and vintage.

| Endpoint | Returns |
|---|---|
| `GET /api/nodes`, `GET /api/edges` | GeoJSON with status and freshness |
| `GET /api/flows`, `GET /api/heatmap` | Time-binned aggregate flows and route intensity |
| `GET /api/nodes/{id}/series` | Observations, baseline and spikes |
| `GET /api/links/{from}/{to}` | Lags, probability, interval, sample size |
| `GET /api/forecasts`, `GET /api/outlook` | Pressure warnings and the 14-day release outlook |
| `GET /api/readiness` | Implied demand against configured capacity |
| `GET /api/statistics` | Landed, attempts, success rate, returns, age and sex bands, deaths |
| `GET /api/alerts`, `POST /api/alerts/{id}/ack` | Alerts and acknowledgement (audited, CSRF-checked) |
| `GET/PUT /api/subscriptions` | The user's notification preferences |
| `POST /api/scenarios` | Scenario results (not persisted as forecasts) |
| `POST /api/explain` | Cited answer from read-model retrieval |
| `GET /api/scorecard`, `GET /api/backtests/{runId}` | Scores and case studies |
| `GET /api/brief` | Morning brief |
| `GET /api/provenance/{ref}` | Receipt, version, hash and source link |
| `GET /api/stream` | Server-sent events for receipts, model runs and alerts |

## 11. Identity and security (hosted Keycloak)

- **Realm and hosting:** realm `migration` on hosted Keycloak with its own database, on
  its own HTTPS hostname. The admin console is reachable only from an operator network
  or VPN.
- **Client:** confidential, authorization code with PKCE, state and nonce. Sessions stay
  server-side.
- **Roles**
  - `viewer`: all read views and alert subscription
  - `duty-officer`: viewer plus alert acknowledgement
  - `analyst`: adds model parameters, review of event tags and policy markers, and
    capacity configuration
  - `operator`: adds interface controls and data alerts
- **Accounts:** invitation only. MFA is required for every role except viewer and
  recommended for viewers. Short idle timeout and brute-force detection.
- **Federation:** identity brokering to a buyer's own identity provider, so their staff
  use work accounts.
- **Tenancy:** each buyer gets its own deployment or namespace, so subscriptions,
  capacity figures and acknowledgements never mix.
- **Web hardening:**
  - Strict Content Security Policy, helped by self-hosted tiles and fonts
  - HTTPS only, secure and `SameSite` cookies, CSRF checks
  - Safe rendering of source text and URLs
  - The explain feature is bounded to retrieved records, and its prompts and answers
    are audited

## 12. Containers and hosting

| Container | Image | Exposure |
|---|---|---|
| `proxy` | Caddy (automatic TLS, static PMTiles with range requests, no Docker socket) | **Public**: portal and identity hostnames |
| `identity`, `identity-db` | Keycloak 26.x (pinned digest), PostgreSQL | Identity via proxy |
| `web` | `migration-portal-web` | Via proxy |
| `portal-api` | `migration-portal-api` (BFF) | Via proxy (`/api`) |
| `projector` | `migration-portal-api` in projector mode | Internal |
| `model` | `migration-model` | Internal |
| `notifier` | `migration-portal-api` in notifier mode | Internal; outbound to the email provider only |
| `protocol` | `migration-integration` (overlay processors) | Internal |
| `workflow`, `adapter`, `ledger`, `source`, `broker`, `database` | Native EIS 0.4.0 images unchanged | Internal; NiFi UI on the management network only |

- **First target:** a single UK-region cloud VM running Docker Compose, generated by the
  compose tooling, with public, integration and management networks.
- **Memory:** strict limits on the JVM containers (NiFi, Kafka, Keycloak).
- **Secrets:** taken from the host secret store.
- **Images:** pinned by digest and recorded in `artifacts.json`.
- **Backups:** nightly backups of the ledger, archive, `migration` and Keycloak
  databases.
- **Scale-out:** Kubernetes only when a buyer requires it.

## 13. Responsible design and commercial readiness

- **Data:** aggregate public data only, with no individual-level fields anywhere.
  Small cells are suppressed or banded.
- **Operational sensitivity:** French coastal sectors stay coarse and the outlook is
  behind login. There are no beach-level launch predictions, which would be
  operationally sensitive and useful to smugglers.
- **Lives lost:** shown with care and visible uncertainty.
- **Wording:** "situational awareness", "early warning" and "readiness". Every alert
  shows its inputs, assumptions, confidence and model run.
- **Buyer assurance pack:** a data protection impact assessment template, an ethics
  statement, the NCSC and OWASP evidence, the accuracy scorecard and the licence
  register.
- **Licences:** every source must permit commercial reuse, or be licensed for it,
  before it ships in the product. Attribution is displayed in the interface.
- **Ownership:** the product is developed and owned by Innovox Limited.
- **Route to market:** a G-Cloud listing under Innovox.

## 14. Security assurance: Defence Secure by Design and NIST

### 14.1 Approach

| Framework | Role in this product |
|---|---|
| **Defence Secure by Design (SbD)** | Governs cyber risk across the whole lifecycle. MOD mandates it through JSP 440 Leaflet 5C, and Industry Security Notice 2023/09 sets supplier requirements. Cross-government SbD principles provide the control headings |
| **NIST SP 800-218 Secure Software Development Framework (SSDF) 1.1** | The development practice baseline: Prepare the Organization (PO), Protect the Software (PS), Produce Well-Secured Software (PW), Respond to Vulnerabilities (RV). Revision 800-218r1 (draft, December 2025, adds PO.6 and PS.4) is tracked and adopted when final |
| **NIST SP 800-218A** | SSDF profile for the generative AI components: event enrichment and explain |
| **NIST SP 800-37 Risk Management Framework (RMF)** | Lifecycle steps used to structure authorisation evidence alongside the Defence lifecycle |
| **NCSC principles and OWASP ASVS 5.0.0** | Carried forward unchanged from the EIS security assurance (`eis/design/security-assurance.md` in the TenderHub repository), including its Bandit, `pip-audit`, Dependency-Check and ZAP gates |

SbD is treated as continual risk management, not a one-time accreditation. Controls are
proportionate to the product's criticality. Security is designed in from the concept
stage and re-assessed on every release.

**Classification:** the product baseline holds public data at OFFICIAL. Buyer-supplied
capacity figures, subscriptions and acknowledgements may be OFFICIAL or
OFFICIAL-SENSITIVE, depending on the buyer. They are therefore tenant-scoped and
optional, and the hosting profile for a buyer must meet the highest classification
that buyer stores.

### 14.2 Lifecycle mapping

The Defence lifecycle stages (Concept, Assessment, Demonstration, Manufacture,
In-Service and Disposal or Termination, known as CADMID/T) are mapped to RMF steps and
product phases. The mapping is indicative; each buyer's assurance team confirms its
own.

| CADMID/T stage | RMF steps | Product phase (section 15) | Security outputs |
|---|---|---|---|
| Concept | Prepare, Categorize | Design 0.x | Security categorisation, context and data-flow diagram, initial threat model, risk register, SbD self-assessment v0, DPIA screening |
| Assessment | Select | 0 | Control baseline (SbD principles, SSDF tasks, ASVS level 2), architecture risk review, supplier and component assessment (cloud, email, LLM, base images) |
| Demonstration | Implement, Assess | 1–6 | SSDF practices enforced in CI, threat model updated per design change, security testing per release, evidence pack maintained |
| Manufacture | Authorize | 7 | Independent IT health check (penetration test), residual risk statement, buyer risk-owner acceptance |
| In-Service | Monitor | Operations | Continuous assurance, vulnerability management, change control, periodic re-assessment |
| Disposal / Termination | — (lifecycle activity) | Exit | Tenant data return and verified destruction, credential and key revocation, archive disposition, records retained as agreed |

### 14.3 Secure by Design principles

| Principle | How the product meets it |
|---|---|
| 1. Create responsibility for cyber security risk | Named Innovox security lead as design authority; buyer risk owner named per deployment; responsibility matrix in the assurance pack |
| 2. Source secure technology products | Digest-pinned images and locked dependencies; SBOM per release; supplier assessment for cloud, email and LLM providers; native EIS reused rather than rebuilt |
| 3. Adopt a risk-driven approach | Threat model per component (STRIDE), risk register reviewed each release; aggregate-only data limits impact |
| 4. Design usable security controls | Keycloak single sign-on and federation to the buyer's identity provider; MFA; roles match real duties (viewer, duty officer, analyst, operator) |
| 5. Build in detect and respond security | Audited actions, centralised logs, alerts on authentication anomalies and data-feed tampering, incident response runbook |
| 6. Design flexible architectures | Overlay pattern; independently released components; containers portable across cloud providers |
| 7. Minimise the attack surface | Only the proxy is public; no Docker socket; admin consoles on the management network; read-only, outbound-only source access; no inbound path to EIS |
| 8. Defend in depth | Segmented networks, row-level security, least-privilege database roles, server-side sessions, CSP, encryption in transit and at rest |
| 9. Embed continuous assurance | CI security gates on every change; scheduled DAST; SbD self-assessment updated per release; annual IT health check |
| 10. Make changes securely | Protected branches, reviewed pull requests, signed commits and artifacts, digest-pinned promotion from dev through test and pre-production to production, additive migrations with rollback qualified |

### 14.4 NIST SSDF practices

| Group | Implementation |
|---|---|
| **PO** Prepare the Organization | Security requirements in each design document; defined roles and training; CI toolchain with defined security check criteria; separate development, test and production environments |
| **PS** Protect the Software | Branch protection and MFA on the repository; signed commits; digest records in `artifacts.json`, extended to signed images and provenance attestations; release archives retained with their SBOM |
| **PW** Produce Well-Secured Software | Threat modelling per design change; reuse of vetted components; code review; SAST with Bandit, extended to the model and portal code and the TypeScript front end; SCA with `pip-audit`, Dependency-Check and `npm audit`; secret scanning; container and infrastructure-as-code scanning; ZAP DAST against the BFF; secure defaults (connectors disabled, invitation-only accounts) |
| **RV** Respond to Vulnerabilities | Published vulnerability disclosure policy and `security.txt`; triage with severity-based fix targets; root-cause analysis; vulnerability exploitability (VEX) statements issued to buyers |

**Generative AI (SP 800-218A)**

- **Untrusted input:** source text from GDELT and other feeds is treated as hostile, and
  prompt-injection tests run in CI for both event enrichment and explain.
- **Bounded and grounded:** explain answers only from retrieved read-model records, and
  every claim must cite provenance.
- **Records:** model and provider versions are recorded on each enrichment and answer,
  and prompts and outputs are audited.
- **Release gate:** an evaluation set of questions and expected citations runs before
  each release.

### 14.5 Evidence pack

Maintained in the repository and versioned with releases:

- Context and data-flow diagrams, threat model and risk register
- SbD self-assessment and principle mapping (14.3)
- SSDF practice evidence (14.4) with CI reports retained as machine-readable artifacts
- SBOMs and signed provenance per release
- IT health check reports and remediation records
- DPIA, ethics statement and licence register
- Incident response plan, change records and residual risk statement

### 14.6 Supplier readiness (Innovox)

- **Cyber Essentials Plus**, which government contracts commonly require.
- **Personnel security:** baseline screening for staff with access to buyer
  deployments.
- **Information security management:** a documented system with an ISO/IEC 27001
  roadmap, if buyers require it.
- **Defence obligations:** ISN 2023/09 supplier obligations reviewed before any
  Defence contract.

## 15. Delivery phases

| Phase | Outcome | Main reuse |
|---|---|---|
| 0. Scaffolding and assurance baseline | `migration-*` components building in CI with SSDF gates; threat model, risk register and SbD self-assessment v0; Keycloak realm; Caddy; hosted compose | EIS images and security gates |
| 1. UK spine and statistics | Arrivals, preventions, returns and illegal-entry tables landing; statistics panels; UK node on the map | Overlay pattern, projector |
| 2. Release layer | Crossing conditions, marine data and tides; 14-day outlook; release alerts; email | New model service, outbox |
| 3. One corridor | Frontex and UNHCR for one route family; pressure layer; heat map | Generic flow pack |
| 4. Rewind and scorecard | Vintage-aware BFF, two or three curated surges, live scoring | Vintages in EIS |
| 5. Explain, brief, scenarios, readiness | The full demonstration set | BFF and model |
| 6. Events, deaths and markers | UCDP, GDELT, IOM Missing Migrants, policy markers | Review-finding pattern |
| 7. Buyer pilot | Federated login, assurance pack, G-Cloud listing | Keycloak, proxy |

Phase 2 comes before the upstream corridor because the release layer gives a working
operational outlook quickly, and it can be scored within weeks.

## 16. External review responses

| Recommendation | Response |
|---|---|
| Separate transit lag from publication lag | **Adopted** (7.2) |
| Weather as an exogenous covariate at the UK node | **Adopted** and extended to tides, backlog and calendar (7.3, 7.4) |
| Log-odds pooling instead of naive chaining | **Adopted** (7.3) |
| Two modes for UK nationality | **Adopted**; Chow-Lin disaggregation deferred (7.3) |
| Serve PMTiles from the proxy; Caddy over Traefik | **Adopted** (8, 12) |
| Debounced event-driven model runs plus a scheduled fallback | **Adopted** (7.9) |
| Fixed phase 2 nationality list | **Not adopted**: set with the first buyer, because the top nationalities shift year to year |
| 4–12 week horizon | **Adopted** for pressure alerts; the release layer covers 0–14 days |

## 17. Open decisions

1. First buyer and the decision they need supported, which sets horizons, alert
   thresholds and wording.
2. Corridor and nationalities for phase 3.
3. Commercial licence position per source (see the [source register](source-register.md)).
4. Cloud provider and email provider (UK region).
5. Probabilistic library for the release layer.
6. Which historical surges form the curated rewind set.
7. Who acts as risk owner and accreditor in the first buyer's Secure by Design
   process, and which assurance artefacts they require at each stage.
8. Who performs the independent IT health check before the buyer pilot.
