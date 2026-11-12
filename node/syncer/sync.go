package syncer

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
)

// ErrReplay means the feed is not newer than the one already applied.
// ErrFutureFeed means the feed claims a creation time in the future, which
// would otherwise extend the offline grace period by the skew.
var ErrFutureFeed = errors.New("syncer: feed created in the future")

// MaxClockSkew is the tolerated clock difference between Innovox and the node.
const MaxClockSkew = 5 * time.Minute

var ErrReplay = errors.New("syncer: feed sequence not newer than current state")

// Apply verifies env against the pinned root and, on success, replaces the
// node's listings with the feed's snapshot. On any error st is unchanged.
// Certificates that fail verification or have expired are skipped.
func Apply(st *State, env attest.Envelope, rootPub ed25519.PublicKey, now time.Time) error {
	kr := trust.NewKeyring(rootPub)
	idx, err := feed.Open(env, kr, now)
	if err != nil {
		return fmt.Errorf("syncer: %w", err)
	}
	if idx.CreatedAt.After(now.Add(MaxClockSkew)) {
		return ErrFutureFeed
	}
	if idx.Seq <= st.Seq {
		return ErrReplay
	}
	revoked := map[string]string{}
	for _, r := range idx.Revocations {
		revoked[r.CertDigest] = r.Reason
	}
	keys := kr.SignerKeys(now)
	listings := map[string]*Listing{}
	for _, c := range idx.Certificates {
		stmt, err := attest.OpenCertificate(c, keys)
		if err != nil || !now.Before(stmt.Predicate.ExpiresAt) {
			continue
		}
		d := feed.CertDigest(c)
		reason, isRevoked := revoked[d]
		listings[ListingKey(stmt.Predicate.Service.Name, stmt.Predicate.Service.Version)] =
			&Listing{Cert: stmt, CertDigest: d, Revoked: isRevoked, RevokeReason: reason}
	}
	st.Seq, st.FeedTime, st.SyncedAt, st.Listings = idx.Seq, idx.CreatedAt, now, listings
	return nil
}
