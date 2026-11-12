package audit

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"innovox.uk/assured/internal/conformance"
)

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("siem down") }

func open5(t *testing.T) (*Log, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "audit.log")
	l, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	for i := 0; i < 5; i++ {
		l.Append("agent", "tools/call", "allow", "n")
	}
	return l, p
}

func TestExportWritesEveryEventToSink(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	p := filepath.Join(t.TempDir(), "audit.log")
	l, _ := Open(p)
	defer l.Close()
	var siem bytes.Buffer
	l.Sink = &siem
	l.Append("a", "x", "allow", "")
	l.Append("a", "y", "deny", "")
	file, _ := os.ReadFile(p)
	if siem.String() != string(file) || strings.Count(siem.String(), "\n") != 2 {
		t.Fatalf("sink does not mirror the log:\n%s", siem.String())
	}
}

func TestSinkFailureDoesNotStopAuditButIsCounted(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	l, p := open5(t)
	l.Sink = failWriter{}
	if _, err := l.Append("a", "x", "allow", ""); err != nil {
		t.Fatalf("local audit must not depend on the SIEM: %v", err)
	}
	if l.SinkErrors() != 1 {
		t.Fatalf("sink errors = %d", l.SinkErrors())
	}
	if err := Verify(p); err != nil {
		t.Fatal(err)
	}
}

func TestCheckpointDetectsTruncation(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	l, p := open5(t)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	cp, err := l.Checkpoint(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyCheckpoint(p, cp, pub); err != nil {
		t.Fatalf("intact log failed: %v", err)
	}
	b, _ := os.ReadFile(p)
	lines := strings.SplitAfter(string(b), "\n")
	os.WriteFile(p, []byte(strings.Join(lines[:3], "")), 0o600)
	if Verify(p) != nil {
		t.Fatal("precondition: a truncated chain still verifies on its own")
	}
	if VerifyCheckpoint(p, cp, pub) == nil {
		t.Fatal("truncation not detected")
	}
}

func TestCheckpointDetectsRewrittenHistory(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	l, p := open5(t)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	cp, _ := l.Checkpoint(priv)
	l.Close()
	os.Remove(p)
	l2, _ := Open(p) // an attacker rebuilds a clean, valid chain of the same length
	for i := 0; i < 5; i++ {
		l2.Append("agent", "tools/call", "allow", "nothing to see")
	}
	l2.Close()
	if VerifyCheckpoint(p, cp, pub) == nil {
		t.Fatal("rewritten history not detected")
	}
}

func TestCheckpointRejectsWrongKey(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	l, p := open5(t)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	other, _, _ := ed25519.GenerateKey(rand.Reader)
	cp, _ := l.Checkpoint(priv)
	if VerifyCheckpoint(p, cp, other) == nil {
		t.Fatal("checkpoint verified under the wrong key")
	}
}

func TestCheckpointIsExported(t *testing.T) {
	conformance.Covers(t, "IAS-AUD-03")
	l, _ := open5(t)
	var siem bytes.Buffer
	l.Sink = &siem
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	l.Checkpoint(priv)
	if !strings.Contains(siem.String(), `{"checkpoint":`) {
		t.Fatalf("checkpoint not exported: %q", siem.String())
	}
}
