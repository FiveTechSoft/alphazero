package main

import "testing"

// Cost observability of EvalBank/SummarizeBank. Cases that depend on a measured
// property of this specific bank (fresh MLP seed 11, first 12 cases per level)
// are labelled MEASURED: they pin an observation, not a general theorem.

func obsBank(k string) []EvalCase {
	ps := LoadProblems("levels.json", k)
	if len(ps) > 12 {
		ps = ps[:12]
	}
	cs := make([]EvalCase, len(ps))
	for i, p := range ps {
		cs[i] = EvalCase{ID: k + "-" + string(rune('a'+i)), Start: p.Cur, Target: p.Target}
	}
	return cs
}

func TestObservabilityNodesWithinCapAndSummary(t *testing.T) {
	net := NewMLP(11)
	for _, k := range []string{"4", "6"} {
		cs := obsBank(k)
		for _, cap := range []int{20, 40} {
			rs := EvalBank(net, cs, EvalConfig{Cap: cap})
			sm := SummarizeBank(rs, cap)
			s, e := 0, 0
			for _, r := range rs {
				if r.Nodes > cap {
					t.Fatalf("level %s cap %d: nodes %d > cap", k, cap, r.Nodes)
				}
				if r.Solved {
					s++
				} else if r.Nodes >= cap {
					e++
				}
			}
			if sm.N != len(cs) || sm.Solved != s || sm.BudgetExhausted != e || s+e > sm.N {
				t.Fatalf("level %s cap %d: inconsistent summary %+v (s=%d e=%d)", k, cap, sm, s, e)
			}
		}
	}
}

// MEASURED (seed 11, 12 cases): level 4 cap 20 -> 11 solved, 1 exhausted;
// level 6 cap 20 -> 2 solved, 10 exhausted; every unsolved case used its whole
// budget; raising the cap 20 -> 40 never lost a solved case.
func TestObservabilityMeasuredBankProfile(t *testing.T) {
	net := NewMLP(11)
	want := map[string][2]int{"4": {11, 1}, "6": {2, 10}}
	for _, k := range []string{"4", "6"} {
		cs := obsBank(k)
		lo := EvalBank(net, cs, EvalConfig{Cap: 20})
		hi := EvalBank(net, cs, EvalConfig{Cap: 40})
		sm := SummarizeBank(lo, 20)
		if sm.Solved != want[k][0] || sm.BudgetExhausted != want[k][1] {
			t.Fatalf("level %s cap 20: solved/exhausted %d/%d, measured %v", k, sm.Solved, sm.BudgetExhausted, want[k])
		}
		for i, r := range lo {
			if !r.Solved && r.Nodes != 20 {
				t.Fatalf("level %s case %d: unsolved with %d nodes < cap", k, i, r.Nodes)
			}
			if r.Solved && !hi[i].Solved {
				t.Fatalf("level %s case %d: solved at cap 20 but not at cap 40", k, i)
			}
		}
	}
}
