package main

import "testing"

// Deterministic failing cases (fresh MLP seed 11, cap 60) found by measurement
// on the inherited search: both exhaust the whole budget and yield a verified
// hindsight path of 3 steps in training mode.
var hindsightCases = []struct {
	level string
	idx   int
}{{"6", 5}, {"8", 5}}

func TestContractHindsightOnlyInTraining(t *testing.T) {
	net := NewMLP(11)
	const cap = 60
	for _, c := range hindsightCases {
		p := LoadProblems("levels.json", c.level)[c.idx]
		var tr, ev bidirResult
		withTraining(true, func() { tr = guidedSolveBidirectional(net, p.Cur, p.Target, cap, true) })
		withTraining(false, func() { ev = guidedSolveBidirectional(net, p.Cur, p.Target, cap, false) })
		if tr.OK || ev.OK {
			t.Fatalf("level %s case %d: expected failure on original target (train %v eval %v)", c.level, c.idx, tr.OK, ev.OK)
		}
		if tr.Nodes != cap || ev.Nodes != cap {
			t.Fatalf("level %s case %d: expected full budget use, got %d/%d", c.level, c.idx, tr.Nodes, ev.Nodes)
		}
		if tr.Hindsight == nil || len(tr.Samples) < 3 || tr.Hindsight.Steps != len(tr.Samples) {
			t.Fatalf("level %s case %d: training hindsight missing or inconsistent: %+v samples %d", c.level, c.idx, tr.Hindsight, len(tr.Samples))
		}
		if tr.Hindsight.Start != RenderPrefix(p.Cur) {
			t.Fatalf("hindsight start differs from original start")
		}
		if tr.Hindsight.Target == RenderPrefix(p.Target) {
			t.Fatalf("hindsight target equals the original target on a failed search")
		}
		for j, s := range tr.Samples {
			if s.Target < 0 || s.Target >= len(s.Feats) {
				t.Fatalf("sample %d: bad target index", j)
			}
		}
		if ev.Hindsight != nil || len(ev.Samples) != 0 {
			t.Fatalf("level %s case %d: eval mode produced hindsight/samples", c.level, c.idx)
		}
	}
}

func TestContractTrainingFailureLibraryCounters(t *testing.T) {
	net := NewMLP(11)
	p := LoadProblems("levels.json", "6")[3] // fails at cap 60, no hindsight
	L := NewLibrary()
	L.Lemmas = []*Lemma{mkLemma("x", 2, 3, 1, 4)}
	prev := LemmaLib
	LemmaLib = L
	defer func() { LemmaLib = prev }()
	before := *L.Lemmas[0]
	var r bidirResult
	withTraining(true, func() { r = guidedSolveBidirectional(net, p.Cur, p.Target, 60, true) })
	if r.OK {
		t.Fatal("expected failure")
	}
	after := *L.Lemmas[0]
	if after.TrainingProblems != before.TrainingProblems+1 {
		t.Fatalf("TrainingProblems %d -> %d, want +1", before.TrainingProblems, after.TrainingProblems)
	}
	if after.WinningApplications != before.WinningApplications || after.Uses != before.Uses {
		t.Fatalf("a failed search credited a win: %+v -> %+v", before, after)
	}
	if after.FailedGeneratedNodes < before.FailedGeneratedNodes {
		t.Fatal("FailedGeneratedNodes decreased")
	}
}
