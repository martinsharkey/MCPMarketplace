SYNTHETIC INTERVIEW. Rehearsal only. Not research evidence. Persona: Supplier security lead, built 10 November 2026.

# Rehearsal: supplier security lead

Interviewer: Innovox (rehearsal) · Date: 10 November 2026 · Topic: default script and platform concept

**Walk me through the last time you had to produce security evidence for a gate.**
I built the pack by hand. Scanner outputs, our threat model, policies, configuration screenshots, a mapping spreadsheet against the customer's requirements. At the same time I was keeping our Def Stan 05-138 controls current for the Cyber Risk Profile on the contract, which is a separate piece of work about the company, not the system. People often confuse the two.

**What took longest? What was redone?**
The mapping. Every customer, and sometimes every assessor, wants evidence cut a different way. And keeping it current: a scan from six weeks ago is not evidence of today.

**What would you need to trust evidence produced by a tool?**
I'd need to see exactly what it checked, with which tool and version, and what it didn't check. The false-positive rate matters more to me than the number of checks. And I'd need assessors to recognise the output. If they don't, I've just made another format to translate.

**Who decides whether a tool like this is used, and how is it bought?**
I'm the technical approver. If I say it gives false comfort, it doesn't get in. Budget sits elsewhere, but nobody buys security tooling over my objection.

**What would make you stop using it?**
The first time it marked something as compliant that wasn't. My name goes on the evidence, not the tool's. One false green and I'd stop trusting all of it.

**Concept: the platform generates evidence continuously from the build and holds gates for a named approver. Where would this fit or fail for you?**
The named-approver part is right, and so is showing what wasn't assessed. That's the bit most tools hide. Where it could fail: if it tries to cover Def Stan 05-138 organisational controls by scanning code. Most of those are about how we run the company. I'd want it to track those as evidence I supply, not pretend to test them.

## Objections raised
- Automated compliance gives false comfort
- One false "compliant" destroys trust in everything
- Organisational controls can't be tested by scanning code
- Another output format assessors don't recognise

## Assumptions this rehearsal suggests
- Security leads value honest coverage and a low false-positive rate above breadth of checks.
- Mapping evidence to each customer's format is the largest manual cost.
- Security leads hold an effective veto over adoption.
- Showing "not assessed" explicitly increases trust rather than reducing perceived value.

## Questions to ask real people
- What share of your evidence effort is mapping to the customer's format rather than producing evidence?
- What would a tool have to show you to trust a "compliant" result?
- How do you currently keep Def Stan 05-138 evidence current, and how is it checked?

## Basis
- **Cited role sources:** ownership of supplier evidence; Def Stan 05-138 at the assigned Cyber Risk Profile; Cyber Essentials Plus; Defence Cyber Certification.
- **Inferred:** the hand-built pack, mapping as the largest cost, the veto, the false-green reaction.
