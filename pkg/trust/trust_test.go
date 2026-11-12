package trust

import (
	"crypto/ed25519"
	"crypto/rand"
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
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.Add(time.Hour)), now)
	if len(kr.SignerKeys(now.Add(2*time.Hour))) != 0 {
		t.Fatal("expired delegation still valid")
	}
}

func TestRootNotInSignerKeys(t *testing.T) {
	rootPub, rootPriv := gen(t)
	sPub, _ := gen(t)
	kr := NewKeyring(rootPub)
	_ = kr.AddDelegation(Delegate(rootPriv, sPub, now.AddDate(1, 0, 0)), now)
	if _, ok := kr.SignerKeys(now)[KeyID(rootPub)]; ok {
		t.Fatal("root must not sign certificates")
	}
}
