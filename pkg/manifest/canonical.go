// Package manifest canonicalises and hashes MCP tool manifests and A2A
// Agent Cards so a certificate can pin exactly what a service exposes.
package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Canonicalize returns the RFC 8785 (JCS) canonical form of a JSON document.
// Unicode is never normalised, so invisible characters change the output.
func Canonicalize(raw []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := readValue(dec)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("canonicalize: trailing data")
	}
	var b bytes.Buffer
	if err := writeValue(&b, v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// readValue decodes one JSON value, rejecting duplicate object keys as
// RFC 8785 (I-JSON) requires. A duplicate key could otherwise let two
// parsers disagree about content that hashes identically.
func readValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			m := map[string]any{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k := kt.(string)
				if _, dup := m[k]; dup {
					return nil, fmt.Errorf("duplicate key %q", k)
				}
				if m[k], err = readValue(dec); err != nil {
					return nil, err
				}
			}
			_, err = dec.Token()
			return m, err
		case '[':
			a := []any{}
			for dec.More() {
				v, err := readValue(dec)
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			}
			_, err = dec.Token()
			return a, err
		}
		return nil, fmt.Errorf("unexpected %v", t)
	default:
		return t, nil
	}
}

func writeValue(b *bytes.Buffer, v any) error {
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case json.Number:
		s, err := formatNumber(string(t))
		if err != nil {
			return err
		}
		b.WriteString(s)
	case string:
		writeString(b, t)
	case []any:
		b.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := writeValue(b, e); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return lessUTF16(keys[i], keys[j]) })
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, k)
			b.WriteByte(':')
			if err := writeValue(b, t[k]); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	default:
		return fmt.Errorf("canonicalize: unsupported type %T", v)
	}
	return nil
}

// lessUTF16 orders keys by UTF-16 code units, as RFC 8785 requires.
func lessUTF16(a, b string) bool {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			return ua[i] < ub[i]
		}
	}
	return len(ua) < len(ub)
}

func writeString(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

// formatNumber serialises a JSON number the way ECMAScript Number.toString does.
func formatNumber(lit string) (string, error) {
	f, err := strconv.ParseFloat(lit, 64)
	if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
		return "", fmt.Errorf("canonicalize: bad number %q", lit)
	}
	if f == 0 {
		return "0", nil
	}
	sign := ""
	if f < 0 {
		sign, f = "-", -f
	}
	e := strconv.FormatFloat(f, 'e', -1, 64) // d.ddde±XX
	mant, expStr, _ := strings.Cut(e, "e")
	digits := strings.Replace(mant, ".", "", 1)
	exp, _ := strconv.Atoi(expStr)
	n, k := exp+1, len(digits)
	var out string
	switch {
	case k <= n && n <= 21:
		out = digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		out = digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		out = "0." + strings.Repeat("0", -n) + digits
	default:
		out = digits[:1]
		if k > 1 {
			out += "." + digits[1:]
		}
		es := "+"
		if n-1 < 0 {
			es = "-"
		}
		out += "e" + es + strconv.Itoa(abs(n-1))
	}
	return sign + out, nil
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
