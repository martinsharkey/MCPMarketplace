// Package trust holds the Innovox key hierarchy: an offline root key that
// delegates time-bound online signing keys.
package trust

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"innovox.uk/assured/pkg/attest"
)

const (
	DelegationPayloadType    = "application/vnd.innovox.delegation+json"
	KeyRevocationPayloadType = "application/vnd.innovox.keyrevocation+json"
)

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

// KeyRevocation withdraws a signer key for good. Only the root can issue one.
type KeyRevocation struct {
	KeyID     string    `json:"keyid"`
	Reason    string    `json:"reason"`
	RevokedAt time.Time `json:"revokedAt"`
}

// RevokeKey has the root permanently revoke a signer key.
func RevokeKey(rootPriv ed25519.PrivateKey, keyID, reason string, at time.Time) attest.Envelope {
	raw, _ := json.Marshal(KeyRevocation{KeyID: keyID, Reason: reason, RevokedAt: at.UTC()})
	rootPub := rootPriv.Public().(ed25519.PublicKey)
	return attest.SignEnvelope(KeyRevocationPayloadType, raw, KeyID(rootPub), rootPriv)
}

// Keyring holds the pinned root key, the delegations it has verified and
// the signer keys it has revoked.
type Keyring struct {
	root        ed25519.PublicKey
	delegations []Delegation
	revoked     map[string]bool
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

// AddKeyRevocation verifies env against the root and revokes the key it names.
func (k *Keyring) AddKeyRevocation(env attest.Envelope) error {
	if env.PayloadType != KeyRevocationPayloadType {
		return fmt.Errorf("trust: payloadType %q", env.PayloadType)
	}
	if _, err := attest.VerifyEnvelope(env, map[string]ed25519.PublicKey{KeyID(k.root): k.root}); err != nil {
		return fmt.Errorf("trust: key revocation not signed by root: %w", err)
	}
	var r KeyRevocation
	if err := json.Unmarshal(env.Payload, &r); err != nil || r.KeyID == "" {
		return fmt.Errorf("trust: malformed key revocation")
	}
	k.Revoke(r.KeyID)
	return nil
}

// Revoke marks a key revoked without a signed statement. It is for
// re-loading revocations the node has already verified and stored.
func (k *Keyring) Revoke(keyID string) {
	if k.revoked == nil {
		k.revoked = map[string]bool{}
	}
	k.revoked[keyID] = true
}

// Revoked returns the revoked key IDs, sorted.
func (k *Keyring) Revoked() []string {
	out := make([]string, 0, len(k.revoked))
	for id := range k.revoked {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// SignerKeys returns delegated keys that are valid at now and not revoked.
// The root is never included.
func (k *Keyring) SignerKeys(now time.Time) map[string]ed25519.PublicKey {
	out := map[string]ed25519.PublicKey{}
	for _, d := range k.delegations {
		if k.revoked[d.KeyID] {
			continue
		}
		if !now.Before(d.NotBefore) && now.Before(d.NotAfter) {
			out[d.KeyID] = ed25519.PublicKey(d.PublicKey)
		}
	}
	return out
}

// AllSignerKeys returns every delegated key that is not revoked, whether or
// not its delegation is current. Use it with ValidAt to check something that
// was signed in the past, such as a certificate.
func (k *Keyring) AllSignerKeys() map[string]ed25519.PublicKey {
	out := map[string]ed25519.PublicKey{}
	for _, d := range k.delegations {
		if !k.revoked[d.KeyID] {
			out[d.KeyID] = ed25519.PublicKey(d.PublicKey)
		}
	}
	return out
}

// ValidAt reports whether keyID was delegated at time t and is not revoked.
func (k *Keyring) ValidAt(keyID string, t time.Time) bool {
	if k.revoked[keyID] {
		return false
	}
	for _, d := range k.delegations {
		if d.KeyID == keyID && !t.Before(d.NotBefore) && t.Before(d.NotAfter) {
			return true
		}
	}
	return false
}
