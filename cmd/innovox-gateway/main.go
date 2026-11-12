// Command innovox-gateway runs the customer node's MCP gateway.
// Usage: innovox-gateway -config gateway.json
//
//	{"listen":":8443","stateFile":"state.json","auditFile":"audit.log","graceHours":168,
//	 "upstreams":[{"Service":"uk.example/echo@1.0.0","URL":"http://echo:8080/mcp","ImageDigest":"sha256:..."}],
//	 "tokens":{"<token>":"alice"},"grants":{"alice":{"uk.example/echo@1.0.0":["*"]}}}
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"innovox.uk/assured/node/audit"
	"innovox.uk/assured/node/gateway"
	"innovox.uk/assured/node/syncer"
)

type config struct {
	Listen     string               `json:"listen"`
	StateFile  string               `json:"stateFile"`
	AuditFile  string               `json:"auditFile"`
	GraceHours int                  `json:"graceHours"`
	Upstreams  []gateway.Upstream   `json:"upstreams"`
	Tokens     gateway.StaticTokens `json:"tokens"`
	Grants     gateway.Grants       `json:"grants"`
}

func main() {
	path := flag.String("config", "gateway.json", "config file")
	flag.Parse()
	b, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}
	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
	}
	if c.GraceHours == 0 {
		c.GraceHours = 168
	}
	lg, err := audit.Open(c.AuditFile)
	if err != nil {
		log.Fatal(err)
	}
	ups := map[string]gateway.Upstream{}
	for _, u := range c.Upstreams {
		ups[u.Service] = u
	}
	g := &gateway.Gateway{
		// State is re-read per request so a sync takes effect immediately;
		// an unreadable state file yields nil, which the gateway denies.
		State:     func() *syncer.State { st, _ := syncer.Load(c.StateFile); return st },
		Upstreams: ups, Auth: c.Tokens, Policy: c.Grants, Audit: lg,
		Grace: time.Duration(c.GraceHours) * time.Hour, Now: time.Now,
		Client: &http.Client{Timeout: 30 * time.Second},
	}
	log.Printf("innovox-gateway listening on %s", c.Listen)
	log.Fatal(http.ListenAndServe(c.Listen, g))
}
