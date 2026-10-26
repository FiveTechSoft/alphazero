//go:build !js

package main

// A1C analysis: paired A-F per seed, exact two-sided sign test (zeros excluded,
// effective n = P+N), exact sign-flip permutation test as secondary. Unit = seed.
// A missing or censored pair makes the primary result INCOMPLETE and NOT
// confirmatory: n is never reduced, nothing is filled in, no p-value is offered as
// the primary. `mathzero d1a1confirm -root R -seeds 19,...,26` reads
// the 5 stage json and status_* files of each R/seed-S. It evaluates nothing.

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type A1CSeed struct {
	Seed     int64  `json:"seed"`
	Complete bool   `json:"complete"`
	Problem  string `json:"problem,omitempty"` // why not complete
	Pre      int    `json:"pre,omitempty"`
	SolvedA  int    `json:"solved_A,omitempty"`
	SolvedF  int    `json:"solved_F,omitempty"`
	D        int    `json:"d,omitempty"`
}

type A1CResult struct {
	Status        string    `json:"status"` // PRIMARY_COMPLETE | INCOMPLETE_NOT_CONFIRMATORY
	ExpectedSeeds []int64   `json:"expected_seeds"`
	Seeds         []A1CSeed `json:"seeds"`
	P             int       `json:"P"`
	N             int       `json:"N"`
	Z             int       `json:"Z"`
	NEff          int       `json:"n_effective"`
	SignP         *float64  `json:"sign_test_p,omitempty"`
	Verdict       string    `json:"verdict"`
	MeanD         *float64  `json:"mean_d,omitempty"`
	PermP         *float64  `json:"signflip_perm_p,omitempty"`
	Note          string    `json:"note"`
}

// SignTestP: p = min(1, 2*P(X <= min(P,N))), X ~ Binomial(P+N, 1/2). 1 when P+N=0.
func SignTestP(pos, neg int) float64 {
	n := pos + neg
	if n == 0 {
		return 1
	}
	k := pos
	if neg < k {
		k = neg
	}
	var cum float64
	for i := 0; i <= k; i++ {
		cum += binom(n, i)
	}
	p := 2 * cum / math.Pow(2, float64(n))
	if p > 1 {
		p = 1
	}
	return p
}

func binom(n, k int) float64 {
	r := 1.0
	for i := 1; i <= k; i++ {
		r = r * float64(n-k+i) / float64(i)
	}
	return math.Round(r)
}

// SignFlipP: exact two-sided p for mean(d) under random sign flips of the d_s.
func SignFlipP(d []int) float64 {
	n := len(d)
	var obs int
	for _, x := range d {
		obs += x
	}
	if obs < 0 {
		obs = -obs
	}
	ge := 0
	for m := 0; m < 1<<n; m++ {
		s := 0
		for i := 0; i < n; i++ {
			if m>>i&1 == 1 {
				s += d[i]
			} else {
				s -= d[i]
			}
		}
		if s < 0 {
			s = -s
		}
		if s >= obs {
			ge++
		}
	}
	return float64(ge) / float64(int(1)<<n)
}

func AnalyzeA1Confirm(expected []int64, seeds []A1CSeed) A1CResult {
	r := A1CResult{ExpectedSeeds: expected, Seeds: seeds}
	have := map[int64]A1CSeed{}
	for _, s := range seeds {
		have[s.Seed] = s
	}
	complete := len(expected) > 0
	var ds []int
	for _, e := range expected {
		s, ok := have[e]
		if !ok || !s.Complete {
			complete = false
			continue
		}
		ds = append(ds, s.D)
		switch {
		case s.D > 0:
			r.P++
		case s.D < 0:
			r.N++
		default:
			r.Z++
		}
	}
	r.NEff = r.P + r.N
	if len(ds) > 0 {
		var sum int
		for _, x := range ds {
			sum += x
		}
		m := float64(sum) / float64(len(ds))
		r.MeanD = &m
	}
	if !complete {
		r.Status = "INCOMPLETE_NOT_CONFIRMATORY"
		r.Verdict = "incomplete"
		r.Note = "at least one expected A-F pair is missing or censored: no primary p, no conclusion about A-F; the counts above are descriptive over completed pairs only and may be biased by what was censored"
		return r
	}
	r.Status = "PRIMARY_COMPLETE"
	p := SignTestP(r.P, r.N)
	pp := SignFlipP(ds)
	r.SignP, r.PermP = &p, &pp
	switch {
	case r.NEff == 0:
		r.Verdict = "not_informative"
	case p < 0.05 && r.P > r.N:
		r.Verdict = "adaptive_A_greater_than_F"
	case p < 0.05 && r.N > r.P:
		r.Verdict = "frozen_F_greater_than_A"
	default:
		r.Verdict = "no_advantage_detected"
	}
	r.Note = "nominal, conditional on bank G3 and this configuration; low power; not equivalence; not general learning; not TUI"
	return r
}

func countSolved(rs []CaseResult) int {
	n := 0
	for _, c := range rs {
		if c.Solved {
			n++
		}
	}
	return n
}

