package trust

import (
	"crypto/ed25519"
	"crypto/rand"
	"innovox.uk/assured/internal/conformance"
	"testing"
	"time"
)

var now = time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)

func gen(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	p, s, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return p, s
}

func TestDelegationAccepted(t *testing.T) {
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	if err := kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)), now); err != nil {
		t.Fatal(err)
	}
	if _, ok := kr.SignerKeys(now)[KeyID(sPub)]; !ok {
		t.Fatal("delegated key missing")
	}
}

func TestDelegationNotSignedByRootRejected(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	rootPub, _ := gen(t)
	_, evilPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	if err := kr.AddDelegation(Delegate(evilPriv, sPub, now.AddDate(1, 0, 0)), now); err == nil {
		t.Fatal("foreign delegation accepted")
	}
	if len(kr.SignerKeys(now)) != 0 {
		t.Fatal("foreign key present")
	}
}

func TestExpiredDelegationExcluded(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.Add(time.Hour)), now)
	if len(kr.SignerKeys(now.Add(2*time.Hour))) != 0 {
		t.Fatal("expired delegation still valid")
	}
}

func TestRootNotInSignerKeys(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-01")
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)), now)
	if _, ok := kr.SignerKeys(now)[KeyID(rootPub)]; ok {
		t.Fatal("root must not sign certificates")
	}
}

func TestRevokedSignerKeyExcluded(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)), now)
	if err := kr.AddKeyRevocation(RevokeKey(rootPriv, KeyID(sPub), "stolen", now)); err != nil {
		t.Fatal(err)
	}
	if len(kr.SignerKeys(now)) != 0 {
		t.Fatal("revoked signer key still trusted")
	}
	if got := kr.Revoked(); len(got) != 1 || got[0] != KeyID(sPub) {
		t.Fatalf("Revoked() = %v", got)
	}
	// Re-delegating a revoked key must not bring it back.
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(2, 0, 0)), now)
	if len(kr.SignerKeys(now)) != 0 {
		t.Fatal("revoked key returned after re-delegation")
	}
}

func TestKeyRevocationMustBeSignedByRoot(t *testing.T) {
	conformance.Covers(t, "IAS-TRUST-03")
	rootPub, rootPriv := gen(t)
	sPub, sPriv := gen(t)
	other, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)), now)
	_ = kr.AddDelegation(Delegate(rootPriv, other, now.AddDate(1, 0, 0)), now)
	// A signer must not be able to revoke another signer.
	if err := kr.AddKeyRevocation(RevokeKey(sPriv, KeyID(other), "sabotage", now)); err == nil {
		t.Fatal("revocation signed by a non-root key accepted")
	}
	if len(kr.SignerKeys(now)) != 2 {
		t.Fatal("forged revocation took effect")
	}
}
