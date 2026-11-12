// Package feed defines the signed index Innovox publishes and nodes consume,
// and the single-file bundle used for air-gapped transfer.
package feed

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/trust"
)

const PayloadType = "application/vnd.innovox.feed+json"

type Revocation struct {
	CertDigest string    `json:"certDigest"`
	Reason     string    `json:"reason"`
	RevokedAt  time.Time `json:"revokedAt"`
}

type Index struct {
	Seq          uint64            `json:"seq"`
	CreatedAt    time.Time         `json:"createdAt"`
	Delegations  []attest.Envelope `json:"delegations"`
	Certificates []attest.Envelope `json:"certificates"`
	Revocations  []Revocation      `json:"revocations"`
	// KeyRevocations are root-signed statements withdrawing signer keys.
	KeyRevocations []attest.Envelope `json:"keyRevocations,omitempty"`
}

// CertDigest identifies a certificate: hex SHA-256 of its envelope payload.
func CertDigest(env attest.Envelope) string {
	sum := sha256.Sum256(env.Payload)
	return hex.EncodeToString(sum[:])
}

// Sign wraps idx in a DSSE envelope.
func Sign(idx Index, keyID string, priv ed25519.PrivateKey) (attest.Envelope, error) {
	raw, err := json.Marshal(idx)
	if err != nil {
		return attest.Envelope{}, err
	}
	return attest.SignEnvelope(PayloadType, raw, keyID, priv), nil
}

// Open loads the feed's delegations and key revocations into kr (each
// verified against the root) and then verifies the feed signature against
// delegated, unrevoked signer keys only.
func Open(env attest.Envelope, kr *trust.Keyring, now time.Time) (Index, error) {
	idx, _, err := OpenWithSigner(env, kr, now)
	return idx, err
}

// OpenWithSigner is Open, also returning the ID of the key that signed the feed.
func OpenWithSigner(env attest.Envelope, kr *trust.Keyring, now time.Time) (Index, string, error) {
	if env.PayloadType != PayloadType {
		return Index{}, "", fmt.Errorf("feed: payloadType %q", env.PayloadType)
	}
	var idx Index
	if err := json.Unmarshal(env.Payload, &idx); err != nil {
		return Index{}, "", fmt.Errorf("feed: decode: %w", err)
	}
	// Invalid delegations and revocations are ignored: they grant and
	// withdraw nothing. They are root-signed, so they are safe to load
	// before the feed's own signature is checked.
	for _, d := range idx.Delegations {
		_ = kr.AddDelegation(d, now)
	}
	for _, r := range idx.KeyRevocations {
		_ = kr.AddKeyRevocation(r)
	}
	signer, err := attest.VerifyEnvelope(env, kr.SignerKeys(now))
	if err != nil {
		return Index{}, "", fmt.Errorf("feed: %w", err)
	}
	return idx, signer, nil
}

func WriteBundle(path string, env attest.Envelope) error {
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func ReadBundle(path string) (attest.Envelope, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return attest.Envelope{}, err
	}
	var env attest.Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return attest.Envelope{}, fmt.Errorf("feed: bundle: %w", err)
	}
	return env, nil
}
