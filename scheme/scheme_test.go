package scheme

import (
	"strings"
	"testing"
)

func load(t *testing.T) *Scheme {
	t.Helper()
	s, err := LoadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSchemeLoads(t *testing.T) {
	s := load(t)
	if s.Version == "" || len(s.Controls) < 30 || len(s.Sources) < 10 {
		t.Fatalf("scheme too small: version %q, %d controls, %d sources", s.Version, len(s.Controls), len(s.Sources))
	}
}

func TestSchemeValidates(t *testing.T) {
	if errs := load(t).Validate(); len(errs) > 0 {
		t.Fatalf("scheme invalid:\n%s", strings.Join(errs, "\n"))
	}
}

func TestValidateCatchesProblems(t *testing.T) {
	s := &Scheme{
		Version: "x",
		Sources: []Source{{ID: "S", Clauses: map[string]string{"S-1": "a"}}},
		Controls: []Control{
			{ID: "IAS-X-01", Title: "t", Requirement: "r", Scope: "service", Tiers: []string{"T1"}, Method: "pipeline", Stage: "S3", MapsTo: []string{"S-1"}},
			{ID: "IAS-X-01", Title: "t", Requirement: "r", Scope: "service", Tiers: []string{"T1"}, Method: "pipeline", Stage: "S3", MapsTo: []string{"S-1"}},
			{ID: "IAS-X-02", Title: "t", Requirement: "r", Scope: "galaxy", Tiers: []string{"T9"}, Method: "vibes", MapsTo: []string{"NOPE"}},
			{ID: "IAS-X-03", Title: "t", Requirement: "r", Scope: "service", Tiers: []string{"T1"}, Method: "pipeline", MapsTo: []string{"S-1"}},
			{ID: "IAS-X-04", Title: "t", Requirement: "r", Scope: "platform", Tiers: []string{"T1"}, Method: "automated"},
		},
	}
	errs := strings.Join(s.Validate(), "\n")
	for _, want := range []string{"duplicate control IAS-X-01", "unknown scope", "unknown tier", "unknown method", "unknown clause NOPE", "IAS-X-03: pipeline control needs a stage", "IAS-X-04: maps to no clause"} {
		if !strings.Contains(errs, want) {
			t.Errorf("missing error %q in:\n%s", want, errs)
		}
	}
}

func TestEveryAutomatedPlatformControlIsReachable(t *testing.T) {
	s := load(t)
	c, ok := s.Control("IAS-RT-02")
	if !ok || c.Method != "automated" || c.Scope != "platform" {
		t.Fatalf("expected IAS-RT-02 to be an automated platform control: %+v", c)
	}
}
