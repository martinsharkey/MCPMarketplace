// Command innovox-publish signs a new feed from a store directory.
// Usage: innovox-publish -dir store -key signer.key
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"innovox.uk/assured/assurance/publisher"
	"innovox.uk/assured/pkg/trust"
)

func main() {
	dir := flag.String("dir", "", "store directory")
	keyPath := flag.String("key", "", "hex Ed25519 signer private key file")
	flag.Parse()
	keyHex, err := os.ReadFile(*keyPath)
	if err != nil {
		log.Fatal(err)
	}
	key, err := hex.DecodeString(strings.TrimSpace(string(keyHex)))
	if err != nil || len(key) != ed25519.PrivateKeySize {
		log.Fatal("invalid key file")
	}
	priv := ed25519.PrivateKey(key)
	p := publisher.Publisher{Dir: *dir, KeyID: trust.KeyID(priv.Public().(ed25519.PublicKey)), Priv: priv, Now: time.Now}
	if _, err := p.Publish(); err != nil {
		log.Fatal(err)
	}
}
