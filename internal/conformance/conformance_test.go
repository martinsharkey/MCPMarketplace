package conformance

import (
	"strings"
	"testing"

	"innovox.uk/assured/scheme"
)

const events = `{"Action":"run","Package":"p/gw","Test":"TestA"}
{"Action":"output","Package":"p/gw","Test":"TestA","Output":"    gw_test.go:10: innovox-control: IAS-RT-02\n"}
{"Action":"output","Package":"p/gw","Test":"TestA","Output":"    gw_test.go:10: innovox-control: IAS-RT-03\n"}
{"Action":"pass","Package":"p/gw","Test":"TestA"}
{"Action":"run","Package":"p/gw","Test":"TestB"}
{"Action":"output","Package":"p/gw","Test":"TestB","Output":"    gw_test.go:20: innovox-control: IAS-RT-03\n"}
{"Action":"fail","Package":"p/gw","Test":"TestB"}
{"Action":"output","Package":"p/gw","Output":"ok\n"}
`

func sch() *scheme.Scheme {
	return &scheme.Scheme{Version: "v0.1", Controls: []scheme.Control{
		{ID: "IAS-RT-02", Family: "Run-time", Title: "Rug pulls", Scope: "platform", Method: "automated"},
		{ID: "IAS-RT-03", Family: "Run-time", Title: "Hidden tools", Scope: "platform", Method: "automated"},
		{ID: "IAS-RT-09", Family: "Run-time", Title: "Output inspection", Scope: "platform", Method: "automated"},
		{ID: "IAS-RT-10", Family: "Run-time", Title: "HITL", Scope: "platform", Method: "automated"},
		{ID: "IAS-SC-02", Family: "Supply chain", Title: "Provenance", Scope: "service", Method: "pipeline", Stage: "S2"},
	}}
}

func TestParseEventsCollectsControls(t *testing.T) {
	ev, err := ParseEvents(strings.NewReader(events))
	if err != nil {
		t.Fatal(err)
	}
	if got := ev["IAS-RT-02"]; len(got) != 1 || got[0].Name != "p/gw.TestA" || !got[0].Passed {
		t.Fatalf("RT-02: %+v", got)
	}
	if got := ev["IAS-RT-03"]; len(got) != 2 || got[1].Passed {
		t.Fatalf("RT-03: %+v", got)
	}
}

func TestStatusRules(t *testing.T) {
	ev, _ := ParseEvents(strings.NewReader(events))
	assess := []Assessment{
		{Control: "IAS-RT-09", Status: "partial", Rationale: "size limit only"},
		{Control: "IAS-SC-02", Status: "not-yet-testable", Rationale: "pipeline not built"},
	}
	r := Build(sch(), ev, assess)
	want := map[string]string{
		"IAS-RT-02": "met",
		"IAS-RT-03": "failing",
		"IAS-RT-09": "partial",
		"IAS-RT-10": "not-met",
		"IAS-SC-02": "not-yet-testable",
	}
	for id, st := range want {
		if got := r.Status(id); got != st {
			t.Errorf("%s: got %q want %q", id, got, st)
		}
	}
}

func TestMetClaimWithoutTestEvidenceIsRejected(t *testing.T) {
	r := Build(sch(), Evidence{}, []Assessment{{Control: "IAS-RT-10", Status: "met", Rationale: "trust me"}})
	if r.Status("IAS-RT-10") != "not-met" || len(r.Problems) == 0 {
		t.Fatalf("unsupported met claim accepted: %q %v", r.Status("IAS-RT-10"), r.Problems)
	}
}

func TestAssessmentForUnknownControlIsAProblem(t *testing.T) {
	r := Build(sch(), Evidence{}, []Assessment{{Control: "IAS-NOPE-01", Status: "met"}})
	if len(r.Problems) == 0 {
		t.Fatal("unknown control accepted")
	}
}

func TestRenderSummarisesAndListsEvidence(t *testing.T) {
	ev, _ := ParseEvents(strings.NewReader(events))
	md := Build(sch(), ev, nil).Markdown()
	for _, want := range []string{"| met | 1 |", "| failing | 1 |", "IAS-RT-02", "p/gw.TestA", "IAS-SC-02"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}

func TestRenderStandardsCoverage(t *testing.T) {
	s := sch()
	s.Sources = []scheme.Source{{ID: "OWASP-MCP", Title: "OWASP MCP Top 10", Clauses: map[string]string{"MCP03": "Tool Poisoning", "MCP08": "Lack of Audit and Telemetry"}}}
	s.Controls[0].MapsTo = []string{"MCP03"}
	ev, _ := ParseEvents(strings.NewReader(events))
	md := Build(s, ev, nil).Markdown()
	for _, want := range []string{"## Coverage by standard", "### OWASP MCP Top 10", "| MCP03 | Tool Poisoning | IAS-RT-02 (met) |", "| MCP08 | Lack of Audit and Telemetry | — |"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in:\n%s", want, md)
		}
	}
}

func TestEvidenceOrderIsStable(t *testing.T) {
	const a = `{"Action":"output","Package":"z","Test":"T","Output":"innovox-control: IAS-RT-02\n"}
{"Action":"pass","Package":"z","Test":"T"}
{"Action":"output","Package":"a","Test":"T","Output":"innovox-control: IAS-RT-02\n"}
{"Action":"pass","Package":"a","Test":"T"}
`
	ev, _ := ParseEvents(strings.NewReader(a))
	if got := ev["IAS-RT-02"]; len(got) != 2 || got[0].Name != "a.T" {
		t.Fatalf("evidence not sorted: %+v", got)
	}
}
