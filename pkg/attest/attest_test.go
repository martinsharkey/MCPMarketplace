package attest

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"innovox.uk/assured/internal/conformance"
	"strings"
	"testing"
	"time"
)

func key(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestPAEVector(t *testing.T) {
	got := string(PAE("http://example.com/HelloWorld", []byte("hello world")))
	if want := "DSSEv1 29 http://example.com/HelloWorld 11 hello world"; got != want {
		t.Fatalf("got %q", got)
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	pub, priv := key(t)
	env := SignEnvelope("text/plain", []byte("hi"), "k1", priv)
	id, err := VerifyEnvelope(env, map[string]ed25519.PublicKey{"k1": pub})
	if err != nil || id != "k1" {
		t.Fatalf("verify: %v %q", err, id)
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	pub, priv := key(t)
	env := SignEnvelope("text/plain", []byte("hi"), "k1", priv)
	env.Payload = []byte("ho")
	if _, err := VerifyEnvelope(env, map[string]ed25519.PublicKey{"k1": pub}); err == nil {
		t.Fatal("tampered payload verified")
	}
}

func TestVerifyRejectsUnknownKey(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	_, priv := key(t)
	other, _ := key(t)
	env := SignEnvelope("text/plain", []byte("hi"), "k1", priv)
	if _, err := VerifyEnvelope(env, map[string]ed25519.PublicKey{"k2": other}); err == nil {
		t.Fatal("unknown key verified")
	}
}

func cert() Certificate {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	return Certificate{SchemeVersion: "v0.1", Tier: "T1", Service: ServiceRef{Name: "uk.example/echo", Version: "1.0.0", Kind: "mcp"},
		ManifestSHA256: strings.Repeat("a", 64), IssuedAt: now, ExpiresAt: now.AddDate(1, 0, 0)}
}

func TestOpenCertificateRoundTrip(t *testing.T) {
	pub, priv := key(t)
	st := Statement{Type: StatementType, Subject: []Subject{{Name: "img", Digest: map[string]string{"sha256": strings.Repeat("b", 64)}}},
		PredicateType: PredicateType, Predicate: cert()}
	raw, _ := json.Marshal(st)
	env := SignEnvelope(PayloadType, raw, "k1", priv)
	got, err := OpenCertificate(env, map[string]ed25519.PublicKey{"k1": pub})
	if err != nil {
		t.Fatal(err)
	}
	if got.Predicate.Service.Name != "uk.example/echo" || !got.Predicate.ExpiresAt.Equal(cert().ExpiresAt) {
		t.Fatalf("bad decode: %+v", got)
	}
}

func TestOpenCertificateRejectsWrongPredicateType(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	pub, priv := key(t)
	st := Statement{Type: StatementType, PredicateType: "https://evil.example/v1", Predicate: cert()}
	raw, _ := json.Marshal(st)
	env := SignEnvelope(PayloadType, raw, "k1", priv)
	if _, err := OpenCertificate(env, map[string]ed25519.PublicKey{"k1": pub}); err == nil {
		t.Fatal("wrong predicate type accepted")
	}
}

func TestClaimText(t *testing.T) {
	conformance.Covers(t, "IAS-GOV-01")
	c := cert().Claim()
	if !strings.Contains(c, "Innovox Assured scheme v0.1") || strings.Contains(strings.ToLower(c), "certified") {
		t.Fatalf("bad claim: %q", c)
	}
}
