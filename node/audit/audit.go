// Package audit is an append-only, hash-chained log of gateway decisions.
// Each line is one JSON event whose Hash covers the event (with Hash empty),
// including PrevHash, so edits, deletions and reordering break the chain.
package audit

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"innovox.uk/assured/pkg/manifest"
)

var genesis = strings.Repeat("0", 64)

type Event struct {
	Seq      uint64    `json:"seq"`
	PrevHash string    `json:"prevHash"`
	Hash     string    `json:"hash"`
	TS       time.Time `json:"ts"`
	Actor    string    `json:"actor"`
	Action   string    `json:"action"`
	Decision string    `json:"decision"`
	Detail   string    `json:"detail"`
}

type Log struct {
	mu   sync.Mutex
	f    *os.File
	seq  uint64
	prev string
	Now  func() time.Time
}

// Open verifies any existing chain and resumes it. A broken chain is an error.
func Open(path string) (*Log, error) {
	seq, prev, err := walk(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Log{f: f, seq: seq, prev: prev, Now: time.Now}, nil
}

func (l *Log) Close() error { return l.f.Close() }

// Append writes one event and syncs it to disk before returning.
func (l *Log) Append(actor, action, decision, detail string) (Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{Seq: l.seq + 1, PrevHash: l.prev, TS: l.Now().UTC(), Actor: actor, Action: action, Decision: decision, Detail: detail}
	h, err := hashEvent(e)
	if err != nil {
		return Event{}, err
	}
	e.Hash = h
	line, _ := json.Marshal(e)
	if _, err := l.f.Write(append(line, '\n')); err != nil {
		return Event{}, err
	}
	if err := l.f.Sync(); err != nil {
		return Event{}, err
	}
	l.seq, l.prev = e.Seq, e.Hash
	return e, nil
}

// Verify checks the whole chain in the file at path.
func Verify(path string) error {
	_, _, err := walk(path)
	return err
}

func walk(path string) (uint64, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, genesis, err
	}
	defer f.Close()
	seq, prev := uint64(0), genesis
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var e Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return 0, "", fmt.Errorf("audit: line %d: %w", seq+1, err)
		}
		want := e.Hash
		e.Hash = ""
		got, err := hashEvent(e)
		if err != nil {
			return 0, "", err
		}
		if e.Seq != seq+1 || e.PrevHash != prev || got != want {
			return 0, "", fmt.Errorf("audit: chain broken at line %d", seq+1)
		}
		seq, prev = e.Seq, want
	}
	return seq, prev, sc.Err()
}

func hashEvent(e Event) (string, error) {
	e.Hash = ""
	raw, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	c, err := manifest.Canonicalize(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(c)
	return hex.EncodeToString(sum[:]), nil
}
