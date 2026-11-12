# Plan 1 — Core Trust Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the end-to-end trust chain. Innovox signs a certificate binding an MCP server's image digest and tool-manifest hash, publishes it in a signed feed, and a customer node syncs and verifies the feed. The node's gateway then lets certified MCP calls through and blocks a tool change, revocation, expiry or stale feed.

**Architecture:** Shared Go libraries (`pkg/`) carry canonical hashing, DSSE attestations, the key hierarchy and the feed format. Thin services (`assurance/signer`, `assurance/publisher`, `node/syncer`, `node/gateway`, `node/audit`) compose them. File-backed stores stand in for PostgreSQL in this plan.

**Tech Stack:** Go 1.24, standard library only (no network for module downloads in this environment).

**Spec:** `docs/superpowers/specs/2026-11-11-innovox-assurance-marketplace-design.md`

**Sequence:** This is Plan 1 of 5. A2A Agent Card verification uses `manifest.HashAgentCard` (Task 1); enforcing it in the gateway is in Plan 5. Plan 2 covers the certification pipeline orchestrator (S1–S9). Plan 3 covers the factory (OpenAPI create and repo import). Plan 4 covers the portals (catalogue and assessor). Plan 5 covers PostgreSQL, OPA, OIDC, Sigstore, Helm and the kind end-to-end run.

## Global Constraints

- Module path: `innovox.uk/assured`. Go 1.24, stdlib only.
- Predicate type: `https://innovox.uk/attestation/certificate/v1`. DSSE payloadType: `application/vnd.in-toto+json`. Statement `_type`: `https://in-toto.io/Statement/v1`.
- Tiers: `T1` (auto-signed, 12-month expiry) and `T2` (assessor required, 12-month expiry). `T3` is rejected in the pilot.
- Remote endpoints: max tier `T1`; subject name `remote:<url>`, subject digest `sha256:<manifest hash>`.
- Offline grace period default: 7 days. After it, the gateway denies everything.
- Fail closed: any verification, lookup or parse error in sync or the gateway means deny.
- Claim text constant (`pkg/attest.ClaimText`): "Independently tested by Innovox against the Innovox Assured scheme %s, with signed evidence mapped to MOD and NCSC guidance."
- Signing: Ed25519. An offline root key signs delegations for online signing keys.
- Every commit ends with the trailers `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` and `Claude-Session: https://claude.ai/code/session_01EhNkxGaAQTQgjxucoeZ2NP`.

## Spec deviations (to be reversed in Plan 5)

- Sigstore/cosign is replaced by self-implemented DSSE with Ed25519. The envelope format is the same, so keyless signing can be swapped in later.
- OPA is replaced by a Go `Policy` interface with a grant-table implementation.
- The official MCP Go SDK is replaced by a minimal JSON-RPC client in `pkg/mcpx` covering only `tools/list` and `tools/call`.

## Review Focus

1. **Same tools, different JSON formatting or tool order** must produce the identical hash. Pinned in Task 1 (`TestHashIgnoresOrderAndWhitespace`).
2. **Invisible Unicode in a description** (a zero-width space or a homoglyph) must change the hash. Canonicalisation must never normalise Unicode. Pinned in Task 1 (`TestHashDetectsZeroWidthChar`).
3. **Paginated `tools/list` (`nextCursor` present)** must be denied by the gateway. It must not hash a partial set. Pinned in Task 8 (`TestGatewayDeniesPaginatedToolsList`).
4. **A certificate that expires during a session** must cause the next call to be denied. Pinned in Task 8 (`TestGatewayDeniesAfterExpiry`).
5. **A feed older than the grace period** must cause every call to be denied. Pinned in Task 6 (`TestStateStaleAfterGrace`) and Task 8 (`TestGatewayDeniesStaleFeed`).

---

### Task 1: Module scaffold and `pkg/manifest`

**Files:** Create `go.mod`, `pkg/manifest/canonical.go`, `pkg/manifest/manifest.go`, `pkg/manifest/manifest_test.go`

**Interfaces:**
- Produces:
  - `type Tool struct { Name string; Description string; InputSchema json.RawMessage; OutputSchema json.RawMessage; Annotations json.RawMessage }` with JSON tags `name, description, inputSchema, outputSchema, annotations` (omitempty on the last three)
  - `func Canonicalize(raw []byte) ([]byte, error)`
  - `func HashTools(tools []Tool) (string, error)`, which returns lowercase hex SHA-256
  - `func HashAgentCard(raw []byte) (string, error)`

