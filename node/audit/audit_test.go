package audit

import (
	"innovox.uk/assured/internal/conformance"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func logWith(t *testing.T, n int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "audit.log")
	l, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		if _, err := l.Append("agent-1", "tools/call", "allow", "echo"); err != nil {
			t.Fatal(err)
		}
	}
	l.Close()
	return p
}

func TestAppendChains(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-02")
	p := filepath.Join(t.TempDir(), "audit.log")
	l, _ := Open(p)
	defer l.Close()
	e1, _ := l.Append("a", "x", "allow", "")
	e2, _ := l.Append("a", "y", "deny", "")
	if e1.Seq != 1 || e2.Seq != 2 || e1.PrevHash != strings.Repeat("0", 64) || e2.PrevHash != e1.Hash {
		t.Fatalf("bad chain %+v %+v", e1, e2)
	}
	if err := Verify(p); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyDetectsEdit(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-02")
	p := logWith(t, 3)
	b, _ := os.ReadFile(p)
	os.WriteFile(p, []byte(strings.Replace(string(b), `"decision":"allow"`, `"decision":"deny"`, 1)), 0o600)
	if Verify(p) == nil {
		t.Fatal("edit not detected")
	}
}

func TestVerifyDetectsDeletion(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-02")
	p := logWith(t, 3)
	b, _ := os.ReadFile(p)
	lines := strings.SplitAfter(string(b), "\n")
	os.WriteFile(p, []byte(lines[0]+lines[2]), 0o600)
	if Verify(p) == nil {
		t.Fatal("deletion not detected")
	}
}

func TestReopenContinuesChain(t *testing.T) {
	p := logWith(t, 2)
	l, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	e, _ := l.Append("a", "z", "allow", "")
	l.Close()
	if e.Seq != 3 {
		t.Fatalf("seq = %d", e.Seq)
	}
	if err := Verify(p); err != nil {
		t.Fatal(err)
	}
}

func TestOpenRefusesTamperedLog(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-02")
	p := logWith(t, 2)
	b, _ := os.ReadFile(p)
	os.WriteFile(p, []byte(strings.Replace(string(b), `"actor":"agent-1"`, `"actor":"agent-2"`, 1)), 0o600)
	if _, err := Open(p); err == nil {
		t.Fatal("opened tampered log")
	}
}
