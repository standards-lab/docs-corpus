package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type object struct {
	keys   []string
	values map[string]any
}

func newObject() *object {
	return &object{values: map[string]any{}}
}

func (o *object) set(key string, value any) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

func (o *object) get(key string) any {
	return o.values[key]
}

func (o *object) has(key string) bool {
	_, ok := o.values[key]
	return ok
}

func (o *object) str(key string) string {
	s, _ := o.values[key].(string)
	return s
}

type pyFloat float64

func parseJSON(data []byte) (any, error) {
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
			o := newObject()
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return nil, err
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				o.set(k.(string), v)
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

func truthy(v any) bool {
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
	case *object:
		return len(t.keys) > 0
	}
	return true
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
	case pyFloat:
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
	case *object:
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
		panic(fmt.Sprintf("place-traits: cannot write %T", v))
	}
}
