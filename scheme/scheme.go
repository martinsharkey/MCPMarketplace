// Package scheme loads the Innovox Assured scheme: the controls Innovox tests
// against and the verified source clauses each control maps to.
package scheme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
)

type Source struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Publisher string            `json:"publisher"`
	Version   string            `json:"version"`
	URL       string            `json:"url"`
	Verified  string            `json:"verified"`
	Note      string            `json:"note,omitempty"`
	Clauses   map[string]string `json:"clauses"`
}

type Control struct {
	ID          string   `json:"id"`
	Family      string   `json:"family"`
	Title       string   `json:"title"`
	Requirement string   `json:"requirement"`
	Scope       string   `json:"scope"`  // service | platform
	Tiers       []string `json:"tiers"`  // T1 | T2 | T3
	Method      string   `json:"method"` // automated | pipeline | document | manual
	Stage       string   `json:"stage,omitempty"`
	MapsTo      []string `json:"maps_to"`
}

type Scheme struct {
	Name     string    `json:"scheme"`
	Version  string    `json:"version"`
	Status   string    `json:"status"`
	Claim    string    `json:"claim"`
	Controls []Control `json:"controls"`
	Sources  []Source  `json:"-"`
}

var (
	scopes  = []string{"service", "platform"}
	tiers   = []string{"T1", "T2", "T3"}
	methods = []string{"automated", "pipeline", "document", "manual"}
	stages  = []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8"}
)

// LoadDir reads controls.json and sources.json from dir.
func LoadDir(dir string) (*Scheme, error) {
	var s Scheme
	if err := readJSON(filepath.Join(dir, "controls.json"), &s); err != nil {
		return nil, err
	}
	var src struct {
		Sources []Source `json:"sources"`
	}
	if err := readJSON(filepath.Join(dir, "sources.json"), &src); err != nil {
		return nil, err
	}
	s.Sources = src.Sources
	return &s, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// Clause returns the text of a source clause and its source.
func (s *Scheme) Clause(id string) (string, Source, bool) {
	for _, src := range s.Sources {
		if t, ok := src.Clauses[id]; ok {
			return t, src, true
		}
	}
	return "", Source{}, false
}

func (s *Scheme) Control(id string) (Control, bool) {
	for _, c := range s.Controls {
		if c.ID == id {
			return c, true
		}
	}
	return Control{}, false
}

// Validate returns every structural problem in the scheme, sorted.
func (s *Scheme) Validate() []string {
	var errs []string
	add := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	clauseOwner := map[string]string{}
	for _, src := range s.Sources {
		for id := range src.Clauses {
			if prev, dup := clauseOwner[id]; dup {
				add("clause %s defined by both %s and %s", id, prev, src.ID)
			}
			clauseOwner[id] = src.ID
		}
	}
	seen := map[string]bool{}
	for _, c := range s.Controls {
		if seen[c.ID] {
			add("duplicate control %s", c.ID)
		}
		seen[c.ID] = true
		if c.Title == "" || c.Requirement == "" {
			add("%s: missing title or requirement", c.ID)
		}
		if !slices.Contains(scopes, c.Scope) {
			add("%s: unknown scope %q", c.ID, c.Scope)
		}
		if len(c.Tiers) == 0 {
			add("%s: no tiers", c.ID)
		}
		for _, t := range c.Tiers {
			if !slices.Contains(tiers, t) {
				add("%s: unknown tier %q", c.ID, t)
			}
		}
		if !slices.Contains(methods, c.Method) {
			add("%s: unknown method %q", c.ID, c.Method)
		}
		if c.Method == "pipeline" && !slices.Contains(stages, c.Stage) {
			add("%s: pipeline control needs a stage (S1-S8)", c.ID)
		}
		if len(c.MapsTo) == 0 {
			add("%s: maps to no clause", c.ID)
		}
		for _, m := range c.MapsTo {
			if _, ok := clauseOwner[m]; !ok {
				add("%s: unknown clause %s", c.ID, m)
			}
		}
	}
	sort.Strings(errs)
	return errs
}
