package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// Kill is the operator's kill switch. It can stop all traffic, one service
// or one principal (an agent or a person). With a path it survives restarts.
type Kill struct {
	mu    sync.RWMutex
	path  string
	state killState
}

type killState struct {
	All        bool            `json:"all"`
	Services   map[string]bool `json:"services"`
	Principals map[string]bool `json:"principals"`
}

// NewKill returns an in-memory kill switch with nothing engaged.
func NewKill() *Kill {
	return &Kill{state: killState{Services: map[string]bool{}, Principals: map[string]bool{}}}
}

// LoadKill reads persisted kill state. A missing file means nothing is
// engaged; an unreadable or corrupt file is an error, so the caller can
// refuse to start rather than run with an unknown kill state.
func LoadKill(path string) (*Kill, error) {
	k := NewKill()
	k.path = path
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return k, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &k.state); err != nil {
		return nil, fmt.Errorf("kill state %s: %w", path, err)
	}
	if k.state.Services == nil {
		k.state.Services = map[string]bool{}
	}
	if k.state.Principals == nil {
		k.state.Principals = map[string]bool{}
	}
	return k, nil
}

// Blocked reports whether a call is stopped, and by which scope.
func (k *Kill) Blocked(principal, service string) (string, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	switch {
	case k.state.All:
		return "all traffic stopped", true
	case k.state.Services[service]:
		return "service stopped", true
	case k.state.Principals[principal]:
		return "principal stopped", true
	}
	return "", false
}

// Set engages or releases the switch for scope "all", "service" or "principal".
func (k *Kill) Set(scope, target string, engaged bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	switch scope {
	case "all":
		k.state.All = engaged
	case "service", "principal":
		if target == "" {
			return fmt.Errorf("kill: scope %q needs a target", scope)
		}
		m := k.state.Services
		if scope == "principal" {
			m = k.state.Principals
		}
		if engaged {
			m[target] = true
		} else {
			delete(m, target)
		}
	default:
		return fmt.Errorf("kill: unknown scope %q", scope)
	}
	if k.path == "" {
		return nil
	}
	b, _ := json.Marshal(k.state)
	tmp := k.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, k.path)
}
