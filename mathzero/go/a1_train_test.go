package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
)

func a1Keys(t *testing.T) map[string]bool {
	rule := G1Rule{Seed: 5, PerLevel: 3, Levels: []int{4, 5}}
	prior, err := G2PriorKeys()
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := GenBankG1(rule, prior)
	if err != nil {
		t.Fatal(err)
	}
	return G1KeySetOf(b)
}

func a1Run(t *testing.T, seed int64, sched []int, per, cp int, frozen bool, keys map[string]bool) (A1Result, *MLP) {
	net := NewMLP(seed)
	r, err := TrainRoundsA1(net, A1Config{Seed: seed, Schedule: sched, PerRound: per, Cap: cp, FrozenGuide: frozen, ExcludeKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	return r, net
}

func TestA1ProblemListIdenticalAcrossArmsAndSearchCaps(t *testing.T) {
	keys := a1Keys(t)
	sched := []int{2, 3, 4}
	a, na := a1Run(t, 7, sched, 6, 6, false, keys)
	f, nf := a1Run(t, 7, sched, 6, 6, true, keys)
	a2, _ := a1Run(t, 7, sched, 6, 40, false, keys)
	j := func(r A1Result) string { return strings.Join(r.Log.RenderedProblems, "\n") }
	if len(a.Log.RenderedProblems) != 18 || j(a) != j(f) || j(a) != j(a2) {
		t.Fatalf("problem lists differ across arms/caps (n=%d)", len(a.Log.RenderedProblems))
	}
	// the cap changes how many samples SGD sees, yet the problems do not move:
	// the SGD stream cannot be consuming the problem stream.
	sa, sa2 := 0, 0
	for i := range a.Log.Rounds {
		sa += a.Log.Rounds[i].Samples
		sa2 += a2.Log.Rounds[i].Samples
	}
	if sa == sa2 {
		t.Fatal("test is vacuous: sample counts did not differ between caps")
	}
	for i, r := range a.Log.Rounds {
		if r.Level != sched[i] {
			t.Fatalf("level schedule not followed: %+v", r)
		}
	}
	if na.Level != NewMLP(7).Level || nf.Level != NewMLP(7).Level || na.Round != 3 || nf.Round != 3 {
		t.Fatalf("Level must stay untouched and Round advance: %d/%d %d/%d", na.Level, nf.Level, na.Round, nf.Round)
	}
}

func TestA1Round1IdenticalInBothArms(t *testing.T) {
	keys := a1Keys(t)
	a, na := a1Run(t, 9, []int{3}, 8, 20, false, keys)
	f, nf := a1Run(t, 9, []int{3}, 8, 20, true, keys)
	if a.Log.Rounds[0] != f.Log.Rounds[0] || strings.Join(a.Log.RenderedProblems, ";") != strings.Join(f.Log.RenderedProblems, ";") {
		t.Fatalf("round 1 differs: %+v vs %+v", a.Log.Rounds[0], f.Log.Rounds[0])
	}
	if WeightsSHA(na) != WeightsSHA(nf) {
		t.Fatal("same samples and same SGD stream must give identical weights after round 1")
	}
	if WeightsSHA(na) == WeightsSHA(NewMLP(9)) {
		t.Fatal("training did not change the weights")
	}
}

func TestA1FrozenGuideIsUnchangedAndNotAliased(t *testing.T) {
	keys := a1Keys(t)
	f, nf := a1Run(t, 4, []int{2, 3, 4}, 8, 20, true, keys)
	init := WeightsSHA(NewMLP(4))
	if f.GuideSHAPre != init || f.GuideSHAPost != init {
		t.Fatalf("guide not the frozen initial net: pre=%s post=%s init=%s", f.GuideSHAPre, f.GuideSHAPost, init)
	}
	if WeightsSHA(nf) == init {
		t.Fatal("learner did not train in arm F")
	}
	// CloneMLP shares no memory with its source
	src := NewMLP(4)
	c, err := CloneMLP(src)
	if err != nil || WeightsSHA(c) != WeightsSHA(src) {
		t.Fatalf("clone differs: %v", err)
	}
	h := WeightsSHA(src)
	c.W1[0][0] += 1
	c.W2[0] += 1
	c.B1[0] += 1
	if WeightsSHA(src) != h {
		t.Fatal("mutating the clone changed the source: aliasing")
	}
	fa, _ := a1Run(t, 4, []int{2}, 4, 20, false, keys)
	if fa.GuideSHAPre != "" || fa.GuideSHAPost != "" {
		t.Fatal("arm A must not report a guide")
	}
}

func TestA1ExcludesStartsTargetsAndHindsightByIndependentNormalization(t *testing.T) {
	ctl, _ := a1Run(t, 1, []int{2, 3, 4}, 10, 6, false, map[string]bool{})
	var bank []BankCase
	for i := 0; i < len(ctl.Log.RenderedProblems); i += 3 {
		p := strings.SplitN(ctl.Log.RenderedProblems[i], "|", 2)
		a, _ := ParsePrefix(p[0])
		b, _ := ParsePrefix(p[1])
		bank = append(bank, BankCase{EvalCase{Start: swapRootArgs(a), Target: swapRootArgs(b)}, 6})
	}
	bn := map[string]bool{}
	for _, c := range bank {
		bn[indepNorm(c.Start)], bn[indepNorm(c.Target)] = true, true
	}
	over := func(l TrainLog) int {
		n := 0
		for _, pr := range append(append([]string{}, l.RenderedProblems...), l.HindsightKeptRendered...) {
			p := strings.SplitN(pr, "|", 2)
			a, _ := ParsePrefix(p[0])
			b, _ := ParsePrefix(p[1])
			if bn[indepNorm(a)] || bn[indepNorm(b)] {
				n++
			}
		}
		return n
	}
	if over(ctl.Log) == 0 {
		t.Fatal("negative control: checker saw no overlap without the filter")
	}
	got, _ := a1Run(t, 1, []int{2, 3, 4}, 10, 6, false, G1KeySetOf(bank))
	if len(got.Log.RenderedProblems) != 30 || over(got.Log) != 0 {
		t.Fatalf("n=%d overlap=%d", len(got.Log.RenderedProblems), over(got.Log))
	}
	rej := 0
	for _, r := range got.Log.Rounds {
		rej += r.KeyRejected
	}
	if rej == 0 {
		t.Fatal("no rejections recorded")
	}
}

func TestA1DropsHindsightWhoseRealSubstituteTargetIsExcluded(t *testing.T) {
	mk := func(keys map[string]bool) A1Result {
		net := NewMLP(11)
		r, err := TrainRoundsA1(net, A1Config{Seed: 1, Schedule: []int{6}, PerRound: 12, Cap: 60, ExcludeKeys: keys})
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	ctl := mk(map[string]bool{})
	if len(ctl.Log.HindsightKeptRendered) == 0 {
		t.Skip("control produced no hindsight pair in this configuration")
	}
	set := map[string]bool{}
	for _, h := range ctl.Log.HindsightKeptRendered {
		b, err := ParsePrefix(strings.SplitN(h, "|", 2)[1])
		if err != nil {
			t.Fatal(err)
		}
		set[Key(b)] = true
	}
	for _, pr := range ctl.Log.RenderedProblems {
		p := strings.SplitN(pr, "|", 2)
		a, _ := ParsePrefix(p[0])
		b, _ := ParsePrefix(p[1])
		if set[Key(a)] || set[Key(b)] {
			t.Skip("substitute target coincides with a natural problem expression")
		}
	}
	got := mk(set)
	if got.Log.Rounds[0].HindsightDropped < 1 || len(got.Log.HindsightKeptRendered) != 0 || got.Log.Rounds[0].KeyRejected != 0 {
		t.Fatalf("round %+v kept %v", got.Log.Rounds[0], got.Log.HindsightKeptRendered)
	}
}

func TestA1RejectsInvalidConfigAndFailsClosedOnExhaustedPool(t *testing.T) {
	if _, err := TrainRoundsA1(NewMLP(1), A1Config{Seed: 1, Schedule: []int{2}, PerRound: 2, Cap: 10}); err == nil {
		t.Fatal("accepted nil ExcludeKeys")
	}
	if _, err := TrainRoundsA1(NewMLP(1), A1Config{Seed: 1, PerRound: 2, Cap: 10, ExcludeKeys: map[string]bool{}}); err == nil {
		t.Fatal("accepted empty schedule")
	}
	// exclude every Key the level-2 stream can produce (the keys are strings; the
	// pointer-keyed Key cache is cleared as we go to bound memory)
	all := map[string]bool{}
	r := randNew(1)
	for i := 0; i < 12000; i++ {
		a, b := genTrain(r, 2)
		all[Key(a)], all[Key(b)] = true, true
		if i%500 == 0 {
			ClearKeyCache()
		}
	}
	ClearKeyCache()
	if _, err := TrainRoundsA1(NewMLP(1), A1Config{Seed: 1, Schedule: []int{2}, PerRound: 2, Cap: 10, ExcludeKeys: all}); err == nil || !strings.Contains(err.Error(), "no training problem") {
		t.Fatalf("expected fail-closed error, got %v", err)
	}
}

func TestA1BankG2FrozenAndDisjointFromG1ByIndependentNormalization(t *testing.T) {
	prior, err := G2PriorKeys()
	if err != nil {
		t.Fatal(err)
	}
	rule := G1Rule{Seed: G2Seed, PerLevel: 30, Levels: []int{6, 7, 8}}
	g2, st, err := GenBankG1(rule, prior)
	if err != nil || len(g2) != 90 || BankDigest(g2) != G2FrozenDigest || G2Seed != 20261025 {
		t.Fatalf("err=%v n=%d digest=%s", err, len(g2), BankDigest(g2))
	}
	if st.PerLevel[0].Attempts != 56 || st.PerLevel[1].Attempts != 66 || st.PerLevel[2].Attempts != 49 || st.PriorKeys != 1739 {
		t.Fatalf("statistics changed: %+v", st)
	}
	g1, _, _ := GenBankG1(G1Rule{Seed: G1Seed, PerLevel: 30, Levels: []int{6, 7, 8}}, G1PriorKeys())
	n1 := map[string]bool{}
	for _, c := range g1 {
		n1[indepNorm(c.Start)], n1[indepNorm(c.Target)] = true, true
	}
	seen := map[string]bool{}
	for _, c := range g2 {
		for _, e := range []*Expr{c.Start, c.Target} {
			n := indepNorm(e)
			if n1[n] || seen[n] {
				t.Fatalf("G2 expression in G1 or repeated: %s", RenderPrefix(e))
			}
			seen[n] = true
		}
	}
}

func TestA1StagedRunReportsBothArms(t *testing.T) {
	rule := G1Rule{Seed: 5, PerLevel: 3, Levels: []int{4, 5}}
	prior, _ := G2PriorKeys()
	bank, st, err := GenBankG1(rule, prior)
	if err != nil {
		t.Fatal(err)
	}
	d := BankDigest(bank)
	dir := t.TempDir()
	var sink bytes.Buffer
	sched := []int{2, 3}
	run := func(stage string, out *bytes.Buffer) error {
		return RunA1StageWithBank(stage, dir, 3, sched, 6, 12, 12, bank, st, d, rule, out, &sink)
	}
	for _, s := range []string{"pre-g2", "train-A", "train-F", "post-A", "post-F"} {
		if err := run(s, &sink); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	var out bytes.Buffer
	if err := run("report", &out); err != nil {
		t.Fatal(err)
	}
	var rep A1Report
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if !rep.ProblemsIdentical || len(rep.Arms) != 2 || rep.D != rep.DeltaA-rep.DeltaF || rep.BankDigest != d {
		t.Fatalf("report %+v", rep)
	}
	for _, a := range rep.Arms {
		if a.Overlap.StartInEval != 0 || a.Overlap.TargetInEval != 0 || a.Overlap.PairInEval != 0 || a.Overlap.TrainProblems != 12 {
			t.Fatalf("arm %s overlap %+v", a.Arm, a.Overlap)
		}
	}
	if rep.Arms[1].GuideSHAPre == "" || rep.Arms[1].GuideSHAPre != rep.Arms[1].GuideSHAPost || rep.Arms[0].GuideSHAPre != "" {
		t.Fatal("guide hashes wrong in the report")
	}
	// refusals
	if err := run("train-A", &sink); err == nil {
		t.Fatal("overwrote train-A")
	}
	if err := RunA1StageWithBank("pre-g2", dir, 3, sched, 6, 12, 12, bank, st, "deadbeef", rule, &sink, &sink); err == nil {
		t.Fatal("wrong digest accepted")
	}
	if err := RunA1StageWithBank("post-A", t.TempDir(), 3, sched, 6, 12, 12, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("post-A ran without a train stage")
	}
	if err := RunA1StageWithBank("pre-g1", dir, 3, sched, 6, 12, 12, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("accepted a G1 stage name")
	}
	if err := RunA1StageWithBank("report", dir, 3, []int{2, 4}, 6, 12, 12, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("report accepted stages from another configuration")
	}
}

func randNew(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }
