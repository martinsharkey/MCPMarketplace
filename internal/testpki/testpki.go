// Package testpki builds an Innovox key hierarchy and certificates for tests.
package testpki

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"time"

	"innovox.uk/assured/assurance/signer"
	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/manifest"
	"innovox.uk/assured/pkg/trust"
)

type PKI struct {
	RootPub    ed25519.PublicKey
	RootPriv   ed25519.PrivateKey
	SignerPriv ed25519.PrivateKey
	SignerID   string
	Delegation attest.Envelope
	Now        time.Time
}

func New(now time.Time) *PKI {
	rp, rs, _ := ed25519.GenerateKey(rand.Reader)
	sp, ss, _ := ed25519.GenerateKey(rand.Reader)
	return &PKI{RootPub: rp, RootPriv: rs, SignerPriv: ss, SignerID: trust.KeyID(sp),
		Delegation: trust.Delegate(rs, sp, now.AddDate(2, 0, 0)), Now: now}
}

var Digest = "sha256:" + strings.Repeat("d", 64)

func EchoTools() []manifest.Tool {
	return []manifest.Tool{
		{Name: "echo", Description: "Echo text", InputSchema: json.RawMessage(`{"type":"object"}`)},
		{Name: "add", Description: "Add numbers", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
}

// Cert issues a T1 certificate for name@1.0.0 at issuedAt.
func (p *PKI) Cert(name string, tools []manifest.Tool, issuedAt time.Time) attest.Envelope {
	env, err := signer.Signer{Priv: p.SignerPriv, Now: func() time.Time { return issuedAt }}.Issue(signer.Request{
		Service: attest.ServiceRef{Name: name, Version: "1.0.0", Kind: "mcp"}, ImageDigest: Digest,
		Tools: tools, Tier: "T1", SchemeVersion: "v0.1"})
	if err != nil {
		panic(err)
	}
	return env
}

// Feed signs an index with the given seq, certs and revocations.
func (p *PKI) Feed(seq uint64, certs []attest.Envelope, revs []feed.Revocation) attest.Envelope {
	env, err := feed.Sign(feed.Index{Seq: seq, CreatedAt: p.Now, Delegations: []attest.Envelope{p.Delegation},
		Certificates: certs, Revocations: revs}, p.SignerID, p.SignerPriv)
	if err != nil {
		panic(err)
	}
	return env
}
