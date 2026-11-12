package e2e

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"innovox.uk/assured/internal/conformance"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"innovox.uk/assured/assurance/publisher"
	"innovox.uk/assured/assurance/signer"
	"innovox.uk/assured/node/audit"
	"innovox.uk/assured/node/gateway"
	"innovox.uk/assured/node/syncer"
	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
	"innovox.uk/assured/test/fixtures/evilmcp"
)

const svc = "uk.mod.example/stock@1.0.0"

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	b, _ := json.Marshal(v)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEndToEnd(t *testing.T) {
	conformance.Covers(t, "IAS-RT-02", "IAS-TRUST-04", "IAS-TRUST-05", "IAS-AUD-01", "IAS-AUD-02")
	now := time.Now().UTC()
	digest := "sha256:" + strings.Repeat("f", 64)

	// 1. Innovox key hierarchy: offline root delegates an online signer.
	rootPub, rootPriv, _ := ed25519.GenerateKey(rand.Reader)
	sPub, sPriv, _ := ed25519.GenerateKey(rand.Reader)
	store := t.TempDir()
	writeJSON(t, filepath.Join(store, "delegations", "signer.json"), trust.Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)))

	// 2. Certify the server as it behaves today (T1).
	up, mutate, calls := evilmcp.NewEvil()
	defer up.Close()
	cert, err := signer.Signer{Priv: sPriv, Now: time.Now}.Issue(signer.Request{
		Service:     attest.ServiceRef{Name: "uk.mod.example/stock", Version: "1.0.0", Kind: "mcp"},
		ImageDigest: digest, Tools: evilmcp.Tools(), Tier: "T1", SchemeVersion: "v0.1"})
	if err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(store, "certs", "stock.json"), cert)

	// 3. Publish, write the air-gap bundle, carry it across, read it back.
	pub := publisher.Publisher{Dir: store, KeyID: trust.KeyID(sPub), Priv: sPriv, Now: time.Now}
	if _, err := pub.Publish(); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(t.TempDir(), "carried.json")
	env, _ := feed.ReadBundle(filepath.Join(store, "feed.json"))
	if err := feed.WriteBundle(bundle, env); err != nil {
		t.Fatal(err)
	}
	carried, err := feed.ReadBundle(bundle)
	if err != nil {
		t.Fatal(err)
	}

	// 4. Customer node applies the feed against its pinned root.
	st := syncer.NewState()
	if err := syncer.Apply(st, carried, rootPub, time.Now()); err != nil {
		t.Fatal(err)
	}
	auditPath := filepath.Join(t.TempDir(), "audit.log")
	lg, _ := audit.Open(auditPath)
	defer lg.Close()
	gw := &gateway.Gateway{
		State:     func() *syncer.State { return st },
		Upstreams: map[string]gateway.Upstream{svc: {Service: svc, URL: up.URL, ImageDigest: digest}},
		Auth:      gateway.StaticTokens{"tok": "planner-agent"},
		Policy:    gateway.Grants{"planner-agent": {svc: {"*"}}},
		Audit:     lg, Grace: 7 * 24 * time.Hour, Now: time.Now, Client: up.Client(),
	}
	callTool := func() int {
		body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lookup","arguments":{"nsn":"1005-99-123-4567"}}}`
		req, _ := http.NewRequest("POST", "/mcp/"+svc, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer tok")
		rec := httptest.NewRecorder()
		gw.ServeHTTP(rec, req)
		return rec.Code
	}

	// 5. Certified call is allowed.
	if c := callTool(); c != http.StatusOK || calls.Load() != 1 {
		t.Fatalf("certified call: %d, upstream calls %d", c, calls.Load())
	}
	// 6. Rug pull: denied, never reaches upstream.
	mutate(true)
	if c := callTool(); c != http.StatusForbidden || calls.Load() != 1 {
		t.Fatalf("rug pull not blocked: %d", c)
	}
	// 7. Server reverts, but Innovox revokes and publishes seq 2.
	mutate(false)
	writeJSON(t, filepath.Join(store, "revocations.json"), []feed.Revocation{{CertDigest: feed.CertDigest(cert), Reason: "behaviour change observed", RevokedAt: now}})
	env2, err := pub.Publish()
	if err != nil {
		t.Fatal(err)
	}
	if err := syncer.Apply(st, env2, rootPub, time.Now()); err != nil {
		t.Fatal(err)
	}
	if c := callTool(); c != http.StatusForbidden {
		t.Fatalf("revoked cert still usable: %d", c)
	}
	// Replaying the old feed must not un-revoke.
	if err := syncer.Apply(st, carried, rootPub, time.Now()); err == nil {
		t.Fatal("old feed replay accepted")
	}
	// 8. Audit chain is intact and records the rug pull and revocation.
	if err := audit.Verify(auditPath); err != nil {
		t.Fatal(err)
	}
	log, _ := os.ReadFile(auditPath)
	for _, want := range []string{`"action":"rug_pull"`, `"action":"revoked"`, `"decision":"allow"`} {
		if !strings.Contains(string(log), want) {
			t.Fatalf("audit missing %s", want)
		}
	}
}
