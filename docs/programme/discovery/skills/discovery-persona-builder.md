---
name: "discovery-persona-builder"
description: "Build or refresh a synthetic discovery stakeholder persona from public role sources (MOD SbD roles, DDaT, Government Security Profession, job adverts), cited and labelled synthetic."
---

# Discovery persona builder

Build a composite persona of a stakeholder role so a team can rehearse discovery before real interviews. A persona is preparation, never evidence.

## Rules

1. **Composite role, never a person.** Never name, imitate or profile a real individual. Do not search for named people. Use invented first names only if the user asks, and label them invented.
2. **Public sources only.** Use official role definitions first, then job adverts, then professional frameworks. Record every source with its URL and the date you read it.
3. **Label it synthetic.** Every persona file starts with: `SYNTHETIC PERSONA. Composite of public role sources. Not research evidence.`
4. **Say what is inferred.** Mark each statement as `source` (stated in a cited source) or `inferred` (your reasonable extrapolation). Keep inferred statements few and plausible.
5. **Refresh.** A persona older than six months, or built before a relevant policy change, is refreshed before use.

## Research order

1. Official role definitions:
   - MOD Secure by Design roles: https://www.digital.mod.uk/secure-by-design/who-is-responsible-for-sbd
   - Government Digital and Data capability framework: https://ddat-capability-framework.service.gov.uk
   - Government Security Profession career framework: https://www.security.gov.uk/government-security-profession-career-framework
2. Policy the role answers to: JSP 440 Leaflet 5C, Def Stan 05-138, ISNs, NCSC CAF, Service Standard, as relevant.
3. Current job adverts for the role (Civil Service Jobs, MOD and defence supplier adverts): responsibilities, clearance, tools, reporting line.
4. Professional guidance or blogs from the relevant government profession.

## Persona template

```markdown
SYNTHETIC PERSONA. Composite of public role sources. Not research evidence.
Built: <date> · Refresh by: <date + 6 months>

# <Role name>

## Role
<One paragraph: where they sit, who they report to, what they deliver.>

## Accountable for
- <item> (source | inferred)

## Answers to
- <standards, policies, people>

## Day-to-day pressures
- <item> (source | inferred)

## How they judge a new tool or service
- <criteria>

## Likely objections
- <objection and the concern behind it>

## Role in buying
<User, influencer, approver, budget holder or blocker; how purchases happen for them.>

## Questions they would ask us
- <question>

## Sources
- <title> · <URL> · read <date>
```

## Steps

1. Confirm the role and the context (for example: supplier delivering to MOD, Home Office programme).
2. Research in the order above; stop when each template section has at least one sourced line.
3. Write the persona with the template. Keep it under 600 words.
4. List the three assumptions about this role that real interviews most need to test.
5. Save it where the user asks (for example `discovery/personas/<role-slug>.md`).