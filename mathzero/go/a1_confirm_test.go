//go:build !js

package main

// Artificial tests: no training, no checkpoint evaluation, no bank generation.

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSignTestTable(t *testing.T) {
	cases := []struct {
		p, n int
		want float64
	}{{8, 0, 0.0078125}, {7, 1, 0.0703125}, {6, 2, 0.2890625}, {5, 3, 0.7265625}, {4, 4, 1},
		{7, 0, 0.015625}, {6, 1, 0.125}, {6, 0, 0.03125}, {5, 0, 0.0625}, {4, 0, 0.125}, {0, 0, 1}, {0, 8, 0.0078125}}
	for _, c := range cases {
		if g := SignTestP(c.p, c.n); math.Abs(g-c.want) > 1e-12 {
			t.Fatalf("P=%d N=%d got %v want %v", c.p, c.n, g, c.want)
		}
	}
}

func seedsOf(ds ...int) []A1CSeed {
	var out []A1CSeed
	for i, d := range ds {
		out = append(out, A1CSeed{Seed: int64(19 + i), Complete: true, SolvedA: 50 + d, SolvedF: 50, D: d})
	}
	return out
}

var exp8 = []int64{19, 20, 21, 22, 23, 24, 25, 26}

func TestZerosExcludedAndEffectiveN(t *testing.T) {
	r := AnalyzeA1Confirm(exp8, seedsOf(3, 4, 5, 6, 7, 8, 9, 0)) // 7:0 with one zero
	if r.NEff != 7 || r.Z != 1 || r.SignP == nil || math.Abs(*r.SignP-0.015625) > 1e-12 || r.Verdict != "adaptive_A_greater_than_F" {
		t.Fatalf("%+v", r)
	}
	r = AnalyzeA1Confirm(exp8, seedsOf(-3, -4, -5, -6, -7, -8, 0, 0)) // 6:0 two zeros, F better
	if r.NEff != 6 || math.Abs(*r.SignP-0.03125) > 1e-12 || r.Verdict != "frozen_F_greater_than_A" {
		t.Fatalf("%+v", r)
	}
	r = AnalyzeA1Confirm(exp8, seedsOf(1, 1, 1, 1, 1, 1, 1, -1)) // 7:1 not significant
	if r.Verdict != "no_advantage_detected" {
		t.Fatalf("%+v", r)
	}
	r = AnalyzeA1Confirm(exp8, seedsOf(0, 0, 0, 0, 0, 0, 0, 0))
	if r.Verdict != "not_informative" || r.NEff != 0 {
		t.Fatalf("%+v", r)
	}
	// A1 historical d (-9,+1,-5,+4,-2,-2,+14,+12): 4:4 p=1
	r = AnalyzeA1Confirm(exp8, seedsOf(-9, 1, -5, 4, -2, -2, 14, 12))
	if r.P != 4 || r.N != 4 || *r.SignP != 1 || r.Verdict != "no_advantage_detected" {
		t.Fatalf("%+v", r)
	}
}

func TestCensoredPairMakesPrimaryIncomplete(t *testing.T) {
	ss := seedsOf(9, 9, 9, 9, 9, 9, 9, 9)
	ss[3].Complete, ss[3].Problem = false, "status_post-A not COMPLETED"
	r := AnalyzeA1Confirm(exp8, ss)
	if r.Status != "INCOMPLETE_NOT_CONFIRMATORY" || r.SignP != nil || r.PermP != nil || r.Verdict != "incomplete" {
		t.Fatalf("%+v", r)
	}
	if len(r.ExpectedSeeds) != 8 {
		t.Fatal("n must not shrink")
	}
	// missing seed entirely
	r = AnalyzeA1Confirm(exp8, seedsOf(9, 9, 9))
	if r.Status != "INCOMPLETE_NOT_CONFIRMATORY" || r.SignP != nil {
		t.Fatalf("%+v", r)
	}
}

func TestSignFlipExact(t *testing.T) {
	// all +1 on n=3: 2 of 8 sign patterns give |sum|>=3 => 0.25
	if p := SignFlipP([]int{1, 1, 1}); math.Abs(p-0.25) > 1e-12 {
		t.Fatal(p)
	}
}

const testBankFile = "../../pilots/a1c/g3bank.json"

func testBank(t *testing.T) (string, map[string]bool) {
	t.Helper()
	d, k, err := a1cBankFromJSON(testBankFile)
	if err != nil {
		t.Fatal(err)
	}
	return d, k
}

