package main

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// indepNorm is an INDEPENDENT associative-commutative normalizer (written without
// Key): flatten same-operator chains with an explicit stack, sort leaves' strings.
func indepNorm(e *Expr) string {
	switch e.Op {
	case "const":
		return "#" + strings.Repeat("+", 0) + itoa(e.N)
	case "var":
		return "$" + e.S
	case "neg":
		return "-[" + indepNorm(e.A) + "]"
	}
	var leaves []string
	stack := []*Expr{e}
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if x.Op == e.Op {
			stack = append(stack, x.B, x.A)
		} else {
			leaves = append(leaves, indepNorm(x))
		}
	}
	sort.Strings(leaves)
	return e.Op + "<" + strings.Join(leaves, ";") + ">"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "m" + string(b)
	}
	return string(b)
}

func TestParsePrefixInvertsRenderPrefix(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for i := 0; i < 400; i++ {
		a, b := genTrain(r, 2+r.Intn(7))
		for _, e := range []*Expr{a, b} {
			p, err := ParsePrefix(RenderPrefix(e))
			if err != nil || RenderPrefix(p) != RenderPrefix(e) {
				t.Fatalf("roundtrip failed for %s: %v", RenderPrefix(e), err)
			}
		}
	}
	for _, bad := range []string{"", "x", "add(c1)", "add(c1,c2", "c", "neg(c1", "add(c1,c2)z"} {
		if _, err := ParsePrefix(bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

func TestIndependentNormalizerAgreesWithKeyPartition(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	var es []*Expr
	for i := 0; i < 150; i++ {
		a, b := genTrain(r, 3+r.Intn(5))
		es = append(es, a, b)
	}
	// AC variants must collide under both; the partitions must be identical.
	es = append(es, EAdd(V("a"), EAdd(V("b"), V("c"))), EAdd(EAdd(V("c"), V("a")), V("b")), EMul(V("a"), V("b")), EMul(V("b"), V("a")))
	for i := range es {
		for j := range es {
			if (Key(es[i]) == Key(es[j])) != (indepNorm(es[i]) == indepNorm(es[j])) {
				t.Fatalf("partitions differ: %s vs %s", RenderPrefix(es[i]), RenderPrefix(es[j]))
			}
		}
	}
	if Key(ENeg(ENeg(V("a")))) == Key(V("a")) {
		t.Fatal("Key must not collapse neg(neg(x))")
	}
}

func TestGenBankG1DeclaredRuleAndDeterminism(t *testing.T) {
	prior := G1PriorKeys()
	rule := G1Rule{Seed: 77, PerLevel: 8, Levels: []int{5, 6}}
	a, st, err := GenBankG1(rule, prior)
	if err != nil || len(a) != 16 {
		t.Fatalf("err %v len %d", err, len(a))
	}
	b, _, _ := GenBankG1(rule, prior)
	if BankDigest(a) != BankDigest(b) {
		t.Fatal("not deterministic")
	}
	seen := map[string]bool{}
	for _, c := range a {
		for _, e := range []*Expr{c.Start, c.Target} {
			n := indepNorm(e) // checked with the independent normalizer, not Key
			if seen[n] {
				t.Fatalf("expression repeated inside bank: %s", RenderPrefix(e))
			}
			seen[n] = true
		}
	}
	priorN := map[string]bool{}
	for _, f := range []string{"levels.json", "blind.json"} {
		for k := 2; k <= 8; k++ {
			for _, p := range LoadProblems(f, itoa(k)) {
				priorN[indepNorm(p.Cur)], priorN[indepNorm(p.Target)] = true, true
			}
		}
	}
	for _, c := range GenBank(BankRule{Seed: 9001, PerLevel: 30, Levels: []int{6, 7, 8}}) {
		priorN[indepNorm(c.Start)], priorN[indepNorm(c.Target)] = true, true
	}
	for n := range seen {
		if priorN[n] {
			t.Fatal("G1 expression found in the prior set (independent check)")
		}
	}
	if len(st.PerLevel) != 2 || st.PerLevel[0].Accepted != 8 || st.PriorKeys != len(prior) {
		t.Fatalf("stats %+v", st)
	}
}

func TestGenBankG1FailsClosedWithoutRelaxing(t *testing.T) {
	cs, st, err := GenBankG1(G1Rule{Seed: 1, PerLevel: 30, Levels: []int{6, 7}, MaxAttemptsPerLevel: 5}, G1PriorKeys())
	if err == nil || len(cs) >= 30 || st.PerLevel[0].Attempts != 5 || len(st.PerLevel) != 1 {
		t.Fatalf("expected fail-closed at level 6: err=%v cases=%d stats=%+v", err, len(cs), st)
	}
}

func TestGenBankG1RejectsEveryPriorKey(t *testing.T) {
	// If every key were prior nothing may be accepted (reasons are counted).
	r := rand.New(rand.NewSource(9))
	all := map[string]bool{}
	for i := 0; i < 5000; i++ {
		a, b := genTrain(r, 5)
		all[Key(a)], all[Key(b)] = true, true
	}
	r2 := rand.New(rand.NewSource(9))
	_ = r2
	cs, st, err := GenBankG1(G1Rule{Seed: 9, PerLevel: 3, Levels: []int{5}, MaxAttemptsPerLevel: 200}, all)
	if err == nil || len(cs) != 0 || st.PerLevel[0].StartPrior+st.PerLevel[0].EqualKey != 200 {
		t.Fatalf("err=%v cases=%d stats=%+v", err, len(cs), st)
	}
}

func g1CheapCfg(seed int64, keys map[string]bool) TrainConfig {
	return TrainConfig{Seed: seed, Rounds: 3, PerRound: 10, Cap: 6, ExcludeKeys: keys}
}

func TestTrainG1ExcludesStartsTargetsAndHindsightByIndependentNormalization(t *testing.T) {
	// Control: record the natural training stream with an empty (non-nil) key set.
	ctl, err := TrainRounds(NewMLP(1), g1CheapCfg(1, map[string]bool{}))
	if err != nil || len(ctl.RenderedProblems) != 30 {
		t.Fatalf("control: %v n=%d", err, len(ctl.RenderedProblems))
	}
	// Bank = AC-permuted variants of some natural Starts and Targets (different
	// RenderPrefix, same Key) plus kept hindsight pairs' targets of the control.
	var bank []BankCase
	for i := 0; i < len(ctl.RenderedProblems); i += 3 {
		p := strings.SplitN(ctl.RenderedProblems[i], "|", 2)
		a, _ := ParsePrefix(p[0])
		b, _ := ParsePrefix(p[1])
		bank = append(bank, BankCase{EvalCase{Start: swapRootArgs(a), Target: swapRootArgs(b)}, 6})
	}
	for _, h := range ctl.HindsightKeptRendered {
		p := strings.SplitN(h, "|", 2)
		a, _ := ParsePrefix(p[0])
		b, _ := ParsePrefix(p[1])
		bank = append(bank, BankCase{EvalCase{Start: a, Target: b}, 6})
	}
	bn := map[string]bool{}
	for _, c := range bank {
		bn[indepNorm(c.Start)], bn[indepNorm(c.Target)] = true, true
	}
	overlaps := func(l TrainLog) (int, int) {
		np, nh := 0, 0
		for _, pr := range l.RenderedProblems {
			p := strings.SplitN(pr, "|", 2)
			a, _ := ParsePrefix(p[0])
			b, _ := ParsePrefix(p[1])
			if bn[indepNorm(a)] || bn[indepNorm(b)] {
				np++
			}
		}
		for _, pr := range l.HindsightKeptRendered {
			p := strings.SplitN(pr, "|", 2)
			a, _ := ParsePrefix(p[0])
			b, _ := ParsePrefix(p[1])
			if bn[indepNorm(a)] || bn[indepNorm(b)] {
				nh++
			}
		}
		return np, nh
	}
	if np, _ := overlaps(ctl); np == 0 {
		t.Fatal("negative control: the independent checker found no overlap in the unfiltered stream")
	}
	got, err := TrainRounds(NewMLP(1), g1CheapCfg(1, G1KeySetOf(bank)))
	if err != nil || len(got.RenderedProblems) != 30 {
		t.Fatalf("filtered run: %v n=%d", err, len(got.RenderedProblems))
	}
	if np, nh := overlaps(got); np != 0 || nh != 0 {
		t.Fatalf("overlap after exclusion: problems=%d hindsight=%d", np, nh)
	}
	rej := 0
	for _, r := range got.Rounds {
		rej += r.KeyRejected
		if r.Attempts < r.Problems {
			t.Fatalf("attempts %d < problems %d", r.Attempts, r.Problems)
		}
	}
	if rej == 0 {
		t.Fatal("expected key rejections to be recorded")
	}
}

func swapRootArgs(e *Expr) *Expr {
	if e.Op == "add" || e.Op == "mul" {
		return &Expr{Op: e.Op, A: e.B, B: e.A}
	}
	return e
}

func TestHindsightExcludedByKeyChecksSubstituteTargetAndFailsClosed(t *testing.T) {
	keys := G1KeySetOf([]BankCase{{EvalCase{Start: V("a"), Target: EAdd(V("b"), V("c"))}, 6}})
	if !hindsightExcludedByKey(&HindsightPair{Start: "va", Target: "vz"}, keys) {
		t.Fatal("start in set not excluded")
	}
	if !hindsightExcludedByKey(&HindsightPair{Start: "vq", Target: "add(vc,vb)"}, keys) {
		t.Fatal("AC variant of the substitute target not excluded")
	}
	if hindsightExcludedByKey(&HindsightPair{Start: "vq", Target: "vz"}, keys) {
		t.Fatal("unrelated pair excluded")
	}
	if !hindsightExcludedByKey(&HindsightPair{Start: "garbage(", Target: "vz"}, keys) {
		t.Fatal("unparseable pair must be dropped")
	}
}

func TestTrainRoundsWithoutExcludeKeysUnchangedAndOmitsG1Fields(t *testing.T) {
	ev := trainEvalBank()
	l1, _ := TrainRounds(NewMLP(2), smallTrainCfg(2, ev))
	if len(l1.RenderedProblems) != 0 || len(l1.HindsightKeptRendered) != 0 || l1.Rounds[0].Attempts != 0 || l1.Rounds[0].KeyRejected != 0 {
		t.Fatal("G1 fields set in P3 mode")
	}
}

func TestTrainG1DropsHindsightWhoseRealSubstituteTargetIsInTheSet(t *testing.T) {
	mk := func(keys map[string]bool) (TrainLog, error) {
		n := NewMLP(11)
		n.Level = 6
		return TrainRounds(n, TrainConfig{Seed: 1, Rounds: 1, PerRound: 12, Cap: 60, ExcludeKeys: keys})
	}
	ctl, err := mk(map[string]bool{})
	if err != nil || len(ctl.HindsightKeptRendered) == 0 {
		t.Fatalf("control must produce a kept hindsight pair: err=%v kept=%d", err, len(ctl.HindsightKeptRendered))
	}
	// The set holds ONLY the substitute targets of the control's hindsight pairs
	// (not their starts), so the problem stream is unchanged and only the
	// hindsight filter can act.
	set := map[string]bool{}
	subst := map[string]bool{}
	for _, h := range ctl.HindsightKeptRendered {
		b, err := ParsePrefix(strings.SplitN(h, "|", 2)[1])
		if err != nil {
			t.Fatal(err)
		}
		set[Key(b)] = true
		subst[indepNorm(b)] = true
	}
	for _, pr := range ctl.RenderedProblems {
		p := strings.SplitN(pr, "|", 2)
		a, _ := ParsePrefix(p[0])
		b, _ := ParsePrefix(p[1])
		if set[Key(a)] || set[Key(b)] {
			t.Skip("substitute target coincides with a natural problem expression; stream would change")
		}
	}
	got, err := mk(set)
	if err != nil {
		t.Fatal(err)
	}
	if got.Rounds[0].KeyRejected != 0 || got.Rounds[0].HindsightDropped < 1 || len(got.HindsightKeptRendered) != 0 {
		t.Fatalf("round %+v kept=%v", got.Rounds[0], got.HindsightKeptRendered)
	}
	for _, h := range got.HindsightKeptRendered {
		b, _ := ParsePrefix(strings.SplitN(h, "|", 2)[1])
		if subst[indepNorm(b)] {
			t.Fatal("kept hindsight has an excluded substitute target")
		}
	}
}
