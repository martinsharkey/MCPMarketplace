package syncer

import (
	"crypto/ed25519"
	"encoding/json"
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
	if st.RevokedCerts == nil {
		st.RevokedCerts = map[string]string{}
	}
	for _, r := range idx.Revocations {
		st.RevokedCerts[r.CertDigest] = r.Reason
	}
	st.Seq, st.FeedTime, st.SyncedAt, st.FeedKeyID, st.Feed = idx.Seq, idx.CreatedAt, now, signer, &env
	st.Listings = derive(idx, kr, now, st.RevokedCerts)
	return nil
}

// derive builds the listings for a verified feed. A certificate is listed if
// it was signed by an unrevoked key that the root had delegated at the time
// the certificate was issued, was not issued in the future, and has not
// expired. The signing key does not need to be current: a certificate
// outlives the routine rotation of the key that signed it.
func derive(idx feed.Index, kr *trust.Keyring, now time.Time, revoked map[string]string) map[string]*Listing {
	keys := kr.AllSignerKeys()
	listings := map[string]*Listing{}
	for _, c := range idx.Certificates {
		stmt, signer, err := attest.OpenCertificateWithSigner(c, keys)
		if err != nil {
			continue
		}
		issued := stmt.Predicate.IssuedAt
		if !kr.ValidAt(signer, issued) || issued.After(now.Add(MaxClockSkew)) || !now.Before(stmt.Predicate.ExpiresAt) {
			continue
		}
		d := feed.CertDigest(c)
		reason, isRevoked := revoked[d]
		listings[ListingKey(stmt.Predicate.Service.Name, stmt.Predicate.Service.Version)] =
			&Listing{Cert: stmt, CertDigest: d, Revoked: isRevoked, RevokeReason: reason}
	}
	return listings
}

// LoadVerified loads the node state and rebuilds its listings from the signed
// feed stored inside it, checked against the pinned root. Anything in the
// file that the signed feed does not support, such as a forged listing or a
// cleared revocation flag, is discarded. A state file whose feed is missing
// or does not verify is an error, and the gateway then denies everything.
//
// The feed's signature is checked as at the time the feed was created, so a
// node keeps working through its grace period after a signer key expires.
// LoadVerified cannot detect the whole file being swapped for an older valid
// one: that needs a monotonic counter held outside the file.
func LoadVerified(path string, rootPub ed25519.PublicKey, now time.Time) (*State, error) {
	st, err := Load(path)
	if err != nil {
		return nil, err
	}
	if st.Feed == nil {
		if st.Seq != 0 || len(st.Listings) != 0 {
			return nil, errors.New("syncer: state has listings but no signed feed")
		}
		return st, nil
	}
	var peek feed.Index
	if err := json.Unmarshal(st.Feed.Payload, &peek); err != nil {
		return nil, fmt.Errorf("syncer: stored feed: %w", err)
	}
	kr := trust.NewKeyring(rootPub)
	for _, id := range st.RevokedKeys {
		kr.Revoke(id)
	}
	idx, signer, err := feed.OpenWithSigner(*st.Feed, kr, peek.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("syncer: stored feed: %w", err)
	}
	if idx.Seq != st.Seq {
		return nil, errors.New("syncer: state sequence does not match its signed feed")
	}
	if st.RevokedCerts == nil {
		st.RevokedCerts = map[string]string{}
	}
	for _, r := range idx.Revocations {
		st.RevokedCerts[r.CertDigest] = r.Reason
	}
	st.RevokedKeys, st.FeedTime, st.FeedKeyID = kr.Revoked(), idx.CreatedAt, signer
	st.Listings = derive(idx, kr, now, st.RevokedCerts)
	return st, nil
}
