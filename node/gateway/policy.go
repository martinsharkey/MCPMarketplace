package gateway

import (
	"errors"
	"net/http"
	"slices"
	"strings"
)

// Authenticator identifies the caller. Replaced by OIDC in Plan 5.
type Authenticator interface {
	Authenticate(r *http.Request) (principal string, err error)
}

// StaticTokens maps bearer tokens to principals (pilot only).
type StaticTokens map[string]string

func (s StaticTokens) Authenticate(r *http.Request) (string, error) {
	tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if p, found := s[tok]; ok && found && tok != "" {
		return p, nil
	}
	return "", errors.New("unauthenticated")
}

// Policy decides whether principal may use tool on service. Replaced by OPA in Plan 5.
type Policy interface {
	Allow(principal, service, tool string) bool
}

// Grants maps principal -> service -> tools; "*" grants every tool.
type Grants map[string]map[string][]string

func (g Grants) Allow(principal, service, tool string) bool {
	tools := g[principal][service]
	return slices.Contains(tools, "*") || slices.Contains(tools, tool)
}
