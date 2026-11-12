package publisher

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
)

func TestPublishIncrementsSeq(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	rp, rs, _ := ed25519.GenerateKey(rand.Reader)
	sp, ss, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "delegations"), 0o755)
	d, _ := json.Marshal(trust.Delegate(rs, sp, now.AddDate(1, 0, 0)))
	os.WriteFile(filepath.Join(dir, "delegations", "s.json"), d, 0o644)
	p := Publisher{Dir: dir, KeyID: trust.KeyID(sp), Priv: ss, Now: func() time.Time { return now }}
	for want := uint64(1); want <= 2; want++ {
		if _, err := p.Publish(); err != nil {
			t.Fatal(err)
		}
		env, err := feed.ReadBundle(filepath.Join(dir, "feed.json"))
		if err != nil {
			t.Fatal(err)
		}
		idx, err := feed.Open(env, trust.NewKeyring(rp), now)
		if err != nil {
			t.Fatal(err)
		}
		if idx.Seq != want {
			t.Fatalf("seq = %d want %d", idx.Seq, want)
		}
	}
}
