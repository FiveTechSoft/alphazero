package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Synthetic only: no solver, no weights. Writes fake transfer-*.json files.
func tcWrite(t *testing.T, root string, seed int64, net string, seenSolved, newSolved int, mut func(*TransferOut)) {
	tr := true
	o := TransferOut{Seed: seed, Net: net, Levels: []int{6, 7}, Cap: 100, NCases: map[string]int{"seen": 80, "new": 80},
		ListSHA: map[string]string{"seen": fmt.Sprintf("S%d", seed), "new": fmt.Sprintf("N%d", seed)}, SeenMatches: &tr, SeenMatchesF: &tr,
		Solved: map[string]map[string][]int{net: {"seen": {seenSolved, 0}, "new": {newSolved, 0}}}}
	if mut != nil {
		mut(&o)
	}
	dir := filepath.Join(root, fmt.Sprintf("seed-%d", seed))
	os.MkdirAll(dir, 0o755)
	b, _ := json.Marshal(o)
	os.WriteFile(filepath.Join(dir, "transfer-"+net+".json"), b, 0o644)
}

// d per seed = gA - gF with 80 cases: gA=(sa-na)/80, gF=(sf-nf)/80
func tcSeedData(t *testing.T, root string, seed int64, sa, na, sf, nf int, mutNet string, mut func(*TransferOut)) {
	for _, n := range []string{"pre", "A", "F"} {
		s, w := 30, 30
		if n == "A" {
			s, w = sa, na
		}
		if n == "F" {
			s, w = sf, nf
		}
		var m func(*TransferOut)
		if n == mutNet {
			m = mut
		}
		tcWrite(t, root, seed, n, s, w, m)
	}
}

func TestTConfirmComplete(t *testing.T) {
	root := t.TempDir()
	// d = ((sa-na)-(sf-nf))/80 : +, +, +, -, 0, +, +, +   => P=6 N=1 Z=1
	cfg := [][4]int{{50, 40, 45, 45}, {60, 40, 50, 40}, {50, 45, 40, 40}, {40, 50, 50, 40}, {45, 45, 45, 45}, {55, 40, 40, 40}, {50, 40, 45, 40}, {60, 50, 50, 45}}
	for i, c := range cfg {
		tcSeedData(t, root, int64(19+i), c[0], c[1], c[2], c[3], "", nil)
	}
	o := TConfirm(root, []int64{19, 20, 21, 22, 23, 24, 25, 26}, []int{6, 7})
	if o.Status != "COMPLETE_EXPLORATORY" || o.P != 6 || o.N != 1 || o.Z != 1 {
		t.Fatalf("got %+v", o)
	}
	if want := SignTestP(6, 1); o.SignP != want {
		t.Fatalf("p %v want %v", o.SignP, want)
	}
}

func TestTConfirmIncompleteCases(t *testing.T) {
	cases := map[string]struct {
		missing string
		net     string
		mut     func(*TransferOut)
	}{
		"missing stage":        {missing: "F"},
		"seen!=new counts":     {net: "A", mut: func(o *TransferOut) { o.NCases["new"] = 79 }},
		"list fingerprint":     {net: "F", mut: func(o *TransferOut) { o.ListSHA["new"] = "OTHER" }},
		"counts across nets":   {net: "pre", mut: func(o *TransferOut) { o.NCases["seen"], o.NCases["new"] = 79, 79 }},
		"levels differ":        {net: "A", mut: func(o *TransferOut) { o.Levels = []int{6, 7, 8} }},
		"seen not verified":    {net: "A", mut: func(o *TransferOut) { f := false; o.SeenMatches = &f }},
		"train-F not verified": {net: "F", mut: func(o *TransferOut) { o.SeenMatchesF = nil }},
		"wrong seed":           {net: "A", mut: func(o *TransferOut) { o.Seed = 99 }},
		"zero cases":           {net: "A", mut: func(o *TransferOut) { o.NCases["seen"], o.NCases["new"] = 0, 0 }},
	}
	for name, c := range cases {
		root := t.TempDir()
		for i := int64(0); i < 8; i++ {
			sd := 19 + i
			mn, mm := "", func(*TransferOut) {}
			if sd == 22 { // one broken seed is enough to make everything INCOMPLETE (no n reduction)
				mn, mm = c.net, c.mut
			}
			tcSeedData(t, root, sd, 50, 40, 45, 45, mn, mm)
			if sd == 22 && c.missing != "" {
				os.Remove(filepath.Join(root, "seed-22", "transfer-"+c.missing+".json"))
			}
		}
		o := TConfirm(root, []int64{19, 20, 21, 22, 23, 24, 25, 26}, []int{6, 7})
		if o.Status != "INCOMPLETE" || o.P+o.N+o.Z != 0 {
			t.Fatalf("%s: got %s P=%d N=%d Z=%d", name, o.Status, o.P, o.N, o.Z)
		}
		bad := 0
		for _, s := range o.Seeds {
			if !s.Complete {
				bad++
			}
		}
		if bad != 1 {
			t.Fatalf("%s: %d incomplete seeds, want exactly 1", name, bad)
		}
	}
}

func TestTConfirmMissingSeedDir(t *testing.T) {
	o := TConfirm(t.TempDir(), []int64{19}, []int{6, 7})
	if o.Status != "INCOMPLETE" {
		t.Fatal("empty root must be INCOMPLETE")
	}
}
