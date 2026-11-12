# Plan 1b: changes from the design review

**Goal:** Carry out recommended changes 1 to 5 of the [design review](../../programme/marketplace/design-review.md) before Plan 2 builds on the trust chain and gateway.

**Approval:** The owner approved changes 1 to 5 on 12 November 2026. Change 6 (independence model, Factory wording, product name) is the owner's decision and is not in this plan.

**Status of each finding:** see [review-response.md](../../programme/marketplace/review-response.md).

## Done

| Review item | What was built | Control |
|---|---|---|
| C2 | Certificates are checked against the signing key's delegation at the time of issue, so they survive key rotation. Certificates issued outside the key's lifetime, or in the future, are refused. | IAS-TRUST-07 |
| C4 | Audit records carry the tool, an arguments digest, certificate digest, manifest hash, outcome and duration. | IAS-AUD-05 |
| C5 | The node rebuilds its listings from the signed feed against the pinned root on every load. | IAS-TRUST-08 |
| C1, in part | The root can revoke signer keys. Key and certificate revocations are kept for good. | IAS-TRUST-03, IAS-TRUST-06 |
| C8, in part | Signed audit checkpoints and an export sink. | IAS-AUD-03 |

## Part A: re-baseline the gateway on MCP 2026-07-28 (C3, T5)

**Blocker:** the MCP specification site could not be opened from the session that wrote this plan, and package downloads were unavailable. Part A must be done where the specification and the official Go SDK can be fetched. No clause of the new revision in `scheme/sources.json` is verified yet.

**Global constraints**

- Open the 2026-07-28 changelog and specification first. Record each clause used in `scheme/sources.json` with the date it was verified.
- Keep serving servers on the earlier revision.
- Every method and result type gets an explicit allow or deny rule. Anything not listed is denied.
- Adding the official SDK is a dependency: record it as an architecture decision (solution architecture §15).

**Tasks**

1. **Method policy table.** Create `node/gateway/methods.go` with one table: method, revisions it exists in, rule (forward, verify-then-forward, deny), and reason. Replace the `switch` in `handle` with a lookup. Test: every method in both revisions has a row; an unknown method is denied.
2. **Discovery.** Handle `server/discover`. Pin its result the same way `initialize` instructions are stripped today: only certified fields reach the agent. Test: a discover result with extra instructions is stripped.
3. **Tool lists.** Rebuilt `tools/list` results carry the fields the revision requires. Decide and test how the certified manifest hash treats each new field.
4. **Result types.** For each result type, including a request for user input, add a rule. Default for input requests: deny, with an audited reason, until a policy for them is designed. Add an adversarial case: a tool result that asks the user to visit a URL.
5. **Server identity.** Check the server identity carried on results against the certificate's subject. Mismatch denies.
6. **Schemas.** The certification pipeline fails a tool whose schema refers to content outside the manifest. The gateway refuses such a schema at run time. Test both.
7. **Headers, extensions, tasks.** Deny by default. Add a row per item to the method policy table.
8. **SDK swap.** Replace `pkg/mcpx` with the official Go SDK behind the same narrow interface. Keep the no-redirect rule and the response size limit. Run the adversarial corpus unchanged.

**Review focus**

1. A server that answers in the old revision to a new-revision client, and the reverse.
2. A result type the table does not list.
3. A tool list that changes between `server/discover` and `tools/call`.

## Part B: rebuild feed trust on the TUF model (C1, T1)

**Global constraints**

- Verify The Update Framework specification text before relying on it. The review described it from general knowledge.
- Existing feeds and certificates must still verify during migration, or the migration must be a single documented cut-over while there is one node.
- No new dependency is required: the roles can be built on `pkg/attest` and `pkg/trust`.

**Tasks**

1. **Roles.** Extend `trust.Delegation.Role` from `signer` to `certificate`, `snapshot` and `timestamp`. `Keyring.KeysFor(role, at)` returns only that role's keys. Tests: a certificate signed by a snapshot key is refused, and a feed signed by a certificate key is refused.
2. **Separate feed and certificate keys.** `syncer.derive` verifies certificates with `certificate` keys. `feed.OpenWithSigner` verifies the index with `snapshot` keys. Update `cmd/innovox-root delegate` with `-role`.
3. **Timestamp.** A small, frequently re-signed statement naming the current snapshot's sequence number and hash. The node's freshness check uses the timestamp, so the snapshot key can stay offline longer. Test: an old snapshot with a fresh timestamp for a different hash is refused.
4. **Thresholds.** `attest.VerifyThreshold(env, keys, n)` requires `n` distinct valid signatures. Root actions (delegation, key revocation, root rotation) require the root threshold. This also closes C12. Test: one of two required signatures is refused.
5. **Root rotation.** A root-rotation statement signed by the threshold of the old root names the new root keys. The node stores the chain and accepts feeds under the new root. Test: a node pinned to root 1 follows a signed chain to root 3; a chain with a gap is refused.
6. **Reinstatement.** A root-signed statement can lift a certificate revocation. Without it, revocations stay (already built).
7. **Self-assessment.** Move IAS-TRUST-09 to met only when tasks 1 to 5 pass.

**Review focus**

1. A key delegated for two roles at once.
2. A threshold met by the same key signing twice.
3. A node that missed several root rotations.

## Part C: transport and runtime digest (C6, C7)

1. **TLS at the gateway.** `cmd/innovox-gateway` takes a certificate and key and refuses to start without them unless an explicit development flag is set.
2. **Mutual TLS to services.** `Upstream` gains a client certificate and a pinned server CA. Test: a service presenting the wrong certificate is refused.
3. **OIDC.** Replace `StaticTokens` with a verifier for the customer's identity provider. This needs a dependency or a careful standard-library JWT verifier: record the decision.
4. **Runtime-read digest.** Blocked on the runtime controller (roadmap Plan 5). Until then IAS-ISO-02 stays partial.

## Order

Part B can start now and needs no network. Part A waits for an environment that can reach the specification and the SDK. Part C tasks 1 and 2 can follow Part B; tasks 3 and 4 depend on later plans.
