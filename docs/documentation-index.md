# Documentation index

| Path | Contents | Maintained by |
|---|---|---|
| `architecture/solution-architecture.md` | Solution architecture: purpose, context, logical, data, interface, security and physical views, NFRs, roadmap, decisions, risks | Hand-written; update when the design changes |
| `conformance/self-assessment.md` | Status of this build against the Innovox Assured scheme, by control and by standard | **Generated**: `go run ./cmd/innovox-conformance -out docs/conformance/self-assessment.md` |
| `conformance/how-conformance-works.md` | How the conformance report and adversarial corpus work | Hand-written |
| `programme/INDEX.md` | Index of the wider programme pack: design review, Assurance Platform, delivery model, discovery rehearsals, migration early-warning | Hand-written |
| `programme/marketplace/design-review.md` | Independent design review of this repository (export; the live document is the master) | Export: do not edit here |
| `programme/marketplace/review-response.md` | Status of every review finding | Hand-written; update whenever a finding's status changes |
| `superpowers/specs/` | Design specs agreed before each build increment, dated | Hand-written; one per increment |
| `superpowers/plans/` | Step-by-step implementation plans derived from a spec, dated | Hand-written; one per increment |

`superpowers/` is named after the planning workflow that produces specs and plans. A spec records what was decided; its plan records how it was built.
