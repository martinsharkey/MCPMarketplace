SYNTHETIC INTERVIEW. Rehearsal only. Not research evidence. Persona: Supplier delivery lead, built 10 November 2026.

# Rehearsal: supplier delivery lead

Interviewer: Innovox (rehearsal) · Date: 10 November 2026 · Topic: default script and platform concept

**Walk me through the last time you had to produce security evidence for a gate.**
The request came in late, as it usually does. We'd been building for months, and then the customer's security people wanted evidence mapped to Secure by Design before the next stage. My security lead spent weeks pulling scan outputs, screenshots and documents into a pack. My engineers kept getting pulled off sprint work to answer questions about things they'd built months before.

**What took longest? What was redone?**
Finding evidence that matched what they asked for, in the format they wanted. Then a second round of questions, because something was out of date by the time they read it. That second round is what hurts: the work has moved on and the evidence hasn't.

**What would you need to trust evidence produced by a tool?**
Honestly, I'd trust it if my security lead trusts it. I'm not the one who'll be challenged on it. What I'd need is to know it doesn't slow the team down and that the customer will accept it.

**Who decides whether a tool like this is used, and how is it bought?**
Small tools I can sometimes get through our own tooling budget. Anything that touches customer evidence needs our security lead to agree, and possibly the customer too. If the customer told us to use it, that would settle it.

**What would make you stop using it?**
If it became another thing the team has to feed. If it flags hundreds of things and my engineers spend their time arguing with it. Or if the customer's assessors ignored its output anyway.

**Concept: the platform generates evidence continuously from the build and holds gates for a named approver. Where would this fit or fail for you?**
It fits if it means the evidence pack exists before anyone asks for it. It fails if it's my team who has to keep it fed. And I'd want to know who approves inside my team. If the approver is my security lead, fine. If it ends up being me, I'm not qualified to sign security risk.

## Objections raised
- Another tool for the team to feed
- Noise that drags engineers into arguing with findings
- The customer's assessors may not accept the output

## Assumptions this rehearsal suggests
- Delivery leads value evidence that exists before the request more than evidence that is faster to assemble.
- The second round of evidence requests, caused by stale evidence, is a larger cost than the first.
- Delivery leads defer the trust decision to their security lead.
- A customer mandate would decide adoption faster than a supplier's own budget.

## Questions to ask real people
- How many rounds of evidence requests did your last gate take, and what triggered each one?
- Who on your team spends the most time on evidence, and on what?
- Would you adopt a tool your customer had not asked for?

## Basis
- **Cited role sources:** accountability for delivery and contract security terms; evidence requests competing with feature work.
- **Inferred:** late requests, second-round questions, budget routes, the approver concern.
