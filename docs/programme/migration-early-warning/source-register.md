# Migration early-warning — source register 0.2.0 (draft)

Candidate public sources for the acquisition overlay. Each needs an interface control
document (ICD) before its connector is qualified, using the EIS interface control document
structure. Phases follow section 15 of the
[product design](product-design.md).

**Licence column:** the terms are expected, not yet verified. Because this is a
commercial product, every row must be checked against the publisher's current terms
for **commercial use**, redistribution and derived outputs before it ships. A source
without a commercial-use basis either stays out of the product or is licensed.

| Source | Measure | Granularity | Native form | Commercial licence (expected, to verify) | Phase |
|---|---|---|---|---|---|
| Home Office: small boat arrivals (GOV.UK) | UK small-boat arrivals, boats | Daily since 2018; provisional | ODS/CSV on a publication page | Open Government Licence v3.0 | 1 |
| Home Office: French preventions (GOV.UK) | Crossing attempts prevented | Weekly; earlier coverage sparser | ODS/CSV | Open Government Licence v3.0 | 1 |
| Home Office: illegal entry routes tables | Arrivals by method, nationality, age band and sex | Quarterly; final | xlsx | Open Government Licence v3.0 | 1 |
| Home Office: returns statistics | Returns, including of small-boat arrivals | Quarterly and periodic factsheets; series structure has changed over time | xlsx and HTML | Open Government Licence v3.0 | 1 |
| Home Office: crossing conditions tables | Met Office crossing-condition assessments | Daily | xlsx/ODS | Open Government Licence v3.0 | 2 |
| Copernicus Climate Change Service: ERA5 reanalysis | Historical wind and sea state | Hourly, gridded | NetCDF/GRIB subsets via API | Copernicus licence, attribution | 2 |
| Copernicus Marine Service | Wave and current analysis and forecasts | Hourly to daily, gridded | NetCDF via API | Copernicus Marine licence, attribution | 2 |
| Tidal predictions for Channel ports (UKHO, SHOM or a harmonic model) | Tide times, ranges, spring/neap phase | Event-level | API, files or computed | Varies by provider; to verify | 2 |
| Public holidays and religious calendars | Calendar flags | Daily | Curated files | Public facts, curated in-house | 2 |
| Frontex: monthly detections of irregular border crossings | Detections by route and nationality | Monthly | xlsx | Reuse with attribution | 3 |
| UNHCR Operational Data Portal | Sea and land arrivals (Italy, Greece, Spain and others) | Weekly or monthly | REST JSON | CC BY 4.0 or portal terms | 3 |
| Eurostat `migr_asyappctzm` | First-time asylum applications by citizenship | Monthly | SDMX/JSON statistics API | Reuse with acknowledgement | 3 |
| IOM Displacement Tracking Matrix | Displacement and flow monitoring | Varies by dataset | Files and API | Dataset-specific | 3 |
| EUAA | Asylum trends | Monthly | Reports and data files | To verify | 3 |
| Italy Ministry of the Interior: statistical dashboard | Sea arrivals by nationality | Daily | PDF | To verify | 3 |
| French Channel crossing data | Attempts and interceptions | Varies | To identify | To verify | 3 |
| IOM Missing Migrants Project | Deaths and missing by route; sex and child/adult status where known | Incident-level, shown aggregated | CSV | Dataset-specific | 6 |
| UCDP Georeferenced Event Dataset | Organised violence events | Event-level; monthly candidate releases | REST JSON | CC BY 4.0 | 6 |
| GDELT 2.0 Events | News-derived events | 15-minute exports | Zipped CSV | Open use with citation | 6 |
| Policy markers | Returns arrangements, enforcement funding, asylum support rules, bilateral agreements | Event-level | Curated and reviewed file citing official sources | In-house, citing public sources | 6 |
| IDMC | Internal displacement | Annual and event-level | API | To verify | Later |
| FEWS NET | Food insecurity outlooks | Monthly | Files and API | To verify | Later |
| ViEWS | Conflict forecasts | Monthly | API | To verify | Later |
| ACLED | Conflict events | Event-level | API with registration | Commercial use requires a paid licence; excluded unless licensed | Optional |

## Comparability notes

- Detections count crossings, not people. One person can be detected more than once.
- Asylum applications lag arrival and occur after movement inside the Schengen area.
- Sea arrivals and small-boat arrivals are counted at landing and are not detections.
- Attempts are derived as arrivals plus preventions, only for periods where both are
  published, and labelled as derived.
- UK daily figures are provisional; the quarterly release supersedes them. Both are
  kept as separate vintages.
- Death and missing records are incomplete by nature. Unknown sex, age and identity
  are shown as unknown, never imputed.
- Nationality, age and sex groupings differ by source and are preserved as published.
- Forecast covariates keep their issue time. Observed values never replace a forecast
  in the outlook's history.

## Per-source ICD checklist

1. Publisher, URL pattern and how the current release is discovered.
2. Native format, encoding and layout version. Changes in layout fail extraction
   visibly and are never silently remapped.
3. Revision behaviour and release calendar.
4. Pacing and polite-use limits.
5. Licence, commercial-use basis, required attribution and redistribution terms.
6. Mapping to nodes, measures, nationalities, age and sex bands, and periods.
7. Known gaps and comparability caveats.
8. Security notes: transport, integrity checks and how untrusted content is handled
   (Secure by Design principle 2, SSDF PW).
