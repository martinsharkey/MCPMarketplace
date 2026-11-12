// Command innovox-gateway runs the customer node's MCP gateway.
// Usage: innovox-gateway -config gateway.json
//
//	{
//	  "listen": ":8443",
//	  "stateFile": "state.json",
//	  "auditFile": "audit.log",
//	  "auditExportFile": "audit-export.jsonl",   // optional: JSON lines for a SIEM forwarder
//	  "checkpointKeyFile": "audit.key",          // optional: signs a checkpoint every checkpointMinutes
//	  "checkpointMinutes": 5,
//	  "graceHours": 168,
//	  "killFile": "kill.json",
//	  "admins": ["alice"],
//	  "principalRate": {"perSecond": 20, "burst": 40},
//	  "serviceRate": {"perSecond": 100, "burst": 200},
//	  "blockMarkings": ["SECRET", "TOP SECRET"],
//	  "inspectWarnOnly": false,
//	  "upstreams": [{"Service": "uk.example/echo@1.0.0", "URL": "http://echo:8080/mcp", "ImageDigest": "sha256:..."}],
//	  "tokens": {"<token>": "alice"},
//	  "grants": {"alice": {"uk.example/echo@1.0.0": ["*"]}}
//	}
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"innovox.uk/assured/node/audit"
	"innovox.uk/assured/node/gateway"
	"innovox.uk/assured/node/syncer"
)

type rate struct {
	PerSecond float64 `json:"perSecond"`
	Burst     int     `json:"burst"`
}

type config struct {
	Listen            string               `json:"listen"`
	StateFile         string               `json:"stateFile"`
	AuditFile         string               `json:"auditFile"`
	AuditExportFile   string               `json:"auditExportFile"`
	CheckpointKeyFile string               `json:"checkpointKeyFile"`
	CheckpointMinutes int                  `json:"checkpointMinutes"`
	GraceHours        int                  `json:"graceHours"`
	KillFile          string               `json:"killFile"`
	Admins            []string             `json:"admins"`
	PrincipalRate     rate                 `json:"principalRate"`
	ServiceRate       rate                 `json:"serviceRate"`
	BlockMarkings     []string             `json:"blockMarkings"`
	InspectWarnOnly   bool                 `json:"inspectWarnOnly"`
	Upstreams         []gateway.Upstream   `json:"upstreams"`
	Tokens            gateway.StaticTokens `json:"tokens"`
	Grants            gateway.Grants       `json:"grants"`
}

func main() {
	path := flag.String("config", "gateway.json", "config file")
	flag.Parse()
	b, err := os.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}
	c := config{GraceHours: 168, CheckpointMinutes: 5, KillFile: "kill.json",
		PrincipalRate: rate{20, 40}, ServiceRate: rate{100, 200}}
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
	}
	lg, err := audit.Open(c.AuditFile)
	if err != nil {
		log.Fatal(err)
	}
	if c.AuditExportFile != "" {
		f, err := os.OpenFile(c.AuditExportFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			log.Fatal(err)
		}
		lg.Sink = f
	}
	if c.CheckpointKeyFile != "" {
		kb, err := os.ReadFile(c.CheckpointKeyFile)
		if err != nil {
			log.Fatal(err)
		}
		key, err := hex.DecodeString(strings.TrimSpace(string(kb)))
		if err != nil || len(key) != ed25519.PrivateKeySize {
			log.Fatal("invalid checkpoint key")
		}
		go func() {
			for range time.Tick(time.Duration(c.CheckpointMinutes) * time.Minute) {
				if _, err := lg.Checkpoint(key); err != nil {
					log.Printf("checkpoint: %v", err)
				}
			}
		}()
	}
	// An unreadable kill file stops the gateway from starting: it must not
	// run without knowing whether an operator has stopped traffic.
	kill, err := gateway.LoadKill(c.KillFile)
	if err != nil {
		log.Fatal(err)
	}
	ups := map[string]gateway.Upstream{}
	for _, u := range c.Upstreams {
		ups[u.Service] = u
	}
	admins := map[string]bool{}
	for _, a := range c.Admins {
		admins[a] = true
	}
	inspector := gateway.DefaultInspector(c.BlockMarkings...)
	inspector.WarnOnly = c.InspectWarnOnly
	g := &gateway.Gateway{
		// State is re-read per request so a sync takes effect immediately;
		// an unreadable state file yields nil, which the gateway denies.
		State:     func() *syncer.State { st, _ := syncer.Load(c.StateFile); return st },
		Upstreams: ups, Auth: c.Tokens, Policy: c.Grants, Audit: lg,
		Grace: time.Duration(c.GraceHours) * time.Hour, Now: time.Now,
		Client:         &http.Client{Timeout: 30 * time.Second},
		Kill:           kill,
		Admins:         admins,
		PrincipalLimit: &gateway.Limiter{Rate: c.PrincipalRate.PerSecond, Burst: c.PrincipalRate.Burst},
		ServiceLimit:   &gateway.Limiter{Rate: c.ServiceRate.PerSecond, Burst: c.ServiceRate.Burst},
		Inspect:        inspector,
	}
	srv := &http.Server{Addr: c.Listen, Handler: g, ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 2 * time.Minute}
	log.Printf("innovox-gateway listening on %s", c.Listen)
	log.Fatal(srv.ListenAndServe())
}
