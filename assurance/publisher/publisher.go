// Package publisher builds and signs the Innovox feed from a store directory:
//
//	Dir/certs/*.json        certificate envelopes
//	Dir/delegations/*.json  root-signed signer delegations
//	Dir/key-revocations/*.json  root-signed signer-key revocations (optional)
//	Dir/revocations.json    []feed.Revocation (optional)
//	Dir/seq                 last published sequence number
//	Dir/feed.json           output
package publisher

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
)

type Publisher struct {
	Dir   string
	KeyID string
	Priv  ed25519.PrivateKey
	Now   func() time.Time
}

// Publish signs a new feed with seq = previous + 1 and writes Dir/feed.json.
func (p Publisher) Publish() (attest.Envelope, error) {
	certs, err := readEnvelopes(filepath.Join(p.Dir, "certs"))
	if err != nil {
		return attest.Envelope{}, err
	}
	dels, err := readEnvelopes(filepath.Join(p.Dir, "delegations"))
	if err != nil {
		return attest.Envelope{}, err
	}
	keyRevs, err := readEnvelopes(filepath.Join(p.Dir, "key-revocations"))
	if err != nil {
		return attest.Envelope{}, err
	}
	var revs []feed.Revocation
	if b, err := os.ReadFile(filepath.Join(p.Dir, "revocations.json")); err == nil {
		if err := json.Unmarshal(b, &revs); err != nil {
			return attest.Envelope{}, fmt.Errorf("publisher: revocations: %w", err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return attest.Envelope{}, err
	}
	held := map[string]bool{}
	for _, c := range certs {
		held[feed.CertDigest(c)] = true
	}
	for _, r := range revs {
		if !held[r.CertDigest] {
			return attest.Envelope{}, fmt.Errorf("publisher: revocation %q matches no held certificate (want 64 lowercase hex)", r.CertDigest)
		}
	}
	var seq uint64
	if b, err := os.ReadFile(filepath.Join(p.Dir, "seq")); err == nil {
		if seq, err = strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64); err != nil {
			return attest.Envelope{}, fmt.Errorf("publisher: seq: %w", err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return attest.Envelope{}, err
	}
	seq++
	env, err := feed.Sign(feed.Index{Seq: seq, CreatedAt: p.Now().UTC(), Delegations: dels, Certificates: certs, Revocations: revs, KeyRevocations: keyRevs}, p.KeyID, p.Priv)
	if err != nil {
		return attest.Envelope{}, err
	}
	if err := feed.WriteBundle(filepath.Join(p.Dir, "feed.json"), env); err != nil {
		return attest.Envelope{}, err
	}
	return env, os.WriteFile(filepath.Join(p.Dir, "seq"), []byte(strconv.FormatUint(seq, 10)), 0o644)
}

func readEnvelopes(dir string) ([]attest.Envelope, error) {
	paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	sort.Strings(paths)
	out := []attest.Envelope{}
	for _, p := range paths {
		env, err := feed.ReadBundle(p)
		if err != nil {
			return nil, fmt.Errorf("publisher: %s: %w", p, err)
		}
		out = append(out, env)
	}
	return out, nil
}
