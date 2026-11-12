package syncer

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"slices"
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
// node's listings with the feed's snapshot. Certificates that fail
// verification or have expired are skipped.
//
// If the feed does not verify, st is unchanged. Once a feed verifies, any
// signer-key revocations it carries are recorded even if the feed is then
// rejected as stale or future-dated, because they are root-signed and must
// never be lost. If the feed the node currently holds was signed by a key
// that is now revoked, its listings and sequence number are discarded, so a
// thief cannot lock out honest feeds with a very high sequence number.
func Apply(st *State, env attest.Envelope, rootPub ed25519.PublicKey, now time.Time) error {
	kr := trust.NewKeyring(rootPub)
	for _, id := range st.RevokedKeys {
		kr.Revoke(id)
	}
	idx, signer, err := feed.OpenWithSigner(env, kr, now)
	if err != nil {
		return fmt.Errorf("syncer: %w", err)
	}
	st.RevokedKeys = kr.Revoked()
	if st.FeedKeyID != "" && slices.Contains(st.RevokedKeys, st.FeedKeyID) {
		st.Seq, st.FeedKeyID, st.FeedTime, st.Listings = 0, "", time.Time{}, map[string]*Listing{}
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
	st.Seq, st.FeedTime, st.SyncedAt, st.Listings, st.FeedKeyID = idx.Seq, idx.CreatedAt, now, listings, signer
	return nil
}
