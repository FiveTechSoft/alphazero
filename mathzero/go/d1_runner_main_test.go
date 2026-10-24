//go:build !js

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// The published bank_digest must equal its declared algorithm, recomputed
// independently here. bank_sha256 keeps its historical, different semantics:
// levels.json mode = sha256 of "start\0target\n" lines (no level); generated
// mode = sha256 of the hex string of BankDigest (a hash of a hash).
func TestRunnerBankDigestMatchesDeclaredAlgorithm(t *testing.T) {
	g := runD1(t, "-bankseed", "7", "-perlevel", "3", "-levels", "4,6", "-cap", "10")
	bank := GenBank(BankRule{Seed: 7, PerLevel: 3, Levels: []int{4, 6}})
	h := sha256.New()
	for _, c := range bank {
		fmt.Fprintf(h, "%d\x00%s\x00%s\n", c.Level, RenderPrefix(c.Start), RenderPrefix(c.Target))
	}
	want := fmt.Sprintf("%x", h.Sum(nil))
	if g.BankDigest != want || g.BankDigestAlgo != BankDigestAlgo {
		t.Fatalf("bank_digest %s want %s", g.BankDigest, want)
	}
	legacy := sha256.Sum256([]byte(want))
	if g.BankSHA != fmt.Sprintf("%x", legacy) {
		t.Fatal("legacy bank_sha256 semantics changed for generated banks")
	}
	f := runD1(t, "-level", "2", "-max", "4", "-cap", "10")
	h = sha256.New()
	lh := sha256.New()
	for i, p := range LoadProblems("levels.json", "2")[:4] {
		_ = i
		fmt.Fprintf(h, "%d\x00%s\x00%s\n", 2, RenderPrefix(p.Cur), RenderPrefix(p.Target))
		fmt.Fprintf(lh, "%s\x00%s\n", RenderPrefix(p.Cur), RenderPrefix(p.Target))
	}
	if f.BankDigest != fmt.Sprintf("%x", h.Sum(nil)) {
		t.Fatal("levels.json bank_digest differs from declared algorithm")
	}
	if f.BankSHA != fmt.Sprintf("%x", lh.Sum(nil)) {
		t.Fatal("legacy bank_sha256 semantics changed for levels.json banks")
	}
}
