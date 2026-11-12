// Package syncer verifies Innovox feeds and maintains the node's view of
// which certified services may run.
package syncer

import (
	"encoding/json"
	"os"
	"time"

	"innovox.uk/assured/pkg/attest"
)

type Listing struct {
	Cert         attest.Statement `json:"cert"`
	CertDigest   string           `json:"certDigest"`
	Revoked      bool             `json:"revoked"`
	RevokeReason string           `json:"revokeReason,omitempty"`
}

type State struct {
	Seq      uint64              `json:"seq"`
	FeedTime time.Time           `json:"feedTime"`
	SyncedAt time.Time           `json:"syncedAt"`
	Listings map[string]*Listing `json:"listings"`
}

func NewState() *State { return &State{Listings: map[string]*Listing{}} }

// ListingKey is "<name>@<version>".
func ListingKey(name, version string) string { return name + "@" + version }

// Stale reports whether the feed is older than grace. Age is measured from
// the feed's creation time, which bounds how late a revocation can arrive.
func (s *State) Stale(now time.Time, grace time.Duration) bool {
	return s.Seq == 0 || now.Sub(s.FeedTime) > grace
}

func Load(path string) (*State, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	st := NewState()
	if err := json.Unmarshal(b, st); err != nil {
		return nil, err
	}
	return st, nil
}

// Save writes state atomically (temp file then rename).
func (s *State) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
