package feed

import (
	"crypto/ed25519"
	"crypto/rand"
	"innovox.uk/assured/internal/conformance"
	"path/filepath"
	"testing"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/trust"
)

var now = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

type keys struct {
	rootPub         ed25519.PublicKey
	rootPriv, sPriv ed25519.PrivateKey
	sID             string
	delegation      attest.Envelope
}

func mk(t *testing.T) keys {
	t.Helper()
	rp, rs, _ := ed25519.GenerateKey(rand.Reader)
	sp, ss, _ := ed25519.GenerateKey(rand.Reader)
	return keys{rp, rs, ss, trust.KeyID(sp), trust.Delegate(rs, sp, now.AddDate(1, 0, 0))}
}

func TestFeedRoundTrip(t *testing.T) {
	k := mk(t)
	cert := attest.SignEnvelope(attest.PayloadType, []byte(`{}`), k.sID, k.sPriv)
	idx := Index{Seq: 3, CreatedAt: now, Delegations: []attest.Envelope{k.delegation}, Certificates: []attest.Envelope{cert},
		Revocations: []Revocation{{CertDigest: "ab", Reason: "test", RevokedAt: now}}}
	env, err := Sign(idx, k.sID, k.sPriv)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Open(env, trust.NewKeyring(k.rootPub), now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Seq != 3 || len(got.Certificates) != 1 || got.Revocations[0].Reason != "test" {
		t.Fatalf("bad index %+v", got)
	}
}

func TestOpenRejectsFeedSignedByRoot(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	k := mk(t)
	env, _ := Sign(Index{Seq: 1, CreatedAt: now, Delegations: []attest.Envelope{k.delegation}}, trust.KeyID(k.rootPub), k.rootPriv)
	if _, err := Open(env, trust.NewKeyring(k.rootPub), now); err == nil {
		t.Fatal("root-signed feed accepted")
	}
}

func TestOpenRejectsForeignDelegation(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	k := mk(t)
	evil := mk(t)
	env, _ := Sign(Index{Seq: 1, CreatedAt: now, Delegations: []attest.Envelope{evil.delegation}}, evil.sID, evil.sPriv)
	if _, err := Open(env, trust.NewKeyring(k.rootPub), now); err == nil {
		t.Fatal("feed signed by foreign hierarchy accepted")
	}
}

func TestBundleRoundTrip(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-04")
	k := mk(t)
	env, _ := Sign(Index{Seq: 1, CreatedAt: now, Delegations: []attest.Envelope{k.delegation}}, k.sID, k.sPriv)
	p := filepath.Join(t.TempDir(), "bundle.json")
	if err := WriteBundle(p, env); err != nil {
		t.Fatal(err)
	}
	back, err := ReadBundle(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(back, trust.NewKeyring(k.rootPub), now); err != nil {
		t.Fatal(err)
	}
}

func TestCertDigestStable(t *testing.T) {
	e := attest.Envelope{Payload: []byte("abc")}
	if CertDigest(e) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal(CertDigest(e))
	}
}
