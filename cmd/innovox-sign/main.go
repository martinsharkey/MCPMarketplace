// Command innovox-sign issues a certificate from a request JSON file.
// Usage: innovox-sign -key signer.key -in req.json -out cert.json
// The key file holds a hex-encoded Ed25519 private key.
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"innovox.uk/assured/assurance/signer"
)

func main() {
	keyPath := flag.String("key", "", "hex Ed25519 private key file")
	in := flag.String("in", "", "request JSON")
	out := flag.String("out", "", "certificate envelope output")
	flag.Parse()
	keyHex, err := os.ReadFile(*keyPath)
	if err != nil {
		log.Fatal(err)
	}
	key, err := hex.DecodeString(strings.TrimSpace(string(keyHex)))
	if err != nil || len(key) != ed25519.PrivateKeySize {
		log.Fatal("invalid key file")
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	var req signer.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		log.Fatal(err)
	}
	env, err := signer.Signer{Priv: key, Now: time.Now}.Issue(req)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(env, "", "  ")
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		log.Fatal(err)
	}
}
