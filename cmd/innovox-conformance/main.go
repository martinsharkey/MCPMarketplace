// Command innovox-conformance measures this repository against the Innovox
// Assured scheme. It runs `go test -json ./...` (or reads saved events),
// combines tagged test results with scheme/self-assessment.json, and writes
// a Markdown report. It exits non-zero if any control is failing or the
// assessment makes a claim the evidence does not support.
//
// Usage: go run ./cmd/innovox-conformance -out docs/conformance/self-assessment.md
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"innovox.uk/assured/internal/conformance"
	"innovox.uk/assured/scheme"
)

func main() {
	dir := flag.String("scheme", "scheme", "scheme directory")
	assessPath := flag.String("assessment", "scheme/self-assessment.json", "written assessment")
	eventsPath := flag.String("events", "", "saved `go test -json` output (default: run the tests)")
	out := flag.String("out", "", "Markdown output file (default: stdout)")
	flag.Parse()

	s, err := scheme.LoadDir(*dir)
	if err != nil {
		log.Fatal(err)
	}
	if errs := s.Validate(); len(errs) > 0 {
		log.Fatalf("scheme invalid: %v", errs)
	}
	var events io.Reader
	if *eventsPath != "" {
		f, err := os.Open(*eventsPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		events = f
	} else {
		var buf bytes.Buffer
		cmd := exec.Command("go", "test", "-count=1", "-json", "./...")
		cmd.Stdout, cmd.Stderr = &buf, os.Stderr
		_ = cmd.Run() // failures are reported per control below
		events = &buf
	}
	ev, err := conformance.ParseEvents(events)
	if err != nil {
		log.Fatal(err)
	}
	var assess struct {
		Assessments []conformance.Assessment `json:"assessments"`
	}
	b, err := os.ReadFile(*assessPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(b, &assess); err != nil {
		log.Fatal(err)
	}
	r := conformance.Build(s, ev, assess.Assessments)
	md := r.Markdown()
	if *out == "" {
		fmt.Print(md)
	} else if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		log.Fatal(err)
	}
	failing := 0
	for _, row := range r.Rows {
		if row.Status == "failing" {
			failing++
		}
	}
	if failing > 0 || len(r.Problems) > 0 {
		log.Fatalf("%d failing controls, %d problems", failing, len(r.Problems))
	}
}
