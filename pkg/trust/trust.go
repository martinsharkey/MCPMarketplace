// Package trust holds the Innovox key hierarchy: an offline root key that
// delegates time-bound online signing keys.
package trust

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"innovox.uk/assured/pkg/attest"
)

const DelegationPayloadType = "application/vnd.innovox.delegation+json"

// KeyID is the first 16 hex characters of SHA-256(public key).
func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])[:16]
}

type Delegation struct {
	KeyID     string    `json:"keyid"`
	PublicKey []byte    `json:"publicKey"`
	Role      string    `json:"role"`
	NotBefore time.Time `json:"notBefore"`
	NotAfter  time.Time `json:"notAfter"`
}

// Delegate has the root authorise signerPub to sign until notAfter.
func Delegate(rootPriv ed25519.PrivateKey, signerPub ed25519.PublicKey, notAfter time.Time) attest.Envelope {
	d := Delegation{KeyID: KeyID(signerPub), PublicKey: signerPub, Role: "signer",
		NotBefore: time.Unix(0, 0).UTC(), NotAfter: notAfter.UTC()}
	raw, _ := json.Marshal(d)
	rootPub := rootPriv.Public().(ed25519.PublicKey)
	return attest.SignEnvelope(DelegationPayloadType, raw, KeyID(rootPub), rootPriv)
}

// Keyring holds the pinned root key and the delegations it has verified.
type Keyring struct {
	root        ed25519.PublicKey
	delegations []Delegation
}

func NewKeyring(rootPub ed25519.PublicKey) *Keyring { return &Keyring{root: rootPub} }

// AddDelegation verifies env against the root and records it.
func (k *Keyring) AddDelegation(env attest.Envelope, now time.Time) error {
	if env.PayloadType != DelegationPayloadType {
		return fmt.Errorf("trust: payloadType %q", env.PayloadType)
	}
	if _, err := attest.VerifyEnvelope(env, map[string]ed25519.PublicKey{KeyID(k.root): k.root}); err != nil {
		return fmt.Errorf("trust: delegation not signed by root: %w", err)
	}
	var d Delegation
	if err := json.Unmarshal(env.Payload, &d); err != nil {
		return fmt.Errorf("trust: decode delegation: %w", err)
	}
	if d.Role != "signer" || len(d.PublicKey) != ed25519.PublicKeySize || d.KeyID != KeyID(d.PublicKey) {
		return fmt.Errorf("trust: malformed delegation")
	}
	k.delegations = append(k.delegations, d)
	return nil
}

// SignerKeys returns delegated keys valid at now. The root is never included.
func (k *Keyring) SignerKeys(now time.Time) map[string]ed25519.PublicKey {
	out := map[string]ed25519.PublicKey{}
	for _, d := range k.delegations {
		if !now.Before(d.NotBefore) && now.Before(d.NotAfter) {
			out[d.KeyID] = ed25519.PublicKey(d.PublicKey)
		}
	}
	return out
}
