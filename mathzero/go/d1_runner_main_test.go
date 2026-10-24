//go:build !js

package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func runD1(t *testing.T, args ...string) D1EvalReport {
	t.Helper()
	var b bytes.Buffer
	if err := RunD1Eval(args, &b); err != nil {
		t.Fatal(err)
	}
	var r D1EvalReport
	if err := json.Unmarshal(b.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunnerDeterministicAndConsistent(t *testing.T) {
	a := runD1(t, "-level", "2", "-cap", "100", "-max", "8", "-seed", "3")
	b := runD1(t, "-level", "2", "-cap", "100", "-max", "8", "-seed", "3")
	if !bytes.Equal(BankJSON(a.Results), BankJSON(b.Results)) || a.BankSHA != b.BankSHA {
		t.Fatal("runs differ")
	}
	if a.Summary.N != 8 || len(a.Results) != 8 {
		t.Fatalf("n=%d", a.Summary.N)
	}
	s := 0
	for _, r := range a.Results {
		if r.Solved {
			s++
		}
	}
	if s != a.Summary.Solved {
		t.Fatal("summary mismatch")
	}
}

func TestRunnerBankDigestDependsOnSelection(t *testing.T) {
	a := runD1(t, "-level", "2", "-max", "4")
	b := runD1(t, "-level", "2", "-max", "5")
	if a.BankSHA == b.BankSHA {
		t.Fatal("digest ignores selection")
	}
}

func TestRunnerRejectsBadInput(t *testing.T) {
	var b bytes.Buffer
	if RunD1Eval([]string{"-level", "nope"}, &b) == nil {
		t.Fatal("unknown level accepted")
	}
	if RunD1Eval([]string{"-cap", "-1"}, &b) == nil {
		t.Fatal("negative cap accepted")
	}
}

func TestRunnerLibraryDigestWiring(t *testing.T) {
	a := runD1(t, "-level", "2", "-cap", "100", "-max", "6")
	l := runD1(t, "-level", "2", "-cap", "100", "-max", "6", "-lib", "lemas.json")
	if l.LibSHA == "" {
		t.Fatal("library digest missing")
	}
	_ = a // library content may legitimately change results; only digest wiring is asserted
}
