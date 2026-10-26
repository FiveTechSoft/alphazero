package main

// d1a1transferconfirm: PREFIXED analysis of the exploratory seen/new evaluation. No solver, no weights.
// Per seed: needs transfer-pre/A/F.json, all with the same levels, cap, case counts (seen==new), identical list
// fingerprints across the three nets (equal counts alone are not accepted). g_arm = solved_seen/n_seen - solved_new/n_new
// (levels 6+7 pooled); d = g_A - g_F. Exact two-sided sign test on the n=8 d values, zeros excluded. Any missing or
// inconsistent seed => INCOMPLETE (no n reduction, no refill). Exploratory: seeds and G3 already exposed; measures the
// guide+data package; not causal; not general learning.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type TConfirmSeed struct {
	Seed     int64   `json:"seed"`
	Complete bool    `json:"complete"`
	Problem  string  `json:"problem,omitempty"`
	N        int     `json:"n_cases_per_list,omitempty"`
	GA       float64 `json:"g_A,omitempty"`
	GF       float64 `json:"g_F,omitempty"`
	D        float64 `json:"d_gA_minus_gF,omitempty"`
	SeenA    float64 `json:"rate_seen_A,omitempty"`
	NewA     float64 `json:"rate_new_A,omitempty"`
	SeenF    float64 `json:"rate_seen_F,omitempty"`
	NewF     float64 `json:"rate_new_F,omitempty"`
	SeenPre  float64 `json:"rate_seen_pre,omitempty"`
	NewPre   float64 `json:"rate_new_pre,omitempty"`
}

type TConfirmOut struct {
	Status string         `json:"status"`
	Seeds  []TConfirmSeed `json:"seeds"`
	P      int            `json:"P,omitempty"`
	N      int            `json:"N,omitempty"`
	Z      int            `json:"Z,omitempty"`
	SignP  float64        `json:"sign_test_p,omitempty"`
	MeanD  float64        `json:"mean_d,omitempty"`
	Note   string         `json:"note"`
}

func tcLoad(dir, net string) (*TransferOut, error) {
	b, err := os.ReadFile(filepath.Join(dir, "transfer-"+net+".json"))
	if err != nil {
		return nil, fmt.Errorf("missing transfer-%s.json", net)
	}
	var o TransferOut
	if err := json.Unmarshal(b, &o); err != nil {
		return nil, fmt.Errorf("transfer-%s.json unreadable", net)
	}
	return &o, nil
}

func tcSeed(root string, seed int64, levels []int) TConfirmSeed {
	r := TConfirmSeed{Seed: seed}
	bad := func(m string) TConfirmSeed { r.Problem = m; return r }
	dir := filepath.Join(root, fmt.Sprintf("seed-%d", seed))
	var os_ [3]*TransferOut
	for i, n := range []string{"pre", "A", "F"} {
		o, err := tcLoad(dir, n)
		if err != nil {
			return bad(err.Error())
		}
		if o.Seed != seed || o.Net != n || o.Cap != 100 {
			return bad("seed/net/cap field mismatch in transfer-" + n)
		}
		if o.SeenMatches == nil || !*o.SeenMatches || o.SeenMatchesF == nil || !*o.SeenMatchesF {
			return bad("seen list not verified against stored lists of both arms in transfer-" + n)
		}
		if fmt.Sprint(o.Levels) != fmt.Sprint(levels) {
			return bad("levels differ from the prefixed levels in transfer-" + n)
		}
		if o.NCases["seen"] <= 0 || o.NCases["seen"] != o.NCases["new"] {
			return bad("seen/new case counts differ or are zero in transfer-" + n)
		}
		s, ok1 := o.Solved[n]["seen"]
		w, ok2 := o.Solved[n]["new"]
		if !ok1 || !ok2 || len(s) != len(levels) || len(w) != len(levels) {
			return bad("solved counts missing in transfer-" + n)
		}
		os_[i] = o
	}
	for i := 1; i < 3; i++ { // same lists, not just the same counts
		if os_[i].ListSHA["seen"] != os_[0].ListSHA["seen"] || os_[i].ListSHA["new"] != os_[0].ListSHA["new"] || os_[0].ListSHA["seen"] == "" || os_[0].ListSHA["new"] == "" {
			return bad("list fingerprints differ across pre/A/F")
		}
		if os_[i].NCases["seen"] != os_[0].NCases["seen"] {
			return bad("case counts differ across nets")
		}
	}
	n := float64(os_[0].NCases["seen"])
	sum := func(v []int) float64 {
		t := 0
		for _, x := range v {
			t += x
		}
		return float64(t)
	}
	rate := func(o *TransferOut, net, l string) float64 { return sum(o.Solved[net][l]) / n }
	r.N = os_[0].NCases["seen"]
	r.SeenPre, r.NewPre = rate(os_[0], "pre", "seen"), rate(os_[0], "pre", "new")
	r.SeenA, r.NewA = rate(os_[1], "A", "seen"), rate(os_[1], "A", "new")
	r.SeenF, r.NewF = rate(os_[2], "F", "seen"), rate(os_[2], "F", "new")
	r.GA, r.GF = r.SeenA-r.NewA, r.SeenF-r.NewF
	r.D = r.GA - r.GF
	r.Complete = true
	return r
}

func TConfirm(root string, seeds []int64, levels []int) TConfirmOut {
	out := TConfirmOut{Note: "exploratory; seeds and G3 already exposed; package guide+data, not pure guide effect; levels 6+7 pooled; paired by generation rule, equal real difficulty not guaranteed; new pairs are on average longer than seen; not causal; not general learning; not TUI"}
	allOK := true
	for _, s := range seeds {
		c := tcSeed(root, s, levels)
		out.Seeds = append(out.Seeds, c)
		if !c.Complete {
			allOK = false
		}
	}
	if !allOK {
		out.Status = "INCOMPLETE"
		return out
	}
	sum := 0.0
	for _, c := range out.Seeds {
		sum += c.D
		switch {
		case c.D > 1e-12:
			out.P++
		case c.D < -1e-12:
			out.N++
		default:
			out.Z++
		}
	}
	out.MeanD = sum / float64(len(out.Seeds))
	out.SignP = SignTestP(out.P, out.N)
	out.Status = "COMPLETE_EXPLORATORY"
	return out
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1a1transferconfirm" {
		fs := flag.NewFlagSet("d1a1transferconfirm", flag.ExitOnError)
		root := fs.String("root", "", "transfer output root containing seed-N/")
		seedsF := fs.String("seeds", "19,20,21,22,23,24,25,26", "expected seeds")
		lv := fs.String("levels", "6,7", "levels pooled")
		fs.Parse(os.Args[2:])
		var seeds []int64
		for _, s := range strings.Split(*seedsF, ",") {
			v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				fmt.Fprintln(os.Stderr, "bad -seeds")
				os.Exit(2)
			}
			seeds = append(seeds, v)
		}
		var levels []int
		for _, s := range strings.Split(*lv, ",") {
			var v int
			fmt.Sscan(s, &v)
			levels = append(levels, v)
		}
		if *root == "" {
			fmt.Fprintln(os.Stderr, "need -root")
			os.Exit(2)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		enc.Encode(TConfirm(*root, seeds, levels))
		os.Exit(0)
	}
}
