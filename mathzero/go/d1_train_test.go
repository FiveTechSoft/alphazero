package main

import "testing"

func trainEvalBank() []EvalCase {
	var cs []EvalCase
	for _, k := range []string{"2", "3", "4"} {
		for i, p := range LoadProblems("levels.json", k) {
			cs = append(cs, EvalCase{ID: k + "-" + string(rune('a'+i%26)), Start: p.Cur, Target: p.Target})
		}
	}
	return cs
}

func smallTrainCfg(seed int64, ev []EvalCase) TrainConfig {
	return TrainConfig{Seed: seed, Rounds: 3, PerRound: 10, Cap: 60, Exclude: ExcludeSetFor(ev), ExcludeRender: RenderExcludeFor(ev)}
}

func TestTrainRoundsDeterministicAndChangesWeights(t *testing.T) {
	ev := trainEvalBank()
	a, b := NewMLP(1), NewMLP(1)
	w0 := WeightsSHA(a)
	la, err := TrainRounds(a, smallTrainCfg(5, ev))
	if err != nil {
		t.Fatal(err)
	}
	lb, _ := TrainRounds(b, smallTrainCfg(5, ev))
	if WeightsSHA(a) != WeightsSHA(b) {
		t.Fatal("same seed, different weights")
	}
	if WeightsSHA(a) == w0 {
		t.Fatal("training did not change weights")
	}
	if len(la.Rounds) != 3 || len(la.Pairs) != 30 || len(lb.Pairs) != 30 {
		t.Fatalf("log sizes %d %d", len(la.Rounds), len(la.Pairs))
	}
	for i := range la.Pairs {
		if la.Pairs[i] != lb.Pairs[i] {
			t.Fatal("training problems differ for the same seed")
		}
	}
	c := NewMLP(1)
	TrainRounds(c, smallTrainCfg(6, ev))
	if WeightsSHA(c) == WeightsSHA(a) {
		t.Fatal("different training seeds gave identical weights")
	}
}

func TestTrainRoundsExcludesEvalPairs(t *testing.T) {
	ev := trainEvalBank()
	log, err := TrainRounds(NewMLP(2), smallTrainCfg(9, ev))
	if err != nil {
		t.Fatal(err)
	}
	ex := ExcludeSetFor(ev)
	for _, p := range log.Pairs {
		if ex[p] {
			t.Fatal("a training pair is in the evaluation set")
		}
	}
	if o := OverlapTrainEval(log, ev); o.PairInEval != 0 || o.TrainProblems != 30 {
		t.Fatalf("overlap %+v", o)
	}
}

func TestTrainRoundsIgnoresLibraryAndRestoresGlobals(t *testing.T) {
	prevT, prevL := SearchTraining, LemmaLib
	defer func() { SearchTraining, LemmaLib = prevT, prevL }()
	L := NewLibrary()
	L.Lemmas = []*Lemma{mkLemma("x", 1, 3, 1, 1)}
	before := *L.Lemmas[0]
	LemmaLib, SearchTraining = L, false
	TrainRounds(NewMLP(3), smallTrainCfg(1, trainEvalBank()))
	if LemmaLib != L || SearchTraining {
		t.Fatal("globals not restored")
	}
	if *L.Lemmas[0] != before {
		t.Fatal("training touched a library it should ignore")
	}
}

func TestEvalBankNeverChangesWeights(t *testing.T) {
	net := NewMLP(4)
	TrainRounds(net, smallTrainCfg(2, trainEvalBank()))
	w := WeightsSHA(net)
	EvalBank(net, adapterCases(t), EvalConfig{Cap: 60})
	if WeightsSHA(net) != w {
		t.Fatal("evaluation changed weights")
	}
}

func TestTrainRoundsRejectsBadConfig(t *testing.T) {
	if _, err := TrainRounds(NewMLP(1), TrainConfig{Rounds: 0, PerRound: 1, Cap: 10}); err == nil {
		t.Fatal("accepted zero rounds")
	}
}

func TestOverlapTrainEvalCountsStartsAndTargets(t *testing.T) {
	ev := trainEvalBank()[:3]
	log := TrainLog{
		Pairs:   []string{"p1", pairKey(ev[0].Start, ev[0].Target)},
		Starts:  []string{Key(ev[0].Start), Key(ev[0].Start)},
		Targets: []string{"zzz", Key(ev[0].Target)},
	}
	o := OverlapTrainEval(log, ev)
	if o.TrainProblems != 2 || o.StartInEval != 2 || o.TargetInEval != 1 || o.PairInEval != 1 {
		t.Fatalf("%+v", o)
	}
}
