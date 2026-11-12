package syncer

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"innovox.uk/assured/internal/conformance"
	"path/filepath"
	"testing"
	"time"

	"innovox.uk/assured/internal/testpki"
	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
)

var now = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func TestApplyListsValidCert(t *testing.T) {
	conformance.Covers(t, "IAS-INV-01")
	p := testpki.New(now)
	st := NewState()
	if err := Apply(st, p.Feed(1, []attest.Envelope{p.Cert("uk.example/echo", testpki.EchoTools(), now)}, nil), p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	l, ok := st.Listings["uk.example/echo@1.0.0"]
	if !ok || l.Revoked || st.Seq != 1 {
		t.Fatalf("listing missing: %+v", st)
	}
}

func TestApplyRejectsReplay(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-02")
	p := testpki.New(now)
	st := NewState()
	f := p.Feed(1, nil, nil)
	if err := Apply(st, f, p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	if err := Apply(st, f, p.RootPub, now); !errors.Is(err, ErrReplay) {
		t.Fatalf("want ErrReplay, got %v", err)
	}
}

func TestApplySkipsExpiredCert(t *testing.T) {
	conformance.Covers(t, "IAS-RT-06")
	p := testpki.New(now)
	st := NewState()
	old := p.Cert("uk.example/old", testpki.EchoTools(), now.AddDate(-2, 0, 0))
	if err := Apply(st, p.Feed(1, []attest.Envelope{old}, nil), p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	if len(st.Listings) != 0 {
		t.Fatal("expired cert listed")
	}
}

func TestApplyMarksRevoked(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-05")
	p := testpki.New(now)
	st := NewState()
	c := p.Cert("uk.example/echo", testpki.EchoTools(), now)
	revs := []feed.Revocation{{CertDigest: feed.CertDigest(c), Reason: "key compromise", RevokedAt: now}}
	if err := Apply(st, p.Feed(1, []attest.Envelope{c}, revs), p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	if l := st.Listings["uk.example/echo@1.0.0"]; l == nil || !l.Revoked || l.RevokeReason != "key compromise" {
		t.Fatalf("not revoked: %+v", l)
	}
}

func TestApplyRejectsForgedFeed(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01", "IAS-TRUST-02")
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, []attest.Envelope{p.Cert("uk.example/echo", testpki.EchoTools(), now)}, nil), p.RootPub, now)
	otherRoot, _, _ := ed25519.GenerateKey(rand.Reader)
	if err := Apply(st, p.Feed(2, nil, nil), otherRoot, now); err == nil {
		t.Fatal("feed from another hierarchy accepted")
	}
	if st.Seq != 1 || len(st.Listings) != 1 {
		t.Fatal("state changed on failed apply")
	}
}

func TestApplySkipsCertFromUntrustedSigner(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	p, evil := testpki.New(now), testpki.New(now)
	st := NewState()
	if err := Apply(st, p.Feed(1, []attest.Envelope{evil.Cert("uk.example/evil", testpki.EchoTools(), now)}, nil), p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	if len(st.Listings) != 0 {
		t.Fatal("cert from untrusted signer listed")
	}
}

func TestStateStaleAfterGrace(t *testing.T) {
	conformance.Covers(t, "IAS-RT-01")
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, nil, nil), p.RootPub, now)
	grace := 7 * 24 * time.Hour
	if !st.Stale(now.AddDate(0, 0, 8), grace) || st.Stale(now.AddDate(0, 0, 6), grace) {
		t.Fatal("staleness wrong")
	}
	if !NewState().Stale(now, grace) {
		t.Fatal("never-synced state must be stale")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, []attest.Envelope{p.Cert("uk.example/echo", testpki.EchoTools(), now)}, nil), p.RootPub, now)
	path := filepath.Join(t.TempDir(), "state.json")
	if err := st.Save(path); err != nil {
		t.Fatal(err)
	}
	back, err := Load(path)
	if err != nil || back.Seq != 1 || back.Listings["uk.example/echo@1.0.0"] == nil {
		t.Fatalf("load: %v %+v", err, back)
	}
}