- [ ] **Step 1: Write the failing tests**
  - `TestCanonicalizeSortsKeys`: `{"b":1,"a":{"d":2,"c":3}}` becomes `{"a":{"c":3,"d":2},"b":1}`.
  - `TestHashIgnoresOrderAndWhitespace`: two tools in either order, with re-indented schemas, give an equal hash.
  - `TestHashChangesOnDescription`: appending "." to a description gives a different hash.
  - `TestHashDetectsZeroWidthChar`: inserting `​` gives a different hash.
  - `TestHashRejectsDuplicateNames`: returns an error.
  - `TestHashGolden`: the tool `{name:"echo",description:"Echo text",inputSchema:{"type":"object"}}` hashes to the value pinned on first green run, recorded in the test.
- [ ] **Step 2:** `go test ./pkg/manifest/` fails (undefined).
- [ ] **Step 3: Implement.**
  - `Canonicalize` follows RFC 8785. Decode with `json.Decoder.UseNumber`, sort object keys by UTF-16 code units, and emit with no whitespace.
  - Number formatting: integers come out as written, and non-integers use `strconv.FormatFloat(f,'g',-1,64)` with ES6 exponent rules (`e+21` boundary).
  - String escaping: escape only `"`, `\`, and control characters below U+0020 (using `\b \f \n \r \t`, otherwise `\u00XX`). Never escape HTML characters (use `SetEscapeHTML(false)` semantics).
  - `HashTools` sorts by `Name`, rejects duplicates, then marshals `{"tools":[...]}` and canonicalises it.
- [ ] **Step 4:** Tests pass. Pin the golden value.
- [ ] **Step 5:** Commit with message `feat(manifest): RFC 8785 canonical tool-manifest hashing`.

### Task 2: `pkg/attest`, DSSE envelope and certificate predicate

**Files:** Create `pkg/attest/dsse.go`, `pkg/attest/certificate.go`, `pkg/attest/attest_test.go`

**Interfaces:**
- Produces:
  - `type Envelope struct { PayloadType string; Payload []byte; Signatures []Signature }` with JSON tags `payloadType, payload, signatures`; `Payload` is base64-encoded by `encoding/json`
  - `type Signature struct { KeyID string; Sig []byte }` with JSON tags `keyid, sig`
  - `func PAE(payloadType string, payload []byte) []byte`, the DSSE v1 pre-authentication encoding
  - `func SignEnvelope(payloadType string, payload []byte, keyID string, priv ed25519.PrivateKey) Envelope`
  - `func VerifyEnvelope(env Envelope, keys map[string]ed25519.PublicKey) (keyID string, err error)`
  - `type Subject struct { Name string; Digest map[string]string }`
  - `type Statement struct { Type string; Subject []Subject; PredicateType string; Predicate Certificate }`
  - `type Certificate struct { SchemeVersion, Tier string; Service ServiceRef; ManifestSHA256 string; Evidence []EvidenceRef; IssuedAt, ExpiresAt time.Time; Assessor, ClaimsMappingRef string }`
  - `type ServiceRef struct { Name, Version, Kind string }` (Kind is `mcp` or `a2a`)
  - `type EvidenceRef struct { Stage, Result, SHA256 string }`
  - `func (c Certificate) Claim() string`, which formats `ClaimText` with `SchemeVersion`
  - `func OpenCertificate(env Envelope, keys map[string]ed25519.PublicKey) (Statement, error)`. It verifies the signature, checks the payloadType, Statement type and predicateType, and decodes.

- [ ] **Step 1: Write the failing tests**
  - `TestPAEVector` checks the DSSE spec vector: `PAE("http://example.com/HelloWorld", "hello world")` equals `"DSSEv1 29 http://example.com/HelloWorld 11 hello world"`.
  - `TestSignVerifyRoundTrip`.
  - `TestVerifyRejectsTamperedPayload`.
  - `TestVerifyRejectsUnknownKey`.
  - `TestOpenCertificateRejectsWrongPredicateType`.
  - `TestClaimText` checks that `Claim()` with `"v0.1"` contains `"Innovox Assured scheme v0.1"` and does not contain `"certified"`.
- [ ] **Step 2:** Run the tests and confirm they fail.
- [ ] **Step 3:** Implement with `crypto/ed25519`. `VerifyEnvelope` succeeds if any signature verifies under a known key ID.
- [ ] **Step 4:** Confirm the tests pass.
- [ ] **Step 5:** Commit with message `feat(attest): DSSE envelopes and Innovox certificate predicate`.

### Task 3: `pkg/trust`, root key and delegated signing keys

**Files:** Create `pkg/trust/trust.go`, `pkg/trust/trust_test.go`

**Interfaces:**
- Consumes: `attest.SignEnvelope`, `attest.VerifyEnvelope`
- Produces:
  - `func KeyID(pub ed25519.PublicKey) string`, which returns the first 16 hex characters of the SHA-256 of the public key
  - `type Delegation struct { KeyID string; PublicKey []byte; Role string; NotBefore, NotAfter time.Time }` with JSON tags `keyid, publicKey, role, notBefore, notAfter`; Role is `"signer"`
  - `func Delegate(rootPriv ed25519.PrivateKey, signerPub ed25519.PublicKey, notAfter time.Time) attest.Envelope`, using payloadType `application/vnd.innovox.delegation+json`
  - `type Keyring struct { /* root + verified delegations */ }`
  - `func NewKeyring(rootPub ed25519.PublicKey) *Keyring`
  - `func (k *Keyring) AddDelegation(env attest.Envelope, now time.Time) error`
  - `func (k *Keyring) SignerKeys(now time.Time) map[string]ed25519.PublicKey`, which returns delegated keys valid at `now`. The root is excluded, because the root never signs certificates.

- [ ] **Step 1: Write the failing tests**
  - `TestDelegationAccepted`.
  - `TestDelegationNotSignedByRootRejected`.
  - `TestExpiredDelegationExcluded`.
  - `TestRootNotInSignerKeys`.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(trust): offline root and delegated signer keys`.

