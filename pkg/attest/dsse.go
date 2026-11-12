// Package attest implements DSSE envelopes and the Innovox certificate
// predicate carried inside an in-toto Statement.
package attest

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

// Envelope is a DSSE v1 envelope. Payload is base64 in JSON form.
type Envelope struct {
	PayloadType string      `json:"payloadType"`
	Payload     []byte      `json:"payload"`
	Signatures  []Signature `json:"signatures"`
}

// Signature is one DSSE signature.
type Signature struct {
	KeyID string `json:"keyid"`
	Sig   []byte `json:"sig"`
}

// PAE is the DSSE v1 pre-authentication encoding.
func PAE(payloadType string, payload []byte) []byte {
	return fmt.Appendf(nil, "DSSEv1 %d %s %d %s", len(payloadType), payloadType, len(payload), payload)
}

// SignEnvelope signs payload with an Ed25519 key.
func SignEnvelope(payloadType string, payload []byte, keyID string, priv ed25519.PrivateKey) Envelope {
	return Envelope{
		PayloadType: payloadType,
		Payload:     payload,
		Signatures:  []Signature{{KeyID: keyID, Sig: ed25519.Sign(priv, PAE(payloadType, payload))}},
	}
}

// ErrNoValidSignature is returned when no signature verifies under a known key.
var ErrNoValidSignature = errors.New("attest: no valid signature from a trusted key")

// VerifyEnvelope returns the key ID of the first signature that verifies.
func VerifyEnvelope(env Envelope, keys map[string]ed25519.PublicKey) (string, error) {
	msg := PAE(env.PayloadType, env.Payload)
	for _, s := range env.Signatures {
		pub, ok := keys[s.KeyID]
		if ok && len(pub) == ed25519.PublicKeySize && ed25519.Verify(pub, msg, s.Sig) {
			return s.KeyID, nil
		}
	}
	return "", ErrNoValidSignature
}
