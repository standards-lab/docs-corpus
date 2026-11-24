package jsonvalue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Object is a JSON object that keeps its keys in the order they were first set.
type Object struct {
	keys   []string
	values map[string]any
}

// NewObject returns an empty Object.
func NewObject() *Object {
	return &Object{values: map[string]any{}}
}

// Set sets key's value, keeping the key's place when it is already set.
func (o *Object) Set(key string, value any) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

// Get returns key's value, or nil when it is not set.
func (o *Object) Get(key string) any {
	return o.values[key]
}

// Has reports whether key is set.
func (o *Object) Has(key string) bool {
	_, ok := o.values[key]
	return ok
}

// Str returns key's value when it is a string, and "" otherwise.
func (o *Object) Str(key string) string {
	return String(o.values[key])
}

// Keys returns the object's keys in order.
func (o *Object) Keys() []string {
	return slices.Clone(o.keys)
}

// Float is a number Marshal writes as Python's repr writes a float, as in 3.0 or 1e+16.
type Float float64

// String returns v when it is a string, and "" otherwise.
func String(v any) string {
	s, _ := v.(string)
	return s
}

// Parse parses one JSON value: an object as an *Object, an array as []any, a number as a
// json.Number, and a string, a bool, or null as Go's own. Data after the value is an error.
func Parse(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing data")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			o := NewObject()
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return nil, err
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				o.Set(k.(string), v)
			}
			_, err := dec.Token()
			return o, err
		case '[':
			var list []any
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				list = append(list, v)
			}
			_, err := dec.Token()
			if list == nil {
				list = []any{}
			}
			return list, err
		}
		return nil, fmt.Errorf("unexpected %v", t)
	default:
		return t, nil
	}
}

// Truthy reports whether v is true as Python's bool reads it: null, false, "", zero, and an empty
// array or object are false; a number that does not parse is true.
func Truthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case json.Number:
		f, err := t.Float64()
		return err != nil || f != 0
	case []any:
		return len(t) > 0
	case *Object:
		return len(t.keys) > 0
	}
	return true
}

// Compare orders two values as Python orders them: strings by their bytes, and json.Numbers by
// their value.
func Compare(a, b any) int {
	if sa, ok := a.(string); ok {
		return strings.Compare(sa, String(b))
	}
	fa, _ := a.(json.Number).Float64()
	fb, _ := b.(json.Number).Float64()
	switch {
	case fa < fb:
		return -1
	case fa > fb:
		return 1
	}
	return 0
}

func pyFloatRepr(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	e := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expText, _ := strings.Cut(e, "e")
	exp, _ := strconv.Atoi(expText)
	if exp < -4 || exp >= 16 {
		sign := "+"
		if exp < 0 {
			sign = "-"
			exp = -exp
		}
		return fmt.Sprintf("%se%s%02d", mantissa, sign, exp)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.ContainsAny(s, ".") {
		s += ".0"
	}
	return s
}

func numberRepr(n json.Number) string {
	s := string(n)
	if !strings.ContainsAny(s, ".eE") {
		return s
	}
	f, _ := n.Float64()
	return pyFloatRepr(f)
}

func writeJSONString(w *bytes.Buffer, s string) {
	w.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			w.WriteString(`\"`)
		case '\\':
			w.WriteString(`\\`)
		case '\n':
			w.WriteString(`\n`)
		case '\r':
			w.WriteString(`\r`)
		case '\t':
			w.WriteString(`\t`)
		case '\b':
			w.WriteString(`\b`)
		case '\f':
			w.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(w, `\u%04x`, r)
			} else {
				w.WriteRune(r)
			}
		}
	}
	w.WriteByte('"')
}

// Marshal writes v as Python's json.dumps writes it with an indent of 2 and ensure_ascii off: nil,
// bool, int, int64, Float, json.Number, string, []string, []int, []any, and *Object. Any other
// type panics.
func Marshal(v any) []byte {
	var buf bytes.Buffer
	writeJSON(&buf, v, 0)
	return buf.Bytes()
}

func writeJSON(w *bytes.Buffer, v any, depth int) {
	indent := func(d int) {
		w.WriteByte('\n')
		w.WriteString(strings.Repeat("  ", d))
	}
	switch t := v.(type) {
	case nil:
		w.WriteString("null")
	case bool:
		w.WriteString(strconv.FormatBool(t))
	case int:
		w.WriteString(strconv.Itoa(t))
	case int64:
		w.WriteString(strconv.FormatInt(t, 10))
	case Float:
		w.WriteString(pyFloatRepr(float64(t)))
	case json.Number:
		w.WriteString(numberRepr(t))
	case string:
		writeJSONString(w, t)
	case []string:
		list := make([]any, len(t))
		for i, s := range t {
			list[i] = s
		}
		writeJSON(w, list, depth)
	case []int:
		list := make([]any, len(t))
		for i, n := range t {
			list[i] = n
		}
		writeJSON(w, list, depth)
	case []any:
		if len(t) == 0 {
			w.WriteString("[]")
			return
		}
		w.WriteByte('[')
		for i, item := range t {
			if i > 0 {
				w.WriteByte(',')
			}
			indent(depth + 1)
			writeJSON(w, item, depth+1)
		}
		indent(depth)
		w.WriteByte(']')
	case *Object:
		if len(t.keys) == 0 {
			w.WriteString("{}")
			return
		}
		w.WriteByte('{')
		for i, k := range t.keys {
			if i > 0 {
				w.WriteByte(',')
			}
			indent(depth + 1)
			writeJSONString(w, k)
			w.WriteString(": ")
			writeJSON(w, t.values[k], depth+1)
		}
		indent(depth)
		w.WriteByte('}')
	default:
		panic(fmt.Sprintf("jsonvalue: cannot write %T", v))
	}
}
