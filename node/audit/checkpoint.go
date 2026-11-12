package audit

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/trust"
)

const CheckpointPayloadType = "application/vnd.innovox.audit-checkpoint+json"

// Checkpoint is a signed statement of where the chain stood at a moment.
// Held outside the node (in the SIEM), it makes truncating or rebuilding
// the log detectable, which the hash chain alone cannot do.
type Checkpoint struct {
	Seq  uint64    `json:"seq"`
	Hash string    `json:"hash"`
	TS   time.Time `json:"ts"`
}

// Checkpoint signs the current head of the chain and exports it to the sink.
func (l *Log) Checkpoint(priv ed25519.PrivateKey) (attest.Envelope, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	raw, err := json.Marshal(Checkpoint{Seq: l.seq, Hash: l.prev, TS: l.Now().UTC()})
	if err != nil {
		return attest.Envelope{}, err
	}
	env := attest.SignEnvelope(CheckpointPayloadType, raw, trust.KeyID(priv.Public().(ed25519.PublicKey)), priv)
	line, _ := json.Marshal(map[string]attest.Envelope{"checkpoint": env})
	l.export(append(line, '\n'))
	return env, nil
}

// VerifyCheckpoint checks that the log at path still contains the event the
// checkpoint names, with the same hash, and that the chain up to it is intact.
func VerifyCheckpoint(path string, cp attest.Envelope, pub ed25519.PublicKey) error {
	if cp.PayloadType != CheckpointPayloadType {
		return fmt.Errorf("audit: payloadType %q", cp.PayloadType)
	}
	if _, err := attest.VerifyEnvelope(cp, map[string]ed25519.PublicKey{trust.KeyID(pub): pub}); err != nil {
		return fmt.Errorf("audit: checkpoint: %w", err)
	}
	var c Checkpoint
	if err := json.Unmarshal(cp.Payload, &c); err != nil {
		return fmt.Errorf("audit: checkpoint: %w", err)
	}
	if c.Seq == 0 {
		return nil // nothing had been logged yet
	}
	seq, hash, err := walkTo(path, c.Seq)
	if err != nil {
		return err
	}
	if seq != c.Seq {
		return fmt.Errorf("audit: log ends at %d but checkpoint covers %d: log truncated", seq, c.Seq)
	}
	if hash != c.Hash {
		return fmt.Errorf("audit: event %d does not match checkpoint: history rewritten", c.Seq)
	}
	return nil
}