func TestApplyRejectsFutureDatedFeed(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-02")
	p := testpki.New(now.Add(48 * time.Hour))
	st := NewState()
	if err := Apply(st, p.Feed(1, nil, nil), p.RootPub, now); err == nil {
		t.Fatal("future-dated feed accepted")
	}
}

// rotate gives p a second signer key and returns a root-signed revocation of the first.
func rotate(p *testpki.PKI) (old *testpki.PKI, revocation attest.Envelope) {
	prev := *p
	fresh := testpki.New(p.Now)
	p.SignerPriv, p.SignerID = fresh.SignerPriv, trust.KeyID(fresh.SignerPriv.Public().(ed25519.PublicKey))
	p.Delegation = trust.Delegate(p.RootPriv, fresh.SignerPriv.Public().(ed25519.PublicKey), p.Now.AddDate(1, 0, 0))
	return &prev, trust.RevokeKey(p.RootPriv, prev.SignerID, "signer key compromised", p.Now)
}

func TestRevokedSignerKeyDropsItsCertificates(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	p := testpki.New(now)
	st := NewState()
	oldCert := p.Cert("uk.example/echo", testpki.EchoTools(), now)
	_ = Apply(st, p.Feed(1, []attest.Envelope{oldCert}, nil), p.RootPub, now)
	old, rev := rotate(p)
	newCert := p.Cert("uk.example/new", testpki.EchoTools(), now)
	f := p.FeedWith(2, []attest.Envelope{oldCert, newCert}, nil, []attest.Envelope{old.Delegation}, []attest.Envelope{rev})
	if err := Apply(st, f, p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	if st.Listings["uk.example/echo@1.0.0"] != nil || st.Listings["uk.example/new@1.0.0"] == nil {
		t.Fatalf("listings after key revocation: %v", st.Listings)
	}
}

func TestRevokedSignerKeyCannotPublishLater(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, nil, nil), p.RootPub, now)
	old, rev := rotate(p)
	_ = Apply(st, p.FeedWith(2, nil, nil, nil, []attest.Envelope{rev}), p.RootPub, now)
	// The thief publishes a higher-numbered feed that leaves the revocation out.
	evil := old.Feed(3, []attest.Envelope{old.Cert("uk.example/evil", testpki.EchoTools(), now)}, nil)
	if err := Apply(st, evil, p.RootPub, now); err == nil {
		t.Fatal("feed signed by a revoked key accepted")
	}
	if st.Seq != 2 || len(st.Listings) != 0 {
		t.Fatalf("state changed by revoked signer: seq %d", st.Seq)
	}
}

func TestRecoversFromSequencePoisoningByStolenKey(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, nil, nil), p.RootPub, now)
	// Before the node learns of the theft, the thief publishes a huge
	// sequence number to lock out honest feeds.
	thief := *p
	poison := thief.Feed(1<<62, []attest.Envelope{thief.Cert("uk.example/evil", testpki.EchoTools(), now)}, nil)
	if err := Apply(st, poison, p.RootPub, now); err != nil {
		t.Fatal(err)
	}
	_, rev := rotate(p)
	honest := p.FeedWith(2, []attest.Envelope{p.Cert("uk.example/good", testpki.EchoTools(), now)}, nil, nil, []attest.Envelope{rev})
	if err := Apply(st, honest, p.RootPub, now); err != nil {
		t.Fatalf("honest feed locked out: %v", err)
	}
	if st.Seq != 2 || st.Listings["uk.example/evil@1.0.0"] != nil || st.Listings["uk.example/good@1.0.0"] == nil {
		t.Fatalf("did not recover: seq %d listings %v", st.Seq, st.Listings)
	}
}

func TestKeyRevocationsSurviveSaveAndLoad(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	p := testpki.New(now)
	st := NewState()
	_ = Apply(st, p.Feed(1, nil, nil), p.RootPub, now)
	old, rev := rotate(p)
	_ = Apply(st, p.FeedWith(2, nil, nil, nil, []attest.Envelope{rev}), p.RootPub, now)
	path := filepath.Join(t.TempDir(), "state.json")
	_ = st.Save(path)
	back, _ := Load(path)
	if err := Apply(back, old.Feed(9, nil, nil), p.RootPub, now); err == nil {
		t.Fatal("revocation forgotten after restart")
	}
}
