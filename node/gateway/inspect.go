package gateway

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Inspector examines a tool result before it reaches the agent.
type Inspector interface {
	// Inspect returns the name of the first rule the result breaks ("" if
	// none) and whether the result must be blocked.
	Inspect(service, tool string, result json.RawMessage) (rule string, block bool)
}

// Rule is one named pattern.
type Rule struct {
	Name string
	Re   *regexp.Regexp
}

// PatternInspector matches rules against every string in a tool result.
// It is a heuristic first line of defence: it catches common injection
// phrasing, hidden characters, leaked keys and classification markings,
// and it can be evaded by paraphrase or encoding.
type PatternInspector struct {
	Rules    []Rule
	WarnOnly bool // flag and audit instead of blocking
}

// DefaultInspector returns the built-in rules. blockMarkings lists
// classification markings that must not pass through this node, for example
// "SECRET" and "TOP SECRET" on a node accredited for OFFICIAL.
func DefaultInspector(blockMarkings ...string) *PatternInspector {
	p := &PatternInspector{Rules: []Rule{
		{"injection-override", regexp.MustCompile(`(?is)\b(ignore|disregard|forget|override)\b.{0,40}\b(previous|prior|above|earlier|preceding)\b.{0,20}\b(instructions?|prompts?|messages?|directions?)`)},
		{"injection-role", regexp.MustCompile(`(?im)^\s*(system|assistant|developer)\s*:|</?(system|assistant)>|\[/?INST\]`)},
		{"injection-concealment", regexp.MustCompile(`(?is)\bdo\s+not\s+(tell|inform|mention|reveal|alert)\b.{0,30}\b(user|human|operator)`)},
		// Zero-width space, word joiner, bidirectional overrides and isolates,
		// and Unicode tag characters. Joiners and directional marks used by
		// real scripts and emoji are deliberately not included.
		{"invisible-characters", regexp.MustCompile(`[\x{200B}\x{2060}\x{202A}-\x{202E}\x{2066}-\x{2069}\x{E0000}-\x{E007F}]`)},
		{"secret-private-key", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	}}
	if len(blockMarkings) > 0 {
		quoted := make([]string, len(blockMarkings))
		for i, m := range blockMarkings {
			quoted[i] = regexp.QuoteMeta(m)
		}
		// Case-sensitive and whole-word, so "secretary" or "secret" in prose
		// does not match the protective marking "SECRET".
		p.Rules = append(p.Rules, Rule{"classification-marking", regexp.MustCompile(`\b(` + strings.Join(quoted, "|") + `)\b`)})
	}
	return p
}

func (p *PatternInspector) Inspect(_, _ string, result json.RawMessage) (string, bool) {
	var v any
	if err := json.Unmarshal(result, &v); err != nil {
		return "undecodable-result", !p.WarnOnly
	}
	if rule := p.walk(v); rule != "" {
		return rule, !p.WarnOnly
	}
	return "", false
}

func (p *PatternInspector) walk(v any) string {
	switch t := v.(type) {
	case string:
		for _, r := range p.Rules {
			if r.Re.MatchString(t) {
				return r.Name
			}
		}
	case []any:
		for _, e := range t {
			if rule := p.walk(e); rule != "" {
				return rule
			}
		}
	case map[string]any:
		for k, e := range t {
			if rule := p.walk(k); rule != "" {
				return rule
			}
			if rule := p.walk(e); rule != "" {
				return rule
			}
		}
	}
	return ""
}
