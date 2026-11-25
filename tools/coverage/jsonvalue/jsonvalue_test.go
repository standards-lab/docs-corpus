package jsonvalue_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/standards-lab/docs-corpus/tools/coverage/jsonvalue"
)

func TestParseAndMarshalKeepKeyOrderAsPythonWritesIt(t *testing.T) {
	v, err := jsonvalue.Parse([]byte(`{"b": 10, "a": [1.50, "x", true, null], "c": {}, "d": []}`))
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "b": 10,
  "a": [
    1.5,
    "x",
    true,
    null
  ],
  "c": {},
  "d": []
}`
	if got := string(jsonvalue.Marshal(v)); got != want {
		t.Errorf("Marshal(Parse(...)) =\n%s\nwant\n%s", got, want)
	}
}

func TestParseRejectsTrailingAndTruncatedData(t *testing.T) {
	for _, in := range []string{`{} x`, `[1,`, `{"a"`, ``} {
		if _, err := jsonvalue.Parse([]byte(in)); err == nil {
			t.Errorf("Parse(%q) succeeded, want an error", in)
		}
	}
}

func TestMarshalWritesFloatsInPythonRepr(t *testing.T) {
	for _, c := range []struct {
		in   float64
		want string
	}{
		{3, "3.0"},
		{45.25, "45.25"},
		{0.0001, "0.0001"},
		{0.00001, "1e-05"},
		{1e16, "1e+16"},
		{123456789012345.0, "123456789012345.0"},
	} {
		if got := string(jsonvalue.Marshal(jsonvalue.Float(c.in))); got != c.want {
			t.Errorf("Marshal(Float(%v)) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestMarshalEscapesControlCharactersAndKeepsUnicode(t *testing.T) {
	got := string(jsonvalue.Marshal("a\"\\\n\t\x01é"))
	if want := `"a\"\\\n\t\u0001é"`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

func TestMarshalPanicsOnATypeItCannotWrite(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Marshal(float64) did not panic")
		}
	}()
	jsonvalue.Marshal(1.5)
}

func TestObjectSetKeepsAKeysFirstPlace(t *testing.T) {
	o := jsonvalue.NewObject()
	o.Set("a", 1)
	o.Set("b", "x")
	o.Set("a", 2)
	if got := o.Keys(); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("Keys = %v, want [a b]", got)
	}
	if o.Get("a") != 2 || o.Str("b") != "x" || o.Str("a") != "" || !o.Has("b") || o.Has("c") {
		t.Errorf("object = %s", jsonvalue.Marshal(o))
	}
}

func TestTruthyReadsAsPythonsBool(t *testing.T) {
	empty := jsonvalue.NewObject()
	full := jsonvalue.NewObject()
	full.Set("k", nil)
	for _, c := range []struct {
		v    any
		want bool
	}{
		{nil, false}, {false, false}, {true, true}, {"", false}, {"0", true},
		{json.Number("0"), false}, {json.Number("0.0"), false}, {json.Number("2"), true},
		{[]any{}, false}, {[]any{nil}, true}, {empty, false}, {full, true},
	} {
		if got := jsonvalue.Truthy(c.v); got != c.want {
			t.Errorf("Truthy(%#v) = %v, want %v", c.v, got, c.want)
		}
	}
}

func TestCompareOrdersNumbersByValueAndStringsByBytes(t *testing.T) {
	values := []any{json.Number("300"), json.Number("72.5"), json.Number("150")}
	slices.SortFunc(values, jsonvalue.Compare)
	if got := string(jsonvalue.Marshal(values)); got != "[\n  72.5,\n  150,\n  300\n]" {
		t.Errorf("sorted numbers = %s", got)
	}
	if jsonvalue.Compare("b", "a") <= 0 || jsonvalue.Compare("a", "a") != 0 {
		t.Error("Compare does not order strings by their bytes")
	}
}
