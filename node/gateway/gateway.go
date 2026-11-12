// Package gateway is the customer node's single MCP entry point. Every
// decision fails closed and is written to the audit log.
package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"innovox.uk/assured/node/audit"
	"innovox.uk/assured/node/syncer"
	"innovox.uk/assured/pkg/manifest"
	"innovox.uk/assured/pkg/mcpx"
)

const codeDenied = -32001

// Upstream is a running certified service. ImageDigest is what the runtime
// reports for the running container (sha256:<hex>).
type Upstream struct {
	Service     string
	URL         string
	ImageDigest string
}

type Gateway struct {
	State     func() *syncer.State
	Upstreams map[string]Upstream
	Auth      Authenticator
	Policy    Policy
	Audit     *audit.Log
	Grace     time.Duration
	Now       func() time.Time
	Client    *http.Client
}

type decision struct {
	code   int
	action string
	detail string
}

func deny(action, format string, a ...any) *decision {
	return &decision{code: http.StatusForbidden, action: action, detail: fmt.Sprintf(format, a...)}
}

// ServeHTTP handles POST /mcp/{name}@{version}.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, ok := strings.CutPrefix(r.URL.Path, "/mcp/")
	if r.Method != http.MethodPost || !ok || service == "" {
		http.NotFound(w, r)
		return
	}
	var req mcpx.Request
	body, err := io.ReadAll(io.LimitReader(r.Body, mcpx.MaxBody))
	if err == nil {
		err = json.Unmarshal(body, &req)
	}
	principal, authErr := g.Auth.Authenticate(r)
	if authErr != nil {
		g.write(w, "anonymous", service, req, &decision{code: http.StatusUnauthorized, action: "authenticate", detail: "invalid credentials"}, nil)
		return
	}
	if err != nil || req.JSONRPC != "2.0" || req.Method == "" {
		g.write(w, principal, service, req, deny("parse", "malformed JSON-RPC request"), nil)
		return
	}
	resp, d := g.handle(principal, service, req)
	g.write(w, principal, service, req, d, resp)
}

func (g *Gateway) handle(principal, service string, req mcpx.Request) (*mcpx.Response, *decision) {
	now := g.Now()
	st := g.State()
	if st == nil || st.Stale(now, g.Grace) {
		return nil, deny("stale_feed", "feed older than grace period")
	}
	listing, ok := st.Listings[service]
	if !ok {
		return nil, deny("not_listed", "service has no valid certificate")
	}
	if listing.Revoked {
		return nil, deny("revoked", "certificate revoked: %s", listing.RevokeReason)
	}
	cert := listing.Cert
	if !now.Before(cert.Predicate.ExpiresAt) {
		return nil, deny("expired", "certificate expired %s", cert.Predicate.ExpiresAt.Format(time.RFC3339))
	}
	up, ok := g.Upstreams[service]
	if !ok {
		return nil, deny("no_upstream", "service not running on this node")
	}
	if len(cert.Subject) != 1 || !subjectMatches(cert.Subject[0].Name, cert.Subject[0].Digest["sha256"], up) {
		return nil, deny("digest_mismatch", "running artefact does not match certificate")
	}

	switch {
	case req.Method == "tools/list":
		tools, d := g.verifiedTools(up, cert.Predicate.ManifestSHA256, req.ID)
		if d != nil {
			return nil, d
		}
		allowed := []manifest.Tool{}
		for _, t := range tools {
			if g.Policy.Allow(principal, service, t.Name) {
				allowed = append(allowed, t)
			}
		}
		raw, _ := json.Marshal(mcpx.ToolsListResult{Tools: allowed})
		return &mcpx.Response{JSONRPC: "2.0", ID: req.ID, Result: raw}, nil

	case req.Method == "tools/call":
		var p mcpx.CallParams
		if err := json.Unmarshal(req.Params, &p); err != nil || p.Name == "" {
			return nil, deny("parse", "invalid tools/call params")
		}
		tools, d := g.verifiedTools(up, cert.Predicate.ManifestSHA256, json.RawMessage(`"gw-verify"`))
		if d != nil {
			return nil, d
		}
		if !hasTool(tools, p.Name) {
			return nil, deny("unlisted_tool", "tool %q is not in the certified manifest", p.Name)
		}
		if !g.Policy.Allow(principal, service, p.Name) {
			return nil, deny("policy", "principal not granted tool %q", p.Name)
		}
		return g.forward(up, req)

	case req.Method == "initialize", req.Method == "ping", strings.HasPrefix(req.Method, "notifications/"):
		return g.forward(up, req)
	}
	return nil, deny("method", "method %q not permitted", req.Method)
}

// verifiedTools fetches the live tool list and checks it against the
// certified manifest hash. Paginated lists are refused, never hashed partially.
func (g *Gateway) verifiedTools(up Upstream, want string, id json.RawMessage) ([]manifest.Tool, *decision) {
	resp, err := mcpx.Call(g.Client, up.URL, mcpx.Request{JSONRPC: "2.0", ID: id, Method: "tools/list"})
	if err != nil || resp.Error != nil {
		return nil, deny("upstream", "tools/list failed")
	}
	var tl mcpx.ToolsListResult
	if err := json.Unmarshal(resp.Result, &tl); err != nil {
		return nil, deny("upstream", "tools/list undecodable")
	}
	if tl.NextCursor != "" {
		return nil, deny("paginated", "paginated tools/list not supported")
	}
	got, err := manifest.HashTools(tl.Tools)
	if err != nil || got != want {
		return nil, deny("rug_pull", "tool manifest %s does not match certified %s", got, want)
	}
	return tl.Tools, nil
}

func (g *Gateway) forward(up Upstream, req mcpx.Request) (*mcpx.Response, *decision) {
	resp, err := mcpx.Call(g.Client, up.URL, req)
	if err != nil {
		return nil, deny("upstream", "upstream call failed")
	}
	return &resp, nil
}

func subjectMatches(name, digest string, up Upstream) bool {
	if remote, ok := strings.CutPrefix(name, "remote:"); ok {
		return remote == up.URL
	}
	return "sha256:"+digest == up.ImageDigest
}

func hasTool(tools []manifest.Tool, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

func (g *Gateway) write(w http.ResponseWriter, principal, service string, req mcpx.Request, d *decision, resp *mcpx.Response) {
	action := req.Method
	result, detail := "allow", ""
	if d != nil {
		result, action, detail = "deny", d.action, d.detail
	}
	if _, err := g.Audit.Append(principal, action, result, service+" "+req.Method+" "+detail); err != nil && d == nil {
		// Fail closed: an unauditable allow becomes a deny.
		d, resp = &decision{code: http.StatusServiceUnavailable, action: "audit", detail: "audit unavailable"}, nil
	}
	w.Header().Set("Content-Type", "application/json")
	if d != nil {
		w.WriteHeader(d.code)
		json.NewEncoder(w).Encode(mcpx.Response{JSONRPC: "2.0", ID: req.ID, Error: &mcpx.Error{Code: codeDenied, Message: d.action + ": " + d.detail}})
		return
	}
	json.NewEncoder(w).Encode(resp)
}