### Task 4: `assurance/signer`, issuing certificates

**Files:** Create `assurance/signer/signer.go`, `assurance/signer/signer_test.go`, `cmd/innovox-sign/main.go`

**Interfaces:**
- Consumes: `manifest.HashTools`, `attest.*`, `trust.KeyID`
- Produces:
  - `type Request struct { Service attest.ServiceRef; ImageDigest string; RemoteURL string; Tools []manifest.Tool; Tier string; Assessor string; SchemeVersion string; Evidence []attest.EvidenceRef }`
  - `type Signer struct { Priv ed25519.PrivateKey; Now func() time.Time }`
  - `func (s Signer) Issue(r Request) (attest.Envelope, error)`
  - The CLI reads the request JSON and key file and writes the envelope JSON: `innovox-sign -key signer.key -in req.json -out cert.json`.

- [ ] **Step 1: Write the failing tests**
  - `TestIssueT1Auto`: no assessor is required, and `ExpiresAt` equals `IssuedAt.AddDate(1,0,0)`.
  - `TestIssueT2RequiresAssessor`: an empty assessor returns an error.
  - `TestIssueRejectsT3`.
  - `TestRemoteCappedAtT1`: `RemoteURL` set with tier T2 returns an error.
  - `TestRemoteSubject`: the subject name is `remote:<url>` and the digest is `sha256:<hash>`.
  - `TestImageSubject`: the subject digest is parsed from `sha256:<hex>`.
  - `TestRejectsBothOrNeitherImageAndRemote`.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(signer): issue tiered Innovox certificates`.

### Task 5: `pkg/feed` and `assurance/publisher`

**Files:** Create `pkg/feed/feed.go`, `pkg/feed/feed_test.go`, `assurance/publisher/publisher.go`, `assurance/publisher/publisher_test.go`, `cmd/innovox-publish/main.go`

**Interfaces:**
- Produces:
  - `type Revocation struct { CertDigest, Reason string; RevokedAt time.Time }`, where `CertDigest` is the hex SHA-256 of the certificate envelope's `Payload`
  - `type Index struct { Seq uint64; CreatedAt time.Time; Delegations []attest.Envelope; Certificates []attest.Envelope; Revocations []Revocation }`
  - `func CertDigest(env attest.Envelope) string`
  - `func Sign(idx Index, keyID string, priv ed25519.PrivateKey) (attest.Envelope, error)`, using payloadType `application/vnd.innovox.feed+json`
  - `func Open(env attest.Envelope, kr *trust.Keyring, now time.Time) (Index, error)`. It adds the delegations to the keyring, then verifies the index signature against the signer keys.
  - `func WriteBundle(path string, env attest.Envelope) error` and `func ReadBundle(path string) (attest.Envelope, error)`. The bundle is a single JSON file, used for air-gap transfer.
  - `type Publisher struct { Dir string; KeyID string; Priv ed25519.PrivateKey; Now func() time.Time }`
  - `func (p Publisher) Publish() (attest.Envelope, error)`. It reads `Dir/certs/*.json`, `Dir/revocations.json`, `Dir/delegations/*.json` and `Dir/seq`, increments `seq`, then signs and writes `Dir/feed.json`.
  - CLI: `innovox-publish -dir store -key signer.key`

- [ ] **Step 1: Write the failing tests**
  - `TestFeedRoundTrip`.
  - `TestOpenRejectsFeedSignedByRoot`.
  - `TestBundleRoundTrip`.
  - `TestPublishIncrementsSeq`: two publishes give seqs 1 then 2.
  - `TestCertDigestStable`.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(feed): signed feed index, bundles and publisher`.

### Task 6: `node/syncer` (package `syncer`, avoids clashing with stdlib `sync`), verifying and accepting the feed

**Files:** Create `node/syncer/sync.go`, `node/syncer/state.go`, `node/syncer/sync_test.go`

**Interfaces:**
- Consumes: `feed.Open`, `attest.OpenCertificate`, `trust.Keyring`
- Produces:
  - `type Listing struct { Cert attest.Statement; CertDigest string; Revoked bool; RevokeReason string }`
  - `type State struct { Seq uint64; FeedTime time.Time; SyncedAt time.Time; Listings map[string]*Listing }`, keyed by `Service.Name + "@" + Service.Version`
  - `func (s *State) Stale(now time.Time, grace time.Duration) bool`, which is true when `now - SyncedAt > grace`
  - `func Apply(st *State, env attest.Envelope, rootPub ed25519.PublicKey, now time.Time) error`. It rejects `idx.Seq <= st.Seq` with `ErrReplay`. It skips (does not list) any certificate that fails to open or is expired. It marks revocations. It is atomic: on error, `st` is unchanged.
  - `func Load(path string) (*State, error)` and `func (s *State) Save(path string) error`

- [ ] **Step 1: Write the failing tests**
  - `TestApplyListsValidCert`.
  - `TestApplyRejectsReplay`: applying the same feed twice returns `ErrReplay`.
  - `TestApplySkipsExpiredCert`.
  - `TestApplyMarksRevoked`.
  - `TestApplyRejectsForgedFeed`: a feed signed by an undelegated key returns an error and leaves state unchanged.
  - `TestStateStaleAfterGrace`: 8 days later is stale with a 7-day grace period, and 6 days later is not.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(sync): verify and apply Innovox feeds`.

### Task 7: `node/audit`, hash-chained log

**Files:** Create `node/audit/audit.go`, `node/audit/audit_test.go`

**Interfaces:**
- Produces:
  - `type Event struct { Seq uint64; PrevHash, Hash string; TS time.Time; Actor, Action, Decision, Detail string }`
  - `type Log struct { /* mutex, file */ }`
  - `func Open(path string) (*Log, error)`, which resumes the chain from the file
  - `func (l *Log) Append(actor, action, decision, detail string) (Event, error)`
  - `func Verify(path string) error`, which detects a modified, deleted or reordered line

  The hash is the hex SHA-256 of the canonical JSON of the event with `Hash` empty. The genesis `PrevHash` is 64 `"0"` characters. Each event is one JSON line.

- [ ] **Step 1: Write the failing tests**
  - `TestAppendChains`.
  - `TestVerifyDetectsEdit`.
  - `TestVerifyDetectsDeletion`.
  - `TestReopenContinuesChain`.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(audit): hash-chained append-only audit log`.

