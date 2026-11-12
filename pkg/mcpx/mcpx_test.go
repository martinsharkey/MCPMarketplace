package mcpx

import (
	"encoding/json"
	"innovox.uk/assured/internal/conformance"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{JSONRPC: "2.0", ID: req.ID, Result: json.RawMessage(`{"tools":[{"name":"echo","description":"Echo"}],"nextCursor":"n1"}`)})
	}))
	defer srv.Close()
	resp, err := Call(srv.Client(), srv.URL, Request{JSONRPC: "2.0", ID: json.RawMessage(`7`), Method: "tools/list"})
	if err != nil {
		t.Fatal(err)
	}
	var tl ToolsListResult
	if err := json.Unmarshal(resp.Result, &tl); err != nil || tl.Tools[0].Name != "echo" || tl.NextCursor != "n1" || string(resp.ID) != "7" {
		t.Fatalf("bad decode %v %+v", err, tl)
	}
}

func TestCallRejectsNonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {}\n\n")
	}))
	defer srv.Close()
	if _, err := Call(srv.Client(), srv.URL, Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list"}); err == nil {
		t.Fatal("SSE response accepted")
	}
}

func TestCallDoesNotFollowRedirects(t *testing.T) {
	conformance.Covers(t, "IAS-RT-01")
	hit := false
	inner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true }))
	defer inner.Close()
	outer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, inner.URL, http.StatusTemporaryRedirect)
	}))
	defer outer.Close()
	if _, err := Call(outer.Client(), outer.URL, Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "ping"}); err == nil || hit {
		t.Fatalf("redirect followed: err=%v hit=%v", err, hit)
	}
}
