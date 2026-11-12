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

	// Optional controls. A nil value disables that control.
	Kill           *Kill           // operator kill switch
	Admins         map[string]bool // principals allowed to operate the kill switch
	PrincipalLimit *Limiter        // rate limit per caller
	ServiceLimit   *Limiter        // rate limit per service, across callers
	Inspect        Inspector       // tool-output inspection
}

type decision struct {
	code   int
	action string
	detail string
	cause  string // audited, never returned to the caller
}

func deny(action, format string, a ...any) *decision {
	return &decision{code: http.StatusForbidden, action: action, detail: fmt.Sprintf(format, a...)}
}

// ServeHTTP handles POST /mcp/{name}@{version} and POST /admin/kill.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/admin/kill" {
		g.serveKill(w, r)
		return
	}
	service, ok := strings.CutPrefix(r.URL.Path, "/mcp/")
	if r.Method != http.MethodPost || !ok || service == "" {
		http.NotFound(w, r)
		return
	}
	// Authenticate before reading the body, so anonymous callers cannot make
	// the gateway parse, or the audit log record, their payload.
	principal, authErr := g.Auth.Authenticate(r)
	if authErr != nil {
		g.write(w, "anonymous", service, mcpx.Request{}, &decision{code: http.StatusUnauthorized, action: "authenticate", detail: "invalid credentials"}, nil)
		return
	}
	// Kill switch and rate limits also run before the body is read.
	if g.Kill != nil {
		if why, blocked := g.Kill.Blocked(principal, service); blocked {
			g.write(w, principal, service, mcpx.Request{}, deny("kill_switch", "%s", why), nil)
			return
		}
	}
	_, known := g.Upstreams[service] // unknown services are denied later and must not grow the limiter
	if (g.PrincipalLimit != nil && !g.PrincipalLimit.Allow(principal)) ||
		(g.ServiceLimit != nil && known && !g.ServiceLimit.Allow(service)) {
		g.write(w, principal, service, mcpx.Request{}, &decision{code: http.StatusTooManyRequests, action: "rate_limited", detail: "rate limit exceeded"}, nil)
		return
	}
	var req mcpx.Request
	body, err := io.ReadAll(io.LimitReader(r.Body, mcpx.MaxBody))
	if err == nil {
		err = json.Unmarshal(body, &req)
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
	if !g.Policy.AllowService(principal, service) {
		return nil, deny("policy", "principal has no grant for this service")
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
		p, err := strictCallParams(req.Params)
		if err != nil {
			return nil, deny("parse", "invalid tools/call params: %v", err)
		}
		// Forward exactly what was authorised, never the caller's raw bytes.
		req.Params, _ = json.Marshal(p)
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
		resp, d := g.forward(principal, service, up, req)
		if d != nil || g.Inspect == nil {
			return resp, d
		}
		// The call has already run upstream; inspection decides whether its
		// output may reach the agent. The matched text is never echoed back.
		raw, _ := json.Marshal(resp)
		if rule, block := g.Inspect.Inspect(service, p.Name, raw); block {
			return nil, deny("output_blocked", "tool output broke rule %s", rule)
		} else if rule != "" {
			if _, err := g.Audit.Append(principal, "output_flagged", "allow", clip(service+" "+p.Name+" rule "+rule, 512)); err != nil {
				return nil, &decision{code: http.StatusServiceUnavailable, action: "audit", detail: "audit unavailable"}
			}
		}
		return resp, nil

	case req.Method == "initialize":
		resp, d := g.forward(principal, service, up, req)
		if d == nil {
			stripInstructions(resp)
		}
		return resp, d

	case req.Method == "ping", strings.HasPrefix(req.Method, "notifications/"):
		return g.forward(principal, service, up, req)
	}
	return nil, deny("method", "method %q not permitted", req.Method)
}

