//go:build !js

package main

// Staged execution of d1train for hosts where one process cannot run the whole
// pilot (call time limits). Each stage is ONE short process; state lives in a
// directory as JSON. Training is a single stage, so the optimizer state never
// crosses a process boundary; weights are stored with Go's exact float64 JSON
// round trip. A config fingerprint refuses stages mixed from different configs.
// The final report equals the single-process report (test), except timings.

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type stageParams struct {
	Seed        int64
	Rounds, Per int
	CapT, CapE  int
	Lv          []int
	SecSeed     int64
	SecPer      int
	// G1 only (zero for P3). Not part of the P3 fingerprint format.
	ExcludeKeys map[string]bool
	G1Info      *G1ReportInfo
}

func (p stageParams) fingerprint(banks map[string][]BankCase, names []string) string {
	h := sha256.New()
	// The P3 fingerprint format is preserved bit for bit: it prints exactly the
	// original fields; G1 adds its own line below.
	fmt.Fprintf(h, "%+v\n", struct {
		Seed        int64
		Rounds, Per int
		CapT, CapE  int
		Lv          []int
		SecSeed     int64
		SecPer      int
	}{p.Seed, p.Rounds, p.Per, p.CapT, p.CapE, p.Lv, p.SecSeed, p.SecPer})
	if p.ExcludeKeys != nil {
		ks := make([]string, 0, len(p.ExcludeKeys))
		for k := range p.ExcludeKeys {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		fmt.Fprintf(h, "G1 exclude-keys %d %x\n", len(ks), sha256.Sum256([]byte(strings.Join(ks, "\n"))))
	}
	for _, n := range names {
		fmt.Fprintf(h, "%s %s\n", n, BankDigest(banks[n]))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

type stageFile struct {
	Fingerprint string       `json:"fingerprint"`
	Stage       string       `json:"stage"`
	Seconds     float64      `json:"seconds"`
	Results     []CaseResult `json:"results,omitempty"`
	Weights     string       `json:"weights_sha256,omitempty"`
	Train       *TrainLog    `json:"train_log,omitempty"`
	Level       int          `json:"final_level,omitempty"`
}

func writeJSON(path string, v any) error {
	tmp := path + ".tmp"
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readStage(dir, name, fp string) (stageFile, error) {
	var sf stageFile
	b, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return sf, fmt.Errorf("stage %s missing: %v", name, err)
	}
	if err := json.Unmarshal(b, &sf); err != nil {
		return sf, err
	}
	if sf.Fingerprint != fp {
		return sf, fmt.Errorf("stage %s was produced with a different configuration", name)
	}
	return sf, nil
}

func runStage(stage, dir string, p stageParams, banks map[string][]BankCase, names []string, w, prog io.Writer) error {
	if dir == "" {
		return fmt.Errorf("-dir is required for staged runs")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fp := p.fingerprint(banks, names)
	toCases := func(b []BankCase) []EvalCase {
		cs := make([]EvalCase, len(b))
		for i, c := range b {
			cs[i] = c.EvalCase
		}
		return cs
	}
	var all []EvalCase
	for _, n := range names {
		all = append(all, toCases(banks[n])...)
	}
	weightsPath := filepath.Join(dir, "weights_post.json")
	done := func(name string, sf stageFile) error {
		sf.Fingerprint, sf.Stage = fp, name
		if _, err := os.Stat(filepath.Join(dir, name+".json")); err == nil {
			return fmt.Errorf("stage %s already exists in %s; refusing to overwrite", name, dir)
		}
		if err := writeJSON(filepath.Join(dir, name+".json"), sf); err != nil {
			return err
		}
		fmt.Fprintf(prog, "stage %s written (%.1fs)\n", name, sf.Seconds)
		return nil
	}
	timed := func(f func()) float64 { t0 := time.Now(); f(); return time.Since(t0).Seconds() }
	kind := stage
	if strings.HasPrefix(stage, "pre-") {
		kind = "pre"
	} else if strings.HasPrefix(stage, "post-") {
		kind = "post"
	}
	switch kind {
	case "pre":
		n := stage[4:]
		if _, ok := banks[n]; !ok {
			return fmt.Errorf("no %s bank in this configuration", n)
		}
		net := NewMLP(p.Seed)
		var rs []CaseResult
		sec := timed(func() { rs = EvalBank(net, toCases(banks[n]), EvalConfig{Cap: p.CapE}) })
		return done(stage, stageFile{Seconds: sec, Results: rs, Weights: WeightsSHA(net)})
	case "train":
		net := NewMLP(p.Seed)
		var lg TrainLog
		var terr error
		sec := timed(func() {
			lg, terr = TrainRounds(net, TrainConfig{Seed: p.Seed, Rounds: p.Rounds, PerRound: p.Per, Cap: p.CapT,
				Exclude: ExcludeSetFor(all), ExcludeRender: RenderExcludeFor(all), ExcludeKeys: p.ExcludeKeys})
		})
		if terr != nil {
			return terr
		}
		if _, err := os.Stat(weightsPath); err == nil {
			return fmt.Errorf("weights already exist in %s; refusing to overwrite", dir)
		}
		net.Save(weightsPath)
		return done("train", stageFile{Seconds: sec, Train: &lg, Weights: WeightsSHA(net), Level: net.Level})
	case "post":
		n := stage[5:]
		if _, ok := banks[n]; !ok {
			return fmt.Errorf("no %s bank in this configuration", n)
		}
		tr, err := readStage(dir, "train", fp)
		if err != nil {
			return err
		}
		net := LoadMLP(weightsPath)
		if net == nil {
			return fmt.Errorf("cannot load trained weights")
		}
		if WeightsSHA(net) != tr.Weights {
			return fmt.Errorf("stored weights do not match the hash recorded by the train stage")
		}
		var rs []CaseResult
		sec := timed(func() { rs = EvalBank(net, toCases(banks[n]), EvalConfig{Cap: p.CapE}) })
		return done(stage, stageFile{Seconds: sec, Results: rs, Weights: WeightsSHA(net)})
	case "report":
		rep := D1TrainReport{Seed: p.Seed, Rounds: p.Rounds, PerRound: p.Per, TrainCap: p.CapT, EvalCap: p.CapE, Library: "none",
			Overlap: map[string]TrainEvalOverlap{}, SecondsPhase: map[string]float64{}}
		tr, err := readStage(dir, "train", fp)
		if err != nil {
			return err
		}
		rep.G1 = p.G1Info
		rep.Train, rep.PostWeights, rep.FinalLevel = *tr.Train, tr.Weights, tr.Level
		rep.SecondsPhase["train"] = tr.Seconds
		for _, n := range names {
			pre, err := readStage(dir, "pre-"+n, fp)
			if err != nil {
				return err
			}
			post, err := readStage(dir, "post-"+n, fp)
			if err != nil {
				return err
			}
			rep.PreWeights = pre.Weights
			rep.SecondsPhase["eval_pre_"+n], rep.SecondsPhase["eval_post_"+n] = pre.Seconds, post.Seconds
			rep.Overlap[n] = OverlapTrainEval(rep.Train, toCases(banks[n]))
			rep.Banks = append(rep.Banks, evalBankReport(n, banks[n], pre.Results, post.Results))
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		return enc.Encode(rep)
	}
	return fmt.Errorf("unknown stage %q", stage)
}
