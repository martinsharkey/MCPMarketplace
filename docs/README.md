# Documentation index

| Path | Contents | Maintained by |
|---|---|---|
| `architecture/solution-architecture.md` | Solution architecture: purpose, context, logical, data, interface, security and physical views, NFRs, roadmap, decisions, risks | Hand-written; update when the design changes |
| `conformance/self-assessment.md` | Status of this build against the Innovox Assured scheme, by control and by standard | **Generated**: `go run ./cmd/innovox-conformance -out docs/conformance/self-assessment.md` |
| `conformance/README.md` | How the conformance report and adversarial corpus work | Hand-written |
| `superpowers/specs/` | Design specs agreed before each build increment, dated | Hand-written; one per increment |
| `superpowers/plans/` | Step-by-step implementation plans derived from a spec, dated | Hand-written; one per increment |

`superpowers/` is named after the planning workflow that produces specs and plans. A spec records what was decided; its plan records how it was built.
