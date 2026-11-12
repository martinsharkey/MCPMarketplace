package publisher

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"innovox.uk/assured/internal/conformance"
	"os"
	"path/filepath"
	"testing"
	"time"

	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
)

func TestPublishIncrementsSeq(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-02")
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	rp, rs, _ := ed25519.GenerateKey(rand.Reader)
	sp, ss, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "delegations"), 0o755)
	d, _ := json.Marshal(trust.Delegate(rs, sp, now.AddDate(1, 0, 0)))
	os.WriteFile(filepath.Join(dir, "delegations", "s.json"), d, 0o644)
	p := Publisher{Dir: dir, KeyID: trust.KeyID(sp), Priv: ss, Now: func() time.Time { return now }}
	for want := uint64(1); want <= 2; want++ {
		if _, err := p.Publish(); err != nil {
			t.Fatal(err)
		}
		env, err := feed.ReadBundle(filepath.Join(dir, "feed.json"))
		if err != nil {
			t.Fatal(err)
		}
		idx, err := feed.Open(env, trust.NewKeyring(rp), now)
		if err != nil {
			t.Fatal(err)
		}
		if idx.Seq != want {
			t.Fatalf("seq = %d want %d", idx.Seq, want)
		}
	}
}

func TestPublishRejectsUnknownRevocationDigest(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-05")
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	_, ss, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	revs, _ := json.Marshal([]feed.Revocation{{CertDigest: "SHA256:ABC", Reason: "typo", RevokedAt: now}})
	os.WriteFile(filepath.Join(dir, "revocations.json"), revs, 0o644)
	p := Publisher{Dir: dir, KeyID: "k", Priv: ss, Now: func() time.Time { return now }}
	if _, err := p.Publish(); err == nil {
		t.Fatal("revocation for unknown certificate published")
	}
}

func TestPublishCarriesKeyRevocations(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	rp, rs, _ := ed25519.GenerateKey(rand.Reader)
	oldPub, _, _ := ed25519.GenerateKey(rand.Reader)
	sp, ss, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	write := func(sub, name string, v any) {
		os.MkdirAll(filepath.Join(dir, sub), 0o755)
		b, _ := json.Marshal(v)
		os.WriteFile(filepath.Join(dir, sub, name), b, 0o644)
	}
	write("delegations", "old.json", trust.Delegate(rs, oldPub, now.AddDate(1, 0, 0)))
	write("delegations", "new.json", trust.Delegate(rs, sp, now.AddDate(1, 0, 0)))
	write("key-revocations", "old.json", trust.RevokeKey(rs, trust.KeyID(oldPub), "stolen", now))
	p := Publisher{Dir: dir, KeyID: trust.KeyID(sp), Priv: ss, Now: func() time.Time { return now }}
	env, err := p.Publish()
	if err != nil {
		t.Fatal(err)
	}
	kr := trust.NewKeyring(rp)
	if _, err := feed.Open(env, kr, now); err != nil {
		t.Fatal(err)
	}
	if got := kr.Revoked(); len(got) != 1 || got[0] != trust.KeyID(oldPub) {
		t.Fatalf("revocation not carried: %v", got)
	}
}
