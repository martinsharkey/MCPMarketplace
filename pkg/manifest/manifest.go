package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Tool is one entry of an MCP tools/list result.
type Tool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"inputSchema,omitempty"`
	OutputSchema json.RawMessage `json:"outputSchema,omitempty"`
	Annotations  json.RawMessage `json:"annotations,omitempty"`
}

// HashTools returns the lowercase hex SHA-256 of the canonical tool manifest.
// Tool order and JSON formatting do not affect the result.
func HashTools(tools []Tool) (string, error) {
	sorted := append([]Tool(nil), tools...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Name == sorted[i-1].Name {
			return "", fmt.Errorf("manifest: duplicate tool name %q", sorted[i].Name)
		}
	}
	raw, err := json.Marshal(struct {
		Tools []Tool `json:"tools"`
	}{sorted})
	if err != nil {
		return "", err
	}
	return hashCanonical(raw)
}

// HashAgentCard returns the lowercase hex SHA-256 of a canonical A2A Agent Card.
func HashAgentCard(raw []byte) (string, error) { return hashCanonical(raw) }

func hashCanonical(raw []byte) (string, error) {
	c, err := Canonicalize(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(c)
	return hex.EncodeToString(sum[:]), nil
}
