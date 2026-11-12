package manifest

import (
	"encoding/json"
	"testing"
)

func TestCanonicalizeSortsKeys(t *testing.T) {
	got, err := Canonicalize([]byte(`{"b":1,"a":{"d":2,"c":3}}`))
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"a":{"c":3,"d":2},"b":1}`; string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestCanonicalizeNoHTMLEscapeAndControlChars(t *testing.T) {
	got, err := Canonicalize([]byte(`{"s":"<a>&\u0001\n"}`))
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"s":"<a>&\u0001\n"}`; string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func tools(desc string) []Tool {
	return []Tool{
		{Name: "echo", Description: desc, InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)},
		{Name: "add", Description: "Add numbers", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
}

func TestHashIgnoresOrderAndWhitespace(t *testing.T) {
	a := tools("Echo text")
	b := []Tool{
		{Name: "add", Description: "Add numbers", InputSchema: json.RawMessage("{ \"type\" : \"object\" }")},
		{Name: "echo", Description: "Echo text", InputSchema: json.RawMessage("{\n  \"properties\": {\"text\": {\"type\": \"string\"}},\n  \"type\": \"object\"\n}")},
	}
	ha, err := HashTools(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := HashTools(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Fatalf("hash differs: %s vs %s", ha, hb)
	}
}

func TestHashChangesOnDescription(t *testing.T) {
	h1, _ := HashTools(tools("Echo text"))
	h2, _ := HashTools(tools("Echo text."))
	if h1 == h2 {
		t.Fatal("hash unchanged after description change")
	}
}

func TestHashDetectsZeroWidthChar(t *testing.T) {
	h1, _ := HashTools(tools("Echo text"))
	h2, _ := HashTools(tools("Echo​ text"))
	if h1 == h2 {
		t.Fatal("zero-width char not detected")
	}
}

func TestHashRejectsDuplicateNames(t *testing.T) {
	ts := append(tools("x"), Tool{Name: "echo", Description: "dup"})
	if _, err := HashTools(ts); err == nil {
		t.Fatal("expected duplicate-name error")
	}
}

func TestHashGolden(t *testing.T) {
	h, err := HashTools([]Tool{{Name: "echo", Description: "Echo text", InputSchema: json.RawMessage(`{"type":"object"}`)}})
	if err != nil {
		t.Fatal(err)
	}
	const golden = "d574706ab20c07bcb1e9a5dde4be7e8cc657b460261a7f58a0a70ac5f79f8a84"
	if h != golden {
		t.Fatalf("golden hash = %s", h)
	}
}

func TestHashAgentCardKeyOrder(t *testing.T) {
	h1, err := HashAgentCard([]byte(`{"name":"a","url":"https://x"}`))
	if err != nil {
		t.Fatal(err)
	}
	h2, _ := HashAgentCard([]byte(`{"url":"https://x","name":"a"}`))
	if h1 != h2 {
		t.Fatal("agent card hash depends on key order")
	}
}

func TestCanonicalizeNumbersES6(t *testing.T) {
	got, err := Canonicalize([]byte(`{"n":1.0,"m":1e21,"k":0.000001,"j":1e-7,"i":-12.5}`))
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"i":-12.5,"j":1e-7,"k":0.000001,"m":1e+21,"n":1}`; string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestCanonicalizeRejectsDuplicateKeys(t *testing.T) {
	if _, err := Canonicalize([]byte(`{"type":"EVIL","type":"object"}`)); err == nil {
		t.Fatal("duplicate keys accepted")
	}
	if _, err := Canonicalize([]byte(`{"a":{"x":1,"x":2}}`)); err == nil {
		t.Fatal("nested duplicate keys accepted")
	}
}
