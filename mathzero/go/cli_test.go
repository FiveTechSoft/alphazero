package main

import "testing"

func TestParseIntArg(t *testing.T) {
	casos := []struct {
		raw  string
		min  int
		want int
		ok   bool
	}{
		{"3", 0, 3, true},
		{"0", 0, 0, true},
		{"-1", 0, 0, false},
		{"0", 1, 0, false},
		{"abc", 0, 0, false},
		{"", 0, 0, false},
	}
	for _, c := range casos {
		got, ok := parseIntArg("x", c.raw, c.min)
		if got != c.want || ok != c.ok {
			t.Errorf("parseIntArg(%q, min=%d) = (%d,%v), esperado (%d,%v)",
				c.raw, c.min, got, ok, c.want, c.ok)
		}
	}
}

func TestParseInt64Arg(t *testing.T) {
	if got, ok := parseInt64Arg("semilla", "-42"); !ok || got != -42 {
		t.Errorf("parseInt64Arg(-42) = (%d,%v)", got, ok)
	}
	if _, ok := parseInt64Arg("semilla", "12x"); ok {
		t.Error("parseInt64Arg acepto un valor invalido")
	}
}
