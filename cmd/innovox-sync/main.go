// Command innovox-sync applies a feed bundle to the node state.
// Usage: innovox-sync -root root.pub -bundle feed.json -state state.json
// root.pub holds the hex-encoded Innovox root public key pinned at install.
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"innovox.uk/assured/node/syncer"
	"innovox.uk/assured/pkg/feed"
)

func main() {
	rootPath := flag.String("root", "", "hex root public key file")
	bundle := flag.String("bundle", "", "feed bundle")
	statePath := flag.String("state", "state.json", "node state file")
	flag.Parse()
	rb, err := os.ReadFile(*rootPath)
	if err != nil {
		log.Fatal(err)
	}
	root, err := hex.DecodeString(strings.TrimSpace(string(rb)))
	if err != nil || len(root) != ed25519.PublicKeySize {
		log.Fatal("invalid root key")
	}
	env, err := feed.ReadBundle(*bundle)
	if err != nil {
		log.Fatal(err)
	}
	st, err := syncer.Load(*statePath)
	if errors.Is(err, fs.ErrNotExist) {
		st = syncer.NewState()
	} else if err != nil {
		log.Fatal(err)
	}
	if err := syncer.Apply(st, env, root, time.Now()); err != nil {
		log.Fatal(err)
	}
	if err := st.Save(*statePath); err != nil {
		log.Fatal(err)
	}
	log.Printf("applied feed seq %d: %d listings", st.Seq, len(st.Listings))
}
