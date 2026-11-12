package signer

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/manifest"
	"innovox.uk/assured/pkg/trust"
)

var now = time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)

func setup(t *testing.T) (Signer, map[string]ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return Signer{Priv: priv, Now: func() time.Time { return now }}, map[string]ed25519.PublicKey{trust.KeyID(pub): pub}
}

var digest = "sha256:" + strings.Repeat("c", 64)

func req(tier string) Request {
	return Request{Service: attest.ServiceRef{Name: "uk.example/echo", Version: "1.0.0", Kind: "mcp"}, ImageDigest: digest,
		Tools: []manifest.Tool{{Name: "echo", Description: "Echo", InputSchema: json.RawMessage(`{"type":"object"}`)}},
		Tier:  tier, SchemeVersion: "v0.1"}
}

func open(t *testing.T, s Signer, keys map[string]ed25519.PublicKey, r Request) attest.Statement {
	t.Helper()
	env, err := s.Issue(r)
	if err != nil {
		t.Fatal(err)
	}
	st, err := attest.OpenCertificate(env, keys)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func TestIssueT1Auto(t *testing.T) {
	s, keys := setup(t)
	st := open(t, s, keys, req("T1"))
	c := st.Predicate
	if c.Tier != "T1" || !c.IssuedAt.Equal(now) || !c.ExpiresAt.Equal(now.AddDate(1, 0, 0)) {
		t.Fatalf("bad cert: %+v", c)
	}
	want, _ := manifest.HashTools(req("T1").Tools)
	if c.ManifestSHA256 != want {
		t.Fatal("manifest hash not bound")
	}
}

func TestIssueT2RequiresAssessor(t *testing.T) {
	s, keys := setup(t)
	if _, err := s.Issue(req("T2")); err == nil {
		t.Fatal("T2 without assessor issued")
	}
	r := req("T2")
	r.Assessor = "a.assessor@innovox.uk"
	if st := open(t, s, keys, r); st.Predicate.Assessor != r.Assessor {
		t.Fatal("assessor missing")
	}
}

func TestIssueRejectsT3(t *testing.T) {
	s, _ := setup(t)
	r := req("T3")
	r.Assessor = "x"
	if _, err := s.Issue(r); err == nil {
		t.Fatal("T3 issued in pilot")
	}
}

func TestRemoteCappedAtT1(t *testing.T) {
	s, _ := setup(t)
	r := req("T2")
	r.Assessor, r.ImageDigest, r.RemoteURL = "x", "", "https://svc.example/mcp"
	if _, err := s.Issue(r); err == nil {
		t.Fatal("remote T2 issued")
	}
}

func TestRemoteSubject(t *testing.T) {
	s, keys := setup(t)
	r := req("T1")
	r.ImageDigest, r.RemoteURL = "", "https://svc.example/mcp"
	st := open(t, s, keys, r)
	sub := st.Subject[0]
	if sub.Name != "remote:https://svc.example/mcp" || sub.Digest["sha256"] != st.Predicate.ManifestSHA256 {
		t.Fatalf("bad remote subject %+v", sub)
	}
}

func TestImageSubject(t *testing.T) {
	s, keys := setup(t)
	st := open(t, s, keys, req("T1"))
	if st.Subject[0].Digest["sha256"] != strings.Repeat("c", 64) {
		t.Fatalf("bad image subject %+v", st.Subject[0])
	}
}

func TestRejectsBothOrNeitherImageAndRemote(t *testing.T) {
	s, _ := setup(t)
	r := req("T1")
	r.RemoteURL = "https://x"
	if _, err := s.Issue(r); err == nil {
		t.Fatal("both accepted")
	}
	r.RemoteURL, r.ImageDigest = "", ""
	if _, err := s.Issue(r); err == nil {
		t.Fatal("neither accepted")
	}
}

func TestRejectsMalformedDigest(t *testing.T) {
	s, _ := setup(t)
	r := req("T1")
	r.ImageDigest = "sha256:xyz"
	if _, err := s.Issue(r); err == nil {
		t.Fatal("malformed digest accepted")
	}
}
