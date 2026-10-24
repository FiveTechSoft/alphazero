//go:build !js

package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func runTrain(t *testing.T, args ...string) D1TrainReport {
	t.Helper()
	var out, prog bytes.Buffer
	if err := RunD1Train(args, &out, &prog); err != nil {
		t.Fatal(err)
	}
	var r D1TrainReport
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

var cheapTrainArgs = []string{"-rounds", "2", "-perround", "6", "-cap", "40", "-evalcap", "40", "-evallevels", "2,3", "-secperlevel", "4"}

func TestD1TrainDeterministicAndConsistent(t *testing.T) {
	a := runTrain(t, append([]string{"-seed", "21"}, cheapTrainArgs...)...)
	b := runTrain(t, append([]string{"-seed", "21"}, cheapTrainArgs...)...)
	if a.PreWeights != b.PreWeights || a.PostWeights != b.PostWeights {
		t.Fatal("weights hashes differ between identical runs")
	}
	if a.PreWeights == a.PostWeights {
		t.Fatal("training did not change the weights")
	}
	if len(a.Banks) != 2 {
		t.Fatalf("banks %d", len(a.Banks))
	}
	for i, bk := range a.Banks {
		if bk.Digest != b.Banks[i].Digest || bk.DigestAlgo != BankDigestAlgo {
			t.Fatal("digest mismatch")
		}
		if len(bk.Pre) != bk.Cases || len(bk.Post) != bk.Cases || bk.Delta != bk.PostSolved-bk.PreSolved {
			t.Fatalf("bank %s inconsistent", bk.Name)
		}
		if bk.PostSolved-bk.PreSolved != bk.GainSolved-bk.LostSolved {
			t.Fatalf("bank %s: transitions do not add up", bk.Name)
		}
		for j := range bk.Pre {
			if bk.Pre[j] != b.Banks[i].Pre[j] || bk.Post[j] != b.Banks[i].Post[j] {
				t.Fatal("per-case results differ between identical runs")
			}
		}
	}
	for n, o := range a.Overlap {
		if o.PairInEval != 0 {
			t.Fatalf("%s: training pair in evaluation set", n)
		}
	}
	if a.Library != "none" || len(a.Train.Rounds) != 2 {
		t.Fatal("report fields")
	}
}

func TestD1TrainDifferentSeedsDiffer(t *testing.T) {
	a := runTrain(t, append([]string{"-seed", "21"}, cheapTrainArgs...)...)
	b := runTrain(t, append([]string{"-seed", "22"}, cheapTrainArgs...)...)
	if a.PreWeights == b.PreWeights || a.PostWeights == b.PostWeights {
		t.Fatal("different seeds, same weights")
	}
}

func TestD1TrainRejectsBadInput(t *testing.T) {
	var o, p bytes.Buffer
	if RunD1Train(nil, &o, &p) == nil {
		t.Fatal("missing seed accepted")
	}
	if RunD1Train([]string{"-seed", "1", "-evallevels", "x"}, &o, &p) == nil {
		t.Fatal("bad levels accepted")
	}
	if RunD1Train([]string{"-seed", "1", "-rounds", "0"}, &o, &p) == nil {
		t.Fatal("zero rounds accepted")
	}
}

// Staged execution must reproduce the single-process report exactly (except timings).
func TestD1TrainStagedEqualsSingleProcess(t *testing.T) {
	base := append([]string{"-seed", "31"}, cheapTrainArgs...)
	single := runTrain(t, base...)
	dir := t.TempDir()
	for _, st := range []string{"pre-primary", "pre-secondary", "train", "post-primary", "post-secondary"} {
		var o, p bytes.Buffer
		if err := RunD1Train(append(append([]string{}, base...), "-stage", st, "-dir", dir), &o, &p); err != nil {
			t.Fatalf("stage %s: %v", st, err)
		}
	}
	staged := runTrain(t, append(append([]string{}, base...), "-stage", "report", "-dir", dir)...)
	single.SecondsPhase, staged.SecondsPhase = nil, nil
	a, _ := json.Marshal(single)
	b, _ := json.Marshal(staged)
	if !bytes.Equal(a, b) {
		t.Fatalf("staged report differs from single-process report")
	}
	// refuses to overwrite a stage, and refuses a different configuration
	var o, p bytes.Buffer
	if RunD1Train(append(append([]string{}, base...), "-stage", "train", "-dir", dir), &o, &p) == nil {
		t.Fatal("overwrote an existing stage")
	}
	other := append([]string{"-seed", "32"}, cheapTrainArgs...)
	if RunD1Train(append(other, "-stage", "report", "-dir", dir), &o, &p) == nil {
		t.Fatal("report accepted stages from a different configuration")
	}
}