func a1cReadStage(dir, name string) (stageFile, error) {
	var sf stageFile
	b, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return sf, err
	}
	if err := json.Unmarshal(b, &sf); err != nil {
		return sf, err
	}
	if sf.Stage != name {
		return sf, fmt.Errorf("%s: stage field is %q", name, sf.Stage)
	}
	return sf, nil
}

var a1cStages = []string{"pre-g3", "train-A", "train-F", "post-A", "post-F"}

// a1cBankIDs: the ids the G3 bank must carry, in order (generator ids embed the bank seed).
func a1cBankIDs() []string {
	var ids []string
	for _, k := range []int{6, 7, 8} {
		for i := 0; i < 30; i++ {
			ids = append(ids, fmt.Sprintf("g1-%d-L%d-%d", G3Seed, k, i))
		}
	}
	return ids
}

func a1cIDsOK(rs []CaseResult) bool {
	ids := a1cBankIDs()
	if len(rs) != len(ids) {
		return false
	}
	for i := range rs {
		if rs[i].ID != ids[i] {
			return false
		}
	}
	return true
}

// a1cBankFromJSON reads the frozen G3 bank file written by `d1g3bank` (no regeneration):
// seed, 90 rendered cases (level|start|target), recomputes BankDigest from the parsed
// expressions and requires it to equal both the file's digest and G3FrozenDigest, and
// returns the Key set of all Starts and Targets (the ExcludeKeys of every stage).
func a1cBankFromJSON(path string) (string, map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var f struct {
		Seed     int64    `json:"seed"`
		Cases    int      `json:"cases"`
		Digest   string   `json:"digest"`
		Rendered []string `json:"rendered"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return "", nil, err
	}
	if f.Seed != G3Seed || f.Cases != 90 || len(f.Rendered) != 90 {
		return "", nil, fmt.Errorf("bank file: seed %d cases %d rendered %d (want %d, 90, 90)", f.Seed, f.Cases, len(f.Rendered), G3Seed)
	}
	var bank []BankCase
	for _, r := range f.Rendered {
		p := strings.SplitN(r, "|", 3)
		if len(p) != 3 {
			return "", nil, fmt.Errorf("bad rendered case %q", r)
		}
		lv, err := strconv.Atoi(p[0])
		if err != nil {
			return "", nil, err
		}
		a, errA := ParsePrefix(p[1])
		c, errB := ParsePrefix(p[2])
		if errA != nil || errB != nil {
			return "", nil, fmt.Errorf("unparseable case %q", r)
		}
		bank = append(bank, BankCase{EvalCase{Start: a, Target: c}, lv})
	}
	d := BankDigest(bank)
	if d != f.Digest || d != G3FrozenDigest {
		return "", nil, fmt.Errorf("bank digest recomputed %s, file %s, frozen %s: must be equal", d, f.Digest, G3FrozenDigest)
	}
	return d, G1KeySetOf(bank), nil
}

// a1cExpectedFingerprint is the fingerprint every stage of this seed must carry
// (A1Params of the frozen A1C configuration: schedule 2..7, 40 per round, caps 100/100).
func a1cExpectedFingerprint(seed int64, digest string, keys map[string]bool) string {
	return A1Params{Seed: seed, Schedule: A1Schedule, Per: 40, CapT: 100, CapE: 100, BankDigest: digest, ExcludeKeys: keys}.fingerprint()
}

// a1cLoadSeed marks a seed complete only if ALL of these hold (all implemented here):
// the 5 status_* files (pre-g3, train-A, train-F, post-A, post-F) exist and say
// outcome=COMPLETED; the 5 stage json files exist and carry their own stage name; one
// common fingerprint; pre/post results carry exactly the 90 G3 case ids in order and
// post ids equal pre ids; post weights hash = train weights hash per arm; arm F guide
// hash before = after = the initial network hash (pre-g3) and arm A has no guide
// hash; the rendered training problems are identical in A and F (240); binary.sha256
// and budget_used_s are present. Not verified here (the binary verifies them at each
// stage): nothing about the fingerprint any more; it is recomputed from the frozen bank file.
func a1cLoadSeed(root string, seed int64, wantFP string) (A1CSeed, string, string, string) {
	out := A1CSeed{Seed: seed}
	dir := filepath.Join(root, fmt.Sprintf("seed-%d", seed))
	bad := func(m string) (A1CSeed, string, string, string) { out.Problem = m; return out, "", "", "" }
	for _, st := range a1cStages {
		b, err := os.ReadFile(filepath.Join(dir, "status_"+st))
		if err != nil {
			return bad("status_" + st + " missing")
		}
		if !strings.Contains(string(b), "outcome=COMPLETED") {
			return bad("status_" + st + " not COMPLETED")
		}
	}
	sf := map[string]stageFile{}
	first := ""
	for _, st := range a1cStages {
		x, err := a1cReadStage(dir, st)
		if err != nil {
			return bad(err.Error())
		}
		if first == "" {
			first = x.Fingerprint
		}
		if x.Fingerprint == "" || x.Fingerprint != first {
			return bad("fingerprints differ at " + st)
		}
		if x.Fingerprint != wantFP {
			return bad("fingerprint at " + st + " is not the expected one for this seed and the G3 configuration")
		}
		sf[st] = x
	}
	pre, ta, tf, pa, pf := sf["pre-g3"], sf["train-A"], sf["train-F"], sf["post-A"], sf["post-F"]
	for _, x := range []stageFile{pre, pa, pf} {
		if !a1cIDsOK(x.Results) {
			return bad(x.Stage + ": case ids are not the G3 bank")
		}
	}
	if ta.Weights == "" || tf.Weights == "" || pa.Weights != ta.Weights || pf.Weights != tf.Weights {
		return bad("post weights hash differs from train weights hash")
	}
	if pre.Weights == "" || tf.GuideSHAPre != pre.Weights || tf.GuideSHAPost != pre.Weights || ta.GuideSHAPre != "" || ta.GuideSHAPost != "" {
		return bad("guide hash check failed (F guide before/after must equal the initial network; A has none)")
	}
	if ta.Train == nil || tf.Train == nil || len(ta.Train.RenderedProblems) != 240 || strings.Join(ta.Train.RenderedProblems, "\n") != strings.Join(tf.Train.RenderedProblems, "\n") {
		return bad("training problems not identical across arms (or not 240)")
	}
	bin, e1 := os.ReadFile(filepath.Join(dir, "binary.sha256"))
	bud, e2 := os.ReadFile(filepath.Join(dir, "budget_used_s"))
	if e1 != nil || e2 != nil || strings.TrimSpace(string(bin)) == "" {
		return bad("binary.sha256 or budget_used_s missing")
	}
	out.Complete, out.Pre, out.SolvedA, out.SolvedF = true, countSolved(pre.Results), countSolved(pa.Results), countSolved(pf.Results)
	out.D = out.SolvedA - out.SolvedF
	return out, pre.Fingerprint + "|" + pre.Weights, strings.TrimSpace(string(bin)), strings.TrimSpace(string(bud))
}

// a1cLoadAll adds the cross-seed checks: a fingerprint, an initial network or a train
// weights hash shared by two seeds means a reused directory (both are marked not
// complete), one binary for all seeds, and total budget used <= budget.
func a1cLoadAll(root string, exp []int64, budget int, wantFP func(int64) string) []A1CSeed {
	var ss []A1CSeed
	fp, ws, bins := map[string][]int{}, map[string][]int{}, map[string]bool{}
	total := 0
	keys := make([]string, len(exp))
	for i, e := range exp {
		s, key, bin, bud := a1cLoadSeed(root, e, wantFP(e))
		ss = append(ss, s)
		keys[i] = key
		if s.Complete {
			kk := strings.SplitN(key, "|", 2)
			fp[kk[0]] = append(fp[kk[0]], i)
			ws[kk[1]] = append(ws[kk[1]], i)
			bins[bin] = true
			if v, err := strconv.Atoi(bud); err == nil {
				total += v
			} else {
				ss[i].Complete, ss[i].Problem = false, "budget_used_s not an integer"
			}
		}
	}
	for _, m := range []map[string][]int{fp, ws} {
		for _, idx := range m {
			if len(idx) > 1 {
				for _, i := range idx {
					ss[i].Complete, ss[i].Problem = false, "fingerprint or initial network shared with another seed (reused directory?)"
				}
			}
		}
	}
	if len(bins) > 1 || total > budget {
		for i := range ss {
			if ss[i].Complete {
				ss[i].Complete, ss[i].Problem = false, fmt.Sprintf("more than one binary across seeds (%d) or total budget %ds > %ds", len(bins), total, budget)
			}
		}
	}
	return ss
}

func init() {
	if len(os.Args) > 1 && os.Args[1] == "d1a1confirm" {
		fs := flag.NewFlagSet("d1a1confirm", flag.ExitOnError)
		root := fs.String("root", "", "directory containing seed-N/")
		seedsF := fs.String("seeds", "19,20,21,22,23,24,25,26", "expected seeds")
		budget := fs.Int("budget", 5400, "total budget in seconds (sum of budget_used_s over seeds)")
		bankF := fs.String("bank", "", "frozen G3 bank file (d1g3bank output); its digest must equal G3FrozenDigest")
		fs.Parse(os.Args[2:])
		var exp []int64
		for _, s := range strings.Split(*seedsF, ",") {
			v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				fmt.Fprintln(os.Stderr, "bad -seeds")
				os.Exit(2)
			}
			exp = append(exp, v)
		}
		digest, keys, err := a1cBankFromJSON(*bankF)
		if err != nil {
			fmt.Fprintln(os.Stderr, "d1a1confirm: bank:", err)
			os.Exit(2)
		}
		ss := a1cLoadAll(*root, exp, *budget, func(sd int64) string { return a1cExpectedFingerprint(sd, digest, keys) })
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		enc.Encode(AnalyzeA1Confirm(exp, ss))
		os.Exit(0)
	}
}