### Task 8: `pkg/mcpx` and `node/gateway`, enforcement

**Files:** Create `pkg/mcpx/jsonrpc.go`, `pkg/mcpx/mcpx_test.go`, `node/gateway/gateway.go`, `node/gateway/policy.go`, `node/gateway/gateway_test.go`, `cmd/innovox-gateway/main.go`

**Interfaces:**
- Consumes: `syncer.State`, `manifest.HashTools`, `audit.Log`
- Produces:
  - `type Request struct { JSONRPC string; ID json.RawMessage; Method string; Params json.RawMessage }` and `type Response struct { JSONRPC string; ID json.RawMessage; Result json.RawMessage; Error *Error }`
  - `type Error struct{ Code int; Message string }`
  - `type ToolsListResult struct { Tools []manifest.Tool; NextCursor string }` with JSON tags `tools, nextCursor`
  - `type CallParams struct { Name string; Arguments json.RawMessage }`
  - `type Authenticator interface { Authenticate(r *http.Request) (principal string, err error) }`, plus `StaticTokens map[string]string` (token to principal) for the pilot
  - `type Policy interface { Allow(principal, service, tool string) bool }`, plus `Grants map[string]map[string][]string` (principal to service to tools, where `"*"` means all)
  - `type Upstream struct { Service string; URL string; ImageDigest string }`, where `ImageDigest` is the running digest reported by the runtime
  - `type Gateway struct { State func() *syncer.State; Upstreams map[string]Upstream; Auth Authenticator; Policy Policy; Audit *audit.Log; Grace time.Duration; Now func() time.Time; Client *http.Client }`
  - `func (g *Gateway) ServeHTTP(w, r)` for route `POST /mcp/{service}`

  **Decision order (deny returns HTTP 403 with a JSON-RPC error `-32001`, and is audited):**
  1. Authenticate.
  2. The state must not be stale.
  3. The listing must exist, must not be revoked, and must not be expired at `Now()`.
  4. For an image subject, the subject digest must equal `Upstream.ImageDigest`.
  5. Method `tools/list`: forward, then deny if `NextCursor != ""`. Hash the tools and deny on mismatch with the action `rug_pull`. Otherwise return only the tools allowed by policy.
  6. Method `tools/call`: re-fetch `tools/list` upstream and re-verify the hash, then check `Policy.Allow` for the tool name, then forward.
  7. Any other method is forwarded only if it is `initialize`, `ping` or `notifications/*`. Everything else is denied.

