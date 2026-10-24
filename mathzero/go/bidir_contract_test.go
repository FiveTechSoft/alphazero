package main

import (
	"fmt"
	"testing"
)

// New contract tests for the search/utility code inherited from the base
// branch. They test stated invariants, not performance.

func contractCases(k string, n int) []MState {
	ps := LoadProblems("levels.json", k)
	if len(ps) > n {
		ps = ps[:n]
	}
	return ps
}

func withTraining(v bool, f func()) {
	prev := SearchTraining
	SearchTraining = v
	defer func() { SearchTraining = prev }()
	f()
}

func TestContractEvalModeNoTrainingArtifacts(t *testing.T) {
	net := NewMLP(7)
	for _, k := range []string{"2", "4"} {
		for i, p := range contractCases(k, 6) {
			withTraining(false, func() {
				for _, cap := range []int{1, 2, 40} {
					r := guidedSolveBidirectional(net, p.Cur, p.Target, cap, false)
					// Observed contract: a solved search returns its path samples in any
					// mode; a failed search returns samples/hindsight only in training.
					if r.Hindsight != nil || (!r.OK && len(r.Samples) != 0) {
						t.Fatalf("level %s case %d cap %d: eval produced hindsight or failure samples", k, i, cap)
					}
					if r.Nodes > cap {
						t.Fatalf("level %s case %d: nodes %d > cap %d", k, i, r.Nodes, cap)
					}
				}
			})
		}
	}
}

func TestContractIdentityAndZeroCap(t *testing.T) {
	net := NewMLP(1)
	p := contractCases("2", 1)[0]
	r := guidedSolveBidirectional(net, p.Cur, p.Cur, 10, false)
	if !r.OK || r.Nodes != 0 {
		t.Fatalf("identity: %+v", r)
	}
	r = guidedSolveBidirectional(net, p.Cur, p.Target, 0, false)
	if r.OK || r.Nodes != 0 {
		t.Fatalf("cap 0: %+v", r)
	}
}

func TestContractTrainingSamplesAreValidIndices(t *testing.T) {
	net := NewMLP(11)
	seen := 0
	for _, k := range []string{"2", "4"} {
		for i, p := range contractCases(k, 8) {
			var r bidirResult
			withTraining(true, func() { r = guidedSolveBidirectional(net, p.Cur, p.Target, 200, true) })
			for j, s := range r.Samples {
				seen++
				if s.Target < 0 || s.Target >= len(s.Feats) {
					t.Fatalf("level %s case %d sample %d: target %d of %d", k, i, j, s.Target, len(s.Feats))
				}
			}
			var e bidirResult
			withTraining(false, func() { e = guidedSolveBidirectional(net, p.Cur, p.Target, 200, false) })
			if e.OK != r.OK {
				t.Fatalf("level %s case %d: solved differs between modes (%v vs %v)", k, i, e.OK, r.OK)
			}
		}
	}
	if seen == 0 {
		t.Skip("no training samples were produced by these cases; invariant not exercised")
	}
}

func TestContractExactActionAndVerifiedEdge(t *testing.T) {
	p := contractCases("2", 1)[0]
	acts := MathActions(p.Cur)
	if len(acts) == 0 {
		t.Skip("no actions")
	}
	for i, a := range acts {
		child := applyAction(p.Cur, a)
		_, idx := exactAction(p.Cur, child)
		if idx < 0 {
			t.Fatalf("action %d not recovered by exactAction", i)
		}
		if RenderPrefix(applyAction(p.Cur, acts[idx])) != RenderPrefix(child) {
			t.Fatalf("action %d: exactAction returned a different edge", i)
		}
		if !verifiedEdge(p.Cur, child) {
			t.Fatalf("action %d: legal edge not verified", i)
		}
	}
}

func TestContractBackwardRejectsInformationLosingRules(t *testing.T) {
	p := contractCases("2", 1)[0]
	for _, rule := range []int{2, 10, 12, 13} {
		if backwardEdgeAllowed(p.Cur, p.Target, MAction{Rule: rule}) {
			t.Fatalf("rule %d allowed backward", rule)
		}
	}
}

func TestContractLemmaUtilityFormula(t *testing.T) {
	l := &Lemma{WinningApplications: 5, Saving: 4, FailedGeneratedNodes: 30}
	if got, want := l.Utility(0.1), 5*3-0.1*30; got != want {
		t.Fatalf("utility %v want %v", got, want)
	}
}

func mkLemma(name string, win, saving, failed, problems int) *Lemma {
	return &Lemma{FromKey: name + "a", ToKey: name + "b", WinningApplications: win, Saving: saving, FailedGeneratedNodes: failed, TrainingProblems: problems}
}

func TestContractPruneUtility(t *testing.T) {
	L := NewLibrary()
	L.UtilityLambda, L.UtilityMinProblems, L.ActiveCap = 0.1, 10, 3
	L.Lemmas = []*Lemma{
		mkLemma("bad", 0, 3, 50, 20),  // utility<=0, enough evidence: dropped
		mkLemma("young", 0, 3, 50, 5), // utility<=0 but too little evidence: kept
		mkLemma("good", 4, 4, 10, 20), // 12-1=11
		mkLemma("best", 6, 5, 0, 20),  // 24
		mkLemma("mid", 2, 3, 0, 20),   // 4
		mkLemma("extra", 1, 3, 0, 20), // 2, cut by ActiveCap
	}
	L.PruneUtility()
	var names []string
	for _, l := range L.Lemmas {
		names = append(names, l.FromKey)
	}
	want := fmt.Sprint([]string{"besta", "gooda", "mida"})
	if fmt.Sprint(names) != want {
		t.Fatalf("kept %v want %v", names, want)
	}
	for i, l := range L.Lemmas {
		for _, k := range []string{l.FromKey, l.ToKey} {
			found := false
			for _, ix := range L.ByKey[k] {
				found = found || ix == i
			}
			if !found {
				t.Fatalf("ByKey missing index %d for %s", i, k)
			}
		}
	}
}

func TestContractPruneKeepsYoungWhenUnderCap(t *testing.T) {
	L := NewLibrary()
	L.UtilityLambda, L.UtilityMinProblems, L.ActiveCap = 0.1, 10, 0
	L.Lemmas = []*Lemma{mkLemma("young", 0, 3, 50, 5)}
	L.PruneUtility()
	if len(L.Lemmas) != 1 {
		t.Fatal("young lemma dropped")
	}
}

func TestContractEvalDoesNotTouchLibraryCounters(t *testing.T) {
	net := NewMLP(5)
	L := NewLibrary()
	L.Lemmas = []*Lemma{mkLemma("x", 1, 3, 2, 4)}
	before := *L.Lemmas[0]
	prev := LemmaLib
	LemmaLib = L
	defer func() { LemmaLib = prev }()
	for _, p := range contractCases("4", 6) {
		guidedSolveBidirectional(net, p.Cur, p.Target, 60, false)
	}
	if *L.Lemmas[0] != before {
		t.Fatalf("eval changed counters: %+v -> %+v", before, *L.Lemmas[0])
	}
}