func fpFor(t *testing.T, seed int64) string {
	d, k := testBank(t)
	return a1cExpectedFingerprint(seed, d, k)
}

// fixture writes a coherent artificial seed directory (no network, no evaluation).
func fixture(t *testing.T, root string, seed int64, tweak func(map[string]*stageFile, map[string]string)) {
	t.Helper()
	dir := filepath.Join(root, "seed-"+strconv.FormatInt(seed, 10))
	os.MkdirAll(dir, 0o755)
	ids := a1cBankIDs()
	res := func(n int) []CaseResult {
		var r []CaseResult
		for i, id := range ids {
			r = append(r, CaseResult{ID: id, Solved: i < n})
		}
		return r
	}
	fp := fpFor(t, seed)
	w0 := "w0-" + strconv.FormatInt(seed, 10)
	var probs []string
	for i := 0; i < 240; i++ {
		probs = append(probs, "p"+strconv.Itoa(i))
	}
	sfs := map[string]*stageFile{
		"pre-g3":  {Fingerprint: fp, Stage: "pre-g3", Results: res(40), Weights: w0},
		"train-A": {Fingerprint: fp, Stage: "train-A", Weights: "wa", Train: &TrainLog{RenderedProblems: append([]string(nil), probs...)}},
		"train-F": {Fingerprint: fp, Stage: "train-F", Weights: "wf", GuideSHAPre: w0, GuideSHAPost: w0, Train: &TrainLog{RenderedProblems: append([]string(nil), probs...)}},
		"post-A":  {Fingerprint: fp, Stage: "post-A", Results: res(60), Weights: "wa"},
		"post-F":  {Fingerprint: fp, Stage: "post-F", Results: res(50), Weights: "wf"},
	}
	files := map[string]string{"binary.sha256": "bin1", "budget_used_s": "300"}
	for _, st := range a1cStages {
		files["status_"+st] = "outcome=COMPLETED guard_rc=0"
	}
	if tweak != nil {
		tweak(sfs, files)
	}
	for st, sf := range sfs {
		if sf != nil {
			b, _ := json.Marshal(sf)
			os.WriteFile(filepath.Join(dir, st+".json"), b, 0o644)
		}
	}
	for n, c := range files {
		if c != "" {
			os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644)
		}
	}
}

func TestSeedLoaderComplete(t *testing.T) {
	root := t.TempDir()
	fixture(t, root, 19, nil)
	s, _, _, _ := a1cLoadSeed(root, 19, fpFor(t, 19))
	if !s.Complete || s.D != 10 || s.Pre != 40 || s.SolvedA != 60 || s.SolvedF != 50 {
		t.Fatalf("%+v", s)
	}
}

func TestSeedLoaderRejects(t *testing.T) {
	cases := map[string]func(map[string]*stageFile, map[string]string){
		"missing status":      func(_ map[string]*stageFile, f map[string]string) { f["status_train-F"] = "" },
		"partial (censored)":  func(_ map[string]*stageFile, f map[string]string) { f["status_post-A"] = "outcome=CENSORED_TIMEOUT" },
		"missing stage json":  func(s map[string]*stageFile, _ map[string]string) { s["post-F"] = nil },
		"fingerprint differs": func(s map[string]*stageFile, _ map[string]string) { s["post-F"].Fingerprint = "other" },
		"wrong bank ids":      func(s map[string]*stageFile, _ map[string]string) { s["post-A"].Results[3].ID = "g1-20261025-L6-3" },
		"short results":       func(s map[string]*stageFile, _ map[string]string) { s["post-F"].Results = s["post-F"].Results[:89] },
		"stage name":          func(s map[string]*stageFile, _ map[string]string) { s["post-A"].Stage = "post-F" },
		"post weights":        func(s map[string]*stageFile, _ map[string]string) { s["post-A"].Weights = "mutated" },
		"guide changed":       func(s map[string]*stageFile, _ map[string]string) { s["train-F"].GuideSHAPost = "x" },
		"guide != initial": func(s map[string]*stageFile, _ map[string]string) {
			s["train-F"].GuideSHAPre, s["train-F"].GuideSHAPost = "y", "y"
		},
		"A has guide hash": func(s map[string]*stageFile, _ map[string]string) { s["train-A"].GuideSHAPre = "z" },
		"problems differ":  func(s map[string]*stageFile, _ map[string]string) { s["train-F"].Train.RenderedProblems[7] = "q" },
		"no binary hash":   func(_ map[string]*stageFile, f map[string]string) { f["binary.sha256"] = "" },
	}
	for name, tw := range cases {
		root := t.TempDir()
		fixture(t, root, 19, tw)
		if s, _, _, _ := a1cLoadSeed(root, 19, fpFor(t, 19)); s.Complete {
			t.Fatalf("%s must not be complete", name)
		}
	}
	if s, _, _, _ := a1cLoadSeed(t.TempDir(), 19, fpFor(t, 19)); s.Complete {
		t.Fatal("empty dir")
	}
}

