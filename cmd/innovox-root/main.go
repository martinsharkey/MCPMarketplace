// Command innovox-root performs the key operations of the Innovox trust
// hierarchy. In production the root key lives in an offline HSM and these
// steps are a witnessed ceremony; this tool is for development and pilots.
//
//	innovox-root keygen     -out NAME                  writes NAME.key and NAME.pub (hex)
//	innovox-root delegate   -root root.key -signer signer.pub -days 90 -out delegation.json
//	innovox-root revoke-key -root root.key -keyid ID -reason TEXT -out revocation.json
//	innovox-root keyid      -pub NAME.pub              prints the key ID
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"innovox.uk/assured/pkg/attest"
	"innovox.uk/assured/pkg/feed"
	"innovox.uk/assured/pkg/trust"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: innovox-root keygen|delegate|revoke-key|keyid [flags]")
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	out := fs.String("out", "", "output file (or name prefix for keygen)")
	root := fs.String("root", "", "root private key file")
	signer := fs.String("signer", "", "signer public key file")
	pub := fs.String("pub", "", "public key file")
	days := fs.Int("days", 90, "delegation lifetime in days")
	keyID := fs.String("keyid", "", "key ID to revoke")
	reason := fs.String("reason", "", "reason for revocation")
	fs.Parse(os.Args[2:])

	switch os.Args[1] {
	case "keygen":
		p, s, err := ed25519.GenerateKey(rand.Reader)
		check(err)
		check(os.WriteFile(*out+".key", []byte(hex.EncodeToString(s)+"\n"), 0o600))
		check(os.WriteFile(*out+".pub", []byte(hex.EncodeToString(p)+"\n"), 0o644))
		fmt.Println(trust.KeyID(p))
	case "delegate":
		env := trust.Delegate(readKey(*root, ed25519.PrivateKeySize), readKey(*signer, ed25519.PublicKeySize), time.Now().AddDate(0, 0, *days))
		write(*out, env)
	case "revoke-key":
		if *keyID == "" {
			log.Fatal("-keyid is required")
		}
		write(*out, trust.RevokeKey(readKey(*root, ed25519.PrivateKeySize), *keyID, *reason, time.Now()))
	case "keyid":
		fmt.Println(trust.KeyID(readKey(*pub, ed25519.PublicKeySize)))
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func readKey(path string, size int) []byte {
	b, err := os.ReadFile(path)
	check(err)
	k, err := hex.DecodeString(strings.TrimSpace(string(b)))
	if err != nil || len(k) != size {
		log.Fatalf("%s: not a valid key", path)
	}
	return k
}

func write(path string, env attest.Envelope) { check(feed.WriteBundle(path, env)) }

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
