package main

import (
	"bytes"
	"fmt"
	"testing"
)

func adapterCases(t *testing.T) []EvalCase {
	t.Helper()
	var cs []EvalCase
	for _, k := range []string{"1", "2"} {
		for i, p := range LoadProblems("levels.json", k) {
			cs = append(cs, EvalCase{ID: fmt.Sprintf("L%s-%d", k, i), Start: p.Cur, Target: p.Target})
			if i >= 5 {
				break
			}
		}
	}
	if len(cs) == 0 {
		t.Fatal("no cases loaded")
	}
	return cs
}

func TestEvalBankDeterministic(t *testing.T) {
	net, cs := NewMLP(1), adapterCases(t)
	a := BankJSON(EvalBank(net, cs, EvalConfig{Cap: 200}))
	b := BankJSON(EvalBank(net, cs, EvalConfig{Cap: 200}))
	if !bytes.Equal(a, b) {
		t.Fatal("two runs differ")
	}
}

func TestEvalBankMatchesGuidedSolve(t *testing.T) {
	net, cs := NewMLP(2), adapterCases(t)
	rs := EvalBank(net, cs, EvalConfig{Cap: 150})
	prev := SearchTraining
	SearchTraining = false
	defer func() { SearchTraining = prev }()
	for i, c := range cs {
		ok, n, _ := GuidedSolve(net, c.Start, c.Target, 150)
		if rs[i].Solved != ok || rs[i].Nodes != n || rs[i].ID != c.ID {
			t.Fatalf("case %s differs: %+v vs %v %d", c.ID, rs[i], ok, n)
		}
	}
}

func TestEvalBankRestoresGlobalsAndLibrary(t *testing.T) {
	net, cs := NewMLP(3), adapterCases(t)
	prevT, prevL := SearchTraining, LemmaLib
	SearchTraining = true
	lib := NewLibrary()
	EvalBank(net, cs, EvalConfig{Cap: 100, Library: lib})
	if !SearchTraining || LemmaLib != prevL {
		t.Fatal("globals not restored")
	}
	SearchTraining = prevT
	if len(lib.Lemmas) != 0 {
		t.Fatal("evaluation mutated the library")
	}
}

func TestEvalBankEmptyLibraryEqualsNil(t *testing.T) {
	net, cs := NewMLP(4), adapterCases(t)
	a := BankJSON(EvalBank(net, cs, EvalConfig{Cap: 120}))
	b := BankJSON(EvalBank(net, cs, EvalConfig{Cap: 120, Library: NewLibrary()}))
	if !bytes.Equal(a, b) {
		t.Fatal("empty library changes outcome")
	}
}

func TestZeroCapSolvesOnlyIdentity(t *testing.T) {
	net, cs := NewMLP(5), adapterCases(t)
	for _, r := range EvalBank(net, cs, EvalConfig{Cap: 0}) {
		if r.Solved || r.Nodes != 0 {
			t.Fatalf("cap 0 result %+v", r)
		}
	}
}