func TestLoadAllReusedDirectoryAndBudget(t *testing.T) {
	root := t.TempDir()
	for s := int64(19); s <= 26; s++ {
		fixture(t, root, s, nil)
	}
	ss := a1cLoadAll(root, exp8, 5400, func(s int64) string { return fpFor(t, s) })
	for _, s := range ss {
		if !s.Complete {
			t.Fatalf("%+v", s)
		}
	}
	// seed 22 directory has its own expected fingerprint but the initial network of seed 21
	root2 := t.TempDir()
	for s := int64(19); s <= 26; s++ {
		fixture(t, root2, s, func(m map[string]*stageFile, f map[string]string) {
			if s == 22 {
				m["pre-g3"].Weights = "w0-21"
				m["train-F"].GuideSHAPre, m["train-F"].GuideSHAPost = "w0-21", "w0-21"
			}
		})
	}
	ss = a1cLoadAll(root2, exp8, 5400, func(s int64) string { return fpFor(t, s) })
	for _, s := range ss {
		if (s.Seed == 21 || s.Seed == 22) == s.Complete { // both share an initial network
			t.Fatalf("reused dir not flagged: %+v", s)
		}
	}
	if r := AnalyzeA1Confirm(exp8, ss); r.Status != "INCOMPLETE_NOT_CONFIRMATORY" {
		t.Fatal(r.Status)
	}
	// total budget above the limit invalidates everything
	ss = a1cLoadAll(root, exp8, 100, func(s int64) string { return fpFor(t, s) })
	for _, s := range ss {
		if s.Complete {
			t.Fatal("budget exceeded must not be complete")
		}
	}
	// two different binaries
	root3 := t.TempDir()
	for s := int64(19); s <= 26; s++ {
		fixture(t, root3, s, func(_ map[string]*stageFile, f map[string]string) {
			if s == 24 {
				f["binary.sha256"] = "bin2"
			}
		})
	}
	for _, s := range a1cLoadAll(root3, exp8, 5400, func(s int64) string { return fpFor(t, s) }) {
		if s.Complete {
			t.Fatal("different binaries must not be complete")
		}
	}
}

func TestG3Guards(t *testing.T) {
	if G3Seed != 20261027 {
		t.Fatal("bank seed changed")
	}
	if a1BankTag(G1Rule{Seed: G3Seed}) != "g3" || a1BankTag(G1Rule{Seed: G2Seed}) != "g2" || a1BankTag(G1Rule{Seed: 7}) != "g2" {
		t.Fatal("tags")
	}
	// the pre stage name of each bank is enforced; wrong name fails before any work
	var sink bytes.Buffer
	rule := G1Rule{Seed: G3Seed, PerLevel: 1, Levels: []int{2}}
	err := RunA1StageWithBank("pre-g2", t.TempDir(), 3, []int{2}, 1, 4, 4, nil, G1GenStats{}, "x", rule, &sink, &sink)
	if err == nil || !strings.Contains(err.Error(), "pre-g3") {
		t.Fatalf("expected pre-g3 guard, got %v", err)
	}
	rule2 := G1Rule{Seed: G2Seed, PerLevel: 1, Levels: []int{2}}
	err = RunA1StageWithBank("pre-g3", t.TempDir(), 3, []int{2}, 1, 4, 4, nil, G1GenStats{}, "x", rule2, &sink, &sink)
	if err == nil {
		t.Fatal("pre-g3 must be rejected for bank g2")
	}
	// wrong digest refuses to run
	err = RunA1StageWithBank("pre-g3", t.TempDir(), 3, []int{2}, 1, 4, 4, nil, G1GenStats{}, "wrong", rule, &sink, &sink)
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("expected digest refusal, got %v", err)
	}
}

