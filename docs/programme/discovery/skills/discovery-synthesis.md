---
name: "discovery-synthesis"
description: "Synthesise discovery findings from real interview notes, comparing them with synthetic persona rehearsals to show where assumptions held or failed; real evidence only counts."
---

# Discovery synthesis

Turn real discovery interview notes into findings, and use synthetic persona rehearsals only to show which assumptions the real evidence confirmed, contradicted or left untested.

## Rules

1. **Only real interviews are evidence.** Synthetic transcripts (labelled SYNTHETIC) never support a finding, a user need or a gate decision.
2. **Trace every finding** to the interview notes it came from, by interview ID and participant role. Keep participants anonymous; never record names in findings.
3. **Count, don't generalise.** Say how many participants of which roles said something. A single mention is a signal, not a finding.
4. **Contradictions are the value.** Where real people disagree with the personas, record it prominently and update or retire the persona assumption.
5. **No invented quotes.** Quote only words in the real notes, kept short.

## Inputs

- Real interview notes in `discovery/interviews/` (one file per interview, with ID, date and role)
- Synthetic rehearsals in `discovery/rehearsals/`
- The current problem statement and assumption list

## Steps

1. Read every real interview note. Extract observations: pain points, current process steps, trust criteria, buying route, objections.
2. Group observations into themes. For each theme record: the roles that raised it, the count, and the interview IDs.
3. Write user needs in the Service Manual form: "As a <role>, I need <need> so that <outcome>", each traced to themes.
4. Compare with rehearsals. For each assumption the rehearsals produced, mark it `confirmed`, `contradicted` or `untested`, citing interview IDs.
5. Update the problem statement where evidence changes it, and list what is still untested.
6. Recommend next interviews: roles under-represented, and assumptions still untested.

## Output

```markdown
# Discovery synthesis · <date>
Interviews: <n> real (<roles and counts>) · Rehearsals compared: <n> synthetic

## Themes
| Theme | Roles | Count | Interviews |

## User needs
- As a <role>, I need <need> so that <outcome>. (themes: …)

## Assumptions
| Assumption | Source | Status | Evidence |

## Problem statement changes
## Untested and next interviews
```

The synthesis feeds the discovery report and the discovery gate. The gate criterion that buyers confirm the problem is met only by real interviews.