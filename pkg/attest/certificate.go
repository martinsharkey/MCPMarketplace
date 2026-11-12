package attest

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"
)

const (
	PayloadType   = "application/vnd.in-toto+json"
	StatementType = "https://in-toto.io/Statement/v1"
	PredicateType = "https://innovox.uk/attestation/certificate/v1"
	// ClaimText is the only wording the mark may use (spec §1).
	ClaimText = "Independently tested by Innovox against the Innovox Assured scheme %s, with signed evidence mapped to MOD and NCSC guidance."
)

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

type Statement struct {
	Type          string      `json:"_type"`
	Subject       []Subject   `json:"subject"`
	PredicateType string      `json:"predicateType"`
	Predicate     Certificate `json:"predicate"`
}

type ServiceRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kind    string `json:"kind"` // mcp | a2a
}

type EvidenceRef struct {
	Stage  string `json:"stage"`
	Result string `json:"result"`
	SHA256 string `json:"sha256"`
}

type Certificate struct {
	SchemeVersion    string        `json:"scheme_version"`
	Tier             string        `json:"tier"`
	Service          ServiceRef    `json:"service"`
	ManifestSHA256   string        `json:"manifest_sha256"`
	Evidence         []EvidenceRef `json:"evidence"`
	IssuedAt         time.Time     `json:"issued_at"`
	ExpiresAt        time.Time     `json:"expires_at"`
	Assessor         string        `json:"assessor,omitempty"`
	ClaimsMappingRef string        `json:"claims_mapping_ref,omitempty"`
}

// Claim returns the mandated public claim for this certificate.
func (c Certificate) Claim() string { return fmt.Sprintf(ClaimText, c.SchemeVersion) }

// OpenCertificate verifies an envelope and decodes the Innovox certificate.
func OpenCertificate(env Envelope, keys map[string]ed25519.PublicKey) (Statement, error) {
	st, _, err := OpenCertificateWithSigner(env, keys)
	return st, err
}

// OpenCertificateWithSigner is OpenCertificate, also returning the ID of the
// key that signed the certificate.
func OpenCertificateWithSigner(env Envelope, keys map[string]ed25519.PublicKey) (Statement, string, error) {
	if env.PayloadType != PayloadType {
		return Statement{}, "", fmt.Errorf("attest: payloadType %q", env.PayloadType)
	}
	signer, err := VerifyEnvelope(env, keys)
	if err != nil {
		return Statement{}, "", err
	}
	var st Statement
	if err := json.Unmarshal(env.Payload, &st); err != nil {
		return Statement{}, "", fmt.Errorf("attest: decode: %w", err)
	}
	if st.Type != StatementType || st.PredicateType != PredicateType {
		return Statement{}, "", fmt.Errorf("attest: unexpected statement %q / %q", st.Type, st.PredicateType)
	}
	return st, signer, nil
}