func TestHindsightKeyFilterArtificial(t *testing.T) {
	// Artificial expressions, no network. Covers only the hindsight filter
	// (hindsightExcludedByKey). The training-problem rejection in TrainRoundsA1 is not
	// exercised here.
	a, b := EAdd(V("x"), C(1)), EMul(V("y"), C(2))
	keys := map[string]bool{Key(a): true}
	h := &HindsightPair{Start: RenderPrefix(a), Target: RenderPrefix(b)}
	if !hindsightExcludedByKey(h, keys) {
		t.Fatal("hindsight with bank Key must be dropped")
	}
	if hindsightExcludedByKey(h, map[string]bool{}) {
		t.Fatal("empty set must not drop")
	}
	if !hindsightExcludedByKey(&HindsightPair{Start: "garbage", Target: "x"}, map[string]bool{}) {
		t.Fatal("unparseable must fail closed")
	}
}

func TestBankFileVerification(t *testing.T) {
	d, k := testBank(t)
	if d != G3FrozenDigest || len(k) == 0 || len(k) > 180 {
		t.Fatalf("%s %d", d, len(k))
	}
	b, _ := os.ReadFile(testBankFile)
	write := func(s string) string {
		p := filepath.Join(t.TempDir(), "bank.json")
		os.WriteFile(p, []byte(s), 0o644)
		return p
	}
	// a changed case: recomputed digest differs from the file and from the frozen one
	var f map[string]any
	json.Unmarshal(b, &f)
	r := f["rendered"].([]any)
	r[5] = strings.Replace(r[5].(string), "|", "|add(c1,", 1) + ")"
	nb, _ := json.Marshal(f)
	if _, _, err := a1cBankFromJSON(write(string(nb))); err == nil {
		t.Fatal("tampered case accepted")
	}
	// digest field edited but cases intact
	f2 := map[string]any{}
	json.Unmarshal(b, &f2)
	f2["digest"] = "00"
	nb, _ = json.Marshal(f2)
	if _, _, err := a1cBankFromJSON(write(string(nb))); err == nil {
		t.Fatal("wrong digest field accepted")
	}
	// foreign bank seed
	f3 := map[string]any{}
	json.Unmarshal(b, &f3)
	f3["seed"] = 20261025
	nb, _ = json.Marshal(f3)
	if _, _, err := a1cBankFromJSON(write(string(nb))); err == nil {
		t.Fatal("foreign seed accepted")
	}
}

func TestExpectedFingerprintRejectsForeign(t *testing.T) {
	d, k := testBank(t)
	want := fpFor(t, 19)
	foreign := map[string]func() string{
		"foreign seed (seed 20's fingerprint)": func() string { return fpFor(t, 20) },
		"foreign bank digest (consistent)": func() string {
			return a1cExpectedFingerprint(19, "f58161c5133bffeb2838ee9b5da9a516f09022bbe7d94367c0b4c658d4bb67b6", k)
		},
		"foreign key set (consistent)": func() string { return a1cExpectedFingerprint(19, d, map[string]bool{"x": true}) },
		"foreign config (per 20)": func() string {
			return A1Params{Seed: 19, Schedule: A1Schedule, Per: 20, CapT: 100, CapE: 100, BankDigest: d, ExcludeKeys: k}.fingerprint()
		},
		"foreign caps (eval 50)": func() string {
			return A1Params{Seed: 19, Schedule: A1Schedule, Per: 40, CapT: 100, CapE: 50, BankDigest: d, ExcludeKeys: k}.fingerprint()
		},
	}
	for name, fp := range foreign {
		fpv := fp()
		if fpv == want {
			t.Fatalf("%s: equals the expected fingerprint", name)
		}
		root := t.TempDir()
		fixture(t, root, 19, func(m map[string]*stageFile, _ map[string]string) {
			for _, sf := range m { // internally consistent: all five stages carry the foreign one
				sf.Fingerprint = fpv
			}
		})
		if s, _, _, _ := a1cLoadSeed(root, 19, want); s.Complete {
			t.Fatalf("%s accepted", name)
		}
	}
	root := t.TempDir()
	fixture(t, root, 19, nil)
	if s, _, _, _ := a1cLoadSeed(root, 19, want); !s.Complete {
		t.Fatal("the expected fingerprint must be accepted")
	}
}
