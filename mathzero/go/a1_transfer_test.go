package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Integrity tests need the finished A1C-B results: set A1C_RESULTS_ROOT. Missing => FAIL, never a silent skip.
func resRoot(t *testing.T) string {
	r := os.Getenv("A1C_RESULTS_ROOT")
	if r == "" {
		t.Fatal("A1C_RESULTS_ROOT not set: integrity tests cannot run (this is a failure, not a skip)")
	}
	return r
}

func TestSeenRegeneratesStoredLists(t *testing.T) {
	root := resRoot(t)
	g3, err := transferG3Keys()
	if err != nil {
		t.Fatal(err)
	}
	for s := int64(19); s <= 26; s++ {
		_, rendered, err := transferSeen(s, A1Schedule, 40, g3)
		if err != nil {
			t.Fatal(err)
		}
		for _, arm := range []string{"A", "F"} {
			b, e := os.ReadFile(fmt.Sprintf("%s/seed-%d/train-%s.json", root, s, arm))
			if e != nil {
				t.Fatal(e)
			}
			var sf struct {
				Train struct{ RenderedProblems []string } `json:"train_log"`
			}
			if json.Unmarshal(b, &sf) != nil || len(sf.Train.RenderedProblems) != len(rendered) {
				t.Fatalf("seed %d arm %s: stored list unreadable or length differs", s, arm)
			}
			for i := range rendered {
				if sf.Train.RenderedProblems[i] != rendered[i] {
					t.Fatalf("seed %d arm %s problem %d differs", s, arm, i)
				}
			}
		}
	}
}

func TestNewRuleDisjointDeterministicBalanced(t *testing.T) {
	g3, err := transferG3Keys()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []int64{19, 23} {
		seen, nw, _, err := transferLists(s, 40)
		if err != nil {
			t.Fatal(err)
		}
		_, nw2, _, _ := transferLists(s, 40)
		if len(nw.Cases) != len(nw2.Cases) {
			t.Fatal("not deterministic (len)")
		}
		for i := range nw.Cases {
			if pairKey(nw.Cases[i].Start, nw.Cases[i].Target) != pairKey(nw2.Cases[i].Start, nw2.Cases[i].Target) {
				t.Fatal("not deterministic")
			}
		}
		for l, n := range seen.Kept {
			if nw.Kept[l] != n {
				t.Fatalf("seed %d level %d: new has %d cases, seen has %d", s, l, nw.Kept[l], n)
			}
		}
		for _, c := range nw.Cases {
			if g3[Key(c.Start)] || g3[Key(c.Target)] {
				t.Fatal("new overlaps G3 keys")
			}
			if seen.Pairs[pairKey(c.Start, c.Target)] {
				t.Fatal("new pair was seen")
			}
		}
		// both lists duplicate-free by pairKey
		for _, tl := range []TransferList{seen, nw} {
			m := map[string]bool{}
			for _, c := range tl.Cases {
				k := pairKey(c.Start, c.Target)
				if m[k] {
					t.Fatal("duplicate pair inside a list")
				}
				m[k] = true
			}
		}
	}
}

func TestWeightsAndInitialRecoverable(t *testing.T) {
	root := resRoot(t)
	for s := int64(19); s <= 26; s++ {
		var pre struct {
			Weights string `json:"weights_sha256"`
		}
		pb, e := os.ReadFile(fmt.Sprintf("%s/seed-%d/pre-g3.json", root, s))
		if e != nil {
			t.Fatal(e)
		}
		json.Unmarshal(pb, &pre)
		if pre.Weights == "" || WeightsSHA(NewMLP(s)) != pre.Weights {
			t.Fatalf("initial net of seed %d not reproduced by NewMLP", s)
		}
		for _, arm := range []string{"A", "F"} {
			var st struct {
				Weights string `json:"weights_sha256"`
			}
			sb, e := os.ReadFile(fmt.Sprintf("%s/seed-%d/train-%s.json", root, s, arm))
			if e != nil {
				t.Fatal(e)
			}
			json.Unmarshal(sb, &st)
			if st.Weights == "" || WeightsSHA(LoadMLP(fmt.Sprintf("%s/seed-%d/weights_%s.json", root, s, arm))) != st.Weights {
				t.Fatalf("stored weights %d/%s do not match recorded sha", s, arm)
			}
		}
	}
}
