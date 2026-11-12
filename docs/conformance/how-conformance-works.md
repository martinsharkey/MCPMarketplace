# Conformance

`self-assessment.md` is generated. Do not edit it by hand.

```bash
go run ./cmd/innovox-conformance -out docs/conformance/self-assessment.md
```

The command does four things:

1. Runs the full test suite.
2. Collects the controls each test proves, through `conformance.Covers(t, "IAS-…")`.
3. Combines the results with the written judgements in `scheme/self-assessment.json`.
4. Renders status by control and coverage by standard.

**Rules the tool enforces:**

- An automated control is `met` only if tagged tests ran and passed.
- A written `met` claim with no test evidence is rejected.
- A document control marked `met` must cite evidence.
- Any failing tagged test marks its control `failing`, and the command exits non-zero.

**Adversarial corpus:** `test/adversarial` runs malicious MCP server behaviours through the gateway. Cases marked `blocked` must be denied. Cases marked `gap` record attacks the gateway does not stop yet. They are not counted as evidence, and they fail on purpose once the gap is closed, so the record stays honest.

This is a self-assessment by the build team. It is not independent.
