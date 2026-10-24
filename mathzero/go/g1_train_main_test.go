package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func smallG1(t *testing.T) ([]BankCase, G1GenStats, G1Rule) {
	rule := G1Rule{Seed: 5, PerLevel: 3, Levels: []int{4, 5}}
	b, st, err := GenBankG1(rule, G1PriorKeys())
	if err != nil {
		t.Fatal(err)
	}
	return b, st, rule
}

func TestG1StagedRunExcludesBankAndReports(t *testing.T) {
	bank, st, rule := smallG1(t)
	dir := t.TempDir()
	d := BankDigest(bank)
	var sink bytes.Buffer
	for _, s := range []string{"pre-g1", "train", "post-g1"} {
		if err := RunG1StageWithBank(s, dir, 3, 2, 8, 20, 20, bank, st, d, rule, &sink, &sink); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	var out bytes.Buffer
	if err := RunG1StageWithBank("report", dir, 3, 2, 8, 20, 20, bank, st, d, rule, &out, &sink); err != nil {
		t.Fatal(err)
	}
	var rep D1TrainReport
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	ov := rep.Overlap["g1"]
	if rep.G1 == nil || rep.G1.BankDigest != d || len(rep.Banks) != 1 || rep.Banks[0].Name != "g1" || ov.StartInEval != 0 || ov.TargetInEval != 0 || ov.PairInEval != 0 || ov.TrainProblems != 16 {
		t.Fatalf("report %+v overlap %+v", rep.G1, ov)
	}
	// independent check of the recorded rendered problems against the bank
	bn := map[string]bool{}
	for _, c := range bank {
		bn[indepNorm(c.Start)], bn[indepNorm(c.Target)] = true, true
	}
	if len(rep.Train.RenderedProblems) != 16 {
		t.Fatalf("rendered problems %d", len(rep.Train.RenderedProblems))
	}
	for _, pr := range rep.Train.RenderedProblems {
		x := strings.SplitN(pr, "|", 2)
		a, _ := ParsePrefix(x[0])
		b, _ := ParsePrefix(x[1])
		if bn[indepNorm(a)] || bn[indepNorm(b)] {
			t.Fatalf("training problem in G1 bank: %s", pr)
		}
	}
	if rep.Train.Rounds[0].Attempts < rep.Train.Rounds[0].Problems {
		t.Fatal("attempts not recorded")
	}
}

func TestG1RefusesWrongDigestOverwriteAndOtherConfig(t *testing.T) {
	bank, st, rule := smallG1(t)
	dir := t.TempDir()
	d := BankDigest(bank)
	var sink bytes.Buffer
	if err := RunG1StageWithBank("pre-g1", dir, 3, 2, 8, 20, 20, bank, st, "deadbeef", rule, &sink, &sink); err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("wrong digest accepted: %v", err)
	}
	if err := RunG1StageWithBank("pre-g1", dir, 3, 2, 8, 20, 20, bank, st, d, rule, &sink, &sink); err != nil {
		t.Fatal(err)
	}
	if err := RunG1StageWithBank("pre-g1", dir, 3, 2, 8, 20, 20, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("overwrote a stage")
	}
	if err := RunG1StageWithBank("post-g1", dir, 3, 5, 8, 20, 20, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("post accepted a train stage from another configuration")
	}
	if err := RunG1StageWithBank("pre-primary", dir, 3, 2, 8, 20, 20, bank, st, d, rule, &sink, &sink); err == nil {
		t.Fatal("accepted a P3 stage name in G1")
	}
	if _, err := os.Stat(filepath.Join(dir, "pre-g1.json")); err != nil {
		t.Fatal(err)
	}
}

func TestG1FrozenBankMatchesFrozenDigestAndConstants(t *testing.T) {
	rule := G1Rule{Seed: G1Seed, PerLevel: 30, Levels: []int{6, 7, 8}}
	b, st, err := GenBankG1(rule, G1PriorKeys())
	if err != nil || len(b) != 90 || BankDigest(b) != G1FrozenDigest || G1Seed != 20261024 {
		t.Fatalf("err=%v n=%d digest=%s", err, len(b), BankDigest(b))
	}
	if st.PerLevel[0].Attempts != 56 || st.PerLevel[1].Attempts != 71 || st.PerLevel[2].Attempts != 58 || st.PriorKeys != 1559 {
		t.Fatalf("statistics changed: %+v", st)
	}
}

func TestP3FingerprintFormatUnchanged(t *testing.T) {
	// The P3 fingerprint of the stored pilot (seed 101) must be reproducible. The
	// expected value is read from a state file only if present (local evidence).
	p := stageParams{Seed: 101, Rounds: 6, Per: 40, CapT: 100, CapE: 100, Lv: []int{6, 7, 8}, SecSeed: 9001, SecPer: 30}
	banks := map[string][]BankCase{"primary": buildPrimary([]int{6, 7, 8})}
	banks["secondary"] = GenBank(BankRule{Seed: 9001, PerLevel: 30, Levels: []int{6, 7, 8}})
	got := p.fingerprint(banks, []string{"primary", "secondary"})
	b, err := os.ReadFile("/home/sandbox/p3-pilot-101/pre-primary.json")
	if err != nil {
		t.Skip("no stored pilot state here")
	}
	var sf struct{ Fingerprint string }
	json.Unmarshal(b, &sf)
	if sf.Fingerprint != got {
		t.Fatalf("P3 fingerprint changed: %s vs stored %s", got, sf.Fingerprint)
	}
}