- [ ] **Step 1: Write the failing tests** (upstream simulated with `httptest`)
  - `TestGatewayAllowsCertifiedCall`.
  - `TestGatewayBlocksRugPull`: the upstream changes a description after certification, so `tools/list` and `tools/call` are both denied and an audit event with action `rug_pull` is written.
  - `TestGatewayFiltersToolsByGrant`.
  - `TestGatewayDeniesUngrantedTool`.
  - `TestGatewayDeniesRevoked`.
  - `TestGatewayDeniesAfterExpiry`: `Now` is advanced past `ExpiresAt` mid-test.
  - `TestGatewayDeniesStaleFeed`.
  - `TestGatewayDeniesDigestMismatch`.
  - `TestGatewayDeniesPaginatedToolsList`.
  - `TestGatewayDeniesBadToken`.
  - `TestGatewayDeniesUnknownMethod`.
- [ ] **Steps 2–4:** Run (fail), implement, run (pass).
- [ ] **Step 5:** Commit with message `feat(gateway): fail-closed MCP enforcement with manifest pinning`.

### Task 9: Malicious fixture and end-to-end test

**Files:** Create `test/fixtures/evilmcp/server.go`, `test/e2e/e2e_test.go`

**Interfaces:**
- Consumes: everything above.
- Produces: `func NewEvil() (*httptest.Server, func(mutate bool))`, an MCP server whose second tool's description gains a hidden instruction when `mutate(true)` is called.

- [ ] **Step 1: Write the failing test `TestEndToEnd`.** It runs the full flow:
  1. Generate a root and a signer key, then delegate.
  2. Sign a T1 certificate for the evil server's tools.
  3. Publish a feed, write it as a bundle, and read the bundle back.
  4. Apply the feed to a fresh state.
  5. A call through the gateway is allowed.
  6. Call `mutate(true)`; the next call is denied with `rug_pull`.
  7. `mutate(false)`, revoke the certificate, publish seq 2 and apply it; the call is denied.
  8. Run `audit.Verify` and confirm the chain is valid.
- [ ] **Steps 2–4:** Run (fail), fix any gaps, run (pass). Then run `go test ./... && go vet ./...` and expect all packages to pass.
- [ ] **Step 5:** Commit with message `test(e2e): full trust-chain lifecycle against malicious fixture`.