// verifiedTools fetches the live tool list and checks it against the
// certified manifest hash. Paginated lists are refused, never hashed partially.
func (g *Gateway) verifiedTools(up Upstream, want string, id json.RawMessage) ([]manifest.Tool, *decision) {
	resp, err := mcpx.Call(g.Client, up.URL, mcpx.Request{JSONRPC: "2.0", ID: id, Method: "tools/list"})
	if err != nil {
		d := deny("upstream", "tools/list failed")
		d.cause = err.Error()
		return nil, d
	}
	if resp.Error != nil {
		d := deny("upstream", "tools/list failed")
		d.cause = resp.Error.Message
		return nil, d
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

// forward records the intent to call upstream before making the call, so no
// side effect can happen without an audit record.
func (g *Gateway) forward(principal, service string, up Upstream, req mcpx.Request) (*mcpx.Response, *decision) {
	if _, err := g.Audit.Append(principal, "forward", "allow", clip(service+" "+req.Method, 512)); err != nil {
		return nil, &decision{code: http.StatusServiceUnavailable, action: "audit", detail: "audit unavailable"}
	}
	resp, err := mcpx.Call(g.Client, up.URL, req)
	if err != nil {
		d := deny("upstream", "upstream call failed")
		d.cause = err.Error()
		return nil, d
	}
	return &resp, nil
}

// serveKill lets an admin engage or release the kill switch.
// Body: {"scope":"all|service|principal","target":"...","engaged":true}.
// Engaging always takes effect, even if it cannot be audited, because
// stopping traffic is the safe direction. Releasing requires an audit record.
func (g *Gateway) serveKill(w http.ResponseWriter, r *http.Request) {
	principal, err := g.Auth.Authenticate(r)
	if err != nil {
		g.Audit.Append("anonymous", "kill_switch", "deny", "invalid credentials")
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	if g.Kill == nil || !g.Admins[principal] {
		g.Audit.Append(clip(principal, 256), "kill_switch", "deny", "not an admin")
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var body struct {
		Scope   string `json:"scope"`
		Target  string `json:"target"`
		Engaged bool   `json:"engaged"`
	}
	b, _ := io.ReadAll(io.LimitReader(r.Body, 4<<10))
	if json.Unmarshal(b, &body) != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	verb := "release"
	if body.Engaged {
		verb = "engage"
	}
	_, auditErr := g.Audit.Append(clip(principal, 256), "kill_switch", "allow", clip(verb+" "+body.Scope+" "+body.Target, 512))
	if auditErr != nil && !body.Engaged {
		http.Error(w, "audit unavailable: kill switch not released", http.StatusServiceUnavailable)
		return
	}
	if err := g.Kill.Set(body.Scope, body.Target, body.Engaged); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if auditErr != nil {
		http.Error(w, "engaged, but the action could not be audited", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// strictCallParams accepts only {"name": string, "arguments": object?} with
// no duplicate or differently-cased keys, so the gateway and the upstream
// cannot disagree about which tool is being called.
func strictCallParams(raw json.RawMessage) (mcpx.CallParams, error) {
	if _, err := manifest.Canonicalize(raw); err != nil {
		return mcpx.CallParams{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return mcpx.CallParams{}, err
	}
	var p mcpx.CallParams
	for k, v := range fields {
		switch k {
		case "name":
			if err := json.Unmarshal(v, &p.Name); err != nil {
				return mcpx.CallParams{}, err
			}
		case "arguments":
			p.Arguments = v
		default:
			return mcpx.CallParams{}, fmt.Errorf("unexpected field %q", clip(k, 64))
		}
	}
	if p.Name == "" {
		return mcpx.CallParams{}, fmt.Errorf("missing name")
	}
	return p, nil
}

// stripInstructions removes server instructions from an initialize result:
// they reach the model's context but are not covered by the manifest hash.
func stripInstructions(resp *mcpx.Response) {
	var m map[string]json.RawMessage
	if resp == nil || json.Unmarshal(resp.Result, &m) != nil {
		return
	}
	delete(m, "instructions")
	resp.Result, _ = json.Marshal(m)
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
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
		if d.cause != "" {
			detail += " cause: " + d.cause
		}
	}
	if _, err := g.Audit.Append(clip(principal, 256), clip(action, 128), result, clip(service+" "+req.Method+" "+detail, 1024)); err != nil && d == nil {
		// Fail closed: an unauditable allow becomes a deny.
		d, resp = &decision{code: http.StatusServiceUnavailable, action: "audit", detail: "audit unavailable"}, nil
	}
	w.Header().Set("Content-Type", "application/json")
	if d != nil {
		w.WriteHeader(d.code)
		json.NewEncoder(w).Encode(mcpx.Response{JSONRPC: "2.0", ID: req.ID, Error: &mcpx.Error{Code: codeDenied, Message: d.action + ": " + d.detail}})
		return
	}
	if resp != nil && resp.JSONRPC == "" {
		w.WriteHeader(http.StatusAccepted) // notification acknowledged upstream
		return
	}
	json.NewEncoder(w).Encode(resp)
}
