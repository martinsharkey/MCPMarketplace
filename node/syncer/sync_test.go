package syncer

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"innovox.uk/assured/internal/testpki"
	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
)

var now = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func TestApplyListsValidCert(t *testing.T) {
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
	p := testpki.New(now.Add(48 * time.Hour))
	st := NewState()
	if err := Apply(st, p.Feed(1, nil, nil), p.RootPub, now); err == nil {
		t.Fatal("future-dated feed accepted")
	}
}
