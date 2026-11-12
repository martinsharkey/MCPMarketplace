// Package signer issues Innovox Assured certificates.
package signer

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/manifest"
	"innovox.uk/assured/pkg/trust"
)

// Request is everything the pipeline hands the signer for one service version.
type Request struct {
	Service       attest.ServiceRef    `json:"service"`
	ImageDigest   string               `json:"imageDigest,omitempty"` // sha256:<hex>
	RemoteURL     string               `json:"remoteUrl,omitempty"`
	Tools         []manifest.Tool      `json:"tools"`
	Tier          string               `json:"tier"`
	Assessor      string               `json:"assessor,omitempty"`
	SchemeVersion string               `json:"schemeVersion"`
	Evidence      []attest.EvidenceRef `json:"evidence,omitempty"`
}

type Signer struct {
	Priv ed25519.PrivateKey
	Now  func() time.Time
}

// Issue validates r against the tier rules and returns a signed certificate.
func (s Signer) Issue(r Request) (attest.Envelope, error) {
	if (r.ImageDigest == "") == (r.RemoteURL == "") {
		return attest.Envelope{}, errors.New("signer: exactly one of imageDigest or remoteUrl is required")
	}
	switch r.Tier {
	case "T1":
	case "T2":
		if r.Assessor == "" {
			return attest.Envelope{}, errors.New("signer: T2 requires an assessor")
		}
		if r.RemoteURL != "" {
			return attest.Envelope{}, errors.New("signer: remote endpoints are capped at T1")
		}
	default:
		return attest.Envelope{}, fmt.Errorf("signer: tier %q not issuable in pilot", r.Tier)
	}
	if r.Service.Name == "" || r.Service.Version == "" || (r.Service.Kind != "mcp" && r.Service.Kind != "a2a") {
		return attest.Envelope{}, errors.New("signer: service name, version and kind (mcp|a2a) are required")
	}
	hash, err := manifest.HashTools(r.Tools)
	if err != nil {
		return attest.Envelope{}, err
	}
	var subject attest.Subject
	if r.RemoteURL != "" {
		subject = attest.Subject{Name: "remote:" + r.RemoteURL, Digest: map[string]string{"sha256": hash}}
	} else {
		hex64, ok := strings.CutPrefix(r.ImageDigest, "sha256:")
		if _, err := hex.DecodeString(hex64); !ok || err != nil || len(hex64) != 64 {
			return attest.Envelope{}, fmt.Errorf("signer: malformed image digest %q", r.ImageDigest)
		}
		subject = attest.Subject{Name: r.Service.Name, Digest: map[string]string{"sha256": hex64}}
	}
	now := s.Now().UTC()
	st := attest.Statement{
		Type:          attest.StatementType,
		Subject:       []attest.Subject{subject},
		PredicateType: attest.PredicateType,
		Predicate: attest.Certificate{
			SchemeVersion: r.SchemeVersion, Tier: r.Tier, Service: r.Service, ManifestSHA256: hash,
			Evidence: r.Evidence, IssuedAt: now, ExpiresAt: now.AddDate(1, 0, 0), Assessor: r.Assessor,
		},
	}
	raw, err := json.Marshal(st)
	if err != nil {
		return attest.Envelope{}, err
	}
	pub := s.Priv.Public().(ed25519.PublicKey)
	return attest.SignEnvelope(attest.PayloadType, raw, trust.KeyID(pub), s.Priv), nil
}
