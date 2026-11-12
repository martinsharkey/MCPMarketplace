// Package mcpx is a minimal MCP JSON-RPC client covering what the gateway
// needs. It is replaced by the official MCP Go SDK in Plan 5.
package mcpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	"innovox.uk/assured/pkg/manifest"
)

const MaxBody = 4 << 20

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolsListResult struct {
	Tools      []manifest.Tool `json:"tools"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

type CallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Call POSTs req to url and decodes a JSON (not SSE) JSON-RPC response.
func Call(c *http.Client, url string, req Request) (Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	hr, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("Accept", "application/json")
	resp, err := c.Do(hr)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	if mt, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); mt != "application/json" {
		return Response{}, fmt.Errorf("mcpx: unsupported response type %q", mt)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, MaxBody))
	if err != nil {
		return Response{}, err
	}
	var out Response
	if err := json.Unmarshal(b, &out); err != nil {
		return Response{}, fmt.Errorf("mcpx: decode: %w", err)
	}
	return out, nil
}
